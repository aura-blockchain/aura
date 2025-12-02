---
status: ready
priority: p2
issue_id: "014"
tags: [code-review, architecture, maintainability]
dependencies: []
---

# Store Keys Defined in Three Separate Locations

## Problem Statement

Store keys are defined in three different locations that must be manually synchronized:
1. `StoreKeyNames()` function (line 303)
2. `app.storeKeys` struct fields (line 1057)
3. `base.MountKVStores()` call (line 412)

**Why it matters:** Adding a module requires editing 3 locations. Missing from one = runtime panic or initialization failure.

## Findings

### Evidence
- **File:** `chain/app/app.go`
- **Lines:** 303-340 (StoreKeyNames), 412-475 (MountKVStores), 1057-1128 (storeKeys struct)

### Impact
- High maintenance burden
- Easy to introduce bugs when adding modules
- No compile-time validation that all keys are synchronized

## Proposed Solutions

### Option A: Centralize Store Key Definitions (Recommended)
**Pros:** Single source of truth
**Cons:** Refactoring effort
**Effort:** Medium (2-4 hours)
**Risk:** Low

```go
type storeKeys struct {
    account *storetypes.KVStoreKey
    bank    *storetypes.KVStoreKey
    // ... all keys
}

func (s *storeKeys) Names() []string {
    return []string{authtypes.StoreKey, banktypes.StoreKey, ...}
}

func (s *storeKeys) AsMap() map[string]*storetypes.KVStoreKey {
    return map[string]*storetypes.KVStoreKey{
        authtypes.StoreKey: s.account,
        banktypes.StoreKey: s.bank,
        ...
    }
}
```

## Technical Details

### Affected Files
- `chain/app/app.go:303-340, 412-475, 1057-1128`

### Acceptance Criteria
- [ ] Single source of truth for store keys
- [ ] Methods to generate Names() and AsMap() from single definition
- [ ] Build-time validation that all keys are mounted

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in architecture review | DRY principle violation |
