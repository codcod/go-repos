APP_NAME=repos
PKG=github.com/codcod/repos
CMD_DIR=./cmd
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build run test lint fmt clean

all: build

build:
	go build -o bin/$(APP_NAME) $(CMD_DIR)/repos

run: build
	./bin/$(APP_NAME)

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

clean:
	rm -rf bin

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