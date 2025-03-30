.PHONY: init
init:
	go mod tidy

.PHONY: test
test: init
	go test $${PACKAGE:-./...} -cover

.PHONY: debug
debug: init
	dlv test $${PACKAGE} --listen=:40000 --headless=true --api-version=2
