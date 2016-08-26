// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", http.NotFound)
	r.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "World, %q", html.EscapeString(r.URL.Path))
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
