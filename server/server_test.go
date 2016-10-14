// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var s = NewServer(Config{
	Port:           ":8080",
	DatabaseType:   "sqlite3",
	DatabaseName:   "test.db",
	ExpirationTime: 5 * time.Minute,
	CleanTime:      30 * time.Second,
	BaseURL:        "127.0.0.1",
})

var ts = httptest.NewUnstartedServer(s.GetHandler())
var goodServerName = "foo"
var goodAppName = "bar"
var badAppID = encodeBase64([]byte("321saWQgc3RyaW5nCg=="))
var goodBaseURL = "2q2r.org"
var goodKeyType = "P256"
var goodPublicKey = "notHidden!"
var goodPermissions = "[]"
var goodAppID string

// Create new app for use in other
func TestMain(m *testing.M) {
	l, err := net.Listen("tcp", s.c.BaseURL+s.c.Port)
	if err != nil {
		panic(fmt.Sprintf("Failed to listen on a port: %+v\n", err))
	}
	ts.Listener = l // overwriting the default random port given by httptest
	ts.Start()
	defer ts.Close()
	res, _ := postJSON("/admin/apps", NewAppRequest{
		AppName: goodAppName,
	})
	appReply := new(NewAppReply)
	unmarshalJSONBody(res, appReply)
	goodAppID = appReply.AppID
	code := m.Run()
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

func checkStatus(t *testing.T, expected int, r *http.Response) {
	if expected != r.StatusCode {
		t.Errorf("Expected status code %d but got %d\n", expected, r.StatusCode)
	}
}

func TestCreateNewApp(t *testing.T) {
	// Create new server
	res, _ := postJSON("/admin/servers", NewServerRequest{
		ServerName:  goodServerName,
		AppID:       goodAppID,
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
	res, _ = http.Get(ts.URL + "/v1/info/" + goodAppID)
	appInfo := new(AppIDInfoReply)
	unmarshalJSONBody(res, appInfo)
	if appInfo.AppName != goodAppName {
		t.Errorf("Expected app name of %s. Got %s", goodAppName, appInfo.AppName)
	}

	// Test server info
	res, _ = postJSON("/admin/servers", AppServerInfoRequest{
		ServerID: newReply.ServerID,
	})
	serverInfo := new(AppServerInfo)
	unmarshalJSONBody(res, serverInfo)
	if serverInfo.ServerName != goodServerName {
		t.Errorf("Expected server name of %s. Got %s", goodServerName, serverInfo.ServerName)
	}

	// Delete server
	res, _ = postJSON("/admin/servers/"+serverInfo.ServerID, DeleteServerRequest{})
	checkStatus(t, http.StatusOK, res)

	// Assert that server was deleted
	res, _ = postJSON("/admin/servers", AppServerInfoRequest{
		ServerID: newReply.ServerID,
	})
	deletedServerInfo := new(AppServerInfo)
	unmarshalJSONBody(res, deletedServerInfo)
	checkStatus(t, http.StatusNotFound, res)

	// Test invalid method but with proper app ID
	res, _ = http.Post(ts.URL+"/v1/info/"+goodAppID, "", nil)
	checkStatus(t, http.StatusMethodNotAllowed, res)
}

func TestNonExistingApp(t *testing.T) {
	// Test nonexisting app
	res, _ := http.Get(ts.URL + "/v1/info/" + badAppID)
	checkStatus(t, http.StatusNotFound, res)
}
