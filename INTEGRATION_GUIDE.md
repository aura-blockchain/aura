# Shared Common Package Integration Guide

## Quick Start

The `chain/pkg/common` package provides production-quality utilities to reduce code duplication across modules.

### 1. Address Validation

**Before:**
```go
if msg.Sender == "" {
    return nil, status.Error(codes.InvalidArgument, "sender cannot be empty")
}
addr, err := sdk.AccAddressFromBech32(msg.Sender)
if err != nil {
    return nil, status.Errorf(codes.InvalidArgument, "invalid sender address: %s", err.Error())
}
```

**After:**
```go
import "github.com/aequitas/aura/chain/pkg/common"

addr, err := common.ValidateAddress(msg.Sender)
if err != nil {
    return nil, status.Error(codes.InvalidArgument, err.Error())
}
```

### 2. Pagination

**Before:**
```go
pagination := req.Pagination
if pagination == nil {
    pagination = &query.PageRequest{Limit: 100}
}
if pagination.Limit == 0 {
    pagination.Limit = 100
}
```

**After:**
```go
import "github.com/aequitas/aura/chain/pkg/common"

pagination := common.NormalizePagination(req.Pagination)
// Now use pagination with query.Paginate
```

### 3. Event Emission

**Before:**
```go
ctx.EventManager().EmitEvent(
    sdk.NewEvent(
        "dex.swap",
        sdk.NewAttribute("sender", msg.Sender),
        sdk.NewAttribute("pool_id", msg.PoolId),
        sdk.NewAttribute("amount_in", msg.CoinIn.String()),
    ),
)
```

**After:**
```go
import "github.com/aequitas/aura/chain/pkg/common"

common.EmitSuccessEvent(ctx, "dex", "swap", msg.Sender, map[string]string{
    "pool_id": msg.PoolId,
    "amount_in": msg.CoinIn.String(),
})
```

### 4. Store Operations

**Before:**
```go
store := ctx.KVStore(k.storeKey)
bz := store.Get(key)
if bz == nil {
    return nil, types.ErrNotFound
}
var pool types.LiquidityPool
if err := k.cdc.Unmarshal(bz, &pool); err != nil {
    return nil, err
}
```

**After:**
```go
import "github.com/aequitas/aura/chain/pkg/common"

store := ctx.KVStore(k.storeKey)
var pool types.LiquidityPool
found, err := common.GetObject(ctx, store, k.cdc, key, &pool)
if err != nil {
    return nil, err
}
if !found {
    return nil, types.ErrNotFound
}
```

### 5. Error Handling

**Before:**
```go
var (
    ErrInvalidAddress = errorsmod.Register("mymodule", 1001, "invalid address")
    ErrNotFound = errorsmod.Register("mymodule", 1002, "not found")
)
```

**After:**
```go
import "github.com/aequitas/aura/chain/pkg/common"

// Use common errors when appropriate
return common.WrapError(common.ErrInvalidAddress, "failed to validate sender: %s", msg.Sender)

// Or define module-specific errors
var ErrPoolNotFound = common.NewError("dex", 2001, "liquidity pool not found")
```

## Integration Checklist

For each module you update:

- [ ] Replace manual address validation with `common.ValidateAddress()`
- [ ] Replace pagination setup with `common.NormalizePagination()`
- [ ] Replace manual event emission with `common.EmitSuccessEvent()` / `common.EmitErrorEvent()`
- [ ] Replace manual marshal/unmarshal with `common.GetObject()` / `common.SetObject()`
- [ ] Use `common.WrapError()` for consistent error wrapping
- [ ] Run tests to ensure behavior is unchanged
- [ ] Update imports: `import "github.com/aequitas/aura/chain/pkg/common"`

## Testing

After integration, verify:

1. **Address validation** still rejects invalid addresses
2. **Pagination** defaults work correctly (100/page, max 1000)
3. **Events** are emitted with correct attributes
4. **Store operations** correctly handle not-found cases
5. **Error messages** remain descriptive and actionable

## Example Integration PR

See commit `9a1f3f4` for the shared package implementation.

Target modules for initial integration:
- identity
- dex  
- bridge
- compliance
- vcregistry

## Benefits

- **Reduced duplication**: Eliminate 100+ lines of repeated code per module
- **Consistency**: All modules use same validation/error patterns  
- **Security**: Centralized validation prevents missed edge cases
- **Maintainability**: Fix bugs once, benefit everywhere
- **Type safety**: Compile-time checks prevent runtime errors

