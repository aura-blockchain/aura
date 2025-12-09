# TODO: Remove unnecessary security abstraction layer

---
status: pending
priority: p3
issue_id: "016"
tags: [code-review, cleanup, simplicity, over-engineering]
dependencies: []
---

## Problem Statement

The `x/common/security/` package provides custom SafeMath, ReentrancyGuard, etc. but Cosmos SDK already provides these protections. Only 2/27 modules use it.

**Impact:** ~500 LOC of redundant code creating false sense of security.

## Findings

**Unnecessary Abstractions:**
- `SafeMath` - Cosmos SDK `math.Int` already has overflow protection
- `ReentrancyGuard` - Context handles deterministic execution
- `PauseGuard` - Should be module params, not separate abstraction

**Evidence - SafeMath is redundant:**
```go
// x/common/security/math.go
func (sm *SafeMath) SafeAddDec(a, b math.LegacyDec) (math.LegacyDec, error) {
    result := a.Add(b)
    // LegacyDec has built-in overflow protection
    // The operation itself will panic on overflow, so if we get here it's safe
    return result, nil  // ← Literally just wraps the native operation!
}
```

**Usage Stats:**
- 164 direct math operations (`.Add`, `.Sub`, `.Mul`) in bridge keeper
- Only 35 SafeMath usages across entire codebase
- **Developers bypass SafeMath 82% of the time** because it's unnecessary

## Proposed Solutions

### Option 1: Delete entire x/common/security package (Recommended)
**Pros:** Removes false security theater, cleaner code
**Cons:** Need to update 2 modules (bridge, dex)
**Effort:** Medium (1 day)
**Risk:** Low (tests catch issues)

### Option 2: Keep but don't expand
**Pros:** No changes needed
**Cons:** Perpetuates over-engineering
**Effort:** None
**Risk:** None

## Acceptance Criteria

- [ ] x/common/security/ package deleted
- [ ] Bridge and DEX modules use SDK primitives directly
- [ ] All tests pass
- [ ] ~500 LOC reduction

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Code Simplicity Reviewer agent |
