#!/bin/bash
# tx-with-memory-keyring.sh
#
# Workaround script for Cosmos SDK 0.53.4 keyring migration bug.
# Uses memory keyring backend to recover a key and execute a transaction
# in a single shell session.
#
# Usage:
#   ./tx-with-memory-keyring.sh <key-name> <mnemonic-file> <tx-command...>
#
# Example:
#   ./tx-with-memory-keyring.sh deployer ./deployer.mnemonic tx bank send deployer aura1... 1000uaura --chain-id aura-local-1
#
# The mnemonic file should contain a 24-word recovery phrase (one line).

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AURAD="${SCRIPT_DIR}/../chain/aurad"

# Check if aurad exists in expected location, fallback to current dir
if [ ! -f "$AURAD" ]; then
    AURAD="./aurad"
fi

if [ ! -f "$AURAD" ]; then
    echo "Error: aurad binary not found"
    exit 1
fi

KEY_NAME="$1"
MNEMONIC_FILE="$2"
shift 2

if [ -z "$KEY_NAME" ] || [ -z "$MNEMONIC_FILE" ]; then
    echo "Usage: $0 <key-name> <mnemonic-file> <tx-command...>"
    echo ""
    echo "Example:"
    echo "  $0 deployer ./deployer.mnemonic tx bank send deployer aura1... 1000uaura"
    exit 1
fi

if [ ! -f "$MNEMONIC_FILE" ]; then
    echo "Error: Mnemonic file not found: $MNEMONIC_FILE"
    exit 1
fi

MNEMONIC=$(cat "$MNEMONIC_FILE")

if [ -z "$MNEMONIC" ]; then
    echo "Error: Mnemonic file is empty"
    exit 1
fi

# Count words in mnemonic
WORD_COUNT=$(echo "$MNEMONIC" | wc -w)
if [ "$WORD_COUNT" -ne 24 ] && [ "$WORD_COUNT" -ne 12 ]; then
    echo "Warning: Mnemonic has $WORD_COUNT words (expected 12 or 24)"
fi

echo "=== Recovering key '$KEY_NAME' in memory keyring ==="
echo "$MNEMONIC" | $AURAD keys add "$KEY_NAME" --recover --keyring-backend memory 2>&1 | head -5

echo ""
echo "=== Executing: aurad $@ ==="
$AURAD "$@" --from "$KEY_NAME" --keyring-backend memory

echo ""
echo "=== Done ==="
