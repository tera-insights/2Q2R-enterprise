// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tstranex/u2f"
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
	rr := RegistrationRequest{
		RequestID: randString(32),
		Challenge: []byte{0x01},
		AppID:     req.AppID,
		UserID:    req.UserID,
	}
	rh.s.cache.SetRegistrationRequest(rr.RequestID, rr)
	server := AppServerInfo{}
	rh.s.DB.Model(AppServerInfo{}).Find(&server,
		AppServerInfo{ServerID: req.AuthenticationData.ServerID})
	writeJSON(w, http.StatusOK, RegistrationSetupReply{
		rr.RequestID,
		server.BaseURL + "/register/" + rr.RequestID,
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
	cachedRequest, err := rh.s.cache.GetRegistrationRequest(requestID)
	if err != nil {
		handleError(w, err)
		return
	}
	var appInfo AppInfo
	query := AppInfo{AppID: cachedRequest.AppID}
	err = rh.s.DB.Model(AppInfo{}).Find(&appInfo, query).Error
	if err != nil {
		handleError(w, err)
		return
	}
	data, err := json.Marshal(registerData{
		RequestID: requestID,
		KeyTypes:  []string{"2q2r", "u2f"},
		Challenge: cachedRequest.Challenge,
		UserID:    cachedRequest.UserID,
		AppID:     cachedRequest.AppID,
		InfoURL:   rh.s.c.BaseURL + "/v1/info/" + cachedRequest.AppID,
		WaitURL:   rh.s.c.BaseURL + "/v1/register/" + requestID + "/wait",
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

// Register registers a new authentication method for a device.
func (rh *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	req := RegisterRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		handleError(w, err)
		return
	}

	// Assert that the registration presented to us was successful
	if !req.Successful {
		failedData := req.Data.(failedRegistrationData)
		writeJSON(w, failedData.ErrorCode, errorResponse{
			Message: string(failedData.ErrorMessage),
		})
		return
	}

	// Assert that the challenge exists
	successData := req.Data.(successfulRegistrationData)
	clientData := u2f.ClientData{}
	decoded, err := base64.StdEncoding.DecodeString(successData.ClientData)
	reader := bytes.NewReader(decoded)
	decoder = json.NewDecoder(reader)
	err = decoder.Decode(&clientData)
	c := u2f.Challenge{}
	requestID, found := rh.s.cache.challengeToRequestID.Get(string(c.Challenge))
	if !found {
		writeJSON(w, http.StatusForbidden, errorResponse{
			Message: "Challenge does not exist",
		})
		return
	}

	// Get challenge data
	rr, err := rh.s.cache.GetRegistrationRequest(requestID.(string))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Message: "Failed to look up data for valid challenge",
		})
	}

	// Verify app id

	// Verify signature
	var counter uint32

	var reg u2f.Registration
	var resp u2f.SignResponse
	newCounter, err := reg.Authenticate(resp, c, counter)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Message: "Could not verify signature",
		})
		return
	}
}
