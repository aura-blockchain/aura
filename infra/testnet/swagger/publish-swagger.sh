#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${ROOT_DIR:-$HOME/blockchain-projects/aura}"
SOURCE_DIR="${SOURCE_DIR:-${ROOT_DIR}/docs/api/swagger}"
OUTPUT_DIR="${OUTPUT_DIR:-/var/www/aura-status/swagger}"

mkdir -p "${OUTPUT_DIR}"
cp -R "${SOURCE_DIR}"/* "${OUTPUT_DIR}/"
cp "${ROOT_DIR}/docs/api/openapi.json" "${OUTPUT_DIR}/openapi.json"
cp "${ROOT_DIR}/infra/testnet/swagger/index.html" "${OUTPUT_DIR}/index.html"

echo "Published swagger docs to ${OUTPUT_DIR}"
