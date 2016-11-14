// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package main

import (
	"2q2r/server"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// 1. Create new admin request from referenced JSON file
// 2. Check that the signature on the admin is valid
// 3. Add the signature, admin to the database as an active admin
func main() {
	var bootstrapPath string
	var configPath string
	var configType string

	flag.StringVar(&bootstrapPath, "file-path", "./bootstrap.json",
		"Path to JSON file containing info to bootstrap the first admin")
	flag.StringVar(&configPath, "config-path", "./config.yaml",
		"Path to server configuration file")
	flag.StringVar(&configType, "config-type", "yaml",
		"Filetype of config file. Case insensitive. Must be either JSON, "+
			"YAML, HCL, or Java")

	flag.Parse()
	raw, err := ioutil.ReadFile(bootstrapPath)
	if err != nil {
		panic(errors.Wrapf(err, "Could not open bootstrap file at path %s",
			bootstrapPath))
	}

	var req server.NewAdminRequest
	err = json.Unmarshal(raw, &req)
	if err != nil {
		panic(errors.Wrapf(err, "Could not unmarshal JSON file at path %s", bootstrapPath))
	}

	r, err := os.Open(configPath)
	if err != nil {
		panic(errors.Wrapf(err, "Could not open config file at path %s", configPath))
	}

	s := server.NewServer(r, configType)

	// Verify the passed signature
	adminID, err := server.RandString(32)
	if err != nil {
		panic(errors.Wrap(err, "Could not generate random ID for admin"))
	}
	signature := server.KeySignature{
		SigningPublicKey: req.SigningPublicKey,
		SignedPublicKey:  req.PublicKey,
		Type:             "signing",
		OwnerID:          adminID,
		Signature:        req.Signature,
	}
	err = s.VerifySignature(signature)
	if err != nil {
		panic(errors.Wrap(err, "Could not verify signature"))
	}

	// Transactionally add the signing key, admin, and key signature to the DB
	tx := s.DB.Begin()

	keyID, err := server.RandString(32)
	if err != nil {
		panic(errors.Wrap(err, "Could not generate random ID for key"))
	}

	if err = tx.Create(&server.SigningKey{
		ID:        keyID,
		IV:        req.IV,
		Salt:      req.Salt,
		PublicKey: req.PublicKey,
	}).Error; err != nil {
		tx.Rollback()
		panic(errors.Wrap(err, "Could not save signing key to the database"))
	}

	encodedPermissions, err := json.Marshal(req.Permissions)
	if err != nil {
		panic(errors.Wrapf(err, "Could not marshal %s as JSON", req.Permissions))
	}

	if err = tx.Create(&server.Admin{
		ID:                  adminID,
		Status:              "active",
		Name:                req.Name,
		Email:               req.Email,
		Permissions:         string(encodedPermissions),
		Role:                "superadmin",
		PrimarySigningKeyID: keyID,
		AdminFor:            "1",
	}).Error; err != nil {
		tx.Rollback()
		panic(errors.Wrap(err, "Could not save admin to the database"))
	}

	if err = tx.Create(&signature).Error; err != nil {
		tx.Rollback()
		panic(errors.Wrap(err, "Could not save key signature to the database"))
	}

	if err = tx.Commit().Error; err != nil {
		panic(errors.Wrap(err, "Could not commit changes to database"))
	}

	fmt.Println("Successfully added admin, signing key, and key signature " +
		"to database")
}
