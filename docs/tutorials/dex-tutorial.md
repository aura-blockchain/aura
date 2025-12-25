# DEX Trading Tutorial

## Overview
Trade tokens and provide liquidity on Aura DEX.

## Token Swap

### CLI
```bash
aurad tx dex swap \
  --input 1000uaura \
  --output-denom usdc \
  --min-output 990 \
  --from mykey
```

### REST
```bash
curl -X POST http://localhost:1317/aura/dex/v1beta1/swap \
  -d '{
    "sender": "aura1...",
    "input": {"denom": "uaura", "amount": "1000"},
    "output_denom": "usdc",
    "min_output": "990"
  }'
```

## Add Liquidity

```bash
aurad tx dex add-liquidity \
  --pool-id 1 \
  --token-a 1000uaura \
  --token-b 1000usdc \
  --from mykey
```

## Remove Liquidity

```bash
aurad tx dex remove-liquidity \
  --pool-id 1 \
  --lp-tokens 500 \
  --from mykey
```

## Query Pool

```bash
aurad query dex pool 1
aurad query dex pools --limit 10
```

## Get Quote

```bash
aurad query dex quote \
  --input 1000uaura \
  --output-denom usdc
```

## Create Limit Order

```bash
aurad tx dex create-order \
  --pool-id 1 \
  --side buy \
  --amount 1000 \
  --price 1.05 \
  --from mykey
```
