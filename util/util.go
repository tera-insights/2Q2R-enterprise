// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package util

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

// CheckBase64 returns any errors encountered when deserializing a
// (supposedly) base-64 encoded string.
func CheckBase64(s string) error {
	_, err := DecodeBase64(s)
	return err
}

// EncodeBase64 encodes bytes in base-64 using web encoding with no padding.
// Its implementation is from go-u2f.
func EncodeBase64(b []byte) string {
	s := base64.URLEncoding.EncodeToString(b)
	return strings.TrimRight(s, "=")
}

// Copied from go-u2f
func DecodeBase64(s string) ([]byte, error) {
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

type BubbledError struct {
	StatusCode int
	Message    string
	Info       interface{}
}

func OptionalPanic(err error, code int, message string) {
	if err != nil {
		panic(BubbledError{
			StatusCode: code,
			Message:    message,
		})
	}
}

func OptionalInternalPanic(err error, message string) {
	OptionalPanic(err, http.StatusInternalServerError, message)
}

func OptionalBadRequestPanic(err error, message string) {
	OptionalPanic(err, http.StatusBadRequest, message)
}

func PanicIfFalse(b bool, c int, m string) {
	if !b {
		panic(BubbledError{
			StatusCode: c,
			Message:    m,
		})
	}
}
