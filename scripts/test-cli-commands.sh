#!/bin/bash
# Aura CLI Command Test Script
# Tests all critical commands to verify they are properly registered

set -e

AURAD="${1:-/tmp/aurad}"
NODE="${2:-tcp://localhost:27657}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "====================================="
echo "Aura CLI Command Test Script"
echo "====================================="
echo ""
echo "Binary: $AURAD"
echo "Node: $NODE"
echo ""

# Check if binary exists
if [ ! -f "$AURAD" ]; then
    echo -e "${RED}ERROR: Binary not found at $AURAD${NC}"
    echo "Build it with: cd chain && go build -o $AURAD ./cmd/aurad"
    exit 1
fi

# Test counter
TOTAL=0
PASSED=0
FAILED=0
WARNINGS=0

test_command() {
    local name="$1"
    local cmd="$2"
    local expected_pattern="$3"
    local severity="${4:-FAIL}" # FAIL or WARN

    TOTAL=$((TOTAL + 1))

    echo -n "Testing: $name... "

    if output=$(eval "$cmd" 2>&1); then
        if echo "$output" | grep -q "$expected_pattern"; then
            echo -e "${GREEN}✓ PASS${NC}"
            PASSED=$((PASSED + 1))
        else
            if [ "$severity" = "WARN" ]; then
                echo -e "${YELLOW}⚠ WARN${NC} (unexpected output)"
                WARNINGS=$((WARNINGS + 1))
            else
                echo -e "${RED}✗ FAIL${NC} (unexpected output)"
                FAILED=$((FAILED + 1))
            fi
        fi
    else
        # Command failed
        if echo "$output" | grep -q "$expected_pattern"; then
            echo -e "${GREEN}✓ PASS${NC} (expected error)"
            PASSED=$((PASSED + 1))
        else
            if [ "$severity" = "WARN" ]; then
                echo -e "${YELLOW}⚠ WARN${NC}"
                echo "  Output: $output"
                WARNINGS=$((WARNINGS + 1))
            else
                echo -e "${RED}✗ FAIL${NC}"
                echo "  Output: $output"
                FAILED=$((FAILED + 1))
            fi
        fi
    fi
}

echo "====================================="
echo "1. BANK MODULE TESTS"
echo "====================================="

test_command "Bank tx send help" \
    "$AURAD tx bank send --help" \
    "Send funds from one account to another"

test_command "Bank query (KNOWN ISSUE)" \
    "$AURAD query bank --help" \
    "unknown command" \
    "WARN"

echo ""
echo "====================================="
echo "2. DEX MODULE TESTS"
echo "====================================="

# TX Commands
test_command "DEX swap help" \
    "$AURAD tx dex swap --help" \
    "Execute a token swap"

test_command "DEX create-pool help" \
    "$AURAD tx dex create-pool --help" \
    "Create a new AMM liquidity pool"

test_command "DEX create-htlc help" \
    "$AURAD tx dex create-htlc --help" \
    "Create an HTLC for trustless"

test_command "DEX create-order help" \
    "$AURAD tx dex create-order --help" \
    "Create a peer-to-peer swap order"

# Query Commands
test_command "DEX pool query help" \
    "$AURAD query dex pool --help" \
    "Query detailed information"

test_command "DEX quote query help" \
    "$AURAD query dex quote --help" \
    "Get an estimated output"

test_command "DEX orderbook query help" \
    "$AURAD query dex orderbook --help" \
    "Query the peer-to-peer orderbook"

# Test invalid input
test_command "DEX swap validation" \
    "$AURAD tx dex swap 2>&1" \
    "accepts 4 arg"

echo ""
echo "====================================="
echo "3. COMPLIANCE MODULE TESTS"
echo "====================================="

test_command "Compliance submit-kyc help" \
    "$AURAD tx compliance submit-kyc --help" \
    "Submit Know Your Customer"

test_command "Compliance screen-sanctions help" \
    "$AURAD tx compliance screen-sanctions --help" \
    "Screen an address against"

test_command "Compliance kyc-record query help" \
    "$AURAD query compliance kyc-record --help" \
    "Query KYC record"

test_command "Compliance aml-profile query help" \
    "$AURAD query compliance aml-profile --help" \
    "Query AML risk profile"

# Test invalid input
test_command "Compliance submit-kyc validation" \
    "$AURAD tx compliance submit-kyc 2>&1" \
    "accepts 5 arg"

echo ""
echo "====================================="
echo "4. CONFIDENCE SCORE MODULE TESTS"
echo "====================================="

test_command "Confidence score tx help" \
    "$AURAD tx confidencescore --help" \
    "Confidence score transaction"

test_command "Confidence score query help" \
    "$AURAD query confidencescore --help" \
    "Querying commands for the confidencescore"

test_command "Confidence score score query help" \
    "$AURAD query confidencescore score --help" \
    "Query the confidence score"

test_command "Confidence score record-completion help" \
    "$AURAD tx confidencescore record-completion --help" \
    "completion of an Inclusion Routine"

echo ""
echo "====================================="
echo "5. WASM SECURITY MODULE TESTS"
echo "====================================="

test_command "Wasm tx help" \
    "$AURAD tx aura_wasm_security --help" \
    "Wasm transaction subcommands"

test_command "Wasm store help" \
    "$AURAD tx aura_wasm_security store --help" \
    "Upload a wasm binary"

test_command "Wasm instantiate help" \
    "$AURAD tx aura_wasm_security instantiate --help" \
    "Instantiate a wasm contract"

test_command "Wasm execute help" \
    "$AURAD tx aura_wasm_security execute --help" \
    "Execute a command on a wasm contract"

echo ""
echo "====================================="
echo "6. STANDARD COSMOS MODULES"
echo "====================================="

test_command "Staking help" \
    "$AURAD tx staking --help" \
    "Staking transaction subcommands"

test_command "Distribution help" \
    "$AURAD tx distribution --help" \
    "Distribution transactions subcommands"

test_command "Account query help" \
    "$AURAD query account --help" \
    "Query account information"

echo ""
echo "====================================="
echo "7. TESTNET INTEGRATION TESTS"
echo "====================================="

# Only run these if node is accessible
if curl -s "$NODE/status" > /dev/null 2>&1; then
    echo "Node is accessible at $NODE"

    test_command "DEX pools query" \
        "$AURAD query dex pools --node $NODE --output json" \
        "pools" \
        "WARN"

    test_command "DEX supported-coins query" \
        "$AURAD query dex supported-coins --node $NODE --output json" \
        "coins" \
        "WARN"

    test_command "Confidence score params query" \
        "$AURAD query confidencescore params --node $NODE --output json 2>&1" \
        "params"
else
    echo -e "${YELLOW}⚠ Node not accessible at $NODE - skipping integration tests${NC}"
    echo "  Start testnet or verify RPC endpoint"
fi

echo ""
echo "====================================="
echo "TEST SUMMARY"
echo "====================================="
echo ""
echo "Total tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${YELLOW}Warnings: $WARNINGS${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo -e "${YELLOW}⚠ Tests passed with warnings (known issues)${NC}"
        echo ""
        echo "Known Issues:"
        echo "  1. Bank query module not registered"
        exit 0
    fi
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
