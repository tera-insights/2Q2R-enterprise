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
	db.AutoMigrate(&AppServerInfo{})
	if err != nil {
		panic(fmt.Errorf("Could not open database: %s", err))
	}
	return db
}

// HandleInvalidMethod returns a function that says that the requested method
// was not allowed.
func HandleInvalidMethod() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleError(w, MethodNotAllowedError(r.Method))
	}
}

// GetHandler returns the routes used by the 2Q2R server.
func (srv *Server) GetHandler() http.Handler {
	router := mux.NewRouter()

	// GET /v1/info/{appID}
	router.HandleFunc("/v1/info/{appID}", AppInfoHandler(srv.DB)).Methods("GET")
	router.HandleFunc("/v1/info/{appID}", HandleInvalidMethod())

	// POST /v1/app/new
	router.HandleFunc("/v1/app/new", NewAppHandler(srv.DB)).Methods("POST")
	router.HandleFunc("/v1/app/new", HandleInvalidMethod())

	// POST /v1/admin/server/new
	router.HandleFunc("/v1/admin/server/new", NewServerHandler(srv.DB)).Methods("POST")
	router.HandleFunc("/v1/admin/server/new", HandleInvalidMethod())

	return router
}
