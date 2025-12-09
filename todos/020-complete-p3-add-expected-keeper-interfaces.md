# TODO: Add expected_keepers.go to 21 modules

---
status: pending
priority: p3
issue_id: "020"
tags: [code-review, patterns, architecture]
dependencies: []
---

## Problem Statement

Only 6/27 modules define explicit `expected_keepers.go` interfaces. The remaining 21 rely on direct keeper imports.

**Impact:** Harder to test with mocks, tighter coupling.

## Findings

**Modules WITH expected_keepers.go (6):**
- governance
- dex
- bridge
- contractregistry
- security
- privacy
- vcregistry

**Modules WITHOUT expected_keepers.go (21):**
- identity, compliance, economics, economicsecurity, networksecurity
- walletsecurity, validatorsecurity, incidentresponse, prevalidation
- cryptography, monitoring, dataregistry, inclusionroutines
- identitychange, confidencescore, aura-bindings, wasm, auth, common, internal

## Proposed Solutions

### Option 1: Add interfaces to all modules with dependencies (Recommended)
**Pros:** Better testability, decoupling
**Cons:** Time investment
**Effort:** Medium (2-3 days)
**Risk:** Low

```go
// x/{module}/types/expected_keepers.go
package types

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
    SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error
    GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

type AccountKeeper interface {
    GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
}
```

## Acceptance Criteria

- [ ] All modules with keeper dependencies have expected_keepers.go
- [ ] Interfaces define only methods actually used
- [ ] Keepers accept interfaces in constructor, not concrete types

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Pattern Recognition Specialist agent |
