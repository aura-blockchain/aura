#!/bin/bash
# Test 01: Chain Status
# Verifies chain is running and producing blocks

RPC="${1:-https://testnet-rpc.aurablockchain.org}"

echo "Checking chain status..."

# Get status
STATUS=$(curl -s "$RPC/status")
if [ -z "$STATUS" ]; then
    echo "ERROR: Cannot reach RPC endpoint"
    exit 1
fi

# Check chain ID
CHAIN_ID=$(echo "$STATUS" | jq -r '.result.node_info.network')
echo "Chain ID: $CHAIN_ID"

# Check catching up
CATCHING_UP=$(echo "$STATUS" | jq -r '.result.sync_info.catching_up')
if [ "$CATCHING_UP" = "true" ]; then
    echo "WARNING: Node is still catching up"
fi

# Check latest block height
HEIGHT=$(echo "$STATUS" | jq -r '.result.sync_info.latest_block_height')
echo "Latest block height: $HEIGHT"

if [ "$HEIGHT" -gt 0 ]; then
    echo "Chain is producing blocks"
    exit 0
else
    echo "ERROR: Chain height is 0"
    exit 1
fi
