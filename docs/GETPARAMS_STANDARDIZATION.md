# GetParams Signature Standardization

**Status:** In Progress (1/7 Medium Priority Modules Complete) | **Priority:** P2 | **Effort:** ~16 hours

**See:** `GETPARAMS_STANDARDIZATION_PROGRESS.md` for detailed progress report

## Problem

18 different GetParams signatures exist across 26 modules, causing inconsistency and maintainability issues.

## Recommended Standard

```go
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
    store := k.storeService.OpenKVStore(ctx)
    bz, err := store.Get(types.ParamsKey)
    if err != nil {
        return types.Params{}, err
    }
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

**Rationale:**
- Value receiver (Keeper doesn't modify state)
- `context.Context` (standard Go, SDK v0.50+ compatible)
- Error return (no panics on unmarshal failure)
- Non-pointer return (simpler semantics)

## Current Patterns

| Pattern | Modules | Count | Issues |
|---------|---------|-------|--------|
| `(k Keeper) GetParams(ctx context.Context) (Params, error)` | aiassistant, cryptography, monitoring, walletsecurity, networksecurity, economics | 6 | ✅ Standard |
| `(k *Keeper) GetParams() Params` (no context) | confidencescore, dataregistry, economicsecurity, identitychange, inclusionroutines, vcregistry | 6 | Cannot access store |
| `(k *Keeper) GetParams(ctx sdk.Context) (*Params, error)` | identity, auth, prevalidation | 3 | Pointer receiver, legacy context |
| `(k Keeper) GetParams(ctx sdk.Context) Params` (no error) | contractregistry, security, wasm, bridge | 4 | Panics on failure |
| `(k Keeper) GetParams(ctx context.Context) Params` (no error) | privacy, validatorsecurity | 2 | No error handling |
| Other variations | governance, dex, incidentresponse | 5 | Mixed patterns |

## Migration Priority

### High Priority (No Context Parameter)
- `confidencescore/keeper/keeper.go:70`
- `dataregistry/keeper/keeper.go:91`
- `economicsecurity/keeper/keeper.go:52`
- `identitychange/keeper/keeper.go:54`
- `inclusionroutines/keeper/keeper.go:74`
- `vcregistry/keeper/keeper.go:114`

### Medium Priority (No Error Return)
- ✅ `contractregistry/keeper/keeper.go` - COMPLETED 2025-12-24
- `governance/keeper/keeper.go:73` - 14 call sites
- `security/keeper/keeper.go:114` - 7 call sites
- `privacy/keeper/keeper.go` - 16 call sites
- `validatorsecurity/keeper/keeper.go` - 13 call sites
- `wasm/keeper/keeper.go` - TBD call sites
- `bridge/keeper/keeper.go` - 17 call sites (special paramstore pattern)

### Low Priority (Legacy sdk.Context)
- identity, auth, prevalidation (functional, just need context type update)

## Migration Steps

1. Update function signature
2. Add error handling in implementation
3. Update all call sites to handle error
4. Run tests
5. Verify no panics

## Modules Already Compliant

aiassistant, cryptography, monitoring, walletsecurity, networksecurity, economics (6 modules)
