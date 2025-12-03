#!/bin/bash

# Verify test files have correct syntax without running them
# This is useful when Go runtime isn't available

echo "Verifying test file syntax..."
echo ""

cd /home/decri/blockchain-projects/aura/chain

# Check if files exist
echo "1. Checking test files exist..."
test -f x/dex/keeper/pool_creation_record_test.go && echo "   ✓ pool_creation_record_test.go exists" || echo "   ✗ pool_creation_record_test.go missing"
test -f x/dex/keeper/query_server_test.go && echo "   ✓ query_server_test.go exists" || echo "   ✗ query_server_test.go missing"
echo ""

# Check for syntax errors (basic checks)
echo "2. Checking for common syntax issues..."

# Check for unclosed braces
for file in x/dex/keeper/pool_creation_record_test.go x/dex/keeper/query_server_test.go; do
    open_braces=$(grep -o '{' "$file" | wc -l)
    close_braces=$(grep -o '}' "$file" | wc -l)
    if [ "$open_braces" -eq "$close_braces" ]; then
        echo "   ✓ $file: braces balanced ($open_braces pairs)"
    else
        echo "   ✗ $file: brace mismatch (open: $open_braces, close: $close_braces)"
    fi
done
echo ""

# Check for proper imports
echo "3. Checking imports..."
for file in x/dex/keeper/pool_creation_record_test.go x/dex/keeper/query_server_test.go; do
    if grep -q 'import (' "$file"; then
        echo "   ✓ $file: has import block"
    else
        echo "   ✗ $file: missing import block"
    fi
done
echo ""

# Check for test function definitions
echo "4. Checking test functions..."
echo "   Pool creation record tests:"
grep -c 'func TestPoolCreationLimit_Enforcement' x/dex/keeper/pool_creation_record_test.go > /dev/null && \
    echo "     ✓ TestPoolCreationLimit_Enforcement found"
grep -c 'func TestPoolCreationCooldown_RespectsCooldownPeriod' x/dex/keeper/pool_creation_record_test.go > /dev/null && \
    echo "     ✓ TestPoolCreationCooldown_RespectsCooldownPeriod found"

echo "   Query server tests:"
grep -c 'func TestQueryMarketPriceUsesStoredValue' x/dex/keeper/query_server_test.go > /dev/null && \
    echo "     ✓ TestQueryMarketPriceUsesStoredValue found"
grep -c 'func TestQuerySpotPrice' x/dex/keeper/query_server_test.go > /dev/null && \
    echo "     ✓ TestQuerySpotPrice found"
echo ""

# Check for proper test setup calls
echo "5. Checking test setup..."
for file in x/dex/keeper/pool_creation_record_test.go x/dex/keeper/query_server_test.go; do
    if grep -q 'SetupKeeperTestSuite\|setupTestKeeper' "$file"; then
        echo "   ✓ $file: uses proper test setup"
    else
        echo "   ⚠ $file: might be missing test setup"
    fi
done
echo ""

# Check for fmt import in pool_creation_record_test.go (needed for fmt.Sprintf)
echo "6. Checking required imports for fixes..."
if grep -q '"fmt"' x/dex/keeper/pool_creation_record_test.go; then
    echo "   ✓ pool_creation_record_test.go: has fmt import (needed for sprintf)"
else
    echo "   ⚠ pool_creation_record_test.go: missing fmt import"
fi
echo ""

echo "Syntax verification complete!"
echo ""
echo "Note: This is a basic syntax check. Run actual Go tests to verify functionality:"
echo "  go test -v ./x/dex/keeper/... -run 'TestPoolCreationLimit_Enforcement|TestPoolCreationCooldown_RespectsCooldownPeriod|TestQueryMarketPriceUsesStoredValue|TestQuerySpotPrice'"
