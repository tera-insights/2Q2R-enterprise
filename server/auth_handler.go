// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/rand"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
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
		var challenge = make([]byte, ah.s.c.ChallengeLength)
		rand.Read(challenge)
		r := AuthenticationRequest{
			randString(32),
			challenge,
			req.AuthenticationData.Counter + 1,
		}
		ah.s.cache.SetAuthenticationRequest(r.requestID, r)
		server := AppServerInfo{}
		ah.s.DB.Model(AppServerInfo{}).Find(&server,
			AppServerInfo{ServerID: req.AuthenticationData.ServerID})
		writeJSON(w, http.StatusOK, AuthenticationSetupReply{
			r.requestID,
			server.BaseURL + "/auth/" + r.requestID,
		})
	}
}

// AuthIFrameHandler returns the iFrame that is used to perform authentication.
// GET /register/:id
func (ah *AuthHandler) AuthIFrameHandler(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	t, err := template.ParseFiles("../assets/all.html")
	if err != nil {
		handleError(w, err)
		return
	}
	data, err := json.Marshal(authenticateData{
		RequestID: requestID,
		Counter:   1,
	})
	if err != nil {
		handleError(w, err)
		return
	}
	t.Execute(w, templateData{
		Name: "Authentication",
		ID:   "auth",
		Data: template.JS(data),
	})
}
