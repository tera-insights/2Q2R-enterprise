// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// CheckBase64 returns any errors encountered when deserializing a
// (supposedly) base-64 encoded string.
func CheckBase64(s string) error {
	_, err := decodeBase64(s)
	return err
}

// EncodeBase64 encodes bytes in base-64 using web encoding with no padding.
// Its implementation is from go-u2f.
func EncodeBase64(b []byte) string {
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

// RandString generates `n` random bytes and returns them in a base-64 string
// that has been web-encoded with no padding.
func RandString(n int) (string, error) {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return EncodeBase64(bytes), nil
}

// Returns error, ID, messageMAC
func getAuthDataFromHeaders(r *http.Request) (string, string, error) {
	parts := strings.Split(r.Header.Get("X-Authentication"), ":")
	if len(parts) != 2 {
		return "", "", errors.Errorf("Found %d parts, expected 2", len(parts))
	}
	return parts[0], parts[1], nil
}

type bubbledError struct {
	StatusCode int
	Message    string
	Info       interface{}
}

func optionalPanic(err error, code int, message string) {
	if err != nil {
		panic(bubbledError{
			StatusCode: code,
			Message:    message,
		})
	}
}

func optionalInternalPanic(err error, message string) {
	optionalPanic(err, http.StatusInternalServerError, message)
}

func optionalBadRequestPanic(err error, message string) {
	optionalPanic(err, http.StatusBadRequest, message)
}

func panicIfFalse(b bool, c int, m string) {
	if !b {
		panic(bubbledError{
			StatusCode: c,
			Message:    m,
		})
	}
}
