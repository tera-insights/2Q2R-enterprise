// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package crypto

import (
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/telehash/gogotelehash/e3x/cipherset/cs1a/ecdh"
)

// KeyGen stores and regenerates keys for the admin-frontend.
type KeyGen struct {
	// Send `generated` new private keys for the admin-frontend
	generated chan []byte

	// Send where to put out the current admin private key
	privOut chan chan wrappedKey

	// Send requests for shared keys
	sharedReq chan sharedReq

	// Send shared keys here
	sharedIn chan sharedIn
}

type wrappedKey struct {
	key     []byte
	expires time.Time
}

type sharedReq struct {
	x, y *big.Int
	priv []byte
	out  chan []byte
}

type sharedIn struct {
	x, y   *big.Int
	shared []byte
}

// NewKeyGen creates a new KeyGen that automatically regenerates its private
// key.
func NewKeyGen() *KeyGen {
	kc := KeyGen{
		make(chan []byte),
		make(chan chan wrappedKey),
		make(chan sharedReq),
		make(chan sharedIn),
	}
	go kc.listen()
	return &kc
}

// PutSharedKey stores `shared` in the cache at the index determined by x, y,
// and the elliptic curve.
func (kg *KeyGen) PutSharedKey(x, y *big.Int, shared []byte) {
	kg.sharedIn <- sharedIn{x, y, shared}
}

// GetAdminPriv returns the current admin-frontend private key along with how
// much longer it will be valid for.
func (kg *KeyGen) GetAdminPriv() ([]byte, time.Duration, error) {
	out := make(chan wrappedKey)
	kg.privOut <- out
	wk := <-out
	return wk.key, wk.expires.Sub(time.Now()), nil
}

// GetShared returns the shared key between x, y and `priv`. If
// `priv == nil`, the admin-frontend private key is used.
// If the key has not yet been generated, it is computed using ECDH and stored
// in the cache.
func (kg *KeyGen) GetShared(x, y *big.Int, priv []byte) []byte {
	out := make(chan []byte)
	kg.sharedReq <- sharedReq{x, y, priv, out}
	return <-out
}

func (kg *KeyGen) listen() {
	shared := cache.New(cache.NoExpiration, cache.NoExpiration)
	priv, _, _, err := elliptic.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(errors.Wrapf(err, "Could not generate private key"))
	}

	wk := wrappedKey{
		priv,
		time.Now().Add(5 * time.Minute),
	}

	go func() {
		priv, _, _, err := elliptic.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(errors.Wrapf(err, "Could not generate private key"))
		}
		kg.generated <- priv
		time.Sleep(5 * time.Minute)
	}()

	if err != nil {
		panic(errors.Wrapf(err, "Could not generate private key"))
	}

	for true {
		select {
		case out := <-kg.privOut:
			out <- wk
		case sr := <-kg.sharedReq:
			index := string(elliptic.Marshal(elliptic.P256(), sr.x, sr.y))
			if val, found := shared.Get(index); found {
				sr.out <- val.([]byte)
			} else {
				var priv []byte
				if sr.priv == nil {
					priv = wk.key
				} else {
					priv = sr.priv
				}
				key := ecdh.ComputeShared(elliptic.P256(), sr.x, sr.y, priv)
				sr.out <- key
			}
		case si := <-kg.sharedIn:
			index := string(elliptic.Marshal(elliptic.P256(), si.x, si.y))
			shared.Set(index, si.shared, cache.NoExpiration)
		case priv := <-kg.generated:
			wk = wrappedKey{priv, time.Now().Add(5 * time.Minute)}
		}
	}
}
