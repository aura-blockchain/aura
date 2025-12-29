---
sidebar_position: 6
---

# DEX Trading

**Difficulty:** Beginner | **Time:** 5 minutes

Trade tokens on Aura's decentralized exchange.

## DEX Overview

Aura DEX features:
- **AMM Pools**: Automated market makers for liquidity
- **Limit Orders**: Set your price
- **IBC Support**: Trade cross-chain assets

## Swapping Tokens

### View Available Pools

```bash
aurad query dex pools
```

### Execute a Swap

```bash
aurad tx dex swap \
  --input 1000000uaura \
  --output-denom uibc/atom \
  --slippage 0.01 \
  --from my-wallet \
  --chain-id aura-testnet-1 \
  -y
```

Parameters:
- `--slippage`: Maximum acceptable price impact (0.01 = 1%)

### Get Swap Quote

```bash
aurad query dex quote \
  --input 1000000uaura \
  --output-denom uibc/atom
```

## Providing Liquidity

### Add Liquidity

```bash
aurad tx dex add-liquidity \
  --pool-id 1 \
  --amounts "1000000uaura,500000uibc/atom" \
  --from my-wallet \
  -y
```

You receive LP tokens representing your share.

### Remove Liquidity

```bash
aurad tx dex remove-liquidity \
  --pool-id 1 \
  --lp-tokens 100000 \
  --from my-wallet \
  -y
```

## Limit Orders

### Place Order

```bash
aurad tx dex place-order \
  --sell 1000000uaura \
  --buy uibc/atom \
  --price 0.5 \
  --from my-wallet \
  -y
```

### Cancel Order

```bash
aurad tx dex cancel-order <order-id> \
  --from my-wallet \
  -y
```

### View Your Orders

```bash
aurad query dex orders-by-address $(aurad keys show my-wallet -a)
```

## Trading Fees

| Action | Fee |
|--------|-----|
| Swap | 0.3% |
| Add Liquidity | 0% |
| Remove Liquidity | 0% |

Fees go to liquidity providers.

## Next Steps

- [Stake Tokens](./stake-tokens) - Earn staking rewards
- [Module Guide: DEX](/docs/modules/dex) - Advanced DEX features
