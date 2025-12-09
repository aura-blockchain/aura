# TODO: Delete backup files from repository

---
status: pending
priority: p3
issue_id: "014"
tags: [code-review, cleanup, simplicity]
dependencies: []
---

## Problem Statement

22 `.bak` and `.old` files exist in the repository. Git already provides version history.

**Impact:** Clutter, confusion, wasted repository space (~2000 lines of duplicate code).

## Findings

**Files to Delete:**
```
./x/wasm/keeper/keeper.go.bak
./x/bridge/types/params.go.bak
./x/bridge/keeper/unmarshal_error_handling_test.go.bak
./x/bridge/keeper/keeper_complete_test.go.bak
./x/vcregistry/keeper/keeper.go.bak2
./x/security/types/genesis.go.bak
./x/security/types/types.go.bak
# ... 15 more
```

## Proposed Solutions

### Option 1: Delete all backup files (Recommended)
**Pros:** Instant cleanup
**Cons:** None (git has history)
**Effort:** Small (5 minutes)
**Risk:** Zero

```bash
find ./chain/x -name "*.bak*" -delete
find ./chain/x -name "*.old" -delete

# Add to .gitignore
echo "*.bak*" >> .gitignore
echo "*.old" >> .gitignore
```

## Acceptance Criteria

- [ ] All .bak and .old files deleted
- [ ] .gitignore updated to prevent future additions
- [ ] ~2000 LOC reduction

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Code Simplicity Reviewer agent |
