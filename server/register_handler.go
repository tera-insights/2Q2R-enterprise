// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/rand"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
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

// RegisterIFrameHandler returns the iFrame that is used to perform registration.
// GET /register/:id
func (rh *RegisterHandler) RegisterIFrameHandler(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	t, err := template.ParseFiles("../assets/all.html")
	if err != nil {
		handleError(w, err)
		return
	}
	data, err := json.Marshal(registerData{
		RequestID: requestID,
		KeyTypes:  []string{"foo", "bar"},
	})
	if err != nil {
		handleError(w, err)
		return
	}
	t.Execute(w, templateData{
		Name: "Registration",
		ID:   "register",
		Data: template.JS(data),
	})
}
