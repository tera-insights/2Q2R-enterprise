// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRegisterIFrameGeneration(t *testing.T) {
	// Set up registration request
	authData := AuthenticationData{
		Counter:  0,
		ServerID: "foo",
	}
	registrationRequest := RegistrationSetupRequest{
		AppID:              goodAppID,
		Timestamp:          time.Now(),
		UserID:             "bar",
		AuthenticationData: authData,
	}
	res, _ := postJSON("/v1/register/request", registrationRequest)
	setupInfo := new(RegistrationSetupReply)
	unmarshalJSONBody(res, setupInfo)

	// Get registration iFrame
	res, _ = http.Get(ts.URL + "/register/" + setupInfo.RequestID)
	bytes, _ := ioutil.ReadAll(res.Body)
	iFrameBody := string(bytes)
	if strings.Index(iFrameBody, "var data = ") == -1 {
		t.Errorf("Could not find data inside iFrameBody")
	}

	// Get app info
	res, _ = http.Get(ts.URL + "/v1/info/" + registrationRequest.AppID)
	appInfo := new(AppIDInfoReply)
	unmarshalJSONBody(res, appInfo)

	// Assert that data embedded in the iFrame is what we expect
	startIndex := strings.Index(iFrameBody, "{")
	endIndex := strings.Index(iFrameBody, ";")
	if startIndex == -1 || endIndex == -1 {
		t.Errorf("Could not find data inside iFrameBody")
	}

	embedded := iFrameBody[startIndex:endIndex]
	gleanedData := new(registerData)
	json.NewDecoder(strings.NewReader(embedded)).Decode(gleanedData)

	cachedRequest, _ := s.cache.GetRegistrationRequest(setupInfo.RequestID)
	correctData := registerData{
		RequestID: setupInfo.RequestID,
		KeyTypes:  []string{"2q2r", "u2f"},
		Challenge: cachedRequest.Challenge,
		UserID:    registrationRequest.UserID,
		AppID:     registrationRequest.AppID,
		InfoURL:   appInfo.BaseURL + "/v1/info/" + registrationRequest.AppID,
		WaitURL:   appInfo.BaseURL + "/v1/register/" + setupInfo.RequestID + "/wait",
	}
	if gleanedData.RequestID != correctData.RequestID {
		t.Errorf("RequestID was not properly templated")
	}
	if !reflect.DeepEqual(gleanedData.KeyTypes, correctData.KeyTypes) {
		t.Errorf("KeyTypes were not properly templated")
	}
	if !reflect.DeepEqual(gleanedData.Challenge, correctData.Challenge) {
		t.Errorf("Challenge was not properly templated")
	}
	if gleanedData.UserID != correctData.UserID {
		t.Errorf("UserID was not properly templated")
	}
	if gleanedData.AppID != correctData.AppID {
		t.Errorf("AppID was not properly templated")
	}
	if gleanedData.InfoURL != correctData.InfoURL {
		t.Errorf("InfoURL was not properly templated")
	}
	if gleanedData.WaitURL != correctData.WaitURL {
		t.Errorf("WaitURL was not properly templated")
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
	bytes, _ := ioutil.ReadAll(res.Body)
	iFrameBody := string(bytes)

	startIndex := strings.Index(iFrameBody, "{")
	endIndex := strings.Index(iFrameBody, ";")
	if startIndex == -1 || endIndex == -1 {
		t.Errorf("Could not find data inside iFrameBody")
	}

	embedded := iFrameBody[startIndex:endIndex]
	gleanedData := new(authenticateData)
	json.NewDecoder(strings.NewReader(embedded)).Decode(gleanedData)

	// Get app info
	res, _ = http.Get(ts.URL + "/v1/info/" + asr.AppID)
	appInfo := new(AppIDInfoReply)
	unmarshalJSONBody(res, appInfo)

	authenticationRequest, _ := s.cache.GetAuthenticationRequest(setupInfo.RequestID)

	keyID := "foo"
	query := Key{AppID: asr.AppID, UserID: asr.UserID}
	var keys []string
	s.DB.Model(Key{}).Where(query).Select("PublicKey").Where(keys)
	correctCounter := 0
	s.DB.Model(Key{}).Where(Key{KeyID: keyID}).Count(correctCounter)

	correctData := authenticateData{
		RequestID:    setupInfo.RequestID,
		Counter:      correctCounter,
		Keys:         keys,
		Challenge:    authenticationRequest.Challenge,
		UserID:       asr.UserID,
		AppID:        asr.AppID,
		InfoURL:      appInfo.BaseURL + "/v1/info/" + asr.AppID,
		WaitURL:      appInfo.BaseURL + "/v1/auth/" + setupInfo.RequestID + "/wait",
		ChallengeURL: appInfo.BaseURL + "/v1/auth/" + setupInfo.RequestID + "/challenge",
	}
	if gleanedData.RequestID != correctData.RequestID {
		t.Errorf("RequestID was not properly templated")
	}
	if gleanedData.Counter != correctData.Counter {
		t.Errorf("Counter was not properly templated")
	}
	if !reflect.DeepEqual(gleanedData.Keys, correctData.Keys) {
		t.Errorf("Keys were not properly templated")
	}
	if !reflect.DeepEqual(gleanedData.Challenge, correctData.Challenge) {
		t.Errorf("Challenge was not properly templated")
	}
	if gleanedData.UserID != correctData.UserID {
		t.Errorf("UserID was not properly templated")
	}
	if gleanedData.AppID != correctData.AppID {
		t.Errorf("AppID was not properly templated")
	}
	if gleanedData.InfoURL != correctData.InfoURL {
		t.Errorf("InfoURL was not properly templated")
	}
	if gleanedData.WaitURL != correctData.WaitURL {
		t.Errorf("WaitURL was not properly templated")
	}
	if gleanedData.ChallengeURL != correctData.ChallengeURL {
		t.Errorf("ChallengeURL was not properly templated")
	}
}
