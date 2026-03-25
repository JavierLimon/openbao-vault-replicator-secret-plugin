.PHONY: build test test-cover lint fmt

BINARY_NAME=vault-openbao-replicator
MAIN_PATH=./cmd/vault-replicator

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
