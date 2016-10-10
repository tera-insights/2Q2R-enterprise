// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Needed for Gorm
	cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/ryanuber/go-glob"
	"github.com/spf13/viper"
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

	HTTPS       bool
	LogRequests bool

	CertFile string
	KeyFile  string

	// A list of regular expressions mapping to routes that require auth
	// headers
	AuthenticationRequiredRoutes []string

	Base64EncodedPublicKey string
	KeyType                string

	// For HMAC-based authentication
	Token string
}

func (c *Config) getBaseURLWithProtocol() string {
	if c.HTTPS {
		return "https://" + c.BaseURL + c.Port
	}
	return "http://" + c.BaseURL + c.Port
}

// MakeConfig reads in r as if it were a config file of type ct and returns the
// resulting config.
func MakeConfig(r io.Reader, ct string) *Config {
	viper.SetConfigType(ct)

	viper.SetDefault("Port", ":8080")
	viper.SetDefault("DatabaseType", "sqlite3")
	viper.SetDefault("DatabaseName", "2q2r.db")
	viper.SetDefault("ExpirationTime", 1*time.Minute)
	viper.SetDefault("CleanTime", 30*time.Second)
	viper.SetDefault("ListenerExpirationTime", 3*time.Minute)
	viper.SetDefault("RecentlyCompletedExpirationTime", 5*time.Second)
	viper.SetDefault("BaseURL", "127.0.0.1")
	viper.SetDefault("HTTPS", true)
	viper.SetDefault("LogRequests", false)
	viper.SetDefault("AuthenticationRequiredRoutes", []string{
		"/*/register/request/*",
	})
	viper.SetDefault("Base64EncodedPublicKey", "mypubkey")
	viper.SetDefault("KeyType", "ECC-P256")
	viper.SetDefault("Token", "mytoken")

	err := viper.ReadConfig(r)
	if err != nil {
		log.Printf("Could not read config file! Using default options\n")
	}
	return &Config{
		Port:                            viper.GetString("Port"),
		DatabaseType:                    viper.GetString("DatabaseType"),
		DatabaseName:                    viper.GetString("DatabaseName"),
		ExpirationTime:                  viper.GetDuration("ExpirationTime"),
		CleanTime:                       viper.GetDuration("CleanTime"),
		ListenerExpirationTime:          viper.GetDuration("ListenerExpirationTime"),
		RecentlyCompletedExpirationTime: viper.GetDuration("RecentlyCompletedExpirationTime"),
		BaseURL:     viper.GetString("BaseURL"),
		HTTPS:       viper.GetBool("HTTPS"),
		LogRequests: viper.GetBool("LogRequests"),
		CertFile:    viper.GetString("CertFile"),
		KeyFile:     viper.GetString("KeyFile"),
		AuthenticationRequiredRoutes: viper.GetStringSlice("AuthenticationRequiredRoutes"),
		Base64EncodedPublicKey:       viper.GetString("Base64EncodedPublicKey"),
		KeyType:                      viper.GetString("KeyType"),
		Token:                        viper.GetString("Token"),
	}
}

// Server is the type that represents the 2Q2R server.
type Server struct {
	c     *Config
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
	RequestID   string   `json:"id"`
	KeyTypes    []string `json:"keyTypes"`
	Challenge   string   `json:"challenge"` // base-64 URL-encoded
	UserID      string   `json:"userID"`
	AppID       string   `json:"appId"`
	BaseURL     string   `json:"baseUrl"`
	InfoURL     string   `json:"infoUrl"`
	RegisterURL string   `json:"registerUrl"`
	WaitURL     string   `json:"waitUrl"`
	AppURL      string   `json:"appUrl"`
}

type keyDataToEmbed struct {
	KeyID string `json:"keyID"`
	Type  string `json:"type"`
	Name  string `json:"name"`
}

// Embedded in the templates
type authenticateData struct {
	RequestID    string           `json:"id"`
	Counter      int              `json:"counter"`
	Keys         []keyDataToEmbed `json:"keys"`
	Challenge    string           `json:"challenge"` // base-64 URL-encoded
	UserID       string           `json:"userID"`
	AppID        string           `json:"appId"`
	BaseURL      string           `json:"baseUrl"`
	AuthURL      string           `json:"authUrl"`
	InfoURL      string           `json:"infoUrl"`
	WaitURL      string           `json:"waitUrl"`
	ChallengeURL string           `json:"challengeUrl"`
	AppURL       string           `json:"appUrl"`
}

// NewServer creates a new 2Q2R server.
func NewServer(c *Config) Server {
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

// MakeDB returns the database specified by the configuration.
func MakeDB(c *Config) *gorm.DB {
	db, err := gorm.Open(c.DatabaseType, c.DatabaseName)
	db.AutoMigrate(&AppInfo{})
	db.AutoMigrate(&AppServerInfo{})
	db.AutoMigrate(&Key{})
	db.AutoMigrate(&Admin{})
	if err != nil {
		panic(errors.Errorf("Could not open database: %s", err))
	}
	return db
}

// MakeCacher returns the cacher specified by the configuration.
func MakeCacher(c *Config) Cacher {
	return Cacher{
		baseURL:                c.getBaseURLWithProtocol(),
		expiration:             c.ExpirationTime,
		clean:                  c.CleanTime,
		registrationRequests:   cache.New(c.ExpirationTime, c.CleanTime),
		authenticationRequests: cache.New(c.ExpirationTime, c.CleanTime),
		challengeToRequestID:   cache.New(c.ExpirationTime, c.CleanTime),
		admins:                 cache.New(c.ExpirationTime, c.CleanTime),
		adminRegistrations:     cache.New(c.ExpirationTime, c.CleanTime),
	}
}

func forMethod(r *mux.Router, s string, h http.HandlerFunc, m string) {
	r.PathPrefix(s).Methods(m).HandlerFunc(h)
}

type errorResponse struct {
	Message string
	Info    interface{} `json:",omitempty"`
}

func recoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				var statusCode int
				var response errorResponse
				if be, ok := err.(bubbledError); ok {
					statusCode = be.StatusCode
					response = errorResponse{
						Message: be.Message,
						Info:    be.Info,
					}
				} else {
					statusCode = http.StatusInternalServerError
					response = errorResponse{
						Message: "Internal server error",
					}
				}
				writingErr := writeJSON(w, statusCode, response)
				if writingErr != nil {
					log.Printf("Failed to encode error as JSON.\nEncoding error: "+
						"%v\nOriginal error:%v\n", writingErr, err)
				}
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func (srv *Server) middleware(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// See if URL is on one of the authentication-required routes
		needsAuthentication := false
		for _, regex := range srv.c.AuthenticationRequiredRoutes {
			if glob.Glob(regex, r.URL.Path) {
				needsAuthentication = true
				break
			}
		}

		if !needsAuthentication {
			handle.ServeHTTP(w, r)
			return
		}

		// Assert that we can validly parse the authentication headers
		serverID, messageMAC := getAuthDataFromHeaders(r)
		authParts := strings.Split(r.Header.Get("authentication"), ":")
		if len(authParts) != 2 {
			writeJSON(w, http.StatusBadRequest, "Authentication header invalid")
			return
		}

		// Determine which authentication mechanism to use
		serverInfo := AppServerInfo{}
		err := srv.DB.Model(AppServerInfo{}).Where(AppServerInfo{ServerID: serverID}).
			First(&serverInfo).Error
		optionalBadRequestPanic(err, "Could not find app server")

		panicIfFalse(serverInfo.AuthType == "token", http.StatusBadRequest,
			"Stored authentication method not supported")

		route := []byte(r.URL.Path)
		var body []byte
		if r.ContentLength > 0 {
			_, err = r.Body.Read(body)
			optionalInternalPanic(err, "Failed to read request body")
		}

		mac := hmac.New(sha256.New, []byte(srv.c.Token))
		mac.Write(route)
		if len(body) > 0 {
			mac.Write(body)
		}
		expectedMAC := mac.Sum(nil)
		bytesOfMessageMAC, err := base64.StdEncoding.DecodeString(messageMAC)
		optionalInternalPanic(err, "Failed to validate headers")

		if !hmac.Equal(bytesOfMessageMAC, expectedMAC) {
			writeJSON(w, http.StatusUnauthorized, "Invalid security headers")
			return
		}

		handle.ServeHTTP(w, r)
	})
}

// GetHandler returns the routes used by the 2Q2R server.
func (srv *Server) GetHandler() http.Handler {
	router := mux.NewRouter()

	// Admin routes
	ah := AdminHandler{
		s: srv,
		q: NewQueue(srv.c.RecentlyCompletedExpirationTime, srv.c.CleanTime,
			srv.c.ListenerExpirationTime, srv.c.CleanTime),
	}
	forMethod(router, "/v1/admin/register/{requestID}", ah.RegisterIFrameHandler, "GET")
	forMethod(router, "/v1/admin/register", ah.Register, "POST")
	forMethod(router, "/v1/admin/{requestID}/wait", ah.Wait, "GET")
	forMethod(router, "/v1/admin/new/{code}", ah.NewAdmin, "POST")

	forMethod(router, "/v1/admin/get", ah.GetAdmins, "GET")
	forMethod(router, "/v1/admin/update", ah.UpdateAdmin, "POST")
	forMethod(router, "/v1/admin/delete", ah.DeleteAdmin, "DELETE")   // super-admins only
	forMethod(router, "/v1/admin/roles", ah.ChangeAdminRoles, "POST") // super-admins only

	forMethod(router, "/v1/admin/app/new", ah.NewApp, "POST")
	forMethod(router, "/v1/admin/app/get", ah.GetApps, "GET")
	forMethod(router, "/v1/admin/app/update", ah.UpdateApp, "POST")
	forMethod(router, "/v1/admin/app/delete", ah.DeleteApp, "DELETE")

	forMethod(router, "/v1/admin/server/new", ah.NewServer, "POST")
	forMethod(router, "/v1/admin/server/get", ah.GetServers, "GET")
	forMethod(router, "/v1/admin/server/update", ah.UpdateServer, "POST")
	forMethod(router, "/v1/admin/server/delete", ah.DeleteServer, "DELETE")

	forMethod(router, "/v1/admin/ltr/new", ah.NewLongTerm, "POST")
	forMethod(router, "/v1/admin/ltr/delete", ah.DeleteLongTerm, "DELETE")

	// Info routes
	ih := InfoHandler{srv}
	forMethod(router, "/v1/info/{appID}", ih.AppInfoHandler, "GET")

	// Key routes
	kh := keyHandler{srv}
	forMethod(router, "/v1/users/{userID}", kh.UserExists, "GET")

	// Auth routes
	th := AuthHandler{
		s: srv,
		q: NewQueue(srv.c.RecentlyCompletedExpirationTime, srv.c.CleanTime,
			srv.c.ListenerExpirationTime, srv.c.CleanTime),
	}
	forMethod(router, "/v1/auth/request/{userID}", th.AuthRequestSetupHandler, "GET")
	forMethod(router, "/v1/auth/{requestID}/wait", th.Wait, "GET")
	forMethod(router, "/v1/auth/{requestID}/challenge", th.SetKey, "POST")
	forMethod(router, "/v1/auth", th.Authenticate, "POST")
	forMethod(router, "/auth/{requestID}", th.AuthIFrameHandler, "GET")

	// Register routes
	rh := RegisterHandler{
		s: srv,
		q: NewQueue(srv.c.RecentlyCompletedExpirationTime, srv.c.CleanTime,
			srv.c.ListenerExpirationTime, srv.c.CleanTime),
	}
	forMethod(router, "/v1/register/request/{userID}", rh.RegisterSetupHandler, "GET")
	forMethod(router, "/v1/register/{requestID}/wait", rh.Wait, "GET")
	forMethod(router, "/v1/register", rh.Register, "POST")
	forMethod(router, "/register/{requestID}", rh.RegisterIFrameHandler, "GET")

	// Static files
	fileServer := http.FileServer(rice.MustFindBox("assets").HTTPBox())
	router.PathPrefix("/").Handler(fileServer)
	h := recoverWrap(srv.middleware(router))
	if srv.c.LogRequests {
		return handlers.LoggingHandler(os.Stdout, h)
	}
	return h
}
