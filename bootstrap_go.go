// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package main

import (
	"github.com/tera-insights/2Q2R-enterprise/server"

	"github.com/jinzhu/gorm"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
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

	err = db.Create(&server.AppServerInfo{
		ID:      "_T-zi0wzr7GCi4vsfsXsUuKOfmiWLiHBVbmJJPidvhA",
		BaseURL: "localhost:8080",
		AppID:   "_T-zi0wzr7GCi4vsfsXsUuKOfmiWLiHBVbmJJPidvhA",
	}).Error
	handle(err)
}
