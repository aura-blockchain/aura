---
status: pending
priority: p3
issue_id: "117"
tags: [code-review, quality, duplication, maintainability]
dependencies: ["100"]
---

# P3 MEDIUM: 44% Code Duplication - Potential for Significant Reduction

## Problem Statement

Code simplicity analysis identified ~44% potential LOC reduction through eliminating duplication and extracting shared utilities.

**Why it matters:** Duplicated code means bugs must be fixed in multiple places, increasing maintenance burden and risk of inconsistency.

## Findings

### Common Duplication Patterns

**1. Address Validation (Found in 15+ modules)**
```go
// Repeated in every module
func validateAddress(addr string) error {
    _, err := sdk.AccAddressFromBech32(addr)
    if err != nil {
        return errorsmod.Wrap(types.ErrInvalidAddress, addr)
    }
    return nil
}
```

**2. Pagination Handling (Found in 10+ modules)**
```go
// Copy-pasted pagination logic
pageReq := req.Pagination
if pageReq == nil {
    pageReq = &query.PageRequest{Limit: 100}
}
if pageReq.Limit > 1000 {
    pageReq.Limit = 1000
}
```

**3. Event Emission (Found in 20+ locations)**
```go
// Similar event patterns repeated
ctx.EventManager().EmitEvent(
    sdk.NewEvent(
        types.EventTypeXXX,
        sdk.NewAttribute(types.AttributeKeyXXX, value),
    ),
)
```

**4. Store Operations (Found in all modules)**
```go
// Same KV store patterns
store := ctx.KVStore(k.storeKey)
key := types.GetXXXKey(id)
bz := store.Get(key)
if bz == nil {
    return nil, types.ErrNotFound
}
var result types.XXX
k.cdc.MustUnmarshal(bz, &result)
return &result, nil
```

### Duplication Statistics

| Pattern | Instances | LOC Duplicated |
|---------|-----------|----------------|
| Address validation | 15+ | ~150 lines |
| Pagination setup | 10+ | ~100 lines |
| Event emission | 20+ | ~200 lines |
| Store CRUD | 27+ | ~500 lines |
| Error wrapping | 50+ | ~250 lines |

**Total estimated duplicate LOC: ~1,200**

## Proposed Solutions

### Solution A: Create Shared Utility Package (Recommended)
**Effort:** 3-5 days | **Risk:** Low

**1. Create shared package:**
```go
// chain/pkg/common/validation.go
package common

func ValidateAddress(addr string) error {
    _, err := sdk.AccAddressFromBech32(addr)
    if err != nil {
        return errorsmod.Wrapf(ErrInvalidAddress, "address: %s", addr)
    }
    return nil
}

func ValidateAddresses(addrs ...string) error {
    for _, addr := range addrs {
        if err := ValidateAddress(addr); err != nil {
            return err
        }
    }
    return nil
}
```

**2. Pagination helper:**
```go
// chain/pkg/common/pagination.go
func NormalizePagination(req *query.PageRequest, defaultLimit, maxLimit uint64) *query.PageRequest {
    if req == nil {
        return &query.PageRequest{Limit: defaultLimit}
    }
    if req.Limit == 0 {
        req.Limit = defaultLimit
    }
    if req.Limit > maxLimit {
        req.Limit = maxLimit
    }
    return req
}
```

**3. Generic store operations:**
```go
// chain/pkg/common/store.go
func GetObject[T any](store sdk.KVStore, cdc codec.BinaryCodec, key []byte) (*T, error) {
    bz := store.Get(key)
    if bz == nil {
        return nil, ErrNotFound
    }
    var result T
    if err := cdc.Unmarshal(bz, &result); err != nil {
        return nil, errorsmod.Wrap(ErrUnmarshal, err.Error())
    }
    return &result, nil
}

func SetObject[T any](store sdk.KVStore, cdc codec.BinaryCodec, key []byte, obj *T) error {
    bz, err := cdc.Marshal(obj)
    if err != nil {
        return errorsmod.Wrap(ErrMarshal, err.Error())
    }
    store.Set(key, bz)
    return nil
}
```

## Recommended Action

**GO WITH SOLUTION A**: Create shared utility package and migrate modules incrementally.

## Technical Details

### New Package Structure

```
chain/pkg/
├── common/
│   ├── validation.go
│   ├── pagination.go
│   ├── store.go
│   ├── events.go
│   └── errors.go
└── testutil/
    ├── keeper.go
    └── mocks.go
```

### Migration Strategy

1. Create shared package with common utilities
2. Add deprecation comments to duplicated code
3. Migrate one module at a time
4. Remove deprecated code after all migrations
5. Update imports

## Acceptance Criteria

- [ ] Shared validation utilities extracted
- [ ] Shared pagination helpers extracted
- [ ] Generic store operations available
- [ ] At least 5 modules migrated to use shared code
- [ ] ~500+ LOC reduction achieved
- [ ] All tests pass after migration

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Code simplicity analysis identified duplication | P3 Medium |

## Resources

- [DRY Principle](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself)
- [Go Generics](https://go.dev/doc/tutorial/generics)
