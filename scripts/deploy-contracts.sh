#!/usr/bin/env bash
# Deploy CosmosWasm contracts to an Aura node using the custom wasm module.
# Uses optimized artifacts produced by `make optimize-wasm`.

set -euo pipefail
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Defaults (can be overridden via env vars or flags)
BINARY="${AURA_BINARY:-aurad}"
CHAIN_ID="${AURA_CHAIN_ID:-aura-local-1}"
NODE="${AURA_NODE:-http://localhost:26657}"
HOME_DIR="${AURA_HOME:-$HOME/.aura}"
KEYRING_BACKEND="${AURA_KEYRING:-test}"
FROM_KEY="${AURA_DEPLOYER:-validator}"
DENOM="${AURA_DENOM:-uaura}"
GAS_PRICES="${AURA_GAS_PRICES:-0.025${DENOM}}"
ARTIFACT="${ROOT_DIR}/contracts/artifacts/vc_issuer.wasm"
LABEL="vc-issuer-$(date +%s)"
ADMIN_ADDRESS=""
NO_ADMIN=false
STORE_ONLY=false
CODE_ID_OVERRIDE=""
FUNDS=""
INIT_JSON_OVERRIDE=""

usage() {
  cat <<EOF
Usage: $(basename "$0") [options]

Options:
  --artifact <path>       Path to optimized wasm artifact (default: ${ARTIFACT})
  --from <key>            Key name to sign transactions (default: ${FROM_KEY})
  --chain-id <id>         Chain ID (default: ${CHAIN_ID})
  --node <addr>           RPC endpoint (default: ${NODE})
  --home <path>           Aura home directory (default: ${HOME_DIR})
  --keyring <backend>     Keyring backend (default: ${KEYRING_BACKEND})
  --gas-prices <price>    Gas prices for transactions (default: ${GAS_PRICES})
  --label <label>         Label for instantiation (default: ${LABEL})
  --admin <address>       Admin address; defaults to signer address
  --no-admin              Instantiate without admin control
  --funds <amount>        Funds to send on instantiate (e.g., "10000${DENOM}")
  --init-json <json>      Override default instantiate message JSON
  --code-id <id>          Skip store step and instantiate existing code ID
  --store-only            Upload code without instantiating
  -h, --help              Show this help

Environment overrides:
  AURA_BINARY, AURA_CHAIN_ID, AURA_NODE, AURA_HOME, AURA_KEYRING,
  AURA_DEPLOYER, AURA_DENOM, AURA_GAS_PRICES
EOF
  exit 0
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Error: missing required command: $1" >&2
    exit 1
  fi
}

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --artifact) ARTIFACT="$2"; shift 2 ;;
    --from) FROM_KEY="$2"; shift 2 ;;
    --chain-id) CHAIN_ID="$2"; shift 2 ;;
    --node) NODE="$2"; shift 2 ;;
    --home) HOME_DIR="$2"; shift 2 ;;
    --keyring) KEYRING_BACKEND="$2"; shift 2 ;;
    --gas-prices) GAS_PRICES="$2"; shift 2 ;;
    --label) LABEL="$2"; shift 2 ;;
    --admin) ADMIN_ADDRESS="$2"; shift 2 ;;
    --no-admin) NO_ADMIN=true; shift 1 ;;
    --funds) FUNDS="$2"; shift 2 ;;
    --init-json) INIT_JSON_OVERRIDE="$2"; shift 2 ;;
    --code-id) CODE_ID_OVERRIDE="$2"; shift 2 ;;
    --store-only) STORE_ONLY=true; shift 1 ;;
    -h|--help) usage ;;
    *) echo "Unknown argument: $1" >&2; usage ;;
  esac
done

require_cmd "$BINARY"
require_cmd jq

if [[ ! -f "$ARTIFACT" ]]; then
  echo "Error: artifact not found at $ARTIFACT" >&2
  exit 1
fi

TX_FLAGS=(--from "$FROM_KEY" --home "$HOME_DIR" --chain-id "$CHAIN_ID" --node "$NODE" \
  --keyring-backend "$KEYRING_BACKEND" --yes --broadcast-mode block --output json \
  --gas auto --gas-adjustment 1.3)
if [[ -n "$GAS_PRICES" ]]; then
  TX_FLAGS+=(--gas-prices "$GAS_PRICES")
fi

echo "Using binary:        $BINARY"
echo "Chain ID:            $CHAIN_ID"
echo "Node:                $NODE"
echo "From key:            $FROM_KEY (keyring: $KEYRING_BACKEND)"
echo "Artifact:            $ARTIFACT"
echo "Label:               $LABEL"
echo "Gas prices:          $GAS_PRICES"

FROM_ADDRESS=$("$BINARY" keys show "$FROM_KEY" --home "$HOME_DIR" --keyring-backend "$KEYRING_BACKEND" --address)
if [[ -z "$FROM_ADDRESS" ]]; then
  echo "Error: failed to resolve address for key '$FROM_KEY'" >&2
  exit 1
fi
if [[ -z "$ADMIN_ADDRESS" && "$NO_ADMIN" = false ]]; then
  ADMIN_ADDRESS="$FROM_ADDRESS"
fi

CODE_ID="$CODE_ID_OVERRIDE"

if [[ -z "$CODE_ID" ]]; then
  echo "Uploading wasm bytecode..."
  STORE_RES=$("$BINARY" tx aura_wasm_security store "$ARTIFACT" "${TX_FLAGS[@]}")
  STORE_HASH=$(echo "$STORE_RES" | jq -r '.txhash')
  CODE_ID=$(echo "$STORE_RES" | jq -r '.logs[]?.events[]? | select(.type=="store_code") | .attributes[]? | select(.key=="code_id") | .value' | head -n1)
  if [[ -z "$CODE_ID" || "$CODE_ID" == "null" ]]; then
    echo "Error: failed to extract code_id from tx response" >&2
    echo "$STORE_RES" >&2
    exit 1
  fi
  echo "✓ Stored code. txhash: ${STORE_HASH}, code_id: ${CODE_ID}"
else
  echo "Using provided code_id: ${CODE_ID} (store step skipped)"
fi

if [[ "$STORE_ONLY" = true ]]; then
  echo "Store-only requested; skipping instantiation."
  exit 0
fi

if [[ -n "$INIT_JSON_OVERRIDE" ]]; then
  # Validate JSON
  if ! echo "$INIT_JSON_OVERRIDE" | jq -e . >/dev/null 2>&1; then
    echo "Error: provided --init-json is not valid JSON" >&2
    exit 1
  fi
  INIT_MSG="$INIT_JSON_OVERRIDE"
else
  # Default instantiate message for vc-issuer
  if [[ -n "$ADMIN_ADDRESS" ]]; then
    INIT_MSG=$(jq -n --arg admin "$ADMIN_ADDRESS" '{admin: $admin}')
  else
    INIT_MSG=$(jq -n '{admin: null}')
  fi
fi

INSTANTIATE_FLAGS=("${TX_FLAGS[@]}")
if [[ -n "$FUNDS" ]]; then
  INSTANTIATE_FLAGS+=(--amount "$FUNDS")
fi
if [[ "$NO_ADMIN" = true ]]; then
  INSTANTIATE_FLAGS+=(--no-admin)
elif [[ -n "$ADMIN_ADDRESS" ]]; then
  INSTANTIATE_FLAGS+=(--admin "$ADMIN_ADDRESS")
fi

echo "Instantiating contract..."
INST_RES=$("$BINARY" tx aura_wasm_security instantiate "$CODE_ID" "$INIT_MSG" --label "$LABEL" "${INSTANTIATE_FLAGS[@]}")
INST_HASH=$(echo "$INST_RES" | jq -r '.txhash')
CONTRACT_ADDR=$(echo "$INST_RES" | jq -r '.logs[]?.events[]? | select(.type=="instantiate") | .attributes[]? | select(.key=="_contract_address") | .value' | head -n1)

if [[ -z "$CONTRACT_ADDR" || "$CONTRACT_ADDR" == "null" ]]; then
  echo "Error: failed to extract contract address from instantiate response" >&2
  echo "$INST_RES" >&2
  exit 1
fi

echo "✓ Instantiated contract."
echo "  txhash:           ${INST_HASH}"
echo "  code_id:          ${CODE_ID}"
echo "  contract_address: ${CONTRACT_ADDR}"

echo ""
echo "Next steps:"
echo "  - Verify code:   $BINARY query aura_wasm_security code $CODE_ID --node $NODE --chain-id $CHAIN_ID"
echo "  - Contract info: $BINARY query aura_wasm_security contract $CONTRACT_ADDR --node $NODE --chain-id $CHAIN_ID"
echo "  - Query state:   $BINARY query aura_wasm_security contract-state-all $CONTRACT_ADDR --node $NODE --chain-id $CHAIN_ID"
