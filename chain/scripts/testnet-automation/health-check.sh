#!/bin/bash
# AURA Testnet Health Check Script
# Checks all nodes and sends alerts for issues
#
# Usage: ./health-check.sh [--alert] [--json]
#
# Environment variables:
#   SLACK_WEBHOOK_URL - Slack webhook for alerts
#   ALERT_THRESHOLD_BLOCKS - Blocks behind before alerting (default: 10)
#   MIN_PEERS - Minimum peer count before alerting (default: 2)

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ALERT_THRESHOLD_BLOCKS="${ALERT_THRESHOLD_BLOCKS:-10}"
STALL_THRESHOLD_SECONDS="${STALL_THRESHOLD_SECONDS:-60}"

# Hermit architecture validation (validators = hermit mode, sentries = public)
# Validators should only peer with their 2 assigned sentries
# Sentries should have at least 4 external peers
VALIDATOR_EXPECTED_PEERS=2
SENTRY_MIN_PEERS=4

# Node definitions (server:name:rpc_port)
declare -A NODES=(
    ["aura-testnet:val1"]="127.0.0.1:26657"
    ["aura-testnet:val2"]="127.0.0.1:26757"
    ["aura-testnet:sentry1"]="127.0.0.1:26680"
    ["services-testnet:val3"]="127.0.0.1:26657"
    ["services-testnet:val4"]="127.0.0.1:26757"
    ["services-testnet:sentry2"]="127.0.0.1:26680"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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

# Alert function
send_alert() {
    local severity="$1"
    local title="$2"
    local message="$3"

    if [[ "$SEND_ALERTS" == "true" ]] && [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local color="warning"
        [[ "$severity" == "critical" ]] && color="danger"
        [[ "$severity" == "info" ]] && color="good"

        curl -s -X POST "$SLACK_WEBHOOK_URL" \
            -H "Content-Type: application/json" \
            -d "{
                \"attachments\": [{
                    \"color\": \"$color\",
                    \"title\": \"AURA Testnet: $title\",
                    \"text\": \"$message\",
                    \"footer\": \"aura-mvp-1 | $(date -u '+%Y-%m-%d %H:%M:%S UTC')\"
                }]
            }" >/dev/null 2>&1 || true
    fi

    # Log to file
    echo "[$(date -u '+%Y-%m-%d %H:%M:%S UTC')] [$severity] $title: $message" >> /tmp/aura-health-check.log
}

# Get node status
get_node_status() {
    local server="$1"
    local rpc="$2"
    local timeout=5

    ssh -o ConnectTimeout=3 -o StrictHostKeyChecking=no "$server" \
        "curl -s --max-time $timeout http://$rpc/status 2>/dev/null" 2>/dev/null || echo "{}"
}

# Get net info
get_net_info() {
    local server="$1"
    local rpc="$2"

    ssh -o ConnectTimeout=3 -o StrictHostKeyChecking=no "$server" \
        "curl -s --max-time 5 http://$rpc/net_info 2>/dev/null" 2>/dev/null || echo "{}"
}

# Main health check
main() {
    local issues=()
    local max_height=0
    local node_heights=()
    local json_results=()
    local all_healthy=true

    echo "AURA Testnet Health Check - $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
    echo "============================================================"
    echo ""

    # Collect status from all nodes
    for node_key in "${!NODES[@]}"; do
        IFS=':' read -r server node_name <<< "$node_key"
        rpc="${NODES[$node_key]}"

        status=$(get_node_status "$server" "$rpc")
        net_info=$(get_net_info "$server" "$rpc")

        # Parse status
        height=$(echo "$status" | jq -r '.result.sync_info.latest_block_height // "0"')
        catching_up=$(echo "$status" | jq -r '.result.sync_info.catching_up // "unknown"')
        latest_block_time=$(echo "$status" | jq -r '.result.sync_info.latest_block_time // ""')
        voting_power=$(echo "$status" | jq -r '.result.validator_info.voting_power // "0"')
        n_peers=$(echo "$net_info" | jq -r '.result.n_peers // "0"')

        # Track max height
        if [[ "$height" =~ ^[0-9]+$ ]] && [[ "$height" -gt "$max_height" ]]; then
            max_height=$height
        fi

        # Store for later comparison
        node_heights+=("$node_key:$height")

        # Determine status
        local node_status="healthy"
        local node_issues=()

        # Check if node is responding
        if [[ "$height" == "0" ]] || [[ -z "$height" ]]; then
            node_status="down"
            node_issues+=("Node not responding")
            all_healthy=false
        else
            # Check if catching up
            if [[ "$catching_up" == "true" ]]; then
                node_status="syncing"
                node_issues+=("Node is catching up")
            fi

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
                    node_issues+=("Low peer count: $n_peers (minimum $SENTRY_MIN_PEERS)")
                    all_healthy=false
                fi
            fi

            # Check block time freshness
            if [[ -n "$latest_block_time" ]] && [[ "$latest_block_time" != "null" ]]; then
                block_timestamp=$(date -d "$latest_block_time" +%s 2>/dev/null || echo "0")
                now=$(date +%s)
                age=$((now - block_timestamp))

                if [[ "$age" -gt "$STALL_THRESHOLD_SECONDS" ]]; then
                    node_status="stalled"
                    node_issues+=("Last block ${age}s ago")
                    all_healthy=false
                fi
            fi
        fi

        # Output
        if [[ "$JSON_OUTPUT" == "true" ]]; then
            local issues_json="[]"
            if [[ ${#node_issues[@]} -gt 0 ]]; then
                issues_json="[$(printf '"%s",' "${node_issues[@]}" | sed 's/,$//')]"
            fi
            json_results+=("{\"server\":\"$server\",\"node\":\"$node_name\",\"height\":$height,\"peers\":$n_peers,\"catching_up\":$catching_up,\"status\":\"$node_status\",\"issues\":$issues_json}")
        else
            local status_color="$GREEN"
            [[ "$node_status" == "warning" ]] && status_color="$YELLOW"
            [[ "$node_status" == "down" ]] || [[ "$node_status" == "stalled" ]] && status_color="$RED"

            printf "%-20s %-10s Height: %-8s Peers: %-3s Status: ${status_color}%s${NC}\n" \
                "$server" "$node_name" "$height" "$n_peers" "$node_status"

            if [[ ${#node_issues[@]} -gt 0 ]]; then
                for issue in "${node_issues[@]}"; do
                    printf "  ${YELLOW}-> %s${NC}\n" "$issue"
                done
            fi
        fi

        # Collect issues for alerting
        if [[ "$node_status" != "healthy" ]]; then
            issues+=("$node_name@$server: ${node_issues[*]:-$node_status}")
        fi
    done

    echo ""

    # Check for nodes falling behind
    for node_info in "${node_heights[@]}"; do
        IFS=':' read -r server node_name height <<< "$node_info"
        if [[ "$height" =~ ^[0-9]+$ ]] && [[ "$max_height" =~ ^[0-9]+$ ]]; then
            behind=$((max_height - height))
            if [[ "$behind" -gt "$ALERT_THRESHOLD_BLOCKS" ]]; then
                issues+=("$node_name is $behind blocks behind")
                all_healthy=false
            fi
        fi
    done

    # JSON output
    if [[ "$JSON_OUTPUT" == "true" ]]; then
        echo "{\"timestamp\":\"$(date -u '+%Y-%m-%dT%H:%M:%SZ')\",\"max_height\":$max_height,\"healthy\":$all_healthy,\"nodes\":[$(IFS=,; echo "${json_results[*]}")]}"
        return
    fi

    # Summary
    echo "Max Height: $max_height"

    if [[ "$all_healthy" == "true" ]]; then
        echo -e "${GREEN}All nodes healthy${NC}"
    else
        echo -e "${RED}Issues detected:${NC}"
        for issue in "${issues[@]}"; do
            echo -e "  ${RED}- $issue${NC}"
        done

        # Send alert
        if [[ ${#issues[@]} -gt 0 ]]; then
            send_alert "warning" "Health Check Issues" "$(printf '%s\n' "${issues[@]}")"
        fi
    fi

    # Return exit code based on health
    [[ "$all_healthy" == "true" ]] && return 0 || return 1
}

main "$@"
