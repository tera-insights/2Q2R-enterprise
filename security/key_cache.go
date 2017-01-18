// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package security

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"io"
	"math/big"
	"time"

	"github.com/jinzhu/gorm"
	cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/telehash/gogotelehash/e3x/cipherset/cs1a/ecdh"
)

// Key is the Gorm model for a user's stored public key.
type Key struct {
	// base-64 web encoded version of the KeyHandle in MarshalledRegistration
	ID     string `json:"keyID"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	UserID string `json:"userID"`
	AppID  string `json:"appID"`

	// unmarshalled by go-u2f
	MarshalledRegistration []byte `json:"marshalledRegistration"`
	Counter                uint32 `json:"counter"`
}

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
	validPublic      *cache.Cache // Stores valid public signing keys
	secondFactorKeys *cache.Cache // Stores valid second-factor keys
	userToAppID      *cache.Cache // User ID to app ID
	shared           *cache.Cache // marshalled public x, y -> shared
	priv             []byte

	serverPub *rsa.PublicKey

	db *gorm.DB
}

// SigningKey is the Gorm model for keys that the admin uses to sign things.
type SigningKey struct {
	ID        string `json:"signingKeyID"`
	IV        string `json:"iv"`        // encoded using encodeBase64
	Salt      string `json:"salt"`      // same encoding
	PublicKey string `json:"publicKey"` // same encoding
}

// NewKeyCache uses the config to create a new KeyCache.
func NewKeyCache(et, ct time.Duration, p *rsa.PublicKey, db *gorm.DB, priv []byte) *KeyCache {
	return &KeyCache{
		cache.New(et, ct),
		cache.New(et, ct),
		cache.New(et, ct),
		cache.New(et, ct),
		priv,
		p,
		db,
	}
}

// Add2FAKey adds a second-factor key to the cache
func (kc *KeyCache) Add2FAKey(k Key) {
	kc.secondFactorKeys.Add(k.ID, k, cache.NoExpiration)
}

// Get2FAKey looks up a second-factor key based on its ID. If the key is not
// in the cache, then Get2FAKey checks the database. If the key is in the DB,
// then it is added to the cache.
func (kc *KeyCache) Get2FAKey(id string) (Key, error) {
	if val, ok := kc.secondFactorKeys.Get(id); ok {
		return val.(Key), nil
	}

	k := Key{}
	err := kc.db.First(&k, &Key{
		ID: id,
	}).Error
	if err == nil {
		kc.secondFactorKeys.Add(id, k, cache.NoExpiration)
		kc.userToAppID.Add(k.UserID, k.AppID, cache.NoExpiration)
	}
	return k, err
}

// GetAppID returns the app ID for a particular user ID. If the user ID is not
// in the userToAppID cache, then GetAppID checks the database. If the user is
// in the DB, then it is added to the cache.
func (kc *KeyCache) GetAppID(userID string) (string, error) {
	if val, ok := kc.userToAppID.Get(userID); ok {
		return val.(string), nil
	}

	k := Key{}
	err := kc.db.First(&k, &Key{
		UserID: userID,
	}).Error
	if err == nil {
		kc.userToAppID.Add(k.UserID, k.AppID, cache.NoExpiration)
	}
	return k.AppID, err
}

// Remove2FAKey a second-factor key from the cache
func (kc *KeyCache) Remove2FAKey(id string) {
	if key, ok := kc.secondFactorKeys.Get(id); ok {
		kc.userToAppID.Delete(key.(Key).UserID)
	}
	kc.secondFactorKeys.Delete(id)
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

// GetShared returns the shared key between x, y and `priv`.
// If the key has not yet been generated, it is computed using ECDH and stored
// in the cache.
func (kc *KeyCache) GetShared(x, y *big.Int) []byte {
	index := string(elliptic.Marshal(elliptic.P256(), x, y))
	if val, found := kc.shared.Get(index); found {
		return val.([]byte)
	}

	s := ecdh.ComputeShared(elliptic.P256(), x, y, kc.priv)
	kc.shared.Add(index, s, cache.NoExpiration)
	return s
}

// PutShared stores `shared` in the cache at the index determined by x, y,
// and the elliptic curve.
func (kc *KeyCache) PutShared(x, y *big.Int, shared []byte) {
	index := string(elliptic.Marshal(elliptic.P256(), x, y))
	kc.shared.Add(index, shared, cache.NoExpiration)
}

// VerifyEphemeralKey verifies the ephemeral public key proposed by an admin.
func (kc *KeyCache) VerifyEphemeralKey(ephemeralPublic, sig string, sk SigningKey) ([]byte, error) {
	// Look up signature of signing key
	var signatureOfAdminsPublic KeySignature
	if err := kc.db.First(&signatureOfAdminsPublic, KeySignature{
		SignedPublicKey: sk.PublicKey,
	}).Error; err != nil {
		return nil, err
	}

	// Assert that admin's public key has been verified
	if err := kc.VerifySignature(signatureOfAdminsPublic); err != nil {
		return nil, err
	}

	// Assert that the signature of the ephemeral key is valid
	marshalled, err := decodeBase64(ephemeralPublic)
	if err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal signing "+
			"public key")
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), marshalled)
	if x == nil {
		return nil, errors.New("Signing public key was not on the " +
			"elliptic curve")
	}

	decoded, err := decodeBase64(sig)
	if err != nil {
		return nil, errors.Wrap(err, "Could not decode signature as "+
			"web-encoded base-64 with no padding")
	}
	r, s := elliptic.Unmarshal(elliptic.P256(), decoded)
	if r == nil {
		return nil, errors.New("Signed public key was not on the " +
			"elliptic curve")
	}

	h := crypto.SHA256.New()
	io.WriteString(h, ephemeralPublic)
	verified := ecdsa.Verify(&ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}, h.Sum(nil), r, s)

	if !verified {
		return nil, errors.Errorf("Could not verify signature of ephemeral key")
	}
	return kc.GetShared(x, y), nil
}

// Copied from go-u2f and 2q2r/server
func decodeBase64(s string) ([]byte, error) {
	for i := 0; i < len(s)%4; i++ {
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
