statik:
	statik -m -src=static/content -dest=static

go-build:
	go build

go-build-release:
	export CGO_ENABLED=0
	repo_flags="-ldflags=-buildid= -trimpath"
	GOOS=darwin GOARCH=amd64 go build $$repo_flags -o gflows-darwin-amd64
	GOOS=linux GOARCH=amd64 go build $$repo_flags -o gflows-linux-amd64

compile: statik go-build

compile-release: statik go-build-release

build: compile test

unit-test:
	go test -coverprofile c.out $$(go list ./... | grep -v /e2e)

e2e-test:
	go test ./e2e

test: unit-test e2e-test

.PHONY: statik go-build compile build unit-test e2e-test test

.DEFAULT_GOAL := build
