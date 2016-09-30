// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"net/http"
)

// AdminHandler is the handler for all registration requests.
type AdminHandler struct {
	s *Server
}

// NewAppHandler creates a new app.
// POst /v1/admin/app/new
func (ah *AdminHandler) NewAppHandler(w http.ResponseWriter, r *http.Request) {
	req := NewAppRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	appID, err := randString(32)
	optionalInternalPanic(err, "Could not generate app ID")

	err = ah.s.DB.Create(&AppInfo{
		AppID:   appID,
		AppName: req.AppName,
	}).Error
	optionalInternalPanic(err, "Could not create app info")

	writeJSON(w, http.StatusOK, NewAppReply{appID})
}

// NewServerHandler creates a new server for an admin with valid credentials.
// POST /v1/admin/server/new
func (ah *AdminHandler) NewServerHandler(w http.ResponseWriter, r *http.Request) {
	req := NewServerRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	serverID, err := randString(32)
	optionalBadRequestPanic(err, "Could not generate server ID")

	err = ah.s.DB.Create(&AppServerInfo{
		ServerID:    serverID,
		ServerName:  req.ServerName,
		BaseURL:     req.BaseURL,
		AppID:       req.AppID,
		KeyType:     req.KeyType,
		PublicKey:   []byte(req.PublicKey),
		Permissions: req.Permissions,
	}).Error
	optionalInternalPanic(err, "Could not create app server")

	writeJSON(w, http.StatusOK, NewServerReply{
		ServerName: req.ServerName,
		ServerID:   serverID,
	})
}

// DeleteServerHandler deletes a server on behalf of a valid admin.
// POST /v1/admin/server/delete
func (ah *AdminHandler) DeleteServerHandler(w http.ResponseWriter, r *http.Request) {
	req := DeleteServerRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	err = CheckBase64(req.ServerID)
	optionalBadRequestPanic(err, "Server ID was not base-64 encoded")

	err = ah.s.DB.Where(AppServerInfo{
		ServerID: req.ServerID,
	}).Delete(AppServerInfo{}).Error
	optionalInternalPanic(err, "Could not delete app server")

	writeJSON(w, http.StatusOK, "Server deleted")
}

// GetServerHandler gets information about a server with a particular ID.
// POST /v1/admin/server/get
func (ah *AdminHandler) GetServerHandler(w http.ResponseWriter, r *http.Request) {
	req := AppServerInfoRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	err = CheckBase64(req.ServerID)
	optionalBadRequestPanic(err, "Server ID was not base-64 encoded")

	g := DBHandler{DB: ah.s.DB.Model(&AppServerInfo{}), Writer: w}
	var info AppServerInfo
	g.FirstWhereWithRespond(AppServerInfo{ServerID: req.ServerID}, &info)
}

// NewUserHandler creates a new user for an admin with valid credentials.
// POST /v1/admin/user/new
func (ah *AdminHandler) NewUserHandler(w http.ResponseWriter, r *http.Request) {
	req := NewUserRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	userID, err := randString(32)
	optionalInternalPanic(err, "Could not generate user ID")

	err = ah.s.DB.Create(&Key{
		UserID: userID,
	}).Error
	optionalInternalPanic(err, "Could not create key for new user")

	writeJSON(w, http.StatusOK, NewUserReply{
		UserID: userID,
	})
}
