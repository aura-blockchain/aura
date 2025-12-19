#!/bin/bash
set -e

# Comprehensive Transaction Testing Script for Aura Blockchain
# Tests 22 transaction flows end-to-end (bank, staking, governance, DEX, WASM, security)
# Target: 100% success rate

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHAIN_DIR="${HOME}/.aura-test"
CHAIN_DIR_SECOND="${HOME}/.aura-test-val2"
BINARY="${SCRIPT_DIR}/../chain/aurad"
CHAIN_ID="aura-test-1"
RPC="tcp://localhost:26657"
WASM_FALLBACK="${SCRIPT_DIR}/../third_party/wasmvm/testdata/hackatom.wasm"
WASM_PRIMARY="${SCRIPT_DIR}/../contracts/artifacts/binding_tester.wasm"

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
    rm -rf "${CHAIN_DIR_SECOND}"
}

init_chain() {
    log_info "Initializing test chain with multi-denom support..."

    # Initialize chain
    # aurad init prints a mnemonic and prompts for Enter; provide a newline to keep this script non-interactive.
    printf '\n' | ${BINARY} init test-node --chain-id ${CHAIN_ID} --home ${CHAIN_DIR}
    printf '\n' | ${BINARY} init validator-two --chain-id ${CHAIN_ID} --home ${CHAIN_DIR_SECOND} > /dev/null 2>&1

    # Add validator key
    echo "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art" | \
        ${BINARY} keys add validator --recover --keyring-backend test --home ${CHAIN_DIR}

    # Add test users
    ${BINARY} keys add testuser --keyring-backend test --home ${CHAIN_DIR} 2>&1 | grep -v "mnemonic"
    ${BINARY} keys add testuser2 --keyring-backend test --home ${CHAIN_DIR} 2>&1 | grep -v "mnemonic"
    ${BINARY} keys add validator2 --keyring-backend test --home ${CHAIN_DIR} 2>&1 | grep -v "mnemonic"

    # Get addresses
    VAL_ADDR=$(${BINARY} keys show validator -a --keyring-backend test --home ${CHAIN_DIR})
    TEST_ADDR=$(${BINARY} keys show testuser -a --keyring-backend test --home ${CHAIN_DIR})
    TEST2_ADDR=$(${BINARY} keys show testuser2 -a --keyring-backend test --home ${CHAIN_DIR})
    VAL2_ADDR=$(${BINARY} keys show validator2 -a --keyring-backend test --home ${CHAIN_DIR})
    VAL2_OPER=$(${BINARY} keys show validator2 --bech val --keyring-backend test --home ${CHAIN_DIR} --output json | jq -r '.address')

    log_info "Validator address: ${VAL_ADDR}"
    log_info "Test user address: ${TEST_ADDR}"
    log_info "Second test user address: ${TEST2_ADDR}"
    log_info "Secondary validator account: ${VAL2_ADDR}"

    # Add genesis accounts with multiple denoms
    ${BINARY} genesis add-genesis-account ${VAL_ADDR} 1000000000000uaura,10000000000ubtc,10000000000usdt,10000000000ueth --home ${CHAIN_DIR}
    ${BINARY} genesis add-genesis-account ${TEST_ADDR} 1000000000uaura,1000000000ubtc,1000000000usdt,1000000000ueth --home ${CHAIN_DIR}
    ${BINARY} genesis add-genesis-account ${TEST2_ADDR} 1000000000uaura --home ${CHAIN_DIR}
    ${BINARY} genesis add-genesis-account ${VAL2_ADDR} 100000000uaura --home ${CHAIN_DIR}

    # Create genesis transaction
    ${BINARY} genesis gentx validator 100000000uaura --chain-id ${CHAIN_ID} --keyring-backend test --home ${CHAIN_DIR}

    # Collect genesis transactions
    ${BINARY} genesis collect-gentxs --home ${CHAIN_DIR}

    # Update config for faster blocks (1s)
    sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ${CHAIN_DIR}/config/config.toml

    # Rebind gRPC server to avoid host conflicts with other running services
    sed -i 's/address = "localhost:9090"/address = "localhost:29090"/g' ${CHAIN_DIR}/config/app.toml

    log_success "Chain initialized with multiple denoms: uaura, ubtc, usdt, ueth"
}

setup_secondary_validator() {
    log_info "Creating secondary validator for redelegation coverage..."
    VAL2_PUBKEY=$(jq -c '{ "@type": "/cosmos.crypto.ed25519.PubKey", "key": .pub_key.value }' "${CHAIN_DIR_SECOND}/config/priv_validator_key.json")
    local create_val_file
    create_val_file=$(mktemp)

    cat <<EOF > "${create_val_file}"
{
  "pubkey": ${VAL2_PUBKEY},
  "amount": "5000000uaura",
  "moniker": "validator-two",
  "identity": "",
  "website": "",
  "security": "",
  "details": "Secondary validator used in automation harness",
  "commission-rate": "0.050000000000000000",
  "commission-max-rate": "0.200000000000000000",
  "commission-max-change-rate": "0.010000000000000000",
  "min-self-delegation": "1"
}
EOF

    local result=0
    if ! ${BINARY} tx staking create-validator "${create_val_file}" \
        --from validator2 --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1; then
        result=1
    fi
    rm -f "${create_val_file}"
    sleep 2
    return ${result}
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

# Test 3: Multi-Send
test_multisend() {
    log_info "Test 3: Multi-Send Split Transfer"
    ${BINARY} tx bank multi-send ${VAL_ADDR} ${TEST_ADDR} ${TEST2_ADDR} 200000uaura \
        --split \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Bank multi-send (uaura split)" $?

    sleep 2
}

# Test 4: Staking Operations
test_staking() {
    log_info "Test 4: Staking Operations"

    # Get validator operator address
    VAL_OPER=$(${BINARY} keys show validator --bech val --keyring-backend test --home ${CHAIN_DIR} --output json | jq -r '.address')

    # Delegate
    ${BINARY} tx staking delegate ${VAL_OPER} 1000000uaura \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Staking delegate" $?

    sleep 2

    if setup_secondary_validator; then
        test_result "Create secondary validator" 0
    else
        test_result "Create secondary validator" 1
    fi

    sleep 2

    local redelegate_result=0
    if ! ${BINARY} tx staking redelegate ${VAL_OPER} ${VAL2_OPER} 300000uaura \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1; then
        redelegate_result=1
    fi
    test_result "Redelegate stake" ${redelegate_result}

    sleep 2

    # Withdraw rewards
    ${BINARY} tx distribution withdraw-rewards ${VAL_OPER} \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Withdraw staking rewards" $?

    sleep 2

    # Undelegate a portion of the stake to verify unbonding flow
    ${BINARY} tx staking unbond ${VAL_OPER} 500000uaura \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Undelegate stake" $?

    sleep 2
}

# Test 5: Governance
test_governance() {
    log_info "Test 5: Governance Operations"

    # Submit text proposal
    ${BINARY} tx governance submit-proposal \
        "Test Proposal" \
        "Testing governance" \
        text \
        --initial-deposit 10000000 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Submit governance proposal" $?

    sleep 2

    # Deposit to proposal from another participant
    ${BINARY} tx governance deposit 1 5000000 \
        --from testuser --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Deposit to governance proposal" $?

    sleep 2

    # Vote on proposal
    ${BINARY} tx governance vote 1 yes \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Vote on proposal" $?

    sleep 2
}

# Test 6: DEX HTLC
test_dex_htlc() {
    log_info "Test 6: DEX HTLC (Atomic Swap)"

    SECRET="test_secret_preimage_123456"
    HASH=$(echo -n "$SECRET" | sha256sum | awk '{print $1}')

    ${BINARY} tx dex create-htlc ${TEST_ADDR} 500000uaura ${HASH} 3600 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Create HTLC" $?

    sleep 2
}

# Test 7: DEX AMM Pool
test_dex_amm() {
    log_info "Test 7: DEX AMM Pool Operations"

    # Create pool (uaura/ubtc)
    ${BINARY} tx dex create-pool uaura 1000000 ubtc 1000000 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Create AMM pool (uaura/ubtc)" $?

    sleep 2

    # Add liquidity
    ${BINARY} tx dex add-liquidity 1 uaura 500000 ubtc 500000 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Add liquidity to pool" $?

    sleep 2

    # Swap
    ${BINARY} tx dex swap 1 100000uaura 90000 500 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Swap tokens in pool" $?

    sleep 2

    # Remove liquidity (burn LP tokens)
    ${BINARY} tx dex remove-liquidity 1 250000 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Remove liquidity from pool" $?

    sleep 2
}

# Test 8: Validator Security
test_validator_security() {
    log_info "Test 8: Validator Security"

    # Generate proper key hashes (32+ chars)
    HOT_KEY=$(echo -n "hot_key_validator_1" | sha256sum | awk '{print $1}')
    COLD_KEY=$(echo -n "cold_key_validator_1" | sha256sum | awk '{print $1}')

    ${BINARY} tx validatorsecurity register-validator \
        ${HOT_KEY} ${COLD_KEY} us-east US \
        --latitude 0 --longitude 0 \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1000uaura --broadcast-mode sync > /dev/null 2>&1
    test_result "Register validator security" $?

    sleep 2
}

# Test 9: Wallet Security
test_wallet_security() {
    log_info "Test 9: Wallet Security (Social Recovery)"

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

# Test 10: WASM Contract Lifecycle
test_wasm_contract() {
    log_info "Test 10: WASM Contract lifecycle (store → instantiate → execute)"

    local contract_path=""
    local init_msg=""
    local exec_msg=""
    if [ -f "${WASM_PRIMARY}" ]; then
        contract_path="${WASM_PRIMARY}"
    elif [ -f "${WASM_FALLBACK}" ]; then
        contract_path="${WASM_FALLBACK}"
    fi

    if [ -z "${contract_path}" ]; then
        log_error "No WASM artifacts found at ${WASM_PRIMARY} or ${WASM_FALLBACK}"
        test_result "Store WASM contract" 1
        test_result "Instantiate WASM contract" 1
        test_result "Execute WASM contract" 1
        return
    fi

    case "$(basename "${contract_path}")" in
        binding_tester.wasm)
            # contracts/binding-tester: InstantiateMsg {}, ExecuteMsg::RegisterVc { address, vc_base64 }
            init_msg="{}"
            exec_msg="{\"register_vc\":{\"address\":\"${TEST_ADDR}\",\"vc_base64\":\"dGVzdA==\"}}"
            ;;
        *)
            # Fallback (e.g., hackatom): InstantiateMsg {verifier,beneficiary}, ExecuteMsg::Release {}
            init_msg="{\"verifier\":\"${VAL_ADDR}\",\"beneficiary\":\"${TEST_ADDR}\"}"
            exec_msg='{"release":{}}'
            ;;
    esac

    local store_log
    local store_err
    store_err=$(mktemp)
    if ! store_log=$(${BINARY} tx aura_wasm_security store "${contract_path}" \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 2000uaura --broadcast-mode sync --gas auto --gas-adjustment 1.4 \
        --output json 2>"${store_err}"); then
        cat "${store_err}" >&2 || true
        echo "${store_log}"
        rm -f "${store_err}"
        test_result "Store WASM contract" 1
        test_result "Instantiate WASM contract" 1
        test_result "Execute WASM contract" 1
        return
    fi
    rm -f "${store_err}"
    test_result "Store WASM contract" 0
    sleep 2

    local code_id=""
    for _ in $(seq 1 10); do
        local code_json
        code_json=$(${BINARY} query aura_wasm_security list-code --home ${CHAIN_DIR} --output json 2>/dev/null || true)
        if [ -n "${code_json}" ]; then
            # Prefer explicit code_id when available; fall back to length-based heuristic for older outputs.
            code_id=$(echo "${code_json}" | jq -r '
                (.code_infos // []) as $infos
                | if ($infos | length) == 0 then "" else (($infos[-1].code_id // $infos[-1].codeID // ($infos | length)) | tostring) end
            ' 2>/dev/null || true)
        fi
        if [ -n "${code_id}" ]; then
            break
        fi
        sleep 2
    done

    if [ -z "${code_id}" ] || [ "${code_id}" = "null" ]; then
        log_error "Unable to determine stored code ID"
        test_result "Instantiate WASM contract" 1
        test_result "Execute WASM contract" 1
        return
    fi

    local instantiate_log
    local instantiate_err
    instantiate_err=$(mktemp)
    if ! instantiate_log=$(${BINARY} tx aura_wasm_security instantiate ${code_id} "${init_msg}" \
        --label "hackatom-${RANDOM}" --admin ${VAL_ADDR} --amount 100000uaura \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1500uaura --broadcast-mode sync --gas auto --gas-adjustment 1.2 \
        --output json 2>"${instantiate_err}"); then
        cat "${instantiate_err}" >&2 || true
        echo "${instantiate_log}"
        rm -f "${instantiate_err}"
        test_result "Instantiate WASM contract" 1
        test_result "Execute WASM contract" 1
        return
    fi
    rm -f "${instantiate_err}"
    test_result "Instantiate WASM contract" 0
    sleep 2

    if ! echo "${instantiate_log}" | jq -e . >/dev/null 2>&1; then
        log_error "Instantiate output was not valid JSON; unable to locate contract address"
        echo "${instantiate_log}"
        test_result "Execute WASM contract" 1
        return
    fi

    local contract_addr=""
    contract_addr=$(echo "${instantiate_log}" | jq -r '
        .logs[]?
        | .events[]?
        | select(.type=="instantiate" or .type=="wasm")
        | .attributes[]?
        | select(.key=="_contract_address" or .key=="contract_address")
        | .value' 2>/dev/null | tail -n 1) || contract_addr=""

    local instantiate_hash
    instantiate_hash=$(echo "${instantiate_log}" | jq -r '.txhash // empty' 2>/dev/null || true)

    if { [ -z "${contract_addr}" ] || [ "${contract_addr}" = "null" ]; } \
        && [ -n "${instantiate_hash}" ] && [ "${instantiate_hash}" != "null" ]; then
        log_info "Falling back to Tendermint RPC for contract address lookup..."
        for _ in $(seq 1 10); do
            local rpc_json
            rpc_json=$(curl -s "http://localhost:26657/tx?hash=0x${instantiate_hash}" 2>/dev/null || true)
            if [ -n "${rpc_json}" ]; then
                local candidate=""
                candidate=$(echo "${rpc_json}" | jq -r '
                    .result.tx_result.events[]?
                    | select(.type=="instantiate")
                    | .attributes[]?
                    | select(.key=="_contract_address")
                    | .value' 2>/dev/null | tail -n 1) || candidate=""
                if [ -n "${candidate}" ] && [ "${candidate}" != "null" ]; then
                    contract_addr="${candidate}"
                    break
                fi
            fi
            sleep 2
        done
    fi

    if [ -z "${contract_addr}" ] || [ "${contract_addr}" = "null" ]; then
        log_error "Unable to resolve contract address for code_id ${code_id}"
        test_result "Execute WASM contract" 1
        return
    fi

    local execute_log
    local execute_err
    execute_err=$(mktemp)
    if execute_log=$(${BINARY} tx aura_wasm_security execute "${contract_addr}" "${exec_msg}" \
        --from validator --keyring-backend test --chain-id ${CHAIN_ID} --home ${CHAIN_DIR} \
        --yes --fees 1200uaura --broadcast-mode sync --gas auto --gas-adjustment 1.2 \
        --output json 2>"${execute_err}"); then
        rm -f "${execute_err}"
        test_result "Execute WASM contract" 0
    else
        cat "${execute_err}" >&2 || true
        rm -f "${execute_err}"
        echo "${execute_log}"
        test_result "Execute WASM contract" 1
    fi

    sleep 2
}

# Main test execution
main() {
    echo "========================================="
    echo "  Aura Blockchain Transaction Testing"
    echo "  Target: 22/22 (100% Success Rate)"
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
    test_multisend
    test_staking
    test_governance
    test_dex_htlc
    test_dex_amm
    test_validator_security
    test_wallet_security
    test_wasm_contract

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

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
