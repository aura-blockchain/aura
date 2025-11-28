# Keeper Builder Pattern Guide

## Quick Start

### Why Builder Pattern?

The builder pattern eliminates circular dependencies by ensuring **all dependencies are set BEFORE the keeper is constructed**, rather than using post-construction mutation.

### Basic Usage

```go
// ❌ OLD WAY (Causes Circular Dependencies)
csKeeper := cskeeper.NewKeeper(csParamsStore, authorityAddr)
csKeeper.SetIRRegistry(irKeeper)  // Post-construction mutation

vcKeeper := vckeeper.NewKeeper(vcParamsStore, authorityAddr).WithStore(vcKey, cdc)
vcKeeper.SetConfidenceScoreKeeper(csKeeper)  // Post-construction mutation

// Problem: Both keepers depend on each other, causing undefined behavior

// ✅ NEW WAY (Builder Pattern)
// Step 1: Initialize dependency first (irKeeper)
irKeeper := irkeeper.NewKeeper(irParamsStore, authorityAddr)

// Step 2: Build csKeeper with all dependencies
csKeeper := cskeeper.NewKeeperBuilder(csParamsStore, authorityAddr).
    WithIRRegistry(irKeeper).  // Set dependency FIRST
    Build()                     // Returns immutable keeper

// Step 3: Build vcKeeper with all dependencies
vcKeeper := vckeeper.NewKeeperBuilder(vcParamsStore, authorityAddr).
    WithStore(vcKey, encoding.Codec).
    WithConfidenceScoreKeeper(csKeeper).  // Set dependency FIRST
    Build()                                // Returns immutable keeper

// Result: Zero circular dependencies, predictable initialization
```

## Implementation Guide

### Creating a New Keeper Builder

**File**: `x/yourmodule/keeper/builder.go`

```go
package keeper

import (
    "sync"
    storetypes "cosmossdk.io/store/types"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/aequitas/aura/chain/x/yourmodule/params"
    "github.com/aequitas/aura/chain/x/yourmodule/types"
)

// KeeperBuilder provides type-safe builder pattern
type KeeperBuilder struct {
    // Required fields
    paramsStore *params.Store
    authority   string

    // Optional dependencies (set via With* methods)
    storeKey    storetypes.StoreKey
    codec       codec.BinaryCodec
    otherKeeper OtherKeeperInterface
}

// NewKeeperBuilder initializes builder with required parameters
func NewKeeperBuilder(paramsStore *params.Store, authority string) *KeeperBuilder {
    if paramsStore == nil {
        paramsStore = params.NewStore(*types.DefaultParams())
    }
    return &KeeperBuilder{
        paramsStore: paramsStore,
        authority:   authority,
    }
}

// WithStore sets the KV store key and codec
func (b *KeeperBuilder) WithStore(storeKey storetypes.StoreKey, codec codec.BinaryCodec) *KeeperBuilder {
    b.storeKey = storeKey
    b.codec = codec
    return b
}

// WithOtherKeeper sets a dependency keeper
func (b *KeeperBuilder) WithOtherKeeper(keeper OtherKeeperInterface) *KeeperBuilder {
    b.otherKeeper = keeper
    return b
}

// Build constructs the immutable keeper
// Panics if required dependencies are missing (fail-fast)
func (b *KeeperBuilder) Build() *Keeper {
    // Validate required dependencies
    if b.storeKey == nil {
        panic("yourmodule keeper builder: storeKey is required")
    }
    if b.codec == nil {
        panic("yourmodule keeper builder: codec is required")
    }
    if b.otherKeeper == nil {
        panic("yourmodule keeper builder: other keeper is required")
    }

    // Create the store
    store := NewStore(b.storeKey, b.codec)

    // Return immutable keeper with all dependencies
    return &Keeper{
        mu:          sync.RWMutex{},
        store:       &store,
        paramsStore: b.paramsStore,
        otherKeeper: b.otherKeeper,
        authority:   b.authority,
    }
}

// Validate checks all dependencies are set
func (b *KeeperBuilder) Validate() error {
    if b.storeKey == nil {
        return types.ErrInvalidRequest.Wrap("storeKey required")
    }
    if b.codec == nil {
        return types.ErrInvalidRequest.Wrap("codec required")
    }
    if b.otherKeeper == nil {
        return types.ErrInvalidRequest.Wrap("other keeper required")
    }
    return nil
}
```

## Best Practices

### 1. Always Use Builder for New Keepers

```go
// ✅ CORRECT: Use builder pattern
keeper := NewKeeperBuilder(paramsStore, authority).
    WithDependency1(dep1).
    WithDependency2(dep2).
    Build()

// ❌ INCORRECT: Direct construction with setters
keeper := NewKeeper(paramsStore, authority)
keeper.SetDependency1(dep1)  // Post-construction mutation
```

### 2. Set Dependencies in Correct Order

```go
// Dependency Graph:
// A -> B -> C (A depends on B, B depends on C)

// ✅ CORRECT: Initialize in reverse dependency order
cKeeper := NewCKeeperBuilder(...).Build()
bKeeper := NewBKeeperBuilder(...).WithC(cKeeper).Build()
aKeeper := NewAKeeperBuilder(...).WithB(bKeeper).Build()

// ❌ INCORRECT: Initialize in wrong order
aKeeper := NewAKeeperBuilder(...).WithB(nil).Build()  // B doesn't exist yet!
```

### 3. Validate Before Build

```go
// Optional: Validate before building
builder := NewKeeperBuilder(paramsStore, authority).
    WithStore(storeKey, codec).
    WithDependency(dep)

if err := builder.Validate(); err != nil {
    logger.Error("invalid builder configuration", "error", err)
    return err
}

keeper := builder.Build()
```

### 4. Panic is Intentional

```go
// Panic on missing dependencies is INTENTIONAL
// This is fail-fast behavior to catch configuration errors at startup

func (b *KeeperBuilder) Build() *Keeper {
    if b.storeKey == nil {
        panic("storeKey required")  // ✅ CORRECT: Fail at startup
    }
    // NOT: return nil, error  // ❌ INCORRECT: Silent failure
}
```

## Integration with app.go

### Initialization Order in app.go

```go
// Tier 1: No dependencies
keeper1 := NewKeeper1Builder(...).Build()

// Tier 2: Depends on Tier 1
keeper2 := NewKeeper2Builder(...).
    WithKeeper1(keeper1).
    Build()

// Tier 3: Depends on Tier 2
keeper3 := NewKeeper3Builder(...).
    WithKeeper2(keeper2).
    Build()
```

### Logging Initialization

```go
logger.Info("initializing keepers", "phase", "tier-1")
keeper1 := NewKeeper1Builder(...).Build()

logger.Info("initializing keepers", "phase", "tier-2")
keeper2 := NewKeeper2Builder(...).WithKeeper1(keeper1).Build()
```

## Testing Builder Pattern

### Test Missing Dependencies

```go
func TestKeeperBuilder_MissingStore(t *testing.T) {
    builder := NewKeeperBuilder(paramsStore, authority)
    // Don't set store

    assert.Panics(t, func() {
        builder.Build()  // Should panic
    })
}

func TestKeeperBuilder_AllDepsSet(t *testing.T) {
    builder := NewKeeperBuilder(paramsStore, authority).
        WithStore(storeKey, codec).
        WithDependency(dep)

    assert.NotPanics(t, func() {
        keeper := builder.Build()  // Should succeed
        assert.NotNil(t, keeper)
    })
}
```

### Test Validate Method

```go
func TestKeeperBuilder_Validate(t *testing.T) {
    builder := NewKeeperBuilder(paramsStore, authority)

    err := builder.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "storeKey required")

    builder.WithStore(storeKey, codec).WithDependency(dep)

    err = builder.Validate()
    assert.NoError(t, err)
}
```

## Migration from Old Pattern

### Before (Old Pattern)

```go
keeper := NewKeeper(paramsStore, authority)
keeper.WithStore(storeKey, codec)
keeper.SetDependency(dep)
```

### After (Builder Pattern)

```go
keeper := NewKeeperBuilder(paramsStore, authority).
    WithStore(storeKey, codec).
    WithDependency(dep).
    Build()
```

### Migration Steps

1. Create `builder.go` file
2. Implement `KeeperBuilder` struct
3. Implement `NewKeeperBuilder()` function
4. Implement `With*()` methods for each dependency
5. Implement `Build()` method with validation
6. Update `app.go` to use builder
7. Remove old setter methods from keeper

## Common Patterns

### Optional Dependencies

```go
type KeeperBuilder struct {
    required    RequiredInterface
    optional    OptionalInterface  // May be nil
}

func (b *KeeperBuilder) Build() *Keeper {
    if b.required == nil {
        panic("required dependency missing")
    }
    // optional can be nil
    return &Keeper{
        required: b.required,
        optional: b.optional,
    }
}
```

### Multiple Store Keys

```go
func (b *KeeperBuilder) WithStores(
    kvStore storetypes.StoreKey,
    memStore storetypes.MemoryStoreKey,
) *KeeperBuilder {
    b.kvStore = kvStore
    b.memStore = memStore
    return b
}
```

### Conditional Dependencies

```go
func (b *KeeperBuilder) WithBankKeeper(keeper BankKeeper) *KeeperBuilder {
    b.bankKeeper = keeper
    return b
}

func (b *KeeperBuilder) WithAccountKeeper(keeper AccountKeeper) *KeeperBuilder {
    b.accountKeeper = keeper
    return b
}

func (b *KeeperBuilder) Build() *Keeper {
    // At least one must be set
    if b.bankKeeper == nil && b.accountKeeper == nil {
        panic("either bank keeper or account keeper required")
    }
    return &Keeper{...}
}
```

## Troubleshooting

### Panic: "required dependency missing"

**Problem**: Forgot to call `With*()` method
**Solution**: Set all required dependencies before `Build()`

```go
// ❌ Missing dependency
keeper := NewKeeperBuilder(paramsStore, authority).Build()

// ✅ All dependencies set
keeper := NewKeeperBuilder(paramsStore, authority).
    WithStore(storeKey, codec).
    WithDependency(dep).
    Build()
```

### Panic: "nil pointer dereference"

**Problem**: Dependency initialized out of order
**Solution**: Check initialization order in app.go

```go
// ❌ B depends on C, but C not initialized yet
bKeeper := NewBKeeperBuilder(...).WithC(cKeeper).Build()  // cKeeper is nil
cKeeper := NewCKeeperBuilder(...).Build()

// ✅ Initialize in correct order
cKeeper := NewCKeeperBuilder(...).Build()
bKeeper := NewBKeeperBuilder(...).WithC(cKeeper).Build()
```

### Circular Dependency Error

**Problem**: Two keepers depend on each other
**Solution**: Refactor to eliminate circular dependency

```go
// ❌ Circular: A -> B -> A
aKeeper needs bKeeper
bKeeper needs aKeeper

// ✅ Solution 1: Extract interface
aKeeper needs BInterface (only part of B)
bKeeper implements BInterface
No circular dependency!

// ✅ Solution 2: Create intermediate keeper
aKeeper -> cKeeper <- bKeeper
Both depend on C, C depends on neither
```

## Summary

✅ **Use builder pattern for all new keepers**
✅ **Set dependencies BEFORE Build()**
✅ **Initialize keepers in dependency order**
✅ **Panic on missing dependencies (fail-fast)**
✅ **Test builder validation**
✅ **Document dependencies in comments**

The builder pattern ensures type-safe, predictable initialization with zero circular dependencies.
