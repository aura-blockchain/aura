# DEX Module CLI Commands Reference

## Query Commands

```bash
# Pool Queries
aurad query dex pool [pool-id]
aurad query dex pools
aurad query dex pool-stats [pool-id]

# Quote and Price Queries  
aurad query dex quote [pool-id] [denom-in] [amount-in]
aurad query dex market-price [coin]
aurad query dex spot-price [pool-id] [base-denom] [quote-denom]

# Orderbook Queries
aurad query dex orderbook [pair]
aurad query dex order [order-id]
aurad query dex user-orders [address]

# System Queries
aurad query dex supported-coins

# HTLC Queries
aurad query dex htlc [htlc-id]
```

## Transaction Commands

```bash
# AMM Liquidity Pool Commands
aurad tx dex create-pool [denom-a] [amount-a] [denom-b] [amount-b]
aurad tx dex add-liquidity [pool-id] [denom-a] [amount-a] [denom-b] [amount-b]
aurad tx dex remove-liquidity [pool-id] [lp-tokens]
aurad tx dex swap [pool-id] [coin-in] [min-amount-out] [max-slippage-bps]

# P2P Orderbook Commands
aurad tx dex create-order [order-type] [aura-amount] [other-coin] [other-amount]
aurad tx dex match-order [order-id]
aurad tx dex cancel-order [order-id]

# HTLC Atomic Swap Commands
aurad tx dex create-htlc [recipient] [amount] [secret-hash] [timelock-seconds]
aurad tx dex claim-htlc [htlc-id] [secret]
aurad tx dex refund-htlc [htlc-id]
```

## Examples

### Create a liquidity pool for AURA/USDT
```bash
aurad tx dex create-pool uaura 1000000 usdt 500000 --from alice --chain-id aura-1
```

### Swap tokens
```bash
aurad tx dex swap uaura-usdt 100000uaura 48000 500 --from alice --chain-id aura-1
```

### Create a P2P sell order
```bash
aurad tx dex create-order sell 1000000 usdt 500000 --from alice --chain-id aura-1
```

### Query pool statistics
```bash
aurad query dex pool-stats uaura-usdt
```

### Get swap quote
```bash
aurad query dex quote uaura-usdt uaura 1000000
```

### Create HTLC for atomic swap
```bash
aurad tx dex create-htlc aura1recipient... 1000000uaura abc123... 3600 --from alice
```

## All Commands Mapped to Proto Definitions

| Proto RPC | CLI Command | Type |
|-----------|-------------|------|
| Pool | pool | Query |
| AllPools | pools | Query |
| GetQuote | quote | Query |
| PoolStats | pool-stats | Query |
| Orderbook | orderbook | Query |
| Order | order | Query |
| UserOrders | user-orders | Query |
| MarketPrice | market-price | Query |
| SpotPrice | spot-price | Query |
| SupportedCoins | supported-coins | Query |
| HTLC | htlc | Query |
| CreatePool | create-pool | Tx |
| AddLiquidity | add-liquidity | Tx |
| RemoveLiquidity | remove-liquidity | Tx |
| SwapExactIn | swap | Tx |
| CreateOrder | create-order | Tx |
| CancelOrder | cancel-order | Tx |
| ExecuteSwap | match-order | Tx |
| CreateHTLC | create-htlc | Tx |
| ClaimHTLC | claim-htlc | Tx |
| RefundHTLC | refund-htlc | Tx |

**Total: 20 commands (10 queries + 10 transactions)**
