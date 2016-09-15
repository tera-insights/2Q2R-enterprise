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
	Port         string
	DatabaseType string
	DatabaseName string

	// How long until auth/register requests are considered invalid (will be
	// cleaned on the next sweep)
	ExpirationTime time.Duration

	// How frequently the auth/register caches are cleaned
	CleanTime time.Duration

	ListenerExpirationTime          time.Duration
	RecentlyCompletedExpirationTime time.Duration

	// BaseURL for this 2Q-2R Server
	BaseURL string
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
type registerData struct {
	RequestID string   `json:"id"`
	KeyTypes  []string `json:"keyTypes"`
	Challenge []byte   `json:"challenge"`
	UserID    string   `json:"userID"`
	AppID     string   `json:"appId"`
	InfoURL   string   `json:"infoUrl"`
	WaitURL   string   `json:"waitUrl"`
}

// Embedded in the templates
type authenticateData struct {
	RequestID    string   `json:"id"`
	Counter      int      `json:"counter"`
	Keys         []string `json:"keys"`
	Challenge    []byte   `json:"challenge"`
	UserID       string   `json:"userID"`
	AppID        string   `json:"appId"`
	InfoURL      string   `json:"infoUrl"`
	WaitURL      string   `json:"waitUrl"`
	ChallengeURL string   `json:"challengeUrl"`
}

// New creates a new 2Q2R server.
func NewServer(c Config) Server {
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
		Message: err.Error(),
	}
	if serr, ok := err.(StatusError); ok {
		statusCode = serr.StatusCode()
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
	db.AutoMigrate(&User{})
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
		challengeToRequestID:   cache.New(c.ExpirationTime, c.CleanTime),
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

	// Admin routes
	ah := AdminHandler{srv}
	forMethod(router, "/v1/admin/app/new", ah.NewAppHandler, "POST")
	forMethod(router, "/v1/admin/server/new", ah.NewServerHandler, "POST")
	forMethod(router, "/v1/admin/server/delete", ah.DeleteServerHandler, "POST")
	forMethod(router, "/v1/admin/server/get", ah.GetServerHandler, "POST")
	forMethod(router, "/v1/admin/user/new", ah.NewUserHandler, "POST")

	// Info routes
	ih := InfoHandler{srv}
	forMethod(router, "/v1/info/{appID}", ih.AppInfoHandler, "GET")

	// Auth routes
	th := AuthHandler{srv}
	forMethod(router, "/v1/auth/request", th.AuthRequestSetupHandler, "POST")
	forMethod(router, "/auth/{requestID}", th.AuthIFrameHandler, "GET")

	// Register routes
	rh := RegisterHandler{
		s: srv,
		q: NewQueue(srv.c.RecentlyCompletedExpirationTime, srv.c.CleanTime,
			srv.c.ListenerExpirationTime, srv.c.CleanTime),
	}
	forMethod(router, "/v1/register/request", rh.RegisterSetupHandler, "POST")
	forMethod(router, "/v1/register", rh.Register, "POST")
	forMethod(router, "/v1/register/{requestID}/wait", rh.Wait, "GET")
	forMethod(router, "/register/{requestID}", rh.RegisterIFrameHandler, "GET")

	return router
}
