# CRITICAL: Missing Unmarshal Error Handling Causes Silent Failures

**Status:** ready
**Priority:** P0 (MAINNET BLOCKER)
**Severity:** CRITICAL
**CWE:** CWE-754 (Improper Check for Unusual or Exceptional Conditions)
**CVSS Score:** 8.6

## Summary

Protobuf unmarshal errors are not properly handled throughout the codebase, leading to silent failures, state corruption, and potential DoS attacks.

## Location

**Affected Files (17+ instances):**
- `chain/x/compliance/keeper/keeper.go` (8 instances)
- `chain/x/bridge/keeper/keeper.go` (5 instances)
- `chain/x/dex/keeper/keeper.go` (4 instances)
- Multiple other keepers

## Vulnerability Pattern

```go
// INSECURE PATTERN - Silent failure:
func (k Keeper) GetKYCRecord(ctx sdk.Context, address string) *types.KYCRecord {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get(types.KYCKey(address))
    if bz == nil {
        return nil
    }

    var record types.KYCRecord
    k.cdc.MustUnmarshal(bz, &record)  // PANICS on error!
    return &record
}

// OR WORSE - Ignores errors:
func (k Keeper) GetKYCRecord(ctx sdk.Context, address string) *types.KYCRecord {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get(types.KYCKey(address))
    if bz == nil {
        return nil
    }

    var record types.KYCRecord
    _ = k.cdc.Unmarshal(bz, &record)  // IGNORES ERROR!
    return &record  // Returns corrupted data
}
```

## Attack Scenarios

### 1. State Corruption Attack
```
Attacker:
1. Stores malformed protobuf data via governance proposal
2. Unmarshal silently fails or panics
3. Keeper returns corrupted/default struct
4. Chain processes invalid state
5. Consensus divergence
```

### 2. DoS via Panic
```
Attacker:
1. Triggers query with corrupted data
2. MustUnmarshal panics
3. Node crashes
4. Repeat for all nodes
5. Network halt
```

### 3. Silent Data Loss
```
User:
1. Submits valid KYC data
2. Data stored successfully
3. Later, database corruption occurs
4. Unmarshal fails silently
5. User appears non-compliant
6. Transactions blocked incorrectly
```

## Impact

- **State Corruption:** Invalid data processed as valid
- **DoS:** Panics crash nodes
- **Silent Failures:** Operations fail without error reporting
- **Consensus Divergence:** Different nodes interpret corrupted data differently
- **Data Loss:** Failed unmarshal returns default/empty structs

## Required Fix

**Proper error handling pattern:**

```go
func (k Keeper) GetKYCRecord(ctx sdk.Context, address string) (*types.KYCRecord, error) {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get(types.KYCKey(address))
    if bz == nil {
        return nil, types.ErrKYCNotFound
    }

    var record types.KYCRecord
    if err := k.cdc.Unmarshal(bz, &record); err != nil {
        // Log the error for debugging
        k.Logger(ctx).Error(
            "failed to unmarshal KYC record",
            "address", address,
            "error", err.Error(),
            "data_length", len(bz),
        )
        return nil, errorsmod.Wrapf(
            types.ErrInvalidData,
            "failed to unmarshal KYC record for %s: %v", address, err,
        )
    }

    // Validate unmarshaled data
    if err := record.ValidateBasic(); err != nil {
        return nil, errorsmod.Wrap(types.ErrInvalidKYCData, err.Error())
    }

    return &record, nil
}
```

## Codebase-Wide Fix Required

### Step 1: Find All Instances

```bash
# Find MustUnmarshal usage (panics on error)
grep -r "MustUnmarshal" chain/x/ | wc -l  # Expected: 200+

# Find ignored Unmarshal errors
grep -r "_ = .*Unmarshal" chain/x/ | wc -l

# Find Unmarshal without error check
grep -A 2 "\.Unmarshal(" chain/x/ | grep -v "if err"
```

### Step 2: Fix Each Instance

For each keeper method:
1. Change return type to include `error`
2. Replace `MustUnmarshal` with `Unmarshal`
3. Properly handle and propagate errors
4. Add error logging
5. Add post-unmarshal validation
6. Update all callers

### Step 3: Add Defensive Checks

```go
// Add to all keeper constructors:
func NewKeeper(...) *Keeper {
    k := &Keeper{...}

    // Verify codec can marshal/unmarshal types
    testTypes := []proto.Message{
        &types.KYCRecord{},
        &types.Params{},
        // ... all types
    }
    for _, t := range testTypes {
        bz, err := k.cdc.Marshal(t)
        if err != nil {
            panic(fmt.Sprintf("codec cannot marshal %T: %v", t, err))
        }
        if err := k.cdc.Unmarshal(bz, t); err != nil {
            panic(fmt.Sprintf("codec cannot unmarshal %T: %v", t, err))
        }
    }

    return k
}
```

## Testing Requirements

```go
func TestUnmarshalErrorHandling(t *testing.T) {
    tests := []struct {
        name          string
        corruptData   []byte
        expectedError string
    }{
        {
            name:          "truncated protobuf",
            corruptData:   []byte{0x08, 0x01},  // Incomplete
            expectedError: "failed to unmarshal",
        },
        {
            name:          "wrong type",
            corruptData:   marshalWrongType(),
            expectedError: "failed to unmarshal",
        },
        {
            name:          "random garbage",
            corruptData:   randomBytes(100),
            expectedError: "failed to unmarshal",
        },
        {
            name:          "empty data",
            corruptData:   []byte{},
            expectedError: "not found",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Store corrupted data
            store.Set(key, tt.corruptData)

            // Attempt to retrieve
            record, err := keeper.GetKYCRecord(ctx, address)

            // Must return error, not panic or silent failure
            require.Error(t, err)
            require.Nil(t, record)
            require.Contains(t, err.Error(), tt.expectedError)
        })
    }
}

func TestNoPanicOnCorruptedState(t *testing.T) {
    // Store corrupted data in all keeper stores
    corruptAllStores(ctx, keeper)

    // All query methods should return errors, not panic
    require.NotPanics(t, func() {
        _, err := keeper.GetKYCRecord(ctx, "addr")
        require.Error(t, err)
    })
}
```

## Acceptance Criteria

- [ ] Audit all 200+ `MustUnmarshal` calls
- [ ] Replace with proper error handling
- [ ] Add comprehensive error logging
- [ ] Update all method signatures to return errors
- [ ] Update all callers to handle errors
- [ ] Add corruption resilience tests (20+ scenarios)
- [ ] Add fuzzing tests for unmarshal paths
- [ ] Document error handling patterns

## Migration Strategy

This is a **breaking change** to keeper interfaces. Coordinate:
1. Update keeper methods (one module at a time)
2. Update message handlers
3. Update query handlers
4. Update tests
5. Update documentation

## References

- [CWE-754: Improper Check for Unusual or Exceptional Conditions](https://cwe.mitre.org/data/definitions/754.html)
- [Protocol Buffers: Error Handling](https://protobuf.dev/programming-guides/api/)
- [Cosmos SDK: Codec Usage](https://docs.cosmos.network/main/build/building-modules/encoding)

## Related Issues

- Security Audit Report: CRITICAL-003
- See also: todos/004-complete-p1-keeper-adapter-error-suppression.md

---

**PRIORITY: Fix before mainnet - State corruption risk**
