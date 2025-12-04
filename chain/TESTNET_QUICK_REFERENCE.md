# Aura Testnet Quick Reference Guide

**Purpose**: Fast command reference for common testnet operations
**For detailed testing**: See `TESTNET_VALIDATION_SUITE.md`

---

## Environment Setup

```bash
# Set these variables before running commands
export CHAIN_ID="aura-testnet-1"
export NODE="http://localhost:26657"
export VALIDATOR_HOME="$HOME/.testnets/aura-testnet/node0/aurad"
export VALIDATOR_KEY="validator0"
```

---

## Quick Health Check (30 seconds)

```bash
# 1. Check node is running
aurad status --node $NODE | jq '.sync_info'

# 2. Check validators
aurad query staking validators --node $NODE | grep moniker

# 3. Check your balance
ADDR=$(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend test --address)
aurad query bank balances $ADDR --node $NODE

# 4. Check latest block
curl -s $NODE/status | jq '.result.sync_info.latest_block_height'
```

**Expected**: Node is synced, validators are bonded, balance > 0, blocks incrementing

---

## Essential Commands

### Node Operations

```bash
# Check node status
aurad status --node $NODE

# Check consensus
curl -s $NODE/consensus_state | jq '.result.round_state.height'

# Check peers
curl -s $NODE/net_info | jq '.result.n_peers'

# Check if catching up
aurad status --node $NODE | jq '.sync_info.catching_up'
```

### Account Management

```bash
# List all keys
aurad keys list --home $VALIDATOR_HOME --keyring-backend test

# Show specific key address
aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend test --address

# Create new account
aurad keys add myaccount --home $VALIDATOR_HOME --keyring-backend test

# Query account balance
aurad query bank balances <address> --node $NODE

# Query account info
aurad query auth account <address> --node $NODE
```

### Transactions

```bash
# Send tokens
aurad tx bank send <from_address> <to_address> <amount>uaura \
  --from <key_name> \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# Query transaction by hash
aurad query tx <txhash> --node $NODE

# Check mempool
curl -s $NODE/num_unconfirmed_txs | jq '.'
```

---

## Module-Specific Commands

### DEX Module

```bash
# Create liquidity pool
aurad tx dex create-pool <denom_a> <denom_b> <amount_a> <amount_b> \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# Query pool
aurad query dex pool <pool_id> --node $NODE

# Execute swap
aurad tx dex swap-exact-in <pool_id> <coin_in> <min_amount_out> <max_slippage_bps> \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# List all pools
aurad query dex pools --node $NODE
```

### Bridge Module

```bash
# Lock tokens for cross-chain transfer
aurad tx bridge lock-tokens <target_chain> <recipient> <amount> \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# Query transfer status
aurad query bridge transfer <transfer_id> --node $NODE

# Query all transfers
aurad query bridge transfers --node $NODE
```

### Compliance Module

```bash
# Submit KYC record
PII_HASH=$(echo -n '{"name":"John Doe"}' | sha256sum | awk '{print $1}')
aurad tx compliance submit-kyc <address> <kyc_level> <provider> $PII_HASH <jurisdiction> \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# Query KYC record
aurad query compliance kyc-record <address> --node $NODE

# Screen sanctions
aurad tx compliance screen-sanctions <address> false \
  --from <address> \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```

### Identity Module

```bash
# Create role
aurad tx identity create-role <role_name> <description> <permissions> \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# Assign role
aurad tx identity assign-role <address> <role_name> <expires_at> \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# Query roles
aurad query identity roles <address> --node $NODE
```

### WASM Module

```bash
# Store contract code
aurad tx wasm store <contract.wasm> \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --gas auto \
  --gas-adjustment 1.3 \
  --fees 50000uaura \
  --yes

# Query code info
aurad query wasm code <code_id> --node $NODE

# Instantiate contract
aurad tx wasm instantiate <code_id> '<init_msg_json>' \
  --from $VALIDATOR_KEY \
  --label "my-contract" \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --admin <admin_address> \
  --fees 20000uaura \
  --yes

# Execute contract
aurad tx wasm execute <contract_address> '<execute_msg_json>' \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 10000uaura \
  --yes

# Query contract state
aurad query wasm contract-state smart <contract_address> '<query_msg_json>' \
  --node $NODE

# List contracts by code
aurad query wasm list-contract-by-code <code_id> --node $NODE
```

### Governance Module

```bash
# Submit proposal
aurad tx governance submit-proposal \
  "<title>" \
  "<description>" \
  <category> \
  <proposer> \
  <initial_deposit>uaura \
  <is_emergency> \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# Add deposit
aurad tx governance deposit <proposal_id> <depositor> <amount>uaura \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# Vote on proposal
aurad tx governance vote <proposal_id> <voter> <vote_option> false "" \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes

# Query proposal
aurad query governance proposal <proposal_id> --node $NODE

# List all proposals
aurad query governance proposals --node $NODE
```

### Staking Operations

```bash
# Query validators
aurad query staking validators --node $NODE

# Query specific validator
VALIDATOR_OPER=$(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend test --bech val --address)
aurad query staking validator $VALIDATOR_OPER --node $NODE

# Query staking pool
aurad query staking pool --node $NODE

# Query delegations to validator
aurad query staking delegations-to $VALIDATOR_OPER --node $NODE

# Query delegations from delegator
aurad query staking delegations <delegator_address> --node $NODE
```

### Distribution & Rewards

```bash
# Query validator commission
aurad query distribution commission $VALIDATOR_OPER --node $NODE

# Query validator outstanding rewards
aurad query distribution validator-outstanding-rewards $VALIDATOR_OPER --node $NODE

# Query delegator rewards
aurad query distribution rewards <delegator_address> --node $NODE

# Query community pool
aurad query distribution community-pool --node $NODE
```

---

## Troubleshooting Commands

### Connection Issues

```bash
# Test node connectivity
curl -s $NODE/health

# Check if node is syncing
aurad status --node $NODE | jq '.sync_info.catching_up'

# Check latest block time
aurad status --node $NODE | jq '.sync_info.latest_block_time'

# List available endpoints
curl -s $NODE/status | jq '.result.node_info.listen_addr'
```

### Transaction Issues

```bash
# Check account sequence
aurad query auth account <address> --node $NODE | jq '.sequence'

# Estimate gas for transaction
# Add --gas auto --gas-adjustment 1.3 to tx commands

# Check minimum gas prices
curl -s $NODE/status | jq '.result.node_info.other.tx_index'

# Decode transaction
aurad tx decode <base64_tx>
```

### Debugging

```bash
# Check logs (if running locally)
tail -f $VALIDATOR_HOME/aurad.log

# Check validator signing info
aurad query slashing signing-info <validator_consensus_pubkey> --node $NODE

# Check validator delegations
aurad query staking delegations-to $VALIDATOR_OPER --node $NODE

# Check module parameters
aurad query bank params --node $NODE
aurad query staking params --node $NODE
aurad query gov params --node $NODE
```

---

## Automated Testing

```bash
# Run critical tests only
./scripts/run-testnet-validation.sh --critical-only

# Run all tests
./scripts/run-testnet-validation.sh --all

# Run specific category
./scripts/run-testnet-validation.sh --category dex

# With custom configuration
CHAIN_ID=aura-testnet-2 NODE=http://localhost:26657 ./scripts/run-testnet-validation.sh --all
```

---

## Performance Testing

```bash
# Monitor block production rate
watch -n 1 'curl -s $NODE/status | jq ".result.sync_info.latest_block_height"'

# Check transaction throughput
curl -s $NODE/num_unconfirmed_txs | jq '.result.n_txs'

# Monitor peer count
watch -n 5 'curl -s $NODE/net_info | jq ".result.n_peers"'

# Check validator uptime
aurad query slashing signing-info $(aurad tendermint show-validator --home $VALIDATOR_HOME) --node $NODE
```

---

## Common Patterns

### Batch Operations

```bash
# Send tokens to multiple addresses
for addr in addr1 addr2 addr3; do
  aurad tx bank send $VALIDATOR_ADDR $addr 1000000uaura \
    --from $VALIDATOR_KEY \
    --chain-id $CHAIN_ID \
    --node $NODE \
    --home $VALIDATOR_HOME \
    --keyring-backend test \
    --fees 5000uaura \
    --yes
  sleep 2
done
```

### Query Loops

```bash
# Monitor balance changes
while true; do
  aurad query bank balances <address> --node $NODE
  sleep 5
done

# Wait for transaction confirmation
TX_HASH=<txhash>
while true; do
  if aurad query tx $TX_HASH --node $NODE &>/dev/null; then
    echo "Transaction confirmed!"
    break
  fi
  echo "Waiting for confirmation..."
  sleep 2
done
```

### Data Export

```bash
# Export all validator info to JSON
aurad query staking validators --node $NODE --output json > validators.json

# Export account balances
aurad query bank balances <address> --node $NODE --output json > balances.json

# Export proposal details
aurad query governance proposal <id> --node $NODE --output json > proposal.json
```

---

## Safety Checklist

Before making transactions on testnet:

- [ ] Verified node is synced: `aurad status --node $NODE | jq '.sync_info.catching_up'` returns `false`
- [ ] Confirmed correct chain ID: Check `aurad status` output
- [ ] Using test keyring: `--keyring-backend test` flag present
- [ ] Sufficient balance: Query balance before transaction
- [ ] Reasonable fees: Start with `--fees 5000uaura` and adjust if needed
- [ ] Transaction parameters are correct: Double-check addresses, amounts, denominations

---

## Emergency Commands

```bash
# If node is stuck/not producing blocks
# 1. Check consensus state
curl -s $NODE/consensus_state | jq '.'

# 2. Check if validator is jailed
aurad query slashing signing-info $(aurad tendermint show-validator --home $VALIDATOR_HOME) --node $NODE

# 3. Restart node (if running locally)
pkill aurad
aurad start --home $VALIDATOR_HOME

# If you need to reset testnet data (WARNING: destroys all data)
aurad tendermint unsafe-reset-all --home $VALIDATOR_HOME
```

---

## Useful Aliases

Add these to your `~/.bashrc` or `~/.zshrc`:

```bash
# Aura testnet aliases
alias aura-status='aurad status --node $NODE | jq ".sync_info"'
alias aura-validators='aurad query staking validators --node $NODE'
alias aura-balance='aurad query bank balances $(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend test --address) --node $NODE'
alias aura-height='curl -s $NODE/status | jq ".result.sync_info.latest_block_height"'
alias aura-peers='curl -s $NODE/net_info | jq ".result.n_peers"'
alias aura-mempool='curl -s $NODE/num_unconfirmed_txs | jq ".result.n_txs"'
```

---

## Resources

- **Full Test Suite**: `TESTNET_VALIDATION_SUITE.md`
- **Automation Script**: `scripts/run-testnet-validation.sh`
- **Cosmos SDK Docs**: https://docs.cosmos.network
- **WASM Docs**: https://docs.cosmwasm.com
- **AURA Documentation**: `../docs/`

---

**Last Updated**: 2025-12-03
