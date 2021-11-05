// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net"
	"net/http"
	"time"

	"github.com/tera-insights/2Q2R-enterprise/security"
	"github.com/tera-insights/2Q2R-enterprise/util"

	"github.com/gorilla/mux"
)

type keyHandler struct {
	s *Server
}

// UserExists checks whether there exists a user with the passed ID.
// GET /v1/users/{userID}
func (kh *keyHandler) UserExists(w http.ResponseWriter, r *http.Request) {
	serverID, _, err := getAuthDataFromHeaders(r)
	util.OptionalInternalPanic(err, "Could not decode authentication headers")

	var asi AppServerInfo
	err = kh.s.DB.Model(AppServerInfo{}).Where(AppServerInfo{ID: serverID}).
		First(&asi).Error
	util.OptionalBadRequestPanic(err, "Could not find server")

	query := security.Key{AppID: asi.AppID, UserID: mux.Vars(r)["userID"]}
	count := 0
	err = kh.s.DB.Model(security.Key{}).Where(query).Count(&count).Error
	util.OptionalInternalPanic(err, "Could not find key")

	writeJSON(w, http.StatusOK, userExistsReply{count > 0})
}

// DeleteUser deletes all the keys for a particular user ID. Note that first it
// removes them from the cache.
// DELETE /v1/users/{userID}
func (kh *keyHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	util.PanicIfFalse(userID != "", http.StatusBadRequest, "User ID cannot be \"\"")

	// kh.s.kc.Remove2FAKey()
	var keys []security.Key
	err := kh.s.DB.Find(&keys, &security.Key{
		UserID: userID,
	}).Error
	util.OptionalInternalPanic(err, "Could not lookup keys to delete")

	for _, key := range keys {
		kh.s.kc.Remove2FAKey(key.ID)
	}

	query := kh.s.DB.Delete(security.Key{}, &security.Key{
		UserID: userID,
	})
	util.OptionalInternalPanic(query.Error, "Could not delete keys from database")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}

// GetKeys lists all the keys in the database.
// GET /v1/keys/get
func (kh *keyHandler) GetKeys(w http.ResponseWriter, r *http.Request) {
	var result []security.Key
	query := kh.s.DB.Model(&security.Key{}).Find(&result)
	util.OptionalInternalPanic(query.Error, "Could not read keys from database")

	writeJSON(w, http.StatusOK, result)
}

// DeleteKey deletes a key that matches a particular query.
// DELETE /v1/keys/{userID}/{keyHandle}
func (kh *keyHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	util.PanicIfFalse(userID != "", http.StatusBadRequest, "User ID cannot be \"\"")

	keyHandle := mux.Vars(r)["keyHandle"]
	util.PanicIfFalse(keyHandle != "", http.StatusBadRequest, "Key handle cannot be \"\"")

	kh.s.kc.Remove2FAKey(keyHandle)

	var k security.Key
	query := kh.s.DB.First(&k, security.Key{
		UserID: userID,
		ID:     keyHandle,
	}).Delete(security.Key{}, &security.Key{
		UserID: userID,
		ID:     keyHandle,
	})

	util.OptionalInternalPanic(query.Error, "Could not delete key")

	// If the key deletion was for an admin, then look up the admin's app ID
	// and log the event under that key
	var appID string
	if k.AppID == "1" {
		var a Admin
		err := kh.s.DB.First(&a, Admin{
			ID: k.UserID,
		}).Error
		util.OptionalBadRequestPanic(err, "Could not find admin")

		appID = a.AdminFor
	} else {
		appID = k.AppID
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)

	kh.s.disperser.addEvent(keyDeletion, time.Now(), appID, "success",
		userID, host, host)
	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}
