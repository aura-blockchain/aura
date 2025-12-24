# GetParams Standardization Progress Report

**Date:** 2025-12-24
**Task:** Standardize GetParams signatures for MEDIUM PRIORITY modules
**Status:** Partial Completion (1 of 7 modules completed)

## Completed Modules

### ✅ contractregistry (COMPLETED)

**Changes Made:**
- Updated signature: `func (k Keeper) GetParams(ctx context.Context) (types.Params, error)`
- Added error handling for unmarshal failures
- Updated 3 production call sites
- Updated 2 test call sites
- All tests passing (0.088s)

**Files Modified:**
1. `x/contractregistry/keeper/keeper.go` - GetParams signature and implementation
2. `x/contractregistry/keeper/invariants.go` - ParamsInvariant error handling
3. `x/contractregistry/keeper/msg_server.go` - RegisterContract error handling
4. `x/contractregistry/keeper/keeper_test.go` - Test updated
5. `x/contractregistry/keeper/msg_server_comprehensive_test.go` - Test updated

**Test Results:**
```
PASS
ok      github.com/aequitas/aura/chain/x/contractregistry/keeper    0.088s
```

## Pending Modules

### ⏳ security (7 call sites - MEDIUM complexity)

**Current:** `func (k Keeper) GetParams(ctx sdk.Context) Params`
**Target:** `func (k Keeper) GetParams(ctx context.Context) (Params, error)`

**Call Sites:**
- genesis_test.go (2)
- privacy.go
- network.go
- invariants.go
- genesis.go
- query_server.go

**Estimated Effort:** 1 hour

### ⏳ privacy (16 call sites - MEDIUM complexity)

**Current:** `func (k Keeper) GetParams(ctx context.Context) Params`
**Target:** `func (k Keeper) GetParams(ctx context.Context) (Params, error)`

**Note:** Already uses `context.Context`, only needs error return added.

**Call Sites:**
- keeper.go
- confidential_transactions.go
- mixing_protocol.go
- ring_signatures.go
- msg_server.go
- query_server.go
- invariants.go
- Plus 9 test files

**Estimated Effort:** 2 hours

### ⏳ validatorsecurity (13 call sites - MEDIUM complexity)

**Current:** `func (k Keeper) GetParams(ctx context.Context) *ValidatorSecurityParams`
**Target:** `func (k Keeper) GetParams(ctx context.Context) (ValidatorSecurityParams, error)`

**Note:** Already uses `context.Context`, needs to change from pointer return to value return.

**Call Sites:**
- sentry.go, jailing.go, monitoring.go, slashing.go
- genesis.go, query_server.go, keeper.go
- abci.go
- Plus 5 test files

**Estimated Effort:** 2 hours

### ⏳ governance (14 call sites - HIGH complexity)

**Current:** `func (k *Keeper) GetParams(ctx sdk.Context) *GovernanceParams`
**Target:** `func (k Keeper) GetParams(ctx context.Context) (GovernanceParams, error)`

**Challenges:**
- Pointer receiver → value receiver (method set change)
- Pointer return → value return
- sdk.Context → context.Context
- Complex error handling in invariants and genesis

**Call Sites:**
- vote_privacy.go
- query_server.go
- invariants.go
- proposal_execution.go
- vote_delegation.go
- proposal_lifecycle.go (2 calls)
- deposit_refund.go
- param_validation.go
- msg_server.go
- genesis.go
- Plus 3 test files

**Estimated Effort:** 4 hours

### ⏳ wasm (call sites TBD - LOW complexity)

**Status:** Needs analysis - standard GetParams method not found in initial scan.

**Estimated Effort:** 1 hour

### ⏳ bridge (17 call sites - SPECIAL CASE)

**Current:** Uses paramstore subspace pattern
```go
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
    if k.paramstore.HasKeyTable() {
        k.paramstore.GetParamSet(ctx, &params)
    }
    return params
}
```

**Target:**
```go
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    var params types.Params
    if k.paramstore.HasKeyTable() {
        k.paramstore.GetParamSet(sdkCtx, &params)
        return params, nil
    }
    return types.DefaultParams(), nil
}
```

**Estimated Effort:** 3 hours

## Standard Signature Template

```go
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    store := sdkCtx.KVStore(k.storeKey)
    bz := store.Get(types.ParamsKey)
    if bz == nil {
        return types.DefaultParams(), nil
    }
    var params types.Params
    if err := k.cdc.Unmarshal(bz, &params); err != nil {
        return types.Params{}, fmt.Errorf("failed to unmarshal params: %w", err)
    }
    return params, nil
}
```

## Call Site Update Patterns

### Production Code

**Functions returning error:**
```go
params, err := k.GetParams(ctx)
if err != nil {
    return err // or fmt.Errorf("failed to get params: %w", err)
}
```

**Invariants:**
```go
params, err := k.GetParams(ctx)
if err != nil {
    return sdk.FormatInvariant(
        moduleName,
        "params-valid",
        fmt.Sprintf("failed to get params: %v", err),
    ), true
}
```

**Genesis Export:**
```go
params, err := k.GetParams(ctx)
if err != nil {
    ctx.Logger().Error("failed to get params, using defaults", "error", err)
    params = types.DefaultParams()
}
```

**Query Servers:**
```go
params, err := qs.Keeper.GetParams(goCtx)
if err != nil {
    return nil, err
}
return &types.QueryParamsResponse{Params: params}, nil
```

### Test Code

```go
params, err := keeper.GetParams(ctx)
require.NoError(t, err)
require.NotNil(t, params)
```

## Progress Summary

| Module | Call Sites | Status | Time Spent | Tests |
|--------|------------|--------|------------|-------|
| contractregistry | 5 (3 prod + 2 test) | ✅ Complete | ~45 min | ✅ Pass |
| security | 7 | ⏳ Pending | - | - |
| privacy | 16 | ⏳ Pending | - | - |
| validatorsecurity | 13 | ⏳ Pending | - | - |
| governance | 14 | ⏳ Pending | - | - |
| wasm | TBD | ⏳ Pending | - | - |
| bridge | 17 | ⏳ Pending | - | - |
| **TOTAL** | **~73** | **14% Complete** | **~45 min** | - |

## Lessons Learned

### What Went Well
1. **Incremental approach** - Testing after each file change caught issues early
2. **Template pattern** - Standard error handling patterns work across modules
3. **Test coverage** - Existing tests validated changes immediately

### Challenges Encountered
1. **Variable shadowing** - `err := k.GetParams()` when `err` already declared required changing `:=` to `=`
2. **Mixed return types** - Invariants return `(string, bool)`, genesis returns `GenesisState`, requiring different error handling
3. **Pointer semantics** - Some modules use pointer returns, others use values

### Recommendations for Remaining Work

1. **Start with simpler modules:**
   - security (7 call sites, straightforward)
   - privacy (16 but already uses context.Context)
   - validatorsecurity (13, similar to privacy)

2. **Leave complex for last:**
   - governance (14 call sites, pointer receiver change)
   - bridge (special paramstore pattern)

3. **Batch test updates:**
   - Grep for all `GetParams` calls in tests
   - Update in single pass to avoid compilation cycles

4. **Create helper script:**
   - Automate common patterns
   - Reduce manual editing errors

## Next Steps

1. **Immediate:** Complete `security` module (estimated 1 hour)
2. **Short-term:** Complete `privacy` and `validatorsecurity` (estimated 4 hours)
3. **Medium-term:** Tackle `governance` module (estimated 4 hours)
4. **Final:** Handle `wasm` and `bridge` special cases (estimated 4 hours)

**Total Remaining Effort:** ~13-14 hours

## References

- Original analysis: `/home/hudson/blockchain-projects/aura/docs/GETPARAMS_STANDARDIZATION.md`
- Detailed module analysis: `/home/hudson/blockchain-projects/aura/docs/GETPARAMS_MEDIUM_PRIORITY_ANALYSIS.md`
- Standard specification: Line 29-33 of GETPARAMS_STANDARDIZATION.md
