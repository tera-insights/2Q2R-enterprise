// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package main

import (
	"2q2r/server"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	r, err := os.Open("./config.yaml")
	if err != nil {
		panic(err)
	}
	c, err := server.MakeConfig(r, "yaml")
	if err != nil {
		panic(err)
	}
	s := server.NewServer(c)
	http.Handle("/", s.GetHandler())
	if c.Secure {
		fmt.Printf("Listening on HTTPS port %s\n", c.Port)
		log.Fatal(http.ListenAndServeTLS(c.Port, c.CertFile, c.KeyFile, nil))
	} else {
		fmt.Printf("Listening on HTTP port %s\n", c.Port)
		log.Fatal(http.ListenAndServe(c.Port, nil))
	}

}
