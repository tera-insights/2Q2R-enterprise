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
	fmt.Printf("c.CleanTime = %s\n", c.CleanTime)
	s := server.NewServer(c)
	http.Handle("/", s.GetHandler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
