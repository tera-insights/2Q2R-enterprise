# Copyright 2016 Tera Insights, LLC. All Rights Reserved.

install_dependencies:
	glide install

test: test_server

assets:
	cd server && rice embed-go

test_server:
	go test -v 2q2r/server
	rm server/test.db

run: 
	go run cmd/server/server.go
	