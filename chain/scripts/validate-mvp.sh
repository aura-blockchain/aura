#!/bin/bash
# AURA MVP Validation Script
# Validates MVP build, tests, and genesis before release.
#
# Usage: ./validate-mvp.sh [--quick]
#        --quick: Skip long-running tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHAIN_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="${CHAIN_DIR}/build"
TESTNETS_DIR="$(dirname "$(dirname "$CHAIN_DIR")")/testnets/aura-testnet-1"
TEMP_HOME="${CHAIN_DIR}/.mvp-validate-temp"

QUICK_MODE=false
if [ "$1" = "--quick" ]; then
    QUICK_MODE=true
fi

cd "$CHAIN_DIR"

echo "=============================================="
echo "       AURA MVP Validation Suite"
echo "=============================================="
echo ""

ERRORS=0
WARNINGS=0

pass() {
    echo "  [PASS] $1"
}

fail() {
    echo "  [FAIL] $1"
    ERRORS=$((ERRORS + 1))
}

warn() {
    echo "  [WARN] $1"
    WARNINGS=$((WARNINGS + 1))
}

# Step 1: Build MVP binary
echo "1. Building MVP binary..."
if make build-mvp > /dev/null 2>&1; then
    pass "make build-mvp succeeded"
else
    fail "make build-mvp failed"
fi

# Check binary exists
if [ -f "${BUILD_DIR}/aurad-mvp" ]; then
    pass "MVP binary exists at ${BUILD_DIR}/aurad-mvp"
else
    fail "MVP binary not found"
    echo "Cannot continue validation without binary."
    exit 1
fi

# Step 2: Run MVP tests
echo ""
echo "2. Running MVP module tests..."
if [ "$QUICK_MODE" = true ]; then
    echo "   (Quick mode - running subset)"
    if go test ./x/dataregistry/... -count=1 > /dev/null 2>&1; then
        pass "dataregistry tests passed"
    else
        fail "dataregistry tests failed"
    fi
else
    if make test-mvp > /dev/null 2>&1; then
        pass "make test-mvp succeeded"
    else
        fail "make test-mvp failed"
    fi
fi

# Step 3: Check test coverage
echo ""
echo "3. Checking test coverage..."

# Governance coverage
GOV_COV=$(go test -coverprofile=/tmp/gov-cov.out ./x/governance/... 2>&1 | grep -oP 'coverage: \K[0-9.]+' | head -1 || echo "0")
if [ -n "$GOV_COV" ] && [ "$(echo "$GOV_COV >= 60" | bc -l)" -eq 1 ]; then
    pass "governance coverage: ${GOV_COV}% (>= 60%)"
else
    warn "governance coverage: ${GOV_COV}% (target: 60%)"
fi

# Dataregistry coverage
DR_COV=$(go test -coverprofile=/tmp/dr-cov.out ./x/dataregistry/... 2>&1 | grep -oP 'coverage: \K[0-9.]+' | head -1 || echo "0")
if [ -n "$DR_COV" ] && [ "$(echo "$DR_COV >= 60" | bc -l)" -eq 1 ]; then
    pass "dataregistry coverage: ${DR_COV}% (>= 60%)"
else
    warn "dataregistry coverage: ${DR_COV}% (target: 60%)"
fi

# Step 4: Validate genesis
echo ""
echo "4. Validating MVP genesis..."

# Clean up temp dir
rm -rf "$TEMP_HOME"

# Initialize with MVP binary
if "${BUILD_DIR}/aurad-mvp" init test-node --chain-id test-mvp-1 --home "$TEMP_HOME" > /dev/null 2>&1; then
    pass "MVP init succeeded"
else
    fail "MVP init failed"
fi

# Check for genesis template
if [ -f "${TESTNETS_DIR}/genesis-mvp-template.json" ]; then
    pass "MVP genesis template exists"

    # Validate template structure
    if jq -e '.app_state.identity' "${TESTNETS_DIR}/genesis-mvp-template.json" > /dev/null 2>&1; then
        pass "genesis template has identity module"
    else
        warn "genesis template missing identity module"
    fi

    if jq -e '.app_state.dataregistry' "${TESTNETS_DIR}/genesis-mvp-template.json" > /dev/null 2>&1; then
        pass "genesis template has dataregistry module"
    else
        warn "genesis template missing dataregistry module"
    fi
else
    warn "MVP genesis template not found at ${TESTNETS_DIR}/genesis-mvp-template.json"
fi

# Step 5: Test node start (briefly)
echo ""
echo "5. Testing node initialization..."

if [ -f "${TEMP_HOME}/config/genesis.json" ]; then
    # Just validate genesis, don't start node
    if "${BUILD_DIR}/aurad-mvp" genesis validate "${TEMP_HOME}/config/genesis.json" --home "$TEMP_HOME" > /dev/null 2>&1; then
        pass "genesis validation passed"
    else
        warn "genesis validation failed (may need manual review)"
    fi
else
    fail "genesis.json not created during init"
fi

# Step 6: Check build tags
echo ""
echo "6. Checking build configuration..."

if grep -q "^//go:build mvp" "${CHAIN_DIR}/app/app_mvp.go"; then
    pass "app_mvp.go has correct build tag"
else
    fail "app_mvp.go missing build tag"
fi

if grep -q "^//go:build mvp" "${CHAIN_DIR}/app/mvp.go"; then
    pass "mvp.go has correct build tag"
else
    fail "mvp.go missing build tag"
fi

# Step 7: Check required files exist
echo ""
echo "7. Checking required files..."

REQUIRED_FILES=(
    "app/app_mvp.go"
    "app/mvp.go"
    "app/store_keys_mvp.go"
    "app/module_basics_mvp.go"
    "scripts/generate-mvp-genesis.sh"
    "scripts/build-mvp-release.sh"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "${CHAIN_DIR}/${file}" ]; then
        pass "$file exists"
    else
        fail "$file missing"
    fi
done

# Step 8: Check chain registry
echo ""
echo "8. Checking chain registry..."

if [ -f "${TESTNETS_DIR}/chain-mvp.json" ]; then
    pass "chain-mvp.json exists"

    # Validate JSON
    if jq empty "${TESTNETS_DIR}/chain-mvp.json" 2>/dev/null; then
        pass "chain-mvp.json is valid JSON"
    else
        fail "chain-mvp.json is invalid JSON"
    fi

    # Check required fields
    CHAIN_ID=$(jq -r '.chain_id' "${TESTNETS_DIR}/chain-mvp.json")
    if [ "$CHAIN_ID" = "aura-mvp-1" ]; then
        pass "chain_id is correct: $CHAIN_ID"
    else
        warn "chain_id unexpected: $CHAIN_ID"
    fi
else
    warn "chain-mvp.json not found"
fi

# Cleanup
rm -rf "$TEMP_HOME"
rm -f /tmp/gov-cov.out /tmp/dr-cov.out

# Summary
echo ""
echo "=============================================="
echo "              Validation Summary"
echo "=============================================="
echo ""
echo "  Errors:   $ERRORS"
echo "  Warnings: $WARNINGS"
echo ""

if [ $ERRORS -eq 0 ]; then
    echo "  Status: READY FOR RELEASE"
    echo ""
    exit 0
else
    echo "  Status: NOT READY - Fix errors above"
    echo ""
    exit 1
fi
