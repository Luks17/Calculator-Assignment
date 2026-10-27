.DEFAULT_GOAL := help

.PHONY: help env generate build test lint format debug

help:
	@echo "env       Copy .env.example to .env (fails if it exists)"
	@echo "generate  Regenerate the server code from the OpenAPI spec"
	@echo "build     Build this module's executables and packages"
	@echo "test      Run this module's tests with the race detector"
	@echo "lint      Lint Go code with golangci-lint"
	@echo "format    Format Go code with gofmt via golangci-lint"
	@echo "debug     Run module under wgo and Delve"

env:
	@test ! -e .env && test ! -L .env || { echo ".env already exists"; exit 1; }
	cp .env.example .env

generate:
	go tool ogen --config=ogen.yml --target=internal/transport/http/server/gen --clean ../../contracts/calc/openapi/schema.yml

build:
	go build ./...

test:
	go test -race ./...

lint:
	golangci-lint run ./...

format:
	golangci-lint fmt --enable=gofmt ./...

debug:
	wgo -file .go -file go.mod -file go.sum \
		dlv debug ./cmd/server/main.go --continue --accept-multiclient --headless=true \
		--output=/tmp/calc-service-debug --listen=":2345" --api-version=2
