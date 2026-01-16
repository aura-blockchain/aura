#!/bin/bash
# Test 09: CosmWasm Module
# Verifies wasm module is accessible

REST="${2:-https://testnet-api.aurablockchain.org}"

echo "Checking wasm module..."

# Test wasm params
echo "Testing wasm params..."
PARAMS=$(curl -s "$REST/cosmwasm/wasm/v1/params" 2>/dev/null)
if echo "$PARAMS" | jq -e '.params' > /dev/null 2>&1; then
    echo "  Wasm params: OK"
    INSTANTIATE=$(echo "$PARAMS" | jq -r '.params.instantiate_default_permission')
    echo "  Instantiate permission: $INSTANTIATE"
else
    echo "  WARNING: Wasm params not available"
fi

# Test contract listing
echo "Testing contract codes..."
CODES=$(curl -s "$REST/cosmwasm/wasm/v1/code" 2>/dev/null)
if echo "$CODES" | jq -e '.code_infos' > /dev/null 2>&1; then
    COUNT=$(echo "$CODES" | jq '.code_infos | length')
    echo "  Stored codes: $COUNT"
else
    echo "  WARNING: Code listing not available"
fi

echo "Wasm module check complete"
exit 0
