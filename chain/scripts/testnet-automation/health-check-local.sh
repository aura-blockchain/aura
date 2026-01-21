#!/bin/bash
# AURA Testnet Local Health Check Script
# Runs on each server to check local nodes only (no SSH)
# Version: 2.0 - Updated with industry-standard thresholds
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

# Industry-standard thresholds
# Disk: <70% healthy, 70-85% warning, >85% critical
DISK_WARN_PERCENT=70
DISK_CRIT_PERCENT=85
# Memory: <80% healthy, 80-90% warning, >90% critical
MEM_WARN_PERCENT=80
MEM_CRIT_PERCENT=90

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

# Check disk usage
check_disk() {
    local disk_percent=$(df / | awk 'NR==2 {gsub(/%/,""); print $5}')
    local disk_status="healthy"
    local disk_issue=""

    if [[ "$disk_percent" -ge "$DISK_CRIT_PERCENT" ]]; then
        disk_status="critical"
        disk_issue="CRITICAL: Disk usage ${disk_percent}% (>${DISK_CRIT_PERCENT}%)"
    elif [[ "$disk_percent" -ge "$DISK_WARN_PERCENT" ]]; then
        disk_status="warning"
        disk_issue="WARNING: Disk usage ${disk_percent}% (>${DISK_WARN_PERCENT}%)"
    fi

    echo "$disk_percent:$disk_status:$disk_issue"
}

# Check memory usage
check_memory() {
    local mem_info=$(free | awk 'NR==2 {printf "%.0f", $3/$2 * 100}')
    local mem_status="healthy"
    local mem_issue=""

    if [[ "$mem_info" -ge "$MEM_CRIT_PERCENT" ]]; then
        mem_status="critical"
        mem_issue="CRITICAL: Memory usage ${mem_info}% (>${MEM_CRIT_PERCENT}%)"
    elif [[ "$mem_info" -ge "$MEM_WARN_PERCENT" ]]; then
        mem_status="warning"
        mem_issue="WARNING: Memory usage ${mem_info}% (>${MEM_WARN_PERCENT}%)"
    fi

    echo "$mem_info:$mem_status:$mem_issue"
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
            # Ensure catching_up is a proper JSON boolean
            local catching_up_json="false"
            [[ "$catching_up" == "true" ]] && catching_up_json="true"
            json_results+=("{\"node\":\"$node_name\",\"height\":$height,\"peers\":$n_peers,\"catching_up\":$catching_up_json,\"status\":\"$node_status\",\"issues\":$issues_json}")
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

    # System resource checks
    local disk_result=$(check_disk)
    local disk_percent=$(echo "$disk_result" | cut -d: -f1)
    local disk_status=$(echo "$disk_result" | cut -d: -f2)
    local disk_issue=$(echo "$disk_result" | cut -d: -f3-)

    local mem_result=$(check_memory)
    local mem_percent=$(echo "$mem_result" | cut -d: -f1)
    local mem_status=$(echo "$mem_result" | cut -d: -f2)
    local mem_issue=$(echo "$mem_result" | cut -d: -f3-)

    # Add system issues to the list
    if [[ "$disk_status" != "healthy" ]]; then
        issues+=("System: $disk_issue")
        all_healthy=false
    fi
    if [[ "$mem_status" != "healthy" ]]; then
        issues+=("System: $mem_issue")
        all_healthy=false
    fi

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        echo "{\"timestamp\":\"$(date -u '+%Y-%m-%dT%H:%M:%SZ')\",\"host\":\"$HOSTNAME\",\"max_height\":$max_height,\"healthy\":$all_healthy,\"system\":{\"disk_percent\":$disk_percent,\"disk_status\":\"$disk_status\",\"mem_percent\":$mem_percent,\"mem_status\":\"$mem_status\"},\"nodes\":[$(IFS=,; echo "${json_results[*]}")]}"
    else
        echo ""
        echo "Max Height: $max_height"

        # Display system resources
        local disk_color="$GREEN"
        [[ "$disk_status" == "warning" ]] && disk_color="$YELLOW"
        [[ "$disk_status" == "critical" ]] && disk_color="$RED"
        printf "Disk:        ${disk_color}%s%%${NC} (%s)\n" "$disk_percent" "$disk_status"

        local mem_color="$GREEN"
        [[ "$mem_status" == "warning" ]] && mem_color="$YELLOW"
        [[ "$mem_status" == "critical" ]] && mem_color="$RED"
        printf "Memory:      ${mem_color}%s%%${NC} (%s)\n" "$mem_percent" "$mem_status"

        echo ""
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
