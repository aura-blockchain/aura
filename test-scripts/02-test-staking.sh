#!/bin/bash
# Test Staking Transactions on Aura Testnet
# Tests: Delegate, Undelegate, Redelegate, Claim rewards

set -e

CHAIN_ID="aura-local-4"
NODE="http://localhost:26657"
KEYRING="test"
DELEGATOR="test-account-1"
DELEGATOR_ADDR="aura1awdvyxzyjjrl2n8hthvft5y6ea5afsa2v00684"
VALIDATOR1="auravaloper147rneetqamym8rjy28u5n3njzee6s4u0we5nlx"

echo "=== Testing Staking Transactions ==="
echo ""
echo "Delegator: $DELEGATOR ($DELEGATOR_ADDR)"
echo "Validator: $VALIDATOR1"
echo ""

# Initial balance
echo "Initial delegator balance:"
docker exec aura-validator-1 aurad query bank balances $DELEGATOR_ADDR --node $NODE | grep -A1 "balances:"
echo ""

# Test 1: Delegate
echo "=== Test 1: Delegate 1000000uaura to validator ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx staking delegate $VALIDATOR1 1000000uaura \
  --from $DELEGATOR \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Delegate TX Hash: $TX_HASH"
sleep 6
echo ""

# Check delegation
echo "Delegation after delegate:"
docker exec aura-validator-1 aurad query distribution rewards $DELEGATOR_ADDR --node $NODE 2>&1 | head -20 || echo "No rewards yet"
echo ""

# Test 2: Claim rewards (might fail if no rewards accumulated)
echo "=== Test 2: Claim rewards ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx distribution withdraw-all-rewards \
  --from $DELEGATOR \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Withdraw rewards TX Hash: $TX_HASH"
sleep 6
echo ""

# Balance after claiming
echo "Balance after claiming rewards:"
docker exec aura-validator-1 aurad query bank balances $DELEGATOR_ADDR --node $NODE | grep -A1 "balances:"
echo ""

# Test 3: Undelegate
echo "=== Test 3: Undelegate 500000uaura from validator ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx staking unbond $VALIDATOR1 500000uaura \
  --from $DELEGATOR \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Unbond TX Hash: $TX_HASH"
sleep 6
echo ""

# Test 4: Redelegate (need another validator)
echo "=== Test 4: Redelegate (checking for second validator) ==="
VALIDATOR2=$(docker exec aura-validator-2 aurad keys show validator-2 --bech=val --keyring-backend test 2>&1 | grep "address:" | awk '{print $2}')
if [ -n "$VALIDATOR2" ]; then
  echo "Second validator: $VALIDATOR2"
  TX_HASH=$(docker exec aura-validator-1 aurad tx staking redelegate $VALIDATOR1 $VALIDATOR2 100000uaura \
    --from $DELEGATOR \
    --chain-id $CHAIN_ID \
    --keyring-backend $KEYRING \
    --node $NODE \
    --yes \
    --fees 5000uaura \
    --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
  echo "Redelegate TX Hash: $TX_HASH"
  sleep 6
else
  echo "No second validator found, skipping redelegate test"
fi
echo ""

# Final balance
echo "Final delegator balance:"
docker exec aura-validator-1 aurad query bank balances $DELEGATOR_ADDR --node $NODE | grep -A1 "balances:"
echo ""

echo "=== Staking tests complete ==="
