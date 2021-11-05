# Copyright 2016 Tera Insights, LLC. All Rights Reserved.

install_dependencies:
	go mod vendor

test: test_server

assets:
	cd server && rice embed-go

test_server:
	go test -v 2q2r/server
	rm server/test.db

run: 
	go run cmd/server/server.go --config-path=config.example.yaml
	
build: **/*.go
	go build -o bin/2Q2R.linux cmd/server/server.go