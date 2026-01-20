#!/bin/bash
# AURA Testnet Local Auto-Recovery Script
# Runs on each server to recover local nodes only (no SSH)
#
# Usage: ./auto-recovery-local.sh [--dry-run]
#
# Cron: */5 * * * * /path/to/auto-recovery-local.sh >> /var/log/aura-recovery.log 2>&1

set -euo pipefail

LOG_FILE="/var/log/aura-auto-recovery.log"
LOCK_FILE="/tmp/aura-auto-recovery.lock"
MAX_RESTART_ATTEMPTS=3
STALL_THRESHOLD=120
RESTART_COOLDOWN=300
HOSTNAME=$(hostname)

# Detect local nodes
declare -A LOCAL_NODES=()

for home in ~/.aura-mvp-*; do
    if [[ -d "$home" ]]; then
        node_name=$(basename "$home" | sed 's/^\.aura-mvp-//')
        case "$node_name" in
            val1) LOCAL_NODES["$node_name"]="127.0.0.1:26657:aurad-mvp-val1:$home" ;;
            val2) LOCAL_NODES["$node_name"]="127.0.0.1:26757:aurad-mvp-val2:$home" ;;
            val3) LOCAL_NODES["$node_name"]="127.0.0.1:26657:aurad-mvp-val3:$home" ;;
            val4) LOCAL_NODES["$node_name"]="127.0.0.1:26757:aurad-mvp-val4:$home" ;;
            sentry1) LOCAL_NODES["$node_name"]="127.0.0.1:26680:aurad-mvp-sentry1:$home" ;;
            sentry2) LOCAL_NODES["$node_name"]="127.0.0.1:26680:aurad-mvp-sentry2:$home" ;;
        esac
    fi
done

DRY_RUN=false
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=true

log() {
    echo "[$(date -u '+%Y-%m-%d %H:%M:%S UTC')] [$HOSTNAME] $*" | tee -a "$LOG_FILE"
}

acquire_lock() {
    if [[ -f "$LOCK_FILE" ]]; then
        local pid=$(cat "$LOCK_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log "Another recovery running (PID $pid)"
            exit 0
        fi
        rm -f "$LOCK_FILE"
    fi
    echo $$ > "$LOCK_FILE"
    trap "rm -f $LOCK_FILE" EXIT
}

get_restart_count() {
    local node="$1"
    local count_file="/tmp/aura-restart-count-$node"
    if [[ -f "$count_file" ]]; then
        local data=$(cat "$count_file")
        local timestamp=$(echo "$data" | cut -d: -f1)
        local count=$(echo "$data" | cut -d: -f2)
        if [[ $(($(date +%s) - timestamp)) -gt $RESTART_COOLDOWN ]]; then
            echo "0"
        else
            echo "$count"
        fi
    else
        echo "0"
    fi
}

increment_restart_count() {
    local node="$1"
    local count_file="/tmp/aura-restart-count-$node"
    local current=$(get_restart_count "$node")
    echo "$(date +%s):$((current + 1))" > "$count_file"
}

reset_restart_count() {
    local node="$1"
    rm -f "/tmp/aura-restart-count-$node"
}

check_node() {
    local rpc="$1"
    local status=$(curl -s --max-time 5 "http://$rpc/status" 2>/dev/null || echo "{}")

    local height=$(echo "$status" | jq -r '.result.sync_info.latest_block_height // "0"')
    local catching_up=$(echo "$status" | jq -r '.result.sync_info.catching_up // "true"')
    local block_time=$(echo "$status" | jq -r '.result.sync_info.latest_block_time // ""')

    if [[ "$height" == "0" ]] || [[ -z "$height" ]]; then
        echo "down:0:0"
        return
    fi

    local age=0
    if [[ -n "$block_time" ]] && [[ "$block_time" != "null" ]]; then
        local block_ts=$(date -d "$block_time" +%s 2>/dev/null || echo "0")
        age=$(($(date +%s) - block_ts))
    fi

    if [[ "$age" -gt "$STALL_THRESHOLD" ]]; then
        echo "stalled:$height:$age"
    elif [[ "$catching_up" == "true" ]]; then
        echo "syncing:$height:$age"
    else
        echo "healthy:$height:$age"
    fi
}

restart_node() {
    local service="$1"
    log "Restarting $service"
    [[ "$DRY_RUN" == "true" ]] && { log "[DRY-RUN] Would restart $service"; return 0; }
    sudo systemctl restart "$service" 2>&1 || { log "ERROR: Restart failed"; return 1; }
    log "Restarted $service"
}

rollback_node() {
    local service="$1"
    local home="$2"
    log "Attempting rollback for $service"
    [[ "$DRY_RUN" == "true" ]] && { log "[DRY-RUN] Would rollback $service"; return 0; }

    sudo systemctl stop "$service"
    sleep 2
    "$home/cosmovisor/current/bin/aurad" rollback --home "$home" 2>&1 || true
    sudo systemctl start "$service"
    log "Rollback complete for $service"
}

main() {
    acquire_lock
    log "Starting auto-recovery check"

    if [[ ${#LOCAL_NODES[@]} -eq 0 ]]; then
        log "No AURA nodes found"
        exit 0
    fi

    local actions=0

    for node_name in "${!LOCAL_NODES[@]}"; do
        IFS=':' read -r rpc _ service home <<< "${LOCAL_NODES[$node_name]}"
        local result=$(check_node "$rpc")
        local status=$(echo "$result" | cut -d: -f1)
        local height=$(echo "$result" | cut -d: -f2)
        local age=$(echo "$result" | cut -d: -f3)

        case "$status" in
            healthy)
                reset_restart_count "$node_name"
                ;;
            syncing)
                log "$node_name syncing (height: $height)"
                ;;
            down)
                log "ALERT: $node_name is DOWN"
                local count=$(get_restart_count "$node_name")
                if [[ "$count" -lt "$MAX_RESTART_ATTEMPTS" ]]; then
                    log "Restart attempt $((count + 1))/$MAX_RESTART_ATTEMPTS"
                    restart_node "$service" && increment_restart_count "$node_name" && ((actions++))
                else
                    log "Max restarts reached, attempting rollback"
                    rollback_node "$service" "$home" && reset_restart_count "$node_name" && ((actions++))
                fi
                ;;
            stalled)
                log "ALERT: $node_name STALLED (height: $height, age: ${age}s)"
                local count=$(get_restart_count "$node_name")
                if [[ "$count" -lt "$MAX_RESTART_ATTEMPTS" ]]; then
                    restart_node "$service" && increment_restart_count "$node_name" && ((actions++))
                fi
                ;;
        esac
    done

    if [[ "$actions" -gt 0 ]] && [[ "$DRY_RUN" == "false" ]]; then
        log "Waiting 30s for recovery..."
        sleep 30

        local unhealthy=0
        for node_name in "${!LOCAL_NODES[@]}"; do
            IFS=':' read -r rpc _ _ _ <<< "${LOCAL_NODES[$node_name]}"
            local result=$(check_node "$rpc")
            local status=$(echo "$result" | cut -d: -f1)
            [[ "$status" != "healthy" ]] && [[ "$status" != "syncing" ]] && ((unhealthy++))
        done

        [[ "$unhealthy" -gt 0 ]] && log "WARNING: $unhealthy nodes still unhealthy"
    fi

    log "Auto-recovery check complete (actions: $actions)"
}

main "$@"
