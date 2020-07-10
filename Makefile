statik:
	statik -m -src=static-content

compile:
	go build ./cmd/gflows

build: statik compile

.PHONY: statik compile build
