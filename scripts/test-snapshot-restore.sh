#!/bin/bash
# State Snapshot & Restore Testing
# Tests snapshot creation, export, and restoration on a new node

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_FILE="${SCRIPT_DIR}/snapshot_restore_results.txt"
AURA_HOME="/home/aura/.aura"

echo "=== State Snapshot & Restore Testing ===" | tee "${RESULTS_FILE}"
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

# Get current block height
log_test "Getting current block height from validator-1"
CURRENT_HEIGHT=$(curl -s localhost:27657/status 2>&1 | jq -r '.result.sync_info.latest_block_height // "0"')
log_result "Current block height: ${CURRENT_HEIGHT}"

if [[ "${CURRENT_HEIGHT}" == "0" ]] || [[ -z "${CURRENT_HEIGHT}" ]]; then
    log_error "Failed to get block height"
    exit 1
fi

# Wait for a few blocks to ensure state updates
log_test "Waiting for 10 more blocks to ensure state changes"
TARGET_HEIGHT=$((CURRENT_HEIGHT + 10))
echo "Waiting for height ${TARGET_HEIGHT}..." | tee -a "${RESULTS_FILE}"

TIMEOUT=120
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
        exit 1
    fi
    sleep 2
done

# Check if snapshots are enabled
log_test "Checking snapshot configuration on validator-1"
SNAPSHOT_INTERVAL=$(docker exec aura-validator-1 grep -A10 "^\[state-sync\]" ${AURA_HOME}/config/app.toml | grep "snapshot-interval" | awk '{print $3}')
log_result "Snapshot interval: ${SNAPSHOT_INTERVAL:-not set}"

if [[ -z "${SNAPSHOT_INTERVAL}" ]] || [[ "${SNAPSHOT_INTERVAL}" == "0" ]]; then
    log_result "Snapshots are disabled. Enabling snapshots..."
    docker exec aura-validator-1 sed -i 's/snapshot-interval = 0/snapshot-interval = 100/' ${AURA_HOME}/config/app.toml
    log_result "Snapshot interval set to 100. Restarting validator-1..."
    docker restart aura-validator-1
    sleep 15
    log_success "Validator restarted with snapshots enabled"
fi

# Wait for snapshot to be created
log_test "Waiting for snapshot to be created at next interval"
CURRENT=$(curl -s localhost:27657/status 2>&1 | jq -r '.result.sync_info.latest_block_height // "0"')
SNAPSHOT_HEIGHT=$(((CURRENT / 100 + 1) * 100))
echo "Waiting for snapshot at height ${SNAPSHOT_HEIGHT}..." | tee -a "${RESULTS_FILE}"

TIMEOUT=300
START_TIME=$(date +%s)
while true; do
    CURRENT=$(curl -s localhost:27657/status 2>&1 | jq -r '.result.sync_info.latest_block_height // "0"')
    if [[ "${CURRENT}" -ge "${SNAPSHOT_HEIGHT}" ]]; then
        log_success "Reached snapshot height ${SNAPSHOT_HEIGHT}"
        sleep 10  # Wait for snapshot to be written
        break
    fi

    ELAPSED=$(($(date +%s) - START_TIME))
    if [[ ${ELAPSED} -gt ${TIMEOUT} ]]; then
        log_error "Timeout waiting for snapshot height"
        exit 1
    fi
    sleep 3
done

# List available snapshots
log_test "Listing available snapshots"
docker exec aura-validator-1 ls -lh ${AURA_HOME}/data/snapshots/ 2>&1 | tee -a "${RESULTS_FILE}" || log_result "No snapshots directory yet"

# Create a test node that will use state sync
log_test "Creating a new test node (aura-snapshot-test) to restore from snapshot"

TEST_NODE_NAME="aura-snapshot-test"

# Stop and remove if exists
docker stop ${TEST_NODE_NAME} 2>/dev/null || true
docker rm ${TEST_NODE_NAME} 2>/dev/null || true
docker volume rm aura-snapshot-test-data 2>/dev/null || true

# Create test node
log_test "Creating test node container"
docker run -d \
    --name ${TEST_NODE_NAME} \
    --network aura_aura-testnet \
    -v aura-snapshot-test-data:/home/aura/.aura \
    -p 27658:26657 \
    -p 27659:26656 \
    --entrypoint /bin/sh \
    aurad:latest \
    -c "sleep infinity"

log_success "Test node container created"

# Initialize test node
log_test "Initializing test node"
docker exec ${TEST_NODE_NAME} aurad init snapshot-test --chain-id aura-mvp-1 --home ${AURA_HOME} 2>&1 | tee -a "${RESULTS_FILE}"

# Copy genesis from validator-1
log_test "Copying genesis.json from validator-1"
docker exec aura-validator-1 cat ${AURA_HOME}/config/genesis.json > /tmp/genesis.json
docker cp /tmp/genesis.json ${TEST_NODE_NAME}:${AURA_HOME}/config/genesis.json
log_success "Genesis copied"

# Configure state sync
log_test "Configuring state sync on test node"

# Get trusted height and hash
TRUSTED_HEIGHT=${SNAPSHOT_HEIGHT}
TRUSTED_HASH=$(docker exec aura-validator-1 aurad q block ${TRUSTED_HEIGHT} --home ${AURA_HOME} 2>&1 | jq -r '.block_id.hash // empty')

if [[ -z "${TRUSTED_HASH}" ]]; then
    log_error "Failed to get trusted hash for height ${TRUSTED_HEIGHT}"
    docker stop ${TEST_NODE_NAME}
    docker rm ${TEST_NODE_NAME}
    exit 1
fi

log_result "Trusted height: ${TRUSTED_HEIGHT}"
log_result "Trusted hash: ${TRUSTED_HASH}"

# Configure state sync in config.toml
docker exec ${TEST_NODE_NAME} bash -c "
sed -i '/^\[statesync\]/,/^\[/ {
    s/^enable = .*/enable = true/
    s|^rpc_servers = .*|rpc_servers = \"aura-validator-1:26657,aura-validator-2:26657\"|
    s/^trust_height = .*/trust_height = ${TRUSTED_HEIGHT}/
    s/^trust_hash = .*/trust_hash = \"${TRUSTED_HASH}\"/
    s/^trust_period = .*/trust_period = \"168h0m0s\"/
}' ${AURA_HOME}/config/config.toml
"

log_success "State sync configured"

# Configure persistent peers
log_test "Configuring persistent peers"
VAL1_ID=$(docker exec aura-validator-1 aurad tendermint show-node-id 2>&1)
VAL2_ID=$(docker exec aura-validator-2 aurad tendermint show-node-id 2>&1)
PEERS="${VAL1_ID}@aura-validator-1:26656,${VAL2_ID}@aura-validator-2:26656"

docker exec ${TEST_NODE_NAME} sed -i "s/^persistent_peers = .*/persistent_peers = \"${PEERS}\"/" ${AURA_HOME}/config/config.toml
log_success "Persistent peers configured: ${PEERS}"

# Start test node and monitor state sync
log_test "Starting test node with state sync"
docker exec -d ${TEST_NODE_NAME} aurad start --home ${AURA_HOME}

log_result "Waiting for state sync to complete (this may take 30-60 seconds)..."

# Monitor state sync progress
TIMEOUT=120
START_TIME=$(date +%s)
SYNCED=false

while true; do
    sleep 5

    # Check if node is catching up
    STATUS=$(docker exec ${TEST_NODE_NAME} aurad status 2>&1 || echo "{}")
    CATCHING_UP=$(echo "${STATUS}" | jq -r '.sync_info.catching_up // "true"')
    SYNCED_HEIGHT=$(echo "${STATUS}" | jq -r '.sync_info.latest_block_height // "0"')

    log_result "Test node status - Height: ${SYNCED_HEIGHT}, Catching up: ${CATCHING_UP}"

    if [[ "${CATCHING_UP}" == "false" ]] && [[ "${SYNCED_HEIGHT}" -gt "0" ]]; then
        SYNCED=true
        log_success "State sync completed! Node synced to height ${SYNCED_HEIGHT}"
        break
    fi

    ELAPSED=$(($(date +%s) - START_TIME))
    if [[ ${ELAPSED} -gt ${TIMEOUT} ]]; then
        log_error "Timeout waiting for state sync"
        docker logs ${TEST_NODE_NAME} --tail 100 | tee -a "${RESULTS_FILE}"
        break
    fi
done

# Verify synced state
if [[ "${SYNCED}" == "true" ]]; then
    log_test "Verifying synced state integrity"

    # Compare block hash at synced height
    VAL1_BLOCK=$(docker exec aura-validator-1 aurad q block ${SYNCED_HEIGHT} --home ${AURA_HOME} 2>&1 | jq -r '.block_id.hash // "N/A"')
    TEST_BLOCK=$(docker exec ${TEST_NODE_NAME} aurad q block ${SYNCED_HEIGHT} --home ${AURA_HOME} 2>&1 | jq -r '.block_id.hash // "N/A"')

    log_result "Validator-1 block hash at ${SYNCED_HEIGHT}: ${VAL1_BLOCK}"
    log_result "Test node block hash at ${SYNCED_HEIGHT}: ${TEST_BLOCK}"

    if [[ "${VAL1_BLOCK}" == "${TEST_BLOCK}" ]] && [[ "${VAL1_BLOCK}" != "N/A" ]]; then
        log_success "Block hashes match - state sync verified!"
    else
        log_error "Block hashes do not match - state sync may be incomplete"
        SYNCED=false
    fi

    # Query some state to verify functionality
    log_test "Querying bank balances to verify state access"
    docker exec ${TEST_NODE_NAME} aurad q bank total --output json --home ${AURA_HOME} 2>&1 | jq '.' | head -20 | tee -a "${RESULTS_FILE}"

    log_test "Querying staking validators"
    VALIDATOR_COUNT=$(docker exec ${TEST_NODE_NAME} aurad q staking validators --output json --home ${AURA_HOME} 2>&1 | jq '.validators | length')
    log_result "Validators found: ${VALIDATOR_COUNT}"

    log_success "State queries successful on synced node"
else
    log_error "State sync did not complete within timeout"
fi

# Cleanup test node
log_test "Cleaning up test node"
docker stop ${TEST_NODE_NAME}
docker rm ${TEST_NODE_NAME}
docker volume rm aura-snapshot-test-data 2>/dev/null || true
log_success "Test node cleaned up"

echo "" | tee -a "${RESULTS_FILE}"
echo "=== Snapshot & Restore Test Complete ===" | tee -a "${RESULTS_FILE}"
echo "End time: $(date)" | tee -a "${RESULTS_FILE}"
echo "" | tee -a "${RESULTS_FILE}"

if [[ "${SYNCED}" == "true" ]]; then
    echo "FINAL RESULT: ✅ PASSED - State snapshot and restore working correctly" | tee -a "${RESULTS_FILE}"
    echo "Results saved to: ${RESULTS_FILE}" | tee -a "${RESULTS_FILE}"
    exit 0
else
    echo "FINAL RESULT: ❌ FAILED - State sync did not complete" | tee -a "${RESULTS_FILE}"
    echo "Results saved to: ${RESULTS_FILE}" | tee -a "${RESULTS_FILE}"
    exit 1
fi
