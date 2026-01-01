# Getting Started

This guide covers building `aurad`, running a local node, and basic CLI usage.

## Prerequisites

- Go 1.21+
- Make
- Git
- jq (for some scripts)

## Build the chain

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

## Keys

```bash
./build/aurad keys add local-key
./build/aurad keys list
```

## Transactions

Use `aurad tx` commands once the account is funded (genesis funds for local dev or testnet faucet for public testnet).

## Public testnet

See `docs/testnet/README.md` for testnet instructions and endpoints.
