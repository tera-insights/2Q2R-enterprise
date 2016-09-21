// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto"
	"fmt"
	"io"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/patrickmn/go-cache"
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
	AppID     string
	UserID    string
}

// Cacher holds various requests. If they are not found, it hits the database.
type Cacher struct {
	baseURL                string
	expiration             time.Duration
	clean                  time.Duration
	registrationRequests   *cache.Cache
	authenticationRequests *cache.Cache
	challengeToRequestID   *cache.Cache // Stores a string of the []byte challenge
	db                     *gorm.DB     // Templated on and holds long-term requests
}

// GetRegistrationRequest returns the registration request for a particular
// request ID.
func (c *Cacher) GetRegistrationRequest(id string) (*RegistrationRequest, error) {
	if val, ok := c.registrationRequests.Get(id); ok {
		rr := val.(RegistrationRequest) // We must convert and then take the address
		ptr := &rr
		return ptr, nil
	}

	// For long-term requests
	ltr := LongTermRequest{}
	h := crypto.SHA256.New()
	io.WriteString(h, id)
	hashedID := string(h.Sum(nil))

	// We transactionally find the long-term request and then delete it from
	// the DB.
	tx := c.db.Begin()
	query := LongTermRequest{HashedRequestID: hashedID}
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
	s := encodeBase64(r.Challenge.Challenge)
	c.challengeToRequestID.Set(s, id, c.expiration)
	return &r, nil
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

// SetAuthenticationRequest puts an AuthenticationRequest into the cache.
func (c *Cacher) SetAuthenticationRequest(id string, r AuthenticationRequest) {
	c.authenticationRequests.Set(id, r, c.expiration)
	s := encodeBase64(r.Challenge.Challenge)
	c.challengeToRequestID.Set(s, id, c.expiration)
}

// SetRegistrationRequest puts a RegistrationRequest into the cache.
func (c *Cacher) SetRegistrationRequest(id string, r RegistrationRequest) {
	c.registrationRequests.Set(id, r, c.expiration)
	s := encodeBase64(r.Challenge.Challenge)
	c.challengeToRequestID.Set(s, id, c.expiration)
}
