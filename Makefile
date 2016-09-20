# Copyright 2016 Tera Insights, LLC. All Rights Reserved.

install_dependencies:
	go get 2q2r/server

test: test_server

assets:
	cd server && rice embed-go

test_server:
	go test 2q2r/server
	rm server/test.db