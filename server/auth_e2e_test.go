// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"
	"testing"
	"time"
)

// Create app server
// Add user to system
// Add key to system

func TestIFrameAuthentication(t *testing.T) {
	serverName := "auth_e2e_test"

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
	authData := AuthenticationData{
		Counter:  0,
		ServerID: "foo",
	}
	registrationRequest := RegistrationSetupRequest{
		AppID:              goodAppID,
		Timestamp:          time.Now(),
		UserID:             userData.UserID,
		AuthenticationData: authData,
	}
	res, _ = postJSON("/v1/register/request", registrationRequest)
	setupInfo := new(RegistrationSetupReply)
	unmarshalJSONBody(res, setupInfo)

	// Get the registration iFrame and extract the challenge
	gleanedData := new(registerData)
	extractEmbeddedData("/register/"+setupInfo.RequestID, gleanedData)

	// In a separate routine, wait for the registration challenge to be met
	c := make(chan *http.Response)
	go func() {
		res, _ := http.Get(ts.URL + "/v1/register/" + setupInfo.RequestID + "/wait")
		c <- res
	}()

	// Sign the challenge and send the result to /v1/register

	// Assert that the waiting thread came to a close
	res = <-c
	if res.StatusCode != 200 {
		t.Errorf("/v1/register/:id/wait failed with status code %d\n",
			res.StatusCode)
	}
}
