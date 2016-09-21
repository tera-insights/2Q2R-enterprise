// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// InfoHandler is the handler for all `/*/info/*` requests.
type InfoHandler struct {
	s *Server
}

// AppInfoHandler returns information about the app specified by `appID`.
// GET /v1/info/:appID
func (ih *InfoHandler) AppInfoHandler(w http.ResponseWriter, r *http.Request) {
	appID := mux.Vars(r)["appID"]
	if CheckBase64(appID) != nil {
		http.Error(w, "appID was not a valid base-64 string",
			http.StatusBadRequest)
		return
	}
	t := DBHandler{DB: ih.s.DB.Model(&AppInfo{})}
	query := AppInfo{AppID: appID}
	count := t.CountWhere(query)
	if count > 0 {
		var info AppInfo
		t.FirstWhere(query, &info)
		reply := AppIDInfoReply{
			AppName:   info.AppName,
			BaseURL:   ih.s.c.getBaseURLWithProtocol(),
			AppID:     info.AppID,
			PublicKey: ih.s.c.Base64EncodedPublicKey,
			KeyType:   ih.s.c.KeyType,
		}
		writeJSON(w, http.StatusOK, reply)
		return
	}
	http.Error(w, "Could not find information for app with ID "+appID,
		http.StatusNotFound)
}
