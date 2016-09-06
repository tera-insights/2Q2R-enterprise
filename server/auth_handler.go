// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
)

// CounterBasedAuthData represents auth data that uses a counter.
type CounterBasedAuthData struct {
	Counter  int
	ServerID string
}

// AuthHandler is the handler for all `/auth/` requests.
type AuthHandler struct {
	s *Server
}

// AuthRequestSetupHandler sets up a two-factor authentication request.
// POST /v1/auth/request
func (ah *AuthHandler) AuthRequestSetupHandler(w http.ResponseWriter, r *http.Request) {
	req := AuthenticationSetupRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		handleError(w, err)
	} else {
		authData := req.AuthenticationData.(CounterBasedAuthData)
		var challenge = make([]byte, ah.s.c.ChallengeLength)
		rand.Read(challenge)
		r := AuthenticationRequest{
			randString(32),
			challenge,
			authData.Counter,
		}
		ah.s.cache.SetAuthenticationRequest(r.requestID, r)
		server := AppServerInfo{}
		ah.s.DB.Model(AppServerInfo{}).Find(&server,
			AppServerInfo{ServerID: authData.ServerID})
		writeJSON(w, http.StatusOK, AuthenticationSetupReply{
			r.requestID,
			server.BaseURL + "/auth/" + r.requestID,
		})
	}
}
