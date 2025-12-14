#!/bin/bash
#
# Verification script for Aura Wallet .deb package
# This script verifies that the .deb package is properly built and contains all required files
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEB_FILE="$SCRIPT_DIR/dist/aura-desktop-wallet_1.0.0_amd64.deb"

echo "================================================"
echo "Aura Wallet .deb Package Verification"
echo "================================================"
echo

# Check if .deb file exists
if [ ! -f "$DEB_FILE" ]; then
    echo "❌ ERROR: .deb package not found at: $DEB_FILE"
    echo "   Run 'npm run build:linux' to build the package"
    exit 1
fi

echo "✅ Package file exists: $DEB_FILE"
echo

# Check file size
FILE_SIZE=$(du -h "$DEB_FILE" | cut -f1)
echo "📦 Package size: $FILE_SIZE"
echo

# Display package information
echo "📋 Package Information:"
echo "---"
dpkg --info "$DEB_FILE" | grep -E "Package:|Version:|Architecture:|Maintainer:|Depends:|Description:" | sed 's/^/ /'
echo

# Check for required files in package
echo "🔍 Verifying package contents..."
echo

REQUIRED_FILES=(
    "./opt/Aura Wallet/aura-desktop-wallet"
    "./usr/share/applications/aura-desktop-wallet.desktop"
    "./usr/share/icons/hicolor/16x16/apps/aura-desktop-wallet.png"
    "./usr/share/icons/hicolor/32x32/apps/aura-desktop-wallet.png"
    "./usr/share/icons/hicolor/48x48/apps/aura-desktop-wallet.png"
    "./usr/share/icons/hicolor/64x64/apps/aura-desktop-wallet.png"
    "./usr/share/icons/hicolor/128x128/apps/aura-desktop-wallet.png"
    "./usr/share/icons/hicolor/256x256/apps/aura-desktop-wallet.png"
    "./usr/share/icons/hicolor/512x512/apps/aura-desktop-wallet.png"
)

MISSING_FILES=0

for file in "${REQUIRED_FILES[@]}"; do
    if dpkg --contents "$DEB_FILE" | grep -q "$file"; then
        echo "  ✅ $file"
    else
        echo "  ❌ MISSING: $file"
        MISSING_FILES=$((MISSING_FILES + 1))
    fi
done

echo

# Check for post-install/remove scripts
echo "📜 Checking maintenance scripts..."
if dpkg --info "$DEB_FILE" 2>&1 | grep -q "postinst"; then
    echo "  ✅ Post-install script present"
else
    echo "  ❌ Post-install script missing"
    MISSING_FILES=$((MISSING_FILES + 1))
fi

if dpkg --info "$DEB_FILE" 2>&1 | grep -q "postrm"; then
    echo "  ✅ Post-remove script present"
else
    echo "  ❌ Post-remove script missing"
    MISSING_FILES=$((MISSING_FILES + 1))
fi

echo

# Check build assets
echo "🎨 Checking build assets..."
BUILD_ASSETS=(
    "build/aura-wallet.desktop"
    "build/deb-postinstall.sh"
    "build/deb-postremove.sh"
    "build/icon.png"
    "build/icons/16x16.png"
    "build/icons/32x32.png"
    "build/icons/48x48.png"
    "build/icons/64x64.png"
    "build/icons/128x128.png"
    "build/icons/256x256.png"
    "build/icons/512x512.png"
)

for asset in "${BUILD_ASSETS[@]}"; do
    if [ -f "$SCRIPT_DIR/$asset" ]; then
        echo "  ✅ $asset"
    else
        echo "  ❌ MISSING: $asset"
        MISSING_FILES=$((MISSING_FILES + 1))
    fi
done

echo

# Final status
echo "================================================"
if [ $MISSING_FILES -eq 0 ]; then
    echo "✅ VERIFICATION PASSED"
    echo "   All required files are present in the package"
    echo
    echo "Package is ready for distribution!"
    echo
    echo "To install:"
    echo "  sudo dpkg -i $DEB_FILE"
    echo "  sudo apt-get install -f  # If dependencies are missing"
    echo
    echo "To test:"
    echo "  aura-wallet  # After installation"
    exit 0
else
    echo "❌ VERIFICATION FAILED"
    echo "   $MISSING_FILES file(s) missing or incorrect"
    echo
    echo "Please rebuild the package:"
    echo "  npm run build:linux"
    exit 1
fi
