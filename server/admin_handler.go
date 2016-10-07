// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

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
	base := ah.s.c.getBaseURLWithProtocol() + "/v1/admin/"
	writeJSON(w, http.StatusOK, newAdminReply{
		RequestID:   requestID,
		IFrameRoute: base + "register/" + requestID,
		WaitRoute:   base + requestID + "/wait",
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

// RegisterIFrameHandler returns the iFrame used for the admin to register.
// GET /v1/admin/register/{requestID}
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
		RegisterURL: base + "/v1/admin/register",
		WaitURL:     base + "/v1/admin/" + requestID + "/wait",
	})
	optionalInternalPanic(err, "Failed to generate template")

	t.Execute(w, templateData{
		Name: "Registration",
		ID:   "register",
		Data: template.JS(data),
	})
}

// Register registers a new second-factor for an admin.
// POST /v1/admin/register
func (ah *AdminHandler) Register(w http.ResponseWriter, r *http.Request) {
	req := adminRegisterRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Failed to decode request body")

	data, found := ah.s.cache.admins.Get(req.RequestID)
	panicIfFalse(found, http.StatusBadRequest,
		"Failed to find admin for registration request")

	admin, ok := data.(Admin)
	panicIfFalse(ok, http.StatusInternalServerError,
		"Failed to load admin for registration request")

	data, found = ah.s.cache.adminRegistrations.Get(req.RequestID)
	panicIfFalse(found, http.StatusBadRequest,
		"Failed to find stored registration request")

	registerRequest, ok := data.(AdminRegistrationRequest)
	panicIfFalse(ok, http.StatusInternalServerError,
		"Failed to load registration request")

	x, y := elliptic.Unmarshal(elliptic.P256(), admin.PublicKey)
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

// NewApp creates, well, a new app.
// POST /v1/admin/app/new
func (ah *AdminHandler) NewApp(w http.ResponseWriter, r *http.Request) {
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

// GetApp gets an app with a particular app ID.
// GET /v1/admin/app/get/{appID}
func (ah *AdminHandler) GetApp(w http.ResponseWriter, r *http.Request) {
	appID := mux.Vars(r)["appID"]
	err := CheckBase64(appID)
	optionalBadRequestPanic(err, "App ID was not base-64")

	var found []AppInfo
	err = ah.s.DB.Model(&AppInfo{}).Find(&found, &AppInfo{AppID: appID}).Error
	optionalInternalPanic(err, "Could not read app infos")
	panicIfFalse(len(found) == 0, http.StatusBadRequest,
		fmt.Sprintf("Could not find app with id %s", appID))

	writeJSON(w, http.StatusOK, found[0])
}

// UpdateApp updates an app with a particular app ID.
// POST /v1/admin/app/update
func (ah *AdminHandler) UpdateApp(w http.ResponseWriter, r *http.Request) {
	req := appUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body as JSON")

	// So that we don't overwrite the app name if there is no app name passed
	panicIfFalse(req.AppName != "", http.StatusBadRequest,
		"Cannot have an empty app name")

	query := ah.s.DB.Model(&AppInfo{}).Where(&AppInfo{
		AppID: req.AppID,
	}).Update(map[string]interface{}{
		gorm.ToDBName("AppName"): req.AppName,
	})
	optionalInternalPanic(query.Error, "Could not update app")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}

// DeleteApp deletes an app with a particular app ID.
// DELETE /v1/admin/app/delete
func (ah *AdminHandler) DeleteApp(w http.ResponseWriter, r *http.Request) {
	req := appDeleteRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	optionalBadRequestPanic(err, "Could not decode request body as JSON")

	query := ah.s.DB.Delete(AppInfo{}, &AppInfo{
		AppID: req.AppID,
	})
	optionalInternalPanic(query.Error, "Could not delete app")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
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
// DELETE /v1/admin/server/delete
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
// GET /v1/admin/server/get/{serverID}
func (ah *AdminHandler) GetServerHandler(w http.ResponseWriter, r *http.Request) {
	serverID := mux.Vars(r)["serverID"]
	err := CheckBase64(serverID)
	optionalBadRequestPanic(err, "Server ID was not base-64 encoded")

	var info AppServerInfo
	err = ah.s.DB.Model(&AppServerInfo{}).Where(&AppServerInfo{ServerID: serverID}).First(&info).Error
	optionalBadRequestPanic(err, "Failed to find server with specified ID")

	writeJSON(w, http.StatusOK, info)
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
