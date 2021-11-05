module 2q2r

go 1.16

replace github.com/tera-insights/2Q2R-enterprise/server => ./server

replace github.com/tera-insights/2Q2R-enterprise/security => ./security

replace github.com/tera-insights/2Q2R-enterprise/util => ./util

require (
	github.com/GeertJohan/go.rice v1.0.2 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/oschwald/maxminddb-golang v1.8.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/spf13/viper v1.9.0 // indirect
	github.com/tera-insights/2Q2R-enterprise v0.2.0 // indirect
	github.com/tera-insights/2Q2R-enterprise/security v0.0.0-00010101000000-000000000000
	github.com/tera-insights/2Q2R-enterprise/server v0.0.0-00010101000000-000000000000
	github.com/tera-insights/2Q2R-enterprise/util v0.0.0-00010101000000-000000000000
	github.com/tstranex/u2f v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
)

replace github.com/tera-insights/2Q2R-enterprise => ./
