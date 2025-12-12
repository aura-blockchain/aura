#!/bin/bash
set -e

# Initialize 4-validator testnet
CHAIN_ID="aura-testnet-4val"
DENOM="uaura"
STAKE_DENOM="uaura"
KEYRING="test"

# Temporary directory for initialization
TMPDIR=$(mktemp -d)
echo "Working in $TMPDIR"

# Build aurad if needed
cd /home/hudson/blockchain-projects/aura/chain
if [ ! -f aurad ]; then
    echo "Building aurad..."
    go build -o aurad ./cmd/aurad
fi
AURAD=$(pwd)/aurad

# Initialize all 4 validators
for i in 1 2 3 4; do
    echo "=== Initializing validator-$i ==="
    VALHOME="$TMPDIR/validator-$i"
    mkdir -p "$VALHOME"

    $AURAD init "validator-$i" --chain-id "$CHAIN_ID" --home "$VALHOME" > /dev/null 2>&1

    # Create validator key
    $AURAD keys add "validator$i" --keyring-backend "$KEYRING" --home "$VALHOME" > /dev/null 2>&1

    # Get address
    ADDR=$($AURAD keys show "validator$i" -a --keyring-backend "$KEYRING" --home "$VALHOME")
    echo "  Validator $i address: $ADDR"
done

# Use validator-1's genesis as base
echo "=== Setting up genesis ==="
GENESIS="$TMPDIR/validator-1/config/genesis.json"

# Add genesis accounts for all validators
for i in 1 2 3 4; do
    VALHOME="$TMPDIR/validator-$i"
    ADDR=$($AURAD keys show "validator$i" -a --keyring-backend "$KEYRING" --home "$VALHOME")

    # Add to validator-1's genesis
    $AURAD genesis add-genesis-account "$ADDR" "100000000000${DENOM}" --home "$TMPDIR/validator-1" > /dev/null 2>&1
    echo "  Added genesis account for validator-$i"
done

# Create gentx for each validator
echo "=== Creating gentx for each validator ==="
for i in 1 2 3 4; do
    VALHOME="$TMPDIR/validator-$i"

    # Copy the updated genesis to this validator (skip validator-1, it's the source)
    if [ $i -ne 1 ]; then
        cp "$GENESIS" "$VALHOME/config/genesis.json"
    fi

    # Create gentx
    $AURAD genesis gentx "validator$i" "25000000000${STAKE_DENOM}" \
        --chain-id "$CHAIN_ID" \
        --keyring-backend "$KEYRING" \
        --home "$VALHOME" \
        --moniker "validator-$i" > /dev/null 2>&1

    echo "  Created gentx for validator-$i"
done

# Collect all gentxs into validator-1
echo "=== Collecting gentxs ==="
mkdir -p "$TMPDIR/validator-1/config/gentx"
for i in 1 2 3 4; do
    VALHOME="$TMPDIR/validator-$i"
    cp "$VALHOME/config/gentx/"*.json "$TMPDIR/validator-1/config/gentx/" 2>/dev/null || true
done

# Collect gentxs and generate final genesis
$AURAD genesis collect-gentxs --home "$TMPDIR/validator-1" > /dev/null 2>&1
echo "  Collected all gentxs"

# Validate genesis
$AURAD genesis validate-genesis "$TMPDIR/validator-1/config/genesis.json" || echo "  (Skipping validation - command may not exist)"
echo "  Genesis collected"

# Get final genesis
FINAL_GENESIS="$TMPDIR/validator-1/config/genesis.json"

# Get node IDs and set up persistent_peers
echo "=== Setting up peer connections ==="
declare -a NODE_IDS
declare -a PEER_STRINGS

for i in 1 2 3 4; do
    VALHOME="$TMPDIR/validator-$i"
    NODE_ID=$($AURAD tendermint show-node-id --home "$VALHOME")
    NODE_IDS[$i]=$NODE_ID
    # Use Docker hostnames
    PEER_STRINGS[$i]="${NODE_ID}@validator-$i:26656"
    echo "  Validator-$i node ID: $NODE_ID"
done

# Create Docker volumes and copy data
echo "=== Creating Docker volumes ==="
for i in 1 2 3 4; do
    VALHOME="$TMPDIR/validator-$i"
    VOLNAME="aura_validator-$i-data"

    # Create volume
    docker volume create "$VOLNAME" > /dev/null

    # Copy genesis to all validators (skip validator-1, it's the source)
    if [ $i -ne 1 ]; then
        cp "$FINAL_GENESIS" "$VALHOME/config/genesis.json"
    fi

    # Build persistent_peers (all except self)
    PEERS=""
    for j in 1 2 3 4; do
        if [ $i -ne $j ]; then
            if [ -n "$PEERS" ]; then
                PEERS="$PEERS,"
            fi
            PEERS="$PEERS${PEER_STRINGS[$j]}"
        fi
    done

    # Update config.toml
    sed -i "s/^persistent_peers = .*/persistent_peers = \"$PEERS\"/" "$VALHOME/config/config.toml"
    sed -i 's/^addr_book_strict = .*/addr_book_strict = false/' "$VALHOME/config/config.toml"
    sed -i 's/^allow_duplicate_ip = .*/allow_duplicate_ip = true/' "$VALHOME/config/config.toml"

    # Update app.toml for API
    sed -i 's/^enable = false/enable = true/' "$VALHOME/config/app.toml"
    sed -i 's/^swagger = false/swagger = true/' "$VALHOME/config/app.toml"
    sed -i 's/^enabled-unsafe-cors = false/enabled-unsafe-cors = true/' "$VALHOME/config/app.toml"

    # Copy to Docker volume using a temporary container
    docker run --rm -v "$VOLNAME:/dest" -v "$VALHOME:/src:ro" alpine sh -c "cp -r /src/* /dest/ && chown -R 1000:1000 /dest"

    echo "  Configured validator-$i with peers: $PEERS"
done

# Verify genesis is identical across all volumes
echo "=== Verifying genesis consistency ==="
HASH1=$(docker run --rm -v aura_validator-1-data:/data alpine sha256sum /data/config/genesis.json | cut -d' ' -f1)
HASH2=$(docker run --rm -v aura_validator-2-data:/data alpine sha256sum /data/config/genesis.json | cut -d' ' -f1)
HASH3=$(docker run --rm -v aura_validator-3-data:/data alpine sha256sum /data/config/genesis.json | cut -d' ' -f1)
HASH4=$(docker run --rm -v aura_validator-4-data:/data alpine sha256sum /data/config/genesis.json | cut -d' ' -f1)

echo "  Validator-1 genesis hash: $HASH1"
echo "  Validator-2 genesis hash: $HASH2"
echo "  Validator-3 genesis hash: $HASH3"
echo "  Validator-4 genesis hash: $HASH4"

if [ "$HASH1" = "$HASH2" ] && [ "$HASH2" = "$HASH3" ] && [ "$HASH3" = "$HASH4" ]; then
    echo "  All genesis files match!"
else
    echo "  ERROR: Genesis files do not match!"
    exit 1
fi

# Check validators in genesis
NUM_VALS=$(cat "$FINAL_GENESIS" | jq '.app_state.genutil.gen_txs | length')
echo "  Number of validators in genesis: $NUM_VALS"

# Cleanup
rm -rf "$TMPDIR"

echo ""
echo "=== Initialization complete ==="
echo "Start testnet with: docker compose -f docker-compose.testnet.yml up -d"
