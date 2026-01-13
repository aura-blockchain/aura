#!/usr/bin/env bash
# One-command joiner for Aura testnet full or pruned light nodes.
# Configures genesis, peers, gas prices, pruning, and optional state sync.

set -euo pipefail

CHAIN_ID="${CHAIN_ID:-aura-mvp-1}"
MONIKER="${MONIKER:-aura-$(hostname)}"
AURA_HOME="${AURA_HOME:-$HOME/.aura}"
RPC_ENDPOINT="${RPC_ENDPOINT:-https://rpc.aura-testnet.com}"
REST_ENDPOINT="${REST_ENDPOINT:-https://api.aura-testnet.com}"
MODE="${MODE:-full}"           # full | light
STATE_SYNC="${STATE_SYNC:-true}"
PEER_LIMIT="${PEER_LIMIT:-8}"
SEEDS="${SEEDS:-}"
PERSISTENT_PEERS="${PERSISTENT_PEERS:-}"
GAS_PRICES="${GAS_PRICES:-0.025uaura}"
TRUST_OFFSET="${TRUST_OFFSET:-200}"
SNAPSHOT_URL="${SNAPSHOT_URL:-}"

command -v aurad >/dev/null || { echo "aurad not found in PATH"; exit 1; }
command -v jq >/dev/null || { echo "jq is required"; exit 1; }

echo ">>> Chain: ${CHAIN_ID}"
echo ">>> Mode: ${MODE} (state sync: ${STATE_SYNC})"
echo ">>> Home: ${AURA_HOME}"
echo ">>> RPC: ${RPC_ENDPOINT}"

mkdir -p "${AURA_HOME}"

# Init if needed
if [ ! -f "${AURA_HOME}/config/genesis.json" ]; then
  echo "Initializing home directory..."
  aurad init "${MONIKER}" --chain-id "${CHAIN_ID}" --home "${AURA_HOME}" >/dev/null
fi

# Download genesis from RPC unless provided
if [ -n "${GENESIS_URL:-}" ]; then
  echo "Fetching genesis from ${GENESIS_URL}"
  curl -fsSL "${GENESIS_URL}" -o "${AURA_HOME}/config/genesis.json"
else
  echo "Fetching genesis from ${RPC_ENDPOINT}/genesis"
  curl -fsSL "${RPC_ENDPOINT}/genesis" | jq -r '.result.genesis' > "${AURA_HOME}/config/genesis.json"
fi

CONFIG="${AURA_HOME}/config/config.toml"
APP_CONFIG="${AURA_HOME}/config/app.toml"

# Build persistent peers automatically if not provided
if [ -z "${PERSISTENT_PEERS}" ]; then
  echo "Building persistent peer list from ${RPC_ENDPOINT}/net_info (limit ${PEER_LIMIT})"
  PERSISTENT_PEERS=$(curl -fsSL "${RPC_ENDPOINT}/net_info" \
    | jq -r --argjson limit "${PEER_LIMIT}" '
      .result.peers
      | map(
          .node_info.id as $id
          | .remote_ip as $rip
          | .node_info.listen_addr as $addr
          | ($addr|sub("^tcp://";"")|split(":")) as $parts
          | "\($id)@\(($rip // $parts[0])):\($parts[1])"
        )
      | .[0:$limit]
      | join(",")
    ')
fi

echo "Seeds: ${SEEDS:-<none>}"
echo "Peers: ${PERSISTENT_PEERS:-<auto-empty>}"

sed -i.bak -E "s|^seeds = \".*\"|seeds = \"${SEEDS}\"|" "${CONFIG}"
sed -i.bak -E "s|^persistent_peers = \".*\"|persistent_peers = \"${PERSISTENT_PEERS}\"|" "${CONFIG}"
sed -i.bak -E "s|^minimum-gas-prices = \".*\"|minimum-gas-prices = \"${GAS_PRICES}\"|" "${APP_CONFIG}"

# Pruning profile
if [ "${MODE}" = "light" ]; then
  sed -i.bak -E 's|^pruning *=.*|pruning = "custom"|' "${APP_CONFIG}"
  sed -i.bak -E 's|^pruning-keep-recent *=.*|pruning-keep-recent = "1000"|' "${APP_CONFIG}"
  sed -i.bak -E 's|^pruning-keep-every *=.*|pruning-keep-every = "0"|' "${APP_CONFIG}"
  sed -i.bak -E 's|^pruning-interval *=.*|pruning-interval = "50"|' "${APP_CONFIG}"
else
  sed -i.bak -E 's|^pruning *=.*|pruning = "default"|' "${APP_CONFIG}"
fi

# State sync configuration
if [ "${STATE_SYNC}" = "true" ]; then
  echo "Configuring state sync..."
  LATEST_HEIGHT=$(curl -fsSL "${RPC_ENDPOINT}/block" | jq -r '.result.block.header.height')
  TRUST_HEIGHT=$((LATEST_HEIGHT - TRUST_OFFSET))
  TRUST_HASH=$(curl -fsSL "${RPC_ENDPOINT}/block?height=${TRUST_HEIGHT}" | jq -r '.result.block_id.hash')
  RPC_SERVERS="${RPC_ENDPOINT},${RPC_ENDPOINT}"

  sed -i.bak -E "s|^enable *=.*|enable = true|" "${CONFIG}"
  sed -i.bak -E "s|^rpc_servers *=.*|rpc_servers = \"${RPC_SERVERS}\"|" "${CONFIG}"
  sed -i.bak -E "s|^trust_height *=.*|trust_height = ${TRUST_HEIGHT}|" "${CONFIG}"
  sed -i.bak -E "s|^trust_hash *=.*|trust_hash = \"${TRUST_HASH}\"|" "${CONFIG}"
  sed -i.bak -E "s|^trust_period *=.*|trust_period = \"168h0m0s\"|" "${CONFIG}"
else
  sed -i.bak -E "s|^enable *=.*|enable = false|" "${CONFIG}"
fi

# Optional snapshot restore
if [ -n "${SNAPSHOT_URL}" ]; then
  echo "Downloading snapshot from ${SNAPSHOT_URL}"
  curl -fsSL "${SNAPSHOT_URL}" | lz4 -d | tar -C "${AURA_HOME}" -xf -
fi

echo ">>> Done."
echo "Start the node:"
echo "  aurad start --home \"${AURA_HOME}\" --chain-id \"${CHAIN_ID}\""
echo "Monitor sync:"
echo "  curl -s ${RPC_ENDPOINT}/status | jq '.result.sync_info'"
