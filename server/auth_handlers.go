// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

// MakeAuthRequestHandler sets up a two-factor authentication request.
// POST /v1/auth/request
func MakeAuthRequestHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// GetAuthResultHandler checks the result of the authentication. Blocks until
// the authentication is complete.
// GET /v1/auth/:id/wait
func GetAuthResultHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// AuthenticateHandler performs the actual two-factor authentication.
// POST /v1/auth
func AuthenticateHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// GetAuthenticationIFrameHandler returns the iFrame that allows user
// interaction.
// GET /auth/:id
func GetAuthenticationIFrameHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
