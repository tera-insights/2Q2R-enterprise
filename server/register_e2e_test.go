// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"testing"
	"time"

	"github.com/tstranex/u2f"
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

	// Sign the challenge and send the result to /v1/register
	cd := u2f.ClientData{
		Typ:       "uh",
		Challenge: string(gleanedData.Challenge),
		Origin: ts.URL,
		CIDPubKey: ,
	}
	rr := RegisterRequest{
		Successful: true,
		Data: successfulRegistrationData{
			ClientData,       // base64 serialized client data
			RegistrationData, // base64 binary registration data
			DeviceName:       "register_e2e_client",
			Type:             "2Q-2R", // device type and key type
			FCMToken,         // Firebase Communicator Device token
		},
	}

	// Assert that the waiting thread came to a close
}
