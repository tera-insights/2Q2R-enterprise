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

func postJSON(route string, d interface{}) (*http.Response, error) {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(d)
	return http.Post(ts.URL+route, "application/json; charset=utf-8", b)
}

func unmarshalJSONBody(r *http.Response, d interface{}) {
	json.NewDecoder(r.Body).Decode(d)
	r.Body.Close()
}

func TestCreateNewApp(t *testing.T) {
	// Create new app
	res, _ := postJSON("/v1/app/new", NewAppRequest{
		AppName: goodAppName,
	})
	appReply := new(NewAppReply)
	unmarshalJSONBody(res, appReply)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Could not create app. Status code: %d\n", res.StatusCode)
	}

	// Create new server
	res, _ = postJSON("/v1/admin/server/new", NewServerRequest{
		ServerName:  goodServerName,
		AppID:       appReply.AppID,
		BaseURL:     goodBaseURL,
		KeyType:     goodKeyType,
		PublicKey:   goodPublicKey,
		Permissions: goodPermissions,
	})
	newReply := new(NewServerReply)
	unmarshalJSONBody(res, newReply)
	if newReply.ServerName != goodServerName {
		t.Errorf("Expected server name of %s. Got %s", goodServerName, newReply.ServerName)
	}

	// Test app info
	res, _ = http.Get(ts.URL + "/v1/info/" + appReply.AppID)
	appInfo := new(AppIDInfoReply)
	unmarshalJSONBody(res, appInfo)
	if appInfo.AppName != goodAppName {
		t.Errorf("Expected app name of %s. Got %s", goodAppName, appInfo.AppName)
	}

	// Test server info
	res, _ = postJSON("/v1/admin/server/get", AppServerInfoRequest{
		ServerID: newReply.ServerID,
	})
	serverInfo := new(AppServerInfo)
	unmarshalJSONBody(res, serverInfo)
	if serverInfo.ServerName != goodServerName {
		t.Errorf("Expected server name of %s. Got %s", goodServerName, serverInfo.ServerName)
	}

	// Delete server
	res, _ = postJSON("/v1/admin/server/delete", DeleteServerRequest{})
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected `http.StatusOK`, but got %d\n", res.StatusCode)
	}

	// Assert that server was deleted
	res, _ = postJSON("/v1/admin/server/get", AppServerInfoRequest{
		ServerID: newReply.ServerID,
	})
	deletedServerInfo := new(AppServerInfo)
	unmarshalJSONBody(res, deletedServerInfo)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected response code of `http.StatusNotFound`. Got %d",
			res.StatusCode)
	}

	// Test invalid method but with proper app ID
	res, _ = http.Post(ts.URL+"/v1/info/"+appReply.AppID, "", nil)
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected `http.StatusMethodNotAllowed, but got %d\n",
			res.StatusCode)
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
