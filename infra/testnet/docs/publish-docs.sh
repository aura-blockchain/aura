#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${ROOT_DIR:-$HOME/blockchain-projects/aura}"
OUTPUT_DIR="${OUTPUT_DIR:-/var/www/aura-docs}"

mkdir -p "${OUTPUT_DIR}"

cp "${ROOT_DIR}/infra/testnet/docs/index.html" "${OUTPUT_DIR}/index.html"
cp "${ROOT_DIR}/docs/GETTING_STARTED.md" "${OUTPUT_DIR}/GETTING_STARTED.md"
cp "${ROOT_DIR}/docs/testnet/README.md" "${OUTPUT_DIR}/testnet/README.md"
mkdir -p "${OUTPUT_DIR}/ops"
cp "${ROOT_DIR}/docs/ops/NODE_OPERATOR_GUIDE.md" "${OUTPUT_DIR}/ops/NODE_OPERATOR_GUIDE.md"
cp "${ROOT_DIR}/docs/ops/VALIDATOR_SETUP_GUIDE.md" "${OUTPUT_DIR}/ops/VALIDATOR_SETUP_GUIDE.md"
cp "${ROOT_DIR}/docs/ops/UPGRADE_PROCEDURES.md" "${OUTPUT_DIR}/ops/UPGRADE_PROCEDURES.md"
cp "${ROOT_DIR}/docs/ops/TROUBLESHOOTING.md" "${OUTPUT_DIR}/ops/TROUBLESHOOTING.md"
cp "${ROOT_DIR}/ARCHITECTURE.md" "${OUTPUT_DIR}/ARCHITECTURE.md"
cp "${ROOT_DIR}/DEVELOPMENT.md" "${OUTPUT_DIR}/DEVELOPMENT.md"
cp "${ROOT_DIR}/GOVERNANCE.md" "${OUTPUT_DIR}/GOVERNANCE.md"
cp "${ROOT_DIR}/SECURITY.md" "${OUTPUT_DIR}/SECURITY.md"
cp "${ROOT_DIR}/CONTRIBUTING.md" "${OUTPUT_DIR}/CONTRIBUTING.md"

echo "Docs published to ${OUTPUT_DIR}"
