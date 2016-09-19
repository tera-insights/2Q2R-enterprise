# Copyright 2016 Tera Insights, LLC. All Rights Reserved.

install_dependencies:
	go get 2q2r/server

test: test_server

server/assets.go: server/assets/*
	cd server && go-bindata-assetfs -pkg server assets/* && mv bindata_assetfs.go assets.go

test_server:
	go test 2q2r/server
	rm server/test.db