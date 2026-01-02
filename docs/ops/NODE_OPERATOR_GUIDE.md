# Node Operator Guide

This guide covers running a non-validator full node.

## Build and configure

```bash
cd chain
make build
```

Initialize and configure the node:

```bash
./build/aurad init <moniker> --chain-id aura-testnet-1
```

For testnet, the quickest path is:

```bash
scripts/join-aura-testnet.sh
```

## Start the node

```bash
./build/aurad start --home "$HOME/.aura"
```

## Check sync status

```bash
curl -s http://localhost:26657/status | jq '.result.sync_info'
```

## Logs

Run in a terminal or configure systemd for long-running nodes.

## Hermes relayer (example)

If you operate IBC relayers, start from the example config at `config/hermes/config.toml.example` and update endpoints and keys for your environment.
