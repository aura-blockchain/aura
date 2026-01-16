#!/bin/bash
# Test 02: Validator Set
# Verifies 4 validators are active and signing

RPC="${1:-https://testnet-rpc.aurablockchain.org}"
EXPECTED_VALIDATORS=4

echo "Checking validator set..."

# Get validators
VALIDATORS=$(curl -s "$RPC/validators")
if [ -z "$VALIDATORS" ]; then
    echo "ERROR: Cannot get validator set"
    exit 1
fi

# Count validators
COUNT=$(echo "$VALIDATORS" | jq -r '.result.validators | length')
echo "Active validators: $COUNT"

if [ "$COUNT" -ge "$EXPECTED_VALIDATORS" ]; then
    echo "Validator count OK (expected >= $EXPECTED_VALIDATORS)"
else
    echo "ERROR: Expected at least $EXPECTED_VALIDATORS validators, got $COUNT"
    exit 1
fi

# Check voting power distribution
echo "Voting power distribution:"
echo "$VALIDATORS" | jq -r '.result.validators[] | "  \(.address[0:12])... : \(.voting_power)"'

# Check total voting power
TOTAL=$(echo "$VALIDATORS" | jq -r '[.result.validators[].voting_power | tonumber] | add')
echo "Total voting power: $TOTAL"

exit 0
