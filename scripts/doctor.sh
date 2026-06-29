#!/usr/bin/env bash

set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

FAILED=0

check() {

    local name="$1"
    local cmd="$2"

    printf "%-30s" "$name"

    if command -v "$cmd" >/dev/null 2>&1; then
        echo -e "\033[32mOK\033[0m"
    else
        echo -e "\033[31mMISSING\033[0m"
        FAILED=1
    fi
}

check_pkg() {

    local pkg="$1"

    printf "%-30s" "$pkg"

    if pkg-config --exists "$pkg"; then
        echo -e "\033[32mOK\033[0m"
    else
        echo -e "\033[31mMISSING\033[0m"
        FAILED=1
    fi
}

banner

echo
info "Checking build environment..."
echo

check "Go" go
check "Git" git
check "pkg-config" pkg-config
check "GCC" gcc
check "Tar" tar
check "Zip" zip

case "$HOST_OS" in

linux)

    ;;

darwin)

    ;;

esac

echo

if [[ "$FAILED" == 1 ]]; then

    error "Environment not ready."

    exit 1

fi

success "Environment OK."