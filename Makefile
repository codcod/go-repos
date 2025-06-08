APP_NAME=repos
PKG=github.com/codcod/repos
CMD_DIR=./cmd
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Version information - can be overridden by environment variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%d")

.PHONY: all build run test test-unit test-integration test-coverage test-bench lint fmt clean

all: build

build:
	VERSION=$(VERSION) COMMIT=$(COMMIT) BUILD_DATE=$(BUILD_DATE) go build -o bin/$(APP_NAME) $(CMD_DIR)/repos

run: build
	./bin/$(APP_NAME)

test: test-unit

test-unit:
	go test -v ./internal/...
	go test -v ./cmd/...

test-integration:
	go test -v -tags=integration .

test-coverage:
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-bench:
	go test -v -bench=. -benchmem ./...

test-all: test-unit test-integration test-coverage

lint:
	golangci-lint run

fmt:
	go fmt ./...

clean:
	rm -rf bin coverage.out coverage.html repos-test

modtidy:
	go mod tidy

deps:
	go mod download

# For convenience, install golangci-lint if not present
GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)
install-lint:
ifndef GOLANGCI_LINT
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.59.1
endif