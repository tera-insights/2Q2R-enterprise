// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
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
	q Queue
}

// AuthRequestSetupHandler sets up a two-factor authentication request.
// GET /v1/auth/request
func (ah *AuthHandler) AuthRequestSetupHandler(w http.ResponseWriter, r *http.Request) {
	req := AuthenticationSetupRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		handleError(w, err)
		return
	}
	challenge, err := u2f.NewChallenge(ah.s.c.getBaseURLWithProtocol(),
		[]string{ah.s.c.getBaseURLWithProtocol()})
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
		ah.s.c.getBaseURLWithProtocol() + "/auth/" + cachedRequest.RequestID,
	})
}

// AuthIFrameHandler returns the iFrame that is used to perform authentication.
// GET /auth/:id
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
		BaseURL:      base,
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

// POST /v1/auth
func (ah *AuthHandler) authenticate(w http.ResponseWriter, r *http.Request) {
	req := authenticateRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		handleError(w, err)
		return
	}

	// Assert that the authentication presented to us was successful
	if !req.Successful {
		failedData := req.Data.(failedAuthenticationData)
		writeJSON(w, failedData.ErrorStatus, errorResponse{
			Message: string(failedData.ErrorMessage),
		})
		return
	}

	mappedValues := req.Data.(map[string]interface{})
	var successData successfulAuthenticationData

	// There were problems with deserialization. This is gross. Will fix later.
	if value, ok := mappedValues["clientData"]; ok {
		successData.ClientData = value.(string)
	}
	if value, ok := mappedValues["registrationData"]; ok {
		successData.RegistrationData = value.(string)
	}

	decoded, err := decodeBase64(successData.ClientData)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Message: "Could not decode client data",
		})
		return
	}
	clientData := u2f.ClientData{}
	reader := bytes.NewReader(decoded)
	decoder = json.NewDecoder(reader)
	err = decoder.Decode(&clientData)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Message: "Could not decode client data",
		})
		return
	}

	requestID, found := ah.s.cache.challengeToRequestID.Get(clientData.Challenge)
	if !found {
		writeJSON(w, http.StatusForbidden, errorResponse{
			Message: "Challenge does not exist",
		})
		return
	}

	// Get authentication request
	ar, err := ah.s.cache.GetAuthenticationRequest(requestID.(string))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Message: "Failed to look up data for valid challenge",
		})
		return
	}

	resp := u2f.SignResponse{}

	var counter uint32
	newCounter, err := reg.Authenticate(resp, ar.Challenge, counter)
	if err != nil {
		// Authentication failed.
	}

	// Store updated counter in the database.

	ah.q.MarkCompleted(requestID.(string))
}
