SHELL := /bin/bash
version=$(shell cat VERSION)
LDFLAGS=-ldflags "-X main.AppVersion=$(version)"
format_output=$(shell gofmt -l .)

.PHONY: all
all: clean build

clean:
	rm -f hyperledger-updates

build: lint-check unit-test
	go build -o hyperledger-updates $(LDFLAGS) ./cmd

unit-test:
	CGO_ENABLED=0 go test -v ./...

lint-check:
	@[ "$(format_output)" == "" ] || exit -1

format:
	go fmt ./...
