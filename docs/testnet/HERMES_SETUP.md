# Hermes Relayer Setup (Aura ↔ Counterparty)

This guide prepares Hermes to relay packets between the Aura testnet (`aura-testnet-1` or `aura-local-4`) and a counterparty chain (example: `gaia-testnet` or another Aura localnet).

## Prerequisites
- Hermes installed (`hermes version`)
- RPC/GRPC/WebSocket endpoints reachable for both chains
- Funded relayer keys on both chains (enough to pay fees)

## Example `~/.hermes/config.toml`
Replace endpoint hosts/ports with your actual RPC/API/GRPC services.

```toml
[global]
log_level = 'info'

[mode]
[mode.clients]
enabled = true
refresh = true
misbehaviour = true
[mode.connections]
enabled = true
[mode.channels]
enabled = true
[mode.packets]
enabled = true
clear_interval = 50
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '127.0.0.1'
port = 3000

[telemetry]
enabled = false

[[chains]]
id = 'aura-testnet-1'
type = 'CosmosSdk'
rpc_addr = 'http://localhost:27657'          # docker-compose validator-1 RPC
grpc_addr = 'http://localhost:10090'         # docker-compose validator-1 gRPC
event_source = { mode = 'push', url = 'ws://localhost:27657/websocket', batch_delay = '200ms' }
rpc_timeout = '10s'
trusted_node = false
account_prefix = 'aura'
key_name = 'relayer-aura'
store_prefix = 'ibc'
default_gas = 300000
max_gas = 800000
gas_multiplier = 1.2
max_msg_num = 15
max_tx_size = 180000
clock_drift = '10s'
max_block_time = '10s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
gas_price = { price = 0.025, denom = 'uaura' }
packet_filter = { policy = 'allow', list = [ ['transfer', '*'] ] }
address_type = { derivation = 'cosmos' }

[[chains]]
id = 'gaia-testnet'                          # replace with actual counterparty chain-id
type = 'CosmosSdk'
rpc_addr = 'https://rpc.testnet.cosmos.network:443'   # replace with real endpoint
grpc_addr = 'https://grpc.testnet.cosmos.network:443' # replace with real endpoint
event_source = { mode = 'push', url = 'wss://rpc.testnet.cosmos.network/websocket', batch_delay = '200ms' }
rpc_timeout = '10s'
trusted_node = false
account_prefix = 'cosmos'
key_name = 'relayer-gaia'
store_prefix = 'ibc'
default_gas = 300000
max_gas = 800000
gas_multiplier = 1.2
max_msg_num = 15
max_tx_size = 180000
clock_drift = '10s'
max_block_time = '10s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
gas_price = { price = 0.025, denom = 'uatom' }
packet_filter = { policy = 'allow', list = [ ['transfer', '*'] ] }
address_type = { derivation = 'cosmos' }
```

## Key Management
```bash
# Add/import keys
hermes keys add aura-testnet-1 --key-name relayer-aura --mnemonic-file ./mnemonic_aura.txt
hermes keys add gaia-testnet    --key-name relayer-gaia --mnemonic-file ./mnemonic_gaia.txt

# Verify balances (replace addresses)
hermes keys list aura-testnet-1
hermes keys list gaia-testnet
```

## Client/Connection/Channel Creation (transfer)
```bash
# Create clients
hermes create client --host-chain aura-testnet-1 --reference-chain gaia-testnet
hermes create client --host-chain gaia-testnet --reference-chain aura-testnet

# Create connection
hermes create connection --a-chain aura-testnet-1 --b-chain gaia-testnet

# Create transfer channel (ordered=false, port=transfer)
hermes create channel --a-chain aura-testnet-1 --b-chain gaia-testnet \
  --a-port transfer --b-port transfer --order unordered --channel-version ics20-1
```

Hermes will print client/connection/channel IDs (e.g., `connection-0`, `channel-0`). Record them in your ops docs.

## Running the Relayer
```bash
hermes start
```

## Sanity Transfer
```bash
# Replace addresses and channel IDs
hermes tx ft-transfer --dst-chain gaia-testnet --src-chain aura-testnet-1 \
  --src-port transfer --src-channel channel-0 \
  --amount 1000 --denom uaura --receiver <cosmos_recipient_address> --timeout-seconds 120
```

## Notes
- For local 4-node docker-compose, use RPC `http://localhost:27657` and gRPC `http://localhost:10090` (validator-1). Ensure ports are published.
- Keep `gas_price` in-line with `MINIMUM_GAS_PRICES` in app.toml (default `0.025uaura`).
- If counterparty uses different packet filter needs, adjust `packet_filter` accordingly.
