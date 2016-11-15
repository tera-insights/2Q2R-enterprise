// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto"
	"crypto/rand"
	"io"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/tstranex/u2f"
)

// registrationRequest stores data used during the registration of a new
// device, etc.
type registrationRequest struct {
	RequestID       string
	Challenge       *u2f.Challenge
	AppID           string
	UserID          string
	OrigininatingIP string
}

// adminRegistrationRequest is the what admins use to add their initial second
// factor.
type adminRegistrationRequest struct {
	Challenge []byte
}

// Holds various requests. If they are not found, it hits the database.
type cacher struct {
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
}

func newCacher(c *Config) *cacher {
	cache := cacher{
		baseURL:                c.getBaseURLWithProtocol(),
		expiration:             c.ExpirationTime,
		clean:                  c.CleanTime,
		registrationRequests:   cache.New(c.ExpirationTime, c.CleanTime),
		authenticationRequests: cache.New(c.ExpirationTime, c.CleanTime),
		challengeToRequestID:   cache.New(c.ExpirationTime, c.CleanTime),
		admins:                 cache.New(c.ExpirationTime, c.CleanTime),
		signingKeys:            cache.New(c.ExpirationTime, c.CleanTime),
		adminRegistrations:     cache.New(c.ExpirationTime, c.CleanTime),
	}
	return &cache
}

// GetRegistrationRequest returns the registration request for a particular
// request ID.
func (c *cacher) GetRegistrationRequest(id string) (*registrationRequest,
	error) {
	if val, ok := c.registrationRequests.Get(id); ok {
		// We must convert and then take the address
		rr := val.(registrationRequest)
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
	r := registrationRequest{
		RequestID: id,
		Challenge: challenge,
		AppID:     ltr.AppID,
	}
	c.registrationRequests.Set(id, r, c.expiration)
	s := EncodeBase64(r.Challenge.Challenge)
	c.challengeToRequestID.Set(s, id, c.expiration)
	return &r, nil
}

// SetRegistrationRequest puts a RegistrationRequest into the cache.
func (c *cacher) SetRegistrationRequest(id string, r registrationRequest) {
	c.registrationRequests.Set(id, r, c.expiration)
	s := EncodeBase64(r.Challenge.Challenge)
	c.challengeToRequestID.Set(s, id, c.expiration)
}

// NewAdminRegisterRequest stores a new admin, signing key, and registration
// request for a particular request ID. If the request is successful, the admin
// is saved to the DB.
func (c *cacher) NewAdminRegisterRequest(id string, a Admin, sk SigningKey) {
	c.admins.Set(id, a, c.expiration)
	c.signingKeys.Set(id, a, c.expiration)

	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		c.admins.Delete(id)
		c.signingKeys.Delete(id)
		optionalInternalPanic(err, "Failed to generate echallenge for admin")
	}

	c.adminRegistrations.Set(id, adminRegistrationRequest{
		Challenge: bytes,
	}, c.expiration)
}

// GetAdmin returns the admin for a particular request ID.
func (c *cacher) GetAdmin(id string) (Admin, SigningKey, error) {
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
