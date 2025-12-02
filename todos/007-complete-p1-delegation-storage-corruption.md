---
status: ready
priority: p1
issue_id: "007"
tags: [code-review, data-integrity, critical-bug]
dependencies: []
---

# Delegation Storage Uses Broken Marshaling (Data Corruption)

## Problem Statement

The `storeDelegation` and `getDelegation` functions use a broken custom marshaling scheme that only stores 2 string fields, losing all other data. The retrieval function returns a stub object!

**Why it matters:** This is a **CRITICAL BUG** that causes total data loss for score delegations. Any delegation data stored is immediately corrupted.

## Findings

### Evidence
- **File:** `chain/x/confidencescore/keeper/score_delegation.go`
- **Lines:** 354-365, 367-378

```go
func (k *Keeper) storeDelegation(ctx sdk.Context, delegation *ScoreDelegation) {
    store := k.storeService.OpenKVStore(ctx)
    key := types.DelegationStoreKey(delegation.DelegationID)

    // Marshal delegation (simplified - in production use protobuf)
    bz := make([]byte, 0)
    bz = append(bz, []byte(delegation.Delegator)...)
    bz = append(bz, 0) // delimiter
    bz = append(bz, []byte(delegation.Delegate)...)  // ONLY stores 2 fields!

    store.Set([]byte(key), bz)
}

func (k *Keeper) getDelegation(ctx sdk.Context, delegationID string) (*ScoreDelegation, bool) {
    // ...
    // Unmarshal delegation (simplified)
    // In production, use proper protobuf unmarshaling
    return &ScoreDelegation{DelegationID: delegationID}, true  // Returns STUB!
}
```

### Lost Fields
The ScoreDelegation struct likely contains:
- `DelegationID` - Partially preserved
- `Delegator` - Stored
- `Delegate` - Stored
- `Amount` - **LOST**
- `StartHeight` - **LOST**
- `EndHeight` - **LOST**
- `Status` - **LOST**
- `RewardsClaimed` - **LOST**

### Impact
- **CRITICAL**: All delegation data except 2 fields is lost
- Delegations cannot function - retrieved data is incomplete
- Any existing delegations in state are corrupted
- Feature is completely broken

## Proposed Solutions

### Option A: Use Protobuf Marshaling (Required)
**Pros:** Correct implementation, preserves all data
**Cons:** None - this is the standard approach
**Effort:** Small (1-2 hours)
**Risk:** Low

```go
func (k *Keeper) storeDelegation(ctx sdk.Context, delegation *ScoreDelegation) error {
    store := k.storeService.OpenKVStore(ctx)
    key := types.DelegationStoreKey(delegation.DelegationID)

    // Use protobuf (delegation should be a proto message type)
    bz, err := k.cdc.Marshal(delegation)
    if err != nil {
        return fmt.Errorf("failed to marshal delegation: %w", err)
    }

    return store.Set([]byte(key), bz)
}

func (k *Keeper) getDelegation(ctx sdk.Context, delegationID string) (*ScoreDelegation, bool) {
    store := k.storeService.OpenKVStore(ctx)
    key := types.DelegationStoreKey(delegationID)

    bz, err := store.Get([]byte(key))
    if err != nil || bz == nil {
        return nil, false
    }

    var delegation ScoreDelegation
    if err := k.cdc.Unmarshal(bz, &delegation); err != nil {
        return nil, false
    }

    return &delegation, true
}
```

## Recommended Action
**IMMEDIATE FIX REQUIRED**: This bug makes the delegation feature completely non-functional. Fix before any testing.

## Technical Details

### Affected Files
- `chain/x/confidencescore/keeper/score_delegation.go:354-378`

### Acceptance Criteria
- [ ] storeDelegation uses proper protobuf marshaling
- [ ] getDelegation properly unmarshals all fields
- [ ] Unit test verifies round-trip serialization
- [ ] Migration plan for any existing corrupted data

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in performance review | Critical bug - feature is broken |

## Resources
- [Cosmos SDK Store Basics](https://docs.cosmos.network/main/build/building-modules/store)
