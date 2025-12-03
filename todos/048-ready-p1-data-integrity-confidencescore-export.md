---
id: "048"
title: "Confidence Score Export Continues After Errors"
status: ready
priority: p1
category: data-integrity
module: confidencescore
severity: HIGH
source: data-integrity-review
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

- [ ] Export returns error on unmarshal failure
- [ ] ExportGenesis signature changed to return error
- [ ] App-level export handles error
- [ ] Tests for export failure propagation
