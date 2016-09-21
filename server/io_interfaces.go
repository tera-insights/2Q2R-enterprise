// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import "time"

// NewAppRequest is the request to `POST /v1/app/new`.
type NewAppRequest struct {
	AppName string `json:"appName"`
}

// NewAppReply is the response to `POST /v1/app/new`.
type NewAppReply struct {
	AppID string `json:"appID"`
}

// AppIDInfoReply is the reply to `GET /v1/info/:appID`.
type AppIDInfoReply struct {
	// string specifying displayable app name
	AppName string `json:"appName"`

	// string specifying the prefix of all routes
	BaseURL string `json:"baseURL"`

	AppID string `json:"appID"`

	// base 64 encoded public key of the 2Q2R server
	PublicKey string `json:"serverPubKey"`

	// Only P256 supported for now
	KeyType string `json:"serverKeyType"`
}

// NewServerRequest is the request to `POST /v1/admin/server/new`.
type NewServerRequest struct {
	ServerName  string `json:"serverName"`
	AppID       string `json:"appID"`
	BaseURL     string `json:"baseURL"`
	KeyType     string `json:"keyType"`
	PublicKey   string `json:"publicKey"` // base-64 encoded byte array
	Permissions string `json:"permissions"`
}

// NewServerReply is the response to `POST `/v1/admin/server/new`.
type NewServerReply struct {
	ServerName string `json:"serverName"`
	ServerID   string `json:"serverID"`
}

// DeleteServerRequest is the request to `POST /v1/admin/server/delete`.
type DeleteServerRequest struct {
	ServerID string `json:"serverID"`
}

// AppServerInfoRequest is the request to `POST /v1/admin/server/info`.
type AppServerInfoRequest struct {
	ServerID string `json:"serverID"`
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

// NewUserRequest is the request to `POST /v1/admin/user/new`.
type NewUserRequest struct {
}

// NewUserReply is the reply to `POST /v1/admin/user/new`.
type NewUserReply struct {
	UserID string `json:"userID"`
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

// Reply to `GET /v1/users/:userID`
type userExistsReply struct {
	Exists bool `json:"exists"`
}

type successfulAuthenticationData struct {
	ClientData       string `json:"clientData"`
	RegistrationData string `json:"registrationData"`
}

type failedAuthenticationData struct {
	Challenge    string `json:"challenge"`
	ErrorMessage string `json:"errorMessage"`
	ErrorStatus  int    `json:"errorStatus"`
}
