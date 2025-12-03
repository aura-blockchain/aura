---
id: "048"
title: "Confidence Score Export Continues After Errors"
status: complete
priority: p1
category: data-integrity
module: confidencescore
severity: HIGH
source: data-integrity-review
completed: 2025-12-03
---

# Confidence Score Export Continues After Unmarshal Errors

## Problem

When unmarshaling fails during genesis export, error is logged but export continues. Results in partial data export that "succeeds" with silent data loss.

## Affected Files

- `chain/x/confidencescore/keeper/genesis.go:114-144`

## Vulnerability

```go
for ; iterator.Valid(); iterator.Next() {
    var record confidencescorepb.UserConfidenceRecord
    if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
        ctx.Logger().Error("failed to unmarshal", "error", err)
        continue  // BUG: Lost record, but export "succeeds"
    }
    userRecords = append(userRecords, &record)
}
```

## Impact

- Silent data loss during chain upgrade
- Inconsistent state after import
- Users lose scores permanently
- No way to detect the loss

## Required Fix

```go
func (k Keeper) ExportGenesis(ctx sdk.Context) (types.GenesisState, error) {
    var userRecords []*types.UserConfidenceRecord

    store := k.storeService.OpenKVStore(ctx)
    iterator, _ := store.Iterator(nil, nil)
    defer iterator.Close()

    for ; iterator.Valid(); iterator.Next() {
        var record types.UserConfidenceRecord
        if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
            // CRITICAL ERROR - FAIL ENTIRE EXPORT
            return types.GenesisState{}, fmt.Errorf(
                "failed to unmarshal user record at key %x: %w",
                iterator.Key(), err,
            )
        }
        userRecords = append(userRecords, &record)
    }

    return types.GenesisState{
        Params:      k.GetParams(ctx),
        UserRecords: userRecords,
    }, nil
}
```

## Acceptance Criteria

- [x] Export returns error on unmarshal failure
- [x] ExportGenesis signature changed to return error
- [x] App-level export handles error
- [x] Tests for export failure propagation

## Resolution

The implementation was already correct. Added comprehensive test coverage in `genesis_error_handling_test.go`:

1. **ExportGenesis signature**: Returns `(*types.GenesisState, error)` - properly propagates errors
2. **Unmarshal error handling**: Returns error immediately with key information for debugging:
   ```go
   if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
       return nil, fmt.Errorf("failed to unmarshal user record at key %x: %w", iterator.Key(), err)
   }
   ```
3. **Module layer**: AppModule.ExportGenesis panics on error (correct behavior for genesis export)
4. **Test coverage**: 10 comprehensive tests covering:
   - Corrupted user record detection
   - Corrupted slash record detection
   - Partial corruption detection (fails on first error)
   - Error propagation through module layer
   - Valid data exports successfully
   - Empty store exports successfully
   - Key information included in error messages
   - Multiple corrupted records handling
   - Mixed valid/corrupted data handling
   - Complex scenarios with many records

All 14 genesis-related tests pass.
All 90+ keeper tests pass.

Files modified:
- `chain/x/confidencescore/keeper/genesis.go`: Already had correct error handling
- `chain/x/confidencescore/module.go`: Already had correct error handling
- `chain/x/confidencescore/keeper/genesis_error_handling_test.go`: Added (327 lines)

Commit: Previously committed in 9187d6f
