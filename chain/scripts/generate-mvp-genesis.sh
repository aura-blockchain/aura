#!/bin/bash
# Generate AURA MVP Genesis File
# Usage: ./generate-mvp-genesis.sh [chain-id] [output-file]
#
# This script generates a minimal MVP genesis with only the 12 essential modules.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHAIN_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="${CHAIN_DIR}/build"
TESTNETS_DIR="$(dirname "$(dirname "$CHAIN_DIR")")/testnets"

CHAIN_ID="${1:-aura-mvp-1}"
OUTPUT="${2:-genesis-mvp.json}"
TEMP_HOME=".mvp-genesis-temp"

echo "=== AURA MVP Genesis Generator ==="
echo "Chain ID: $CHAIN_ID"
echo "Output: $OUTPUT"
echo ""

# Check for MVP binary
if [ ! -f "${BUILD_DIR}/aurad-mvp" ]; then
    echo "MVP binary not found. Building..."
    cd "$CHAIN_DIR"
    make build-mvp
fi

BINARY="${BUILD_DIR}/aurad-mvp"
echo "Using binary: $BINARY"
echo ""

# Clean up any previous temp directory
rm -rf "$TEMP_HOME"

# Initialize chain with MVP binary
echo "1. Initializing chain..."
"$BINARY" init mvp-genesis-node --chain-id "$CHAIN_ID" --home "$TEMP_HOME" > /dev/null 2>&1

# Check if template exists
TEMPLATE="${TESTNETS_DIR}/aura-mvp-1/genesis-mvp-template.json"
if [ -f "$TEMPLATE" ]; then
    echo "2. Using MVP template: $TEMPLATE"

    # Update chain_id in template
    jq --arg chain_id "$CHAIN_ID" '.chain_id = $chain_id' "$TEMPLATE" > "${TEMP_HOME}/config/genesis.json"
else
    echo "2. No template found, using default genesis"
fi

# Update genesis time to now
GENESIS_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
jq --arg time "$GENESIS_TIME" '.genesis_time = $time' "${TEMP_HOME}/config/genesis.json" > "${TEMP_HOME}/config/genesis-updated.json"
mv "${TEMP_HOME}/config/genesis-updated.json" "${TEMP_HOME}/config/genesis.json"

echo "3. Genesis time set to: $GENESIS_TIME"

# Validate genesis
echo "4. Validating genesis..."
if "$BINARY" genesis validate "${TEMP_HOME}/config/genesis.json" --home "$TEMP_HOME" > /dev/null 2>&1; then
    echo "   Genesis validation PASSED"
else
    echo "   WARNING: Genesis validation failed, continuing anyway..."
fi

# Copy to output
cp "${TEMP_HOME}/config/genesis.json" "$OUTPUT"

# Show summary
echo ""
echo "=== Genesis Summary ==="
MODULES=$(jq '.app_state | keys | length' "$OUTPUT")
ACCOUNTS=$(jq '.app_state.auth.accounts | length' "$OUTPUT")
SUPPLY=$(jq -r '.app_state.bank.supply[0].amount // "0"' "$OUTPUT")

echo "Chain ID:    $CHAIN_ID"
echo "Modules:     $MODULES"
echo "Accounts:    $ACCOUNTS"
echo "Total Supply: $SUPPLY uaura"
echo ""
echo "Module list:"
jq -r '.app_state | keys[]' "$OUTPUT" | while read module; do
    echo "  - $module"
done

# Cleanup
rm -rf "$TEMP_HOME"

echo ""
echo "=== MVP Genesis Generated ==="
echo "Output: $OUTPUT"
echo ""
echo "To use:"
echo "  1. Add genesis accounts: aurad-mvp genesis add-genesis-account <address> <amount>uaura --home <home>"
echo "  2. Create gentx: aurad-mvp genesis gentx <key> <amount>uaura --chain-id $CHAIN_ID --home <home>"
echo "  3. Collect gentxs: aurad-mvp genesis collect-gentxs --home <home>"
