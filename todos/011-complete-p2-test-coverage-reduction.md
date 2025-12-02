---
status: ready
priority: p2
issue_id: "011"
tags: [code-review, testing, security]
dependencies: []
---

# Test Coverage Significantly Reduced in Refactoring

## Problem Statement

The refactoring deleted 2,000+ lines of comprehensive tests across multiple modules. Focused edge case tests were replaced with simplified "happy path" tests.

**Why it matters:** The deleted tests covered security edge cases, error paths, and attack vectors. Reduced coverage increases risk of regressions and security vulnerabilities.

## Findings

### Evidence
Major test deletions:
- `chain/x/confidencescore/keeper/slash_test.go`: -736 lines
- `chain/x/compliance/keeper/invariants_test.go`: -339 lines
- `chain/x/compliance/keeper/genesis_test.go`: -237 lines
- Multiple module-level test files deleted

### Before/After Example (slash_test.go)
**Before: 16 focused test functions**
```go
func TestSlashScore_Success(t *testing.T) { ... }
func TestSlashScore_InvalidInputs(t *testing.T) { ... }
func TestSlashScore_UserNotFound(t *testing.T) { ... }
func TestSlashScore_ExceedsTotalScore(t *testing.T) { ... }
func TestSlashScore_VerificationRevoked(t *testing.T) { ... }
func TestAppealSlash_Success(t *testing.T) { ... }
func TestAppealSlash_NotFound(t *testing.T) { ... }
// ... 9 more specific test cases
```

**After: 2 combined tests**
```go
func TestSlashScore(t *testing.T) { ... }       // Combined "happy path"
func TestAppealSlashFlow(t *testing.T) { ... }  // Combined flow
```

### Impact
- Edge cases no longer tested
- Error paths no longer verified
- Security attack vectors untested
- Regressions more likely to reach production

## Proposed Solutions

### Option A: Restore Deleted Tests (Recommended)
**Pros:** Comprehensive coverage restored
**Cons:** More test code to maintain
**Effort:** Medium (4-8 hours to review and restore)
**Risk:** Low

### Option B: Rewrite with New Test Helpers
**Pros:** May be cleaner
**Cons:** More effort, may miss cases
**Effort:** Large (1-2 days)
**Risk:** Medium

## Recommended Action
Restore the deleted test functions. The Prime Directive explicitly requires:
> - Unit tests: Every public function, every edge case
> - Test coverage: Happy path: 100%, Error paths: 100%, Edge cases: 100%

## Technical Details

### Affected Files
- `chain/x/confidencescore/keeper/slash_test.go`
- `chain/x/compliance/keeper/invariants_test.go`
- `chain/x/compliance/keeper/genesis_test.go`
- `chain/app/app_test.go`
- Multiple other test files

### Acceptance Criteria
- [ ] All deleted edge case tests restored
- [ ] Error path tests restored
- [ ] Coverage metrics back to previous levels
- [ ] Security-focused tests verified present

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in simplicity review | Tests should not be simplified away |
