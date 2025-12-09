# TODO: Fix invariant registration in app.go

---
status: pending
priority: p1
issue_id: "004"
tags: [code-review, data-integrity, invariants, critical]
dependencies: []
---

## Problem Statement

The `registerInvariants()` function in app.go only logs that invariants are registered but does not actually call the invariant registration functions. All invariants are effectively disabled.

**Impact:** State corruption will go undetected. No automatic validation of critical state properties.

## Findings

**Location:** `/home/decri/blockchain-projects/aura/chain/app/app.go` (lines 1348-1417)

**Current Implementation (BROKEN):**
```go
func (a *App) registerInvariants() {
    // Logs but doesn't register:
    a.Logger().Info("registered bank invariants", "checks", "total-supply")
    a.Logger().Info("registered staking invariants", ...)
    a.Logger().Info("registered dex invariants", ...)
    // etc.
}
```

**Expected Implementation:**
```go
func (a *App) registerInvariants() {
    crisisKeeper := a.CrisisKeeper

    // Actually register invariants
    dexkeeper.RegisterInvariants(crisisKeeper, a.dexKeeper)
    economicsecuritykeeper.RegisterInvariants(crisisKeeper, a.economicsecurityKeeper)
    networksecuritykeeper.RegisterInvariants(crisisKeeper, a.networksecurityKeeper)
    bridgekeeper.RegisterInvariants(crisisKeeper, a.bridgeKeeper)
    // ... etc
}
```

**Modules with Invariants (should be registered):**
- DEX: 6 invariants (pool reserves, LP consistency, HTLC validity)
- EconomicSecurity: 7 invariants (params, vesting, treasury)
- NetworkSecurity: 5 invariants (peer reputation, rate limits)
- Bridge: Multiple invariants (supply, signatures)
- Compliance: AML/KYC invariants

## Proposed Solutions

### Option 1: Register with Crisis Keeper (Recommended)
**Pros:** Standard Cosmos SDK pattern, enables crisis module halt
**Cons:** Requires crisis keeper setup
**Effort:** Medium (2-3 hours)
**Risk:** Low

### Option 2: Custom invariant runner without crisis module
**Pros:** Simpler, doesn't need crisis keeper
**Cons:** Non-standard, no automatic halt on violation
**Effort:** Medium (2-3 hours)
**Risk:** Medium

## Recommended Action

Option 1 - Properly register invariants with the crisis keeper following Cosmos SDK patterns.

## Technical Details

**Files to Modify:**
- `chain/app/app.go` - Fix registerInvariants()

**Modules with RegisterInvariants functions:**
- `chain/x/dex/keeper/invariants.go`
- `chain/x/economicsecurity/keeper/invariants.go`
- `chain/x/networksecurity/keeper/invariants.go`
- `chain/x/bridge/keeper/invariants.go`

## Acceptance Criteria

- [ ] All module invariants registered with crisis keeper
- [ ] `aurad invariant-check` CLI command works
- [ ] Invariant violations are detected and logged
- [ ] Chain halts on critical invariant violation (configurable)

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Data Integrity Guardian agent review |

## Resources

- Cosmos SDK Crisis Module: https://docs.cosmos.network/main/build/modules/crisis
- DEX invariants: `chain/x/dex/keeper/invariants.go`
