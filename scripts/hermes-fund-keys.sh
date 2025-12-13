#!/bin/bash
# ============================================================================
# Hermes Relayer Funding Helper
# ============================================================================
# Funds the local Hermes relayer account (default: relayer-aura on aura-local-4)
# from a validator key so ICS20 setup can proceed without manual CLI steps.
# The script looks up the relayer address via `hermes keys list`, sends tokens
# from the specified validator container, and verifies the resulting balance.
# ============================================================================

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

CHAIN_ID="${CHAIN_ID:-aura-local-4}"
RELAYER_CHAIN_ID="${RELAYER_CHAIN_ID:-${CHAIN_ID}}"
RELAYER_KEY_NAME="${RELAYER_KEY_NAME:-relayer-aura}"
DENOM="${DENOM:-uaura}"
RELAYER_AMOUNT="${RELAYER_AMOUNT:-500000000000}" # 500,000 AURA (in uaura)
KEYRING_BACKEND="${AURA_KEYRING_BACKEND:-test}"
KEY_PASSWORD="${AURA_KEY_PASSWORD:-password123}"
HERMES_CONFIG="${HERMES_CONFIG:-${REPO_ROOT}/config/hermes/config.toml}"
VALIDATOR_KEY_NAME="${VALIDATOR_KEY_NAME:-validator-1}"
VALIDATOR_CONTAINER="${VALIDATOR_CONTAINER:-aura-${VALIDATOR_KEY_NAME}}"
VALIDATOR_RPC="${VALIDATOR_RPC:-tcp://localhost:26657}"
LOG_DIR="${REPO_ROOT}/logs"
FUNDING_LOG="${LOG_DIR}/hermes-${RELAYER_KEY_NAME}-funding.json"

usage() {
  cat <<EOF
Hermes relayer funding helper

Usage:
  $(basename "$0") [--amount <uaura>] [--relayer-key <name>] [--validator <validator-1>]

Environment overrides:
  RELAYER_AMOUNT        Amount in base denom (default: ${RELAYER_AMOUNT})
  RELAYER_KEY_NAME      Hermes key name to fund (default: ${RELAYER_KEY_NAME})
  RELAYER_CHAIN_ID      Chain ID as configured in Hermes (default: ${RELAYER_CHAIN_ID})
  CHAIN_ID              Aura chain-id for aurad tx (default: ${CHAIN_ID})
  VALIDATOR_KEY_NAME    Local validator key to debit (default: ${VALIDATOR_KEY_NAME})
  VALIDATOR_CONTAINER   Docker container running the validator (default: ${VALIDATOR_CONTAINER})
  HERMES_CONFIG         Path to Hermes config (default: ${HERMES_CONFIG})
  KEY_PASSWORD          Password fed into aurad keyring (default: ${KEY_PASSWORD})
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --amount)
      RELAYER_AMOUNT="$2"
      shift 2
      ;;
    --relayer-key)
      RELAYER_KEY_NAME="$2"
      shift 2
      ;;
    --validator)
      VALIDATOR_KEY_NAME="$2"
      VALIDATOR_CONTAINER="aura-${VALIDATOR_KEY_NAME}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

command -v hermes >/dev/null 2>&1 || { echo "hermes binary not found" >&2; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq is required" >&2; exit 1; }

if [ ! -f "${HERMES_CONFIG}" ]; then
  echo "Hermes config not found at ${HERMES_CONFIG}" >&2
  exit 1
fi

if ! docker ps --format '{{.Names}}' | grep -Fxq "${VALIDATOR_CONTAINER}"; then
  echo "Validator container ${VALIDATOR_CONTAINER} is not running" >&2
  exit 1
fi

mkdir -p "${LOG_DIR}"

log() {
  printf '[hermes-fund] %s\n' "$*"
}

HERMES_BIN=(hermes --config "${HERMES_CONFIG}")

log "Fetching relayer address for key '${RELAYER_KEY_NAME}' on ${RELAYER_CHAIN_ID}..."
RELAYER_ADDRESS="$(
  "${HERMES_BIN[@]}" keys list --chain "${RELAYER_CHAIN_ID}" 2>/dev/null | \
  awk -v key="${RELAYER_KEY_NAME}" '
    $1 == "-" && $2 == key {
      match($0, /\(([^)]+)\)/, m);
      if (m[1] != "") {
        print m[1];
        exit
      }
    }
  '
)"

if [ -z "${RELAYER_ADDRESS}" ]; then
  echo "Failed to find relayer address for key '${RELAYER_KEY_NAME}'. Import the key via 'hermes keys add' first." >&2
  exit 1
fi

log "Relayer address: ${RELAYER_ADDRESS}"
log "Funding amount: ${RELAYER_AMOUNT}${DENOM}"

log "Fetching validator address for key '${VALIDATOR_KEY_NAME}'..."
VALIDATOR_ADDRESS="$(docker exec "${VALIDATOR_CONTAINER}" bash -c \
  "echo '${KEY_PASSWORD}' | aurad keys show ${VALIDATOR_KEY_NAME} --keyring-backend ${KEYRING_BACKEND} --address" \
  | tr -d '\r')"

if [ -z "${VALIDATOR_ADDRESS}" ]; then
  echo "Failed to obtain address for validator key '${VALIDATOR_KEY_NAME}'" >&2
  exit 1
fi

log "Validator address: ${VALIDATOR_ADDRESS}"

TX_CMD=$(cat <<EOF
echo '${KEY_PASSWORD}' | aurad tx bank send ${VALIDATOR_KEY_NAME} ${RELAYER_ADDRESS} ${RELAYER_AMOUNT}${DENOM} \
  --chain-id ${CHAIN_ID} \
  --node ${VALIDATOR_RPC} \
  --keyring-backend ${KEYRING_BACKEND} \
  --gas auto \
  --gas-adjustment 1.3 \
  --gas-prices 0.025${DENOM} \
  --broadcast-mode block \
  --yes \
  --output json
EOF
)

log "Broadcasting transaction from ${VALIDATOR_KEY_NAME} via ${VALIDATOR_CONTAINER}..."
docker exec "${VALIDATOR_CONTAINER}" bash -c "${TX_CMD}" | tee "${FUNDING_LOG}"

log "Waiting for balance query..."
sleep 6

RELAYER_BALANCE=$(docker exec "${VALIDATOR_CONTAINER}" aurad query bank balances "${RELAYER_ADDRESS}" \
  --chain-id "${CHAIN_ID}" \
  --node ${VALIDATOR_RPC} \
  --output json | jq -r ".balances[] | select(.denom==\"${DENOM}\") | .amount" || true)

if [ -z "${RELAYER_BALANCE}" ]; then
  echo "Failed to confirm relayer balance. Check ${FUNDING_LOG} for transaction details." >&2
  exit 1
fi

log "Relayer balance: ${RELAYER_BALANCE}${DENOM}"
log "Funding complete. Transaction saved to ${FUNDING_LOG}"
