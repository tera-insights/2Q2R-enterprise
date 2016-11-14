// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package security

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"io"
	"time"

	"github.com/jinzhu/gorm"
	cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

// KeySignature is the Gorm model for signatures of both signing and
// second-factor keys.
type KeySignature struct {
	// base-64 web encoded
	// "1" if the signing public key is the TI public key
	SigningPublicKey string `gorm:"primary_key"`
	SignedPublicKey  string `gorm:"primary_key"`

	Type    string // either "signing" or "second-factor"
	OwnerID string // the admin's ID for `type == "signing"`, user's ID else

	// signature of the sha-256 of: SignedPublicKey | Type | OwnerID
	Signature string // same encoding as above
}

// KeyCache stores and checks the validity of signing keys.
type KeyCache struct {
	// Stores valid public signing keys
	validPublic *cache.Cache

	serverPub *rsa.PublicKey

	db *gorm.DB
}

// NewKeyCache uses the config to create a new KeyCache.
func NewKeyCache(et, ct time.Duration, p *rsa.PublicKey, db *gorm.DB) *KeyCache {
	return &KeyCache{
		cache.New(et, ct),
		p,
		db,
	}
}

// VerifySignature validates the passed key signature using ECDSA. It uses the
// cache as much as possible to avoid database accesses. Additionally, if it
// ever reaches the Tera Insights public key (`SigningPublicKey == "1"`), then
// the signature is verified using `rsa.VerifyPSS`.
func (kc *KeyCache) VerifySignature(sig KeySignature) error {
	if sig.Type != "signing" {
		return errors.Errorf("Signature had type %s, not \"signing\"",
			sig.Type)
	}

	// Represents a stack. New elements are added to the back of the slice.
	// Elements are also popped from the back of the slice.
	s := []*KeySignature{&sig}

	// Add elements to the slice when they are not yet verified.
	// Stop adding elements to the stack when we reach a key that has been
	// verified. That is, stop once we hit a key whose ID is in the cache of
	// verified IDs.
	for len(s) != 0 {
		toVerify := s[len(s)-1]
		s = s[:len(s)-1]
		if toVerify.SigningPublicKey == "1" {
			// Verify this Tera Insights signature using `rsa.VerifyPSS`.

			decoded, err := decodeBase64(toVerify.Signature)
			if err != nil {
				return errors.Wrap(err, "Could not decode signature as "+
					"web-encoded base-64 with no padding")
			}
			h := crypto.SHA256.New()
			io.WriteString(h, toVerify.SignedPublicKey)
			io.WriteString(h, toVerify.Type)
			io.WriteString(h, toVerify.OwnerID)
			err = rsa.VerifyPSS(kc.serverPub, crypto.SHA256, h.Sum(nil),
				decoded, nil)
			if err != nil {
				return errors.Wrap(err, "Could not verify Tera Insights "+
					"signature")
			}
			kc.validPublic.Set(toVerify.SignedPublicKey, true,
				cache.NoExpiration)
		} else {
			_, found := kc.validPublic.Get(toVerify.SigningPublicKey)
			if found {
				// We have verified the key used to sign `toVerify`. Now,
				// verify this signature using `ecdsa.Verify`.
				marshalled, err := decodeBase64(toVerify.SigningPublicKey)
				if err != nil {
					return errors.Wrap(err, "Could not unmarshal signing "+
						"public key")
				}
				x, y := elliptic.Unmarshal(elliptic.P256(), marshalled)
				if x == nil {
					return errors.New("Signing public key was not on the " +
						"elliptic curve")
				}
				decoded, err := decodeBase64(toVerify.Signature)
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
				io.WriteString(h, toVerify.SignedPublicKey)
				io.WriteString(h, toVerify.Type)
				io.WriteString(h, toVerify.OwnerID)
				verified := ecdsa.Verify(&ecdsa.PublicKey{
					Curve: elliptic.P256(),
					X:     x,
					Y:     y,
				}, h.Sum(nil), r, s)
				if !verified {
					return errors.Errorf("Could not verify signature of "+
						"public key %s", toVerify.SigningPublicKey)
				}
				kc.validPublic.Set(toVerify.SignedPublicKey, true,
					cache.NoExpiration)
			} else {
				// We have not yet verified the key used to sign `toVerify`.
				// So, we need to verify both `toVerify` and the key used to
				// sign `toVerify`.
				var fetched KeySignature
				count := 0
				err := kc.db.Where(KeySignature{
					SignedPublicKey: toVerify.SigningPublicKey,
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
				s = append(s, toVerify, &fetched)
			}
		}
	}

	return nil
}

// Copied from go-u2f and 2q2r/server
func decodeBase64(s string) ([]byte, error) {
	for i := 0; i < len(s)%4; i++ {
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}