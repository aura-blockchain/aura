# Privacy Module BankKeeper Interface - Quick Reference

## Summary
The privacy module's BankKeeper interface is **correctly implemented** for Cosmos SDK v0.53.4. The fix involved updating mock implementations in tests to use `context.Context` instead of `sdk.Context`.

## Correct Interface (Production-Ready)

### File: `/home/decri/blockchain-projects/aura/chain/x/privacy/types/expected_keepers.go`

```go
package types

import (
    "context"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
    SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
    SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
    SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
    MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error  // ✓ string, not []byte
    BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error  // ✓ string, not []byte
    GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
```

## Key Points

### ✅ Correct Signatures (Cosmos SDK v0.53.4)
- **Context Type**: `context.Context` (modern SDK pattern)
- **Module Name Type**: `string` (NOT `[]byte`)
- **Coins Type**: `sdk.Coins`

### ✅ Fixed Files
1. `/home/decri/blockchain-projects/aura/chain/x/privacy/keeper/keeper_type_safety_test.go`
   - Updated `MockBankKeeper` to use `context.Context`

2. `/home/decri/blockchain-projects/aura/chain/x/privacy/types/expected_keepers_test.go`
   - Updated `MockBankKeeper` to use `context.Context`

### ✅ Compatibility
- Works with `bankkeeper.BaseKeeper` from Cosmos SDK v0.53.4
- Follows modern Cosmos SDK v0.50+ patterns
- Type-safe interface implementation

## Common Mistakes to Avoid

### ❌ Wrong: Using `[]byte` for module name
```go
// This is WRONG for Cosmos SDK v0.53.4
BurnCoins(ctx context.Context, moduleName []byte, amt sdk.Coins) error
```

### ✅ Correct: Using `string` for module name
```go
// This is CORRECT for Cosmos SDK v0.53.4
BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
```

### ❌ Wrong: Using `sdk.Context` in mocks
```go
// This is WRONG - creates interface mismatch
func (m MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
```

### ✅ Correct: Using `context.Context` in mocks
```go
// This is CORRECT - matches interface
func (m MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
```

## Verification Commands

```bash
# Compile privacy module
cd /home/decri/blockchain-projects/aura/chain
go build ./x/privacy/...

# Verify interface compatibility
cat > /tmp/test.go << 'EOF'
package main
import (
    bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
    privacytypes "github.com/aequitas/aura/chain/x/privacy/types"
)
func main() {
    var bk bankkeeper.BaseKeeper
    var _ privacytypes.BankKeeper = bk // Should compile
}
EOF
go build /tmp/test.go
```

## Integration Example

```go
// In app/app.go
import (
    privacykeeper "github.com/aequitas/aura/chain/x/privacy/keeper"
    bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// Initialize bank keeper
bankKeeper := bankkeeper.NewBaseKeeper(
    cdc,
    runtime.NewKVStoreService(bankKey),
    accountKeeper,
    blockedModuleAddresses(moduleAccountPermissions),
    authorityAddr,
    logger,
)

// Pass directly to privacy keeper (no adapter needed)
privacyKeeper := privacykeeper.NewKeeper(
    cdc,
    privacyKey,
    accountKeeper,
    bankKeeper, // bankkeeper.BaseKeeper implements types.BankKeeper
)
```

## Module Comparison

| Module | Context Type | Module Name Type | SDK Version Compatibility |
|--------|--------------|------------------|---------------------------|
| Privacy | `context.Context` | `string` | v0.50+ / v0.53.4 ✅ |
| Bridge | `sdk.Context` | `string` | Legacy (needs adapter) |
| Dex | `sdk.Context` | `string` | Legacy (needs adapter) |

## Status

✅ **Production Ready**
- Interface correctly defined
- Mocks updated
- Compilation successful
- SDK v0.53.4 compatible
