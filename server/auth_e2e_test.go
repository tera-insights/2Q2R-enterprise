// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import "testing"

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

	// Set up an authentication request

	// Get the registration iFrame and extract the challenge

	// In a separate routine, wait for the registration challenge to be met

	// Sign the challenge and send the result to /v1/register

	// Assert that the waiting thread came to a close
}
