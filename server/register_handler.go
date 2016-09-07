// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
)

// RegisterHandler is the handler for all registration requests.
type RegisterHandler struct {
	s *Server
}

// RegisterSetupHandler sets up the registration of a new two-factor device.
// POST /v1/auth/request
func (rh *RegisterHandler) RegisterSetupHandler(w http.ResponseWriter, r *http.Request) {
	req := RegistrationSetupRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		handleError(w, err)
		return
	}
	var challenge = make([]byte, rh.s.c.ChallengeLength)
	rand.Read(challenge)
	rr := RegistrationRequest{
		randString(32),
		challenge,
	}
	rh.s.cache.SetRegistrationRequest(rr.requestID, rr)
	server := AppServerInfo{}
	rh.s.DB.Model(AppServerInfo{}).Find(&server,
		AppServerInfo{ServerID: req.AuthenticationData.ServerID})
	writeJSON(w, http.StatusOK, RegistrationSetupReply{
		rr.requestID,
		server.BaseURL + "/register/" + rr.requestID,
	})
}
