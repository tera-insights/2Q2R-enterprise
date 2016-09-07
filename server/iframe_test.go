// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRegisterIFrameGeneration(t *testing.T) {
	// Set up registration request
	authData := CounterBasedAuthData{
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
	gleanedData := registerData{}
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
	bytes, _ := ioutil.ReadAll(res.Body)
	iFrameBody := string(bytes)
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
		RequestID:    setupInfo.RequestID,
		Counter:      authenticationRequest.Counter,
		Keys:         keys,
		Challenge:    authenticationRequest.Challenge,
		UserID:       asr.UserID,
		AppID:        asr.AppID,
		InfoURL:      appInfo.BaseURL + "/v1/info/" + asr.AppID,
		WaitURL:      appInfo.BaseURL + "/v1/auth/" + setupInfo.RequestID + "/wait",
		ChallengeURL: appInfo.BaseURL + "/v1/auth/" + setupInfo.RequestID + "/challenge",
	}
	if !reflect.DeepEqual(gleanedData, correctData) {
		t.Errorf("Gleaned data was not expected")
	}
}
