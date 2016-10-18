// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package main

import (
	"2q2r/server"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

func main() {
	var configPath string
	var configType string

	flag.StringVar(&configPath, "config-path", "./config.yaml",
		"Path to server configuration file")
	flag.StringVar(&configType, "config-type", "yaml",
		"Filetype of config file. Case insensitive. Must be either JSON, "+
			"YAML, HCL, or Java")
	flag.Parse()
	pathSet := false
	typeSet := false
	flag.CommandLine.Visit(func(f *flag.Flag) {
		if f.Name == "config-path" {
			pathSet = true
		}
		if f.Name == "config-type" {
			typeSet = true
		}
	})
	if !pathSet {
		fmt.Printf("No config file set! Using default path %s\n",
			flag.Lookup("config-path").DefValue)
	}
	if !typeSet {
		fmt.Printf("No config type set! Using default type %s\n",
			flag.Lookup("config-type").DefValue)
	}
	r, err := os.Open(configPath)
	if err != nil {
		s := fmt.Sprintf("Failed to open config file at path %s\n", configPath)
		panic(errors.Wrap(err, s))
	}
	c := server.MakeConfig(r, configType)
	s := server.NewServer(c)
	http.Handle("/", s.GetHandler())
	if c.HTTPS {
		fmt.Printf("Listening on HTTPS port %s\n", c.Port)
		log.Fatal(http.ListenAndServeTLS(c.Port, c.CertFile, c.KeyFile, nil))
	} else {
		fmt.Printf("Listening on HTTP port %s\n", c.Port)
		log.Fatal(http.ListenAndServe(c.Port, nil))
	}

}
