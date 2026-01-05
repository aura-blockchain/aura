# AURA Project

## Repository Separation

**This repo (`aura/`)** → github:aura-blockchain/aura (source code)
**Testnet repo (`aura-testnets/`)** → github:aura-blockchain/testnets (network config)

### Save HERE (aura/)
- Go source code, modules, CLI
- Protobuf definitions
- Tests, Makefiles, Dockerfiles
- General docs (README, CONTRIBUTING)

### Save to TESTNET REPO (aura-testnets/aura-testnet-1/)
- genesis.json, chain.json, assetlist.json, versions.json
- peers.txt, seeds.txt
- config/app.toml, config/config.toml
- SNAPSHOTS.md, state_sync.md, README.md
- bin/SHA256SUMS

## Testnet SSH Access
```bash
ssh aura-testnet  # 158.69.119.76
```

## Health Check
Run `./scripts/health-check-all.sh` for AURA-specific health check.
