#!/bin/bash
# Test 04: REST API Endpoints
# Verifies REST API is accessible and responding

REST="${2:-https://testnet-api.aurablockchain.org}"

echo "Checking REST API endpoints..."

# Test node info
echo "Testing /cosmos/base/tendermint/v1beta1/node_info..."
NODE_INFO=$(curl -s "$REST/cosmos/base/tendermint/v1beta1/node_info")
if echo "$NODE_INFO" | jq -e '.default_node_info.network' > /dev/null 2>&1; then
    echo "  Node info: OK"
else
    echo "  Node info: FAIL"
    exit 1
fi

# Test latest block
echo "Testing /cosmos/base/tendermint/v1beta1/blocks/latest..."
BLOCK=$(curl -s "$REST/cosmos/base/tendermint/v1beta1/blocks/latest")
if echo "$BLOCK" | jq -e '.block.header.height' > /dev/null 2>&1; then
    HEIGHT=$(echo "$BLOCK" | jq -r '.block.header.height')
    echo "  Latest block: OK (height $HEIGHT)"
else
    echo "  Latest block: FAIL"
    exit 1
fi

# Test bank params
echo "Testing /cosmos/bank/v1beta1/params..."
BANK=$(curl -s "$REST/cosmos/bank/v1beta1/params")
if echo "$BANK" | jq -e '.params' > /dev/null 2>&1; then
    echo "  Bank params: OK"
else
    echo "  Bank params: FAIL"
    exit 1
fi

# Test staking params
echo "Testing /cosmos/staking/v1beta1/params..."
STAKING=$(curl -s "$REST/cosmos/staking/v1beta1/params")
if echo "$STAKING" | jq -e '.params.bond_denom' > /dev/null 2>&1; then
    DENOM=$(echo "$STAKING" | jq -r '.params.bond_denom')
    echo "  Staking params: OK (denom: $DENOM)"
else
    echo "  Staking params: FAIL"
    exit 1
fi

echo "All REST API endpoints OK"
exit 0
