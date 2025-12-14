#!/bin/bash
# Test WASM Transactions on Aura Testnet
# Tests: Store code, Instantiate contract, Execute contract, Migrate contract

set -e

CHAIN_ID="aura-local-4"
NODE="http://localhost:26657"
KEYRING="test"
UPLOADER="test-account-1"
WASM_FILE="/workspace/third_party/cosmwasm-vm/testdata/hackatom_1.2.wasm"

echo "=== Testing WASM Transactions ==="
echo ""
echo "Uploader: $UPLOADER"
echo "WASM file: $WASM_FILE"
echo ""

# Copy WASM file into container
echo "Copying WASM file to container..."
docker cp /home/hudson/blockchain-projects/aura/third_party/cosmwasm-vm/testdata/hackatom_1.2.wasm aura-validator-1:/tmp/test.wasm
echo ""

# Test 1: Store WASM code
echo "=== Test 1: Store WASM code ==="
TX_RESULT=$(docker exec aura-validator-1 aurad tx aura_wasm_security store /tmp/test.wasm \
  --from $UPLOADER \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --gas 3000000 \
  --fees 50000uaura \
  --broadcast-mode sync 2>&1)
TX_HASH=$(echo "$TX_RESULT" | grep "txhash:" | awk '{print $2}')
echo "Store Code TX Hash: $TX_HASH"
echo "$TX_RESULT" | grep -E "(code:|txhash:)"
sleep 6
echo ""

# Get code ID from events (will query contracts to get it)
echo "Querying stored code..."
CODE_ID=$(docker exec aura-validator-1 aurad query aura_wasm_security list-code --node $NODE --output json 2>&1 | jq -r '.code_infos[-1].code_id // "1"' 2>/dev/null || echo "1")
echo "Code ID: $CODE_ID"
echo ""

# Test 2: Instantiate contract
echo "=== Test 2: Instantiate WASM contract ==="
INIT_MSG='{"verifier":"aura1awdvyxzyjjrl2n8hthvft5y6ea5afsa2v00684","beneficiary":"aura199hdwfncvp98g5dkj8jaqz4eytvsm6pr4xzehq"}'
TX_HASH=$(docker exec aura-validator-1 aurad tx aura_wasm_security instantiate $CODE_ID "$INIT_MSG" \
  --label "test-contract-1" \
  --from $UPLOADER \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --gas 1000000 \
  --fees 20000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Instantiate Contract TX Hash: $TX_HASH"
sleep 6
echo ""

# Get contract address
echo "Querying contract address..."
CONTRACT_ADDR=$(docker exec aura-validator-1 aurad query aura_wasm_security list-contract-by-code $CODE_ID --node $NODE --output json 2>&1 | jq -r '.contracts[-1] // "aura14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0phg4d"' 2>/dev/null || echo "aura14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0phg4d")
echo "Contract Address: $CONTRACT_ADDR"
echo ""

# Test 3: Execute contract
echo "=== Test 3: Execute WASM contract ==="
EXEC_MSG='{"release":{}}'
TX_HASH=$(docker exec aura-validator-1 aurad tx aura_wasm_security execute $CONTRACT_ADDR "$EXEC_MSG" \
  --from $UPLOADER \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --gas 500000 \
  --fees 10000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Execute Contract TX Hash: $TX_HASH"
sleep 6
echo ""

# Test 4: Pause contract (requires governance/authority)
echo "=== Test 4: Pause contract ==="
TX_RESULT=$(docker exec aura-validator-1 aurad tx aura_wasm_security pause-contract $CONTRACT_ADDR \
  "Security testing pause" \
  --from validator-1 \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 10000uaura \
  --broadcast-mode sync 2>&1)
TX_HASH=$(echo "$TX_RESULT" | grep "txhash:" | awk '{print $2}')
echo "Pause Contract TX Hash: $TX_HASH"
echo "$TX_RESULT" | grep -E "(code:|txhash:)"
sleep 6
echo ""

# Test 5: Migrate contract (requires new code and admin)
echo "=== Test 5: Migrate contract (setting admin first) ==="
# Set admin
TX_HASH=$(docker exec aura-validator-1 aurad tx aura_wasm_security set-admin $CONTRACT_ADDR \
  aura1awdvyxzyjjrl2n8hthvft5y6ea5afsa2v00684 \
  --from $UPLOADER \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 10000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Set Admin TX Hash: $TX_HASH"
sleep 6
echo ""

# Migrate (to same code for testing)
MIGRATE_MSG='{}'
TX_HASH=$(docker exec aura-validator-1 aurad tx aura_wasm_security migrate $CONTRACT_ADDR $CODE_ID "$MIGRATE_MSG" \
  --from $UPLOADER \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --gas 500000 \
  --fees 10000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Migrate Contract TX Hash: $TX_HASH"
sleep 6
echo ""

echo "=== WASM tests complete ==="
echo "Code ID: $CODE_ID"
echo "Contract Address: $CONTRACT_ADDR"
