#!/bin/bash

# Mini-Proxy Multi-Platform Build Script
# Script ini membuild untuk semua platform yang didukung

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

VERSION="${1:-dev}"

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Mini-Proxy Multi-Platform Build Script    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""
echo "Version: $VERSION"
echo "This will build for all supported platforms"
echo ""

# Array of platforms
PLATFORMS=(
    "linux:amd64"
    "linux:arm64"
    "darwin:amd64"
    "darwin:arm64"
    "windows:amd64"
)

# Sync modules first
echo -e "${YELLOW}Syncing Go modules...${NC}"
go work sync
echo -e "${GREEN}✓ Modules synced${NC}"
echo ""

# Create build directory
mkdir -p build
BUILD_DIR="build"

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS=':' read -r OS ARCH <<< "$PLATFORM"
    
    echo -e "${BLUE}─────────────────────────────────────────${NC}"
    echo -e "${YELLOW}Building for: $OS-$ARCH${NC}"
    
    BINARY_NAME="main"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="main.exe"
    fi
    
    OUTPUT_PATH="$BUILD_DIR/${OS}-${ARCH}/$BINARY_NAME"
    mkdir -p "$(dirname "$OUTPUT_PATH")"
    
    # Build
    GOOS="$OS" GOARCH="$ARCH" go build \
        -o "$OUTPUT_PATH" \
        -ldflags="-X main.Version=$VERSION" \
        ./webcore/main.go 2>&1
    
    if [ -f "$OUTPUT_PATH" ]; then
        SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
        echo -e "${GREEN}✓ Built: $(basename "$OUTPUT_PATH") ($SIZE)${NC}"
    else
        echo -e "${RED}✗ Build failed for $OS-$ARCH${NC}"
    fi
done

echo ""
echo -e "${BLUE}─────────────────────────────────────────${NC}"
echo ""

# Create release packages for each platform
RELEASE_DIR_PREFIX="mini-proxy"

for PLATFORM in "${PLATFORMS[@]}"; do
    IFS=':' read -r OS ARCH <<< "$PLATFORM"
    
    echo -e "${YELLOW}Creating package for: $OS-$ARCH${NC}"
    
    RELEASE_DIR="${RELEASE_DIR_PREFIX}-${OS}-${ARCH}"
    BINARY_NAME="main"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="main.exe"
    fi
    
    # Clean previous
    rm -rf "$RELEASE_DIR"
    
    # Create structure
    mkdir -p "$RELEASE_DIR/app"
    cp "$BUILD_DIR/${OS}-${ARCH}/$BINARY_NAME" "$RELEASE_DIR/app/main"
    chmod +x "$RELEASE_DIR/app/main" 2>/dev/null || true
    
    cp config.yaml.example "$RELEASE_DIR/app/config.yaml.example"
    cp go.work "$RELEASE_DIR/app/go.work"
    cp docker-compose-barebone.yml "$RELEASE_DIR/docker-compose.yml"
    cp Dockerfile.barebone "$RELEASE_DIR/Dockerfile"
    cp -r webcore "$RELEASE_DIR/"
    rm -rf "$RELEASE_DIR/webcore/.git" 2>/dev/null || true
    
    # Create archive
    ARCHIVE="${RELEASE_DIR}.tar.gz"
    tar -czf "$ARCHIVE" "$RELEASE_DIR" 2>&1
    
    SIZE=$(du -h "$ARCHIVE" | cut -f1)
    echo -e "${GREEN}✓ Package: $ARCHIVE ($SIZE)${NC}"
done

echo ""
echo -e "${BLUE}═════════════════════════════════════════${NC}"
echo -e "${GREEN}✓ All builds completed successfully!${NC}"
echo -e "${BLUE}═════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}Release packages created:${NC}"
ls -lh mini-proxy-*.tar.gz 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'
echo ""
