# Aura Wallet Tools CLI

Command-line utilities for generating wallets, moving funds, staking, and querying Aura DEX telemetry. The toolset borrows the secure mnemonic-handling patterns from the PAW examples (`paw/examples/go/basic/create_wallet.go`) and the hardened workflow checklist used in the Crypto repo's HD wallet helpers (`crypto/src/xai/security/hd_wallet.py`), but is fully adapted for Aura's `aura`/`uaura` prefixes.

## Capabilities

- **Account lifecycle** – Generate new mnemonics or inspect balances over REST.
- **Banking flows** – Broadcast multi-asset transfers via the Cosmos SDK gRPC/RPC stack with fee sanity checks.
- **Staking management** – Delegate stake to validators with guardrails for chain-id mismatches.
- **DEX telemetry** – Query pool spot prices and consolidated market prices from the Aura DEX module REST endpoints so operators can validate Grafana dashboards from the CLI.

## Installation

```bash
cd wallet-tools
npm install
npm link   # optional globally-available binary
```

The CLI targets Node.js 18+ and uses ESM modules.

## Usage

```
aura-wallet --help
```

Global flags:

- `--rest` – REST/LCD endpoint (default `http://localhost:1317`)
- `--rpc` – Tendermint RPC endpoint for signing (`http://localhost:26657`)
- `--chain-id` – Expected chain ID; mismatches abort the command
- `--gas-price` – Gas price string (defaults to `0.025uaura`)

### Commands

| Command | Description |
| --- | --- |
| `account new` | Generates a 12/24-word mnemonic plus Aura address (printed once). |
| `account balance <address>` | Fetches balances via REST with optional denom filtering. |
| `bank transfer <to> --amount 1.25 --mnemonic-file ~/.aura.mnemonic` | Broadcasts a transfer from the provided mnemonic. Amounts are interpreted in display units (e.g., `1.25` AURA -> `1250000uaura`). |
| `staking delegate <validator>` | Delegates stake to a validator operator address. |
| `dex spot-price <poolId> --base uaura --quote uusdc` | Shows the instantaneous pool price. |
| `dex market-price <coin>` | Returns the aggregated market price with `sample_size` telemetry. |

### Secure key handling

- Use `--mnemonic-file` whenever possible so mnemonics never hit process args history.
- Files should be `chmod 600` or stored in an offline secrets manager.
- The CLI refuses to proceed without an explicit mnemonic source—there is no default keyring that could leak secrets inadvertently.

### Environment variables

| Variable | Purpose |
| --- | --- |
| `AURA_REST_ENDPOINT` | Overrides `--rest` default |
| `AURA_RPC_ENDPOINT` | Overrides `--rpc` default |
| `AURA_CHAIN_ID` | Default chain id |
| `AURA_GAS_PRICE` | Default gas price |

## Testing

Basic smoke test (ensures the CLI boots and prints help):

```bash
npm test
```

For end-to-end tests, export `AURA_MNEMONIC` or point to a mnemonic file and run transfer/delegate commands against a devnet.

## Future extensions

- Ledger/offline transaction signing
- Batch delegation withdrawals
- Automated DEX watchlists fed directly into the Grafana `dex` dashboard series
