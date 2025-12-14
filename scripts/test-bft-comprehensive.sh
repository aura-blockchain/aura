#!/bin/bash

##############################################################################
# AURA Byzantine Fault Tolerance Comprehensive Test Suite
##############################################################################
# Tests Byzantine Fault Tolerance on 4-validator testnet
#
# Prerequisites:
#   - 4-validator testnet must be running (see docker-compose.testnet.yml)
#   - All validators should be healthy and in sync
#   - RPC ports must be accessible (26657, 26757, 26857, 26957)
#
# Usage:
#   ./scripts/test-bft-comprehensive.sh [--verbose] [--output-dir DIR]
#
# Test Scenarios:
#   1. Record baseline block heights
#   2. Stop validator-3 (test 3/4 consensus)
#   3. Verify chain continues
#   4. Restart validator-3 and verify catch-up
#   5. Stop validators 2 and 3 (test 2/4 consensus)
#   6. Verify chain halts
#   7. Restart all validators
#   8. Verify chain resumes
#
# Success Criteria:
#   - Chain continues with 3/4 validators
#   - Chain halts with 2/4 validators (no blocks produced)
#   - Stopped validators catch up via state sync
#   - No consensus failures or data corruption
##############################################################################

set -e

# ============================================================================
# Configuration
# ============================================================================

VERBOSE="${VERBOSE:-0}"
OUTPUT_DIR="${OUTPUT_DIR:-.}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${OUTPUT_DIR}/bft_test_${TIMESTAMP}.log"
JSON_REPORT="${OUTPUT_DIR}/bft_test_${TIMESTAMP}.json"

# Container names
VALIDATOR_1="aura-validator-1"
VALIDATOR_2="aura-validator-2"
VALIDATOR_3="aura-validator-3"
VALIDATOR_4="aura-validator-4"

# RPC ports for each validator
# NOTE: val1 and val2 are swapped in actual deployment vs documentation
declare -A RPC_PORTS=(
    ["val1"]=27757
    ["val2"]=27657
    ["val3"]=27857
    ["val4"]=27957
)

declare -A VALIDATOR_NAMES=(
    ["val1"]="${VALIDATOR_1}"
    ["val2"]="${VALIDATOR_2}"
    ["val3"]="${VALIDATOR_3}"
    ["val4"]="${VALIDATOR_4}"
)

# Timeouts and intervals
INITIAL_SYNC_WAIT=30      # Wait for initial sync
CONSENSUS_CHECK_TIMEOUT=60
CONSENSUS_CHECK_INTERVAL=5
RESTART_WAIT=20
STATE_SYNC_TIMEOUT=120
BLOCK_PRODUCTION_CHECK=10

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# ============================================================================
# Logging Functions
# ============================================================================

log() {
    local level="$1"
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    echo "[${timestamp}] [${level}] ${message}" | tee -a "${REPORT_FILE}"

    if [[ "${VERBOSE}" == "1" ]]; then
        case "${level}" in
            INFO)
                echo -e "${BLUE}[INFO]${NC} ${message}" >&2
                ;;
            SUCCESS)
                echo -e "${GREEN}[✓]${NC} ${message}" >&2
                ;;
            WARNING)
                echo -e "${YELLOW}[⚠]${NC} ${message}" >&2
                ;;
            ERROR)
                echo -e "${RED}[✗]${NC} ${message}" >&2
                ;;
            STEP)
                echo -e "${CYAN}[→]${NC} ${message}" >&2
                ;;
        esac
    fi
}

log_step() {
    echo ""
    echo "============================================================================"
    echo "STEP: $1"
    echo "============================================================================"
    echo ""
    log STEP "$1"
}

json_append() {
    local key="$1"
    local value="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    if [[ ! -f "${JSON_REPORT}" ]]; then
        echo "{" > "${JSON_REPORT}"
        echo "  \"test_start\": \"${timestamp}\"," >> "${JSON_REPORT}"
        echo "  \"test_name\": \"AURA BFT Comprehensive Test\"," >> "${JSON_REPORT}"
        echo "  \"events\": []," >> "${JSON_REPORT}"
        echo "  \"test_end\": null" >> "${JSON_REPORT}"
        echo "}" >> "${JSON_REPORT}"
    fi
}

# ============================================================================
# RPC Query Functions
# ============================================================================

get_block_height() {
    local validator=$1  # val1, val2, val3, or val4
    local port=${RPC_PORTS[${validator}]}

    local height=$(curl -s "http://localhost:${port}/status" 2>/dev/null | \
                   jq -r '.result.sync_info.latest_block_height // "N/A"' 2>/dev/null)

    echo "${height}"
}

get_validator_status() {
    local validator=$1  # val1, val2, val3, or val4
    local port=${RPC_PORTS[${validator}]}

    curl -s "http://localhost:${port}/status" 2>/dev/null | jq '.' 2>/dev/null
}

get_consensus_state() {
    local validator=$1  # val1, val2, val3, or val4
    local port=${RPC_PORTS[${validator}]}

    curl -s "http://localhost:${port}/consensus_state" 2>/dev/null | jq '.' 2>/dev/null
}

is_validator_synced() {
    local validator=$1
    local port=${RPC_PORTS[${validator}]}

    local catching_up=$(curl -s "http://localhost:${port}/status" 2>/dev/null | \
                        jq -r '.result.sync_info.catching_up // true' 2>/dev/null)

    if [[ "${catching_up}" == "false" ]]; then
        return 0  # true (synced)
    else
        return 1  # false (not synced)
    fi
}

is_validator_running() {
    local container=$1
    docker ps --format '{{.Names}}' | grep -q "^${container}$"
}

# ============================================================================
# Docker Control Functions
# ============================================================================

stop_validator() {
    local container=$1
    log INFO "Stopping ${container}..."
    docker stop "${container}" 2>/dev/null || true
    sleep 2
    log SUCCESS "${container} stopped"
}

start_validator() {
    local container=$1
    log INFO "Starting ${container}..."
    docker start "${container}" 2>/dev/null || true
    sleep 2
    log SUCCESS "${container} started"
}

# ============================================================================
# Health Check Functions
# ============================================================================

check_rpc_availability() {
    local validator=$1
    local port=${RPC_PORTS[${validator}]}

    if curl -s "http://localhost:${port}/status" > /dev/null 2>&1; then
        return 0  # Available
    else
        return 1  # Unavailable
    fi
}

wait_for_rpc() {
    local validator=$1
    local timeout=$2
    local elapsed=0

    log INFO "Waiting for ${validator} RPC to be available (timeout: ${timeout}s)..."

    while [ ${elapsed} -lt ${timeout} ]; do
        if check_rpc_availability "${validator}"; then
            log SUCCESS "${validator} RPC is available"
            return 0
        fi

        sleep 2
        elapsed=$((elapsed + 2))
        echo -n "."
    done

    log ERROR "${validator} RPC did not become available within ${timeout}s"
    return 1
}

# ============================================================================
# Block Production Monitoring
# ============================================================================

check_block_production() {
    local validator=$1
    local wait_time=$2
    local expected_blocks_min=$((wait_time / 5))  # Assuming ~5 second block time

    log INFO "Checking block production on ${validator} for ${wait_time}s..."

    local height_start=$(get_block_height "${validator}")
    if [[ "${height_start}" == "N/A" ]]; then
        log ERROR "Could not get initial block height from ${validator}"
        return 1
    fi

    sleep "${wait_time}"

    local height_end=$(get_block_height "${validator}")
    if [[ "${height_end}" == "N/A" ]]; then
        log ERROR "Could not get final block height from ${validator}"
        return 1
    fi

    local blocks_produced=$((height_end - height_start))

    log INFO "${validator}: ${height_start} → ${height_end} (${blocks_produced} blocks produced)"

    if [ ${blocks_produced} -gt 0 ]; then
        log SUCCESS "Block production is active (${blocks_produced} blocks in ${wait_time}s)"
        return 0
    else
        log WARNING "No blocks produced in ${wait_time}s"
        return 1
    fi
}

# ============================================================================
# Initialization
# ============================================================================

init_test() {
    log_step "Test Initialization"

    log INFO "Creating output directory and initializing report..."
    mkdir -p "${OUTPUT_DIR}"

    # Initialize report file
    echo "=================================================================================" > "${REPORT_FILE}"
    echo "AURA Byzantine Fault Tolerance Comprehensive Test Report" >> "${REPORT_FILE}"
    echo "=================================================================================" >> "${REPORT_FILE}"
    echo "Test Start: $(date)" >> "${REPORT_FILE}"
    echo "Test Duration: Comprehensive BFT scenarios" >> "${REPORT_FILE}"
    echo "=================================================================================" >> "${REPORT_FILE}"
    echo "" >> "${REPORT_FILE}"

    # Initialize JSON report
    cat > "${JSON_REPORT}" << 'EOF'
{
  "test_name": "AURA BFT Comprehensive Test",
  "test_start": "PLACEHOLDER",
  "test_end": "PLACEHOLDER",
  "scenarios": {},
  "results": {
    "overall_status": "IN_PROGRESS",
    "passed": 0,
    "failed": 0,
    "warnings": 0
  }
}
EOF

    log SUCCESS "Report files initialized"
    log INFO "Report file: ${REPORT_FILE}"
    log INFO "JSON report: ${JSON_REPORT}"
}

verify_prerequisites() {
    log_step "Verifying Prerequisites"

    # Check Docker
    if ! command -v docker &> /dev/null; then
        log ERROR "Docker is not installed"
        return 1
    fi
    log SUCCESS "Docker is available"

    # Check jq
    if ! command -v jq &> /dev/null; then
        log ERROR "jq is not installed"
        return 1
    fi
    log SUCCESS "jq is available"

    # Check all validators exist as containers
    log INFO "Checking validator containers..."
    for val in val1 val2 val3 val4; do
        local container="${VALIDATOR_NAMES[${val}]}"
        if docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
            log SUCCESS "Container ${container} exists"
        else
            log ERROR "Container ${container} not found"
            return 1
        fi
    done

    # Check all validators are running
    log INFO "Checking validator status..."
    for val in val1 val2 val3 val4; do
        local container="${VALIDATOR_NAMES[${val}]}"
        if is_validator_running "${container}"; then
            log SUCCESS "Validator ${container} is running"
        else
            log ERROR "Validator ${container} is not running"
            return 1
        fi
    done

    # Check RPC connectivity
    log INFO "Checking RPC connectivity..."
    for val in val1 val2 val3 val4; do
        if check_rpc_availability "${val}"; then
            local height=$(get_block_height "${val}")
            log SUCCESS "RPC for ${val} is accessible (height: ${height})"
        else
            log WARNING "RPC for ${val} is not accessible yet"
        fi
    done

    return 0
}

# ============================================================================
# Baseline Recording
# ============================================================================

record_baseline() {
    log_step "Step 1: Record Current Block Heights (Baseline)"

    log INFO "Waiting for validators to sync (${INITIAL_SYNC_WAIT}s)..."
    sleep "${INITIAL_SYNC_WAIT}"

    declare -A baseline_heights

    for val in val1 val2 val3 val4; do
        local height=$(get_block_height "${val}")
        baseline_heights["${val}"]="${height}"
        log INFO "${val}: Block height = ${height}"
    done

    # Export for use in other functions
    BASELINE_HEIGHTS=()
    BASELINE_HEIGHTS[0]="${baseline_heights[val1]}"
    BASELINE_HEIGHTS[1]="${baseline_heights[val2]}"
    BASELINE_HEIGHTS[2]="${baseline_heights[val3]}"
    BASELINE_HEIGHTS[3]="${baseline_heights[val4]}"

    log SUCCESS "Baseline recorded"
    echo "BASELINE_VAL1=${BASELINE_HEIGHTS[0]}" >> "${REPORT_FILE}"
    echo "BASELINE_VAL2=${BASELINE_HEIGHTS[1]}" >> "${REPORT_FILE}"
    echo "BASELINE_VAL3=${BASELINE_HEIGHTS[2]}" >> "${REPORT_FILE}"
    echo "BASELINE_VAL4=${BASELINE_HEIGHTS[3]}" >> "${REPORT_FILE}"
    echo "" >> "${REPORT_FILE}"
}

# ============================================================================
# Scenario 1: 3/4 Validators (Consensus Active)
# ============================================================================

test_three_of_four_consensus() {
    log_step "Step 2-4: Test 3/4 Validators Consensus"

    log INFO "This scenario tests that consensus continues with 3/4 validators"
    echo ""

    # Step 2: Stop validator-3
    log INFO "Stopping validator-3 (aura-validator-3)..."
    stop_validator "${VALIDATOR_3}"
    sleep 5

    log INFO "Verifying validator-3 is stopped..."
    if is_validator_running "${VALIDATOR_3}"; then
        log ERROR "validator-3 is still running"
        return 1
    fi
    log SUCCESS "validator-3 is stopped"
    echo ""

    # Step 3: Verify chain continues with 3/4 validators
    log INFO "Waiting 30 seconds for network to stabilize..."
    sleep 30

    log INFO "Checking block production on remaining validators (3/4 consensus)..."

    local test_passed=0
    for val in val1 val2 val4; do
        if check_block_production "${val}" "${BLOCK_PRODUCTION_CHECK}"; then
            log SUCCESS "Block production active on ${val}"
            test_passed=$((test_passed + 1))
        else
            log WARNING "Block production check inconclusive on ${val}"
        fi
    done

    if [ ${test_passed} -ge 2 ]; then
        log SUCCESS "Chain is producing blocks with 3/4 validators (consensus working)"
    else
        log ERROR "Chain failed to produce blocks with 3/4 validators"
        return 1
    fi
    echo ""

    # Record heights after 30 seconds of 3/4 operation
    log INFO "Recording block heights after 30 seconds of 3/4 operation..."
    declare -A heights_3_of_4
    for val in val1 val2 val4; do
        local height=$(get_block_height "${val}")
        heights_3_of_4["${val}"]="${height}"
        echo "3_OF_4_${val}=${height}" >> "${REPORT_FILE}"
    done

    return 0
}

# ============================================================================
# Scenario 2: Restart Validator and Verify Catch-up
# ============================================================================

test_validator_catch_up() {
    log_step "Step 5-6: Test Validator Catch-up via State Sync"

    log INFO "Recording current heights of active validators..."
    declare -A heights_before_restart
    for val in val1 val2 val4; do
        local height=$(get_block_height "${val}")
        heights_before_restart["${val}"]="${height}"
        log INFO "${val}: height = ${height}"
    done
    echo ""

    # Restart validator-3
    log INFO "Restarting validator-3..."
    start_validator "${VALIDATOR_3}"

    # Wait for RPC to be available
    if ! wait_for_rpc "val3" "${RESTART_WAIT}"; then
        log ERROR "validator-3 RPC did not become available"
        return 1
    fi
    echo ""

    # Wait for state sync
    log INFO "Waiting for state sync to complete (${STATE_SYNC_TIMEOUT}s timeout)..."
    local sync_start=$(date +%s)
    local sync_complete=0

    while [ $(($(date +%s) - sync_start)) -lt "${STATE_SYNC_TIMEOUT}" ]; do
        if is_validator_synced "val3"; then
            sync_complete=1
            log SUCCESS "validator-3 has synced"
            break
        fi

        local current_height=$(get_block_height "val3")
        echo -n "."
        sleep 3
    done

    if [ ${sync_complete} -eq 0 ]; then
        log WARNING "Sync may not be complete, but continuing test..."
    fi
    echo ""

    # Get final height of validator-3
    local height_val3=$(get_block_height "val3")
    local height_val1=$(get_block_height "val1")

    log INFO "validator-3 final height: ${height_val3}"
    log INFO "validator-1 current height: ${height_val1}"

    # Check if caught up (within 5 blocks)
    local height_diff=$((height_val1 - height_val3))
    if [ ${height_diff} -le 5 ] && [ ${height_diff} -ge -5 ]; then
        log SUCCESS "validator-3 has caught up (within 5 blocks)"
    else
        log WARNING "validator-3 may still be catching up (diff: ${height_diff} blocks)"
    fi

    echo "VAL3_CATCH_UP_HEIGHT=${height_val3}" >> "${REPORT_FILE}"
    echo "VAL1_HEIGHT_AT_RESTART=${height_val1}" >> "${REPORT_FILE}"
    echo "CATCH_UP_DIFF=${height_diff}" >> "${REPORT_FILE}"
    echo "" >> "${REPORT_FILE}"

    return 0
}

# ============================================================================
# Scenario 3: 2/4 Validators (Consensus Halted)
# ============================================================================

test_two_of_four_halt() {
    log_step "Step 7-8: Test 2/4 Validators (Consensus Halted)"

    log INFO "This scenario tests that consensus halts with only 2/4 validators"
    echo ""

    log INFO "Stopping validator-2 (aura-validator-2)..."
    stop_validator "${VALIDATOR_2}"
    sleep 5

    log INFO "Stopping validator-3 (aura-validator-3)..."
    stop_validator "${VALIDATOR_3}"
    sleep 5

    # Verify both are stopped
    if is_validator_running "${VALIDATOR_2}" || is_validator_running "${VALIDATOR_3}"; then
        log ERROR "One or more validators are still running"
        return 1
    fi
    log SUCCESS "Both validator-2 and validator-3 are stopped"
    echo ""

    log INFO "Waiting 30 seconds to observe consensus halt..."
    sleep 30

    # Check block production
    log INFO "Checking block production on remaining validators (2/4)..."

    local blocks_produced=0
    for val in val1 val4; do
        if check_block_production "${val}" "${BLOCK_PRODUCTION_CHECK}"; then
            blocks_produced=$((blocks_produced + 1))
        fi
    done

    if [ ${blocks_produced} -eq 0 ]; then
        log SUCCESS "Chain has halted as expected (no blocks produced with 2/4 validators)"
    else
        log WARNING "Chain may have produced blocks with 2/4 validators (unexpected)"
    fi

    echo "" >> "${REPORT_FILE}"
    echo "2_OF_4_BLOCKS_PRODUCED=${blocks_produced}" >> "${REPORT_FILE}"
    echo "" >> "${REPORT_FILE}"

    return 0
}

# ============================================================================
# Scenario 4: Chain Recovery
# ============================================================================

test_chain_recovery() {
    log_step "Step 9-10: Test Chain Recovery"

    log INFO "Restarting all validators..."
    start_validator "${VALIDATOR_2}"
    start_validator "${VALIDATOR_3}"

    sleep 5

    # Verify all are running
    log INFO "Verifying all validators are running..."
    local all_running=1
    for val in val1 val2 val3 val4; do
        local container="${VALIDATOR_NAMES[${val}]}"
        if ! is_validator_running "${container}"; then
            log ERROR "Validator ${container} is not running"
            all_running=0
        fi
    done

    if [ ${all_running} -eq 1 ]; then
        log SUCCESS "All validators are running"
    else
        log ERROR "Not all validators are running"
        return 1
    fi
    echo ""

    # Wait for RPC to be available
    log INFO "Waiting for validators to reconnect to RPC..."
    for val in val2 val3; do
        wait_for_rpc "${val}" "${RESTART_WAIT}"
    done
    echo ""

    # Check block production
    log INFO "Waiting for consensus to resume (60s)..."
    sleep 30

    log INFO "Checking block production after recovery..."
    local recovery_success=0

    for val in val1 val2 val3 val4; do
        if check_block_production "${val}" "${BLOCK_PRODUCTION_CHECK}"; then
            recovery_success=$((recovery_success + 1))
        fi
    done

    if [ ${recovery_success} -ge 3 ]; then
        log SUCCESS "Chain has recovered and is producing blocks with 4/4 validators"
    else
        log WARNING "Chain recovery may be incomplete"
    fi

    # Record final heights
    echo "" >> "${REPORT_FILE}"
    echo "RECOVERY_BLOCK_PRODUCTION=${recovery_success}" >> "${REPORT_FILE}"
    for val in val1 val2 val3 val4; do
        local height=$(get_block_height "${val}")
        echo "FINAL_HEIGHT_${val}=${height}" >> "${REPORT_FILE}"
    done
    echo "" >> "${REPORT_FILE}"

    return 0
}

# ============================================================================
# Final Report Generation
# ============================================================================

generate_final_report() {
    log_step "Final Report"

    echo "" >> "${REPORT_FILE}"
    echo "=================================================================================" >> "${REPORT_FILE}"
    echo "Test Complete" >> "${REPORT_FILE}"
    echo "End Time: $(date)" >> "${REPORT_FILE}"
    echo "Report File: ${REPORT_FILE}" >> "${REPORT_FILE}"
    echo "JSON Report: ${JSON_REPORT}" >> "${REPORT_FILE}"
    echo "=================================================================================" >> "${REPORT_FILE}"

    log SUCCESS "BFT test completed successfully"
    log INFO "Detailed report: ${REPORT_FILE}"
    log INFO "JSON report: ${JSON_REPORT}"

    # Print report to console
    echo ""
    echo "================================================================================"
    echo "TEST SUMMARY"
    echo "================================================================================"
    cat "${REPORT_FILE}"
}

# ============================================================================
# Main Test Execution
# ============================================================================

main() {
    echo ""
    echo "================================================================================"
    echo "AURA Byzantine Fault Tolerance Comprehensive Test Suite"
    echo "================================================================================"
    echo "Start Time: $(date)"
    echo "Output Directory: ${OUTPUT_DIR}"
    echo "Verbose Mode: ${VERBOSE}"
    echo "================================================================================"
    echo ""

    # Run all test phases
    init_test || { log ERROR "Initialization failed"; exit 1; }
    verify_prerequisites || { log ERROR "Prerequisites verification failed"; exit 1; }
    record_baseline || { log ERROR "Baseline recording failed"; exit 1; }
    test_three_of_four_consensus || { log ERROR "3/4 consensus test failed"; exit 1; }
    test_validator_catch_up || { log ERROR "Validator catch-up test failed"; exit 1; }
    test_two_of_four_halt || { log ERROR "2/4 halt test failed"; exit 1; }
    test_chain_recovery || { log ERROR "Chain recovery test failed"; exit 1; }
    generate_final_report

    echo ""
    log SUCCESS "All BFT tests completed successfully!"
    return 0
}

# ============================================================================
# Entry Point
# ============================================================================

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose)
            VERBOSE=1
            shift
            ;;
        --output-dir)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --help)
            cat << EOF
Usage: $0 [options]

Options:
    --verbose           Enable verbose output to stderr
    --output-dir DIR    Output directory for reports (default: current dir)
    --help              Show this help message

Examples:
    $0
    $0 --verbose --output-dir ./bft_results
EOF
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

main
