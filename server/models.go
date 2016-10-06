// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import "github.com/jinzhu/gorm"

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

	// base-64 web encoded version of the KeyHandle in MarshalledRegistration
	KeyID                  string
	Type                   string
	Name                   string
	UserID                 string
	AppID                  string
	MarshalledRegistration []byte // unmarshalled by go-u2f
	Counter                uint32
}

// Admin is the Gorm model for a (super-) admin.
type Admin struct {
	gorm.Model

	AdminID     string `gorm:"primary_key"` // can be joined with Key.UserID
	Active      bool
	Name        string
	Email       string
	Permissions string // comma-separated list of permissions
	SuperAdmin  bool   // if so, this essentially has all the permissions
	IV          string // encoded using encodeBase64 (web encoding, no padding)
	Seed        string // same encoding
	PublicKey   []byte
}
