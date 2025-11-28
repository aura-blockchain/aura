# AURA Blockchain Production Deployment Guide

**Version:** 1.0
**Last Updated:** 2025-11-25
**Target Audience:** DevOps Engineers, System Administrators, Node Operators

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Hardware Requirements](#hardware-requirements)
4. [Software Prerequisites](#software-prerequisites)
5. [Network Topology](#network-topology)
6. [Security Best Practices](#security-best-practices)
7. [Deployment Steps](#deployment-steps)
8. [Post-Deployment Verification](#post-deployment-verification)
9. [Monitoring Setup](#monitoring-setup)
10. [Backup Procedures](#backup-procedures)
11. [Disaster Recovery](#disaster-recovery)
12. [Appendix](#appendix)

---

## Overview

AURA is a sophisticated blockchain platform built on Cosmos SDK with 24+ custom modules providing advanced functionality including:

- **Identity & Privacy**: Verifiable Credentials (VCs), privacy-preserving transactions, confidential computing
- **DeFi**: Decentralized Exchange (DEX) with liquidity pools, HTLC support
- **Cross-chain**: Bridge module with Merkle proof validation
- **Smart Contracts**: CosmWasm integration with custom security bindings
- **Governance**: On-chain governance with proposal voting
- **Security**: Advanced validator security, wallet security, network security, cryptography modules
- **Compliance**: KYC/AML compliance tracking, GDPR support
- **Monitoring**: AI-powered anomaly detection, security event tracking

This guide covers deploying AURA for production use on both testnet and mainnet environments.

### Key Features

- **Consensus**: CometBFT (Tendermint) with Byzantine Fault Tolerance
- **Block Time**: ~6 seconds (configurable)
- **Finality**: Instant finality
- **State Machine**: Cosmos SDK v0.50+
- **VM**: CosmWasm for smart contracts
- **Network**: Peer-to-peer with gossip protocol

---

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     AURA Network Layer                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │  Validator   │◄──►│  Validator   │◄──►│  Validator   │  │
│  │   Node 1     │    │   Node 2     │    │   Node 3     │  │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘  │
│         │                   │                   │            │
│         └───────────┬───────┴───────────┬───────┘            │
│                     │                   │                    │
│         ┌───────────▼─────┐ ┌───────────▼─────┐             │
│         │  Sentry Node 1  │ │  Sentry Node 2  │             │
│         └───────────┬─────┘ └───────────┬─────┘             │
│                     │                   │                    │
│     ┌───────────────┼───────────────────┼────────────┐      │
│     │               │                   │            │      │
│ ┌───▼────┐    ┌────▼─────┐      ┌─────▼────┐  ┌───▼────┐ │
│ │ Full   │    │  Full    │      │  Archive │  │  Seed  │ │
│ │ Node 1 │    │  Node 2  │      │   Node   │  │  Node  │ │
│ └────────┘    └──────────┘      └──────────┘  └────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                          │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Cosmos   │  │ AURA     │  │ CosmWasm │  │ Custom   │   │
│  │ Modules  │  │ Modules  │  │ Runtime  │  │ Bindings │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                          │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │Prometheus│  │ Grafana  │  │   ELK    │  │  Backup  │   │
│  │          │  │Dashboards│  │   Stack  │  │  System  │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### Node Types

1. **Validator Nodes**: Participate in consensus, sign blocks
2. **Sentry Nodes**: Protect validators from DDoS attacks
3. **Full Nodes**: Maintain full state, serve RPC/API requests
4. **Archive Nodes**: Maintain complete historical state
5. **Seed Nodes**: Provide peer discovery for the network

---

## Hardware Requirements

### Validator Nodes

**Minimum Requirements:**
- **CPU**: 8 cores (3.0+ GHz)
- **RAM**: 32 GB
- **Storage**: 1 TB NVMe SSD
- **Network**: 100 Mbps symmetric, low latency (<50ms to peers)
- **Uptime**: 99.9%+ availability

**Recommended Production:**
- **CPU**: 16 cores (3.5+ GHz) AMD EPYC or Intel Xeon
- **RAM**: 64 GB ECC
- **Storage**: 2 TB NVMe SSD (RAID 1 for redundancy)
- **Network**: 1 Gbps symmetric, dedicated line
- **Backup**: Redundant power supply, UPS
- **Location**: Tier 3+ data center

### Sentry Nodes

**Minimum:**
- **CPU**: 8 cores
- **RAM**: 16 GB
- **Storage**: 500 GB SSD
- **Network**: 100 Mbps

**Recommended:**
- **CPU**: 8-16 cores
- **RAM**: 32 GB
- **Storage**: 1 TB NVMe SSD
- **Network**: 1 Gbps with DDoS protection

### Full Nodes (RPC/API)

**Minimum:**
- **CPU**: 4 cores
- **RAM**: 16 GB
- **Storage**: 500 GB SSD
- **Network**: 100 Mbps

**Recommended:**
- **CPU**: 8 cores
- **RAM**: 32 GB
- **Storage**: 1 TB SSD
- **Network**: 1 Gbps

### Archive Nodes

**Minimum:**
- **CPU**: 8 cores
- **RAM**: 64 GB
- **Storage**: 4 TB SSD
- **Network**: 100 Mbps

**Recommended:**
- **CPU**: 16 cores
- **RAM**: 128 GB
- **Storage**: 8+ TB NVMe SSD array
- **Network**: 1 Gbps

### Seed Nodes

**Minimum:**
- **CPU**: 2 cores
- **RAM**: 4 GB
- **Storage**: 100 GB SSD
- **Network**: 100 Mbps

---

## Software Prerequisites

### Operating System

**Recommended:**
- Ubuntu 22.04 LTS or 24.04 LTS
- Debian 12
- RHEL 9 / Rocky Linux 9

**Minimum Kernel:** 5.15+

### Required Software

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install build essentials
sudo apt install -y build-essential git curl wget jq

# Install Go 1.22+
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Verify Go installation
go version  # Should be 1.22.0 or higher
```

### System Tuning

```bash
# Increase file descriptors
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Network tuning
sudo tee /etc/sysctl.d/99-aura.conf > /dev/null <<EOF
# Network performance
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.tcp_rmem = 4096 87380 67108864
net.ipv4.tcp_wmem = 4096 65536 67108864
net.core.netdev_max_backlog = 5000
net.ipv4.tcp_congestion_control = bbr

# Connection tracking
net.netfilter.nf_conntrack_max = 1048576
net.nf_conntrack_max = 1048576

# File system
fs.file-max = 2097152
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
EOF

sudo sysctl -p /etc/sysctl.d/99-aura.conf
```

### Time Synchronization

```bash
# Install and configure chrony for time sync (critical for consensus)
sudo apt install -y chrony

sudo tee /etc/chrony/chrony.conf > /dev/null <<EOF
pool time.google.com iburst
pool time.cloudflare.com iburst
pool pool.ntp.org iburst

makestep 1.0 3
rtcsync
EOF

sudo systemctl restart chrony
sudo systemctl enable chrony

# Verify time sync
chronyc tracking
```

---

## Network Topology

### Recommended Validator Setup

For production validators, implement a **sentry node architecture**:

```
                    Internet
                       │
        ┌──────────────┼──────────────┐
        │              │              │
    ┌───▼────┐    ┌───▼────┐    ┌───▼────┐
    │Sentry 1│    │Sentry 2│    │Sentry 3│
    │(Public)│    │(Public)│    │(Public)│
    └───┬────┘    └───┬────┘    └───┬────┘
        │             │             │
        │    Private Network        │
        │             │             │
        └─────────────┼─────────────┘
                      │
                 ┌────▼────┐
                 │Validator│
                 │(Private)│
                 └─────────┘
```

**Benefits:**
- DDoS protection (validator not directly exposed)
- Connection filtering
- Geographic distribution
- Failover capability

### Network Configuration

**Validator Node (Private):**
```toml
# config.toml
[p2p]
laddr = "tcp://0.0.0.0:26656"
pex = false  # Disable peer exchange
persistent_peers = "sentry1_id@10.0.1.10:26656,sentry2_id@10.0.1.11:26656"
private_peer_ids = ""  # Leave empty
addr_book_strict = true
max_num_inbound_peers = 5
max_num_outbound_peers = 5
```

**Sentry Nodes (Public):**
```toml
# config.toml
[p2p]
laddr = "tcp://0.0.0.0:26656"
pex = true  # Enable peer exchange
persistent_peers = "validator_id@10.0.1.1:26656,other_sentries"
private_peer_ids = "validator_node_id"  # Protect validator
unconditional_peer_ids = "validator_node_id"
addr_book_strict = false
max_num_inbound_peers = 100
max_num_outbound_peers = 50
```

### Firewall Rules

**Validator Node:**
```bash
# Only allow connections from sentry nodes
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow from 10.0.1.10 to any port 26656 proto tcp
sudo ufw allow from 10.0.1.11 to any port 26656 proto tcp
sudo ufw allow from 10.0.1.0/24 to any port 22 proto tcp
sudo ufw enable
```

**Sentry/Full Nodes:**
```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC (if public)
sudo ufw allow 1317/tcp   # REST API (if public)
sudo ufw allow 9090/tcp   # gRPC (if public)
sudo ufw allow 22/tcp     # SSH (restrict to your IPs)
sudo ufw enable
```

---

## Security Best Practices

### Key Management

1. **Generate keys on air-gapped machine**
2. **Use Hardware Security Module (HSM)** for validator keys (production)
3. **Encrypt key files** at rest
4. **Backup keys** to multiple secure locations
5. **Implement key rotation** procedures

### Access Control

```bash
# Create dedicated user
sudo useradd -m -s /bin/bash aura
sudo usermod -aG sudo aura

# Set proper permissions
sudo chown -R aura:aura /home/aura/.aura
sudo chmod 700 /home/aura/.aura
sudo chmod 600 /home/aura/.aura/config/*
```

### SSH Hardening

```bash
# /etc/ssh/sshd_config
sudo tee -a /etc/ssh/sshd_config > /dev/null <<EOF
PermitRootLogin no
PubkeyAuthentication yes
PasswordAuthentication no
PermitEmptyPasswords no
X11Forwarding no
MaxAuthTries 3
ClientAliveInterval 300
ClientAliveCountMax 2
EOF

sudo systemctl restart sshd
```

### Enable Fail2ban

```bash
sudo apt install -y fail2ban

sudo tee /etc/fail2ban/jail.local > /dev/null <<EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
EOF

sudo systemctl restart fail2ban
sudo systemctl enable fail2ban
```

### DDoS Protection

For public-facing nodes (sentries, RPC):

1. **Use CloudFlare or similar** for RPC endpoints
2. **Implement rate limiting** at load balancer
3. **Configure connection limits** in config.toml
4. **Monitor for attack patterns**

### Security Monitoring

- Enable audit logging (`auditd`)
- Monitor failed authentication attempts
- Track sudo usage
- Alert on configuration changes
- Review logs daily

---

## Deployment Steps

### Step 1: Build AURA Binary

```bash
# Clone repository
cd $HOME
git clone https://github.com/aequitas/aura.git
cd aura/chain

# Checkout specific version (use tagged release for production)
git checkout v1.0.0  # Replace with actual version

# Build binary
make install

# Verify installation
aurad version --long
```

### Step 2: Initialize Node

```bash
# Set chain ID
export CHAIN_ID="aura-mainnet-1"  # or "aura-testnet-1"
export MONIKER="your-node-name"

# Initialize node
aurad init "$MONIKER" --chain-id "$CHAIN_ID"

# This creates ~/.aura directory with:
# - config/config.toml
# - config/app.toml
# - config/genesis.json (template)
# - data/
```

### Step 3: Download Genesis File

```bash
# For Mainnet
wget -O ~/.aura/config/genesis.json https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json

# For Testnet
wget -O ~/.aura/config/genesis.json https://raw.githubusercontent.com/aequitas/aura/main/networks/testnet/genesis.json

# Verify genesis file
aurad validate-genesis
sha256sum ~/.aura/config/genesis.json
# Compare with official checksum
```

### Step 4: Configure Node

#### config.toml

```bash
# Edit ~/.aura/config/config.toml

# Set seeds (for initial peer discovery)
seeds = "seed1_id@seed1.aura.network:26656,seed2_id@seed2.aura.network:26656"

# Set persistent peers (for validators, set to your sentries)
persistent_peers = ""

# Set mempool size
[mempool]
size = 5000
max_txs_bytes = 1073741824  # 1GB

# Set consensus timeouts
[consensus]
timeout_propose = "3s"
timeout_propose_delta = "500ms"
timeout_prevote = "1s"
timeout_prevote_delta = "500ms"
timeout_precommit = "1s"
timeout_precommit_delta = "500ms"
timeout_commit = "6s"

# State sync (for fast sync)
[statesync]
enable = false  # Enable for new nodes to sync quickly
rpc_servers = ""
trust_height = 0
trust_hash = ""
```

#### app.toml

```bash
# Edit ~/.aura/config/app.toml

# Minimum gas prices
minimum-gas-prices = "0.025uaura"

# Pruning (adjust based on node type)
pruning = "default"  # "default", "nothing" (archive), "everything"

# State sync snapshots (for serving snapshots to other nodes)
[state-sync]
snapshot-interval = 1000  # Take snapshot every 1000 blocks
snapshot-keep-recent = 2

# API Configuration
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"

# gRPC Configuration
[grpc]
enable = true
address = "0.0.0.0:9090"

# Telemetry
[telemetry]
enabled = true
prometheus-retention-time = 60
```

### Step 5: Set Up Systemd Service

```bash
sudo tee /etc/systemd/system/aurad.service > /dev/null <<EOF
[Unit]
Description=AURA Node
After=network-online.target

[Service]
User=aura
ExecStart=$(which aurad) start
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable aurad
```

### Step 6: Sync Blockchain

**Option A: Sync from Genesis (Slow)**

```bash
sudo systemctl start aurad
sudo journalctl -u aurad -f
```

**Option B: State Sync (Fast)**

```bash
# Get trust height and hash from RPC
LATEST_HEIGHT=$(curl -s https://rpc.aura.network:26657/block | jq -r .result.block.header.height)
TRUST_HEIGHT=$((LATEST_HEIGHT - 1000))
TRUST_HASH=$(curl -s "https://rpc.aura.network:26657/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)

# Update config.toml
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true|" ~/.aura/config/config.toml
sed -i.bak -E "s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"https://rpc1.aura.network:26657,https://rpc2.aura.network:26657\"|" ~/.aura/config/config.toml
sed -i.bak -E "s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$TRUST_HEIGHT|" ~/.aura/config/config.toml
sed -i.bak -E "s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"|" ~/.aura/config/config.toml

# Start node
sudo systemctl start aurad
sudo journalctl -u aurad -f
```

**Option C: Snapshot Download (Fastest)**

```bash
# Stop node if running
sudo systemctl stop aurad

# Download and extract snapshot
cd ~/.aura
wget https://snapshots.aura.network/aura-mainnet-latest.tar.gz
tar -xzf aura-mainnet-latest.tar.gz

# Start node
sudo systemctl start aurad
```

### Step 7: Create Validator (Validators Only)

**Prerequisites:**
- Node fully synced
- Sufficient AURA tokens staked
- Validator key secured

```bash
# Create validator transaction
aurad tx staking create-validator \
  --amount=1000000uaura \
  --pubkey=$(aurad tendermint show-validator) \
  --moniker="$MONIKER" \
  --chain-id="$CHAIN_ID" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --from=validator-key \
  --gas=auto \
  --gas-adjustment=1.4 \
  --fees=5000uaura

# Verify validator is active
aurad query staking validator $(aurad keys show validator-key --bech val -a)
```

---

## Post-Deployment Verification

### Health Checks

```bash
# Check node status
aurad status

# Check sync status
aurad status 2>&1 | jq .SyncInfo

# Check validator info (if validator)
aurad query staking validator $(aurad tendermint show-validator)

# Check peer connections
curl -s http://localhost:26657/net_info | jq .result.n_peers
```

### Performance Checks

```bash
# Check block time
curl -s http://localhost:26657/status | jq -r .result.sync_info

# Check transaction throughput
aurad query txs --limit 100

# Check memory usage
free -h
htop

# Check disk usage
df -h
du -sh ~/.aura/data
```

### Security Checks

```bash
# Verify firewall rules
sudo ufw status

# Check open ports
sudo netstat -tulpn | grep aurad

# Verify SSH configuration
sudo sshd -T | grep -E "^(permitrootlogin|passwordauthentication)"

# Check fail2ban status
sudo fail2ban-client status sshd
```

---

## Monitoring Setup

AURA includes built-in Prometheus metrics and Grafana dashboards.

### Prometheus Configuration

```bash
# Install Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvfz prometheus-2.45.0.linux-amd64.tar.gz
cd prometheus-2.45.0.linux-amd64

# Configure Prometheus
tee prometheus.yml > /dev/null <<EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

# Load alerting rules
rule_files:
  - "monitoring-alerts.yml"

# Scrape configurations
scrape_configs:
  - job_name: 'aura-node'
    static_configs:
      - targets: ['localhost:26660']  # Tendermint metrics
        labels:
          instance: 'validator-1'

  - job_name: 'aura-app'
    static_configs:
      - targets: ['localhost:1317']  # AURA metrics
        labels:
          instance: 'validator-1'
EOF

# Copy alert rules from repository
cp /path/to/aura/prometheus/rules/monitoring-alerts.yml .

# Run Prometheus
./prometheus --config.file=prometheus.yml
```

### Grafana Setup

```bash
# Install Grafana
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
sudo apt-get update
sudo apt-get install grafana

# Start Grafana
sudo systemctl start grafana-server
sudo systemctl enable grafana-server

# Access Grafana at http://localhost:3000 (admin/admin)
# Import dashboard from /path/to/aura/grafana/dashboards/security-monitoring.json
```

### Key Metrics to Monitor

1. **Consensus Metrics:**
   - Block height
   - Block time
   - Missing blocks (validators)
   - Validator voting power

2. **Network Metrics:**
   - Peer count
   - Network congestion
   - Inbound/outbound connections

3. **Security Metrics:**
   - Failed transactions
   - Anomaly detections
   - Large transaction alerts
   - Gossip duplicate rate

4. **System Metrics:**
   - CPU usage
   - Memory usage
   - Disk I/O
   - Network bandwidth

### Alert Configuration

Reference existing alert rules at `/prometheus/rules/monitoring-alerts.yml`:

- **Critical**: Validator down, network congestion, security threats
- **High**: Low consensus health, validator jailed, high anomaly rate
- **Medium**: Gas price spikes, rapid TVL changes
- **Warning**: High mempool size, sync lag

---

## Backup Procedures

### What to Backup

1. **Validator Keys** (critical):
   - `~/.aura/config/priv_validator_key.json`
   - `~/.aura/data/priv_validator_state.json`

2. **Node Configuration**:
   - `~/.aura/config/config.toml`
   - `~/.aura/config/app.toml`
   - `~/.aura/config/node_key.json`

3. **Wallet Keys**:
   - `~/.aura/keyring-file/` (or keyring backend directory)

4. **State Data** (optional, can re-sync):
   - `~/.aura/data/`

### Backup Script

```bash
#!/bin/bash
# backup-aura.sh

BACKUP_DIR="/backup/aura"
DATE=$(date +%Y%m%d_%H%M%S)
AURA_HOME="$HOME/.aura"

# Create backup directory
mkdir -p "$BACKUP_DIR/$DATE"

# Backup validator keys (encrypted)
tar -czf - "$AURA_HOME/config/priv_validator_key.json" \
  "$AURA_HOME/data/priv_validator_state.json" | \
  gpg --symmetric --cipher-algo AES256 -o "$BACKUP_DIR/$DATE/validator_keys.tar.gz.gpg"

# Backup configuration
cp "$AURA_HOME/config/config.toml" "$BACKUP_DIR/$DATE/"
cp "$AURA_HOME/config/app.toml" "$BACKUP_DIR/$DATE/"
cp "$AURA_HOME/config/node_key.json" "$BACKUP_DIR/$DATE/"

# Backup wallet keyring
if [ -d "$AURA_HOME/keyring-file" ]; then
  tar -czf "$BACKUP_DIR/$DATE/keyring.tar.gz" "$AURA_HOME/keyring-file"
fi

# Optional: Backup state snapshot
# tar -czf "$BACKUP_DIR/$DATE/data.tar.gz" "$AURA_HOME/data"

# Upload to remote storage (S3, GCS, etc.)
# aws s3 sync "$BACKUP_DIR/$DATE" "s3://aura-backups/$DATE/"

# Retention: Keep last 30 days
find "$BACKUP_DIR" -type d -mtime +30 -exec rm -rf {} +

echo "Backup completed: $BACKUP_DIR/$DATE"
```

### Automated Backups

```bash
# Add to crontab
crontab -e

# Backup daily at 2 AM
0 2 * * * /home/aura/backup-aura.sh >> /var/log/aura-backup.log 2>&1
```

### Backup Verification

```bash
# Test restore procedure quarterly
gpg --decrypt validator_keys.tar.gz.gpg | tar -xzf -
# Verify files extracted correctly
```

---

## Disaster Recovery

### Recovery Scenarios

#### Scenario 1: Node Corruption

```bash
# Stop node
sudo systemctl stop aurad

# Restore from backup
cd ~/.aura
rm -rf data config

# Restore configuration
tar -xzf /backup/aura/latest/config.tar.gz

# Restore keys
gpg --decrypt /backup/aura/latest/validator_keys.tar.gz.gpg | tar -xzf -

# Re-sync blockchain (use state sync or snapshot)
# ... follow sync procedures from deployment steps

# Start node
sudo systemctl start aurad
```

#### Scenario 2: Validator Key Compromise

```bash
# IMMEDIATE ACTIONS:
# 1. Stop validator immediately
sudo systemctl stop aurad

# 2. Alert network validators via governance channels

# 3. Generate new validator key on air-gapped machine
aurad init recovery-validator --chain-id aura-mainnet-1

# 4. Submit validator update transaction
aurad tx staking edit-validator \
  --new-pubkey=$(aurad tendermint show-validator) \
  --from=validator-key \
  --chain-id=aura-mainnet-1

# 5. Update all configurations with new key
# 6. Restart with new key
# 7. Monitor for double-sign attempts
```

#### Scenario 3: Data Center Failure

**Prerequisites**: Geographic redundancy with hot standby

```bash
# Standby node should be synced and ready
# Switch DNS/load balancer to standby
# For validators: coordinate failover to avoid double-signing

# On standby node:
# 1. Ensure fully synced
aurad status | jq .SyncInfo.catching_up  # Should be false

# 2. Copy validator state from primary (if not already done)
# 3. Start validator
sudo systemctl start aurad

# 4. Verify signing
aurad query slashing signing-info $(aurad tendermint show-validator)
```

### Recovery Time Objectives (RTO)

- **Validator Node**: < 5 minutes (with hot standby)
- **Full Node**: < 30 minutes (from snapshot)
- **Archive Node**: < 2 hours (from backup)

### Recovery Point Objectives (RPO)

- **Validator Keys**: 0 data loss (encrypted backups)
- **Configuration**: 0 data loss (daily backups)
- **Blockchain State**: 0 data loss (can re-sync)

---

## Appendix

### A. Useful Commands

```bash
# Check node status
aurad status | jq

# Query account balance
aurad query bank balances <address>

# Send transaction
aurad tx bank send <from> <to> <amount> --chain-id aura-mainnet-1

# Query validator
aurad query staking validator <validator-address>

# Check consensus state
curl http://localhost:26657/consensus_state

# Export state
aurad export > genesis_export.json

# Reset node
aurad unsafe-reset-all
```

### B. Network Endpoints

**Mainnet:**
- RPC: `https://rpc.aura.network:26657`
- API: `https://api.aura.network:1317`
- gRPC: `grpc.aura.network:9090`
- Explorer: `https://explorer.aura.network`

**Testnet:**
- RPC: `https://rpc-testnet.aura.network:26657`
- API: `https://api-testnet.aura.network:1317`
- gRPC: `grpc-testnet.aura.network:9090`
- Explorer: `https://testnet-explorer.aura.network`

### C. Chain Parameters

- **Chain ID**: `aura-mainnet-1` (mainnet), `aura-testnet-1` (testnet)
- **Bech32 Prefix**: `aura`
- **Native Denom**: `uaura` (1 AURA = 1,000,000 uaura)
- **Block Time**: ~6 seconds
- **Block Size**: 22 MB
- **Unbonding Period**: 21 days
- **Slashing**: 5% (double sign), 0.01% (downtime)

### D. Support Resources

- **Documentation**: https://docs.aura.network
- **Discord**: https://discord.gg/aura
- **GitHub**: https://github.com/aequitas/aura
- **Forum**: https://forum.aura.network
- **Security**: security@aura.network

### E. Version History

| Version | Date | Changes |
|---------|------|---------|
| v1.0.0 | 2025-11-25 | Initial production release |

---

**Document Status**: Production Ready
**Review Cycle**: Quarterly
**Next Review**: 2026-02-25
