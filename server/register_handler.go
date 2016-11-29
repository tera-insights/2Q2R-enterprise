// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"2q2r/util"
	"bytes"
	"crypto"
	"encoding/json"
	"html/template"
	"io"
	"net"
	"net/http"
	"time"

	"sync"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/tstranex/u2f"
)

func withLocking(l *sync.RWMutex, f func()) {
	defer l.Unlock()
	l.Lock()
	f()
}

type registerHandler struct {
	s *Server

	// Migrated from cache
	registrationReqs     *cache.Cache
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

type registrationReq struct {
	RequestID  string
	Challenge  *u2f.Challenge
	AppID      string
	UserID     string
	OriginalIP string
}

func newRegisterHandler(s *Server) *registerHandler {
	rcet := s.Config.RecentlyCompletedExpirationTime
	ct := s.Config.CleanTime

	return &registerHandler{
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

// GetRequest returns the request for a particular request ID.
func (rh *registerHandler) GetRequest(id string) (*registrationReq, error) {
	if val, found := rh.registrationReqs.Get(id); found {
		rr := val.(registrationReq)
		ptr := &rr
		return ptr, nil
	}

	// For long-term requests
	ltr := LongTermRequest{}
	h := crypto.SHA256.New()
	io.WriteString(h, id)

	// We transactionally find the long-term request and then delete it from
	// the DB.
	tx := rh.s.DB.Begin()
	query := LongTermRequest{ID: string(h.Sum(nil))}
	if err := tx.First(ltr, query).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Delete(LongTermRequest{}, query).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	base := rh.s.Config.getBaseURLWithProtocol()
	challenge, err := u2f.NewChallenge(base, []string{base})
	if err != nil {
		return nil, err
	}

	rh.registrationReqs.Set(id, registrationReq{
		RequestID: id,
		Challenge: challenge,
		AppID:     ltr.AppID,
	}, rh.expiration)
	s := util.EncodeBase64(challenge.Challenge)
	rh.challengeToRequestID.Set(s, id, rh.expiration)
	return nil, errors.Errorf("Could not find request with id %s", id)
}

// RegisterSetupHandler sets up the registration of a new two-factor device.
// GET /v1/register/request/:userID
func (rh *registerHandler) RegisterSetupHandler(w http.ResponseWriter, r *http.Request) {
	serverID, _, err := getAuthDataFromHeaders(r)
	util.OptionalInternalPanic(err, "Could not decode authentication headers")

	userID := mux.Vars(r)["userID"]
	server := AppServerInfo{}
	err = rh.s.DB.First(&server, AppServerInfo{ID: serverID}).Error
	util.OptionalBadRequestPanic(err, "Could not find app server")

	challenge, err := u2f.NewChallenge(rh.s.Config.getBaseURLWithProtocol(),
		[]string{rh.s.Config.getBaseURLWithProtocol()})
	util.OptionalInternalPanic(err, "Could not generate challenge")

	requestID, err := util.RandString(32)
	util.OptionalInternalPanic(err, "Could not generate request ID")

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	rr := registrationReq{
		RequestID:  requestID,
		Challenge:  challenge,
		AppID:      server.AppID,
		UserID:     userID,
		OriginalIP: host,
	}
	rh.registrationReqs.Set(requestID, rr, rh.expiration)
	s := util.EncodeBase64(rr.Challenge.Challenge)
	rh.challengeToRequestID.Set(s, requestID, rh.expiration)

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
	util.OptionalInternalPanic(err, "Failed to load assets")

	templateString, err := templateBox.String("all.html")
	util.OptionalInternalPanic(err, "Failed to load template")

	t, err := template.New("register").Parse(templateString)
	util.OptionalInternalPanic(err, "Failed to generate registration iFrame")

	cachedRequest, err := rh.GetRequest(requestID)
	util.OptionalBadRequestPanic(err, "Failed to get registration request")

	var appInfo AppInfo
	query := AppInfo{ID: cachedRequest.AppID}
	err = rh.s.DB.Model(AppInfo{}).Find(&appInfo, query).Error
	util.OptionalInternalPanic(err, "Failed to find app information")

	base := rh.s.Config.getBaseURLWithProtocol()
	data, err := json.Marshal(registerData{
		RequestID:   requestID,
		KeyTypes:    []string{"2q2r", "u2f"},
		Challenge:   util.EncodeBase64(cachedRequest.Challenge.Challenge),
		UserID:      cachedRequest.UserID,
		AppID:       cachedRequest.AppID,
		BaseURL:     base,
		AppURL:      base,
		InfoURL:     base + "/v1/info/" + cachedRequest.AppID,
		RegisterURL: base + "/v1/register",
		WaitURL:     base + "/v1/register/" + requestID + "/wait",
	})
	util.OptionalInternalPanic(err, "Failed to generate template")

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
	util.OptionalBadRequestPanic(err, "Could not decode request body")

	// Assert that the registration presented to us was successful
	if !req.Successful {
		failedData := req.Data.(failedRegistrationData)
		panic(util.BubbledError{
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
	decoded, err := util.DecodeBase64(successData.ClientData)
	util.OptionalBadRequestPanic(err, "Could not decode client data")

	clientData := u2f.ClientData{}
	reader := bytes.NewReader(decoded)
	decoder = json.NewDecoder(reader)
	err = decoder.Decode(&clientData)
	util.OptionalBadRequestPanic(err, "Could not decode client data")

	// Assert that the challenge exists
	val, found := rh.challengeToRequestID.Get(clientData.Challenge)
	util.PanicIfFalse(found, http.StatusForbidden, "Challenge does not exist")

	requestID, ok := val.(string)
	util.PanicIfFalse(ok, http.StatusInternalServerError, "Invalid cached data")

	// Get challenge data
	val, found = rh.registrationReqs.Get(requestID)
	util.PanicIfFalse(found, http.StatusInternalServerError, "Failed to look up "+
		"data for valid challenge")

	rr, ok := val.(registrationReq)
	util.PanicIfFalse(ok, http.StatusInternalServerError, "Invalid cached data")

	// Verify signature
	resp := u2f.RegisterResponse{
		RegistrationData: successData.RegistrationData,
		ClientData:       successData.ClientData,
	}
	reg, err := u2f.Register(resp, *rr.Challenge, &u2f.Config{
		SkipAttestationVerify: true,
	})
	util.OptionalBadRequestPanic(err, "Could not verify signature")

	// Record valid public key in database
	marshalledRegistration, err := reg.MarshalBinary()

	tx := rh.s.DB.Begin()

	// Save key
	err = rh.s.DB.Model(&Key{}).Create(&Key{
		ID:     util.EncodeBase64(reg.KeyHandle),
		Type:   successData.Type,
		Name:   successData.DeviceName,
		UserID: rr.UserID,
		AppID:  rr.AppID,
		MarshalledRegistration: marshalledRegistration,
		Counter:                0,
	}).Error
	if err != nil {
		tx.Rollback()
		util.OptionalInternalPanic(err, "Could not save key to database")
	}

	// Mark the request as completed
	withLocking(rh.stateLock, func() {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				rh.recent.Delete(requestID)
			}
		}()

		if _, found = rh.recent.Get(requestID); found {
			writeJSON(w, http.StatusUnauthorized, "Request timed out")
			tx.Rollback()
			return
		}
		rh.recent.Set(requestID, http.StatusOK, rh.rcTimeout)
	})

	// Tell all the listeners that we finished
	withLocking(rh.stateLock, func() {
		if cached, found := rh.listeners.Get(requestID); found {
			listeners := cached.([]chan int)
			for _, listener := range listeners {
				select {
				case listener <- http.StatusOK:
				default:
				}
			}
			rh.listeners.Delete(requestID)
		}
	})

	if err != nil {
		tx.Rollback()
		util.OptionalInternalPanic(err, "Could not notify request listeners")
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
func (rh *registerHandler) Wait(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["requestID"]
	c := make(chan int, 1)

	rh.stateLock.RLock()
	status, found := rh.recent.Get(id)
	rh.stateLock.RUnlock()

	if found {
		c <- status.(int)
	} else {
		withLocking(rh.stateLock, func() {
			if val, found := rh.listeners.Get(id); found {
				listeners := val.([]chan int)
				newListeners := append(listeners, c)
				rh.listeners.Set(id, newListeners, rh.lTimeout)
			} else {
				rh.listeners.Set(id, []chan int{c}, rh.lTimeout)
			}
		})

		go func() {
			time.Sleep(rh.lTimeout)
			withLocking(rh.stateLock, func() {
				// Only time the request out if it did not complete
				if _, found := rh.recent.Get(id); !found {
					rr, _ := rh.GetRequest(id)
					rh.recent.Set(id, http.StatusRequestTimeout,
						rh.rcTimeout)
					rh.listeners.Delete(id)
					rh.s.disperser.addEvent(registration, time.Now(),
						rr.AppID, "timeout", rr.UserID, rr.OriginalIP, "")
				}
			})
		}()
	}
	w.WriteHeader(<-c)
}
