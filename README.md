# 2Q2R-enterprise
Enterprise of the 2Q2R server

## Setup

1. Install [Glide](https://github.com/Masterminds/glide#install). Make sure
you've either set your `$GOBIN` or your `$PATH` includes `$GOPATH/bin`.
2. `make install_dependencies`

## Running

1. If not already done, bootstrap the database with `go run
cmd/bootstrap/bootstrap.go`. 
2. `go run cmd/server/server.go` 

## Documentation

Run `godoc -http=:6060` and then navigate to `localhost:6060/pkg/2q2r`. 