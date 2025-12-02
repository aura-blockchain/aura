# Centralized Store Key Architecture

## Overview

The Aura blockchain uses a **centralized store key architecture** to eliminate duplication and prevent runtime panics when adding new modules. All store key definitions are managed from a single source of truth.

## The Problem (Before Centralization)

Previously, store keys were defined in **THREE separate locations**:

1. `StoreKeyNames()` function - list of key names
2. `app.storeKeys` struct fields - key storage
3. `base.MountKVStores()` call - key mounting

**Adding a module required editing all 3 locations.** Missing one = runtime panic.

## The Solution (Centralized Architecture)

All store key operations now derive from a single `storeKeys` struct with three methods:

```go
type storeKeys struct {
    account    *storetypes.KVStoreKey
    bank       *storetypes.KVStoreKey
    staking    *storetypes.KVStoreKey
    // ... all other keys
}

// Single source of truth
func initStoreKeys() *storeKeys {
    return &storeKeys{
        account: storetypes.NewKVStoreKey(authtypes.StoreKey),
        bank:    storetypes.NewKVStoreKey(banktypes.StoreKey),
        // ... all keys initialized
    }
}

// Derive list of names
func (s *storeKeys) Names() []string {
    return []string{
        authtypes.StoreKey,
        banktypes.StoreKey,
        // ... derived from initStoreKeys
    }
}

// Derive map for mounting
func (s *storeKeys) AsMap() map[string]*storetypes.KVStoreKey {
    return map[string]*storetypes.KVStoreKey{
        authtypes.StoreKey: s.account,
        banktypes.StoreKey: s.bank,
        // ... derived from initStoreKeys
    }
}
```

## How to Add a New Module

**Only 4 steps (all in one place):**

1. **Add field to `storeKeys` struct:**
   ```go
   type storeKeys struct {
       // ... existing fields
       newModule *storetypes.KVStoreKey
   }
   ```

2. **Initialize in `initStoreKeys()`:**
   ```go
   func initStoreKeys() *storeKeys {
       return &storeKeys{
           // ... existing initializations
           newModule: storetypes.NewKVStoreKey(newmoduletypes.StoreKey),
       }
   }
   ```

3. **Add to `Names()` method:**
   ```go
   func (s *storeKeys) Names() []string {
       return []string{
           // ... existing names
           newmoduletypes.StoreKey,
       }
   }
   ```

4. **Add to `AsMap()` method:**
   ```go
   func (s *storeKeys) AsMap() map[string]*storetypes.KVStoreKey {
       return map[string]*storetypes.KVStoreKey{
           // ... existing mappings
           newmoduletypes.StoreKey: s.newModule,
       }
   }
   ```

**That's it!** The following are automatically handled:

- ✅ `MountKVStores()` uses `AsMap()`
- ✅ `StoreKeyNames()` uses `Names()`
- ✅ `validateStoreKeys()` uses `AsMap()`
- ✅ `allStoreKeys()` uses struct fields

## Benefits

### Before (Fragile)
```
Adding a module:
1. Edit StoreKeyNames() - add key name
2. Edit storeKeys struct - add field
3. Edit MountKVStores() - add to map
4. Edit allStoreKeys() - add to slice
5. Edit validateStoreKeys() - add to validation map

Miss one? → RUNTIME PANIC
```

### After (Robust)
```
Adding a module:
1. Edit initStoreKeys() - add initialization
2. Edit Names() - add name
3. Edit AsMap() - add mapping
4. Edit storeKeys struct - add field

Everything else is automatic!
```

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    initStoreKeys()                           │
│              (Single Source of Truth)                        │
│                                                              │
│  Creates all store keys with proper initialization          │
└──────────────────────┬──────────────────────────────────────┘
                       │
           ┌───────────┴───────────┐
           │                       │
           ▼                       ▼
    ┌─────────────┐         ┌─────────────┐
    │  Names()    │         │  AsMap()    │
    │  method     │         │  method     │
    └──────┬──────┘         └──────┬──────┘
           │                       │
           │                       │
           ▼                       ▼
    ┌─────────────────┐    ┌─────────────────┐
    │ StoreKeyNames() │    │ MountKVStores() │
    │                 │    │ validateKeys()  │
    └─────────────────┘    └─────────────────┘
```

## Testing

The centralized architecture is validated by comprehensive tests:

### Store Key Tests (`store_keys_test.go`)
- `TestStoreKeysCentralization` - Verifies basic centralization
- `TestAppInitializationWithCentralizedKeys` - Verifies app creation
- `TestSingleSourceOfTruthConsistency` - Verifies consistency across methods

### Validation Tests (`validation_centralized_test.go`)
- `TestValidationUsesCentralizedStoreKeys` - Verifies validation uses centralized keys
- `TestValidateStoreKeysUsesCentralizedMap` - Verifies AsMap() usage
- `TestStoreKeyConsistencyAcrossAppAndValidation` - Verifies consistency
- `TestAddingNewModuleWorkflow` - Documents and tests the workflow
- `TestNoDuplicateStoreKeys` - Prevents duplicates
- `TestStoreKeysMatchExpectedModules` - Verifies all modules present

### Running Tests
```bash
cd chain
go test ./app -v -run TestStoreKeys
go test ./app -v -run TestValidation
go test ./app -v  # Run all app tests
```

## File Locations

- **Main implementation:** `/home/decri/blockchain-projects/aura/chain/app/app.go`
  - Lines 188-371: `storeKeys` struct and methods

- **Tests:**
  - `/home/decri/blockchain-projects/aura/chain/app/store_keys_test.go`
  - `/home/decri/blockchain-projects/aura/chain/app/validation_centralized_test.go`

- **Usage in validation:** `/home/decri/blockchain-projects/aura/chain/app/validation.go`
  - Line 82: `storeKeys := app.storeKeys.AsMap()`

## Best Practices

### DO ✅
- Add new module keys to all 4 locations in `app.go` (struct, init, Names, AsMap)
- Run tests after adding a new module
- Use `AsMap()` for operations that need the full key map
- Use `Names()` for operations that only need key names

### DON'T ❌
- Create hardcoded lists of store keys elsewhere in the codebase
- Directly access `app.storeKeys` fields from outside the app package
- Skip updating `Names()` or `AsMap()` when adding a key
- Add keys in different orders across the methods

## Security Considerations

The centralized architecture improves security by:

1. **Preventing runtime panics** - Missing keys are caught at compile time or test time
2. **Enforcing consistency** - Single source of truth prevents mismatches
3. **Simplifying audits** - All keys defined in one location
4. **Reducing bugs** - Less manual coordination reduces human error

## Migration from Old Architecture

The migration from the old 3-location system to the centralized system involved:

1. Creating the `storeKeys` struct with all fields
2. Implementing `initStoreKeys()` with all initializations
3. Implementing `Names()` and `AsMap()` methods
4. Updating `MountKVStores()` to use `AsMap()`
5. Updating `StoreKeyNames()` to use `Names()`
6. Updating `validateStoreKeys()` to use `AsMap()`
7. Removing all hardcoded lists
8. Adding comprehensive tests

The migration was completed without any runtime errors or test failures.

## Future Improvements

Potential future enhancements:

1. **Code generation** - Generate `Names()` and `AsMap()` from struct fields automatically
2. **Reflection-based validation** - Use reflection to verify consistency
3. **Module metadata** - Attach additional metadata to each key (priority, dependencies)
4. **Dynamic key loading** - Support loading keys from configuration

## References

- Cosmos SDK Store Documentation: https://docs.cosmos.network/main/core/store
- IAVL Store: https://github.com/cosmos/iavl
- Cosmos SDK BaseApp: https://docs.cosmos.network/main/core/baseapp
