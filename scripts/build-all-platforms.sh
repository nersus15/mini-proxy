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

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
BASE_DIR="$(dirname "$SCRIPT_DIR")"
cd $BASE_DIR

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
    "windows:amd64"
)

# Sync modules first
echo -e "${YELLOW}Syncing Go modules...${NC}"
go work sync
echo -e "${GREEN}✓ Modules synced${NC}"
echo ""

if [ -d "dist" ]; then rm -rf dist; fi
mkdir -p dist

BUILD_DIR=dist

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

    
    if [ "$OS" = "linux" ] && [ "$ARCH" = "amd64" ]; then
        echo -e "${YELLOW}Running build inside Docker container to handle CGO (Kafka)...${NC}"
        
        # Jalankan build di dalam container golang resmi Linux
        # Kita install librdkafka-dev agar confluent-kafka-go bisa dicompile
        docker run --rm \
            -v "$BASE_DIR":/app \
            -w /app \
            golang:1.25 \
            sh -c "apt-get update && apt-get install -y librdkafka-dev && \
                   GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build \
                   -o $OUTPUT_PATH \
                   -ldflags=\"-X main.Version=$VERSION\" \
                   ./webcore/main.go" 2>&1
                   
    elif [ "$OS" = "windows" ]; then
        # Untuk Windows, jika library kafka kamu tidak dipakai di windows, kamu bisa matikan CGO
        # Namun jika dipakai, windows juga membutuhkan gcc-mingw (lebih aman dibuild terpisah atau pakai docker image multi-arch)
        echo -e "${YELLOW}Building for Windows (CGO Disabled)...${NC}"
        GOOS="$OS" GOARCH="$ARCH" CGO_ENABLED=0 go build \
            -o "$OUTPUT_PATH" \
            -ldflags="-X main.Version=$VERSION" \
            ./webcore/main.go 2>&1
    else
        # Default fallback
        GOOS="$OS" GOARCH="$ARCH" CGO_ENABLED=0 go build \
            -o "$OUTPUT_PATH" \
            -ldflags="-X main.Version=$VERSION" \
            ./webcore/main.go 2>&1
    fi

    
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
