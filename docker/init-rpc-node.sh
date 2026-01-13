#!/bin/bash
# Initialize Aura RPC Node
# This script runs once during container initialization

set -e

AURA_HOME="${AURA_HOME:-/root/.aura}"
CHAIN_ID="${AURA_CHAIN_ID:-aura-mvp-1}"
MONIKER="${MONIKER:-aura-rpc-node}"

echo "Initializing Aura RPC Node..."
echo "Home directory: $AURA_HOME"
echo "Chain ID: $CHAIN_ID"
echo "Moniker: $MONIKER"

# Check if node is already initialized
if [ -f "$AURA_HOME/config/genesis.json" ]; then
    echo "Node already initialized. Skipping initialization."
    exit 0
fi

# Initialize the node
echo "Running aurad init..."
aurad init "$MONIKER" --chain-id "$CHAIN_ID" --home "$AURA_HOME"

# Download genesis file (if available)
# TODO: Update with actual genesis file URL when available
GENESIS_URL="${GENESIS_URL:-}"
if [ -n "$GENESIS_URL" ]; then
    echo "Downloading genesis file from $GENESIS_URL..."
    curl -L "$GENESIS_URL" -o "$AURA_HOME/config/genesis.json"
else
    echo "No genesis URL provided. Using default genesis."
fi

# Add persistent peers (if provided)
PERSISTENT_PEERS="${PERSISTENT_PEERS:-}"
if [ -n "$PERSISTENT_PEERS" ]; then
    echo "Setting persistent peers: $PERSISTENT_PEERS"
    sed -i "s/^persistent_peers = .*/persistent_peers = \"$PERSISTENT_PEERS\"/" "$AURA_HOME/config/config.toml"
fi

# Add seeds (if provided)
SEEDS="${SEEDS:-}"
if [ -n "$SEEDS" ]; then
    echo "Setting seeds: $SEEDS"
    sed -i "s/^seeds = .*/seeds = \"$SEEDS\"/" "$AURA_HOME/config/config.toml"
fi

# Set external address (if provided)
EXTERNAL_ADDRESS="${EXTERNAL_ADDRESS:-}"
if [ -n "$EXTERNAL_ADDRESS" ]; then
    echo "Setting external address: $EXTERNAL_ADDRESS"
    sed -i "s/^external_address = .*/external_address = \"$EXTERNAL_ADDRESS\"/" "$AURA_HOME/config/config.toml"
fi

echo "Aura RPC Node initialization complete!"
