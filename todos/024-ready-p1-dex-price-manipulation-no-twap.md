---
id: "024"
title: "DEX Price Manipulation - No TWAP Protection"
status: ready
priority: p1
category: security
module: dex
severity: HIGH
cvss: 8.2
source: security-audit-report
---

# DEX Price Manipulation - No TWAP Protection

## Problem

The `GetAuraPrice` function (line 76) calculates AURA price from a single pool without TWAP (Time-Weighted Average Price) or multiple oracle sources. Vulnerable to flash loan attacks.

## Affected Files

- `chain/x/dex/keeper/keeper.go:76`

## Vulnerability

```go
func (k Keeper) GetAuraPrice(ctx sdk.Context) sdkmath.LegacyDec {
    pool := k.GetPoolByDenoms(ctx, "uaura", "usdt")
    if pool == nil {
        return sdkmath.LegacyNewDecWithPrec(10, 2) // $0.10
    }

    // Single-block price - VULNERABLE TO MANIPULATION
    price := sdkmath.LegacyNewDecFromInt(reserveB).Quo(sdkmath.LegacyNewDecFromInt(reserveA))
    return price
}
```

## Attack Scenario

```
1. Borrow large USDT amount via flash loan
2. Swap USDT -> AURA in pool (manipulates price up)
3. Use inflated price to bypass minimum liquidity requirements
4. Perform malicious operation (e.g., undercollateralized borrow)
5. Swap back and repay loan
6. Price returns to normal, but damage done
```

## Impact

- Minimum liquidity bypass
- Pool drain attacks
- Economic exploit
- Flash loan manipulation

## Required Fix

```go
func (k Keeper) GetAuraPrice(ctx sdk.Context) sdkmath.LegacyDec {
    // Use Time-Weighted Average Price (TWAP) over 30 minutes
    twapPrice := k.GetTWAP(ctx, "uaura", "usdt", 30*time.Minute)
    if twapPrice.IsZero() {
        // Fallback to multiple oracle sources
        prices := []sdkmath.LegacyDec{
            k.getChainlinkPrice(ctx, "AURA/USD"),
            k.getBandProtocolPrice(ctx, "AURA/USD"),
            k.getPoolPrice(ctx, "uaura", "usdt"),
        }

        // Use median price
        twapPrice = calculateMedian(prices)
    }

    // Sanity check: price shouldn't move more than 10% per block
    lastPrice := k.GetLastRecordedPrice(ctx)
    if !lastPrice.IsZero() {
        maxChange := lastPrice.Mul(sdkmath.LegacyNewDecWithPrec(10, 2))
        if twapPrice.Sub(lastPrice).Abs().GT(maxChange) {
            return lastPrice // Reject suspicious price movement
        }
    }

    k.SetLastRecordedPrice(ctx, twapPrice)
    return twapPrice
}

// TWAP implementation
func (k Keeper) GetTWAP(ctx sdk.Context, denomA, denomB string, duration time.Duration) sdkmath.LegacyDec {
    // Get price observations over duration
    observations := k.GetPriceObservations(ctx, denomA, denomB, duration)
    if len(observations) == 0 {
        return sdkmath.LegacyZeroDec()
    }

    // Calculate time-weighted average
    var totalWeight int64
    var weightedSum sdkmath.LegacyDec = sdkmath.LegacyZeroDec()

    for i := 1; i < len(observations); i++ {
        timeDelta := observations[i].Timestamp - observations[i-1].Timestamp
        weightedSum = weightedSum.Add(observations[i-1].Price.MulInt64(timeDelta))
        totalWeight += timeDelta
    }

    if totalWeight == 0 {
        return sdkmath.LegacyZeroDec()
    }

    return weightedSum.QuoInt64(totalWeight)
}
```

## Acceptance Criteria

- [ ] TWAP implementation with configurable window
- [ ] Price observation recording in BeginBlocker/EndBlocker
- [ ] Sanity check for excessive price movements
- [ ] Multiple price source aggregation
- [ ] Tests for flash loan manipulation resistance
