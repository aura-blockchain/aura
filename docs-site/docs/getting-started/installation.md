---
sidebar_position: 1
---

# Installation

This guide will help you install and set up the Aura blockchain on your system.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+**: [Download](https://golang.org/dl/)
- **Git**: [Download](https://git-scm.com/downloads)
- **Make**: Usually pre-installed on Unix systems
- **Docker** (optional): [Download](https://www.docker.com/get-started)

### System Requirements

- **Minimum**: 4 CPU cores, 8GB RAM, 100GB SSD
- **Recommended**: 8 CPU cores, 32GB RAM, 500GB SSD
- **OS**: Linux (Ubuntu 20.04+), macOS 11+, or Windows WSL2

## Installation Methods

### Option 1: From Source (Recommended)

Building from source gives you the latest version and allows for custom configurations.

```bash
# Clone the repository
git clone https://github.com/aura-blockchain/aura.git
cd aura/chain

# Build the binary
make install

# Verify installation
aurad version
```

The `aurad` binary will be installed to `$GOPATH/bin`. Make sure this directory is in your `PATH`.

### Option 2: Using Docker

Docker provides an isolated environment and simplifies dependency management.

```bash
# Pull the latest image
docker pull aura-blockchain/aura:latest

# Run a container
docker run -it aura-blockchain/aura:latest aurad version

# For persistent data, mount a volume
docker run -it -v ~/.aura:/root/.aura aura-blockchain/aura:latest aurad init mynode
```

### Option 3: Pre-built Binaries

Download pre-compiled binaries for quick installation.

```bash
# Linux AMD64
curl -LO https://github.com/aura-blockchain/aura/releases/latest/download/aurad-linux-amd64
chmod +x aurad-linux-amd64
sudo mv aurad-linux-amd64 /usr/local/bin/aurad

# macOS ARM64
curl -LO https://github.com/aura-blockchain/aura/releases/latest/download/aurad-darwin-arm64
chmod +x aurad-darwin-arm64
sudo mv aurad-darwin-arm64 /usr/local/bin/aurad

# Verify installation
aurad version
```

## Additional Dependencies

### For Development

If you plan to build custom modules or contribute to Aura development:

```bash
# Install development tools
go install github.com/cosmwasm/wasmvm@latest
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install Ignite CLI (optional, for scaffolding)
curl https://get.ignite.com/cli! | bash
```

### For Validators

Additional tools for running a validator node:

```bash
# Install Cosmovisor (for automated upgrades)
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

# Install monitoring tools
# Prometheus, Grafana, etc. (see Monitoring guide)
```

## Environment Setup

### Configure Go Environment

```bash
# Add to ~/.bashrc or ~/.zshrc
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export PATH=$PATH:/usr/local/go/bin

# Reload shell configuration
source ~/.bashrc  # or source ~/.zshrc
```

### Configure Docker (if using Docker)

```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker

# Verify Docker works without sudo
docker ps
```

## Verification

Verify your installation is working correctly:

```bash
# Check version
aurad version

# Check available commands
aurad --help

# Initialize a test node
aurad init test-node --chain-id test-1

# Check configuration was created
ls -la ~/.aura/config/
```

## Troubleshooting

### Command not found: aurad

If `aurad` is not found after installation:

```bash
# Check if binary exists
ls $GOPATH/bin/aurad

# Add GOPATH/bin to PATH
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc
```

### Build errors

If you encounter build errors:

```bash
# Update Go to latest version
go version

# Clear Go cache
go clean -cache -modcache

# Try building again
cd aura/chain && make clean && make install
```

### Docker permission denied

If Docker commands require sudo:

```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

## Next Steps

Now that Aura is installed, proceed to:

- [Quick Start Guide](/docs/getting-started/quick-start) - Run your first node
- [Developer Guide](/docs/developers/overview) - Start building applications
- [Validator Setup](/docs/validators/setup) - Become a validator

## Updating

To update to the latest version:

```bash
# Pull latest code
cd aura
git pull origin main

# Rebuild
cd chain
make install

# Verify new version
aurad version
```

For Docker installations:

```bash
docker pull aura-blockchain/aura:latest
```
