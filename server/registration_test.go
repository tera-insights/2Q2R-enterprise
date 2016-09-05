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

func TestRegister(t *testing.T) {
	// Similar process to auth_e2e_test.go
}
