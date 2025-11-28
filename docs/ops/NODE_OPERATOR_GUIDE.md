# AURA Node Operator Guide

**Version:** 1.0
**Last Updated:** 2025-11-25
**Target Audience:** Node Operators, Infrastructure Teams, RPC Providers

---

## Table of Contents

1. [Introduction](#introduction)
2. [Node Types](#node-types)
3. [Quick Start](#quick-start)
4. [Full Node Setup](#full-node-setup)
5. [Archive Node Setup](#archive-node-setup)
6. [State Sync](#state-sync)
7. [Snapshot Management](#snapshot-management)
8. [RPC Configuration](#rpc-configuration)
9. [API Endpoints](#api-endpoints)
10. [Performance Tuning](#performance-tuning)
11. [Maintenance](#maintenance)
12. [Troubleshooting](#troubleshooting)

---

## Introduction

This guide covers operating non-validator nodes on the AURA blockchain, including:
- **Full Nodes**: Maintain current state, serve RPC/API requests
- **Archive Nodes**: Maintain complete historical state
- **Seed Nodes**: Provide peer discovery
- **RPC Nodes**: Optimized for serving API requests

### Use Cases

**Full Node:**
- Wallet backends
- DApp backends
- Block explorers (recent blocks)
- Personal node for transaction submission

**Archive Node:**
- Block explorers (full history)
- Analytics platforms
- Historical data queries
- Compliance and auditing

**RPC Node:**
- Public API services
- Development environments
- Integration testing

---

## Node Types

### Full Node

**Characteristics:**
- Maintains current blockchain state
- Prunes old state periodically
- Suitable for most applications
- Lower storage requirements

**Storage Requirements:**
- Initial: ~200 GB
- Growth: ~50 GB/month
- Pruning reduces growth

### Archive Node

**Characteristics:**
- Maintains complete historical state
- Never prunes data
- Required for historical queries
- High storage requirements

**Storage Requirements:**
- Initial: ~500 GB
- Growth: ~100 GB/month
- Plan for multi-TB storage

### Seed Node

**Characteristics:**
- Provides peer discovery
- Doesn't participate in consensus
- Minimal resource requirements
- Runs in seed mode

**Storage Requirements:**
- Minimal: ~100 GB

---

## Quick Start

### Prerequisites

```bash
# Install dependencies
sudo apt update
sudo apt install -y build-essential git curl wget jq

# Install Go 1.22+
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

### Install AURA

```bash
# Clone repository
git clone https://github.com/aequitas/aura.git
cd aura/chain

# Checkout stable version
git checkout v1.0.0

# Build and install
make install

# Verify installation
aurad version
```

### Initialize Node

```bash
# Set variables
export CHAIN_ID="aura-mainnet-1"
export MONIKER="my-full-node"

# Initialize node
aurad init "$MONIKER" --chain-id "$CHAIN_ID"

# Download genesis
wget -O ~/.aura/config/genesis.json \
  https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json

# Verify genesis
aurad validate-genesis
```

### Configure and Start

```bash
# Set minimum gas prices
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.025uaura"/' \
  ~/.aura/config/app.toml

# Set seeds
SEEDS="seed1_id@seed1.aura.network:26656,seed2_id@seed2.aura.network:26656"
sed -i "s/seeds = \"\"/seeds = \"$SEEDS\"/" ~/.aura/config/config.toml

# Create systemd service
sudo tee /etc/systemd/system/aurad.service > /dev/null <<EOF
[Unit]
Description=AURA Full Node
After=network-online.target

[Service]
User=$USER
ExecStart=$(which aurad) start
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# Start node
sudo systemctl daemon-reload
sudo systemctl enable aurad
sudo systemctl start aurad

# Check logs
sudo journalctl -u aurad -f
```

---

## Full Node Setup

### Hardware Requirements

**Minimum:**
- CPU: 4 cores
- RAM: 16 GB
- Storage: 500 GB SSD
- Network: 100 Mbps

**Recommended:**
- CPU: 8 cores @ 3.0+ GHz
- RAM: 32 GB
- Storage: 1 TB NVMe SSD
- Network: 1 Gbps

### Configuration

#### config.toml

```toml
# ~/.aura/config/config.toml

#######################################################
###           Base Configuration                    ###
#######################################################

# Logging level
log_level = "info"

#######################################################
###           P2P Configuration                     ###
#######################################################
[p2p]

# Address to listen for incoming connections
laddr = "tcp://0.0.0.0:26656"

# Address to advertise to peers
external_address = ""

# Seed nodes for initial peer discovery
seeds = "seed1_id@seed1.aura.network:26656,seed2_id@seed2.aura.network:26656"

# Persistent peers (recommended to set some)
persistent_peers = ""

# Enable peer exchange
pex = true

# Maximum number of inbound peers
max_num_inbound_peers = 40

# Maximum number of outbound peers
max_num_outbound_peers = 10

#######################################################
###          Mempool Configuration                  ###
#######################################################
[mempool]

# Maximum number of transactions in the mempool
size = 5000

# Maximum size in bytes of all txs in mempool
max_txs_bytes = 1073741824  # 1GB

# Size of cache for already seen transactions
cache_size = 10000

#######################################################
###           State Sync Configuration              ###
#######################################################
[statesync]

# State sync rapidly bootstraps a new node
# Set to true to use state sync
enable = false

# RPC servers to fetch snapshots from
rpc_servers = ""

# Trust height and hash (set when enabling)
trust_height = 0
trust_hash = ""

#######################################################
###           Consensus Configuration               ###
#######################################################
[consensus]

# How long to wait for a proposal block
timeout_propose = "3s"
timeout_propose_delta = "500ms"

# How long to wait for vote
timeout_prevote = "1s"
timeout_prevote_delta = "500ms"

# How long to wait for precommit
timeout_precommit = "1s"
timeout_precommit_delta = "500ms"

# How long to wait after committing a block
timeout_commit = "6s"

#######################################################
###            RPC Server Configuration             ###
#######################################################
[rpc]

# TCP or UNIX socket address for RPC server
laddr = "tcp://0.0.0.0:26657"

# CORS allowed origins
cors_allowed_origins = []

# Enable unsafe RPC commands
unsafe = false

# Maximum number of simultaneous connections
max_open_connections = 900

# Maximum subscriptions per client
max_subscription_clients = 100
max_subscriptions_per_client = 5
```

#### app.toml

```toml
# ~/.aura/config/app.toml

#######################################################
###         Application Configuration               ###
#######################################################

# Minimum gas prices
minimum-gas-prices = "0.025uaura"

#######################################################
###           Pruning Configuration                 ###
#######################################################

# Pruning strategy (default | nothing | everything | custom)
pruning = "default"

# Applied only if pruning is custom
pruning-keep-recent = "100"
pruning-keep-every = "0"
pruning-interval = "10"

#######################################################
###         State Sync Snapshot Configuration       ###
#######################################################
[state-sync]

# Snapshot interval (blocks between snapshots)
# Set to 0 to disable snapshots
snapshot-interval = 1000

# Number of recent snapshots to keep
snapshot-keep-recent = 2

#######################################################
###                API Configuration                ###
#######################################################
[api]

# Enable API server
enable = true

# Swagger documentation
swagger = true

# API server listen address
address = "tcp://0.0.0.0:1317"

# Maximum simultaneous connections
max-open-connections = 1000

# Enable unsafe CORS
enabled-unsafe-cors = false

#######################################################
###              gRPC Configuration                 ###
#######################################################
[grpc]

# Enable gRPC server
enable = true

# gRPC server listen address
address = "0.0.0.0:9090"

#######################################################
###           gRPC Web Configuration                ###
#######################################################
[grpc-web]

# Enable gRPC-web server
enable = true

# gRPC-web server listen address
address = "0.0.0.0:9091"

# Enable unsafe CORS
enable-unsafe-cors = false

#######################################################
###            Telemetry Configuration              ###
#######################################################
[telemetry]

# Enable prometheus metrics
enabled = true

# Prometheus retention time
prometheus-retention-time = 60

# Telemetry server address
service-name = "aura-node"

# Enable hostname label
enable-hostname = true

# Enable service label
enable-service-label = true
```

### Pruning Strategies

**Default Pruning (Recommended for Full Nodes):**
```toml
pruning = "default"
# Keeps recent 100 states + every 500th state
```

**No Pruning (Archive Node):**
```toml
pruning = "nothing"
# Keeps all historical states
```

**Everything Pruning (Minimal Storage):**
```toml
pruning = "everything"
# Keeps only recent states
```

**Custom Pruning:**
```toml
pruning = "custom"
pruning-keep-recent = "100"     # Keep last 100 states
pruning-keep-every = "500"      # Keep every 500th state
pruning-interval = "10"         # Prune every 10 blocks
```

### Sync from Genesis

```bash
# Start node (will sync from block 0)
sudo systemctl start aurad

# Monitor sync progress
aurad status 2>&1 | jq .SyncInfo

# Sync can take 24-48 hours depending on chain size
```

---

## Archive Node Setup

### Hardware Requirements

**Minimum:**
- CPU: 8 cores
- RAM: 64 GB
- Storage: 4 TB SSD
- Network: 100 Mbps

**Recommended:**
- CPU: 16 cores
- RAM: 128 GB
- Storage: 8 TB NVMe SSD (RAID)
- Network: 1 Gbps

### Configuration for Archive

```toml
# app.toml

# No pruning - keep all historical state
pruning = "nothing"

# Still create snapshots for other nodes
[state-sync]
snapshot-interval = 1000
snapshot-keep-recent = 5  # Keep more snapshots

# Archive nodes often serve more API requests
[api]
enable = true
max-open-connections = 2000

[grpc]
enable = true
```

### Storage Management

```bash
# Monitor storage growth
du -sh ~/.aura/data

# Expected growth: ~100 GB/month
# Plan storage expansion accordingly

# Set up storage alerts
# Alert when < 20% free space
```

---

## State Sync

State sync allows rapid bootstrapping of a new node by downloading a recent snapshot instead of replaying all blocks.

### Enable State Sync

```bash
# Get latest block height and hash
LATEST_HEIGHT=$(curl -s https://rpc.aura.network:26657/block | jq -r .result.block.header.height)
TRUST_HEIGHT=$((LATEST_HEIGHT - 1000))
TRUST_HASH=$(curl -s "https://rpc.aura.network:26657/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)

echo "Latest Height: $LATEST_HEIGHT"
echo "Trust Height: $TRUST_HEIGHT"
echo "Trust Hash: $TRUST_HASH"

# Configure state sync in config.toml
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true|" ~/.aura/config/config.toml
sed -i.bak -E "s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"https://rpc1.aura.network:26657,https://rpc2.aura.network:26657\"|" ~/.aura/config/config.toml
sed -i.bak -E "s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$TRUST_HEIGHT|" ~/.aura/config/config.toml
sed -i.bak -E "s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"|" ~/.aura/config/config.toml

# Start node
sudo systemctl start aurad

# State sync typically completes in 10-30 minutes
sudo journalctl -u aurad -f
```

### State Sync RPC Servers

**Mainnet:**
```
https://rpc1.aura.network:26657
https://rpc2.aura.network:26657
```

**Testnet:**
```
https://rpc1-testnet.aura.network:26657
https://rpc2-testnet.aura.network:26657
```

### Troubleshooting State Sync

```bash
# If state sync fails:

# 1. Verify RPC servers are reachable
curl https://rpc1.aura.network:26657/status

# 2. Try different trust height (older snapshot)
TRUST_HEIGHT=$((LATEST_HEIGHT - 2000))

# 3. Check logs for specific errors
sudo journalctl -u aurad -n 100

# 4. If all else fails, use snapshot download instead
```

---

## Snapshot Management

### Download Snapshot

```bash
# Stop node
sudo systemctl stop aurad

# Backup priv_validator_state.json (if validator)
cp ~/.aura/data/priv_validator_state.json ~/priv_validator_state.json.backup

# Remove old data
rm -rf ~/.aura/data

# Download snapshot
cd ~/.aura
wget https://snapshots.aura.network/aura-mainnet-latest.tar.gz

# Verify checksum
wget https://snapshots.aura.network/aura-mainnet-latest.tar.gz.sha256
sha256sum -c aura-mainnet-latest.tar.gz.sha256

# Extract snapshot
tar -xzf aura-mainnet-latest.tar.gz

# Restore priv_validator_state.json (if validator)
cp ~/priv_validator_state.json.backup ~/.aura/data/priv_validator_state.json

# Start node
sudo systemctl start aurad

# Monitor sync (should start from snapshot height)
sudo journalctl -u aurad -f
```

### Create Snapshots (For Serving)

```bash
# Configure snapshot creation
# Edit app.toml
[state-sync]
snapshot-interval = 1000      # Create snapshot every 1000 blocks
snapshot-keep-recent = 5      # Keep 5 most recent snapshots

# Snapshots stored in: ~/.aura/data/snapshots/

# Export snapshot for distribution
cd ~/.aura/data
tar -czf ~/aura-snapshot-$(date +%Y%m%d).tar.gz \
  --exclude=data/cs.wal \
  application.db snapshots

# Upload to distribution server
# aws s3 cp ~/aura-snapshot-*.tar.gz s3://aura-snapshots/
```

### Automated Snapshot Script

```bash
#!/bin/bash
# create-snapshot.sh

SNAPSHOT_DIR="/snapshots"
AURA_HOME="$HOME/.aura"
DATE=$(date +%Y%m%d_%H%M)
SNAPSHOT_NAME="aura-mainnet-$DATE"

# Create snapshot directory
mkdir -p "$SNAPSHOT_DIR"

# Create compressed snapshot
cd "$AURA_HOME"
tar -czf "$SNAPSHOT_DIR/$SNAPSHOT_NAME.tar.gz" \
  --exclude=data/cs.wal \
  --exclude=data/tx_index.db \
  data/application.db data/snapshots

# Generate checksum
cd "$SNAPSHOT_DIR"
sha256sum "$SNAPSHOT_NAME.tar.gz" > "$SNAPSHOT_NAME.tar.gz.sha256"

# Upload to storage (S3, GCS, etc.)
# aws s3 cp "$SNAPSHOT_NAME.tar.gz" s3://aura-snapshots/
# aws s3 cp "$SNAPSHOT_NAME.tar.gz.sha256" s3://aura-snapshots/

# Create "latest" symlink
ln -sf "$SNAPSHOT_NAME.tar.gz" "$SNAPSHOT_DIR/aura-mainnet-latest.tar.gz"
ln -sf "$SNAPSHOT_NAME.tar.gz.sha256" "$SNAPSHOT_DIR/aura-mainnet-latest.tar.gz.sha256"

# Cleanup old snapshots (keep last 7 days)
find "$SNAPSHOT_DIR" -name "aura-mainnet-*.tar.gz" -mtime +7 -delete

echo "Snapshot created: $SNAPSHOT_NAME.tar.gz"
```

---

## RPC Configuration

### Public RPC Node

For public-facing RPC nodes, additional configuration is needed:

#### Enable CORS

```toml
# config.toml
[rpc]
cors_allowed_origins = ["*"]  # Or specific domains
cors_allowed_methods = ["HEAD", "GET", "POST"]
cors_allowed_headers = ["Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"]

# app.toml
[api]
enabled-unsafe-cors = true  # Only if needed

[grpc-web]
enable-unsafe-cors = false  # Keep disabled unless necessary
```

#### Rate Limiting

Use nginx reverse proxy for rate limiting:

```nginx
# /etc/nginx/sites-available/aura-rpc

limit_req_zone $binary_remote_addr zone=rpc_limit:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=20r/s;

upstream aura_rpc {
    server 127.0.0.1:26657;
}

upstream aura_api {
    server 127.0.0.1:1317;
}

server {
    listen 80;
    server_name rpc.yournode.com;

    # RPC endpoint
    location / {
        limit_req zone=rpc_limit burst=20 nodelay;
        proxy_pass http://aura_rpc;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}

server {
    listen 80;
    server_name api.yournode.com;

    # API endpoint
    location / {
        limit_req zone=api_limit burst=50 nodelay;
        proxy_pass http://aura_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

#### SSL/TLS with Let's Encrypt

```bash
# Install certbot
sudo apt install -y certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d rpc.yournode.com -d api.yournode.com

# Auto-renewal (already configured by certbot)
sudo certbot renew --dry-run
```

### RPC Methods

Common RPC methods exposed:

**Node Information:**
- `/status` - Node status and sync info
- `/net_info` - Network information
- `/health` - Health check endpoint

**Blockchain:**
- `/blockchain` - Blockchain information
- `/block?height=X` - Get block at height
- `/block_results?height=X` - Block results

**Transactions:**
- `/tx?hash=X` - Get transaction by hash
- `/tx_search?query=X` - Search transactions
- `/broadcast_tx_sync` - Broadcast transaction (sync)
- `/broadcast_tx_async` - Broadcast transaction (async)
- `/broadcast_tx_commit` - Broadcast transaction (commit)

**ABCI:**
- `/abci_info` - ABCI application info
- `/abci_query?path=X&data=X` - Query application state

---

## API Endpoints

### REST API

AURA exposes a REST API (Cosmos SDK standard) at port 1317:

**Account Information:**
```bash
# Get account balance
curl http://localhost:1317/cosmos/bank/v1beta1/balances/aura1...

# Get account info
curl http://localhost:1317/cosmos/auth/v1beta1/accounts/aura1...
```

**Staking:**
```bash
# Get validators
curl http://localhost:1317/cosmos/staking/v1beta1/validators

# Get delegations
curl http://localhost:1317/cosmos/staking/v1beta1/delegations/aura1...
```

**Governance:**
```bash
# Get proposals
curl http://localhost:1317/cosmos/gov/v1beta1/proposals

# Get proposal details
curl http://localhost:1317/cosmos/gov/v1beta1/proposals/1
```

**AURA Custom Modules:**
```bash
# VC Registry
curl http://localhost:1317/aura/vcregistry/v1beta1/credentials

# DEX
curl http://localhost:1317/aura/dex/v1beta1/pools

# Bridge
curl http://localhost:1317/aura/bridge/v1beta1/status
```

### gRPC

gRPC server runs on port 9090:

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:9090 list

# Query balance
grpcurl -plaintext \
  -d '{"address":"aura1..."}' \
  localhost:9090 \
  cosmos.bank.v1beta1.Query/Balance
```

### WebSocket

Subscribe to events via WebSocket (port 26657):

```javascript
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:26657/websocket');

ws.on('open', function open() {
  // Subscribe to new blocks
  ws.send(JSON.stringify({
    jsonrpc: '2.0',
    method: 'subscribe',
    params: ["tm.event='NewBlock'"],
    id: 1
  }));
});

ws.on('message', function message(data) {
  console.log('Received:', data);
});
```

---

## Performance Tuning

### Database Optimization

```toml
# app.toml

# Use faster DB backend (if compiled with support)
# goleveldb (default) | rocksdb | badgerdb
db_backend = "goleveldb"

# For RocksDB (best performance, requires build flag)
[app-db-backend]
db_backend = "rocksdb"
```

### Memory Management

```bash
# Increase node memory limits
# Edit systemd service
sudo systemctl edit aurad

# Add:
[Service]
LimitNOFILE=65535
MemoryLimit=16G  # Adjust based on RAM

# Apply changes
sudo systemctl daemon-reload
sudo systemctl restart aurad
```

### P2P Optimization

```toml
# config.toml

[p2p]
# Use more outbound peers for better sync
max_num_outbound_peers = 20

# Enable seed mode for seed nodes
seed_mode = false  # true for seed nodes

# Reduce handshake timeout
handshake_timeout = "20s"

# Dial timeout
dial_timeout = "3s"

# Maximum packet payload size
max_packet_msg_payload_size = 1024  # KB
```

### Consensus Optimization

```toml
# config.toml

[consensus]
# For non-validator nodes, can increase timeouts
timeout_commit = "6s"

# Skip timeout for faster sync
skip_timeout_commit = false

# Create empty blocks
create_empty_blocks = true
create_empty_blocks_interval = "0s"
```

### Mempool Optimization

```toml
# config.toml

[mempool]
# Reduce mempool size for lower memory usage
size = 5000

# Max transaction bytes
max_txs_bytes = 1073741824  # 1GB

# Cache size
cache_size = 10000

# Recheck transactions after blocks
recheck = true
```

### Monitoring Performance

```bash
# CPU usage
htop

# Memory usage
free -h

# Disk I/O
iostat -x 1

# Network
iftop

# Node metrics
curl localhost:26660/metrics | grep -E "(height|peers|mempool)"
```

---

## Maintenance

### Regular Maintenance Tasks

**Daily:**
```bash
# Check node status
aurad status

# Check disk space
df -h

# Check logs for errors
sudo journalctl -u aurad --since "1 day ago" | grep -i error

# Verify node is synced
CATCHING_UP=$(aurad status 2>&1 | jq -r .SyncInfo.catching_up)
if [ "$CATCHING_UP" = "true" ]; then
  echo "WARNING: Node is catching up"
fi
```

**Weekly:**
```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Check storage growth
du -sh ~/.aura/data

# Review peer connections
curl -s localhost:26657/net_info | jq '.result.peers | length'

# Check for new releases
curl -s https://api.github.com/repos/aequitas/aura/releases/latest | jq -r .tag_name
```

**Monthly:**
```bash
# Security updates
sudo apt update && sudo apt upgrade -y
sudo reboot  # During maintenance window

# Backup configuration
tar -czf ~/aura-config-backup-$(date +%Y%m%d).tar.gz ~/.aura/config

# Review and rotate logs
sudo journalctl --vacuum-time=30d
```

### Log Management

```bash
# View recent logs
sudo journalctl -u aurad -n 100

# Follow logs
sudo journalctl -u aurad -f

# Filter by severity
sudo journalctl -u aurad -p err

# Logs since timestamp
sudo journalctl -u aurad --since "2025-01-01 00:00:00"

# Limit journal size
sudo journalctl --vacuum-size=1G
```

### Cleanup and Optimization

```bash
# Clear address book (if connectivity issues)
rm ~/.aura/config/addrbook.json
sudo systemctl restart aurad

# Compact database (requires stop)
sudo systemctl stop aurad
aurad compact
sudo systemctl start aurad

# Reset to snapshot (nuclear option)
sudo systemctl stop aurad
aurad unsafe-reset-all
# Download and extract snapshot
sudo systemctl start aurad
```

---

## Troubleshooting

### Node Won't Sync

**Symptoms:**
- `catching_up: true` for extended period
- Block height not increasing

**Solutions:**
```bash
# Check peer connections
curl -s localhost:26657/net_info | jq '.result.peers | length'
# Should have > 5 peers

# If no peers, reset address book
rm ~/.aura/config/addrbook.json
sudo systemctl restart aurad

# Add known peers manually
# Edit config.toml and add persistent_peers

# Try state sync instead
# Follow state sync procedure above

# Check for fork or wrong genesis
sha256sum ~/.aura/config/genesis.json
# Compare with official
```

### Out of Memory

**Symptoms:**
- Node crashes
- `journalctl` shows OOM errors

**Solutions:**
```bash
# Increase pruning (reduce state kept)
# Edit app.toml
pruning = "everything"

# Reduce mempool size
# Edit config.toml
[mempool]
size = 1000
cache_size = 1000

# Add swap (temporary measure)
sudo fallocate -l 8G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Long-term: Upgrade RAM
```

### High Disk I/O

**Symptoms:**
- Slow sync
- High `iowait`

**Solutions:**
```bash
# Use faster storage (NVMe SSD)

# Enable database caching
# Requires rebuild with RocksDB

# Reduce state sync snapshot frequency
# Edit app.toml
snapshot-interval = 5000  # Increase interval

# Use separate disk for logs
sudo systemctl edit aurad
# Add: StandardOutput=file:/mnt/logs/aura.log
```

### RPC Not Responding

**Symptoms:**
- API timeout errors
- Unable to query node

**Solutions:**
```bash
# Check if RPC is listening
sudo netstat -tulpn | grep 26657

# Check connection limits
# Edit config.toml
[rpc]
max_open_connections = 1000

# Check if nginx/proxy is healthy
sudo systemctl status nginx

# Check firewall
sudo ufw status | grep 26657

# Restart node
sudo systemctl restart aurad
```

### Peer Connection Issues

**Symptoms:**
- Low peer count
- Frequent disconnections

**Solutions:**
```bash
# Check P2P port is open
sudo ufw status | grep 26656

# Test external connectivity
telnet your-external-ip 26656

# Add more seeds
# Edit config.toml
seeds = "seed1@host1:26656,seed2@host2:26656"

# Increase connection limits
[p2p]
max_num_inbound_peers = 50
max_num_outbound_peers = 20

# Check for IP ban
# Review logs for "banned" messages
```

### Database Corruption

**Symptoms:**
- Crash on startup
- Database errors in logs

**Solutions:**
```bash
# Stop node
sudo systemctl stop aurad

# Backup current data
mv ~/.aura/data ~/.aura/data.corrupt

# Restore from snapshot or resync
# Download latest snapshot
wget https://snapshots.aura.network/aura-mainnet-latest.tar.gz
tar -xzf aura-mainnet-latest.tar.gz -C ~/.aura/

# Or unsafe reset (full resync)
aurad unsafe-reset-all

# Start node
sudo systemctl start aurad
```

---

## Appendix

### Useful Commands

```bash
# Node status
aurad status | jq

# Validator info
aurad query staking validators

# Account balance
aurad query bank balances aura1...

# Submit transaction
aurad tx bank send from to amount --chain-id aura-mainnet-1

# Query block
aurad query block <height>

# Query transaction
aurad query tx <hash>

# Node ID
aurad tendermint show-node-id

# Validator address
aurad tendermint show-address

# Export state
aurad export > genesis_export.json
```

### Quick Reference

**Default Ports:**
- 26656: P2P
- 26657: RPC
- 26658: Prometheus metrics
- 1317: REST API
- 9090: gRPC
- 9091: gRPC-Web

**Config Files:**
- `~/.aura/config/config.toml`: Node configuration
- `~/.aura/config/app.toml`: Application configuration
- `~/.aura/config/genesis.json`: Genesis file
- `~/.aura/data/`: Blockchain data

**Log Locations:**
- Systemd: `journalctl -u aurad`
- File: `/var/log/aura/` (if configured)

---

**Document Status**: Production Ready
**Review Cycle**: Quarterly
**Next Review**: 2026-02-25
