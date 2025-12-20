# Aura DEX Quick Start Guide

Get started with Aura's decentralized exchange in minutes.

## What is Aura DEX?

Aura DEX is an automated market maker (AMM) with P2P orderbook and atomic swaps:
- AMM liquidity pools with constant product formula
- P2P orderbook for limit orders
- HTLC atomic swaps for trustless cross-chain trades
- Privacy-preserving trades via identity module integration

## Prerequisites

- Aura daemon installed (`aurad`)
- Wallet with AURA tokens
- Node running or access to RPC endpoint

## Basic Setup

### 1. Check Your Balance

```bash
aurad query bank balances $(aurad keys show alice -a)
```

### 2. View Available Pools

```bash
aurad query dex pools
```

### 3. Get a Price Quote

```bash
aurad query dex quote 1 uaura 1000000
```

## AMM Liquidity Pools

### Create a Pool

```bash
aurad tx dex create-pool uaura 1000000000 uusdt 2000000000 \
  --from alice \
  --chain-id aura-testnet-1 \
  --gas auto
```

### Add Liquidity

```bash
aurad tx dex add-liquidity 1 uaura 100000000 uusdt 200000000 \
  --from alice \
  --chain-id aura-testnet-1
```

### Swap Tokens

```bash
# Swap 1000 AURA for USDT with 1% max slippage
aurad tx dex swap 1 1000uaura 990 100 \
  --from alice \
  --chain-id aura-testnet-1
```

### Remove Liquidity

```bash
aurad tx dex remove-liquidity 1 500000 \
  --from alice \
  --chain-id aura-testnet-1
```

## P2P Orderbook

### Create a Limit Order

```bash
# Sell 1000 AURA for 2000 USDT
aurad tx dex create-order sell 1000000000uaura uusdt 2000000000 \
  --from alice \
  --chain-id aura-testnet-1
```

### View Orderbook

```bash
aurad query dex orderbook uaura/uusdt
```

### Match an Order

```bash
aurad tx dex match-order [order-id] \
  --from bob \
  --chain-id aura-testnet-1
```

### Cancel an Order

```bash
aurad tx dex cancel-order [order-id] \
  --from alice \
  --chain-id aura-testnet-1
```

## HTLC Atomic Swaps

For trustless cross-chain trades without an intermediary.

### Create an HTLC

```bash
# Lock 1000 AURA for recipient with 24hr timelock
aurad tx dex create-htlc aura1recipient... 1000000000uaura \
  [secret-hash] 86400 \
  --from alice \
  --chain-id aura-testnet-1
```

### Claim an HTLC

```bash
aurad tx dex claim-htlc [htlc-id] [secret] \
  --from bob \
  --chain-id aura-testnet-1
```

### Refund Expired HTLC

```bash
aurad tx dex refund-htlc [htlc-id] \
  --from alice \
  --chain-id aura-testnet-1
```

## Useful Queries

```bash
# Pool statistics
aurad query dex pool-stats 1

# Market price
aurad query dex market-price uaura

# Spot price between tokens
aurad query dex spot-price 1 uaura uusdt

# Your open orders
aurad query dex user-orders $(aurad keys show alice -a)

# Supported trading pairs
aurad query dex supported-coins
```

## Next Steps

- [IBC Quick Start](IBC_QUICK_START.md) - Cross-chain transfers
- [DEX CLI Reference](modules/dex/DEX_CLI_COMMANDS.md) - Full command reference
- [Identity Module](modules/identity/) - Privacy features for trading
