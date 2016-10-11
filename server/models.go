// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import "github.com/jinzhu/gorm"

// AppInfo is the Gorm model that holds information about an app.
type AppInfo struct {
	gorm.Model

	AppID   string `json:"appID"`
	AppName string `json:"appName"`
}

// AppServerInfo is the Gorm model that holds information about an app server.
type AppServerInfo struct {
	gorm.Model

	ServerID   string `json:"serverID"`
	ServerName string `json:"serverName"`

	// Base URL for users to connect to
	BaseURL string `json:"baseURL"`

	// A server can only serve one app
	AppID string `json:"appID"`

	// P256, etc.
	KeyType string `json:"keyType"`

	// JSON
	PublicKey []byte `json:"publicKey"`

	// JSON array containing a subset of ["Register", "Delete", "Login"]
	Permissions string `json:"permissions"`

	// Either token or DSA
	AuthType string `json:"authType"`
}

// LongTermRequest is the Gorm model for a long-term registration request set
// up by an admin.
type LongTermRequest struct {
	gorm.Model

	HashedRequestID string `json:"hashedRequestID"` // sha-256 hashed
	AppID           string `json:"appID"`
}

// Key is the Gorm model for a user's stored public key.
type Key struct {
	gorm.Model

	// base-64 web encoded version of the KeyHandle in MarshalledRegistration
	KeyID  string `gorm:"primary_key" json:"keyID"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	UserID string `json:"userID"`
	AppID  string `json:"appID"`

	// unmarshalled by go-u2f
	MarshalledRegistration []byte `json:"marshalledRegistration"`
	Counter                uint32 `json:"counter"`
}

// Admin is the Gorm model for a (super-) admin.
type Admin struct {
	gorm.Model

	AdminID     string `gorm:"primary_key" json:"activeID"` // can be joined with Key.UserID
	Active      bool   `json:"active"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Permissions string `json:"permissions"` // JSON-encoded array
	SuperAdmin  bool   `json:"superAdmin"`  // if so, this essentially has all the permissions
	IV          string `json:"iv"`          // encoded using encodeBase64
	Seed        string `json:"seed"`        // same encoding
	PublicKey   []byte `json:"publicKey"`
}
