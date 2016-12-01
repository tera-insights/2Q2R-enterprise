// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package security

import (
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"time"

	"sync"

	"github.com/pkg/errors"
	"github.com/telehash/gogotelehash/e3x/cipherset/cs1a/ecdh"
)

// KeyGen stores and regenerates keys for the admin-frontend.
type KeyGen struct {
	stateLock   sync.RWMutex
	shared      map[string][]byte
	priv        []byte
	privExpires time.Time
}

// NewKeyGen creates a new KeyGen that automatically regenerates its private
// key.
func NewKeyGen() *KeyGen {
	priv, _, _, err := elliptic.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(errors.Wrapf(err, "Could not generate private key"))
	}

	kg := KeyGen{
		sync.RWMutex{},
		make(map[string][]byte),
		priv,
		time.Now().Add(5 * time.Minute),
	}

	go func() {
		for true {
			newPriv, _, _, err := elliptic.GenerateKey(elliptic.P256(),
				rand.Reader)
			if err != nil {
				panic(errors.Wrapf(err, "Could not regenerate private key"))
			}
			kg.stateLock.Lock()
			kg.priv = newPriv
			kg.privExpires = time.Now().Add(5 * time.Minute)
			kg.stateLock.Unlock()
			time.Sleep(5 * time.Minute)
		}
	}()

	return &kg
}

// PutShared stores `shared` in the cache at the index determined by x, y,
// and the elliptic curve.
func (kg *KeyGen) PutShared(x, y *big.Int, shared []byte) {
	defer func() {
		kg.stateLock.Unlock()
	}()
	kg.stateLock.Lock()
	index := string(elliptic.Marshal(elliptic.P256(), x, y))
	kg.shared[index] = shared
}

// GetAdminPriv returns the current admin-frontend private key along with how
// much longer it will be valid for.
func (kg *KeyGen) GetAdminPriv() ([]byte, time.Duration, error) {
	defer func() {
		kg.stateLock.RUnlock()
	}()
	kg.stateLock.RLock()
	return kg.priv, kg.privExpires.Sub(time.Now()), nil
}

// GetShared returns the shared key between x, y and `priv`. If
// `priv == nil`, the admin-frontend private key is used.
// If the key has not yet been generated, it is computed using ECDH and stored
// in the cache.
func (kg *KeyGen) GetShared(x, y *big.Int, priv []byte) []byte {
	defer func() {
		kg.stateLock.Unlock()
	}()
	kg.stateLock.Lock()

	index := string(elliptic.Marshal(elliptic.P256(), x, y))
	if val, found := kg.shared[index]; found {
		return val
	}

	if priv == nil {
		kg.shared[index] = ecdh.ComputeShared(elliptic.P256(), x, y, kg.priv)
	} else {
		kg.shared[index] = ecdh.ComputeShared(elliptic.P256(), x, y, priv)
	}
	return kg.shared[index]
}
