// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Needed for Gorm
	cache "github.com/patrickmn/go-cache"
)

// Config is the configuration for the server.
type Config struct {
	Port         int
	DatabaseType string
	DatabaseName string

	// How long until auth/register requests are considered invalid (will be
	// cleaned on the next sweep)
	ExpirationTime time.Duration

	// How frequently the auth/register caches are cleaned
	CleanTime time.Duration

	// In bytes
	ChallengeLength int
}

// Server is the type that represents the 2Q2R server.
type Server struct {
	c     Config
	DB    *gorm.DB
	cache Cacher
}

// Used in registration and authentication templates
type templateData struct {
	Name string
	ID   string
	Data template.JS
}

// Embedded in the templates
type authData struct {
	ID      string `json:"id"`
	Counter int    `json:"counter"`
}

// New creates a new 2Q2R server.
func New(c Config) Server {
	var s = Server{c, MakeDB(c), MakeCacher(c)}
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

// MakeCacher returns the cacher specified by the configuration.
func MakeCacher(c Config) Cacher {
	return Cacher{
		expiration:             c.ExpirationTime,
		clean:                  c.CleanTime,
		registrationRequests:   cache.New(c.ExpirationTime, c.CleanTime),
		authenticationRequests: cache.New(c.ExpirationTime, c.CleanTime),
	}
}

// HandleInvalidMethod returns a function that says that the requested method
// was not allowed.
func HandleInvalidMethod() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleError(w, MethodNotAllowedError(r.Method))
	}
}

func forMethod(r *mux.Router, s string, h http.HandlerFunc, m string) {
	r.HandleFunc(s, h).Methods(m)
	r.HandleFunc(s, HandleInvalidMethod())
}

// GetHandler returns the routes used by the 2Q2R server.
func (srv *Server) GetHandler() http.Handler {
	router := mux.NewRouter()

	forMethod(router, "/v1/info/{appID}", AppInfoHandler(srv.DB), "GET")
	forMethod(router, "/v1/app/new", NewAppHandler(srv.DB), "POST")
	forMethod(router, "/v1/admin/server/new", NewServerHandler(srv.DB), "POST")
	forMethod(router, "/v1/admin/server/delete", DeleteServerHandler(srv.DB), "POST")
	forMethod(router, "/v1/admin/server/get", GetServerHandler(srv.DB), "POST")

	// Auth routes
	ah := AuthHandler{srv}
	forMethod(router, "/v1/auth/request", ah.AuthRequestSetupHandler, "POST")

	// Register routes
	rh := RegisterHandler{srv}
	forMethod(router, "/v1/register/request", rh.RegisterSetupHandler, "POST")
	forMethod(router, "/register/{requestID}", rh.RegisterIFrameHandler, "GET")

	return router
}
