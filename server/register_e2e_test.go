// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"
	"testing"
)

// Create app server
// Add user to system
// Add key to system

func TestIFrameAuthentication(t *testing.T) {
	serverName := "register_e2e_test"

	// Create app server
	postJSON("/v1/admin/server/new", NewServerRequest{
		ServerName:  serverName,
		AppID:       goodAppID,
		BaseURL:     goodBaseURL,
		KeyType:     goodKeyType,
		PublicKey:   goodPublicKey,
		Permissions: goodPermissions,
	})

	// Add user to system
	res, _ := postJSON("/v1/admin/user/new", NewUserRequest{})
	userData := new(NewUserReply)
	unmarshalJSONBody(res, userData)

	// Set up a registration request
	res, _ = http.Get("/v1/register/request" + userData.UserID)
	setupInfo := new(RegistrationSetupReply)
	unmarshalJSONBody(res, setupInfo)

	// Get the registration iFrame and extract the challenge
	gleanedData := new(registerData)
	extractEmbeddedData("/register/"+setupInfo.RequestID, gleanedData)

	// In a separate routine, wait for the registration challenge to be met

	// Sign the challenge and send the result to /v1/register

	// Assert that the waiting thread came to a close
}
