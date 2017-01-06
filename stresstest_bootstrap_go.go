// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package main

import (
	"2q2r/security"
	"2q2r/server"
	"2q2r/util"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"crypto/elliptic"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

// openssl ecparam -name prime256v1 -genkey -out app_server_priv.pem -noout
func main() {
	appServerKeyPath := "app_server_priv.pem"
	keyBytes, err := ioutil.ReadFile(appServerKeyPath)
	if keyBytes == nil {
		panic(errors.Errorf("Couldn't open app server private key at path %s", appServerKeyPath))
	}
	handle(err)

	p, _ := pem.Decode(keyBytes)
	if p == nil {
		panic(errors.New("File was not PEM-formatted"))
	}

	ec, err := x509.ParseECPrivateKey(p.Bytes)
	handle(err)

	db, err := gorm.Open("sqlite3", "test.db")
	handle(err)

	err = db.AutoMigrate(&server.AppInfo{}).
		AutoMigrate(&server.AppServerInfo{}).Error
	handle(err)

	err = db.Create(&server.AppInfo{
		ID:      "_T-zi0wzr7GCi4vsfsXsUuKOfmiWLiHBVbmJJPidvhA",
		AppName: "test-app",
	}).Error
	handle(err)

	pub := elliptic.Marshal(elliptic.P256(), ec.PublicKey.X,
		ec.PublicKey.Y)
	err = db.Create(&server.AppServerInfo{
		ID:        "_T-zi0wzr7GCi4vsfsXsUuKOfmiWLiHBVbmJJPidvhA",
		BaseURL:   "localhost:8080",
		AppID:     "_T-zi0wzr7GCi4vsfsXsUuKOfmiWLiHBVbmJJPidvhA",
		PublicKey: pub,
	}).Error
	handle(err)

	goServerKeyPath := "priv.pem"
	keyBytes, err = ioutil.ReadFile(goServerKeyPath)
	if keyBytes == nil {
		panic(errors.Errorf("Couldn't open 2Q2R server private key at path %s", goServerKeyPath))
	}
	handle(err)

	p, _ = pem.Decode(keyBytes)
	if p == nil {
		panic(errors.New("File was not PEM-formatted"))
	}

	ec, err = x509.ParseECPrivateKey(p.Bytes)
	handle(err)

	x, y := elliptic.Unmarshal(elliptic.P256(), pub)
	key := security.NewKeyGen().GetShared(x, y, ec.D.Bytes())
	fmt.Printf("Shared key = %s\n", util.EncodeBase64(key))
}
