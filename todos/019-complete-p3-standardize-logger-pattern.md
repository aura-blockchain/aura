# TODO: Standardize Logger signature across all modules

---
status: pending
priority: p3
issue_id: "019"
tags: [code-review, patterns, consistency]
dependencies: []
---

## Problem Statement

Logger method signature varies across modules - some take `sdk.Context`, some take `context.Context`, and some take no parameters.

**Impact:** Inconsistent patterns make code harder to maintain.

## Findings

**Pattern A: Takes SDK context (12 modules)**
```go
func (k Keeper) Logger(ctx sdk.Context) log.Logger
```

**Pattern B: Takes context.Context (3 modules)**
```go
func (k Keeper) Logger(ctx context.Context) log.Logger
```

**Pattern C: No context parameter (1 module - identity)**
```go
func (k *Keeper) Logger() log.Logger
```

**Pattern D: Returns interface{} (dataregistry)**
```go
func (k *Keeper) Logger(ctx sdk.Context) interface{}
```

## Proposed Solutions

### Option 1: Standardize to sdk.Context pattern (Recommended)
**Pros:** Cosmos SDK standard, consistent
**Cons:** Need to update 4 modules
**Effort:** Small (2-3 hours)
**Risk:** Low

```go
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
    return ctx.Logger().With("module", types.ModuleName)
}
```

## Acceptance Criteria

- [ ] All 27 modules use same Logger signature
- [ ] Returns log.Logger, not interface{}
- [ ] Takes sdk.Context parameter

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Pattern Recognition Specialist agent |
