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
	if err != nil {
		handleError(w, err)
		return
	}
	appID := randString(32)
	if CheckBase64(appID) != nil {
		http.Error(w, "Could not generate app ID", http.StatusInternalServerError)
		return
	}
	err = ah.s.DB.Create(&AppInfo{
		AppID:   appID,
		AppName: req.AppName,
	}).Error
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, NewAppReply{appID})
}

// NewServerHandler creates a new server for an admin with valid credentials.
// POST /v1/admin/server/new
func (ah *AdminHandler) NewServerHandler(w http.ResponseWriter, r *http.Request) {
	req := NewServerRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		handleError(w, err)
		return
	}
	serverID := randString(32)
	if CheckBase64(serverID) != nil {
		http.Error(w, "Server ID was not base-64 encoded",
			http.StatusBadRequest)
		return
	}
	ah.s.DB.Create(&AppServerInfo{
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

// DeleteServerHandler deletes a server on behalf of a valid admin.
// POST /v1/admin/server/delete
func (ah *AdminHandler) DeleteServerHandler(w http.ResponseWriter, r *http.Request) {
	req := DeleteServerRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		handleError(w, err)
		return
	}
	if CheckBase64(req.ServerID) != nil {
		http.Error(w, "Server ID was not base-64 encoded", http.StatusBadRequest)
		return
	}
	query := AppServerInfo{
		ServerID: req.ServerID,
	}
	err = ah.s.DB.Where(query).Delete(AppServerInfo{}).Error
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, "Server deleted")
}

// GetServerHandler gets information about a server with a particular ID.
// POST /v1/admin/server/get
func (ah *AdminHandler) GetServerHandler(w http.ResponseWriter, r *http.Request) {
	req := AppServerInfoRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		handleError(w, err)
		return
	}
	if CheckBase64(req.ServerID) != nil {
		http.Error(w, "Server ID was not base-64 encoded", http.StatusBadRequest)
		return
	}
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
	if err != nil {
		handleError(w, err)
		return
	}
	userID := randString(32)
	if CheckBase64(userID) != nil {
		http.Error(w, "User ID was not base-64 encoded",
			http.StatusInternalServerError)
		return
	}
	ah.s.DB.Create(&User{
		UserID: userID,
	})
	writeJSON(w, http.StatusOK, NewUserReply{
		UserID: userID,
	})
}
