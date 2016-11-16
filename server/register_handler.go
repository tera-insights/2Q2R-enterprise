// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net"
	"net/http"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/tstranex/u2f"
)

type registerHandler struct {
	s *Server
	q queue
}

type registerData struct {
	RequestID   string   `json:"id"`
	KeyTypes    []string `json:"keyTypes"`
	Challenge   string   `json:"challenge"` // base-64 URL-encoded
	UserID      string   `json:"userID"`
	AppID       string   `json:"appId"`
	BaseURL     string   `json:"baseUrl"`
	InfoURL     string   `json:"infoUrl"`
	RegisterURL string   `json:"registerUrl"`
	WaitURL     string   `json:"waitUrl"`
	AppURL      string   `json:"appUrl"`
}

// RegisterSetupHandler sets up the registration of a new two-factor device.
// GET /v1/register/request/:userID
func (rh *registerHandler) RegisterSetupHandler(w http.ResponseWriter, r *http.Request) {
	serverID, _, err := getAuthDataFromHeaders(r)
	optionalInternalPanic(err, "Could not decode authentication headers")

	userID := mux.Vars(r)["userID"]
	server := AppServerInfo{}
	err = rh.s.DB.First(&server, AppServerInfo{ID: serverID}).Error
	optionalBadRequestPanic(err, "Could not find app server")

	challenge, err := u2f.NewChallenge(rh.s.Config.getBaseURLWithProtocol(),
		[]string{rh.s.Config.getBaseURLWithProtocol()})
	optionalInternalPanic(err, "Could not generate challenge")

	requestID, err := RandString(32)
	optionalInternalPanic(err, "Could not generate request ID")

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	rh.s.cache.SetRegistrationRequest(requestID, registrationRequest{
		RequestID:  requestID,
		Challenge:  challenge,
		AppID:      server.AppID,
		UserID:     userID,
		OriginalIP: host,
	})
	writeJSON(w, http.StatusOK, registrationSetupReply{
		requestID,
		rh.s.Config.getBaseURLWithProtocol() + "/v1/register/" + requestID,
	})
}

// RegisterIFrameHandler returns the iFrame that is used to perform registration.
// GET /v1/register/:id
func (rh *registerHandler) RegisterIFrameHandler(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	templateBox, err := rice.FindBox("assets")
	optionalInternalPanic(err, "Failed to load assets")

	templateString, err := templateBox.String("all.html")
	optionalInternalPanic(err, "Failed to load template")

	t, err := template.New("register").Parse(templateString)
	optionalInternalPanic(err, "Failed to generate registration iFrame")

	cachedRequest, err := rh.s.cache.GetRegistrationRequest(requestID)
	optionalBadRequestPanic(err, "Failed to get registration request")

	var appInfo AppInfo
	query := AppInfo{ID: cachedRequest.AppID}
	err = rh.s.DB.Model(AppInfo{}).Find(&appInfo, query).Error
	optionalInternalPanic(err, "Failed to find app information")

	base := rh.s.Config.getBaseURLWithProtocol()
	data, err := json.Marshal(registerData{
		RequestID:   requestID,
		KeyTypes:    []string{"2q2r", "u2f"},
		Challenge:   EncodeBase64(cachedRequest.Challenge.Challenge),
		UserID:      cachedRequest.UserID,
		AppID:       cachedRequest.AppID,
		BaseURL:     base,
		AppURL:      base,
		InfoURL:     base + "/v1/info/" + cachedRequest.AppID,
		RegisterURL: base + "/v1/register",
		WaitURL:     base + "/v1/register/" + requestID + "/wait",
	})
	optionalInternalPanic(err, "Failed to generate template")

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
func (rh *registerHandler) Register(w http.ResponseWriter, r *http.Request) {
	req := registerRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	// Assert that the registration presented to us was successful
	if !req.Successful {
		failedData := req.Data.(failedRegistrationData)
		panic(bubbledError{
			StatusCode: failedData.ErrorCode,
			Message:    failedData.ErrorMessage,
		})
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
	optionalBadRequestPanic(err, "Could not decode client data")

	clientData := u2f.ClientData{}
	reader := bytes.NewReader(decoded)
	decoder = json.NewDecoder(reader)
	err = decoder.Decode(&clientData)
	optionalBadRequestPanic(err, "Could not decode client data")

	// Assert that the challenge exists
	requestID, found := rh.s.cache.challengeToRequestID.Get(clientData.Challenge)
	panicIfFalse(found, http.StatusForbidden, "Challenge does not exist")

	// Get challenge data
	rr, err := rh.s.cache.GetRegistrationRequest(requestID.(string))
	optionalInternalPanic(err, "Failed to look up data for valid challenge")

	// Verify signature
	resp := u2f.RegisterResponse{
		RegistrationData: successData.RegistrationData,
		ClientData:       successData.ClientData,
	}
	reg, err := u2f.Register(resp, *rr.Challenge, &u2f.Config{
		SkipAttestationVerify: true,
	})
	optionalBadRequestPanic(err, "Could not verify signature")

	// Record valid public key in database
	marshalledRegistration, err := reg.MarshalBinary()

	tx := rh.s.DB.Begin()

	// Save key
	err = rh.s.DB.Model(&Key{}).Create(&Key{
		ID:     EncodeBase64(reg.KeyHandle),
		Type:   successData.Type,
		Name:   successData.DeviceName,
		UserID: rr.UserID,
		AppID:  rr.AppID,
		MarshalledRegistration: marshalledRegistration,
		Counter:                0,
	}).Error
	if err != nil {
		tx.Rollback()
		optionalInternalPanic(err, "Could not save key to database")
	}

	// Notify request listeners
	err = rh.q.MarkCompleted(requestID.(string))
	if err != nil {
		tx.Rollback()
		optionalInternalPanic(err, "Could not notify request listeners")
	}

	tx.Commit()

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	rh.s.disperser.addEvent(registration, time.Now(), rr.AppID,
		"success", rr.UserID, rr.OriginalIP, host)
	writeJSON(w, http.StatusOK, registerResponse{
		Successful: true,
		Message:    "OK",
	})
}

// Wait allows the requester to check the result of the registration. It blocks
// until the registration is complete.
// GET /v1/register/{requestID}/wait
func (rh registerHandler) Wait(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	c, err := rh.q.Listen(requestID, false)
	if err != nil { // Should never happen
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(<-c)
	}

}
