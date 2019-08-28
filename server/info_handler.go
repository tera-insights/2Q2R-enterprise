// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"github.com/alinVD/2Q2R-enterprise/util"
	"net/http"

	"github.com/gorilla/mux"
)

type infoHandler struct {
	s *Server
}

// AppinfoHandler returns information about the app specified by `appID`.
// GET /v1/info/:appID
func (ih *infoHandler) AppinfoHandler(w http.ResponseWriter, r *http.Request) {
	appID := mux.Vars(r)["appID"]
	err := util.CheckBase64(appID)
	util.OptionalBadRequestPanic(err, "App ID was not a valid base-64 string")

	query := AppInfo{ID: appID}
	count := 0
	err = ih.s.DB.Model(&AppInfo{}).Where(&query).Count(&count).Error
	util.OptionalInternalPanic(err, "Failed to count apps inside database")

	if count > 0 {
		var info AppInfo
		err = ih.s.DB.Model(&AppInfo{}).Where(&query).First(&info).Error
		reply := appIDInfoReply{
			AppName:   info.AppName,
			BaseURL:   ih.s.Config.getBaseURLWithProtocol(),
			AppURL:    ih.s.Config.getBaseURLWithProtocol(),
			AppID:     info.ID,
			PublicKey: ih.s.Config.Base64EncodedPublicKey,
			KeyType:   ih.s.Config.KeyType,
		}
		writeJSON(w, http.StatusOK, reply)
		return
	}
	http.Error(w, "Could not find information for app with ID "+appID,
		http.StatusNotFound)
}
