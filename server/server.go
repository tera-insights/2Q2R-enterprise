// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"2q2r/security"
	"2q2r/util"
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
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Needed for Gorm
	_ "github.com/jinzhu/gorm/dialects/sqlite"   // Needed for Gorm
	"github.com/pkg/errors"
	glob "github.com/ryanuber/go-glob"
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

	Base64EncodedPublicKey string
	KeyType                string

	// Path to the private ECDSA key that verifies server-to-server
	// authentication
	PrivateKeyFile string

	// Whether or nor the private ECDSA key file is encrypted
	PrivateKeyEncrypted bool
	PrivateKeyPassword  string

	AdminSessionLength time.Duration
	MaxMindPath        string
}

func (c *Config) getBaseURLWithProtocol() string {
	if c.HTTPS {
		return "https://" + c.BaseURL + c.Port
	}
	return "http://" + c.BaseURL + c.Port
}

// Server is the type that represents the 2Q2R server.
type Server struct {
	Config    *Config
	DB        *gorm.DB
	disperser *disperser
	Pub       *rsa.PublicKey
	priv      *ecdsa.PrivateKey
	sc        *securecookie.SecureCookie
	kc        *security.KeyCache
	kg        *security.KeyGen
}

// Used in registration and authentication templates
type templateData struct {
	Name string
	ID   string
	Data template.JS
}

// NewServer creates a new 2Q2R server.
func NewServer(r io.Reader, ct string) (s Server) {
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
	viper.SetDefault("Base64EncodedPublicKey", "mypubkey")
	viper.SetDefault("KeyType", "ECC-P256")
	viper.SetDefault("PrivateKeyFile", "priv.pem")
	viper.SetDefault("PrivateKeyEncrypted", false)
	viper.SetDefault("AdminSessionLength", 15*time.Minute)
	viper.SetDefault("MaxMindPath", "db.mmdb")

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
		BaseURL:                viper.GetString("BaseURL"),
		HTTPS:                  viper.GetBool("HTTPS"),
		LogRequests:            viper.GetBool("LogRequests"),
		CertFile:               viper.GetString("CertFile"),
		KeyFile:                viper.GetString("KeyFile"),
		Base64EncodedPublicKey: viper.GetString("Base64EncodedPublicKey"),
		KeyType:                viper.GetString("KeyType"),
		PrivateKeyFile:         viper.GetString("PrivateKeyFile"),
		PrivateKeyEncrypted:    viper.GetBool("PrivateKeyEncrypted"),
		PrivateKeyPassword:     viper.GetString("PrivateKeyPassword"),
		AdminSessionLength:     viper.GetDuration("AdminSessionLength"),
		MaxMindPath:            viper.GetString("MaxMindPath"),
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

	info, err := file.Stat()
	if err != nil {
		panic(errors.Wrap(err, "Couldn't get info about private key file"))
	}

	bytes := make([]byte, info.Size())
	_, err = file.Read(bytes)
	if err != nil {
		panic(errors.Wrap(err, "Couldn't read private key file"))
	}

	p, _ := pem.Decode(bytes)
	if p == nil {
		panic(errors.New("File was not PEM-formatted"))
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
	if c.DatabaseType == "sqlite3" {
		db.DB().SetMaxOpenConns(1)
	}

	err = db.AutoMigrate(&AppInfo{}).
		AutoMigrate(&AppServerInfo{}).
		AutoMigrate(&security.Key{}).
		AutoMigrate(&Admin{}).
		AutoMigrate(&security.KeySignature{}).
		AutoMigrate(&SigningKey{}).
		AutoMigrate(&Permission{}).Error
	if err != nil {
		panic(errors.Wrap(err, "Could not migrate schemas"))
	}

	d, err := newDisperser(c.MaxMindPath)
	if err != nil {
		panic(errors.Wrap(err, "Could not create event disperser"))
	}

	rsa, ok := pub.(*rsa.PublicKey)
	if !ok {
		panic(errors.New("Could not cast key as RSA"))
	}

	s = Server{
		c,
		db,
		d,
		rsa,
		priv,
		securecookie.New(securecookie.GenerateRandomKey(64), nil),
		security.NewKeyCache(c.ExpirationTime, c.CleanTime, rsa, db),
		security.NewKeyGen(),
	}
	return s
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

func (s *Server) recoverWrap(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				var statusCode int
				var response errorResponse
				if be, ok := err.(util.BubbledError); ok {
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

		headerAuthPatterns := []string{
			"/*/register/request/*",
		}
		for _, pattern := range headerAuthPatterns {
			if glob.Glob(pattern, r.URL.Path) {
				s.headerAuthentication(w, r)
				break
			}
		}

		if glob.Glob("/admin/*", r.URL.Path) {
			cookie, err := r.Cookie("admin-session")
			util.OptionalPanic(err, http.StatusUnauthorized, "No session cookie")

			var m map[string]interface{}
			err = s.sc.Decode("admin-session", cookie.Value, &m)
			util.OptionalPanic(err, http.StatusUnauthorized, "Invalid session "+
				"cookie")

			var expires time.Time
			val, found := m["set"]
			util.PanicIfFalse(found, http.StatusBadRequest, "Invalid cookie")

			expires = val.(time.Time)

			distance := time.Now().Sub(expires).Nanoseconds()
			valid := distance < s.Config.AdminSessionLength.Nanoseconds()
			util.PanicIfFalse(valid, http.StatusUnauthorized, "Session expired")

			encoded, err := s.sc.Encode("admin-session", time.Now())
			util.OptionalInternalPanic(err, "Could not update session cookie")

			http.SetCookie(w, &http.Cookie{
				Name:  "admin-session",
				Value: encoded,
				Path:  "/",
			})
		}

		// If none of the middleware panicked, serve the main route
		handle.ServeHTTP(w, r)
	})
}

// Returns error, ID, messageMAC
func getAuthDataFromHeaders(r *http.Request) (string, string, error) {
	parts := strings.Split(r.Header.Get("X-Authentication"), ":")
	if len(parts) != 2 {
		return "", "", errors.Errorf("Found %d parts, expected 2", len(parts))
	}
	return parts[0], parts[1], nil
}

// For requests coming from an app sever or the admin frontend
func (s *Server) headerAuthentication(w http.ResponseWriter, r *http.Request) {
	id, received, err := getAuthDataFromHeaders(r)
	util.OptionalBadRequestPanic(err, "Invalid X-Authentication header")

	hmacBytes, err := util.DecodeBase64(received)
	util.OptionalInternalPanic(err, "Failed to decode MAC")

	var key []byte
	var x, y *big.Int
	if r.Header.Get("X-Authentication-Type") == "admin-frontend" {
		var a Admin
		err = s.DB.Find(&a, Admin{ID: id}).Error
		util.OptionalBadRequestPanic(err, "Could not find admin with ID "+id)

		var sk SigningKey
		err = s.DB.Find(&sk, SigningKey{ID: a.PrimarySigningKeyID}).Error
		util.OptionalInternalPanic(err, "Could not find admin's signing key")

		skBytes, err := util.DecodeBase64(sk.PublicKey)
		util.OptionalInternalPanic(err, "Could not decode admin's signing key")

		x, y = elliptic.Unmarshal(elliptic.P256(), skBytes)
		key = s.kg.GetShared(x, y, nil)
	} else {
		var app AppServerInfo
		err := s.DB.First(&app, AppServerInfo{ID: id}).Error
		util.OptionalBadRequestPanic(err, "Could not find app server")

		x, y = elliptic.Unmarshal(elliptic.P256(), app.PublicKey)
		key = s.kg.GetShared(x, y, s.priv.D.Bytes())
	}

	route := []byte(r.URL.Path)
	var body []byte
	if r.ContentLength > 0 {
		_, err := r.Body.Read(body)
		util.OptionalInternalPanic(err, "Failed to read request body")
	}

	hash := hmac.New(sha256.New, []byte(util.EncodeBase64(key)))
	hash.Write(route)
	if len(body) > 0 {
		hash.Write(body)
	}

	match := hmac.Equal(hmacBytes, hash.Sum(nil))
	util.PanicIfFalse(match, http.StatusUnauthorized, "Invalid security headers")

	s.kg.PutShared(x, y, key)
}

// GetHandler returns the routes used by the 2Q2R server.
func (s *Server) GetHandler() http.Handler {
	router := mux.NewRouter()

	// Get the server's public key
	forMethod(router, "/v1/public", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.priv.PublicKey)
	}, "GET")

	// Admin routes
	ah := adminHandler{s}
	forMethod(router, "/admin/new", ah.NewAdmin, "POST")

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

	forMethod(router, "/admin/stats/listen", ah.RegisterListener, "GET")
	forMethod(router, "/admin/stats/recent", ah.GetMostRecent, "GET")

	forMethod(router, "/admin/public", func(w http.ResponseWriter,
		r *http.Request) {
		priv, exp, err := s.kg.GetAdminPriv()
		util.OptionalInternalPanic(err, "Could not get keys for admin frontend")

		x, y := elliptic.P256().ScalarBaseMult(priv)
		encoded := util.EncodeBase64(elliptic.Marshal(elliptic.P256(), x, y))
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"public":  encoded,
			"expires": exp.Seconds(),
		})
	}, "GET")

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
	th := newAuthHandler(s)
	forMethod(router, "/v1/auth/request/{userID}/{nonce}", th.Setup, "GET")
	forMethod(router, "/v1/auth/wait", th.Wait, "POST")
	forMethod(router, "/v1/auth/challenge", th.SetKey, "POST")
	forMethod(router, "/v1/auth/iframe", th.IFrame, "POST")
	forMethod(router, "/v1/auth", th.Authenticate, "POST")

	// Register routes
	rh := newRegisterHandler(s)
	forMethod(router, "/v1/register/request/{userID}", rh.Setup, "GET")
	forMethod(router, "/v1/register/wait", rh.Wait, "POST")
	forMethod(router, "/v1/register/challenge", rh.GetChallenge, "POST")
	forMethod(router, "/v1/register/iframe", rh.IFrame, "POST")
	forMethod(router, "/v1/register", rh.Register, "POST")

	// Static files
	fileServer := http.FileServer(rice.MustFindBox("assets").HTTPBox())
	router.PathPrefix("/").Handler(fileServer)
	h := s.recoverWrap(router)
	if s.Config.LogRequests {
		return handlers.LoggingHandler(os.Stdout, h)
	}
	return h
}
