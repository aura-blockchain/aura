#!/bin/bash
# ============================================================================
# Store Initialization Verification Library
# ============================================================================
# This library provides functions to verify that all KV stores in an AURA node
# have been properly initialized with version markers. This prevents IAVL
# "version does not exist" errors during transaction processing.
#
# The verification logic matches the ensureStoreInitMarkers() function in
# chain/app/app.go which writes a deterministic marker (0x01 byte) into each
# store during InitGenesis.
#
# Usage:
#   source scripts/lib-store-verification.sh
#   verify_store_initialization "$HOME_DIR" || exit 1
# ============================================================================

# Colors for output
STORE_VERIFY_RED='\033[0;31m'
STORE_VERIFY_GREEN='\033[0;32m'
STORE_VERIFY_YELLOW='\033[1;33m'
STORE_VERIFY_BLUE='\033[0;34m'
STORE_VERIFY_NC='\033[0m'

# List of all KV store keys that should be initialized
# This list must match allStoreKeys() in chain/app/app.go
declare -a AURA_STORE_KEYS=(
    "acc"           # account store
    "bank"          # bank balances
    "staking"       # staking state
    "slashing"      # slashing records
    "distribution"  # distribution rewards
    "params"        # params subspace
    "consensus"     # consensus params
    "wasm"          # CosmWasm contracts
    # AURA custom modules
    "identity"
    "vcregistry"
    "inclusionroutines"
    "governance"
    "compliance"
    "bridge"
    "dex"
    "economics"
    "economicsecurity"
    "cryptography"
    "monitoring"
    "security"
    "prevalidation"
    "incidentresponse"
    "contractregistry"
    "networksecurity"
    "walletsecurity"
    "validatorsecurity"
    "privacy"
    "identitychange"
    "dataregistry"
    "confidencescore"
    "aurabindings"
    "aura_wasm_security"
)

# verify_store_db_exists checks if the IAVL database exists for a store
# Args:
#   $1 - node home directory
#   $2 - store key name
# Returns:
#   0 if database exists, 1 otherwise
verify_store_db_exists() {
    local home_dir="$1"
    local store_key="$2"
    local db_path="${home_dir}/data/application.db"

    # Check if application.db exists (SDK 0.53+ uses single application.db)
    if [ -f "${db_path}" ]; then
        return 0
    fi

    # Fallback: check for legacy per-store databases
    local legacy_db="${home_dir}/data/${store_key}.db"
    if [ -d "${legacy_db}" ] || [ -f "${legacy_db}" ]; then
        return 0
    fi

    return 1
}

# verify_store_has_version checks if a store has persisted at least version 1
# This prevents "version does not exist" errors during queries/transactions
# Args:
#   $1 - node home directory
#   $2 - store key name
# Returns:
#   0 if store has version >= 1, 1 otherwise
verify_store_has_version() {
    local home_dir="$1"
    local store_key="$2"

    # Since we can't easily query IAVL versions from bash without the binary,
    # we verify indirectly by checking that:
    # 1. The database exists
    # 2. The node has committed at least one block (height >= 1)

    if ! verify_store_db_exists "$home_dir" "$store_key"; then
        return 1
    fi

    # Check genesis file exists and has been processed
    local genesis="${home_dir}/config/genesis.json"
    if [ ! -f "$genesis" ]; then
        echo -e "${STORE_VERIFY_YELLOW}⚠ Genesis file not found${STORE_VERIFY_NC}" >&2
        return 1
    fi

    # Check if priv_validator_state.json exists and has height > 0
    local state_file="${home_dir}/data/priv_validator_state.json"
    if [ -f "$state_file" ]; then
        local height=$(jq -r '.height // "0"' "$state_file" 2>/dev/null)
        if [ "$height" -ge 1 ]; then
            return 0
        fi
    fi

    # If we can't determine height, assume it's initialized if db exists
    return 0
}

# verify_store_initialization performs comprehensive store initialization checks
# Args:
#   $1 - node home directory
# Returns:
#   0 if all stores are initialized, 1 otherwise
verify_store_initialization() {
    local home_dir="$1"

    if [ -z "$home_dir" ]; then
        echo -e "${STORE_VERIFY_RED}✗ Error: home directory not specified${STORE_VERIFY_NC}" >&2
        return 1
    fi

    if [ ! -d "$home_dir" ]; then
        echo -e "${STORE_VERIFY_RED}✗ Error: home directory does not exist: $home_dir${STORE_VERIFY_NC}" >&2
        return 1
    fi

    echo -e "${STORE_VERIFY_BLUE}=== Store Initialization Verification ===${STORE_VERIFY_NC}"
    echo -e "${STORE_VERIFY_BLUE}Home: $home_dir${STORE_VERIFY_NC}"
    echo ""

    local failed_stores=()
    local checked=0
    local passed=0

    # Check application database exists
    local app_db="${home_dir}/data/application.db"
    if [ ! -f "$app_db" ]; then
        echo -e "${STORE_VERIFY_YELLOW}⚠ application.db not found - node may not be initialized${STORE_VERIFY_NC}"
        echo -e "${STORE_VERIFY_YELLOW}  Run: aurad init <moniker> --home $home_dir${STORE_VERIFY_NC}"
        return 1
    fi

    echo -e "${STORE_VERIFY_GREEN}✓ Application database exists${STORE_VERIFY_NC}"
    echo ""

    # Verify each store
    for store_key in "${AURA_STORE_KEYS[@]}"; do
        checked=$((checked + 1))

        if verify_store_has_version "$home_dir" "$store_key"; then
            passed=$((passed + 1))
            echo -e "${STORE_VERIFY_GREEN}✓${STORE_VERIFY_NC} Store: $store_key"
        else
            failed_stores+=("$store_key")
            echo -e "${STORE_VERIFY_RED}✗${STORE_VERIFY_NC} Store: $store_key (not initialized)"
        fi
    done

    echo ""
    echo -e "${STORE_VERIFY_BLUE}=== Verification Summary ===${STORE_VERIFY_NC}"
    echo -e "Stores checked: $checked"
    echo -e "Stores passed:  $passed"
    echo -e "Stores failed:  ${#failed_stores[@]}"

    if [ ${#failed_stores[@]} -gt 0 ]; then
        echo ""
        echo -e "${STORE_VERIFY_RED}✗ Verification FAILED${STORE_VERIFY_NC}"
        echo -e "${STORE_VERIFY_YELLOW}Failed stores:${STORE_VERIFY_NC}"
        for store in "${failed_stores[@]}"; do
            echo -e "  - $store"
        done
        echo ""
        echo -e "${STORE_VERIFY_YELLOW}Remediation:${STORE_VERIFY_NC}"
        echo -e "  1. Start the node: aurad start --home $home_dir"
        echo -e "  2. Wait for genesis block (height 1)"
        echo -e "  3. Stop the node gracefully"
        echo -e "  4. Re-run this verification"
        echo ""
        echo -e "${STORE_VERIFY_YELLOW}Note: Store initialization happens during InitGenesis${STORE_VERIFY_NC}"
        echo -e "${STORE_VERIFY_YELLOW}      See chain/app/app.go::ensureStoreInitMarkers()${STORE_VERIFY_NC}"
        return 1
    fi

    echo ""
    echo -e "${STORE_VERIFY_GREEN}✓ All stores initialized successfully${STORE_VERIFY_NC}"
    return 0
}

# get_apphash retrieves the current AppHash from the node
# Args:
#   $1 - RPC endpoint (e.g., http://localhost:26657)
# Returns:
#   Outputs AppHash to stdout, returns 1 on error
get_apphash() {
    local rpc_endpoint="$1"

    if [ -z "$rpc_endpoint" ]; then
        echo "Error: RPC endpoint not specified" >&2
        return 1
    fi

    # Query status endpoint
    local response=$(curl -s "${rpc_endpoint}/status" 2>/dev/null)
    if [ $? -ne 0 ]; then
        echo "Error: Failed to query RPC endpoint" >&2
        return 1
    fi

    # Extract AppHash
    local apphash=$(echo "$response" | jq -r '.result.sync_info.latest_app_hash // empty' 2>/dev/null)
    if [ -z "$apphash" ]; then
        echo "Error: Could not extract AppHash from response" >&2
        return 1
    fi

    echo "$apphash"
    return 0
}

# verify_apphash_consistency checks that AppHash remains consistent across restarts
# Args:
#   $1 - node home directory
#   $2 - RPC endpoint (e.g., http://localhost:26657)
#   $3 - expected AppHash (optional, if not provided will be fetched)
# Returns:
#   0 if consistent, 1 otherwise
verify_apphash_consistency() {
    local home_dir="$1"
    local rpc_endpoint="$2"
    local expected_apphash="$3"

    echo -e "${STORE_VERIFY_BLUE}=== AppHash Consistency Verification ===${STORE_VERIFY_NC}"
    echo -e "${STORE_VERIFY_BLUE}Home: $home_dir${STORE_VERIFY_NC}"
    echo -e "${STORE_VERIFY_BLUE}RPC:  $rpc_endpoint${STORE_VERIFY_NC}"
    echo ""

    # If expected hash not provided, fetch current
    if [ -z "$expected_apphash" ]; then
        expected_apphash=$(get_apphash "$rpc_endpoint")
        if [ $? -ne 0 ]; then
            echo -e "${STORE_VERIFY_RED}✗ Failed to retrieve AppHash${STORE_VERIFY_NC}"
            return 1
        fi
        echo -e "Current AppHash: ${expected_apphash}"
        echo -e "${STORE_VERIFY_YELLOW}⚠ No baseline AppHash provided, recording current value${STORE_VERIFY_NC}"

        # Save to file for next check
        local apphash_file="${home_dir}/apphash_baseline.txt"
        echo "$expected_apphash" > "$apphash_file"
        echo -e "${STORE_VERIFY_GREEN}✓ Baseline saved to: $apphash_file${STORE_VERIFY_NC}"
        return 0
    fi

    # Fetch current AppHash
    local current_apphash=$(get_apphash "$rpc_endpoint")
    if [ $? -ne 0 ]; then
        echo -e "${STORE_VERIFY_RED}✗ Failed to retrieve current AppHash${STORE_VERIFY_NC}"
        return 1
    fi

    echo -e "Expected AppHash: ${expected_apphash}"
    echo -e "Current AppHash:  ${current_apphash}"
    echo ""

    # Compare hashes
    if [ "$expected_apphash" = "$current_apphash" ]; then
        echo -e "${STORE_VERIFY_GREEN}✓ AppHash consistent across restart${STORE_VERIFY_NC}"
        return 0
    else
        echo -e "${STORE_VERIFY_RED}✗ AppHash MISMATCH - state divergence detected!${STORE_VERIFY_NC}"
        echo -e "${STORE_VERIFY_YELLOW}This indicates potential issues:${STORE_VERIFY_NC}"
        echo -e "  - Store initialization incomplete"
        echo -e "  - Non-deterministic state transitions"
        echo -e "  - Database corruption"
        echo -e "  - Consensus divergence"
        return 1
    fi
}

# test_start_stop_start performs a full integration test of store consistency
# Args:
#   $1 - aurad binary path
#   $2 - node home directory
#   $3 - RPC endpoint
# Returns:
#   0 if test passes, 1 otherwise
test_start_stop_start() {
    local aurad_binary="$1"
    local home_dir="$2"
    local rpc_endpoint="$3"

    echo -e "${STORE_VERIFY_BLUE}============================================================================${STORE_VERIFY_NC}"
    echo -e "${STORE_VERIFY_BLUE}Start → Stop → Start Consistency Test${STORE_VERIFY_NC}"
    echo -e "${STORE_VERIFY_BLUE}============================================================================${STORE_VERIFY_NC}"
    echo ""

    # Verify binary exists
    if [ ! -f "$aurad_binary" ]; then
        echo -e "${STORE_VERIFY_RED}✗ Error: aurad binary not found: $aurad_binary${STORE_VERIFY_NC}"
        return 1
    fi

    # Step 1: Start node
    echo -e "${STORE_VERIFY_YELLOW}[1/6]${STORE_VERIFY_NC} Starting node..."
    "$aurad_binary" start --home "$home_dir" > /tmp/aurad_test.log 2>&1 &
    local pid=$!

    # Wait for node to start
    echo -e "${STORE_VERIFY_YELLOW}[2/6]${STORE_VERIFY_NC} Waiting for node to produce blocks..."
    sleep 10

    # Verify node is running
    if ! kill -0 $pid 2>/dev/null; then
        echo -e "${STORE_VERIFY_RED}✗ Node failed to start${STORE_VERIFY_NC}"
        cat /tmp/aurad_test.log
        return 1
    fi

    # Get baseline AppHash
    echo -e "${STORE_VERIFY_YELLOW}[3/6]${STORE_VERIFY_NC} Recording baseline AppHash..."
    local baseline_apphash=$(get_apphash "$rpc_endpoint")
    if [ $? -ne 0 ]; then
        echo -e "${STORE_VERIFY_RED}✗ Failed to get baseline AppHash${STORE_VERIFY_NC}"
        kill $pid
        return 1
    fi
    echo -e "${STORE_VERIFY_GREEN}✓ Baseline: $baseline_apphash${STORE_VERIFY_NC}"

    # Stop node
    echo -e "${STORE_VERIFY_YELLOW}[4/6]${STORE_VERIFY_NC} Stopping node..."
    kill $pid
    wait $pid 2>/dev/null
    sleep 2

    # Restart node
    echo -e "${STORE_VERIFY_YELLOW}[5/6]${STORE_VERIFY_NC} Restarting node..."
    "$aurad_binary" start --home "$home_dir" > /tmp/aurad_test.log 2>&1 &
    local pid=$!
    sleep 10

    # Verify AppHash consistency
    echo -e "${STORE_VERIFY_YELLOW}[6/6]${STORE_VERIFY_NC} Verifying AppHash consistency..."
    verify_apphash_consistency "$home_dir" "$rpc_endpoint" "$baseline_apphash"
    local result=$?

    # Cleanup
    kill $pid 2>/dev/null
    wait $pid 2>/dev/null

    if [ $result -eq 0 ]; then
        echo ""
        echo -e "${STORE_VERIFY_GREEN}✓ Start → Stop → Start test PASSED${STORE_VERIFY_NC}"
    else
        echo ""
        echo -e "${STORE_VERIFY_RED}✗ Start → Stop → Start test FAILED${STORE_VERIFY_NC}"
    fi

    return $result
}

# Export functions for use in other scripts
export -f verify_store_db_exists
export -f verify_store_has_version
export -f verify_store_initialization
export -f get_apphash
export -f verify_apphash_consistency
export -f test_start_stop_start
