#!/bin/bash
# Test 03: Consensus Health
# Verifies consensus is functioning correctly

RPC="${1:-https://testnet-rpc.aurablockchain.org}"

echo "Checking consensus health..."

# Get consensus state
CONSENSUS=$(curl -s "$RPC/consensus_state")
if [ -z "$CONSENSUS" ]; then
    echo "ERROR: Cannot get consensus state"
    exit 1
fi

# Check prevotes
PREVOTES=$(echo "$CONSENSUS" | jq -r '.result.round_state.votes[0].prevotes_bit_array' 2>/dev/null)
echo "Prevotes: $PREVOTES"

# Get two blocks to verify chain is advancing
HEIGHT1=$(curl -s "$RPC/status" | jq -r '.result.sync_info.latest_block_height')
sleep 6  # Wait for ~1 block
HEIGHT2=$(curl -s "$RPC/status" | jq -r '.result.sync_info.latest_block_height')

echo "Height check: $HEIGHT1 -> $HEIGHT2"

if [ "$HEIGHT2" -gt "$HEIGHT1" ]; then
    echo "Chain is advancing"
    exit 0
else
    echo "WARNING: Chain may be stalled"
    exit 1
fi
