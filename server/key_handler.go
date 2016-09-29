// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type keyHandler struct {
	s *Server
}

// UserExists checks whether there exists a user with the passed ID.
// `GET /v1/users/:userID`
func (kh keyHandler) UserExists(w http.ResponseWriter, r *http.Request) {
	serverID, _ := getAuthDataFromHeaders(r)
	var asi AppServerInfo
	err := kh.s.DB.Model(AppServerInfo{}).Where(AppServerInfo{ServerID: serverID}).
		First(&asi).Error
	optionalBadRequestPanic(err, "Could not find server")

	query := Key{AppID: asi.AppID, UserID: mux.Vars(r)["userID"]}
	count := 0
	err = kh.s.DB.Model(Key{}).Where(query).Count(&count).Error
	optionalInternalPanic(err, "Could not find key")

	writeJSON(w, http.StatusOK, userExistsReply{count > 0})
}
