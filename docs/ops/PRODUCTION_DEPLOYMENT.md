# AURA Production Deployment Guide

This guide covers deploying AURA blockchain nodes in production environments.

## Prerequisites

### Hardware Requirements

| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| CPU | 4 cores | 8+ cores | AMD64 architecture |
| RAM | 16 GB | 32 GB | More for archive nodes |
| Storage | 500 GB SSD | 1 TB NVMe | IOPS > 10,000 |
| Network | 100 Mbps | 1 Gbps | Low latency preferred |

### Software Requirements

- **OS:** Ubuntu 22.04 LTS or RHEL 8+
- **Go:** 1.21+
- **Docker:** 24.0+ (optional, for containerized deployment)
- **Kubernetes:** 1.28+ (optional, for orchestrated deployment)

## Deployment Options

### Option 1: Binary Deployment (Recommended for Validators)

#### 1. Build from Source

```bash
# Clone repository
git clone https://github.com/aequitas/aura.git
cd aura/chain

# Build optimized binary
CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=$(git describe --tags)" -o aurad ./cmd/aurad

# Verify build
./aurad version
```

#### 2. Initialize Node

```bash
# Set chain ID
export CHAIN_ID="aura-mainnet-1"
export MONIKER="your-validator-name"

# Initialize node
./aurad init $MONIKER --chain-id $CHAIN_ID

# Download genesis file
curl -o ~/.aura/config/genesis.json \
  https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json

# Verify genesis
./aurad validate-genesis
```

#### 3. Configure Node

```bash
# Download config template
curl -o ~/.aura/config/config.toml \
  https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/config.toml

# Update config with your settings
sed -i 's/moniker = .*/moniker = "'$MONIKER'"/' ~/.aura/config/config.toml
```

Edit `~/.aura/config/config.toml`:

```toml
# P2P settings
[p2p]
seeds = "seed1.aura.network:26656,seed2.aura.network:26656"
persistent_peers = ""
addr_book_strict = true
max_num_inbound_peers = 40
max_num_outbound_peers = 10

# Consensus settings
[consensus]
timeout_propose = "3s"
timeout_prevote = "1s"
timeout_precommit = "1s"
timeout_commit = "5s"

# Mempool settings
[mempool]
size = 5000
max_txs_bytes = 1073741824
cache_size = 10000
```

Edit `~/.aura/config/app.toml`:

```toml
# Minimum gas prices (required)
minimum-gas-prices = "0.025uaura"

# API settings
[api]
enable = true
swagger = false
address = "tcp://0.0.0.0:1317"

# gRPC settings
[grpc]
enable = true
address = "0.0.0.0:9090"

# State sync (optional, for fast sync)
[state-sync]
snapshot-interval = 1000
snapshot-keep-recent = 2
```

#### 4. Create Systemd Service

```bash
sudo tee /etc/systemd/system/aurad.service > /dev/null << EOF
[Unit]
Description=AURA Blockchain Node
After=network-online.target

[Service]
User=aura
Group=aura
ExecStart=/usr/local/bin/aurad start --home /home/aura/.aura
Restart=always
RestartSec=3
LimitNOFILE=65535
LimitNPROC=65535
Environment="GOGC=100"
Environment="GOMEMLIMIT=28GiB"

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable aurad
```

#### 5. Start Node

```bash
sudo systemctl start aurad

# Monitor logs
journalctl -u aurad -f
```

### Option 2: Docker Deployment

#### 1. Build Docker Image

```bash
cd aura
docker build -t aura:latest -f docker/Dockerfile .
```

#### 2. Run Container

```bash
# Create data directory
mkdir -p /opt/aura/data

# Initialize
docker run --rm -v /opt/aura/data:/root/.aura aura:latest \
  aurad init $MONIKER --chain-id $CHAIN_ID

# Download genesis
curl -o /opt/aura/data/config/genesis.json \
  https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json

# Run node
docker run -d \
  --name aura-node \
  --restart unless-stopped \
  -v /opt/aura/data:/root/.aura \
  -p 26656:26656 \
  -p 26657:26657 \
  -p 1317:1317 \
  -p 9090:9090 \
  aura:latest \
  aurad start
```

### Option 3: Kubernetes Deployment

Use the provided Kubernetes manifests:

```bash
# Deploy to production cluster
kubectl apply -k k8s/overlays/production/

# Monitor deployment
kubectl -n aura get pods -w
```

See `/k8s/overlays/production/` for full configuration.

## Network Synchronization

### Option A: Full Sync (Most Secure)

Start from genesis and sync all blocks:

```bash
./aurad start
# Takes several days for mainnet
```

### Option B: State Sync (Faster)

Sync from recent snapshot:

```bash
# Get trusted height and hash
LATEST_HEIGHT=$(curl -s https://rpc.aura.network/block | jq -r .result.block.header.height)
TRUST_HEIGHT=$((LATEST_HEIGHT - 2000))
TRUST_HASH=$(curl -s "https://rpc.aura.network/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)

# Configure state sync
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true|" ~/.aura/config/config.toml
sed -i.bak -E "s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"https://rpc.aura.network:443,https://rpc2.aura.network:443\"|" ~/.aura/config/config.toml
sed -i.bak -E "s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$TRUST_HEIGHT|" ~/.aura/config/config.toml
sed -i.bak -E "s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"|" ~/.aura/config/config.toml

./aurad start
```

### Option C: Snapshot Restore (Fastest)

Download and restore from snapshot:

```bash
# Stop node
sudo systemctl stop aurad

# Download snapshot (example)
wget https://snapshots.aura.network/aura-mainnet-1_latest.tar.lz4

# Extract
lz4 -d aura-mainnet-1_latest.tar.lz4 | tar -xvf - -C ~/.aura/

# Start node
sudo systemctl start aurad
```

## Validator Setup

### 1. Create Validator Key

```bash
# Generate new key
aurad keys add validator --keyring-backend file

# Or recover from mnemonic
aurad keys add validator --keyring-backend file --recover
```

### 2. Fund Validator Account

Transfer AURA tokens to your validator address for:
- Self-delegation (minimum recommended: 1,000,000 AURA)
- Transaction fees

### 3. Create Validator

```bash
aurad tx staking create-validator \
  --amount=1000000000000uaura \
  --pubkey=$(aurad tendermint show-validator) \
  --moniker="$MONIKER" \
  --chain-id=$CHAIN_ID \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1000000000000" \
  --gas="auto" \
  --gas-adjustment="1.5" \
  --gas-prices="0.025uaura" \
  --from=validator \
  --keyring-backend=file
```

### 4. Verify Validator

```bash
# Check validator status
aurad query staking validator $(aurad keys show validator --bech val -a)

# Check signing info
aurad query slashing signing-info $(aurad tendermint show-validator)
```

## Security Hardening

### Firewall Configuration

```bash
# Allow SSH (restrict to your IP)
sudo ufw allow from YOUR_IP to any port 22

# Allow P2P
sudo ufw allow 26656/tcp

# Allow RPC (internal only)
sudo ufw allow from 10.0.0.0/8 to any port 26657

# Allow API (internal only)
sudo ufw allow from 10.0.0.0/8 to any port 1317

# Enable firewall
sudo ufw enable
```

### Sentry Node Architecture

For validators, use sentry nodes to protect against DDoS:

```
                    ┌─────────────┐
                    │   Sentry 1   │◄──── Public P2P
                    └──────┬──────┘
                           │
    ┌─────────────┐        │        ┌─────────────┐
    │   Sentry 2   │◄──────┼───────►│   Sentry 3   │
    └──────┬──────┘        │        └──────┬──────┘
           │               │               │
           └───────────────┼───────────────┘
                           │
                    ┌──────▼──────┐
                    │  Validator   │ (private network only)
                    └─────────────┘
```

Validator config:
```toml
[p2p]
pex = false
persistent_peers = "sentry1_id@sentry1_ip:26656,sentry2_id@sentry2_ip:26656"
addr_book_strict = false
```

Sentry config:
```toml
[p2p]
pex = true
private_peer_ids = "validator_id"
unconditional_peer_ids = "validator_id"
```

### HSM Integration

For production validators, use HSM for key protection. See [HSM Integration Guide](/docs/security/HSM_INTEGRATION.md).

## Monitoring

### Prometheus Metrics

Enable metrics in `config.toml`:

```toml
[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
```

### Key Metrics to Monitor

| Metric | Alert Threshold | Description |
|--------|-----------------|-------------|
| `tendermint_consensus_height` | Stalled > 5 min | Block height |
| `tendermint_consensus_validators` | < expected | Active validators |
| `tendermint_consensus_missing_validators` | > 0 | Missing validators |
| `tendermint_p2p_peers` | < 10 | Connected peers |
| `process_resident_memory_bytes` | > 28 GB | Memory usage |

### Grafana Dashboard

Import dashboard from `/grafana/dashboards/validator.json`.

### Alerting

Configure alerts in `/prometheus/alerts/validator.yml`:

```yaml
groups:
  - name: validator
    rules:
      - alert: ValidatorMissedBlocks
        expr: increase(tendermint_consensus_validator_missed_blocks[1h]) > 100
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Validator missing blocks"

      - alert: NodeDown
        expr: up{job="aurad"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "AURA node is down"
```

## Backup and Recovery

### What to Backup

| File/Directory | Priority | Notes |
|----------------|----------|-------|
| `~/.aura/config/priv_validator_key.json` | **CRITICAL** | Validator signing key |
| `~/.aura/config/node_key.json` | High | Node identity |
| `~/.aura/config/genesis.json` | Medium | Can re-download |
| `~/.aura/data/priv_validator_state.json` | High | Prevents double signing |

### Backup Script

```bash
#!/bin/bash
BACKUP_DIR="/backup/aura/$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# Backup configs (encrypted)
tar czf - ~/.aura/config/*.json | \
  gpg --symmetric --cipher-algo AES256 > $BACKUP_DIR/config.tar.gz.gpg

# Upload to secure storage
aws s3 cp $BACKUP_DIR/config.tar.gz.gpg s3://aura-backups/
```

### Recovery

1. Install aurad on new server
2. Restore configs from backup
3. **IMPORTANT:** Stop old node before starting new one (prevent double signing)
4. Sync node using state sync or snapshot
5. Verify validator is signing

## Upgrades

### Planned Upgrades

```bash
# Download new binary
wget https://github.com/aequitas/aura/releases/download/v2.0.0/aurad-v2.0.0-linux-amd64

# Verify checksum
sha256sum aurad-v2.0.0-linux-amd64
# Compare with release notes

# Install Cosmovisor for automatic upgrades
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

# Setup upgrade
mkdir -p ~/.aura/cosmovisor/upgrades/v2.0.0/bin
mv aurad-v2.0.0-linux-amd64 ~/.aura/cosmovisor/upgrades/v2.0.0/bin/aurad
```

### Emergency Procedures

See [Emergency Procedures Runbook](/docs/runbooks/EMERGENCY_PROCEDURES.md).

## Troubleshooting

### Node Not Syncing

```bash
# Check peer count
curl localhost:26657/net_info | jq '.result.n_peers'

# Check for errors
journalctl -u aurad -n 100 --no-pager | grep -i error

# Reset and resync
aurad tendermint unsafe-reset-all
```

### High Memory Usage

```bash
# Adjust Go garbage collector
export GOGC=100
export GOMEMLIMIT=28GiB

# Prune old state
aurad prune everything
```

### Validator Jailed

```bash
# Check jail status
aurad query slashing signing-info $(aurad tendermint show-validator)

# Unjail (after fixing issue)
aurad tx slashing unjail --from validator --chain-id $CHAIN_ID
```

## Support

- **Discord:** discord.aura.network
- **Telegram:** t.me/auravalidators
- **Documentation:** docs.aura.network
- **GitHub Issues:** github.com/aequitas/aura/issues
