---
status: pending
priority: p1
issue_id: "105"
tags: [code-review, architecture, integration, testnet-blocker]
dependencies: ["100"]
---

# P1 CRITICAL: Keeper Dependency Wiring Incomplete - Cross-Module Features Broken

## Problem Statement

Several keeper dependencies are commented out due to "interface mismatches", meaning **cross-module features don't work** (VC issuance, confidence scoring, contract registry).

**Why it matters:** Core blockchain functionality depends on modules working together. Without proper wiring, the identity/VC system, confidence scoring, and smart contract policies are non-functional.

## Findings

### Evidence

**File:** `/home/decri/blockchain-projects/aura/chain/app/app.go` (lines 787-790)

```go
// Note: Keeper dependencies not wired due to interface mismatches - fix in follow-up
// contractRegistryKeeper.SetVCRegistryKeeper(vcKeeper)
// contractRegistryKeeper.SetComplianceKeeper(complianceKeeper)
// contractRegistryKeeper.SetConfidenceScoreKeeper(csKeeper)
```

**File:** `/home/decri/blockchain-projects/aura/chain/app/depinject.go` (lines 197, 220, 250)

```go
// TODO: Wire dependency after interface alignment (ConfidenceScoreKeeper)
// TODO: Wire dependency after interface alignment (VCRegistryKeeper)
// TODO: Wire dependencies after interface alignment
```

### Specific Interface Mismatches

1. **ConfidenceScoreKeeper.SetIRRegistry** - Missing `GetIRArena` method
2. **VCRegistryKeeper.SetConfidenceScoreKeeper** - `sdk.Context` signature mismatch

### Impact
- Contract registry cannot enforce VC-based policies
- Confidence scoring cannot query inclusion routines
- VC issuance cannot update confidence scores
- Cross-module authorization broken

## Proposed Solutions

### Solution A: Create Adapter Interfaces (Recommended)
**Effort:** 2-3 days | **Risk:** Low

Create adapter wrappers that bridge the interface mismatches:

```go
// adapter.go
type ConfidenceScoreKeeperAdapter struct {
    inner *confidencescore.Keeper
}

func (a *ConfidenceScoreKeeperAdapter) GetUserScore(ctx context.Context, addr string) (uint64, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    return a.inner.GetUserScore(sdkCtx, addr)
}
```

### Solution B: Refactor Keeper Interfaces
**Effort:** 1 week | **Risk:** Medium

Standardize all keeper interfaces to use `context.Context` consistently.

## Recommended Action

**GO WITH SOLUTION A**: Adapters are faster and don't require refactoring all modules.

## Technical Details

### Affected Files
- `chain/app/app.go`
- `chain/app/depinject.go`
- `chain/app/keeper_adapters.go` (exists, extend it)

### Database/State Changes
None - this is wiring.

## Acceptance Criteria

- [ ] All commented-out SetXXXKeeper calls uncommented and working
- [ ] Contract registry can query VC status
- [ ] Confidence score updates on VC issuance
- [ ] Integration tests for cross-module flows pass

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Architecture review identified gap | P1 Critical |

## Resources

- [Cosmos SDK Keeper Pattern](https://docs.cosmos.network/main/building-modules/keeper)
