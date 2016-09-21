// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tstranex/u2f"
)

// AuthenticationData represents auth data.
type AuthenticationData struct {
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
		return
	}
	challenge, err := u2f.NewChallenge(req.AppID, []string{ah.s.c.getBaseURLWithProtocol()})
	if err != nil {
		handleError(w, err)
		return
	}
	cachedRequest := AuthenticationRequest{
		RequestID: randString(32),
		Challenge: challenge,
		AppID:     req.AppID,
		UserID:    req.UserID,
	}
	ah.s.cache.SetAuthenticationRequest(cachedRequest.RequestID, cachedRequest)
	server := AppServerInfo{}
	ah.s.DB.Model(AppServerInfo{}).Find(&server,
		AppServerInfo{ServerID: req.AuthenticationData.ServerID})
	writeJSON(w, http.StatusOK, AuthenticationSetupReply{
		cachedRequest.RequestID,
		server.BaseURL + "/auth/" + cachedRequest.RequestID,
	})
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
	cachedRequest, err := ah.s.cache.GetAuthenticationRequest(requestID)
	if err != nil {
		handleError(w, err)
		return
	}
	query := Key{AppID: cachedRequest.AppID, UserID: cachedRequest.UserID}
	var keys []string
	err = ah.s.DB.Model(Key{}).Where(query).Select("PublicKey").Where(keys).Error
	base := ah.s.c.getBaseURLWithProtocol()
	data, err := json.Marshal(authenticateData{
		RequestID:    requestID,
		Counter:      1,
		Keys:         keys,
		Challenge:    encodeBase64(cachedRequest.Challenge.Challenge),
		UserID:       cachedRequest.UserID,
		AppID:        cachedRequest.AppID,
		AuthURL:      base + "/v1/auth/",
		InfoURL:      base + "/v1/info/" + cachedRequest.AppID,
		WaitURL:      base + "/v1/auth/" + cachedRequest.RequestID + "/wait",
		ChallengeURL: base + "/v1/auth/" + cachedRequest.RequestID + "/challenge",
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
