// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package main

import (
	"github.com/alinVD/2Q2R-enterprise/util"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh/terminal"

	"flag"

	"io"

	"os"

	"github.com/pkg/errors"
)

func main() {
	var keyPath string
	var encrypted bool
	var infoPath string
	flag.StringVar(&keyPath, "key-path", "./app_server_priv.pem",
		"Path to PEM file containing the Tera Insights Private Key")
	flag.BoolVar(&encrypted, "encrypted", false,
		"When true, prompts for a password to decrypt the private key")
	flag.StringVar(&infoPath, "info-path", "./admin_info.json",
		"Path to JSON file containing info about the first admin")
	flag.Parse()

	fmt.Printf("Using key path = %s\n", keyPath)
	fmt.Printf("Using info path = %s\n", infoPath)

	keyBytes, err := ioutil.ReadFile(keyPath)
	if keyBytes == nil {
		panic(errors.Errorf("Couldn't open private key at path %s", keyPath))
	}

	p, _ := pem.Decode(keyBytes)
	if p == nil {
		panic(errors.New("File was not PEM-formatted"))
	}

	var key []byte
	if encrypted {
		fmt.Print("Please enter private key password:\t")
		pwd, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			panic(errors.Wrap(err, "Couldn't read key password"))
		}

		key, err = x509.DecryptPEMBlock(p, pwd)
		if err != nil {
			panic(errors.Wrap(err, "Couldn't decrypt private key"))
		}
	} else {
		key = p.Bytes
	}

	priv, err := x509.ParsePKCS1PrivateKey(key)
	if err != nil {
		panic(errors.Wrap(err, "Couldn't open DER bytes as RSA private key"))
	}

	raw, err := ioutil.ReadFile(infoPath)
	var info map[string]string
	err = json.Unmarshal(raw, &info)
	if err != nil {
		panic(errors.Wrapf(err, "Couldn't unmarshal JSON file at path %s", infoPath))
	}

	pub, found := info["publicKey"]
	if !found {
		panic(errors.New("No public key at key \"publicKey\""))
	}

	ownerID, err := util.RandString(32)
	if err != nil {
		panic(errors.Wrap(err, "Couldn't generate admin ID"))
	}

	h := crypto.SHA256.New()
	io.WriteString(h, pub)
	io.WriteString(h, "signing")
	io.WriteString(h, ownerID)
	signature, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, h.Sum(nil), nil)
	if err != nil {
		panic(errors.Wrap(err, "Couldn't sign admin's public key"))
	}

	info["ownerID"] = ownerID;
	info["signature"] = util.EncodeBase64(signature)
	bytes, _ := json.Marshal(info)
	err = ioutil.WriteFile(infoPath, bytes, os.ModePerm)
	if err != nil {
		panic(errors.Wrap(err, "Couldn't save file with signature"))
	}

	fmt.Println("Output written to " + infoPath)
}
