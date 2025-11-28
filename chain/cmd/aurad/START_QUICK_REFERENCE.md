# Start Command Quick Reference

## File Location
`/home/decri/blockchain-projects/aura/chain/cmd/aurad/cmd/start.go`

## Quick Start

```bash
# 1. Initialize the node
aurad init mynode --chain-id aura-1

# 2. Start the node
aurad start
```

## Key Functions

### Main Entry Point
- `StartCmd(auraApp **app.App, logger log.Logger) *cobra.Command` - Creates the start command

### Node Startup
- `startInProcess()` - Starts node with CometBFT consensus (production mode)
- `startStandAlone()` - Starts without consensus (development mode)

### CometBFT Integration
- `loadCometConfig()` - Loads CometBFT configuration
- `loadOrGenPrivValidator()` - Loads/generates validator keys
- `createCometLogger()` - Creates CometBFT logger
- `NewCometABCIWrapper()` - Wraps app for ABCI interface

### Server Management
- `startGRPCServer()` - Starts gRPC query server
- `startAPIServer()` - Starts REST API server
- `waitForShutdown()` - Handles graceful shutdown

### Database & Storage
- `openDB()` - Opens GoLevelDB database
- `loadBaseAppOptions()` - Configures BaseApp options

## Common Commands

### Basic Start
```bash
aurad start
```

### Custom gRPC/API Addresses
```bash
aurad start \
  --grpc.address=0.0.0.0:9090 \
  --api.address=tcp://0.0.0.0:1317
```

### With Gas Prices
```bash
aurad start --minimum-gas-prices="0.025uaura"
```

### Custom Pruning
```bash
aurad start \
  --pruning=custom \
  --pruning-keep-recent=100000 \
  --pruning-interval=10
```

### Development Mode (No Consensus)
```bash
aurad start --with-comet=false
```

### With Tracing
```bash
aurad start --trace-store=/tmp/trace.log --trace=true
```

## Important Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--with-comet` | bool | true | Enable CometBFT consensus |
| `--grpc.enable` | bool | true | Enable gRPC server |
| `--grpc.address` | string | localhost:9090 | gRPC listen address |
| `--api.enable` | bool | true | Enable API server |
| `--api.address` | string | localhost:1317 | API listen address |
| `--pruning` | string | default | Pruning strategy |
| `--minimum-gas-prices` | string | "" | Min gas prices |
| `--halt-height` | uint64 | 0 | Halt at block height |
| `--home` | string | ~/.aura | Node home directory |

## Default Ports

- **CometBFT P2P**: 26656
- **CometBFT RPC**: 26657
- **ABCI**: 26658
- **gRPC**: 9090
- **REST API**: 1317

## Health Checks

```bash
# API health
curl http://localhost:1317/health

# CometBFT status
curl http://localhost:26657/status

# gRPC services
grpcurl -plaintext localhost:9090 list
```

## File Structure Created

After initialization, the home directory contains:

```
~/.aura/
├── config/
│   ├── config.toml          # CometBFT configuration
│   ├── app.toml             # Application configuration
│   ├── genesis.json         # Genesis file
│   ├── node_key.json        # Node P2P key
│   ├── priv_validator_key.json  # Validator consensus key
│   └── priv_validator_state.json # Validator state
├── data/
│   └── application.db/      # Blockchain state database
└── logs/                    # Log files
```

## Shutdown

The node handles graceful shutdown on:
- `Ctrl+C` (SIGINT)
- `SIGTERM`
- `SIGQUIT`

Shutdown order:
1. API server stops accepting requests
2. gRPC server completes in-flight requests
3. CometBFT stops consensus
4. Database closes

Timeout: 30 seconds

## Troubleshooting

### Port Already in Use
```bash
# Change ports
aurad start \
  --grpc.address=localhost:9091 \
  --api.address=tcp://localhost:1318
```

### Database Locked
```bash
# Ensure no other instance is running
ps aux | grep aurad
kill <pid>

# Or remove lock file
rm ~/.aura/data/application.db/LOCK
```

### Permission Denied
```bash
# Check home directory permissions
ls -la ~/.aura
chmod 755 ~/.aura
chmod 700 ~/.aura/config
```

### Cannot Connect to Peers
```bash
# Check P2P port is accessible
netstat -an | grep 26656

# Check firewall
sudo ufw allow 26656/tcp
```

## Production Recommendations

1. **Enable TLS**: Configure TLS certificates for gRPC and API
2. **Set Gas Prices**: Use `--minimum-gas-prices` to prevent spam
3. **Configure Pruning**: Use custom pruning to manage disk space
4. **Monitor Health**: Set up health check monitoring
5. **Use Systemd**: Run as systemd service for auto-restart
6. **Backup Data**: Regular backups of `~/.aura/data` and config files

## Systemd Service Example

```ini
[Unit]
Description=Aura Blockchain Node
After=network.target

[Service]
Type=simple
User=aura
WorkingDirectory=/home/aura
ExecStart=/usr/local/bin/aurad start
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

## Monitoring

```bash
# Check if node is syncing
curl -s http://localhost:26657/status | jq .result.sync_info.catching_up

# Check latest block
curl -s http://localhost:26657/status | jq .result.sync_info.latest_block_height

# Check peers
curl -s http://localhost:26657/net_info | jq .result.n_peers
```

## Development Tips

1. **Fast Restart**: Use `--with-comet=false` for quick testing
2. **Trace Logs**: Enable `--trace=true` for debugging
3. **Custom Home**: Use `--home` to run multiple nodes
4. **Low Gas Prices**: Set `--minimum-gas-prices="0uaura"` for testing

## Security Notes

- **Never expose** 26656, 26657, 26658 to public internet on mainnet
- **Always use TLS** for gRPC and API in production
- **Restrict API access** with firewall rules
- **Rotate validator keys** regularly
- **Keep backups** of private keys in secure location

## Code Statistics

- **Total Lines**: ~770
- **Functions**: 20+
- **ABCI Methods**: 13 (full ABCI 2.0 compliance)
- **Server Types**: 2 (gRPC, HTTP)
- **Configuration Flags**: 15+

## Implementation Status

✅ CometBFT node initialization
✅ ABCI application wrapper
✅ Block production and consensus
✅ gRPC server with TLS support
✅ REST API server with security
✅ Graceful shutdown handling
✅ Database management
✅ Configuration loading
✅ Validator key management
✅ Comprehensive error handling
✅ Production-ready logging
