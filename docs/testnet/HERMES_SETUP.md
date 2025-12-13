# Hermes Relayer Setup (Aura ↔ Counterparty)

This guide prepares Hermes to relay packets between the Aura local testnet (`aura-local-4`) and the dedicated counterparty chain (`aura-counter-1`). Swap the endpoints/IDs as needed if you connect to a different network.

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
id = 'aura-local-4'
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
id = 'aura-counter-1'
type = 'CosmosSdk'
rpc_addr = 'http://localhost:29657'
grpc_addr = 'http://localhost:12092'
event_source = { mode = 'push', url = 'ws://localhost:29657/websocket', batch_delay = '200ms' }
rpc_timeout = '10s'
trusted_node = false
account_prefix = 'aura'
key_name = 'relayer-counter'
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
gas_price = { price = 0.025, denom = 'stake' }
packet_filter = { policy = 'allow', list = [ ['transfer', '*'] ] }
address_type = { derivation = 'cosmos' }
```

## Key Management
```bash
# Add/import keys
hermes keys add aura-local-4    --key-name relayer-aura --mnemonic-file </secure/path/aura.txt>
hermes keys add aura-counter-1  --key-name relayer-counter --mnemonic-file </secure/path/counter.txt>

# Verify balances (replace addresses)
hermes keys list --chain aura-local-4
hermes keys list --chain aura-counter-1
```

## Funding Relayer Keys (Local Testnet)
Use the helper script to fund the Aura-side relayer key from validator-1 once the key has been imported into Hermes:

```bash
# Fund relayer-aura with 500,000 AURA (in uaura) from validator-1
./scripts/hermes-fund-keys.sh --amount 500000000000
```

Environment variables allow overrides when running on different hosts:
- `RELAYER_KEY_NAME`, `RELAYER_CHAIN_ID` – select a different Hermes key/chain.
- `VALIDATOR_KEY_NAME`, `VALIDATOR_CONTAINER` – debit a different local validator.
- `RELAYER_AMOUNT` – change the uaura amount sent.

The script will:
1. Read the relayer address via `hermes keys list`.
2. Send the requested amount from the validator container (`aura-validator-1` by default).
3. Verify the relayer balance and record the tx JSON under `logs/hermes-<key>-funding.json`.

> For external counterparties (e.g., theta-testnet), run the respective chain CLI to fund the relayer key once a reliable RPC endpoint is available.
### Counterparty Funding (aura-counter-1)
After running the counterparty init script (see below) and importing the `relayer-counter` Hermes key, fund it from the counterparty validator:

```bash
CHAIN_ID=aura-counter-1 RELAYER_CHAIN_ID=aura-counter-1 DENOM=stake \
  VALIDATOR_KEY_NAME=counterparty VALIDATOR_CONTAINER=aura-counter \
  RELAYER_KEY_NAME=relayer-counter \
  ./scripts/hermes-fund-keys.sh --amount 500000000000
```

> Until the chain exposes the `cosmos.tx.v1beta1.Service/Simulate` gRPC endpoint, `aurad tx ...` must be signed offline. Retrieve the validator account number/sequence via `aurad query account <validator> --node <rpc> --chain-id <id> --output json`, then:
> 1. `aurad tx bank send <from> <relayer> <amount><denom> ... --generate-only --gas 300000 --fees 7500000<denom> --output json > /tmp/tx.json`
> 2. `aurad tx sign /tmp/tx.json --from <key> --offline --account-number <N> --sequence <S> --chain-id <id> --output-document /tmp/signed.json`
> 3. `aurad tx broadcast /tmp/signed.json --node tcp://<rpc-host>:<rpc-port> --chain-id <id> --broadcast-mode sync --output json`
> 
> Example tx hashes: aura-local-4 → `947F37E039D053875173F7B4BD643651D9B700829B3CED6EFEEB92404C439552`, aura-counter-1 → `A4C93ECA2820D5C0465B67C76C28F63EC87CB21AF3CA62E1A6E827AC2143BEB5`.

## Local Counterparty Chain
1. Initialize (or reinitialize) the counterparty home under `testnet-data/aura-counter`:
   ```bash
   ./scripts/counterparty-init.sh
   # set COUNTERPARTY_FORCE_REINIT=1 to wipe existing data
   ```
2. Start the container:
   ```bash
   docker compose -f docker-compose.counterparty.yml up -d aura-counter
   ```
3. Import the counterparty relayer key (`relayer-counter`) into Hermes using a Vault-provided mnemonic file.
4. Fund the relayer key with the command shown above. The container exposes RPC `http://localhost:29657` and gRPC `http://localhost:12092`, matching the default Hermes config in `config/hermes/config.toml`.

Once both relayer keys are funded you can proceed with the Hermes bootstrap commands below.

## Client/Connection/Channel Creation (transfer)
```bash
# Create clients
hermes create client --host-chain aura-local-4 --reference-chain aura-counter-1
hermes create client --host-chain aura-counter-1 --reference-chain aura-local-4

# Create connection
hermes create connection --a-chain aura-local-4 --b-chain aura-counter-1

# Create transfer channel (ordered=false, port=transfer)
hermes create channel --a-chain aura-local-4 --b-chain aura-counter-1 \
  --a-port transfer --b-port transfer --order unordered --channel-version ics20-1
```

Hermes will print client/connection/channel IDs (e.g., `connection-0`, `channel-0`). Record them in your ops docs.

## Running the Relayer
```bash
# Optional helper to create clients/connection/channel via the hardened endpoints
./scripts/hermes-bootstrap.sh

# Start the relayer loop
hermes start
```

## Sanity Transfer
```bash
# Replace addresses and channel IDs
hermes tx ft-transfer --dst-chain aura-counter-1 --src-chain aura-local-4 \
  --src-port transfer --src-channel channel-0 \
  --amount 1000 --denom uaura --receiver <counterparty_address> --timeout-seconds 120
```

## Notes
- For local 4-node docker-compose with the observer/proxy stack, use RPC `http://localhost:8080/rpc` (nginx fronting `aura-observer-1`) and gRPC `http://localhost:12091`. Ensure the proxy is running (`docker-compose.proxy.yml`) so Hermes never talks directly to validators.
- Keep `gas_price` in-line with `MINIMUM_GAS_PRICES` in app.toml (default `0.025uaura`).
- If you hook Hermes to a different counterparty (e.g., theta-testnet), swap the second `[[chains]]` entry and re-import keys with the external mnemonics.
- Hermes requires the chain to expose `cosmos.tx.v1beta1.Service/Simulate` over gRPC; register the Tx service (or add a simulation bypass) before re-running `scripts/hermes-bootstrap.sh`, otherwise Hermes reports `unknown service cosmos.tx.v1beta1.Service`.
