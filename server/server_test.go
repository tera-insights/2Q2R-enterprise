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
	defer ts.Close()
	os.Exit(m.Run())
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
	appID := "really_fake"
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
	if info.ServerPubKey != "missing" {
		t.Errorf("Expected a missing public key. Got %s", info.ServerPubKey)
	}
}
