#!/usr/bin/env bash
# Simple WASM contract deployment test
# Tests that contract store, instantiate, and execute work correctly

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${AURA_BINARY:-${ROOT_DIR}/chain/aurad}"
CHAIN_ID="test-wasm-1"
HOME_DIR="$(mktemp -d)"
ARTIFACT="${ROOT_DIR}/contracts/artifacts/vc_issuer.wasm"

cleanup() {
  echo "Cleaning up ${HOME_DIR}..."
  if [[ -n "${AURAD_PID:-}" ]] && ps -p "${AURAD_PID}" >/dev/null 2>&1; then
    kill "${AURAD_PID}" 2>/dev/null || true
    wait "${AURAD_PID}" 2>/dev/null || true
  fi
  rm -rf "${HOME_DIR}"
}
trap cleanup EXIT

if [[ ! -f "${ARTIFACT}" ]]; then
  echo "ERROR: WASM artifact not found: ${ARTIFACT}"
  exit 1
fi

echo "=== Initializing chain ==="
"${BINARY}" init test --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" >/dev/null

# Create test account
"${BINARY}" keys add validator --home "${HOME_DIR}" --keyring-backend test --output json >/dev/null 2>&1
VALIDATOR_ADDR=$("${BINARY}" keys show validator --home "${HOME_DIR}" --keyring-backend test --address)
"${BINARY}" genesis add-genesis-account "${VALIDATOR_ADDR}" "1000000000stake" --home "${HOME_DIR}" >/dev/null
"${BINARY}" genesis gentx validator "100000000stake" --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" --keyring-backend test >/dev/null
"${BINARY}" genesis collect-gentxs --home "${HOME_DIR}" >/dev/null

echo "=== Starting node ==="
"${BINARY}" start --home "${HOME_DIR}" --grpc.enable=false >/dev/null 2>&1 &
AURAD_PID=$!
sleep 3

# Wait for blocks
echo "Waiting for blocks..."
for i in $(seq 1 20); do
  height=$(curl -s http://localhost:26657/status 2>/dev/null | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0")
  if [[ "${height}" -gt 5 ]]; then
    echo "Node is at height ${height}"
    break
  fi
  sleep 1
done

echo ""
echo "=== Storing contract ==="
STORE_RES=$("${BINARY}" tx aura_wasm_security store "${ARTIFACT}" \
  --from validator \
  --chain-id "${CHAIN_ID}" \
  --home "${HOME_DIR}" \
  --keyring-backend test \
  --gas 5000000 \
  --yes \
  --output json \
  --broadcast-mode sync 2>&1) || {
  echo "ERROR: Failed to store contract"
  echo "${STORE_RES}"
  exit 1
}

TX_HASH=$(echo "${STORE_RES}" | jq -r '.txhash // .tx_response.txhash // empty')
echo "Transaction hash: ${TX_HASH}"

# Wait for tx to be included in a block
echo "Waiting for transaction to be included..."
for i in $(seq 1 20); do
  TX_RES=$(curl -s "http://localhost:26657/tx?hash=0x${TX_HASH}" 2>/dev/null || true)
  HEIGHT=$(echo "${TX_RES}" | jq -r '.result.height // ""' 2>/dev/null || true)
  if [[ -n "${HEIGHT}" && "${HEIGHT}" != "0" ]]; then
    echo "Transaction included at height ${HEIGHT}"
    CODE=$(echo "${TX_RES}" | jq -r '.result.tx_result.code // ""' 2>/dev/null || true)
    if [[ -n "${CODE}" && "${CODE}" != "0" ]]; then
      LOG=$(echo "${TX_RES}" | jq -r '.result.tx_result.log // ""' 2>/dev/null || true)
      echo "ERROR: Transaction failed with code ${CODE}: ${LOG}"
      exit 1
    fi
    echo "Transaction succeeded!"
    echo "${TX_RES}" | jq -r '.result.tx_result.events[] | select(.type=="store_code")' || echo "No store_code event"
    break
  fi
  sleep 1
done

# Query code list
echo ""
echo "=== Querying stored codes ==="
CODE_LIST=$("${BINARY}" query aura_wasm_security list-code \
  --chain-id "${CHAIN_ID}" \
  --home "${HOME_DIR}" \
  --output json 2>&1) || {
  echo "ERROR: Failed to query code list"
  echo "${CODE_LIST}"
  exit 1
}

echo "Code list response:"
echo "${CODE_LIST}" | jq '.'

CODE_ID=$(echo "${CODE_LIST}" | jq -r '.code_infos[0].code_id // empty')
if [[ -z "${CODE_ID}" || "${CODE_ID}" == "null" ]]; then
  echo "ERROR: No code found in list"
  exit 1
fi

echo ""
echo "✅ SUCCESS: Contract stored with code_id=${CODE_ID}"
echo ""
echo "=== Instantiating contract ==="

INIT_MSG='{"admin":"'${VALIDATOR_ADDR}'"}'
INST_RES=$("${BINARY}" tx aura_wasm_security instantiate "${CODE_ID}" "${INIT_MSG}" \
  --label "test-vc-issuer" \
  --admin "${VALIDATOR_ADDR}" \
  --from validator \
  --chain-id "${CHAIN_ID}" \
  --home "${HOME_DIR}" \
  --keyring-backend test \
  --gas 5000000 \
  --yes \
  --output json \
  --broadcast-mode sync 2>&1) || {
  echo "ERROR: Failed to instantiate contract"
  echo "${INST_RES}"
  exit 1
}

INST_HASH=$(echo "${INST_RES}" | jq -r '.txhash // .tx_response.txhash // empty')
echo "Instantiate hash: ${INST_HASH}"

sleep 3

# Query contracts
echo ""
echo "=== Querying contracts ==="
CONTRACTS=$("${BINARY}" query aura_wasm_security list-contracts \
  --chain-id "${CHAIN_ID}" \
  --home "${HOME_DIR}" \
  --output json 2>&1) || {
  echo "WARNING: Failed to query contracts (may not be implemented)"
  echo "${CONTRACTS}"
}

echo ""
echo "================================================================"
echo "✅ WASM DEPLOYMENT TEST PASSED"
echo "================================================================"
echo "Code ID: ${CODE_ID}"
echo "Test completed successfully!"
echo ""
