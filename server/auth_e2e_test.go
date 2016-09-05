// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http/httptest"
	"testing"
)

var s = New(Config{
	8080,
	"sqlite3",
	"test.db",
})
var ts = httptest.NewServer(s.GetHandler())

// Create app and app server
// Add user to system
// Add key to system

func TestIFrameAuthentication(t *testing.T) {
	// Set up an authentication request

	// Get the registration iFrame and extract the challenge

	// In a separate routine, wait for the registration challenge to be met

	// Sign the challenge and send the result to /v1/register

	// Assert that the waiting thread came to a close
}
