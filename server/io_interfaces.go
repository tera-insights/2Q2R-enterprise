// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

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

	// base64Web encoded appID
	AppID string `json:"appID"`

	// The server public key. Depends on key type
	ServerPubKey string `json:"serverPubKey"`

	// The key type. Only P256 supported for now.
	ServerKeyType string `json:"serverKeyType"`
}

// NewServerRequest is the request to `POST /v1/admin/server/new`.
type NewServerRequest struct {
	ServerName  string `json:"serverName"`
	AppID       string `json:"appID"`
	BaseURL     string `json:"baseURL"`
	KeyType     string `json:"keyType"`
	PublicKey   string `json:"publicKey"`
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

// RegistrationSetupRequest is the request to `POST /v1/register/request`.
type RegistrationSetupRequest struct {
	AppID     string `json:"appID"`
	Timestamp string `json:"timestamp"`
	UserID    string `json:"userID"`
}

// RegistrationRequestReply is the response to `POST /v1/register/request`.
type RegistrationRequestReply struct {
	// base64Web encoded random reply id
	RequestID string `json:"id"`

	// Url at which the registration iframe can be found. Pass to frontend.
	RegisterURL string `json:"registerUrl"`
}

// AuthenticationSetupRequest is the request to `POST /v1/auth/request`.
type AuthenticationSetupRequest struct {
	AppID     string `json:"appID"`
	Timestamp string `json:"timestamp"`
	UserID    string `json:"userID"`
	KeyID     string `json:"keyID"`
}

// AuthRequestReply is the response to `POST /v1/auth/request`.
type AuthRequestReply struct {
	// base64Web encoded random reply id
	RequestID string `json:"id"`

	// Url at which the registration iframe can be found. Pass to frontend.
	AuthURL string `json:"authUrl"`
}
