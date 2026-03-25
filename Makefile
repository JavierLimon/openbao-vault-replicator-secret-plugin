.PHONY: build test test-cover lint fmt

NAME := vault-replicator
PLUGIN_NAME := replicator
VERSION := 1.0.0
GO := go
GOFLAGS := -ldflags "-X github.com/JavierLimon/openbao-vault-replicator-secret-plugin/plugin.version=$(VERSION)"
BUILD_DIR := ./dist
GOPATH := $(shell go env GOPATH)

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME) ./cmd/vault-replicator

test:
	$(GO) test -v ./...

test-cover:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint:
	$(GOPATH)/bin/golangci-lint run

fmt:
	$(GO) fmt ./...
	$(GOPATH)/bin/goimports -w .
