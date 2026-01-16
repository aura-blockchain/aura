#!/bin/bash
# AURA MVP Integration Test Suite
# Runs all integration tests against the live testnet

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RPC_ENDPOINT="${RPC_ENDPOINT:-https://testnet-rpc.aurablockchain.org}"
REST_ENDPOINT="${REST_ENDPOINT:-https://testnet-api.aurablockchain.org}"
CHAIN_ID="${CHAIN_ID:-aura-mvp-1}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0
SKIPPED=0

log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; ((PASSED++)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; ((FAILED++)); }
log_skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; ((SKIPPED++)); }
log_info() { echo -e "[INFO] $1"; }

echo "=========================================="
echo "AURA MVP Integration Test Suite"
echo "=========================================="
echo "RPC: $RPC_ENDPOINT"
echo "REST: $REST_ENDPOINT"
echo "Chain ID: $CHAIN_ID"
echo "Time: $(date -u)"
echo "=========================================="
echo

# Run individual test scripts
for test_script in "$SCRIPT_DIR"/[0-9][0-9]-*.sh; do
    if [ -f "$test_script" ]; then
        test_name=$(basename "$test_script" .sh)
        log_info "Running: $test_name"

        if bash "$test_script" "$RPC_ENDPOINT" "$REST_ENDPOINT" "$CHAIN_ID"; then
            log_pass "$test_name"
        else
            log_fail "$test_name"
        fi
        echo
    fi
done

echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Passed:  ${GREEN}$PASSED${NC}"
echo -e "Failed:  ${RED}$FAILED${NC}"
echo -e "Skipped: ${YELLOW}$SKIPPED${NC}"
echo "=========================================="

if [ $FAILED -gt 0 ]; then
    exit 1
fi
