// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// AdminHandler is the handler for all registration requests.
type AdminHandler struct {
	s *Server
	q Queue
}

// NewAdmin creates a new admin. If code == "bootstrap", attempts to bootstrap
// the system. Bootstrapping only works if there are no keys and no admins
// interface in the system.
// Replies with a request ID that must be used in order to add a second-factor
// authentication mechanism.
// POST /v1/admin/new/{code}
func (ah *AdminHandler) NewAdmin(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	panicIfFalse(code == "bootstrap", http.StatusBadRequest,
		"Unrecognized activation code")

	req := newAdminRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	adminCount := 0
	err = ah.s.DB.Model(&Admin{}).Count(&adminCount).Error
	optionalInternalPanic(err, "Failed to count the admins in the database")
	panicIfFalse(adminCount == 0, http.StatusBadRequest,
		"There are already admins in the database")

	keyCount := 0
	err = ah.s.DB.Model(&Key{}).Count(&keyCount).Error
	optionalInternalPanic(err, "Failed to count the keys in the database")
	panicIfFalse(adminCount == 0, http.StatusBadRequest,
		"There are already keys in the database")

	requestID, err := randString(32)
	optionalInternalPanic(err, "Could not generate request ID")

	ah.s.cache.NewAdminRegisterRequest(requestID, Admin{
		AdminID:     req.AdminID,
		Name:        req.Name,
		Email:       req.Email,
		Active:      code == "bootstrap",
		SuperAdmin:  code == "bootstrap",
		Permissions: strings.Join(req.Permissions, ","),
		IV:          req.IV,
		Seed:        req.Seed,
		PublicKey:   req.PublicKey,
	})
	writeJSON(w, http.StatusOK, newAdminReply{
		RequestID: requestID,
		Route:     ah.s.c.getBaseURLWithProtocol() + "/admin/register",
	})
}

// Wait waits for the admin to authenticate a particular request. If the
// authentication is successful, writes the admin and key to the database.
// GET /v1/admin/{requestID}/wait
func (ah *AdminHandler) Wait(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	c := ah.q.Listen(requestID)
	response := <-c
	if response != http.StatusOK {
		w.WriteHeader(response)
		return
	}

	admin, err := ah.s.cache.GetAdmin(requestID)
	optionalInternalPanic(err, "Failed to find admin to save to database")

	err = ah.s.DB.Model(&Admin{}).Create(&admin).Error
	optionalInternalPanic(err, "Failed to save admin to database")

	keyID, err := randString(32)
	optionalInternalPanic(err, "Failed to generate key ID")
	key := Key{
		KeyID:                  keyID,
		UserID:                 admin.AdminID,
		AppID:                  "1", // special app for admins
		Counter:                0,
		MarshalledRegistration: admin.PublicKey, // this is bad, will change
	}
	err = ah.s.DB.Model(&Key{}).Create(&key).Error
	optionalInternalPanic(err, "Failed to save key to database")

	w.WriteHeader(response)
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
