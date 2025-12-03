---
id: "053"
title: "DEX No Slippage Protection Enforcement"
status: ready
priority: p2
category: security
module: dex
severity: HIGH
source: dex-security-audit
---

# DEX No Slippage Protection Enforcement

## Problem

Swap functions accept slippage parameters but don't enforce them properly, allowing sandwich attacks.

## Affected Files

- `chain/x/dex/keeper/liquidity_pool.go`

## Impact

- Sandwich attacks extract value from swaps
- Front-running profitable for attackers
- Users receive less than expected

## Required Fix

```go
func (k Keeper) ExecuteSwap(..., minAmountOut sdkmath.Int) (sdkmath.Int, error) {
    // Calculate output
    amountOut := k.calculateSwapOutput(...)

    // ENFORCE slippage protection
    if amountOut.LT(minAmountOut) {
        return sdkmath.ZeroInt(), fmt.Errorf(
            "slippage exceeded: output %s < minimum %s",
            amountOut, minAmountOut)
    }

    // Proceed with swap
    return amountOut, nil
}
```

## Acceptance Criteria

- [ ] minAmountOut parameter enforced in all swap paths
- [ ] maxAmountIn parameter for reverse swaps
- [ ] Slippage check happens BEFORE state changes
- [ ] Tests for slippage rejection
