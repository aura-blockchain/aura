# Aura Testnet CLI Examples

Real commands you can run against the live Aura testnet.

**Testnet RPC:** tcp://localhost:27657
**Chain ID:** aura-testnet-1
**Latest Block:** 13466+ (actively producing)

---

## Prerequisites

```bash
# Build aurad
cd /home/hudson/blockchain-projects/aura/chain
go build -o /tmp/aurad ./cmd/aurad

# Set up environment (for convenience)
export AURAD=/tmp/aurad
export NODE=tcp://localhost:27657
export CHAIN_ID=aura-testnet-1
```

---

## Query Commands (No Keys Required)

These commands work without any keys or accounts:

### 1. Check Blockchain Status

```bash
# Get latest block info via curl (recommended - status command has a bug)
curl -s http://localhost:27657/status | jq '.result.sync_info'

# Get current block height
curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height'

# Check if node is catching up
curl -s http://localhost:27657/status | jq -r '.result.sync_info.catching_up'
```

**Expected Output:**
```json
{
  "latest_block_hash": "...",
  "latest_app_hash": "...",
  "latest_block_height": "13466",
  "latest_block_time": "2025-12-14T02:43:56.486545956Z",
  "catching_up": false
}
```

### 2. Query Validator Set

```bash
$AURAD query comet-validator-set --node $NODE
```

**Expected Output:** List of active validators with voting power

### 3. Query DEX Pools

```bash
# List all liquidity pools
$AURAD query dex pools --node $NODE --output json

# Query specific pool (if exists)
$AURAD query dex pool uaura-usdt --node $NODE
```

**Expected Output (empty testnet):**
```json
{"pools":[],"pagination":{"next_key":null,"total":"0"}}
```

### 4. Query Supported Coins

```bash
$AURAD query dex supported-coins --node $NODE --output json
```

**Expected Output (empty testnet):**
```json
{"coins":[]}
```

### 5. Query Account (if you know an address)

```bash
# Replace with actual address
$AURAD query account aura1... --node $NODE --output json
```

### 6. Query Transaction by Hash

```bash
# If you have a tx hash
$AURAD query tx <HASH> --node $NODE --output json
```

---

## Transaction Commands (Require Keys)

These examples show the command structure. To actually execute them, you need:
1. An account with keys in the keyring
2. Tokens in that account

### Setting Up Test Keys

```bash
# Create a test key
$AURAD keys add alice --keyring-backend test

# Or import existing key
$AURAD keys add alice --recover --keyring-backend test
# (then paste mnemonic)

# List keys
$AURAD keys list --keyring-backend test
```

### DEX Examples

#### Create a Liquidity Pool

```bash
$AURAD tx dex create-pool uaura 1000000 usdt 500000 \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --gas auto \
  --gas-adjustment 1.3 \
  --fees 5000uaura \
  --yes
```

**What this does:**
- Creates pool with 1,000,000 uaura and 500,000 usdt
- Pool ID will be: uaura-usdt
- Initial price: 1 AURA = 0.5 USDT

#### Execute a Swap

```bash
$AURAD tx dex swap uaura-usdt 100000uaura 48000 500 \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --gas auto \
  --fees 5000uaura \
  --yes
```

**What this does:**
- Swaps 100,000 uaura for USDT
- Minimum output: 48,000 (slippage protection)
- Max slippage: 500 bps (5%)

#### Get Swap Quote First

```bash
# Check price before swapping
$AURAD query dex quote uaura-usdt uaura 100000 --node $NODE --output json
```

**Returns:**
- Estimated output amount
- Effective price
- Price impact %
- Fee amount

#### Create P2P Order

```bash
# Alice creates a sell order
$AURAD tx dex create-order sell 1000000 usdt 500000 \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --fees 5000uaura \
  --yes
```

**What this does:**
- Alice offers to sell 1,000,000 AURA
- For 500,000 USDT
- Order stays in orderbook until matched or cancelled

#### Create HTLC for Atomic Swap

```bash
# Generate secret and hash
SECRET=$(openssl rand -hex 32)
SECRET_HASH=$(echo -n $SECRET | sha256sum | cut -d' ' -f1)

# Create HTLC
$AURAD tx dex create-htlc aura1recipient... 1000000uaura $SECRET_HASH 3600 \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --fees 5000uaura \
  --yes
```

**What this does:**
- Locks 1,000,000 uaura for recipient
- Recipient must reveal secret within 3600 seconds (1 hour)
- If timeout expires, Alice can refund

### Compliance Examples

#### Submit KYC Verification

```bash
# Generate PII commitment (off-chain)
PII_HASH=$(echo -n "user-pii-data" | sha256sum | cut -d' ' -f1)

# Submit KYC
$AURAD tx compliance submit-kyc aura1user... 3 aura1provider... $PII_HASH US \
  --from provider \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --fees 5000uaura \
  --yes
```

**What this does:**
- Sets KYC level to 3 (INTERMEDIATE) for user
- Stores hash of PII (not actual PII - GDPR compliant)
- Jurisdiction: US

#### Screen for Sanctions

```bash
$AURAD tx compliance screen-sanctions aura1user... \
  --from anyone \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --fees 5000uaura \
  --yes
```

**What this does:**
- Checks address against OFAC/EU/UN sanctions lists
- Results cached for efficiency
- Use --force-refresh to bypass cache

#### Query KYC Record

```bash
$AURAD query compliance kyc-record aura1user... --node $NODE --output json
```

**Returns:**
- KYC level
- Provider address
- PII commitment hash
- Jurisdiction
- Verification timestamp

### Confidence Score Examples

#### Record IR Completion (Assistant Only)

```bash
$AURAD tx confidencescore record-completion aura1user... IR-042-FINANCE 500 \
  --from assistant \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --fees 5000uaura \
  --yes
```

**What this does:**
- Records user completed IR-042-FINANCE
- Adds 500 points to user's confidence score
- Only AI assistant account can call this

#### Query User Score

```bash
$AURAD query confidencescore score aura1user... --node $NODE --output json
```

**Returns:**
- Total confidence score
- Verification status (verified if >= 10,000)
- Arena breakdown
- IR completion count
- Last update time

### Wasm Contract Examples

#### Upload Contract

```bash
# Upload wasm binary
$AURAD tx aura_wasm_security store /path/to/contract.wasm \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --gas 3000000 \
  --fees 50000uaura \
  --yes
```

**What this does:**
- Uploads compiled wasm contract
- Returns code ID
- Requires authorization (governance or authorized uploader)

#### Instantiate Contract

```bash
# Instantiate with init message
$AURAD tx aura_wasm_security instantiate 1 '{"count":0}' \
  --label "Counter Contract" \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --gas 500000 \
  --fees 10000uaura \
  --yes
```

**What this does:**
- Creates new instance of code ID 1
- Initializes with {"count":0}
- Returns contract address

### Standard Cosmos Commands

#### Send Tokens

```bash
$AURAD tx bank send alice aura1recipient... 1000000uaura \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --fees 5000uaura \
  --yes
```

#### Delegate to Validator

```bash
$AURAD tx staking delegate auravaloper1... 1000000uaura \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --fees 5000uaura \
  --yes
```

#### Withdraw Staking Rewards

```bash
$AURAD tx distribution withdraw-all-rewards \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --fees 5000uaura \
  --yes
```

---

## Advanced Query Examples

### Query by Block Height

```bash
# Query pools at specific height
$AURAD query dex pools --node $NODE --height 10000 --output json

# Query account at specific height
$AURAD query account aura1... --node $NODE --height 10000
```

### Filter Transaction Events

```bash
# Query transactions by events
$AURAD query txs --events 'message.action=/cosmos.bank.v1beta1.MsgSend' --node $NODE

# Query DEX swaps
$AURAD query txs --events 'message.action=/aura.dex.MsgSwap' --node $NODE
```

### Paginated Queries

```bash
# Query pools with pagination
$AURAD query dex pools --limit 10 --page 1 --node $NODE --output json

# Query with count total
$AURAD query dex pools --count-total --node $NODE --output json
```

---

## Dry Run & Simulation

Test transactions without broadcasting:

```bash
# Simulate gas usage
$AURAD tx dex swap uaura-usdt 100000uaura 48000 500 \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --dry-run

# Generate unsigned transaction
$AURAD tx bank send alice aura1... 1000uaura \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --generate-only > unsigned_tx.json

# Sign transaction
$AURAD tx sign unsigned_tx.json \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE > signed_tx.json

# Broadcast signed transaction
$AURAD tx broadcast signed_tx.json --node $NODE
```

---

## Troubleshooting

### Check if Node is Reachable

```bash
curl -s http://localhost:27657/status
```

If this fails, the node is not running or not accessible.

### Check if Account Has Tokens

```bash
$AURAD query account aura1... --node $NODE --output json | jq '.account.base_account.address'
```

### Check Transaction Status

```bash
# After broadcasting, check tx status
$AURAD query tx <TX_HASH> --node $NODE --output json
```

### View Validator RPC Endpoints

```bash
# Validator 1
curl -s http://localhost:27657/status

# Validator 2
curl -s http://localhost:27757/status

# Validator 3
curl -s http://localhost:27857/status

# Validator 4
curl -s http://localhost:27957/status
```

---

## Common Flags Reference

### Transaction Flags

```bash
--from <key>                    # Signer (required)
--chain-id <chain-id>           # Chain ID (required for broadcast)
--node <rpc-endpoint>           # Node to broadcast to
--keyring-backend <backend>     # test|os|file|kwallet|pass|memory
--fees <amount>                 # Fixed fees (e.g., 5000uaura)
--gas-prices <price>            # Gas price (e.g., 0.025uaura)
--gas <limit>                   # Gas limit (or "auto")
--gas-adjustment <factor>       # Multiplier for auto gas (e.g., 1.3)
--yes                           # Skip confirmation
--dry-run                       # Simulate without broadcasting
--generate-only                 # Output unsigned tx
--broadcast-mode <mode>         # sync|async|block
--note <text>                   # Memo/note for tx
```

### Query Flags

```bash
--node <rpc-endpoint>           # Node to query
--output <format>               # text|json|yaml
--height <block-height>         # Query at specific height
--page <number>                 # Page number for pagination
--limit <number>                # Results per page
--count-total                   # Include total count
```

---

## Real Workflow Examples

### Example 1: Check if Pool Exists Before Swapping

```bash
# Step 1: Check if pool exists
$AURAD query dex pool uaura-usdt --node $NODE

# Step 2: Get quote
$AURAD query dex quote uaura-usdt uaura 100000 --node $NODE --output json

# Step 3: Execute swap
$AURAD tx dex swap uaura-usdt 100000uaura 48000 500 \
  --from alice \
  --keyring-backend test \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --fees 5000uaura \
  --yes
```

### Example 2: Complete KYC Verification Flow

```bash
# Step 1: Generate PII hash off-chain
PII_HASH=$(echo -n "FirstName LastName DOB SSN" | sha256sum | cut -d' ' -f1)

# Step 2: Submit KYC
$AURAD tx compliance submit-kyc aura1user... 3 aura1provider... $PII_HASH US \
  --from provider --keyring-backend test --chain-id $CHAIN_ID --node $NODE --fees 5000uaura --yes

# Step 3: Screen for sanctions
$AURAD tx compliance screen-sanctions aura1user... \
  --from anyone --keyring-backend test --chain-id $CHAIN_ID --node $NODE --fees 5000uaura --yes

# Step 4: Query KYC status
$AURAD query compliance kyc-record aura1user... --node $NODE --output json

# Step 5: Query sanctions results
$AURAD query compliance sanctions aura1user... --node $NODE --output json

# Step 6: Query AML profile
$AURAD query compliance aml-profile aura1user... --node $NODE --output json
```

### Example 3: Cross-Chain Atomic Swap

```bash
# On Aura Chain (Alice):
SECRET=$(openssl rand -hex 32)
SECRET_HASH=$(echo -n $SECRET | sha256sum | cut -d' ' -f1)
echo "Secret: $SECRET"
echo "Hash: $SECRET_HASH"

# Alice creates HTLC on Aura
$AURAD tx dex create-htlc aura1bob... 1000000uaura $SECRET_HASH 7200 \
  --from alice --keyring-backend test --chain-id $CHAIN_ID --node $NODE --fees 5000uaura --yes

# Get HTLC ID from transaction receipt
HTLC_ID=<from-tx-output>

# Bob creates matching HTLC on other chain (using same hash)
# ... (on other chain)

# Alice claims Bob's HTLC, revealing secret
# ... (on other chain)

# Bob claims Alice's HTLC using revealed secret
$AURAD tx dex claim-htlc $HTLC_ID $SECRET \
  --from bob --keyring-backend test --chain-id $CHAIN_ID --node $NODE --fees 5000uaura --yes
```

---

## Next Steps

1. **Set up test keys** using `aurad keys add`
2. **Get tokens** from the faucet (if available) or genesis accounts
3. **Try query commands** to explore the chain state
4. **Create test pools** and execute swaps
5. **Experiment with compliance** features
6. **Deploy test contracts** using wasm security module

For more help:
- Use `--help` flag on any command
- See CLI_QUICK_REFERENCE.md for common patterns
- See CLI_COMMAND_TEST_REPORT.md for comprehensive documentation
