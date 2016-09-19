// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package main

import (
	"2q2r/server"
	"os"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	appID := "_T-zi0wzr7GCi4vsfsXsUuKOfmiWLiHBVbmJJPidvhA"

	f, err := os.Open("./config.yaml")
	handle(err)

	config, err := server.MakeConfig(f, "yaml")
	handle(err)

	db := server.MakeDB(config)

	err = db.Create(&server.AppInfo{
		AppID:   appID,
		AppName: "test-app",
	}).Error
	handle(err)

	err = db.Create(&server.AppServerInfo{
		ServerID: "_T-zi0wzr7GCi4vsfsXsUuKOfmiWLiHBVbmJJPidvhA",
		BaseURL:  config.BaseURL,
		AppID:    appID,
		AuthType: "token",
	}).Error
	handle(err)
}
