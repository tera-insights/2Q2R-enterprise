// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// Base64Encode turns a string into its base-64 representation.
func Base64Encode(s string) string {
	bytes := []byte(s)
	return base64.StdEncoding.EncodeToString(bytes)
}

// AppInfoHandler returns information about the app specified by `appID`.
func AppInfoHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appID := mux.Vars(r)["appID"]
		_, err := base64.StdEncoding.DecodeString(appID)
		if err != nil {
			http.Error(w, "appID was not a valid base-64 string",
				http.StatusBadRequest)
		} else {
			var count = 0
			db.Model(&AppInfo{}).Where(AppInfo{AppID: appID}).Count(&count)
			if count > 0 {
				var info AppInfo
				db.Model(&AppInfo{}).Where(AppInfo{AppID: appID}).First(&info)
				reply := AppIDInfoReply{
					AppName:       info.AppName,
					BaseURL:       "example.com",
					AppID:         info.AppID,
					ServerPubKey:  "my_pub_key",
					ServerKeyType: "ECC-P256",
				}
				writeJSON(w, http.StatusOK, reply)
			} else {
				http.Error(w, "Could not find information for app with ID "+
					appID, http.StatusNotFound)
			}
		}
	}
}

// NewAppHandler creates a new app.
func NewAppHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := NewAppRequest{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			handleError(w, err)
		} else {
			appID := "123saWQgc3RyaW5nCg=="
			db.Create(&AppInfo{
				AppID:   appID,
				AppName: req.AppName,
			})
			writeJSON(w, http.StatusOK, NewAppReply{appID})
		}
	}
}

// NewServerHandler creates a new server for an admin with valid credentials.
// POST /v1/admin/server/new
func NewServerHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := NewServerRequest{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			handleError(w, err)
		} else {
			writeJSON(w, http.StatusOK, NewServerReply{})
		}
	}
}
