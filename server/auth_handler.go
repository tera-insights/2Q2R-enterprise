// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	rice "github.com/GeertJohan/go.rice"
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
// GET /v1/auth/request/{userID}
func (ah *AuthHandler) AuthRequestSetupHandler(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	key := Key{}
	err := ah.s.DB.Model(&key).First(&key, &Key{
		UserID: userID,
	}).Error
	optionalInternalPanic(err, "Failed to load key")

	challenge, err := u2f.NewChallenge(ah.s.c.getBaseURLWithProtocol(),
		[]string{ah.s.c.getBaseURLWithProtocol()})
	optionalInternalPanic(err, "Failed to generate challenge")

	cachedRequest := AuthenticationRequest{
		RequestID: randString(32),
		Challenge: challenge,
		AppID:     key.AppID,
		UserID:    userID,
	}
	ah.s.cache.SetAuthenticationRequest(cachedRequest.RequestID, cachedRequest)
	writeJSON(w, http.StatusOK, AuthenticationSetupReply{
		cachedRequest.RequestID,
		ah.s.c.getBaseURLWithProtocol() + "/auth/" + cachedRequest.RequestID,
	})
}

// AuthIFrameHandler returns the iFrame that is used to perform authentication.
// GET /auth/:id
func (ah *AuthHandler) AuthIFrameHandler(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	templateBox, err := rice.FindBox("assets")
	optionalInternalPanic(err, "Failed to load assets")

	templateString, err := templateBox.String("all.html")
	optionalInternalPanic(err, "Failed to load template")

	t, err := template.New("auth").Parse(templateString)
	optionalInternalPanic(err, "Failed to generate authentication iFrame")

	cachedRequest, err := ah.s.cache.GetAuthenticationRequest(requestID)
	optionalPanic(err, http.StatusBadRequest, "Failed to load cached request")

	query := Key{AppID: cachedRequest.AppID, UserID: cachedRequest.UserID}
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
	base := ah.s.c.getBaseURLWithProtocol()
	data, err := json.Marshal(authenticateData{
		RequestID:    requestID,
		Counter:      1,
		Keys:         keys,
		Challenge:    encodeBase64(cachedRequest.Challenge.Challenge),
		UserID:       cachedRequest.UserID,
		AppID:        cachedRequest.AppID,
		BaseURL:      base,
		AppURL:       base,
		AuthURL:      base + "/v1/auth/",
		InfoURL:      base + "/v1/info/" + cachedRequest.AppID,
		WaitURL:      base + "/v1/auth/" + cachedRequest.RequestID + "/wait",
		ChallengeURL: base + "/v1/auth/" + cachedRequest.RequestID + "/challenge",
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
func (ah *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	req := authenticateRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode JSON body")

	// Assert that the authentication presented to us was successful
	if !req.Successful {
		failedData := req.Data.(failedAuthenticationData)
		panic(bubbledError{
			StatusCode: failedData.ErrorStatus,
			Message:    failedData.ErrorMessage,
		})
	}

	mappedValues := req.Data.(map[string]interface{})
	var successData successfulAuthenticationData

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
	ar, err := ah.s.cache.GetAuthenticationRequest(requestID.(string))
	optionalInternalPanic(err, "Failed to look up data for valid challenge")

	storedKey := Key{}
	err = ah.s.DB.Model(&Key{}).Where(&Key{
		UserID: ar.UserID,
		KeyID:  ar.KeyID,
	}).First(&storedKey).Error
	optionalInternalPanic(err, "Failed to look up stored key")

	var reg u2f.Registration
	err = reg.UnmarshalBinary(storedKey.MarshalledRegistration)
	optionalInternalPanic(err, "Failed to unmarshal stored registration data")

	resp := u2f.SignResponse{
		KeyHandle:     ar.KeyID,
		SignatureData: successData.SignatureData,
		ClientData:    successData.ClientData,
	}
	newCounter, err := reg.Authenticate(resp, *ar.Challenge, storedKey.Counter)
	optionalPanic(err, http.StatusBadRequest, "Authentication failed")

	// Store updated counter in the database.
	err = ah.s.DB.Model(&Key{}).Where(&Key{
		UserID: ar.UserID,
		KeyID:  ar.KeyID,
	}).Update("counter", newCounter).Error
	optionalPanic(err, http.StatusInternalServerError, "Failed to update counter")

	ah.q.MarkCompleted(requestID.(string))
}

// Wait allows the requester to check the result of the authentication. It
// blocks until the authentication is complete.
// GET /v1/auth/{requestID}/wait
func (ah *AuthHandler) Wait(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	c := ah.q.Listen(requestID)
	w.WriteHeader(<-c)
}

// SetKey sets the key for a given authentication request.
// POST /v1/auth/{requestID}/challenge
func (ah *AuthHandler) SetKey(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	req := setKeyRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalPanic(err, http.StatusBadRequest, "Could not decode request body")

	err = ah.s.cache.SetKeyForAuthenticationRequest(requestID, req.KeyID)
	optionalInternalPanic(err, "Failed to set key for authentication request")

	ar, err := ah.s.cache.GetAuthenticationRequest(requestID)
	optionalInternalPanic(err, "Failed to get authentication request")

	storedKey := Key{}
	err = ah.s.DB.Model(&Key{}).Where(&Key{KeyID: req.KeyID}).First(&storedKey).Error
	optionalBadRequestPanic(err, "Failed to get stored key")

	writeJSON(w, http.StatusOK, setKeyReply{
		KeyID:     req.KeyID,
		Challenge: encodeBase64(ar.Challenge.Challenge),
		Counter:   storedKey.Counter,
		AppID:     storedKey.AppID,
	})
}
