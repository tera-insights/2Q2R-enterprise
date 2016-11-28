# 2Q2R-enterprise

## Setup

1. Install [Glide](https://github.com/Masterminds/glide#install). Make sure
you've either set your `$GOBIN` or your `$PATH` includes `$GOPATH/bin`. 
2. `make install_dependencies`
3. Install a MaxMind City database. The free tier, GeoLite2, is available 
[here](http://dev.maxmind.com/geoip/geoip2/geolite2/). Make sure its path is
either set in the `config.yaml` or is the default `./db.mmdb`.
4. Bootstrap the database with `go run cmd/bootstrap/bootstrap.go`. This script
requires a `bootstrap.json` config file that is a `server.NewAdminRequest`.

## Signing
The first admin's public key must be signed by Tera Insights by
`go run cmd/sign/sign.go`. This script takes an info file that has the admin's 
public key and generates a file with the admin ID and signature.  

## Running
`go run cmd/server/server.go` 

## Documentation

Run `godoc -http=:6060` and then navigate to `localhost:6060/pkg/2q2r`. 