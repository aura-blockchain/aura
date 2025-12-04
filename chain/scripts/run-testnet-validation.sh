#!/bin/bash
set -e

#####################################################################
# Aura Testnet Validation Automation Script
#####################################################################
# Purpose: Automate testnet validation tests with logging and reporting
# Usage:
#   ./run-testnet-validation.sh --critical-only
#   ./run-testnet-validation.sh --all
#   ./run-testnet-validation.sh --category dex
#####################################################################

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHAIN_DIR="$(dirname "$SCRIPT_DIR")"
RESULTS_DIR="$HOME/testnet-validation-results"
RESULTS_FILE="$RESULTS_DIR/validation-$(date +%Y%m%d-%H%M%S).log"

# Default values
CHAIN_ID="${CHAIN_ID:-aura-testnet-1}"
NODE="${NODE:-http://localhost:26657}"
VALIDATOR_HOME="${VALIDATOR_HOME:-$HOME/.testnets/aura-testnet/node0/aurad}"
VALIDATOR_KEY="${VALIDATOR_KEY:-validator0}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Test mode
MODE="critical"  # critical, all, or specific category

#####################################################################
# Utility Functions
#####################################################################

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_test() {
    echo -e "${YELLOW}[TEST $1]${NC} $2"
}

print_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
}

print_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
}

print_skip() {
    echo -e "${YELLOW}⊘ SKIP${NC}: $1"
}

log_result() {
    echo "$1" | tee -a "$RESULTS_FILE"
}

increment_total() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

increment_passed() {
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

increment_failed() {
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

increment_skipped() {
    SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
}

#####################################################################
# Test Execution Functions
#####################################################################

run_test() {
    local test_num="$1"
    local test_name="$2"
    local test_cmd="$3"
    local expected_pattern="$4"
    local is_critical="$5"

    increment_total
    print_test "$test_num" "$test_name"

    # Skip non-critical tests if in critical-only mode
    if [ "$MODE" == "critical" ] && [ "$is_critical" != "true" ]; then
        print_skip "$test_name"
        increment_skipped
        log_result "Test $test_num: SKIPPED - $test_name"
        return 0
    fi

    # Execute test command
    local output
    if output=$(eval "$test_cmd" 2>&1); then
        # Check if output matches expected pattern (if provided)
        if [ -n "$expected_pattern" ]; then
            if echo "$output" | grep -q "$expected_pattern"; then
                print_pass "$test_name"
                increment_passed
                log_result "Test $test_num: PASS - $test_name"
                log_result "Output: $output"
            else
                print_fail "$test_name - Expected pattern not found: $expected_pattern"
                increment_failed
                log_result "Test $test_num: FAIL - $test_name"
                log_result "Output: $output"
                log_result "Expected: $expected_pattern"

                # Fail fast for critical tests
                if [ "$is_critical" == "true" ]; then
                    echo -e "${RED}Critical test failed. Stopping validation.${NC}"
                    exit 1
                fi
            fi
        else
            print_pass "$test_name"
            increment_passed
            log_result "Test $test_num: PASS - $test_name"
            log_result "Output: $output"
        fi
    else
        print_fail "$test_name - Command failed"
        increment_failed
        log_result "Test $test_num: FAIL - $test_name"
        log_result "Error: $output"

        # Fail fast for critical tests
        if [ "$is_critical" == "true" ]; then
            echo -e "${RED}Critical test failed. Stopping validation.${NC}"
            exit 1
        fi
    fi

    echo ""
}

#####################################################################
# Test Categories
#####################################################################

run_basic_chain_tests() {
    print_header "Category 1: Basic Chain Operations"

    run_test "1" "Query Chain Status" \
        "aurad status --node $NODE --output json" \
        "catching_up" \
        "true"

    run_test "2" "Check Validator Set" \
        "aurad query staking validators --node $NODE --output json" \
        "validators" \
        "true"

    run_test "3" "Query Latest Block" \
        "aurad query block --node $NODE --output json" \
        "block" \
        "true"

    run_test "4" "Check Transaction Mempool" \
        "curl -s $NODE/num_unconfirmed_txs" \
        "n_txs" \
        "true"

    run_test "5" "Query Network Info" \
        "curl -s $NODE/net_info" \
        "peers" \
        "true"

    run_test "6" "Check Consensus State" \
        "curl -s $NODE/consensus_state" \
        "round_state" \
        "true"

    run_test "7" "Query Module Accounts" \
        "aurad query auth module-accounts --node $NODE --output json" \
        "accounts" \
        "true"

    run_test "8" "Check Bank Total Supply" \
        "aurad query bank total --node $NODE --output json" \
        "supply" \
        "true"
}

run_account_tests() {
    print_header "Category 2: Account Operations"

    run_test "9" "List Keys in Keyring" \
        "aurad keys list --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND" \
        "$VALIDATOR_KEY" \
        "true"

    # Test 10: Create test account (skip if already exists)
    if ! aurad keys show testuser1 --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND &>/dev/null; then
        run_test "10" "Create Test Account" \
            "echo -e '\\n' | aurad keys add testuser1 --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND" \
            "address:" \
            "true"
    else
        print_skip "Test 10: Test account already exists"
        increment_total
        increment_skipped
        log_result "Test 10: SKIPPED - Test account already exists"
    fi

    # Get validator address
    VALIDATOR_ADDR=$(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --address)

    run_test "11" "Query Account Balance (Validator)" \
        "aurad query bank balances $VALIDATOR_ADDR --node $NODE --output json" \
        "balances" \
        "true"

    run_test "12" "Query Account Info" \
        "aurad query auth account $VALIDATOR_ADDR --node $NODE --output json" \
        "account" \
        "true"

    # Test 13: Send transfer
    TEST_ADDR=$(aurad keys show testuser1 --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --address)

    run_test "13" "Send Simple Transfer" \
        "aurad tx bank send $VALIDATOR_ADDR $TEST_ADDR 1000000uaura --from $VALIDATOR_KEY --chain-id $CHAIN_ID --node $NODE --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --fees 5000uaura --yes --output json" \
        "txhash" \
        "true"

    # Wait for tx to be included
    sleep 5

    run_test "14" "Verify Transfer Completed" \
        "aurad query bank balances $TEST_ADDR --node $NODE --output json" \
        "1000000uaura" \
        "false"

    # Test 15 requires TX hash from test 13 - skip for now
    print_skip "Test 15: Query Transaction by Hash (manual verification required)"
    increment_total
    increment_skipped
}

run_dex_tests() {
    print_header "Category 3A: DEX Module Tests"

    VALIDATOR_ADDR=$(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --address)

    # Note: These tests require the DEX module to be properly configured
    # May need to create test tokens first

    print_skip "Test 16-20: DEX tests require pre-configuration (manual execution recommended)"
    for i in {16..20}; do
        increment_total
        increment_skipped
    done
}

run_bridge_tests() {
    print_header "Category 3B: Bridge Module Tests"

    print_skip "Test 21-23: Bridge tests require cross-chain setup (manual execution recommended)"
    for i in {21..23}; do
        increment_total
        increment_skipped
    done
}

run_compliance_tests() {
    print_header "Category 3C: Compliance Module Tests"

    TEST_ADDR=$(aurad keys show testuser1 --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --address 2>/dev/null || echo "")

    if [ -z "$TEST_ADDR" ]; then
        print_skip "Test 24-26: Compliance tests require testuser1 account"
        for i in {24..26}; do
            increment_total
            increment_skipped
        done
        return
    fi

    # Create PII commitment
    PII_HASH=$(echo -n '{"name":"Test User","dob":"1990-01-01"}' | sha256sum | awk '{print $1}')

    run_test "24" "Submit KYC Record" \
        "aurad tx compliance submit-kyc $TEST_ADDR 3 kyc-provider-1 $PII_HASH US --from $VALIDATOR_KEY --chain-id $CHAIN_ID --node $NODE --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --fees 5000uaura --yes --output json" \
        "txhash" \
        "true"

    sleep 3

    run_test "25" "Query KYC Record" \
        "aurad query compliance kyc-record $TEST_ADDR --node $NODE --output json" \
        "kyc_level" \
        "false"

    run_test "26" "Screen Sanctions" \
        "aurad tx compliance screen-sanctions $TEST_ADDR false --from $TEST_ADDR --chain-id $CHAIN_ID --node $NODE --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --fees 5000uaura --yes --output json" \
        "txhash" \
        "false"
}

run_identity_tests() {
    print_header "Category 3D: Identity Module Tests"

    print_skip "Test 27-28: Identity tests require manual verification"
    for i in {27..28}; do
        increment_total
        increment_skipped
    done
}

run_wasm_tests() {
    print_header "Category 4: WASM Contract Tests"

    # Check if WASM contract exists
    WASM_CONTRACT="$CHAIN_DIR/../contracts/artifacts/vc_issuer.wasm"

    if [ ! -f "$WASM_CONTRACT" ]; then
        print_skip "Test 29-33: WASM contract not found at $WASM_CONTRACT"
        for i in {29..33}; do
            increment_total
            increment_skipped
        done
        return
    fi

    run_test "29" "Store WASM Contract Code" \
        "aurad tx wasm store $WASM_CONTRACT --from $VALIDATOR_KEY --chain-id $CHAIN_ID --node $NODE --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --gas auto --gas-adjustment 1.3 --fees 50000uaura --yes --output json" \
        "code_id" \
        "true"

    sleep 3

    run_test "30" "Query Stored WASM Code" \
        "aurad query wasm code 1 --node $NODE --output json" \
        "creator" \
        "true"

    # Tests 31-33 require contract instantiation - skip for automation
    print_skip "Test 31-33: WASM execution tests require manual verification"
    for i in {31..33}; do
        increment_total
        increment_skipped
    done
}

run_governance_tests() {
    print_header "Category 5A: Governance Tests"

    VALIDATOR_ADDR=$(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --address)

    run_test "34" "Submit Governance Proposal" \
        "aurad tx governance submit-proposal 'Test Parameter Change' 'Testing governance mechanism' 1 $VALIDATOR_ADDR 1000000uaura false --from $VALIDATOR_KEY --chain-id $CHAIN_ID --node $NODE --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --fees 5000uaura --yes --output json" \
        "proposal_id" \
        "true"

    sleep 3

    run_test "35" "Add Deposit to Proposal" \
        "aurad tx governance deposit 1 $VALIDATOR_ADDR 500000uaura --from $VALIDATOR_KEY --chain-id $CHAIN_ID --node $NODE --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --fees 5000uaura --yes --output json" \
        "txhash" \
        "false"

    sleep 3

    run_test "36" "Vote on Proposal" \
        "aurad tx governance vote 1 $VALIDATOR_ADDR 1 false '' --from $VALIDATOR_KEY --chain-id $CHAIN_ID --node $NODE --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --fees 5000uaura --yes --output json" \
        "txhash" \
        "false"

    run_test "37" "Query Proposal Status" \
        "aurad query governance proposal 1 --node $NODE --output json" \
        "proposal_id" \
        "false"
}

run_staking_tests() {
    print_header "Category 5B: Staking Operations"

    run_test "38" "Query Staking Pool" \
        "aurad query staking pool --node $NODE --output json" \
        "bonded_tokens" \
        "true"

    VALIDATOR_OPER=$(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --bech val --address)

    run_test "39" "Query Validator Details" \
        "aurad query staking validator $VALIDATOR_OPER --node $NODE --output json" \
        "operator_address" \
        "false"

    run_test "40" "Query Delegations" \
        "aurad query staking delegations-to $VALIDATOR_OPER --node $NODE --output json" \
        "delegation_responses" \
        "false"
}

run_distribution_tests() {
    print_header "Category 5C: Distribution & Rewards"

    VALIDATOR_OPER=$(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend $KEYRING_BACKEND --bech val --address)

    run_test "41" "Query Outstanding Rewards" \
        "aurad query distribution validator-outstanding-rewards $VALIDATOR_OPER --node $NODE --output json" \
        "rewards" \
        "false"

    run_test "42" "Query Community Pool" \
        "aurad query distribution community-pool --node $NODE --output json" \
        "pool" \
        "false"
}

#####################################################################
# Main Execution
#####################################################################

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    --critical-only     Run only critical tests (default)
    --all               Run all tests including nice-to-have
    --category <name>   Run specific category: basic, accounts, dex, bridge, compliance, identity, wasm, governance, staking, distribution
    --help              Show this help message

Environment Variables:
    CHAIN_ID            Chain ID (default: aura-testnet-1)
    NODE                Node RPC URL (default: http://localhost:26657)
    VALIDATOR_HOME      Validator home directory (default: ~/.testnets/aura-testnet/node0/aurad)
    VALIDATOR_KEY       Validator key name (default: validator0)
    KEYRING_BACKEND     Keyring backend (default: test)

Examples:
    $0 --critical-only
    $0 --all
    $0 --category dex
    CHAIN_ID=aura-testnet-2 $0 --all

EOF
}

main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --critical-only)
                MODE="critical"
                shift
                ;;
            --all)
                MODE="all"
                shift
                ;;
            --category)
                MODE="category"
                CATEGORY="$2"
                shift 2
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    # Create results directory
    mkdir -p "$RESULTS_DIR"

    # Print header
    print_header "Aura Testnet Validation Suite"
    echo "Date: $(date)"
    echo "Chain ID: $CHAIN_ID"
    echo "Node: $NODE"
    echo "Validator Home: $VALIDATOR_HOME"
    echo "Mode: $MODE"
    echo ""

    # Log configuration
    log_result "Aura Testnet Validation - $(date)"
    log_result "Chain ID: $CHAIN_ID"
    log_result "Node: $NODE"
    log_result "Validator Home: $VALIDATOR_HOME"
    log_result "Mode: $MODE"
    log_result "---"

    # Check node is accessible
    if ! curl -s "$NODE/health" &>/dev/null; then
        echo -e "${RED}ERROR: Cannot connect to node at $NODE${NC}"
        echo "Please ensure the testnet is running and NODE is set correctly."
        exit 1
    fi

    # Run tests based on mode
    if [ "$MODE" == "category" ]; then
        case "$CATEGORY" in
            basic)
                run_basic_chain_tests
                ;;
            accounts)
                run_account_tests
                ;;
            dex)
                run_dex_tests
                ;;
            bridge)
                run_bridge_tests
                ;;
            compliance)
                run_compliance_tests
                ;;
            identity)
                run_identity_tests
                ;;
            wasm)
                run_wasm_tests
                ;;
            governance)
                run_governance_tests
                ;;
            staking)
                run_staking_tests
                ;;
            distribution)
                run_distribution_tests
                ;;
            *)
                echo "Unknown category: $CATEGORY"
                show_usage
                exit 1
                ;;
        esac
    else
        # Run all or critical tests
        run_basic_chain_tests
        run_account_tests
        run_dex_tests
        run_bridge_tests
        run_compliance_tests
        run_identity_tests
        run_wasm_tests
        run_governance_tests
        run_staking_tests
        run_distribution_tests
    fi

    # Print summary
    print_header "Test Summary"
    echo "Total Tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo -e "Skipped: ${YELLOW}$SKIPPED_TESTS${NC}"
    echo ""

    # Calculate pass rate
    if [ $TOTAL_TESTS -gt 0 ]; then
        EXECUTED_TESTS=$((TOTAL_TESTS - SKIPPED_TESTS))
        if [ $EXECUTED_TESTS -gt 0 ]; then
            PASS_RATE=$(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS / $EXECUTED_TESTS) * 100}")
            echo "Pass Rate: $PASS_RATE% ($PASSED_TESTS/$EXECUTED_TESTS executed)"
        fi
    fi

    echo ""
    echo "Results saved to: $RESULTS_FILE"

    # Log summary
    log_result "---"
    log_result "Test Summary:"
    log_result "Total Tests: $TOTAL_TESTS"
    log_result "Passed: $PASSED_TESTS"
    log_result "Failed: $FAILED_TESTS"
    log_result "Skipped: $SKIPPED_TESTS"

    # Exit with appropriate code
    if [ $FAILED_TESTS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

# Run main function
main "$@"
