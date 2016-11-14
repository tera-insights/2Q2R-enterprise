// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type keyHandler struct {
	s *Server
}

// UserExists checks whether there exists a user with the passed ID.
// GET /v1/users/{userID}
func (kh *keyHandler) UserExists(w http.ResponseWriter, r *http.Request) {
	serverID, _, err := getAuthDataFromHeaders(r)
	optionalInternalPanic(err, "Could not decode authentication headers")

	var asi AppServerInfo
	err = kh.s.DB.Model(AppServerInfo{}).Where(AppServerInfo{ID: serverID}).
		First(&asi).Error
	optionalBadRequestPanic(err, "Could not find server")

	query := Key{AppID: asi.AppID, UserID: mux.Vars(r)["userID"]}
	count := 0
	err = kh.s.DB.Model(Key{}).Where(query).Count(&count).Error
	optionalInternalPanic(err, "Could not find key")

	writeJSON(w, http.StatusOK, userExistsReply{count > 0})
}

// DeleteUser deletes all the keys for a particular user ID.
// DELETE /v1/users/{userID}
func (kh *keyHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	panicIfFalse(userID != "", http.StatusBadRequest, "User ID cannot be \"\"")

	query := kh.s.DB.Delete(Key{}, &Key{
		UserID: userID,
	})
	optionalInternalPanic(query.Error, "Could not delete keys from database")

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}

// GetKeys lists all the keys in the database.
// GET /v1/keys/get
func (kh *keyHandler) GetKeys(w http.ResponseWriter, r *http.Request) {
	var result []Key
	query := kh.s.DB.Model(&Key{}).Find(&result)
	optionalInternalPanic(query.Error, "Could not read keys from database")

	writeJSON(w, http.StatusOK, result)
}

// DeleteKey deletes a key that matches a particular query.
// DELETE /v1/keys/{userID}/{keyHandle}
func (kh *keyHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	panicIfFalse(userID != "", http.StatusBadRequest, "User ID cannot be \"\"")

	keyHandle := mux.Vars(r)["keyHandle"]
	panicIfFalse(keyHandle != "", http.StatusBadRequest, "Key handle cannot be \"\"")

	var k Key
	query := kh.s.DB.First(&k, Key{
		UserID: userID,
		ID:     keyHandle,
	}).Delete(Key{}, &Key{
		UserID: userID,
		ID:     keyHandle,
	})

	optionalInternalPanic(query.Error, "Could not delete key")

	// If the key deletion was for an admin, then look up the admin's app ID
	// and log the event under that key
	var appID string
	if k.AppID == "1" {
		var a Admin
		err := kh.s.DB.First(&a, Admin{
			ID: k.UserID,
		}).Error
		optionalBadRequestPanic(err, "Could not find admin")

		appID = a.AdminFor
	} else {
		appID = k.AppID
	}
	kh.s.disperser.addEvent(keyDeletion, time.Now(), []string{appID})

	writeJSON(w, http.StatusOK, modificationReply{
		NumAffected: query.RowsAffected,
	})
}
