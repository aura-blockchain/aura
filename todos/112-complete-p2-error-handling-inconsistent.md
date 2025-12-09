---
status: pending
priority: p2
issue_id: "112"
tags: [code-review, quality, error-handling, consistency]
dependencies: ["100"]
---

# P2 HIGH: Inconsistent Error Handling Across Modules

## Problem Statement

Error handling patterns vary wildly across modules: some use `errorsmod.Wrap`, some use `fmt.Errorf`, some return raw errors, and some swallow errors silently.

**Why it matters:** Inconsistent errors make debugging difficult, break error type assertions, and confuse users with mixed error formats.

## Findings

### Pattern Variations Found

**1. Modern SDK Pattern (Good)**
```go
// identity module
return errorsmod.Wrapf(types.ErrInvalidDID, "DID %s not found", did)
```

**2. Legacy Pattern (Acceptable)**
```go
// vcregistry module
return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid VC: %s", err)
```

**3. Plain Go Errors (Bad)**
```go
// bridge module
return fmt.Errorf("signature verification failed: %w", err)
```

**4. Silent Error Swallowing (Critical)**
```go
// dex module
if err != nil {
    ctx.Logger().Error("failed to update pool", "error", err)
    return // Silently continues
}
```

### Error Code Collisions

| Module | Error Code 1 | Conflict |
|--------|--------------|----------|
| identity | ErrInvalidDID | Code 1 |
| vcregistry | ErrInvalidVC | Code 1 |
| bridge | ErrInvalidSignature | Code 1 |

All using the same error code (1) makes debugging impossible.

### Missing Error Types

Many modules lack specific error types:

```go
// Currently:
return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "something went wrong")

// Should be:
var ErrPoolNotFound = errorsmod.Register(ModuleName, 2, "pool not found")
return errorsmod.Wrapf(ErrPoolNotFound, "pool %d", poolID)
```

## Proposed Solutions

### Solution A: Standardize Error Handling (Recommended)
**Effort:** 2-3 days | **Risk:** Low

1. **Define unique error codes per module**
2. **Use errorsmod.Wrapf consistently**
3. **Include context in all errors**
4. **Never swallow errors**

**Error code allocation:**

```go
// Each module gets a range
// identity: 100-199
// vcregistry: 200-299
// bridge: 300-399
// dex: 400-499
// compliance: 500-599
// etc.

// identity/types/errors.go
var (
    ErrInvalidDID      = errorsmod.Register(ModuleName, 100, "invalid DID")
    ErrDIDNotFound     = errorsmod.Register(ModuleName, 101, "DID not found")
    ErrDIDExists       = errorsmod.Register(ModuleName, 102, "DID already exists")
    ErrUnauthorized    = errorsmod.Register(ModuleName, 103, "unauthorized")
)
```

**Standard pattern:**

```go
func (k Keeper) DoSomething(ctx sdk.Context, id string) error {
    // Always wrap with context
    thing, err := k.GetThing(ctx, id)
    if err != nil {
        return errorsmod.Wrapf(err, "failed to get thing %s", id)
    }

    // Use specific error types
    if thing == nil {
        return errorsmod.Wrapf(ErrThingNotFound, "id: %s", id)
    }

    // Never swallow errors - propagate or handle explicitly
    if err := k.ProcessThing(ctx, thing); err != nil {
        return errorsmod.Wrap(err, "failed to process")
    }

    return nil
}
```

## Recommended Action

**GO WITH SOLUTION A**: Standardize error handling across all modules.

## Technical Details

### Affected Files

All `errors.go` and keeper files across all modules.

### Database/State Changes

None - code changes only.

## Acceptance Criteria

- [ ] Each module has unique error code range
- [ ] All errors use errorsmod.Wrapf pattern
- [ ] All errors include contextual information
- [ ] No error swallowing (logging without returning)
- [ ] Error codes documented in module README
- [ ] Tests verify specific error types

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Code quality review identified inconsistency | P2 High |

## Resources

- [Cosmos SDK Error Handling](https://docs.cosmos.network/main/building-modules/errors)
- [Error Best Practices](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
