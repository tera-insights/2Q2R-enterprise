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

// CheckBase64 returns any errors encountered when deserializing a
// (supposedly) base-64 encoded string.
func CheckBase64(s string) error {
	_, err := base64.StdEncoding.DecodeString(s)
	return err
}

type genericGormTable struct {
	DB     *gorm.DB
	Writer http.ResponseWriter
}

func (g *genericGormTable) CountWhere(q interface{}) int {
	c := 0
	g.DB.Where(q).Count(&c)
	return c
}

func (g *genericGormTable) FirstWhere(q interface{}, r interface{}) {
	g.DB.Where(q).First(r)
}

func (g *genericGormTable) FirstWhereWithRespond(q interface{}, r interface{}) {
	count := g.CountWhere(q)
	if count > 0 {
		g.FirstWhere(q, r)
		writeJSON(g.Writer, http.StatusOK, r)
	} else {
		http.Error(g.Writer, "Could not find resource", http.StatusNotFound)
	}
}

// AppInfoHandler returns information about the app specified by `appID`.
func AppInfoHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appID := mux.Vars(r)["appID"]
		if CheckBase64(appID) != nil {
			http.Error(w, "appID was not a valid base-64 string",
				http.StatusBadRequest)
		} else {
			t := genericGormTable{DB: db.Model(&AppInfo{})}
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
			appID := "123saWQgc3RyaW5nCg=="
			if CheckBase64(appID) != nil {
				http.Error(w, "App ID was not base-64 encoded", http.StatusBadRequest)
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
			serverID := "777saJMgc7RyaW5nCg=="
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
			serverID := "777saJMgc7RyaW5nCg=="
			if CheckBase64(serverID) != nil {
				http.Error(w, "Server ID was not base-64 encoded", http.StatusBadRequest)
			} else {
				query := AppServerInfo{
					ServerID: serverID,
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
				g := genericGormTable{DB: db.Model(&AppServerInfo{}), Writer: w}
				var info AppServerInfo
				g.FirstWhereWithRespond(AppServerInfo{ServerID: req.ServerID}, &info)
			}
		}
	}
}
