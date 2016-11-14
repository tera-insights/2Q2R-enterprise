// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"2q2r/security"
	"crypto"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
)

type adminHandler struct {
	s *Server
	q queue
}

// NewAdmin challenges the incoming admin, replying with a request ID that must
// be used in order to add a second-factor authentication mechanism. If the
// challenge signature is valid, then we store the admin.
// POST /admin/new
func (ah *adminHandler) NewAdmin(w http.ResponseWriter, r *http.Request) {
	req := NewAdminRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	encodedPermissions, err := json.Marshal(req.Permissions)
	optionalInternalPanic(err, "Could not encode permissions for storage")

	adminID, err := RandString(32)
	optionalInternalPanic(err, "Could not generate admin ID")

	keyID, err := RandString(32)
	optionalInternalPanic(err, "Could not generate key ID")

	err = ah.s.kc.VerifySignature(security.KeySignature{
		SigningPublicKey: req.SigningPublicKey,
		SignedPublicKey:  req.PublicKey,
		Type:             "signing",
		OwnerID:          adminID,
		Signature:        req.Signature,
	})
	optionalPanic(err, http.StatusForbidden,
		"Could not verify public key signature")

	err = ah.s.DB.Create(&Admin{
		ID:          adminID,
		Name:        req.Name,
		Email:       req.Email,
		Role:        "admin",
		Status:      "active",
		Permissions: string(encodedPermissions),
		AdminFor:    req.AdminFor,
	}).Error
	optionalInternalPanic(err, "Could not save admin")

	err = ah.s.DB.Create(&SigningKey{
		ID:        keyID,
		IV:        req.IV,
		Salt:      req.Salt,
		PublicKey: req.PublicKey,
	}).Error
	optionalInternalPanic(err, "Could not save signing key")

	requestID, err := RandString(32)
	optionalInternalPanic(err, "Could not generate request ID")

	h := crypto.SHA256.New()
	io.WriteString(h, requestID)
	query := ah.s.DB.Create(&LongTermRequest{
		AppID: "1",
		ID:    string(h.Sum(nil)),
	})
	optionalInternalPanic(query.Error, "Could not save long-term request "+
		"to the database")

	writeJSON(w, http.StatusOK, newAdminReply{
		RequestID: requestID,
	})
}

// GetAdmins lists all the admins.
// GET /admin/admin
func (ah *adminHandler) GetAdmins(w http.ResponseWriter, r *http.Request) {
	var result []Admin
	err := ah.s.DB.Model(&Admin{}).Find(&result).Error
	optionalBadRequestPanic(err, "Failed to read admins")

	writeJSON(w, http.StatusOK, result)
}

// UpdateAdmin updates an admin with a specific ID.
// PUT /admin/admin/{adminID}
func (ah *adminHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	req := adminUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	adminID := mux.Vars(r)["adminID"]
	err = CheckBase64(adminID)
	optionalBadRequestPanic(err, "Admin ID was not base-64 encoded")

	err = ah.s.DB.Where(&Admin{
		ID: adminID,
	}).Updates(Admin{
		Name:                req.Name,
		Email:               req.Email,
		PrimarySigningKeyID: req.PrimarySigningKeyID,
		AdminFor:            req.AdminFor,
	}).Error
	optionalInternalPanic(err, "Failed to update admin")

	var updated Admin
	err = ah.s.DB.First(&updated, &Admin{
		ID: adminID,
	}).Error
	optionalInternalPanic(err, "Failed to read updated admin")

	writeJSON(w, http.StatusOK, updated)
}

// DeleteAdmin deletes an admin that matches a query.
// DELETE /admin/admin/{adminID}
func (ah *adminHandler) DeleteAdmin(w http.ResponseWriter, r *http.Request) {
	adminID := mux.Vars(r)["adminID"]
	err := CheckBase64(adminID)
	optionalBadRequestPanic(err, "Admin ID was not base-64 encoded")

	query := ah.s.DB.Delete(Admin{}, &Admin{
		ID: adminID,
	})
	optionalInternalPanic(query.Error, "Failed to delete admins")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}

// ChangeAdminRoles can (de-)activate an admin or make the admin a super.
// POST /admin/admin/roles
func (ah *adminHandler) ChangeAdminRoles(w http.ResponseWriter, r *http.Request) {
	req := adminRoleChangeRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	err = ah.s.DB.Model(&Admin{}).Where(&Admin{
		ID: req.AdminID,
	}).Updates(Admin{
		Role:        req.Role,
		Status:      req.Status,
		Permissions: req.Permissions,
		AdminFor:    req.AdminFor,
	}).Error
	optionalInternalPanic(err, "Failed to change admin roles")

	var updated Admin
	err = ah.s.DB.First(&updated, &Admin{
		ID: req.AdminID,
	}).Error
	optionalInternalPanic(err, "Could not read updated admin")

	writeJSON(w, http.StatusOK, updated)
}

// GetApps gets all AppInfos.
// GET /admin/app
func (ah *adminHandler) GetApps(w http.ResponseWriter, r *http.Request) {
	var found []AppInfo
	err := ah.s.DB.Model(&AppInfo{}).Find(&found).Error
	optionalInternalPanic(err, "Could not read app infos")

	writeJSON(w, http.StatusOK, found)
}

// NewApp creates a new app.
// POST /admin/app
func (ah *adminHandler) NewApp(w http.ResponseWriter, r *http.Request) {
	req := newAppRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	appID, err := RandString(32)
	optionalInternalPanic(err, "Could not generate app ID")

	info := AppInfo{
		ID:      appID,
		AppName: req.AppName,
	}
	err = ah.s.DB.Create(&info).Error
	optionalInternalPanic(err, "Could not create app info")

	writeJSON(w, http.StatusOK, info)
}

// UpdateApp updates an app with a particular app ID.
// PUT /admin/app/{appID}
func (ah *adminHandler) UpdateApp(w http.ResponseWriter, r *http.Request) {
	req := appUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body as JSON")

	appID := mux.Vars(r)["appID"]
	err = CheckBase64(appID)
	optionalBadRequestPanic(err, "App ID was not base-64 encoded")

	// So that we don't overwrite the app name if there is no app name passed
	panicIfFalse(req.AppName != "", http.StatusBadRequest,
		"Cannot have an empty app name")

	err = ah.s.DB.Model(&AppInfo{}).Where(&AppInfo{
		ID: appID,
	}).Update(map[string]interface{}{
		gorm.ToDBName("AppName"): req.AppName,
	}).Error
	optionalInternalPanic(err, "Could not update app")

	var updated AppInfo
	err = ah.s.DB.Where(&updated, &AppInfo{
		ID: appID,
	}).Error
	optionalInternalPanic(err, "Could not read updated app")

	writeJSON(w, http.StatusOK, updated)
}

// DeleteApp deletes an app with a particular app ID.
// DELETE /admin/app/{appID}
func (ah *adminHandler) DeleteApp(w http.ResponseWriter, r *http.Request) {
	appID := mux.Vars(r)["appID"]
	err := CheckBase64(appID)
	optionalBadRequestPanic(err, "App ID was not base-64 encoded")

	query := ah.s.DB.Delete(AppInfo{}, &AppInfo{
		ID: appID,
	})
	optionalInternalPanic(query.Error, "Could not delete app")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}

// NewServer creates a new server for an admin with valid credentials.
// POST /admin/server
func (ah *adminHandler) NewServer(w http.ResponseWriter, r *http.Request) {
	req := newServerRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	serverID, err := RandString(32)
	optionalBadRequestPanic(err, "Could not generate server ID")

	info := AppServerInfo{
		ID:          serverID,
		BaseURL:     req.BaseURL,
		AppID:       req.AppID,
		KeyType:     req.KeyType,
		PublicKey:   []byte(req.PublicKey),
		Permissions: req.Permissions,
	}
	err = ah.s.DB.Create(&info).Error
	optionalInternalPanic(err, "Could not create app server")

	writeJSON(w, http.StatusOK, info)
}

// DeleteServer deletes a server on behalf of a valid admin.
// DELETE /admin/server/{serverID}
func (ah *adminHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	serverID := mux.Vars(r)["serverID"]
	err := CheckBase64(serverID)
	optionalBadRequestPanic(err, "Server ID was not base-64 encoded")

	err = ah.s.DB.Where(AppServerInfo{
		ID: serverID,
	}).Delete(AppServerInfo{}).Error
	optionalInternalPanic(err, "Could not delete app server")

	writeJSON(w, http.StatusOK, "Server deleted")
}

// GetServers gets information about app servers.
// GET /admin/server
func (ah *adminHandler) GetServers(w http.ResponseWriter, r *http.Request) {
	var info []AppServerInfo
	err := ah.s.DB.Model(&AppServerInfo{}).Find(&info).Error
	optionalBadRequestPanic(err, "Failed to find servers")

	writeJSON(w, http.StatusOK, info)
}

// UpdateServer updates an app server with `ServerID == req.ServerID`.
// PUT /admin/server/{serverID}
func (ah *adminHandler) UpdateServer(w http.ResponseWriter, r *http.Request) {
	req := serverUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body as JSON")

	serverID := mux.Vars(r)["serverID"]
	err = CheckBase64(serverID)
	optionalBadRequestPanic(err, "Server ID was not base-64 encoded")

	pub, err := decodeBase64(req.PublicKey)
	optionalBadRequestPanic(err, "Public key was not properly encoded")

	err = ah.s.DB.Where(&AppServerInfo{
		ID: serverID,
	}).Updates(AppServerInfo{
		BaseURL:     req.BaseURL,
		KeyType:     req.KeyType,
		PublicKey:   pub,
		Permissions: req.Permissions,
	}).Error
	optionalInternalPanic(err, "Failed to update app server info")

	var updated AppServerInfo
	err = ah.s.DB.First(&updated, &AppServerInfo{
		ID: serverID,
	}).Error
	optionalInternalPanic(err, "Failed to read updated app server info")

	writeJSON(w, http.StatusOK, updated)
}

// NewLongTerm stores a long-term request in the database.
// POST /admin/ltr
func (ah *adminHandler) NewLongTerm(w http.ResponseWriter, r *http.Request) {
	req := newLTRRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body as JSON")

	id, err := RandString(32)
	optionalInternalPanic(err, "Could not generate request ID")
	h := crypto.SHA256.New()
	io.WriteString(h, id)
	hashedID := string(h.Sum(nil))

	query := ah.s.DB.Create(&LongTermRequest{
		AppID: req.AppID,
		ID:    hashedID,
	})
	optionalInternalPanic(query.Error,
		"Could not save long-term request to the database")

	writeJSON(w, http.StatusOK, newLTRResponse{
		RequestID: id,
	})
}

// DeleteLongTerm deletes a long-term request from the database.
// DELETE /admin/ltr
func (ah *adminHandler) DeleteLongTerm(w http.ResponseWriter, r *http.Request) {
	req := deleteLTRRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body as JSON")

	query := ah.s.DB.Delete(LongTermRequest{}, &LongTermRequest{
		AppID: req.AppID,
		ID:    req.HashedRequestID,
	})
	optionalInternalPanic(query.Error, "Could not delete long-term request")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}

// GetSigningKeys returns all signing keys in the database.
// GET /admin/signing-key
func (ah *adminHandler) GetSigningKeys(w http.ResponseWriter, r *http.Request) {
	var result []SigningKey
	err := ah.s.DB.Model(&SigningKey{}).Find(&result).Error
	optionalInternalPanic(err, "Could not read signing keys")

	writeJSON(w, http.StatusOK, result)
}

// GetPermissions returns all permission in the DB.
// GET /admin/permission
func (ah *adminHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	var result []Permission
	err := ah.s.DB.Model(&Permission{}).Find(&result).Error
	optionalInternalPanic(err, "Could not read permissions")

	writeJSON(w, http.StatusOK, result)
}

// NewPermissions creates a list of new permissions.
// POST /admin/permission
func (ah *adminHandler) NewPermissions(w http.ResponseWriter,
	r *http.Request) {
	req := newPermissionsRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	tx := ah.s.DB.Begin()
	for _, p := range req.Permissions {
		err = ah.s.DB.Create(&p).Error
		if err != nil {
			tx.Rollback()
			optionalInternalPanic(err, "Could not save permission")
		}
	}

	err = tx.Commit().Error
	optionalInternalPanic(err, "Could not commit transaction to database")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: int64(len(req.Permissions)),
	})
}

// DeletePermission deletes a specified admin permission.
// DELETE /admin/permission/{appID}/{adminID}/{permission}
func (ah *adminHandler) DeletePermission(w http.ResponseWriter,
	r *http.Request) {
	appID := mux.Vars(r)["appID"]
	adminID := mux.Vars(r)["adminID"]
	permission := mux.Vars(r)["permission"]

	err := CheckBase64(appID)
	optionalBadRequestPanic(err, "App ID was not base-64 encoded")

	err = CheckBase64(adminID)
	optionalBadRequestPanic(err, "Admin ID was not base-64 encoded")

	err = CheckBase64(permission)
	optionalBadRequestPanic(err, "Permission was not base-64 encoded")

	query := ah.s.DB.Delete(Permission{}, &Permission{
		AppID:      appID,
		AdminID:    adminID,
		Permission: permission,
	})
	optionalInternalPanic(query.Error, "Could not delete permission")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}

// RegisterListener creates a new websocket-based stats listener from the
// request.
// GET /admin/stats/listen
func (ah *adminHandler) RegisterListener(w http.ResponseWriter,
	r *http.Request) {
	cookie, err := r.Cookie("admin-session")
	optionalPanic(err, http.StatusUnauthorized, "No session cookie")

	var m map[string]interface{}
	err = ah.s.sc.Decode("admin-session", cookie.Value, &m)
	optionalPanic(err, http.StatusUnauthorized, "Invalid session "+
		"cookie")

	val, found := m["app"]
	panicIfFalse(found, http.StatusBadRequest, "Invalid cookie")

	appID, ok := val.(string)
	panicIfFalse(ok, http.StatusBadRequest, "Invalid app ID in cookie")

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	optionalBadRequestPanic(err, "Could not upgrade request to a websocket")

	ah.s.disperser.addListener(listener{conn, appID})
	ah.s.disperser.addEvent(listenerRegistered, time.Now(), []string{appID})
	writeJSON(w, http.StatusOK, "Socket created")
}

func (ah *adminHandler) GetMostRecent(w http.ResponseWriter,
	r *http.Request) {
	writeJSON(w, http.StatusOK, ah.s.disperser.getRecent())
}
