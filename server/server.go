// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Needed for Gorm
	cache "github.com/patrickmn/go-cache"
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
}

// MakeConfig reads in r as if it were a config file of type ct and returns the
// resulting config.
func MakeConfig(r io.Reader, ct string) (Config, error) {
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
		"/*/register*",
	})
	viper.SetDefault("Base64EncodedPublicKey", "mypubkey")
	viper.SetDefault("KeyType", "ECC-P256")

	err := viper.ReadConfig(r)
	return Config{
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
	}, err
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

// NewServer creates a new 2Q2R server.
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
	db.AutoMigrate(&Key{})
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

		serverID, messageMAC := getAuthDataFromHeaders(r)

		// Determine which authentication mechanism to use
		serverInfo := AppServerInfo{}
		err := srv.DB.Model(AppServerInfo{}).Where(AppServerInfo{ServerID: serverID}).
			First(&serverInfo).Error
		if err != nil {
			handleError(w, err)
			return
		}

		if serverInfo.AuthType != "token" {
			writeJSON(w, http.StatusBadRequest, "Stored authentication method not supported")
			return
		}
		route := []byte(r.URL.Path)
		var body []byte
		_, err = r.Body.Read(body)
		if err != nil {
			handleError(w, err)
			return
		}
		mac := hmac.New(sha256.New, serverInfo.PublicKey)
		mac.Write(route)
		if len(body) > 0 {
			mac.Write(body)
		}
		expectedMAC := mac.Sum(nil)
		bytesOfMessageMAC, err := base64.StdEncoding.DecodeString(messageMAC)
		if err != nil {
			handleError(w, err)
			return
		}

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
	ah := AdminHandler{srv}
	forMethod(router, "/v1/admin/app/new", ah.NewAppHandler, "POST")
	forMethod(router, "/v1/admin/server/new", ah.NewServerHandler, "POST")
	forMethod(router, "/v1/admin/server/delete", ah.DeleteServerHandler, "POST")
	forMethod(router, "/v1/admin/server/get", ah.GetServerHandler, "POST")
	forMethod(router, "/v1/admin/user/new", ah.NewUserHandler, "POST")

	// Info routes
	ih := InfoHandler{srv}
	forMethod(router, "/v1/info/{appID}", ih.AppInfoHandler, "GET")

	// Key routes
	kh := keyHandler{srv}
	forMethod(router, "/v1/users/{userID}", kh.UserExists, "GET")

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
	forMethod(router, "/v1/register/request/{userID}", rh.RegisterSetupHandler, "GET")
	forMethod(router, "/v1/register", rh.Register, "POST")
	forMethod(router, "/v1/register/{requestID}/wait", rh.Wait, "GET")
	forMethod(router, "/register/{requestID}", rh.RegisterIFrameHandler, "GET")

	if srv.c.LogRequests {
		return handlers.LoggingHandler(os.Stdout, srv.middleware(router))
	}
	return srv.middleware(router)
}
