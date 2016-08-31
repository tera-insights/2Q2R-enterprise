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
var goodServerName = "foo"
var goodAppName = "bar"
var goodAppID = "123saWQgc3RyaW5nCg=="
var badAppID = "321saWQgc3RyaW5nCg=="
var goodBaseURL = "2q2r.org"
var goodKeyType = "P256"
var goodPublicKey = "notHidden!"
var goodPermissions = "[]"

func TestMain(m *testing.M) {
	code := m.Run()
	ts.Close()
	os.Exit(code)
}

func PostJSON(route string, d interface{}) (*http.Response, error) {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(d)
	return http.Post(ts.URL+route, "application/json; charset=utf-8", b)
}

func TestCreateNewApp(t *testing.T) {
	// Create new app
	res, _ := PostJSON("/v1/app/new", NewAppRequest{
		AppName: goodAppName,
	})
	appReply := new(NewAppReply)
	json.NewDecoder(res.Body).Decode(appReply)
	res.Body.Close()
	if appReply.AppID != goodAppID {
		t.Errorf("Expected app ID of %s. Got %s", goodAppID, appReply.AppID)
	}

	// Create new server
	res, _ = PostJSON("/v1/admin/server/new", NewServerRequest{
		ServerName:  goodServerName,
		AppID:       appReply.AppID,
		BaseURL:     goodBaseURL,
		KeyType:     goodKeyType,
		PublicKey:   goodPublicKey,
		Permissions: goodPermissions,
	})
	reply := new(NewServerReply)
	json.NewDecoder(res.Body).Decode(reply)
	res.Body.Close()
	if reply.ServerName != goodServerName {
		t.Errorf("Expected server name of %s. Got %s", goodServerName, reply.ServerName)
	}

	// Test app info
	res, _ = http.Get(ts.URL + "/v1/info/" + appReply.AppID)
	appInfo := new(AppIDInfoReply)
	json.NewDecoder(res.Body).Decode(appInfo)
	res.Body.Close()
	if appInfo.AppName != goodAppName {
		t.Errorf("Expected app name of %s. Got %s", goodAppName, appInfo.AppName)
	}

	// Test invalid method but with proper app ID
	res, _ = http.Post(ts.URL+"/v1/info/"+appReply.AppID, "", nil)
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Error("Expected `StatusMethodNotAllowed` when sending `POST` to " +
			"/v1/info/{appID}")
	}
}

func TestNonExistingApp(t *testing.T) {
	// Test nonexisting app
	res, _ := http.Get(ts.URL + "/v1/info/" + badAppID)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected response code of `http.StatusNotFound`. Got %d",
			res.StatusCode)
	}
}
