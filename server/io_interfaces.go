// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"math/big"
	"time"
)

// NewAdminRequest the request to add a new admin. It is used both in HTTP
// requests and in the bootstrap script.
// REQUEST POST /admin/new/:code
type NewAdminRequest struct {
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	IV          string   `json:"iv"`   // encoded with web encoding, no padding
	Salt        string   `json:"salt"` // same encoding
	PublicKey   []byte   `json:"publicKey"`
}

// REPLY POST /admin/new/:code
type newAdminReply struct {
	RequestID   string `json:"requestID"`
	IFrameRoute string `json:"iFrameRoute"`
	WaitRoute   string `json:"waitRoute"`
}

// Request to POST /admin/admin/{adminID}
type adminUpdateRequest struct {
	Name                string `json:"name"`
	Email               string `json:"email"`
	PrimarySigningKeyID string `json:"primarySigningKeyID"`
}

// Request to POST /admin/admin/roles
type adminRoleChangeRequest struct {
	AdminID     string `json:"adminID"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	Permissions string `json:"permissions"`
}

// NewAppRequest is the request to POST /admin/app
type NewAppRequest struct {
	AppName string `json:"appName"`
}

// Request to PUT /admin/app/{appID}
type appUpdateRequest struct {
	AppName string `json:"appName"`
}

type modificationReply struct {
	NumAffected int64 `json:"numAffected"`
}

// AppIDInfoReply is the reply to `GET /v1/info/:appID`.
type AppIDInfoReply struct {
	// string specifying displayable app name
	AppName string `json:"appName"`

	// string specifying the prefix of all routes
	BaseURL string `json:"baseURL"`

	AppURL string `json:"appURL"`

	AppID string `json:"appID"`

	// base 64 encoded public key of the 2Q2R server
	PublicKey string `json:"serverPubKey"`

	// Only P256 supported for now
	KeyType string `json:"serverKeyType"`
}

type adminRegisterRequest struct {
	RequestID string  `json:"requestID"`
	R         big.Int `json:"r"` // can just be specified as a string
	S         big.Int `json:"s"`
}

// NewServerRequest is the request to POST /admin/server
type NewServerRequest struct {
	ServerName  string `json:"serverName"`
	AppID       string `json:"appID"`
	BaseURL     string `json:"baseURL"`
	KeyType     string `json:"keyType"`
	PublicKey   string `json:"publicKey"` // base-64 encoded byte array
	Permissions string `json:"permissions"`
}

// Request to PUT /admin/server/{serverID}
type serverUpdateRequest struct {
	ServerName  string `json:"serverName"`
	BaseURL     string `json:"baseURL"`
	KeyType     string `json:"keyType"`
	PublicKey   []byte `json:"publicKey"`
	Permissions string `json:"permissions"`
	AuthType    string `json:"authType"`
}

// RegistrationSetupReply is the reply to `GET /v1/register/request/:userID`.
type RegistrationSetupReply struct {
	// base64Web encoded random reply id
	RequestID string `json:"id"`

	// Url at which the registration iframe can be found. Pass to frontend.
	RegisterURL string `json:"registerUrl"`
}

// RegisterRequest is the request to `POST /v1/register`.
type RegisterRequest struct {
	Successful bool `json:"successful"`
	// Either a successfulRegistrationData or a failedRegistrationData
	Data interface{} `json:"data"`
}

// RegisterResponse is the response to `POST /v1/register`.
type RegisterResponse struct {
	Successful bool   `json:"successful"`
	Message    string `json:"message"`
}

type successfulRegistrationData struct {
	ClientData       string `json:"clientData"`       // base64 serialized client data
	RegistrationData string `json:"registrationData"` // base64 binary registration data
	DeviceName       string `json:"deviceName"`
	Type             string `json:"type"`     // device type and key type
	FCMToken         string `json:"fcmToken"` // Firebase Communicator Device token
}

type failedRegistrationData struct {
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    int    `json:"errorStatus"`
}

// AuthenticationSetupRequest is the request to `POST /v1/auth/request`.
type AuthenticationSetupRequest struct {
	AppID              string             `json:"appID"`
	Timestamp          time.Time          `json:"timestamp"`
	UserID             string             `json:"userID"`
	KeyID              string             `json:"keyID"`
	AuthenticationData AuthenticationData `json:"authentication"`
}

// AuthenticationSetupReply is the response to `POST /v1/auth/request`.
type AuthenticationSetupReply struct {
	// base64Web encoded random reply id
	RequestID string `json:"id"`

	// Url at which the registration iframe can be found. Pass to frontend.
	AuthURL string `json:"authUrl"`
}

type authenticateRequest struct {
	Successful bool        `json:"successful"`
	Data       interface{} `json:"data"`
}

// Request to `POST /v1/auth/{requestID}/challenge`
type setKeyRequest struct {
	KeyHandle string `json:"keyID"`
}

// Response to `POST /v1/auth/{requestID}/challenge`
type setKeyReply struct {
	KeyHandle string `json:"keyID"`
	Challenge string `json:"challenge"`
	Counter   uint32 `json:"counter"`
	AppID     string `json:"AppID"`
}

// Reply to `GET /v1/users/:userID`
type userExistsReply struct {
	Exists bool `json:"exists"`
}

type successfulAuthenticationData struct {
	ClientData    string `json:"clientData"`
	SignatureData string `json:"signatureData"`
}

type failedAuthenticationData struct {
	Challenge    string `json:"challenge"`
	ErrorMessage string `json:"errorMessage"`
	ErrorStatus  int    `json:"errorStatus"`
}

// Request to POST /admin/ltr
type newLTRRequest struct {
	AppID string `json:"appID"`
}

// Reply to POST /admin/ltr
type newLTRResponse struct {
	RequestID string `json:"requestID"`
}

// Request to DELETE /admin/ltr
type deleteLTRRequest struct {
	AppID           string `json:"appID"`
	HashedRequestID string `json:"hashedRequestID"`
}
