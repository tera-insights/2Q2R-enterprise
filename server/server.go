// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Needed for Gorm
)

// Config is the configuration for the server.
type Config struct {
	Port         int
	DatabaseType string
	DatabaseName string
}

// Server is the type that represents the 2Q2R server.
type Server struct {
	c  Config
	DB *gorm.DB
}

// New creates a new 2Q2R server.
func New(c Config) Server {
	var s = Server{c, MakeDB(c)}
	return s
}

// Taken from https://git.io/v6xHB.
func writeJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}

// Taken from https://git.io/viJaE.
func handleError(w http.ResponseWriter, err error) {
	var statusCode = http.StatusInternalServerError
	var response = errorResponse{
		ErrorCode: "unknown",
		Message:   err.Error(),
	}
	if serr, ok := err.(StatusError); ok {
		statusCode = serr.StatusCode()
		response.ErrorCode = serr.ErrorCode()
		response.Info = serr.Info()
	}
	writingErr := writeJSON(w, statusCode, response)
	if writingErr != nil {
		log.Printf("Failed to encode error as JSON.\nEncoding error: "+
			"%v\nOriginal error:%v\n", writingErr, err)
	}
}

// MakeDB returns the database specified by the configuration.
func MakeDB(c Config) *gorm.DB {
	db, err := gorm.Open(c.DatabaseType, c.DatabaseName)
	db.AutoMigrate(&AppInfo{})
	if err != nil {
		panic(fmt.Errorf("Could not open database: %s", err))
	}
	return db
}

// GetHandler returns the routes used by the 2Q2R server.
func (srv *Server) GetHandler() http.Handler {
	router := mux.NewRouter()

	// Get app info
	router.HandleFunc("/v1/info/{appID}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			appID := mux.Vars(r)["appID"]
			var count = 0
			srv.DB.Model(&AppInfo{}).Where("app_id = ?", appID).Count(&count)
			if count > 0 {
				var info AppInfo
				srv.DB.Model(&AppInfo{}).Where(AppInfo{AppID: appID}).First(&info)
				reply := AppIDInfoReply{
					AppName:       info.Name,
					BaseURL:       "example.com",
					AppID:         info.AppID,
					ServerPubKey:  "my_pub_key",
					ServerKeyType: "ECC-P256",
				}
				writeJSON(w, http.StatusOK, reply)
			} else {
				http.Error(w, "Could not find information for app with ID "+
					appID, http.StatusNotFound)
			}
		default:
			handleError(w, MethodNotAllowedError(r.Method))
		}
	})

	// Upload app info
	router.HandleFunc("/v1/app/new", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			req := NewAppRequest{}
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&req)
			if err != nil {
				handleError(w, err)
				return
			}
			appID := "123"
			srv.DB.Create(&AppInfo{
				AppID:    appID,
				Name:     req.AppName,
				AuthType: req.AuthType,
				AuthData: req.AuthData,
			})
			if err != nil {
				handleError(w, err)
			} else {
				writeJSON(w, http.StatusOK, NewAppReply{appID})
			}
		default:
			handleError(w, MethodNotAllowedError(r.Method))
		}
	})

	return router
}
