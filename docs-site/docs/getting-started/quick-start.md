---
sidebar_position: 2
---

# Quick Start

Get up and running with Aura in under 10 minutes. This guide will walk you through starting a local node and performing your first transactions.

## Start a Local Node

### Initialize Node

First, initialize your node with a moniker (your node's name):

```bash
# Initialize node
aurad init my-node --chain-id aura-testnet-1

# This creates configuration files in ~/.aura/
```

### Configure Genesis

For a local development node:

```bash
# Create a key for your validator
aurad keys add my-wallet

# Save the mnemonic phrase securely!
# Output shows: address: aura1...

# Add genesis account
aurad genesis add-genesis-account my-wallet 100000000000uaura

# Create genesis transaction
aurad genesis gentx my-wallet 1000000000uaura \
  --chain-id aura-testnet-1 \
  --moniker my-node

# Collect genesis transactions
aurad genesis collect-gentxs
```

### Start the Node

```bash
# Start the node
aurad start

# You should see blocks being produced
# Leave this terminal running
```

## Connect to Testnet

To connect to the public testnet instead of running a local node:

### Download Genesis

```bash
# Initialize node
aurad init my-node --chain-id aura-testnet-1

# Download testnet genesis
curl https://raw.githubusercontent.com/aura-blockchain/networks/main/testnet/genesis.json \
  > ~/.aura/config/genesis.json
```

### Configure Peers

Edit `~/.aura/config/config.toml`:

```toml
# Add persistent peers
persistent_peers = "node1@seed1.testnet.aura.network:26656,node2@seed2.testnet.aura.network:26656"

# Add seeds
seeds = "seed1.testnet.aura.network:26656,seed2.testnet.aura.network:26656"
```

### Start Syncing

```bash
# Start the node
aurad start

# Check sync status (in another terminal)
aurad status | jq '.SyncInfo'

# Wait until "catching_up": false
```

## Create Your First Wallet

```bash
# Create a new wallet
aurad keys add my-wallet

# IMPORTANT: Save the mnemonic phrase!
# Example output:
# - address: aura1abcd1234...
# - mnemonic: word1 word2 word3...

# View your address
aurad keys show my-wallet -a

# List all keys
aurad keys list
```

## Get Testnet Tokens

### Using the Faucet

Visit the testnet faucet to get tokens:

```bash
# Get your address
MY_ADDRESS=$(aurad keys show my-wallet -a)

# Request tokens from faucet
curl -X POST https://faucet.testnet.aura.network/request \
  -H "Content-Type: application/json" \
  -d "{\"address\":\"$MY_ADDRESS\"}"
```

### Check Your Balance

```bash
# Query your balance
aurad query bank balances $(aurad keys show my-wallet -a)

# Expected output:
# balances:
# - amount: "10000000"
#   denom: uaura
```

## Perform Your First Transaction

### Send Tokens

```bash
# Send 1 AURA (1000000 uaura) to another address
aurad tx bank send \
  my-wallet \
  aura1recipient... \
  1000000uaura \
  --chain-id aura-testnet-1 \
  --gas auto \
  --gas-adjustment 1.3 \
  --gas-prices 0.025uaura \
  --yes

# Check transaction status
aurad query tx <TX_HASH>
```

### Delegate to a Validator

```bash
# List validators
aurad query staking validators

# Delegate tokens
aurad tx staking delegate \
  auravaloper1... \
  5000000uaura \
  --from my-wallet \
  --chain-id aura-testnet-1 \
  --gas auto \
  --gas-prices 0.025uaura \
  --yes
```

## Interact with Identity Module

### Issue a Verifiable Credential

```bash
# Issue a credential (requires appropriate permissions)
aurad tx vcregistry issue-vc \
  did:aura:subject123 \
  "ProofOfHumanity" \
  '{"verified":true}' \
  --from my-wallet \
  --chain-id aura-testnet-1 \
  --gas auto \
  --gas-prices 0.025uaura \
  --yes
```

### Query Credentials

```bash
# Query a specific credential
aurad query vcregistry credential <vc-id>

# List credentials for a DID
aurad query vcregistry credentials-by-subject did:aura:subject123
```

## Using the REST API

The REST API provides an HTTP interface to the chain:

```bash
# Enable API in config/app.toml
# [api]
# enable = true
# address = "tcp://0.0.0.0:1317"

# Query account
curl http://localhost:1317/cosmos/auth/v1beta1/accounts/aura1...

# Query balance
curl http://localhost:1317/cosmos/bank/v1beta1/balances/aura1...

# Query validators
curl http://localhost:1317/cosmos/staking/v1beta1/validators
```

## Using gRPC

For programmatic access:

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:9090 list

# Query balance
grpcurl -plaintext \
  -d '{"address":"aura1..."}' \
  localhost:9090 \
  cosmos.bank.v1beta1.Query/Balance
```

## Docker Quickstart

Run a complete local testnet with Docker Compose:

```bash
# Clone repository
git clone https://github.com/aura-blockchain/aura.git
cd aura

# Start testnet (3 validators)
docker-compose -f docker-compose.testnet.yml up -d

# View logs
docker-compose -f docker-compose.testnet.yml logs -f

# Access validator container
docker exec -it aura-validator-1 bash

# Stop testnet
docker-compose -f docker-compose.testnet.yml down
```

## Common Commands Reference

### Node Management

```bash
aurad start                    # Start node
aurad status                   # Check node status
aurad version                  # Show version
```

### Key Management

```bash
aurad keys add <name>          # Create new key
aurad keys list                # List all keys
aurad keys show <name>         # Show key details
aurad keys delete <name>       # Delete key
```

### Queries

```bash
aurad query bank balances <address>      # Check balance
aurad query tx <hash>                    # Query transaction
aurad query staking validators           # List validators
aurad query gov proposals                # List proposals
```

### Transactions

```bash
aurad tx bank send <from> <to> <amount>        # Send tokens
aurad tx staking delegate <val> <amount>       # Delegate
aurad tx gov vote <proposal-id> <option>       # Vote
```

## Troubleshooting

### Node won't start

```bash
# Check logs
aurad start --log_level debug

# Reset node (WARNING: deletes all data)
aurad tendermint unsafe-reset-all
```

### Transaction fails

```bash
# Check account sequence
aurad query auth account $(aurad keys show my-wallet -a)

# Increase gas
--gas 300000

# Check sufficient balance
aurad query bank balances $(aurad keys show my-wallet -a)
```

### Sync issues

```bash
# Check peers
aurad tendermint show-node-id

# Verify genesis file
sha256sum ~/.aura/config/genesis.json

# Enable state sync for faster sync
# See State Sync guide
```

## Next Steps

Now that you have a running node, explore:

- [Module Development](/docs/developers/module-development) - Build custom modules
- [SDK Integration](/docs/developers/sdk-integration) - Integrate Aura into your app
- [Validator Setup](/docs/validators/setup) - Run a production validator
- [Network Parameters](https://docs.aura.network/network-parameters) - Understanding chain configuration

## Need Help?

- [Discord](https://discord.gg/aura) - Community support
- [Documentation](https://docs.aura.network) - Full docs
- [GitHub Issues](https://github.com/aura-blockchain/aura/issues) - Report bugs
- [Forum](https://forum.aura.network) - Discussions
