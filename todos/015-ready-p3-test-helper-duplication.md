---
status: ready
priority: p3
issue_id: "015"
tags: [code-review, maintainability, testing]
dependencies: []
---

# Duplicated Test Helper Functions Across Modules

## Problem Statement

Nearly identical test setup functions are duplicated across 4+ modules (~160 lines total) despite global helpers existing in `chain/testing/testutil/`.

**Why it matters:** Maintenance burden - changes require updating multiple files. Inconsistencies can creep in.

## Findings

### Evidence
Duplicated files:
- `chain/x/confidencescore/keeper/test_helpers_test.go` (51 lines)
- `chain/x/inclusionroutines/keeper/test_helpers_test.go` (50 lines)
- `chain/x/compliance/keeper/test_suite.go` (42 lines)
- `chain/x/dex/keeper/keeper_test_suite.go` (49 lines)

### Impact
- ~160 lines of duplicate code
- Maintenance burden
- Inconsistencies between modules

## Proposed Solutions

### Option A: Use Global Test Helpers (Recommended)
**Effort:** Small (2-4 hours)

```go
// DELETE per-module test_helpers_test.go files
// USE existing global helpers:
import keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"

func TestFoo(t *testing.T) {
    input := keepertest.CreateTestInputWithKeys(t, "confidencescore")
    k := NewKeeper(input.StoreKey, input.Cdc, ...)
}
```

## Technical Details

### Affected Files
- `chain/x/*/keeper/test_helpers_test.go` (delete)
- `chain/x/*/keeper/test_suite.go` (refactor)
- `chain/testing/testutil/keeper/setup.go` (may need enhancement)

### Acceptance Criteria
- [ ] Per-module test helpers deleted
- [ ] All tests use global helpers
- [ ] No duplicated SDK configuration code

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in pattern analysis | DRY principle |
