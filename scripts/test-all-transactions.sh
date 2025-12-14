#!/bin/bash
set -e

# Comprehensive Transaction Testing Script for Aura Blockchain
# Tests all 10 transaction types end-to-end
# Target: 100% success rate

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHAIN_DIR="${HOME}/.aura-test"
BINARY="${SCRIPT_DIR}/../chain/aurad"
CHAIN_ID="aura-test-1"
RPC="tcp://localhost:26657"

# Color output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

log_error() {
    echo -e "${RED}✗ $1${NC}"
}

log_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Initialize counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

test_result() {
    local test_name="$1"
    local result="$2"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ "$result" = "0" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "$test_name"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "$test_name"
    fi
}

# Cleanup and initialization
cleanup() {
    log_info "Cleaning up previous test environment..."
    pkill aurad || true
    sleep 2
    rm -rf "${CHAIN_DIR}"
}

init_chain() {
    log_info "Initializing test chain with multi-denom support..."

    # Initialize chain
    ${BINARY} init test-node --chain-id ${CHAIN_ID} --home ${CHAIN_DIR}

    # Add validator key
    echo "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art" | \
        ${BINARY} keys add validator --recover --keyring-backend test --home ${CHAIN_DIR}

    # Add test user
    ${BINARY} keys add testuser --keyring-backend test --home ${CHAIN_DIR} 2>&1 | grep -v "mnemonic"

    # Get addresses
    VAL_ADDR=$(${BINARY} keys show validator -a --keyring-backend test --home ${CHAIN_DIR})
    TEST_ADDR=$(${BINARY} keys show testuser -a --keyring-backend test --home ${CHAIN_DIR})

    log_info "Validator address: ${VAL_ADDR}"
    log_info "Test user address: ${TEST_ADDR}"

    # Add genesis accounts with multiple denoms
    ${BINARY} genesis add-genesis-account ${VAL_ADDR} 1000000000000uaura,10000000000ubtc,10000000000usdt,10000000000ueth --home ${CHAIN_DIR}
    ${BINARY} genesis add-genesis-account ${TEST_ADDR} 1000000000uaura,1000000000ubtc,1000000000usdt,1000000000ueth --home ${CHAIN_DIR}

    # Create genesis transaction
    ${BINARY} genesis gentx validator 100000000uaura --chain-id ${CHAIN_ID} --keyring-backend test --home ${CHAIN_DIR}

    # Collect genesis transactions
    ${BINARY} genesis collect-gentxs --home ${CHAIN_DIR}

    # Update config for faster blocks (1s)
    sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ${CHAIN_DIR}/config/config.toml

    log_success "Chain initialized with multiple denoms: uaura, ubtc, usdt, ueth"
}

start_chain() {
    log_info "Starting blockchain..."
    ${BINARY} start --home ${CHAIN_DIR} --pruning=nothing > ${CHAIN_DIR}/node.log 2>&1 &
    CHAIN_PID=$!

    log_info "Waiting for chain to start (PID: ${CHAIN_PID})..."
    sleep 10

    # Verify chain is running
    if ! ps -p ${CHAIN_PID} > /dev/null; then
        log_error "Chain failed to start"
        cat ${CHAIN_DIR}/node.log
        exit 1
    fi

    log_success "Chain started successfully"
}

# Test 1: Bank Transfers
test_bank_transfer() {
    log_info "Test 1: Bank Transfer (uaura)"
    ${BINARY} tx bank send ${VAL_ADDR} ${TEST_ADDR} 5000000uaura \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Bank transfer (uaura)" $?

    sleep 2
}

# Test 2: Multi-Denom Transfers
test_multidenom_transfer() {
    log_info "Test 2: Multi-Denom Transfers"
    ${BINARY} tx bank send ${VAL_ADDR} ${TEST_ADDR} 1000000ubtc \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Bank transfer (ubtc)" $?

    sleep 2

    ${BINARY} tx bank send ${VAL_ADDR} ${TEST_ADDR} 1000000usdt \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Bank transfer (usdt)" $?

    sleep 2
}

# Test 3: Staking Operations
test_staking() {
    log_info "Test 3: Staking Operations"

    # Get validator operator address
    VAL_OPER=$(${BINARY} query staking validators --home ${CHAIN_DIR} --output json | jq -r '.validators[0].operator_address')

    # Delegate
    ${BINARY} tx staking delegate ${VAL_OPER} 1000000uaura \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Staking delegate" $?

    sleep 2

    # Withdraw rewards
    ${BINARY} tx distribution withdraw-rewards ${VAL_OPER} \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Withdraw staking rewards" $?

    sleep 2
}

# Test 4: Governance
test_governance() {
    log_info "Test 4: Governance Operations"

    # Submit text proposal
    ${BINARY} tx governance submit-proposal \
        --title "Test Proposal" \
        --description "Testing governance" \
        --category text \
        --initial-deposit 10000000uaura \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Submit governance proposal" $?

    sleep 2

    # Vote on proposal
    ${BINARY} tx governance vote 1 yes \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Vote on proposal" $?

    sleep 2
}

# Test 5: DEX HTLC
test_dex_htlc() {
    log_info "Test 5: DEX HTLC (Atomic Swap)"

    SECRET="test_secret_preimage_123456"
    HASH=$(echo -n "$SECRET" | sha256sum | awk '{print $1}')

    ${BINARY} tx dex create-htlc ${TEST_ADDR} 500000uaura ${HASH} 3600 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Create HTLC" $?

    sleep 2
}

# Test 6: DEX AMM Pool
test_dex_amm() {
    log_info "Test 6: DEX AMM Pool Operations"

    # Create pool (uaura/ubtc)
    ${BINARY} tx dex create-pool uaura ubtc 1000000uaura 1000000ubtc \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Create AMM pool (uaura/ubtc)" $?

    sleep 2

    # Add liquidity
    ${BINARY} tx dex add-liquidity 1 500000uaura 500000ubtc \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Add liquidity to pool" $?

    sleep 2

    # Swap
    ${BINARY} tx dex swap 1 100000uaura ubtc 90000 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Swap tokens in pool" $?

    sleep 2
}

# Test 7: Validator Security
test_validator_security() {
    log_info "Test 7: Validator Security"

    # Generate proper key hashes (32+ chars)
    HOT_KEY=$(echo -n "hot_key_validator_1" | sha256sum | awk '{print $1}')
    COLD_KEY=$(echo -n "cold_key_validator_1" | sha256sum | awk '{print $1}')

    ${BINARY} tx validatorsecurity register-validator \
        ${HOT_KEY} ${COLD_KEY} us-east US \
        --latitude 40.7128 --longitude -74.0060 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Register validator security" $?

    sleep 2
}

# Test 8: Wallet Security
test_wallet_security() {
    log_info "Test 8: Wallet Security (Social Recovery)"

    # Generate wallet ID (32+ chars)
    WALLET_ID=$(echo -n "test_wallet_001" | sha256sum | awk '{print $1}')

    ${BINARY} tx walletsecurity configure-social-recovery \
        ${WALLET_ID} \
        "${TEST_ADDR}" \
        1 "24h" \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Configure social recovery" $?

    sleep 2
}

# Main test execution
main() {
    echo "========================================="
    echo "  Aura Blockchain Transaction Testing"
    echo "  Target: 10/10 (100% Success Rate)"
    echo "========================================="
    echo

    cleanup
    init_chain
    start_chain

    echo
    echo "Running transaction tests..."
    echo

    test_bank_transfer
    test_multidenom_transfer
    test_staking
    test_governance
    test_dex_htlc
    test_dex_amm
    test_validator_security
    test_wallet_security

    echo
    echo "========================================="
    echo "  Test Results"
    echo "========================================="
    echo "Total Tests:  ${TOTAL_TESTS}"
    echo -e "Passed:       ${GREEN}${PASSED_TESTS}${NC}"
    echo -e "Failed:       ${RED}${FAILED_TESTS}${NC}"
    echo

    SUCCESS_RATE=$(echo "scale=1; ${PASSED_TESTS} * 100 / ${TOTAL_TESTS}" | bc)
    echo "Success Rate: ${SUCCESS_RATE}%"
    echo "========================================="

    # Cleanup
    log_info "Stopping blockchain..."
    kill ${CHAIN_PID} 2>/dev/null || true

    # Return exit code based on success
    if [ "${FAILED_TESTS}" = "0" ]; then
        exit 0
    else
        exit 1
    fi
}

main "$@"
