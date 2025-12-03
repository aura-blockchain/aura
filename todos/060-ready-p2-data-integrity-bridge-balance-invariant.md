---
id: "060"
title: "Bridge Module Balance Invariant Incomplete"
status: ready
priority: p2
category: data-integrity
module: bridge
severity: HIGH
source: data-integrity-review
---

# Bridge Transfer Balance Invariant Incomplete

## Problem

The balance invariant comment says it should check module balance, but the actual check is skipped. Transfers can be created without locking funds.

## Affected Files

- `chain/x/bridge/keeper/invariants.go:45-94`

## Current Code

```go
// Note: We skip the module balance check since we don't have GetBalance method
return "", false  // INVARIANT NOT ACTUALLY CHECKED
```

## Impact

- No validation that funds are actually locked
- Bridge can be drained
- Module can become insolvent

## Required Fix

Actually check that module balance >= sum of locked transfers.

```go
for denom, totalLocked := range lockedAmounts {
    moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
    moduleBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, denom)

    if moduleBalance.Amount.LT(totalLocked) {
        return sdk.FormatInvariant(...), true
    }
}
```

## Acceptance Criteria

- [ ] Module balance check implemented
- [ ] Bank keeper injected for balance queries
- [ ] Invariant actually validates state
- [ ] Tests for invariant violations
