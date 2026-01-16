#!/bin/bash
# AURA MVP Genesis Distribution Script
# Distributes genesis file to all validators

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHAIN_DIR="$(dirname "$SCRIPT_DIR")"

# Configuration
GENESIS_FILE="${1:-$CHAIN_DIR/../testnets/aura-mvp-1/genesis.json}"
CHAIN_ID="${CHAIN_ID:-aura-mvp-1}"

# Validator servers and homes
declare -A VALIDATORS=(
    ["aura-testnet:val1"]="~/.aura-val1"
    ["aura-testnet:val2"]="~/.aura-val2"
    ["services-testnet:val3"]="~/.aura-val3"
    ["services-testnet:val4"]="~/.aura-val4"
)

echo "=========================================="
echo "AURA MVP Genesis Distribution"
echo "=========================================="
echo "Genesis file: $GENESIS_FILE"
echo "Chain ID: $CHAIN_ID"
echo "=========================================="

# Validate genesis file exists
if [ ! -f "$GENESIS_FILE" ]; then
    echo "ERROR: Genesis file not found: $GENESIS_FILE"
    exit 1
fi

# Calculate checksum
CHECKSUM=$(sha256sum "$GENESIS_FILE" | awk '{print $1}')
echo "Genesis checksum: $CHECKSUM"
echo

# Validate genesis
echo "Validating genesis file..."
if command -v aurad &> /dev/null; then
    aurad genesis validate "$GENESIS_FILE" || {
        echo "ERROR: Genesis validation failed"
        exit 1
    }
    echo "Genesis validation: OK"
else
    echo "WARNING: aurad not found locally, skipping validation"
fi
echo

# Distribute to each validator
for key in "${!VALIDATORS[@]}"; do
    IFS=':' read -r server val <<< "$key"
    home="${VALIDATORS[$key]}"

    echo "Distributing to $server ($val)..."

    # Copy genesis
    scp "$GENESIS_FILE" "$server:$home/config/genesis.json"

    # Verify checksum on remote
    REMOTE_CHECKSUM=$(ssh "$server" "sha256sum $home/config/genesis.json | awk '{print \$1}'" 2>/dev/null)

    if [ "$REMOTE_CHECKSUM" = "$CHECKSUM" ]; then
        echo "  $val: OK (checksum verified)"
    else
        echo "  $val: ERROR (checksum mismatch)"
        echo "    Expected: $CHECKSUM"
        echo "    Got: $REMOTE_CHECKSUM"
        exit 1
    fi
done

echo
echo "=========================================="
echo "Genesis Distribution Complete"
echo "=========================================="
echo "Checksum: $CHECKSUM"
echo
echo "Next steps:"
echo "1. Configure peers on each validator"
echo "2. Set genesis time (if not already set)"
echo "3. Run coordinated launch script"
