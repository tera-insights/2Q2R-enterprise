// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"2q2r/common"
)

var ts = httptest.NewServer(Handler())

func TestMain(m *testing.M) {
	code := m.Run()
	ts.Close()
	os.Exit(code)
}

func CreateNewApp(t *testing.T) {
	// Create new app

	// Test app info

	// Test nonexisting app info
}

func TestExistingAppID(t *testing.T) {
	appID := "really_here"
	res, err := http.Get(ts.URL + "/v1/info/" + appID)
	if err != nil {
		t.Error(err)
	}
	info := new(common.AppIDInfoReply)
	err = json.NewDecoder(res.Body).Decode(info)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if info.AppID != appID {
		t.Errorf("Expected appID %s. Got response %s", appID, info.AppID)
	}
	if info.ServerPubKey == "missing" {
		t.Errorf("Expected a valid public key. Got %s", info.ServerPubKey)
	}
}

func TestNonExistingAppID(t *testing.T) {
	_, err := http.Get(ts.URL + "/v1/info/really_fake")
	if err == nil {
		t.Error("Expected error when accessing fake appID. Received none.")
	}
}
