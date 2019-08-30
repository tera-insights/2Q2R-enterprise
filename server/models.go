// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

// AppInfo is the Gorm model that holds information about an app.
type AppInfo struct {
	ID      string `json:"appID"`
	AppName string `json:"appName"`
}

// AppServerInfo is the Gorm model that holds information about an app server.
type AppServerInfo struct {
	ID string `json:"serverID"`

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
}

// LongTermRequest is the Gorm model for a long-term registration request set
// up by an admin.
type LongTermRequest struct {
	ID    []byte `json:"hashedRequestID"` // sha-256 hashed
	AppID string `json:"appID"`
}

// Admin is the Gorm model for a (super-) admin.
type Admin struct {
	ID          string `json:"activeID"` // can be joined with Key.UserID
	Status      string `json:"status"`   // either active or inactive
	Name        string `json:"name"`
	Email       string `json:"email"`
	Permissions string `json:"permissions"` // JSON-encoded array

	// if superadmin, this has all permissions
	Role string `json:"role"`

	// FK into the SigningKey relation
	PrimarySigningKeyID string `json:"primarySigningKeyID"`

	// The AppID for which this admin can act. "1" for superadmins
	AdminFor string `json:"adminFor"`
}

// Permission is the schema for an admin's permissions.
type Permission struct {
	AdminID string `gorm:"primary_key" json:"adminID"`

	// If `AppID == 1`, then this is a global permission
	AppID string `gorm:"primary_key" json:"appID"`

	// Must be inside the valid list of permissions
	Permission string `gorm:"primary_key" json:"permission"`
}
