.PHONY: init
init:
	go mod tidy

.PHONY: test
test: init
	go test -cover $$( \
		if [ -n "$$PACKAGE" ]; then \
			echo $$PACKAGE; \
		else \
			go list ./... | grep -v './examples'; \
		fi \
	)

.PHONY: debug
debug: init
	dlv test $${PACKAGE} --listen=:40000 --headless=true --api-version=2
