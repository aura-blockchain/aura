#!/bin/bash

# Complete Identity Module Proto Migration Script
# This script automates the remaining JSON to Codec conversions

set -e

KEEPER_DIR="/home/decri/blockchain-projects/aura/chain/x/identity/keeper"

cd "$KEEPER_DIR"

echo "Starting identity module proto migration..."

# Step 1: Remove encoding/json imports
echo "Step 1: Removing encoding/json imports..."
sed -i '/^[[:space:]]*"encoding\/json"$/d' auth.go changes.go sessions.go

# Step 2: Convert JSON marshal calls to codec marshal
echo "Step 2: Converting json.Marshal to k.cdc.Marshal..."

# For already-pointer types (need to keep as-is)
# For value types (need to add &)
# We'll do a conservative replacement

for file in auth.go changes.go sessions.go; do
    echo "Processing $file..."

    # Replace json.Marshal with k.cdc.Marshal
    # This is a conservative replacement - may need manual adjustment
    sed -i 's/json\.Marshal(/k.cdc.Marshal(/g' "$file"

    # Replace json.Unmarshal with k.cdc.Unmarshal
    sed -i 's/json\.Unmarshal(/k.cdc.Unmarshal(/g' "$file"
done

echo "Step 3: Manual adjustments needed..."
echo ""
echo "The following manual changes are still required:"
echo ""
echo "1. Update function signatures to use pointers:"
echo "   - Change return types from 'Type' to '*Type'"
echo "   - Change parameters from 'Type' to '*Type'"
echo ""
echo "2. Update struct instantiations to use pointers:"
echo "   - Change 'types.Type{...}' to '&types.Type{...}'"
echo ""
echo "3. Update field names to match proto conventions:"
echo "   - DID → Did"
echo "   - RequestID → RequestId"
echo "   - SessionID → SessionId"
echo "   - ID → Id"
echo ""
echo "4. Update timestamp handling:"
echo "   - Change 'time.Time' to '*timestamppb.Timestamp'"
echo "   - Use '&now' instead of 'now' for timestamp fields"
echo ""
echo "5. Update params access:"
echo "   - params.MaxRolesPerAddress → params.Auth.MaxRolesPerAccount"
echo "   - params.MaxChangeRequestsPerMonth → params.Change.MaxRequestsPerWalletPerMonth"
echo ""
echo "6. Add nil checks for proto pointer fields"
echo ""
echo "7. Ensure k.cdc.Marshal calls use pointers:"
echo "   - If variable is already pointer: k.cdc.Marshal(ptr)"
echo "   - If variable is value: k.cdc.Marshal(&value)"
echo ""

echo "Conversion complete! Please review and make manual adjustments as needed."
echo ""
echo "To test compilation:"
echo "cd /home/decri/blockchain-projects/aura/chain"
echo "go build ./x/identity/..."
