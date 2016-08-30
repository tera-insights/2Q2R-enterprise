// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

// NewAppInfoRequest is the request to `POST /v1/app/new`.
type NewAppRequest struct {
	AppName  string `json:"appName"`
	AuthType string `json:"authType"`
	AuthData string `json:"authData"`
}

// NewAppReply is the response to `POST /v1/app/new`.
type NewAppReply struct {
	AppID string `json:"appID"`
}

// AppIDInfoReply is the response to `GET /v1/info/:appID`.
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

// MakeAppIDInfoReply creates a new `AppIDInfoReply`.
func MakeAppIDInfoReply(name string, url string, id string, pub string,
	kt string) AppIDInfoReply {
	return AppIDInfoReply{name, url, id, pub, kt}
}

// RegistrationRequestReply is the response to `POST /v1/register/request`.
type RegistrationRequestReply struct {
	// base64Web encoded random reply id
	ID string `json:"id"`

	// Url at which the registration iframe can be found. Pass to frontend.
	RegisterURL string `json:"registerUrl"`
}

// AuthRequestReply is the response to `POST /v1/auth/request`.
type AuthRequestReply struct {
	// base64Web encoded random reply id
	ID string `json:"id"`

	// Url at which the registration iframe can be found. Pass to frontend.
	AuthURL string `json:"authUrl"`
}
