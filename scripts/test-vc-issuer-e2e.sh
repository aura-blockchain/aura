#!/usr/bin/env bash
# End-to-end vc-issuer deployment and exercise against a fresh local Aura node.
# Spins up an ephemeral chain home, uploads the optimized wasm artifact, instantiates
# the contract, registers an issuer, requests a VC, fulfills it, and verifies the
# issued record via query. Fails fast with actionable diagnostics when any step
# does not return expected data (useful for catching wasm integration gaps).

set -euo pipefail
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${AURA_BINARY:-aurad}"
CHAIN_ID="${AURA_CHAIN_ID:-aura-vc-e2e-1}"
HOME_DIR="$(mktemp -d "${TMPDIR:-/tmp}/aura-vc-e2e.XXXXXX")"
ARTIFACT="${ROOT_DIR}/contracts/artifacts/vc_issuer.wasm"
LOG_FILE="${HOME_DIR}/aurad.log"
TX_LOG_FILE="${HOME_DIR}/tx_operations.log"
GAS_LOG_FILE="${HOME_DIR}/gas_measurements.log"
LABEL="vc-issuer-e2e-$(date +%s)"
GRPC_ENABLE="${AURA_GRPC_ENABLE:-true}"

# Gas tracking for each operation
declare -A GAS_USED

port_free() {
  local port=$1
  if command -v lsof >/dev/null 2>&1; then
    ! lsof -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1
  else
    ! ss -ltn 2>/dev/null | awk '{print $4}' | grep -q ":${port}$"
  fi
}

pick_port() {
  local base=${1:-20000}
  for _ in $(seq 1 50); do
    local candidate
    if command -v shuf >/dev/null 2>&1; then
      candidate=$(shuf -i "${base}"-$((base + 20000)) -n1)
    else
      candidate=$((base + RANDOM % 20000))
    fi
    if port_free "${candidate}"; then
      echo "${candidate}"
      return 0
    fi
  done
  echo "${base}"
}

RPC_PORT="${AURA_RPC_PORT:-$(pick_port 26657)}"
P2P_PORT="${AURA_P2P_PORT:-$(pick_port 26656)}"
API_PORT="${AURA_API_PORT:-$(pick_port 1317)}"
GRPC_PORT="${AURA_GRPC_PORT:-$(pick_port 19090)}"
DENOM="${AURA_DENOM:-uaura}"

echo "Selected ports - RPC:${RPC_PORT} P2P:${P2P_PORT} API:${API_PORT} GRPC:${GRPC_PORT}"
if ! port_free "${RPC_PORT}" || ! port_free "${P2P_PORT}"; then
  echo "selected ports are already in use (RPC:${RPC_PORT} P2P:${P2P_PORT}); set AURA_RPC_PORT/AURA_P2P_PORT to free ports or stop the conflicting node." >&2
  exit 1
fi

cleanup() {
  # Always keep temp directory for debugging
  echo ""
  echo "============================================"
  echo "Test completed. Temp directory preserved at:"
  echo "${HOME_DIR}"
  echo "============================================"
  echo "Logs available:"
  echo "  - Node log: ${LOG_FILE}"
  echo "  - Transaction log: ${TX_LOG_FILE}"
  echo "  - Gas measurements: ${GAS_LOG_FILE}"
  echo ""

  if [[ -n "${AURAD_PID:-}" ]] && ps -p "${AURAD_PID}" >/dev/null 2>&1; then
    echo "Stopping aurad (PID ${AURAD_PID})..."
    kill "${AURAD_PID}" >/dev/null 2>&1 || true
    wait "${AURAD_PID}" 2>/dev/null || true
    echo "Node stopped."
  fi

  # Only cleanup if explicitly requested
  if [[ -n "${CLEANUP_VC_E2E:-}" ]]; then
    echo "CLEANUP_VC_E2E set; removing ${HOME_DIR}"
    rm -rf "${HOME_DIR}"
  fi
}
trap cleanup EXIT

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_cmd "${BINARY}"
require_cmd jq
require_cmd curl

if [[ ! -f "${ARTIFACT}" ]]; then
  echo "artifact not found: ${ARTIFACT}" >&2
  exit 1
fi

GENESIS_HELP="$("${BINARY}" genesis --help 2>/dev/null || true)"
if ! echo "${GENESIS_HELP}" | grep -q "add-genesis-account"; then
  echo "aurad genesis helpers (add-genesis-account/gentx) are unavailable; cannot auto-build a validator set." >&2
  echo "Install a build of aurad that includes standard genesis commands or inject a validator/allocations manually before rerunning." >&2
  exit 1
fi

echo "Using home: ${HOME_DIR}"
echo "Initializing chain..."
"${BINARY}" init vc-e2e --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" >/dev/null

# Update ports to avoid collisions with any running nodes on the host.
CONFIG_TOML="${HOME_DIR}/config/config.toml"
APP_TOML="${HOME_DIR}/config/app.toml"
CLIENT_TOML="${HOME_DIR}/config/client.toml"

if [[ -f "${CONFIG_TOML}" ]]; then
  sed -i.bak \
    -e "s#^laddr = \".*26657\"#laddr = \"tcp://127.0.0.1:${RPC_PORT}\"#" \
    -e "s#^pprof_laddr = \".*6060\"#pprof_laddr = \"localhost:0\"#" \
    -e "s#^laddr = \".*26656\"#laddr = \"tcp://0.0.0.0:${P2P_PORT}\"#" \
    "${CONFIG_TOML}"
fi

if [[ -f "${APP_TOML}" ]]; then
  sed -i.bak \
    -e "s#^address = \".*1317\"#address = \"tcp://127.0.0.1:${API_PORT}\"#" \
    -e "s#^address = \".*9090\"#address = \"localhost:${GRPC_PORT}\"#" \
    -e "s#^address = \".*9091\"#address = \"localhost:0\"#" \
    "${APP_TOML}"
fi

if [[ -f "${CLIENT_TOML}" ]]; then
  sed -i.bak \
    -e "s#^grpc-address = \".*\"#grpc-address = \"localhost:${GRPC_PORT}\"#" \
    -e "s#^node = \".*\"#node = \"tcp://127.0.0.1:${RPC_PORT}\"#" \
    "${CLIENT_TOML}"
  CLIENT_TOML_PATH="${CLIENT_TOML}" python3 - <<'PY'
from pathlib import Path
import re
import os

path = Path(os.environ["CLIENT_TOML_PATH"])
text = path.read_text()
sign_mode_line = 'sign-mode = "amino-json"'
if re.search(r"^sign-mode\\s*=\\s*.+$", text, flags=re.M):
    text = re.sub(r"^sign-mode\\s*=\\s*.+$", sign_mode_line, text, flags=re.M)
else:
    text = text.rstrip() + "\n" + sign_mode_line + "\n"
path.write_text(text)
PY
fi

BOND_DENOM=$(jq -r '.app_state.staking.params.bond_denom' "${HOME_DIR}/config/genesis.json")
DENOM="${AURA_DENOM:-${BOND_DENOM}}"
GAS_PRICES="${AURA_GAS_PRICES:-0.025${DENOM}}"

KEYRING_FLAGS=(--home "${HOME_DIR}" --keyring-backend test)
declare -A ACC_NUM
declare -A SEQ
ACCOUNT_SEQ_FLAGS=()

add_key() {
  local name=$1
  "${BINARY}" keys add "${name}" "${KEYRING_FLAGS[@]}" --output json >/dev/null 2>&1
  local addr
  addr=$("${BINARY}" keys show "${name}" "${KEYRING_FLAGS[@]}" --address 2>/dev/null)
  "${BINARY}" genesis add-genesis-account "${addr}" "500000000${DENOM}" --home "${HOME_DIR}" --append
  if ! jq -e --arg addr "${addr}" '.app_state.bank.balances[] | select(.address==$addr)' "${HOME_DIR}/config/genesis.json" >/dev/null; then
    echo "failed to inject genesis balance for ${name} (${addr})" >&2
    exit 1
  fi
}

add_key validator
add_key issuer
add_key subject

"${BINARY}" genesis gentx validator "100000000${DENOM}" --chain-id "${CHAIN_ID}" "${KEYRING_FLAGS[@]}" >/dev/null
"${BINARY}" genesis collect-gentxs --home "${HOME_DIR}" >/dev/null

# Remove collected gentxs to avoid signature verification in single-node dev runs.
jq '.app_state.genutil.gen_txs = []' "${HOME_DIR}/config/genesis.json" > "${HOME_DIR}/config/genesis.tmp" \
  && mv "${HOME_DIR}/config/genesis.tmp" "${HOME_DIR}/config/genesis.json"

echo "Starting aurad..."
"${BINARY}" start \
  --home "${HOME_DIR}" \
  --grpc.enable="${GRPC_ENABLE}" \
  --grpc.address "localhost:${GRPC_PORT}" \
  --api.address "tcp://127.0.0.1:${API_PORT}" \
  --rpc.address "tcp://127.0.0.1:${RPC_PORT}" \
  --p2p.address "tcp://0.0.0.0:${P2P_PORT}" \
  >"${LOG_FILE}" 2>&1 &
AURAD_PID=$!
sleep 1
if ! kill -0 "${AURAD_PID}" >/dev/null 2>&1; then
  echo "aurad exited immediately; see ${LOG_FILE}" >&2
  tail -n 60 "${LOG_FILE}" >&2 || true
  exit 1
fi

wait_for_height() {
  local attempts=$1
  for _ in $(seq 1 "${attempts}"); do
    local height
    height=$(curl -s "http://127.0.0.1:${RPC_PORT}/status" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "")
    if [[ -n "${height}" && "${height}" =~ ^[0-9]+$ && "${height}" -gt 5 ]]; then
      return 0
    fi
    if ! kill -0 "${AURAD_PID}" >/dev/null 2>&1; then
      echo "aurad process exited while waiting for blocks; see ${LOG_FILE}" >&2
      tail -n 60 "${LOG_FILE}" >&2 || true
      return 1
    fi
    sleep 1
  done
  return 1
}

if ! wait_for_height 40; then
  echo "node failed to start; tail ${LOG_FILE} for details" >&2
  exit 1
fi

sleep 2
echo "Node is producing blocks. Uploading contract..."
# TX_FLAGS without signing mode (will be added per-transaction to avoid conflicts)
TX_FLAGS=(--chain-id "${CHAIN_ID}" --node "http://127.0.0.1:${RPC_PORT}" --broadcast-mode sync --yes --gas 5000000 --output json --gas-prices "${GAS_PRICES}")

# Logging helper
log_tx() {
  local message="$1"
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] ${message}" | tee -a "${TX_LOG_FILE}"
}

# Gas tracking helper
log_gas() {
  local operation="$1"
  local gas_used="$2"
  local gas_wanted="$3"
  echo "${operation}: gas_used=${gas_used}, gas_wanted=${gas_wanted}" | tee -a "${GAS_LOG_FILE}"
  GAS_USED["${operation}"]="${gas_used}"
}

# Account state validation helper
validate_account_state() {
  local name=$1
  local expected_acc_num=$2
  local expected_seq=$3

  load_account_state "${name}"
  local actual_acc_num="${ACC_NUM[${name}]}"
  local actual_seq="${SEQ[${name}]}"

  if [[ "${actual_acc_num}" != "${expected_acc_num}" ]]; then
    log_tx "ERROR: Account number mismatch for ${name}: expected=${expected_acc_num}, actual=${actual_acc_num}"
    return 1
  fi

  if [[ "${actual_seq}" != "${expected_seq}" ]]; then
    log_tx "ERROR: Sequence mismatch for ${name}: expected=${expected_seq}, actual=${actual_seq}"
    return 1
  fi

  log_tx "✓ Account state validated for ${name}: account_number=${actual_acc_num}, sequence=${actual_seq}"
  return 0
}

load_account_state() {
  local name=$1
  local addr
  addr=$("${BINARY}" keys show "${name}" "${KEYRING_FLAGS[@]}" --address)
  local account_raw=""

  # Prefer CLI query (gRPC-backed) to get authoritative account metadata.
  account_raw=$("${BINARY}" query account "${addr}" --node "http://127.0.0.1:${RPC_PORT}" --chain-id "${CHAIN_ID}" --output json 2>/dev/null || true)
  if [[ -n "${account_raw}" && "${account_raw}" != "{}" ]]; then
    ACC_NUM["${name}"]=$(echo "${account_raw}" | jq -r '.account.base_account.account_number // .account.account_number // "0"')
    SEQ["${name}"]=$(echo "${account_raw}" | jq -r '.account.base_account.sequence // .account.sequence // "0"')
    echo "Loaded ${name} via CLI: account_number=${ACC_NUM[$name]}, sequence=${SEQ[$name]}"
    return
  fi

  # Query account info from running node via REST API (more reliable than gRPC)
  local rest_response=""
  rest_response=$(curl -s "http://127.0.0.1:${API_PORT}/cosmos/auth/v1beta1/accounts/${addr}" 2>/dev/null || true)

  if [[ -n "${rest_response}" ]]; then
    # Extract account number and sequence from REST response
    # The response format varies: check for both BaseAccount and ModuleAccount patterns
    local acc_num seq_num
    acc_num=$(echo "${rest_response}" | jq -r '.account.account_number // .account.base_account.account_number // empty' 2>/dev/null || true)
    seq_num=$(echo "${rest_response}" | jq -r '.account.sequence // .account.base_account.sequence // "0"' 2>/dev/null || true)

    if [[ -n "${acc_num}" && "${acc_num}" != "null" ]]; then
      ACC_NUM["${name}"]="${acc_num}"
      SEQ["${name}"]="${seq_num:-0}"
      echo "Loaded ${name} (${addr}): account_number=${acc_num}, sequence=${seq_num:-0}"
      return
    fi
  fi

  # Fallback: try CLI query
  account_raw=$("${BINARY}" query auth account "${addr}" --node "http://127.0.0.1:${RPC_PORT}" --chain-id "${CHAIN_ID}" -o json 2>/dev/null || true)
  if [[ -n "${account_raw}" && "${account_raw}" != "{}" ]]; then
    ACC_NUM["${name}"]=$(echo "${account_raw}" | jq -r '.account.base_account.account_number // .account.account_number // .base_account.account_number // "0"')
    SEQ["${name}"]=$(echo "${account_raw}" | jq -r '.account.base_account.sequence // .account.sequence // .base_account.sequence // "0"')
    echo "Loaded ${name} via CLI: account_number=${ACC_NUM[$name]}, sequence=${SEQ[$name]}"
    return
  fi

  # Final fallback: parse genesis (for accounts not yet on-chain)
  local genesis="${HOME_DIR}/config/genesis.json"
  ACC_NUM["${name}"]=$(jq -r --arg addr "${addr}" '.app_state.auth.accounts[] | select(.address==$addr or .base_account.address==$addr) | (.account_number // .base_account.account_number // "0")' "${genesis}" 2>/dev/null || echo "0")
  SEQ["${name}"]=0
  echo "Loaded ${name} from genesis: account_number=${ACC_NUM[$name]}, sequence=0"
}

next_seq_flags() {
  local name=$1
  # Always refresh account/sequence from the chain so signing never reuses stale values.
  load_account_state "${name}"
  local acc=${ACC_NUM["${name}"]:-0}
  local seq=${SEQ["${name}"]:-0}
  ACCOUNT_SEQ_FLAGS=(--account-number "${acc}" --sequence "${seq}")
  SEQ["${name}"]=$((seq + 1))
}

wait_for_tx() {
  local hash=$1
  local label=${2:-tx}
  if [[ -z "${hash}" || "${hash}" == "null" ]]; then
    log_tx "ERROR: missing tx hash for ${label}"
    return 1
  fi

  local lookup_hash="${hash}"
  if [[ "${lookup_hash}" != 0x* ]]; then
    lookup_hash="0x${lookup_hash}"
  fi

  log_tx "Waiting for tx ${label} (${hash})..."

  for _ in $(seq 1 30); do
    local resp
    resp=$(curl -s "http://127.0.0.1:${RPC_PORT}/tx?hash=${lookup_hash}" 2>/dev/null || true)
    local height code gas_used gas_wanted
    height=$(echo "${resp}" | jq -r '.result.height // ""' 2>/dev/null || true)
    code=$(echo "${resp}" | jq -r '.result.tx_result.code // ""' 2>/dev/null || true)
    gas_used=$(echo "${resp}" | jq -r '.result.tx_result.gas_used // "0"' 2>/dev/null || true)
    gas_wanted=$(echo "${resp}" | jq -r '.result.tx_result.gas_wanted // "0"' 2>/dev/null || true)

    if [[ -n "${height}" && "${height}" != "0" ]]; then
      if [[ -n "${code}" && "${code}" != "0" ]]; then
        local raw_log
        raw_log=$(echo "${resp}" | jq -r '.result.tx_result.log // ""' 2>/dev/null || true)
        log_tx "ERROR: tx ${label} failed with code ${code}: ${raw_log}"
        echo "${resp}" >&2
        return 1
      fi

      log_tx "✓ tx ${label} succeeded at height ${height}"
      log_gas "${label}" "${gas_used}" "${gas_wanted}"

      echo "${resp}"
      return 0
    fi
    sleep 1
  done

  log_tx "ERROR: tx ${label} (${hash}) not found after waiting"
  return 1
}

run_tx_and_wait() {
  local label=$1
  shift
  log_tx "Broadcasting tx: ${label}"
  local res hash
  res=$("$@" "${TX_FLAGS[@]}" --sign-mode amino-json 2>&1)
  hash=$(echo "${res}" | jq -r '.txhash // .tx_response.txhash // empty' 2>/dev/null || true)
  if [[ -z "${hash}" || "${hash}" == "null" ]]; then
    log_tx "ERROR: failed to broadcast ${label}"
    echo "Broadcast response: ${res}" >&2
    return 1
  fi
  wait_for_tx "${hash}" "${label}"
}

log_tx "=== Starting E2E Test Flow ==="
log_tx "Loading initial account state from running node..."
load_account_state validator
load_account_state issuer
load_account_state subject

# Store the initial account/sequence values for validation
VALIDATOR_ACC_NUM="${ACC_NUM[validator]}"
VALIDATOR_INIT_SEQ="${SEQ[validator]}"
ISSUER_ACC_NUM="${ACC_NUM[issuer]}"
ISSUER_INIT_SEQ="${SEQ[issuer]}"
SUBJECT_ACC_NUM="${ACC_NUM[subject]}"
SUBJECT_INIT_SEQ="${SEQ[subject]}"

log_tx "Initial account state: validator(${VALIDATOR_ACC_NUM},${VALIDATOR_INIT_SEQ}), issuer(${ISSUER_ACC_NUM},${ISSUER_INIT_SEQ}), subject(${SUBJECT_ACC_NUM},${SUBJECT_INIT_SEQ})"

# === STORE CONTRACT ===
log_tx ""
log_tx "=== STORE CONTRACT ==="

# Preflight: Validate validator account state before transaction
if ! validate_account_state validator "${VALIDATOR_ACC_NUM}" "${VALIDATOR_INIT_SEQ}"; then
  log_tx "ERROR: Preflight validation failed for validator before store"
  exit 1
fi

next_seq_flags validator
STORE_UNSIGNED="${HOME_DIR}/store_unsigned.json"
SIGNED_STORE_PATH="${HOME_DIR}/store_signed.json"

log_tx "Generating unsigned store transaction..."
if ! STORE_RES=$("${BINARY}" tx aura_wasm_security store "${ARTIFACT}" \
  --from validator \
  "${ACCOUNT_SEQ_FLAGS[@]}" \
  "${KEYRING_FLAGS[@]}" \
  "${TX_FLAGS[@]}" \
  --generate-only 2>&1); then
  log_tx "ERROR: Failed to generate unsigned store transaction"
  echo "${STORE_RES}" >&2
  exit 1
fi

echo "${STORE_RES}" > "${STORE_UNSIGNED}"
log_tx "Unsigned transaction saved to ${STORE_UNSIGNED}"

# Sign transaction with explicit amino-json mode
log_tx "Signing transaction with amino-json mode..."
if ! "${BINARY}" tx sign "${STORE_UNSIGNED}" \
  --from validator \
  "${KEYRING_FLAGS[@]}" \
  --sign-mode amino-json \
  "${ACCOUNT_SEQ_FLAGS[@]}" \
  --chain-id "${CHAIN_ID}" \
  --node "http://127.0.0.1:${RPC_PORT}" \
  --output json > "${SIGNED_STORE_PATH}" 2>&1; then
  log_tx "ERROR: Failed to sign store transaction"
  cat "${SIGNED_STORE_PATH}" >&2
  exit 1
fi

if [[ ! -s "${SIGNED_STORE_PATH}" ]]; then
  log_tx "ERROR: Signed store tx is empty"
  exit 1
fi

# Verify signing mode in the signed transaction
SIGNING_MODE=$(jq -r '.auth_info.signer_infos[0].mode_info.single.mode // "UNKNOWN"' "${SIGNED_STORE_PATH}")
log_tx "Signed transaction uses mode: ${SIGNING_MODE}"

if [[ "${SIGNING_MODE}" != "SIGN_MODE_LEGACY_AMINO_JSON" ]]; then
  log_tx "WARNING: Expected SIGN_MODE_LEGACY_AMINO_JSON but got ${SIGNING_MODE}"
  # Continue anyway - some SDK versions may accept this
fi

log_tx "Broadcasting signed transaction..."
if ! STORE_RES=$("${BINARY}" tx broadcast "${SIGNED_STORE_PATH}" \
  --node "http://127.0.0.1:${RPC_PORT}" \
  --broadcast-mode sync \
  --output json 2>&1); then
  log_tx "ERROR: Failed to broadcast store transaction"
  echo "${STORE_RES}" >&2
  exit 1
fi

STORE_HASH=$(echo "${STORE_RES}" | jq -r '.txhash // .tx_response.txhash // empty')
if [[ -z "${STORE_HASH}" || "${STORE_HASH}" == "null" ]]; then
  log_tx "ERROR: Failed to extract tx hash from broadcast response"
  echo "${STORE_RES}" >&2
  exit 1
fi

STORE_TX=$(wait_for_tx "${STORE_HASH}" "store") || exit 1

# Postflight: Validate validator sequence incremented
EXPECTED_SEQ=$((VALIDATOR_INIT_SEQ + 1))
if ! validate_account_state validator "${VALIDATOR_ACC_NUM}" "${EXPECTED_SEQ}"; then
  log_tx "ERROR: Postflight validation failed - sequence did not increment correctly"
  exit 1
fi
CODE_ID=$(echo "${STORE_TX}" | jq -r '
  (.tx_response.logs[]?.events[]? | select(.type=="store_code") | .attributes[]? | select(.key=="code_id") | .value) //
  (.result.tx_result.events[]? | select(.type=="store_code") | .attributes[]? | select(.key=="code_id") | .value)
' | head -n1)
if [[ -z "${CODE_ID}" || "${CODE_ID}" == "null" ]]; then
  log_tx "ERROR: Failed to extract code_id from store response"
  echo "${STORE_TX}" >&2
  exit 1
fi
log_tx "✓ Stored contract with code_id=${CODE_ID}"

# === INSTANTIATE CONTRACT ===
log_tx ""
log_tx "=== INSTANTIATE CONTRACT ==="

# Preflight: Validate validator sequence before instantiate
EXPECTED_SEQ=$((VALIDATOR_INIT_SEQ + 1))
if ! validate_account_state validator "${VALIDATOR_ACC_NUM}" "${EXPECTED_SEQ}"; then
  log_tx "ERROR: Preflight validation failed for validator before instantiate"
  exit 1
fi

VALIDATOR_ADDR=$("${BINARY}" keys show validator "${KEYRING_FLAGS[@]}" --address)
INIT_MSG=$(jq -n --arg admin "${VALIDATOR_ADDR}" '{admin: $admin}')
next_seq_flags validator
INST_TX=$(run_tx_and_wait "instantiate" "${BINARY}" tx aura_wasm_security instantiate "${CODE_ID}" "${INIT_MSG}" --label "${LABEL}" --admin "${VALIDATOR_ADDR}" --from validator "${ACCOUNT_SEQ_FLAGS[@]}" "${KEYRING_FLAGS[@]}")

# Postflight: Validate validator sequence incremented
EXPECTED_SEQ=$((VALIDATOR_INIT_SEQ + 2))
if ! validate_account_state validator "${VALIDATOR_ACC_NUM}" "${EXPECTED_SEQ}"; then
  log_tx "ERROR: Postflight validation failed after instantiate"
  exit 1
fi
CONTRACT_ADDR=$(echo "${INST_TX}" | jq -r '
  (.tx_response.logs[]?.events[]? | select(.type=="instantiate") | .attributes[]? | select(.key=="_contract_address" or .key=="contract") | .value) //
  (.result.tx_result.events[]? | select(.type=="instantiate") | .attributes[]? | select(.key=="_contract_address" or .key=="contract") | .value)
' | head -n1)
if [[ -z "${CONTRACT_ADDR}" || "${CONTRACT_ADDR}" == "null" ]]; then
  log_tx "ERROR: Failed to extract contract address from instantiate response"
  echo "${INST_TX}" >&2
  exit 1
fi
log_tx "✓ Instantiated contract at ${CONTRACT_ADDR}"

ISSUER_ADDR=$("${BINARY}" keys show issuer "${KEYRING_FLAGS[@]}" --address)
SUBJECT_ADDR=$("${BINARY}" keys show subject "${KEYRING_FLAGS[@]}" --address)

# === REGISTER ISSUER ===
log_tx ""
log_tx "=== REGISTER ISSUER ==="

# Preflight
EXPECTED_SEQ=$((VALIDATOR_INIT_SEQ + 2))
if ! validate_account_state validator "${VALIDATOR_ACC_NUM}" "${EXPECTED_SEQ}"; then
  log_tx "ERROR: Preflight validation failed for validator before register-issuer"
  exit 1
fi

next_seq_flags validator
run_tx_and_wait "register-issuer" "${BINARY}" tx aura_wasm_security execute "${CONTRACT_ADDR}" \
  "$(jq -n --arg issuer "${ISSUER_ADDR}" '{register_issuer:{issuer:$issuer,policy_id:"kyc-basic",daily_limit:2}}')" \
  --from validator "${ACCOUNT_SEQ_FLAGS[@]}" "${KEYRING_FLAGS[@]}"

# Postflight
EXPECTED_SEQ=$((VALIDATOR_INIT_SEQ + 3))
if ! validate_account_state validator "${VALIDATOR_ACC_NUM}" "${EXPECTED_SEQ}"; then
  log_tx "ERROR: Postflight validation failed after register-issuer"
  exit 1
fi

# === REQUEST VC ===
log_tx ""
log_tx "=== REQUEST VC ==="

# Preflight - this is subject's first transaction
if ! validate_account_state subject "${SUBJECT_ACC_NUM}" "${SUBJECT_INIT_SEQ}"; then
  log_tx "ERROR: Preflight validation failed for subject before request-vc"
  exit 1
fi

next_seq_flags subject
REQ_TX=$(run_tx_and_wait "request-vc" "${BINARY}" tx aura_wasm_security execute "${CONTRACT_ADDR}" \
  "$(jq -n --arg issuer "${ISSUER_ADDR}" --arg subject "${SUBJECT_ADDR}" '{request_vc:{issuer:$issuer,subject:$subject,vc_type:"kyc",metadata:"{\"tier\":\"gold\"}"}}')" \
  --from subject "${ACCOUNT_SEQ_FLAGS[@]}" "${KEYRING_FLAGS[@]}")

# Postflight
EXPECTED_SEQ=$((SUBJECT_INIT_SEQ + 1))
if ! validate_account_state subject "${SUBJECT_ACC_NUM}" "${EXPECTED_SEQ}"; then
  log_tx "ERROR: Postflight validation failed after request-vc"
  exit 1
fi
REQUEST_ID=$(echo "${REQ_TX}" | jq -r '
  (.tx_response.logs[]?.events[]? | select(.type=="wasm") | .attributes[]? | select(.key=="request_id") | .value) //
  (.result.tx_result.events[]? | select(.type=="wasm") | .attributes[]? | select(.key=="request_id") | .value)
' | head -n1)
if [[ -z "${REQUEST_ID}" || "${REQUEST_ID}" == "null" ]]; then
  log_tx "Request ID not found in events, querying pending_requests..."
  # Fallback: query pending requests for issuer and grab the first ID.
  PENDING_RAW=$("${BINARY}" query aura_wasm_security query-smart "${CONTRACT_ADDR}" \
    "$(jq -n --arg issuer "${ISSUER_ADDR}" '{pending_requests:{issuer:$issuer}}')" \
    --node "http://127.0.0.1:${RPC_PORT}" --chain-id "${CHAIN_ID}" --output json 2>/dev/null || true)
  if [[ -n "${PENDING_RAW}" ]]; then
    PENDING_DATA=$(echo "${PENDING_RAW}" | jq -r '.data // empty')
    if [[ -n "${PENDING_DATA}" ]]; then
      PENDING_DECODED=$(echo "${PENDING_DATA}" | base64 --decode 2>/dev/null || true)
      REQUEST_ID=$(echo "${PENDING_DECODED}" | jq -r '.requests[0].id // empty' 2>/dev/null || true)
    fi
  fi
fi
if [[ -z "${REQUEST_ID}" || "${REQUEST_ID}" == "null" ]]; then
  log_tx "ERROR: Failed to extract request_id from request response or pending_requests query"
  echo "tx response: ${REQ_TX}" >&2
  echo "pending_requests: ${PENDING_RAW:-<empty>}" >&2
  exit 1
fi
log_tx "✓ Captured request_id=${REQUEST_ID}"

# === FULFILL VC ===
log_tx ""
log_tx "=== FULFILL VC ==="

# Preflight - this is issuer's first transaction
if ! validate_account_state issuer "${ISSUER_ACC_NUM}" "${ISSUER_INIT_SEQ}"; then
  log_tx "ERROR: Preflight validation failed for issuer before fulfill-vc"
  exit 1
fi

CRED_BASE64=$(printf '{"credential":"demo","ts":%s}' "$(date +%s)" | base64 | tr -d '\n')
next_seq_flags issuer
run_tx_and_wait "fulfill-vc" "${BINARY}" tx aura_wasm_security execute "${CONTRACT_ADDR}" \
  "$(jq -n --arg id "${REQUEST_ID}" --arg cred "${CRED_BASE64}" '{fulfill_request:{request_id:$id,credential_base64:$cred}}')" \
  --from issuer "${ACCOUNT_SEQ_FLAGS[@]}" "${KEYRING_FLAGS[@]}"

# Postflight
EXPECTED_SEQ=$((ISSUER_INIT_SEQ + 1))
if ! validate_account_state issuer "${ISSUER_ACC_NUM}" "${EXPECTED_SEQ}"; then
  log_tx "ERROR: Postflight validation failed after fulfill-vc"
  exit 1
fi

# === QUERY CREDENTIALS ===
log_tx ""
log_tx "=== QUERY CREDENTIALS ==="

log_tx "Querying issued credentials for subject..."
QUERY_RAW=$("${BINARY}" query aura_wasm_security query-smart "${CONTRACT_ADDR}" \
  "$(jq -n --arg subject "${SUBJECT_ADDR}" '{issued_by_subject:{subject:$subject}}')" \
  --node "http://127.0.0.1:${RPC_PORT}" --chain-id "${CHAIN_ID}" -o json)

DATA_B64=$(echo "${QUERY_RAW}" | jq -r '.data // empty')
if [[ -z "${DATA_B64}" ]]; then
  log_tx "ERROR: Query returned empty data"
  echo "${QUERY_RAW}" >&2
  exit 1
fi

DECODED=$(echo "${DATA_B64}" | base64 --decode 2>/dev/null || true)
if [[ -z "${DECODED}" ]]; then
  log_tx "ERROR: Failed to decode wasm query data"
  echo "${QUERY_RAW}" >&2
  exit 1
fi

CREDS_COUNT=$(echo "${DECODED}" | jq '.credentials | length' 2>/dev/null || echo "0")
if [[ "${CREDS_COUNT}" -lt 1 ]]; then
  log_tx "ERROR: No issued credentials returned"
  echo "Decoded payload: ${DECODED}" >&2
  exit 1
fi

log_tx "✓ Found ${CREDS_COUNT} issued credential(s)"

# === FINAL SUMMARY ===
log_tx ""
log_tx "============================================"
log_tx "✅ VC-ISSUER E2E TEST COMPLETED SUCCESSFULLY"
log_tx "============================================"
log_tx ""
log_tx "Contract Address: ${CONTRACT_ADDR}"
log_tx "Code ID: ${CODE_ID}"
log_tx "Request ID: ${REQUEST_ID}"
log_tx "Credentials Count: ${CREDS_COUNT}"
log_tx ""
log_tx "Final Account States:"
log_tx "  - validator: seq ${VALIDATOR_INIT_SEQ} -> $((VALIDATOR_INIT_SEQ + 3)) (3 txs)"
log_tx "  - subject: seq ${SUBJECT_INIT_SEQ} -> $((SUBJECT_INIT_SEQ + 1)) (1 tx)"
log_tx "  - issuer: seq ${ISSUER_INIT_SEQ} -> $((ISSUER_INIT_SEQ + 1)) (1 tx)"
log_tx ""
log_tx "Gas Usage Summary:"
for op in store instantiate register-issuer request-vc fulfill-vc; do
  if [[ -n "${GAS_USED[$op]:-}" ]]; then
    log_tx "  - ${op}: ${GAS_USED[$op]} gas"
  fi
done

echo ""
echo "✅ vc-issuer flow completed successfully."
echo "Contract: ${CONTRACT_ADDR}"
echo "Request:  ${REQUEST_ID}"
echo ""
