package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHello(t *testing.T) {
	handler := Handler()
	ts := httptest.NewServer(handler)
	defer ts.Close()
	res, err := http.Get("/hello")
	if err != nil {
		t.Error(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if greeting != "world" {
		t.Errorf("Expected response \"world\". Got response %s", greeting)
	}
}
