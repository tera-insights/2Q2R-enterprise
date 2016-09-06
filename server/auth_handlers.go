// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/rand"
	"encoding/json"
	"net/http"

	"github.com/jinzhu/gorm"
)

// CounterBasedAuthData represents auth data that uses a counter.
type CounterBasedAuthData struct {
	Counter  int
	ServerID string
}

// AuthRequestSetupHandler sets up a two-factor authentication request.
// POST /v1/auth/request
func AuthRequestSetupHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := AuthenticationSetupRequest{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			handleError(w, err)
		} else {
			authData := req.AuthenticationData.(CounterBasedAuthData)
			var challenge = make([]byte, s.c.ChallengeLength)
			rand.Read(challenge)
			r := AuthenticationRequest{
				randString(32),
				challenge,
				authData.Counter,
			}
			s.cache.SetAuthenticationRequest(r.requestID, r)
			server := AppServerInfo{}
			s.DB.Model(AppServerInfo{}).Find(&server,
				AppServerInfo{ServerID: authData.ServerID})
			writeJSON(w, http.StatusOK, AuthenticationSetupReply{
				r.requestID,
				server.BaseURL + "/auth/" + r.requestID,
			})
		}
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
