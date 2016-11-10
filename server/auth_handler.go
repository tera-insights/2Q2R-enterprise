// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/tstranex/u2f"
)

// authenticationData represents auth data.
type authenticationData struct {
	Counter  int
	ServerID string
}

type authHandler struct {
	s  *Server
	sc *securecookie.SecureCookie
	a  *authenticator
}

// AuthRequestSetupHandler sets up a two-factor authentication request.
// GET /v1/auth/request/{userID}
func (ah *authHandler) AuthRequestSetupHandler(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	key := Key{}
	err := ah.s.DB.Model(&key).First(&key, &Key{
		UserID: userID,
	}).Error
	optionalInternalPanic(err, "Failed to load key")

	challenge, err := u2f.NewChallenge(ah.s.Config.getBaseURLWithProtocol(),
		[]string{ah.s.Config.getBaseURLWithProtocol()})
	optionalInternalPanic(err, "Failed to generate challenge")

	requestID, err := RandString(32)
	optionalInternalPanic(err, "Failed to generate request ID")

	req := authReq{
		RequestID: requestID,
		Challenge: challenge,
		AppID:     key.AppID,
		UserID:    userID,
	}
	ah.a.PutRequest(requestID, req)
	writeJSON(w, http.StatusOK, authenticationSetupReply{
		requestID,
		ah.s.Config.getBaseURLWithProtocol() + "/v1/auth/" + requestID,
	})
}

// AuthIFrameHandler returns the iFrame that is used to perform authentication.
// GET /v1/auth/:id
func (ah *authHandler) AuthIFrameHandler(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	templateBox, err := rice.FindBox("assets")
	optionalInternalPanic(err, "Failed to load assets")

	templateString, err := templateBox.String("all.html")
	optionalInternalPanic(err, "Failed to load template")

	t, err := template.New("auth").Parse(templateString)
	optionalInternalPanic(err, "Failed to generate authentication iFrame")

	cached, err := ah.a.GetRequest(requestID)
	optionalPanic(err, http.StatusBadRequest, "Failed to load cached request")

	query := Key{AppID: cached.AppID, UserID: cached.UserID}
	rows, err := ah.s.DB.Model(&Key{}).Where(query).Select([]string{"key_id", "type", "name"}).Rows()
	optionalInternalPanic(err, "Could not load keys")

	defer rows.Close()

	var keys []keyDataToEmbed
	for rows.Next() {
		var keyID string
		var keyType string
		var name string
		err := rows.Scan(&keyID, &keyType, &name)
		optionalInternalPanic(err, "Internal server error")
		keys = append(keys, keyDataToEmbed{
			KeyID: keyID,
			Type:  keyType,
			Name:  name,
		})
	}
	base := ah.s.Config.getBaseURLWithProtocol()
	data, err := json.Marshal(authenticateData{
		RequestID:    requestID,
		Counter:      1,
		Keys:         keys,
		Challenge:    EncodeBase64(cached.Challenge.Challenge),
		UserID:       cached.UserID,
		AppID:        cached.AppID,
		BaseURL:      base,
		AppURL:       base,
		AuthURL:      base + "/v1/auth/",
		InfoURL:      base + "/v1/info/" + cached.AppID,
		WaitURL:      base + "/v1/auth/" + cached.RequestID + "/wait",
		ChallengeURL: base + "/v1/auth/" + cached.RequestID + "/challenge",
	})
	optionalInternalPanic(err, "Failed to render template")

	t.Execute(w, templateData{
		Name: "Authentication",
		ID:   "auth",
		Data: template.JS(data),
	})
}

// Authenticate performs authentication for a U2F device.
// POST /v1/auth
func (ah *authHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	req := authenticateRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode JSON body")

	// Assert that the authentication presented to us was successful
	if !req.Successful {
		failedData := req.Data.(failedauthenticationData)
		panic(bubbledError{
			StatusCode: failedData.ErrorStatus,
			Message:    failedData.ErrorMessage,
		})
	}

	mappedValues := req.Data.(map[string]interface{})
	var successData successfulauthenticationData

	// There were problems with deserialization. This is gross. Will fix later.
	if value, ok := mappedValues["clientData"]; ok {
		successData.ClientData = value.(string)
	}
	if value, ok := mappedValues["signatureData"]; ok {
		successData.SignatureData = value.(string)
	}

	decoded, err := decodeBase64(successData.ClientData)
	optionalBadRequestPanic(err, "Could not decode client data")

	clientData := u2f.ClientData{}
	reader := bytes.NewReader(decoded)
	decoder = json.NewDecoder(reader)
	err = decoder.Decode(&clientData)
	optionalBadRequestPanic(err, "Could not decode client data")

	requestID, found := ah.s.cache.challengeToRequestID.Get(clientData.Challenge)
	if !found {
		panic(bubbledError{
			StatusCode: http.StatusForbidden,
			Message:    "Challenge does not exist",
		})
	}

	// Get authentication request
	ar, err := ah.a.GetRequest(requestID.(string))
	optionalInternalPanic(err, "Failed to look up data for valid challenge")

	storedKey := Key{}
	err = ah.s.DB.Model(&Key{}).Where(&Key{
		UserID: ar.UserID,
		ID:     ar.KeyHandle,
	}).First(&storedKey).Error
	optionalInternalPanic(err, "Failed to look up stored key")

	var reg u2f.Registration
	err = reg.UnmarshalBinary(storedKey.MarshalledRegistration)
	optionalInternalPanic(err, "Failed to unmarshal stored registration data")

	resp := u2f.SignResponse{
		KeyHandle:     ar.KeyHandle,
		SignatureData: successData.SignatureData,
		ClientData:    successData.ClientData,
	}
	newCounter, err := reg.Authenticate(resp, *ar.Challenge, storedKey.Counter)
	optionalPanic(err, http.StatusBadRequest, "Authentication failed")

	tx := ah.s.DB.Begin()

	// Store updated counter in the database.
	err = tx.Model(&Key{}).Where(&Key{
		UserID: ar.UserID,
		ID:     ar.KeyHandle,
	}).Update("counter", newCounter).Error
	if err != nil {
		tx.Rollback()
		optionalInternalPanic(err, "Failed to update counter")
	}

	// Notify request listeners
	err = ah.a.MarkCompleted(requestID.(string))
	if err != nil {
		tx.Rollback()
		optionalInternalPanic(err, "Could not notify request listeners")
	}

	err = tx.Commit().Error
	optionalInternalPanic(err, "Could not commit transaction to database")

	writeJSON(w, http.StatusOK, "Authentication successful")
}

// Wait allows the requester to check the result of the authentication. It
// blocks until the authentication is complete.
// GET /v1/auth/{requestID}/wait
func (ah *authHandler) Wait(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["requestID"]
	c, ar, err := ah.a.Listen(id)
	optionalBadRequestPanic(err, "Could not listen for unknown request")

	status := <-c
	if status == http.StatusOK && ar.AppID == "1" {
		encoded, err := ah.sc.Encode("admin-session", time.Now())
		optionalInternalPanic(err, "Could not set session cookie")

		http.SetCookie(w, &http.Cookie{
			Name:  "admin-session",
			Value: encoded,
			Path:  "/",
		})
	}
	w.WriteHeader(status)
}

// SetKey sets the key for a given authentication request.
// POST /v1/auth/{requestID}/challenge
func (ah *authHandler) SetKey(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	req := setKeyRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalPanic(err, http.StatusBadRequest, "Could not decode request body")

	err = ah.a.SetKey(requestID, req.KeyHandle)
	optionalInternalPanic(err, "Failed to set key for authentication request")

	ar, err := ah.a.GetRequest(requestID)
	optionalInternalPanic(err, "Failed to get authentication request")

	storedKey := Key{}
	err = ah.s.DB.Model(&Key{}).Where(&Key{ID: req.KeyHandle}).First(&storedKey).Error
	optionalBadRequestPanic(err, "Failed to get stored key")

	writeJSON(w, http.StatusOK, setKeyReply{
		KeyHandle: req.KeyHandle,
		Challenge: EncodeBase64(ar.Challenge.Challenge),
		Counter:   storedKey.Counter,
		AppID:     storedKey.AppID,
	})
}
