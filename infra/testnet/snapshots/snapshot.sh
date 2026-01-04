#!/usr/bin/env bash
set -euo pipefail

CHAIN_ID="${CHAIN_ID:-aura-testnet-1}"
NODE_HOME="${NODE_HOME:-$HOME/.aura}"
SNAPSHOT_DIR="${SNAPSHOT_DIR:-/var/lib/aura/snapshots}"
RETENTION_DAYS="${RETENTION_DAYS:-3}"

TIMESTAMP=$(date -u +"%Y%m%d-%H%M%S")
ARCHIVE_NAME="aura-${CHAIN_ID}-${TIMESTAMP}.tar.lz4"

mkdir -p "${SNAPSHOT_DIR}"

tar -C "${NODE_HOME}" -cf - data | lz4 -c > "${SNAPSHOT_DIR}/${ARCHIVE_NAME}"
find "${SNAPSHOT_DIR}" -type f -name "aura-${CHAIN_ID}-*.tar.lz4" -mtime +"${RETENTION_DAYS}" -delete

echo "Snapshot written: ${SNAPSHOT_DIR}/${ARCHIVE_NAME}"
