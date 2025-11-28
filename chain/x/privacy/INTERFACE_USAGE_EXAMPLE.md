# Privacy Module BankKeeper Interface - Usage Examples

## Correct Interface Definition

```go
// File: chain/x/privacy/types/expected_keepers.go
package types

import (
    "context"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected bank keeper interface for privacy operations
type BankKeeper interface {
    SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
    SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
    SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
    MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
    BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
    GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
```

## Using the BankKeeper in Production Code

### Example 1: Minting Coins for Privacy Operations

```go
package keeper

import (
    "context"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// PrivacyMint mints coins for privacy pool
func (k Keeper) PrivacyMint(ctx context.Context, amount sdk.Coins) error {
    // Use string for module name (not []byte)
    return k.bankKeeper.MintCoins(ctx, "privacy", amount)
}
```

### Example 2: Burning Coins from Privacy Pool

```go
// PrivacyBurn burns coins from privacy pool
func (k Keeper) PrivacyBurn(ctx context.Context, amount sdk.Coins) error {
    // Use string for module name (not []byte)
    return k.bankKeeper.BurnCoins(ctx, "privacy", amount)
}
```

### Example 3: Shielded Transfer with Bank Operations

```go
// ShieldFunds shields user funds by moving to privacy module
func (k Keeper) ShieldFunds(ctx context.Context, from sdk.AccAddress, amount sdk.Coins) error {
    // Use context.Context (not sdk.Context)
    return k.bankKeeper.SendCoinsFromAccountToModule(ctx, from, "privacy", amount)
}

// UnshieldFunds unshields funds by moving from privacy module to user
func (k Keeper) UnshieldFunds(ctx context.Context, to sdk.AccAddress, amount sdk.Coins) error {
    // Use context.Context (not sdk.Context)
    return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, "privacy", to, amount)
}
```

## Creating Mock Implementations for Tests

### Correct Mock Implementation

```go
// File: chain/x/privacy/keeper/keeper_type_safety_test.go
package keeper_test

import (
    "context"  // IMPORTANT: Use context.Context
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockBankKeeper implements types.BankKeeper for testing
type MockBankKeeper struct{}

// All methods use context.Context (not sdk.Context)
func (m MockBankKeeper) SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
    return nil
}

func (m MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
    return nil
}

func (m MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
    return nil
}

// Use string for moduleName (not []byte)
func (m MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
    return nil
}

// Use string for moduleName (not []byte)
func (m MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
    return nil
}

func (m MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
    return sdk.NewInt64Coin(denom, 0)
}
```

## Integration with App

### Keeper Initialization

```go
// File: chain/app/app.go
package app

import (
    bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
    privacykeeper "github.com/aequitas/aura/chain/x/privacy/keeper"
)

func (app *AuraApp) initKeepers() {
    // Initialize bank keeper (Cosmos SDK v0.53.4)
    bankKeeper := bankkeeper.NewBaseKeeper(
        app.encoding.Codec,
        runtime.NewKVStoreService(app.storeKeys.bank),
        app.accountKeeper,
        blockedModuleAddresses(moduleAccountPermissions),
        authorityAddr,
        logger,
    )

    // Initialize privacy keeper
    // bankKeeper implements types.BankKeeper directly - no adapter needed!
    app.privacyKeeper = privacykeeper.NewKeeper(
        app.encoding.Codec,
        app.storeKeys.privacy,
        app.accountKeeper,
        bankKeeper, // Direct compatibility ✓
    )
}
```

## Common Mistakes and How to Avoid Them

### ❌ WRONG: Using []byte for module name

```go
// This will NOT compile
func (k Keeper) BadExample(ctx context.Context) error {
    moduleName := []byte("privacy")
    return k.bankKeeper.BurnCoins(ctx, moduleName, coins) // ERROR!
}
```

### ✅ CORRECT: Using string for module name

```go
// This will compile correctly
func (k Keeper) GoodExample(ctx context.Context) error {
    moduleName := "privacy"
    return k.bankKeeper.BurnCoins(ctx, moduleName, coins) // ✓
}
```

### ❌ WRONG: Using sdk.Context in mocks

```go
// This creates interface mismatch
type MockBankKeeper struct{}

func (m MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
    return nil // ERROR: ctx type mismatch!
}
```

### ✅ CORRECT: Using context.Context in mocks

```go
// This matches the interface
type MockBankKeeper struct{}

func (m MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
    return nil // ✓
}
```

## Context Conversion (When Needed)

If you need to convert between context types:

```go
import (
    "context"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// Convert sdk.Context to context.Context
func convertToContext(sdkCtx sdk.Context) context.Context {
    return sdk.WrapSDKContext(sdkCtx)
}

// Extract sdk.Context from context.Context
func extractSDKContext(ctx context.Context) sdk.Context {
    return sdk.UnwrapSDKContext(ctx)
}
```

However, in the privacy module keeper, you should use `context.Context` directly since that's what the SDK provides in v0.50+.

## Type Safety Verification

```go
// This code should compile without errors
import (
    bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
    privacytypes "github.com/aequitas/aura/chain/x/privacy/types"
)

func verifyInterfaceCompatibility() {
    var bk bankkeeper.BaseKeeper
    var _ privacytypes.BankKeeper = bk // ✓ Compiles successfully
}
```

## Summary

✅ **Use `context.Context`** for all keeper method parameters
✅ **Use `string`** for module names in MintCoins/BurnCoins
✅ **Match interface signatures** exactly in mock implementations
✅ **No adapters needed** - bankkeeper.BaseKeeper implements the interface directly

This ensures:
- Type safety
- Cosmos SDK v0.53.4 compatibility
- Modern SDK patterns
- Production-ready code
