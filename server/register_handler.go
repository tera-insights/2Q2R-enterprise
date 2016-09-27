// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/GeertJohan/go.rice"
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
	challenge, err := u2f.NewChallenge(rh.s.c.getBaseURLWithProtocol(),
		[]string{rh.s.c.getBaseURLWithProtocol()})
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
	writeJSON(w, http.StatusOK, RegistrationSetupReply{
		rr.RequestID,
		rh.s.c.getBaseURLWithProtocol() + "/register/" + rr.RequestID,
	})
}

// RegisterIFrameHandler returns the iFrame that is used to perform registration.
// GET /register/:id
func (rh *RegisterHandler) RegisterIFrameHandler(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	templateBox, err := rice.FindBox("assets")
	if err != nil {
		handleError(w, err)
		return
	}
	templateString, err := templateBox.String("all.html")
	if err != nil {
		handleError(w, err)
		return
	}
	t, err := template.New("register").Parse(templateString)
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
	base := rh.s.c.getBaseURLWithProtocol()

	data, err := json.Marshal(registerData{
		RequestID:   requestID,
		KeyTypes:    []string{"2q2r", "u2f"},
		Challenge:   encodeBase64(cachedRequest.Challenge.Challenge),
		UserID:      cachedRequest.UserID,
		AppID:       cachedRequest.AppID,
		BaseURL:     base,
		AppURL:      base,
		InfoURL:     base + "/v1/info/" + cachedRequest.AppID,
		RegisterURL: base + "/v1/register",
		WaitURL:     base + "/v1/register/" + requestID + "/wait",
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

	mappedValues := req.Data.(map[string]interface{})
	var successData successfulRegistrationData

	// There were problems with deserialization. This is gross. Will fix later.
	if value, ok := mappedValues["clientData"]; ok {
		successData.ClientData = value.(string)
	}
	if value, ok := mappedValues["registrationData"]; ok {
		successData.RegistrationData = value.(string)
	}
	if value, ok := mappedValues["deviceName"]; ok {
		successData.DeviceName = value.(string)
	}
	if value, ok := mappedValues["type"]; ok {
		successData.Type = value.(string)
	}
	if value, ok := mappedValues["fcmToken"]; ok {
		successData.FCMToken = value.(string)
	}

	// Decode the client data
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
	reg, err := u2f.Register(resp, *rr.Challenge, &u2f.Config{
		SkipAttestationVerify: true,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Message: "Could not verify signature when registering",
		})
		return
	}

	// Record valid public key in database
	marshalledRegistration, err := reg.MarshalBinary()
	err = rh.s.DB.Model(&Key{}).Create(&Key{
		KeyID:  encodeBase64(reg.KeyHandle),
		Type:   successData.Type,
		Name:   successData.DeviceName,
		UserID: rr.UserID,
		AppID:  rr.AppID,
		MarshalledRegistration: marshalledRegistration,
		Counter:                0,
	}).Error

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Message: "Could not save key to database",
		})
		return
	}

	rh.q.MarkCompleted(requestID.(string))

	writeJSON(w, http.StatusOK, RegisterResponse{
		Successful: true,
		Message:    "OK",
	})
}

// Wait allows the requester to check the result of the registration. It blocks
// until the registration is complete.
// GET /v1/register/{requestID}/wait
func (rh RegisterHandler) Wait(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	c := rh.q.Listen(requestID)
	w.WriteHeader(<-c)
}
