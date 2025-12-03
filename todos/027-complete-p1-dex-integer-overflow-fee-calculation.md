---
id: "027"
title: "Integer Overflow in DEX Fee Calculation"
status: ready
priority: p1
category: security
module: dex
severity: HIGH
cvss: 7.2
source: security-audit-report
---

# Integer Overflow in DEX Fee Calculation

## Problem

The `CalculateFeeBoost` function multiplies user-controlled values without overflow protection.

## Affected Files

- `chain/x/dex/keeper/keeper.go`

## Vulnerability

When calculating fee boosts, large user-controlled values can cause integer overflow leading to:
- Negative fees (attacker gets paid)
- Zero fees (attacker pays nothing)
- Incorrect fee distribution

## Impact

- Economic exploit
- Fee bypass
- Protocol revenue loss

## Required Fix

```go
import (
    sdkmath "cosmossdk.io/math"
)

func (k Keeper) CalculateFeeBoost(ctx sdk.Context, amount sdkmath.Int, boostMultiplier sdkmath.Int) (sdkmath.Int, error) {
    // Validate inputs
    if amount.IsNegative() {
        return sdkmath.ZeroInt(), fmt.Errorf("amount cannot be negative")
    }

    if boostMultiplier.IsNegative() {
        return sdkmath.ZeroInt(), fmt.Errorf("boost multiplier cannot be negative")
    }

    // Check for potential overflow before multiplication
    maxSafeValue := sdkmath.NewIntFromBigInt(new(big.Int).Sqrt(sdkmath.MaxInt256.BigInt()))

    if amount.GT(maxSafeValue) || boostMultiplier.GT(maxSafeValue) {
        return sdkmath.ZeroInt(), fmt.Errorf("values too large, potential overflow")
    }

    // Safe multiplication using SDK math (automatically handles overflow)
    result := amount.Mul(boostMultiplier)

    // Sanity check result
    if result.IsNegative() {
        return sdkmath.ZeroInt(), fmt.Errorf("overflow detected: result is negative")
    }

    return result, nil
}

// Use SafeMath for all fee calculations
func (k Keeper) CalculateSwapFee(ctx sdk.Context, amountIn sdkmath.Int) (sdkmath.Int, error) {
    params := k.GetParams(ctx)

    // Convert fee rate to Int (e.g., 30 basis points = 0.003)
    // Fee = amountIn * feeRate / 10000

    feeNumerator, overflow := amountIn.SafeMul(sdkmath.NewInt(int64(params.SwapFeeBasisPoints)))
    if overflow {
        return sdkmath.ZeroInt(), fmt.Errorf("fee calculation overflow")
    }

    fee := feeNumerator.Quo(sdkmath.NewInt(10000))

    return fee, nil
}
```

## Acceptance Criteria

- [ ] SafeMath used for all arithmetic operations
- [ ] Overflow checks on all multiplications
- [ ] Input validation for all user-controlled values
- [ ] Tests for overflow edge cases
- [ ] Tests for negative value rejection
