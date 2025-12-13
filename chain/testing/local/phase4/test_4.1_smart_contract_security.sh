#!/bin/bash
# Phase 4.1: Smart Contract Security Analysis
# Tests: cargo audit + manual security review for common vulnerabilities

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
RESULTS_FILE="$SCRIPT_DIR/test_4.1_results.txt"

echo "=== Phase 4.1: Smart Contract Security Analysis ===" | tee "$RESULTS_FILE"
echo "Started at: $(date)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter for issues found
CRITICAL_ISSUES=0
HIGH_ISSUES=0
MEDIUM_ISSUES=0
LOW_ISSUES=0

echo "Step 1: Running cargo audit on all contracts..." | tee -a "$RESULTS_FILE"
echo "=================================================" | tee -a "$RESULTS_FILE"

# Check if cargo-audit is installed
if ! command -v cargo-audit &> /dev/null; then
    echo -e "${YELLOW}cargo-audit not installed. Installing...${NC}" | tee -a "$RESULTS_FILE"
    cargo install cargo-audit --locked
fi

# Run cargo audit on each contract workspace
for cargo_toml in "$PROJECT_ROOT"/contracts/*/Cargo.toml "$PROJECT_ROOT"/contracts/Cargo.toml; do
    if [ -f "$cargo_toml" ]; then
        contract_dir=$(dirname "$cargo_toml")
        contract_name=$(basename "$contract_dir")

        echo "" | tee -a "$RESULTS_FILE"
        echo "Auditing: $contract_name" | tee -a "$RESULTS_FILE"
        echo "---" | tee -a "$RESULTS_FILE"

        cd "$contract_dir"

        # Run cargo audit
        if cargo audit 2>&1 | tee -a "$RESULTS_FILE"; then
            echo -e "${GREEN}✓ No vulnerabilities found in $contract_name${NC}" | tee -a "$RESULTS_FILE"
        else
            echo -e "${RED}✗ Vulnerabilities found in $contract_name${NC}" | tee -a "$RESULTS_FILE"
            CRITICAL_ISSUES=$((CRITICAL_ISSUES + 1))
        fi
    fi
done

echo "" | tee -a "$RESULTS_FILE"
echo "Step 2: Manual Security Review - Common Vulnerabilities" | tee -a "$RESULTS_FILE"
echo "=========================================================" | tee -a "$RESULTS_FILE"

cd "$PROJECT_ROOT"

# 2.1: Check for reentrancy vulnerabilities (external calls before state changes)
echo "" | tee -a "$RESULTS_FILE"
echo "2.1: Checking for potential reentrancy vulnerabilities..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

# Look for patterns where external calls happen before state updates
REENTRANCY_PATTERNS=$(grep -r "\.execute(" contracts/ --include="*.rs" | grep -v "test" | grep -v "//" || true)
if [ -n "$REENTRANCY_PATTERNS" ]; then
    echo -e "${YELLOW}Found external execute calls - manual review required:${NC}" | tee -a "$RESULTS_FILE"
    echo "$REENTRANCY_PATTERNS" | head -20 | tee -a "$RESULTS_FILE"
    MEDIUM_ISSUES=$((MEDIUM_ISSUES + 1))
else
    echo -e "${GREEN}✓ No obvious reentrancy patterns detected${NC}" | tee -a "$RESULTS_FILE"
fi

# 2.2: Check for integer overflow/underflow (should use checked arithmetic)
echo "" | tee -a "$RESULTS_FILE"
echo "2.2: Checking for unchecked arithmetic operations..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

# Look for direct arithmetic without checked_ methods
UNCHECKED_MATH=$(grep -rE "\+ |\- |\* |/" contracts/ --include="*.rs" | \
    grep -v "checked_" | \
    grep -v "saturating_" | \
    grep -v "test" | \
    grep -v "//" | \
    grep -v "String" | \
    grep -v "format!" | \
    grep -v "#\[" | \
    head -20 || true)

if [ -n "$UNCHECKED_MATH" ]; then
    echo -e "${YELLOW}Found potential unchecked arithmetic - review for overflow safety:${NC}" | tee -a "$RESULTS_FILE"
    echo "$UNCHECKED_MATH" | head -10 | tee -a "$RESULTS_FILE"
    LOW_ISSUES=$((LOW_ISSUES + 1))
else
    echo -e "${GREEN}✓ No unchecked arithmetic detected${NC}" | tee -a "$RESULTS_FILE"
fi

# 2.3: Check for proper access control
echo "" | tee -a "$RESULTS_FILE"
echo "2.3: Checking for access control patterns..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

# Look for execute handlers without authorization checks
MISSING_AUTH=$(grep -r "pub fn execute" contracts/ --include="*.rs" -A 10 | \
    grep -v "ensure!" | \
    grep -v "assert_eq!" | \
    grep -v "require" | \
    grep -v "test" || true)

if [ -n "$MISSING_AUTH" ]; then
    echo -e "${YELLOW}Found execute functions - verify authorization checks:${NC}" | tee -a "$RESULTS_FILE"
    # Count unique execute functions
    EXEC_COUNT=$(grep -r "pub fn execute" contracts/ --include="*.rs" | wc -l)
    echo "Total execute functions found: $EXEC_COUNT" | tee -a "$RESULTS_FILE"
    echo "Manual review recommended for each" | tee -a "$RESULTS_FILE"
    MEDIUM_ISSUES=$((MEDIUM_ISSUES + 1))
else
    echo -e "${GREEN}✓ Access control patterns appear present${NC}" | tee -a "$RESULTS_FILE"
fi

# 2.4: Check for unsafe unwrap() usage
echo "" | tee -a "$RESULTS_FILE"
echo "2.4: Checking for unsafe unwrap() usage..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

UNSAFE_UNWRAP=$(grep -r "\.unwrap()" contracts/ --include="*.rs" | \
    grep -v "test" | \
    grep -v "//" || true)

UNWRAP_COUNT=$(echo "$UNSAFE_UNWRAP" | grep -c "unwrap()" || echo "0")
if [ "$UNWRAP_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}Found $UNWRAP_COUNT unwrap() calls - should use proper error handling:${NC}" | tee -a "$RESULTS_FILE"
    echo "$UNSAFE_UNWRAP" | head -10 | tee -a "$RESULTS_FILE"
    MEDIUM_ISSUES=$((MEDIUM_ISSUES + 1))
else
    echo -e "${GREEN}✓ No unsafe unwrap() usage detected${NC}" | tee -a "$RESULTS_FILE"
fi

# 2.5: Check for proper input validation
echo "" | tee -a "$RESULTS_FILE"
echo "2.5: Checking for input validation patterns..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

# Look for validation patterns
VALIDATION_PRESENT=$(grep -r "validate" contracts/ --include="*.rs" | wc -l)
ENSURE_PRESENT=$(grep -r "ensure!" contracts/ --include="*.rs" | wc -l)

echo "Validation patterns found: $VALIDATION_PRESENT" | tee -a "$RESULTS_FILE"
echo "Ensure! macros found: $ENSURE_PRESENT" | tee -a "$RESULTS_FILE"

if [ "$VALIDATION_PRESENT" -lt 5 ] && [ "$ENSURE_PRESENT" -lt 5 ]; then
    echo -e "${YELLOW}⚠ Limited input validation detected - review recommended${NC}" | tee -a "$RESULTS_FILE"
    LOW_ISSUES=$((LOW_ISSUES + 1))
else
    echo -e "${GREEN}✓ Input validation patterns present${NC}" | tee -a "$RESULTS_FILE"
fi

# 2.6: Check for storage key collisions
echo "" | tee -a "$RESULTS_FILE"
echo "2.6: Checking for storage key collision risks..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

# Look for storage definitions
STORAGE_KEYS=$(grep -r "const.*STORAGE" contracts/ --include="*.rs" || true)
if [ -n "$STORAGE_KEYS" ]; then
    echo "Storage keys defined:" | tee -a "$RESULTS_FILE"
    echo "$STORAGE_KEYS" | tee -a "$RESULTS_FILE"

    # Check for duplicate keys
    DUPLICATES=$(echo "$STORAGE_KEYS" | awk -F'"' '{print $2}' | sort | uniq -d)
    if [ -n "$DUPLICATES" ]; then
        echo -e "${RED}✗ CRITICAL: Duplicate storage keys found!${NC}" | tee -a "$RESULTS_FILE"
        echo "$DUPLICATES" | tee -a "$RESULTS_FILE"
        CRITICAL_ISSUES=$((CRITICAL_ISSUES + 1))
    else
        echo -e "${GREEN}✓ No duplicate storage keys detected${NC}" | tee -a "$RESULTS_FILE"
    fi
else
    echo -e "${GREEN}✓ No hardcoded storage keys found (likely using Item/Map)${NC}" | tee -a "$RESULTS_FILE"
fi

# 2.7: Check for proper error handling
echo "" | tee -a "$RESULTS_FILE"
echo "2.7: Checking error handling patterns..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

# Look for Result types and error handling
ERROR_TYPES=$(grep -r "StdError\|ContractError" contracts/ --include="*.rs" | wc -l)
QUESTION_MARKS=$(grep -r "?" contracts/ --include="*.rs" | grep -v "//" | wc -l)

echo "Error type usages: $ERROR_TYPES" | tee -a "$RESULTS_FILE"
echo "Error propagation (?) operators: $QUESTION_MARKS" | tee -a "$RESULTS_FILE"

if [ "$ERROR_TYPES" -gt 20 ]; then
    echo -e "${GREEN}✓ Comprehensive error handling present${NC}" | tee -a "$RESULTS_FILE"
else
    echo -e "${YELLOW}⚠ Limited error handling - review recommended${NC}" | tee -a "$RESULTS_FILE"
    LOW_ISSUES=$((LOW_ISSUES + 1))
fi

# 2.8: Check for timestamp dependency issues
echo "" | tee -a "$RESULTS_FILE"
echo "2.8: Checking for timestamp manipulation risks..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

TIMESTAMP_USAGE=$(grep -r "block\.time\|env\.block\.time" contracts/ --include="*.rs" | \
    grep -v "test" | \
    grep -v "//" || true)

if [ -n "$TIMESTAMP_USAGE" ]; then
    TIMESTAMP_COUNT=$(echo "$TIMESTAMP_USAGE" | wc -l)
    echo -e "${YELLOW}Found $TIMESTAMP_COUNT timestamp usages - verify not used for critical logic:${NC}" | tee -a "$RESULTS_FILE"
    echo "$TIMESTAMP_USAGE" | head -5 | tee -a "$RESULTS_FILE"
    LOW_ISSUES=$((LOW_ISSUES + 1))
else
    echo -e "${GREEN}✓ No timestamp dependencies detected${NC}" | tee -a "$RESULTS_FILE"
fi

# 2.9: Check for proper migration handling
echo "" | tee -a "$RESULTS_FILE"
echo "2.9: Checking for migration implementations..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

MIGRATE_FUNCS=$(grep -r "pub fn migrate" contracts/ --include="*.rs" | wc -l)
if [ "$MIGRATE_FUNCS" -gt 0 ]; then
    echo -e "${GREEN}✓ Found $MIGRATE_FUNCS migration functions${NC}" | tee -a "$RESULTS_FILE"

    # Check for version checks in migrations
    VERSION_CHECKS=$(grep -r "CONTRACT_VERSION\|cw2::set_contract_version" contracts/ --include="*.rs" | wc -l)
    if [ "$VERSION_CHECKS" -gt 0 ]; then
        echo -e "${GREEN}✓ Version tracking present in contracts${NC}" | tee -a "$RESULTS_FILE"
    else
        echo -e "${YELLOW}⚠ No version tracking found - review migration safety${NC}" | tee -a "$RESULTS_FILE"
        MEDIUM_ISSUES=$((MEDIUM_ISSUES + 1))
    fi
else
    echo -e "${YELLOW}⚠ No migration functions found - upgrades may be difficult${NC}" | tee -a "$RESULTS_FILE"
    LOW_ISSUES=$((LOW_ISSUES + 1))
fi

# 2.10: Check for front-running vulnerabilities
echo "" | tee -a "$RESULTS_FILE"
echo "2.10: Checking for front-running risks (DEX/auction patterns)..." | tee -a "$RESULTS_FILE"
echo "---" | tee -a "$RESULTS_FILE"

FRONTRUN_PATTERNS=$(grep -rE "swap|bid|offer" contracts/ --include="*.rs" | \
    grep -v "test" | \
    grep -v "//" || true)

if [ -n "$FRONTRUN_PATTERNS" ]; then
    echo -e "${YELLOW}Found patterns that may be vulnerable to front-running:${NC}" | tee -a "$RESULTS_FILE"
    echo "$FRONTRUN_PATTERNS" | head -5 | tee -a "$RESULTS_FILE"
    echo "Review: Use commit-reveal schemes or minimum delays where applicable" | tee -a "$RESULTS_FILE"
    MEDIUM_ISSUES=$((MEDIUM_ISSUES + 1))
else
    echo -e "${GREEN}✓ No obvious front-running patterns detected${NC}" | tee -a "$RESULTS_FILE"
fi

# Step 3: Build all contracts to verify compilation
echo "" | tee -a "$RESULTS_FILE"
echo "Step 3: Building all contracts..." | tee -a "$RESULTS_FILE"
echo "===================================" | tee -a "$RESULTS_FILE"

cd "$PROJECT_ROOT/contracts"

if cargo build --release 2>&1 | tee -a "$RESULTS_FILE"; then
    echo -e "${GREEN}✓ All contracts built successfully${NC}" | tee -a "$RESULTS_FILE"
else
    echo -e "${RED}✗ Contract build failed${NC}" | tee -a "$RESULTS_FILE"
    CRITICAL_ISSUES=$((CRITICAL_ISSUES + 1))
fi

# Step 4: Run contract tests
echo "" | tee -a "$RESULTS_FILE"
echo "Step 4: Running contract tests..." | tee -a "$RESULTS_FILE"
echo "===================================" | tee -a "$RESULTS_FILE"

if cargo test 2>&1 | tee -a "$RESULTS_FILE"; then
    echo -e "${GREEN}✓ All contract tests passed${NC}" | tee -a "$RESULTS_FILE"
else
    echo -e "${RED}✗ Some contract tests failed${NC}" | tee -a "$RESULTS_FILE"
    HIGH_ISSUES=$((HIGH_ISSUES + 1))
fi

# Final summary
echo "" | tee -a "$RESULTS_FILE"
echo "========================================" | tee -a "$RESULTS_FILE"
echo "PHASE 4.1 SECURITY ANALYSIS SUMMARY" | tee -a "$RESULTS_FILE"
echo "========================================" | tee -a "$RESULTS_FILE"
echo "Completed at: $(date)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"
echo "Issues found by severity:" | tee -a "$RESULTS_FILE"
echo "  Critical: $CRITICAL_ISSUES" | tee -a "$RESULTS_FILE"
echo "  High:     $HIGH_ISSUES" | tee -a "$RESULTS_FILE"
echo "  Medium:   $MEDIUM_ISSUES" | tee -a "$RESULTS_FILE"
echo "  Low:      $LOW_ISSUES" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

TOTAL_ISSUES=$((CRITICAL_ISSUES + HIGH_ISSUES + MEDIUM_ISSUES + LOW_ISSUES))

if [ "$CRITICAL_ISSUES" -gt 0 ]; then
    echo -e "${RED}✗ CRITICAL ISSUES FOUND - MUST FIX BEFORE PRODUCTION${NC}" | tee -a "$RESULTS_FILE"
    exit 1
elif [ "$HIGH_ISSUES" -gt 0 ]; then
    echo -e "${YELLOW}⚠ HIGH ISSUES FOUND - REVIEW RECOMMENDED${NC}" | tee -a "$RESULTS_FILE"
    exit 1
elif [ "$TOTAL_ISSUES" -gt 0 ]; then
    echo -e "${YELLOW}⚠ Security review complete with $TOTAL_ISSUES minor findings${NC}" | tee -a "$RESULTS_FILE"
    echo "Review findings and address before production deployment" | tee -a "$RESULTS_FILE"
else
    echo -e "${GREEN}✓ SECURITY ANALYSIS PASSED - No critical issues found${NC}" | tee -a "$RESULTS_FILE"
fi

echo "" | tee -a "$RESULTS_FILE"
echo "Detailed results saved to: $RESULTS_FILE" | tee -a "$RESULTS_FILE"
