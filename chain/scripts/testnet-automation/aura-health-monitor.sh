#!/bin/bash
# AURA Testnet Local Health Check Script
# Runs on each server to check local nodes only (no SSH)
#
# Usage: ./health-check-local.sh [--alert] [--json]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ALERT_THRESHOLD_SECONDS="${ALERT_THRESHOLD_SECONDS:-60}"
HOSTNAME=$(hostname)

# Hermit architecture validation (validators = hermit mode, sentries = public)
# Validators should only peer with their 2 assigned sentries
# Sentries should have at least 4 external peers
VALIDATOR_EXPECTED_PEERS=2
SENTRY_MIN_PEERS=4

# Detect which nodes are on this server based on existing directories
declare -A LOCAL_NODES=()

for home in ~/.aura-mvp-*; do
    if [[ -d "$home" ]]; then
        node_name=$(basename "$home" | sed 's/^\.aura-mvp-//')
        case "$node_name" in
            val1) LOCAL_NODES["$node_name"]="127.0.0.1:26657" ;;
            val2) LOCAL_NODES["$node_name"]="127.0.0.1:26757" ;;
            val3) LOCAL_NODES["$node_name"]="127.0.0.1:26657" ;;
            val4) LOCAL_NODES["$node_name"]="127.0.0.1:26757" ;;
            sentry1) LOCAL_NODES["$node_name"]="127.0.0.1:26680" ;;
            sentry2) LOCAL_NODES["$node_name"]="127.0.0.1:26680" ;;
        esac
    fi
done

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Parse arguments
SEND_ALERTS=false
JSON_OUTPUT=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --alert) SEND_ALERTS=true; shift ;;
        --json) JSON_OUTPUT=true; shift ;;
        *) shift ;;
    esac
done

send_alert() {
    local severity="$1"
    local title="$2"
    local message="$3"

    if [[ "$SEND_ALERTS" == "true" ]] && [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        curl -s -X POST "$SLACK_WEBHOOK_URL" \
            -H "Content-Type: application/json" \
            -d "{
                \"attachments\": [{
                    \"color\": \"$([ \"$severity\" == \"critical\" ] && echo danger || echo warning)\",
                    \"title\": \"AURA Testnet ($HOSTNAME): $title\",
                    \"text\": \"$message\",
                    \"footer\": \"aura-mvp-1 | $(date -u '+%Y-%m-%d %H:%M:%S UTC')\"
                }]
            }" >/dev/null 2>&1 || true
    fi
    echo "[$(date -u '+%Y-%m-%d %H:%M:%S UTC')] [$severity] $title: $message" >> /tmp/aura-health-check.log
}

main() {
    local issues=()
    local max_height=0
    local all_healthy=true
    local json_results=()

    [[ "$JSON_OUTPUT" != "true" ]] && echo "AURA Local Health Check ($HOSTNAME) - $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
    [[ "$JSON_OUTPUT" != "true" ]] && echo "============================================================"

    if [[ ${#LOCAL_NODES[@]} -eq 0 ]]; then
        echo "No AURA nodes found on this server"
        exit 0
    fi

    for node_name in "${!LOCAL_NODES[@]}"; do
        rpc="${LOCAL_NODES[$node_name]}"

        status=$(curl -s --max-time 5 "http://$rpc/status" 2>/dev/null || echo "{}")
        net_info=$(curl -s --max-time 5 "http://$rpc/net_info" 2>/dev/null || echo "{}")

        height=$(echo "$status" | jq -r '.result.sync_info.latest_block_height // "0"')
        catching_up=$(echo "$status" | jq -r '.result.sync_info.catching_up // "unknown"')
        latest_block_time=$(echo "$status" | jq -r '.result.sync_info.latest_block_time // ""')
        n_peers=$(echo "$net_info" | jq -r '.result.n_peers // "0"')

        [[ "$height" =~ ^[0-9]+$ ]] && [[ "$height" -gt "$max_height" ]] && max_height=$height

        local node_status="healthy"
        local node_issues=()

        if [[ "$height" == "0" ]] || [[ -z "$height" ]]; then
            node_status="down"
            node_issues+=("Node not responding")
            all_healthy=false
        else
            [[ "$catching_up" == "true" ]] && node_status="syncing"

            # Hermit architecture peer validation
            if [[ "$node_name" == val* ]]; then
                # Validators should have exactly 2 peers (their sentries)
                if [[ "$n_peers" -ne "$VALIDATOR_EXPECTED_PEERS" ]]; then
                    node_status="warning"
                    node_issues+=("HERMIT VIOLATION: validator has $n_peers peers (expected $VALIDATOR_EXPECTED_PEERS)")
                    all_healthy=false
                fi
            elif [[ "$node_name" == sentry* ]]; then
                # Sentries should have at least 4 external peers
                if [[ "$n_peers" -lt "$SENTRY_MIN_PEERS" ]]; then
                    node_status="warning"
                    node_issues+=("Low peers: $n_peers (minimum $SENTRY_MIN_PEERS)")
                    all_healthy=false
                fi
            fi

            if [[ -n "$latest_block_time" ]] && [[ "$latest_block_time" != "null" ]]; then
                block_ts=$(date -d "$latest_block_time" +%s 2>/dev/null || echo "0")
                now=$(date +%s)
                age=$((now - block_ts))
                if [[ "$age" -gt "$ALERT_THRESHOLD_SECONDS" ]]; then
                    node_status="stalled"
                    node_issues+=("Last block ${age}s ago")
                    all_healthy=false
                fi
            fi
        fi

        if [[ "$JSON_OUTPUT" == "true" ]]; then
            local issues_json="[]"
            if [[ ${#node_issues[@]} -gt 0 ]]; then
                issues_json="[$(printf '"%s",' "${node_issues[@]}" | sed 's/,$//')]"
            fi
            json_results+=("{\"node\":\"$node_name\",\"height\":$height,\"peers\":$n_peers,\"catching_up\":$catching_up,\"status\":\"$node_status\",\"issues\":$issues_json}")
        else
            local color="$GREEN"
            [[ "$node_status" == "warning" ]] && color="$YELLOW"
            [[ "$node_status" == "down" ]] || [[ "$node_status" == "stalled" ]] && color="$RED"
            printf "%-12s Height: %-8s Peers: %-3s Status: ${color}%s${NC}\n" "$node_name" "$height" "$n_peers" "$node_status"
            for issue in "${node_issues[@]:-}"; do
                [[ -n "$issue" ]] && printf "  ${YELLOW}-> %s${NC}\n" "$issue"
            done
        fi

        [[ "$node_status" != "healthy" ]] && [[ "$node_status" != "syncing" ]] && issues+=("$node_name: ${node_issues[*]:-$node_status}")
    done

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        echo "{\"timestamp\":\"$(date -u '+%Y-%m-%dT%H:%M:%SZ')\",\"host\":\"$HOSTNAME\",\"max_height\":$max_height,\"healthy\":$all_healthy,\"nodes\":[$(IFS=,; echo "${json_results[*]}")]}"
    else
        echo ""
        echo "Max Height: $max_height"
        if [[ "$all_healthy" == "true" ]]; then
            echo -e "${GREEN}All local nodes healthy${NC}"
        else
            echo -e "${RED}Issues detected:${NC}"
            for issue in "${issues[@]}"; do
                echo -e "  ${RED}- $issue${NC}"
            done
            [[ ${#issues[@]} -gt 0 ]] && send_alert "warning" "Local Health Issues" "$(printf '%s\n' "${issues[@]}")"
        fi
    fi

    [[ "$all_healthy" == "true" ]] && return 0 || return 1
}

main "$@"
