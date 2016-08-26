package common

// AppIDInfoReply is the response to `GET /v1/info/:appID`.
type AppIDInfoReply struct {
	appName       string // string specifying displayable app name
	baseURL       string // string specifying the prefix of all routes
	appID         string // base64Web encoded appID
	serverPubKey  string // The server public key. Depends on key type
	serverKeyType string // The key type. Only P256 supported for now.
}

// RegistrationRequestReply is the response to `POST /v1/register/request`.
type RegistrationRequestReply struct {
	id          string // base64Web encoded random reply id
	registerUrl string // Url at which the registration iframe can be found. Pass to frontend.
}

// AuthRequestReply is the response to `POST /v1/auth/request`.
type AuthRequestReply struct {
	id      string // base64Web encoded random reply id
	authUrl string // Url at which the registration iframe can be found. Pass to frontend.
}
