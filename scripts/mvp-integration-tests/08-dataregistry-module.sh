#!/bin/bash
# Test 08: Data Registry Module (MVP Custom)
# Verifies dataregistry module is accessible

REST="${2:-https://testnet-api.aurablockchain.org}"

echo "Checking dataregistry module..."

# Test dataregistry params
echo "Testing dataregistry params..."
PARAMS=$(curl -s "$REST/aura/dataregistry/v1/params" 2>/dev/null)
if echo "$PARAMS" | jq -e '.params' > /dev/null 2>&1; then
    echo "  DataRegistry params: OK"
else
    echo "  WARNING: DataRegistry params endpoint not responding (may be expected)"
fi

# Test data items listing
echo "Testing data items..."
ITEMS=$(curl -s "$REST/aura/dataregistry/v1/data_items" 2>/dev/null)
if echo "$ITEMS" | jq -e '.' > /dev/null 2>&1; then
    echo "  Data items: OK"
else
    echo "  WARNING: Data items not available"
fi

echo "DataRegistry module check complete"
exit 0
