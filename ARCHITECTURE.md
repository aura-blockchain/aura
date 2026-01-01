# Architecture

Aura is a Cosmos SDK application with custom modules and supporting services.

## Core components

- `chain/` Cosmos SDK app and `aurad` binary.
- `proto/` protobuf definitions for modules and services.
- `contracts/` smart contracts and bindings.
- `explorer/`, `faucet/`, `wallet/` supporting services used for testnet operations.
- `sdk/` client SDKs.

## Configuration

- Chain configuration is generated under `~/.aura/config/` after `aurad init`.
- Reference configurations live under `networks/`.
