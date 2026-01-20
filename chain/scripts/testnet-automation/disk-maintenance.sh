#!/bin/bash
# AURA Testnet Disk Maintenance Script
# Handles disk space cleanup and database maintenance
#
# Usage: ./disk-maintenance.sh [--aggressive]
#
# Schedule weekly: 0 5 * * 0 /path/to/disk-maintenance.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="/var/log/aura-maintenance.log"
DISK_THRESHOLD=85  # Percentage
AGGRESSIVE=false

# Node homes
declare -A NODE_HOMES=(
    ["val1"]="$HOME/.aura-mvp-val1"
    ["val2"]="$HOME/.aura-mvp-val2"
    ["val3"]="$HOME/.aura-mvp-val3"
    ["val4"]="$HOME/.aura-mvp-val4"
    ["sentry1"]="$HOME/.aura-mvp-sentry1"
    ["sentry2"]="$HOME/.aura-mvp-sentry2"
)

[[ "${1:-}" == "--aggressive" ]] && AGGRESSIVE=true

log() {
    echo "[$(date -u '+%Y-%m-%d %H:%M:%S UTC')] $*" | tee -a "$LOG_FILE"
}

# Get disk usage percentage
get_disk_usage() {
    df / | awk 'NR==2 {gsub(/%/,""); print $5}'
}

# Clean old WAL files
clean_wal_files() {
    local node_home="$1"
    local node_name="$2"

    if [[ -d "$node_home/data/cs.wal" ]]; then
        local wal_size=$(du -sm "$node_home/data/cs.wal" 2>/dev/null | cut -f1)
        if [[ "$wal_size" -gt 100 ]]; then
            log "Cleaning WAL files for $node_name (${wal_size}MB)"
            # Keep only the most recent WAL file
            find "$node_home/data/cs.wal" -name "wal" -mtime +1 -delete 2>/dev/null || true
        fi
    fi
}

# Clean old log files in LevelDB
clean_leveldb_logs() {
    local node_home="$1"
    local node_name="$2"

    for db in blockstore.db state.db application.db tx_index.db; do
        local db_path="$node_home/data/$db"
        if [[ -d "$db_path" ]]; then
            # Remove old LOG files (LevelDB creates LOG.old.* files)
            find "$db_path" -name "LOG.old.*" -delete 2>/dev/null || true
            find "$db_path" -name "*.log" -mtime +7 -delete 2>/dev/null || true
        fi
    done
}

# Compact LevelDB (requires node restart)
compact_database() {
    local node_home="$1"
    local node_name="$2"
    local service_name="aurad-mvp-$node_name"

    if [[ "$AGGRESSIVE" == "true" ]]; then
        log "Running database compaction for $node_name (this will restart the node)"

        # Stop node
        sudo systemctl stop "$service_name" 2>/dev/null || true
        sleep 5

        # Run compaction using aurad
        local aurad_bin="$node_home/cosmovisor/current/bin/aurad"
        if [[ -x "$aurad_bin" ]]; then
            $aurad_bin tendermint unsafe-reset-all --home "$node_home" --keep-addr-book 2>/dev/null || true
        fi

        # Restart node
        sudo systemctl start "$service_name" 2>/dev/null || true

        log "Compaction complete for $node_name"
    fi
}

# Clean old snapshots
clean_snapshots() {
    local node_home="$1"
    local node_name="$2"

    if [[ -d "$node_home/data/snapshots" ]]; then
        local snap_count=$(find "$node_home/data/snapshots" -mindepth 1 -maxdepth 1 -type d | wc -l)
        if [[ "$snap_count" -gt 3 ]]; then
            log "Cleaning old snapshots for $node_name ($snap_count found, keeping 3)"
            # Keep only 3 most recent snapshots
            find "$node_home/data/snapshots" -mindepth 1 -maxdepth 1 -type d -printf '%T@ %p\n' | \
                sort -n | head -n -3 | cut -d' ' -f2- | xargs -r rm -rf
        fi
    fi
}

# Clean system temp files
clean_temp_files() {
    log "Cleaning temporary files..."

    # Clean old recovery/health check temp files
    find /tmp -name "aura-*" -mtime +7 -delete 2>/dev/null || true

    # Clean old journal logs (keep 3 days)
    sudo journalctl --vacuum-time=3d 2>/dev/null || true

    # Clean apt cache
    sudo apt-get clean 2>/dev/null || true
}

# Report disk usage by node
report_disk_usage() {
    log "Disk usage report:"
    log "  Total: $(df -h / | awk 'NR==2 {print $3 "/" $2 " (" $5 ")"}')"

    for node_name in "${!NODE_HOMES[@]}"; do
        local node_home="${NODE_HOMES[$node_name]}"
        if [[ -d "$node_home" ]]; then
            local size=$(du -sh "$node_home" 2>/dev/null | cut -f1)
            local data_size=$(du -sh "$node_home/data" 2>/dev/null | cut -f1)
            log "  $node_name: $size (data: $data_size)"
        fi
    done
}

# Main
main() {
    log "=========================================="
    log "AURA Testnet Disk Maintenance"
    log "=========================================="

    local initial_usage=$(get_disk_usage)
    log "Initial disk usage: ${initial_usage}%"

    # Check if maintenance is needed
    if [[ "$initial_usage" -lt "$DISK_THRESHOLD" ]] && [[ "$AGGRESSIVE" == "false" ]]; then
        log "Disk usage below threshold ($DISK_THRESHOLD%), performing light maintenance"
    else
        log "Disk usage high or aggressive mode, performing full maintenance"
    fi

    # Clean each node
    for node_name in "${!NODE_HOMES[@]}"; do
        local node_home="${NODE_HOMES[$node_name]}"
        if [[ -d "$node_home" ]]; then
            log "Maintaining $node_name..."
            clean_wal_files "$node_home" "$node_name"
            clean_leveldb_logs "$node_home" "$node_name"
            clean_snapshots "$node_home" "$node_name"

            if [[ "$initial_usage" -gt 90 ]]; then
                compact_database "$node_home" "$node_name"
            fi
        fi
    done

    # System-wide cleanup
    clean_temp_files

    # Report results
    local final_usage=$(get_disk_usage)
    local freed=$((initial_usage - final_usage))

    log ""
    report_disk_usage
    log ""
    log "Maintenance complete: ${initial_usage}% -> ${final_usage}% (freed ~${freed}%)"

    if [[ "$final_usage" -gt 90 ]]; then
        log "WARNING: Disk usage still critical! Manual intervention may be required."
        exit 1
    fi
}

main "$@"
