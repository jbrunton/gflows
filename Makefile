statik:
	statik -m -src=static-content

compile:
	go build ./cmd/jflows

build: statik compile

.PHONY: statik compile build
