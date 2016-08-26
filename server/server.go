// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"2q2r/common"
)

// Taken from https://git.io/v6xHB
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encErr := encoder.Encode(data)
	if encErr != nil {
		log.Printf("Failed to encode data as JSON: %v", encErr)
	}
}

func Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/v1/info/{appID}", func(w http.ResponseWriter, r *http.Request) {
		data := new(common.AppIDInfoReply)
		data.AppName = "foo_app"
		data.BaseURL = "https://example.com/baz"
		data.AppID = mux.Vars(r)["appID"]
		data.ServerPubKey = "Call me beep me"
		data.ServerKeyType = "P256"
		writeJSON(w, http.StatusOK, data)
	})
	return r
}

func New() (*http.Server, error) {
	s := &http.Server{
		Addr:           ":8080",
		Handler:        Handler(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return s, nil
}
