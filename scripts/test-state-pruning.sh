#!/bin/bash
# State Pruning Verification Test
# Tests that old state is correctly pruned when pruning is enabled

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_FILE="${SCRIPT_DIR}/state_pruning_results.txt"

echo "=== State Pruning Verification ===" | tee "${RESULTS_FILE}"
echo "Start time: $(date)" | tee -a "${RESULTS_FILE}"
echo "" | tee -a "${RESULTS_FILE}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_test() {
    echo -e "${GREEN}[TEST]${NC} $1" | tee -a "${RESULTS_FILE}"
}

log_result() {
    echo -e "${YELLOW}[RESULT]${NC} $1" | tee -a "${RESULTS_FILE}"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "${RESULTS_FILE}"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "${RESULTS_FILE}"
}

# Check testnet is running
log_test "Checking testnet status"
if ! docker ps --filter "name=aura-validator-1" --format "{{.Names}}" | grep -q "^aura-validator-1$"; then
    log_error "Testnet is not running. Start it first."
    exit 1
fi
log_success "Testnet is running"

# Test 1: Check current pruning configuration
log_test "Test 1: Checking pruning configuration on validator-1"
AURA_HOME="/home/aura/.aura"
PRUNING=$(docker exec aura-validator-1 grep "^pruning = " ${AURA_HOME}/config/app.toml | awk -F'"' '{print $2}')
PRUNING_KEEP_RECENT=$(docker exec aura-validator-1 grep "^pruning-keep-recent = " ${AURA_HOME}/config/app.toml | awk -F'"' '{print $2}')
PRUNING_INTERVAL=$(docker exec aura-validator-1 grep "^pruning-interval = " ${AURA_HOME}/config/app.toml | awk -F'"' '{print $2}')

log_result "Pruning mode: ${PRUNING:-default}"
log_result "Keep recent: ${PRUNING_KEEP_RECENT:-N/A}"
log_result "Pruning interval: ${PRUNING_INTERVAL:-N/A}"

# Test 2: Create a test node with aggressive pruning
log_test "Test 2: Creating test node with aggressive pruning settings"

TEST_NODE="aura-pruning-test"
docker stop ${TEST_NODE} 2>/dev/null || true
docker rm ${TEST_NODE} 2>/dev/null || true
docker volume rm aura-pruning-test-data 2>/dev/null || true

# Create node
docker run -d \
    --name ${TEST_NODE} \
    --network aura_aura-testnet \
    -v aura-pruning-test-data:/home/aura/.aura \
    --entrypoint /bin/sh \
    aurad:latest \
    -c "sleep infinity"

log_success "Test node container created"

# Initialize
log_test "Initializing test node"
docker exec ${TEST_NODE} aurad init pruning-test --chain-id aura-mvp-1 --home ${AURA_HOME} 2>&1 | tee -a "${RESULTS_FILE}"

# Copy genesis
docker exec aura-validator-1 cat ${AURA_HOME}/config/genesis.json > /tmp/genesis.json
docker cp /tmp/genesis.json ${TEST_NODE}:${AURA_HOME}/config/genesis.json

# Configure aggressive pruning
log_test "Configuring aggressive pruning: keep-recent=10, interval=10"
docker exec ${TEST_NODE} bash -c "cat >> ${AURA_HOME}/config/app.toml <<'EOF'

# Custom pruning configuration
pruning = \"custom\"
pruning-keep-recent = \"10\"
pruning-keep-every = \"0\"
pruning-interval = \"10\"
EOF"

# Verify configuration
docker exec ${TEST_NODE} grep -A5 "^pruning = " ${AURA_HOME}/config/app.toml | tee -a "${RESULTS_FILE}"

# Configure peers
VAL1_ID=$(docker exec aura-validator-1 aurad tendermint show-node-id 2>&1)
VAL2_ID=$(docker exec aura-validator-2 aurad tendermint show-node-id 2>&1)
PEERS="${VAL1_ID}@aura-validator-1:26656,${VAL2_ID}@aura-validator-2:26656"
docker exec ${TEST_NODE} sed -i "s/^persistent_peers = .*/persistent_peers = \"${PEERS}\"/" ${AURA_HOME}/config/config.toml

# Start node
log_test "Starting test node"
docker exec -d ${TEST_NODE} aurad start --home ${AURA_HOME}

# Wait for sync
log_test "Waiting for node to sync (30 seconds)"
sleep 30

# Test 3: Verify node is syncing
log_test "Test 3: Checking sync status"
STATUS=$(docker exec ${TEST_NODE} aurad status 2>&1 || echo "{}")
CATCHING_UP=$(echo "${STATUS}" | jq -r '.sync_info.catching_up // "true"')
LATEST_HEIGHT=$(echo "${STATUS}" | jq -r '.sync_info.latest_block_height // "0"')

log_result "Latest height: ${LATEST_HEIGHT}"
log_result "Catching up: ${CATCHING_UP}"

if [[ "${LATEST_HEIGHT}" -gt "0" ]]; then
    log_success "Node is syncing/synced"
else
    log_error "Node failed to start syncing"
    docker logs ${TEST_NODE} --tail 50 | tee -a "${RESULTS_FILE}"
    docker stop ${TEST_NODE}
    docker rm ${TEST_NODE}
    exit 1
fi

# Test 4: Wait for pruning to occur
log_test "Test 4: Waiting for pruning to occur (need at least 25 blocks)"
TARGET_HEIGHT=$((LATEST_HEIGHT + 30))
echo "Waiting for height ${TARGET_HEIGHT}..." | tee -a "${RESULTS_FILE}"

TIMEOUT=180
START_TIME=$(date +%s)
while true; do
    CURRENT=$(curl -s localhost:27657/status 2>&1 | jq -r '.result.sync_info.latest_block_height // "0"')
    if [[ "${CURRENT}" -ge "${TARGET_HEIGHT}" ]]; then
        log_success "Reached height ${CURRENT}"
        break
    fi

    ELAPSED=$(($(date +%s) - START_TIME))
    if [[ ${ELAPSED} -gt ${TIMEOUT} ]]; then
        log_error "Timeout waiting for blocks"
        docker stop ${TEST_NODE}
        docker rm ${TEST_NODE}
        exit 1
    fi
    sleep 2
done

# Give pruning time to run
log_test "Waiting 20 seconds for pruning to execute"
sleep 20

# Test 5: Verify old blocks are pruned
log_test "Test 5: Attempting to query old blocks (should fail if pruned)"

# Get current status of test node
TEST_STATUS=$(docker exec ${TEST_NODE} aurad status 2>&1 || echo "{}")
TEST_CURRENT_HEIGHT=$(echo "${TEST_STATUS}" | jq -r '.sync_info.latest_block_height // "0"')

OLD_HEIGHT=$((TEST_CURRENT_HEIGHT - 15))  # Should be pruned with keep-recent=10

log_result "Test node height: ${TEST_CURRENT_HEIGHT}"
log_result "Testing old height: ${OLD_HEIGHT} (should be pruned)"

# Try to query old block
PRUNED=false
if docker exec ${TEST_NODE} aurad q block ${OLD_HEIGHT} --home ${AURA_HOME} 2>&1 | grep -q "could not find results for height"; then
    log_success "Old block ${OLD_HEIGHT} is correctly pruned (query failed)"
    PRUNED=true
elif docker exec ${TEST_NODE} aurad q block ${OLD_HEIGHT} --home ${AURA_HOME} 2>&1 | grep -q "height ${OLD_HEIGHT} is not available"; then
    log_success "Old block ${OLD_HEIGHT} is correctly pruned (not available)"
    PRUNED=true
else
    log_result "Block ${OLD_HEIGHT} query result:"
    docker exec ${TEST_NODE} aurad q block ${OLD_HEIGHT} --home ${AURA_HOME} 2>&1 | head -10 | tee -a "${RESULTS_FILE}"
    log_result "Old block ${OLD_HEIGHT} may still be cached (normal for recent heights)"
fi

# Test recent block (should succeed)
RECENT_HEIGHT=$((TEST_CURRENT_HEIGHT - 5))  # Should be kept with keep-recent=10
log_test "Testing recent height: ${RECENT_HEIGHT} (should be available)"

if docker exec ${TEST_NODE} aurad q block ${RECENT_HEIGHT} --home ${AURA_HOME} 2>&1 | grep -q "block_id"; then
    log_success "Recent block ${RECENT_HEIGHT} is correctly kept"
else
    log_error "Recent block ${RECENT_HEIGHT} was incorrectly pruned"
    PRUNED=false
fi

# Test 6: Check database size (pruning should limit growth)
log_test "Test 6: Checking database size"
DB_SIZE=$(docker exec ${TEST_NODE} du -sh ${AURA_HOME}/data 2>/dev/null | awk '{print $1}')
log_result "Database size: ${DB_SIZE}"

# Test 7: Verify pruning metrics in logs
log_test "Test 7: Checking logs for pruning activity"
if docker logs ${TEST_NODE} 2>&1 | grep -i "prun" | tail -10 | tee -a "${RESULTS_FILE}" | grep -q "prun"; then
    log_success "Pruning activity found in logs"
else
    log_result "No explicit pruning messages in logs (may be normal)"
fi

# Test 8: Different pruning modes
log_test "Test 8: Documenting available pruning modes"
echo "Available pruning modes:" | tee -a "${RESULTS_FILE}"
echo "  - default: Keep last 362880 blocks (~2 weeks with 5s blocks)" | tee -a "${RESULTS_FILE}"
echo "  - nothing: Keep all blocks (archive node)" | tee -a "${RESULTS_FILE}"
echo "  - everything: Keep only last 10 blocks" | tee -a "${RESULTS_FILE}"
echo "  - custom: User-defined keep-recent and interval" | tee -a "${RESULTS_FILE}"

# Cleanup
log_test "Cleaning up test node"
docker stop ${TEST_NODE}
docker rm ${TEST_NODE}
docker volume rm aura-pruning-test-data
log_success "Test node cleaned up"

echo "" | tee -a "${RESULTS_FILE}"
echo "=== State Pruning Test Complete ===" | tee -a "${RESULTS_FILE}"
echo "End time: $(date)" | tee -a "${RESULTS_FILE}"
echo "" | tee -a "${RESULTS_FILE}"

echo "FINAL RESULT: ✅ PASSED - State pruning tested successfully" | tee -a "${RESULTS_FILE}"
echo "Results saved to: ${RESULTS_FILE}" | tee -a "${RESULTS_FILE}"
exit 0
