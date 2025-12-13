#!/bin/bash
# Phase 7.1: Database Corruption Test
# Tests node's ability to detect and recover from database corruption

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
TEST_DIR="$HOME/.aura-corruption-test"
RESULTS_FILE="$SCRIPT_DIR/test_7.1_results.txt"
CHAIN_ID="aura-corruption-test"

echo "==================================================================="
echo "Phase 7.1: Database Corruption Test"
echo "==================================================================="
echo ""

# Initialize results file
cat > "$RESULTS_FILE" << EOF
=================================================================
Phase 7.1: Database Corruption Test Results
=================================================================
Timestamp: $(date)
Test Directory: $TEST_DIR

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

# Function to get latest height
get_height() {
    local log_file="${1:-$TEST_DIR/node.log}"
    # Try status command first
    local height=$($BINARY status --node tcp://localhost:36657 --home "$TEST_DIR" 2>/dev/null | jq -r '.sync_info.latest_block_height' 2>/dev/null || echo "0")

    # Fallback to log parsing if status fails
    if [ "$height" = "0" ] || [ "$height" = "null" ] || [ -z "$height" ]; then
        height=$(grep "committed state" "$log_file" 2>/dev/null | tail -1 | grep -oP 'height=\K[0-9]+' || echo "0")
    fi

    echo "$height"
}

# Function to cleanup
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"

    # Stop any running processes
    pkill -f "aurad.*$CHAIN_ID" || true

    # Wait a moment
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

# Clean test directory
log_section "Initializing test node"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# Initialize node
if $BINARY init test-node --chain-id "$CHAIN_ID" --home "$TEST_DIR" &>> "$RESULTS_FILE"; then
    log_result "✅ Node initialized successfully"
else
    log_result "❌ Node initialization failed"
    exit 1
fi

# Configure unique ports to avoid conflicts
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/127.0.0.1:36657"/' "$TEST_DIR/config/config.toml"
sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:36656"/' "$TEST_DIR/config/config.toml"
sed -i 's/address = "localhost:9090"/address = "localhost:9190"/' "$TEST_DIR/config/app.toml"
sed -i 's/address = "localhost:9091"/address = "localhost:9191"/' "$TEST_DIR/config/app.toml"
sed -i 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/localhost:1417"/' "$TEST_DIR/config/app.toml"
sed -i 's/prometheus_listen_addr = ":26660"/prometheus_listen_addr = ":36660"/' "$TEST_DIR/config/config.toml"

# Add genesis account and create validator
TEST_ADDR=$($BINARY keys add test-validator --keyring-backend test --home "$TEST_DIR" 2>&1 | grep -oP 'aura[a-z0-9]{39}' | head -1)
log_result "Test address: $TEST_ADDR"

$BINARY genesis add-genesis-account "$TEST_ADDR" 1000000000stake --home "$TEST_DIR" --keyring-backend test &>> "$RESULTS_FILE"
$BINARY genesis gentx test-validator 500000000stake --chain-id "$CHAIN_ID" --home "$TEST_DIR" --keyring-backend test &>> "$RESULTS_FILE"
$BINARY genesis collect-gentxs --home "$TEST_DIR" &>> "$RESULTS_FILE"

# Start node and let it run for a bit to create some state
log_section "Starting node to generate initial state"
$BINARY start --home "$TEST_DIR" &> "$TEST_DIR/node.log" &
NODE_PID=$!
log_result "Node PID: $NODE_PID"

echo "Waiting for node to produce blocks..."
sleep 10

# Check if blocks are being produced
LATEST_HEIGHT=$(get_height "$TEST_DIR/node.log")

if [ "$LATEST_HEIGHT" -gt 0 ]; then
    log_result "✅ Node is producing blocks (height: $LATEST_HEIGHT)"
else
    log_result "❌ Node failed to produce blocks"
    kill $NODE_PID || true
    exit 1
fi

# Stop the node
log_section "Stopping node for corruption test"
kill $NODE_PID
sleep 3

# Verify node is stopped
if ps -p $NODE_PID > /dev/null 2>&1; then
    log_result "⚠️  Node still running, forcing kill"
    kill -9 $NODE_PID || true
    sleep 2
fi
log_result "✅ Node stopped"

# Test 1: Corrupt application.db
log_section "Test 1: Application DB Corruption"
APP_DB_DIR="$TEST_DIR/data/application.db"

if [ -d "$APP_DB_DIR" ]; then
    log_result "Application DB directory found: $APP_DB_DIR"

    # Find the largest .ldb or .log file to corrupt
    APP_DB_FILE=$(find "$APP_DB_DIR" -type f \( -name "*.ldb" -o -name "*.log" \) -exec stat -c "%s %n" {} \; 2>/dev/null | sort -rn | head -1 | awk '{print $2}')

    if [ -n "$APP_DB_FILE" ] && [ -f "$APP_DB_FILE" ]; then
        log_result "Corrupting file: $APP_DB_FILE"

        # Backup entire directory
        cp -r "$APP_DB_DIR" "$APP_DB_DIR.backup"

        # Corrupt the database file by writing garbage at random offset
        DB_SIZE=$(stat -f%z "$APP_DB_FILE" 2>/dev/null || stat -c%s "$APP_DB_FILE")
        CORRUPT_OFFSET=$((DB_SIZE / 2))
        log_result "DB file size: $DB_SIZE bytes, corrupting at offset: $CORRUPT_OFFSET"

        # Write 1KB of garbage data
        dd if=/dev/urandom of="$APP_DB_FILE" bs=1024 count=1 seek=$((CORRUPT_OFFSET / 1024)) conv=notrunc 2>&1 | tee -a "$RESULTS_FILE"

        log_result "✅ Database corrupted"
    else
        log_result "⚠️  No suitable database file found to corrupt"
        APP_DB_DIR=""
    fi
else
    log_result "⚠️  Application DB directory not found"
    APP_DB_DIR=""
fi

if [ -n "$APP_DB_DIR" ]; then

    # Try to restart node
    log_result "Attempting to restart with corrupted application.db..."

    # Capture startup output
    timeout 10 $BINARY start --home "$TEST_DIR" &> "$TEST_DIR/corrupt_app_db.log" &
    CORRUPT_PID=$!

    sleep 5

    # Check if node is still running (it shouldn't be)
    if ps -p $CORRUPT_PID > /dev/null 2>&1; then
        log_result "⚠️  Node unexpectedly running with corrupted DB"
        kill $CORRUPT_PID || true

        # Check log for error messages
        if grep -i "error\|corrupt\|panic" "$TEST_DIR/corrupt_app_db.log" > /dev/null; then
            log_result "✅ Node logged errors about corruption"
            cat "$TEST_DIR/corrupt_app_db.log" | tail -20 >> "$RESULTS_FILE"
        else
            log_result "❌ No corruption errors logged"
        fi
    else
        log_result "✅ Node failed to start with corrupted DB (expected)"

        # Check for clear error message
        if grep -i "corrupt\|invalid\|error.*database" "$TEST_DIR/corrupt_app_db.log" > /dev/null; then
            log_result "✅ Clear corruption error message found"
            echo "Error excerpt:" >> "$RESULTS_FILE"
            grep -i "error\|corrupt\|panic" "$TEST_DIR/corrupt_app_db.log" | head -10 >> "$RESULTS_FILE"
        else
            log_result "⚠️  No clear corruption error message"
            echo "Log excerpt:" >> "$RESULTS_FILE"
            tail -20 "$TEST_DIR/corrupt_app_db.log" >> "$RESULTS_FILE"
        fi
    fi

    # Restore from backup
    log_result "Restoring application.db from backup..."
    rm -rf "$APP_DB_DIR"
    mv "$APP_DB_DIR.backup" "$APP_DB_DIR"

    # Verify restoration
    log_result "Verifying node can restart with restored DB..."
    timeout 10 $BINARY start --home "$TEST_DIR" &> "$TEST_DIR/restored_app_db.log" &
    RESTORE_PID=$!

    sleep 5

    if ps -p $RESTORE_PID > /dev/null 2>&1; then
        sleep 2  # Give it a moment to start producing blocks
        NEW_HEIGHT=$(get_height "$TEST_DIR/restored_app_db.log")
        log_result "✅ Node successfully restarted with restored DB (height: $NEW_HEIGHT)"
        kill $RESTORE_PID || true
    else
        log_result "❌ Node failed to restart even with restored DB"
    fi

else
    log_result "⚠️  Application DB not found, skipping test"
fi

sleep 2

# Test 2: Corrupt state.db (blockstore)
log_section "Test 2: State/Blockstore DB Corruption"
STATE_DB_DIR="$TEST_DIR/data/state.db"

if [ -d "$STATE_DB_DIR" ]; then
    log_result "State DB directory found: $STATE_DB_DIR"

    # Find the largest .ldb or .log file to corrupt
    STATE_DB_FILE=$(find "$STATE_DB_DIR" -type f \( -name "*.ldb" -o -name "*.log" \) -exec stat -c "%s %n" {} \; 2>/dev/null | sort -rn | head -1 | awk '{print $2}')

    if [ -n "$STATE_DB_FILE" ] && [ -f "$STATE_DB_FILE" ]; then
        log_result "Corrupting file: $STATE_DB_FILE"

        # Backup entire directory
        cp -r "$STATE_DB_DIR" "$STATE_DB_DIR.backup"

        # Corrupt the database file
        DB_SIZE=$(stat -f%z "$STATE_DB_FILE" 2>/dev/null || stat -c%s "$STATE_DB_FILE")
        CORRUPT_OFFSET=$((DB_SIZE / 3))
        log_result "DB file size: $DB_SIZE bytes, corrupting at offset: $CORRUPT_OFFSET"

        # Write 1KB of garbage data
        dd if=/dev/urandom of="$STATE_DB_FILE" bs=1024 count=1 seek=$((CORRUPT_OFFSET / 1024)) conv=notrunc 2>&1 | tee -a "$RESULTS_FILE"

        log_result "✅ State database corrupted"
    else
        log_result "⚠️  No suitable database file found to corrupt"
        STATE_DB_DIR=""
    fi
else
    log_result "⚠️  State DB directory not found"
    STATE_DB_DIR=""
fi

if [ -n "$STATE_DB_DIR" ]; then

    # Try to restart node
    log_result "Attempting to restart with corrupted state.db..."

    timeout 10 $BINARY start --home "$TEST_DIR" &> "$TEST_DIR/corrupt_state_db.log" &
    CORRUPT_PID=$!

    sleep 5

    # Check if node is still running
    if ps -p $CORRUPT_PID > /dev/null 2>&1; then
        log_result "⚠️  Node unexpectedly running with corrupted state DB"
        kill $CORRUPT_PID || true

        # Check log for error messages
        if grep -i "error\|corrupt\|panic" "$TEST_DIR/corrupt_state_db.log" > /dev/null; then
            log_result "✅ Node logged errors about corruption"
            cat "$TEST_DIR/corrupt_state_db.log" | tail -20 >> "$RESULTS_FILE"
        fi
    else
        log_result "✅ Node failed to start with corrupted state DB (expected)"

        # Check for clear error message
        if grep -i "corrupt\|invalid\|error.*database\|panic" "$TEST_DIR/corrupt_state_db.log" > /dev/null; then
            log_result "✅ Clear corruption error message found"
            echo "Error excerpt:" >> "$RESULTS_FILE"
            grep -i "error\|corrupt\|panic" "$TEST_DIR/corrupt_state_db.log" | head -10 >> "$RESULTS_FILE"
        else
            log_result "⚠️  No clear corruption error message"
            echo "Log excerpt:" >> "$RESULTS_FILE"
            tail -20 "$TEST_DIR/corrupt_state_db.log" >> "$RESULTS_FILE"
        fi
    fi

    # Restore from backup
    log_result "Restoring state.db from backup..."
    rm -rf "$STATE_DB_DIR"
    mv "$STATE_DB_DIR.backup" "$STATE_DB_DIR"

    # Verify restoration
    log_result "Verifying node can restart with restored DB..."
    timeout 10 $BINARY start --home "$TEST_DIR" &> "$TEST_DIR/restored_state_db.log" &
    RESTORE_PID=$!

    sleep 5

    if ps -p $RESTORE_PID > /dev/null 2>&1; then
        sleep 2  # Give it a moment to start producing blocks
        NEW_HEIGHT=$(get_height "$TEST_DIR/restored_state_db.log")
        log_result "✅ Node successfully restarted with restored DB (height: $NEW_HEIGHT)"
        kill $RESTORE_PID || true
    else
        log_result "❌ Node failed to restart even with restored DB"
    fi

else
    log_result "⚠️  State DB not found, skipping test"
fi

sleep 2

# Test 3: Test recovery from unsafe-reset-all
log_section "Test 3: Recovery via unsafe-reset-all"

log_result "Creating fresh state for reset test..."
$BINARY start --home "$TEST_DIR" &> "$TEST_DIR/pre_reset.log" &
PRE_RESET_PID=$!
sleep 5

LATEST_HEIGHT=$(get_height "$TEST_DIR/pre_reset.log")
log_result "Current height before reset: $LATEST_HEIGHT"

kill $PRE_RESET_PID
sleep 2

# Perform unsafe-reset-all
log_result "Executing unsafe-reset-all..."
if $BINARY comet unsafe-reset-all --home "$TEST_DIR" &>> "$RESULTS_FILE"; then
    log_result "✅ unsafe-reset-all executed successfully"

    # Verify data was cleared
    if [ ! -f "$TEST_DIR/data/application.db" ] || [ $(stat -f%z "$TEST_DIR/data/application.db" 2>/dev/null || stat -c%s "$TEST_DIR/data/application.db") -lt 1000 ]; then
        log_result "✅ Database cleared/reset"
    else
        log_result "⚠️  Database may not have been fully reset"
    fi

    # Try to restart node after reset
    log_result "Attempting to restart after unsafe-reset-all..."
    $BINARY start --home "$TEST_DIR" &> "$TEST_DIR/post_reset.log" &
    POST_RESET_PID=$!

    sleep 5

    if ps -p $POST_RESET_PID > /dev/null 2>&1; then
        sleep 2  # Give it a moment to start producing blocks
        NEW_HEIGHT=$(get_height "$TEST_DIR/post_reset.log")
        log_result "✅ Node restarted successfully after reset (height: $NEW_HEIGHT)"

        if [ "$NEW_HEIGHT" -lt "$LATEST_HEIGHT" ] || [ "$NEW_HEIGHT" -eq 0 ]; then
            log_result "✅ Height reset confirmed (was $LATEST_HEIGHT, now $NEW_HEIGHT)"
        fi

        kill $POST_RESET_PID
    else
        log_result "❌ Node failed to restart after unsafe-reset-all"
    fi
else
    log_result "❌ unsafe-reset-all command failed"
fi

# Summary
log_section "Test 7.1 Summary"

echo ""
echo "=================================================================="
echo "Test 7.1 Complete - Results saved to:"
echo "$RESULTS_FILE"
echo "=================================================================="
echo ""

# Display summary
cat "$RESULTS_FILE" | grep -E "^(✅|❌|⚠️)" | tee -a "$RESULTS_FILE"

echo ""
log_result "Test execution completed at: $(date)"
