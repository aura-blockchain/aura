#!/bin/bash
# Test Security Transactions on Aura Testnet
# Tests: Validator security, Network security, Economic security

set -e

CHAIN_ID="aura-local-4"
NODE="http://localhost:26657"
KEYRING="test"
ACCOUNT="test-account-1"
VALIDATOR_ADDR="auravaloper147rneetqamym8rjy28u5n3njzee6s4u0we5nlx"

echo "=== Testing Security Transactions ==="
echo ""

# Test 1: Validator Security - Register Validator
echo "=== Test 1: Register validator with security info ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx validatorsecurity register-validator \
  "validator-1" \
  "security@validator1.com" \
  "https://validator1.com" \
  --from validator-1 \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Register Validator TX Hash: $TX_HASH"
sleep 6
echo ""

# Test 2: Report Double Sign (will likely fail as no actual double sign)
echo "=== Test 2: Report double signing evidence (expected to fail without real evidence) ==="
docker exec aura-validator-1 aurad tx validatorsecurity report-double-sign --help 2>&1 | head -20
echo "Skipping actual execution - requires real double-sign evidence"
echo ""

# Test 3: Economic Security - Lock voting tokens
echo "=== Test 3: Lock tokens for voting power boost ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx economicsecurity lock-voting \
  500000uaura \
  --duration 86400 \
  --from $ACCOUNT \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Lock Voting TX Hash: $TX_HASH"
sleep 6
echo ""

# Test 4: Network Security - Add trusted peer
echo "=== Test 4: Add trusted peer (requires authority) ==="
PEER_ID="26a3efc2bb5a1a84e467f90b45f5f59c79b93685"
PEER_ADDR="tcp://192.168.1.100:26656"
TX_RESULT=$(docker exec aura-validator-1 aurad tx networksecurity add-trusted-peer \
  $PEER_ID \
  $PEER_ADDR \
  --from validator-1 \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Add Trusted Peer TX Hash: $TX_RESULT"
sleep 6
echo ""

# Test 5: Update peer reputation
echo "=== Test 5: Update peer reputation score ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx networksecurity update-reputation \
  $PEER_ID \
  75 \
  --from validator-1 \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Update Reputation TX Hash: $TX_HASH"
sleep 6
echo ""

echo "=== Security tests complete ==="
