// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/tstranex/u2f"
)

// RegistrationRequest stores data used during the registration of a new
// device, etc.
type RegistrationRequest struct {
	RequestID string
	Challenge *u2f.Challenge
	AppID     string
	UserID    string
}

// AuthenticationRequest stores data used during authentication.
type AuthenticationRequest struct {
	RequestID string
	Challenge *u2f.Challenge
	KeyHandle string
	AppID     string
	UserID    string
}

// AdminRegistrationRequest is the what admins use to add their initial second
// factor.
type AdminRegistrationRequest struct {
	Challenge []byte
}

// Cacher holds various requests. If they are not found, it hits the database.
type Cacher struct {
	baseURL                string
	s                      *Server
	expiration             time.Duration
	clean                  time.Duration
	registrationRequests   *cache.Cache
	authenticationRequests *cache.Cache

	// Stores a string of the []byte challenge
	challengeToRequestID *cache.Cache

	// Request ID to admin to be saved
	admins *cache.Cache

	// Request ID to signing key to be saved
	signingKeys *cache.Cache

	// request ID to AdminRegistrationRequest
	adminRegistrations *cache.Cache

	// signed public key to true
	validPublicKeys *cache.Cache
}

// GetRegistrationRequest returns the registration request for a particular
// request ID.
func (c *Cacher) GetRegistrationRequest(id string) (*RegistrationRequest,
	error) {
	if val, ok := c.registrationRequests.Get(id); ok {
		// We must convert and then take the address
		rr := val.(RegistrationRequest)
		ptr := &rr
		return ptr, nil
	}

	// For long-term requests
	ltr := LongTermRequest{}
	h := crypto.SHA256.New()
	io.WriteString(h, id)

	// We transactionally find the long-term request and then delete it from
	// the DB.
	tx := c.s.DB.Begin()
	query := LongTermRequest{ID: string(h.Sum(nil))}
	if err := tx.First(ltr, query).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Delete(LongTermRequest{}, query).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	challenge, err := u2f.NewChallenge(c.baseURL, []string{c.baseURL})
	if err != nil {
		return nil, err
	}
	r := RegistrationRequest{
		RequestID: id,
		Challenge: challenge,
		AppID:     ltr.AppID,
	}
	c.registrationRequests.Set(id, r, c.expiration)
	s := EncodeBase64(r.Challenge.Challenge)
	c.challengeToRequestID.Set(s, id, c.expiration)
	return &r, nil
}

// GetAuthenticationRequest returns the string(h.Sum(nil))n request for a
// particular request ID.
func (c *Cacher) GetAuthenticationRequest(id string) (*AuthenticationRequest,
	error) {
	val, ok := c.authenticationRequests.Get(id)
	if ok {
		ar := val.(AuthenticationRequest)
		ptr := &ar
		return ptr, nil
	}
	return nil, errors.Errorf("Could not find authentication request with "+
		"id %s", id)
}

// SetAuthenticationRequest puts an AuthenticationRequest into the cache.
func (c *Cacher) SetAuthenticationRequest(id string, r AuthenticationRequest) {
	c.authenticationRequests.Set(id, r, c.expiration)
	s := EncodeBase64(r.Challenge.Challenge)
	c.challengeToRequestID.Set(s, id, c.expiration)
}

// SetRegistrationRequest puts a RegistrationRequest into the cache.
func (c *Cacher) SetRegistrationRequest(id string, r RegistrationRequest) {
	c.registrationRequests.Set(id, r, c.expiration)
	s := EncodeBase64(r.Challenge.Challenge)
	c.challengeToRequestID.Set(s, id, c.expiration)
}

// SetKeyForAuthenticationRequest sets the key handle used by an authentication
// request.
func (c *Cacher) SetKeyForAuthenticationRequest(requestID,
	keyHandle string) error {
	if val, found := c.authenticationRequests.Get(requestID); found {
		ar := val.(AuthenticationRequest)
		ar.KeyHandle = keyHandle
		c.authenticationRequests.Set(requestID, ar, c.expiration)
		return nil
	}
	return errors.Errorf("Could not find authentication request with id %s",
		requestID)
}

// NewAdminRegisterRequest stores a new admin, signing key, and registration
// request for a particular request ID. If the request is successful, the admin
// is saved to the DB.
func (c *Cacher) NewAdminRegisterRequest(id string, a Admin, sk SigningKey) {
	c.admins.Set(id, a, c.expiration)
	c.signingKeys.Set(id, a, c.expiration)

	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	optionalInternalPanic(err, "Failed to generate echallenge for admin")

	c.adminRegistrations.Set(id, AdminRegistrationRequest{
		Challenge: bytes,
	}, c.expiration)
}

// GetAdmin returns the admin for a particular request ID.
func (c *Cacher) GetAdmin(id string) (Admin, SigningKey, error) {
	if val, found := c.admins.Get(id); found {
		admin := val.(Admin)
		if val, found = c.signingKeys.Get(id); found {
			signingKey := val.(SigningKey)
			return admin, signingKey, nil
		}
		return admin, SigningKey{},
			errors.Errorf("Could not find signing key for request %s", id)
	}
	return Admin{}, SigningKey{}, errors.Errorf("Could not find admin for "+
		"request %s", id)
}

type stack struct {
	top *element
}

type element struct {
	data *KeySignature
	next *element
}

func (s stack) pop() *element {
	oldTop := s.top
	if oldTop != nil {
		s.top = oldTop.next
	}
	return oldTop
}

func (s stack) push(e *element) {
	e.next = s.top
	s.top = e
}

// VerifySignature validates the passed key signature using ECDSA. It uses the
// cache as much as possible to avoid database accesses. Additionally, if it
// ever reaches the Tera Insights public key (`SigningPublicKey == "1"`), then
// the signature is verified using `rsa.VerifyPSS`.
func (c *Cacher) VerifySignature(sig KeySignature) error {
	if sig.Type != "signing" {
		return errors.Errorf("Signature had type %s, not \"signing\"",
			sig.Type)
	}

	s := stack{
		top: &element{
			data: &sig,
			next: nil,
		},
	}

	// Build a stack that holds the signatures to verify.
	// Add elements to the stack when they are not yet verified.
	// Stop adding elements to the stack when we reach a key that has been
	// verified. That is, stop once we hit a key whose ID is in the cache of
	// verified IDs.
	for s.top != nil {
		toVerify := s.pop()
		if toVerify.data.SigningPublicKey == "1" {
			// Verify this Tera Insights signature using `rsa.VerifyPSS`.

			decoded, err := decodeBase64(toVerify.data.Signature)
			if err != nil {
				return errors.Wrap(err, "Could not decode signature as "+
					"web-encoded base-64 with no padding")
			}
			h := crypto.SHA256.New()
			io.WriteString(h, toVerify.data.SignedPublicKey)
			io.WriteString(h, toVerify.data.Type)
			io.WriteString(h, toVerify.data.OwnerID)
			err = rsa.VerifyPSS(c.s.pub, crypto.SHA256, h.Sum(nil), decoded,
				nil)
			if err != nil {
				return errors.Wrap(err, "Could not verify Tera Insights "+
					"signature")
			}
			c.validPublicKeys.Set(toVerify.data.SignedPublicKey, true,
				cache.NoExpiration)
		} else {
			_, found := c.validPublicKeys.Get(toVerify.data.SigningPublicKey)
			if found {
				// We have verified the key used to sign `toVerify`. Now,
				// verify this signature using `ecdsa.Verify`.
				marshalled, err := decodeBase64(toVerify.data.SigningPublicKey)
				if err != nil {
					return errors.Wrap(err, "Could not unmarshal signing "+
						"public key")
				}
				x, y := elliptic.Unmarshal(elliptic.P256(), marshalled)
				if x == nil {
					return errors.New("Signing public key was not on the " +
						"elliptic curve")
				}
				decoded, err := decodeBase64(toVerify.data.Signature)
				if err != nil {
					return errors.Wrap(err, "Could not decode signature as "+
						"web-encoded base-64 with no padding")
				}
				r, s := elliptic.Unmarshal(elliptic.P256(), decoded)
				if r == nil {
					return errors.New("Signed public key was not on the " +
						"elliptic curve")
				}
				h := crypto.SHA256.New()
				io.WriteString(h, toVerify.data.SignedPublicKey)
				io.WriteString(h, toVerify.data.Type)
				io.WriteString(h, toVerify.data.OwnerID)
				verified := ecdsa.Verify(&ecdsa.PublicKey{
					Curve: elliptic.P256(),
					X:     x,
					Y:     y,
				}, h.Sum(nil), r, s)
				if !verified {
					return errors.Errorf("Could not verify signature of "+
						"public key %s", toVerify.data.SigningPublicKey)
				}
				c.validPublicKeys.Set(toVerify.data.SignedPublicKey, true,
					cache.NoExpiration)
			} else {
				// We have not yet verified the key used to sign `toVerify`.
				// So, we need to verify both `toVerify` and the key used to
				// sign `toVerify`.
				var fetched KeySignature
				count := 0
				err := c.s.DB.Where(KeySignature{
					SignedPublicKey: toVerify.data.SigningPublicKey,
				}).Count(&count).First(&fetched).Error
				if err != nil {
					return err
				}
				if count == 0 {
					return errors.Errorf("Signing key was not in the database")
				}
				if fetched.Type != "signing" {
					return errors.Errorf("Signing key had type %s, not "+
						"\"signing\"", fetched.Type)
				}
				s.push(toVerify)
				s.push(&element{
					data: &fetched,
					next: toVerify,
				})
			}
		}
	}

	return nil
}
