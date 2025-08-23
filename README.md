# KeyNginx CLI - Phase 1

A simple SSL certificate generator CLI tool built with Go.

## Phase 1 Features

- ✅ Generate RSA private keys (2048, 3072, 4096 bits)
- ✅ Create self-signed SSL certificates
- ✅ Support for localhost and custom domains
- ✅ Configurable certificate validity period
- ✅ Proper file permissions (600 for private keys)
- ✅ Certificate validation and information display

## Installation

```bash
# Clone repository
git clone https://github.com/sinhaparth5/keynginx
cd keynginx

# Build
make build

# Or install directly
make install
```

## Usage

### Generate certificates for localhost
```bash
keynginx certs --domain localhost --out ./ssl
```

### Generate certificates for custom domain
```bash
keynginx certs --domain myapp.local --out ./ssl --key-size 4096 --validity 730
```

### Verbose output with certificate details
```bash
keynginx certs --domain example.com --out ./certs --verbose
```

### Full example with all options
```bash
keynginx certs \
  --domain myapp.local \
  --out ./ssl \
  --key-size 2048 \
  --validity 365 \
  --country US \
  --state CA \
  --city "San Francisco" \
  --organization "My Company" \
  --unit "IT Department" \
  --email admin@myapp.local \
  --overwrite \
  --verbose
```

## Development

```bash
# Install dependencies
make deps

# Format code
make fmt

# Run tests
make test

# Build
make build

# Test certificate generation
make test-certs
```

# KeyNginx CLI Makefile - Phase 2

BINARY_NAME=keynginx
VERSION?=1.0.0-phase2
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/yourusername/keynginx/cmd.Version=$(VERSION) -X github.com/yourusername/keynginx/cmd.GitCommit=$(GIT_COMMIT) -X github.com/yourusername/keynginx/cmd.BuildTime=$(BUILD_TIME)"

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
```

## 10. Updated README.md
```markdown
# KeyNginx CLI - Phase 2

SSL certificate generator with Nginx configuration and Docker Compose generation.

## Phase 2 Features

- ✅ Generate RSA SSL certificates
- ✅ Create complete Nginx configurations with security headers
- ✅ Generate Docker Compose files
- ✅ Interactive project setup
- ✅ Service proxy configuration
- ✅ Security levels (strict/balanced/permissive)
- ✅ Project configuration management

## Installation

```bash
# Install dependencies
go mod download

# Build
make build
```

## Usage

### Basic project initialization
```bash
keynginx init --domain myapp.local --output ./myapp
```

### Interactive mode
```bash
keynginx init --interactive
```

### Advanced configuration with services
```bash
keynginx init \
  --domain api.local \
  --services "frontend:3000:/,backend:8000:/api" \
  --custom-headers "X-API-Version:v2.0,X-Environment:dev" \
  --security-level strict \
  --output ./api-project
```

### Just generate certificates (from Phase 1)
```bash
keynginx certs --domain localhost --out ./ssl
```

## Generated Project Structure

```
myapp/
├── ssl/
│   ├── private.key      # SSL private key
│   └── certificate.crt  # SSL certificate
├── nginx.conf           # Complete Nginx configuration
├── docker-compose.yml   # Docker setup
└── keynginx.yaml       # Project configuration
```

## Security Levels

- **strict**: Maximum security headers, strict CSP, HSTS
- **balanced**: Good security with compatibility (default)
- **permissive**: Basic security headers only

## Testing

```bash
# Test all Phase 2 features
make test-init
make test-services
make test-interactive
```

## Next Phase

**Phase 3**: Docker integration with container management commands
```

This completes Phase 2 with full Nginx configuration generation, security headers, templates, and project initialization! 🎉

## Next Phases

- **Phase 3**: Docker integration
- **Phase 4**: Advanced features and polish
```