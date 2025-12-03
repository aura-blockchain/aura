---
id: "041"
title: "DEX LP Token Inflation Attack (First Depositor)"
status: ready
priority: p1
category: security
module: dex
severity: CRITICAL
cvss: 9.5
source: dex-security-audit
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

- [ ] Minimum liquidity burn (1000) on first deposit
- [ ] Minimum LP token check for subsequent deposits
- [ ] Genesis export/import of locked liquidity
- [ ] Tests for first depositor attack prevention
- [ ] Tests for dust attack prevention
- [ ] Invariant: locked_liquidity + sum(provider_lp) = total_lp
