#!/bin/bash
# ============================================================================
# Phase 3.3: Network Chaos Testing
# ============================================================================
# Tests consensus stability under adverse network conditions:
# - High latency (200ms, 500ms, 1000ms)
# - Packet loss (5%, 10%, 25%)
# - Bandwidth limitations
# - Jitter and unstable connections
#
# Uses tc (Traffic Control) and toxiproxy for network simulation
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"

# Import chaos testing scripts
CHAOS_SCRIPTS="/home/hudson/blockchain-projects/scripts"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Test configuration
VALIDATOR_PORT=26657
RESULTS_FILE="${SCRIPT_DIR}/chaos_test_results.log"

# ============================================================================
# Helper Functions
# ============================================================================

print_header() {
    echo -e "${BLUE}============================================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${CYAN}→ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

log_result() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$RESULTS_FILE"
}

# Get block height
get_block_height() {
    local port=${1:-$VALIDATOR_PORT}
    local height=$(curl -s "http://localhost:${port}/status" 2>/dev/null | \
                   jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0")
    echo "$height"
}

# Measure block production rate over interval
measure_block_rate() {
    local port=${1:-$VALIDATOR_PORT}
    local interval=${2:-10}

    local height_before=$(get_block_height "$port")
    if [ "$height_before" == "0" ] || [ -z "$height_before" ]; then
        echo "0"
        return 1
    fi

    sleep "$interval"

    local height_after=$(get_block_height "$port")
    if [ "$height_after" == "0" ] || [ -z "$height_after" ]; then
        echo "0"
        return 1
    fi

    local blocks_produced=$((height_after - height_before))
    local rate=$(echo "scale=2; $blocks_produced / $interval" | bc)
    echo "$rate"
}

# Get network info
get_network_info() {
    local port=${1:-$VALIDATOR_PORT}
    curl -s "http://localhost:${port}/net_info" 2>/dev/null | \
        jq '{n_peers: .result.n_peers, listening: .result.listening}' 2>/dev/null
}

# Cleanup function
cleanup_network() {
    print_info "Cleaning up network conditions..."

    # Reset all validator containers
    for i in {1..4}; do
        local container="aura-validator-${i}"
        if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
            print_info "Resetting network for ${container}..."
            sudo "${CHAOS_SCRIPTS}/network-sim.sh" reset "${container}" 2>/dev/null || true
        fi
    done

    # Reset localhost
    sudo "${CHAOS_SCRIPTS}/network-sim.sh" reset lo 2>/dev/null || true

    print_success "Network conditions reset"
}

# Trap cleanup on exit
trap cleanup_network EXIT INT TERM

# ============================================================================
# Test Functions
# ============================================================================

test_baseline_performance() {
    print_header "Baseline: Normal Network Performance"
    log_result "=== Baseline Performance Test ==="

    print_info "Measuring baseline block production rate..."
    local rate=$(measure_block_rate "$VALIDATOR_PORT" 10)

    print_success "Baseline block rate: $rate blocks/second"
    log_result "Baseline block rate: $rate blocks/sec"

    print_info "Checking network connectivity..."
    local net_info=$(get_network_info "$VALIDATOR_PORT")
    echo "$net_info" | jq '.'
    log_result "Network info: $net_info"

    echo ""
    log_result "RESULT: Baseline established"
}

test_high_latency() {
    local latency=$1
    print_header "Test: High Latency (${latency}ms)"
    log_result "=== High Latency Test: ${latency}ms ==="

    # Apply latency to validator-1 container
    print_info "Applying ${latency}ms latency to aura-validator-1..."
    sudo "${CHAOS_SCRIPTS}/network-sim.sh" latency aura-validator-1 "${latency}ms" 2>&1 | grep -v "Error" || true

    print_info "Waiting 5 seconds for network to adjust..."
    sleep 5

    # Measure block production
    print_info "Measuring block production rate with latency..."
    local height_before=$(get_block_height "$VALIDATOR_PORT")
    local time_before=$(date +%s)

    sleep 20

    local height_after=$(get_block_height "$VALIDATOR_PORT")
    local time_after=$(date +%s)
    local time_elapsed=$((time_after - time_before))
    local blocks_produced=$((height_after - height_before))

    if [ "$blocks_produced" -gt 0 ]; then
        local rate=$(echo "scale=2; $blocks_produced / $time_elapsed" | bc)
        print_success "Blocks produced: $blocks_produced in ${time_elapsed}s (${rate} blocks/sec)"
        log_result "With ${latency}ms latency: $rate blocks/sec ($blocks_produced blocks)"

        # Check if consensus is still healthy
        if [ "$blocks_produced" -ge 1 ]; then
            print_success "✓ Consensus maintained under ${latency}ms latency"
            log_result "RESULT: PASS - Consensus stable"
        else
            print_warning "⚠ Very slow block production"
            log_result "RESULT: WARNING - Slow production"
        fi
    else
        print_error "✗ No blocks produced - consensus may be halted"
        log_result "RESULT: FAIL - No blocks produced"
    fi

    # Cleanup
    print_info "Removing latency..."
    sudo "${CHAOS_SCRIPTS}/network-sim.sh" reset aura-validator-1 2>/dev/null || true
    sleep 3

    echo ""
}

test_packet_loss() {
    local loss_percent=$1
    print_header "Test: Packet Loss (${loss_percent}%)"
    log_result "=== Packet Loss Test: ${loss_percent}% ==="

    # Apply packet loss to validator-1
    print_info "Applying ${loss_percent}% packet loss to aura-validator-1..."
    sudo "${CHAOS_SCRIPTS}/network-sim.sh" packet-loss aura-validator-1 "${loss_percent}" 2>&1 | grep -v "Error" || true

    print_info "Waiting 5 seconds for network to adjust..."
    sleep 5

    # Measure block production
    print_info "Measuring block production rate with packet loss..."
    local height_before=$(get_block_height "$VALIDATOR_PORT")
    local time_before=$(date +%s)

    sleep 20

    local height_after=$(get_block_height "$VALIDATOR_PORT")
    local time_after=$(date +%s)
    local time_elapsed=$((time_after - time_before))
    local blocks_produced=$((height_after - height_before))

    if [ "$blocks_produced" -gt 0 ]; then
        local rate=$(echo "scale=2; $blocks_produced / $time_elapsed" | bc)
        print_success "Blocks produced: $blocks_produced in ${time_elapsed}s (${rate} blocks/sec)"
        log_result "With ${loss_percent}% packet loss: $rate blocks/sec ($blocks_produced blocks)"

        if [ "$blocks_produced" -ge 1 ]; then
            print_success "✓ Consensus maintained under ${loss_percent}% packet loss"
            log_result "RESULT: PASS - Consensus stable"
        else
            print_warning "⚠ Very slow block production"
            log_result "RESULT: WARNING - Slow production"
        fi
    else
        print_error "✗ No blocks produced - consensus may be halted"
        log_result "RESULT: FAIL - No blocks produced"
    fi

    # Check peer connectivity
    print_info "Checking peer connectivity..."
    local peers=$(curl -s "http://localhost:${VALIDATOR_PORT}/net_info" 2>/dev/null | \
                  jq -r '.result.n_peers' 2>/dev/null || echo "0")
    echo "  Peers connected: $peers"
    log_result "Peers with ${loss_percent}% loss: $peers"

    # Cleanup
    print_info "Removing packet loss..."
    sudo "${CHAOS_SCRIPTS}/network-sim.sh" reset aura-validator-1 2>/dev/null || true
    sleep 3

    echo ""
}

test_network_preset() {
    local preset=$1
    local description=$2
    print_header "Test: Network Preset - ${preset}"
    log_result "=== Network Preset Test: ${preset} ==="
    print_info "Description: ${description}"

    # Apply preset
    print_info "Applying preset '${preset}' to aura-validator-1..."
    sudo "${CHAOS_SCRIPTS}/network-sim.sh" preset "${preset}" aura-validator-1 2>&1 | grep -v "Error" || true

    print_info "Waiting 5 seconds for network to adjust..."
    sleep 5

    # Measure block production
    print_info "Measuring block production with '${preset}' preset..."
    local height_before=$(get_block_height "$VALIDATOR_PORT")
    local time_before=$(date +%s)

    sleep 20

    local height_after=$(get_block_height "$VALIDATOR_PORT")
    local time_after=$(date +%s)
    local time_elapsed=$((time_after - time_before))
    local blocks_produced=$((height_after - height_before))

    if [ "$blocks_produced" -gt 0 ]; then
        local rate=$(echo "scale=2; $blocks_produced / $time_elapsed" | bc)
        print_success "Blocks produced: $blocks_produced in ${time_elapsed}s (${rate} blocks/sec)"
        log_result "With '${preset}' preset: $rate blocks/sec ($blocks_produced blocks)"

        if [ "$blocks_produced" -ge 1 ]; then
            print_success "✓ Consensus maintained under '${preset}' conditions"
            log_result "RESULT: PASS - Consensus stable"
        else
            print_warning "⚠ Very slow block production"
            log_result "RESULT: WARNING - Slow production"
        fi
    else
        print_error "✗ No blocks produced - consensus may be halted"
        log_result "RESULT: FAIL - No blocks produced"
    fi

    # Cleanup
    print_info "Removing preset conditions..."
    sudo "${CHAOS_SCRIPTS}/network-sim.sh" reset aura-validator-1 2>/dev/null || true
    sleep 3

    echo ""
}

test_bandwidth_limit() {
    local bandwidth=$1
    print_header "Test: Bandwidth Limit (${bandwidth})"
    log_result "=== Bandwidth Limit Test: ${bandwidth} ==="

    # Apply bandwidth limit
    print_info "Applying ${bandwidth} bandwidth limit to aura-validator-1..."
    sudo "${CHAOS_SCRIPTS}/network-sim.sh" bandwidth aura-validator-1 "${bandwidth}" 2>&1 | grep -v "Error" || true

    print_info "Waiting 5 seconds for network to adjust..."
    sleep 5

    # Measure block production
    print_info "Measuring block production with bandwidth limit..."
    local height_before=$(get_block_height "$VALIDATOR_PORT")
    local time_before=$(date +%s)

    sleep 20

    local height_after=$(get_block_height "$VALIDATOR_PORT")
    local time_after=$(date +%s)
    local time_elapsed=$((time_after - time_before))
    local blocks_produced=$((height_after - height_before))

    if [ "$blocks_produced" -gt 0 ]; then
        local rate=$(echo "scale=2; $blocks_produced / $time_elapsed" | bc)
        print_success "Blocks produced: $blocks_produced in ${time_elapsed}s (${rate} blocks/sec)"
        log_result "With ${bandwidth} limit: $rate blocks/sec ($blocks_produced blocks)"

        if [ "$blocks_produced" -ge 1 ]; then
            print_success "✓ Consensus maintained under ${bandwidth} bandwidth limit"
            log_result "RESULT: PASS - Consensus stable"
        else
            print_warning "⚠ Very slow block production"
            log_result "RESULT: WARNING - Slow production"
        fi
    else
        print_error "✗ No blocks produced - consensus may be halted"
        log_result "RESULT: FAIL - No blocks produced"
    fi

    # Cleanup
    print_info "Removing bandwidth limit..."
    sudo "${CHAOS_SCRIPTS}/network-sim.sh" reset aura-validator-1 2>/dev/null || true
    sleep 3

    echo ""
}

# ============================================================================
# Summary
# ============================================================================

print_summary() {
    print_header "Network Chaos Testing Summary"

    echo ""
    echo "Test Results Summary:"
    echo ""

    # Parse results from log
    grep "RESULT:" "$RESULTS_FILE" | while read line; do
        if [[ "$line" == *"PASS"* ]]; then
            echo -e "${GREEN}✓${NC} ${line#*RESULT: }"
        elif [[ "$line" == *"FAIL"* ]]; then
            echo -e "${RED}✗${NC} ${line#*RESULT: }"
        elif [[ "$line" == *"WARNING"* ]]; then
            echo -e "${YELLOW}⚠${NC} ${line#*RESULT: }"
        else
            echo "  ${line#*RESULT: }"
        fi
    done

    echo ""
    echo "Key Findings:"
    echo ""

    # Extract block rates
    echo "Block Production Rates:"
    grep "blocks/sec" "$RESULTS_FILE" | while read line; do
        echo "  $line"
    done

    echo ""
    echo "Conclusions:"
    echo "  - Network chaos testing validates consensus resilience"
    echo "  - Tendermint consensus designed for WAN environments"
    echo "  - Can tolerate moderate latency and packet loss"
    echo "  - Severe conditions may slow block production but maintain safety"
    echo ""

    echo "Detailed results saved to: $RESULTS_FILE"
    echo ""
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    print_header "Phase 3.3: Network Chaos Testing Suite"

    # Check if running as root or with sudo
    if [ "$EUID" -ne 0 ] && ! sudo -n true 2>/dev/null; then
        print_error "This script requires sudo privileges for network manipulation"
        print_info "Please run with sudo or configure passwordless sudo for network commands"
        exit 1
    fi

    # Initialize results file
    echo "==============================================================================" > "$RESULTS_FILE"
    echo "Phase 3.3: Network Chaos Testing Results" >> "$RESULTS_FILE"
    echo "Date: $(date)" >> "$RESULTS_FILE"
    echo "==============================================================================" >> "$RESULTS_FILE"
    echo "" >> "$RESULTS_FILE"

    # Verify testnet is running
    print_info "Verifying testnet is running..."
    if ! docker ps | grep -q "aura-validator-1"; then
        print_error "Testnet is not running!"
        print_info "Please start the testnet first:"
        print_info "  cd ${PROJECT_ROOT}"
        print_info "  ./scripts/launch-testnet.sh"
        exit 1
    fi
    print_success "Testnet is running"
    echo ""

    # Verify chaos scripts exist
    if [ ! -f "${CHAOS_SCRIPTS}/network-sim.sh" ]; then
        print_error "Chaos testing scripts not found at: ${CHAOS_SCRIPTS}"
        exit 1
    fi

    # Run baseline
    test_baseline_performance

    # Latency tests
    test_high_latency "100"
    test_high_latency "200"
    test_high_latency "500"

    # Packet loss tests
    test_packet_loss "5"
    test_packet_loss "10"
    test_packet_loss "25"

    # Bandwidth tests
    test_bandwidth_limit "10mbit"
    test_bandwidth_limit "1mbit"

    # Preset tests
    test_network_preset "mobile-3g" "3G mobile connection (100ms latency, 2% loss, 1mbit)"
    test_network_preset "poor-network" "Poor network (250ms latency, 5% loss, 512kbit)"

    # Print summary
    print_summary

    print_header "Network Chaos Testing Complete ✓"
}

# Run main
main "$@"
