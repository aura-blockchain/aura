# Aura

Aura is a Cosmos SDK chain with custom modules for identity, compliance, privacy, governance, and exchange functionality. This repository also contains the supporting services used for testnet operations (explorer, faucet, wallets, dashboards) and client SDKs.

## Build the node

```bash
cd chain
make build
./build/aurad version
```

## Run a local node

```bash
cd chain
./build/aurad init local-node --chain-id aura-localnet
./build/aurad start --home "$HOME/.aura"
```

## Join the public testnet

```bash
scripts/join-aura-testnet.sh
```

## Documentation

- `docs/GETTING_STARTED.md`
- `docs/ops/NODE_OPERATOR_GUIDE.md`
- `docs/ops/VALIDATOR_SETUP_GUIDE.md`
- `docs/ops/UPGRADE_PROCEDURES.md`
- `docs/ops/TROUBLESHOOTING.md`

## Repository layout

- `chain/` Cosmos SDK application (`aurad`)
- `contracts/` smart contracts and bindings
- `explorer/` block explorer service
- `faucet/` testnet faucet service
- `wallet/` wallet apps
- `sdk/` client SDKs
- `dashboards/` monitoring dashboards
