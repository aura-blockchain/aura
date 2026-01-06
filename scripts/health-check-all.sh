#!/bin/bash
# ============================================================================
# AURA Testnet Health Check (4-Validator Local Testnet)
# ============================================================================
# Comprehensive health checks for the current docker-compose testnet layout.
#
# Usage:
#   ./health-check-all.sh [--json] [--quiet] [--validator validator-1]
#   ./health-check-all.sh --no-docker --no-logs
#
# Environment:
#   CHAIN_ID (default: aura-local-4)
#   RPC_HOST (default: localhost)
#   TIMEOUT (default: 5)
#   MAX_HEIGHT_DIFF (default: 5)
#   MAX_BLOCK_AGE (default: 30)
#   MIN_PEERS (default: 1)
#   EXPECTED_VALIDATORS (default: 4)
#   CHECK_DOCKER (default: true)
#   CHECK_LOGS (default: true)
#   LOG_LINES (default: 200)
# ============================================================================

set -uo pipefail

# ----------------------------------------------------------------------------
# Defaults
# ----------------------------------------------------------------------------
VALIDATORS=("validator-1" "validator-2" "validator-3" "validator-4")
RPC_PORTS=(27657 27757 27857 27957)
API_PORTS=(2317 2417 2517 2617)
GRPC_PORTS=(10090 10190 10290 10390)
METRICS_PORTS=(27660 27760 27860 27960)

SENTRY_NODES=("sentry-1" "sentry-2")
SENTRY_RPC_PORTS=(28659 28663)
SENTRY_API_PORTS=(2320 2321)
SENTRY_GRPC_PORTS=(11091 11092)
SENTRY_METRICS_PORTS=(28661 28664)

PROMETHEUS_PORT=9094
GRAFANA_PORT=3002

CHAIN_ID="${CHAIN_ID:-aura-local-4}"
RPC_HOST="${RPC_HOST:-localhost}"
TIMEOUT="${TIMEOUT:-5}"
MAX_HEIGHT_DIFF="${MAX_HEIGHT_DIFF:-5}"
MAX_BLOCK_AGE="${MAX_BLOCK_AGE:-30}"
MIN_PEERS="${MIN_PEERS:-1}"
EXPECTED_VALIDATORS="${EXPECTED_VALIDATORS:-4}"
CHECK_DOCKER="${CHECK_DOCKER:-true}"
CHECK_LOGS="${CHECK_LOGS:-true}"
LOG_LINES="${LOG_LINES:-200}"

JSON_OUTPUT=false
QUIET=false
VALIDATOR_FILTER=""

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
        --no-docker)
            CHECK_DOCKER=false
            shift
            ;;
        --no-logs)
            CHECK_LOGS=false
            shift
            ;;
        --validator)
            VALIDATOR_FILTER="${2:-}"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

# ----------------------------------------------------------------------------
# Helpers
# ----------------------------------------------------------------------------
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

is_true() {
    [[ "$1" == "true" || "$1" == "1" || "$1" == "yes" ]]
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

# ----------------------------------------------------------------------------
# State
# ----------------------------------------------------------------------------
OVERALL_HEALTHY=true
OVERALL_DEGRADED=false
TIMESTAMP_UTC=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

VALIDATOR_JSON=()
SENTRY_JSON=()
MONITORING_JSON=()

declare -A HEIGHTS

declare -A CHAINS

# ----------------------------------------------------------------------------
# Validator Check
# ----------------------------------------------------------------------------
check_validator() {
    local idx="$1"
    local name="${VALIDATORS[$idx]}"

    if [[ -n "$VALIDATOR_FILTER" && "$VALIDATOR_FILTER" != "$name" ]]; then
        return 0
    fi

    local rpc_port="${RPC_PORTS[$idx]}"
    local api_port="${API_PORTS[$idx]}"
    local grpc_port="${GRPC_PORTS[$idx]}"
    local metrics_port="${METRICS_PORTS[$idx]}"

    local rpc_url="http://${RPC_HOST}:${rpc_port}"
    local api_url="http://${RPC_HOST}:${api_port}"
    local metrics_url="http://${RPC_HOST}:${metrics_port}"

    local status_json=""
    local net_json=""
    local val_json=""

    local moniker=""
    local chain_id=""
    local height=0
    local catching_up="true"
    local latest_time=""
    local block_age=-1
    local peers=0
    local validator_set=0
    local peer_validator_count=0

    local -a errors=()
    local -a warnings=()

    if ! status_json=$(curl_json "${rpc_url}/status" 2>/dev/null); then
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
            local ts_epoch
            ts_epoch=$(to_epoch "$latest_time")
            if [[ "$ts_epoch" -gt 0 ]]; then
                local now_epoch
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

    if ! net_json=$(curl_json "${rpc_url}/net_info" 2>/dev/null); then
        warnings+=("net_info_unavailable")
    else
        peers=$(echo "$net_json" | jq -r '.result.n_peers // 0')
        if ! is_int "$peers"; then
            peers=0
        fi

        if [[ "$peers" -lt "$MIN_PEERS" ]]; then
            errors+=("low_peers")
        fi

        peer_validator_count=$(echo "$net_json" | jq -r '[.result.peers[]?.node_info.moniker] | map(select(. != null)) | join("\n")' | \
            awk 'NF' | while read -r peer; do
                for v in "${VALIDATORS[@]}"; do
                    if [[ "$peer" == "$v" ]]; then
                        echo 1
                    fi
                done
            done | wc -l)

        if [[ "$peer_validator_count" -eq 0 ]]; then
            warnings+=("no_validator_peers")
        fi
    fi

    if ! val_json=$(curl_json "${rpc_url}/validators" 2>/dev/null); then
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

    if [[ -n "$moniker" && "$moniker" != "$name" ]]; then
        warnings+=("moniker_mismatch")
    fi

    if ! curl -fsS --max-time "${TIMEOUT}" "${api_url}/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
        warnings+=("api_unreachable")
    fi

    if ! port_open "$RPC_HOST" "$grpc_port"; then
        warnings+=("grpc_unreachable")
    fi

    if ! curl -fsS --max-time "${TIMEOUT}" "${metrics_url}/metrics" >/dev/null 2>&1; then
        warnings+=("metrics_unreachable")
    fi

    HEIGHTS["$name"]="$height"
    CHAINS["$name"]="$chain_id"

    local status="healthy"
    if [[ ${#errors[@]} -gt 0 ]]; then
        status="unhealthy"
        log_fail "$name rpc=${rpc_port} height=${height} peers=${peers}"
    elif [[ ${#warnings[@]} -gt 0 ]]; then
        status="degraded"
        log_warn "$name rpc=${rpc_port} height=${height} peers=${peers}"
    else
        log_ok "$name rpc=${rpc_port} height=${height} peers=${peers}"
    fi

    if [[ "$JSON_OUTPUT" == "false" && "$QUIET" == "false" ]]; then
        log "     chain_id=${chain_id} catching_up=${catching_up} block_age=${block_age}s validator_set=${validator_set}"
    fi

    local errors_json
    local warnings_json
    errors_json=$(array_to_json "${errors[@]}")
    warnings_json=$(array_to_json "${warnings[@]}")

    local validator_json
    validator_json=$(jq -n \
        --arg name "$name" \
        --arg moniker "$moniker" \
        --arg chain_id "$chain_id" \
        --arg rpc "$rpc_url" \
        --arg api "$api_url" \
        --arg metrics "$metrics_url" \
        --arg status "$status" \
        --argjson height "$height" \
        --argjson peers "$peers" \
        --argjson block_age "$block_age" \
        --argjson catching_up "$catching_up" \
        --argjson validator_set "$validator_set" \
        --argjson errors "$errors_json" \
        --argjson warnings "$warnings_json" \
        '{name:$name, moniker:$moniker, chain_id:$chain_id, rpc:$rpc, api:$api, metrics:$metrics, height:$height, peers:$peers, block_age_seconds:$block_age, catching_up:$catching_up, validator_set:$validator_set, status:$status, errors:$errors, warnings:$warnings}')

    VALIDATOR_JSON+=("$validator_json")
}

# ----------------------------------------------------------------------------
# Optional Node Check (Sentries)
# ----------------------------------------------------------------------------
check_sentry() {
    local idx="$1"
    local name="${SENTRY_NODES[$idx]}"

    local rpc_port="${SENTRY_RPC_PORTS[$idx]}"
    local api_port="${SENTRY_API_PORTS[$idx]}"
    local grpc_port="${SENTRY_GRPC_PORTS[$idx]}"
    local metrics_port="${SENTRY_METRICS_PORTS[$idx]}"

    local rpc_url="http://${RPC_HOST}:${rpc_port}"
    local api_url="http://${RPC_HOST}:${api_port}"
    local metrics_url="http://${RPC_HOST}:${metrics_port}"

    local status="unknown"
    local -a warnings=()

    if curl -fsS --max-time "${TIMEOUT}" "${rpc_url}/status" >/dev/null 2>&1; then
        status="healthy"
        log_ok "${name} rpc=${rpc_port}"
    else
        status="degraded"
        warnings+=("rpc_unreachable")
        log_warn "${name} rpc=${rpc_port}"
    fi

    if ! curl -fsS --max-time "${TIMEOUT}" "${api_url}/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
        warnings+=("api_unreachable")
    fi

    if ! port_open "$RPC_HOST" "$grpc_port"; then
        warnings+=("grpc_unreachable")
    fi

    if ! curl -fsS --max-time "${TIMEOUT}" "${metrics_url}/metrics" >/dev/null 2>&1; then
        warnings+=("metrics_unreachable")
    fi

    local warnings_json
    warnings_json=$(array_to_json "${warnings[@]}")

    local sentry_json
    sentry_json=$(jq -n \
        --arg name "$name" \
        --arg rpc "$rpc_url" \
        --arg api "$api_url" \
        --arg metrics "$metrics_url" \
        --arg status "$status" \
        --argjson warnings "$warnings_json" \
        '{name:$name, rpc:$rpc, api:$api, metrics:$metrics, status:$status, warnings:$warnings}')

    SENTRY_JSON+=("$sentry_json")
}

# ----------------------------------------------------------------------------
# Monitoring Check
# ----------------------------------------------------------------------------
check_monitoring() {
    local prom_url="http://${RPC_HOST}:${PROMETHEUS_PORT}/-/healthy"
    local graf_url="http://${RPC_HOST}:${GRAFANA_PORT}/api/health"

    local prom_status="unknown"
    local graf_status="unknown"

    if curl -fsS --max-time "${TIMEOUT}" "$prom_url" >/dev/null 2>&1; then
        prom_status="healthy"
        log_ok "prometheus port=${PROMETHEUS_PORT}"
    else
        prom_status="degraded"
        log_warn "prometheus port=${PROMETHEUS_PORT}"
    fi

    if curl -fsS --max-time "${TIMEOUT}" "$graf_url" >/dev/null 2>&1; then
        graf_status="healthy"
        log_ok "grafana port=${GRAFANA_PORT}"
    else
        graf_status="degraded"
        log_warn "grafana port=${GRAFANA_PORT}"
    fi

    MONITORING_JSON+=("$(jq -n --arg service "prometheus" --arg status "$prom_status" --arg url "$prom_url" '{service:$service, status:$status, url:$url}')")
    MONITORING_JSON+=("$(jq -n --arg service "grafana" --arg status "$graf_status" --arg url "$graf_url" '{service:$service, status:$status, url:$url}')")
}

# ----------------------------------------------------------------------------
# Docker Checks
# ----------------------------------------------------------------------------
check_docker_containers() {
    if ! command -v docker >/dev/null 2>&1; then
        log_warn "docker not available; skipping container checks"
        return 0
    fi

    if ! docker ps >/dev/null 2>&1; then
        log_warn "docker not running; skipping container checks"
        return 0
    fi

    for v in "${VALIDATORS[@]}"; do
        local container="aura-${v}"
        if ! docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
            log_fail "container not running: ${container}"
            continue
        fi
        local health=""
        health=$(docker inspect --format '{{.State.Health.Status}}' "$container" 2>/dev/null || echo "none")
        if [[ "$health" == "unhealthy" ]]; then
            log_fail "container health unhealthy: ${container}"
        elif [[ "$health" == "starting" ]]; then
            log_warn "container health starting: ${container}"
        fi
    done
}

check_logs() {
    if ! command -v docker >/dev/null 2>&1; then
        return 0
    fi

    if ! docker ps >/dev/null 2>&1; then
        return 0
    fi

    local patterns=("consensus failure" "double sign" "panic")
    local found=false

    for v in "${VALIDATORS[@]}"; do
        local container="aura-${v}"
        if ! docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
            continue
        fi
        local logs
        logs=$(docker logs --tail "${LOG_LINES}" "$container" 2>/dev/null || true)
        for p in "${patterns[@]}"; do
            if echo "$logs" | grep -qi "$p"; then
                log_fail "log pattern '${p}' found in ${container}"
                found=true
            fi
        done
    done

    if [[ "$found" == "false" && "$QUIET" == "false" && "$JSON_OUTPUT" == "false" ]]; then
        log_ok "no critical log patterns found"
    fi
}

# ----------------------------------------------------------------------------
# Cross-Validator Checks
# ----------------------------------------------------------------------------
check_consensus() {
    local max_height=0
    local min_height=0

    for v in "${VALIDATORS[@]}"; do
        if [[ -n "$VALIDATOR_FILTER" && "$VALIDATOR_FILTER" != "$v" ]]; then
            continue
        fi
        local h="${HEIGHTS[$v]:-0}"
        if ! is_int "$h"; then
            continue
        fi
        if [[ "$h" -gt "$max_height" ]]; then
            max_height="$h"
        fi
        if [[ "$min_height" -eq 0 || "$h" -lt "$min_height" ]]; then
            min_height="$h"
        fi
    done

    if [[ "$max_height" -eq 0 || "$min_height" -eq 0 ]]; then
        log_fail "unable to compute height range"
        return
    fi

    local diff=$((max_height - min_height))
    if [[ "$diff" -gt "$MAX_HEIGHT_DIFF" ]]; then
        log_fail "height divergence diff=${diff} (min=${min_height} max=${max_height})"
    else
        log_ok "height divergence diff=${diff} (min=${min_height} max=${max_height})"
    fi
}

check_chain_id_consistency() {
    if [[ -z "$CHAIN_ID" ]]; then
        return
    fi

    local mismatch=false
    for v in "${VALIDATORS[@]}"; do
        if [[ -n "$VALIDATOR_FILTER" && "$VALIDATOR_FILTER" != "$v" ]]; then
            continue
        fi
        local chain="${CHAINS[$v]:-}"
        if [[ -n "$chain" && "$chain" != "$CHAIN_ID" ]]; then
            mismatch=true
        fi
    done

    if [[ "$mismatch" == "true" ]]; then
        log_fail "chain_id mismatch (expected ${CHAIN_ID})"
    else
        log_ok "chain_id consistent (${CHAIN_ID})"
    fi
}

# ----------------------------------------------------------------------------
# Main
# ----------------------------------------------------------------------------
main() {
    if [[ "$JSON_OUTPUT" == "false" ]]; then
        log "======================================================================"
        log "AURA Testnet Health Check - ${TIMESTAMP_UTC}"
        log "======================================================================"
    fi

    for i in "${!VALIDATORS[@]}"; do
        check_validator "$i"
    done

    if [[ -z "$VALIDATOR_FILTER" ]]; then
        check_consensus
        check_chain_id_consistency
    fi

    if is_true "$CHECK_DOCKER"; then
        check_docker_containers
    fi

    if is_true "$CHECK_LOGS"; then
        check_logs
    fi

    if [[ -z "$VALIDATOR_FILTER" ]]; then
        for i in "${!SENTRY_NODES[@]}"; do
            check_sentry "$i"
        done
        check_monitoring
    fi

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        local validators_json
        local sentries_json
        local monitoring_json

        if [[ ${#VALIDATOR_JSON[@]} -eq 0 ]]; then
            validators_json="[]"
        else
            validators_json=$(printf '%s\n' "${VALIDATOR_JSON[@]}" | jq -s .)
        fi

        if [[ ${#SENTRY_JSON[@]} -eq 0 ]]; then
            sentries_json="[]"
        else
            sentries_json=$(printf '%s\n' "${SENTRY_JSON[@]}" | jq -s .)
        fi

        if [[ ${#MONITORING_JSON[@]} -eq 0 ]]; then
            monitoring_json="[]"
        else
            monitoring_json=$(printf '%s\n' "${MONITORING_JSON[@]}" | jq -s .)
        fi

        local overall_status="healthy"
        if [[ "$OVERALL_HEALTHY" == "false" ]]; then
            overall_status="unhealthy"
        elif [[ "$OVERALL_DEGRADED" == "true" ]]; then
            overall_status="degraded"
        fi

        jq -n \
            --arg timestamp "$TIMESTAMP_UTC" \
            --arg chain_id "$CHAIN_ID" \
            --arg status "$overall_status" \
            --argjson validators "$validators_json" \
            --argjson sentries "$sentries_json" \
            --argjson monitoring "$monitoring_json" \
            '{timestamp:$timestamp, chain_id:$chain_id, status:$status, validators:$validators, sentries:$sentries, monitoring:$monitoring}'
    else
        log "---------------------------------------------------------------------"
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

    if [[ "$OVERALL_HEALTHY" == "true" ]]; then
        exit 0
    fi
    exit 1
}

main
