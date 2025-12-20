---
sidebar_position: 1
---

# Validator Setup

Learn how to set up and run a validator node on the Aura blockchain.

## Prerequisites

### Hardware Requirements

**Minimum:**
- 4 CPU cores
- 8GB RAM
- 200GB SSD storage
- 100 Mbps network

**Recommended:**
- 8+ CPU cores
- 32GB RAM
- 500GB+ NVMe SSD
- 1 Gbps network
- Dedicated server with static IP

### Software Requirements

- Ubuntu 20.04+ or similar Linux distribution
- Go 1.21+
- Aura binary (aurad)
- Cosmovisor (for automated upgrades)

## Initial Setup

### 1. Install Aura

```bash
# Clone repository
git clone https://github.com/aura-blockchain/aura.git
cd aura/chain

# Build and install
make install

# Verify installation
aurad version
```

### 2. Initialize Node

```bash
# Initialize node
aurad init <your-moniker> --chain-id aura-mainnet-1

# Download genesis file
curl https://raw.githubusercontent.com/aura-blockchain/networks/main/mainnet/genesis.json \
  > ~/.aura/config/genesis.json

# Verify genesis
sha256sum ~/.aura/config/genesis.json
```

### 3. Configure Node

Edit `~/.aura/config/config.toml`:

```toml
# Set minimum gas prices
minimum-gas-prices = "0.025uaura"

# Set persistent peers
persistent_peers = "node1@peer1.aura.network:26656,node2@peer2.aura.network:26656"

# Set seeds
seeds = "seed1.aura.network:26656,seed2.aura.network:26656"

# Enable Prometheus metrics
prometheus = true
```

Edit `~/.aura/config/app.toml`:

```toml
# Pruning configuration
pruning = "custom"
pruning-keep-recent = "100"
pruning-keep-every = "0"
pruning-interval = "10"

# Enable API
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"

# Enable gRPC
[grpc]
enable = true
address = "0.0.0.0:9090"
```

### 4. Create Validator Key

```bash
# Create validator wallet
aurad keys add validator

# IMPORTANT: Save mnemonic securely!
# Backup to secure location (encrypted USB, hardware wallet, etc.)

# Fund validator wallet
# Transfer minimum stake amount to validator address
```

### 5. Start Node and Sync

```bash
# Start node
aurad start

# Check sync status
aurad status | jq '.SyncInfo.catching_up'

# Wait until catching_up is false
```

## Create Validator

Once your node is fully synced:

```bash
# Create validator
aurad tx staking create-validator \
  --amount=1000000uaura \
  --pubkey=$(aurad tendermint show-validator) \
  --moniker="<your-moniker>" \
  --chain-id=aura-mainnet-1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1000000" \
  --gas="auto" \
  --gas-adjustment=1.5 \
  --gas-prices="0.025uaura" \
  --from=validator
```

### Verify Validator

```bash
# Get validator address
aurad keys show validator --bech val

# Query validator info
aurad query staking validator $(aurad keys show validator --bech val -a)
```

## Security Best Practices

### Key Management

- Store mnemonic offline in secure location
- Use hardware wallets for production
- Never expose private keys
- Enable firewall rules

### Network Security

```bash
# Configure firewall
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 26656/tcp # P2P
sudo ufw deny 26657/tcp  # RPC (only allow from monitoring)
sudo ufw deny 1317/tcp   # API (only allow from monitoring)
sudo ufw enable
```

### Sentry Node Architecture

Use sentry nodes to protect your validator:

```
Internet
    |
Sentry Nodes (Public)
    |
Validator Node (Private)
```

## Monitoring Setup

Install monitoring tools:

```bash
# Node exporter for metrics
wget https://github.com/prometheus/node_exporter/releases/download/v1.7.0/node_exporter-1.7.0.linux-amd64.tar.gz
tar xvfz node_exporter-1.7.0.linux-amd64.tar.gz
sudo mv node_exporter-1.7.0.linux-amd64/node_exporter /usr/local/bin/
sudo useradd -rs /bin/false node_exporter

# Create systemd service
sudo nano /etc/systemd/system/node_exporter.service
```

See [Monitoring Guide](/docs/validators/monitoring) for complete setup.

## Automated Upgrades with Cosmovisor

### Install Cosmovisor

```bash
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest
```

### Configure Cosmovisor

```bash
# Create directory structure
mkdir -p ~/.aura/cosmovisor/genesis/bin
mkdir -p ~/.aura/cosmovisor/upgrades

# Copy current binary
cp $(which aurad) ~/.aura/cosmovisor/genesis/bin/

# Set environment variables
export DAEMON_NAME=aurad
export DAEMON_HOME=$HOME/.aura
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false
export DAEMON_RESTART_AFTER_UPGRADE=true
```

### Create Systemd Service

```bash
sudo nano /etc/systemd/system/aurad.service
```

```ini
[Unit]
Description=Aura Validator
After=network-online.target

[Service]
User=<your-user>
ExecStart=/home/<your-user>/go/bin/cosmovisor run start
Restart=always
RestartSec=3
LimitNOFILE=4096
Environment="DAEMON_NAME=aurad"
Environment="DAEMON_HOME=/home/<your-user>/.aura"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"

[Install]
WantedBy=multi-user.target
```

### Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable aurad
sudo systemctl start aurad
sudo systemctl status aurad
```

## Common Operations

### Check Validator Status

```bash
# Validator info
aurad query staking validator $(aurad keys show validator --bech val -a)

# Signing info
aurad query slashing signing-info $(aurad tendermint show-validator)
```

### Unjail Validator

```bash
aurad tx slashing unjail \
  --from=validator \
  --chain-id=aura-mainnet-1 \
  --gas=auto \
  --gas-prices=0.025uaura
```

### Edit Validator

```bash
aurad tx staking edit-validator \
  --moniker="new-moniker" \
  --website="https://yourwebsite.com" \
  --details="Validator description" \
  --from=validator \
  --chain-id=aura-mainnet-1
```

## Backup and Recovery

### Backup Important Files

```bash
# Backup validator key
cp ~/.aura/config/priv_validator_key.json ~/validator-key-backup.json

# Backup node key
cp ~/.aura/config/node_key.json ~/node-key-backup.json

# Store backups securely offline
```

### Recovery

```bash
# Restore from backup
cp ~/validator-key-backup.json ~/.aura/config/priv_validator_key.json
cp ~/node-key-backup.json ~/.aura/config/node_key.json

# Restart node
sudo systemctl restart aurad
```

## Troubleshooting

### Node not syncing

```bash
# Check peers
aurad tendermint show-node-id

# Add more peers in config.toml
# Check network connectivity
ping peer1.aura.network
```

### Validator jailed

Check slashing info:
```bash
aurad query slashing signing-info $(aurad tendermint show-validator)
```

Common reasons:
- Downtime (missed blocks)
- Double signing
- Network issues

## Resources

- [Validator Guide](https://github.com/aura-blockchain/aura/blob/main/docs/validators/)
- [Monitoring Setup](/docs/validators/monitoring)
- [Upgrade Procedures](/docs/validators/upgrades)

## Support

- [Discord - Validator Channel](https://discord.gg/aura)
- [Validator Forum](https://forum.aura.network/c/validators)
- [Emergency Contact](mailto:validators@aura.network)
