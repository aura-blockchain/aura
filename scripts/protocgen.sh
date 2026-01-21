#!/usr/bin/env bash
# Proto generation script for AURA blockchain
# Generates both legacy (gocosmos) and modern (pulsar) proto types

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== AURA Proto Generation ==="
echo "Root: $ROOT_DIR"

cd "$ROOT_DIR/proto"

# Update buf dependencies
echo "Updating buf dependencies..."
buf mod update

# Clean previous generated files
echo "Cleaning old generated files..."
rm -rf "$ROOT_DIR/api"
mkdir -p "$ROOT_DIR/api"

# Generate with gocosmos (legacy compatibility)
echo "Generating with gocosmos (legacy)..."
buf generate --template buf.gen.gocosmos.yaml

# Generate with pulsar (modern API types)
echo "Generating with pulsar (modern)..."
buf generate --template buf.gen.pulsar.yaml

cd "$ROOT_DIR"

# Move gocosmos generated files to correct locations
if [ -d "github.com" ]; then
    echo "Moving gocosmos generated files..."
    cp -r github.com/aequitas/aura/* ./
    rm -rf github.com
fi

# Add build tags to pulsar files for conditional compilation
echo "Adding build tags to pulsar files..."
find ./api -name "*.pulsar.go" 2>/dev/null | while read -r file; do
    if ! head -1 "$file" | grep -q "//go:build pulsar"; then
        # Create temp file with build tag
        echo "//go:build pulsar" > "$file.tmp"
        echo "" >> "$file.tmp"
        cat "$file" >> "$file.tmp"
        mv "$file.tmp" "$file"
    fi
done

# Add build tags to grpc-gateway files
find . -name "*.pb.gw.go" 2>/dev/null | while read -r file; do
    if ! head -1 "$file" | grep -q "//go:build"; then
        echo "//go:build !pulsar" > "$file.tmp"
        echo "" >> "$file.tmp"
        cat "$file" >> "$file.tmp"
        mv "$file.tmp" "$file"
    fi
done

# Run go mod tidy
echo "Running go mod tidy..."
cd "$ROOT_DIR/chain"
go mod tidy

echo "=== Proto generation complete ==="
echo ""
echo "Generated files:"
echo "  - Legacy (gocosmos): chain/x/*/types/*.pb.go"
echo "  - Modern (pulsar): api/aura/*/*.pulsar.go"
echo "  - gRPC Gateway: proto/aura/*/*.pb.gw.go"
