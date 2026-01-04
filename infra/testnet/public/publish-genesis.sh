#!/usr/bin/env bash
set -euo pipefail

RPC_URL="${RPC_URL:-http://127.0.0.1:26657}"
OUTPUT_DIR="${OUTPUT_DIR:-/var/www/aura-status}"

mkdir -p "${OUTPUT_DIR}"

curl -fsSL "${RPC_URL}/genesis" | jq -r '.result.genesis' > "${OUTPUT_DIR}/genesis.json"
echo "Published genesis.json to ${OUTPUT_DIR}/genesis.json"
