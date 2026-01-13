#!/bin/bash
# ============================================================================
# AURA Public Testnet Health Check (aura-mvp-1)
# ============================================================================
# Checks public endpoints for the AURA testnet only (no XAI/PAW).
# Uses the registered Cloudflare URLs from docs/TESTNET_ENDPOINTS.md.
#
# Usage:
#   ./public-testnet-health-check.sh [--json] [--quiet]
#
# Environment:
#   TIMEOUT (default: 8)
#   MAX_BLOCK_AGE (default: 60)
#   MIN_PEERS (default: 1)
#   EXPECTED_VALIDATORS (default: 4)
#   EXPECTED_MONIKERS (optional comma-separated list)
# ============================================================================

set -uo pipefail

# ----------------------------------------------------------------------------
# Endpoints (registered URLs only)
# ----------------------------------------------------------------------------
CHAIN_ID="aura-mvp-1"

RPC_URL="https://testnet-rpc.aurablockchain.org"
API_URL="https://testnet-api.aurablockchain.org"
GRPC_HOST="testnet-grpc.aurablockchain.org"
GRPC_PORT="443"
WS_HOST="testnet-ws.aurablockchain.org"
WS_PORT="443"

EXPLORER_URL="https://testnet-explorer.aurablockchain.org"
FAUCET_URL="https://testnet-faucet.aurablockchain.org"
MONITORING_URL="https://monitoring.aurablockchain.org"
SNAPSHOTS_URL="https://snapshots.aurablockchain.org"
ARTIFACTS_URL="https://artifacts.aurablockchain.org"
GRAPHQL_URL="https://testnet-graphql.aurablockchain.org/graphql"
DOCS_URL="https://testnet-docs.aurablockchain.org"

# ----------------------------------------------------------------------------
# Settings
# ----------------------------------------------------------------------------
TIMEOUT="${TIMEOUT:-8}"
MAX_BLOCK_AGE="${MAX_BLOCK_AGE:-60}"
MIN_PEERS="${MIN_PEERS:-1}"
EXPECTED_VALIDATORS="${EXPECTED_VALIDATORS:-4}"
EXPECTED_MONIKERS="${EXPECTED_MONIKERS:-}"

JSON_OUTPUT=false
QUIET=false

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

http_code() {
    local url="$1"
    curl -s -o /dev/null -L -w "%{http_code}" --max-time "${TIMEOUT}" "$url" 2>/dev/null || echo "000"
}

http_ok() {
    local code="$1"
    [[ "$code" =~ ^[23] ]]
}

header_present() {
    local url="$1"
    local header="$2"
    curl -sIL --max-time "${TIMEOUT}" "$url" 2>/dev/null | tr -d '\r' | awk -v h="$header" 'tolower($0) ~ tolower(h ":") { found=1 } END { exit found ? 0 : 1 }'
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

RPC_JSON="{}"
REST_JSON="{}"
GRPC_JSON="{}"
WS_JSON="{}"
RESOURCES_JSON="[]"
ARTIFACTS_JSON="[]"
NETWORK_HEALTH_JSON="[]"

# ----------------------------------------------------------------------------
# RPC Checks
# ----------------------------------------------------------------------------
check_rpc() {
    local -a errors=()
    local -a warnings=()

    local status_json=""
    local net_json=""
    local validators_json=""

    local height=0
    local catching_up="true"
    local peers=0
    local chain_id=""
    local latest_time=""
    local block_age=-1
    local validator_set=0
    local moniker=""
    local health_code="000"
    local unsafe_exposed=false
    local pprof_exposed=false

    health_code=$(http_code "${RPC_URL}/health")
    if ! http_ok "$health_code"; then
        warnings+=("rpc_health_unavailable")
    fi

    if ! status_json=$(curl_json "${RPC_URL}/status" 2>/dev/null); then
        errors+=("rpc_unreachable")
    else
        chain_id=$(echo "$status_json" | jq -r '.result.node_info.network // ""')
        moniker=$(echo "$status_json" | jq -r '.result.node_info.moniker // ""')
        height=$(echo "$status_json" | jq -r '.result.sync_info.latest_block_height // "0"')
        catching_up=$(echo "$status_json" | jq -r '.result.sync_info.catching_up // "true"')
        latest_time=$(echo "$status_json" | jq -r '.result.sync_info.latest_block_time // ""')

        if [[ "$chain_id" != "$CHAIN_ID" ]]; then
            errors+=("chain_id_mismatch")
        fi

        if ! is_int "$height" || [[ "$height" -le 0 ]]; then
            errors+=("invalid_height")
        fi

        if [[ "$catching_up" != "false" ]]; then
            errors+=("catching_up")
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

    if ! net_json=$(curl_json "${RPC_URL}/net_info" 2>/dev/null); then
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

    if ! validators_json=$(curl_json "${RPC_URL}/validators" 2>/dev/null); then
        warnings+=("validators_unavailable")
    else
        validator_set=$(echo "$validators_json" | jq -r '.result.validators | length')
        if ! is_int "$validator_set"; then
            validator_set=0
        fi
        if [[ "$validator_set" -lt "$EXPECTED_VALIDATORS" ]]; then
            errors+=("validator_set_small")
        fi
    fi

    # Unsafe endpoints should never be exposed publicly
    for endpoint in unsafe dial_peers dial_seeds unsafe_flush_mempool; do
        local code
        code=$(http_code "${RPC_URL}/${endpoint}")
        if [[ "$code" == "200" ]]; then
            unsafe_exposed=true
        fi
    done
    if [[ "$unsafe_exposed" == "true" ]]; then
        errors+=("unsafe_rpc_exposed")
    fi

    # pprof/debug endpoints should not be exposed on public RPC
    if [[ "$(http_code "${RPC_URL}/debug/pprof/")" == "200" ]]; then
        pprof_exposed=true
        warnings+=("pprof_exposed")
    fi

    local status="healthy"
    if [[ ${#errors[@]} -gt 0 ]]; then
        status="unhealthy"
        log_fail "rpc status height=${height} peers=${peers}"
    elif [[ ${#warnings[@]} -gt 0 ]]; then
        status="degraded"
        log_warn "rpc status height=${height} peers=${peers}"
    else
        log_ok "rpc status height=${height} peers=${peers}"
    fi

    if [[ "$JSON_OUTPUT" == "false" && "$QUIET" == "false" ]]; then
        log "     chain_id=${chain_id} moniker=${moniker} catching_up=${catching_up} block_age=${block_age}s validator_set=${validator_set}"
    fi

    local errors_json
    local warnings_json
    errors_json=$(array_to_json "${errors[@]}")
    warnings_json=$(array_to_json "${warnings[@]}")

    RPC_JSON=$(jq -n \
        --arg status "$status" \
        --arg rpc "$RPC_URL" \
        --arg chain_id "$chain_id" \
        --arg moniker "$moniker" \
        --argjson height "$height" \
        --argjson peers "$peers" \
        --argjson block_age "$block_age" \
        --argjson validator_set "$validator_set" \
        --arg health_code "$health_code" \
        --argjson unsafe_exposed "$unsafe_exposed" \
        --argjson pprof_exposed "$pprof_exposed" \
        --argjson errors "$errors_json" \
        --argjson warnings "$warnings_json" \
        '{service:"rpc", url:$rpc, status:$status, chain_id:$chain_id, moniker:$moniker, height:$height, peers:$peers, block_age_seconds:$block_age, validator_set:$validator_set, health_http_code:$health_code, unsafe_exposed:$unsafe_exposed, pprof_exposed:$pprof_exposed, errors:$errors, warnings:$warnings}')
}

# ----------------------------------------------------------------------------
# Security Headers
# ----------------------------------------------------------------------------
check_security_headers() {
    local -a warnings=()

    if ! header_present "$RPC_URL" "Strict-Transport-Security"; then
        warnings+=("rpc_hsts_missing")
        log_warn "rpc missing HSTS header"
    else
        log_ok "rpc HSTS header"
    fi

    if ! header_present "$API_URL" "Strict-Transport-Security"; then
        warnings+=("api_hsts_missing")
        log_warn "api missing HSTS header"
    else
        log_ok "api HSTS header"
    fi

    if [[ ${#warnings[@]} -gt 0 ]]; then
        OVERALL_DEGRADED=true
    fi
}

# ----------------------------------------------------------------------------
# REST API Checks
# ----------------------------------------------------------------------------
check_rest() {
    local -a errors=()
    local -a warnings=()

    local node_info=""
    local syncing_json=""
    local bonded_json=""

    local chain_id=""
    local syncing="true"
    local bonded_count=0

    if ! node_info=$(curl_json "${API_URL}/cosmos/base/tendermint/v1beta1/node_info" 2>/dev/null); then
        errors+=("api_unreachable")
    else
        chain_id=$(echo "$node_info" | jq -r '.default_node_info.network // ""')
        if [[ "$chain_id" != "$CHAIN_ID" ]]; then
            errors+=("chain_id_mismatch")
        fi
    fi

    if ! syncing_json=$(curl_json "${API_URL}/cosmos/base/tendermint/v1beta1/syncing" 2>/dev/null); then
        warnings+=("syncing_unavailable")
    else
        syncing=$(echo "$syncing_json" | jq -r '.syncing // "true"')
        if [[ "$syncing" != "false" ]]; then
            errors+=("api_syncing")
        fi
    fi

    if ! bonded_json=$(curl_json "${API_URL}/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.limit=100" 2>/dev/null); then
        warnings+=("bonded_validators_unavailable")
    else
        bonded_count=$(echo "$bonded_json" | jq -r '.validators | length')
        if ! is_int "$bonded_count"; then
            bonded_count=0
        fi
        if [[ "$bonded_count" -lt "$EXPECTED_VALIDATORS" ]]; then
            errors+=("bonded_validators_low")
        fi

        if [[ -n "$EXPECTED_MONIKERS" ]]; then
            local missing=()
            IFS=',' read -r -a expected <<< "$EXPECTED_MONIKERS"
            for m in "${expected[@]}"; do
                m=$(echo "$m" | xargs)
                if [[ -z "$m" ]]; then
                    continue
                fi
                if ! echo "$bonded_json" | jq -e --arg moniker "$m" '.validators[]?.description.moniker == $moniker' >/dev/null 2>&1; then
                    missing+=("$m")
                fi
            done
            if [[ ${#missing[@]} -gt 0 ]]; then
                warnings+=("missing_expected_monikers")
            fi
        fi
    fi

    local status="healthy"
    if [[ ${#errors[@]} -gt 0 ]]; then
        status="unhealthy"
        log_fail "rest api"
    elif [[ ${#warnings[@]} -gt 0 ]]; then
        status="degraded"
        log_warn "rest api"
    else
        log_ok "rest api"
    fi

    local errors_json
    local warnings_json
    errors_json=$(array_to_json "${errors[@]}")
    warnings_json=$(array_to_json "${warnings[@]}")

    REST_JSON=$(jq -n \
        --arg status "$status" \
        --arg api "$API_URL" \
        --arg chain_id "$chain_id" \
        --argjson syncing "$syncing" \
        --argjson bonded "$bonded_count" \
        --argjson errors "$errors_json" \
        --argjson warnings "$warnings_json" \
        '{service:"rest", url:$api, status:$status, chain_id:$chain_id, syncing:$syncing, bonded_validators:$bonded, errors:$errors, warnings:$warnings}')
}

# ----------------------------------------------------------------------------
# gRPC Check
# ----------------------------------------------------------------------------
check_grpc() {
    local -a warnings=()
    local status="healthy"

    if ! port_open "$GRPC_HOST" "$GRPC_PORT"; then
        status="degraded"
        warnings+=("grpc_unreachable")
        log_warn "grpc ${GRPC_HOST}:${GRPC_PORT}"
    else
        log_ok "grpc ${GRPC_HOST}:${GRPC_PORT}"
    fi

    local warnings_json
    warnings_json=$(array_to_json "${warnings[@]}")

    GRPC_JSON=$(jq -n --arg status "$status" --arg host "$GRPC_HOST" --arg port "$GRPC_PORT" --argjson warnings "$warnings_json" '{service:"grpc", host:$host, port:$port, status:$status, warnings:$warnings}')
}

# ----------------------------------------------------------------------------
# WebSocket Check (TCP)
# ----------------------------------------------------------------------------
check_ws() {
    local -a warnings=()
    local status="healthy"

    if ! port_open "$WS_HOST" "$WS_PORT"; then
        status="degraded"
        warnings+=("ws_unreachable")
        log_warn "websocket ${WS_HOST}:${WS_PORT}"
    else
        log_ok "websocket ${WS_HOST}:${WS_PORT}"
    fi

    local warnings_json
    warnings_json=$(array_to_json "${warnings[@]}")

    WS_JSON=$(jq -n --arg status "$status" --arg host "$WS_HOST" --arg port "$WS_PORT" --argjson warnings "$warnings_json" '{service:"websocket", host:$host, port:$port, status:$status, warnings:$warnings}')
}

# ----------------------------------------------------------------------------
# Resources & Artifacts
# ----------------------------------------------------------------------------
check_resource_url() {
    local name="$1"
    local url="$2"
    local code
    code=$(http_code "$url")

    if http_ok "$code"; then
        log_ok "$name"
        RESOURCES_JSON=$(echo "$RESOURCES_JSON" | jq -c --arg name "$name" --arg url "$url" --arg status "healthy" --arg code "$code" '. + [{name:$name, url:$url, status:$status, http_code:$code}]')
    else
        log_warn "$name"
        RESOURCES_JSON=$(echo "$RESOURCES_JSON" | jq -c --arg name "$name" --arg url "$url" --arg status "degraded" --arg code "$code" '. + [{name:$name, url:$url, status:$status, http_code:$code}]')
    fi
}

check_artifact_url() {
    local name="$1"
    local url="$2"
    local code
    code=$(http_code "$url")

    if http_ok "$code"; then
        log_ok "$name"
        ARTIFACTS_JSON=$(echo "$ARTIFACTS_JSON" | jq -c --arg name "$name" --arg url "$url" --arg status "healthy" --arg code "$code" '. + [{name:$name, url:$url, status:$status, http_code:$code}]')
    else
        log_warn "$name"
        ARTIFACTS_JSON=$(echo "$ARTIFACTS_JSON" | jq -c --arg name "$name" --arg url "$url" --arg status "degraded" --arg code "$code" '. + [{name:$name, url:$url, status:$status, http_code:$code}]')
    fi
}

check_graphql() {
    local code
    local status="healthy"
    local -a warnings=()

    code=$(curl -s -o /dev/null -w "%{http_code}" --max-time "${TIMEOUT}" \
        -H "Content-Type: application/json" \
        -d '{"query":"{__typename}"}' \
        "$GRAPHQL_URL" 2>/dev/null || echo "000")

    if ! http_ok "$code"; then
        status="degraded"
        warnings+=("graphql_unreachable")
        log_warn "graphql"
    else
        log_ok "graphql"
    fi

    local warnings_json
    warnings_json=$(array_to_json "${warnings[@]}")

    RESOURCES_JSON=$(echo "$RESOURCES_JSON" | jq -c --arg name "graphql" --arg url "$GRAPHQL_URL" --arg status "$status" --arg code "$code" --argjson warnings "$warnings_json" '. + [{name:$name, url:$url, status:$status, http_code:$code, warnings:$warnings}]')
}

# ----------------------------------------------------------------------------
# Network Health Modules (optional)
# ----------------------------------------------------------------------------
check_network_health_modules() {
    local endpoints=(
        "${API_URL}/aura/monitoring/v1beta1/network_health"
        "${API_URL}/aura/networksecurity/v1beta1/health"
        "${API_URL}/aura/security/v1beta1/network/health"
    )

    for url in "${endpoints[@]}"; do
        local status="unknown"
        local healthy=""
        local code
        local -a warnings=()

        code=$(http_code "$url")
        if ! http_ok "$code"; then
            status="degraded"
            warnings+=("unavailable")
        else
            local payload
            payload=$(curl_json "$url" 2>/dev/null || echo "")
            if [[ -n "$payload" ]]; then
                healthy=$(echo "$payload" | jq -r '.health.isHealthy // .health.networkHealthy // .health.networkHealthy // .networkHealthy // empty' 2>/dev/null || echo "")
                if [[ "$healthy" == "false" ]]; then
                    status="unhealthy"
                else
                    status="healthy"
                fi
            else
                status="degraded"
                warnings+=("empty_payload")
            fi
        fi

        if [[ "$status" == "unhealthy" ]]; then
            log_fail "network health module"
        elif [[ "$status" == "degraded" ]]; then
            log_warn "network health module"
        else
            log_ok "network health module"
        fi

        local warnings_json
        warnings_json=$(array_to_json "${warnings[@]}")

        NETWORK_HEALTH_JSON=$(echo "$NETWORK_HEALTH_JSON" | jq -c --arg url "$url" --arg status "$status" --arg healthy "$healthy" --arg code "$code" --argjson warnings "$warnings_json" '. + [{url:$url, status:$status, healthy:$healthy, http_code:$code, warnings:$warnings}]')
    done
}

# ----------------------------------------------------------------------------
# Main
# ----------------------------------------------------------------------------
main() {
    if [[ "$JSON_OUTPUT" == "false" ]]; then
        log "======================================================================"
        log "AURA Public Testnet Health Check - ${TIMESTAMP_UTC}"
        log "======================================================================"
        log "Chain ID: ${CHAIN_ID}"
    fi

    check_rpc
    check_rest
    check_grpc
    check_ws
    check_security_headers

    check_resource_url "explorer" "$EXPLORER_URL"
    check_resource_url "faucet" "$FAUCET_URL"
    check_resource_url "monitoring" "$MONITORING_URL"
    check_resource_url "snapshots" "$SNAPSHOTS_URL"
    check_resource_url "artifacts" "$ARTIFACTS_URL"
    check_resource_url "docs" "$DOCS_URL"

    check_graphql

    check_artifact_url "genesis.json" "${ARTIFACTS_URL}/genesis.json"
    check_artifact_url "chain.json" "${ARTIFACTS_URL}/chain.json"
    check_artifact_url "peers.txt" "${ARTIFACTS_URL}/peers.txt"
    check_artifact_url "seeds.txt" "${ARTIFACTS_URL}/seeds.txt"
    check_artifact_url "addrbook.json" "${ARTIFACTS_URL}/addrbook.json"

    check_network_health_modules

    if [[ "$JSON_OUTPUT" == "true" ]]; then
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
            --argjson rpc "$RPC_JSON" \
            --argjson rest "$REST_JSON" \
            --argjson grpc "$GRPC_JSON" \
            --argjson websocket "$WS_JSON" \
            --argjson resources "$RESOURCES_JSON" \
            --argjson artifacts "$ARTIFACTS_JSON" \
            --argjson network_health "$NETWORK_HEALTH_JSON" \
            '{timestamp:$timestamp, chain_id:$chain_id, status:$status, rpc:$rpc, rest:$rest, grpc:$grpc, websocket:$websocket, resources:$resources, artifacts:$artifacts, network_health:$network_health}'
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
