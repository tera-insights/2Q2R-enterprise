// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package main

import (
	"2q2r/server"
	"log"
	"net/http"
	"time"
)

func main() {
	c := server.Config{
		Port:                            ":8080",
		DatabaseType:                    "sqlite3",
		DatabaseName:                    "server/test.db",
		ExpirationTime:                  5 * time.Minute,
		CleanTime:                       30 * time.Second,
		BaseURL:                         "127.0.0.1",
		ListenerExpirationTime:          3 * time.Minute,
		RecentlyCompletedExpirationTime: 1 * time.Minute,
	}
	s := server.NewServer(c)
	http.Handle("/", s.GetHandler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
