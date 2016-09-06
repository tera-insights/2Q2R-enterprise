// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/rand"
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

// CheckBase64 returns any errors encountered when deserializing a
// (supposedly) base-64 encoded string.
func CheckBase64(s string) error {
	_, err := base64.StdEncoding.DecodeString(s)
	return err
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

// AppInfoHandler returns information about the app specified by `appID`.
func AppInfoHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appID := mux.Vars(r)["appID"]
		if CheckBase64(appID) != nil {
			http.Error(w, "appID was not a valid base-64 string",
				http.StatusBadRequest)
		} else {
			t := DBHandler{DB: db.Model(&AppInfo{})}
			query := AppInfo{AppID: appID}
			count := t.CountWhere(query)
			if count > 0 {
				var info AppInfo
				t.FirstWhere(AppInfo{AppID: appID}, &info)
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
			appID := randString(32)
			if CheckBase64(appID) != nil {
				http.Error(w, "Could not generate app ID",
					http.StatusInternalServerError)
			} else {
				db.Create(&AppInfo{
					AppID:   appID,
					AppName: req.AppName,
				})
				writeJSON(w, http.StatusOK, NewAppReply{appID})
			}
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
			serverID := randString(32)
			if CheckBase64(serverID) != nil {
				http.Error(w, "Server ID was not base-64 encoded", http.StatusBadRequest)
			} else {
				db.Create(&AppServerInfo{
					ServerID:    serverID,
					ServerName:  req.ServerName,
					BaseURL:     req.BaseURL,
					AppID:       req.AppID,
					KeyType:     req.KeyType,
					PublicKey:   req.PublicKey,
					Permissions: req.Permissions,
				})
				writeJSON(w, http.StatusOK, NewServerReply{
					ServerName: req.ServerName,
					ServerID:   serverID,
				})
			}
		}
	}
}

// DeleteServerHandler deletes a server on behalf of a valid admin.
func DeleteServerHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := DeleteServerRequest{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			handleError(w, err)
		} else {
			if CheckBase64(req.ServerID) != nil {
				http.Error(w, "Server ID was not base-64 encoded", http.StatusBadRequest)
			} else {
				query := AppServerInfo{
					ServerID: req.ServerID,
				}
				db.Where(query).Delete(AppServerInfo{})
				writeJSON(w, http.StatusOK, "Server deleted")
			}
		}
	}
}

// GetServerHandler gets information about a server with a particular ID.
func GetServerHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := AppServerInfoRequest{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			handleError(w, err)
		} else {
			if CheckBase64(req.ServerID) != nil {
				http.Error(w, "Server ID was not base-64 encoded", http.StatusBadRequest)
			} else {
				g := DBHandler{DB: db.Model(&AppServerInfo{}), Writer: w}
				var info AppServerInfo
				g.FirstWhereWithRespond(AppServerInfo{ServerID: req.ServerID}, &info)
			}
		}
	}
}
