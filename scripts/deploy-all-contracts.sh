#!/usr/bin/env bash
# Comprehensive smart contract deployment script for Aura testnet
# Deploys all WASM contracts and provides detailed deployment tracking
#
# Security considerations:
# - Uses test keyring for testnet deployment only
# - Validates all artifacts before deployment
# - Tracks all deployed code IDs and contract addresses
# - Provides rollback information

set -euo pipefail
IFS=$'\n\t'

# Color output for better readability
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project paths
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARTIFACTS_DIR="${ROOT_DIR}/contracts/artifacts"
DEPLOYMENT_LOG="${ROOT_DIR}/contract-deployments.json"

# Testnet configuration
BINARY="${ROOT_DIR}/aurad"
CHAIN_ID="aura-local-4"
NODE="http://localhost:27657"
HOME_DIR="${ROOT_DIR}/testnet-data/validator-1"
KEYRING_BACKEND="test"
FROM_KEY="validator-1"
DENOM="uaura"
GAS_PRICES="0.025${DENOM}"

# Contract artifacts to deploy
declare -A CONTRACTS=(
    ["vc-issuer"]="vc_issuer.wasm"
    ["schema"]="schema.wasm"
    ["binding-tester"]="binding_tester.wasm"
)

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Validate prerequisites
validate_prerequisites() {
    log_info "Validating prerequisites..."

    # Check binary exists and is executable
    if [[ ! -x "$BINARY" ]]; then
        log_error "aurad binary not found or not executable at: $BINARY"
        log_info "Run: cd ${ROOT_DIR}/chain && go build -o aurad ./cmd/aurad"
        exit 1
    fi

    # Check jq is installed
    if ! command -v jq >/dev/null 2>&1; then
        log_error "jq is required but not installed"
        exit 1
    fi

    # Check node is accessible
    if ! curl -sf "${NODE}/status" >/dev/null 2>&1; then
        log_error "Cannot connect to node at: $NODE"
        log_info "Ensure the testnet is running. Try: docker ps | grep aura-validator"
        exit 1
    fi

    # Verify chain ID
    ACTUAL_CHAIN_ID=$(curl -s "${NODE}/status" | jq -r '.result.node_info.network')
    if [[ "$ACTUAL_CHAIN_ID" != "$CHAIN_ID" ]]; then
        log_warning "Expected chain ID: $CHAIN_ID, got: $ACTUAL_CHAIN_ID"
        log_info "Updating chain ID to: $ACTUAL_CHAIN_ID"
        CHAIN_ID="$ACTUAL_CHAIN_ID"
    fi

    # Verify key exists
    if ! "$BINARY" keys show "$FROM_KEY" --home "$HOME_DIR" --keyring-backend "$KEYRING_BACKEND" >/dev/null 2>&1; then
        log_error "Key '$FROM_KEY' not found in keyring"
        exit 1
    fi

    # Check artifacts directory
    if [[ ! -d "$ARTIFACTS_DIR" ]]; then
        log_error "Artifacts directory not found: $ARTIFACTS_DIR"
        exit 1
    fi

    # Validate all contract files exist
    local missing_files=0
    for contract_name in "${!CONTRACTS[@]}"; do
        local file="${CONTRACTS[$contract_name]}"
        if [[ ! -f "${ARTIFACTS_DIR}/${file}" ]]; then
            log_error "Contract artifact missing: ${file}"
            missing_files=$((missing_files + 1))
        fi
    done

    if [[ $missing_files -gt 0 ]]; then
        log_error "Missing $missing_files contract artifacts"
        log_info "Build contracts with: cd ${ROOT_DIR}/contracts && cargo wasm"
        exit 1
    fi

    log_success "All prerequisites validated"
}

# Get deployer address
get_deployer_address() {
    "$BINARY" keys show "$FROM_KEY" --home "$HOME_DIR" --keyring-backend "$KEYRING_BACKEND" --address 2>/dev/null
}

# Store contract code
store_contract() {
    local contract_name="$1"
    local artifact_file="$2"
    local artifact_path="${ARTIFACTS_DIR}/${artifact_file}"

    log_info "Uploading $contract_name contract code..."

    # Prepare transaction flags
    local tx_flags=(
        --from "$FROM_KEY"
        --home "$HOME_DIR"
        --chain-id "$CHAIN_ID"
        --node "$NODE"
        --keyring-backend "$KEYRING_BACKEND"
        --yes
        --broadcast-mode sync
        --output json
        --gas 5000000
        --gas-prices "$GAS_PRICES"
    )

    # Upload contract code
    local store_res
    if ! store_res=$("$BINARY" tx aura_wasm_security store "$artifact_path" "${tx_flags[@]}" 2>&1); then
        log_error "Failed to store $contract_name: $store_res"
        return 1
    fi

    local txhash
    txhash=$(echo "$store_res" | jq -r '.txhash')

    if [[ -z "$txhash" || "$txhash" == "null" ]]; then
        log_error "Failed to extract txhash for $contract_name"
        echo "$store_res" >&2
        return 1
    fi

    log_info "Transaction submitted: $txhash. Waiting for confirmation..."
    sleep 6  # Wait for block inclusion

    # Query transaction to get code ID from events
    local tx_result
    if ! tx_result=$(curl -s "${NODE}/tx?hash=0x${txhash}" 2>&1); then
        log_error "Failed to query transaction: $tx_result"
        return 1
    fi

    # Extract code ID from events
    local code_id
    code_id=$(echo "$tx_result" | jq -r '.result.tx_result.events[]? | select(.type=="store_code") | .attributes[]? | select(.key=="code_id") | .value' | head -n1)

    if [[ -z "$code_id" || "$code_id" == "null" ]]; then
        log_error "Failed to extract code_id for $contract_name from tx events"
        echo "$tx_result" >&2
        return 1
    fi

    log_success "Stored $contract_name: code_id=$code_id, txhash=$txhash"

    echo "$code_id"
}

# Instantiate contract
instantiate_contract() {
    local contract_name="$1"
    local code_id="$2"
    local admin_address="$3"

    log_info "Instantiating $contract_name contract..."

    # Prepare init message based on contract type
    local init_msg
    case "$contract_name" in
        vc-issuer)
            init_msg=$(jq -n --arg admin "$admin_address" '{admin: $admin}')
            ;;
        schema)
            init_msg=$(jq -n '{}')
            ;;
        binding-tester)
            init_msg=$(jq -n '{}')
            ;;
        *)
            log_error "Unknown contract type: $contract_name"
            return 1
            ;;
    esac

    local label="${contract_name}-$(date +%s)"

    # Prepare transaction flags
    local tx_flags=(
        --from "$FROM_KEY"
        --home "$HOME_DIR"
        --chain-id "$CHAIN_ID"
        --node "$NODE"
        --keyring-backend "$KEYRING_BACKEND"
        --yes
        --broadcast-mode sync
        --output json
        --gas 3000000
        --gas-prices "$GAS_PRICES"
        --admin "$admin_address"
    )

    # Instantiate contract
    local inst_res
    if ! inst_res=$("$BINARY" tx aura_wasm_security instantiate "$code_id" "$init_msg" --label "$label" "${tx_flags[@]}" 2>&1); then
        log_error "Failed to instantiate $contract_name: $inst_res"
        return 1
    fi

    local txhash
    txhash=$(echo "$inst_res" | jq -r '.txhash')

    if [[ -z "$txhash" || "$txhash" == "null" ]]; then
        log_error "Failed to extract txhash for $contract_name instantiation"
        echo "$inst_res" >&2
        return 1
    fi

    log_info "Instantiation transaction submitted: $txhash. Waiting for confirmation..."
    sleep 6  # Wait for block inclusion

    # Query transaction to get contract address from events
    local tx_result
    if ! tx_result=$(curl -s "${NODE}/tx?hash=0x${txhash}" 2>&1); then
        log_error "Failed to query instantiation transaction: $tx_result"
        return 1
    fi

    # Extract contract address from events
    local contract_addr
    contract_addr=$(echo "$tx_result" | jq -r '.result.tx_result.events[]? | select(.type=="instantiate") | .attributes[]? | select(.key=="_contract_address") | .value' | head -n1)

    if [[ -z "$contract_addr" || "$contract_addr" == "null" ]]; then
        log_error "Failed to extract contract address for $contract_name from tx events"
        echo "$tx_result" >&2
        return 1
    fi

    log_success "Instantiated $contract_name: address=$contract_addr, txhash=$txhash"

    echo "$contract_addr"
}

# Query contract info to verify deployment
verify_contract() {
    local contract_name="$1"
    local contract_addr="$2"

    log_info "Verifying $contract_name deployment..."

    local contract_info
    if ! contract_info=$("$BINARY" query aura_wasm_security contract "$contract_addr" --node "$NODE" --chain-id "$CHAIN_ID" --output json 2>&1); then
        log_error "Failed to query contract info: $contract_info"
        return 1
    fi

    local code_id
    code_id=$(echo "$contract_info" | jq -r '.code_id')

    local admin
    admin=$(echo "$contract_info" | jq -r '.admin // "none"')

    log_success "$contract_name verified: code_id=$code_id, admin=$admin"

    return 0
}

# Save deployment information
save_deployment_info() {
    local deployer_address="$1"
    shift
    local -n deployments=$1

    log_info "Saving deployment information to $DEPLOYMENT_LOG..."

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Create deployment record
    local deployment_record
    deployment_record=$(jq -n \
        --arg timestamp "$timestamp" \
        --arg chain_id "$CHAIN_ID" \
        --arg node "$NODE" \
        --arg deployer "$deployer_address" \
        --argjson contracts "$(declare -p deployments | sed 's/declare -A [^=]*=//' | jq -R 'fromjson')" \
        '{
            timestamp: $timestamp,
            chain_id: $chain_id,
            node: $node,
            deployer: $deployer,
            contracts: []
        }')

    # Build contracts array
    local contracts_json="[]"
    for contract_name in "${!deployments[@]}"; do
        local contract_data="${deployments[$contract_name]}"
        local code_id=$(echo "$contract_data" | cut -d'|' -f1)
        local address=$(echo "$contract_data" | cut -d'|' -f2)

        contracts_json=$(echo "$contracts_json" | jq \
            --arg name "$contract_name" \
            --arg code_id "$code_id" \
            --arg address "$address" \
            '. += [{name: $name, code_id: $code_id, address: $address}]')
    done

    deployment_record=$(echo "$deployment_record" | jq --argjson contracts "$contracts_json" '.contracts = $contracts')

    # Append to deployment log (create if doesn't exist)
    if [[ -f "$DEPLOYMENT_LOG" ]]; then
        local existing_deployments
        existing_deployments=$(cat "$DEPLOYMENT_LOG")
        echo "$existing_deployments" | jq --argjson new "$deployment_record" '. += [$new]' > "$DEPLOYMENT_LOG"
    else
        echo "[$deployment_record]" > "$DEPLOYMENT_LOG"
    fi

    log_success "Deployment information saved"
}

# Print deployment summary
print_summary() {
    local deployer_address="$1"
    shift
    local -n deployments=$1

    echo ""
    echo "=========================================="
    echo "  AURA CONTRACT DEPLOYMENT SUMMARY"
    echo "=========================================="
    echo ""
    echo "Chain ID:        $CHAIN_ID"
    echo "Node:            $NODE"
    echo "Deployer:        $deployer_address"
    echo "Timestamp:       $(date -u +"%Y-%m-%d %H:%M:%S UTC")"
    echo ""
    echo "Deployed Contracts:"
    echo "------------------------------------------"

    for contract_name in "${!deployments[@]}"; do
        local contract_data="${deployments[$contract_name]}"
        local code_id=$(echo "$contract_data" | cut -d'|' -f1)
        local address=$(echo "$contract_data" | cut -d'|' -f2)

        echo ""
        echo "Contract: $contract_name"
        echo "  Code ID:  $code_id"
        echo "  Address:  $address"
    done

    echo ""
    echo "------------------------------------------"
    echo "Verification Commands:"
    echo ""

    for contract_name in "${!deployments[@]}"; do
        local contract_data="${deployments[$contract_name]}"
        local address=$(echo "$contract_data" | cut -d'|' -f2)

        echo "# Query $contract_name"
        echo "$BINARY query aura_wasm_security contract $address --node $NODE --chain-id $CHAIN_ID"
        echo ""
    done

    echo "Deployment log saved to: $DEPLOYMENT_LOG"
    echo "=========================================="
}

# Main deployment workflow
main() {
    log_info "Starting Aura smart contract deployment"
    echo ""

    # Validate environment
    validate_prerequisites
    echo ""

    # Get deployer address (will be admin for contracts)
    local deployer_address
    deployer_address=$(get_deployer_address)
    log_info "Deployer address: $deployer_address"
    echo ""

    # Track deployments
    declare -A deployed_contracts

    # Deploy each contract
    for contract_name in "${!CONTRACTS[@]}"; do
        local artifact_file="${CONTRACTS[$contract_name]}"

        echo "=========================================="
        log_info "Deploying: $contract_name"
        echo "=========================================="

        # Store contract code
        local code_id
        if ! code_id=$(store_contract "$contract_name" "$artifact_file"); then
            log_error "Failed to deploy $contract_name"
            continue
        fi

        # Instantiate contract
        local contract_addr
        if ! contract_addr=$(instantiate_contract "$contract_name" "$code_id" "$deployer_address"); then
            log_error "Failed to instantiate $contract_name"
            continue
        fi

        # Verify deployment
        if ! verify_contract "$contract_name" "$contract_addr"; then
            log_warning "$contract_name deployment verification failed"
        fi

        # Record deployment
        deployed_contracts["$contract_name"]="${code_id}|${contract_addr}"

        echo ""
    done

    # Save deployment information
    local num_deployed=${#deployed_contracts[@]}
    if [[ $num_deployed -gt 0 ]]; then
        save_deployment_info "$deployer_address" deployed_contracts
        echo ""
        print_summary "$deployer_address" deployed_contracts
    else
        log_error "No contracts were successfully deployed"
        exit 1
    fi

    log_success "Contract deployment complete!"
}

# Run main function
main "$@"
