# KeyNginx CLI Makefile - Phase 3

BINARY_NAME=keynginx
VERSION?=1.0.0-phase3
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/sinhaparth5/keynginx/cmd.Version=$(VERSION) -X github.com/sinhaparth5/keynginx/cmd.GitCommit=$(GIT_COMMIT) -X github.com/sinhaparth5/keynginx/cmd.BuildTime=$(BUILD_TIME)"

BUILD_DIR=dist

.PHONY: help build clean test fmt lint deps docker-check

all: clean deps fmt test build

help: 
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $1, $2}' $(MAKEFILE_LIST)

deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy

fmt: 
	@echo "🎨 Formatting code..."
	go fmt ./...

test: 
	@echo "🧪 Running tests..."
	go test -v ./...

build: clean fmt
	@echo "🔨 Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean:
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)

install: build 
	@echo "📥 Installing..."
	go install $(LDFLAGS) ./main.go

docker-check: 
	@echo "🐳 Checking Docker..."
	@docker --version >/dev/null 2>&1 || (echo "❌ Docker not found. Please install Docker." && exit 1)
	@docker ps >/dev/null 2>&1 || (echo "❌ Docker daemon not running. Please start Docker." && exit 1)
	@echo "✅ Docker is available"

test-workflow: build docker-check
	@echo "🧪 Testing complete Phase 3 workflow..."
	
	# Initialize project
	./$(BUILD_DIR)/$(BINARY_NAME) init --domain test.local --output ./test-workflow --overwrite
	
	# Start server
	cd ./test-workflow && ../$(BUILD_DIR)/$(BINARY_NAME) up
	
	# Check status
	cd ./test-workflow && ../$(BUILD_DIR)/$(BINARY_NAME) status
	
	# Wait a bit
	sleep 2
	
	# View logs (non-follow)
	cd ./test-workflow && timeout 5s ../$(BUILD_DIR)/$(BINARY_NAME) logs || true
	
	# Stop server
	cd ./test-workflow && ../$(BUILD_DIR)/$(BINARY_NAME) down
	
	@echo "✅ Workflow test complete!"

test-up: build docker-check 
	@echo "🧪 Testing up command..."
	./$(BUILD_DIR)/$(BINARY_NAME) init --domain test-up.local --output ./test-up --overwrite
	cd ./test-up && ../$(BUILD_DIR)/$(BINARY_NAME) up

test-status: build  
	@echo "🧪 Testing status command..."
	./$(BUILD_DIR)/$(BINARY_NAME) status --all

test-logs: build 
	@echo "🧪 Testing logs command..."
	cd ./test-workflow && timeout 3s ../$(BUILD_DIR)/$(BINARY_NAME) logs || true

clean-tests:
	@echo "🧹 Cleaning test projects..."
	rm -rf ./test-*
	@echo "✅ Test projects cleaned"

example: build docker-check 
	@echo "📋 Running Phase 3 example..."
	./$(BUILD_DIR)/$(BINARY_NAME) version
	./$(BUILD_DIR)/$(BINARY_NAME) init --domain example.local --output ./example --overwrite
	cd ./example && ../$(BUILD_DIR)/$(BINARY_NAME) up
	cd ./example && ../$(BUILD_DIR)/$(BINARY_NAME) status
	@echo "🌐 Visit https://localhost:8443 to see your server!"
	@echo "🛑 Run 'cd example && ../$(BUILD_DIR)/$(BINARY_NAME) down' to stop"

.DEFAULT_GOAL := help