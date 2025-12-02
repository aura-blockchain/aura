---
status: ready
priority: p2
issue_id: "009"
tags: [code-review, data-integrity, dex]
dependencies: []
---

# DEX Liquidity Provider Invariant Race Condition

## Problem Statement

The `LiquidityProviderConsistencyInvariant` iterates over all pools and sums provider LP tokens, but without atomic snapshots. Concurrent liquidity operations can cause false positives/negatives.

**Why it matters:** False positive invariant failures could trigger emergency governance responses. False negatives could miss actual economic exploits.

## Findings

### Evidence
- **File:** `chain/x/dex/keeper/invariants.go`
- **Lines:** 230-289

### Race Condition Example
```
Time    Thread A (Invariant Check)          Thread B (Add Liquidity)
T0      Read pool.TotalLpTokens = 1000
T1      Read provider[0].LpTokens = 300
T2      Read provider[1].LpTokens = 400
T3                                          AddLiquidity(200) -> Total=1200
T4      Read provider[2].LpTokens = 500     provider[2].LpTokens = 700
T5      Sum = 300+400+500 = 1200
T6      Compare: 1000 != 1200 -> BROKEN!    (False positive)
```

### Impact
- False positive invariant failures
- Unnecessary governance emergency responses
- Potential to miss real economic exploits (false negatives)

## Proposed Solutions

### Option A: Use Cache Context (Recommended)
**Pros:** SDK best practice, consistent snapshot
**Cons:** Small memory overhead
**Effort:** Small (1 hour)
**Risk:** Low

```go
func LiquidityProviderConsistencyInvariant(k *Keeper) sdk.Invariant {
    return func(ctx sdk.Context) (string, bool) {
        // Create cache context - reads will be consistent snapshot
        cacheCtx, _ := ctx.CacheContext()
        store := cacheCtx.KVStore(k.storeKey)
        // ... rest of invariant using cacheCtx
    }
}
```

## Recommended Action
Apply cache context pattern to all DEX invariants.

## Technical Details

### Affected Files
- `chain/x/dex/keeper/invariants.go:230-289`

### Acceptance Criteria
- [ ] All DEX invariants use cache context for consistent reads
- [ ] No false positives under concurrent load testing
- [ ] Performance impact measured and acceptable

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in data integrity review | Affects economic security |
