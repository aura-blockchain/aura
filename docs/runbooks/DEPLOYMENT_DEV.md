# Development Environment Deployment

## Prerequisites
- Go 1.21+
- Docker & Docker Compose
- 8GB RAM minimum

## Quick Start

```bash
# Clone and build
git clone https://github.com/aura-blockchain/aura.git
cd aura/chain
go build -o aurad ./cmd/aurad

# Initialize single node
./aurad init dev-node --chain-id aura-dev-1
./aurad keys add validator --keyring-backend test

# Start node
./aurad start
```

## Docker Development

```bash
cd docker
docker-compose -f docker-compose.dev.yml up -d
```

## Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| AURA_HOME | ~/.aura | Node home directory |
| AURA_LOG_LEVEL | info | Log verbosity |
| AURA_RPC_PORT | 26657 | RPC port |

## Troubleshooting
- **Port conflicts**: Change ports in config.toml
- **Build errors**: Run `go mod tidy`
- **Genesis errors**: Remove ~/.aura and reinitialize
