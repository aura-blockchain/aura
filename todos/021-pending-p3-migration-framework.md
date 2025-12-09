# TODO: Create migration framework for all modules

---
status: pending
priority: p3
issue_id: "021"
tags: [code-review, data-integrity, upgrades]
dependencies: []
---

## Problem Statement

Only 1 migration file exists in entire codebase. No migration infrastructure for other modules.

**Impact:** Difficult to perform breaking changes safely. Risk of data loss during upgrades.

## Findings

**Single Migration Found:**
`/home/decri/blockchain-projects/aura/chain/x/privacy/migrations/v1_remove_private_keys.go`

This migration is well-implemented with:
- Defensive error handling
- Event emission for audit trail
- Verification function
- Security-focused (removing sensitive data)

**Missing:** All other modules have no migration infrastructure.

## Proposed Solutions

### Option 1: Create migration framework template (Recommended)
**Pros:** Ready for future upgrades
**Cons:** Time investment upfront
**Effort:** Medium (1 week)
**Risk:** Low

**Structure:**
```
chain/x/{module}/migrations/
├── v2_description.go          // Migration logic
├── v2_description_test.go     // Migration tests
└── migrate.go                  // Version coordinator
```

**Template:**
```go
// migrations/migrate.go
package migrations

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type Migrator struct {
    keeper Keeper
}

func NewMigrator(k Keeper) Migrator {
    return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2(ctx sdk.Context) error {
    return v2_migration.Run(ctx, m.keeper)
}
```

## Acceptance Criteria

- [ ] Migration template created
- [ ] All modules have migrations/ directory structure
- [ ] Module.go registers migrations with configurator
- [ ] Documentation for writing migrations

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Data Integrity Guardian agent review |
