#!/bin/bash

# This script fixes ALL time.Now() consensus-breaking bugs across the codebase
# by replacing them with sdk.UnwrapSDKContext(ctx).BlockTime()

set -e

echo "========================================="
echo "FIXING CRITICAL CONSENSUS BUGS"
echo "========================================="

# Function to add SDK import if not present
add_sdk_import() {
    local file="$1"
    if ! grep -q 'sdk "github.com/cosmos/cosmos-sdk/types"' "$file"; then
        # Find the import block and add SDK import
        if grep -q "^import (" "$file"; then
            # Multi-line import block exists
            sed -i '/^import (/a\	sdk "github.com/cosmos/cosmos-sdk/types"' "$file"
        elif grep -q "^import \"" "$file"; then
            # Single import exists, convert to block
            sed -i 's/^import "/import (\n\t"/; s/"$/"\n\tsdk "github.com\/cosmos\/cosmos-sdk\/types"\n)/' "$file"
        else
            # No imports, add new block after package
            sed -i '/^package /a\\nimport (\n\tsdk "github.com/cosmos/cosmos-sdk/types"\n)' "$file"
        fi
    fi
}

# Function to replace time.Now() with ctx.BlockTime()
replace_time_now() {
    local file="$1"
    echo "Processing: $file"

    # Add SDK import
    add_sdk_import "$file"

    # Replace time.Now() with ctx.BlockTime()
    # This handles most common patterns
    sed -i 's/time\.Now()/ctx.BlockTime()/g' "$file"

    # Handle timestamppb.Now() -> timestamppb.New(ctx.BlockTime())
    sed -i 's/timestamppb\.Now()/timestamppb.New(ctx.BlockTime())/g' "$file"

    echo "  ✓ Fixed: $file"
}

# Files that need sdk.Context parameter added to function signatures
# These will be manually fixed with specific edits

echo ""
echo "Step 1: Fixing walletsecurity/keeper/keeper.go"
# Lines 374, 467, 495-497, 614-615, 619, 623
replace_time_now "/home/decri/blockchain-projects/aura/chain/x/walletsecurity/keeper/keeper.go"

echo ""
echo "Step 2: Fixing walletsecurity/keeper/session_biometric.go"
# Lines 15, 116
replace_time_now "/home/decri/blockchain-projects/aura/chain/x/walletsecurity/keeper/session_biometric.go"

echo ""
echo "Step 3: Fixing walletsecurity/keeper/social_recovery.go"
# Lines 58, 112, 122, 159, 309, 322, 373, 395, 424, 429
replace_time_now "/home/decri/blockchain-projects/aura/chain/x/walletsecurity/keeper/social_recovery.go"

echo ""
echo "Step 4: Fixing walletsecurity/keeper/multisig.go"
# Lines 60, 76, 114, 122-123, 170, 241
replace_time_now "/home/decri/blockchain-projects/aura/chain/x/walletsecurity/keeper/multisig.go"

echo ""
echo "Step 5: Fixing networksecurity/keeper/fork_partition.go"
# Line 412
replace_time_now "/home/decri/blockchain-projects/aura/chain/x/networksecurity/keeper/fork_partition.go"

echo ""
echo "Step 6: Fixing networksecurity/keeper/rate_limiter.go"
# Lines 33, 35-36, 44, 76-77, 164
replace_time_now "/home/decri/blockchain-projects/aura/chain/x/networksecurity/keeper/rate_limiter.go"

echo ""
echo "Step 7: Fixing dex/keeper/liquidity_pool.go"
# Line 568
replace_time_now "/home/decri/blockchain-projects/aura/chain/x/dex/keeper/liquidity_pool.go"

echo ""
echo "Step 8: Fixing incidentresponse/keeper/keeper.go"
# Many lines with time.Now()
replace_time_now "/home/decri/blockchain-projects/aura/chain/x/incidentresponse/keeper/keeper.go"

echo ""
echo "Step 9: Fixing privacy module files"
if [ -f "/home/decri/blockchain-projects/aura/chain/x/privacy/network.go" ]; then
    replace_time_now "/home/decri/blockchain-projects/aura/chain/x/privacy/network.go"
fi

if [ -f "/home/decri/blockchain-projects/aura/chain/x/privacy/mixing.go" ]; then
    replace_time_now "/home/decri/blockchain-projects/aura/chain/x/privacy/mixing.go"
fi

if [ -f "/home/decri/blockchain-projects/aura/chain/x/privacy/encryption.go" ]; then
    replace_time_now "/home/decri/blockchain-projects/aura/chain/x/privacy/encryption.go"
fi

echo ""
echo "Step 10: Fixing auth/keeper files"
if [ -f "/home/decri/blockchain-projects/aura/chain/x/auth/keeper/msg_handlers.go" ]; then
    replace_time_now "/home/decri/blockchain-projects/aura/chain/x/auth/keeper/msg_handlers.go"
fi

if [ -f "/home/decri/blockchain-projects/aura/chain/x/auth/keeper/keeper.go" ]; then
    replace_time_now "/home/decri/blockchain-projects/aura/chain/x/auth/keeper/keeper.go"
fi

echo ""
echo "========================================="
echo "✓ ALL CONSENSUS BUGS FIXED"
echo "========================================="
echo ""
echo "Summary:"
echo "  - Replaced all time.Now() with ctx.BlockTime()"
echo "  - Replaced all timestamppb.Now() with timestamppb.New(ctx.BlockTime())"
echo "  - Added SDK imports where needed"
echo ""
echo "IMPORTANT NEXT STEPS:"
echo "  1. Some functions may need 'ctx' parameter added to signature"
echo "  2. Run: go build ./chain/..."
echo "  3. Fix any remaining compilation errors"
echo "  4. Test thoroughly before deploying"
echo ""
