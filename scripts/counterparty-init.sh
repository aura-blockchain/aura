#!/bin/bash
# ============================================================================
# Aura Counterparty Chain Initializer
# ============================================================================
# Seeds a single-validator Aura chain (default: aura-counter-1) that acts as
# the local IBC counterparty for Hermes. Generates keys, funds the validator,
# and configures RPC/API/gRPC to listen on all interfaces so Docker Compose
# services (docker-compose.counterparty.yml) can expose the ports.
# ============================================================================

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${REPO_ROOT}/chain/aurad"

CHAIN_ID="${COUNTERPARTY_CHAIN_ID:-aura-counter-1}"
MONIKER="${COUNTERPARTY_MONIKER:-aura-counter}"
KEY_NAME="${COUNTERPARTY_KEY_NAME:-counterparty}"
DENOM="${COUNTERPARTY_DENOM:-stake}"
GENESIS_TOKENS="${COUNTERPARTY_GENESIS_TOKENS:-1000000000000}"
SELF_DELEGATION_TOKENS="${COUNTERPARTY_SELF_DELEGATION_TOKENS:-900000000000}"
KEYRING_BACKEND="${AURA_KEYRING_BACKEND:-test}"
KEY_PASSWORD="${AURA_KEY_PASSWORD:-password123}"
TESTNET_DIR="${REPO_ROOT}/testnet-data"
CHAIN_HOME="${TESTNET_DIR}/${MONIKER}"
FORCE_REINIT="${COUNTERPARTY_FORCE_REINIT:-0}"

log() {
  printf '[counterparty-init] %s\n' "$*"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing dependency: $1" >&2
    exit 1
  fi
}

require_cmd jq

if [ ! -x "${BINARY}" ]; then
  log "aurad binary not found, building..."
  if ! command -v go >/dev/null 2>&1; then
    echo "Go toolchain is required to build aurad" >&2
    exit 1
  fi
  (cd "${REPO_ROOT}/chain" && go build -o aurad ./cmd/aurad)
fi

if [ -d "${CHAIN_HOME}" ]; then
  if [ "${FORCE_REINIT}" != "1" ]; then
    echo "Chain home ${CHAIN_HOME} already exists. Set COUNTERPARTY_FORCE_REINIT=1 to overwrite." >&2
    exit 1
  fi
  log "Removing existing chain home at ${CHAIN_HOME}"
  rm -rf "${CHAIN_HOME}"
fi

mkdir -p "${CHAIN_HOME}"

log "Initializing ${CHAIN_ID} (${MONIKER})..."
"${BINARY}" init "${MONIKER}" --chain-id "${CHAIN_ID}" --home "${CHAIN_HOME}"

log "Creating validator key (${KEY_NAME})..."
KEY_INFO="$(printf '%s\n' "${KEY_PASSWORD}" | "${BINARY}" keys add "${KEY_NAME}" \
  --home "${CHAIN_HOME}" \
  --keyring-backend "${KEYRING_BACKEND}" \
  --output json 2>&1)"
echo "${KEY_INFO}" > "${CHAIN_HOME}/${KEY_NAME}.json"

VALIDATOR_ADDR="$(printf '%s' "${KEY_INFO}" | jq -r '.address')"

if [ -z "${VALIDATOR_ADDR}" ] || [ "${VALIDATOR_ADDR}" = "null" ]; then
  echo "Failed to parse validator address for key ${KEY_NAME}" >&2
  exit 1
fi

GENESIS_FILE="${CHAIN_HOME}/config/genesis.json"

if command -v jq >/dev/null 2>&1; then
  jq '.app_state.security.params.network.rate_limit.max_requests_per_second = "200"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.rate_limit.burst_size = "400"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.rate_limit.window_duration = "10s"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.rate_limit.ban_duration = "60s"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.connection.max_inbound_connections = 100' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.connection.max_outbound_connections = 100' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.connection.max_connections_per_ip = 20' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.connection.connection_timeout = "5s"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.mempool.max_size = "5000"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.mempool.max_bytes = "200000000"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.network.mempool.min_priority_fee = "0"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.validator.signed_blocks_window = "1000"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.crypto.min_threshold_participants = 2' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.privacy.min_ring_size = 3' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.privacy.max_ring_size = 16' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.privacy.min_mixing_participants = 3' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.security.params.privacy.mixing_fee = "1000"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
  jq '.app_state.crisis.constant_fee.denom = "'"${DENOM}"'"' "${GENESIS_FILE}" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "${GENESIS_FILE}"
else
  log "jq not available; using default security parameters (may fail validation)"
fi

log "Adding genesis account (${GENESIS_TOKENS}${DENOM})..."
"${BINARY}" genesis add-genesis-account "${VALIDATOR_ADDR}" "${GENESIS_TOKENS}${DENOM}" \
  --home "${CHAIN_HOME}" \
  --keyring-backend "${KEYRING_BACKEND}"

log "Generating gentx (${SELF_DELEGATION_TOKENS}${DENOM})..."
printf '%s\n' "${KEY_PASSWORD}" | "${BINARY}" genesis gentx "${KEY_NAME}" \
  "${SELF_DELEGATION_TOKENS}${DENOM}" \
  --home "${CHAIN_HOME}" \
  --chain-id "${CHAIN_ID}" \
  --keyring-backend "${KEYRING_BACKEND}" >/dev/null

log "Collecting gentx..."
"${BINARY}" genesis collect-gentxs --home "${CHAIN_HOME}" >/dev/null

APP_TOML="${CHAIN_HOME}/config/app.toml"
CONFIG_TOML="${CHAIN_HOME}/config/config.toml"

log "Configuring app.toml (API/gRPC/gas prices)..."
python3 <<PY
import pathlib, re

app = pathlib.Path("${APP_TOML}")
text = app.read_text()
# Ensure API + telemetry sections updated without clobbering comments
text = re.sub(r'(\[api\][\s\S]*?address\s*=\s*)".*"', r'\1"tcp://0.0.0.0:1317"', text)
text = re.sub(r'(\[api\][\s\S]*?enable\s*=\s*)\w+', r'\1true', text)
text = re.sub(r'(enabled-unsafe-cors\s*=\s*)\w+', r'\1true', text)
text = re.sub(r'(\[grpc\][\s\S]*?address\s*=\s*)".*"', r'\1"0.0.0.0:9090"', text)
text = re.sub(r'(\[grpc\][\s\S]*?enable\s*=\s*)\w+', r'\1true', text)
text = re.sub(r'(\[grpc-web\][\s\S]*?address\s*=\s*)".*"', r'\1"0.0.0.0:9091"', text)
text = re.sub(r'(\[grpc-web\][\s\S]*?enable\s*=\s*)\w+', r'\1true', text)
text = re.sub(r'^minimum-gas-prices\s*=.*$', 'minimum-gas-prices = "0.025${DENOM}"', text, flags=re.MULTILINE)
app.write_text(text)
PY

log "Configuring config.toml (RPC/P2P)..."
python3 <<PY
import pathlib, re

cfg = pathlib.Path("${CONFIG_TOML}")
text = cfg.read_text()
text = re.sub(r'(\[rpc\][\s\S]*?laddr\s*=\s*)"[^"]+"', r'\1"tcp://0.0.0.0:26657"', text)
text = re.sub(r'(\[rpc\][\s\S]*?cors_allowed_origins\s*=\s*)\[.*?\]', r'\1["*"]', text)
text = re.sub(r'(\[p2p\][\s\S]*?addr_book_strict\s*=\s*)\w+', r'\1false', text)
text = re.sub(r'(\[p2p\][\s\S]*?allow_duplicate_ip\s*=\s*)\w+', r'\1true', text)
cfg.write_text(text)
PY

cat <<EOF > "${CHAIN_HOME}/keys/README.txt"
Validator Key
=============
Name: ${KEY_NAME}
Address: ${VALIDATOR_ADDR}
Home: ${CHAIN_HOME}
EOF

log "Counterparty chain initialized at ${CHAIN_HOME}"
cat <<EOF

Next steps:
  1. Start the counterparty container:
         docker compose -f docker-compose.counterparty.yml up -d aura-counter
  2. Import the Hermes relayer key for ${CHAIN_ID} (key name suggestion: relayer-counter).
  3. Fund the Hermes key from the counterparty validator using:
         CHAIN_ID=${CHAIN_ID} RELAYER_CHAIN_ID=${CHAIN_ID} DENOM=${DENOM} \\
         VALIDATOR_KEY_NAME=${KEY_NAME} VALIDATOR_CONTAINER=aura-counter \\
         RELAYER_KEY_NAME=relayer-counter ./scripts/hermes-fund-keys.sh --amount 500000000000
EOF
