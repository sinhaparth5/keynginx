# KeyNginx CLI Makefile - Phase 2

BINARY_NAME=keynginx
VERSION?=1.0.0-phase2
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/sinhaparth5/keynginx/cmd.Version=$(VERSION) -X github.com/sinhaparth5/keynginx/cmd.GitCommit=$(GIT_COMMIT) -X github.com/sinhaparth5/keynginx/cmd.BuildTime=$(BUILD_TIME)"

BUILD_DIR=dist

.PHONY: help build clean test fmt lint deps

all: clean deps fmt test build

help: ## Display help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $1, $2}' $(MAKEFILE_LIST)

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

# Test Phase 2 features
test-init: build ## Test init command
	@echo "🧪 Testing init command..."
	./$(BUILD_DIR)/$(BINARY_NAME) init --domain test.local --output ./test-project --overwrite
	@echo "✅ Test complete - check ./test-project/"

test-interactive: build ## Test interactive mode
	@echo "🧪 Testing interactive mode..."
	@echo "localhost\nbalanced\nn\n" | ./$(BUILD_DIR)/$(BINARY_NAME) init --interactive --output ./test-interactive --overwrite

test-services: build ## Test with services
	@echo "🧪 Testing with services..."
	./$(BUILD_DIR)/$(BINARY_NAME) init \
		--domain api.local \
		--output ./test-services \
		--services "frontend:3000:/,backend:8000:/api" \
		--custom-headers "X-API-Version:v2.0" \
		--security-level strict \
		--overwrite

example: build ## Run example commands
	@echo "📋 Running example commands..."
	./$(BUILD_DIR)/$(BINARY_NAME) version
	./$(BUILD_DIR)/$(BINARY_NAME) init --help
	./$(BUILD_DIR)/$(BINARY_NAME) init --domain example.local --output ./example --overwrite

.DEFAULT_GOAL := help