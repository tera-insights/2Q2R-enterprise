// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import "github.com/jinzhu/gorm"

// AppInfo is the Gorm model that holds information about an app.
type AppInfo struct {
	gorm.Model

	AppID   string `gorm:"not null" json:"appID"`
	AppName string `gorm:"not null" json:"appName"`
}

// AppServerInfo is the Gorm model that holds information about an app server.
type AppServerInfo struct {
	gorm.Model

	ServerID   string `gorm:"not null" json:"serverID"`
	ServerName string `gorm:"not null;unique" json:"serverName"`

	// Base URL for users to connect to
	BaseURL string `gorm:"not null" json:"baseURL"`

	// A server can only serve one app
	AppID string `gorm:"not null" json:"appID"`

	// P256, etc.
	KeyType string `gorm:"not null" json:"keyType"`

	// JSON
	PublicKey []byte `gorm:"not null;unique" json:"publicKey"`

	// JSON array containing a subset of ["Register", "Delete", "Login"]
	Permissions string `gorm:"not null" json:"permissions"`

	// Either token or DSA
	AuthType string `gorm:"not null" json:"authType"`
}

// LongTermRequest is the Gorm model for a long-term registration request set
// up by an admin.
type LongTermRequest struct {
	gorm.Model

	HashedRequestID string `gorm:"not null" json:"hashedRequestID"` // sha-256 hashed
	AppID           string `gorm:"not null" json:"appID"`
}

// Key is the Gorm model for a user's stored public key.
type Key struct {
	gorm.Model

	// base-64 web encoded version of the KeyHandle in MarshalledRegistration
	KeyID  string `gorm:"primary_key;not null" json:"keyID"`
	Type   string `gorm:"not null" json:"type"`
	Name   string `gorm:"not null" json:"name"`
	UserID string `gorm:"not null" json:"userID"`
	AppID  string `gorm:"not null" json:"appID"`

	// unmarshalled by go-u2f
	MarshalledRegistration []byte `gorm:"not null" json:"marshalledRegistration"`
	Counter                uint32 `gorm:"not null" json:"counter"`
}

// Admin is the Gorm model for a (super-) admin.
type Admin struct {
	gorm.Model

	AdminID     string `gorm:"primary_key;not null" json:"activeID"` // can be joined with Key.UserID
	Status      string `gorm:"not null" json:"status"`               // either active or inactive
	Name        string `gorm:"not null" json:"name"`
	Email       string `gorm:"not null" json:"email"`
	Permissions string `gorm:"not null" json:"permissions"` // JSON-encoded array
	Role        string `gorm:"not null" json:"role"`        // if superadmin, this has all permissions
	IV          string `gorm:"not null" json:"iv"`          // encoded using encodeBase64
	Seed        string `gorm:"not null" json:"seed"`        // same encoding
	PublicKey   []byte `gorm:"not null" json:"publicKey"`
}
