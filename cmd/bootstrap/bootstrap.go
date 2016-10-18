// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package main

import (
	"2q2r/server"
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/pkg/errors"
)

// 1. Create new admin request from referenced JSON file
// 2. Check that the signature on the admin is valid
// 3. Add the signature, admin to the database as an active admin
func main() {
	var filePath string

	flag.StringVar(&filePath, "file-path", "./bootstrap.json",
		"Path to JSON file containing info to bootstrap the first admin")
	flag.Parse()
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(errors.Errorf("Could not open bootstrap file at path %s\n"+
			"Reason: %+v", filePath, err))
	}

	var request server.NewAdminRequest
	err = json.Unmarshal(raw, &request)
	if err != nil {
		panic(errors.Errorf("Could not unmarshal JSON file at path %s\n"+
			"Reason: %+v", filePath, err))
	}
}
