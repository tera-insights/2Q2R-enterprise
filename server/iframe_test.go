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
	challenge string
	userID    string
	appId     string
	infoUrl   string
	waitUrl   string
}

type authenticateData struct {
	id           string
	counter      int
	keys         []string
	challenge    string
	userID       string
	appId        string
	infoUrl      string
	waitUrl      string
	challengeUrl string
}

func TestRegisterIFrameGeneration(t *testing.T) {
	// Set up registration request
	requestID := "foo"
	appID := "bar"

	// Get registration iFrame
	res, _ := http.Get(ts.URL + "/register/" + appID)
	var bodyBytes = make([]byte, 0)
	n, err := res.Body.Read(bodyBytes)
	iFrameBody := string(bodyBytes[:n])
	if strings.Index(iFrameBody, "var data = ") == -1 {
		t.Errorf("Could not find data inside iFrameBody")
	}

	// Assert that data embedded in the iFrame is what we expect
	gleanedData := registerData{}
	var appInfo AppServerInfo
	appInfo = GetAppInfo(appID)
	registrationRequest := GetRegistrationRequest(requestID)
	correctData := registerData{
		id:        requestID,
		keyTypes:  []string{"2q2r", "u2f"},
		challenge: registrationRequest.challenge,
		userID:    registrationRequest.userID,
		appId:     appID,
		infoUrl:   appInfo.BaseURL + "/v1/info/" + registrationRequest.appID,
		waitUrl:   appInfo.BaseURL + "/v1/register/" + requestID + "/wait",
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
	var appInfo AppServerInfo
	appInfo = GetAppInfo(appID)
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
