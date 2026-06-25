#!/bin/bash

# Mini-Proxy Release Checklist Helper
# Script ini membantu ensure semua step sebelum release sudah done

set -e

echo "╔════════════════════════════════════════════╗"
echo "║  Mini-Proxy Pre-Release Checklist          ║"
echo "╚════════════════════════════════════════════╝"
echo ""

# Function to check
check_item() {
    echo -n "▹ $1 ... "
    read -p "(y/n) " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "✓"
    else
        echo "✗ Please complete this before release"
        exit 1
    fi
}

# Checklist
check_item "Code changes complete and tested"
check_item "All tests passing (make test)"
check_item "Dependencies updated (go work sync)"
check_item "Version bumped in config.yaml and relevant files"
check_item "CHANGELOG updated"
check_item "README updated with new features/fixes"
check_item "All commits pushed to main/dev branch"

echo ""
echo "╔════════════════════════════════════════════╗"
echo "║  Ready to Release!                         ║"
echo "╚════════════════════════════════════════════╝"
echo ""

# Get version
read -p "Enter version (v1.0.0): " VERSION
VERSION=${VERSION:-v1.0.0}

echo ""
echo "Next steps:"
echo "1. Create and push tag:"
echo "   git tag $VERSION"
echo "   git push origin $VERSION"
echo ""
echo "2. GitHub Actions will automatically:"
echo "   - Build for all platforms"
echo "   - Create release packages"
echo "   - Upload to GitHub Releases"
echo ""
echo "3. Monitor release:"
echo "   - Check: https://github.com/YOUR_REPO/releases"
echo "   - Or run: gh release list"
echo ""
