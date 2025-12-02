---
status: ready
priority: p1
issue_id: "008"
tags: [code-review, performance, scalability]
dependencies: []
---

# Unbounded Storage Iterations Will Halt Chain at Scale

## Problem Statement

Multiple functions load entire collections from storage into memory without pagination. At scale (10K+ records), these operations will cause chain halts due to timeout or out-of-memory errors.

**Why it matters:** The chain will become unusable once user base grows. These functions are called in EndBlocker (every block) and query handlers.

## Findings

### Evidence

#### Location 1: `chain/x/confidencescore/keeper/score_delegation.go:380-397`
```go
func (k *Keeper) getAllDelegations(ctx sdk.Context) []ScoreDelegation {
    store := k.storeService.OpenKVStore(ctx)
    prefix := []byte(types.DelegationStoreKeyPrefix)
    iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
    // ... loads ALL delegations into memory
    delegations := []ScoreDelegation{}
    for ; iterator.Valid(); iterator.Next() {
        delegations = append(delegations, ScoreDelegation{})
    }
    return delegations
}
```

**Used by:** `ProcessExpiredDelegations()` (EndBlocker), `GetUserDelegations()`, `DistributeDelegationRewards()`

#### Location 2: `chain/x/inclusionroutines/keeper/ir_crud.go:60-70, 82-93`
```go
func (k *Keeper) ListIRs(ctx sdk.Context, statusFilter, arenaFilter, localeFilter string, offset, limit int) ([]types.IRDefinition, int) {
    allIRs := k.GetAllIRs(ctx)  // Loads ALL IRs even with pagination params!

    matching := make([]types.IRDefinition, 0)
    for _, ir := range allIRs {
        // Apply filters...
    }
    // Apply pagination AFTER loading everything
    return matching[offset:end], total
}
```

### Projected Impact at Scale

| Delegations | getAllDelegations Time | Chain Impact |
|-------------|------------------------|--------------|
| 1,000 | ~50ms | Acceptable |
| 10,000 | ~500ms | Noticeable lag |
| 100,000 | ~5 seconds | Block timeout warnings |
| 1,000,000 | ~50 seconds | **Chain halts** |

### Impact
- Chain becomes unusable with >10K delegations
- Query endpoints timeout
- EndBlocker exceeds block time
- Users cannot use the chain

## Proposed Solutions

### Option A: Add Pagination to Storage Layer (Recommended)
**Pros:** Scalable to millions of records
**Cons:** Requires API changes
**Effort:** Medium (4-8 hours)
**Risk:** Low

```go
// Paginated delegation retrieval
func (k *Keeper) getDelegationsPaginated(ctx sdk.Context, offset, limit int) []ScoreDelegation {
    store := k.storeService.OpenKVStore(ctx)
    prefix := []byte(types.DelegationStoreKeyPrefix)
    iterator, _ := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
    defer iterator.Close()

    delegations := []ScoreDelegation{}
    count := 0
    for ; iterator.Valid() && count < offset+limit; iterator.Next() {
        if count >= offset {
            var d ScoreDelegation
            k.cdc.Unmarshal(iterator.Value(), &d)
            delegations = append(delegations, d)
        }
        count++
    }
    return delegations
}
```

### Option B: Add Expiration Index for EndBlocker
**Pros:** EndBlocker becomes O(m) where m = expiring delegations
**Cons:** Requires new index storage
**Effort:** Medium (3-4 hours)
**Risk:** Low

```go
// Store delegation with expiration index
func (k *Keeper) storeDelegation(ctx sdk.Context, delegation *ScoreDelegation) error {
    // Store main delegation...

    // Add to expiration index if has end height
    if delegation.EndHeight > 0 {
        expirationKey := types.ExpirationIndexKey(delegation.EndHeight, delegation.DelegationID)
        store.Set(expirationKey, []byte(delegation.DelegationID))
    }
    return nil
}

// Process only expiring delegations
func (k *Keeper) ProcessExpiredDelegations(ctx sdk.Context) (int, error) {
    currentHeight := uint64(ctx.BlockHeight())
    prefix := types.ExpirationIndexPrefix(currentHeight)
    iterator, _ := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
    // Only iterates delegations expiring THIS block
}
```

## Recommended Action
Implement both Option A (for queries) and Option B (for EndBlocker) before any scale testing.

## Technical Details

### Affected Files
- `chain/x/confidencescore/keeper/score_delegation.go:380-397`
- `chain/x/inclusionroutines/keeper/ir_crud.go:60-70, 82-93, 186-196`
- `chain/x/inclusionroutines/keeper/prerequisites.go:82-93, 157-162`
- `chain/x/confidencescore/keeper/queries.go:37-68`

### Acceptance Criteria
- [ ] getAllDelegations replaced with paginated version
- [ ] ListIRs filters at storage layer, not memory
- [ ] ProcessExpiredDelegations uses expiration index
- [ ] Load tests pass with 100K+ records
- [ ] EndBlocker completes in <100ms with 100K delegations

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in performance review | Critical for scalability |

## Resources
- [Cosmos SDK Pagination](https://docs.cosmos.network/main/build/packages/pagination)
