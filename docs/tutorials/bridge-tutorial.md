# Bridge Operations Tutorial

## Overview
Transfer tokens between Aura and external chains.

## Supported Chains
- Ethereum (ETH, ERC-20)
- Polygon (MATIC)
- Cosmos Hub (ATOM via IBC - when enabled)

## Lock Tokens (Aura → External)

### CLI
```bash
aurad tx bridge lock-tokens \
  --amount 1000uaura \
  --destination-chain ethereum \
  --recipient 0x742d35Cc6634C0532925a3b844Bc9e7595f... \
  --from mykey
```

### REST
```bash
curl -X POST http://localhost:1317/aura/bridge/v1beta1/lock \
  -d '{
    "sender": "aura1...",
    "amount": {"denom": "uaura", "amount": "1000"},
    "destination_chain": "ethereum",
    "recipient": "0x742d35..."
  }'
```

## Mint Tokens (External → Aura)

```bash
aurad tx bridge mint-tokens \
  --amount 1000 \
  --source-chain ethereum \
  --source-tx-hash 0xabc123... \
  --recipient aura1... \
  --from relayer
```

## Query Pending Transfers

```bash
aurad query bridge pending-transfers --limit 10
```

## Check Transfer Status

```bash
aurad query bridge transfer <transfer-id>
```

## Relay Transfer (Relayers Only)

```bash
aurad tx bridge relay-transfer \
  --transfer-id abc123 \
  --proof <merkle-proof> \
  --from relayer
```
