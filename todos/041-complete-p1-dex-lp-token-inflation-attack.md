---
id: "041"
title: "DEX LP Token Inflation Attack (First Depositor)"
status: complete
priority: p1
category: security
module: dex
severity: CRITICAL
cvss: 9.5
source: dex-security-audit
completed_at: 2025-12-03
---

# DEX LP Token Inflation Attack (First Depositor)

## Problem

No minimum liquidity burn on first deposit. First depositor can manipulate LP token value to steal from subsequent depositors.

## Affected Files

- `chain/x/dex/keeper/liquidity_pool.go:75-77`

## Vulnerability

```go
lpTokens := sdkmath.NewIntFromBigInt(new(big.Int).Sqrt(
    amountA.Amount.Mul(amountB.Amount).BigInt(),
))
// BUG: No minimum liquidity burn
// First depositor gets ALL LP tokens
```

## Attack Scenario (Classic First Depositor Attack)

```
1. Attacker creates pool with 1 WEI of each token
2. Receives sqrt(1*1) = 1 LP token
3. Attacker donates (not via AddLiquidity) 10000 tokens to pool
4. Pool now has 10001 WEI TokenA, 1 WEI TokenB
5. But LP supply is still 1

6. Victim adds 1000 TokenA, 1000 TokenB
7. LP calculation: min(1000*1/10001, 1000*1/1) = ~0.099 LP tokens
8. Due to rounding: Victim gets 0 LP tokens!
9. Victim's tokens go to pool, attacker extracts them

Result: Attacker steals victim's entire deposit
```

## Required Fix

```go
const (
    MinimumLiquidity = 1000 // Locked forever, sent to zero address
)

func (k Keeper) AddLiquidity(ctx sdk.Context, pool *types.LiquidityPool, provider string, amountA, amountB sdk.Coin) (sdkmath.Int, sdkmath.LegacyDec, error) {
    // ...

    var lpTokens sdkmath.Int

    totalLpTokens, _ := k.parseLPTokens(pool.TotalLpTokens)

    if totalLpTokens.IsZero() {
        // FIRST DEPOSIT: Burn minimum liquidity to zero address
        lpTokens = sdkmath.NewIntFromBigInt(new(big.Int).Sqrt(
            amountA.Amount.Mul(amountB.Amount).BigInt(),
        ))

        // Subtract minimum liquidity (locked forever)
        if lpTokens.LTE(sdkmath.NewInt(MinimumLiquidity)) {
            return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(),
                fmt.Errorf("insufficient initial liquidity: need > %d", MinimumLiquidity)
        }

        // Burn minimum liquidity to zero address
        minimumLiquidity := sdkmath.NewInt(MinimumLiquidity)
        lpTokens = lpTokens.Sub(minimumLiquidity)

        // Track burned minimum liquidity
        pool.LockedLiquidity = minimumLiquidity.String()

        ctx.EventManager().EmitEvent(
            sdk.NewEvent(
                "minimum_liquidity_locked",
                sdk.NewAttribute("pool_id", pool.PoolId),
                sdk.NewAttribute("amount", minimumLiquidity.String()),
            ),
        )
    } else {
        // Subsequent deposits: Pro-rata calculation
        reserveA, _ := sdkmath.NewIntFromString(pool.ReserveA)
        reserveB, _ := sdkmath.NewIntFromString(pool.ReserveB)

        // Calculate LP tokens based on minimum ratio
        lpFromA := amountA.Amount.Mul(totalLpTokens).Quo(reserveA)
        lpFromB := amountB.Amount.Mul(totalLpTokens).Quo(reserveB)

        lpTokens = lpFromA
        if lpFromB.LT(lpTokens) {
            lpTokens = lpFromB
        }

        // Minimum LP token issuance to prevent dust attacks
        if lpTokens.IsZero() {
            return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(),
                fmt.Errorf("liquidity too small: would receive 0 LP tokens")
        }
    }

    // ...
}
```

## Acceptance Criteria

- [x] Minimum liquidity burn (1000) on first deposit
- [x] Minimum LP token check for subsequent deposits
- [x] Genesis export/import of locked liquidity
- [x] Tests for first depositor attack prevention
- [x] Tests for dust attack prevention
- [x] Invariant: locked_liquidity + sum(provider_lp) = total_lp

## Resolution

Fixed LP token inflation attack vulnerability through multiple layers of protection:

1. **Minimum Liquidity Burn**: Added `MinimumLiquidity = 1000` constant that is permanently burned on pool creation (lines 29, 117-141 in `liquidity_pool.go`)
2. **Zero LP Token Rejection**: `AddLiquidity` rejects deposits that would receive 0 LP tokens (lines 306-318)
3. **LP Token Invariant Validation**: Added `validateLPTokenInvariant()` function called after all LP-modifying operations (lines 997-1072)
4. **Locked Liquidity Tracking**: Added `LockedLiquidity` field to track permanently locked tokens (line 159)
5. **Event Emission**: Emits `minimum_liquidity_locked` event for audit trail (lines 193-199)

### Implementation Details

- **File**: `chain/x/dex/keeper/liquidity_pool.go`
- **Test File**: `chain/x/dex/keeper/lp_inflation_attack_test.go` (851 lines, 15 comprehensive tests)
- **Exported Function**: `ValidateLPTokenInvariantExported()` for test access (lines 1068-1072)

### Test Coverage

Created comprehensive test suite with 15 tests covering:
- First depositor attack prevention
- Dust attack prevention
- Donation attack invariant protection
- Minimum threshold edge cases
- LP burn with realistic pools
- Locked liquidity permanence
- Event emission verification
- Multiple provider scenarios
- Genesis export/import persistence
- Invariant validation after all operations

All tests pass: `go test -v ./x/dex/keeper/ -run "LPInflation"`
