#!/usr/bin/env bash
set -euo pipefail

NODE_ID="${NODE_ID:-}"
P2P_HOST="${P2P_HOST:-158.69.119.76}"
P2P_PORT="${P2P_PORT:-26656}"
EXTRA_SEEDS="${EXTRA_SEEDS:-}"
EXTRA_PEERS="${EXTRA_PEERS:-}"
OUTPUT_DIR="${OUTPUT_DIR:-/var/www/aura-status}"

if [ -z "${NODE_ID}" ]; then
  if command -v aurad >/dev/null 2>&1; then
    NODE_ID=$(aurad tendermint show-node-id --home "${AURA_HOME:-$HOME/.aura}")
  fi
fi

if [ -z "${NODE_ID}" ]; then
  echo "NODE_ID is required (set NODE_ID env or ensure aurad is installed)." >&2
  exit 1
fi

mkdir -p "${OUTPUT_DIR}"

PRIMARY_ENTRY="${NODE_ID}@${P2P_HOST}:${P2P_PORT}"

seed_entries=("${PRIMARY_ENTRY}")
peer_entries=("${PRIMARY_ENTRY}")

if [ -n "${EXTRA_SEEDS}" ]; then
  IFS=',' read -ra EXTRA_SEED_LIST <<< "${EXTRA_SEEDS}"
  for entry in "${EXTRA_SEED_LIST[@]}"; do
    trimmed=$(echo "${entry}" | xargs)
    if [ -n "${trimmed}" ]; then
      seed_entries+=("${trimmed}")
    fi
  done
fi

if [ -n "${EXTRA_PEERS}" ]; then
  IFS=',' read -ra EXTRA_PEER_LIST <<< "${EXTRA_PEERS}"
  for entry in "${EXTRA_PEER_LIST[@]}"; do
    trimmed=$(echo "${entry}" | xargs)
    if [ -n "${trimmed}" ]; then
      peer_entries+=("${trimmed}")
    fi
  done
elif [ -n "${EXTRA_SEEDS}" ]; then
  peer_entries+=("${seed_entries[@]:1}")
fi

seed_json=$(printf '%s\n' "${seed_entries[@]}" | jq -R . | jq -s .)
peer_json=$(printf '%s\n' "${peer_entries[@]}" | jq -R . | jq -s .)

cat > "${OUTPUT_DIR}/seed-nodes.json" <<EOF
{
  "chain_id": "aura-testnet-1",
  "seeds": ${seed_json},
  "peers": ${peer_json}
}
EOF

echo "Published seed-nodes.json to ${OUTPUT_DIR}/seed-nodes.json"
