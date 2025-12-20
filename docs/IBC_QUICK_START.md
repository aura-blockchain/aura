# Aura IBC Quick Start Guide

Cross-chain communication with Inter-Blockchain Communication protocol.

## What is IBC?

IBC enables Aura to communicate with other Cosmos SDK chains:
- Transfer tokens across chains
- Cross-chain contract calls
- Interchain accounts
- Cross-chain DEX aggregation

## Prerequisites

- Aura node synced and running
- Hermes relayer (or another IBC relayer)
- Connection to target chain

## Check IBC Status

### View Existing Channels

```bash
aurad query ibc channel channels
```

### View Connections

```bash
aurad query ibc connection connections
```

### View Clients

```bash
aurad query ibc client states
```

## Cross-Chain Token Transfers

### Send Tokens to Another Chain

```bash
aurad tx ibc-transfer transfer transfer channel-0 \
  cosmos1recipient... 1000000uaura \
  --from alice \
  --chain-id aura-testnet-1 \
  --gas auto
```

### Check Transfer Status

```bash
aurad query ibc-transfer denom-traces
```

### Receive Tokens

Tokens received via IBC appear as `ibc/HASH` denominations:

```bash
aurad query bank balances $(aurad keys show alice -a)
```

## Setting Up Hermes Relayer

### 1. Install Hermes

```bash
cargo install ibc-relayer-cli --bin hermes
```

### 2. Configure Chains

Create `~/.hermes/config.toml`:

```toml
[global]
log_level = 'info'

[[chains]]
id = 'aura-testnet-1'
rpc_addr = 'http://localhost:26657'
grpc_addr = 'http://localhost:9090'
account_prefix = 'aura'
key_name = 'relayer'
gas_price = { price = 0.025, denom = 'uaura' }

[[chains]]
id = 'cosmoshub-4'
rpc_addr = 'https://rpc.cosmos.network:443'
grpc_addr = 'https://grpc.cosmos.network:443'
account_prefix = 'cosmos'
key_name = 'relayer'
gas_price = { price = 0.025, denom = 'uatom' }
```

### 3. Add Relayer Keys

```bash
hermes keys add --chain aura-testnet-1 --mnemonic-file relayer.mnemonic
hermes keys add --chain cosmoshub-4 --mnemonic-file relayer.mnemonic
```

### 4. Create Client and Connection

```bash
# Create client on both chains
hermes create client --host-chain aura-testnet-1 --reference-chain cosmoshub-4

# Create connection
hermes create connection --a-chain aura-testnet-1 --b-chain cosmoshub-4

# Create transfer channel
hermes create channel --a-chain aura-testnet-1 --a-connection connection-0 \
  --a-port transfer --b-port transfer
```

### 5. Start Relayer

```bash
hermes start
```

## Useful Queries

```bash
# Check channel state
aurad query ibc channel end transfer channel-0

# View packet commitments
aurad query ibc channel packet-commitments transfer channel-0

# Check client status
aurad query ibc client state 07-tendermint-0
```

## Troubleshooting

### Stuck Transfers

Check for pending packets:

```bash
hermes query packet pending --chain aura-testnet-1 --port transfer --channel channel-0
```

Clear stuck packets:

```bash
hermes clear packets --chain aura-testnet-1 --port transfer --channel channel-0
```

### Client Expired

Update the client:

```bash
hermes update client --host-chain aura-testnet-1 --client 07-tendermint-0
```

## Next Steps

- [DEX Quick Start](DEX_QUICK_START.md) - Trade on Aura DEX
- [Bridge Module](modules/bridge/) - Bridge to non-IBC chains
- [Hermes Documentation](https://hermes.informal.systems/)
