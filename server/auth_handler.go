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
			RequestID: randString(32),
			Challenge: challenge,
			Counter:   req.AuthenticationData.Counter + 1,
			AppID:     req.AppID,
			UserID:    req.UserID,
		}
		ah.s.cache.SetAuthenticationRequest(r.RequestID, r)
		server := AppServerInfo{}
		ah.s.DB.Model(AppServerInfo{}).Find(&server,
			AppServerInfo{ServerID: req.AuthenticationData.ServerID})
		writeJSON(w, http.StatusOK, AuthenticationSetupReply{
			r.RequestID,
			server.BaseURL + "/auth/" + r.RequestID,
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
	cachedRequest, err := ah.s.cache.GetAuthenticationRequest(requestID)
	if err != nil {
		handleError(w, err)
		return
	}
	query := Key{AppID: cachedRequest.AppID, UserID: cachedRequest.UserID}
	var keys []string
	err = ah.s.DB.Model(Key{}).Where(query).Select("PublicKey").Where(keys).Error
	data, err := json.Marshal(authenticateData{
		RequestID:    requestID,
		Counter:      1,
		Keys:         keys,
		Challenge:    cachedRequest.Challenge,
		UserID:       cachedRequest.UserID,
		AppID:        cachedRequest.AppID,
		InfoURL:      ah.s.c.BaseURL + "/v1/info/" + cachedRequest.AppID,
		WaitURL:      ah.s.c.BaseURL + "/v1/auth/" + cachedRequest.RequestID + "/wait",
		ChallengeURL: ah.s.c.BaseURL + "/v1/auth/" + cachedRequest.RequestID + "/challenge",
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
