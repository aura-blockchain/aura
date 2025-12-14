#!/bin/bash
# Test Multi-send Transaction on Aura Testnet
# Sends tokens from one account to multiple recipients

set -e

CHAIN_ID="aura-local-4"
NODE="http://localhost:26657"
KEYRING="test"

echo "=== Testing Multi-send Transaction ==="
echo ""

# Get addresses
SENDER="test-account-1"
RECIPIENT1="aura199hdwfncvp98g5dkj8jaqz4eytvsm6pr4xzehq"  # test-account-2
RECIPIENT2="aura192x8cq5ut6hyf95zhdfluh5cmtap7fhp7zgpnu"  # test-account-3

echo "Sender: $SENDER"
echo "Recipient 1: $RECIPIENT1"
echo "Recipient 2: $RECIPIENT2"
echo ""

# Check initial balances
echo "Initial balances:"
docker exec aura-validator-1 aurad query bank balances aura1awdvyxzyjjrl2n8hthvft5y6ea5afsa2v00684 --node $NODE | grep -A1 "balances:"
docker exec aura-validator-1 aurad query bank balances $RECIPIENT1 --node $NODE | grep -A1 "balances:"
docker exec aura-validator-1 aurad query bank balances $RECIPIENT2 --node $NODE | grep -A1 "balances:"
echo ""

# Execute multi-send (sends 150000uaura to each recipient)
echo "Executing multi-send transaction (150000uaura to each recipient)..."
TX_HASH=$(docker exec aura-validator-1 aurad tx bank multi-send $SENDER \
  $RECIPIENT1 $RECIPIENT2 150000uaura \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')

echo "Transaction Hash: $TX_HASH"
echo ""

# Wait for transaction to be included in a block
echo "Waiting for transaction to be included..."
sleep 6
echo ""

# Query transaction
echo "Transaction details:"
docker exec aura-validator-1 aurad query tx $TX_HASH --node $NODE 2>&1 | grep -E "(code:|txhash:|gas_used:)"
echo ""

# Check final balances
echo "Final balances:"
docker exec aura-validator-1 aurad query bank balances aura1awdvyxzyjjrl2n8hthvft5y6ea5afsa2v00684 --node $NODE | grep -A1 "balances:"
docker exec aura-validator-1 aurad query bank balances $RECIPIENT1 --node $NODE | grep -A1 "balances:"
docker exec aura-validator-1 aurad query bank balances $RECIPIENT2 --node $NODE | grep -A1 "balances:"
echo ""

echo "=== Multi-send test complete ==="
