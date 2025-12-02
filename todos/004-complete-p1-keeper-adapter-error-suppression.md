---
status: ready
priority: p1
issue_id: "004"
tags: [code-review, architecture, data-integrity]
dependencies: []
---

# Keeper Adapter Error Suppression Causes State Inconsistency

## Problem Statement

The keeper adapter layer in `chain/app/keeper_adapters.go` suppresses errors from critical operations like `Jail()`, `Unjail()`, and `Slash()`, only logging them instead of returning them to callers.

**Why it matters:** When a security operation fails silently, the calling module believes the operation succeeded. This can lead to state inconsistencies where ConfidenceScore thinks a validator is slashed but staking module disagrees.

## Findings

### Evidence
- **File:** `chain/app/keeper_adapters.go`
- **Lines:** 76-78, 326-332, 334-344

```go
// Line 76-78: securityStakingAdapter.Jail
func (a securityStakingAdapter) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) {
    if err := a.inner.Jail(sdk.WrapSDKContext(ctx), consAddr); err != nil {
        ctx.Logger().Error("security staking jail failed", "error", err)
        // ERROR NOT RETURNED - caller thinks jail succeeded!
    }
}

// Line 334-344: securityStakingKeeperAdapter.Slash
func (a securityStakingKeeperAdapter) Slash(...) string {
    factor, err := sdkmath.LegacyNewDecFromStr(slashFactor)
    if err != nil {
        return "0"  // Parse failure returns "0" instead of error
    }
    slashed, err := a.inner.Slash(...)
    if err != nil {
        return "0"  // Slash failure returns "0" instead of error
    }
    return slashed.String()
}
```

### Impact
- State inconsistencies between modules
- Security module believes validator is jailed, but staking module may disagree
- Slashing operations may silently fail, leaving malicious validators active
- Cannot distinguish "slashed 0 tokens" from "slash operation failed"

## Proposed Solutions

### Option A: Refactor to Return Errors (Recommended)
**Pros:** Proper error propagation, state consistency
**Cons:** Requires updating all caller sites
**Effort:** Medium (2-4 hours)
**Risk:** Low

```go
func (a securityStakingAdapter) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) error {
    if err := a.inner.Jail(sdk.WrapSDKContext(ctx), consAddr); err != nil {
        return fmt.Errorf("staking jail failed for %s: %w", consAddr, err)
    }
    return nil
}

// For Slash, use a result type
type SlashResult struct {
    TokensSlashed sdkmath.Int
    Error         error
}

func (a securityStakingKeeperAdapter) Slash(...) SlashResult {
    factor, err := sdkmath.LegacyNewDecFromStr(slashFactor)
    if err != nil {
        return SlashResult{TokensSlashed: sdkmath.ZeroInt(), Error: err}
    }
    slashed, err := a.inner.Slash(...)
    return SlashResult{TokensSlashed: slashed, Error: err}
}
```

### Option B: Add Panic on Critical Failures
**Pros:** Prevents inconsistent state
**Cons:** May halt chain on edge cases
**Effort:** Small (30 min)
**Risk:** Medium

## Recommended Action
Implement Option A. Update all modules that use these adapters to handle the returned errors.

## Technical Details

### Affected Files
- `chain/app/keeper_adapters.go:76-78` (Jail)
- `chain/app/keeper_adapters.go:326-332` (Unjail)
- `chain/app/keeper_adapters.go:334-344` (Slash)
- All modules that call these adapter methods

### Acceptance Criteria
- [ ] All adapter methods return errors instead of suppressing
- [ ] Callers updated to handle errors appropriately
- [ ] Unit tests verify error propagation
- [ ] Integration tests verify state consistency on failures

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in architecture review | Critical for state consistency |

## Resources
- Related: P1-005 (Genesis validation), P2-001 (Module adapter complexity)
