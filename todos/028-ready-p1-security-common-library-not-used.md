---
id: "028"
title: "Common Security Library Exists But Not Used"
status: ready
priority: p1
category: architecture
module: security
severity: CRITICAL
cvss: 9.0
source: security-audit-report
---

# Common Security Library Exists But Not Used

## Problem

A comprehensive security library exists at `chain/x/security/keeper/common.go` with:
- `ReentrancyGuard`
- `PauseGuard`
- `InputValidator`
- `SafeMath`
- `AccessControl`
- `RateLimiter`

**NO OTHER MODULE IN THE CODEBASE USES ANY OF THESE SECURITY PRIMITIVES.**

## Affected Files

- `chain/x/security/keeper/common.go` (exists but unused)
- ALL other modules (missing security controls)

## Current State

```go
// common.go defines these interfaces:
type ReentrancyGuard interface {
    EnterNoReentrant(ctx sdk.Context, key string) error
    ExitNoReentrant(ctx sdk.Context, key string)
}

type PauseGuard interface {
    RequireNotPaused(ctx sdk.Context) error
    Pause(ctx sdk.Context) error
    Unpause(ctx sdk.Context) error
}

// BUT: Search for usage returns 0 results
// grep -r "ReentrancyGuard" chain/x/*/keeper/ -> common.go only
// grep -r "PauseGuard" chain/x/*/keeper/ -> common.go only
```

## Impact

- No reentrancy protection across entire codebase
- No emergency pause capability
- No rate limiting
- Duplicated validation logic everywhere
- Security vulnerabilities in every module

## Required Fix

### Step 1: Create Security Module Interface

```go
// chain/x/security/keeper/interfaces.go
package keeper

type SecurityKeeper interface {
    // Reentrancy protection
    EnterNoReentrant(ctx sdk.Context, key string) error
    ExitNoReentrant(ctx sdk.Context, key string)

    // Emergency pause
    RequireNotPaused(ctx sdk.Context, moduleName string) error
    PauseModule(ctx sdk.Context, moduleName string) error
    UnpauseModule(ctx sdk.Context, moduleName string) error

    // Rate limiting
    CheckRateLimit(ctx sdk.Context, key string, limit uint64, window time.Duration) error
    IncrementRateLimit(ctx sdk.Context, key string)

    // Input validation
    ValidateAddress(address string) error
    ValidateAmount(amount sdkmath.Int, min, max sdkmath.Int) error
}
```

### Step 2: Inject into All Modules

```go
// Example: chain/x/dex/keeper/keeper.go
type Keeper struct {
    cdc             codec.BinaryCodec
    storeKey        storetypes.StoreKey
    bankKeeper      types.BankKeeper
    securityKeeper  securitytypes.SecurityKeeper  // ADD THIS
}

func NewKeeper(
    cdc codec.BinaryCodec,
    storeKey storetypes.StoreKey,
    bankKeeper types.BankKeeper,
    securityKeeper securitytypes.SecurityKeeper,  // ADD THIS
) *Keeper {
    return &Keeper{
        cdc:            cdc,
        storeKey:       storeKey,
        bankKeeper:     bankKeeper,
        securityKeeper: securityKeeper,  // ADD THIS
    }
}
```

### Step 3: Use in All State-Changing Functions

```go
func (k Keeper) Swap(ctx sdk.Context, msg *types.MsgSwap) (*types.MsgSwapResponse, error) {
    // Check pause status
    if err := k.securityKeeper.RequireNotPaused(ctx, types.ModuleName); err != nil {
        return nil, err
    }

    // Reentrancy guard
    if err := k.securityKeeper.EnterNoReentrant(ctx, "swap:"+msg.PoolId); err != nil {
        return nil, err
    }
    defer k.securityKeeper.ExitNoReentrant(ctx, "swap:"+msg.PoolId)

    // Rate limit check
    if err := k.securityKeeper.CheckRateLimit(ctx, "swap:"+msg.Sender, 100, time.Minute); err != nil {
        return nil, err
    }
    k.securityKeeper.IncrementRateLimit(ctx, "swap:"+msg.Sender)

    // Rest of swap logic...
}
```

## Modules Requiring Integration

1. `dex` - Swap, AddLiquidity, RemoveLiquidity, PlaceOrder
2. `bridge` - LockTokens, UnlockTokens, ValidatorOperations
3. `governance` - Vote, Delegate, Execute
4. `walletsecurity` - All operations
5. `compliance` - KYC, AML operations
6. `vcregistry` - VC operations
7. `identity` - Identity operations
8. `privacy` - Privacy operations

## Acceptance Criteria

- [ ] SecurityKeeper interface defined
- [ ] SecurityKeeper injected into all modules
- [ ] ReentrancyGuard used in all state-changing functions
- [ ] PauseGuard used in all modules
- [ ] RateLimiter used for expensive operations
- [ ] Tests for reentrancy protection
- [ ] Tests for pause functionality
- [ ] Tests for rate limiting
