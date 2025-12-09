# TODO: Standardize error handling with errorsmod.Register

---
status: pending
priority: p3
issue_id: "018"
tags: [code-review, patterns, error-handling]
dependencies: []
---

## Problem Statement

All modules use `errors.New()` instead of `errorsmod.Register()`. This prevents unique error codes for debugging.

**Impact:** Difficult debugging, no error code registry for clients.

## Findings

**Current Pattern (inconsistent):**
```go
// x/bridge/types/errors.go
var (
    ErrInvalidParam = errors.New("invalid parameter")
    ErrDuplicateAttestation = errors.New("duplicate attestation")
)
```

**Cosmos SDK Best Practice:**
```go
var (
    ErrInvalidParam = errorsmod.Register(ModuleName, 1, "invalid parameter")
    ErrDuplicateAttestation = errorsmod.Register(ModuleName, 2, "duplicate attestation")
)
```

**Benefits of errorsmod.Register:**
- Unique error codes across chain
- Machine-parseable error responses
- Easier debugging and monitoring
- Client SDKs can handle specific errors

## Proposed Solutions

### Option 1: Migrate all errors to errorsmod.Register (Recommended)
**Pros:** Standard pattern, better debugging
**Cons:** Requires updating all error definitions
**Effort:** Large (1-2 weeks)
**Risk:** Low

### Option 2: Keep as-is for testnet
**Pros:** No changes needed
**Cons:** Tech debt continues
**Effort:** None
**Risk:** Low

## Acceptance Criteria

- [ ] All modules use errorsmod.Register
- [ ] Error codes documented
- [ ] No duplicate error codes across modules
- [ ] Client SDKs updated if needed

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Pattern Recognition Specialist agent |
