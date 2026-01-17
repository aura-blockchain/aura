# Aura Blockchain Daemon (aurad)

The official command-line interface and daemon for the Aura blockchain network.

## Overview

`aurad` is the main entry point for running an Aura blockchain node. It provides a comprehensive CLI for:

- **Node Operations**: Initialize and start blockchain nodes
- **Key Management**: Create and manage cryptographic keys
- **Transactions**: Create, sign, and broadcast transactions
- **Queries**: Query blockchain state and module-specific data
- **Governance**: Participate in on-chain governance

## Installation

### Building from Source

```bash
# Clone the repository
git clone https://github.com/aequitas/aura
cd aura/chain

# Build the binary
make build

# Or install to $GOPATH/bin
make install
```

### Using Pre-built Binaries

Download the latest release from the [releases page](https://github.com/aequitas/aura/releases).

## Quick Start

### 1. Initialize a Node

```bash
# Initialize with default settings
aurad init my-node

# Initialize with custom chain ID
aurad init my-node --chain-id aura-mvp-1
```

This creates the necessary configuration files in `~/.aura/`:
- `config/config.toml` - Node configuration
- `config/app.toml` - Application configuration
- `data/` - Blockchain data directory

### 2. Start the Node

```bash
# Start with default settings
aurad start

# Start with custom addresses
aurad start --grpc-address localhost:9090 --api-address localhost:1317
```

### 3. Check Node Status

```bash
aurad status
```

## Commands

### Node Management

```bash
# Initialize node
aurad init [moniker]

# Start node
aurad start

# Check node status
aurad status

# Display version
aurad version
```

### Key Management

```bash
# Add a new key
aurad keys add my-key

# List all keys
aurad keys list

# Show key details
aurad keys show my-key

# Delete a key
aurad keys delete my-key
```

### Querying

```bash
# Query block
aurad query block [height]

# Query transaction
aurad query tx [hash]

# Query account
aurad query account [address]

# Module-specific queries
aurad query identitychange params
aurad query vcregistry credential [id]
aurad query confidencescore score [address]
aurad query governance proposal [id]
```

### Transactions

```bash
# Identity Change Module
aurad tx identitychange register [did]
aurad tx identitychange update [did]
aurad tx identitychange deactivate [did]

# VC Registry Module
aurad tx vcregistry issue [credential-json]
aurad tx vcregistry revoke [credential-id]

# Confidence Score Module
aurad tx confidencescore update [address] [score]

# Governance Module
aurad tx governance submit-proposal [proposal-json]
aurad tx governance vote [proposal-id] [vote]
aurad tx governance deposit [proposal-id] [amount]
```

## Configuration

### Configuration Files

The configuration files are located in `~/.aura/config/`:

#### config.toml

Main node configuration including:
- Chain ID and moniker
- RPC server settings
- P2P networking configuration
- Consensus parameters

#### app.toml

Application-level configuration including:
- gRPC server settings
- API server settings
- Logging configuration
- Module-specific settings

### Environment Variables

All configuration values can be overridden using environment variables with the `AURA_` prefix:

```bash
export AURA_HOME=/custom/path
export AURA_LOG_LEVEL=debug
export AURA_LOG_FORMAT=json
```

### Command-line Flags

Global flags available for all commands:

```bash
--home         # Home directory (default: ~/.aura)
--config       # Config file path
--log-level    # Logging level (trace|debug|info|warn|error|fatal|panic)
--log-format   # Logging format (json|plain)
```

## Modules

Aura blockchain includes the following modules:

### Identity Change Module
Manages decentralized identifiers (DIDs) and identity lifecycle operations.

### Inclusion Routines Module
Handles inclusion routine proposals and voting mechanisms.

### Confidence Score Module
Calculates and manages confidence scores for addresses based on participation.

### VC Registry Module
Manages issuance, verification, and revocation of verifiable credentials.

### Data Registry Module
Handles registration and management of data on-chain.

### Governance Module
Provides on-chain governance with proposal submission and voting.

### DEX Module (Planned)
Decentralized exchange functionality (requires full Cosmos SDK integration).

### Bridge Module (Planned)
Cross-chain bridge functionality (requires full Cosmos SDK integration).

## Development

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linters
make lint
```

### Version Information

Build with version information:

```bash
# The Makefile automatically injects version info
make build

# Check version
./build/aurad version
```

## Architecture

```
cmd/aurad/
├── main.go              # Entry point
└── cmd/                 # Command implementations
    ├── root.go          # Root command and configuration
    ├── init.go          # Node initialization
    ├── start.go         # Node startup
    ├── status.go        # Status queries
    ├── version.go       # Version information
    ├── keys.go          # Key management
    ├── query.go         # Query commands
    └── tx.go            # Transaction commands
```

## Network Endpoints

When a node is running, the following endpoints are available:

- **gRPC**: `localhost:9090` - For gRPC queries and transactions
- **API**: `localhost:1317` - REST API endpoints
- **RPC**: `localhost:26657` - Tendermint RPC

## Troubleshooting

### Node won't start

1. Check if the node is already running
2. Verify configuration files in `~/.aura/config/`
3. Check logs for error messages

### Connection refused errors

1. Verify the node is running: `aurad status`
2. Check that ports are not blocked by firewall
3. Ensure correct addresses in configuration

### Permission errors

1. Verify home directory permissions
2. Ensure write access to data directory

## Support

- **Documentation**: [https://docs.aura.network](https://docs.aura.network)
- **Discord**: [https://discord.gg/RwQ8pma6](https://discord.gg/RwQ8pma6)
- **GitHub Issues**: [https://github.com/aequitas/aura/issues](https://github.com/aequitas/aura/issues)

## License

Copyright 2024 Aura Blockchain

Licensed under the Apache License, Version 2.0.
