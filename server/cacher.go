// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto"
	"fmt"
	"io"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/patrickmn/go-cache"
)

// Request represents a challenge-based request.
type Request struct {
	requestID string
	challenge []byte
	userID    string
	appID     string
}

// AuthenticationRequest is a specialization of Request. It has an additional
// field, counter, that is used during authentication.
type AuthenticationRequest struct {
	Request
	counter int
}

// NewCacher creates a new cacher that cleans itself after a set amount of time.
func NewCacher(expiration time.Duration, clean time.Duration) *Cacher {
	return &Cacher{
		expiration:             expiration,
		clean:                  clean,
		registrationRequests:   cache.New(expiration, clean),
		authenticationRequests: cache.New(expiration, clean),
	}
}

// Cacher holds various requests. If they are not found, it hits the database.
type Cacher struct {
	expiration             time.Duration
	clean                  time.Duration
	registrationRequests   *cache.Cache
	authenticationRequests *cache.Cache
	db                     *gorm.DB // Templated on and holds long-term requests
}

// GetRegistrationRequest returns the registration request for a particular
// request ID.
func (c *Cacher) GetRegistrationRequest(id string) (*Request, error) {
	if val, ok := c.registrationRequests.Get(id); ok {
		rr := val.(Request) // We must convert and then take the address
		ptr := &rr
		return ptr, nil
	}
	ltr := LongTermRequest{}
	var hashedID []byte
	h := crypto.SHA256.New()
	io.WriteString(h, id)
	return c.db.FindAndDelete(LongTermRequest{hashedID: h.Sum(nil)}, ltr), nil
}

// GetAuthenticationRequest returns the string(h.Sum(nil))n request for a
// particular request ID.
func (c *Cacher) GetAuthenticationRequest(id string) (*AuthenticationRequest, error) {
	val, ok := c.authenticationRequests.Get(id)
	if ok {
		ar := val.(AuthenticationRequest)
		ptr := &ar
		return ptr, nil
	}
	return nil, fmt.Errorf("Could not find authentication request with id %s", id)
}

// RecordRegistrationRequest puts a registration request inside the cache.
func (c *Cacher) RecordRegistrationRequest(id string, r Request) {
	c.registrationRequests.Set(id, r, c.expiration)
}

// RecordAuthenticationRequest puts an authentication request inside the cache.
func (c *Cacher) RecordAuthenticationRequest(id string, r AuthenticationRequest) {
	c.authenticationRequests.Set(id, r, c.expiration)
}
