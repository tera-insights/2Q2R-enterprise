// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
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
	registrationRequest := RegistrationRequest{}
	res, _ := postJSON("/v1/register/request", registrationRequest)
	setupInfo := new(RegistrationRequestReply)
	unmarshalJSONBody(res, setupInfo)

	// Get registration iFrame
	res, _ = http.Get(ts.URL + "/register/" + registrationRequest.AppID)
	var bodyBytes = make([]byte, 0)
	n, err := res.Body.Read(bodyBytes)
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
	requestID := "foo"
	appID := "bar"

	// Get authentication iFrame
	res, _ := http.Get(ts.URL + "/auth/" + appID)
	var bodyBytes = make([]byte, 0)
	n, err := res.Body.Read(bodyBytes)
	iFrameBody := string(bodyBytes[:n])
	if strings.Index(iFrameBody, "var data = ") == -1 {
		t.Errorf("Could not find data inside iFrameBody")
	}

	// Assert that data embedded in the iFrame is what we expect
	gleanedData := authenticateData{}

	// Get app info
	res, _ = http.Get(ts.URL + "/v1/info/" + appID)
	appInfo := new(AppIDInfoReply)
	unmarshalJSONBody(res, appInfo)

	authenticationRequest := GetAuthenticationRequest(requestID)
	correctData := authenticateData{
		id:           requestID,
		counter:      authenticationRequest.counter,
		keys:         GetKeys(appID, authenticationRequest.userID),
		challenge:    GetAuthenticationChallenge(requestID),
		userID:       authenticationRequest.userID,
		appId:        authenticationRequest.appId,
		infoUrl:      appInfo.BaseURL + "/v1/info/" + authenticationRequest.appId,
		waitUrl:      appInfo.BaseURL + "/v1/auth/" + requestID + "/wait",
		challengeUrl: appInfo.BaseURL + "/v1/auth/" + requestID + "/challenge",
	}
	if !reflect.DeepEqual(gleanedData, correctData) {
		t.Errorf("Gleaned data was not expected")
	}
}
