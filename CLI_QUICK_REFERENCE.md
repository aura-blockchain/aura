# Aura CLI Quick Reference

Quick reference for the most commonly used Aura blockchain commands.

---

## Setup

```bash
# Build the binary
cd /home/hudson/blockchain-projects/aura/chain
go build -o aurad ./cmd/aurad

# Initialize node
./aurad init <moniker> --chain-id aura-testnet-1

# Check version
./aurad version
```

---

## DEX Commands

### AMM Pool Operations

```bash
# Create a new liquidity pool
aurad tx dex create-pool uaura 1000000 usdt 500000 --from alice

# Add liquidity to existing pool
aurad tx dex add-liquidity uaura-usdt 500000uaura 250000usdt --from alice

# Remove liquidity from pool
aurad tx dex remove-liquidity uaura-usdt 100000 --from alice

# Swap tokens
aurad tx dex swap uaura-usdt 100000uaura 48000 500 --from alice
#                 [pool-id] [coin-in] [min-out] [slippage-bps]
```

### P2P Orderbook

```bash
# Create sell order
aurad tx dex create-order sell 1000000 usdt 500000 --from alice

# Create buy order
aurad tx dex create-order buy 2000000 btc 5000 --from bob

# Match existing order
aurad tx dex match-order <order-id> --from bob

# Cancel pending order
aurad tx dex cancel-order <order-id> --from alice
```

### Atomic Swaps (HTLC)

```bash
# Create HTLC
aurad tx dex create-htlc aura1recipient... 1000000uaura abc123def456... 3600 --from alice
#                         [recipient] [amount] [secret-hash] [timelock-seconds]

# Claim HTLC (reveal secret)
aurad tx dex claim-htlc <htlc-id> <secret-preimage> --from bob

# Refund expired HTLC
aurad tx dex refund-htlc <htlc-id> --from alice
```

### DEX Queries

```bash
# List all pools
aurad query dex pools --node tcp://localhost:27657

# Query specific pool
aurad query dex pool uaura-usdt --node tcp://localhost:27657

# Get swap quote (no execution)
aurad query dex quote uaura-usdt uaura 1000000 --node tcp://localhost:27657

# Check spot price
aurad query dex spot-price uaura-usdt uaura --node tcp://localhost:27657

# View orderbook
aurad query dex orderbook AURA/USDT --node tcp://localhost:27657

# Query user's orders
aurad query dex user-orders aura1abc... --node tcp://localhost:27657

# Query HTLC status
aurad query dex htlc <htlc-id> --node tcp://localhost:27657

# Get market price
aurad query dex market-price uaura --node tcp://localhost:27657

# List supported coins
aurad query dex supported-coins --node tcp://localhost:27657
```

---

## Compliance Commands

### KYC & Sanctions Screening

```bash
# Submit KYC verification
aurad tx compliance submit-kyc aura1abc... 3 cosmos1provider... a1b2c3d4...f0 US --from provider
#                              [address] [level] [provider] [pii-commitment] [jurisdiction]
# Levels: 1=NONE, 2=BASIC, 3=INTERMEDIATE, 4=ADVANCED

# Screen address against sanctions
aurad tx compliance screen-sanctions aura1abc... --from alice

# Force refresh sanctions check
aurad tx compliance screen-sanctions aura1abc... --force-refresh --from alice

# Report suspicious activity
aurad tx compliance report-suspicious aura1abc... "reason" --from alice

# Record GDPR consent
aurad tx compliance record-consent --from alice

# Request GDPR data
aurad tx compliance request-data --from alice

# Generate tax report
aurad tx compliance generate-tax-report 2024 --from alice
```

### Compliance Queries

```bash
# Query KYC record
aurad query compliance kyc-record aura1abc... --node tcp://localhost:27657

# Query AML risk profile
aurad query compliance aml-profile aura1abc... --node tcp://localhost:27657

# Query sanctions screening results
aurad query compliance sanctions aura1abc... --node tcp://localhost:27657

# Query transaction monitoring alerts
aurad query compliance alerts aura1abc... --node tcp://localhost:27657

# Query tax report
aurad query compliance tax-report aura1abc... 2024 --node tcp://localhost:27657
```

---

## Confidence Score Commands

### Score Management

```bash
# Record IR completion (assistant only)
aurad tx confidencescore record-completion aura1user... IR-042-FINANCE 500 --from assistant

# Appeal a slash
aurad tx confidencescore appeal <slash-id> "appeal reason" --from alice

# Recalculate score (governance only)
aurad tx confidencescore recalculate-score aura1abc... --from governance

# Slash user (governance only)
aurad tx confidencescore slash aura1abc... 1000 "fraud detected" --from governance

# Resolve appeal (governance only)
aurad tx confidencescore resolve-appeal <appeal-id> approved --from governance
```

### Score Queries

```bash
# Query user's confidence score
aurad query confidencescore score aura1abc... --node tcp://localhost:27657

# Query IR completion history
aurad query confidencescore completions aura1abc... --node tcp://localhost:27657

# Query specific IR completion
aurad query confidencescore ir-completion <completion-id> --node tcp://localhost:27657

# Query score change history
aurad query confidencescore history aura1abc... --node tcp://localhost:27657

# Query arena breakdown
aurad query confidencescore arena-breakdown aura1abc... --node tcp://localhost:27657

# Query slash records
aurad query confidencescore slash-records aura1abc... --node tcp://localhost:27657

# Query verification thresholds
aurad query confidencescore thresholds --node tcp://localhost:27657

# List verified users
aurad query confidencescore verified-users --node tcp://localhost:27657
```

---

## Wasm Security Commands

### Contract Management

```bash
# Upload wasm binary
aurad tx aura_wasm_security store contract.wasm --from alice

# Instantiate contract
aurad tx aura_wasm_security instantiate <code-id> '{"init":"msg"}' --label "My Contract" --from alice

# Execute contract
aurad tx aura_wasm_security execute <contract-address> '{"method":"value"}' --from alice

# Migrate contract
aurad tx aura_wasm_security migrate <contract-address> <new-code-id> '{"migrate":"msg"}' --from admin

# Set contract admin
aurad tx aura_wasm_security set-admin <contract-address> <new-admin-address> --from admin

# Clear contract admin (prevent migrations)
aurad tx aura_wasm_security clear-admin <contract-address> --from admin

# Pause contract (governance only)
aurad tx aura_wasm_security pause-contract <contract-address> --from governance

# Unpause contract (governance only)
aurad tx aura_wasm_security unpause-contract <contract-address> --from governance

# Authorize uploader (governance only)
aurad tx aura_wasm_security authorize-uploader <address> --from governance

# Revoke uploader (governance only)
aurad tx aura_wasm_security revoke-uploader <address> --from governance
```

---

## Standard Cosmos SDK Commands

### Bank

```bash
# Send tokens
aurad tx bank send alice aura1recipient... 1000000uaura --from alice
```

### Staking

```bash
# Delegate to validator
aurad tx staking delegate auravaloper1abc... 1000000uaura --from alice

# Redelegate to different validator
aurad tx staking redelegate auravaloper1abc... auravaloper1def... 500000uaura --from alice

# Unbond from validator
aurad tx staking unbond auravaloper1abc... 500000uaura --from alice

# Cancel unbonding
aurad tx staking cancel-unbond auravaloper1abc... 500000uaura 12345 --from alice
```

### Distribution

```bash
# Withdraw rewards from single validator
aurad tx distribution withdraw-rewards auravaloper1abc... --from alice

# Withdraw all rewards
aurad tx distribution withdraw-all-rewards --from alice

# Change withdrawal address
aurad tx distribution set-withdraw-addr aura1newaddress... --from alice

# Fund community pool
aurad tx distribution fund-community-pool 1000000uaura --from alice
```

---

## Query Commands

### Account & Transaction

```bash
# Query account info
aurad query account aura1abc... --node tcp://localhost:27657

# Query transaction by hash
aurad query tx <hash> --node tcp://localhost:27657

# Query transaction by account sequence
aurad query tx --type=acc_seq aura1abc.../5 --node tcp://localhost:27657

# Query validator set
aurad query comet-validator-set --node tcp://localhost:27657
```

---

## Common Flags

### Transaction Flags

```bash
--from <key-name>           # Signer key name or address
--chain-id <chain-id>       # Network chain ID
--fees 1000uaura            # Transaction fees
--gas auto                  # Auto-calculate gas
--gas-prices 0.025uaura     # Gas price (alternative to --fees)
--node tcp://host:port      # Node RPC endpoint
--broadcast-mode sync       # Broadcasting mode (sync|async)
--yes                       # Skip confirmation prompt
--dry-run                   # Simulate transaction
--generate-only             # Generate unsigned transaction
--offline                   # Offline mode
```

### Query Flags

```bash
--node tcp://host:port      # Node RPC endpoint
--output json               # Output format (json|text)
--height <block-height>     # Query at specific height
--chain-id <chain-id>       # Network chain ID
```

---

## Testnet Connection

### Validator Endpoints

```bash
# Validator 1
--node tcp://localhost:27657

# Validator 2
--node tcp://localhost:27757

# Validator 3
--node tcp://localhost:27857

# Validator 4
--node tcp://localhost:27957
```

### Check Node Status

```bash
# Via curl (recommended - status command has a bug)
curl -s http://localhost:27657/status | jq '.result.sync_info'

# Get latest block height
curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height'

# Check catching up status
curl -s http://localhost:27657/status | jq -r '.result.sync_info.catching_up'
```

---

## Tips & Best Practices

### Gas & Fees

```bash
# Auto-calculate gas with adjustment
--gas auto --gas-adjustment 1.3

# Set gas price (will calculate fees automatically)
--gas-prices 0.025uaura

# Set explicit fees
--fees 5000uaura
```

### Testnet Usage

```bash
# Always specify node for testnet
--node tcp://localhost:27657

# Use test keyring backend
--keyring-backend test

# Skip confirmation for automated scripts
--yes
```

### Debugging

```bash
# Dry run to test transaction
--dry-run

# Generate transaction without broadcasting
--generate-only

# Enable debug output
--debug

# Enable verbose output
--verbose
```

### Output Formats

```bash
# JSON output (for scripting)
--output json

# Text output (human readable)
--output text

# YAML output
--output yaml

# CSV output (for some queries)
--output csv
```

---

## Module Aliases

Some modules have convenient aliases:

```bash
# Compliance module
aurad tx compliance ...
aurad tx kyc ...
aurad tx comp ...

# Confidence Score module
aurad tx confidencescore ...
aurad tx cs ...
aurad tx score ...
```

---

## Examples for Common Workflows

### 1. Creating and Using a DEX Pool

```bash
# Step 1: Create pool
aurad tx dex create-pool uaura 10000000 usdt 5000000 --from alice --yes

# Step 2: Query pool
aurad query dex pool uaura-usdt --node tcp://localhost:27657

# Step 3: Get swap quote
aurad query dex quote uaura-usdt uaura 100000 --node tcp://localhost:27657

# Step 4: Execute swap
aurad tx dex swap uaura-usdt 100000uaura 48000 500 --from bob --yes
```

### 2. Cross-Chain Atomic Swap

```bash
# Alice on Chain A:
SECRET=$(openssl rand -hex 32)
SECRET_HASH=$(echo -n $SECRET | sha256sum | cut -d' ' -f1)
aurad tx dex create-htlc aura1bob... 1000000uaura $SECRET_HASH 3600 --from alice

# Bob on Chain B (after seeing Alice's HTLC):
aurad tx dex create-htlc <bob-chain-recipient> 500000usdt $SECRET_HASH 3600 --from bob

# Alice claims Bob's HTLC (reveals secret):
aurad tx dex claim-htlc <htlc-id> $SECRET --from alice

# Bob claims Alice's HTLC (using revealed secret):
aurad tx dex claim-htlc <htlc-id> $SECRET --from bob
```

### 3. KYC Verification Process

```bash
# Step 1: Generate PII commitment off-chain
PII_HASH=$(echo -n "sensitive-data" | sha256sum | cut -d' ' -f1)

# Step 2: Submit KYC
aurad tx compliance submit-kyc aura1user... 3 aura1provider... $PII_HASH US --from provider --yes

# Step 3: Query KYC record
aurad query compliance kyc-record aura1user... --node tcp://localhost:27657

# Step 4: Screen for sanctions
aurad tx compliance screen-sanctions aura1user... --from anyone --yes

# Step 5: Query sanctions results
aurad query compliance sanctions aura1user... --node tcp://localhost:27657
```

---

## Troubleshooting

### Common Errors

```bash
# Error: unknown command "bank" for "query"
# Fix: Bank query module not registered (known issue)
# Workaround: Use account query instead

# Error: method Params not implemented
# Issue: Confidence score params query not implemented
# Status: Known issue, low priority

# Error: failed to connect to node
# Fix: Check node is running and use correct RPC port
curl -s http://localhost:27657/status

# Error: key not found
# Fix: Import or create key first
aurad keys add alice --keyring-backend test
```

### Getting Help

```bash
# Command help
aurad <command> --help

# Module help
aurad tx dex --help
aurad query compliance --help

# Global help
aurad --help
```
