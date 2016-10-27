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
	ID    string `json:"hashedRequestID"` // sha-256 hashed
	AppID string `json:"appID"`
}

// Key is the Gorm model for a user's stored public key.
type Key struct {
	// base-64 web encoded version of the KeyHandle in MarshalledRegistration
	ID     string `json:"keyID"`
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
	ID                  string `json:"activeID"` // can be joined with Key.UserID
	Status              string `json:"status"`   // either active or inactive
	Name                string `json:"name"`
	Email               string `json:"email"`
	Permissions         string `json:"permissions"`         // JSON-encoded array
	Role                string `json:"role"`                // if superadmin, this has all permissions
	PrimarySigningKeyID string `json:"primarySigningKeyID"` // FK into the SigningKey relation
}

// Permission is the schema for an admin's permissions.
type Permission struct {
	AdminID string `gorm:"primary_key" json:"adminID"`

	// If `AppID == 1`, then this is a global permission
	AppID string `gorm:"primary_key" json:"appID"`

	// Must be inside the valid list of permissions
	Permission string `gorm:"primary_key" json:"permission"`
}

// SigningKey is the Gorm model for keys that the admin uses to sign things.
type SigningKey struct {
	ID        string `json:"signingKeyID"`
	IV        string `json:"iv"`        // encoded using encodeBase64
	Salt      string `json:"salt"`      // same encoding
	PublicKey string `json:"publicKey"` // same encoding
}

// KeySignature is the Gorm model for signatures of both signing and
// second-factor keys.
type KeySignature struct {
	// base-64 web encoded
	// "1" if the signing public key is the TI public key
	SigningPublicKey string `gorm:"primary_key"`
	SignedPublicKey  string `gorm:"primary_key"`

	Type    string // either "signing" or "second-factor"
	OwnerID string // the admin's ID for `type == "signing"`, user's ID else

	// signature of the sha-256 of: SignedPublicKey | Type | OwnerID
	Signature string // same encoding as above
}
