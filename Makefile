GO_DEPENDENCIES = github.com/gorilla/mux

install_dependencies:
	for dep in $(GO_DEPENDENCIES) ; do \
		go get $$dep; \
	done

test: test_server

test_server:
	go test 2q2r/server