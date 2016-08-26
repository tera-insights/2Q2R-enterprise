// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"2q2r/common"
)

func TestHello(t *testing.T) {
	ts := httptest.NewServer(Handler())
	defer ts.Close()
	appID := "my_id"
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
}
