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

	// Get registration iframe
	res, _ := http.Get(ts.URL + "/v1/info/" + badAppID)
	var bodyBytes = make([]byte, 0)
	n, err := res.Body.Read(bodyBytes)
	iFrameBody := string(bodyBytes[:n])
	if strings.Index(iFrameBody, "var data = ") == -1 {
		t.Errorf("Could not find data inside iFrameBody")
	}

	// Assert that data embedded in the iframe is what we expect
	gleanedData := registerData{}
	correctData := registerData{
		id:        requestID,
		keyTypes:  {"2q2r", "u2f"},
		challenge: GetRegistrationChallenge(requestID),
		userID:    cReq.userID,
		appId:     cReq.appId,
		infoUrl:   info.baseURL + "/v1/info/" + cReq.appId,
		waitUrl:   info.baseURL + "/v1/register/" + requestID + "/wait",
	}
	if !reflect.DeepEqual(gleanedData) {
		t.Errorf("Gleaned data was not expected")
	}
}

func TestAuthenticateIFrameGeneration(t *testing.T) {
	iFrameBody := GenerateAuthenticateIFrame()
	requestID := "foo"
	gleanedData := authenticateData{}
	correctData := authenticateData{
		id:           requestID,
		counter:      cReq.counter,
		keys:         keys,
		challenge:    GetAuthenticationChallenge(requestID),
		userID:       cReq.userID,
		appId:        cReq.appId,
		infoUrl:      info.baseURL + "/v1/info/" + cReq.appId,
		waitUrl:      info.baseURL + "/v1/auth/" + requestID + "/wait",
		challengeUrl: info.baseURL + "/v1/auth/" + requestID + "/challenge",
	}
	if !reflect.DeepEqual(gleanedData, correctData) {
		t.Errorf("Gleaned data was not expected")
	}
}
