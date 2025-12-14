#!/bin/bash
# Test Governance Transactions on Aura Testnet
# Tests: Submit proposal, Deposit, Vote

set -e

CHAIN_ID="aura-local-4"
NODE="http://localhost:26657"
KEYRING="test"
PROPOSER="test-account-2"
PROPOSER_ADDR="aura199hdwfncvp98g5dkj8jaqz4eytvsm6pr4xzehq"
VOTER="test-account-3"

echo "=== Testing Governance Transactions ==="
echo ""
echo "Proposer: $PROPOSER ($PROPOSER_ADDR)"
echo "Voter: $VOTER"
echo ""

# Initial balance
echo "Initial proposer balance:"
docker exec aura-validator-1 aurad query bank balances $PROPOSER_ADDR --node $NODE | grep -A1 "balances:"
echo ""

# Test 1: Submit Proposal
echo "=== Test 1: Submit a text proposal ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx governance submit-proposal \
  "Test Proposal for Aura" \
  "This is a test governance proposal to verify the governance module functionality" \
  text \
  --from $PROPOSER \
  --initial-deposit 1000000uaura \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Submit Proposal TX Hash: $TX_HASH"
sleep 6
echo ""

# Get proposal ID
echo "Querying proposals to get proposal ID:"
PROPOSAL_ID=$(docker exec aura-validator-1 aurad query governance proposals --node $NODE --output json 2>&1 | jq -r '.proposals[-1].id' 2>/dev/null || echo "1")
echo "Proposal ID: $PROPOSAL_ID"
echo ""

# Test 2: Add deposit to proposal
echo "=== Test 2: Add deposit to proposal ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx governance deposit $PROPOSAL_ID 500000uaura \
  --from $VOTER \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Deposit TX Hash: $TX_HASH"
sleep 6
echo ""

# Test 3: Vote on proposal
echo "=== Test 3: Vote YES on proposal ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx governance vote $PROPOSAL_ID yes \
  --from $VOTER \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Vote TX Hash: $TX_HASH"
sleep 6
echo ""

# Test 4: Weighted Vote
echo "=== Test 4: Submit weighted vote ==="
TX_HASH=$(docker exec aura-validator-1 aurad tx governance vote-weighted $PROPOSAL_ID \
  "yes=0.6,no=0.3,abstain=0.1" \
  --from $PROPOSER \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING \
  --node $NODE \
  --yes \
  --fees 5000uaura \
  --broadcast-mode sync 2>&1 | grep "txhash:" | awk '{print $2}')
echo "Weighted Vote TX Hash: $TX_HASH"
sleep 6
echo ""

# Check proposal status
echo "Final proposal status:"
docker exec aura-validator-1 aurad query governance proposal $PROPOSAL_ID --node $NODE 2>&1 | head -30 || echo "Query not available"
echo ""

# Final balance
echo "Final proposer balance:"
docker exec aura-validator-1 aurad query bank balances $PROPOSER_ADDR --node $NODE | grep -A1 "balances:"
echo ""

echo "=== Governance tests complete ==="
echo "Proposal ID: $PROPOSAL_ID"
