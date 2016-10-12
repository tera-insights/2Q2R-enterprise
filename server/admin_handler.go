// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"html/template"
	"io"
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
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
// POST /admin/new/{code}
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

	encodedPermissions, err := json.Marshal(req.Permissions)
	optionalInternalPanic(err, "Could not encode permissions for storage")

	role := "admin"
	status := "inactive"
	if code == "bootstrap" {
		role = "superadmin"
		status = "active"
	}

	adminID, err := randString(32)
	optionalInternalPanic(err, "Could not generate admin ID")

	keyID, err := randString(32)
	optionalInternalPanic(err, "Could not generate key ID")

	ah.s.cache.NewAdminRegisterRequest(requestID, Admin{
		AdminID:     adminID,
		Name:        req.Name,
		Email:       req.Email,
		Role:        role,
		Status:      status,
		Permissions: string(encodedPermissions),
	}, SigningKey{
		SigningKeyID: keyID,
		IV:           req.IV,
		Salt:         req.Salt,
		PublicKey:    req.PublicKey,
	})
	base := ah.s.c.getBaseURLWithProtocol() + "/admin/"
	writeJSON(w, http.StatusOK, newAdminReply{
		RequestID:   requestID,
		IFrameRoute: base + "register/" + requestID,
		WaitRoute:   base + requestID + "/wait",
	})
}

// Wait waits for the admin to authenticate a particular request. If the
// authentication is successful, writes the admin and key to the database.
// GET /admin/{requestID}/wait
func (ah *AdminHandler) Wait(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	c := ah.q.Listen(requestID)
	response := <-c
	if response != http.StatusOK {
		w.WriteHeader(response)
		return
	}

	admin, signingKey, err := ah.s.cache.GetAdmin(requestID)
	optionalInternalPanic(err, "Failed to find admin to save to database")

	err = ah.s.DB.Model(&Admin{}).Create(&admin).Error
	optionalInternalPanic(err, "Failed to save admin to database")

	optionalInternalPanic(err, "Failed to generate key ID")
	err = ah.s.DB.Model(&SigningKey{}).Create(&signingKey).Error
	optionalInternalPanic(err, "Failed to save key to database")

	w.WriteHeader(response)
}

// RegisterIFrameHandler returns the iFrame used for the admin to register.
// GET /admin/register/{requestID}
func (ah *AdminHandler) RegisterIFrameHandler(w http.ResponseWriter, r *http.Request) {
	requestID := mux.Vars(r)["requestID"]
	templateBox, err := rice.FindBox("assets")
	optionalInternalPanic(err, "Failed to load assets")

	templateString, err := templateBox.String("all.html")
	optionalInternalPanic(err, "Failed to load template")

	t, err := template.New("register").Parse(templateString)
	optionalInternalPanic(err, "Failed to generate registration iFrame")

	cachedData, found := ah.s.cache.adminRegistrations.Get(requestID)
	panicIfFalse(found, http.StatusBadRequest, "Failed to find registration request")
	request := cachedData.(AdminRegistrationRequest)

	cachedData, found = ah.s.cache.admins.Get(requestID)
	panicIfFalse(found, http.StatusInternalServerError, "Failed to find cached admin")
	admin := cachedData.(Admin)

	base := ah.s.c.getBaseURLWithProtocol()
	data, err := json.Marshal(registerData{
		RequestID:   requestID,
		KeyTypes:    []string{"2q2r", "u2f"},
		Challenge:   encodeBase64(request.Challenge),
		UserID:      admin.AdminID,
		BaseURL:     base,
		AppURL:      base,
		RegisterURL: base + "/admin/register",
		WaitURL:     base + "/admin/" + requestID + "/wait",
	})
	optionalInternalPanic(err, "Failed to generate template")

	t.Execute(w, templateData{
		Name: "Registration",
		ID:   "register",
		Data: template.JS(data),
	})
}

// Register registers a new second-factor for an admin.
// POST /admin/register
func (ah *AdminHandler) Register(w http.ResponseWriter, r *http.Request) {
	req := adminRegisterRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Failed to decode request body")

	_, signingKey, err := ah.s.cache.GetAdmin(req.RequestID)
	optionalBadRequestPanic(err, "Failed to find signing key for registration request")

	data, found := ah.s.cache.adminRegistrations.Get(req.RequestID)
	panicIfFalse(found, http.StatusBadRequest,
		"Failed to find stored registration request")

	registerRequest, ok := data.(AdminRegistrationRequest)
	panicIfFalse(ok, http.StatusInternalServerError,
		"Failed to load registration request")

	x, y := elliptic.Unmarshal(elliptic.P256(), signingKey.PublicKey)
	if x == nil {
		panic(bubbledError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Could not unmarshal stored public key for admin",
		})
	}

	pubKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	hash := sha256.Sum256(registerRequest.Challenge)
	verified := ecdsa.Verify(&pubKey, hash[:], &req.R, &req.S)
	panicIfFalse(verified, http.StatusBadRequest, "Failed to verify signature")

	ah.q.MarkCompleted(req.RequestID)
	writeJSON(w, http.StatusOK, RegisterResponse{
		Successful: true,
		Message:    "OK",
	})
}

// GetAdmins lists all the admins.
// GET /admin/admin
func (ah *AdminHandler) GetAdmins(w http.ResponseWriter, r *http.Request) {
	var result []Admin
	err := ah.s.DB.Model(&Admin{}).Find(&result).Error
	optionalBadRequestPanic(err, "Failed to read admins")

	writeJSON(w, http.StatusOK, result)
}

// UpdateAdmin updates an admin with a specific ID.
// PUT /admin/admin/{adminID}
func (ah *AdminHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	req := adminUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	adminID := mux.Vars(r)["adminID"]
	err = CheckBase64(adminID)
	optionalBadRequestPanic(err, "Admin ID was not base-64 encoded")

	err = ah.s.DB.Where(&Admin{
		AdminID: adminID,
	}).Updates(Admin{
		Name:                req.Name,
		Email:               req.Email,
		PrimarySigningKeyID: req.PrimarySigningKeyID,
	}).Error
	optionalInternalPanic(err, "Failed to update admin")

	var updated Admin
	err = ah.s.DB.First(&updated, &Admin{
		AdminID: adminID,
	}).Error
	optionalInternalPanic(err, "Failed to read updated admin")

	writeJSON(w, http.StatusOK, updated)
}

// DeleteAdmin deletes an admin that matches a query.
// DELETE /admin/admin/{adminID}
func (ah *AdminHandler) DeleteAdmin(w http.ResponseWriter, r *http.Request) {
	adminID := mux.Vars(r)["adminID"]
	err := CheckBase64(adminID)
	optionalBadRequestPanic(err, "Admin ID was not base-64 encoded")

	query := ah.s.DB.Delete(Admin{}, &Admin{
		AdminID: adminID,
	})
	optionalInternalPanic(query.Error, "Failed to delete admins")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}

// ChangeAdminRoles can (de-)activate an admin or make the admin a super.
// POST /admin/admin/roles
func (ah *AdminHandler) ChangeAdminRoles(w http.ResponseWriter, r *http.Request) {
	req := adminRoleChangeRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	err = ah.s.DB.Model(&Admin{}).Where(&Admin{
		AdminID: req.AdminID,
	}).Updates(Admin{
		Role:        req.Role,
		Status:      req.Status,
		Permissions: req.Permissions,
	}).Error
	optionalInternalPanic(err, "Failed to change admin roles")

	var updated Admin
	err = ah.s.DB.First(&updated, &Admin{
		AdminID: req.AdminID,
	}).Error
	optionalInternalPanic(err, "Could not read updated admin")

	writeJSON(w, http.StatusOK, updated)
}

// GetApps gets all AppInfos.
// GET /admin/app
func (ah *AdminHandler) GetApps(w http.ResponseWriter, r *http.Request) {
	var found []AppInfo
	err := ah.s.DB.Model(&AppInfo{}).Find(&found).Error
	optionalInternalPanic(err, "Could not read app infos")

	writeJSON(w, http.StatusOK, found)
}

// NewApp creates a new app.
// POST /admin/app
func (ah *AdminHandler) NewApp(w http.ResponseWriter, r *http.Request) {
	req := NewAppRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	appID, err := randString(32)
	optionalInternalPanic(err, "Could not generate app ID")

	info := AppInfo{
		AppID:   appID,
		AppName: req.AppName,
	}
	err = ah.s.DB.Create(&info).Error
	optionalInternalPanic(err, "Could not create app info")

	writeJSON(w, http.StatusOK, info)
}

// UpdateApp updates an app with a particular app ID.
// PUT /admin/app/{appID}
func (ah *AdminHandler) UpdateApp(w http.ResponseWriter, r *http.Request) {
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
		AppID: appID,
	}).Update(map[string]interface{}{
		gorm.ToDBName("AppName"): req.AppName,
	}).Error
	optionalInternalPanic(err, "Could not update app")

	var updated AppInfo
	err = ah.s.DB.Where(&updated, &AppInfo{
		AppID: appID,
	}).Error
	optionalInternalPanic(err, "Could not read updated app")

	writeJSON(w, http.StatusOK, updated)
}

// DeleteApp deletes an app with a particular app ID.
// DELETE /admin/app/{appID}
func (ah *AdminHandler) DeleteApp(w http.ResponseWriter, r *http.Request) {
	appID := mux.Vars(r)["appID"]
	err := CheckBase64(appID)
	optionalBadRequestPanic(err, "App ID was not base-64 encoded")

	query := ah.s.DB.Delete(AppInfo{}, &AppInfo{
		AppID: appID,
	})
	optionalInternalPanic(query.Error, "Could not delete app")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}

// NewServer creates a new server for an admin with valid credentials.
// POST /admin/server
func (ah *AdminHandler) NewServer(w http.ResponseWriter, r *http.Request) {
	req := NewServerRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body")

	serverID, err := randString(32)
	optionalBadRequestPanic(err, "Could not generate server ID")

	info := AppServerInfo{
		ServerID:    serverID,
		ServerName:  req.ServerName,
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
func (ah *AdminHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	serverID := mux.Vars(r)["serverID"]
	err := CheckBase64(serverID)
	optionalBadRequestPanic(err, "Server ID was not base-64 encoded")

	err = ah.s.DB.Where(AppServerInfo{
		ServerID: serverID,
	}).Delete(AppServerInfo{}).Error
	optionalInternalPanic(err, "Could not delete app server")

	writeJSON(w, http.StatusOK, "Server deleted")
}

// GetServers gets information about app servers.
// GET /admin/server
func (ah *AdminHandler) GetServers(w http.ResponseWriter, r *http.Request) {
	var info []AppServerInfo
	err := ah.s.DB.Model(&AppServerInfo{}).Find(&info).Error
	optionalBadRequestPanic(err, "Failed to find servers")

	writeJSON(w, http.StatusOK, info)
}

// UpdateServer updates an app server with `ServerID == req.ServerID`.
// PUT /admin/server/{serverID}
func (ah *AdminHandler) UpdateServer(w http.ResponseWriter, r *http.Request) {
	req := serverUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body as JSON")

	serverID := mux.Vars(r)["serverID"]
	err = CheckBase64(serverID)
	optionalBadRequestPanic(err, "Server ID was not base-64 encoded")

	err = ah.s.DB.Where(&AppServerInfo{
		ServerID: serverID,
	}).Updates(AppServerInfo{
		ServerName:  req.ServerName,
		BaseURL:     req.BaseURL,
		KeyType:     req.KeyType,
		PublicKey:   req.PublicKey,
		Permissions: req.Permissions,
		AuthType:    req.AuthType,
	}).Error
	optionalInternalPanic(err, "Failed to update app server info")

	var updated AppServerInfo
	err = ah.s.DB.First(&updated, &AppServerInfo{
		ServerID: serverID,
	}).Error
	optionalInternalPanic(err, "Failed to read updated app server info")

	writeJSON(w, http.StatusOK, updated)
}

// NewLongTerm stores a long-term request in the database.
// POST /admin/ltr
func (ah *AdminHandler) NewLongTerm(w http.ResponseWriter, r *http.Request) {
	req := newLTRRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body as JSON")

	id, err := randString(32)
	optionalInternalPanic(err, "Could not generate request ID")
	h := crypto.SHA256.New()
	io.WriteString(h, id)
	hashedID := string(h.Sum(nil))

	query := ah.s.DB.Create(&LongTermRequest{
		AppID:           req.AppID,
		HashedRequestID: hashedID,
	})
	optionalInternalPanic(query.Error,
		"Could not save long-term request to the database")

	writeJSON(w, http.StatusOK, newLTRResponse{
		RequestID: id,
	})
}

// DeleteLongTerm deletes a long-term request from the database.
// DELETE /admin/ltr
func (ah *AdminHandler) DeleteLongTerm(w http.ResponseWriter, r *http.Request) {
	req := deleteLTRRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body as JSON")

	query := ah.s.DB.Delete(LongTermRequest{}, &LongTermRequest{
		AppID:           req.AppID,
		HashedRequestID: req.HashedRequestID,
	})
	optionalInternalPanic(query.Error, "Could not delete long-term request")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}
