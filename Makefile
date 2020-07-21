statik:
	statik -m -src=static-content

compile:
	go build

build: statik compile

unit-test:
	go test -coverprofile coverage.out $$(go list ./... | grep -v /e2e)

e2e-test:
	go test ./e2e

test: unit-test e2e-test

.PHONY: statik compile build unit-test e2e-test test
