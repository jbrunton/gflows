statik:
	statik -m -src=static-content

go-build:
	go build

compile: statik go-build

build: compile test

unit-test:
	go test -coverprofile c.out $$(go list ./... | grep -v /e2e)

e2e-test:
	go test ./e2e

test: unit-test e2e-test

.PHONY: statik go-build compile build unit-test e2e-test test

.DEFAULT_GOAL := build
