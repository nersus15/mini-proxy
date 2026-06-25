#!/bin/bash

# Mini-Proxy Local Build Script
# Script ini membantu build dan membuat release package secara lokal untuk testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="main"
RELEASE_DIR_PREFIX="mini-proxy"
BUILD_DIR="build"
VERSION="${1:-dev}"

echo -e "${YELLOW}=== Mini-Proxy Local Build Script ===${NC}"
echo "Version: $VERSION"
echo ""

# Create build directory
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Get current OS and ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$OS" = "darwin" ]; then
    OS="darwin"
elif [ "$OS" = "linux" ]; then
    OS="linux"
fi

ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

echo -e "${YELLOW}Building for: $OS-$ARCH${NC}"

# Clean previous build
rm -f "$BINARY_NAME"

# Sync go modules
echo -e "${YELLOW}Syncing Go modules...${NC}"
cd ..
go work sync
cd "$BUILD_DIR"

# Build binary
echo -e "${YELLOW}Building binary...${NC}"
GOOS="$OS" GOARCH="$ARCH" go build \
    -o "$BINARY_NAME" \
    -ldflags="-X main.Version=$VERSION" \
    ../webcore/main.go

if [ -f "$BINARY_NAME" ]; then
    echo -e "${GREEN}✓ Binary built successfully${NC}"
    chmod +x "$BINARY_NAME"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Create release structure
cd ..
RELEASE_DIR="${RELEASE_DIR_PREFIX}-${OS}-${ARCH}"
echo -e "${YELLOW}Creating release structure in: $RELEASE_DIR${NC}"

# Clean previous release
rm -rf "$RELEASE_DIR"

# Create directories
mkdir -p "$RELEASE_DIR/app"

# Copy binary
cp "$BUILD_DIR/$BINARY_NAME" "$RELEASE_DIR/app/main"
chmod +x "$RELEASE_DIR/app/main"

# Copy configuration
cp config.yaml.example "$RELEASE_DIR/app/config.yaml.example"
cp go.work "$RELEASE_DIR/app/go.work"

# Copy docker files
cp docker-compose-barebone.yml "$RELEASE_DIR/docker-compose.yml"
cp Dockerfile.barebone "$RELEASE_DIR/Dockerfile"

# Copy webcore directory
cp -r webcore "$RELEASE_DIR/"

# Remove unnecessary Go build cache
rm -rf "$RELEASE_DIR/webcore/.git"

echo -e "${GREEN}✓ Release structure created${NC}"
echo ""

# Create tar.gz archive
ARCHIVE_NAME="${RELEASE_DIR}.tar.gz"
echo -e "${YELLOW}Creating archive: $ARCHIVE_NAME${NC}"
tar -czf "$ARCHIVE_NAME" "$RELEASE_DIR"

if [ -f "$ARCHIVE_NAME" ]; then
    SIZE=$(du -h "$ARCHIVE_NAME" | cut -f1)
    echo -e "${GREEN}✓ Archive created: $ARCHIVE_NAME ($SIZE)${NC}"
else
    echo -e "${RED}✗ Archive creation failed${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== Build Complete ===${NC}"
echo ""
echo "Release package: $ARCHIVE_NAME"
echo "Release directory: $RELEASE_DIR"
echo ""
echo -e "${YELLOW}To test extraction:${NC}"
echo "  tar -xzf $ARCHIVE_NAME"
echo "  cd $RELEASE_DIR"
echo "  ./app/main proxy"
echo ""
