// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net"
	"net/http"
	"sync"

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

	// Migrated from authenticator
	authReqs             *cache.Cache
	challengeToRequestID *cache.Cache
	expiration           time.Duration

	// Migrated from queue
	recent     *cache.Cache // Maps request ID to status code
	rcTimeout  time.Duration
	rcInterval time.Duration
	listeners  *cache.Cache
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
		cache.New(rcet, ct),
		rcet,
		ct,
		cache.New(s.Config.ListenerExpirationTime, ct),
		s.Config.ListenerExpirationTime,
		ct,
		&sync.RWMutex{},
	}
}

type authReq struct {
	RequestID  string
	Challenge  *u2f.Challenge
	KeyHandle  string
	AppID      string
	UserID     string
	OriginalIP string
}

// GetRequest returns the request for a particular request ID.
func (ah *authHandler) GetRequest(id string) (*authReq, error) {
	if val, found := ah.authReqs.Get(id); found {
		ar := val.(authReq)
		ptr := &ar
		return ptr, nil
	}
	return nil, errors.Errorf("Could not find auth request with id %s", id)
}

// Listen returns a chan that emits an HTTP status code corresponding to the
// authentication request. It also returns a pointer to the request so that, if
// appropriate, handlers can attach cookies.
func (ah *authHandler) Listen(id string) (c chan int, ar *authReq, err error) {
	ar, err = ah.GetRequest(id)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not listen to unknown request")
	}

	c = make(chan int, 1)

	withLocking(ah.stateLock, func() {
		status, found := ah.recent.Get(id)
		if found {
			c <- status.(int)
		} else {
			if _, found := ah.listeners.Get(id); found {
				c = nil
				ar = nil
				err = errors.New("Someone is already listening")
			} else {
				ah.listeners.Set(id, []chan int{c}, ah.lTimeout)
			}
		}
	})

	if err != nil {
		go func() {
			time.Sleep(ah.lTimeout)
			withLocking(ah.stateLock, func() {
				if _, found := ah.recent.Get(id); !found {
					ar, _ := ah.GetRequest(id)
					ah.recent.Set(id, http.StatusRequestTimeout, ah.rcTimeout)
					ah.listeners.Delete(id)
					ah.s.disperser.addEvent(authentication, time.Now(),
						ar.AppID, "timeout", ar.UserID, ar.OriginalIP, "")
				}
			})
		}()
	}

	return
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

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ar := authReq{
		RequestID:  requestID,
		Challenge:  challenge,
		AppID:      key.AppID,
		UserID:     userID,
		OriginalIP: host,
	}
	ah.authReqs.Set(requestID, ar, ah.expiration)
	s := EncodeBase64(ar.Challenge.Challenge)
	ah.challengeToRequestID.Set(s, requestID, ah.expiration)

	writeJSON(w, http.StatusOK, authenticationSetupReply{
		requestID,
		ah.s.Config.getBaseURLWithProtocol() + "/v1/auth/" + requestID,
	})
}

// AuthIFrameHandler returns the iFrame that is used to perform authentication.
// GET /v1/auth/:id
func (ah *authHandler) AuthIFrameHandler(w http.ResponseWriter,
	r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	templateBox, err := rice.FindBox("assets")
	optionalInternalPanic(err, "Failed to load assets")

	templateString, err := templateBox.String("all.html")
	optionalInternalPanic(err, "Failed to load template")

	t, err := template.New("auth").Parse(templateString)
	optionalInternalPanic(err, "Failed to generate authentication iFrame")

	cached, err := ah.GetRequest(requestID)
	optionalPanic(err, http.StatusBadRequest, "Failed to load cached request")

	query := Key{AppID: cached.AppID, UserID: cached.UserID}
	rows, err := ah.s.DB.Model(&Key{}).Where(query).Select([]string{
		"key_id", "type", "name",
	}).Rows()
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

	requestID, found := ah.challengeToRequestID.Get(clientData.Challenge)
	if !found {
		panic(bubbledError{
			StatusCode: http.StatusForbidden,
			Message:    "Challenge does not exist",
		})
	}

	// Get authentication request
	ar, err := ah.GetRequest(requestID.(string))
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
	withLocking(ah.stateLock, func() {
		defer func() {
			if r := recover(); r != nil {
				ah.recent.Delete(requestID.(string))
			}
		}()

		ah.recent.Set(requestID.(string), http.StatusOK, ah.rcTimeout)
		if cached, found := ah.listeners.Get(requestID.(string)); found {
			listeners := cached.([]chan int)
			for _, listener := range listeners {
				select {
				case listener <- http.StatusOK:
				default:
				}
			}
			ah.listeners.Delete(requestID.(string))
		}
	})

	if err != nil {
		tx.Rollback()
		optionalInternalPanic(err, "Could not notify request listeners")
	}

	err = tx.Commit().Error
	optionalInternalPanic(err, "Could not commit transaction to database")

	// If the authentication was for an admin, then look up the admin's app ID
	// and log the event under that key
	var appID string
	if ar.AppID == "1" {
		var a Admin
		err = ah.s.DB.First(&a, Admin{
			ID: ar.UserID,
		}).Error
		optionalBadRequestPanic(err, "Could not find admin")

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
// GET /v1/auth/{requestID}/wait
func (ah *authHandler) Wait(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["requestID"]
	c, ar, err := ah.Listen(id)
	optionalBadRequestPanic(err, "Could not listen for unknown request")

	status := <-c
	if status == http.StatusOK && ar.AppID == "1" {
		var a Admin
		err = ah.s.DB.First(&a, Admin{
			ID: ar.UserID,
		}).Error
		optionalBadRequestPanic(err, "Could not find admin with id "+ar.UserID)

		encoded, err := ah.s.sc.Encode("admin-session", map[string]interface{}{
			"set":   time.Now(),
			"app":   a.AdminFor,
			"admin": a.ID,
		})

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

	val, found := ah.authReqs.Get(requestID)
	panicIfFalse(found, http.StatusBadRequest, "Could not find auth request "+
		"with id "+requestID)

	ar := val.(authReq)
	ar.KeyHandle = req.KeyHandle
	ah.authReqs.Set(requestID, ar, ah.expiration)

	var stored Key
	err = ah.s.DB.Model(&Key{}).Where(&Key{
		ID: req.KeyHandle,
	}).First(&stored).Error
	optionalBadRequestPanic(err, "Failed to get stored key")

	writeJSON(w, http.StatusOK, setKeyReply{
		KeyHandle: req.KeyHandle,
		Challenge: EncodeBase64(ar.Challenge.Challenge),
		Counter:   stored.Counter,
		AppID:     stored.AppID,
	})
}
