# KeyNginx CLI Makefile - Phase 1

BINARY_NAME=keynginx
VERSION?=1.0.0-phase1
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/yourusername/keynginx/cmd.Version=$(VERSION) -X github.com/yourusername/keynginx/cmd.GitCommit=$(GIT_COMMIT) -X github.com/yourusername/keynginx/cmd.BuildTime=$(BUILD_TIME)"

BUILD_DIR=dist

.PHONY: help build clean test fmt lint deps

# Default target
all: clean deps fmt test build

help: ## Display help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy

fmt: ## Format code
	@echo "🎨 Formatting code..."
	go fmt ./...

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

build: clean fmt ## Build binary
	@echo "🔨 Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean: ## Clean build directory
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)

install: build ## Install to GOPATH/bin
	@echo "📥 Installing..."
	go install $(LDFLAGS) ./main.go

# Test certificate generation
test-certs: build ## Test certificate generation
	@echo "🧪 Testing certificate generation..."
	./$(BUILD_DIR)/$(BINARY_NAME) certs --domain test.local --out ./test-ssl --verbose
	@echo "✅ Test complete - check ./test-ssl/"

example: build ## Run example commands
	@echo "📋 Running example commands..."
	./$(BUILD_DIR)/$(BINARY_NAME) version
	./$(BUILD_DIR)/$(BINARY_NAME) certs --help
	./$(BUILD_DIR)/$(BINARY_NAME) certs --domain localhost --out ./example-ssl --verbose --overwrite

.DEFAULT_GOAL := help