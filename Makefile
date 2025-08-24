# KeyNginx CLI Makefile - Fixed Version Handling

BINARY_NAME=keynginx
VERSION?=1.0.0
CLEAN_VERSION=$(shell echo $(VERSION) | sed 's/^v//')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X github.com/sinhaparth5/keynginx/cmd.Version=$(VERSION) -X github.com/sinhaparth5/keynginx/cmd.GitCommit=$(GIT_COMMIT) -X github.com/sinhaparth5/keynginx/cmd.BuildTime=$(BUILD_TIME)"

BUILD_DIR=dist
RELEASE_DIR=release
PACKAGE_DIR=packages

RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m 

.PHONY: help build clean test fmt lint deps docker-check

all: clean deps fmt test build

help:
	@echo "$(BLUE)KeyNginx CLI Build System$(NC)"
	@echo "=========================="
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "$(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: 
	@echo "$(BLUE)📦 Downloading dependencies...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✅ Dependencies ready$(NC)"

fmt:
	@echo "$(BLUE)🎨 Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

test:
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	go test -v ./...
	@echo "$(GREEN)✅ Tests passed$(NC)"

build: clean fmt 
	@echo "$(BLUE)🔨 Building $(BINARY_NAME) v$(VERSION) for current platform...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go
	@echo "$(GREEN)✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-all: clean fmt 
	@echo "$(BLUE)🔨 Building $(BINARY_NAME) v$(VERSION) for all platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	
	# Linux
	@echo "$(YELLOW)Building for Linux...$(NC)"
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./main.go
	GOOS=linux GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-386 ./main.go
	
	# macOS
	@echo "$(YELLOW)Building for macOS...$(NC)"
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./main.go
	
	# Windows
	@echo "$(YELLOW)Building for Windows...$(NC)"
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./main.go
	GOOS=windows GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-386.exe ./main.go
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe ./main.go
	
	@echo "$(GREEN)✅ Cross-compilation complete!$(NC)"
	@ls -la $(BUILD_DIR)/

package: build-all 
	@echo "$(BLUE)📦 Creating packages...$(NC)"
	@mkdir -p $(PACKAGE_DIR)/{linux,macos,windows}
	
	# Linux packages
	@$(MAKE) package-deb
	@$(MAKE) package-linux-tar
	
	# macOS packages
	@$(MAKE) package-macos-tar
	
	# Windows packages
	@$(MAKE) package-windows-zip
	
	@echo "$(GREEN)✅ All packages created in $(PACKAGE_DIR)/$(NC)"

package-deb: 
	@echo "$(YELLOW)Creating Debian package...$(NC)"
	@mkdir -p $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN
	@mkdir -p $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/bin
	@mkdir -p $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/share/doc/$(BINARY_NAME)
	@mkdir -p $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/share/man/man1
	
	# Copy binary
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/bin/$(BINARY_NAME)
	chmod 755 $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/bin/$(BINARY_NAME)
	
	# Create control file - FIXED: Use CLEAN_VERSION without 'v' prefix
	@echo "Package: $(BINARY_NAME)" > $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	@echo "Version: $(CLEAN_VERSION)" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	@echo "Section: utils" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	@echo "Priority: optional" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	@echo "Architecture: amd64" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	@echo "Depends: docker.io | docker-ce" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	@echo "Maintainer: KeyNginx Team <admin@keynginx.dev>" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	@echo "Description: SSL-enabled Nginx automation with Docker" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	@echo " KeyNginx automates SSL certificate generation, Nginx configuration," >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	@echo " and Docker container management for secure web development." >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/DEBIAN/control
	
	# Copy documentation
	cp README.md $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/share/doc/$(BINARY_NAME)/ || echo "README.md not found, skipping"
	echo "$(BINARY_NAME) ($(CLEAN_VERSION)) unstable; urgency=low" > $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/share/doc/$(BINARY_NAME)/changelog
	echo "" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/share/doc/$(BINARY_NAME)/changelog
	echo "  * Release $(VERSION)" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/share/doc/$(BINARY_NAME)/changelog
	echo "" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/share/doc/$(BINARY_NAME)/changelog
	echo " -- KeyNginx Team <admin@keynginx.dev>  $(shell date -R)" >> $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/share/doc/$(BINARY_NAME)/changelog
	gzip -9 $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION)/usr/share/doc/$(BINARY_NAME)/changelog
	
	# Build package - Use CLEAN_VERSION for filename
	dpkg-deb --build $(PACKAGE_DIR)/linux/$(BINARY_NAME)-$(CLEAN_VERSION) $(PACKAGE_DIR)/$(BINARY_NAME)-$(CLEAN_VERSION)-amd64.deb
	@echo "$(GREEN)✅ Debian package: $(PACKAGE_DIR)/$(BINARY_NAME)-$(CLEAN_VERSION)-amd64.deb$(NC)"

package-linux-tar: 
	@echo "$(YELLOW)Creating Linux tar packages...$(NC)"
	
	# AMD64
	@mkdir -p $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64/$(BINARY_NAME)
	cp README.md $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64/ 2>/dev/null || echo "README.md not found, skipping"
	echo "#!/bin/bash" > $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64/install.sh
	echo "cp $(BINARY_NAME) /usr/local/bin/" >> $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64/install.sh
	echo "chmod +x /usr/local/bin/$(BINARY_NAME)" >> $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64/install.sh
	echo "echo 'KeyNginx installed successfully!'" >> $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64/install.sh
	chmod +x $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64/install.sh
	cd $(PACKAGE_DIR)/tmp && tar -czf ../$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64
	
	# ARM64
	@mkdir -p $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-arm64
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-arm64/$(BINARY_NAME)
	cp README.md $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-arm64/ 2>/dev/null || echo "README.md not found, skipping"
	cp $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-amd64/install.sh $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-arm64/
	cd $(PACKAGE_DIR)/tmp && tar -czf ../$(BINARY_NAME)-$(CLEAN_VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-$(CLEAN_VERSION)-linux-arm64
	
	rm -rf $(PACKAGE_DIR)/tmp
	@echo "$(GREEN)✅ Linux packages: $(PACKAGE_DIR)/$(BINARY_NAME)-$(CLEAN_VERSION)-linux-*.tar.gz$(NC)"

package-macos-tar: 
	@echo "$(YELLOW)Creating macOS tar packages...$(NC)"
	
	# AMD64
	@mkdir -p $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64
	cp $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64/$(BINARY_NAME)
	cp README.md $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64/ 2>/dev/null || echo "README.md not found, skipping"
	echo "#!/bin/bash" > $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64/install.sh
	echo "cp $(BINARY_NAME) /usr/local/bin/" >> $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64/install.sh
	echo "chmod +x /usr/local/bin/$(BINARY_NAME)" >> $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64/install.sh
	echo "echo 'KeyNginx installed successfully! Run: keynginx --help'" >> $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64/install.sh
	chmod +x $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64/install.sh
	cd $(PACKAGE_DIR)/tmp && tar -czf ../$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64
	
	# ARM64 (Apple Silicon)
	@mkdir -p $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-arm64
	cp $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-arm64/$(BINARY_NAME)
	cp README.md $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-arm64/ 2>/dev/null || echo "README.md not found, skipping"
	cp $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-amd64/install.sh $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-arm64/
	cd $(PACKAGE_DIR)/tmp && tar -czf ../$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-$(CLEAN_VERSION)-darwin-arm64
	
	rm -rf $(PACKAGE_DIR)/tmp
	@echo "$(GREEN)✅ macOS packages: $(PACKAGE_DIR)/$(BINARY_NAME)-$(CLEAN_VERSION)-darwin-*.tar.gz$(NC)"

package-windows-zip: 
	@echo "$(YELLOW)Creating Windows zip packages...$(NC)"
	
	# AMD64
	@mkdir -p $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64
	cp $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64/$(BINARY_NAME).exe
	cp README.md $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64/ 2>/dev/null || echo "README.md not found, skipping"
	echo "@echo off" > $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64/install.bat
	echo "copy $(BINARY_NAME).exe %USERPROFILE%\\AppData\\Local\\Microsoft\\WindowsApps\\" >> $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64/install.bat
	echo "echo KeyNginx installed successfully!" >> $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64/install.bat
	echo "echo Run: keynginx --help" >> $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64/install.bat
	cd $(PACKAGE_DIR)/tmp && zip -r ../$(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64.zip $(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64
	
	# 386
	@mkdir -p $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-386
	cp $(BUILD_DIR)/$(BINARY_NAME)-windows-386.exe $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-386/$(BINARY_NAME).exe
	cp README.md $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-386/ 2>/dev/null || echo "README.md not found, skipping"
	cp $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-amd64/install.bat $(PACKAGE_DIR)/tmp/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-386/
	cd $(PACKAGE_DIR)/tmp && zip -r ../$(BINARY_NAME)-$(CLEAN_VERSION)-windows-386.zip $(BINARY_NAME)-$(CLEAN_VERSION)-windows-386
	
	rm -rf $(PACKAGE_DIR)/tmp
	@echo "$(GREEN)✅ Windows packages: $(PACKAGE_DIR)/$(BINARY_NAME)-$(CLEAN_VERSION)-windows-*.zip$(NC)"

release: clean deps test build-all package 
	@echo "$(BLUE)🚀 Creating release $(VERSION)...$(NC)"
	@mkdir -p $(RELEASE_DIR)
	
	# Copy all packages to release directory
	cp $(PACKAGE_DIR)/*.deb $(RELEASE_DIR)/ 2>/dev/null || true
	cp $(PACKAGE_DIR)/*.tar.gz $(RELEASE_DIR)/ 2>/dev/null || true
	cp $(PACKAGE_DIR)/*.zip $(RELEASE_DIR)/ 2>/dev/null || true
	
	# Copy plain binaries
	cp $(BUILD_DIR)/$(BINARY_NAME)-* $(RELEASE_DIR)/
	
	# Generate checksums
	cd $(RELEASE_DIR) && sha256sum * > checksums.txt
	
	@echo "$(GREEN)✅ Release $(VERSION) ready in $(RELEASE_DIR)/$(NC)"
	@echo "$(BLUE)Release contents:$(NC)"
	@ls -la $(RELEASE_DIR)/

clean:
	@echo "$(BLUE)🧹 Cleaning...$(NC)"
	rm -rf $(BUILD_DIR) $(RELEASE_DIR) $(PACKAGE_DIR)
	@echo "$(GREEN)✅ Cleaned$(NC)"

install: build 
	@echo "$(BLUE)📥 Installing $(BINARY_NAME)...$(NC)"
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ $(BINARY_NAME) installed to /usr/local/bin/$(NC)"
	@echo "Run: $(BINARY_NAME) --help"

.DEFAULT_GOAL := help