#!/bin/bash

set -e

# KeyNginx Install Script
BINARY_NAME="keynginx"
GITHUB_REPO="sinhaparth5/keynginx"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    i386|i686)
        ARCH="386"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case $OS in
    linux)
        PLATFORM="linux"
        ;;
    darwin)
        PLATFORM="darwin"
        ;;
    *)
        echo -e "${RED}Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

echo -e "${BLUE}KeyNginx Installer${NC}"
echo "=================="
echo -e "Detected platform: ${YELLOW}$PLATFORM-$ARCH${NC}"

# Get latest release version
echo -e "${BLUE}Fetching latest release...${NC}"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}Failed to get latest version${NC}"
    exit 1
fi

echo -e "Latest version: ${GREEN}$LATEST_VERSION${NC}"

# Download URL
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$LATEST_VERSION/${BINARY_NAME}-${LATEST_VERSION}-${PLATFORM}-${ARCH}.tar.gz"

if [ "$PLATFORM" = "darwin" ]; then
    DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$LATEST_VERSION/${BINARY_NAME}-${LATEST_VERSION}-${PLATFORM}-${ARCH}.tar.gz"
fi

echo -e "${BLUE}Downloading ${BINARY_NAME}...${NC}"
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

curl -sL "$DOWNLOAD_URL" -o "${BINARY_NAME}.tar.gz"

if [ $? -ne 0 ]; then
    echo -e "${RED}Download failed${NC}"
    exit 1
fi

# Extract
echo -e "${BLUE}Extracting...${NC}"
tar -xzf "${BINARY_NAME}.tar.gz"

# Find binary in extracted files
BINARY_PATH=$(find . -name "$BINARY_NAME" -type f | head -1)

if [ -z "$BINARY_PATH" ]; then
    echo -e "${RED}Binary not found in archive${NC}"
    exit 1
fi

# Install
echo -e "${BLUE}Installing to $INSTALL_DIR...${NC}"

if [ -w "$INSTALL_DIR" ]; then
    cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    echo -e "${YELLOW}Need sudo to install to $INSTALL_DIR${NC}"
    sudo cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Cleanup
cd /
rm -rf "$TEMP_DIR"

# Verify installation
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ KeyNginx installed successfully!${NC}"
    echo -e "${BLUE}Version: $($BINARY_NAME version)${NC}"
    echo ""
    echo -e "${GREEN}Get started:${NC}"
    echo -e "  ${BINARY_NAME} init --domain myapp.local"
    echo -e "  ${BINARY_NAME} --help"
else
    echo -e "${RED}Installation failed - binary not found in PATH${NC}"
    echo -e "${YELLOW}Try adding $INSTALL_DIR to your PATH${NC}"
    exit 1
fi