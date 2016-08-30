# Copyright 2016 Tera Insights, LLC. All Rights Reserved.

GO_DEPENDENCIES = github.com/gorilla/mux github.com/jinzhu/gorm github.com/jinzhu/gorm/dialects/sqlite github.com/spf13/viper

install_dependencies:
	for dep in $(GO_DEPENDENCIES) ; do \
		go get $$dep; \
	done

test: test_server

test_server:
	go test 2q2r/server
	rm server/test.db