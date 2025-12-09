# TODO: Remove deprecated functions (no backward compatibility needed)

---
status: pending
priority: p3
issue_id: "015"
tags: [code-review, cleanup, simplicity]
dependencies: []
---

## Problem Statement

20+ functions marked "deprecated - kept for backward compatibility" but there's no mainnet yet, so no users to maintain compatibility for.

**Impact:** ~400 LOC of dead weight, confusion, potential misuse.

## Findings

**Deprecated Functions:**
```go
// x/economics/keeper/governance.go
// SetVetoRequest stores a veto request (deprecated - kept for backward compatibility)
// GetVetoRequest retrieves a veto request (deprecated - kept for backward compatibility)
// SetSnapshotVote stores a snapshot vote (deprecated - now part of Vote)
// GetSnapshotVote retrieves a snapshot vote (deprecated - now part of Vote)
// SetVoteCommitment stores a vote commitment (deprecated - now part of Vote)
// GetVoteCommitment retrieves a vote commitment (deprecated - now part of Vote)
// SetTokenLock stores a token lock (deprecated - use VoteLock)
// GetTokenLock retrieves a token lock (deprecated - use VoteLock)
```

**Also in:**
- `x/walletsecurity/keeper/`
- `x/privacy/keeper/key_rotation.go` (panic-throwing deprecated functions)
- `x/compliance/keeper/transaction_monitor.go:311`

## Proposed Solutions

### Option 1: Delete all deprecated functions (Recommended)
**Pros:** Clean API surface, less code
**Cons:** None (no users)
**Effort:** Small (1-2 hours)
**Risk:** Zero (pre-mainnet)

## Acceptance Criteria

- [ ] All deprecated functions deleted
- [ ] No "deprecated" comments remain in production code
- [ ] Tests updated if any used deprecated functions
- [ ] ~400 LOC reduction

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Code Simplicity Reviewer agent |
