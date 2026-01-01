# Chain

This directory contains the Cosmos SDK application and the `aurad` binary.

## Build

```bash
make build
```

## Run

```bash
./build/aurad init local-node --chain-id aura-localnet
./build/aurad start --home "$HOME/.aura"
```
