# Getting Started with AURA Blockchain

Welcome to AURA - a Cosmos SDK-based blockchain with advanced identity verification, compliance, and cross-chain capabilities.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Running a Node](#running-a-node)
- [Creating Your First Transaction](#creating-your-first-transaction)
- [Using the CLI](#using-the-cli)
- [Next Steps](#next-steps)

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+**: [Download](https://golang.org/dl/)
- **Git**: [Download](https://git-scm.com/downloads)
- **Make**: Usually pre-installed on Unix systems
- **Docker** (optional): [Download](https://www.docker.com/get-started)

### System Requirements

- **Minimum**: 4 CPU cores, 8GB RAM, 100GB SSD
- **Recommended**: 8 CPU cores, 32GB RAM, 500GB SSD
- **OS**: Linux (Ubuntu 20.04+), macOS 11+, or Windows WSL2

## Installation

### Option 1: From Source

```bash
# Clone the repository
git clone https://github.com/aequitas/aura.git
cd aura/chain

# Build the binary
make install

# Verify installation
aurad version
```

### Option 2: Using Docker

```bash
# Pull the latest image
docker pull aequitas/aura:latest

# Run a container
docker run -it aequitas/aura:latest aurad version
```

### Option 3: Download Pre-built Binary

```bash
# Download for your platform
curl -LO https://github.com/aequitas/aura/releases/latest/download/aurad-linux-amd64
chmod +x aurad-linux-amd64
sudo mv aurad-linux-amd64 /usr/local/bin/aurad
```

## Running a Node

### Initialize Your Node

```bash
# Initialize node with a moniker (your node's name)
aurad init <your-moniker> --chain-id aura-mainnet-1

# This creates:
# ~/.aura/config/config.toml  - Tendermint configuration
# ~/.aura/config/app.toml     - Application configuration
# ~/.aura/config/genesis.json - Genesis state
```

### Download Genesis File

```bash
# For mainnet
curl https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json > ~/.aura/config/genesis.json

# For testnet
curl https://raw.githubusercontent.com/aequitas/aura/main/networks/testnet/genesis.json > ~/.aura/config/genesis.json
```

### Configure Seeds and Peers

Edit `~/.aura/config/config.toml`:

```toml
# Persistent peers
persistent_peers = "node1@seed1.aura.network:26656,node2@seed2.aura.network:26656"

# Seeds
seeds = "seed1.aura.network:26656,seed2.aura.network:26656"
```

### Start Your Node

```bash
# Start the node
aurad start

# Or use systemd for production
sudo systemctl start aurad
sudo systemctl enable aurad

# Check logs
journalctl -u aurad -f
```

### Sync Status

```bash
# Check sync status
aurad status | jq '.SyncInfo'

# Wait until "catching_up": false
```

## Creating Your First Transaction

### Create a Wallet

```bash
# Create a new wallet
aurad keys add my-wallet

# Save the mnemonic phrase securely!
# Example output:
# - address: aura1abcd1234...
# - mnemonic: word1 word2 word3...
```

### Get Testnet Tokens

For testnet, request tokens from the faucet:

```bash
# Visit https://faucet.aura.network
# Or use CLI
curl -X POST https://faucet.aura.network/request \
  -H "Content-Type: application/json" \
  -d '{"address":"aura1abcd1234..."}'
```

### Check Balance

```bash
aurad query bank balances $(aurad keys show my-wallet -a)
```

### Send Tokens

```bash
aurad tx bank send \
  my-wallet \
  aura1recipient... \
  1000000uaura \
  --chain-id aura-testnet-1 \
  --gas auto \
  --gas-adjustment 1.3 \
  --gas-prices 0.025uaura
```

## Using the CLI

### Key Management

```bash
# List all keys
aurad keys list

# Show specific key
aurad keys show my-wallet

# Export key
aurad keys export my-wallet

# Import key
aurad keys import my-wallet key.json

# Recover from mnemonic
aurad keys add recovered-wallet --recover
```

### Query Commands

```bash
# Query account info
aurad query auth account $(aurad keys show my-wallet -a)

# Query transaction
aurad query tx <TX_HASH>

# Query block
aurad query block <HEIGHT>

# Query all validators
aurad query staking validators

# Query specific module
aurad query <module> --help
```

### Transaction Commands

```bash
# Bank module - send tokens
aurad tx bank send <from> <to> <amount> --chain-id <chain-id>

# Staking - delegate tokens
aurad tx staking delegate <validator-addr> <amount> --from my-wallet

# Governance - submit proposal
aurad tx gov submit-proposal \
  --title "Title" \
  --description "Description" \
  --type Text \
  --deposit 1000000uaura \
  --from my-wallet

# Vote on proposal
aurad tx gov vote 1 yes --from my-wallet
```

### Module-Specific Commands

#### VC Registry Module

```bash
# Issue a verifiable credential
aurad tx vcregistry issue-vc \
  subject-did \
  "credential-type" \
  '{"claim":"value"}' \
  --from issuer-wallet

# Query VC
aurad query vcregistry vc <vc-id>
```

#### Bridge Module

```bash
# Initiate cross-chain transfer
aurad tx bridge transfer \
  ethereum \
  <recipient-addr> \
  1000000uaura \
  --from my-wallet

# Query transfer status
aurad query bridge transfer <transfer-id>
```

#### Compliance Module

```bash
# Submit KYC verification
aurad tx compliance submit-kyc \
  <verification-data> \
  --from my-wallet

# Query compliance status
aurad query compliance status $(aurad keys show my-wallet -a)
```

## Configuration

### Node Configuration

Edit `~/.aura/config/config.toml`:

```toml
# Set minimum gas price
minimum-gas-prices = "0.025uaura"

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

### Pruning Settings

For lower disk usage, configure pruning:

```toml
pruning = "custom"
pruning-keep-recent = "100"
pruning-keep-every = "0"
pruning-interval = "10"
```

### State Sync (Fast Sync)

To sync faster, enable state sync in `config.toml`:

```toml
[statesync]
enable = true
rpc_servers = "https://rpc1.aura.network:443,https://rpc2.aura.network:443"
trust_height = 1000000
trust_hash = "ABC123..."
trust_period = "168h0m0s"
```

## Becoming a Validator

### Prerequisites

- Fully synced node
- Minimum stake: 1,000,000 AURA
- Dedicated server with static IP
- Domain name (recommended)

### Create Validator

```bash
aurad tx staking create-validator \
  --amount=1000000uaura \
  --pubkey=$(aurad tendermint show-validator) \
  --moniker="My Validator" \
  --chain-id=aura-mainnet-1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1000000" \
  --gas="auto" \
  --gas-adjustment=1.3 \
  --gas-prices="0.025uaura" \
  --from=my-wallet
```

For complete validator setup, see [Validator Setup Guide](ops/VALIDATOR_SETUP_GUIDE.md).

## Troubleshooting

### Node Won't Start

```bash
# Check logs
aurad start --log_level debug

# Reset data (WARNING: loses all data)
aurad tendermint unsafe-reset-all
```

### Sync Issues

```bash
# Check peers
aurad tendermint show-node-id

# Add more peers in config.toml
# Ensure firewall allows port 26656
```

### Transaction Fails

```bash
# Check account sequence
aurad query auth account $(aurad keys show my-wallet -a)

# Increase gas
--gas=300000

# Check balance
aurad query bank balances $(aurad keys show my-wallet -a)
```

## Next Steps

Now that you have AURA running, explore these guides:

- [Module User Guides](modules/) - Learn about each module
- [CLI Reference](CLI_REFERENCE.md) - Complete command reference
- [API Documentation](api/) - REST and gRPC APIs
- [Validator Guide](ops/VALIDATOR_SETUP_GUIDE.md) - Run a validator
- [Developer Guide](developers/) - Build on AURA
- [Security Best Practices](SECURITY_BEST_PRACTICES.md)

## Getting Help

- **Documentation**: https://docs.aura.network
- **Discord**: https://discord.gg/aura
- **Forum**: https://forum.aura.network
- **GitHub Issues**: https://github.com/aequitas/aura/issues
- **Email**: support@aura.network

## Quick Reference Card

```bash
# Node operations
aurad init <moniker>              # Initialize node
aurad start                        # Start node
aurad status                       # Check status
aurad version                      # Show version

# Key management
aurad keys add <name>              # Create key
aurad keys list                    # List keys
aurad keys show <name>             # Show key details

# Queries
aurad query bank balances <addr>   # Check balance
aurad query tx <hash>              # Query transaction
aurad query staking validators     # List validators

# Transactions
aurad tx bank send <from> <to> <amount>  # Send tokens
aurad tx staking delegate <val> <amt>     # Delegate
aurad tx gov vote <prop-id> <option>      # Vote

# Common flags
--chain-id        # Specify chain
--from            # Sender account
--gas auto        # Auto-calculate gas
--gas-prices      # Gas price
--node            # RPC endpoint
-y                # Skip confirmation
```

Welcome to the AURA ecosystem! 🚀
