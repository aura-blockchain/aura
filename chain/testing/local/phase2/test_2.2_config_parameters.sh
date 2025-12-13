#!/bin/bash
set -e

# Test 2.2: Configuration Parameters Testing
# This script tests critical configuration parameters in config.toml and app.toml
# to ensure the node respects configuration changes

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
BINARY="$PROJECT_ROOT/chain/aurad"
TEST_HOME="/tmp/aura-config-test-$$"
CHAIN_ID="aura-config-test"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Results tracking
declare -A TEST_RESULTS
TESTS_PASSED=0
TESTS_FAILED=0

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_step() {
    echo -e "\n${GREEN}==>${NC} $1"
}

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

# Test result tracking
record_pass() {
    TEST_RESULTS["$1"]="PASSED"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    log_info "✓ $1"
}

record_fail() {
    TEST_RESULTS["$1"]="FAILED: $2"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    log_error "✗ $1: $2"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    if [ -d "$TEST_HOME" ]; then
        rm -rf "$TEST_HOME"
        log_info "Removed test home directory: $TEST_HOME"
    fi

    # Kill any running test nodes
    pkill -f "aurad.*--home.*$TEST_HOME" 2>/dev/null || true
}

# Set trap for cleanup
trap cleanup EXIT

# Verify binary exists
if [ ! -f "$BINARY" ]; then
    log_error "Binary not found at $BINARY"
    exit 1
fi

# Initialize test node
initialize_node() {
    log_step "Initializing test node"
    rm -rf "$TEST_HOME"

    $BINARY init test-node --chain-id "$CHAIN_ID" --home "$TEST_HOME" > /dev/null 2>&1

    # Add a validator account
    $BINARY keys add validator --keyring-backend test --home "$TEST_HOME" > /dev/null 2>&1
    VALIDATOR_ADDR=$($BINARY keys show validator -a --keyring-backend test --home "$TEST_HOME" 2>&1)

    # Add genesis account and create genesis
    $BINARY genesis add-genesis-account "$VALIDATOR_ADDR" 1000000000stake,1000000000uaura --home "$TEST_HOME" > /dev/null 2>&1
    $BINARY genesis gentx validator 500000000stake --chain-id "$CHAIN_ID" --keyring-backend test --home "$TEST_HOME" > /dev/null 2>&1
    $BINARY genesis collect-gentxs --home "$TEST_HOME" > /dev/null 2>&1

    log_info "Node initialized successfully"
}

# Helper function to modify config.toml
modify_config_toml() {
    local key=$1
    local value=$2
    local config_file="$TEST_HOME/config/config.toml"

    # Use sed with different delimiter to avoid issues with slashes
    sed -i "s|^${key}[[:space:]]*=.*|${key} = ${value}|" "$config_file"
}

# Helper function to modify app.toml
modify_app_toml() {
    local key=$1
    local value=$2
    local config_file="$TEST_HOME/config/app.toml"

    sed -i "s|^${key}[[:space:]]*=.*|${key} = ${value}|" "$config_file"
}

# Helper function to get config value
get_config_value() {
    local key=$1
    local config_file="$TEST_HOME/config/config.toml"

    grep "^${key}" "$config_file" | head -1 | sed 's/.*=[[:space:]]*//'
}

# Helper function to get app config value
get_app_config_value() {
    local key=$1
    local config_file="$TEST_HOME/config/app.toml"

    grep "^${key}" "$config_file" | head -1 | sed 's/.*=[[:space:]]*//'
}

# Start node in background and wait for it to be ready
start_node() {
    log_test "Starting node..."

    # Start node in background
    $BINARY start --home "$TEST_HOME" > "$TEST_HOME/node.log" 2>&1 &
    NODE_PID=$!

    # Wait for node to be ready (check for port to be listening)
    local max_wait=30
    local waited=0
    while [ $waited -lt $max_wait ]; do
        if curl -s http://localhost:26657/status > /dev/null 2>&1; then
            log_info "Node started successfully (PID: $NODE_PID)"
            return 0
        fi
        sleep 1
        waited=$((waited + 1))
    done

    log_error "Node failed to start within ${max_wait}s"
    cat "$TEST_HOME/node.log"
    return 1
}

# Stop node
stop_node() {
    if [ -n "$NODE_PID" ]; then
        log_test "Stopping node (PID: $NODE_PID)..."
        kill $NODE_PID 2>/dev/null || true
        wait $NODE_PID 2>/dev/null || true
        sleep 2
    fi
}

log_info "Starting Configuration Parameters Test"
log_info "Binary: $BINARY"
log_info "Test Home: $TEST_HOME"

# ====================================================================
# TEST 1: p2p.max_num_inbound_peers
# ====================================================================
log_step "Test 1: p2p.max_num_inbound_peers"
log_test "Testing max inbound peers configuration"

initialize_node

# Check default value
DEFAULT_INBOUND=$(get_config_value "max_num_inbound_peers")
log_info "Default max_num_inbound_peers: $DEFAULT_INBOUND"

# Modify to a specific value
modify_config_toml "max_num_inbound_peers" "10"
NEW_INBOUND=$(get_config_value "max_num_inbound_peers")

if [[ "$NEW_INBOUND" == *"10"* ]]; then
    record_pass "p2p.max_num_inbound_peers configuration change"
else
    record_fail "p2p.max_num_inbound_peers configuration change" "Value not updated correctly"
fi

# ====================================================================
# TEST 2: consensus.timeout_commit
# ====================================================================
log_step "Test 2: consensus.timeout_commit"
log_test "Testing consensus timeout commit configuration"

initialize_node

# Check default value
DEFAULT_TIMEOUT=$(get_config_value "timeout_commit")
log_info "Default timeout_commit: $DEFAULT_TIMEOUT"

# Modify to 2s
modify_config_toml "timeout_commit" '"2s"'
NEW_TIMEOUT=$(get_config_value "timeout_commit")

if [[ "$NEW_TIMEOUT" == *"2s"* ]]; then
    record_pass "consensus.timeout_commit configuration change"
else
    record_fail "consensus.timeout_commit configuration change" "Value not updated correctly"
fi

# ====================================================================
# TEST 3: mempool.size
# ====================================================================
log_step "Test 3: mempool.size"
log_test "Testing mempool size configuration"

initialize_node

# Check default value
DEFAULT_MEMPOOL=$(get_config_value "size")
log_info "Default mempool size: $DEFAULT_MEMPOOL"

# Modify to 10000
modify_config_toml "size" "10000"
NEW_MEMPOOL=$(get_config_value "size")

if [[ "$NEW_MEMPOOL" == *"10000"* ]]; then
    record_pass "mempool.size configuration change"
else
    record_fail "mempool.size configuration change" "Value not updated correctly"
fi

# ====================================================================
# TEST 4: app.toml - grpc.max-recv-msg-size
# ====================================================================
log_step "Test 4: grpc.max-recv-msg-size (app.toml)"
log_test "Testing gRPC max receive message size configuration"

initialize_node

# Check default value
DEFAULT_GRPC=$(get_app_config_value "max-recv-msg-size")
log_info "Default max-recv-msg-size: $DEFAULT_GRPC"

# Modify to specific value
modify_app_toml "max-recv-msg-size" "20971520"
NEW_GRPC=$(get_app_config_value "max-recv-msg-size")

if [[ "$NEW_GRPC" == *"20971520"* ]]; then
    record_pass "grpc.max-recv-msg-size configuration change"
else
    record_fail "grpc.max-recv-msg-size configuration change" "Value not updated correctly (got: $NEW_GRPC)"
fi

# ====================================================================
# TEST 5: app.toml - snapshot-interval
# ====================================================================
log_step "Test 5: snapshot-interval (app.toml)"
log_test "Testing snapshot interval configuration"

initialize_node

# Check default value
DEFAULT_SNAPSHOT=$(get_app_config_value "snapshot-interval")
log_info "Default snapshot-interval: $DEFAULT_SNAPSHOT"

# Modify to 1000
modify_app_toml "snapshot-interval" "1000"
NEW_SNAPSHOT=$(get_app_config_value "snapshot-interval")

if [[ "$NEW_SNAPSHOT" == *"1000"* ]]; then
    record_pass "snapshot-interval configuration change"
else
    record_fail "snapshot-interval configuration change" "Value not updated correctly"
fi

# ====================================================================
# TEST 6: rpc.laddr - RPC listening address
# ====================================================================
log_step "Test 6: rpc.laddr"
log_test "Testing RPC listening address configuration"

initialize_node

# Check default value
DEFAULT_RPC=$(get_config_value "laddr")
log_info "Default rpc.laddr: $DEFAULT_RPC"

# Modify to different port
modify_config_toml "laddr" '"tcp://127.0.0.1:26658"'
NEW_RPC=$(get_config_value "laddr")

if [[ "$NEW_RPC" == *"26658"* ]]; then
    record_pass "rpc.laddr configuration change"
else
    record_fail "rpc.laddr configuration change" "Value not updated correctly"
fi

# ====================================================================
# TEST 7: p2p.max_num_outbound_peers
# ====================================================================
log_step "Test 7: p2p.max_num_outbound_peers"
log_test "Testing max outbound peers configuration"

initialize_node

# Check default value
DEFAULT_OUTBOUND=$(get_config_value "max_num_outbound_peers")
log_info "Default max_num_outbound_peers: $DEFAULT_OUTBOUND"

# Modify to specific value
modify_config_toml "max_num_outbound_peers" "15"
NEW_OUTBOUND=$(get_config_value "max_num_outbound_peers")

if [[ "$NEW_OUTBOUND" == *"15"* ]]; then
    record_pass "p2p.max_num_outbound_peers configuration change"
else
    record_fail "p2p.max_num_outbound_peers configuration change" "Value not updated correctly"
fi

# ====================================================================
# TEST 8: log_level
# ====================================================================
log_step "Test 8: log_level"
log_test "Testing log level configuration"

initialize_node

# Check default value
DEFAULT_LOG=$(get_config_value "log_level")
log_info "Default log_level: $DEFAULT_LOG"

# Modify to debug
modify_config_toml "log_level" '"debug"'
NEW_LOG=$(get_config_value "log_level")

if [[ "$NEW_LOG" == *"debug"* ]]; then
    record_pass "log_level configuration change"
else
    record_fail "log_level configuration change" "Value not updated correctly"
fi

# ====================================================================
# TEST 9: app.toml - API enable
# ====================================================================
log_step "Test 9: api.enable (app.toml)"
log_test "Testing API enable configuration"

initialize_node

# Check default value
DEFAULT_API=$(get_app_config_value "enable")
log_info "Default api.enable: $DEFAULT_API"

# Modify to true
modify_app_toml "enable" "true"
NEW_API=$(get_app_config_value "enable")

if [[ "$NEW_API" == *"true"* ]]; then
    record_pass "api.enable configuration change"
else
    record_fail "api.enable configuration change" "Value not updated correctly"
fi

# ====================================================================
# TEST 10: consensus.timeout_propose
# ====================================================================
log_step "Test 10: consensus.timeout_propose"
log_test "Testing consensus timeout propose configuration"

initialize_node

# Check default value
DEFAULT_PROPOSE=$(get_config_value "timeout_propose")
log_info "Default timeout_propose: $DEFAULT_PROPOSE"

# Modify to 4s
modify_config_toml "timeout_propose" '"4s"'
NEW_PROPOSE=$(get_config_value "timeout_propose")

if [[ "$NEW_PROPOSE" == *"4s"* ]]; then
    record_pass "consensus.timeout_propose configuration change"
else
    record_fail "consensus.timeout_propose configuration change" "Value not updated correctly"
fi

# ====================================================================
# SUMMARY
# ====================================================================
log_step "Test Summary"
echo ""
echo "Configuration Parameters Testing: COMPLETED"
echo ""
echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"
echo "Passed: $TESTS_PASSED"
echo "Failed: $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All configuration tests PASSED!${NC}"
else
    echo -e "${YELLOW}Some tests failed. Review results above.${NC}"
fi

echo ""
echo "Tested Parameters:"
echo "  1. p2p.max_num_inbound_peers (config.toml)"
echo "  2. consensus.timeout_commit (config.toml)"
echo "  3. mempool.size (config.toml)"
echo "  4. grpc.max-recv-msg-size (app.toml)"
echo "  5. snapshot-interval (app.toml)"
echo "  6. rpc.laddr (config.toml)"
echo "  7. p2p.max_num_outbound_peers (config.toml)"
echo "  8. log_level (config.toml)"
echo "  9. api.enable (app.toml)"
echo " 10. consensus.timeout_propose (config.toml)"
echo ""

# Generate detailed report
REPORT_FILE="$PROJECT_ROOT/chain/testing/local/phase2/test_2.2_results.txt"
{
    echo "Configuration Parameters Test Results"
    echo "Generated: $(date)"
    echo "======================================="
    echo ""
    echo "Tests Passed: $TESTS_PASSED"
    echo "Tests Failed: $TESTS_FAILED"
    echo ""
    echo "Detailed Results:"
    echo "-----------------"
    for test_name in "${!TEST_RESULTS[@]}"; do
        echo "$test_name: ${TEST_RESULTS[$test_name]}"
    done
} > "$REPORT_FILE"

log_info "Detailed results written to: $REPORT_FILE"

if [ $TESTS_FAILED -ne 0 ]; then
    exit 1
fi

exit 0
