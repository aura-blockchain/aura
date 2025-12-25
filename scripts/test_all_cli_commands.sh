#!/bin/bash
# Comprehensive CLI Commands Test Script for Aura
# Tests all 29 CLI command categories against running testnet

AURAD="/home/hudson/blockchain-projects/aura/chain/aurad"
NODE="tcp://localhost:10501"
HOME="$HOME/.aura"
CHAIN_ID="aura-testnet-1"
TEST_ADDR="aura1g2nnl0x5hgvarpgfuwnyxfztss7h62yyj24k5p"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0
TOTAL=0

# Result arrays
declare -a PASS_COMMANDS
declare -a FAIL_COMMANDS
declare -a FAIL_MESSAGES

test_command() {
    local category="$1"
    local cmd="$2"
    local description="$3"

    TOTAL=$((TOTAL + 1))
    echo "=================================================="
    echo "Test $TOTAL: $description"
    echo "Category: $category"
    echo "Command: $cmd"
    echo "--------------------------------------------------"

    # Execute command and capture output
    output=$(eval "$cmd" 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
        echo -e "${GREEN}✅ PASS${NC}"
        echo "Output (first 5 lines):"
        echo "$output" | head -5
        PASSED=$((PASSED + 1))
        PASS_COMMANDS+=("$description")
    else
        echo -e "${RED}❌ FAIL${NC}"
        echo "Exit code: $exit_code"
        echo "Error output:"
        echo "$output" | head -10
        FAILED=$((FAILED + 1))
        FAIL_COMMANDS+=("$description")
        FAIL_MESSAGES+=("$output")
    fi
    echo ""
}

echo "=========================================="
echo "Aura CLI Commands Comprehensive Test"
echo "=========================================="
echo "Testnet: $CHAIN_ID"
echo "RPC: $NODE"
echo "Test Address: $TEST_ADDR"
echo "=========================================="
echo ""

# 1. BANK MODULE QUERIES
echo "======================================"
echo "1. TESTING BANK MODULE QUERIES"
echo "======================================"

test_command "bank" \
    "$AURAD query bank balances $TEST_ADDR --node $NODE --home $HOME" \
    "Bank: Query balances"

test_command "bank" \
    "$AURAD query bank total --node $NODE --home $HOME" \
    "Bank: Query total supply"

test_command "bank" \
    "$AURAD query bank denom-metadata --node $NODE --home $HOME" \
    "Bank: Query denom metadata"

test_command "bank" \
    "$AURAD query bank spendable-balances $TEST_ADDR --node $NODE --home $HOME" \
    "Bank: Query spendable balances"

# 2. DEX MODULE QUERIES
echo "======================================"
echo "2. TESTING DEX MODULE QUERIES"
echo "======================================"

test_command "dex" \
    "$AURAD query dex params --node $NODE --home $HOME" \
    "DEX: Query params"

test_command "dex" \
    "$AURAD query dex list-order-book --node $NODE --home $HOME" \
    "DEX: List order books"

test_command "dex" \
    "$AURAD query dex list-order --node $NODE --home $HOME" \
    "DEX: List orders"

test_command "dex" \
    "$AURAD query dex list-trade --node $NODE --home $HOME" \
    "DEX: List trades"

test_command "dex" \
    "$AURAD query dex get-order-book uaura usdc --node $NODE --home $HOME" \
    "DEX: Get specific order book"

# 3. COMPLIANCE MODULE QUERIES
echo "======================================"
echo "3. TESTING COMPLIANCE MODULE QUERIES"
echo "======================================"

test_command "compliance" \
    "$AURAD query compliance params --node $NODE --home $HOME" \
    "Compliance: Query params"

test_command "compliance" \
    "$AURAD query compliance list-verification --node $NODE --home $HOME" \
    "Compliance: List verifications"

test_command "compliance" \
    "$AURAD query compliance list-operator --node $NODE --home $HOME" \
    "Compliance: List operators"

test_command "compliance" \
    "$AURAD query compliance show-verification $TEST_ADDR --node $NODE --home $HOME" \
    "Compliance: Show verification"

# 4. CONFIDENCE SCORE MODULE QUERIES
echo "======================================"
echo "4. TESTING CONFIDENCE SCORE QUERIES"
echo "======================================"

test_command "confidencescore" \
    "$AURAD query confidencescore params --node $NODE --home $HOME" \
    "ConfidenceScore: Query params"

test_command "confidencescore" \
    "$AURAD query confidencescore get-score $TEST_ADDR --node $NODE --home $HOME" \
    "ConfidenceScore: Get score"

test_command "confidencescore" \
    "$AURAD query confidencescore list-score --node $NODE --home $HOME" \
    "ConfidenceScore: List scores"

# 5. STAKING MODULE QUERIES
echo "======================================"
echo "5. TESTING STAKING MODULE QUERIES"
echo "======================================"

test_command "staking" \
    "$AURAD query staking params --node $NODE --home $HOME" \
    "Staking: Query params"

test_command "staking" \
    "$AURAD query staking validators --node $NODE --home $HOME" \
    "Staking: List validators"

test_command "staking" \
    "$AURAD query staking delegations $TEST_ADDR --node $NODE --home $HOME" \
    "Staking: Query delegations"

test_command "staking" \
    "$AURAD query staking pool --node $NODE --home $HOME" \
    "Staking: Query pool"

# 6. GOVERNANCE MODULE QUERIES
echo "======================================"
echo "6. TESTING GOVERNANCE MODULE QUERIES"
echo "======================================"

test_command "gov" \
    "$AURAD query gov params --node $NODE --home $HOME" \
    "Gov: Query params"

test_command "gov" \
    "$AURAD query gov proposals --node $NODE --home $HOME" \
    "Gov: List proposals"

# 7. DISTRIBUTION MODULE QUERIES
echo "======================================"
echo "7. TESTING DISTRIBUTION QUERIES"
echo "======================================"

test_command "distribution" \
    "$AURAD query distribution params --node $NODE --home $HOME" \
    "Distribution: Query params"

test_command "distribution" \
    "$AURAD query distribution rewards $TEST_ADDR --node $NODE --home $HOME" \
    "Distribution: Query rewards"

# 8. AUTH MODULE QUERIES
echo "======================================"
echo "8. TESTING AUTH MODULE QUERIES"
echo "======================================"

test_command "auth" \
    "$AURAD query auth params --node $NODE --home $HOME" \
    "Auth: Query params"

test_command "auth" \
    "$AURAD query auth account $TEST_ADDR --node $NODE --home $HOME" \
    "Auth: Query account"

# 9. ADDITIONAL MODULE QUERIES
echo "======================================"
echo "9. TESTING ADDITIONAL QUERIES"
echo "======================================"

test_command "privacy" \
    "$AURAD query privacy params --node $NODE --home $HOME" \
    "Privacy: Query params"

test_command "did" \
    "$AURAD query did params --node $NODE --home $HOME" \
    "DID: Query params"

# SUMMARY
echo "=========================================="
echo "TEST SUMMARY"
echo "=========================================="
echo "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo "Pass Rate: $(awk "BEGIN {printf \"%.1f\", ($PASSED/$TOTAL)*100}")%"
echo "=========================================="

if [ $FAILED -gt 0 ]; then
    echo ""
    echo "FAILED COMMANDS:"
    echo "=========================================="
    for i in "${!FAIL_COMMANDS[@]}"; do
        echo -e "${RED}❌ ${FAIL_COMMANDS[$i]}${NC}"
    done
fi

if [ $PASSED -gt 0 ]; then
    echo ""
    echo "PASSED COMMANDS:"
    echo "=========================================="
    for cmd in "${PASS_COMMANDS[@]}"; do
        echo -e "${GREEN}✅ $cmd${NC}"
    done
fi

echo ""
echo "=========================================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}ALL TESTS PASSED! 🎉${NC}"
else
    echo -e "${YELLOW}Some tests failed. See details above.${NC}"
fi
echo "=========================================="

exit $FAILED
