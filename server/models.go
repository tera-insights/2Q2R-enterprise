// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import "github.com/jinzhu/gorm"

// AppInfo is the Gorm model that holds information about an app.
type AppInfo struct {
	gorm.Model

	AppID    string
	Name     string
	AuthType string
	AuthData string // JSON
}
