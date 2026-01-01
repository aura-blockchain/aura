# AURA Project Guidelines

**Read `../CLAUDE.md` first** - contains all general instructions.

## Project-Specific
- **Framework**: Cosmos SDK (Go)
- **Binary**: `go build -o aurad ./cmd/...`
- **Init**: `./aurad init <moniker> --chain-id aura-testnet-1`

## Testnet Access
- **Server**: `ssh aura-testnet` (158.69.119.76)
- **Chain ID**: aura-testnet-1
- **Binary**: `~/.aura/cosmovisor/genesis/bin/aurad`
- **Home**: `~/.aura`
- **VPN**: 10.10.0.1

### Quick Commands
```bash
aurad status --home ~/.aura
aurad query staking validators --home ~/.aura
```

**Full docs**: See `TESTNET_INFRASTRUCTURE.md`
