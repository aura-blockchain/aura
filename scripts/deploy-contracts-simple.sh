#!/usr/bin/env bash
# Simple contract deployment script for Aura testnet
# Deploys all WASM contracts with minimal complexity

set -euo pipefail

# Color output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${ROOT_DIR}/aurad"
CHAIN_ID="aura-local-4"
NODE="http://localhost:27657"
HOME_DIR="${ROOT_DIR}/testnet-data/validator-1"
FROM_KEY="validator-1"
KEYRING="test"
GAS_PRICES="0.025uaura"

# Get deployer address
DEPLOYER=$("$BINARY" keys show "$FROM_KEY" --home "$HOME_DIR" --keyring-backend "$KEYRING" --address)

echo -e "${BLUE}Deploying contracts to Aura testnet${NC}"
echo "Chain ID: $CHAIN_ID"
echo "Deployer: $DEPLOYER"
echo ""

# Array to track deployments
declare -a DEPLOYMENTS

# Deploy vc-issuer
echo -e "${BLUE}[1/3] Deploying vc-issuer contract...${NC}"

# Store code
echo "  Uploading WASM code..."
STORE_TX=$("$BINARY" tx aura_wasm_security store "${ROOT_DIR}/contracts/artifacts/vc_issuer.wasm" \
    --from "$FROM_KEY" --home "$HOME_DIR" --chain-id "$CHAIN_ID" --node "$NODE" \
    --keyring-backend "$KEYRING" --yes --broadcast-mode sync --gas 5000000 \
    --gas-prices "$GAS_PRICES" --output json)

STORE_HASH=$(echo "$STORE_TX" | jq -r '.txhash')
echo "  TX submitted: $STORE_HASH"
sleep 6

# Get code ID
CODE_ID=$(curl -s "${NODE}/tx?hash=0x${STORE_HASH}" | jq -r '.result.tx_result.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
echo -e "  ${GREEN}Code stored: ID=$CODE_ID${NC}"

# Instantiate
echo "  Instantiating contract..."
INST_TX=$("$BINARY" tx aura_wasm_security instantiate "$CODE_ID" "{\"admin\":\"$DEPLOYER\"}" \
    --label "vc-issuer-$(date +%s)" --from "$FROM_KEY" --home "$HOME_DIR" --chain-id "$CHAIN_ID" \
    --node "$NODE" --keyring-backend "$KEYRING" --yes --broadcast-mode sync --gas 3000000 \
    --gas-prices "$GAS_PRICES" --admin "$DEPLOYER" --output json)

INST_HASH=$(echo "$INST_TX" | jq -r '.txhash')
echo "  TX submitted: $INST_HASH"
sleep 6

# Get contract address
CONTRACT=$(curl -s "${NODE}/tx?hash=0x${INST_HASH}" | jq -r '.result.tx_result.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
echo -e "  ${GREEN}Contract deployed: $CONTRACT${NC}"
DEPLOYMENTS+=("vc-issuer|$CODE_ID|$CONTRACT")
echo ""

# Deploy schema
echo -e "${BLUE}[2/3] Deploying schema contract...${NC}"

# Store code
echo "  Uploading WASM code..."
STORE_TX=$("$BINARY" tx aura_wasm_security store "${ROOT_DIR}/contracts/artifacts/schema.wasm" \
    --from "$FROM_KEY" --home "$HOME_DIR" --chain-id "$CHAIN_ID" --node "$NODE" \
    --keyring-backend "$KEYRING" --yes --broadcast-mode sync --gas 5000000 \
    --gas-prices "$GAS_PRICES" --output json)

STORE_HASH=$(echo "$STORE_TX" | jq -r '.txhash')
echo "  TX submitted: $STORE_HASH"
sleep 6

# Get code ID
CODE_ID=$(curl -s "${NODE}/tx?hash=0x${STORE_HASH}" | jq -r '.result.tx_result.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
echo -e "  ${GREEN}Code stored: ID=$CODE_ID${NC}"

# Instantiate
echo "  Instantiating contract..."
INST_TX=$("$BINARY" tx aura_wasm_security instantiate "$CODE_ID" '{}' \
    --label "schema-$(date +%s)" --from "$FROM_KEY" --home "$HOME_DIR" --chain-id "$CHAIN_ID" \
    --node "$NODE" --keyring-backend "$KEYRING" --yes --broadcast-mode sync --gas 3000000 \
    --gas-prices "$GAS_PRICES" --admin "$DEPLOYER" --output json)

INST_HASH=$(echo "$INST_TX" | jq -r '.txhash')
echo "  TX submitted: $INST_HASH"
sleep 6

# Get contract address
CONTRACT=$(curl -s "${NODE}/tx?hash=0x${INST_HASH}" | jq -r '.result.tx_result.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
echo -e "  ${GREEN}Contract deployed: $CONTRACT${NC}"
DEPLOYMENTS+=("schema|$CODE_ID|$CONTRACT")
echo ""

# Deploy binding-tester
echo -e "${BLUE}[3/3] Deploying binding-tester contract...${NC}"

# Store code
echo "  Uploading WASM code..."
STORE_TX=$("$BINARY" tx aura_wasm_security store "${ROOT_DIR}/contracts/artifacts/binding_tester.wasm" \
    --from "$FROM_KEY" --home "$HOME_DIR" --chain-id "$CHAIN_ID" --node "$NODE" \
    --keyring-backend "$KEYRING" --yes --broadcast-mode sync --gas 5000000 \
    --gas-prices "$GAS_PRICES" --output json)

STORE_HASH=$(echo "$STORE_TX" | jq -r '.txhash')
echo "  TX submitted: $STORE_HASH"
sleep 6

# Get code ID
CODE_ID=$(curl -s "${NODE}/tx?hash=0x${STORE_HASH}" | jq -r '.result.tx_result.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
echo -e "  ${GREEN}Code stored: ID=$CODE_ID${NC}"

# Instantiate
echo "  Instantiating contract..."
INST_TX=$("$BINARY" tx aura_wasm_security instantiate "$CODE_ID" '{}' \
    --label "binding-tester-$(date +%s)" --from "$FROM_KEY" --home "$HOME_DIR" --chain-id "$CHAIN_ID" \
    --node "$NODE" --keyring-backend "$KEYRING" --yes --broadcast-mode sync --gas 3000000 \
    --gas-prices "$GAS_PRICES" --admin "$DEPLOYER" --output json)

INST_HASH=$(echo "$INST_TX" | jq -r '.txhash')
echo "  TX submitted: $INST_HASH"
sleep 6

# Get contract address
CONTRACT=$(curl -s "${NODE}/tx?hash=0x${INST_HASH}" | jq -r '.result.tx_result.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
echo -e "  ${GREEN}Contract deployed: $CONTRACT${NC}"
DEPLOYMENTS+=("binding-tester|$CODE_ID|$CONTRACT")
echo ""

# Save deployment info
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
DEPLOY_FILE="${ROOT_DIR}/contract-deployments.json"

# Build JSON
JSON_CONTRACTS="[]"
for deployment in "${DEPLOYMENTS[@]}"; do
    IFS='|' read -r name code_id address <<< "$deployment"
    JSON_CONTRACTS=$(echo "$JSON_CONTRACTS" | jq \
        --arg name "$name" \
        --arg code_id "$code_id" \
        --arg address "$address" \
        '. += [{name: $name, code_id: $code_id, address: $address}]')
done

DEPLOYMENT_RECORD=$(jq -n \
    --arg timestamp "$TIMESTAMP" \
    --arg chain_id "$CHAIN_ID" \
    --arg node "$NODE" \
    --arg deployer "$DEPLOYER" \
    --argjson contracts "$JSON_CONTRACTS" \
    '{
        timestamp: $timestamp,
        chain_id: $chain_id,
        node: $node,
        deployer: $deployer,
        contracts: $contracts
    }')

if [[ -f "$DEPLOY_FILE" ]]; then
    EXISTING=$(cat "$DEPLOY_FILE")
    echo "$EXISTING" | jq --argjson new "$DEPLOYMENT_RECORD" '. += [$new]' > "$DEPLOY_FILE"
else
    echo "[$DEPLOYMENT_RECORD]" > "$DEPLOY_FILE"
fi

# Print summary
echo "=========================================="
echo "  DEPLOYMENT SUMMARY"
echo "=========================================="
echo ""
echo "Timestamp:  $TIMESTAMP"
echo "Chain ID:   $CHAIN_ID"
echo "Node:       $NODE"
echo "Deployer:   $DEPLOYER"
echo ""
echo "Deployed Contracts:"
echo ""

for deployment in "${DEPLOYMENTS[@]}"; do
    IFS='|' read -r name code_id address <<< "$deployment"
    echo "$name"
    echo "  Code ID:  $code_id"
    echo "  Address:  $address"
    echo ""
done

echo "Deployment log: $DEPLOY_FILE"
echo ""
echo "Verification Commands:"
echo ""

for deployment in "${DEPLOYMENTS[@]}"; do
    IFS='|' read -r name code_id address <<< "$deployment"
    echo "# Query $name"
    echo "$BINARY query aura_wasm_security contract $address --node $NODE --chain-id $CHAIN_ID"
    echo ""
done

echo "=========================================="
echo -e "${GREEN}All contracts deployed successfully!${NC}"
