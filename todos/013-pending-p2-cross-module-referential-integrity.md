# TODO: Add cross-module referential integrity invariants

---
status: pending
priority: p2
issue_id: "013"
tags: [code-review, data-integrity, invariants]
dependencies: ["004", "008"]
---

## Problem Statement

No invariants check cross-module references. Orphaned references can exist after deletions or corruption.

**Impact:** Ghost identities, incorrect compliance checks, state bloat from orphaned data.

## Findings

**Risk Scenarios:**
1. Identity DID deleted → VCRegistry still holds VCs referencing it
2. Pool deleted → Orders still reference non-existent pool
3. Chain config deleted → Bridge transfers reference invalid chain

**Current State:**
- No explicit foreign key constraint checking
- Cross-module references validated only at message handling time
- No invariants detect orphaned references

## Proposed Solutions

### Option 1: Add cross-module invariants (Recommended)
**Pros:** Catches corruption early
**Cons:** Adds keeper dependencies
**Effort:** Medium (1-2 days)
**Risk:** Low

```go
// vcregistry/keeper/invariants.go
func VCSubjectIntegrityInvariant(vcK VCKeeper, idK IdentityKeeper) sdk.Invariant {
    return func(ctx sdk.Context) (string, bool) {
        for _, vc := range vcK.GetAllVCs(ctx) {
            if !idK.HasDID(ctx, vc.Subject) {
                return sdk.FormatInvariant("vcregistry", "orphaned-vc",
                    fmt.Sprintf("VC %s references non-existent DID %s",
                        vc.ID, vc.Subject)), true
            }
        }
        return "", false
    }
}

// dex/keeper/invariants.go
func OrderPoolIntegrityInvariant(k Keeper) sdk.Invariant {
    return func(ctx sdk.Context) (string, bool) {
        orders := k.GetAllOrders(ctx)
        for _, order := range orders {
            if !k.PoolExists(ctx, order.PoolID) {
                return sdk.FormatInvariant("dex", "orphaned-order",
                    fmt.Sprintf("Order %s references non-existent pool %d",
                        order.ID, order.PoolID)), true
            }
        }
        return "", false
    }
}
```

## Acceptance Criteria

- [ ] VCRegistry → Identity invariant
- [ ] DEX Orders → Pools invariant
- [ ] Bridge Transfers → Chain configs invariant
- [ ] All invariants registered
- [ ] Tests verify detection of orphaned references

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Data Integrity Guardian agent review |
