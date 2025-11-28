# AURA Backup and Recovery Procedures

This runbook covers backup strategies and disaster recovery procedures for AURA blockchain nodes.

## Critical Data Overview

| Data | Location | Priority | Backup Frequency |
|------|----------|----------|------------------|
| Validator Key | `~/.aura/config/priv_validator_key.json` | **CRITICAL** | Once (at creation) |
| Node Key | `~/.aura/config/node_key.json` | High | Once (at creation) |
| Validator State | `~/.aura/data/priv_validator_state.json` | **CRITICAL** | After every restart |
| Config Files | `~/.aura/config/*.toml` | Medium | After changes |
| Genesis | `~/.aura/config/genesis.json` | Medium | Once |
| Chain Data | `~/.aura/data/` | Low | Daily snapshots |

## Backup Procedures

### 1. Validator Key Backup (One-Time, CRITICAL)

The validator key controls your stake and must be backed up securely.

```bash
#!/bin/bash
# backup-validator-key.sh

BACKUP_DIR="/secure/backup/$(date +%Y%m%d)"
KEY_FILE="$HOME/.aura/config/priv_validator_key.json"

mkdir -p "$BACKUP_DIR"

# Encrypt with GPG
gpg --symmetric --cipher-algo AES256 \
    --output "$BACKUP_DIR/priv_validator_key.json.gpg" \
    "$KEY_FILE"

# Generate checksum
sha256sum "$BACKUP_DIR/priv_validator_key.json.gpg" > "$BACKUP_DIR/checksum.sha256"

echo "Backup complete: $BACKUP_DIR"
echo "IMPORTANT: Store this backup OFFLINE in multiple secure locations"
```

**Storage Recommendations:**
- Hardware security module (HSM)
- Safety deposit box
- Encrypted USB drives (multiple copies)
- **NEVER** store in cloud storage unencrypted
- **NEVER** email or transmit over network

### 2. Validator State Backup (Before Every Restart)

The validator state prevents double signing. **Always backup before restarting.**

```bash
#!/bin/bash
# backup-validator-state.sh

STATE_FILE="$HOME/.aura/data/priv_validator_state.json"
BACKUP_DIR="$HOME/.aura/backups/state"

mkdir -p "$BACKUP_DIR"

# Create timestamped backup
cp "$STATE_FILE" "$BACKUP_DIR/priv_validator_state_$(date +%Y%m%d_%H%M%S).json"

# Keep only last 10 backups
ls -t "$BACKUP_DIR"/*.json | tail -n +11 | xargs -r rm

echo "Validator state backed up"
cat "$STATE_FILE"
```

### 3. Configuration Backup

```bash
#!/bin/bash
# backup-config.sh

BACKUP_DIR="/backup/aura/config/$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# Backup all configs (excluding sensitive keys)
tar --exclude='priv_validator_key.json' \
    --exclude='node_key.json' \
    -czvf "$BACKUP_DIR/config.tar.gz" \
    -C "$HOME/.aura" config/

echo "Config backup: $BACKUP_DIR/config.tar.gz"
```

### 4. Chain Data Snapshots

For fast recovery, create periodic snapshots:

```bash
#!/bin/bash
# create-snapshot.sh

# Stop node first to ensure consistent state
sudo systemctl stop aurad

SNAPSHOT_DIR="/backup/aura/snapshots"
SNAPSHOT_NAME="aura-mainnet-1_$(date +%Y%m%d_%H%M%S)"

mkdir -p "$SNAPSHOT_DIR"

# Create compressed snapshot
tar -cvf - -C "$HOME/.aura" data | lz4 > "$SNAPSHOT_DIR/$SNAPSHOT_NAME.tar.lz4"

# Generate checksum
sha256sum "$SNAPSHOT_DIR/$SNAPSHOT_NAME.tar.lz4" > "$SNAPSHOT_DIR/$SNAPSHOT_NAME.sha256"

# Restart node
sudo systemctl start aurad

echo "Snapshot created: $SNAPSHOT_DIR/$SNAPSHOT_NAME.tar.lz4"
```

### 5. Automated Backup Script

```bash
#!/bin/bash
# /opt/aura/scripts/daily-backup.sh
# Run via cron: 0 2 * * * /opt/aura/scripts/daily-backup.sh

set -e

LOG_FILE="/var/log/aura/backup.log"
BACKUP_ROOT="/backup/aura"
RETENTION_DAYS=30

log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log "Starting daily backup"

# Backup validator state
cp "$HOME/.aura/data/priv_validator_state.json" \
   "$BACKUP_ROOT/state/priv_validator_state_$(date +%Y%m%d).json"

# Backup configs
tar -czvf "$BACKUP_ROOT/config/config_$(date +%Y%m%d).tar.gz" \
    -C "$HOME/.aura" config/ 2>/dev/null

# Create weekly snapshot (Sundays only)
if [ "$(date +%u)" -eq 7 ]; then
    log "Creating weekly snapshot..."
    sudo systemctl stop aurad
    tar -cvf - -C "$HOME/.aura" data | lz4 > "$BACKUP_ROOT/snapshots/weekly_$(date +%Y%m%d).tar.lz4"
    sudo systemctl start aurad
    log "Weekly snapshot complete"
fi

# Clean old backups
find "$BACKUP_ROOT" -type f -mtime +$RETENTION_DAYS -delete

# Upload to remote storage (optional)
# aws s3 sync "$BACKUP_ROOT" s3://aura-backups/$(hostname)/

log "Daily backup complete"
```

## Recovery Procedures

### Scenario 1: Node Data Corruption

If chain data is corrupted but configs are intact:

```bash
# 1. Stop node
sudo systemctl stop aurad

# 2. Remove corrupted data
rm -rf ~/.aura/data

# 3. Recreate data directory
mkdir -p ~/.aura/data

# 4. Option A: Restore from snapshot
lz4 -d /backup/aura/snapshots/latest.tar.lz4 | tar -xvf - -C ~/.aura/

# 4. Option B: State sync from network
# Configure state sync in config.toml first
aurad start

# 5. Restore validator state (CRITICAL for validators)
cp /backup/aura/state/priv_validator_state_YYYYMMDD.json \
   ~/.aura/data/priv_validator_state.json
```

### Scenario 2: Complete Server Loss

Full recovery on new server:

```bash
# 1. Install aurad on new server
# (Follow installation docs)

# 2. Initialize node (creates directory structure)
aurad init new-validator --chain-id aura-mainnet-1

# 3. Restore validator key (from secure backup)
gpg --decrypt /secure/backup/priv_validator_key.json.gpg > \
    ~/.aura/config/priv_validator_key.json
chmod 600 ~/.aura/config/priv_validator_key.json

# 4. Restore config files
tar -xzvf /backup/aura/config/config_latest.tar.gz -C ~/.aura/

# 5. Download genesis (or restore from backup)
curl -o ~/.aura/config/genesis.json \
    https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json

# 6. CRITICAL: Restore validator state
# Get the LATEST backup to prevent double signing!
cp /backup/aura/state/priv_validator_state_LATEST.json \
   ~/.aura/data/priv_validator_state.json

# 7. Verify state height
cat ~/.aura/data/priv_validator_state.json
# Ensure height >= last signed height

# 8. Sync node (state sync recommended for speed)
# Configure state sync first, then:
sudo systemctl start aurad

# 9. Monitor sync progress
watch -n 5 'aurad status | jq ".SyncInfo"'
```

### Scenario 3: Validator Key Compromise

If your validator key may be compromised:

```bash
# 1. IMMEDIATELY stop signing
sudo systemctl stop aurad

# 2. Notify the team
# Contact security@aura.network

# 3. Unbond all delegations
aurad tx staking unbond \
    $(aurad keys show validator --bech val -a) \
    $(aurad query staking validator $(aurad keys show validator --bech val -a) --output json | jq -r '.tokens')uaura \
    --from validator \
    --chain-id aura-mainnet-1

# 4. Create new validator with new key
# Generate new key
aurad keys add new-validator

# 5. After unbonding period, re-stake with new validator
```

### Scenario 4: Double Sign Prevention During Migration

When migrating to new hardware:

```bash
# ON OLD SERVER
# 1. Stop old node
sudo systemctl stop aurad

# 2. Copy validator state (shows last signed height)
cat ~/.aura/data/priv_validator_state.json
# Note the height, round, and step

# 3. Disable auto-restart
sudo systemctl disable aurad

# ON NEW SERVER
# 4. Restore validator key
# 5. Copy exact same validator state file
scp old-server:~/.aura/data/priv_validator_state.json \
    ~/.aura/data/priv_validator_state.json

# 6. Verify state matches
cat ~/.aura/data/priv_validator_state.json

# 7. Start new node ONLY AFTER old is stopped
sudo systemctl start aurad

# WARNING: NEVER run both nodes simultaneously with same key!
```

## Disaster Recovery Testing

### Monthly DR Test

```bash
#!/bin/bash
# dr-test.sh - Run monthly to verify backups

# 1. Spin up test environment
docker run -d --name dr-test -v /tmp/dr-test:/root/.aura ubuntu:22.04 sleep infinity

# 2. Install aurad in container
docker exec dr-test bash -c "apt update && apt install -y wget && ..."

# 3. Restore from backups
docker exec dr-test bash -c "
    tar -xzvf /backup/config_latest.tar.gz -C /root/.aura/
    lz4 -d /backup/snapshot_latest.tar.lz4 | tar -xvf - -C /root/.aura/
"

# 4. Verify restoration
docker exec dr-test aurad validate-genesis
docker exec dr-test ls -la /root/.aura/

# 5. Cleanup
docker rm -f dr-test
rm -rf /tmp/dr-test
```

## Backup Verification Checklist

Weekly verification:

- [ ] Validator key backup is accessible and decryptable
- [ ] Latest validator state is backed up
- [ ] Config backups are current
- [ ] Snapshot checksums verify correctly
- [ ] Remote backup sync is working
- [ ] DR test passed

## Cloud Backup Integration

### AWS S3

```bash
# Install AWS CLI
pip install awscli

# Configure credentials
aws configure

# Sync backups to S3
aws s3 sync /backup/aura s3://aura-backups/$(hostname)/ \
    --exclude "*.json" \  # Exclude unencrypted keys
    --storage-class STANDARD_IA

# Enable versioning for recovery
aws s3api put-bucket-versioning \
    --bucket aura-backups \
    --versioning-configuration Status=Enabled
```

### GCP Cloud Storage

```bash
# Install gsutil
curl https://sdk.cloud.google.com | bash

# Sync backups
gsutil -m rsync -r /backup/aura gs://aura-backups/$(hostname)/
```

## Monitoring Backup Health

Add to Prometheus alerts:

```yaml
groups:
  - name: backups
    rules:
      - alert: BackupMissing
        expr: time() - backup_last_success_timestamp > 86400
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Backup older than 24 hours"

      - alert: ValidatorStateBackupMissing
        expr: time() - validator_state_backup_timestamp > 3600
        for: 30m
        labels:
          severity: critical
        annotations:
          summary: "Validator state not backed up in 1 hour"
```

## Emergency Contacts

- **Infrastructure Lead:** ops@aura.network
- **Security Team:** security@aura.network
- **Discord:** #validator-support
- **24/7 Emergency:** +1-XXX-XXX-XXXX
