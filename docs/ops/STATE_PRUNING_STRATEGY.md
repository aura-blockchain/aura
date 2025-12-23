# State Pruning Strategy for Production

**Version:** 1.0 | **Last Updated:** 2025-12-23

---

## Overview

State pruning controls how much historical blockchain data is retained. Proper configuration is critical for:
- **Disk usage:** Archive nodes can grow to TB+; pruned nodes stay under 100GB
- **Query capabilities:** Historical queries require retained state
- **Sync time:** Less state = faster syncs for new validators
- **Performance:** Smaller databases = faster reads/writes

---

## Pruning Options

| Mode | Retained Blocks | Use Case | Disk Growth |
|------|----------------|----------|-------------|
| `nothing` | All | Archive nodes, block explorers | ~50GB/month |
| `default` | 362,880 (~25 days) | Standard validators | ~10GB stable |
| `everything` | 2 | Exchange hot wallets | ~2GB stable |
| `custom` | Configurable | Specialized needs | Variable |

---

## Node Type Configurations

### 1. Validator Node (Recommended)

**Goal:** Minimize disk usage, maximize performance, maintain consensus participation.

```toml
# ~/.aura/config/app.toml

[pruning]
# Keep only recent state - validators don't need historical queries
pruning = "custom"
pruning-keep-recent = "1000"     # ~1.7 hours at 6s blocks
pruning-interval = "100"          # Prune every 100 blocks

[state-sync]
# Enable snapshot serving for new validators
snapshot-interval = 1000          # Create snapshot every 1000 blocks
snapshot-keep-recent = 2          # Keep last 2 snapshots
```

**Disk estimate:** 5-15GB stable after initial sync.

### 2. RPC/API Node

**Goal:** Support client queries while managing disk growth.

```toml
[pruning]
# Keep 25 days of history for queries
pruning = "default"

# Or custom for specific retention:
# pruning = "custom"
# pruning-keep-recent = "100000"  # ~7 days
# pruning-interval = "50"

[state-sync]
snapshot-interval = 2000
snapshot-keep-recent = 2
```

**Disk estimate:** 20-50GB stable.

### 3. Archive Node (Block Explorer, Analytics)

**Goal:** Full historical data access.

```toml
[pruning]
# Keep everything - required for historical queries
pruning = "nothing"

[state-sync]
# Serve snapshots for network health
snapshot-interval = 5000
snapshot-keep-recent = 5
```

**Disk estimate:** 50GB+ and growing (~1-2GB/day).

### 4. Light Client / Exchange Hot Wallet

**Goal:** Absolute minimum disk, fast syncs.

```toml
[pruning]
# Keep almost nothing - only for submitting transactions
pruning = "everything"

[state-sync]
snapshot-interval = 0             # Don't create snapshots
snapshot-keep-recent = 0
```

**Disk estimate:** 1-3GB stable.

---

## Configuration Reference

### app.toml Pruning Options

```toml
[pruning]
# Pruning strategy: default, nothing, everything, custom
pruning = "default"

# Only used when pruning = "custom":
pruning-keep-recent = "100"       # Blocks of state to keep
pruning-interval = "10"           # Prune every N blocks

# Minimum interval to avoid constant pruning overhead
# Recommended: 10-100 depending on block time
```

### State Sync Configuration

```toml
[state-sync]
# How often to take snapshots (0 = disabled)
snapshot-interval = 1000

# Number of recent snapshots to keep
snapshot-keep-recent = 2
```

---

## Migration Procedures

### From Archive to Pruned Node

```bash
# 1. Stop node
sudo systemctl stop aurad

# 2. Backup critical files
cp ~/.aura/config/priv_validator_key.json /backup/
cp ~/.aura/data/priv_validator_state.json /backup/

# 3. Update config
sed -i 's/pruning = "nothing"/pruning = "default"/' ~/.aura/config/app.toml

# 4. Option A: Keep current state, let it prune over time
sudo systemctl start aurad

# 4. Option B: Fresh start with state sync (faster)
rm -rf ~/.aura/data
# Configure state sync in config.toml
sudo systemctl start aurad
```

### From Pruned to Archive Node

```bash
# WARNING: Cannot recover pruned data
# Must resync from genesis or restore from archive backup

# 1. Stop node
sudo systemctl stop aurad

# 2. Update config
sed -i 's/pruning = "default"/pruning = "nothing"/' ~/.aura/config/app.toml

# 3. Remove data and resync
rm -rf ~/.aura/data
sudo systemctl start aurad

# Wait for full sync from genesis (can take days)
```

---

## Monitoring Disk Usage

### Prometheus Metrics

```yaml
# Add to monitoring
- job_name: 'aura-disk'
  static_configs:
    - targets: ['localhost:26660']
  metrics_path: /metrics
```

**Key metrics:**
- `tendermint_store_size_bytes`: Total store size
- `tendermint_state_block_height`: Current height
- `tendermint_consensus_latest_block_height`: Sync status

### Disk Alerts

```yaml
# Prometheus alert rules
groups:
  - name: aura-disk
    rules:
      - alert: HighDiskUsage
        expr: node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} < 0.15
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Disk space below 15%"

      - alert: CriticalDiskUsage
        expr: node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} < 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Disk space below 5% - pruning may be needed"
```

### Manual Check

```bash
# Check Aura data directory size
du -sh ~/.aura/data/

# Check individual databases
du -sh ~/.aura/data/*.db

# Check disk usage trend
watch -n 60 'du -sh ~/.aura/data/'
```

---

## Recommended Network Topology

For a healthy network, maintain this distribution:

| Node Type | Count | Pruning | State Sync |
|-----------|-------|---------|------------|
| Validators | 50+ | custom/default | Serve snapshots |
| RPC Nodes | 5-10 | default | Serve snapshots |
| Archive Nodes | 2-3 | nothing | Serve snapshots |
| Seed Nodes | 3-5 | default | Serve snapshots |

---

## Emergency Disk Recovery

If disk is full:

```bash
# 1. Stop node immediately
sudo systemctl stop aurad

# 2. Clear WAL (safe to delete)
rm -rf ~/.aura/data/cs.wal/

# 3. Clear evidence.db (optional, can rebuild)
rm -rf ~/.aura/data/evidence.db/

# 4. If still full, switch to aggressive pruning
sed -i 's/pruning = "default"/pruning = "everything"/' ~/.aura/config/app.toml

# 5. Restart
sudo systemctl start aurad

# 6. Monitor for pruning progress
watch 'du -sh ~/.aura/data/'
```

---

## Best Practices

1. **Validators:** Use `custom` with 1000 recent blocks
2. **RPCs:** Use `default` for reasonable query support
3. **Archives:** Use `nothing` but provision 500GB+ disk
4. **Monitor:** Set up disk alerts before 80% usage
5. **Snapshots:** Enable on all non-light nodes
6. **Backups:** Backup keys before any pruning changes
7. **Network Health:** Maintain at least 2 archive nodes

---

## Reference: Aura Module Store Sizes

Approximate sizes per module (at 100K blocks):

| Module | Typical Size | Notes |
|--------|-------------|-------|
| bank | 2-5 GB | Account balances |
| staking | 500MB-2GB | Validator state |
| dex | 1-10 GB | Order books, trades |
| bridge | 500MB-2GB | Transfer records |
| identity | 1-5 GB | DIDs, credentials |
| compliance | 500MB-1GB | KYC records |
| governance | 100-500MB | Proposals, votes |
| Other | 2-5 GB | Combined |

**Total with default pruning:** 10-30GB
**Total archive (no pruning):** 50GB+ (growing)
