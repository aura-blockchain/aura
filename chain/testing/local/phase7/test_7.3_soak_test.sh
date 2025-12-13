#!/bin/bash
# Phase 7.3: Long-Running Stability Test (Soak Test)
# Runs a testnet under moderate load for an extended period to detect memory leaks,
# performance degradation, or stability issues

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
RESULTS_FILE="$SCRIPT_DIR/test_7.3_results.txt"

# Test duration (default: 120 minutes = 2 hours for initial validation)
# For full 24-48 hour test, set DURATION_MINUTES=1440 or DURATION_MINUTES=2880
DURATION_MINUTES=${DURATION_MINUTES:-120}
DURATION_SECONDS=$((DURATION_MINUTES * 60))
DURATION_HOURS=$(echo "scale=2; $DURATION_MINUTES / 60" | bc)

# Transaction rate (txs per minute)
TX_RATE=${TX_RATE:-10}

echo "==================================================================="
echo "Phase 7.3: Long-Running Stability Test (Soak Test)"
echo "Duration: $DURATION_HOURS hours ($DURATION_SECONDS seconds)"
echo "Transaction rate: $TX_RATE tx/min"
echo "==================================================================="
echo ""

# Initialize results file
cat > "$RESULTS_FILE" << EOF
=================================================================
Phase 7.3: Long-Running Stability Test Results
=================================================================
Timestamp: $(date)
Duration: $DURATION_HOURS hours
Transaction Rate: $TX_RATE tx/min

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

# Function to get memory usage for a PID
get_memory_mb() {
    local pid="$1"
    local mem=$(ps -p $pid -o rss= 2>/dev/null || echo "0")
    echo $((mem / 1024))
}

# Function to cleanup
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"

    # Stop load generator
    if [ -n "$LOAD_PID" ] && ps -p $LOAD_PID > /dev/null 2>&1; then
        kill $LOAD_PID || true
    fi

    # Stop nodes
    for i in {1..4}; do
        NODE_PID_VAR="NODE${i}_PID"
        NODE_PID=${!NODE_PID_VAR}
        if [ -n "$NODE_PID" ] && ps -p $NODE_PID > /dev/null 2>&1; then
            kill $NODE_PID || true
        fi
    done

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

# Setup 4-node testnet
log_section "Setting up 4-node testnet"

BASE_DIR="$HOME/.aura-soak-test"
rm -rf "$BASE_DIR"
mkdir -p "$BASE_DIR"

# Initialize 4 nodes
for i in {1..4}; do
    NODE_DIR="$BASE_DIR/node$i"
    mkdir -p "$NODE_DIR"

    log_result "Initializing node$i..."
    $BINARY init "node$i" --chain-id soak-test --home "$NODE_DIR" &>> "$RESULTS_FILE"

    # Configure unique ports
    RPC_PORT=$((26657 + (i - 1) * 100))
    P2P_PORT=$((26656 + (i - 1) * 100))
    GRPC_PORT=$((9090 + (i - 1) * 100))
    API_PORT=$((1317 + (i - 1) * 100))
    PROM_PORT=$((26660 + (i - 1) * 100))

    sed -i "s/laddr = \"tcp:\/\/127.0.0.1:26657\"/laddr = \"tcp:\/\/127.0.0.1:$RPC_PORT\"/" "$NODE_DIR/config/config.toml"
    sed -i "s/laddr = \"tcp:\/\/0.0.0.0:26656\"/laddr = \"tcp:\/\/0.0.0.0:$P2P_PORT\"/" "$NODE_DIR/config/config.toml"
    sed -i "s/address = \"localhost:9090\"/address = \"localhost:$GRPC_PORT\"/" "$NODE_DIR/config/app.toml"
    sed -i "s/address = \"tcp:\/\/localhost:1317\"/address = \"tcp:\/\/localhost:$API_PORT\"/" "$NODE_DIR/config/app.toml"
    sed -i "s/prometheus_listen_addr = \":26660\"/prometheus_listen_addr = \":$PROM_PORT\"/" "$NODE_DIR/config/config.toml"

    # Enable Prometheus
    sed -i 's/prometheus = false/prometheus = true/' "$NODE_DIR/config/config.toml"
done

# Create genesis with all validators
log_result "Creating genesis file with 4 validators..."

# Generate keys and add genesis accounts
for i in {1..4}; do
    NODE_DIR="$BASE_DIR/node$i"
    ADDR=$($BINARY keys add "validator$i" --keyring-backend test --home "$NODE_DIR" 2>&1 | grep -oP 'aura[a-z0-9]{39}' | head -1)
    log_result "Node$i address: $ADDR"

    # Add genesis account to node1's genesis (we'll copy it to others later)
    $BINARY genesis add-genesis-account "$ADDR" 1000000000stake --home "$BASE_DIR/node1" --keyring-backend test &>> "$RESULTS_FILE"
done

# First, copy node1's genesis (with all accounts) to all nodes
for i in {2..4}; do
    cp "$BASE_DIR/node1/config/genesis.json" "$BASE_DIR/node$i/config/genesis.json"
done

# Create gentxs from each node
for i in {1..4}; do
    NODE_DIR="$BASE_DIR/node$i"
    $BINARY genesis gentx "validator$i" 250000000stake --chain-id soak-test --home "$NODE_DIR" --keyring-backend test &>> "$RESULTS_FILE"

    # Copy gentx to node1
    cp "$NODE_DIR/config/gentx/"* "$BASE_DIR/node1/config/gentx/" 2>/dev/null || true
done

# Collect gentxs on node1
$BINARY genesis collect-gentxs --home "$BASE_DIR/node1" &>> "$RESULTS_FILE"

# Copy genesis to all nodes
for i in {2..4}; do
    cp "$BASE_DIR/node1/config/genesis.json" "$BASE_DIR/node$i/config/genesis.json"
done

# Get node1's ID for peer configuration
NODE1_ID=$($BINARY comet show-node-id --home "$BASE_DIR/node1")

# Configure peers (all nodes connect to node1)
for i in {2..4}; do
    sed -i "s/persistent_peers = \"\"/persistent_peers = \"${NODE1_ID}@127.0.0.1:26656\"/" "$BASE_DIR/node$i/config/config.toml"
done

log_result "✅ Testnet configured with 4 validators"

# Start all nodes
log_section "Starting 4-node testnet"

for i in {1..4}; do
    NODE_DIR="$BASE_DIR/node$i"
    log_result "Starting node$i..."
    $BINARY start --home "$NODE_DIR" &> "$NODE_DIR/node.log" &
    declare "NODE${i}_PID=$!"
    NODE_PID_VAR="NODE${i}_PID"
    log_result "Node$i PID: ${!NODE_PID_VAR}"
done

# Wait for network to start
log_result "Waiting for network to initialize..."
sleep 15

# Verify all nodes are running
NODES_RUNNING=0
for i in {1..4}; do
    NODE_PID_VAR="NODE${i}_PID"
    NODE_PID=${!NODE_PID_VAR}
    if ps -p $NODE_PID > /dev/null 2>&1; then
        HEIGHT=$(get_height "$BASE_DIR/node$i/node.log")
        log_result "✅ Node$i running (height: $HEIGHT, PID: $NODE_PID)"
        NODES_RUNNING=$((NODES_RUNNING + 1))
    else
        log_result "❌ Node$i failed to start"
    fi
done

if [ $NODES_RUNNING -lt 3 ]; then
    log_result "❌ Insufficient nodes running ($NODES_RUNNING/4)"
    exit 1
fi

log_result "✅ Testnet is operational with $NODES_RUNNING/4 nodes"

# Create load generator script
log_section "Starting load generator"

LOAD_SCRIPT="$BASE_DIR/load_generator.sh"
cat > "$LOAD_SCRIPT" << 'LOADEOF'
#!/bin/bash
BINARY="$1"
NODE_DIR="$2"
TX_RATE="$3"
DURATION="$4"

TX_COUNT=0
START_TIME=$(date +%s)
END_TIME=$((START_TIME + DURATION))

echo "Load generator started at $(date)"
echo "Target: $TX_RATE tx/min for $((DURATION / 3600)) hours"

# Calculate sleep time between transactions
SLEEP_TIME=$(echo "scale=2; 60 / $TX_RATE" | bc)

while [ $(date +%s) -lt $END_TIME ]; do
    # Send a simple bank transfer
    $BINARY tx bank send validator1 aura1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a 1000stake \
        --chain-id soak-test \
        --home "$NODE_DIR" \
        --keyring-backend test \
        --yes \
        --fees 1000stake \
        --broadcast-mode async \
        &> /dev/null

    if [ $? -eq 0 ]; then
        TX_COUNT=$((TX_COUNT + 1))
        if [ $((TX_COUNT % 100)) -eq 0 ]; then
            echo "[$(date)] Sent $TX_COUNT transactions"
        fi
    fi

    sleep $SLEEP_TIME
done

echo "Load generator completed: $TX_COUNT transactions sent"
LOADEOF

chmod +x "$LOAD_SCRIPT"

# Start load generator
"$LOAD_SCRIPT" "$BINARY" "$BASE_DIR/node1" "$TX_RATE" "$DURATION_SECONDS" &> "$BASE_DIR/load.log" &
LOAD_PID=$!
log_result "✅ Load generator started (PID: $LOAD_PID)"

# Monitoring loop
log_section "Monitoring testnet stability"

log_result "Monitoring for $DURATION_HOURS hours..."
log_result "Samples will be taken every 5 minutes"
log_result ""

START_TIME=$(date +%s)
END_TIME=$((START_TIME + DURATION_SECONDS))
SAMPLE_INTERVAL=300  # 5 minutes

SAMPLE_COUNT=0
declare -A INITIAL_MEM
declare -A INITIAL_HEIGHT

# Record initial state
for i in {1..4}; do
    NODE_PID_VAR="NODE${i}_PID"
    NODE_PID=${!NODE_PID_VAR}
    if ps -p $NODE_PID > /dev/null 2>&1; then
        INITIAL_MEM[$i]=$(get_memory_mb $NODE_PID)
        INITIAL_HEIGHT[$i]=$(get_height "$BASE_DIR/node$i/node.log")
    fi
done

# Monitoring loop
while [ $(date +%s) -lt $END_TIME ]; do
    sleep $SAMPLE_INTERVAL
    SAMPLE_COUNT=$((SAMPLE_COUNT + 1))

    ELAPSED=$(($(date +%s) - START_TIME))
    ELAPSED_MIN=$((ELAPSED / 60))
    REMAINING=$((END_TIME - $(date +%s)))
    REMAINING_MIN=$((REMAINING / 60))

    log_result "=== Sample #$SAMPLE_COUNT (Elapsed: ${ELAPSED_MIN}min, Remaining: ${REMAINING_MIN}min) ==="

    # Check each node
    FAILED_NODES=0
    for i in {1..4}; do
        NODE_PID_VAR="NODE${i}_PID"
        NODE_PID=${!NODE_PID_VAR}

        if ps -p $NODE_PID > /dev/null 2>&1; then
            MEM=$(get_memory_mb $NODE_PID)
            HEIGHT=$(get_height "$BASE_DIR/node$i/node.log")

            # Calculate memory growth
            INITIAL=${INITIAL_MEM[$i]}
            MEM_GROWTH=$((MEM - INITIAL))
            MEM_GROWTH_PCT=$(echo "scale=1; ($MEM_GROWTH * 100) / $INITIAL" | bc)

            # Calculate block rate
            BLOCK_DIFF=$((HEIGHT - ${INITIAL_HEIGHT[$i]}))
            BLOCK_RATE=$(echo "scale=2; $BLOCK_DIFF / $ELAPSED" | bc)

            log_result "Node$i: MEM=${MEM}MB (+${MEM_GROWTH}MB, +${MEM_GROWTH_PCT}%), HEIGHT=$HEIGHT, RATE=${BLOCK_RATE} blk/s"

            # Check for memory leak (>50% growth)
            if [ $MEM_GROWTH_PCT -gt 50 ]; then
                log_result "⚠️  Node$i: Possible memory leak detected (${MEM_GROWTH_PCT}% growth)"
            fi
        else
            log_result "❌ Node$i: Process died"
            FAILED_NODES=$((FAILED_NODES + 1))
        fi
    done

    # Check if consensus can continue
    if [ $FAILED_NODES -gt 1 ]; then
        log_result "❌ Too many nodes failed ($FAILED_NODES), halting test"
        break
    fi

    log_result ""
done

# Final assessment
log_section "Final Assessment"

# Stop load generator if still running
if ps -p $LOAD_PID > /dev/null 2>&1; then
    kill $LOAD_PID || true
fi

# Get final stats
log_result "Final node states:"
for i in {1..4}; do
    NODE_PID_VAR="NODE${i}_PID"
    NODE_PID=${!NODE_PID_VAR}

    if ps -p $NODE_PID > /dev/null 2>&1; then
        FINAL_MEM=$(get_memory_mb $NODE_PID)
        FINAL_HEIGHT=$(get_height "$BASE_DIR/node$i/node.log")

        TOTAL_BLOCKS=$((FINAL_HEIGHT - ${INITIAL_HEIGHT[$i]}))
        AVG_BLOCK_RATE=$(echo "scale=2; $TOTAL_BLOCKS / $DURATION_SECONDS" | bc)

        log_result "Node$i: ✅ ALIVE | Final height: $FINAL_HEIGHT | Blocks produced: $TOTAL_BLOCKS | Avg rate: $AVG_BLOCK_RATE blk/s | Final mem: ${FINAL_MEM}MB"
    else
        log_result "Node$i: ❌ DEAD"
    fi
done

# Check load generator results
if [ -f "$BASE_DIR/load.log" ]; then
    TX_SENT=$(grep "transactions sent" "$BASE_DIR/load.log" | grep -oP '\d+' | head -1 || echo "0")
    log_result ""
    log_result "Load generator results: $TX_SENT transactions sent"
fi

# Summary
log_section "Test 7.3 Summary"

log_result "Test Duration: $DURATION_HOURS hours"
log_result "Network stayed operational: $(if [ $FAILED_NODES -lt 2 ]; then echo "YES"; else echo "NO"; fi)"
log_result "Samples collected: $SAMPLE_COUNT"
log_result ""

echo ""
echo "=================================================================="
echo "Test 7.3 Complete - Results saved to:"
echo "$RESULTS_FILE"
echo "=================================================================="
echo ""
echo "To run full 24-hour test:"
echo "  DURATION_MINUTES=1440 $0"
echo ""
echo "To run full 48-hour test:"
echo "  DURATION_MINUTES=2880 $0"
echo ""

# Display summary
cat "$RESULTS_FILE" | grep -E "^(✅|❌|⚠️|Node[0-9]:)" | tail -30 | tee -a "$RESULTS_FILE"

echo ""
log_result "Test execution completed at: $(date)"
