#!/usr/bin/env bash
set -euo pipefail

ENDPOINTS=(
  "RPC https://testnet-rpc.aurablockchain.org/status"
  "REST https://testnet-api.aurablockchain.org/cosmos/base/tendermint/v1beta1/blocks/latest"
  "Faucet https://testnet-faucet.aurablockchain.org/api/v1/health"
  "Explorer https://testnet-explorer.aurablockchain.org"
  "Status https://testnet-status.aurablockchain.org"
  "Docs https://testnet-docs.aurablockchain.org"
  "GraphQL https://testnet-graphql.aurablockchain.org/graphql"
  "Genesis https://testnet-status.aurablockchain.org/genesis.json"
  "Seeds https://testnet-status.aurablockchain.org/seed-nodes.json"
  "Registry https://testnet-status.aurablockchain.org/chain-registry.json"
  "Swagger https://testnet-status.aurablockchain.org/swagger/"
  "ChainRegistry https://testnet-status.aurablockchain.org/chain-registry.json"
)

for entry in "${ENDPOINTS[@]}"; do
  name=$(awk '{print $1}' <<< "$entry")
  url=$(awk '{print $2}' <<< "$entry")
  if [ "$name" = "GraphQL" ]; then
    if curl -fsSL --max-time 10 -H 'Content-Type: application/json' \
      -d '{"query":"{ latestBlockHeight }"}' "$url" >/dev/null; then
      echo "✅ ${name} OK"
    else
      echo "❌ ${name} FAILED (${url})"
    fi
  else
    if curl -fsSL --max-time 10 "$url" >/dev/null; then
      echo "✅ ${name} OK"
    else
      echo "❌ ${name} FAILED (${url})"
    fi
  fi
done
