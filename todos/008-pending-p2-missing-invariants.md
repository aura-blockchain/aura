# TODO: Add missing invariants for 6 modules

---
status: pending
priority: p2
issue_id: "008"
tags: [code-review, data-integrity, invariants]
dependencies: ["004"]
---

## Problem Statement

Six modules lack `invariants.go` files, meaning state corruption in these modules will go undetected.

**Impact:** Silent data corruption, difficult debugging after incidents.

## Findings

**Modules Missing Invariants:**
1. **economics** - Economic parameters and fee configurations unchecked
2. **identity** - Core identity records without consistency validation
3. **incidentresponse** - Incident records could become inconsistent
4. **prevalidation** - Validation state lacks integrity checks
5. **security** - Security configurations unchecked
6. **walletsecurity** - Wallet security state unvalidated

**Comparison - Modules WITH Good Invariants:**
- DEX: 6 invariants (pool reserves, order validity, LP consistency)
- EconomicSecurity: 7 invariants (params, vesting, treasury, MEV)
- NetworkSecurity: 5 invariants (peer reputation, rate limits, mempool)

## Proposed Solutions

### Option 1: Implement invariants for all 6 modules (Recommended)
**Pros:** Complete coverage, early corruption detection
**Cons:** Time investment
**Effort:** Large (2-3 hours per module)
**Risk:** Low

**Template:**
```go
// RegisterInvariants registers all module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
    ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
    ir.RegisterRoute(types.ModuleName, "state-consistent", StateConsistencyInvariant(k))
}

func ParamsInvariant(k Keeper) sdk.Invariant {
    return func(ctx sdk.Context) (string, bool) {
        params := k.GetParams(ctx)
        if err := params.Validate(); err != nil {
            return sdk.FormatInvariant(types.ModuleName, "params-valid",
                fmt.Sprintf("invalid params: %v", err)), true
        }
        return "", false
    }
}
```

## Technical Details

**Files to Create:**
- `chain/x/economics/keeper/invariants.go`
- `chain/x/identity/keeper/invariants.go`
- `chain/x/incidentresponse/keeper/invariants.go`
- `chain/x/prevalidation/keeper/invariants.go`
- `chain/x/security/keeper/invariants.go`
- `chain/x/walletsecurity/keeper/invariants.go`

## Acceptance Criteria

- [ ] All 6 modules have invariants.go
- [ ] Each module has at least params validation invariant
- [ ] Identity module checks DID structure validity
- [ ] Economics module validates fee configurations
- [ ] Tests verify invariants detect corruption

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Data Integrity Guardian agent review |
