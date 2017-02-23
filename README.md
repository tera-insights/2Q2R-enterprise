# 2Q2R-enterprise

## Setup

0. Goto $GOPATH/src and then `git clone  git@github.com:alinVD/2Q2R-enterprise.git 2q2r`
1. Install [Glide](https://github.com/Masterminds/glide#install). Make sure
you've either set your `$GOBIN` or your `$PATH` includes `$GOPATH/bin`. 
2. `make install_dependencies` will install dependencies inside `vendor/`. 
`go run` is not always aware of the vendor folder. Specifically, as of 1.7.3,
go ignores `vendor` unless it is run inside the `$GOPATH`. See more
[here](https://github.com/golang/go/issues/14566). So, make sure you have
installed this package inside your `$GOPATH`.
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
1. If not already done, bootstrap the database with `go run cmd/bootstrap/bootstrap.go`.
Also make sure you have the latest bootstrap file from the 2Q2R Slack; if you replace an
older bootstrap file, you will need to delete `test.db` before running the new one.
2. Generate a private key with the following shell script: `openssl ecparam -name secp521r1 -genkey -out priv.pem -noout`
2. `make run` (which simply executes `go run cmd/server/server.go`)

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

