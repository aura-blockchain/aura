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

## Public Testnet

To join the AURA public testnet, see the [testnets repository](https://github.com/aura-blockchain/testnets) for genesis files, peer lists, and network details.

| Network | Chain ID | Status |
|---------|----------|--------|
| [aura-testnet-1](https://github.com/aura-blockchain/testnets/tree/main/aura-testnet-1) | `aura-testnet-1` | Active |

**Quick join:**
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
