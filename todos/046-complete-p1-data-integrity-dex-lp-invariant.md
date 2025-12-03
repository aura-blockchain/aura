---
id: "046"
title: "DEX LP Token Supply Not Validated Against Reserves"
status: ready
priority: p1
category: data-integrity
module: dex
severity: CRITICAL
source: data-integrity-review
---

# DEX LP Token Supply Not Validated Against Reserves

## Problem

LP tokens are minted without atomic verification that total shares match total reserves. Pool state updated in multiple separate operations without transaction boundaries.

## Affected Files

- `chain/x/dex/keeper/liquidity_pool.go:121-200`

## Impact

- LP token supply inflation without reserve backing
- Theft of liquidity from other providers
- Pool insolvency

## Required Fix

Add invariant check at end of every liquidity operation:

```go
func (k Keeper) AddLiquidity(...) (sdkmath.Int, sdkmath.LegacyDec, error) {
    // ... existing logic ...

    // CRITICAL: Verify invariant at end
    sumProviderShares := sdkmath.ZeroInt()
    for _, provider := range pool.Providers {
        providerTokens, _ := k.parseLPTokens(provider.LpTokens)
        sumProviderShares = sumProviderShares.Add(providerTokens)
    }

    updatedTotal, _ := k.parseLPTokens(pool.TotalLpTokens)
    lockedLiquidity, _ := sdkmath.NewIntFromString(pool.LockedLiquidity)

    expectedTotal := sumProviderShares.Add(lockedLiquidity)

    if !expectedTotal.Equal(updatedTotal) {
        return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), fmt.Errorf(
            "CRITICAL: LP token invariant violated - sum(%s) + locked(%s) != total(%s)",
            sumProviderShares, lockedLiquidity, updatedTotal,
        )
    }

    return lpTokens, sharePercentage, nil
}
```

## Acceptance Criteria

- [ ] Invariant check added to AddLiquidity
- [ ] Invariant check added to RemoveLiquidity
- [ ] Invariant check added to Swap
- [ ] Module-level invariant for all pools
- [ ] Tests for invariant enforcement
