#!/bin/bash
# Fast test runner for AURA blockchain - optimized for AI agents with timeouts

set -e

echo "=================================================="
echo "AURA Fast Test Runner"
echo "=================================================="
echo ""

SHORT_TEST_PACKAGES=(./app ./x/vcregistry)
SHORT_TEST_PARALLEL=4
SHORT_TEST_TIMEOUT=8m

run_short_suite() {
    echo "Running SHORT tests only (focused packages, ~5 mins)"
    for pkg in "${SHORT_TEST_PACKAGES[@]}"; do
        echo "  -> $pkg"
        go test -v -short -parallel=${SHORT_TEST_PARALLEL} -timeout=${SHORT_TEST_TIMEOUT} "$pkg"
    done
}

# Check for --short flag
if [ "$1" = "--short" ] || [ "$1" = "-s" ]; then
    run_short_suite
    exit 0
fi

# Check if specific package requested
if [ -n "$1" ]; then
    echo "Testing specific package: $1"
    go test -v -parallel=4 -timeout=10m $1
    exit 0
fi

# Default: run optimized parallel tests
echo "Running OPTIMIZED tests with 8 parallel workers"
echo "Estimated time: 10-15 minutes"
echo ""

# Show progress
go test -v -parallel=8 -timeout=20m -count=1 ./...

echo ""
echo "=================================================="
echo "Tests completed!"
echo "=================================================="
