// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package security

import (
	"math/rand"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

// NonceGen holds information about the nonces used to secure the admin frontend.
type NonceGen struct {
	valid  time.Duration
	nonces *cache.Cache
}

// NewNonceGen creates a new KeyGen that automatically regenerates its private
// key.
func NewNonceGen(d time.Duration) *NonceGen {
	return &NonceGen{
		d,
		cache.New(d, d),
	}
}

// GenerateNonce generates a nonce for an admin. If the admin already has a nonce,
// returns an error.
func (ng *NonceGen) GenerateNonce(adminID string) (string, time.Duration, error) {
	if _, found := ng.nonces.Get(adminID); found {
		return "", 0, errors.Errorf("Admin with id %s already has a nonce",
			adminID)
	}
	n := string(rand.Int())
	ng.nonces.Add(adminID, n, 30*time.Second)
	return n, 30 * time.Second, nil
}

// GetNonce returns the nonce for a particular admin. If the admin does not
// have one, returns an error.
func (ng *NonceGen) GetNonce(adminID string) (string, error) {
	if val, ok := ng.nonces.Get(adminID); ok {
		ng.nonces.Delete(adminID)
		return val.(string), nil
	}
	return "", errors.Errorf("No nonce for id %s", adminID)
}
