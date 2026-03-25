.PHONY: build test test-cover lint fmt

NAME := vault-replicator
PLUGIN_NAME := replicator
VERSION := 1.0.0
GO := go
GOFLAGS := -ldflags "-X github.com/JavierLimon/openbao-vault-replicator-secret-plugin/plugin.version=$(VERSION)"
BUILD_DIR := ./dist

build: fmt
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME) ./cmd/vault-replicator

test:
	$(GO) test -v ./...

test-cover:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

GOBIN := $(shell go env GOPATH)/bin

lint:
	@which $(GOBIN)/golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	$(GOBIN)/golangci-lint run

fmt:
	$(GO) fmt ./...
	@which $(GOBIN)/goimports > /dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	$(GOBIN)/goimports -w .
