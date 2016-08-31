// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import "github.com/jinzhu/gorm"

// AppInfo is the Gorm model that holds information about an app.
type AppInfo struct {
	gorm.Model

	AppID   string
	AppName string
}

// AppServerInfo is the Gorm model that holds information about an app server.
type AppServerInfo struct {
	gorm.Model

	ServerID string

	// Base URL for users to connect to
	BaseURL string

	// A server can only serve one app
	AppID string

	// P256, etc.
	KeyType string

	// JSON
	PublicKey string

	// JSON array containing a subset of ["Register", "Delete", "Login"]
	Permissions string
}
