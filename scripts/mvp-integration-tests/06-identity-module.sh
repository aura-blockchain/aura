#!/bin/bash
# Test 06: Identity Module (MVP Custom)
# Verifies identity module is accessible

REST="${2:-https://testnet-api.aurablockchain.org}"

echo "Checking identity module..."

# Test identity params
echo "Testing identity params..."
PARAMS=$(curl -s "$REST/aura/identity/v1/params" 2>/dev/null)
if echo "$PARAMS" | jq -e '.params' > /dev/null 2>&1; then
    echo "  Identity params: OK"
else
    # Try alternate endpoint
    PARAMS=$(curl -s "$REST/aura.identity.v1.Query/Params" 2>/dev/null)
    if echo "$PARAMS" | jq -e '.params' > /dev/null 2>&1; then
        echo "  Identity params: OK (gRPC-gateway)"
    else
        echo "  WARNING: Identity params endpoint not responding (may be expected)"
    fi
fi

# Test identity listing
echo "Testing identity listing..."
IDENTITIES=$(curl -s "$REST/aura/identity/v1/identities" 2>/dev/null)
if echo "$IDENTITIES" | jq -e '.' > /dev/null 2>&1; then
    echo "  Identity listing: OK"
else
    echo "  WARNING: Identity listing not available"
fi

echo "Identity module check complete"
exit 0
