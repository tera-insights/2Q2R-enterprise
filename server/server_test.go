// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var s = New(Config{
	8080,
	"sqlite3",
	"test.db",
})
var ts = httptest.NewServer(s.GetHandler())

func TestMain(m *testing.M) {
	code := m.Run()
	ts.Close()
	os.Exit(code)
}

func CreateNewApp(t *testing.T) {
	name := "foo"

	// Create new app
	req := NewAppRequest{
		AppName:  name,
		AuthType: "public-key",
		AuthData: "{bar: baz}",
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(req)
	res, _ := http.Post(ts.URL+"/v1/app/new", "application/json; charset=utf-8", b)
	reply := new(NewAppReply)
	json.NewDecoder(res.Body).Decode(reply)
	res.Body.Close()
	if reply.AppID != "123" {
		t.Errorf("Expected app ID of 123. Got %s", reply.AppID)
	}

	// Test app info
	res, _ = http.Get(ts.URL + "/v1/info/" + reply.AppID)
	appInfo := new(AppIDInfoReply)
	json.NewDecoder(res.Body).Decode(appInfo)
	res.Body.Close()
	if appInfo.AppName != name {
		t.Errorf("Expected app name of %s. Got %s", name, appInfo.AppName)
	}
}

func TestExistingAppID(t *testing.T) {
	appID := "really_here"
	res, err := http.Get(ts.URL + "/v1/info/" + appID)
	if err != nil {
		t.Error(err)
	}
	info := new(AppIDInfoReply)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code to be `http.StatusOK` Got %d",
			res.StatusCode)
	}
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
	res, _ := http.Get(ts.URL + "/v1/info/really_fake")
	if res.StatusCode != http.StatusNotFound {
		t.Error("Expected `StatusNotFound` when accessing fake appID.")
	}
}

func TestInvalidMethod(t *testing.T) {
	res, _ := http.Post(ts.URL+"/v1/info/doesnt_matter", "", nil)
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Error("Expected `StatusMethodNotAllowed` when sending `POST` to " +
			"/v1/info/{appID}")
	}
}
