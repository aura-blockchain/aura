#!/usr/bin/env bash
# Smoke-test a running CometBFT light-client proxy.
# Defaults: LC_RPC=http://localhost:8888 (matches docker-compose.light-client.yml)
# Optional: ADDRESS=<bech32> to fetch a bank balance proof.

set -euo pipefail

LC_RPC="${LC_RPC:-http://localhost:8888}"

if ! command -v curl >/dev/null; then
  echo "curl is required" >&2
  exit 1
fi

if ! command -v jq >/dev/null; then
  echo "jq is required" >&2
  exit 1
fi

echo "Light client endpoint: ${LC_RPC}"

echo "Checking /status ..."
curl -fsSL "${LC_RPC}/status" | jq '.result.sync_info | {latest_block_height, catching_up}'

echo "Checking latest block ..."
curl -fsSL "${LC_RPC}/block" | jq '.result.block.header | {height, time, hash:.last_block_id.hash}'

if [ -n "${ADDRESS:-}" ]; then
  echo "Querying bank balance proof for ${ADDRESS} ..."
  KEY=$(aurad debug addr "${ADDRESS}" 2>/dev/null | jq -r '.address_bytes' || true)
  if [ -z "${KEY}" ] || [ "${KEY}" = "null" ]; then
    echo "Could not derive address bytes with aurad; skipping proof query." >&2
  else
    curl -fsSL "${LC_RPC}/abci_query?path=%22/store/bank/key%22&data=0x${KEY}&prove=true" | jq .
  fi
else
  echo "ADDRESS not set; skipping bank balance proof."
fi

echo "Done."
