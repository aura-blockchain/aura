#!/bin/bash
# Build AURA MVP Release Artifacts
# Usage: ./build-mvp-release.sh [version]
#
# Produces multi-platform binaries with checksums for MVP release.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHAIN_DIR="$(dirname "$SCRIPT_DIR")"

VERSION="${1:-v1.0.0-mvp}"
BUILD_DIR="${CHAIN_DIR}/release-${VERSION}"
LDFLAGS="-X main.Version=${VERSION} -X main.Commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

echo "=== AURA MVP Release Builder ==="
echo "Version: $VERSION"
echo "Output:  $BUILD_DIR"
echo ""

# Clean previous build
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

cd "$CHAIN_DIR"

# Build for each platform
echo "Building MVP release ${VERSION}..."
echo ""

# Linux AMD64
echo "1/4 Building linux/amd64..."
GOOS=linux GOARCH=amd64 go build -tags=mvp -ldflags="${LDFLAGS}" \
    -o "${BUILD_DIR}/aurad-${VERSION}-linux-amd64" ./cmd/aurad
echo "     ${BUILD_DIR}/aurad-${VERSION}-linux-amd64"

# Linux ARM64
echo "2/4 Building linux/arm64..."
GOOS=linux GOARCH=arm64 go build -tags=mvp -ldflags="${LDFLAGS}" \
    -o "${BUILD_DIR}/aurad-${VERSION}-linux-arm64" ./cmd/aurad
echo "     ${BUILD_DIR}/aurad-${VERSION}-linux-arm64"

# Darwin AMD64
echo "3/4 Building darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build -tags=mvp -ldflags="${LDFLAGS}" \
    -o "${BUILD_DIR}/aurad-${VERSION}-darwin-amd64" ./cmd/aurad
echo "     ${BUILD_DIR}/aurad-${VERSION}-darwin-amd64"

# Darwin ARM64 (Apple Silicon)
echo "4/4 Building darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build -tags=mvp -ldflags="${LDFLAGS}" \
    -o "${BUILD_DIR}/aurad-${VERSION}-darwin-arm64" ./cmd/aurad
echo "     ${BUILD_DIR}/aurad-${VERSION}-darwin-arm64"

echo ""

# Generate checksums
echo "Generating checksums..."
cd "$BUILD_DIR"
sha256sum aurad-* > SHA256SUMS
echo ""

# Show results
echo "=== Release Artifacts ==="
ls -lh aurad-* SHA256SUMS
echo ""
echo "Checksums:"
cat SHA256SUMS
echo ""

# Verify linux binary (if on linux)
if [ "$(uname)" = "Linux" ]; then
    echo "=== Binary Verification ==="
    LINUX_BIN="aurad-${VERSION}-linux-amd64"
    echo "Testing ${LINUX_BIN}..."
    chmod +x "${LINUX_BIN}"
    ./"${LINUX_BIN}" version 2>/dev/null || echo "(version command not available)"
    echo ""
fi

echo "=== MVP Release Complete ==="
echo "Artifacts: ${BUILD_DIR}/"
echo ""
echo "To upload to R2:"
echo "  wrangler r2 object put aura-testnet-artifacts/mvp/${VERSION}/aurad-${VERSION}-linux-amd64 --file ${BUILD_DIR}/aurad-${VERSION}-linux-amd64 --remote"
echo "  wrangler r2 object put aura-testnet-artifacts/mvp/${VERSION}/SHA256SUMS --file ${BUILD_DIR}/SHA256SUMS --remote"
