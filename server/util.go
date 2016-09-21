// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

// CheckBase64 returns any errors encountered when deserializing a
// (supposedly) base-64 encoded string.
func CheckBase64(s string) error {
	_, err := decodeBase64(s)
	return err
}

// Copied from go-u2f
func encodeBase64(b []byte) string {
	s := base64.URLEncoding.EncodeToString(b)
	return strings.TrimRight(s, "=")
}

// Copied from go-u2f
func decodeBase64(s string) ([]byte, error) {
	for i := 0; i < len(s)%4; i++ {
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// Returns serverID, messageMAC
func getAuthDataFromHeaders(r *http.Request) (string, string) {
	authParts := strings.Split(r.Header.Get("authentication"), ":")
	return authParts[0], authParts[1]
}
