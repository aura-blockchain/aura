#!/bin/bash
# ============================================================================
# AURA Blockchain - Comprehensive Health Check
# ============================================================================
# Checks all AURA services including node, explorer, monitoring, and system.
# Run on the testnet server for full diagnostics.
#
# Usage: ./health-check-all.sh [--json] [--quiet]
# ============================================================================

set -uo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Options
JSON_OUTPUT=false
QUIET=false
for arg in "$@"; do
    case $arg in
        --json) JSON_OUTPUT=true ;;
        --quiet|-q) QUIET=true ;;
    esac
done

OVERALL_HEALTHY=true
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

log() {
    if [ "$QUIET" = false ] && [ "$JSON_OUTPUT" = false ]; then
        echo -e "$@"
    fi
}

log_ok() { log "${GREEN}✓${NC} $*"; }
log_warn() { log "${YELLOW}⚠${NC} $*"; }
log_fail() { log "${RED}✗${NC} $*"; OVERALL_HEALTHY=false; }

# ============================================================================
# Service Endpoints (must match server config)
# ============================================================================
declare -A SERVICES=(
    ["aurad-rpc"]="http://127.0.0.1:10657/status"
    ["aurad-api"]="http://127.0.0.1:1317/cosmos/base/tendermint/v1beta1/syncing"
    ["aurad-grpc"]="127.0.0.1:9190"
    ["prometheus"]="http://127.0.0.1:9090/-/healthy"
    ["grafana"]="http://127.0.0.1:3000/api/health"
)

# ============================================================================
# Node Health
# ============================================================================
check_node() {
    log "${BLUE}=== AURA Node ===${NC}"

    local status
    status=$(curl -sf --max-time 5 "http://127.0.0.1:10657/status" 2>/dev/null)

    if [ -z "$status" ]; then
        log_fail "Node RPC not responding"
        return 1
    fi

    local height=$(echo "$status" | jq -r '.result.sync_info.latest_block_height // 0')
    local catching_up=$(echo "$status" | jq -r '.result.sync_info.catching_up')
    local moniker=$(echo "$status" | jq -r '.result.node_info.moniker // "unknown"')
    local network=$(echo "$status" | jq -r '.result.node_info.network // "unknown"')

    log_ok "Moniker: ${moniker}, Network: ${network}"
    log_ok "Block height: ${height}"

    if [ "$catching_up" = "false" ]; then
        log_ok "Sync status: synchronized"
    else
        log_warn "Sync status: catching up"
    fi

    # Check peers
    local net_info
    net_info=$(curl -sf --max-time 5 "http://127.0.0.1:10657/net_info" 2>/dev/null)
    if [ -n "$net_info" ]; then
        local peer_count=$(echo "$net_info" | jq -r '.result.n_peers // 0')
        local peer_names=$(echo "$net_info" | jq -r '[.result.peers[]?.node_info.moniker] | join(", ") // ""')
        if [ "$peer_count" -gt 0 ]; then
            log_ok "Peers: ${peer_count} [${peer_names}]"
        else
            log_fail "No peers connected"
        fi
    fi

    # Check validators
    local val_info
    val_info=$(curl -sf --max-time 5 "http://127.0.0.1:10657/validators" 2>/dev/null)
    if [ -n "$val_info" ]; then
        local val_count=$(echo "$val_info" | jq -r '.result.validators | length')
        log_ok "Active validators: ${val_count}"
    fi
}

# ============================================================================
# gRPC Health
# ============================================================================
check_grpc() {
    log ""
    log "${BLUE}=== gRPC ===${NC}"

    if command -v grpcurl &>/dev/null; then
        if grpcurl -plaintext 127.0.0.1:9190 list &>/dev/null; then
            log_ok "gRPC endpoint responding on :9190"
        else
            log_warn "gRPC endpoint not responding"
        fi
    else
        if nc -z 127.0.0.1 9190 &>/dev/null; then
            log_ok "gRPC port 9190 is open"
        else
            log_fail "gRPC port 9190 not accessible"
        fi
    fi
}

# ============================================================================
# REST API Health
# ============================================================================
check_api() {
    log ""
    log "${BLUE}=== REST API ===${NC}"

    local syncing
    syncing=$(curl -sf --max-time 5 "http://127.0.0.1:1317/cosmos/base/tendermint/v1beta1/syncing" 2>/dev/null)

    if [ -n "$syncing" ]; then
        log_ok "REST API responding on :1317"
    else
        log_warn "REST API not responding"
    fi
}

# ============================================================================
# Cosmovisor Health
# ============================================================================
check_cosmovisor() {
    log ""
    log "${BLUE}=== Cosmovisor ===${NC}"

    if [ -d "$HOME/.aura/cosmovisor" ]; then
        log_ok "Cosmovisor directory exists"

        if [ -x "$HOME/.aura/cosmovisor/genesis/bin/aurad" ]; then
            local version=$("$HOME/.aura/cosmovisor/genesis/bin/aurad" version 2>&1 | head -1)
            log_ok "Binary version: ${version}"
        else
            log_fail "aurad binary not found or not executable"
        fi

        # Check for pending upgrades
        if [ -d "$HOME/.aura/cosmovisor/upgrades" ]; then
            local upgrade_count=$(ls -1 "$HOME/.aura/cosmovisor/upgrades" 2>/dev/null | wc -l)
            if [ "$upgrade_count" -gt 0 ]; then
                log_ok "Upgrade binaries prepared: ${upgrade_count}"
            fi
        fi
    else
        log_warn "Cosmovisor not configured"
    fi
}

# ============================================================================
# System Health
# ============================================================================
check_system() {
    log ""
    log "${BLUE}=== System ===${NC}"

    # Disk space
    local disk_usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ "$disk_usage" -gt 90 ]; then
        log_fail "Disk usage critical: ${disk_usage}%"
    elif [ "$disk_usage" -gt 80 ]; then
        log_warn "Disk usage high: ${disk_usage}%"
    else
        log_ok "Disk usage: ${disk_usage}%"
    fi

    # Memory
    local mem_usage=$(free | awk '/^Mem:/ {printf "%.0f", $3/$2 * 100}')
    if [ "$mem_usage" -gt 90 ]; then
        log_fail "Memory usage critical: ${mem_usage}%"
    elif [ "$mem_usage" -gt 80 ]; then
        log_warn "Memory usage high: ${mem_usage}%"
    else
        log_ok "Memory usage: ${mem_usage}%"
    fi

    # CPU load
    local load=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | tr -d ',')
    local cores=$(nproc)
    log_ok "CPU load: ${load} (${cores} cores)"

    # Process check
    if pgrep -f "aurad start" > /dev/null; then
        log_ok "aurad process is running"
    else
        log_fail "aurad process is not running"
    fi
}

# ============================================================================
# Monitoring Health
# ============================================================================
check_monitoring() {
    log ""
    log "${BLUE}=== Monitoring ===${NC}"

    # Prometheus
    if curl -sf --max-time 3 "http://127.0.0.1:9090/-/healthy" &>/dev/null; then
        log_ok "Prometheus healthy on :9090"
    else
        log_warn "Prometheus not responding"
    fi

    # Grafana
    if curl -sf --max-time 3 "http://127.0.0.1:3000/api/health" &>/dev/null; then
        log_ok "Grafana healthy on :3000"
    else
        log_warn "Grafana not responding"
    fi

    # Node metrics endpoint
    if curl -sf --max-time 3 "http://127.0.0.1:26660/metrics" &>/dev/null; then
        log_ok "Node metrics available on :26660"
    else
        log_warn "Node metrics not available"
    fi
}

# ============================================================================
# Main
# ============================================================================
main() {
    log "======================================================================"
    log "AURA Health Check Report - ${TIMESTAMP}"
    log "======================================================================"
    log ""

    check_node
    check_grpc
    check_api
    check_cosmovisor
    check_system
    check_monitoring

    log ""
    log "======================================================================"
    if [ "$OVERALL_HEALTHY" = true ]; then
        log "${GREEN}Overall Status: HEALTHY${NC}"
        exit 0
    else
        log "${RED}Overall Status: ISSUES DETECTED${NC}"
        exit 1
    fi
}

main "$@"
