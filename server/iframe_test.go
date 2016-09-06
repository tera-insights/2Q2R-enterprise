// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

type registerData struct {
	id        string
	keyTypes  []string
	challenge []byte
	userID    string
	appId     string
	infoUrl   string
	waitUrl   string
}

type authenticateData struct {
	id           string
	counter      int
	keys         []string
	challenge    []byte
	userID       string
	appId        string
	infoUrl      string
	waitUrl      string
	challengeUrl string
}

func TestRegisterIFrameGeneration(t *testing.T) {
	// Set up registration request
	registrationRequest := RegistrationSetupRequest{
		AppID:     goodAppID,
		Timestamp: time.Now(),
		UserID:    "bar",
	}
	res, _ := postJSON("/v1/register/request", registrationRequest)
	setupInfo := new(RegistrationSetupReply)
	unmarshalJSONBody(res, setupInfo)

	// Get registration iFrame
	res, _ = http.Get(ts.URL + "/register/" + registrationRequest.AppID)
	var bodyBytes = make([]byte, 0)
	n, _ := res.Body.Read(bodyBytes)
	iFrameBody := string(bodyBytes[:n])
	if strings.Index(iFrameBody, "var data = ") == -1 {
		t.Errorf("Could not find data inside iFrameBody")
	}

	// Get app info
	res, _ = http.Get(ts.URL + "/v1/info/" + registrationRequest.AppID)
	appInfo := new(AppIDInfoReply)
	unmarshalJSONBody(res, appInfo)

	// Assert that data embedded in the iFrame is what we expect
	gleanedData := registerData{}
	cachedRequest, _ := s.cache.GetRegistrationRequest(setupInfo.RequestID)
	correctData := registerData{
		id:        setupInfo.RequestID,
		keyTypes:  []string{"2q2r", "u2f"},
		challenge: cachedRequest.challenge,
		userID:    registrationRequest.UserID,
		appId:     registrationRequest.AppID,
		infoUrl:   appInfo.BaseURL + "/v1/info/" + registrationRequest.AppID,
		waitUrl:   appInfo.BaseURL + "/v1/register/" + setupInfo.RequestID + "/wait",
	}
	if !reflect.DeepEqual(gleanedData, correctData) {
		t.Errorf("Gleaned data was not expected")
	}
}

func TestAuthenticateIFrameGeneration(t *testing.T) {
	// Set up registration request
	asr := AuthenticationSetupRequest{
		AppID:     goodAppID,
		Timestamp: time.Now(),
		UserID:    "bar",
		KeyID:     "baz",
	}
	res, _ := postJSON("/v1/auth/request", asr)
	setupInfo := new(AuthenticationSetupReply)
	unmarshalJSONBody(res, setupInfo)

	// Get authentication iFrame
	res, _ = http.Get(ts.URL + "/auth/" + setupInfo.RequestID)
	var bodyBytes = make([]byte, 0)
	n, _ := res.Body.Read(bodyBytes)
	iFrameBody := string(bodyBytes[:n])
	if strings.Index(iFrameBody, "var data = ") == -1 {
		t.Errorf("Could not find data inside iFrameBody")
	}

	// Assert that data embedded in the iFrame is what we expect
	gleanedData := authenticateData{}

	// Get app info
	res, _ = http.Get(ts.URL + "/v1/info/" + asr.AppID)
	appInfo := new(AppIDInfoReply)
	unmarshalJSONBody(res, appInfo)

	authenticationRequest, _ := s.cache.GetAuthenticationRequest(setupInfo.RequestID)

	query := Key{AppID: asr.AppID, UserID: asr.UserID}
	var keys []string
	s.DB.Model(Key{}).Where(query).Select("PublicKey").Where(keys)

	correctData := authenticateData{
		id:           setupInfo.RequestID,
		counter:      authenticationRequest.counter,
		keys:         keys,
		challenge:    authenticationRequest.challenge,
		userID:       asr.UserID,
		appId:        asr.AppID,
		infoUrl:      appInfo.BaseURL + "/v1/info/" + asr.AppID,
		waitUrl:      appInfo.BaseURL + "/v1/auth/" + setupInfo.RequestID + "/wait",
		challengeUrl: appInfo.BaseURL + "/v1/auth/" + setupInfo.RequestID + "/challenge",
	}
	if !reflect.DeepEqual(gleanedData, correctData) {
		t.Errorf("Gleaned data was not expected")
	}
}
