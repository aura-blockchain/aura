---
sidebar_position: 3
---

# Validator Guide

This guide covers running an AURA validator node on the testnet.

## Requirements

### Hardware

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 4 cores | 8 cores |
| RAM | 16 GB | 32 GB |
| Storage | 500 GB SSD | 1 TB NVMe |
| Network | 100 Mbps | 1 Gbps |

### Software

- Ubuntu 22.04 LTS
- Go 1.23+
- Cosmovisor (recommended)

## Installation

### Build from Source

```bash
# Clone repository
git clone https://github.com/aura-blockchain/aura.git
cd aura/chain

# Build
make build

# Install
sudo cp build/aurad /usr/local/bin/
```

### Initialize Node

```bash
# Initialize
aurad init my-validator --chain-id aura-mvp-1

# Download genesis
curl -s https://artifacts.aurablockchain.org/aura-mvp-1/genesis.json > ~/.aura/config/genesis.json

# Configure peers
PEERS="f5ce5e5ce5dd77bdbfd636fb8148756f6df9c531@158.69.119.76:26681,35fdadb8b017fc95023a384c7769b946f363294e@139.99.149.160:26681"
sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" ~/.aura/config/config.toml
```

## State Sync

For faster syncing, use state sync:

```bash
# Get state sync parameters
RPC="https://testnet-rpc.aurablockchain.org"
LATEST=$(curl -s "$RPC/block" | jq -r '.result.block.header.height')
TRUST_HEIGHT=$((LATEST - 2000))
TRUST_HASH=$(curl -s "$RPC/block?height=$TRUST_HEIGHT" | jq -r '.result.block_id.hash')

# Configure state sync
sed -i "s/enable = false/enable = true/" ~/.aura/config/config.toml
sed -i "s|rpc_servers = \"\"|rpc_servers = \"$RPC,$RPC\"|" ~/.aura/config/config.toml
sed -i "s/trust_height = 0/trust_height = $TRUST_HEIGHT/" ~/.aura/config/config.toml
sed -i "s/trust_hash = \"\"/trust_hash = \"$TRUST_HASH\"/" ~/.aura/config/config.toml
```

## Create Validator

### Create Wallet

```bash
# Create new key
aurad keys add validator

# Or recover existing
aurad keys add validator --recover
```

### Get Testnet Tokens

Visit the [testnet faucet](https://testnet-faucet.aurablockchain.org) to get tokens.

### Submit Create Validator Transaction

```bash
aurad tx staking create-validator \
  --amount=1000000uaura \
  --pubkey=$(aurad tendermint show-validator) \
  --moniker="my-validator" \
  --chain-id=aura-mvp-1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --from=validator \
  --gas=auto \
  --gas-adjustment=1.5 \
  --fees=5000uaura
```

## Cosmovisor Setup

```bash
# Install cosmovisor
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

# Set environment
export DAEMON_NAME=aurad
export DAEMON_HOME=$HOME/.aura

# Create directories
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
mkdir -p $DAEMON_HOME/cosmovisor/upgrades

# Copy binary
cp $(which aurad) $DAEMON_HOME/cosmovisor/genesis/bin/

# Create systemd service
sudo tee /etc/systemd/system/aurad.service > /dev/null <<EOF
[Unit]
Description=AURA Node
After=network.target

[Service]
Type=simple
User=$USER
ExecStart=$(which cosmovisor) run start
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
Environment="DAEMON_NAME=aurad"
Environment="DAEMON_HOME=$HOME/.aura"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable aurad
sudo systemctl start aurad
```

## Monitoring

### Check Sync Status

```bash
aurad status 2>&1 | jq '.SyncInfo'
```

### View Logs

```bash
journalctl -u aurad -f
```

### Check Validator Status

```bash
aurad query staking validator $(aurad keys show validator --bech val -a)
```
