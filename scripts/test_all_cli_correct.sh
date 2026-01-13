#!/bin/bash
# Comprehensive CLI Commands Test Script for Aura (with correct command names)
# Tests all available CLI commands against running testnet

AURAD="/home/hudson/blockchain-projects/aura/chain/aurad"
NODE="tcp://localhost:10501"
HOME="$HOME/.aura"
CHAIN_ID="aura-mvp-1"
TEST_ADDR="aura1g2nnl0x5hgvarpgfuwnyxfztss7h62yyj24k5p"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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
echo -e "${BLUE}======================================"
echo "1. TESTING BANK MODULE QUERIES"
echo "======================================${NC}"

test_command "bank" \
    "$AURAD query bank balances $TEST_ADDR --node $NODE --home $HOME" \
    "Bank: Query balances"

test_command "bank" \
    "$AURAD query bank total --node $NODE --home $HOME" \
    "Bank: Query total supply"

test_command "bank" \
    "$AURAD query bank denom-metadata --node $NODE --home $HOME" \
    "Bank: Query denom metadata"

# 2. DEX MODULE QUERIES
echo -e "${BLUE}======================================"
echo "2. TESTING DEX MODULE QUERIES"
echo "======================================${NC}"

test_command "dex" \
    "$AURAD query dex params --node $NODE --home $HOME" \
    "DEX: Query params"

test_command "dex" \
    "$AURAD query dex orderbook uaura usdc --node $NODE --home $HOME" \
    "DEX: Query orderbook"

test_command "dex" \
    "$AURAD query dex pools --node $NODE --home $HOME" \
    "DEX: List liquidity pools"

test_command "dex" \
    "$AURAD query dex supported-coins --node $NODE --home $HOME" \
    "DEX: Query supported coins"

test_command "dex" \
    "$AURAD query dex user-orders $TEST_ADDR --node $NODE --home $HOME" \
    "DEX: Query user orders"

# 3. COMPLIANCE MODULE QUERIES
echo -e "${BLUE}======================================"
echo "3. TESTING COMPLIANCE MODULE QUERIES"
echo "======================================${NC}"

test_command "compliance" \
    "$AURAD query compliance params --node $NODE --home $HOME" \
    "Compliance: Query params"

test_command "compliance" \
    "$AURAD query compliance kyc-record $TEST_ADDR --node $NODE --home $HOME" \
    "Compliance: Query KYC record"

test_command "compliance" \
    "$AURAD query compliance aml-profile $TEST_ADDR --node $NODE --home $HOME" \
    "Compliance: Query AML profile"

test_command "compliance" \
    "$AURAD query compliance sanctions $TEST_ADDR --node $NODE --home $HOME" \
    "Compliance: Query sanctions"

test_command "compliance" \
    "$AURAD query compliance alerts $TEST_ADDR --node $NODE --home $HOME" \
    "Compliance: Query alerts"

# 4. CONFIDENCE SCORE MODULE QUERIES
echo -e "${BLUE}======================================"
echo "4. TESTING CONFIDENCE SCORE QUERIES"
echo "======================================${NC}"

test_command "confidencescore" \
    "$AURAD query confidencescore params --node $NODE --home $HOME" \
    "ConfidenceScore: Query params"

test_command "confidencescore" \
    "$AURAD query confidencescore score $TEST_ADDR --node $NODE --home $HOME" \
    "ConfidenceScore: Query score"

test_command "confidencescore" \
    "$AURAD query confidencescore history $TEST_ADDR --node $NODE --home $HOME" \
    "ConfidenceScore: Query history"

test_command "confidencescore" \
    "$AURAD query confidencescore thresholds --node $NODE --home $HOME" \
    "ConfidenceScore: Query thresholds"

test_command "confidencescore" \
    "$AURAD query confidencescore verified-users --node $NODE --home $HOME" \
    "ConfidenceScore: Query verified users"

# 5. GOVERNANCE MODULE QUERIES
echo -e "${BLUE}======================================"
echo "5. TESTING GOVERNANCE MODULE QUERIES"
echo "======================================${NC}"

test_command "governance" \
    "$AURAD query governance params --node $NODE --home $HOME" \
    "Governance: Query params"

test_command "governance" \
    "$AURAD query governance proposals --node $NODE --home $HOME" \
    "Governance: List proposals"

# 6. PRIVACY MODULE QUERIES
echo -e "${BLUE}======================================"
echo "6. TESTING PRIVACY MODULE QUERIES"
echo "======================================${NC}"

test_command "privacy" \
    "$AURAD query privacy params --node $NODE --home $HOME" \
    "Privacy: Query params"

# 7. BRIDGE MODULE QUERIES
echo -e "${BLUE}======================================"
echo "7. TESTING BRIDGE MODULE QUERIES"
echo "======================================${NC}"

test_command "bridge" \
    "$AURAD query bridge params --node $NODE --home $HOME" \
    "Bridge: Query params"

# 8. CRYPTOGRAPHY MODULE QUERIES
echo -e "${BLUE}======================================"
echo "8. TESTING CRYPTOGRAPHY QUERIES"
echo "======================================${NC}"

test_command "cryptography" \
    "$AURAD query cryptography params --node $NODE --home $HOME" \
    "Cryptography: Query params"

# 9. DATA REGISTRY MODULE QUERIES
echo -e "${BLUE}======================================"
echo "9. TESTING DATA REGISTRY QUERIES"
echo "======================================${NC}"

test_command "dataregistry" \
    "$AURAD query dataregistry params --node $NODE --home $HOME" \
    "DataRegistry: Query params"

# 10. ECONOMIC SECURITY MODULE QUERIES
echo -e "${BLUE}======================================"
echo "10. TESTING ECONOMIC SECURITY QUERIES"
echo "======================================${NC}"

test_command "economicsecurity" \
    "$AURAD query economicsecurity params --node $NODE --home $HOME" \
    "EconomicSecurity: Query params"

# 11. IDENTITY CHANGE MODULE QUERIES
echo -e "${BLUE}======================================"
echo "11. TESTING IDENTITY CHANGE QUERIES"
echo "======================================${NC}"

test_command "identitychange" \
    "$AURAD query identitychange params --node $NODE --home $HOME" \
    "IdentityChange: Query params"

# 12. MONITORING MODULE QUERIES
echo -e "${BLUE}======================================"
echo "12. TESTING MONITORING QUERIES"
echo "======================================${NC}"

test_command "monitoring" \
    "$AURAD query monitoring params --node $NODE --home $HOME" \
    "Monitoring: Query params"

# 13. NETWORK SECURITY MODULE QUERIES
echo -e "${BLUE}======================================"
echo "13. TESTING NETWORK SECURITY QUERIES"
echo "======================================${NC}"

test_command "networksecurity" \
    "$AURAD query networksecurity params --node $NODE --home $HOME" \
    "NetworkSecurity: Query params"

# 14. PREVALIDATION MODULE QUERIES
echo -e "${BLUE}======================================"
echo "14. TESTING PREVALIDATION QUERIES"
echo "======================================${NC}"

test_command "prevalidation" \
    "$AURAD query prevalidation params --node $NODE --home $HOME" \
    "Prevalidation: Query params"

# 15. VALIDATOR SECURITY MODULE QUERIES
echo -e "${BLUE}======================================"
echo "15. TESTING VALIDATOR SECURITY QUERIES"
echo "======================================${NC}"

test_command "validatorsecurity" \
    "$AURAD query validatorsecurity params --node $NODE --home $HOME" \
    "ValidatorSecurity: Query params"

# 16. VC REGISTRY MODULE QUERIES
echo -e "${BLUE}======================================"
echo "16. TESTING VC REGISTRY QUERIES"
echo "======================================${NC}"

test_command "vcregistry" \
    "$AURAD query vcregistry params --node $NODE --home $HOME" \
    "VCRegistry: Query params"

# 17. WALLET SECURITY MODULE QUERIES
echo -e "${BLUE}======================================"
echo "17. TESTING WALLET SECURITY QUERIES"
echo "======================================${NC}"

test_command "walletsecurity" \
    "$AURAD query walletsecurity params --node $NODE --home $HOME" \
    "WalletSecurity: Query params"

# 18. AURA WASM SECURITY MODULE QUERIES
echo -e "${BLUE}======================================"
echo "18. TESTING AURA WASM SECURITY QUERIES"
echo "======================================${NC}"

test_command "aura_wasm_security" \
    "$AURAD query aura_wasm_security params --node $NODE --home $HOME" \
    "AuraWasmSecurity: Query params"

# 19. ACCOUNT QUERY
echo -e "${BLUE}======================================"
echo "19. TESTING ACCOUNT QUERY"
echo "======================================${NC}"

test_command "account" \
    "$AURAD query account $TEST_ADDR --node $NODE --home $HOME" \
    "Account: Query account info"

# SUMMARY
echo "=========================================="
echo "TEST SUMMARY"
echo "=========================================="
echo "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
if [ $TOTAL -gt 0 ]; then
    echo "Pass Rate: $(awk "BEGIN {printf \"%.1f\", ($PASSED/$TOTAL)*100}")%"
fi
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
    echo "Target: 29/29 commands"
    echo "Actual: $PASSED/$TOTAL commands"
else
    echo -e "${YELLOW}Some tests failed. See details above.${NC}"
fi
echo "=========================================="

exit $FAILED
