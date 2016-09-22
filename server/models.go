// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/ecdsa"

	"github.com/jinzhu/gorm"
)

// AppInfo is the Gorm model that holds information about an app.
type AppInfo struct {
	gorm.Model

	AppID   string
	AppName string
}

// AppServerInfo is the Gorm model that holds information about an app server.
type AppServerInfo struct {
	gorm.Model

	ServerID string

	ServerName string

	// Base URL for users to connect to
	BaseURL string

	// A server can only serve one app
	AppID string

	// P256, etc.
	KeyType string

	// JSON
	PublicKey []byte

	// JSON array containing a subset of ["Register", "Delete", "Login"]
	Permissions string

	AuthType string // Either token or DSA
}

// LongTermRequest is the Gorm model for a long-term registration request set
// up by an admin.
type LongTermRequest struct {
	gorm.Model

	HashedRequestID string
	AppID           string
	UserID          string
}

// Key is the Gorm model for a user's stored public key.
type Key struct {
	gorm.Model

	KeyID  string
	UserID string
	AppID  string
	// Raw serialization data as received from the token. Used by go-u2f.
	Raw       []byte
	PublicKey ecdsa.PublicKey
	KeyHandle []byte
	Counter   uint32
}
