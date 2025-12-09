---
status: pending
priority: p1
issue_id: "101"
tags: [code-review, security, dex, oracle, critical]
dependencies: ["100"]
---

# P1 CRITICAL: DEX Price Oracle Manipulation Vulnerability

## Problem Statement

The DEX module's `GetAuraPrice()` function falls back to spot price when TWAP data is insufficient, creating a **critical oracle manipulation vulnerability**.

**Why it matters:** An attacker can manipulate prices to bypass minimum liquidity requirements, steal funds, or cause economic exploits. This is a common DeFi attack vector (see Mango Markets, Cream Finance exploits).

## Findings

### Vulnerable Code

**File:** `/home/decri/blockchain-projects/aura/chain/x/dex/keeper/keeper.go` (lines 97-140)

```go
func (k Keeper) GetAuraPrice(ctx sdk.Context, poolID string) sdkmath.LegacyDec {
    params := k.GetParams(ctx)

    // Try TWAP first (good)
    twapPrice, err := k.GetTWAPPrice(ctx, poolID, params.TwapWindowBlocks)
    if err == nil && !twapPrice.IsZero() {
        return twapPrice
    }

    // VULNERABILITY: Falls back to manipulable spot price
    spotPrice := k.calculateSpotPrice(ctx, poolID)
    return spotPrice
}
```

### Attack Vector

```
1. Attacker creates AURA-USDT pool with minimal liquidity
2. No TWAP data exists (new pool)
3. GetAuraPrice() falls back to manipulable spot price
4. Attacker inflates AURA price to $1000 via large USDT deposit
5. CreatePool() calculates minimum liquidity based on fake $1000 price
6. Result: Minimum liquidity requirement bypassed
```

### Security Impact
- **CVSS Score:** 8.5 (High)
- **Attack Complexity:** Low
- **Funds at Risk:** All protocol liquidity

## Proposed Solutions

### Solution A: Governance Fallback Price (Recommended)
**Effort:** 1 day | **Risk:** Low

```go
const MinTWAPObservations = 100

func (k Keeper) GetAuraPrice(ctx sdk.Context, poolID string) sdkmath.LegacyDec {
    params := k.GetParams(ctx)

    twapPrice, observationCount, err := k.GetTWAPPriceWithCount(ctx, poolID, params.TwapWindowBlocks)
    if err == nil && !twapPrice.IsZero() && observationCount >= MinTWAPObservations {
        return twapPrice
    }

    // NEVER use spot price - use governance-set fallback
    fallbackPrice := k.GetGovernanceFallbackPrice(ctx)
    if fallbackPrice.IsZero() {
        return sdkmath.LegacyNewDecWithPrec(10, 2) // $0.10 conservative default
    }
    return fallbackPrice
}
```

**Pros:**
- Eliminates manipulation attack entirely
- Governance can update price as needed

**Cons:**
- Requires governance proposal for price updates

### Solution B: Multi-Pool TWAP
**Effort:** 3-5 days | **Risk:** Medium

Average TWAP across multiple pools to reduce manipulation surface.

### Solution C: External Oracle Integration
**Effort:** 1-2 weeks | **Risk:** High

Integrate Chainlink or Band Protocol oracles.

## Recommended Action

**GO WITH SOLUTION A**: Implement governance fallback price. Simple, secure, no external dependencies.

## Technical Details

### Affected Files
- `chain/x/dex/keeper/keeper.go`
- `chain/x/dex/types/params.go` (add GovernanceFallbackPrice param)

### Database/State Changes
- New param: `governance_fallback_price` in DEX params

## Acceptance Criteria

- [ ] GetAuraPrice() never uses spot price from manipulable pools
- [ ] TWAP requires minimum 100 observations before use
- [ ] Fallback price configurable via governance
- [ ] Unit tests for manipulation resistance
- [ ] Integration test simulating attack scenario

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Security audit identified vulnerability | P1 Critical |

## Resources

- [Mango Markets Exploit Analysis](https://rekt.news/mango-markets-rekt/)
- [TWAP Oracle Best Practices](https://docs.uniswap.org/concepts/protocol/oracle)
