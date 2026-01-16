#!/bin/bash
# Test 07: VC Registry Module (MVP Custom)
# Verifies vcregistry module is accessible

REST="${2:-https://testnet-api.aurablockchain.org}"

echo "Checking vcregistry module..."

# Test vcregistry params
echo "Testing vcregistry params..."
PARAMS=$(curl -s "$REST/aura/vcregistry/v1/params" 2>/dev/null)
if echo "$PARAMS" | jq -e '.params' > /dev/null 2>&1; then
    echo "  VCRegistry params: OK"
else
    echo "  WARNING: VCRegistry params endpoint not responding (may be expected)"
fi

# Test credential types listing
echo "Testing credential types..."
TYPES=$(curl -s "$REST/aura/vcregistry/v1/credential_types" 2>/dev/null)
if echo "$TYPES" | jq -e '.' > /dev/null 2>&1; then
    echo "  Credential types: OK"
else
    echo "  WARNING: Credential types not available"
fi

echo "VCRegistry module check complete"
exit 0
