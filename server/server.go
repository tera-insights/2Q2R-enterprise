// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
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
	"github.com/telehash/gogotelehash/e3x/cipherset/cs1a/ecdh"
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

	// A list of glob patterns mapping to routes that require auth headers
	AuthenticationRequiredRoutes []string

	Base64EncodedPublicKey string
	KeyType                string

	// Path to the private ECDSA key that verifies server-to-server
	// authentication
	PrivateKeyFile string

	// Whether or nor the private ECDSA key file is encrypted
	PrivateKeyEncrypted bool
	PrivateKeyPassword  string
}

func (c *Config) getBaseURLWithProtocol() string {
	if c.HTTPS {
		return "https://" + c.BaseURL + c.Port
	}
	return "http://" + c.BaseURL + c.Port
}

// Server is the type that represents the 2Q2R server.
type Server struct {
	Config *Config
	DB     *gorm.DB
	cache  cacher
	pub    *rsa.PublicKey
	priv   *ecdsa.PrivateKey
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
func NewServer(r io.Reader, ct string) Server {
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
	viper.SetDefault("PrivateKeyEncrypted", false)

	err := viper.ReadConfig(r)
	if err != nil {
		log.Printf("Could not read config file! Using default options\n")
	}

	c := &Config{
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
		PrivateKeyFile:               viper.GetString("PrivateKeyFile"),
		PrivateKeyEncrypted:          viper.GetBool("PrivateKeyEncrypted"),
		PrivateKeyPassword:           viper.GetString("PrivateKeyPassword"),
	}

	// Load the Tera Insights RSA public key
	pubKey := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA16QwDL9Hyk1vKK2a8" +
		"wCmdiz/0da1ciRJ6z08jQxkfEzPVgrM+Vb8Qq/yS3tcLEA/VD+tucTzwzmZxbg5GvLz" +
		"ygyGoYuIVKhaCq598FCZlnqVHlOqa3b0Gg28I9CsJNXOntiYKff3d0KJ7v2HC2kZvL7" +
		"AnJkw7HxFv5bJCb3NPzfZJ3uLCKuWlG6lowG9pcoys7fogdJP8yrcQQarTQMDxPucY2" +
		"4HBvnP44mBzN3cBLg7sy6p7ZqBJbggrP6EQx2uwFyd5pW0INNW7wBx/wf/kEAQJEuBz" +
		"OKkBQWuR4q7aThFfKNyfklRZ0dgrRQegjMkMy5s9Bwe2cou45VzzA7rSQIDAQAB"
	block, _ := base64.StdEncoding.DecodeString(pubKey)
	pub, err := x509.ParsePKIXPublicKey(block)
	if err != nil {
		panic(errors.Wrap(err, "Failed to parse server's public key"))
	}

	// Read the elliptic private key
	file, err := os.Open(c.PrivateKeyFile)
	if err != nil {
		panic(errors.Wrap(err, "Couldn't open private key file"))
	}

	der := []byte{}
	_, err = file.Read(der)
	if err != nil {
		panic(errors.Wrap(err, "Couldn't read private key file"))
	}

	p, _ := pem.Decode(der)
	if p == nil {
		panic(errors.Wrap(err, "File was not PEM-formatted"))
	}

	var key []byte
	if c.PrivateKeyEncrypted {
		key, err = x509.DecryptPEMBlock(p, []byte(c.PrivateKeyPassword))
		if err != nil {
			panic(errors.Wrap(err, "Couldn't decrypt private key"))
		}
	} else {
		key = p.Bytes
	}

	priv, err := x509.ParseECPrivateKey(key)
	if err != nil {
		panic(errors.Wrap(err, "Couldn't parse file as DER-encoded ECDSA "+
			"private key"))
	}

	db, err := gorm.Open(c.DatabaseType, c.DatabaseName)
	if err != nil {
		panic(errors.Wrap(err, "Could not open database"))
	}

	err = db.AutoMigrate(&AppInfo{}).
		AutoMigrate(&AppServerInfo{}).
		AutoMigrate(&Key{}).
		AutoMigrate(&Admin{}).
		AutoMigrate(&KeySignature{}).
		AutoMigrate(&SigningKey{}).
		AutoMigrate(&Permission{}).Error
	if err != nil {
		panic(errors.Wrap(err, "Could not migrate schemas"))
	}

	return Server{
		c,
		db,
		cacher{
			baseURL:                c.getBaseURLWithProtocol(),
			expiration:             c.ExpirationTime,
			clean:                  c.CleanTime,
			registrationRequests:   cache.New(c.ExpirationTime, c.CleanTime),
			authenticationRequests: cache.New(c.ExpirationTime, c.CleanTime),
			challengeToRequestID:   cache.New(c.ExpirationTime, c.CleanTime),
			admins:                 cache.New(c.ExpirationTime, c.CleanTime),
			signingKeys:            cache.New(c.ExpirationTime, c.CleanTime),
			adminRegistrations:     cache.New(c.ExpirationTime, c.CleanTime),
			validPublicKeys: cache.New(cache.NoExpiration,
				cache.NoExpiration),
		},
		pub.(*rsa.PublicKey),
		priv,
	}
}

// Taken from https://git.io/v6xHB.
func writeJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
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
					log.Printf("Failed to encode error as JSON.\nEncoding "+
						"error: %v\nOriginal error:%v\n", writingErr, err)
				}
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// 1. If we don't need to authenticate the route, do nothing
// 2. Else, generate the ECDH key using our private key and the server's public
// key. If we can't find the server, return this is a bad request
// 3. Use the shared key to check the SHA-256 HMAC of the message. If it's OK,
// pass the request along. Else, refuse and log this unauthorized request.
func (s *Server) middleware(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// See if URL is on one of the authentication-required routes
		needsAuthentication := false
		for _, pattern := range s.Config.AuthenticationRequiredRoutes {
			if glob.Glob(pattern, r.URL.Path) {
				needsAuthentication = true
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
			writeJSON(w, http.StatusBadRequest,
				"Authentication header invalid")
			return
		}

		// If possible, get the shared key from the cache
		fetched, found := s.cache.sharedKeys.Get(serverID)
		var key []byte
		if found {
			key = fetched.([]byte)
		} else {
			// Use ECDH to generate the shared key
			serverInfo := AppServerInfo{}
			err := s.DB.Model(AppServerInfo{}).Where(AppServerInfo{
				ID: serverID,
			}).First(&serverInfo).Error
			optionalBadRequestPanic(err, "Could not find app server")

			x, y := elliptic.Unmarshal(elliptic.P256(), serverInfo.PublicKey)
			key = ecdh.ComputeShared(elliptic.P256(), x, y, s.priv.D.Bytes())
			s.cache.sharedKeys.Set(serverID, key, cache.NoExpiration)
		}

		// Hash the path and, if apprpriate, the request body
		route := []byte(r.URL.Path)
		var body []byte
		if r.ContentLength > 0 {
			_, err := r.Body.Read(body)
			optionalInternalPanic(err, "Failed to read request body")
		}
		mac := hmac.New(sha256.New, key)
		mac.Write(route)
		if len(body) > 0 {
			mac.Write(body)
		}
		expectedMAC := mac.Sum(nil)
		bytesOfMessageMAC, err := decodeBase64(messageMAC)
		optionalInternalPanic(err, "Failed to decode MAC as web-encoded "+
			"base-64 without padding")

		if !hmac.Equal(bytesOfMessageMAC, expectedMAC) {
			writeJSON(w, http.StatusUnauthorized, "Invalid security headers")
			return
		}

		handle.ServeHTTP(w, r)
	})
}

// GetHandler returns the routes used by the 2Q2R server.
func (s *Server) GetHandler() http.Handler {
	router := mux.NewRouter()

	// Get the server's public key
	forMethod(router, "/v1/public", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.priv.PublicKey)
	}, "GET")

	// Admin routes
	ah := adminHandler{
		s: s,
		q: newQueue(s.Config.RecentlyCompletedExpirationTime,
			s.Config.CleanTime, s.Config.ListenerExpirationTime,
			s.Config.CleanTime),
	}
	forMethod(router, "/admin/new", ah.NewAdmin, "POST")
	forMethod(router, "/admin/register/{requestID}", ah.RegisterIFrameHandler,
		"GET")
	forMethod(router, "/admin/register", ah.Register, "POST")
	forMethod(router, "/admin/{requestID}/wait", ah.Wait, "GET")

	forMethod(router, "/admin/admin", ah.GetAdmins, "GET")
	// super-admins only
	forMethod(router, "/admin/admin/roles", ah.ChangeAdminRoles, "POST")
	forMethod(router, "/admin/admin/{adminID}", ah.UpdateAdmin, "PUT")
	// super-admins only
	forMethod(router, "/admin/admin/{adminID}", ah.DeleteAdmin, "DELETE")

	forMethod(router, "/admin/app", ah.GetApps, "GET")
	forMethod(router, "/admin/app", ah.NewApp, "POST")
	forMethod(router, "/admin/app/{appID}", ah.UpdateApp, "POST")
	forMethod(router, "/admin/app/{appID}", ah.DeleteApp, "DELETE")

	forMethod(router, "/admin/server", ah.GetServers, "GET")
	forMethod(router, "/admin/server", ah.NewServer, "POST")
	forMethod(router, "/admin/server/{serverID}", ah.UpdateServer, "PUT")
	forMethod(router, "/admin/server/{serverID}", ah.DeleteServer, "DELETE")

	forMethod(router, "/admin/signing-key", ah.GetSigningKeys, "GET")

	forMethod(router, "/admin/ltr", ah.NewLongTerm, "POST")
	forMethod(router, "/admin/ltr", ah.DeleteLongTerm, "DELETE")

	// Super-admins only
	forMethod(router, "/admin/permission", ah.GetPermissions, "GET")
	forMethod(router, "/admin/permission", ah.NewPermissions, "POST")
	forMethod(router, "/admin/permission/{appID}/{adminID}/{permission}",
		ah.DeletePermission, "DELETE")

	// Info routes
	ih := infoHandler{s}
	forMethod(router, "/v1/info/{appID}", ih.AppinfoHandler, "GET")

	// Key routes
	kh := keyHandler{s}
	forMethod(router, "/v1/users/{userID}", kh.UserExists, "GET")
	forMethod(router, "/v1/keys/get", kh.GetKeys, "GET")
	forMethod(router, "/v1/users/{userID}", kh.DeleteUser, "DELETE")
	forMethod(router, "/v1/keys/{userID}/{keyHandle}", kh.DeleteKey, "DELETE")

	// Auth routes
	th := authHandler{
		s: s,
		q: newQueue(s.Config.RecentlyCompletedExpirationTime,
			s.Config.CleanTime, s.Config.ListenerExpirationTime,
			s.Config.CleanTime),
	}
	forMethod(router, "/v1/auth/request/{userID}", th.AuthRequestSetupHandler,
		"GET")
	forMethod(router, "/v1/auth/{requestID}/wait", th.Wait, "GET")
	forMethod(router, "/v1/auth/{requestID}/challenge", th.SetKey, "POST")
	forMethod(router, "/v1/auth", th.Authenticate, "POST")
	forMethod(router, "/v1/auth/{requestID}", th.AuthIFrameHandler, "GET")

	// Register routes
	rh := registerHandler{
		s: s,
		q: newQueue(s.Config.RecentlyCompletedExpirationTime,
			s.Config.CleanTime, s.Config.ListenerExpirationTime,
			s.Config.CleanTime),
	}
	forMethod(router, "/v1/register/request/{userID}", rh.RegisterSetupHandler,
		"GET")
	forMethod(router, "/v1/register/{requestID}/wait", rh.Wait, "GET")
	forMethod(router, "/v1/register", rh.Register, "POST")
	forMethod(router, "/v1/register/{requestID}", rh.RegisterIFrameHandler,
		"GET")

	// Static files
	fileServer := http.FileServer(rice.MustFindBox("assets").HTTPBox())
	router.PathPrefix("/").Handler(fileServer)
	h := recoverWrap(s.middleware(router))
	if s.Config.LogRequests {
		return handlers.LoggingHandler(os.Stdout, h)
	}
	return h
}
