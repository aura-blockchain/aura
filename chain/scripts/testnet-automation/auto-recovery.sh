#!/bin/bash
# AURA Testnet Auto-Recovery Script
# Automatically detects and recovers failed nodes
#
# Usage: ./auto-recovery.sh [--dry-run]
#
# This script should be run via cron every 5 minutes:
#   */5 * * * * /path/to/auto-recovery.sh >> /var/log/aura-recovery.log 2>&1

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="/tmp/aura-auto-recovery.log"
LOCK_FILE="/tmp/aura-auto-recovery.lock"
MAX_RESTART_ATTEMPTS=3
STALL_THRESHOLD=120  # seconds
RESTART_COOLDOWN=300  # 5 minutes between restart attempts

# Node definitions
declare -A NODES=(
    ["aura-testnet:val1"]="127.0.0.1:26657:aurad-mvp-val1:~/.aura-mvp-val1"
    ["aura-testnet:val2"]="127.0.0.1:26757:aurad-mvp-val2:~/.aura-mvp-val2"
    ["aura-testnet:sentry1"]="127.0.0.1:26680:aurad-mvp-sentry1:~/.aura-mvp-sentry1"
    ["services-testnet:val3"]="127.0.0.1:26657:aurad-mvp-val3:~/.aura-mvp-val3"
    ["services-testnet:val4"]="127.0.0.1:26757:aurad-mvp-val4:~/.aura-mvp-val4"
    ["services-testnet:sentry2"]="127.0.0.1:26680:aurad-mvp-sentry2:~/.aura-mvp-sentry2"
)

# Parse arguments
DRY_RUN=false
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=true

log() {
    echo "[$(date -u '+%Y-%m-%d %H:%M:%S UTC')] $*" | tee -a "$LOG_FILE"
}

# Prevent concurrent runs
acquire_lock() {
    if [[ -f "$LOCK_FILE" ]]; then
        local pid=$(cat "$LOCK_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log "Another recovery process is running (PID $pid), exiting"
            exit 0
        fi
        rm -f "$LOCK_FILE"
    fi
    echo $$ > "$LOCK_FILE"
    trap "rm -f $LOCK_FILE" EXIT
}

# Get restart attempt count for a node
get_restart_count() {
    local node="$1"
    local count_file="/tmp/aura-restart-count-${node//[:\/]/_}"
    if [[ -f "$count_file" ]]; then
        local data=$(cat "$count_file")
        local timestamp=$(echo "$data" | cut -d: -f1)
        local count=$(echo "$data" | cut -d: -f2)
        local now=$(date +%s)

        # Reset count if cooldown has passed
        if [[ $((now - timestamp)) -gt $RESTART_COOLDOWN ]]; then
            echo "0"
            return
        fi
        echo "$count"
    else
        echo "0"
    fi
}

# Increment restart count
increment_restart_count() {
    local node="$1"
    local count_file="/tmp/aura-restart-count-${node//[:\/]/_}"
    local current=$(get_restart_count "$node")
    echo "$(date +%s):$((current + 1))" > "$count_file"
}

# Reset restart count
reset_restart_count() {
    local node="$1"
    local count_file="/tmp/aura-restart-count-${node//[:\/]/_}"
    rm -f "$count_file"
}

# Check node status
check_node() {
    local server="$1"
    local rpc="$2"

    local status=$(ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no "$server" \
        "curl -s --max-time 5 http://$rpc/status 2>/dev/null" 2>/dev/null || echo "{}")

    local height=$(echo "$status" | jq -r '.result.sync_info.latest_block_height // "0"')
    local catching_up=$(echo "$status" | jq -r '.result.sync_info.catching_up // "true"')
    local block_time=$(echo "$status" | jq -r '.result.sync_info.latest_block_time // ""')

    # Check if responding
    if [[ "$height" == "0" ]] || [[ -z "$height" ]]; then
        echo "down:0:0"
        return
    fi

    # Check block freshness
    local age=0
    if [[ -n "$block_time" ]] && [[ "$block_time" != "null" ]]; then
        local block_ts=$(date -d "$block_time" +%s 2>/dev/null || echo "0")
        local now=$(date +%s)
        age=$((now - block_ts))
    fi

    if [[ "$age" -gt "$STALL_THRESHOLD" ]]; then
        echo "stalled:$height:$age"
    elif [[ "$catching_up" == "true" ]]; then
        echo "syncing:$height:$age"
    else
        echo "healthy:$height:$age"
    fi
}

# Restart a node
restart_node() {
    local server="$1"
    local service="$2"

    log "Restarting $service on $server"

    if [[ "$DRY_RUN" == "true" ]]; then
        log "[DRY-RUN] Would restart $service on $server"
        return 0
    fi

    ssh -o ConnectTimeout=10 "$server" "sudo systemctl restart $service" 2>&1 || {
        log "ERROR: Failed to restart $service"
        return 1
    }

    log "Successfully restarted $service"
    return 0
}

# Attempt rollback recovery
rollback_node() {
    local server="$1"
    local service="$2"
    local home="$3"

    log "Attempting rollback for $service on $server"

    if [[ "$DRY_RUN" == "true" ]]; then
        log "[DRY-RUN] Would rollback $service on $server"
        return 0
    fi

    ssh -o ConnectTimeout=10 "$server" "
        sudo systemctl stop $service
        sleep 2
        $home/cosmovisor/current/bin/aurad rollback --home $home 2>&1 || true
        sudo systemctl start $service
    " 2>&1 || {
        log "ERROR: Rollback failed for $service"
        return 1
    }

    log "Rollback completed for $service"
    return 0
}

# Get reference node for state copy
get_healthy_reference() {
    for node_key in "${!NODES[@]}"; do
        IFS=':' read -r server node_name <<< "$node_key"
        IFS=':' read -r rpc service home <<< "${NODES[$node_key]}"

        local result=$(check_node "$server" "$rpc")
        local status=$(echo "$result" | cut -d: -f1)

        if [[ "$status" == "healthy" ]]; then
            echo "$server:$home"
            return 0
        fi
    done
    echo ""
}

# Main recovery logic
main() {
    acquire_lock
    log "Starting auto-recovery check"

    local recovery_needed=false
    local actions_taken=0

    for node_key in "${!NODES[@]}"; do
        IFS=':' read -r server node_name <<< "$node_key"
        IFS=':' read -r rpc service home <<< "${NODES[$node_key]}"

        local result=$(check_node "$server" "$rpc")
        local status=$(echo "$result" | cut -d: -f1)
        local height=$(echo "$result" | cut -d: -f2)
        local age=$(echo "$result" | cut -d: -f3)

        case "$status" in
            healthy)
                # Node is healthy, reset restart counter
                reset_restart_count "$node_key"
                ;;

            syncing)
                # Node is syncing, this is OK - just log it
                log "$node_name on $server is syncing (height: $height)"
                ;;

            down)
                log "ALERT: $node_name on $server is DOWN"
                local restart_count=$(get_restart_count "$node_key")

                if [[ "$restart_count" -lt "$MAX_RESTART_ATTEMPTS" ]]; then
                    log "Attempting restart ($((restart_count + 1))/$MAX_RESTART_ATTEMPTS)"
                    if restart_node "$server" "$service"; then
                        increment_restart_count "$node_key"
                        ((actions_taken++))
                    fi
                else
                    log "Max restart attempts reached for $node_name, attempting rollback"
                    if rollback_node "$server" "$service" "$home"; then
                        reset_restart_count "$node_key"
                        ((actions_taken++))
                    fi
                fi
                recovery_needed=true
                ;;

            stalled)
                log "ALERT: $node_name on $server is STALLED (height: $height, age: ${age}s)"
                local restart_count=$(get_restart_count "$node_key")

                if [[ "$restart_count" -lt "$MAX_RESTART_ATTEMPTS" ]]; then
                    log "Attempting restart for stalled node"
                    if restart_node "$server" "$service"; then
                        increment_restart_count "$node_key"
                        ((actions_taken++))
                    fi
                else
                    log "Stall persists after restarts, may need manual intervention"
                    # Could attempt more aggressive recovery here
                fi
                recovery_needed=true
                ;;
        esac
    done

    if [[ "$recovery_needed" == "true" ]]; then
        log "Recovery actions taken: $actions_taken"

        # Wait and verify recovery
        if [[ "$actions_taken" -gt 0 ]] && [[ "$DRY_RUN" == "false" ]]; then
            log "Waiting 30s for nodes to recover..."
            sleep 30

            # Re-check nodes
            local still_unhealthy=0
            for node_key in "${!NODES[@]}"; do
                IFS=':' read -r server node_name <<< "$node_key"
                IFS=':' read -r rpc service home <<< "${NODES[$node_key]}"

                local result=$(check_node "$server" "$rpc")
                local status=$(echo "$result" | cut -d: -f1)

                if [[ "$status" != "healthy" ]] && [[ "$status" != "syncing" ]]; then
                    ((still_unhealthy++))
                    log "Node $node_name still unhealthy after recovery attempt"
                fi
            done

            if [[ "$still_unhealthy" -gt 0 ]]; then
                log "WARNING: $still_unhealthy nodes still unhealthy - manual intervention may be required"
            else
                log "All nodes recovered successfully"
            fi
        fi
    else
        log "All nodes healthy, no recovery needed"
    fi

    log "Auto-recovery check complete"
}

main "$@"
