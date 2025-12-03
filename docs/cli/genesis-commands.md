# Genesis CLI Commands

This document provides usage examples for the three critical genesis CLI helpers that have been restored and are fully wired to Aura's 27 modules.

## Commands Overview

1. **add-genesis-account** - Add genesis accounts with initial balances
2. **gentx** - Generate genesis transaction for validator self-delegation
3. **collect-gentxs** - Collect all genesis transactions into genesis.json

## Command: add-genesis-account

### Purpose
Adds an account with an initial token balance to the genesis file.

### Usage
```bash
aurad genesis add-genesis-account <address_or_key_name> <coins> [flags]
```

### Examples

**Add account with address:**
```bash
aurad genesis add-genesis-account aura1d2pj3t52hkuju4rqq2527l2dkqx8h67t8u92u4 100000000stake
```

**Add account with key name (from keyring):**
```bash
# First create a key
aurad keys add validator --keyring-backend test

# Then add to genesis using key name
aurad genesis add-genesis-account validator 100000000stake --keyring-backend test
```

**Add vesting account:**
```bash
aurad genesis add-genesis-account validator 100000000stake \
  --vesting-amount 50000000stake \
  --vesting-start-time 1609459200 \
  --vesting-end-time 1672531200
```

**Add module account:**
```bash
aurad genesis add-genesis-account my-module 1000000stake --module-name my-module
```

### Flags
- `--home` - Home directory (default: ~/.aura)
- `--keyring-backend` - Keyring backend (os|file|test)
- `--vesting-amount` - Amount of coins for vesting
- `--vesting-start-time` - Vesting start time (unix epoch)
- `--vesting-end-time` - Vesting end time (unix epoch)
- `--module-name` - Module account name
- `--append` - Append coins to existing account

## Command: gentx

### Purpose
Generates a genesis transaction that creates a validator with self-delegation. This transaction is signed by the validator's key.

### Usage
```bash
aurad genesis gentx <key_name> <amount> [flags]
```

### Examples

**Basic validator genesis transaction:**
```bash
aurad genesis gentx validator 90000000stake \
  --chain-id aura-testnet-1 \
  --keyring-backend test
```

**Full validator with metadata:**
```bash
aurad genesis gentx validator 90000000stake \
  --chain-id aura-testnet-1 \
  --keyring-backend test \
  --moniker "My Validator" \
  --commission-rate 0.10 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1000000 \
  --details "Production validator for Aura network" \
  --security-contact "security@example.com" \
  --website "https://example.com" \
  --identity "1234567890ABCDEF"
```

**Custom output location:**
```bash
aurad genesis gentx validator 90000000stake \
  --chain-id aura-testnet-1 \
  --output-document /path/to/custom-gentx.json
```

### Flags
- `--chain-id` - Network chain ID (required)
- `--keyring-backend` - Keyring backend (os|file|test)
- `--moniker` - Validator moniker
- `--commission-rate` - Initial commission rate (default: 0.1)
- `--commission-max-rate` - Maximum commission rate (default: 0.2)
- `--commission-max-change-rate` - Max commission change per day (default: 0.01)
- `--min-self-delegation` - Minimum self delegation (default: 1)
- `--details` - Validator details
- `--security-contact` - Security contact email
- `--website` - Validator website
- `--identity` - Keybase.io identity
- `--ip` - Public P2P IP
- `--p2p-port` - Public P2P port (default: 26656)
- `--output-document` - Custom output file path

## Command: collect-gentxs

### Purpose
Collects all genesis transactions from the `config/gentx/` directory and adds them to genesis.json.

### Usage
```bash
aurad genesis collect-gentxs [flags]
```

### Examples

**Collect from default location:**
```bash
aurad genesis collect-gentxs
```

**Collect from custom directory:**
```bash
aurad genesis collect-gentxs --gentx-dir /path/to/gentx/dir
```

**Collect with custom home:**
```bash
aurad genesis collect-gentxs --home /custom/home/dir
```

### Flags
- `--home` - Home directory (default: ~/.aura)
- `--gentx-dir` - Override default gentx directory

### Output
The command will:
1. Read all .json files from the gentx directory
2. Validate each genesis transaction
3. Add validators to the genesis file
4. Update the genesis file with collected transactions

## Complete Workflow Example

Here's a complete example of setting up a single-node testnet:

```bash
# 1. Initialize the node
aurad init mynode --chain-id aura-testnet-1

# 2. Create a validator key
aurad keys add validator --keyring-backend test

# 3. Add the validator as a genesis account with tokens
aurad genesis add-genesis-account validator 100000000stake --keyring-backend test

# 4. Generate the validator's genesis transaction
aurad genesis gentx validator 90000000stake \
  --chain-id aura-testnet-1 \
  --keyring-backend test \
  --moniker "My Validator"

# 5. Collect all genesis transactions
aurad genesis collect-gentxs

# 6. Validate the final genesis file
aurad genesis validate-genesis

# 7. Start the chain
aurad start
```

## Multi-Node Testnet Workflow

For a multi-node testnet, each validator needs to:

```bash
# On each validator node:

# 1. Initialize
aurad init validator1 --chain-id aura-testnet-1

# 2. Create validator key
aurad keys add validator1 --keyring-backend test

# 3. Share the genesis account address with coordinator
aurad keys show validator1 -a --keyring-backend test

# Coordinator collects all validator addresses and creates genesis file:

# 4. Add all validators as genesis accounts
aurad genesis add-genesis-account aura1abc... 100000000stake
aurad genesis add-genesis-account aura1def... 100000000stake
aurad genesis add-genesis-account aura1ghi... 100000000stake

# 5. Distribute genesis.json to all validators

# On each validator:

# 6. Generate gentx on each node
aurad genesis gentx validator1 90000000stake \
  --chain-id aura-testnet-1 \
  --keyring-backend test

# 7. Collect all gentx files to coordinator node's config/gentx/ directory

# Coordinator collects gentxs:

# 8. Collect all genesis transactions
aurad genesis collect-gentxs

# 9. Validate genesis
aurad genesis validate-genesis

# 10. Distribute final genesis.json to all validators

# 11. Start all nodes
aurad start
```

## Integration with Aura Modules

These genesis commands are fully integrated with all 27 Aura modules:

- **Authentication & Identity**: identity, aiassistant, vcregistry
- **Security**: security, walletsecurity, validatorsecurity, networksecurity, economicsecurity
- **Finance**: bank, dex, bridge
- **Governance**: governance, compliance
- **Infrastructure**: wasm, contractregistry, dataregistry, incidentresponse, monitoring, aurabindings
- **Core SDK**: auth, staking, distribution, slashing, gov, params, upgrade, consensus, genutil

All module genesis states are properly validated and initialized through these commands.

## Troubleshooting

### Error: "key not found"
Make sure you're using the correct keyring backend and the key exists:
```bash
aurad keys list --keyring-backend test
```

### Error: "invalid genesis"
Validate your genesis file:
```bash
aurad genesis validate-genesis
```

### Error: "account already exists"
Use the `--append` flag to add more coins to an existing account:
```bash
aurad genesis add-genesis-account validator 50000000stake --append
```

## Testing

All genesis commands have comprehensive test coverage in `chain/cmd/aurad/cmd/genesis_test.go`:

- Command structure tests
- Flag validation tests
- Help text verification
- Integration workflow tests

Run tests:
```bash
go test ./cmd/aurad/cmd -run TestGenesis -v
```
