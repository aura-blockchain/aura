#!/bin/bash
# Phase 7.2: Resource Constraint Test (Simplified)
# Tests node performance under resource constraints using cgroups/systemd-run

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
TEST_DIR_BASE="$HOME/.aura-resource-test"
RESULTS_FILE="$SCRIPT_DIR/test_7.2_results.txt"

echo "==================================================================="
echo "Phase 7.2: Resource Constraint Test"
echo "==================================================================="
echo ""

# Initialize results file
cat > "$RESULTS_FILE" << EOF
=================================================================
Phase 7.2: Resource Constraint Test Results
=================================================================
Timestamp: $(date)

EOF

# Function to log results
log_result() {
    echo "$1" | tee -a "$RESULTS_FILE"
}

# Function to log section header
log_section() {
    echo ""
    echo "-----------------------------------------------------------------"
    echo "$1"
    echo "-----------------------------------------------------------------"
    echo "" | tee -a "$RESULTS_FILE"
    echo "$1" >> "$RESULTS_FILE"
    echo "-----------------------------------------------------------------" >> "$RESULTS_FILE"
}

# Function to get height from log
get_height() {
    local log_file="$1"
    grep "committed state" "$log_file" 2>/dev/null | tail -1 | grep -oP 'height=\K[0-9]+' || echo "0"
}

# Function to cleanup
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"

    # Kill any running aurad processes
    pkill -f "aurad.*resource-test" || true

    sleep 2

    echo "Cleanup complete"
}

trap cleanup EXIT

# Build binary
log_section "Building aurad binary"
cd "$PROJECT_ROOT/chain"
if go build -o "$PROJECT_ROOT/chain/aurad" ./cmd/aurad 2>&1 | tee -a "$RESULTS_FILE"; then
    log_result "✅ Binary built successfully"
else
    log_result "❌ Binary build failed"
    exit 1
fi

BINARY="$PROJECT_ROOT/chain/aurad"

# Test 1: Baseline (no constraints)
log_section "Test 1: Baseline Performance (No Resource Constraints)"

TEST_DIR="$TEST_DIR_BASE-baseline"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# Initialize
$BINARY init baseline-node --chain-id resource-test --home "$TEST_DIR" &>> "$RESULTS_FILE"

# Configure ports
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/127.0.0.1:36657"/' "$TEST_DIR/config/config.toml"
sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:36656"/' "$TEST_DIR/config/config.toml"
sed -i 's/address = "localhost:9090"/address = "localhost:9190"/' "$TEST_DIR/config/app.toml"
sed -i 's/address = "localhost:9091"/address = "localhost:9191"/' "$TEST_DIR/config/app.toml"

# Add genesis account
ADDR=$($BINARY keys add validator --keyring-backend test --home "$TEST_DIR" 2>&1 | grep -oP 'aura[a-z0-9]{39}' | head -1)
$BINARY genesis add-genesis-account "$ADDR" 1000000000stake --home "$TEST_DIR" --keyring-backend test &>> "$RESULTS_FILE"
$BINARY genesis gentx validator 500000000stake --chain-id resource-test --home "$TEST_DIR" --keyring-backend test &>> "$RESULTS_FILE"
$BINARY genesis collect-gentxs --home "$TEST_DIR" &>> "$RESULTS_FILE"

log_result "Starting baseline node..."
$BINARY start --home "$TEST_DIR" &> "$TEST_DIR/node.log" &
BASELINE_PID=$!

sleep 10

# Check performance
BASELINE_HEIGHT=$(get_height "$TEST_DIR/node.log")
if [ "$BASELINE_HEIGHT" -gt 0 ]; then
    log_result "✅ Baseline node producing blocks (height: $BASELINE_HEIGHT)"

    # Get resource usage
    BASELINE_MEM=$(ps -p $BASELINE_PID -o rss= 2>/dev/null || echo "0")
    BASELINE_MEM_MB=$((BASELINE_MEM / 1024))
    log_result "Baseline memory usage: ${BASELINE_MEM_MB}MB"
else
    log_result "❌ Baseline node failed to produce blocks"
fi

# Stop baseline
kill $BASELINE_PID 2>/dev/null || true
sleep 2

# Test 2: Memory-constrained test using ulimit
log_section "Test 2: Memory Constrained (512MB limit)"

TEST_DIR="$TEST_DIR_BASE-512mb"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# Initialize
$BINARY init constrained-node --chain-id resource-test --home "$TEST_DIR" &>> "$RESULTS_FILE"

# Configure ports (different from baseline)
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/127.0.0.1:37657"/' "$TEST_DIR/config/config.toml"
sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:37656"/' "$TEST_DIR/config/config.toml"
sed -i 's/address = "localhost:9090"/address = "localhost:9290"/' "$TEST_DIR/config/app.toml"
sed -i 's/address = "localhost:9091"/address = "localhost:9291"/' "$TEST_DIR/config/app.toml"

# Add genesis account
ADDR=$($BINARY keys add validator --keyring-backend test --home "$TEST_DIR" 2>&1 | grep -oP 'aura[a-z0-9]{39}' | head -1)
$BINARY genesis add-genesis-account "$ADDR" 1000000000stake --home "$TEST_DIR" --keyring-backend test &>> "$RESULTS_FILE"
$BINARY genesis gentx validator 500000000stake --chain-id resource-test --home "$TEST_DIR" --keyring-backend test &>> "$RESULTS_FILE"
$BINARY genesis collect-gentxs --home "$TEST_DIR" &>> "$RESULTS_FILE"

log_result "Starting node with 512MB memory limit..."

# Use systemd-run if available, otherwise just run normally with monitoring
if command -v systemd-run &> /dev/null; then
    log_result "Using systemd-run for resource limiting..."
    systemd-run --user --scope -p MemoryMax=512M -p MemorySwapMax=0 \
        $BINARY start --home "$TEST_DIR" &> "$TEST_DIR/node.log" &
    CONSTRAINED_PID=$!
else
    log_result "systemd-run not available, running with manual monitoring..."
    $BINARY start --home "$TEST_DIR" &> "$TEST_DIR/node.log" &
    CONSTRAINED_PID=$!
fi

sleep 15

# Monitor for 60 seconds
log_result "Monitoring for 60 seconds..."
SURVIVED=true
for i in {1..12}; do
    sleep 5

    if ! ps -p $CONSTRAINED_PID > /dev/null 2>&1; then
        log_result "❌ Node stopped at $(($i * 5))s"
        SURVIVED=false
        break
    fi

    # Check memory usage
    MEM=$(ps -p $CONSTRAINED_PID -o rss= 2>/dev/null || echo "0")
    MEM_MB=$((MEM / 1024))
    log_result "[$((i * 5))s] Memory usage: ${MEM_MB}MB"

    if [ $MEM_MB -gt 512 ]; then
        log_result "⚠️  Memory usage exceeded 512MB limit"
    fi
done

if $SURVIVED; then
    CONSTRAINED_HEIGHT=$(get_height "$TEST_DIR/node.log")
    if [ "$CONSTRAINED_HEIGHT" -gt 0 ]; then
        log_result "✅ Node survived 60s and produced blocks (height: $CONSTRAINED_HEIGHT)"

        # Calculate performance ratio
        if [ "$BASELINE_HEIGHT" -gt 0 ]; then
            RATIO=$(echo "scale=2; ($CONSTRAINED_HEIGHT * 100) / $BASELINE_HEIGHT" | bc)
            log_result "Performance under constraint: ${RATIO}% of baseline"
        fi
    else
        log_result "⚠️  Node survived but didn't produce blocks"
    fi

    kill $CONSTRAINED_PID 2>/dev/null || true
else
    log_result "❌ Node did not survive 60 seconds under 512MB constraint"
fi

sleep 2

# Test 3: CPU throttling test
log_section "Test 3: CPU Limited Test (using cpulimit if available)"

if command -v cpulimit &> /dev/null; then
    TEST_DIR="$TEST_DIR_BASE-cpulimit"
    rm -rf "$TEST_DIR"
    mkdir -p "$TEST_DIR"

    # Initialize
    $BINARY init cpulimit-node --chain-id resource-test --home "$TEST_DIR" &>> "$RESULTS_FILE"

    # Configure ports
    sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/127.0.0.1:38657"/' "$TEST_DIR/config/config.toml"
    sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:38656"/' "$TEST_DIR/config/config.toml"
    sed -i 's/address = "localhost:9090"/address = "localhost:9390"/' "$TEST_DIR/config/app.toml"
    sed -i 's/address = "localhost:9091"/address = "localhost:9391"/' "$TEST_DIR/config/app.toml"

    # Add genesis account
    ADDR=$($BINARY keys add validator --keyring-backend test --home "$TEST_DIR" 2>&1 | grep -oP 'aura[a-z0-9]{39}' | head -1)
    $BINARY genesis add-genesis-account "$ADDR" 1000000000stake --home "$TEST_DIR" --keyring-backend test &>> "$RESULTS_FILE"
    $BINARY genesis gentx validator 500000000stake --chain-id resource-test --home "$TEST_DIR" --keyring-backend test &>> "$RESULTS_FILE"
    $BINARY genesis collect-gentxs --home "$TEST_DIR" &>> "$RESULTS_FILE"

    log_result "Starting node with 50% CPU limit..."
    $BINARY start --home "$TEST_DIR" &> "$TEST_DIR/node.log" &
    CPU_PID=$!

    sleep 2

    # Apply CPU limit (50% of one core)
    cpulimit -p $CPU_PID -l 50 -b &>> "$RESULTS_FILE"
    CPU_LIMIT_PID=$!

    sleep 30

    if ps -p $CPU_PID > /dev/null 2>&1; then
        CPU_HEIGHT=$(get_height "$TEST_DIR/node.log")
        if [ "$CPU_HEIGHT" -gt 0 ]; then
            log_result "✅ Node with 50% CPU limit produced blocks (height: $CPU_HEIGHT)"
        else
            log_result "⚠️  Node running but no blocks under CPU limit"
        fi

        kill $CPU_PID 2>/dev/null || true
        kill $CPU_LIMIT_PID 2>/dev/null || true
    else
        log_result "❌ Node stopped under CPU limit"
    fi
else
    log_result "⚠️  cpulimit not installed, skipping CPU throttling test"
    log_result "Install with: sudo apt-get install cpulimit"
fi

# Summary
log_section "Test 7.2 Summary"

log_result ""
log_result "System Requirements Assessment:"
log_result "- Baseline memory usage: ${BASELINE_MEM_MB:-N/A}MB"
log_result "- Node survives with 512MB RAM: $(if $SURVIVED; then echo "YES"; else echo "NO"; fi)"
log_result ""
log_result "Recommendations:"
log_result "- Minimum RAM: 512MB (may be limited)"
log_result "- Recommended RAM: ≥ 1GB for stable operation"
log_result "- Optimal RAM: ≥ 2GB for production use"
log_result "- CPU: ≥ 1 core recommended, ≥ 2 cores for production"
log_result ""

echo ""
echo "=================================================================="
echo "Test 7.2 Complete - Results saved to:"
echo "$RESULTS_FILE"
echo "=================================================================="
echo ""

# Display summary
cat "$RESULTS_FILE" | grep -E "^(✅|❌|⚠️)" | tee -a "$RESULTS_FILE"

echo ""
log_result "Test execution completed at: $(date)"
