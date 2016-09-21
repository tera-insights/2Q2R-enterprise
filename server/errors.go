// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"fmt"
	"net/http"
)

// Taken from https://git.io/viJiC.
type errorResponse struct {
	Message string
	Info    interface{} `json:",omitempty"`
}

// StatusError represents an error that has a built-in HTTP status code,
// descriptive error code and optionally extra information about the error.
// Taken from https://git.io/viJis.
type StatusError interface {
	error
	StatusCode() int
	ErrorCode() string
	Info() interface{}
}

// MethodNotAllowedError represents the "HTTP 405 Method Not Allowed" error.
// Taken from https://git.io/viJPm.
type MethodNotAllowedError string

func (e MethodNotAllowedError) Error() string {
	return fmt.Sprintf("Method %v not allowed", string(e))
}

// StatusCode implements part of the StatusError interface.
func (e MethodNotAllowedError) StatusCode() int {
	return http.StatusMethodNotAllowed
}

// ErrorCode implements part of the StatusError interface.
func (e MethodNotAllowedError) ErrorCode() string {
	return "method-not-allowed"
}

// Info implements part of the StatusError interface.
func (e MethodNotAllowedError) Info() interface{} {
	return nil
}
