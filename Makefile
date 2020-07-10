statik:
	statik -m -src=static-content

compile:
	go build

build: statik compile

.PHONY: statik compile build
