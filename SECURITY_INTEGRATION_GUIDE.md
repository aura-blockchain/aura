# Security Integration Guide for Remaining Modules

This guide provides step-by-step instructions for integrating the security utilities into the remaining Aura modules.

## Quick Start Template

### Step 1: Import Security Package

Add to keeper imports:
```go
import (
    "github.com/aequitas/aura/chain/x/common/security"
)
```

### Step 2: Add Security Fields to Keeper

```go
type Keeper struct {
    // ... existing fields ...

    // Security features
    reentrancyGuard *security.ReentrancyGuard
    pauseGuard      *security.PauseGuard
    inputValidator  *security.InputValidator
    safeMath        *security.SafeMath
    gasLimitGuard   *security.GasLimitGuard
    accessControl   *security.AccessControl
}
```

### Step 3: Initialize Security in NewKeeper

```go
func NewKeeper(...) *Keeper {
    return &Keeper{
        // ... existing initialization ...

        // Initialize security features
        reentrancyGuard: security.NewReentrancyGuard(),
        pauseGuard:      security.NewPauseGuard(""), // Admin via governance
        inputValidator:  security.NewInputValidator(),
        safeMath:        security.NewSafeMath(),
        gasLimitGuard:   security.NewGasLimitGuard(1000000),
        accessControl:   security.NewAccessControl([]string{}),
    }
}
```

### Step 4: Add Pause/Unpause Methods

```go
// Pause pauses the module (emergency use only)
func (k *Keeper) Pause(ctx sdk.Context, caller string) error {
    return k.pauseGuard.Pause(ctx, caller)
}

// Unpause unpauses the module
func (k *Keeper) Unpause(ctx sdk.Context, caller string) error {
    return k.pauseGuard.Unpause(ctx, caller)
}

// IsPaused checks if the module is paused
func (k *Keeper) IsPaused() bool {
    return k.pauseGuard.IsPaused()
}
```

### Step 5: Wrap Critical Functions

Pattern for wrapping existing functions:

```go
func (k *Keeper) CriticalOperation(
    ctx sdk.Context,
    address string,
    amount sdk.Int,
) error {
    // 1. Check if module is paused
    if err := k.pauseGuard.CheckNotPaused(); err != nil {
        return err
    }

    // 2. Reentrancy protection
    return k.reentrancyGuard.WithReentrancyGuard(func() error {
        // 3. Input validation
        if err := k.inputValidator.ValidateAddress(address); err != nil {
            return fmt.Errorf("invalid address: %w", err)
        }
        if err := k.inputValidator.ValidateAmount(amount); err != nil {
            return fmt.Errorf("invalid amount: %w", err)
        }

        // 4. Execute operation
        // ... your existing code ...

        // 5. Emit event
        ctx.EventManager().EmitEvent(
            sdk.NewEvent(
                "operation_completed",
                sdk.NewAttribute("address", address),
                sdk.NewAttribute("amount", amount.String()),
            ),
        )

        return nil
    })
}
```

---

## Module-Specific Integration

### VCRegistry Module

**Priority Functions to Secure:**

1. **SetVCRecord** (keeper.go:135)
```go
func (k *Keeper) SetVCRecord(record types.VCRecord) error {
    if err := k.pauseGuard.CheckNotPaused(); err != nil {
        return err
    }

    return k.reentrancyGuard.WithReentrancyGuard(func() error {
        // Validate inputs
        if err := k.inputValidator.ValidateString(record.VcId, "vc_id"); err != nil {
            return err
        }
        if err := k.inputValidator.ValidateAddress(record.HolderAddress); err != nil {
            return err
        }

        k.mu.Lock()
        defer k.mu.Unlock()

        k.vcRecords[record.VcId] = record
        k.userVCs[record.HolderAddress] = append(k.userVCs[record.HolderAddress], record.VcId)

        return nil
    })
}
```

2. **RevokeVC** (keeper.go:215)
```go
func (k *Keeper) RevokeVC(vcID string, reason types.RevocationReason, revoker string, evidence string) error {
    if err := k.pauseGuard.CheckNotPaused(); err != nil {
        return err
    }

    return k.reentrancyGuard.WithReentrancyGuard(func() error {
        if err := k.inputValidator.ValidateString(vcID, "vc_id"); err != nil {
            return err
        }
        if err := k.inputValidator.ValidateAddress(revoker); err != nil {
            return err
        }

        // ... rest of existing code ...
    })
}
```

**Event Additions:**
- Add events to SetVCRecord, RevokeVC, RegisterDID, etc.

---

### DataRegistry Module

**Priority Functions to Secure:**

1. **SetDataItem** (keeper.go:119)
```go
func (k *Keeper) SetDataItem(item types.DataItem) error {
    if err := k.pauseGuard.CheckNotPaused(); err != nil {
        return err
    }

    return k.reentrancyGuard.WithReentrancyGuard(func() error {
        if err := k.inputValidator.ValidateString(item.DataID, "data_id"); err != nil {
            return err
        }
        if err := k.inputValidator.ValidateAddress(item.OwnerAddress); err != nil {
            return err
        }

        // ... existing code ...

        // Add event
        ctx.EventManager().EmitEvent(
            sdk.NewEvent(
                "data_item_set",
                sdk.NewAttribute("data_id", item.DataID),
                sdk.NewAttribute("owner", item.OwnerAddress),
                sdk.NewAttribute("type", item.DataType.String()),
            ),
        )

        return nil
    })
}
```

2. **DeleteDataItem** (keeper.go:142)
- Add pause check
- Add reentrancy guard
- Add validation
- Add event emission

---

### Prevalidation Module

**Priority Functions to Secure:**

1. **CreatePreValidatedTransaction** (keeper.go:318)
```go
func (k *Keeper) CreatePreValidatedTransaction(
    txType types.TransactionType,
    templateID string,
    txData []byte,
    signer string,
    estimatedGas uint64,
    context map[string]string,
) (*types.PreValidatedTransaction, error) {
    if err := k.pauseGuard.CheckNotPaused(); err != nil {
        return nil, err
    }

    // Validate gas
    if err := k.gasLimitGuard.ValidateGasLimit(estimatedGas); err != nil {
        return nil, err
    }

    // ... rest of existing code ...
}
```

2. **ExecutePreValidatedTransaction** (keeper.go:493)
- Add pause check
- Already has good security, add event emissions

---

### InclusionRoutines Module

This module needs the most work. Here's a complete template:

```go
package keeper

import (
    "fmt"
    "sync"

    "github.com/aequitas/aura/chain/x/inclusionroutines/params"
    "github.com/aequitas/aura/chain/x/inclusionroutines/types"
    "github.com/aequitas/aura/chain/x/common/security"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
    mu            sync.RWMutex
    irs           map[string]types.IRDefinition
    prerequisites map[string]types.IRPrerequisite
    rateLimits    map[string]types.IRRateLimit
    rateLimitUsage map[string]int32
    paramsStore   *params.Store

    // Security features
    reentrancyGuard *security.ReentrancyGuard
    pauseGuard      *security.PauseGuard
    inputValidator  *security.InputValidator
    accessControl   *security.AccessControl
}

func NewKeeper(store *params.Store) *Keeper {
    if store == nil {
        store = params.NewStore(types.DefaultParams())
    }
    return &Keeper{
        irs:            make(map[string]types.IRDefinition),
        prerequisites:  make(map[string]types.IRPrerequisite),
        rateLimits:     make(map[string]types.IRRateLimit),
        rateLimitUsage: make(map[string]int32),
        paramsStore:    store,

        // Initialize security
        reentrancyGuard: security.NewReentrancyGuard(),
        pauseGuard:      security.NewPauseGuard(""),
        inputValidator:  security.NewInputValidator(),
        accessControl:   security.NewAccessControl([]string{}),
    }
}

// Pause/Unpause methods
func (k *Keeper) Pause(ctx sdk.Context, caller string) error {
    return k.pauseGuard.Pause(ctx, caller)
}

func (k *Keeper) Unpause(ctx sdk.Context, caller string) error {
    return k.pauseGuard.Unpause(ctx, caller)
}

func (k *Keeper) IsPaused() bool {
    return k.pauseGuard.IsPaused()
}

// Add secure wrapper for any IR registration/modification functions
```

---

### ConfidenceScore Module

**Priority Functions to Secure:**

1. **SetUserRecord** (keeper.go:98)
```go
func (k *Keeper) SetUserRecord(record types.UserConfidenceRecord) error {
    if err := k.pauseGuard.CheckNotPaused(); err != nil {
        return err
    }

    return k.reentrancyGuard.WithReentrancyGuard(func() error {
        if err := k.inputValidator.ValidateAddress(record.WalletAddress); err != nil {
            return err
        }

        k.mu.Lock()
        defer k.mu.Unlock()

        record.LastUpdatedHeight = k.currentHeight
        record.LastUpdated = k.currentTime

        k.userRecords[record.WalletAddress] = record

        // Add event
        ctx.EventManager().EmitEvent(
            sdk.NewEvent(
                "user_record_updated",
                sdk.NewAttribute("wallet", record.WalletAddress),
                sdk.NewAttribute("score", fmt.Sprintf("%d", record.TotalScore)),
                sdk.NewAttribute("status", record.Status.String()),
            ),
        )

        return nil
    })
}
```

2. **AddSlashRecord** (keeper.go:214)
- Add pause check
- Add validation
- Add event emission

---

## Testing Template

For each module, create `keeper_security_test.go`:

```go
package keeper

import (
    "testing"

    "github.com/aequitas/aura/chain/x/common/security"
    "github.com/stretchr/testify/require"
)

func TestPauseFunctionality(t *testing.T) {
    keeper := NewKeeper(nil)

    // Test pause
    err := keeper.Pause(ctx, "admin")
    require.NoError(t, err)
    require.True(t, keeper.IsPaused())

    // Test operation while paused
    err = keeper.SomeOperation(ctx, ...)
    require.Error(t, err)
    require.Equal(t, security.ErrModulePaused, err)

    // Test unpause
    err = keeper.Unpause(ctx, "admin")
    require.NoError(t, err)
    require.False(t, keeper.IsPaused())
}

func TestReentrancyProtection(t *testing.T) {
    // Test that nested calls are prevented
}

func TestInputValidation(t *testing.T) {
    // Test that invalid inputs are rejected
}
```

---

## Checklist for Each Module

### Pre-Integration
- [ ] Review existing code for security issues
- [ ] Identify critical functions
- [ ] Document current security measures

### Integration
- [ ] Add security imports
- [ ] Add security fields to Keeper
- [ ] Initialize security in NewKeeper
- [ ] Add Pause/Unpause methods
- [ ] Wrap critical functions with guards
- [ ] Add input validation
- [ ] Add event emissions
- [ ] Update error handling

### Post-Integration
- [ ] Write security tests
- [ ] Update documentation
- [ ] Run integration tests
- [ ] Performance testing
- [ ] Security audit

---

## Common Pitfalls to Avoid

### 1. Forgetting to Check Pause State
❌ **Wrong:**
```go
func (k *Keeper) Operation() error {
    // Missing pause check
    return k.doSomething()
}
```

✅ **Correct:**
```go
func (k *Keeper) Operation() error {
    if err := k.pauseGuard.CheckNotPaused(); err != nil {
        return err
    }
    return k.doSomething()
}
```

### 2. Nested Reentrancy Guards
❌ **Wrong:**
```go
func (k *Keeper) Operation1() error {
    return k.reentrancyGuard.WithReentrancyGuard(func() error {
        return k.Operation2() // Operation2 also has guard!
    })
}
```

✅ **Correct:**
```go
func (k *Keeper) Operation1() error {
    return k.reentrancyGuard.WithReentrancyGuard(func() error {
        return k.operation2Internal() // Internal version without guard
    })
}
```

### 3. Missing Event Emissions
❌ **Wrong:**
```go
func (k *Keeper) UpdateState() error {
    // Update state
    // Missing event emission
    return nil
}
```

✅ **Correct:**
```go
func (k *Keeper) UpdateState() error {
    // Update state

    ctx.EventManager().EmitEvent(
        sdk.NewEvent("state_updated", ...),
    )
    return nil
}
```

### 4. Unsafe Math Operations
❌ **Wrong:**
```go
result := a.Add(b) // No overflow check
```

✅ **Correct:**
```go
result, err := k.safeMath.SafeAdd(a, b)
if err != nil {
    return err
}
```

---

## Performance Considerations

### Gas Impact
- Each security check adds ~10-15k gas
- Optimize by:
  - Batching operations
  - Caching pause state
  - Minimizing validation for trusted calls

### Memory Impact
- Security guards are lightweight
- Minimal memory overhead per Keeper
- Use sync.RWMutex for read-heavy operations

---

## Support and Questions

For questions or issues:
1. Review the security implementation summary
2. Check existing implementations in bridge/dex modules
3. Refer to security test examples
4. Consult the security utilities documentation

---

## Version History

- v1.0 (2025-11-13): Initial security framework
- Future: External audit, formal verification

