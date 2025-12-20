#!/bin/bash
# Initialize 3-validator testnet for Kubernetes
set -e

CHAIN_ID="aura-testnet-1"
DENOM="uaura"
STAKE_DENOM="uaura"
# Staking amount must be >= DefaultPowerReduction (~825 billion) for 1 voting power
STAKE_AMOUNT="1000000000000"  # 1 trillion uaura
ACCOUNT_BALANCE="10000000000000"  # 10 trillion uaura (enough for staking + fees)
KEYRING="test"
NAMESPACE="aura"
NUM_VALIDATORS=3

# Check for required tools
command -v kubectl >/dev/null 2>&1 || { echo "kubectl required"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq required"; exit 1; }

# Temporary directory for initialization
TMPDIR=$(mktemp -d)
echo "Working in $TMPDIR"
trap "rm -rf $TMPDIR" EXIT

# Build aurad if needed
cd /home/hudson/blockchain-projects/aura/chain
if [ ! -f aurad ]; then
    echo "Building aurad..."
    go build -o aurad ./cmd/aurad
fi
AURAD=$(pwd)/aurad

# Initialize all validators
echo "=== Initializing $NUM_VALIDATORS validators ==="
for i in $(seq 1 $NUM_VALIDATORS); do
    VALHOME="$TMPDIR/validator-$i"
    mkdir -p "$VALHOME"
    $AURAD init "aura-validator-$((i-1))" --chain-id "$CHAIN_ID" --home "$VALHOME" > /dev/null 2>&1
    $AURAD keys add "validator$i" --keyring-backend "$KEYRING" --home "$VALHOME" > /dev/null 2>&1
    ADDR=$($AURAD keys show "validator$i" -a --keyring-backend "$KEYRING" --home "$VALHOME")
    echo "  Validator $i address: $ADDR"
done

# Use validator-1's genesis as base
GENESIS="$TMPDIR/validator-1/config/genesis.json"

# Add genesis accounts
echo "=== Adding genesis accounts ==="
for i in $(seq 1 $NUM_VALIDATORS); do
    VALHOME="$TMPDIR/validator-$i"
    ADDR=$($AURAD keys show "validator$i" -a --keyring-backend "$KEYRING" --home "$VALHOME")
    $AURAD genesis add-genesis-account "$ADDR" "${ACCOUNT_BALANCE}${DENOM}" --home "$TMPDIR/validator-1" > /dev/null 2>&1
done

# Copy genesis with all accounts to all validators first
echo "=== Distributing genesis to all validators ==="
for i in $(seq 2 $NUM_VALIDATORS); do
    cp "$GENESIS" "$TMPDIR/validator-$i/config/genesis.json"
done

# Create gentx for each validator
echo "=== Creating gentx ==="
for i in $(seq 1 $NUM_VALIDATORS); do
    VALHOME="$TMPDIR/validator-$i"
    if $AURAD genesis gentx "validator$i" "${STAKE_AMOUNT}${STAKE_DENOM}" \
        --chain-id "$CHAIN_ID" --keyring-backend "$KEYRING" --home "$VALHOME" \
        --moniker "aura-validator-$((i-1))" 2>&1 | grep -v "^$"; then
        echo "  Created gentx for validator-$((i-1))"
    else
        echo "  ERROR: Failed to create gentx for validator-$((i-1))"
    fi
done

# Collect gentxs - copy from validators 2 and 3 to validator-1's gentx folder
echo "=== Collecting gentxs ==="
for i in $(seq 2 $NUM_VALIDATORS); do
    echo "  Copying gentx from validator-$i..."
    cp "$TMPDIR/validator-$i/config/gentx/"*.json "$TMPDIR/validator-1/config/gentx/" 2>&1 || echo "  WARNING: No gentx found for validator-$i"
done
echo "  Gentx files in collection:"
ls -1 "$TMPDIR/validator-1/config/gentx/"*.json 2>/dev/null | wc -l | xargs -I{} echo "    {} gentx files"
$AURAD genesis collect-gentxs --home "$TMPDIR/validator-1" 2>&1 | grep -v "^$" || true

FINAL_GENESIS="$TMPDIR/validator-1/config/genesis.json"

# Update genesis - keep default validator but update with correct pubkey and denom
echo "=== Updating genesis ==="
# Get validator-0's consensus pubkey
VAL0_PUBKEY=$(cat "$TMPDIR/validator-1/config/priv_validator_key.json" | jq -r '.pub_key.value')
VAL0_ADDR=$($AURAD keys show "validator1" -a --keyring-backend "$KEYRING" --home "$TMPDIR/validator-1")
VAL0_VALOPER=$($AURAD keys show "validator1" --bech val -a --keyring-backend "$KEYRING" --home "$TMPDIR/validator-1")

echo "  Validator-0 pubkey: $VAL0_PUBKEY"
echo "  Validator-0 address: $VAL0_ADDR"
echo "  Validator-0 valoper: $VAL0_VALOPER"

# Update the default validator with correct pubkey and set bond_denom
# Keep the gentxs for future use but work with the pre-initialized validator
jq --arg denom "$STAKE_DENOM" \
   --arg pubkey "$VAL0_PUBKEY" \
   --arg valoper "$VAL0_VALOPER" '
  .app_state.staking.params.bond_denom = $denom |
  .app_state.staking.validators[0].consensus_pubkey.key = $pubkey |
  .app_state.staking.validators[0].operator_address = $valoper |
  .consensus.validators[0].pub_key.value = $pubkey
' "$FINAL_GENESIS" > "$FINAL_GENESIS.tmp" && mv "$FINAL_GENESIS.tmp" "$FINAL_GENESIS"
echo "  Updated validator with correct keys, set bond_denom=$STAKE_DENOM"

# Get node IDs
echo "=== Getting node IDs ==="
declare -a NODE_IDS
for i in $(seq 1 $NUM_VALIDATORS); do
    VALHOME="$TMPDIR/validator-$i"
    NODE_ID=$($AURAD tendermint show-node-id --home "$VALHOME")
    NODE_IDS[$i]=$NODE_ID
    echo "  Validator-$i: $NODE_ID"
done

# Build persistent peers (using K8s headless service DNS)
PEERS=""
for i in $(seq 1 $NUM_VALIDATORS); do
    [ -n "$PEERS" ] && PEERS="$PEERS,"
    PEERS="${PEERS}${NODE_IDS[$i]}@aura-validator-$((i-1)).aura-validator-headless.${NAMESPACE}.svc.cluster.local:26656"
done
echo "Peers: $PEERS"

# Create Kubernetes resources
echo "=== Creating Kubernetes resources ==="

# Create genesis ConfigMap
kubectl create configmap aura-genesis \
    --from-file=genesis.json="$FINAL_GENESIS" \
    --namespace="$NAMESPACE" --dry-run=client -o yaml > "$TMPDIR/genesis-configmap.yaml"

# Create config ConfigMap with app.toml and config.toml templates
cat > "$TMPDIR/config.toml" << 'EOF'
proxy_app = "tcp://127.0.0.1:26658"
moniker = "aura-validator"
db_backend = "goleveldb"
db_dir = "data"
log_level = "info"
log_format = "plain"

[rpc]
laddr = "tcp://0.0.0.0:26657"
cors_allowed_origins = ["*"]
max_open_connections = 900

[p2p]
laddr = "tcp://0.0.0.0:26656"
persistent_peers = ""
addr_book_strict = false
allow_duplicate_ip = true
max_num_inbound_peers = 40
max_num_outbound_peers = 10

[mempool]
size = 5000
cache_size = 10000

[consensus]
timeout_propose = "3s"
timeout_prevote = "1s"
timeout_precommit = "1s"
timeout_commit = "5s"

[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
EOF

cat > "$TMPDIR/app.toml" << 'EOF'
minimum-gas-prices = "0uaura"
pruning = "default"

[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"
enabled-unsafe-cors = true

[grpc]
enable = true
address = "0.0.0.0:9090"

[grpc-web]
enable = true
address = "0.0.0.0:9091"

[telemetry]
enabled = true
prometheus-retention-time = 60
EOF

kubectl create configmap aura-config \
    --from-file=config.toml="$TMPDIR/config.toml" \
    --from-file=app.toml="$TMPDIR/app.toml" \
    --namespace="$NAMESPACE" --dry-run=client -o yaml > "$TMPDIR/config-configmap.yaml"

# Create secrets for each validator
for i in $(seq 1 $NUM_VALIDATORS); do
    VALHOME="$TMPDIR/validator-$i"
    kubectl create secret generic "aura-validator-keys-$((i-1))" \
        --from-file=priv_validator_key.json="$VALHOME/config/priv_validator_key.json" \
        --from-file=node_key.json="$VALHOME/config/node_key.json" \
        --namespace="$NAMESPACE" --dry-run=client -o yaml > "$TMPDIR/validator-$i-keys.yaml"
done

# Create combined peers secret
echo "$PEERS" > "$TMPDIR/peers.txt"
kubectl create secret generic aura-peers \
    --from-literal=persistent_peers="$PEERS" \
    --namespace="$NAMESPACE" --dry-run=client -o yaml > "$TMPDIR/peers-secret.yaml"

# Apply resources
echo "=== Applying Kubernetes resources ==="
kubectl apply -f "$TMPDIR/genesis-configmap.yaml"
kubectl apply -f "$TMPDIR/config-configmap.yaml"
kubectl apply -f "$TMPDIR/peers-secret.yaml"
for i in $(seq 1 $NUM_VALIDATORS); do
    kubectl apply -f "$TMPDIR/validator-$i-keys.yaml"
done

echo ""
echo "=== Resources created ==="
echo "  - ConfigMap: aura-genesis"
echo "  - ConfigMap: aura-config"
echo "  - Secret: aura-peers"
for i in $(seq 1 $NUM_VALIDATORS); do
    echo "  - Secret: aura-validator-keys-$((i-1))"
done
echo ""
echo "Now deploy the real StatefulSet with:"
echo "  kubectl apply -f k8s/testnet-deploy/statefulset-real.yaml"
