---
id: "053"
title: "DEX No Slippage Protection Enforcement"
status: done
priority: p2
category: security
module: dex
severity: HIGH
source: dex-security-audit
completed: 2025-12-02
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

- [x] minAmountOut parameter enforced in all swap paths
- [N/A] maxAmountIn parameter for reverse swaps (no SwapExactOut function exists)
- [x] Slippage check happens BEFORE state changes
- [x] Tests for slippage rejection

## Resolution

After thorough analysis and testing, confirmed that **slippage protection is already properly implemented**:

### Existing Protection (SwapExactIn)

The `SwapExactIn` function in `chain/x/dex/keeper/liquidity_pool.go` already enforces slippage protection correctly:

1. **Line 609-616**: Enforces `minAmountOut` check BEFORE any state changes
2. **Line 631-638**: Enforces `maxSlippageBps` price impact check BEFORE state changes
3. **Line 671-683**: State updates only happen AFTER all checks pass
4. **Line 701-719**: Token transfers happen last

The implementation correctly follows the checks-effects-interactions pattern, making it secure against sandwich attacks and price manipulation.

### Comprehensive Test Suite Added

Added 8 comprehensive tests in `chain/x/dex/keeper/slippage_protection_test.go`:

1. ✅ `TestSlippageEnforcementBeforeStateChanges` - Verifies state isn't modified on slippage failure
2. ✅ `TestMinAmountOutEnforcement` - Tests minAmountOut parameter enforcement
3. ✅ `TestMaxSlippageBpsEnforcement` - Tests price impact limits
4. ✅ `TestSlippageProtectionAgainstSandwichAttack` - Simulates and prevents sandwich attack
5. ✅ `TestSlippageProtectionBidirectional` - Tests both swap directions
6. ✅ `TestSlippageProtectionEdgeCases` - Tests edge cases
7. ✅ `TestSlippageProtectionWithMultipleSwaps` - Tests consistency
8. ✅ `TestSlippageRejectionErrorMessages` - Verifies clear error messages

All tests pass, confirming the protection works as designed.

### No Action Required

The existing code is **SECURE**. No code changes were necessary - only comprehensive tests were added to verify and document the existing security measures.

### Notes

- No `SwapExactOut` (reverse swap) function exists, so `maxAmountIn` parameter is not applicable
- The orderbook `ExecuteSwap` function uses fixed prices from orders, not AMM calculation, so slippage protection works differently there (price is agreed upon when order is created)

Commit: 8fa24ba8c12980a52b0ab2aaefb1391e51409cb9
