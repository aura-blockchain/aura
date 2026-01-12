# AURA MVP CLI Reference

Quick reference for MVP module CLI commands.

## Identity Module

```bash
# Register identity
aurad tx identity register [did] --from <key>

# Query identity by DID
aurad query identity did [did]

# List all identities
aurad query identity list

# Update identity document
aurad tx identity update [did] [document-json] --from <key>
```

## VC Registry Module

```bash
# Register schema
aurad tx vcregistry register-schema [schema-json] --from <key>

# Query schema
aurad query vcregistry schema [schema-id]

# Issue credential
aurad tx vcregistry issue [holder-did] [schema-id] [claims-json] --from <key>

# Verify credential
aurad query vcregistry verify [vc-id]

# Revoke credential
aurad tx vcregistry revoke [vc-id] [reason] --from <key>

# List credentials by holder
aurad query vcregistry credentials [holder-did]
```

## Data Registry Module

```bash
# Store data item
aurad tx dataregistry store [data-id] [content-hash] \
  --access-mode [private|public|whitelist|verified-users] \
  --from <key>

# Query data item
aurad query dataregistry item [data-id]

# List data items by owner
aurad query dataregistry list [owner-address]

# Update data item
aurad tx dataregistry update [data-id] [new-content-hash] --from <key>

# Delete data item
aurad tx dataregistry delete [data-id] --from <key>

# Verify data item
aurad tx dataregistry verify [data-id] [level] --from <key>

# Query verifications
aurad query dataregistry verifications [data-id]
```

## Compliance Module

```bash
# Query compliance rules
aurad query compliance rules

# Check address compliance
aurad query compliance check [address]

# Query compliance status
aurad query compliance status [address]
```

## Governance Module

```bash
# Submit proposal
aurad tx governance submit-proposal [proposal-json] --deposit [amount] --from <key>

# Query proposal
aurad query governance proposal [proposal-id]

# List proposals
aurad query governance proposals

# Vote on proposal
aurad tx governance vote [proposal-id] [yes|no|abstain|no-with-veto] --from <key>

# Deposit to proposal
aurad tx governance deposit [proposal-id] [amount] --from <key>

# Query votes
aurad query governance votes [proposal-id]

# Query tally
aurad query governance tally [proposal-id]
```

## Standard Cosmos SDK Commands

### Auth

```bash
# Query account
aurad query auth account [address]

# Query accounts
aurad query auth accounts
```

### Bank

```bash
# Query balance
aurad query bank balances [address]

# Send tokens
aurad tx bank send [from] [to] [amount] --from <key>

# Query supply
aurad query bank total
```

### Staking

```bash
# Delegate
aurad tx staking delegate [validator-addr] [amount] --from <key>

# Unbond
aurad tx staking unbond [validator-addr] [amount] --from <key>

# Redelegate
aurad tx staking redelegate [src-validator] [dst-validator] [amount] --from <key>

# Query validators
aurad query staking validators

# Query delegations
aurad query staking delegations [delegator-addr]
```

### Distribution

```bash
# Withdraw rewards
aurad tx distribution withdraw-rewards [validator-addr] --from <key>

# Withdraw all rewards
aurad tx distribution withdraw-all-rewards --from <key>

# Query rewards
aurad query distribution rewards [delegator-addr]
```

## Key Management

```bash
# Add key
aurad keys add [name]

# List keys
aurad keys list

# Show key
aurad keys show [name]

# Export key
aurad keys export [name]

# Import key
aurad keys import [name] [keyfile]

# Delete key
aurad keys delete [name]
```

## Node Operations

```bash
# Initialize node
aurad init [moniker] --chain-id [chain-id]

# Start node
aurad start

# Show node ID
aurad tendermint show-node-id

# Show validator key
aurad tendermint show-validator

# Export state
aurad export

# Validate genesis
aurad genesis validate [genesis-file]
```

## Transaction Flags

Common flags for transactions:

| Flag | Description |
|------|-------------|
| `--from` | Key name or address to sign with |
| `--fees` | Transaction fees (e.g., `1000uaura`) |
| `--gas` | Gas limit (or `auto`) |
| `--gas-adjustment` | Adjustment factor for auto gas |
| `--chain-id` | Chain ID |
| `--node` | RPC endpoint |
| `--broadcast-mode` | `sync`, `async`, or `block` |
| `--yes` | Skip confirmation prompt |

## Query Flags

Common flags for queries:

| Flag | Description |
|------|-------------|
| `--node` | RPC endpoint |
| `--output` | Output format (`json` or `text`) |
| `--height` | Query at specific height |

## Examples

### Register Identity and Issue Credential

```bash
# 1. Register identity
aurad tx identity register did:aura:alice123 --from alice

# 2. Register credential schema
aurad tx vcregistry register-schema '{"name":"degree","fields":["institution","degree","year"]}' --from issuer

# 3. Issue credential
aurad tx vcregistry issue did:aura:alice123 schema-1 '{"institution":"MIT","degree":"PhD","year":"2024"}' --from issuer

# 4. Verify credential
aurad query vcregistry verify vc-1
```

### Store and Share Data

```bash
# 1. Store private data
aurad tx dataregistry store mydata-001 QmHash123 --access-mode private --from alice

# 2. Update to whitelist access
aurad tx dataregistry update mydata-001 QmHash123 \
  --access-mode whitelist \
  --allowed-addresses aura1bob,aura1carol \
  --from alice

# 3. Check access
aurad query dataregistry item mydata-001
```
