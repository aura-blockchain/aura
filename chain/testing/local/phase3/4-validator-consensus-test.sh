#!/bin/bash
#
# 4-Validator Consensus Test Suite
# Tests BFT consensus properties with 4 validators
#
# Requirements:
# - 4 validators running (setup by phase3 setup script)
# - Each validator has equal voting power (25% each)
#
# BFT Consensus Rules:
# - Requires >2/3 voting power to produce blocks
# - 75% (3 validators) > 2/3 → Should produce blocks
# - 50% (2 validators) < 2/3 → Should HALT
# - 25% (1 validator) < 2/3 → Should HALT
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test configuration
RESULTS_DIR="/home/hudson/blockchain-projects/aura/chain/testing/local/phase3"
RESULTS_FILE="$RESULTS_DIR/consensus_test_results_$(date +%Y%m%d_%H%M%S).log"
BINARY="/home/hudson/blockchain-projects/aura/chain/aurad"
CHAIN_ID="aura-local-testnet"
WAIT_BLOCKS=15
BLOCK_TIME=5

# Validator configuration
VAL1_HOME="/home/hudson/.aura/validator1"
VAL2_HOME="/home/hudson/.aura/validator2"
VAL3_HOME="/home/hudson/.aura/validator3"
VAL4_HOME="/home/hudson/.aura/validator4"

VAL1_RPC="http://localhost:26657"
VAL2_RPC="http://localhost:26667"
VAL3_RPC="http://localhost:26677"
VAL4_RPC="http://localhost:26687"

# Function to print colored output
print_header() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}\n"
}

print_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Function to log to both console and file
log() {
    echo -e "$1" | tee -a "$RESULTS_FILE"
}

# Function to get current block height from RPC
get_block_height() {
    local rpc_url=$1
    local height
    height=$(curl -s "$rpc_url/status" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0")
    echo "$height"
}

# Function to get validator count from RPC
get_validator_count() {
    local rpc_url=$1
    local count
    count=$(curl -s "$rpc_url/validators" | jq -r '.result.total' 2>/dev/null || echo "0")
    echo "$count"
}

# Function to get voting power from RPC
get_voting_power() {
    local rpc_url=$1
    local power
    power=$(curl -s "$rpc_url/validators" | jq -r '[.result.validators[].voting_power | tonumber] | add' 2>/dev/null || echo "0")
    echo "$power"
}

# Function to check if node is running
is_node_running() {
    local rpc_url=$1
    if curl -s "$rpc_url/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to get validator status summary
get_validator_status() {
    local val_name=$1
    local rpc_url=$2

    if is_node_running "$rpc_url"; then
        local height=$(get_block_height "$rpc_url")
        echo -e "${GREEN}RUNNING${NC} (height: $height)"
    else
        echo -e "${RED}STOPPED${NC}"
    fi
}

# Function to display current network status
display_network_status() {
    log "\n$(print_info "Network Status:")"
    log "  Validator 1: $(get_validator_status "validator1" "$VAL1_RPC")"
    log "  Validator 2: $(get_validator_status "validator2" "$VAL2_RPC")"
    log "  Validator 3: $(get_validator_status "validator3" "$VAL3_RPC")"
    log "  Validator 4: $(get_validator_status "validator4" "$VAL4_RPC")"

    # Calculate active voting power
    local active_validators=0
    is_node_running "$VAL1_RPC" && ((active_validators++)) || true
    is_node_running "$VAL2_RPC" && ((active_validators++)) || true
    is_node_running "$VAL3_RPC" && ((active_validators++)) || true
    is_node_running "$VAL4_RPC" && ((active_validators++)) || true

    local voting_power_pct=$((active_validators * 25))
    log "  Active Validators: $active_validators/4"
    log "  Active Voting Power: ${voting_power_pct}%"

    if [ $voting_power_pct -gt 66 ]; then
        log "  Consensus Status: ${GREEN}CAN PRODUCE BLOCKS${NC} (>66% voting power)"
    else
        log "  Consensus Status: ${RED}HALTED${NC} (<66% voting power)"
    fi
}

# Function to stop a validator
stop_validator() {
    local val_home=$1
    local val_name=$2

    print_info "Stopping $val_name..."

    # Find PID using ps and grep
    local pid=$(ps aux | grep "aurad start --home $val_home" | grep -v grep | awk '{print $2}' | head -1)

    if [ -n "$pid" ]; then
        kill "$pid" 2>/dev/null || true
        sleep 2
        log "$(print_success "$val_name stopped (PID: $pid)")"
    else
        log "$(print_info "$val_name was not running")"
    fi
}

# Function to start a validator
start_validator() {
    local val_home=$1
    local val_name=$2
    local rpc_port=$3
    local p2p_port=$4
    local grpc_port=$5

    print_info "Starting $val_name..."

    nohup "$BINARY" start \
        --home "$val_home" \
        --rpc.laddr "tcp://0.0.0.0:$rpc_port" \
        --p2p.laddr "tcp://0.0.0.0:$p2p_port" \
        --grpc.address "0.0.0.0:$grpc_port" \
        > "$RESULTS_DIR/${val_name}_$(date +%Y%m%d_%H%M%S).log" 2>&1 &

    sleep 3

    if is_node_running "http://localhost:$rpc_port"; then
        log "$(print_success "$val_name started successfully")"
    else
        log "$(print_error "$val_name failed to start")"
        return 1
    fi
}

# Function to wait and monitor block production
monitor_block_production() {
    local duration=$1
    local rpc_url=$2
    local description=$3

    print_info "Monitoring block production for ${duration}s ($description)..."

    local start_height=$(get_block_height "$rpc_url")
    local start_time=$(date +%s)

    sleep "$duration"

    local end_height=$(get_block_height "$rpc_url")
    local end_time=$(date +%s)
    local elapsed=$((end_time - start_time))
    local blocks_produced=$((end_height - start_height))

    log "\n  Start Height: $start_height"
    log "  End Height: $end_height"
    log "  Blocks Produced: $blocks_produced in ${elapsed}s"

    if [ "$blocks_produced" -gt 0 ]; then
        local block_rate=$(echo "scale=2; $blocks_produced / $elapsed" | bc)
        log "  Block Rate: ${block_rate} blocks/sec"
        log "  $(print_success "Chain is producing blocks")"
        return 0
    else
        log "  $(print_error "Chain is HALTED - no blocks produced")"
        return 1
    fi
}

# Function to run a consensus test
run_consensus_test() {
    local test_num=$1
    local test_description=$2
    local expected_result=$3
    shift 3
    local active_validators=("$@")

    log "\n$(print_header "TEST $test_num: $test_description")"

    # Calculate expected voting power
    local validator_count=${#active_validators[@]}
    local voting_power=$((validator_count * 25))

    log "$(print_test "Active Validators: $validator_count/4")"
    log "$(print_test "Active Voting Power: ${voting_power}%")"
    log "$(print_test "Expected Result: $expected_result")"

    # Display current status
    display_network_status

    # Monitor block production
    local monitor_rpc="${active_validators[0]}"
    local blocks_produced=false

    if monitor_block_production "$WAIT_BLOCKS" "$monitor_rpc" "$test_description"; then
        blocks_produced=true
    fi

    # Verify expected result
    log "\n$(print_header "TEST $test_num VERIFICATION")"

    if [ "$expected_result" = "PRODUCE_BLOCKS" ]; then
        if [ "$blocks_produced" = true ]; then
            log "$(print_success "PASS: Chain produced blocks as expected with ${voting_power}% voting power")"
            return 0
        else
            log "$(print_error "FAIL: Chain halted unexpectedly with ${voting_power}% voting power")"
            return 1
        fi
    elif [ "$expected_result" = "HALT" ]; then
        if [ "$blocks_produced" = false ]; then
            log "$(print_success "PASS: Chain halted as expected with ${voting_power}% voting power")"
            return 0
        else
            log "$(print_error "FAIL: Chain produced blocks unexpectedly with ${voting_power}% voting power")"
            return 1
        fi
    fi
}

# Main test execution
main() {
    log "$(print_header "4-VALIDATOR CONSENSUS TEST SUITE")"
    log "Started: $(date)"
    log "Results file: $RESULTS_FILE"

    # Verify binary exists
    if [ ! -f "$BINARY" ]; then
        log "$(print_error "Binary not found: $BINARY")"
        log "Please build the binary first: cd chain && make build"
        exit 1
    fi

    # Verify all validator homes exist
    for home in "$VAL1_HOME" "$VAL2_HOME" "$VAL3_HOME" "$VAL4_HOME"; do
        if [ ! -d "$home" ]; then
            log "$(print_error "Validator home not found: $home")"
            log "Please run the 4-validator setup script first"
            exit 1
        fi
    done

    # Initial network status
    log "\n$(print_header "INITIAL NETWORK STATUS")"
    display_network_status

    # Wait for initial sync
    print_info "Waiting for network to stabilize..."
    sleep 5

    local test_results=()

    # TEST 1: All 4 validators running (100% voting power)
    if run_consensus_test \
        "1" \
        "All 4 validators running (100% voting power)" \
        "PRODUCE_BLOCKS" \
        "$VAL1_RPC" "$VAL2_RPC" "$VAL3_RPC" "$VAL4_RPC"; then
        test_results+=("TEST 1: PASS")
    else
        test_results+=("TEST 1: FAIL")
    fi

    # TEST 2: 3 validators running (75% voting power > 2/3)
    log "\n$(print_header "PREPARING TEST 2")"
    stop_validator "$VAL4_HOME" "validator4"
    sleep 5

    if run_consensus_test \
        "2" \
        "3 validators running (75% voting power > 2/3)" \
        "PRODUCE_BLOCKS" \
        "$VAL1_RPC" "$VAL2_RPC" "$VAL3_RPC"; then
        test_results+=("TEST 2: PASS")
    else
        test_results+=("TEST 2: FAIL")
    fi

    # TEST 3: 2 validators running (50% voting power < 2/3)
    log "\n$(print_header "PREPARING TEST 3")"
    stop_validator "$VAL3_HOME" "validator3"
    sleep 5

    if run_consensus_test \
        "3" \
        "2 validators running (50% voting power < 2/3)" \
        "HALT" \
        "$VAL1_RPC" "$VAL2_RPC"; then
        test_results+=("TEST 3: PASS")
    else
        test_results+=("TEST 3: FAIL")
    fi

    # TEST 4: 1 validator running (25% voting power < 2/3)
    log "\n$(print_header "PREPARING TEST 4")"
    stop_validator "$VAL2_HOME" "validator2"
    sleep 5

    if run_consensus_test \
        "4" \
        "1 validator running (25% voting power < 2/3)" \
        "HALT" \
        "$VAL1_RPC"; then
        test_results+=("TEST 4: PASS")
    else
        test_results+=("TEST 4: FAIL")
    fi

    # TEST 5: Restart validators and verify sync
    log "\n$(print_header "TEST 5: Restart validators and verify sync")"

    # Get current height from val1
    local height_before=$(get_block_height "$VAL1_RPC")
    log "Current height (validator1): $height_before"

    # Restart all validators
    log "\n$(print_info "Restarting all validators...")"
    start_validator "$VAL2_HOME" "validator2" "26667" "26666" "9092"
    start_validator "$VAL3_HOME" "validator3" "26677" "26676" "9093"
    start_validator "$VAL4_HOME" "validator4" "26687" "26686" "9094"

    # Wait for sync
    print_info "Waiting for validators to sync and produce blocks..."
    sleep 10

    # Monitor block production
    if monitor_block_production "$WAIT_BLOCKS" "$VAL1_RPC" "All validators restarted"; then
        log "$(print_success "PASS: Validators restarted and chain resumed")"
        test_results+=("TEST 5: PASS")
    else
        log "$(print_error "FAIL: Chain did not resume after validator restart")"
        test_results+=("TEST 5: FAIL")
    fi

    # Verify all validators are at similar height
    log "\n$(print_info "Verifying validator synchronization...")"
    local val1_height=$(get_block_height "$VAL1_RPC")
    local val2_height=$(get_block_height "$VAL2_RPC")
    local val3_height=$(get_block_height "$VAL3_RPC")
    local val4_height=$(get_block_height "$VAL4_RPC")

    log "  Validator 1: $val1_height"
    log "  Validator 2: $val2_height"
    log "  Validator 3: $val3_height"
    log "  Validator 4: $val4_height"

    # Check if heights are within 2 blocks of each other
    local max_height=$(printf "%s\n%s\n%s\n%s\n" "$val1_height" "$val2_height" "$val3_height" "$val4_height" | sort -n | tail -1)
    local min_height=$(printf "%s\n%s\n%s\n%s\n" "$val1_height" "$val2_height" "$val3_height" "$val4_height" | sort -n | head -1)
    local height_diff=$((max_height - min_height))

    if [ "$height_diff" -le 2 ]; then
        log "$(print_success "All validators are synchronized (max difference: $height_diff blocks)")"
    else
        log "$(print_error "Validators are not synchronized (difference: $height_diff blocks)")"
    fi

    # Final summary
    log "\n$(print_header "TEST SUITE SUMMARY")"
    log "Completed: $(date)"
    log "\nResults:"
    for result in "${test_results[@]}"; do
        if [[ "$result" == *"PASS"* ]]; then
            log "  $(print_success "$result")"
        else
            log "  $(print_error "$result")"
        fi
    done

    # Count passes and fails
    local pass_count=$(printf "%s\n" "${test_results[@]}" | grep -c "PASS" || echo "0")
    local fail_count=$(printf "%s\n" "${test_results[@]}" | grep -c "FAIL" || echo "0")
    local total_tests=${#test_results[@]}

    log "\n$(print_info "Total: $pass_count/$total_tests tests passed")"

    if [ "$fail_count" -eq 0 ]; then
        log "\n$(print_success "ALL CONSENSUS TESTS PASSED")"
        log "\nResults saved to: $RESULTS_FILE"
        exit 0
    else
        log "\n$(print_error "$fail_count TESTS FAILED")"
        log "\nResults saved to: $RESULTS_FILE"
        exit 1
    fi
}

# Trap errors and cleanup
cleanup() {
    log "\n$(print_info "Test interrupted. Network state preserved for inspection.")"
    log "To view current status, run: display_network_status"
}

trap cleanup EXIT

# Run main
main "$@"
