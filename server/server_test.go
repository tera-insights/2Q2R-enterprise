// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHello(t *testing.T) {
	ts := httptest.NewServer(Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Error(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if string(greeting) != "world" {
		t.Errorf("Expected response \"world\". Got response %s", string(greeting))
	}
}
