# Aura

Aura is a Cosmos SDK chain with custom modules for identity, compliance, privacy, governance, and exchange functionality. This repository also contains the supporting services used for testnet operations (explorer, faucet, wallets, dashboards) and client SDKs.

## Build the Node

### Full Build (All Modules)

```bash
cd chain
make build
./build/aurad version
```

### MVP Build (12 Essential Modules)

The MVP release includes only the core modules needed for credential verification:

```bash
cd chain
make build-mvp
./build/aurad-mvp version
```

**MVP Modules:** auth, bank, staking, slashing, distribution, governance, identity, vcregistry, dataregistry, compliance, prevalidation, wasm

See [docs/MVP_MODULES.md](docs/MVP_MODULES.md) for full MVP documentation.

## Run a local node

```bash
cd chain
./build/aurad init local-node --chain-id aura-localnet
./build/aurad start --home "$HOME/.aura"
```

## Devnet

To join the AURA devnet, see the [networks repository](https://github.com/aura-blockchain/testnets) for genesis files, peer lists, and network details.

| Network | Chain ID | Status |
|---------|----------|--------|
| [aura-mvp-1](https://github.com/aura-blockchain/testnets/tree/main/aura-mvp-1) | `aura-mvp-1` | Devnet |

**Quick join:**
```bash
scripts/join-aura-testnet.sh
```

## Documentation

- `docs/MVP_MODULES.md` - MVP module reference
- `docs/GETTING_STARTED.md` - Quick start guide
- `docs/ops/NODE_OPERATOR_GUIDE.md` - Running a node
- `docs/ops/VALIDATOR_SETUP_GUIDE.md` - Validator setup
- `docs/ops/UPGRADE_PROCEDURES.md` - Upgrade procedures
- `docs/ops/TROUBLESHOOTING.md` - Common issues

## Release Versions

| Version | Type | Modules | Status |
|---------|------|---------|--------|
| v1.0.0-mvp | MVP | 12 | Current |
| v2.0.0 | Full | 28+ | Planned |

## Repository layout

- `chain/` Cosmos SDK application (`aurad`)
- `contracts/` smart contracts and bindings
- `explorer/` block explorer service
- `faucet/` testnet faucet service
- `wallet/` wallet apps
- `sdk/` client SDKs
- `dashboards/` monitoring dashboards
