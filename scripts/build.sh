#!/usr/bin/env bash

set -euo pipefail
while [[ $# -gt 0 ]]; do

    case "$1" in
        --all)
            BUILD_ALL=1
            shift
            ;;

        *)
            VERSION="$1"
            shift
            ;;
    esac

done

export VERSION
export BUILD_ALL

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source "$SCRIPT_DIR/lib.sh"

TARGET="${HOST_OS}/${HOST_ARCH}"

if [[ "${BUILD_ALL:-0}" == "1" ]]; then TARGET="--all"; fi


banner

info "Builder   : ${BUILDER}"
info "Version   : ${VERSION}"
info "Host      : ${HOST_OS}/${HOST_ARCH}"
info "Go        : ${GO_VERSION}"
info "Target    : ${TARGET}"

echo

check_environment
sync_workspace

echo
build_all

echo
package_all
echo
checksum

echo

success "==========================================="
success " Build completed successfully"
success "==========================================="

echo

find "$DIST_DIR" -maxdepth 1 -type f | while read file
do
    echo "  $(basename "$file")"
done

echo
