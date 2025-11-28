# Aura Blockchain Project Guidelines

**Read the parent guidelines first:** `../CLAUDE.md` and `../AGENTS.md` contain general agent instructions that apply to all projects.

This file contains Aura-specific conventions and instructions.

---

## Project Overview

Aura is a Cosmos SDK-based blockchain with custom modules for identity, privacy, compliance, DEX, bridge, and more.

## Project Structure

```
aura/
├── chain/              # Go source code (Cosmos SDK)
│   ├── cmd/aurad/      # CLI daemon entry point
│   ├── app/            # Application wiring
│   └── x/              # Custom modules
├── proto/              # Protobuf definitions
├── contracts/          # Smart contracts (if any)
├── docs/               # Documentation
└── scripts/            # Build and utility scripts
```

## Node Data Directory

**The node data directory is `~/.aura/` (in user's home directory), NOT in the repo.**

This is the Cosmos SDK convention. The `~/.aura/` directory contains:
- `config/` - Node configuration (app.toml, config.toml, genesis.json)
- `data/` - Blockchain state database
- `keyring-*/` - Wallet keys
- Private validator and node keys

**Do NOT:**
- Put blockchain data in the repo
- Commit private keys or node identity files
- Change the default data directory without documenting it

**To initialize a fresh node:**
```bash
cd chain
go build -o aurad ./cmd/aurad
./aurad init <moniker> --chain-id aura-testnet-1
```

**To specify a custom data directory (if needed):**
```bash
./aurad init <moniker> --home /custom/path
# or
export AURA_HOME=/custom/path
```

## Building

```bash
cd chain
go build -o aurad ./cmd/aurad

# Optimized build (smaller binary):
go build -ldflags="-s -w" -o aurad ./cmd/aurad
```

**Note:** The `aurad` binary is excluded from git (too large). Always rebuild from source.

## Testing

```bash
cd chain

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific module tests
go test ./x/identity/...
```

**Pre-commit hooks are configured.** Run `pre-commit install` to enable them.

## Protobuf Generation

When modifying `.proto` files:
```bash
cd chain
make proto-gen
```

## Module Development

When adding or modifying modules in `chain/x/`:
1. Update protobuf definitions in `proto/`
2. Regenerate protobuf files
3. Implement keeper methods
4. Add genesis import/export
5. Register in `app/app.go`
6. Write comprehensive tests
7. Update documentation

## Git Workflow

- Commit frequently after completing each task
- Push to GitHub after each commit
- Use clear commit messages
- GitHub Actions are DISABLED (local testing only via pre-commit hooks)

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AURA_HOME` | Node data directory | `~/.aura` |
| `AURA_CHAIN_ID` | Chain identifier | - |

## Common Issues

**"aurad: command not found"**
- Build the binary first: `cd chain && go build -o aurad ./cmd/aurad`

**"genesis.json not found"**
- Initialize the node: `./aurad init <moniker>`

**Large file rejected by GitHub**
- The `aurad` binary should NOT be committed. Check `.gitignore`.
