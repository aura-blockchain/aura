# Genesis Commands Quick Reference

## Overview
Quick reference guide for Aura blockchain genesis commands.

## Common Use Cases

### 1. Start a New Blockchain

```bash
# Initialize node
aurad init mynode --chain-id aura-1 --home ~/.aura

# Create validator key
aurad keys add validator --keyring-backend test

# Add genesis account (1 billion tokens)
aurad genesis add-genesis-account validator 1000000000stake --keyring-backend test

# Create genesis validator transaction
aurad genesis gentx validator 100000000stake \
    --chain-id aura-1 \
    --moniker "My Validator" \
    --keyring-backend test

# Collect genesis transactions
aurad genesis collect-gentxs

# Validate genesis file
aurad genesis validate-genesis

# Start the chain
aurad start
```

### 2. Quick Development Setup

```bash
aurad quickstart --validator-name dev
# Follow the printed instructions
```

### 3. Add Multiple Genesis Accounts

```bash
# Create keys
aurad keys add alice --keyring-backend test
aurad keys add bob --keyring-backend test
aurad keys add charlie --keyring-backend test

# Add genesis accounts
aurad genesis add-genesis-account alice 1000000000stake --keyring-backend test
aurad genesis add-genesis-account bob 500000000stake --keyring-backend test
aurad genesis add-genesis-account charlie 250000000stake --keyring-backend test
```

### 4. Multi-Validator Setup

```bash
# On Node 1
aurad genesis gentx validator1 100000000stake \
    --chain-id aura-1 \
    --moniker "Validator 1" \
    --keyring-backend test

# On Node 2
aurad genesis gentx validator2 100000000stake \
    --chain-id aura-1 \
    --moniker "Validator 2" \
    --keyring-backend test

# Copy all gentx files to node1's config/gentx directory
# Then on node1:
aurad genesis collect-gentxs

# Distribute the final genesis.json to all nodes
```

### 5. Vesting Account

```bash
# Add account with 50% vesting over 1 year
# End time: January 1, 2025 00:00:00 UTC = 1735689600
aurad genesis add-genesis-account alice 1000000000stake \
    --vesting-amount 500000000stake \
    --vesting-end-time 1735689600 \
    --keyring-backend test
```

### 6. Validate Existing Genesis

```bash
aurad genesis validate-genesis
aurad genesis validate-genesis --home /custom/path
```

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| `add-genesis-account` | Add account to genesis | `aurad genesis add-genesis-account alice 1000stake` |
| `gentx` | Generate validator tx | `aurad genesis gentx validator 100stake --chain-id aura-1` |
| `collect-gentxs` | Collect all gentxs | `aurad genesis collect-gentxs` |
| `validate-genesis` | Validate genesis file | `aurad genesis validate-genesis` |
| `migrate` | Migrate genesis version | `aurad genesis migrate v0.46` |
| `export` | Export chain state | `aurad genesis export > genesis.json` |

## Common Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--home` | Node home directory | `~/.aura` |
| `--chain-id` | Chain identifier | `aura-1` |
| `--keyring-backend` | Keyring backend | `os` |
| `--moniker` | Node name | `aura-node` |

## Validator Commission Defaults

| Parameter | Default | Description |
|-----------|---------|-------------|
| `--commission-rate` | 0.1 (10%) | Initial commission rate |
| `--commission-max-rate` | 0.2 (20%) | Maximum commission rate |
| `--commission-max-change-rate` | 0.01 (1%) | Max daily commission change |
| `--min-self-delegation` | 1 | Minimum self delegation |

## File Locations

```
~/.aura/
├── config/
│   ├── genesis.json          # Genesis file
│   ├── config.toml            # Node configuration
│   ├── app.toml               # App configuration
│   └── gentx/                 # Genesis transactions
│       ├── gentx-validator1.json
│       └── gentx-validator2.json
├── data/                      # Blockchain data
└── keys/                      # Keyring data
```

## Troubleshooting

### Error: "account already exists"
```bash
# Remove the existing account from genesis.json manually
# or start with a fresh genesis file
aurad init mynode --chain-id aura-1 --overwrite
```

### Error: "failed to get key from keyring"
```bash
# Verify key exists
aurad keys list --keyring-backend test

# Create key if missing
aurad keys add mykey --keyring-backend test
```

### Error: "invalid coins"
```bash
# Use correct format: amount + denom
# Correct: 1000000stake
# Wrong: 1000000 stake
aurad genesis add-genesis-account alice 1000000stake
```

### Error: "genesis time cannot be zero"
```bash
# Re-initialize with proper genesis time
aurad init mynode --chain-id aura-1
```

## Best Practices

1. **Always validate genesis**: Run `validate-genesis` before starting the chain
2. **Use test keyring for development**: `--keyring-backend test`
3. **Use os keyring for production**: `--keyring-backend os`
4. **Backup keys**: Store validator keys securely
5. **Document validators**: Use descriptive monikers and details
6. **Test locally first**: Set up a single-node testnet before multi-node
7. **Set proper commission**: Configure commission rates carefully

## Security Tips

1. **Protect private keys**: Never share validator private keys
2. **Use hardware wallets**: For production validators
3. **Secure home directory**: Set proper permissions (0700)
4. **Use TLS**: Enable TLS for production nodes
5. **Backup regularly**: Back up genesis and keys
6. **Validate inputs**: Always check addresses and amounts

## Development Tips

```bash
# Reset everything
rm -rf ~/.aura

# Quick single-node testnet
aurad quickstart

# Or manual setup for more control
aurad init dev --chain-id test-1
aurad keys add dev --keyring-backend test
aurad genesis add-genesis-account dev 100000000stake --keyring-backend test
aurad genesis gentx dev 10000000stake --chain-id test-1 --keyring-backend test
aurad genesis collect-gentxs
aurad start
```

## Production Checklist

Before launching a production chain:

- [ ] Genesis file validated
- [ ] All validator gentxs collected
- [ ] Chain ID is unique and permanent
- [ ] Initial supply is correct
- [ ] Validator commission rates are reasonable
- [ ] All validator keys are backed up securely
- [ ] Genesis time is set correctly
- [ ] Module parameters are reviewed
- [ ] All validators have confirmed setup
- [ ] Network connectivity tested
- [ ] Monitoring and alerts configured

## Getting Help

```bash
# Command-specific help
aurad genesis --help
aurad genesis add-genesis-account --help
aurad genesis gentx --help

# General help
aurad help
aurad help genesis

# Version information
aurad version
```

## Additional Resources

- Full documentation: See `GENESIS_CLI_IMPLEMENTATION.md`
- Cosmos SDK docs: https://docs.cosmos.network
- Aura blockchain docs: https://docs.aura.network
