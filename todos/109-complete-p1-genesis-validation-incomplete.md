---
status: pending
priority: p1
issue_id: "109"
tags: [code-review, data-integrity, genesis, validation, critical]
dependencies: ["100"]
---

# P1 CRITICAL: Genesis Validation Incomplete - Data Corruption Risk

## Problem Statement

Multiple modules have incomplete or missing genesis validation, allowing invalid state to be imported that could corrupt the blockchain or enable exploits.

**Why it matters:** Genesis is the foundation of all state. Invalid genesis data propagates corruption throughout the chain's lifetime.

## Findings

### Missing Validations

**1. DEX Module - No Pool Invariant Check**

```go
// File: chain/x/dex/types/genesis.go
func (gs GenesisState) Validate() error {
    // MISSING: Validate k = x * y invariant for pools
    // MISSING: Validate pool reserve amounts > 0
    // MISSING: Validate no duplicate pool IDs
    return nil
}
```

**2. Bridge Module - No Validator Set Validation**

```go
// File: chain/x/bridge/types/genesis.go
func (gs GenesisState) Validate() error {
    // MISSING: Validate validator addresses are valid
    // MISSING: Validate threshold <= validator count
    // MISSING: Validate no duplicate validators
    return nil
}
```

**3. Identity Module - No Credential Expiry Check**

```go
// File: chain/x/identity/types/genesis.go
func (gs GenesisState) Validate() error {
    // MISSING: Validate credential expiry > creation time
    // MISSING: Validate issuer exists for each credential
    // MISSING: Validate no duplicate DIDs
    return nil
}
```

### Impact

| Module | Missing Validation | Impact |
|--------|-------------------|--------|
| dex | Pool invariants | Economic exploit via invalid reserves |
| bridge | Validator set | Unauthorized withdrawals |
| identity | Credential expiry | Expired credentials accepted |
| vcregistry | VC schema | Invalid VCs accepted |
| compliance | Rule integrity | Bypassed compliance checks |

## Proposed Solutions

### Solution A: Comprehensive Genesis Validation (Recommended)
**Effort:** 2-3 days | **Risk:** Low

Implement thorough validation for each module:

```go
// Example: DEX module
func (gs GenesisState) Validate() error {
    // Validate params
    if err := gs.Params.Validate(); err != nil {
        return errorsmod.Wrap(ErrInvalidGenesis, "invalid params")
    }

    // Validate pools
    seenPoolIDs := make(map[uint64]bool)
    for i, pool := range gs.Pools {
        // Check for duplicates
        if seenPoolIDs[pool.PoolID] {
            return errorsmod.Wrapf(ErrInvalidGenesis, "duplicate pool ID %d at index %d", pool.PoolID, i)
        }
        seenPoolIDs[pool.PoolID] = true

        // Validate reserves
        if pool.ReserveA.IsZero() || pool.ReserveB.IsZero() {
            return errorsmod.Wrapf(ErrInvalidGenesis, "pool %d has zero reserves", pool.PoolID)
        }

        // Validate invariant (k = x * y)
        k := pool.ReserveA.Mul(pool.ReserveB)
        if !k.IsPositive() {
            return errorsmod.Wrapf(ErrInvalidGenesis, "pool %d has invalid invariant", pool.PoolID)
        }
    }

    return nil
}
```

## Recommended Action

**GO WITH SOLUTION A**: Implement comprehensive genesis validation for all modules.

## Technical Details

### Affected Files

- `chain/x/dex/types/genesis.go`
- `chain/x/bridge/types/genesis.go`
- `chain/x/identity/types/genesis.go`
- `chain/x/vcregistry/types/genesis.go`
- `chain/x/compliance/types/genesis.go`
- All other module genesis files

### Database/State Changes

None - validation only.

## Acceptance Criteria

- [ ] All modules validate param integrity
- [ ] All modules check for duplicate IDs
- [ ] DEX validates pool invariants and reserve amounts
- [ ] Bridge validates validator set and threshold
- [ ] Identity validates credential expiry and issuer existence
- [ ] VCRegistry validates schema references
- [ ] Compliance validates rule integrity
- [ ] Unit tests for each validation rule
- [ ] Integration test: invalid genesis rejected

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Data integrity review identified gap | P1 Critical |

## Resources

- [Cosmos SDK Genesis](https://docs.cosmos.network/main/building-modules/genesis)
- [Genesis Best Practices](https://docs.cosmos.network/main/building-modules/genesis#validategenesis)
