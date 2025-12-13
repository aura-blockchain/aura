#!/bin/bash
# Quick 4-Validator Testnet Setup Script
# This script creates a working 4-validator testnet by directly modifying genesis
# Bypasses the problematic gentx processing in Cosmos SDK

set -e

CHAIN_ID="aura-local-4"
DENOM="uaura"
# Use 1 trillion per validator to be well above DefaultPowerReduction
STAKING_AMOUNT="1000000000000"
BINARY="../chain/aurad"

echo "=== Aura 4-Validator Quick Start ==="
echo ""

# Clean existing data
echo "[1/5] Cleaning existing testnet data..."
rm -rf testnet-data-quick
mkdir -p testnet-data-quick

# Initialize 4 validators
echo "[2/5] Initializing 4 validators..."
for i in 1 2 3 4; do
  MONIKER="validator-$i"
  NODE_HOME="testnet-data-quick/$MONIKER"

  $BINARY init "$MONIKER" --chain-id "$CHAIN_ID" --home "$NODE_HOME" &> /dev/null

  # Create validator key
  echo "password" | $BINARY keys add "$MONIKER" --keyring-backend test --home "$NODE_HOME" &> /dev/null

  # Get address
  ADDR=$($BINARY keys show "$MONIKER" --keyring-backend test --home "$NODE_HOME" --address)

  # Add genesis account with tokens
  $BINARY genesis add-genesis-account "$ADDR" "10000000000000${DENOM}" --home "$NODE_HOME"

  echo "  ✓ $MONIKER initialized ($ADDR)"
done

echo ""
echo "[3/5] Building combined genesis with 4 validators..."

# Use validator-1 as the base
BASE_HOME="testnet-data-quick/validator-1"
GENESIS="$BASE_HOME/config/genesis.json"

# Add other validator accounts to base genesis
for i in 2 3 4; do
  MONIKER="validator-$i"
  NODE_HOME="testnet-data-quick/$MONIKER"
  ADDR=$($BINARY keys show "$MONIKER" --keyring-backend test --home "$NODE_HOME" --address)

  $BINARY genesis add-genesis-account "$ADDR" "10000000000000${DENOM}" --home "$BASE_HOME"
done

# Create gentx for all 4 validators
echo ""
echo "[4/5] Creating genesis transactions..."
for i in 1 2 3 4; do
  MONIKER="validator-$i"
  NODE_HOME="testnet-data-quick/$MONIKER"

  # Copy base genesis to validator home
  cp "$GENESIS" "$NODE_HOME/config/genesis.json"

  # Create gentx
  echo "password" | $BINARY genesis gentx "$MONIKER" "${STAKING_AMOUNT}${DENOM}" \
    --chain-id "$CHAIN_ID" \
    --keyring-backend test \
    --home "$NODE_HOME" \
    --moniker "$MONIKER" \
    --commission-rate "0.10" \
    --commission-max-rate "0.20" \
    --commission-max-change-rate "0.01" &> /dev/null

  # Copy gentx to base
  cp "$NODE_HOME/config/gentx/"*.json "$BASE_HOME/config/gentx/"

  echo "  ✓ $MONIKER gentx created"
done

# Collect all gentx
echo ""
echo "Collecting genesis transactions..."
$BINARY genesis collect-gentxs --home "$BASE_HOME" &> /dev/null

# Apply critical fixes
echo "Applying fixes..."
jq '.app_state.staking.params.bond_denom = "uaura"' "$GENESIS" > /tmp/g.json && mv /tmp/g.json "$GENESIS"
jq '.app_state.staking.validators = []' "$GENESIS" > /tmp/g.json && mv /tmp/g.json "$GENESIS"

# Distribute final genesis
for i in 2 3 4; do
  cp "$GENESIS" "testnet-data-quick/validator-$i/config/genesis.json"
done

echo "  ✓ Genesis finalized"

# Configure peers
echo ""
echo "[5/5] Configuring network..."
declare -a NODE_IDS
for i in 1 2 3 4; do
  NODE_HOME="testnet-data-quick/validator-$i"
  NODE_ID=$($BINARY tendermint show-node-id --home "$NODE_HOME")
  NODE_IDS[$i]="$NODE_ID"
done

# Set persistent peers for each validator
for i in 1 2 3 4; do
  NODE_HOME="testnet-data-quick/validator-$i"
  PEERS=""
  for j in 1 2 3 4; do
    if [ $i -ne $j ]; then
      if [ -n "$PEERS" ]; then
        PEERS="$PEERS,"
      fi
      PEERS="$PEERS${NODE_IDS[$j]}@validator-$j:26656"
    fi
  done
  sed -i.bak "s/^persistent_peers =.*/persistent_peers = \"$PEERS\"/" "$NODE_HOME/config/config.toml"
done

echo "  ✓ Peer configuration complete"
echo ""
echo "=== Setup Complete ==="
echo ""
echo "Validators created:"
for i in 1 2 3 4; do
  NODE_HOME="testnet-data-quick/validator-$i"
  ADDR=$($BINARY keys show "validator-$i" --keyring-backend test --home "$NODE_HOME" --address)
  echo "  validator-$i: $ADDR"
done
echo ""
echo "To start the testnet:"
echo "  cd testnet-data-quick"
echo "  For each validator in separate terminals:"
echo "    $BINARY start --home validator-1"
echo "    $BINARY start --home validator-2"
echo "    $BINARY start --home validator-3"
echo "    $BINARY start --home validator-4"
