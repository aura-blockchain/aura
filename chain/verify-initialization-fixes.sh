#!/bin/bash
# Verification script for critical app initialization fixes
# This script checks that all required changes are in place

set -e

echo "=========================================="
echo "AURA Chain Initialization Fixes Verification"
echo "=========================================="
echo ""

ERRORS=0
WARNINGS=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_pass() {
    echo -e "${GREEN}✓${NC} $1"
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    ERRORS=$((ERRORS + 1))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    WARNINGS=$((WARNINGS + 1))
}

echo "1. Checking builder files exist..."
if [ -f "x/vcregistry/keeper/builder.go" ]; then
    check_pass "VCRegistry builder exists"
else
    check_fail "VCRegistry builder missing"
fi

if [ -f "x/confidencescore/keeper/builder.go" ]; then
    check_pass "ConfidenceScore builder exists"
else
    check_fail "ConfidenceScore builder missing"
fi

echo ""
echo "2. Checking store keys in app.go..."
if grep -q "confidenceScore.*KVStoreKey" app/app.go; then
    check_pass "confidenceScore store key declared"
else
    check_fail "confidenceScore store key missing"
fi

if grep -q "inclusionRoutines.*KVStoreKey" app/app.go; then
    check_pass "inclusionRoutines store key declared"
else
    check_fail "inclusionRoutines store key missing"
fi

if grep -q "identityChange.*KVStoreKey" app/app.go; then
    check_pass "identityChange store key declared"
else
    check_fail "identityChange store key missing"
fi

if grep -q "dataRegistry.*KVStoreKey" app/app.go; then
    check_pass "dataRegistry store key declared"
else
    check_fail "dataRegistry store key missing"
fi

echo ""
echo "3. Checking store key creation..."
if grep -q "confidenceScoreKey.*:=.*NewKVStoreKey" app/app.go; then
    check_pass "confidenceScore store key created"
else
    check_fail "confidenceScore store key not created"
fi

if grep -q "inclusionRoutinesKey.*:=.*NewKVStoreKey" app/app.go; then
    check_pass "inclusionRoutines store key created"
else
    check_fail "inclusionRoutines store key not created"
fi

if grep -q "identityChangeKey.*:=.*NewKVStoreKey" app/app.go; then
    check_pass "identityChange store key created"
else
    check_fail "identityChange store key not created"
fi

if grep -q "dataRegistryKey.*:=.*NewKVStoreKey" app/app.go; then
    check_pass "dataRegistry store key created"
else
    check_fail "dataRegistry store key not created"
fi

echo ""
echo "4. Checking DEX permissions fix..."
if grep -q "dextypes.ModuleName.*{authtypes.Minter" app/app.go; then
    check_fail "DEX still has Minter permission (SECURITY ISSUE)"
else
    check_pass "DEX Minter permission removed"
fi

if grep -q "dextypes.ModuleName.*{authtypes.Burner}" app/app.go; then
    check_pass "DEX has Burner permission"
else
    check_warn "DEX might be missing Burner permission"
fi

echo ""
echo "5. Checking builder pattern usage..."
if grep -q "cskeeper.NewKeeperBuilder" app/app.go; then
    check_pass "ConfidenceScore uses builder pattern"
else
    check_fail "ConfidenceScore not using builder pattern"
fi

if grep -q "vckeeper.NewKeeperBuilder" app/app.go; then
    check_pass "VCRegistry uses builder pattern"
else
    check_fail "VCRegistry not using builder pattern"
fi

echo ""
echo "6. Checking initialization logging..."
if grep -q "tier-1-no-deps" app/app.go; then
    check_pass "Tier 1 initialization logging present"
else
    check_warn "Missing tier 1 initialization logging"
fi

if grep -q "tier-4-confidence-score" app/app.go; then
    check_pass "Tier 4 initialization logging present"
else
    check_warn "Missing tier 4 initialization logging"
fi

if grep -q "tier-5-vcregistry" app/app.go; then
    check_pass "Tier 5 initialization logging present"
else
    check_warn "Missing tier 5 initialization logging"
fi

echo ""
echo "7. Checking RunMigrations method..."
if grep -q "func.*RunMigrations" app/module_manager.go; then
    check_pass "RunMigrations method exists"
else
    check_fail "RunMigrations method missing"
fi

if grep -q "migrateModules.*confidencescore" app/module_manager.go; then
    check_pass "ConfidenceScore migration registered"
else
    check_warn "ConfidenceScore migration might be missing"
fi

if grep -q "migrateModules.*vcregistry" app/module_manager.go; then
    check_pass "VCRegistry migration registered"
else
    check_warn "VCRegistry migration might be missing"
fi

echo ""
echo "8. Checking invariant registration..."
if grep -q "registerInvariants" app/app.go; then
    check_pass "registerInvariants method exists"
else
    check_fail "registerInvariants method missing"
fi

if grep -q "func.*registerInvariants" app/app.go; then
    check_pass "registerInvariants implementation found"
else
    check_fail "registerInvariants implementation missing"
fi

echo ""
echo "9. Checking SupplyMonitor..."
if grep -q "type SupplyMonitor struct" app/app.go; then
    check_pass "SupplyMonitor type defined"
else
    check_fail "SupplyMonitor type missing"
fi

if grep -q "func.*NewSupplyMonitor" app/app.go; then
    check_pass "NewSupplyMonitor constructor exists"
else
    check_fail "NewSupplyMonitor constructor missing"
fi

if grep -q "RecordMint.*blockHeight.*module.*amount" app/app.go; then
    check_pass "RecordMint method exists"
else
    check_fail "RecordMint method missing"
fi

echo ""
echo "10. Checking documentation..."
if [ -f "CRITICAL_APP_INITIALIZATION_FIXES_REPORT.md" ]; then
    check_pass "Implementation report exists"
else
    check_warn "Implementation report missing"
fi

if [ -f "KEEPER_BUILDER_PATTERN_GUIDE.md" ]; then
    check_pass "Builder pattern guide exists"
else
    check_warn "Builder pattern guide missing"
fi

if [ -f "CHANGES_SUMMARY.md" ]; then
    check_pass "Changes summary exists"
else
    check_warn "Changes summary missing"
fi

echo ""
echo "=========================================="
echo "Verification Summary"
echo "=========================================="
echo ""

if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}All critical checks passed!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Run: go build -o /tmp/aura-test ./app"
    echo "2. Run: go test ./app/..."
    echo "3. Run: go test ./x/vcregistry/keeper/..."
    echo "4. Run: go test ./x/confidencescore/keeper/..."
    echo "5. Review CRITICAL_APP_INITIALIZATION_FIXES_REPORT.md"
else
    echo -e "${RED}$ERRORS critical errors found!${NC}"
    echo ""
    echo "Please fix the errors above before deploying."
    exit 1
fi

if [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}$WARNINGS warnings found.${NC}"
    echo "Review warnings to ensure everything is configured correctly."
fi

echo ""
echo "=========================================="
