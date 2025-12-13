#!/bin/bash
# =============================================================================
# Hermes Bootstrap Helper
# =============================================================================
# Brings up Hermes relayer primitives (clients, connection, transfer channel)
# using the local config at config/hermes/config.toml. This script assumes the
# observer/proxy endpoints are running (docker-compose.observer.yml +
# docker-compose.proxy.yml) so Hermes only talks to the hardened RPC/API stack.
# =============================================================================

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_FILE="${HERMES_CONFIG:-${REPO_ROOT}/config/hermes/config.toml}"
HOST_CHAIN="${HOST_CHAIN:-aura-local-4}"
COUNTER_CHAIN="${COUNTER_CHAIN:-aura-counter-1}"
HOST_PORT="${HOST_PORT:-transfer}"
COUNTER_PORT="${COUNTER_PORT:-transfer}"
CHANNEL_ORDER="${CHANNEL_ORDER:-unordered}"
CHANNEL_VERSION="${CHANNEL_VERSION:-ics20-1}"

log() {
  printf '[hermes-bootstrap] %s\n' "$*" >&2
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    log "Missing dependency: $1"
    exit 1
  fi
}

require_cmd hermes

HERMES_BIN=(hermes --config "${CONFIG_FILE}")

if [ ! -f "${CONFIG_FILE}" ]; then
  log "Hermes config not found at ${CONFIG_FILE}"
  exit 1
fi

log "Using config ${CONFIG_FILE}"

run_or_skip() {
  local description="$1"
  shift
  log "${description}"
  if ! "$@"; then
    log "Command failed (see output above)."
    exit 1
  fi
}

# Display currently registered keys for visibility (helps catch missing imports)
log "Registered keys on ${HOST_CHAIN}:"
"${HERMES_BIN[@]}" keys list --chain "${HOST_CHAIN}" || true
log "Registered keys on ${COUNTER_CHAIN}:"
"${HERMES_BIN[@]}" keys list --chain "${COUNTER_CHAIN}" || true

run_or_skip "Creating/refreshing clients..." \
  "${HERMES_BIN[@]}" create client --host-chain "${HOST_CHAIN}" --reference-chain "${COUNTER_CHAIN}" || true

run_or_skip "Creating/refreshing counterparty client..." \
  "${HERMES_BIN[@]}" create client --host-chain "${COUNTER_CHAIN}" --reference-chain "${HOST_CHAIN}" || true

run_or_skip "Creating connection (may be reused if it already exists)..." \
  "${HERMES_BIN[@]}" create connection --a-chain "${HOST_CHAIN}" --b-chain "${COUNTER_CHAIN}" || true

run_or_skip "Creating ICS20 transfer channel..." \
  "${HERMES_BIN[@]}" create channel \
    --a-chain "${HOST_CHAIN}" \
    --b-chain "${COUNTER_CHAIN}" \
    --a-port "${HOST_PORT}" \
    --b-port "${COUNTER_PORT}" \
    --order "${CHANNEL_ORDER}" \
    --channel-version "${CHANNEL_VERSION}" || true

log "Bootstrap complete. Use 'hermes start' to begin relaying."
