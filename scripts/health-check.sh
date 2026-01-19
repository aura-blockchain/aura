#!/bin/bash
# ============================================================================
# AURA Testnet Node Health Check (Single Node)
# ============================================================================
# Usage:
#   ./health-check.sh
#   ./health-check.sh --validator validator-2
#   ./health-check.sh --rpc-endpoint http://localhost:27657
#
# Environment overrides:
#   RPC_ENDPOINT, API_ENDPOINT, GRPC_ENDPOINT, METRICS_ENDPOINT
#   CHAIN_ID, TIMEOUT, MAX_BLOCK_AGE, MIN_PEERS, EXPECTED_VALIDATORS
#
# Port Configuration:
#   Ports can be configured via environment variables or config file.
#   Config file: /etc/aura/local-testnet-ports.conf
#
#   PORT SENTINEL INTEGRATION
#   -------------------------
#   For production, ports should be verified via Port Sentinel:
#     python scripts/port_sentinel.py verify
#     python scripts/port_sentinel.py status
# ============================================================================

set -uo pipefail

# ----------------------------------------------------------------------------
# Port Configuration (Port Sentinel compliant)
# ----------------------------------------------------------------------------
# Load port config from file if it exists
AURA_PORT_CONFIG="${AURA_PORT_CONFIG:-/etc/aura/local-testnet-ports.conf}"
if [[ -f "$AURA_PORT_CONFIG" ]]; then
    # shellcheck source=/dev/null
    source "$AURA_PORT_CONFIG"
fi

VALIDATORS=("validator-1" "validator-2" "validator-3" "validator-4")

# RPC ports - override via VAL1_RPC_PORT, VAL2_RPC_PORT, etc.
VAL1_RPC_PORT="${VAL1_RPC_PORT:-27657}"
VAL2_RPC_PORT="${VAL2_RPC_PORT:-27757}"
VAL3_RPC_PORT="${VAL3_RPC_PORT:-27857}"
VAL4_RPC_PORT="${VAL4_RPC_PORT:-27957}"
RPC_PORTS=("$VAL1_RPC_PORT" "$VAL2_RPC_PORT" "$VAL3_RPC_PORT" "$VAL4_RPC_PORT")

# API ports - override via VAL1_API_PORT, VAL2_API_PORT, etc.
VAL1_API_PORT="${VAL1_API_PORT:-2317}"
VAL2_API_PORT="${VAL2_API_PORT:-2417}"
VAL3_API_PORT="${VAL3_API_PORT:-2517}"
VAL4_API_PORT="${VAL4_API_PORT:-2617}"
API_PORTS=("$VAL1_API_PORT" "$VAL2_API_PORT" "$VAL3_API_PORT" "$VAL4_API_PORT")

# gRPC ports - override via VAL1_GRPC_PORT, VAL2_GRPC_PORT, etc.
VAL1_GRPC_PORT="${VAL1_GRPC_PORT:-10090}"
VAL2_GRPC_PORT="${VAL2_GRPC_PORT:-10190}"
VAL3_GRPC_PORT="${VAL3_GRPC_PORT:-10290}"
VAL4_GRPC_PORT="${VAL4_GRPC_PORT:-10390}"
GRPC_PORTS=("$VAL1_GRPC_PORT" "$VAL2_GRPC_PORT" "$VAL3_GRPC_PORT" "$VAL4_GRPC_PORT")

# Metrics ports - override via VAL1_METRICS_PORT, VAL2_METRICS_PORT, etc.
VAL1_METRICS_PORT="${VAL1_METRICS_PORT:-27660}"
VAL2_METRICS_PORT="${VAL2_METRICS_PORT:-27760}"
VAL3_METRICS_PORT="${VAL3_METRICS_PORT:-27860}"
VAL4_METRICS_PORT="${VAL4_METRICS_PORT:-27960}"
METRICS_PORTS=("$VAL1_METRICS_PORT" "$VAL2_METRICS_PORT" "$VAL3_METRICS_PORT" "$VAL4_METRICS_PORT")

# ----------------------------------------------------------------------------
# Defaults
# ----------------------------------------------------------------------------
CHAIN_ID="${CHAIN_ID:-aura-local-4}"
TIMEOUT="${TIMEOUT:-5}"
MAX_BLOCK_AGE="${MAX_BLOCK_AGE:-30}"
MIN_PEERS="${MIN_PEERS:-1}"
EXPECTED_VALIDATORS="${EXPECTED_VALIDATORS:-4}"

JSON_OUTPUT=false
QUIET=false
VALIDATOR="${VALIDATOR:-validator-1}"
RPC_ENDPOINT="${RPC_ENDPOINT:-}"
API_ENDPOINT="${API_ENDPOINT:-}"
GRPC_ENDPOINT="${GRPC_ENDPOINT:-}"
METRICS_ENDPOINT="${METRICS_ENDPOINT:-}"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --json)
            JSON_OUTPUT=true
            shift
            ;;
        --quiet|-q)
            QUIET=true
            shift
            ;;
        --validator)
            VALIDATOR="${2:-$VALIDATOR}"
            shift 2
            ;;
        --rpc-endpoint)
            RPC_ENDPOINT="${2:-}"
            shift 2
            ;;
        --api-endpoint)
            API_ENDPOINT="${2:-}"
            shift 2
            ;;
        --grpc-endpoint)
            GRPC_ENDPOINT="${2:-}"
            shift 2
            ;;
        --metrics-endpoint)
            METRICS_ENDPOINT="${2:-}"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

require_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "ERROR: required command not found: $1" >&2
        exit 2
    fi
}

require_cmd curl
require_cmd jq
require_cmd date

log() {
    if [[ "$QUIET" == "false" && "$JSON_OUTPUT" == "false" ]]; then
        echo -e "$@"
    fi
}

log_ok() {
    log "OK   $*"
}

log_warn() {
    log "WARN $*"
    OVERALL_DEGRADED=true
}

log_fail() {
    log "FAIL $*"
    OVERALL_HEALTHY=false
}

is_int() {
    [[ "$1" =~ ^[0-9]+$ ]]
}

port_open() {
    local host="$1"
    local port="$2"
    if command -v timeout >/dev/null 2>&1; then
        timeout "${TIMEOUT}s" bash -c "</dev/tcp/${host}/${port}" >/dev/null 2>&1
    else
        bash -c "</dev/tcp/${host}/${port}" >/dev/null 2>&1
    fi
}

curl_json() {
    local url="$1"
    curl -fsS --max-time "${TIMEOUT}" "$url"
}

to_epoch() {
    local ts="$1"
    date -u -d "$ts" +%s 2>/dev/null || echo 0
}

array_to_json() {
    local -a items=("$@")
    if [[ ${#items[@]} -eq 0 ]]; then
        echo "[]"
    else
        printf '%s\n' "${items[@]}" | jq -R . | jq -s .
    fi
}

# Resolve endpoints from validator if not provided
idx_for_validator=-1
for i in "${!VALIDATORS[@]}"; do
    if [[ "${VALIDATORS[$i]}" == "$VALIDATOR" ]]; then
        idx_for_validator="$i"
        break
    fi
done

if [[ -z "$RPC_ENDPOINT" ]]; then
    if [[ "$idx_for_validator" -ge 0 ]]; then
        RPC_ENDPOINT="http://localhost:${RPC_PORTS[$idx_for_validator]}"
    else
        # DEV DEFAULT - use first port
        RPC_ENDPOINT="http://localhost:${VAL1_RPC_PORT}"
    fi
fi

if [[ -z "$API_ENDPOINT" ]]; then
    if [[ "$idx_for_validator" -ge 0 ]]; then
        API_ENDPOINT="http://localhost:${API_PORTS[$idx_for_validator]}"
    else
        # DEV DEFAULT - use first port
        API_ENDPOINT="http://localhost:${VAL1_API_PORT}"
    fi
fi

if [[ -z "$GRPC_ENDPOINT" ]]; then
    if [[ "$idx_for_validator" -ge 0 ]]; then
        GRPC_ENDPOINT="localhost:${GRPC_PORTS[$idx_for_validator]}"
    else
        # DEV DEFAULT - use first port
        GRPC_ENDPOINT="localhost:${VAL1_GRPC_PORT}"
    fi
fi

if [[ -z "$METRICS_ENDPOINT" ]]; then
    if [[ "$idx_for_validator" -ge 0 ]]; then
        METRICS_ENDPOINT="http://localhost:${METRICS_PORTS[$idx_for_validator]}"
    else
        # DEV DEFAULT - use first port
        METRICS_ENDPOINT="http://localhost:${VAL1_METRICS_PORT}"
    fi
fi

# ----------------------------------------------------------------------------
# Health Check
# ----------------------------------------------------------------------------
OVERALL_HEALTHY=true
OVERALL_DEGRADED=false
TIMESTAMP_UTC=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

errors=()
warnings=()

status_json=""
net_json=""
val_json=""

moniker=""
chain_id=""
height=0
catching_up="true"
latest_time=""
block_age=-1
peers=0
validator_set=0

if ! status_json=$(curl_json "${RPC_ENDPOINT}/status" 2>/dev/null); then
    errors+=("rpc_unreachable")
else
    moniker=$(echo "$status_json" | jq -r '.result.node_info.moniker // ""')
    chain_id=$(echo "$status_json" | jq -r '.result.node_info.network // ""')
    height=$(echo "$status_json" | jq -r '.result.sync_info.latest_block_height // "0"')
    catching_up=$(echo "$status_json" | jq -r '.result.sync_info.catching_up // "true"')
    latest_time=$(echo "$status_json" | jq -r '.result.sync_info.latest_block_time // ""')

    if ! is_int "$height"; then
        height=0
        errors+=("invalid_height")
    fi

    if [[ "$catching_up" != "false" ]]; then
        errors+=("catching_up")
    fi

    if [[ "$height" -le 0 ]]; then
        errors+=("no_blocks")
    fi

    if [[ -n "$latest_time" ]]; then
        ts_epoch=$(to_epoch "$latest_time")
        if [[ "$ts_epoch" -gt 0 ]]; then
            now_epoch=$(date -u +%s)
            block_age=$((now_epoch - ts_epoch))
            if [[ "$block_age" -gt "$MAX_BLOCK_AGE" ]]; then
                errors+=("stale_block_time")
            fi
        else
            warnings+=("block_time_parse_failed")
        fi
    else
        warnings+=("missing_block_time")
    fi
fi

if ! net_json=$(curl_json "${RPC_ENDPOINT}/net_info" 2>/dev/null); then
    warnings+=("net_info_unavailable")
else
    peers=$(echo "$net_json" | jq -r '.result.n_peers // 0')
    if ! is_int "$peers"; then
        peers=0
    fi
    if [[ "$peers" -lt "$MIN_PEERS" ]]; then
        errors+=("low_peers")
    fi
fi

if ! val_json=$(curl_json "${RPC_ENDPOINT}/validators" 2>/dev/null); then
    warnings+=("validators_unavailable")
else
    validator_set=$(echo "$val_json" | jq -r '.result.validators | length')
    if ! is_int "$validator_set"; then
        validator_set=0
    fi
    if [[ "$validator_set" -lt "$EXPECTED_VALIDATORS" ]]; then
        errors+=("validator_set_small")
    fi
fi

if [[ -n "$CHAIN_ID" && -n "$chain_id" && "$chain_id" != "$CHAIN_ID" ]]; then
    errors+=("chain_id_mismatch")
fi

if ! curl -fsS --max-time "${TIMEOUT}" "${API_ENDPOINT}/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
    warnings+=("api_unreachable")
fi

# gRPC endpoint check
grpc_host="${GRPC_ENDPOINT%:*}"
grpc_port="${GRPC_ENDPOINT##*:}"
if [[ -z "$grpc_host" || "$grpc_host" == "$GRPC_ENDPOINT" ]]; then
    grpc_host="localhost"
    grpc_port="$GRPC_ENDPOINT"
fi
if ! port_open "$grpc_host" "$grpc_port"; then
    warnings+=("grpc_unreachable")
fi

if ! curl -fsS --max-time "${TIMEOUT}" "${METRICS_ENDPOINT}/metrics" >/dev/null 2>&1; then
    warnings+=("metrics_unreachable")
fi

status="healthy"
if [[ ${#errors[@]} -gt 0 ]]; then
    status="unhealthy"
    OVERALL_HEALTHY=false
elif [[ ${#warnings[@]} -gt 0 ]]; then
    status="degraded"
    OVERALL_DEGRADED=true
fi

if [[ "$JSON_OUTPUT" == "false" ]]; then
    log "======================================================================"
    log "AURA Node Health Check - ${TIMESTAMP_UTC}"
    log "======================================================================"
    if [[ "$status" == "healthy" ]]; then
        log_ok "${VALIDATOR} rpc=${RPC_ENDPOINT} height=${height} peers=${peers}"
    elif [[ "$status" == "degraded" ]]; then
        log_warn "${VALIDATOR} rpc=${RPC_ENDPOINT} height=${height} peers=${peers}"
    else
        log_fail "${VALIDATOR} rpc=${RPC_ENDPOINT} height=${height} peers=${peers}"
    fi
    log "     chain_id=${chain_id} catching_up=${catching_up} block_age=${block_age}s validator_set=${validator_set}"

    if [[ "$OVERALL_HEALTHY" == "true" ]]; then
        if [[ "$OVERALL_DEGRADED" == "true" ]]; then
            log "OVERALL STATUS: DEGRADED"
        else
            log "OVERALL STATUS: HEALTHY"
        fi
    else
        log "OVERALL STATUS: UNHEALTHY"
    fi
fi

if [[ "$JSON_OUTPUT" == "true" ]]; then
    errors_json=$(array_to_json "${errors[@]}")
    warnings_json=$(array_to_json "${warnings[@]}")
    jq -n \
        --arg timestamp "$TIMESTAMP_UTC" \
        --arg chain_id "$CHAIN_ID" \
        --arg status "$status" \
        --arg validator "$VALIDATOR" \
        --arg moniker "$moniker" \
        --arg rpc "$RPC_ENDPOINT" \
        --arg api "$API_ENDPOINT" \
        --arg grpc "$GRPC_ENDPOINT" \
        --arg metrics "$METRICS_ENDPOINT" \
        --argjson height "$height" \
        --argjson peers "$peers" \
        --argjson block_age "$block_age" \
        --argjson catching_up "$catching_up" \
        --argjson validator_set "$validator_set" \
        --argjson errors "$errors_json" \
        --argjson warnings "$warnings_json" \
        '{timestamp:$timestamp, chain_id:$chain_id, status:$status, validator:$validator, moniker:$moniker, rpc:$rpc, api:$api, grpc:$grpc, metrics:$metrics, height:$height, peers:$peers, block_age_seconds:$block_age, catching_up:$catching_up, validator_set:$validator_set, errors:$errors, warnings:$warnings}'
fi

if [[ "$OVERALL_HEALTHY" == "true" ]]; then
    exit 0
fi
exit 1
