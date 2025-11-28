#!/bin/bash

# Comprehensive Test Runner for AURA Blockchain
# This script runs all test suites and generates reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Directories
CHAIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPORTS_DIR="${CHAIN_DIR}/test-reports"
COVERAGE_DIR="${REPORTS_DIR}/coverage"

# Create reports directory
mkdir -p "${REPORTS_DIR}"
mkdir -p "${COVERAGE_DIR}"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo "========================================"
echo "AURA Blockchain Test Suite"
echo "========================================"
echo ""

# Function to run tests and track results
run_test_suite() {
    local name=$1
    local command=$2
    local report_file="${REPORTS_DIR}/${name}.log"

    echo -e "${YELLOW}Running ${name}...${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if eval "${command}" > "${report_file}" 2>&1; then
        echo -e "${GREEN}✓ ${name} passed${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}✗ ${name} failed${NC}"
        echo "  See ${report_file} for details"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# ============================================================================
# 1. Unit Tests (by module)
# ============================================================================
echo ""
echo "========================================"
echo "Unit Tests"
echo "========================================"

MODULES=(
    "aiassistant"
    "auth"
    "bridge"
    "compliance"
    "confidencescore"
    "contractregistry"
    "cryptography"
    "dataregistry"
    "dex"
    "economicsecurity"
    "governance"
    "identitychange"
    "inclusionroutines"
    "monitoring"
    "networksecurity"
    "prevalidation"
    "privacy"
    "validatorsecurity"
    "vcregistry"
    "walletsecurity"
    "wasm"
)

for module in "${MODULES[@]}"; do
    if [ -d "x/${module}" ]; then
        run_test_suite "unit-${module}" "go test -short -v -race ./x/${module}/... -coverprofile=${COVERAGE_DIR}/${module}.out"
    fi
done

# ============================================================================
# 2. Integration Tests
# ============================================================================
echo ""
echo "========================================"
echo "Integration Tests"
echo "========================================"

run_test_suite "integration-all" "go test -tags=integration -v -timeout=45m ./testing/integration/..."
run_test_suite "integration-module-interaction" "go test -v -run TestModuleInteraction ./testing/integration/..."
run_test_suite "integration-upgrade" "go test -v -run TestUpgrade ./testing/integration/..."
run_test_suite "integration-security" "go test -v -run TestSecurity ./testing/integration/..."

# ============================================================================
# 3. End-to-End Tests
# ============================================================================
echo ""
echo "========================================"
echo "End-to-End Tests"
echo "========================================"

run_test_suite "e2e-all" "go test -v -timeout=60m ./testing/e2e/..."

# ============================================================================
# 4. Stress Tests
# ============================================================================
echo ""
echo "========================================"
echo "Stress & Performance Tests"
echo "========================================"

run_test_suite "stress-load" "go test -v -timeout=30m ./testing/stress/..."
run_test_suite "benchmark" "go test -bench=. -benchmem ./testing/benchmark/..."

# ============================================================================
# 5. Fuzz Tests (short duration for CI)
# ============================================================================
echo ""
echo "========================================"
echo "Fuzz Tests"
echo "========================================"

run_test_suite "fuzz-message" "go test -fuzz=FuzzMessageValidation -fuzztime=30s ./testing/fuzz/..." || true
run_test_suite "fuzz-address" "go test -fuzz=FuzzAddressValidation -fuzztime=30s ./testing/fuzz/..." || true
run_test_suite "fuzz-amount" "go test -fuzz=FuzzAmountValidation -fuzztime=30s ./testing/fuzz/..." || true

# ============================================================================
# 6. Coverage Analysis
# ============================================================================
echo ""
echo "========================================"
echo "Coverage Analysis"
echo "========================================"

# Merge coverage files
echo "Merging coverage files..."
echo "mode: atomic" > "${COVERAGE_DIR}/merged.out"
for f in "${COVERAGE_DIR}"/*.out; do
    if [ -f "$f" ] && [ "$f" != "${COVERAGE_DIR}/merged.out" ]; then
        tail -n +2 "$f" >> "${COVERAGE_DIR}/merged.out" 2>/dev/null || true
    fi
done

# Generate coverage reports
if [ -f "${COVERAGE_DIR}/merged.out" ]; then
    echo "Generating coverage report..."
    go tool cover -html="${COVERAGE_DIR}/merged.out" -o "${COVERAGE_DIR}/coverage.html"
    go tool cover -func="${COVERAGE_DIR}/merged.out" > "${COVERAGE_DIR}/coverage.txt"

    # Display coverage summary
    echo ""
    echo "Coverage Summary:"
    echo "----------------"
    tail -n 1 "${COVERAGE_DIR}/coverage.txt"

    # Check coverage threshold
    COVERAGE=$(go tool cover -func="${COVERAGE_DIR}/merged.out" | grep total | awk '{print $3}' | sed 's/%//')
    THRESHOLD=70

    if (( $(echo "$COVERAGE >= $THRESHOLD" | bc -l) )); then
        echo -e "${GREEN}✓ Coverage ${COVERAGE}% meets threshold ${THRESHOLD}%${NC}"
    else
        echo -e "${YELLOW}⚠ Coverage ${COVERAGE}% below threshold ${THRESHOLD}%${NC}"
    fi

    echo ""
    echo "Coverage reports:"
    echo "  HTML: ${COVERAGE_DIR}/coverage.html"
    echo "  Text: ${COVERAGE_DIR}/coverage.txt"
fi

# ============================================================================
# 7. Test Summary
# ============================================================================
echo ""
echo "========================================"
echo "Test Summary"
echo "========================================"
echo ""
echo "Total test suites run: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"
echo ""

if [ ${FAILED_TESTS} -gt 0 ]; then
    echo -e "${RED}Some tests failed. See individual logs in ${REPORTS_DIR}${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
