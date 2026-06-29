#!/usr/bin/env bash

set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

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
# Need CGO
###############################################################################

need_cgo() {
    echo 1
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
    # local -a build_args=()

    binary=$(binary_name "$os")

    info "Building ${os}/${arch}"

    export GOOS="$os"
    export GOARCH="$arch"
    export CGO_ENABLED="$(need_cgo)"

    if [[ -n "$compiler" ]]; then
        export CC="$compiler"
    else
        unset CC
    fi

    # if [[ "$os" == "windows" ]]; then
    #     build_args+=("-tags" "dynamic")
    # fi

    go build \
        # "${build_args[@]}" \
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