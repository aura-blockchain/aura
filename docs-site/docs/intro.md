---
sidebar_position: 1
---

# Introduction to AURA

AURA is a Cosmos SDK-based blockchain designed for decentralized credential verification. It enables organizations to issue, manage, and verify credentials on a trustless, transparent network.

## Key Features

- **Decentralized Verification**: Verify credentials without relying on centralized authorities
- **Privacy-Preserving**: Zero-knowledge proofs for selective disclosure
- **Interoperable**: Built on Cosmos SDK with IBC support
- **Developer-Friendly**: Comprehensive SDKs for JavaScript, Go, and mobile platforms

## Quick Start

### Prerequisites

- Go 1.23+
- Make
- Git

### Installation

```bash
# Clone the repository
git clone https://github.com/aura-blockchain/aura.git
cd aura

# Build the binary
cd chain
make build

# Verify installation
./build/aurad version
```

### Connect to Testnet

```bash
# Initialize node
aurad init my-node --chain-id aura-mvp-1

# Download genesis
curl -s https://artifacts.aurablockchain.org/aura-mvp-1/genesis.json > ~/.aura/config/genesis.json

# Add persistent peers
PEERS="f5ce5e5ce5dd77bdbfd636fb8148756f6df9c531@158.69.119.76:26681,35fdadb8b017fc95023a384c7769b946f363294e@139.99.149.160:26681"
sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" ~/.aura/config/config.toml

# Start node
aurad start
```

## Architecture

AURA consists of several custom modules:

| Module | Description |
|--------|-------------|
| `x/credential` | Core credential issuance and verification |
| `x/identity` | Decentralized identity management |
| `x/schema` | Credential schema definitions |

## Resources

- [API Reference](/docs/api)
- [Validator Guide](/docs/validators)
- [Testnet Explorer](https://testnet-explorer.aurablockchain.org)
- [Testnet Faucet](https://testnet-faucet.aurablockchain.org)
- [GitHub](https://github.com/aura-blockchain/aura)
