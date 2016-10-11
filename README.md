# 2Q2R-enterprise
Enterprise of the 2Q2R server

## Installation 

Make sure that `$GOPATH` is defined. Then: 
```
mkdir $GOPATH/src/2q2r
ln -sf $PATH2Q2R/server $GOPATH/src/2q2r
```
where `$PATH2Q2R` is the path to the GIT repository of `2q2r-enterprise`.

Run
```
make install_dependencies
```

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
# Documentation

Run `godoc -http=:6060` and then navigate to `localhost:6060/pkg/2q2r`. 

