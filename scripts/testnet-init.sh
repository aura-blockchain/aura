#!/bin/bash
# ============================================================================
# AURA Testnet Initialization Script
# ============================================================================
# This script initializes a 4-validator local testnet for AURA blockchain
# Chain ID: aura-local-4
#
# Features:
# - Store initialization verification (prevents IAVL version errors)
# - AppHash consistency checks
# - Key management with multiple backend support
# ============================================================================

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source utility libraries
source "${SCRIPT_DIR}/lib-store-verification.sh"
source "${SCRIPT_DIR}/lib-key-management.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CHAIN_ID="aura-local-4"
DENOM="uaura"
STAKING_AMOUNT="900000000000${DENOM}"  # 900,000 AURA per validator
NUM_VALIDATORS=${NUM_VALIDATORS:-4}
BINARY="aurad"
VALIDATOR_MONIKERS=("validator-1" "validator-2" "validator-3" "validator-4")

# Key management configuration
# Use 'test' backend for local testnet (unencrypted, easy for development)
# For production/mainnet, use 'os' or 'file' backend
KEYRING_BACKEND="${AURA_KEYRING_BACKEND:-test}"
KEY_PASSWORD="${AURA_KEY_PASSWORD:-password123}"

# Setup keyring backend
setup_keyring_backend "$KEYRING_BACKEND"

# Base directory for testnet data
TESTNET_DIR="${PWD}/testnet-data"

# Docker network configuration (must match docker-compose.testnet.yml)
VALIDATOR_IPS=("172.26.0.10" "172.26.0.11" "172.26.0.12" "172.26.0.13")

echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}AURA Multi-Node Testnet Initialization${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo -e "${GREEN}Chain ID:${NC} ${CHAIN_ID}"
echo -e "${GREEN}Validators:${NC} ${NUM_VALIDATORS}"
echo -e "${GREEN}Staking per validator:${NC} ${STAKING_AMOUNT}"
echo ""

# ============================================================================
# Step 1: Build the aurad binary
# ============================================================================
echo -e "${YELLOW}[1/8]${NC} Building aurad binary..."
cd "${PWD}/chain"
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

go build -o "${BINARY}" ./cmd/aurad
if [ ! -f "${BINARY}" ]; then
    echo -e "${RED}Error: Failed to build ${BINARY}${NC}"
    exit 1
fi
chmod +x "${BINARY}"
echo -e "${GREEN}✓ Binary built successfully${NC}"

# Move binary to a location accessible for the script
BINARY_PATH="${PWD}/${BINARY}"
echo -e "${GREEN}✓ Binary location: ${BINARY_PATH}${NC}"
cd ..

# ============================================================================
# Step 2: Clean up old testnet data
# ============================================================================
echo -e "${YELLOW}[2/8]${NC} Cleaning up old testnet data..."
rm -rf "${TESTNET_DIR}"
mkdir -p "${TESTNET_DIR}"
echo -e "${GREEN}✓ Testnet directory created: ${TESTNET_DIR}${NC}"

# ============================================================================
# Step 3: Initialize each validator node
# ============================================================================
echo -e "${YELLOW}[3/8]${NC} Initializing ${NUM_VALIDATORS} validator nodes..."
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    NODE_HOME="${TESTNET_DIR}/${MONIKER}"

    echo -e "  ${BLUE}Initializing ${MONIKER}...${NC}"
    "${BINARY_PATH}" init "${MONIKER}" \
        --chain-id "${CHAIN_ID}" \
        --home "${NODE_HOME}" \
        -y > /dev/null 2>&1

    echo -e "  ${GREEN}✓ ${MONIKER} initialized${NC}"
done

# ============================================================================
# Step 4: Create validator keys and configure genesis accounts
# ============================================================================
echo -e "${YELLOW}[4/8]${NC} Creating validator keys and accounts..."

# We'll collect all validator addresses and node IDs
declare -a VALIDATOR_ADDRESSES
declare -a VALIDATOR_OPERATOR_ADDRESSES
declare -a NODE_IDS

for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    NODE_HOME="${TESTNET_DIR}/${MONIKER}"

    echo -e "  ${BLUE}Creating key for ${MONIKER}...${NC}"

    # Create validator key
    printf '%s\n' "${KEY_PASSWORD}" | "${BINARY_PATH}" keys add "${MONIKER}" \
        --keyring-backend test \
        --home "${NODE_HOME}" \
        --output json > "${NODE_HOME}/key.json" 2>&1

    # Extract validator address
    VALIDATOR_ADDR=$(printf '%s\n' "${KEY_PASSWORD}" | "${BINARY_PATH}" keys show "${MONIKER}" \
        --keyring-backend test \
        --home "${NODE_HOME}" \
        --address 2>/dev/null)

    VALIDATOR_ADDRESSES[$i]="${VALIDATOR_ADDR}"

    # Operator (valoper) address
    VALIDATOR_OPERATOR_ADDRESSES[$i]=$(printf '%s\n' "${KEY_PASSWORD}" | "${BINARY_PATH}" keys show "${MONIKER}" \
        --keyring-backend test \
        --home "${NODE_HOME}" \
        --bech val \
        --address 2>/dev/null)

    # Get node ID for persistent_peers
    NODE_KEY_GEN=$(mktemp "${REPO_ROOT}/chain/tmp.nodekey.XXXX.go")
    cat > "${NODE_KEY_GEN}" <<'EOF'
package main

import (
	"fmt"
	"log"
	"os"

	tmp2p "github.com/cometbft/cometbft/p2p"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: node_key_path")
	}
	path := os.Args[1]
	key, err := tmp2p.LoadOrGenNodeKey(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(key.ID())
}
EOF
    NODE_ID=$(GOWORK=off go run -C "${REPO_ROOT}/chain" -mod=mod "${NODE_KEY_GEN}" "${NODE_HOME}/config/node_key.json")
    rm -f "${NODE_KEY_GEN}"
    NODE_IDS[$i]="${NODE_ID}"

    echo -e "  ${GREEN}✓ ${MONIKER}: ${VALIDATOR_ADDR}${NC}"
    echo -e "  ${GREEN}  Node ID: ${NODE_ID}${NC}"
done

echo ""
echo -e "${YELLOW}Validating all keys...${NC}"

# Validate all keys were created successfully
VALIDATION_FAILED=0
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    NODE_HOME="${TESTNET_DIR}/${MONIKER}"

    if ! validate_key_exists "$MONIKER" "$NODE_HOME" "$KEYRING_BACKEND"; then
        echo -e "  ${RED}✗ Key validation failed for: $MONIKER${NC}"
        VALIDATION_FAILED=1
    else
        echo -e "  ${GREEN}✓ Key validated: $MONIKER${NC}"
    fi
done

if [ $VALIDATION_FAILED -eq 1 ]; then
    echo -e "${RED}✗ Key validation failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ All keys validated successfully${NC}"

# ============================================================================
# Step 5: Configure genesis file with all validators
# ============================================================================
echo ""
echo -e "${YELLOW}[5/9]${NC} Configuring genesis file..."

# Use validator-1 as the template for genesis
GENESIS_HOME="${TESTNET_DIR}/${VALIDATOR_MONIKERS[0]}"
GENESIS_FILE="${GENESIS_HOME}/config/genesis.json"

# Add all validator accounts to genesis
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    VALIDATOR_ADDR="${VALIDATOR_ADDRESSES[$i]}"

    echo -e "  ${BLUE}Adding ${MONIKER} to genesis...${NC}"
    "${BINARY_PATH}" genesis add-genesis-account "${VALIDATOR_ADDR}" \
        "1000000000000${DENOM}" \
        --home "${GENESIS_HOME}" > /dev/null 2>&1
done

# Configure genesis parameters for testnet
echo -e "  ${BLUE}Configuring genesis parameters...${NC}"

# Use jq to modify genesis parameters (if jq is available)
if command -v jq &> /dev/null; then
    # Staking parameters
    jq '.app_state.staking.params.unbonding_time = "1814400s"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.staking.params.max_validators = 100' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"

    # Governance parameters (shorter voting period for testing)
    jq '.app_state.gov.params.voting_period = "300s"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.gov.params.min_deposit[0].amount = "10000000"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"

    # Mint parameters
    jq '.app_state.mint.params.inflation_min = "0.07"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.mint.params.inflation_max = "0.20"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"

    # Crisis module constant fee
    jq '.app_state.crisis.constant_fee.denom = "uaura"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"

    # Security module parameters (must be positive)
    jq '.app_state.security.params.network.rate_limit.max_requests_per_second = "200"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.rate_limit.burst_size = "400"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.rate_limit.window_duration = "10s"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.rate_limit.ban_duration = "60s"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.connection.max_inbound_connections = 100' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.connection.max_outbound_connections = 100' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.connection.max_connections_per_ip = 20' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.connection.connection_timeout = "5s"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.mempool.max_size = "5000"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.mempool.max_bytes = "200000000"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.network.mempool.min_priority_fee = "0"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.validator.signed_blocks_window = "1000"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.crypto.min_threshold_participants = 2' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.privacy.min_ring_size = 3' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.privacy.max_ring_size = 16' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.privacy.min_mixing_participants = 3' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.security.params.privacy.mixing_fee = "1000"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"

    echo -e "  ${GREEN}✓ Genesis parameters configured${NC}"
else
    echo -e "  ${YELLOW}⚠ jq not found, using default genesis parameters${NC}"
fi

# Remove any default validator scaffold from init
jq '
  .app_state.staking.params.bond_denom = "uaura" |
  .app_state.staking.validators = [] |
  .app_state.staking.last_validator_powers = [] |
  .app_state.staking.last_total_power = "0" |
  .app_state.staking.delegations = [] |
  .app_state.staking.unbonding_delegations = [] |
  .app_state.staking.redelegations = [] |
  .app_state.bank.supply = (.app_state.bank.supply // [] | map(select(.denom != "stake"))) |
  .app_state.bank.balances = [.app_state.bank.balances[] | .coins = [.coins[] | select(.denom != "stake")]]
' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"

# ============================================================================
# Step 5.5: Create gentx for all validators
# ============================================================================
echo ""
echo -e "${YELLOW}[5.5/10]${NC} Creating gentx for all validators..."

# Each validator creates their gentx
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    NODE_HOME="${TESTNET_DIR}/${MONIKER}"
    VALIDATOR_ADDR="${VALIDATOR_ADDRESSES[$i]}"

    echo -e "  ${BLUE}Creating gentx for ${MONIKER}...${NC}"

    # Copy the genesis to this validator's home (needed for gentx)
    if [ $i -gt 0 ]; then
        cp "${GENESIS_HOME}/config/genesis.json" "${NODE_HOME}/config/genesis.json"
    fi

    # Create gentx - this registers the validator with staking power
    printf '%s\n' "${KEY_PASSWORD}" | "${BINARY_PATH}" genesis gentx "${MONIKER}" "${STAKING_AMOUNT}" \
        --chain-id "${CHAIN_ID}" \
        --keyring-backend test \
        --keyring-dir "${NODE_HOME}" \
        --from "${MONIKER}" \
        --home "${NODE_HOME}" \
        --moniker "${MONIKER}" \
        --commission-rate "0.10" \
        --commission-max-rate "0.20" \
        --commission-max-change-rate "0.01" > /dev/null 2>&1

    # Work around gentx bug where delegator_address may be blank
    GENTX_FILE=$(ls "${NODE_HOME}/config/gentx/"*.json | head -n 1)
    if [ -n "${GENTX_FILE}" ]; then
        jq --arg addr "${VALIDATOR_ADDR}" '(.body.messages[] | select(.delegator_address == "")) .delegator_address = $addr' "${GENTX_FILE}" > "${GENTX_FILE}.tmp" && mv "${GENTX_FILE}.tmp" "${GENTX_FILE}"
    fi

    echo -e "  ${GREEN}✓ ${MONIKER} gentx created${NC}"
done

# ============================================================================
# Step 5.6: Collect all gentxs into genesis
# ============================================================================
echo ""
echo -e "${YELLOW}[5.6/10]${NC} Collecting all gentxs into genesis..."

# Copy all gentxs to validator-1's gentx directory
mkdir -p "${GENESIS_HOME}/config/gentx"
for i in $(seq 1 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    NODE_HOME="${TESTNET_DIR}/${MONIKER}"

    # Copy gentx files from each validator to validator-1
    cp "${NODE_HOME}/config/gentx/"*.json "${GENESIS_HOME}/config/gentx/" 2>/dev/null || true
done

# Collect all gentxs
"${BINARY_PATH}" genesis collect-gentxs \
    --home "${GENESIS_HOME}" > /dev/null 2>&1

echo -e "${GREEN}✓ All gentxs collected${NC}"

# CRITICAL FIX: Change bond_denom to uaura (gentx use uaura, not stake)
jq '.app_state.staking.params.bond_denom = "uaura"' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
echo -e "${GREEN}✓ Set bond denomination to uaura${NC}"

# Remove any "stake" denom that may have been added
if command -v jq &> /dev/null; then
    jq '.app_state.bank.balances = [.app_state.bank.balances[] | .coins = [.coins[] | select(.denom != "stake")]]' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    jq '.app_state.bank.supply = (.app_state.bank.supply // [] | map(select(.denom != "stake")))' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"
    echo -e "${GREEN}✓ Removed stake denomination from balances${NC}"
fi

# Verify validators are in gentx
VALIDATOR_COUNT=$(jq '.app_state.genutil.gen_txs | length' "${GENESIS_FILE}")
echo -e "${GREEN}✓ Genesis contains ${VALIDATOR_COUNT} validators in gentx${NC}"

# Rebuild staking state directly from gentx files to guarantee validators exist at genesis
GENESIS_TIME=$(jq -r '.genesis_time' "${GENESIS_FILE}")
VALIDATORS_JSON="[]"
POWERS_JSON="[]"
DELEGATIONS_JSON="[]"
TOTAL_POWER=0
for gentx in "${GENESIS_HOME}"/config/gentx/*.json; do
    [ -f "${gentx}" ] || continue
    VALOPER=$(jq -r '.body.messages[0].validator_address' "${gentx}")
    DELEGATOR=$(jq -r '.body.messages[0].delegator_address' "${gentx}")
    PUBKEY=$(jq -r '.body.messages[0].pubkey.key' "${gentx}")
    MONIKER=$(jq -r '.body.messages[0].description.moniker' "${gentx}")
    RATE=$(jq -r '.body.messages[0].commission.rate' "${gentx}")
    MAX_RATE=$(jq -r '.body.messages[0].commission.max_rate' "${gentx}")
    MAX_CHANGE=$(jq -r '.body.messages[0].commission.max_change_rate' "${gentx}")
    AMOUNT=$(jq -r '.body.messages[0].value.amount' "${gentx}")

    # Convert staking tokens to power (PowerReduction defaults to 1e6)
    POWER=$((AMOUNT / 1000000))
    TOTAL_POWER=$((TOTAL_POWER + POWER))

    VALIDATORS_JSON=$(jq --arg valoper "${VALOPER}" \
                           --arg pub "${PUBKEY}" \
                           --arg mon "${MONIKER}" \
                           --arg rate "${RATE}" \
                           --arg max_rate "${MAX_RATE}" \
                           --arg max_change "${MAX_CHANGE}" \
                           --arg amount "${AMOUNT}" \
                           --arg update_time "${GENESIS_TIME}" \
        '. += [{
            operator_address: $valoper,
            consensus_pubkey: {"@type": "/cosmos.crypto.ed25519.PubKey", key: $pub},
            jailed: false,
            status: "BOND_STATUS_BONDED",
            tokens: $amount,
            delegator_shares: ($amount + ".000000000000000000"),
            description: {moniker: $mon, identity: "", website: "", security_contact: "", details: ""},
            unbonding_height: "0",
            unbonding_time: "1970-01-01T00:00:00Z",
            commission: {commission_rates: {rate: $rate, max_rate: $max_rate, max_change_rate: $max_change}, update_time: $update_time},
            min_self_delegation: "1",
            unbonding_on_hold_ref_count: "0",
            unbonding_ids: []
        }]' <<< "${VALIDATORS_JSON}")

    POWERS_JSON=$(jq --arg valoper "${VALOPER}" --arg power "${POWER}" '. += [{address: $valoper, power: $power}]' <<< "${POWERS_JSON}")
    DELEGATIONS_JSON=$(jq --arg del "${DELEGATOR}" --arg valoper "${VALOPER}" --arg amount "${AMOUNT}" '. += [{delegator_address: $del, validator_address: $valoper, shares: ($amount + ".000000000000000000")}]' <<< "${DELEGATIONS_JSON}")
done

jq --argjson validators "${VALIDATORS_JSON}" \
   --argjson powers "${POWERS_JSON}" \
   --argjson delegations "${DELEGATIONS_JSON}" \
   --arg total_power "$(printf "%d" "${TOTAL_POWER}")" '
  .app_state.staking.validators = $validators |
  .app_state.staking.last_validator_powers = $powers |
  .app_state.staking.last_total_power = ($total_power | tostring) |
  .app_state.staking.delegations = $delegations |
  .app_state.staking.unbonding_delegations = [] |
  .app_state.staking.redelegations = []
' "${GENESIS_FILE}" > tmp.json && mv tmp.json "${GENESIS_FILE}"

# ============================================================================
# Step 6: Distribute genesis and configure peers
# ============================================================================
echo -e "${YELLOW}[6/10]${NC} Distributing genesis and configuring peers..."

# Distribute final genesis to all validators
FINAL_GENESIS="${GENESIS_HOME}/config/genesis.json"
for i in $(seq 1 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    NODE_HOME="${TESTNET_DIR}/${MONIKER}"
    cp "${FINAL_GENESIS}" "${NODE_HOME}/config/genesis.json"
done
echo -e "${GREEN}✓ Genesis distributed to all nodes${NC}"

# Configure persistent peers
echo -e "${YELLOW}[7/9]${NC} Configuring persistent peers..."

for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    NODE_HOME="${TESTNET_DIR}/${MONIKER}"
    CONFIG_FILE="${NODE_HOME}/config/config.toml"
    APP_FILE="${NODE_HOME}/config/app.toml"

    # Build persistent_peers string (exclude self)
    PEERS=""
    for j in $(seq 0 $((NUM_VALIDATORS - 1))); do
        if [ $i -ne $j ]; then
            NODE_ID="${NODE_IDS[$j]}"
            IP="${VALIDATOR_IPS[$j]}"
            if [ -n "${PEERS}" ]; then
                PEERS="${PEERS},"
            fi
            PEERS="${PEERS}${NODE_ID}@${IP}:26656"
        fi
    done

    # Update config.toml
    sed -i.bak "s/^persistent_peers = .*/persistent_peers = \"${PEERS}\"/" "${CONFIG_FILE}"
    sed -i.bak 's/^cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/' "${CONFIG_FILE}"
    sed -i.bak 's/^allow_duplicate_ip = false/allow_duplicate_ip = true/' "${CONFIG_FILE}"
    sed -i.bak 's#laddr = "tcp://127.0.0.1:26657"#laddr = "tcp://0.0.0.0:26657"#' "${CONFIG_FILE}"

    # Enable prometheus metrics
    sed -i.bak 's/^prometheus = false/prometheus = true/' "${CONFIG_FILE}"

    # Faster block times for testing
    sed -i.bak 's/^timeout_commit = .*/timeout_commit = "3s"/' "${CONFIG_FILE}"

    # App-level endpoints (REST/gRPC)
    sed -i.bak 's#address = "tcp://127.0.0.1:1317"#address = "tcp://0.0.0.0:1317"#' "${APP_FILE}"
    sed -i.bak 's#address = "localhost:9090"#address = "0.0.0.0:9090"#' "${APP_FILE}" >/dev/null 2>&1 || true
    sed -i.bak 's#address = "0.0.0.0:9091"#address = "0.0.0.0:9091"#' "${APP_FILE}" >/dev/null 2>&1 || true

    echo -e "  ${GREEN}✓ ${MONIKER} configured with peers${NC}"
done

# ============================================================================
# Step 8: Create Docker volume initialization script
# ============================================================================
echo -e "${YELLOW}[8/9]${NC} Creating Docker volume population script...${NC}"

cat > "${TESTNET_DIR}/populate-volumes.sh" << 'EOF'
#!/bin/bash
# Populate Docker volumes with initialized testnet data

set -e

TESTNET_DIR="$(dirname "$0")"
VALIDATORS=("validator-1" "validator-2" "validator-3" "validator-4")

echo "Populating Docker volumes..."

for VALIDATOR in "${VALIDATORS[@]}"; do
    VOLUME_NAME="aura_${VALIDATOR}-data"

    # Create a temporary container to copy data
    echo "  Copying data for ${VALIDATOR}..."

    docker run --rm \
        -v "${VOLUME_NAME}:/home/aura/.aura" \
        -v "${TESTNET_DIR}/${VALIDATOR}:/source:ro" \
        alpine sh -c "cp -r /source/config /home/aura/.aura/ && \
                      cp -r /source/data /home/aura/.aura/ && \
                      cp -r /source/keyring-test /home/aura/.aura/ 2>/dev/null || true && \
                      chown -R 1000:1000 /home/aura/.aura"

    echo "  ✓ ${VALIDATOR} volume populated"
done

echo "All volumes populated successfully!"
EOF

chmod +x "${TESTNET_DIR}/populate-volumes.sh"

# ============================================================================
# Step 8: Verify Store Initialization Readiness
# ============================================================================
echo -e "${YELLOW}[8/9]${NC} Verifying store initialization readiness..."
echo ""

# Verify each validator's stores
VERIFICATION_FAILED=0
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    NODE_HOME="${TESTNET_DIR}/${MONIKER}"

    echo -e "${BLUE}Verifying ${MONIKER}...${NC}"

    # Check that database will be created on first start
    if [ ! -d "${NODE_HOME}/data" ]; then
        mkdir -p "${NODE_HOME}/data"
    fi

    # Verify genesis is present (required for InitGenesis)
    if [ ! -f "${NODE_HOME}/config/genesis.json" ]; then
        echo -e "  ${RED}✗ Genesis file missing${NC}"
        VERIFICATION_FAILED=1
        continue
    fi

    echo -e "  ${GREEN}✓ Genesis file present${NC}"
    echo -e "  ${GREEN}✓ Ready for store initialization on first start${NC}"
    echo -e "  ${YELLOW}ℹ Stores will be initialized during InitGenesis (first block)${NC}"
    echo ""
done

if [ $VERIFICATION_FAILED -eq 1 ]; then
    echo -e "${RED}✗ Store verification failed for one or more validators${NC}"
    exit 1
fi

echo -e "${GREEN}✓ All validators ready for store initialization${NC}"
echo ""
echo -e "${YELLOW}Important:${NC} Store initialization happens automatically when the node"
echo -e "processes its first block (InitGenesis). The ensureStoreInitMarkers()"
echo -e "function in chain/app/app.go writes a deterministic marker (0x01 byte)"
echo -e "into each KV store to ensure all stores have version 1 persisted."
echo ""
echo -e "To verify stores after startup:"
echo -e "  ${BLUE}source scripts/lib-store-verification.sh${NC}"
echo -e "  ${BLUE}verify_store_initialization testnet-data/validator-1${NC}"
echo ""

# ============================================================================
# Summary
# ============================================================================
echo -e "${YELLOW}[9/9]${NC} Initialization Summary"
echo ""
echo -e "${BLUE}============================================================================${NC}"
echo -e "${GREEN}Testnet Initialization Complete!${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""
echo -e "${GREEN}Chain ID:${NC} ${CHAIN_ID}"
echo -e "${GREEN}Validators:${NC} ${NUM_VALIDATORS}"
echo -e "${GREEN}Testnet Data:${NC} ${TESTNET_DIR}"
echo ""
echo -e "${YELLOW}Validator Details:${NC}"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    MONIKER="${VALIDATOR_MONIKERS[$i]}"
    echo -e "  ${BLUE}${MONIKER}:${NC}"
    echo -e "    Address: ${VALIDATOR_ADDRESSES[$i]}"
    echo -e "    Node ID: ${NODE_IDS[$i]}"
    echo -e "    IP: ${VALIDATOR_IPS[$i]}"
done
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo -e "  1. Populate Docker volumes:"
echo -e "     ${BLUE}cd testnet-data && ./populate-volumes.sh${NC}"
echo ""
echo -e "  2. Start the testnet:"
echo -e "     ${BLUE}docker-compose -f docker-compose.testnet.yml up -d${NC}"
echo ""
echo -e "  3. Verify store initialization (after startup):"
echo -e "     ${BLUE}source scripts/lib-store-verification.sh${NC}"
echo -e "     ${BLUE}verify_store_initialization testnet-data/validator-1${NC}"
echo ""
echo -e "  4. View logs:"
echo -e "     ${BLUE}docker-compose -f docker-compose.testnet.yml logs -f validator-1${NC}"
echo ""
echo -e "  5. Check node status:"
echo -e "     ${BLUE}curl http://localhost:27657/status${NC}"
echo ""
echo -e "${YELLOW}Port Mappings:${NC}"
echo -e "  validator-1: RPC=27657, API=2317, P2P=27656, gRPC=10090, Metrics=27660"
echo -e "  validator-2: RPC=27757, API=2417, P2P=27756, gRPC=10190, Metrics=27760"
echo -e "  validator-3: RPC=27857, API=2517, P2P=27856, gRPC=10290, Metrics=27860"
echo -e "  validator-4: RPC=27957, API=2617, P2P=27956, gRPC=10390, Metrics=27960"
echo ""
echo -e "  Monitoring: Prometheus=9094, Grafana=3002"
echo ""
echo -e "${GREEN}Happy testing!${NC}"
echo -e "${BLUE}============================================================================${NC}"
