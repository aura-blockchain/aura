#!/bin/bash

# Monitoring Module KV Store Migration Verification Script

echo "=========================================="
echo "Monitoring Module Migration Verification"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0

# Check function
check() {
    local test_name="$1"
    local command="$2"
    local expected="$3"

    echo -n "Testing: $test_name ... "

    result=$(eval "$command" 2>&1)
    exit_code=$?

    if [ $exit_code -eq $expected ]; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Expected exit code: $expected, Got: $exit_code"
        echo "  Output: $result"
        ((FAILED++))
        return 1
    fi
}

# Navigate to monitoring module
cd /home/decri/blockchain-projects/aura/chain/x/monitoring

echo "1. Checking for in-memory state removal..."
echo "-------------------------------------------"

# Check for in-memory maps (should NOT exist in keeper.go)
if ! grep -q "map\[string\]\*types\." keeper/keeper.go; then
    echo -e "${GREEN}✓ No in-memory maps found in keeper.go${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ In-memory maps still exist in keeper.go${NC}"
    grep "map\[string\]\*types\." keeper/keeper.go
    ((FAILED++))
fi

# Check for mutex (should NOT exist)
if ! grep -q "sync.RWMutex\|sync.Mutex" keeper/keeper.go; then
    echo -e "${GREEN}✓ No mutexes found in keeper.go${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Mutexes still exist in keeper.go${NC}"
    grep "sync\." keeper/keeper.go
    ((FAILED++))
fi

# Check for WaitGroup (should NOT exist)
if ! grep -q "sync.WaitGroup" keeper/keeper.go; then
    echo -e "${GREEN}✓ No WaitGroups found in keeper.go${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ WaitGroups still exist in keeper.go${NC}"
    ((FAILED++))
fi

# Check for background workers (should NOT exist in keeper.go)
if ! grep -q "go k\." keeper/keeper.go; then
    echo -e "${GREEN}✓ No background goroutines in keeper.go${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Background goroutines still exist in keeper.go${NC}"
    ((FAILED++))
fi

echo ""
echo "2. Checking for KV store implementation..."
echo "-------------------------------------------"

# Check for KV store usage
kv_count=$(grep -c "storeService.OpenKVStore" keeper/keeper.go)
if [ "$kv_count" -ge 38 ]; then
    echo -e "${GREEN}✓ Found $kv_count KV store operations (expected >= 38)${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Found only $kv_count KV store operations (expected >= 38)${NC}"
    ((FAILED++))
fi

# Check for Marshal/Unmarshal operations
marshal_count=$(grep -c "cdc.Marshal\|cdc.Unmarshal" keeper/keeper.go)
if [ "$marshal_count" -ge 30 ]; then
    echo -e "${GREEN}✓ Found $marshal_count codec operations${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Found only $marshal_count codec operations (expected >= 30)${NC}"
    ((FAILED++))
fi

# Check for Iterator usage
iterator_count=$(grep "store.Iterator" keeper/keeper.go | wc -l)
if [ "$iterator_count" -ge 7 ]; then
    echo -e "${GREEN}✓ Found $iterator_count iterator operations${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Found only $iterator_count iterator operations (expected >= 7)${NC}"
    ((FAILED++))
fi

echo ""
echo "3. Checking for proper method signatures..."
echo "-------------------------------------------"

# Check for context.Context in methods
context_methods=$(grep "func (k \*\?Keeper)" keeper/keeper.go | grep -c "ctx context.Context")
total_methods=$(grep -c "func (k \*\?Keeper)" keeper/keeper.go)

if [ "$context_methods" -ge 45 ]; then
    echo -e "${GREEN}✓ Found $context_methods methods with context.Context${NC}"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠ Found only $context_methods methods with context.Context (total: $total_methods)${NC}"
    echo "  Some methods may need updating"
fi

echo ""
echo "4. Checking key prefix definitions..."
echo "-------------------------------------------"

# Check for key prefix definitions
if grep -q "AlertKeyPrefix.*=.*\[\]byte{0x01}" keeper/keeper.go; then
    echo -e "${GREEN}✓ AlertKeyPrefix defined${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ AlertKeyPrefix not properly defined${NC}"
    ((FAILED++))
fi

if grep -q "TransactionKeyPrefix.*=.*\[\]byte{0x02}" keeper/keeper.go; then
    echo -e "${GREEN}✓ TransactionKeyPrefix defined${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ TransactionKeyPrefix not properly defined${NC}"
    ((FAILED++))
fi

if grep -q "ParamsKey.*=.*\[\]byte{0x0B}" keeper/keeper.go; then
    echo -e "${GREEN}✓ ParamsKey defined${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ ParamsKey not properly defined${NC}"
    ((FAILED++))
fi

echo ""
echo "5. Checking for consensus-safe ID generation..."
echo "-------------------------------------------"

if grep -q "func (k Keeper) generateID(ctx context.Context" keeper/keeper.go; then
    echo -e "${GREEN}✓ Consensus-safe generateID method exists${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ generateID method not found or incorrect signature${NC}"
    ((FAILED++))
fi

if grep -q "sdkCtx.BlockTime().UnixNano()" keeper/keeper.go; then
    echo -e "${GREEN}✓ Uses block time for ID generation${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Does not use block time for ID generation${NC}"
    ((FAILED++))
fi

echo ""
echo "6. Checking Keeper struct..."
echo "-------------------------------------------"

# Check Keeper struct has correct fields
if grep -q "storeService store.KVStoreService" keeper/keeper.go; then
    echo -e "${GREEN}✓ Keeper has storeService field${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Keeper missing storeService field${NC}"
    ((FAILED++))
fi

if grep -q "cdc.*codec.BinaryCodec" keeper/keeper.go; then
    echo -e "${GREEN}✓ Keeper has codec field${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Keeper missing codec field${NC}"
    ((FAILED++))
fi

if grep -q "authority.*string" keeper/keeper.go; then
    echo -e "${GREEN}✓ Keeper has authority field${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Keeper missing authority field${NC}"
    ((FAILED++))
fi

echo ""
echo "7. File statistics..."
echo "-------------------------------------------"

lines=$(wc -l < keeper/keeper.go)
echo "Total lines: $lines"

if [ "$lines" -ge 800 ]; then
    echo -e "${GREEN}✓ Keeper implementation is comprehensive ($lines lines)${NC}"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠ Keeper implementation may be incomplete ($lines lines)${NC}"
fi

echo ""
echo "8. Documentation check..."
echo "-------------------------------------------"

if [ -f "KEEPER_MIGRATION_COMPLETE.md" ]; then
    echo -e "${GREEN}✓ Migration documentation exists${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Migration documentation missing${NC}"
    ((FAILED++))
fi

if [ -f "QUICK_REFERENCE.md" ]; then
    echo -e "${GREEN}✓ Quick reference guide exists${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Quick reference guide missing${NC}"
    ((FAILED++))
fi

if [ -f "MIGRATION_SUMMARY.md" ]; then
    echo -e "${GREEN}✓ Migration summary exists${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Migration summary missing${NC}"
    ((FAILED++))
fi

echo ""
echo "=========================================="
echo "Verification Results"
echo "=========================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ "$FAILED" -eq 0 ]; then
    echo -e "${GREEN}✓✓✓ All checks passed! Migration successful! ✓✓✓${NC}"
    exit 0
else
    echo -e "${RED}✗✗✗ Some checks failed. Please review the issues above. ✗✗✗${NC}"
    exit 1
fi
