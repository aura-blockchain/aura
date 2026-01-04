#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${ROOT_DIR:-$HOME/blockchain-projects/aura}"
OUTPUT_DIR="${OUTPUT_DIR:-/var/www/aura-status}"

mkdir -p "${OUTPUT_DIR}"
cp "${ROOT_DIR}/docs/chain-registry/aura.json" "${OUTPUT_DIR}/chain-registry.json"

echo "Published chain-registry.json to ${OUTPUT_DIR}/chain-registry.json"
