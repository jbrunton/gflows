statik:
	statik -m -src=static-content

compile:
	go build

build: statik compile test

unit-test:
	go test -coverprofile c.out $$(go list ./... | grep -v /e2e)

e2e-test:
	go test ./e2e

test: unit-test e2e-test

.PHONY: statik compile build unit-test e2e-test test
