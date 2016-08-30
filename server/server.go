// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/spf13/viper"
)

// AppInfo is the Gorm model that holds information about an app.
type AppInfo struct {
	gorm.Model

	ID       string
	Name     string
	AuthType string
	AuthData string // JSON
}

// Config is the configuration for the server.
type Config struct {
	Port         int
	DatabaseType string
	DatabaseName string
}

func main() {
	viper.SetDefault("Port", 8080)
	viper.SetDefault("DatabaseType", "sqlite3")
	viper.SetDefault("DatabaseName", "test.db")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error when reading config file: %s\n", err))
	}
	c := Config{
		viper.GetInt("Port"),
		viper.GetString("DatabaseType"),
		viper.GetString("DatabaseName"),
	}
	s, err := MakeServer(c)
	if err != nil {
		panic(fmt.Errorf("Could not start server: %s\n", err))
	}
	s.ListenAndServe()
}

// Taken from https://git.io/v6xHB
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encErr := encoder.Encode(data)
	if encErr != nil {
		log.Printf("Failed to encode data as JSON: %v", encErr)
	}
}

func appInfoExists(appID string) bool {
	return false
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

// MakeHandler returns the routes used by the 2Q2R server.
func MakeHandler(db *gorm.DB) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/v1/info/{appID}", func(w http.ResponseWriter, r *http.Request) {
		appID := mux.Vars(r)["appID"]
		if !appInfoExists(appID) {
			msg := "Could not find information for app with ID " + appID
			writeJSON(w, http.StatusNotFound, msg)
		} else {
			var info AppInfo
			db.First(&info, "ID = ?", appID)
			writeJSON(w, http.StatusOK, info)
		}
	})
	return router
}

// MakeServer returns a new server initialized by the given configuration.
func MakeServer(c Config) (*http.Server, error) {
	s := &http.Server{
		Addr:           ":" + string(c.Port),
		Handler:        MakeHandler(MakeDB(c)),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return s, nil
}
