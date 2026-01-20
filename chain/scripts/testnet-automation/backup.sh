#!/bin/bash
# AURA Testnet Backup Script
# Backs up critical node files (keys, configs) and optionally state
#
# Usage: ./backup.sh [--full] [--node NODE_NAME]
#
# Schedule via cron:
#   # Daily key/config backup at 2 AM
#   0 2 * * * /path/to/backup.sh >> /var/log/aura-backup.log 2>&1
#
#   # Weekly full backup on Sunday at 3 AM
#   0 3 * * 0 /path/to/backup.sh --full >> /var/log/aura-backup.log 2>&1

set -euo pipefail

# Configuration
BACKUP_BASE_DIR="${BACKUP_DIR:-/home/ubuntu/backups/aura}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
FULL_RETENTION_DAYS="${FULL_RETENTION_DAYS:-30}"
REMOTE_BACKUP="${REMOTE_BACKUP:-}"  # Set to rsync destination for remote backup

# Node homes (run on each server)
declare -A NODE_HOMES=(
    ["val1"]="$HOME/.aura-mvp-val1"
    ["val2"]="$HOME/.aura-mvp-val2"
    ["val3"]="$HOME/.aura-mvp-val3"
    ["val4"]="$HOME/.aura-mvp-val4"
    ["sentry1"]="$HOME/.aura-mvp-sentry1"
    ["sentry2"]="$HOME/.aura-mvp-sentry2"
)

# Parse arguments
FULL_BACKUP=false
TARGET_NODE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --full) FULL_BACKUP=true; shift ;;
        --node) TARGET_NODE="$2"; shift 2 ;;
        *) shift ;;
    esac
done

log() {
    echo "[$(date -u '+%Y-%m-%d %H:%M:%S UTC')] $*"
}

# Backup a single node
backup_node() {
    local node_name="$1"
    local node_home="$2"
    local date_stamp=$(date +%Y%m%d_%H%M%S)
    local backup_dir="$BACKUP_BASE_DIR/$node_name"

    # Check if node home exists
    if [[ ! -d "$node_home" ]]; then
        log "Skipping $node_name: directory $node_home not found"
        return 0
    fi

    mkdir -p "$backup_dir"

    if [[ "$FULL_BACKUP" == "true" ]]; then
        # Full backup including data
        local backup_file="$backup_dir/full_${date_stamp}.tar.gz"
        log "Creating full backup of $node_name to $backup_file"

        # Stop node for consistent backup
        local service_name="aurad-mvp-$node_name"
        log "Stopping $service_name for consistent backup..."
        sudo systemctl stop "$service_name" 2>/dev/null || true
        sleep 5

        tar -czf "$backup_file" \
            -C "$node_home" \
            config/ \
            data/ \
            --exclude='data/cs.wal' \
            --exclude='data/blockstore.db/*.log' \
            --exclude='data/state.db/*.log' \
            --exclude='data/tx_index.db/*.log' \
            2>/dev/null || {
                log "WARNING: tar completed with warnings"
            }

        # Restart node
        log "Restarting $service_name..."
        sudo systemctl start "$service_name" 2>/dev/null || true

        log "Full backup complete: $(du -h "$backup_file" | cut -f1)"

        # Cleanup old full backups
        find "$backup_dir" -name "full_*.tar.gz" -mtime +$FULL_RETENTION_DAYS -delete 2>/dev/null || true
    else
        # Light backup (keys and configs only)
        local backup_file="$backup_dir/keys_config_${date_stamp}.tar.gz"
        log "Creating keys/config backup of $node_name to $backup_file"

        tar -czf "$backup_file" \
            -C "$node_home" \
            config/priv_validator_key.json \
            config/node_key.json \
            config/genesis.json \
            config/config.toml \
            config/app.toml \
            config/client.toml \
            2>/dev/null || {
                log "WARNING: Some config files may not exist"
            }

        log "Config backup complete: $(du -h "$backup_file" | cut -f1)"

        # Cleanup old config backups
        find "$backup_dir" -name "keys_config_*.tar.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
    fi

    return 0
}

# Sync to remote storage
sync_remote() {
    if [[ -n "$REMOTE_BACKUP" ]]; then
        log "Syncing backups to remote: $REMOTE_BACKUP"
        rsync -avz --delete "$BACKUP_BASE_DIR/" "$REMOTE_BACKUP" 2>&1 || {
            log "WARNING: Remote sync failed"
            return 1
        }
        log "Remote sync complete"
    fi
}

# Verify backup integrity
verify_backup() {
    local backup_file="$1"
    log "Verifying backup: $backup_file"

    if tar -tzf "$backup_file" >/dev/null 2>&1; then
        log "Backup verification passed"
        return 0
    else
        log "ERROR: Backup verification FAILED"
        return 1
    fi
}

# Main
main() {
    log "Starting AURA testnet backup"
    log "Backup type: $([ "$FULL_BACKUP" == "true" ] && echo "FULL" || echo "CONFIG ONLY")"

    local success_count=0
    local fail_count=0

    # Determine which nodes to backup
    if [[ -n "$TARGET_NODE" ]]; then
        if [[ -v NODE_HOMES[$TARGET_NODE] ]]; then
            backup_node "$TARGET_NODE" "${NODE_HOMES[$TARGET_NODE]}" && ((success_count++)) || ((fail_count++))
        else
            log "ERROR: Unknown node: $TARGET_NODE"
            exit 1
        fi
    else
        # Backup all nodes that exist on this server
        for node_name in "${!NODE_HOMES[@]}"; do
            backup_node "$node_name" "${NODE_HOMES[$node_name]}" && ((success_count++)) || ((fail_count++))
        done
    fi

    # Sync to remote if configured
    sync_remote || true

    # Summary
    log "Backup complete: $success_count succeeded, $fail_count failed"

    # Show disk usage
    if [[ -d "$BACKUP_BASE_DIR" ]]; then
        log "Backup storage usage:"
        du -sh "$BACKUP_BASE_DIR"/* 2>/dev/null || true
    fi

    [[ $fail_count -eq 0 ]] && exit 0 || exit 1
}

main "$@"
