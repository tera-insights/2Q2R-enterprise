# 2Q2R-enterprise
Enterprise of the 2Q2R server

## Setup

1. Install [Glide](https://github.com/Masterminds/glide#install). Make sure
you've either set your `$GOBIN` or your `$PATH` includes `$GOPATH/bin`. 
2. `make install_dependencies`
3. Install a MaxMind City database. The free tier, GeoLite2, is available 
[here](http://dev.maxmind.com/geoip/geoip2/geolite2/). Make sure its path is
either set in the `config.yaml` or is the default `./db.mmdb`.
4. If not already done, bootstrap the database with `go run
cmd/bootstrap/bootstrap.go`.

## Running
1. `go run cmd/server/server.go` 

## Documentation

Run `godoc -http=:6060` and then navigate to `localhost:6060/pkg/2q2r`. 