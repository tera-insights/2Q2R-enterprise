// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http"

	bintemplate "github.com/arschles/go-bindata-html-template"
	"github.com/gorilla/mux"
	"github.com/tstranex/u2f"
)

// RegisterHandler is the handler for all registration requests.
type RegisterHandler struct {
	s *Server
	q Queue
}

// RegisterSetupHandler sets up the registration of a new two-factor device.
// GET /v1/register/request/:userID
func (rh *RegisterHandler) RegisterSetupHandler(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	serverID, _ := getAuthDataFromHeaders(r)
	serverInfo := AppServerInfo{}
	err := rh.s.DB.Model(AppServerInfo{}).Where(AppServerInfo{ServerID: serverID}).
		First(&serverInfo).Error
	if err != nil {
		handleError(w, err)
		return
	}
	challenge, err := u2f.NewChallenge(serverInfo.AppID, []string{serverInfo.AppID})
	if err != nil {
		handleError(w, err)
		return
	}
	rr := RegistrationRequest{
		RequestID: randString(32),
		Challenge: challenge,
		AppID:     serverInfo.AppID,
		UserID:    userID,
	}
	rh.s.cache.SetRegistrationRequest(rr.RequestID, rr)
	var url string
	if rh.s.c.HTTPS {
		url = "https://"
	} else {
		url = "http://"
	}
	writeJSON(w, http.StatusOK, RegistrationSetupReply{
		rr.RequestID,
		url + rh.s.c.BaseURL + rh.s.c.Port + "/register/" + rr.RequestID,
	})
}

// RegisterIFrameHandler returns the iFrame that is used to perform registration.
// GET /register/:id
func (rh *RegisterHandler) RegisterIFrameHandler(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	t, err := bintemplate.New("register", Asset).Parse("server/assets/all.html")
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
		Challenge: cachedRequest.Challenge.Challenge,
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
// Steps:
// 1. Parse request
// 2. Assert that we have a pending registration request for the challenge
// 3. Verify the signature in the request
// 4. Record the valid public key in the database
// POST /v1/register
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

	successData := req.Data.(successfulRegistrationData)

	// Decode the client data
	clientData := u2f.ClientData{}
	decoded, err := base64.StdEncoding.DecodeString(successData.ClientData)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Message: "Could not decode client data",
		})
		return
	}
	reader := bytes.NewReader(decoded)
	decoder = json.NewDecoder(reader)
	err = decoder.Decode(&clientData)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Message: "Could not decode client data",
		})
		return
	}

	// Assert that the challenge exists
	requestID, found := rh.s.cache.challengeToRequestID.Get(clientData.Challenge)
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
		return
	}

	// Verify signature
	resp := u2f.RegisterResponse{
		RegistrationData: successData.RegistrationData,
		ClientData:       successData.ClientData,
	}
	reg, err := u2f.Register(resp, *rr.Challenge, nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Message: "Could not verify signature when registering",
		})
		return
	}

	// Record valid public key in database
	err = rh.s.DB.Model(&Key{}).Create(Key{
		KeyID:     randString(32),
		UserID:    rr.UserID,
		AppID:     rr.AppID,
		PublicKey: reg.PubKey,
		Counter:   0,
	}).Error

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Message: "Could not save key to database",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Wait allows the requester to check the result of the registration. It blocks
// until the registration is complete.
// GET /v1/register/{requestID}/wait
func (rh RegisterHandler) Wait(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	c := rh.q.Listen(requestID)
	w.WriteHeader(<-c)
}
