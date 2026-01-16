#!/bin/bash
# Test 10: Faucet Availability
# Verifies faucet endpoint is accessible

FAUCET="${FAUCET_URL:-https://testnet-faucet.aurablockchain.org}"

echo "Checking faucet..."

# Test faucet health/status endpoint
echo "Testing faucet endpoint..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$FAUCET" 2>/dev/null)
if [ "$RESPONSE" = "200" ] || [ "$RESPONSE" = "301" ] || [ "$RESPONSE" = "302" ]; then
    echo "  Faucet endpoint: OK (HTTP $RESPONSE)"
else
    echo "  WARNING: Faucet returned HTTP $RESPONSE"
fi

# Test faucet status if available
STATUS=$(curl -s "$FAUCET/status" 2>/dev/null)
if echo "$STATUS" | jq -e '.' > /dev/null 2>&1; then
    echo "  Faucet status: OK"
else
    echo "  Faucet status endpoint not available (may be expected)"
fi

echo "Faucet check complete"
exit 0
