# TODO: Implement missing genesis handlers for 4 modules

---
status: pending
priority: p1
issue_id: "002"
tags: [code-review, data-integrity, genesis, critical]
dependencies: []
---

## Problem Statement

Four modules lack genesis.go files in their keeper directory, which means they cannot properly export/import state during chain upgrades or migrations.

**Impact:** Data loss during any genesis export/import operation. Critical for chain upgrades.

## Findings

**Missing Files:**
1. `/home/decri/blockchain-projects/aura/chain/x/dataregistry/keeper/genesis.go`
2. `/home/decri/blockchain-projects/aura/chain/x/economics/keeper/genesis.go`
3. `/home/decri/blockchain-projects/aura/chain/x/identity/keeper/genesis.go`
4. `/home/decri/blockchain-projects/aura/chain/x/security/keeper/genesis.go`

**Risk Level:** CRITICAL - Chain halt during upgrade if these modules have non-empty state.

**Evidence:** Module registration exists in `app.go` (lines 933-967) but no corresponding `InitGenesis`/`ExportGenesis` implementations.

## Proposed Solutions

### Option 1: Implement complete genesis handlers (Recommended)
**Pros:** Full state persistence, production-ready
**Cons:** Requires understanding each module's state
**Effort:** Large (4-6 hours per module)
**Risk:** Low if tested properly

Template:
```go
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
    if err := types.ValidateGenesis(&data); err != nil {
        return err
    }
    if err := k.SetParams(ctx, data.Params); err != nil {
        return err
    }
    for _, entity := range data.Entities {
        k.SetEntity(ctx, entity)
    }
    return nil
}

func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
    return types.GenesisState{
        Params:   k.GetParams(ctx),
        Entities: k.GetAllEntities(ctx),
    }
}
```

## Recommended Action

Implement genesis handlers for all 4 modules before testnet deployment.

## Technical Details

**Affected Modules:**
- dataregistry - Data storage and IPFS integration
- economics - Economic parameters and fee structures
- identity - DID and identity records (CRITICAL for users)
- security - Security configurations

**Reference Implementation:**
- Good example: `chain/x/bridge/keeper/genesis.go` (126 lines, comprehensive)
- Good example: `chain/x/dex/keeper/genesis.go` (complete state handling)

## Acceptance Criteria

- [ ] All 4 modules have genesis.go files
- [ ] `InitGenesis` validates input and imports all state
- [ ] `ExportGenesis` exports all state correctly
- [ ] Round-trip test: export → import → export matches
- [ ] Integration test passes with complex state

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Data Integrity Guardian agent review |

## Resources

- Bridge genesis (reference): `chain/x/bridge/keeper/genesis.go`
- DEX genesis (reference): `chain/x/dex/keeper/genesis.go`
