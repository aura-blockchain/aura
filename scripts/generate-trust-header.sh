#!/usr/bin/env bash
# Generate a trusted header tuple (height/hash) for CometBFT light client setup.
# Usage:
#   RPC=https://rpc.aura-testnet.com ./scripts/generate-trust-header.sh
#   RPC=http://localhost:8080/rpc ./scripts/generate-trust-header.sh

set -euo pipefail

RPC="${RPC:-http://localhost:8080/rpc}"

if ! command -v curl >/dev/null; then
  echo "curl is required" >&2
  exit 1
fi

if ! command -v jq >/dev/null; then
  echo "jq is required" >&2
  exit 1
fi

echo "Fetching trusted header from ${RPC}/commit ..."

COMMIT_JSON=$(curl -fsSL "${RPC}/commit") || {
  echo "Failed to fetch /commit from ${RPC}" >&2
  echo "Ensure a reachable RPC is running (e.g., start proxy via 'docker-compose -f docker-compose.proxy.yml up -d' or point RPC=http://<node>:26657)" >&2
  exit 1
}

TRUST_HEIGHT=$(printf '%s' "${COMMIT_JSON}" | jq -r '.result.signed_header.header.height')
TRUST_HASH=$(printf '%s' "${COMMIT_JSON}" | jq -r '.result.signed_header.commit.block_id.hash')

if [ -z "${TRUST_HEIGHT}" ] || [ -z "${TRUST_HASH}" ] || [ "${TRUST_HEIGHT}" = "null" ] || [ "${TRUST_HASH}" = "null" ]; then
  echo "Could not parse trusted header; response was:" >&2
  echo "${COMMIT_JSON}" >&2
  exit 1
fi

cat <<EOF
TRUST_HEIGHT=${TRUST_HEIGHT}
TRUST_HASH=${TRUST_HASH}
# Example .env.light-client overrides:
# CHAIN_ID=aura-testnet-1
# PRIMARY_RPC=${RPC}
# WITNESS_RPC=${RPC}
EOF
