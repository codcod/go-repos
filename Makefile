# Makefile for repos CLI tool

APP_NAME=repos
PKG=github.com/codcod/repos
CMD_DIR=./cmd
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Version information - can be overridden by environment variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%d")

# Go build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

.PHONY: all build run test test-unit test-integration test-coverage test-bench test-race test-all lint fmt clean help

# Default target
all: build

## Build targets
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	go build $(LDFLAGS) -o bin/$(APP_NAME) $(CMD_DIR)/repos

build-all: ## Build for all platforms
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 $(CMD_DIR)/repos
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 $(CMD_DIR)/repos
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe $(CMD_DIR)/repos

## Development targets
run: build ## Build and run the application
	./bin/$(APP_NAME)

dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	go mod tidy
	go mod download
	@$(MAKE) install-tools

## Testing targets
test: test-unit ## Run unit tests (default)

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v ./internal/... ./cmd/...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -tags=integration .

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | tail -1

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	go test -race ./...

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	go test -v -bench=. -benchmem -benchtime=5s ./...

test-all: test-unit test-integration test-coverage test-race ## Run all tests

## Code quality targets
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	elif [ -f "$$(go env GOPATH)/bin/golangci-lint" ]; then \
		$$(go env GOPATH)/bin/golangci-lint run; \
	else \
		echo "golangci-lint not found, installing..."; \
		$(MAKE) install-lint; \
		$$(go env GOPATH)/bin/golangci-lint run; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found, installing..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: fmt vet lint ## Run all code quality checks

## Security targets
security: ## Run security scans
	@echo "Running security checks..."
	@echo "Running govulncheck vulnerability scanner..."
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	govulncheck ./...
	@echo "Running staticcheck..."
	@if ! command -v staticcheck >/dev/null 2>&1; then \
		echo "Installing staticcheck..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	staticcheck ./...

## Cleanup targets
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	rm -rf bin/ coverage.out coverage.html repos-test

clean-all: clean ## Clean everything including dependencies
	go clean -cache
	go clean -modcache

## Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	go mod tidy

## Tools installation
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@$(MAKE) install-go-tools
	@$(MAKE) install-lint
	@$(MAKE) install-go-tools
	@$(MAKE) setup-commitlint

install-go-tools: ## Install Go development tools
	@echo "Installing Go development tools..."
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	else \
		echo "goimports already installed"; \
	fi
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	else \
		echo "govulncheck already installed"; \
	fi

install-lint: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	@# Remove existing golangci-lint to avoid version conflicts
	@rm -f $$(go env GOPATH)/bin/golangci-lint 2>/dev/null || true
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest
	@echo "golangci-lint installed successfully"

setup-commitlint: ## Setup commitlint Git hooks
	@echo "Setting up commitlint..."
	@./scripts/setup-commitlint.sh

install-commitlint: ## Install commitlint dependencies
	@echo "Installing commitlint dependencies..."
	@if command -v npm >/dev/null 2>&1; then \
		npm install; \
		echo "✅ Commitlint dependencies installed"; \
	else \
		echo "❌ npm not found. Please install Node.js and npm first."; \
		exit 1; \
	fi

## Pre-commit workflow
pre-commit: clean fmt vet lint test-all ## Run complete pre-commit checks
	@echo "✅ Pre-commit checks completed successfully!"

## Help
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""