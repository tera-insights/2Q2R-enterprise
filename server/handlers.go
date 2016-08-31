// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// AppInfoHandler returns information about the app specified by `appID`.
func AppInfoHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appID := mux.Vars(r)["appID"]
		if isNotBase64String(appID) {
			http.Error(w, "appID was not a valid base-64 string", http.StatusBadRequest)
		} else {
			var count = 0
			db.Model(&AppInfo{}).Where("app_id = ?", appID).Count(&count)
			if count > 0 {
				var info AppInfo
				db.Model(&AppInfo{}).Where(AppInfo{AppID: appID}).First(&info)
				reply := AppIDInfoReply{
					AppName:       info.Name,
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
			return
		}
		appID := "123"
		db.Create(&AppInfo{
			AppID:    appID,
			Name:     req.AppName,
			AuthType: req.AuthType,
			AuthData: req.AuthData,
		})
		if err != nil {
			handleError(w, err)
		} else {
			writeJSON(w, http.StatusOK, NewAppReply{appID})
		}
	}
}
