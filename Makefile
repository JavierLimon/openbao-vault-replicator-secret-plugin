.PHONY: build test test-cover lint fmt clean

NAME := vault-replicator
PLUGIN_NAME := replicator
VERSION := 1.0.0
GO := go
GOFLAGS := -ldflags "-X github.com/JavierLimon/openbao-vault-replicator-secret-plugin/plugin.version=$(VERSION)"
BUILD_DIR := ./dist

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME) ./cmd/vault-replicator

test:
	$(GO) test -v ./plugin/...

test-cover:
	$(GO) test -v -coverprofile=coverage.out ./plugin/...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint:
	$(GO) vet ./plugin/...
	golangci-lint run ./plugin/...

fmt:
	$(GO) fmt ./plugin/...
	gofmt -s -w ./plugin/...

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
