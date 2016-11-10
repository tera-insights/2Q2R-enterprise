// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/tstranex/u2f"
)

type authReq struct {
	RequestID string
	Challenge *u2f.Challenge
	KeyHandle string
	AppID     string
	UserID    string
}

type authenticator struct {
	authReqs             *cache.Cache
	challengeToRequestID *cache.Cache
	expiration           time.Duration
	q                    queue
}

func newAuthenticator(c *Config) *authenticator {
	rcet := c.RecentlyCompletedExpirationTime
	ct := c.CleanTime
	return &authenticator{
		authReqs:             cache.New(rcet, ct),
		challengeToRequestID: cache.New(rcet, ct),
		expiration:           c.ExpirationTime,
		q:                    newQueue(rcet, ct, c.ListenerExpirationTime, ct),
	}
}

// GetRequest returns the request for a particular request ID.
func (a *authenticator) GetRequest(id string) (*authReq, error) {
	if val, found := a.authReqs.Get(id); found {
		ar := val.(authReq)
		ptr := &ar
		return ptr, nil
	}
	return nil, errors.Errorf("Could not find auth request with id %s", id)
}

// PutRequest puts an authReq into the cache.
func (a *authenticator) PutRequest(id string, r authReq) {
	a.authReqs.Set(id, r, a.expiration)
	s := EncodeBase64(r.Challenge.Challenge)
	a.challengeToRequestID.Set(s, id, a.expiration)
}

// SetKey sets the key handle used by an authReq.
func (a *authenticator) SetKey(id, handle string) error {
	if val, found := a.authReqs.Get(id); found {
		ar := val.(authReq)
		ar.KeyHandle = handle
		a.authReqs.Set(id, ar, a.expiration)
		return nil
	}
	return errors.Errorf("Could not find auth request with id %s", id)
}

func (a *authenticator) MarkCompleted(id string) error {
	return a.q.MarkCompleted(id)
}

// Listen returns a chan that emits an HTTP status code corresponding to the
// authentication request. It also returns a pointer to the request so that, if
// appropriate, handlers can attach cookies.
func (a *authenticator) Listen(id string) (chan int, *authReq, error) {
	ar, err := a.GetRequest(id)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not listen to unknown request")
	}
	return a.q.Listen(id), ar, nil
}
