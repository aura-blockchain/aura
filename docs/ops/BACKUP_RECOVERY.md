# AURA Backup & Recovery Guide

**Version:** 1.0
**Last Updated:** 2025-11-25
**Target Audience:** Validators, Node Operators, DevOps

---

## Table of Contents

1. [Overview](#overview)
2. [What to Backup](#what-to-backup)
3. [Backup Strategies](#backup-strategies)
4. [Automated Backups](#automated-backups)
5. [State Snapshots](#state-snapshots)
6. [Recovery Procedures](#recovery-procedures)
7. [Disaster Recovery](#disaster-recovery)
8. [RTO/RPO Targets](#rtorpo-targets)

---

## Overview

Comprehensive backup procedures protect against:
- Hardware failures
- Data corruption
- Accidental deletion
- Security breaches
- Disaster events

**Critical Items Requiring Backup:**
1. Validator private keys (HIGHEST PRIORITY)
2. Operator wallet keys
3. Node configuration files
4. Validator state file
5. Blockchain data (optional, can resync)

---

## What to Backup

### Critical Files (MUST BACKUP)

**Validator Keys:**
```bash
~/.aura/config/priv_validator_key.json       # Validator signing key
~/.aura/data/priv_validator_state.json       # Validator state (last signed block)
~/.aura/config/node_key.json                 # P2P identity key
```

**Operator Keys:**
```bash
~/.aura/keyring-file/                        # Wallet keys (if file backend)
~/.aura/keyring-os/                          # Or OS keyring
```

**Configuration:**
```bash
~/.aura/config/config.toml                   # Node configuration
~/.aura/config/app.toml                      # Application configuration
~/.aura/config/genesis.json                  # Genesis file
```

### Optional Files

**Blockchain Data:**
```bash
~/.aura/data/application.db                  # State database (large)
~/.aura/data/blockstore.db                   # Block database (large)
~/.aura/data/cs.wal                          # Consensus WAL (don't backup)
~/.aura/data/evidence.db                     # Evidence database
~/.aura/data/state.db                        # State database
~/.aura/data/tx_index.db                     # Transaction index
```

**Note**: Blockchain data can be resynced from network, but takes time.

---

## Backup Strategies

### Strategy 1: Full Backup (Archive Nodes)

**What**: Everything including blockchain data
**Frequency**: Weekly
**Retention**: 4 weeks
**Use Case**: Archive nodes, quick restore needed

```bash
#!/bin/bash
# full-backup.sh

BACKUP_DIR="/backup/aura/full"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Full backup including data
tar -czf "$BACKUP_DIR/aura-full-$DATE.tar.gz" \
  --exclude=.aura/data/cs.wal \
  ~/.aura

# Encrypt backup
gpg --symmetric --cipher-algo AES256 \
  "$BACKUP_DIR/aura-full-$DATE.tar.gz"

# Upload to remote storage
# aws s3 cp "$BACKUP_DIR/aura-full-$DATE.tar.gz.gpg" s3://aura-backups/full/

# Cleanup old backups (keep 4 weeks)
find "$BACKUP_DIR" -name "aura-full-*.tar.gz.gpg" -mtime +28 -delete
```

### Strategy 2: Critical Files Only (Validators)

**What**: Keys + config only
**Frequency**: Daily
**Retention**: 90 days
**Use Case**: Validators (can resync blockchain)

```bash
#!/bin/bash
# critical-backup.sh

BACKUP_DIR="/backup/aura/critical"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup critical files only
tar -czf "$BACKUP_DIR/aura-critical-$DATE.tar.gz" \
  ~/.aura/config/priv_validator_key.json \
  ~/.aura/data/priv_validator_state.json \
  ~/.aura/config/node_key.json \
  ~/.aura/config/config.toml \
  ~/.aura/config/app.toml \
  ~/.aura/keyring-file

# Encrypt with strong encryption
gpg --symmetric --cipher-algo AES256 \
  --s2k-digest-algo SHA512 \
  --s2k-cipher-algo AES256 \
  --s2k-count 65536 \
  "$BACKUP_DIR/aura-critical-$DATE.tar.gz"

# Remove unencrypted backup
rm "$BACKUP_DIR/aura-critical-$DATE.tar.gz"

# Upload to multiple locations
aws s3 cp "$BACKUP_DIR/aura-critical-$DATE.tar.gz.gpg" s3://aura-backups-primary/critical/
aws s3 cp "$BACKUP_DIR/aura-critical-$DATE.tar.gz.gpg" s3://aura-backups-secondary/critical/ --region us-west-2

# Keep 90 days
find "$BACKUP_DIR" -name "aura-critical-*.tar.gz.gpg" -mtime +90 -delete

echo "Critical backup completed: $DATE"
```

### Strategy 3: Incremental Backup

**What**: Changed files only
**Frequency**: Hourly
**Retention**: 7 days
**Use Case**: High-frequency backup needs

```bash
#!/bin/bash
# incremental-backup.sh

BACKUP_DIR="/backup/aura/incremental"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Rsync with incremental backup
rsync -avz --backup --backup-dir="$BACKUP_DIR/incremental-$DATE" \
  ~/.aura/config/ \
  "$BACKUP_DIR/current/"

# Compress incremental changes
tar -czf "$BACKUP_DIR/incremental-$DATE.tar.gz" \
  "$BACKUP_DIR/incremental-$DATE"

rm -rf "$BACKUP_DIR/incremental-$DATE"

# Keep 7 days
find "$BACKUP_DIR" -name "incremental-*.tar.gz" -mtime +7 -delete
```

---

## Automated Backups

### Systemd Timer (Recommended)

```bash
# Create backup script
sudo tee /usr/local/bin/aura-backup.sh > /dev/null <<'EOF'
#!/bin/bash
# Production backup script

BACKUP_DIR="/backup/aura"
DATE=$(date +%Y%m%d_%H%M%S)
LOG_FILE="/var/log/aura-backup.log"

exec > >(tee -a "$LOG_FILE") 2>&1

echo "Starting backup: $DATE"

# Create backup directory
mkdir -p "$BACKUP_DIR/critical"

# Backup critical files
tar -czf "/tmp/aura-critical-$DATE.tar.gz" \
  -C /home/aura/.aura/config priv_validator_key.json node_key.json config.toml app.toml \
  -C /home/aura/.aura/data priv_validator_state.json

# Encrypt
gpg --batch --yes --passphrase-file /etc/aura-backup-key \
  --symmetric --cipher-algo AES256 \
  "/tmp/aura-critical-$DATE.tar.gz"

# Move to backup directory
mv "/tmp/aura-critical-$DATE.tar.gz.gpg" "$BACKUP_DIR/critical/"

# Verify backup
if [ ! -f "$BACKUP_DIR/critical/aura-critical-$DATE.tar.gz.gpg" ]; then
  echo "ERROR: Backup file not found!"
  exit 1
fi

# Upload to S3
aws s3 cp "$BACKUP_DIR/critical/aura-critical-$DATE.tar.gz.gpg" \
  s3://aura-backups/critical/$(hostname)/

# Cleanup old local backups (keep 30 days)
find "$BACKUP_DIR/critical" -name "aura-critical-*.tar.gz.gpg" -mtime +30 -delete

echo "Backup completed successfully: $DATE"
EOF

sudo chmod +x /usr/local/bin/aura-backup.sh

# Create systemd service
sudo tee /etc/systemd/system/aura-backup.service > /dev/null <<EOF
[Unit]
Description=AURA Backup Service
Wants=aura-backup.timer

[Service]
Type=oneshot
ExecStart=/usr/local/bin/aura-backup.sh
User=root

[Install]
WantedBy=multi-user.target
EOF

# Create systemd timer
sudo tee /etc/systemd/system/aura-backup.timer > /dev/null <<EOF
[Unit]
Description=AURA Backup Timer
Requires=aura-backup.service

[Timer]
OnCalendar=daily
OnCalendar=*:0/6  # Every 6 hours
Persistent=true

[Install]
WantedBy=timers.target
EOF

# Enable and start timer
sudo systemctl daemon-reload
sudo systemctl enable aura-backup.timer
sudo systemctl start aura-backup.timer

# Check timer status
sudo systemctl list-timers --all | grep aura-backup
```

### Cron Alternative

```bash
# Add to crontab
crontab -e

# Daily at 2 AM
0 2 * * * /usr/local/bin/aura-backup.sh >> /var/log/aura-backup.log 2>&1

# Every 6 hours
0 */6 * * * /usr/local/bin/aura-backup.sh >> /var/log/aura-backup.log 2>&1
```

---

## State Snapshots

### Creating Snapshots

```bash
#!/bin/bash
# create-snapshot.sh

SNAPSHOT_DIR="/snapshots/aura"
DATE=$(date +%Y%m%d_%H%M)
CHAIN_ID="aura-mainnet-1"

mkdir -p "$SNAPSHOT_DIR"

# Get current height
HEIGHT=$(aurad status 2>&1 | jq -r .SyncInfo.latest_block_height)

echo "Creating snapshot at height: $HEIGHT"

# Stop node for consistent snapshot (optional)
# sudo systemctl stop aurad

# Create snapshot
cd ~/.aura
tar -czf "$SNAPSHOT_DIR/aura-snapshot-$HEIGHT-$DATE.tar.gz" \
  --exclude=data/cs.wal \
  --exclude=data/tx_index.db \
  data/

# Resume node
# sudo systemctl start aurad

# Generate metadata
cat > "$SNAPSHOT_DIR/aura-snapshot-$HEIGHT-$DATE.json" <<EOF
{
  "chain_id": "$CHAIN_ID",
  "height": $HEIGHT,
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "size": $(stat -f%z "$SNAPSHOT_DIR/aura-snapshot-$HEIGHT-$DATE.tar.gz" 2>/dev/null || stat -c%s "$SNAPSHOT_DIR/aura-snapshot-$HEIGHT-$DATE.tar.gz"),
  "checksum": "$(sha256sum "$SNAPSHOT_DIR/aura-snapshot-$HEIGHT-$DATE.tar.gz" | awk '{print $1}')"
}
EOF

# Upload to public storage
# aws s3 cp "$SNAPSHOT_DIR/aura-snapshot-$HEIGHT-$DATE.tar.gz" \
#   s3://aura-public-snapshots/ --acl public-read

# Create latest symlink
ln -sf "$SNAPSHOT_DIR/aura-snapshot-$HEIGHT-$DATE.tar.gz" \
  "$SNAPSHOT_DIR/aura-snapshot-latest.tar.gz"

# Cleanup old snapshots (keep last 5)
ls -t "$SNAPSHOT_DIR"/aura-snapshot-*.tar.gz | tail -n +6 | xargs rm -f

echo "Snapshot created: aura-snapshot-$HEIGHT-$DATE.tar.gz"
```

---

## Recovery Procedures

### Scenario 1: Restore Validator Keys

```bash
# CRITICAL: Validator key restoration

# 1. Stop validator
sudo systemctl stop aurad

# 2. Download backup
aws s3 cp s3://aura-backups/critical/aura-critical-20250125.tar.gz.gpg /tmp/

# 3. Decrypt backup
gpg --decrypt /tmp/aura-critical-20250125.tar.gz.gpg > /tmp/aura-critical.tar.gz

# 4. Extract to temporary location
mkdir /tmp/aura-restore
tar -xzf /tmp/aura-critical.tar.gz -C /tmp/aura-restore

# 5. Verify files
ls -la /tmp/aura-restore/

# 6. Restore validator key
cp /tmp/aura-restore/priv_validator_key.json ~/.aura/config/
cp /tmp/aura-restore/priv_validator_state.json ~/.aura/data/

# 7. Set proper permissions
chmod 600 ~/.aura/config/priv_validator_key.json
chmod 600 ~/.aura/data/priv_validator_state.json

# 8. Verify key matches
aurad tendermint show-validator
# Compare with known validator pubkey

# 9. Restart validator
sudo systemctl start aurad

# 10. Monitor signing
journalctl -u aurad -f | grep "signed"

# 11. Cleanup
rm -rf /tmp/aura-restore /tmp/aura-critical*
```

### Scenario 2: Full Node Restore

```bash
# Full restore from backup

# 1. Stop node
sudo systemctl stop aurad

# 2. Backup current state (safety)
mv ~/.aura ~/.aura.old

# 3. Download and restore backup
aws s3 cp s3://aura-backups/full/aura-full-latest.tar.gz.gpg /tmp/
gpg --decrypt /tmp/aura-full-latest.tar.gz.gpg | tar -xzf - -C ~/

# 4. Verify configuration
cat ~/.aura/config/config.toml
cat ~/.aura/config/app.toml

# 5. Start node
sudo systemctl start aurad

# 6. Monitor sync
aurad status 2>&1 | jq .SyncInfo

# 7. If sync looks good, remove old backup
rm -rf ~/.aura.old
```

### Scenario 3: Restore from Snapshot

```bash
# Fast recovery using snapshot

# 1. Stop node
sudo systemctl stop aurad

# 2. Backup critical files
cp ~/.aura/config/priv_validator_key.json /tmp/
cp ~/.aura/data/priv_validator_state.json /tmp/

# 3. Remove old data
rm -rf ~/.aura/data

# 4. Download snapshot
wget https://snapshots.aura.network/aura-snapshot-latest.tar.gz -O /tmp/snapshot.tar.gz

# 5. Verify checksum
wget https://snapshots.aura.network/aura-snapshot-latest.json
# Check checksum matches

# 6. Extract snapshot
tar -xzf /tmp/snapshot.tar.gz -C ~/.aura/

# 7. Restore validator state (validators only)
cp /tmp/priv_validator_state.json ~/.aura/data/

# 8. Start node
sudo systemctl start aurad

# 9. Verify sync starting from snapshot height
aurad status 2>&1 | jq .SyncInfo
```

---

## Disaster Recovery

### DR Plan

**RTO (Recovery Time Objective):**
- Validator: < 10 minutes (critical)
- Full Node: < 1 hour
- Archive Node: < 4 hours

**RPO (Recovery Point Objective):**
- Validator Keys: 0 (real-time replication)
- Configuration: < 6 hours
- Blockchain State: 0 (resync from network)

### Geographic Redundancy

```bash
# Multi-region backup strategy

# Primary backup (same region)
aws s3 sync /backup/aura/ s3://aura-backups-us-east-1/ --region us-east-1

# Secondary backup (different region)
aws s3 sync /backup/aura/ s3://aura-backups-us-west-2/ --region us-west-2

# Tertiary backup (different cloud provider)
gsutil -m rsync -r /backup/aura/ gs://aura-backups-gcp/

# Offline backup (physical media)
# Quarterly: Copy to encrypted USB, store in bank vault
```

### Disaster Recovery Drill

```bash
#!/bin/bash
# dr-drill.sh - Disaster recovery test

echo "Starting DR drill at $(date)"

# 1. Provision new server
# (Manual or IaC)

# 2. Install dependencies
sudo apt update && sudo apt install -y build-essential git jq

# 3. Download backups
aws s3 cp s3://aura-backups-us-west-2/critical/latest.tar.gz.gpg /tmp/

# 4. Restore configuration
gpg --decrypt /tmp/latest.tar.gz.gpg | tar -xzf - -C ~/

# 5. Install aurad
wget https://github.com/aequitas/aura/releases/download/v1.0.0/aurad
chmod +x aurad && sudo mv aurad /usr/local/bin/

# 6. Download genesis
wget -O ~/.aura/config/genesis.json \
  https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json

# 7. State sync or snapshot
# Use state sync for fast recovery

# 8. Start node
aurad start --log_level info

# 9. Verify operation
# - Node syncing
# - RPC responsive
# - Validator signing (if validator)

# 10. Document time to recovery
echo "DR drill completed at $(date)"
echo "Recovery time: calculate manually"

# Rollback (don't actually use this node)
```

**Drill Frequency:** Quarterly
**Documented:** Yes
**Team Training:** Required

---

## RTO/RPO Targets

### Recovery Time Objectives

| Scenario | RTO | Notes |
|----------|-----|-------|
| Validator key restore | 5 min | From encrypted backup |
| Validator full restore | 30 min | Using state sync |
| Full node restore | 1 hour | Using snapshot |
| Archive node restore | 4 hours | Full resync or restore |
| Data center failure | 15 min | Hot standby validator |

### Recovery Point Objectives

| Data Type | RPO | Backup Frequency |
|-----------|-----|------------------|
| Validator keys | 0 | Real-time replication |
| Config files | 6 hours | Every 6 hours |
| Blockchain state | 0 | Resync from network |
| Snapshots | 24 hours | Daily |

### Backup Retention

| Backup Type | Retention | Storage |
|-------------|-----------|---------|
| Critical (encrypted) | 90 days | S3 + GCS + offline |
| Full backups | 30 days | S3 |
| Snapshots | 7 days | S3 public |
| Incremental | 7 days | Local |
| Audit logs | 1 year | S3 Glacier |

---

## Best Practices

1. **Test Restores Regularly** - Monthly restore drills
2. **Encrypt Everything** - Use strong encryption for all backups
3. **Multi-Location** - Store backups in 3+ locations
4. **Automate** - Use systemd timers or cron
5. **Monitor** - Alert on backup failures
6. **Document** - Maintain runbooks for recovery
7. **Version Control** - Track configuration changes
8. **Offline Copies** - Physical backup quarterly
9. **Access Control** - Limit who can restore keys
10. **Audit Trail** - Log all backup/restore operations

---

**Document Status**: Production Ready
**Review Cycle**: Quarterly
**Next Review**: 2026-02-25
