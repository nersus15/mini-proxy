#!/usr/bin/env bash

set -euo pipefail

###############################################################################
# Root
###############################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

source "$SCRIPT_DIR/build.conf"

###############################################################################
# Color
###############################################################################

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

###############################################################################
# Logger
###############################################################################

log() {
    printf "%b%s%b\n" "$1" "$2" "$NC"
}

info() {
    log "$BLUE" "[INFO] $*"
}

warn() {
    log "$YELLOW" "[WARN] $*"
}

success() {
    log "$GREEN" "[ OK ] $*"
}

error() {
    log "$RED" "[FAIL] $*"
}

die() {
    error "$*"
    exit 1
}

###############################################################################
# Banner
###############################################################################

banner() {

cat <<EOF

=========================================================
                 MINI PROXY BUILD SYSTEM
=========================================================

Application : ${APP_NAME}

EOF

}

###############################################################################
# Directory
###############################################################################

BUILD_DIR="$ROOT_DIR/build"

DIST_DIR="$ROOT_DIR/dist"

mkdir -p "$BUILD_DIR"

mkdir -p "$DIST_DIR"

###############################################################################
# Metadata
###############################################################################

VERSION="${VERSION:-dev}"

COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"

BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)"

TAG="$(git describe --tags --always 2>/dev/null || echo dev)"

BUILD_DATE="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

GO_VERSION="$(go version | awk '{print $3}')"

BUILD_HOST="$(hostname 2>/dev/null || uname -n)"

###############################################################################
# Host Platform
###############################################################################

case "$(uname -s)" in

    Linux*)
        HOST_OS="linux"
        ;;

    Darwin*)
        HOST_OS="darwin"
        ;;

    MINGW*|MSYS*|CYGWIN*)
        HOST_OS="windows"
        ;;

    *)
        HOST_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
        ;;

esac

case "$(uname -m)" in

    x86_64|amd64)
        HOST_ARCH="amd64"
        ;;

    arm64|aarch64)
        HOST_ARCH="arm64"
        ;;

    *)
        HOST_ARCH="$(uname -m)"
        ;;

esac

###############################################################################
# Builder
###############################################################################

if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
    BUILDER="GitHub Actions"
elif [[ -n "${CI:-}" ]]; then
    BUILDER="CI"
else
    BUILDER="${USER:-local}"
fi

###############################################################################
# Tool Detection
###############################################################################

exists() {

    command -v "$1" >/dev/null 2>&1

}

require() {

    exists "$1" || die "$1 not installed"

}

###############################################################################
# Environment Check ("doctor")
###############################################################################

check_environment() {

    info "Checking build environment..."
    echo

    local failed=0

    local checks=(
        "Go:go"
        "Git:git"
        "pkg-config:pkg-config"
        "GCC:gcc"
        "Tar:tar"
        "Zip:zip"
    )

    local entry name cmd
    for entry in "${checks[@]}"; do

        name="${entry%%:*}"
        cmd="${entry#*:}"

        printf "%-30s" "$name"

        if exists "$cmd"; then
            echo -e "${GREEN}OK${NC}"
        else
            echo -e "${RED}MISSING${NC}"
            failed=1
        fi

    done

    echo

    if [[ "$failed" == 1 ]]; then
        die "Environment not ready."
    fi

    success "Environment OK."
    echo
}

###############################################################################
# Workspace
###############################################################################

sync_workspace() {

    info "Syncing Go Workspace..."

    go work sync

    success "Workspace synced."

}

###############################################################################
# Binary
###############################################################################

binary_name() {

    case "$1" in

        windows)

            echo "${OUTPUT_NAME}.exe"

            ;;

        *)

            echo "${OUTPUT_NAME}"

            ;;

    esac

}

###############################################################################
# Output Folder
###############################################################################

output_dir() {

    echo "$BUILD_DIR/$1-$2"

}

###############################################################################
# Build Metadata
###############################################################################

ldflags() {

cat <<EOF
-s -w \
-X '${VERSION_PACKAGE}.AppName=${APP_NAME}' \
-X '${VERSION_PACKAGE}.Version=${VERSION}' \
-X '${VERSION_PACKAGE}.Commit=${COMMIT}' \
-X '${VERSION_PACKAGE}.Branch=${BRANCH}' \
-X '${VERSION_PACKAGE}.Tag=${TAG}' \
-X '${VERSION_PACKAGE}.BuildDate=${BUILD_DATE}' \
-X '${VERSION_PACKAGE}.Builder=${BUILDER}' \
-X '${VERSION_PACKAGE}.Host=${BUILD_HOST}' \
-X '${VERSION_PACKAGE}.GoVersion=${GO_VERSION}' \
-X '${VERSION_PACKAGE}.Platform=$1' \
-X '${VERSION_PACKAGE}.Architecture=$2'
EOF

}

###############################################################################
# Build Helper
###############################################################################

verify_binary() {

    if [[ ! -f "$1" ]]; then

        die "Binary not found: $1"

    fi

    SIZE=$(du -h "$1" | cut -f1)

    success "$(basename "$1") ($SIZE)"

}

###############################################################################
# SHA256
###############################################################################

checksum() {

    info "Generating SHA256SUMS..."

    pushd "$DIST_DIR" >/dev/null

    rm -f SHA256SUMS

    shopt -s nullglob

    local files=()

    files+=( *.tar.gz )
    files+=( *.zip )

    if [[ ${#files[@]} -eq 0 ]]; then
        warn "No release archive found."
        popd >/dev/null
        return
    fi

    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "${files[@]}" > SHA256SUMS
    else
        shasum -a 256 "${files[@]}" > SHA256SUMS
    fi

    popd >/dev/null

    success "SHA256SUMS created."
}

###############################################################################
# Verify SHA256
###############################################################################

verify_checksum() {

    info "Verifying SHA256SUMS..."

    pushd "$DIST_DIR" >/dev/null

    if [[ ! -f SHA256SUMS ]]; then
        die "SHA256SUMS not found."
    fi

    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum -c SHA256SUMS
    else
        shasum -a 256 -c SHA256SUMS
    fi

    popd >/dev/null

    success "Checksum verification passed."
}

###############################################################################
# Manifest
###############################################################################

create_manifest() {

local DIR="$1"
local OS="$2"
local ARCH="$3"

cat > "$DIR/build.json" <<EOF
{
    "application":"${APP_NAME}",
    "version":"${VERSION}",
    "commit":"${COMMIT}",
    "branch":"${BRANCH}",
    "tag":"${TAG}",
    "go":"${GO_VERSION}",
    "builder":"${BUILDER}",
    "host": "${BUILD_HOST}",
    "built_at":"${BUILD_DATE}",
    "platform":"${OS}",
    "architecture":"${ARCH}"
}
EOF

}

###############################################################################
# Compiler Mapping
###############################################################################

compiler_for() {

    local os="$1"
    local arch="$2"

    case "${os}/${arch}" in

        linux/amd64)
            echo gcc
            ;;

        linux/arm64)
            echo aarch64-linux-gnu-gcc
            ;;

        windows/amd64)
            echo ""
            ;;

        windows/arm64)
            echo ""
            ;;

        darwin/*)

            echo clang

            ;;

        *)

            echo ""

            ;;

    esac

}

###############################################################################
# Check Compiler
###############################################################################

check_compiler() {

    local cc="$1"

    [[ -z "$cc" ]] && return 1

    command -v "$cc" >/dev/null 2>&1

}

###############################################################################
# Build Platform
###############################################################################

build_platform() {

    local os="$1"
    local arch="$2"

    local compiler
    compiler=$(compiler_for "$os" "$arch")

    if [[ -n "$compiler" ]]; then

        if ! check_compiler "$compiler"; then

            warn "Skip ${os}/${arch}"
            warn "Compiler ${compiler} not found"

            return 0

        fi

    fi

    local output
    output="$(output_dir "$os" "$arch")"

    mkdir -p "$output"

    local binary
    binary=$(binary_name "$os")

    info "Building ${os}/${arch}"

    export GOOS="$os"
    export GOARCH="$arch"
    export CGO_ENABLED=1

    if [[ -n "$compiler" ]]; then
        export CC="$compiler"
    else
        unset CC
    fi

    go build \
        -trimpath \
        -buildvcs=false \
        -ldflags="$(ldflags "$os" "$arch")" \
        -o "$output/$binary" \
        "$MAIN_PACKAGE"

    verify_binary "$output/$binary"

}

###############################################################################
# Build Host
###############################################################################

build_host() {

    info "Building native host..."

    case "${HOST_OS}" in

        linux)

            build_platform linux amd64

            if command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then
                build_platform linux arm64
            fi

            ;;

        darwin)

            if [[ "${HOST_ARCH}" == "arm64" ]]; then
                build_platform darwin arm64
            else
                build_platform darwin amd64
            fi

            ;;
        windows)
            if [[ "${HOST_ARCH}" == "arm64" ]]; then
                build_platform windows arm64
            else
                build_platform windows amd64
            fi
            ;;
        *)

            die "Unsupported host: ${HOST_OS}"

            ;;

    esac

}

###############################################################################
# Build All Targets
###############################################################################

build_targets() {

    info "Building configured targets..."

    for target in "${PLATFORMS[@]}"; do

        IFS="/" read -r os arch <<< "$target"

        build_platform "$os" "$arch"

    done

}

###############################################################################
# Entry
###############################################################################

build_all() {

    if [[ "${BUILD_ALL:-0}" == "1" ]]; then

        build_targets

    else

        build_host

    fi

}

###############################################################################
# Package One
###############################################################################

package_one() {

    local PLATFORM="$1"

    local OS="${PLATFORM%%-*}"
    local ARCH="${PLATFORM#*-}"

    local BIN
    BIN=$(binary_name "$OS")

    local BUILD_PATH="$BUILD_DIR/$PLATFORM/$BIN"

    [[ ! -f "$BUILD_PATH" ]] && return

    info "Packaging ${PLATFORM}"

    local RELEASE_DIR
    RELEASE_DIR="$DIST_DIR/${RELEASE_PREFIX}-${VERSION}-${OS}-${ARCH}"

    rm -rf "$RELEASE_DIR"

    mkdir -p "$RELEASE_DIR/db"

    ###########################################################################
    # Binary
    ###########################################################################

    cp "$BUILD_PATH" "$RELEASE_DIR/$BIN"

    chmod +x "$RELEASE_DIR/$BIN" 2>/dev/null || true

    ###########################################################################
    # Config
    ###########################################################################

    cp "$ROOT_DIR/config.yaml.example" \
        "$RELEASE_DIR/config.yaml"

    cp "$ROOT_DIR/docker-compose.yml" \
        "$RELEASE_DIR/docker-compose.yml"

    cp "$ROOT_DIR/Dockerfile" \
        "$RELEASE_DIR/Dockerfile"

    ###########################################################################
    # Migration
    ###########################################################################

    mkdir -p "$RELEASE_DIR/webcore/init"

    cp -R \
        "$ROOT_DIR/webcore/init/migrations" \
        "$RELEASE_DIR/webcore/init"

    ###########################################################################
    # Build Manifest
    ###########################################################################

    create_manifest "$RELEASE_DIR" "$OS" "$ARCH"

    ###########################################################################
    # Archive
    ###########################################################################

    pushd "$DIST_DIR" >/dev/null

    if [[ "$OS" == "windows" ]]; then

        zip -qr \
            "${RELEASE_PREFIX}-${VERSION}-${OS}-${ARCH}.zip" \
            "$(basename "$RELEASE_DIR")"

        ARCHIVE="${RELEASE_PREFIX}-${VERSION}-${OS}-${ARCH}.zip"

    else

        tar -czf \
            "${RELEASE_PREFIX}-${VERSION}-${OS}-${ARCH}.tar.gz" \
            "$(basename "$RELEASE_DIR")"

        ARCHIVE="${RELEASE_PREFIX}-${VERSION}-${OS}-${ARCH}.tar.gz"

    fi

    popd >/dev/null

    rm -rf "$RELEASE_DIR"

    success "${PLATFORM} packaged -> ${ARCHIVE}"
}

###############################################################################
# Package All
###############################################################################

package_all() {

    info "Packaging Release"

    for dir in "$BUILD_DIR"/*; do

        [[ ! -d "$dir" ]] && continue

        package_one "$(basename "$dir")"

    done

    success "All package created."

}
