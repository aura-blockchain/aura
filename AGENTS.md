# AURA Project

## Repository Separation

**This repo (`aura/`)** → github:aura-blockchain/aura (source code)
**Testnet repo (`aura-testnets/`)** → github:aura-blockchain/testnets (network config)

### Save HERE (aura/)
- Go source code, modules, CLI
- Protobuf definitions
- Tests, Makefiles, Dockerfiles
- General docs (README, CONTRIBUTING)

### Save to TESTNET REPO (aura-testnets/aura-mvp-1/)
- genesis.json, chain.json, assetlist.json, versions.json
- peers.txt, seeds.txt
- config/app.toml, config/config.toml
- SNAPSHOTS.md, state_sync.md, README.md
- bin/SHA256SUMS

## Testnet Public Endpoints (Cloudflare A Records)

**IMPORTANT**: Always use these registered URLs. Never use raw IPs or localhost.

| Service | URL |
|---------|-----|
| RPC | https://testnet-rpc.aurablockchain.org |
| REST API | https://testnet-api.aurablockchain.org |
| gRPC | testnet-grpc.aurablockchain.org:443 |
| WebSocket | wss://testnet-ws.aurablockchain.org |
| Explorer | https://testnet-explorer.aurablockchain.org |
| Faucet | https://testnet-faucet.aurablockchain.org |
| Monitoring | https://monitoring.aurablockchain.org |
| Snapshots | https://snapshots.aurablockchain.org |
| Artifacts | https://artifacts.aurablockchain.org |

## Testnet SSH Access
```bash
ssh aura-testnet  # 158.69.119.76
```

## Health Check
Run `./scripts/health-check-all.sh` for AURA-specific health check.

## ⚠️ ACTUAL Port Configuration (AURA-Specific)

**CRITICAL**: AURA uses NON-STANDARD ports\!

| Service | Port | Bind Address | Notes |
|---------|------|--------------|-------|
| RPC | **10657** | 127.0.0.1 | Custom (not 26657\!) |
| gRPC | **9190** | 0.0.0.0 | Custom (not 9090\!) |
| REST API | 1317 | 127.0.0.1 | Standard |
| P2P | 26656 | 0.0.0.0 | Standard |
| Prometheus | 26660 | 0.0.0.0 | Standard |

### Internal Access (when SSH'd to aura-testnet)
- RPC: http://127.0.0.1:10657
- gRPC: 127.0.0.1:9190
- REST: http://127.0.0.1:1317

### Source of Truth
See `~/blockchain-sites/testnet-registry/aura-testnet/chain.json` (on WSL2/decri)

## Services Server (services-testnet / 139.99.149.160 / 10.10.0.4)

Secondary nodes for redundancy, indexers, and shared infrastructure.

```bash
ssh services-testnet  # 139.99.149.160
```

### Secondary Node Ports (on services-testnet)
| Chain | RPC | gRPC | REST | P2P |
|-------|-----|------|------|-----|
| AURA | 26657 | 9190 | 1317 | 26656 |
| PAW | 27657 | 9091 | 1327 | 27656 |
| XAI | 8546 | - | - | 8766 |

### Indexers & WebSocket Proxies
| Service | Port |
|---------|------|
| AURA Indexer API | 4101 |
| PAW Indexer API | 4102 |
| AURA WS Proxy | 4201 |
| PAW WS Proxy | 4202 |
| XAI WS Proxy | 4203 |
