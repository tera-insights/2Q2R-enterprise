// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/tera-insights/2Q2R-enterprise/security"
	"github.com/tera-insights/2Q2R-enterprise/util"

	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/tstranex/u2f"
)

// authenticationData represents auth data.
type authenticationData struct {
	Counter  int
	ServerID string
}

type authHandler struct {
	s *Server

	requests             *cache.Cache // Request ID to *authReq
	challengeToRequestID *cache.Cache
	expiration           time.Duration

	// Migrated from queue
	rcTimeout  time.Duration
	rcInterval time.Duration
	lTimeout   time.Duration
	lInterval  time.Duration

	stateLock *sync.RWMutex
}

type keyDataToEmbed struct {
	KeyID string `json:"keyID"`
	Type  string `json:"type"`
	Name  string `json:"name"`
}

type authenticateData struct {
	RequestID    string           `json:"id"`
	Counter      int              `json:"counter"`
	Keys         []keyDataToEmbed `json:"keys"`
	Challenge    string           `json:"challenge"` // base-64 URL-encoded
	UserID       string           `json:"userID"`
	AppID        string           `json:"appId"`
	BaseURL      string           `json:"baseUrl"`
	AuthURL      string           `json:"authUrl"`
	InfoURL      string           `json:"infoUrl"`
	WaitURL      string           `json:"waitUrl"`
	ChallengeURL string           `json:"challengeUrl"`
	AppURL       string           `json:"appUrl"`
}

func newAuthHandler(s *Server) *authHandler {
	rcet := s.Config.RecentlyCompletedExpirationTime
	ct := s.Config.CleanTime

	return &authHandler{
		s,
		cache.New(rcet, ct),
		cache.New(rcet, ct),
		s.Config.ExpirationTime,
		rcet,
		ct,
		s.Config.ListenerExpirationTime,
		ct,
		&sync.RWMutex{},
	}
}

type authReq struct {
	RequestID     string
	Challenge     *u2f.Challenge
	KeyHandle     string
	AppID         string
	UserID        string
	OriginalIP    string
	Nonce         string
	Closed        chan struct{}
	Status        int
	NumListeners  int32 // Used atomically
	SettingResult int32 // Used atomically
}

// GetRequest returns the request for a particular request ID.
func (ah *authHandler) GetRequest(id string) (*authReq, error) {
	if val, found := ah.requests.Get(id); found {
		ptr := val.(*authReq)
		return ptr, nil
	}
	return nil, errors.Errorf("Could not find auth request with id %s", id)
}

// Listen returns a chan that emits an HTTP status code corresponding to the
// authentication request. It also returns a pointer to the request so that, if
// appropriate, handlers can attach cookies.
func (ah *authHandler) Listen(id string) (chan int, *authReq, error) {
	ar, err := ah.GetRequest(id)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not listen to unknown request")
	}

	if atomic.AddInt32(&ar.NumListeners, 1) != 1 {
		return nil, nil, errors.New("Someone is already listening")
	}

	c := make(chan int)
	go func() {
		<-ar.Closed
		c <- ar.Status
	}()

	go func() {
		time.Sleep(ah.lTimeout)
		// If no one set result
		if atomic.CompareAndSwapInt32(&ar.SettingResult, 0, 1) {
			ar.Status = http.StatusRequestTimeout
			ah.requests.Set(id, ar, ah.rcTimeout)
			ah.s.disperser.addEvent(authentication, time.Now(),
				ar.AppID, "timeout", ar.UserID, ar.OriginalIP, "")
			close(ar.Closed)
		}
	}()

	return c, ar, nil
}

// AuthRequestSetupHandler sets up a two-factor authentication request.
// GET /v1/auth/request/{userID}/{nonce}
func (ah *authHandler) Setup(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	appID, err := ah.s.kc.GetAppID(userID)
	util.OptionalInternalPanic(err, "Failed to load key")

	challenge, err := u2f.NewChallenge(ah.s.Config.getBaseURLWithProtocol(),
		[]string{ah.s.Config.getBaseURLWithProtocol()})
	util.OptionalInternalPanic(err, "Failed to generate challenge")

	requestID, err := util.RandString(32)
	util.OptionalInternalPanic(err, "Failed to generate request ID")

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ar := authReq{
		RequestID:  requestID,
		Challenge:  challenge,
		AppID:      appID,
		UserID:     userID,
		OriginalIP: host,
		Nonce:      mux.Vars(r)["nonce"],
		Closed:     make(chan struct{}),
	}
	ah.requests.Set(requestID, &ar, ah.expiration)
	s := util.EncodeBase64(ar.Challenge.Challenge)
	ah.challengeToRequestID.Set(s, requestID, ah.expiration)

	writeJSON(w, http.StatusOK, authenticationSetupReply{
		requestID,
		ah.s.Config.getBaseURLWithProtocol() + "/v1/auth/iframe",
	})
}

// IFrame returns the iFrame that is used to perform authentication.
// POST /v1/auth/iframe
func (ah *authHandler) IFrame(w http.ResponseWriter, r *http.Request) {
	var req requestIDWrapper
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	util.OptionalBadRequestPanic(err, "Could not decode request body")

	templateBox, err := rice.FindBox("assets")
	util.OptionalInternalPanic(err, "Failed to load assets")

	templateString, err := templateBox.String("all.html")
	util.OptionalInternalPanic(err, "Failed to load template")

	t, err := template.New("auth").Parse(templateString)
	util.OptionalInternalPanic(err, "Failed to generate authentication iFrame")

	cached, err := ah.GetRequest(req.RequestID)
	util.OptionalPanic(err, http.StatusBadRequest, "Failed to load cached request")

	query := security.Key{AppID: cached.AppID, UserID: cached.UserID}
	rows, err := ah.s.DB.Model(&security.Key{}).Where(query).Select([]string{
		"key_id", "type", "name",
	}).Rows()
	util.OptionalInternalPanic(err, "Could not load keys")

	defer rows.Close()

	var keys []keyDataToEmbed
	for rows.Next() {
		var keyID string
		var keyType string
		var name string
		err := rows.Scan(&keyID, &keyType, &name)
		util.OptionalInternalPanic(err, "Internal server error")
		keys = append(keys, keyDataToEmbed{
			KeyID: keyID,
			Type:  keyType,
			Name:  name,
		})
	}
	base := ah.s.Config.getBaseURLWithProtocol()
	data, err := json.Marshal(authenticateData{
		RequestID:    req.RequestID,
		Counter:      1,
		Keys:         keys,
		Challenge:    util.EncodeBase64(cached.Challenge.Challenge),
		UserID:       cached.UserID,
		AppID:        cached.AppID,
		BaseURL:      base,
		AppURL:       base,
		AuthURL:      base + "/v1/auth/",
		InfoURL:      base + "/v1/info/" + cached.AppID,
		WaitURL:      base + "/v1/auth/wait",
		ChallengeURL: base + "/v1/auth/challenge",
	})
	util.OptionalInternalPanic(err, "Failed to render template")

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
	util.OptionalBadRequestPanic(err, "Could not decode JSON body")

	// Assert that the authentication presented to us was successful
	if !req.Successful {
		failedData := req.Data.(failedauthenticationData)
		panic(util.BubbledError{
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

	decoded, err := util.DecodeBase64(successData.ClientData)
	util.OptionalBadRequestPanic(err, "Could not decode client data")

	clientData := u2f.ClientData{}
	reader := bytes.NewReader(decoded)
	decoder = json.NewDecoder(reader)
	err = decoder.Decode(&clientData)
	util.OptionalBadRequestPanic(err, "Could not decode client data")

	requestID, found := ah.challengeToRequestID.Get(clientData.Challenge)
	if !found {
		panic(util.BubbledError{
			StatusCode: http.StatusForbidden,
			Message:    "Challenge does not exist",
		})
	}

	// Get authentication request
	ar, err := ah.GetRequest(requestID.(string))
	util.OptionalInternalPanic(err, "Failed to look up data for valid challenge")

	storedKey, err := ah.s.kc.Get2FAKey(ar.KeyHandle)
	util.OptionalInternalPanic(err, "Failed to look up stored key")

	var reg u2f.Registration
	err = reg.UnmarshalBinary(storedKey.MarshalledRegistration)
	util.OptionalInternalPanic(err, "Failed to unmarshal stored registration data")

	resp := u2f.SignResponse{
		KeyHandle:     ar.KeyHandle,
		SignatureData: successData.SignatureData,
		ClientData:    successData.ClientData,
	}
	newCounter, err := reg.Authenticate(resp, *ar.Challenge, storedKey.Counter)
	util.OptionalPanic(err, http.StatusBadRequest, "Authentication failed")

	tx := ah.s.DB.Begin()

	// Store updated counter in the database.
	err = tx.Model(&security.Key{}).Where(&security.Key{
		UserID: ar.UserID,
		ID:     ar.KeyHandle,
	}).Update("counter", newCounter).Error
	if err != nil {
		tx.Rollback()
		util.OptionalInternalPanic(err, "Failed to update counter")
	}

	// Notify request listeners
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			ah.requests.Delete(requestID.(string))
		}
	}()

	util.PanicIfFalse(atomic.CompareAndSwapInt32(&ar.SettingResult, 0, 1),
		http.StatusConflict, "Request already timed out")

	ar.Status = http.StatusOK
	ah.requests.Set(requestID.(string), ar, ah.rcTimeout)
	close(ar.Closed)

	err = tx.Commit().Error
	util.OptionalInternalPanic(err, "Could not commit transaction to database")

	// If the authentication was for an admin, then look up the admin's app ID
	// and log the event under that key
	var appID string
	if ar.AppID == "1" {
		var a Admin
		err = ah.s.DB.First(&a, Admin{
			ID: ar.UserID,
		}).Error
		util.OptionalBadRequestPanic(err, "Could not find admin")

		appID = a.AdminFor
	} else {
		appID = ar.AppID
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ah.s.disperser.addEvent(authentication, time.Now(), appID, "success",
		ar.UserID, ar.OriginalIP, host)
	writeJSON(w, http.StatusOK, "Authentication successful")
}

// Wait allows the requester to check the result of the authentication. It
// blocks until the authentication is complete.
// POST /v1/auth/wait
func (ah *authHandler) Wait(w http.ResponseWriter, r *http.Request) {
	var req requestIDWrapper
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	util.OptionalBadRequestPanic(err, "Could not decode request body")

	c, ar, err := ah.Listen(req.RequestID)
	util.OptionalBadRequestPanic(err, "Could not listen for unknown request")

	status := <-c
	if status == http.StatusOK && ar.AppID == "1" {
		var a Admin
		err = ah.s.DB.First(&a, Admin{
			ID: ar.UserID,
		}).Error
		util.OptionalBadRequestPanic(err, "Could not find admin with id "+ar.UserID)

		encoded, err := ah.s.sc.Encode("admin-session", map[string]interface{}{
			"set":   time.Now(),
			"app":   a.AdminFor,
			"admin": a.ID,
		})

		util.OptionalInternalPanic(err, "Could not set session cookie")

		http.SetCookie(w, &http.Cookie{
			Name:  "admin-session",
			Value: encoded,
			Path:  "/",
		})
	}
	writeJSON(w, status, ar.Nonce)
}

// SetKey sets the key for a given authentication request.
// POST /v1/auth/challenge
func (ah *authHandler) SetKey(w http.ResponseWriter, r *http.Request) {
	req := setKeyRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	util.OptionalPanic(err, http.StatusBadRequest, "Could not decode request body")

	ar, err := ah.GetRequest(req.RequestID)
	util.OptionalBadRequestPanic(err, "Could not find auth "+
		"request with id "+req.RequestID)

	ar.KeyHandle = req.KeyHandle
	ah.requests.Set(req.RequestID, ar, ah.expiration)

	stored, err := ah.s.kc.Get2FAKey(req.KeyHandle)
	util.OptionalBadRequestPanic(err, "Failed to get stored key")

	writeJSON(w, http.StatusOK, setKeyReply{
		KeyHandle: req.KeyHandle,
		Challenge: util.EncodeBase64(ar.Challenge.Challenge),
		Counter:   stored.Counter,
		AppID:     stored.AppID,
	})
}
