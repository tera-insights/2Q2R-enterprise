# 2Q2R-enterprise
Enterprise of the 2Q2R server

## Setup

1. Install [Glide](https://github.com/Masterminds/glide#install). Make sure
you've either set your `$GOBIN` or your `$PATH` includes `$GOPATH/bin`.
2. `make install_dependencies` will install dependencies inside `vendor/`. 
`go run` is not always aware of the vendor folder. Specifically, as of 1.7.3,
go ignores `vendor` unless it is run inside the `$GOPATH`. See more
[here](https://github.com/golang/go/issues/14566). So, make sure you have
installed this package inside your `$GOPATH`.

## Running

1. If not already done, bootstrap the database with `go run
cmd/bootstrap/bootstrap.go`.  
2. `go run cmd/server/server.go`

## Configuring

Edit the file `config.yaml` and set at least `BaseURL`. The local changes 
will not be checked into the GIT repository. Potentially, the `Port` needs 
to be changed as well. Make sure the port is prefixed by :, e.g. `:8080`

To bootstrap the database, grab the file `bootstrap.go` from slack. 
Edit it to contain desired info (such as appIDs and such). The info here
 has to match the info in the demo app. 

Run it with
```
go run bootstrap.go
``` 

## Running

```
make run
```

## Checking the info in the database

```
sqlite3 test.db
.schema
select * from app_server_infos;
select * from keys;
```
## Documentation

Run `godoc -http=:6060` and then navigate to `localhost:6060/pkg/2q2r`. 

