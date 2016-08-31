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
var goodAppID = "123saWQgc3RyaW5nCg=="
var badAppID = "321saWQgc3RyaW5nCg=="

func TestMain(m *testing.M) {
	code := m.Run()
	ts.Close()
	os.Exit(code)
}

func TestCreateNewApp(t *testing.T) {
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
	if reply.AppID != goodAppID {
		t.Errorf("Expected app ID of %s. Got %s", goodAppID, reply.AppID)
	}

	// Test app info
	res, _ = http.Get(ts.URL + "/v1/info/" + reply.AppID)
	appInfo := new(AppIDInfoReply)
	json.NewDecoder(res.Body).Decode(appInfo)
	res.Body.Close()
	if appInfo.AppName != name {
		t.Errorf("Expected app name of %s. Got %s", name, appInfo.AppName)
	}

	// Test nonexisting app
	res, _ = http.Get(ts.URL + "/v1/info/" + badAppID)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected response code of `http.StatusNotFound`. Got %d",
			res.StatusCode)
	}

	// Test invalid method but with proper app ID
	res, _ = http.Post(ts.URL+"/v1/info/"+reply.AppID, "", nil)
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Error("Expected `StatusMethodNotAllowed` when sending `POST` to " +
			"/v1/info/{appID}")
	}
}
