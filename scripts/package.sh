#!/usr/bin/env bash

set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

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

    ###############################################################
    # Binary
    ###############################################################

    cp "$BUILD_PATH" "$RELEASE_DIR/$BIN"

    chmod +x "$RELEASE_DIR/$BIN" 2>/dev/null || true

    ###############################################################
    # Config
    ###############################################################

    cp "$ROOT_DIR/config.yaml.example" \
        "$RELEASE_DIR/config.yaml"

    cp "$ROOT_DIR/docker-compose.yml" \
        "$RELEASE_DIR/docker-compose.yml"

    cp "$ROOT_DIR/Dockerfile" \
        "$RELEASE_DIR/Dockerfile"

    ###############################################################
    # Migration
    ###############################################################

    mkdir -p "$RELEASE_DIR/webcore/init"

    cp -R \
        "$ROOT_DIR/webcore/init/migrations" \
        "$RELEASE_DIR/webcore/init"

    ###############################################################
    # Build Manifest
    ###############################################################

    create_manifest "$RELEASE_DIR" "$OS" "$ARCH"

    ###############################################################################
    # Archive
    ###############################################################################

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

    ###############################################################################
    # Cleanup
    ###############################################################################

    rm -rf "$RELEASE_DIR"

    success "${PLATFORM} packaged -> ${ARCHIVE}"
}

###############################################################################
# Main
###############################################################################

main() {

    info "Packaging Release"

    for dir in "$BUILD_DIR"/*; do

        [[ ! -d "$dir" ]] && continue

        package_one "$(basename "$dir")"

    done

    checksum

    success "All package created."

}

main