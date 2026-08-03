#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source "$SCRIPT_DIR/lib.sh"

VERSION="${1:-}"

if [[ -z "$VERSION" ]]; then
    die "Usage: release.sh <version>"
fi

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

banner

info "Starting Release ${VERSION}"

###############################################################################
# Build (includes environment check, workspace sync, package, checksum)
###############################################################################

bash "$SCRIPT_DIR/build.sh" "$VERSION"

verify_checksum

###############################################################################
# Release Notes
###############################################################################

CHANGELOG="$DIST_DIR/CHANGELOG.md"

cat > "$CHANGELOG" <<EOF
# ${APP_NAME} ${VERSION}

Release Date : ${BUILD_DATE}

Commit       : ${COMMIT}

Branch       : ${BRANCH}

Go           : ${GO_VERSION}

EOF

###############################################################################
# GitHub CLI
###############################################################################

if command -v gh >/dev/null 2>&1; then

    info "Publishing GitHub Release..."

    gh release create "$VERSION" \
        "$DIST_DIR"/* \
        --title "$VERSION" \
        --notes-file "$CHANGELOG"

else

    warn "GitHub CLI not installed."

fi

success "Release completed."
