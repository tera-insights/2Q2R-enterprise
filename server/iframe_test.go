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

	"github.com/stretchr/testify/require"
)

func extractEmbeddedData(route string, o interface{}) {
	res, _ := http.Get(ts.URL + route)
	bytes, _ := ioutil.ReadAll(res.Body)
	iFrameBody := string(bytes)

	// require that data embedded in the iFrame is what we expect
	startIndex := strings.Index(iFrameBody, "{")
	endIndex := strings.Index(iFrameBody, ";")
	embedded := iFrameBody[startIndex:endIndex]
	json.NewDecoder(strings.NewReader(embedded)).Decode(o)
}

func TestRegisterIFrameGeneration(t *testing.T) {
	res, err := http.Get("/v1/register/request/bar")
	require.Nil(t, err)
	setupInfo := new(RegistrationSetupReply)
	unmarshalJSONBody(res, setupInfo)

	// Get registration iFrame
	gleanedData := new(registerData)
	extractEmbeddedData("/register/"+setupInfo.RequestID, gleanedData)

	// Get app info
	res, _ = http.Get(ts.URL + "/v1/info/" + goodAppID)
	appInfo := new(AppIDInfoReply)
	unmarshalJSONBody(res, appInfo)

	cachedRequest, _ := s.cache.GetRegistrationRequest(setupInfo.RequestID)
	correctData := registerData{
		RequestID: setupInfo.RequestID,
		KeyTypes:  []string{"2q2r", "u2f"},
		Challenge: encodeBase64(cachedRequest.Challenge.Challenge),
		UserID:    "bar",
		AppID:     goodAppID,
		InfoURL:   appInfo.BaseURL + "/v1/info/" + goodAppID,
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
	require.NotEqual(t, startIndex, -1)
	require.NotEqual(t, endIndex, -1)

	embedded := iFrameBody[startIndex:endIndex]
	gleanedData := new(authenticateData)
	json.NewDecoder(strings.NewReader(embedded)).Decode(gleanedData)

	// Get app info
	res, _ = http.Get(ts.URL + "/v1/info/" + asr.AppID)
	appInfo := new(AppIDInfoReply)
	unmarshalJSONBody(res, appInfo)

	authenticationRequest, _ := s.cache.GetAuthenticationRequest(setupInfo.RequestID)

	query := Key{AppID: asr.AppID, UserID: asr.UserID}
	rows, err := s.DB.Model(&Key{}).Where(query).Select([]string{"key_id", "type", "name"}).Rows()
	require.Nil(t, err)
	defer rows.Close()
	var keys []keyDataToEmbed
	for rows.Next() {
		var keyID string
		var keyType string
		var name string
		err := rows.Scan(&keyID, &keyType, &name)
		require.Nil(t, err)
		keys = append(keys, keyDataToEmbed{
			KeyID: keyID,
			Type:  keyType,
			Name:  name,
		})
	}

	correctData := authenticateData{
		RequestID:    setupInfo.RequestID,
		Counter:      1,
		Keys:         keys,
		Challenge:    encodeBase64(authenticationRequest.Challenge.Challenge),
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
