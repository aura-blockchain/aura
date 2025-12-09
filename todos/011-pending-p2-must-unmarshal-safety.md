# TODO: Replace MustUnmarshal with safe error handling

---
status: pending
priority: p2
issue_id: "011"
tags: [code-review, security, reliability]
dependencies: []
---

## Problem Statement

The codebase uses `MustUnmarshal` extensively (321 occurrences), which panics on unmarshaling errors. Corrupted state data causes chain halt.

**Impact:** Chain halt on state corruption or proto schema changes during upgrades.

## Findings

**Usage Count:** 321 occurrences of `MustUnmarshal` across 53+ keeper implementations.

**Example:**
```go
// cryptography/keeper/keeper.go:111
var key cryptoproto.QuantumResistantKey
k.cdc.MustUnmarshal(bz, &key)  // Panics if bz is corrupted
```

**Risk Scenarios:**
1. State corruption from hardware failure
2. Proto schema changes during upgrade
3. Invalid data written by bug

## Proposed Solutions

### Option 1: Replace with safe unmarshal + defaults (Recommended)
**Pros:** Graceful degradation, no chain halt
**Cons:** May mask underlying issues
**Effort:** Large (2-3 weeks to audit all 321 usages)
**Risk:** Medium

```go
var key cryptoproto.QuantumResistantKey
if err := k.cdc.Unmarshal(bz, &key); err != nil {
    k.Logger(ctx).Error("failed to unmarshal quantum key, using defaults",
        "error", err, "key_id", keyID)
    return types.DefaultQuantumKey(keyID), nil
}
```

### Option 2: Keep MustUnmarshal but add recovery (Middle ground)
**Pros:** Maintains fail-fast, but handles gracefully
**Cons:** More complex
**Effort:** Medium (1 week)
**Risk:** Low

```go
func (k Keeper) safeGetKey(ctx sdk.Context, keyID string) (key types.Key, err error) {
    defer func() {
        if r := recover(); r != nil {
            k.Logger(ctx).Error("panic in key unmarshal", "keyID", keyID, "panic", r)
            err = ErrCorruptedState
        }
    }()

    bz := store.Get(keyID)
    k.cdc.MustUnmarshal(bz, &key)
    return key, nil
}
```

## Technical Details

**High-Priority Files (critical state):**
- `chain/x/bridge/keeper/*.go` - Cross-chain asset state
- `chain/x/dex/keeper/*.go` - AMM pool state
- `chain/x/identity/keeper/*.go` - DID records
- `chain/x/economics/keeper/*.go` - Economic parameters

## Acceptance Criteria

- [ ] All 321 MustUnmarshal usages audited
- [ ] Critical paths use safe unmarshal with fallback
- [ ] Logging for all unmarshal failures
- [ ] No chain halt on single corrupted record

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Security Sentinel agent review |
