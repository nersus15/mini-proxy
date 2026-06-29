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
CYAN='\033[0;36m'
GRAY='\033[0;90m'
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

BUILD_HOST="$(hostname)"


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

