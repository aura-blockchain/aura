# Timestamp Type Conflict Fix Guide

## Problem Summary

Proto files generate different timestamp types based on annotations:
- Fields with `(gogoproto.stdtime) = true` → `time.Time` (value type)
- Fields without stdtime → `*gogotypes.Timestamp` (pointer to gogo Timestamp)
- Some proto fields use `*time.Time` (pointer to time.Time)

## Import Changes Required

Replace:
```go
"google.golang.org/protobuf/types/known/timestamppb"
```

With:
```go
gogotypes "github.com/cosmos/gogoproto/types"
```

Also ensure `"time"` is imported.

## Conversion Patterns

### Pattern 1: Creating gogotypes.Timestamp from time values

**OLD (WRONG):**
```go
timestamppb.New(time.Unix(timestamp, 0))
timestamppb.Now()
```

**NEW (CORRECT):**
```go
now := time.Now()
&gogotypes.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())}

// OR for Unix timestamp:
&gogotypes.Timestamp{Seconds: unixTimestamp, Nanos: 0}
```

### Pattern 2: Converting gogotypes.Timestamp to time.Time

**OLD (WRONG):**
```go
timestamp.AsTime()
```

**NEW (CORRECT):**
```go
time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
```

### Pattern 3: Converting gogotypes.Duration to time.Duration

**OLD (WRONG):**
```go
duration.AsDuration()
```

**NEW (CORRECT):**
```go
time.Duration(duration.Seconds)*time.Second + time.Duration(duration.Nanos)*time.Nanosecond
```

### Pattern 4: Nil checks for time.Time fields

**OLD (WRONG):**
```go
if timestamp == nil {  // time.Time is not a pointer
}
```

**NEW (CORRECT):**
```go
if timestamp.IsZero() {
}
```

### Pattern 5: Nil checks for *time.Time fields

**OLD (WRONG):**
```go
if timestamp.AsTime() // *time.Time doesn't have AsTime
```

**NEW (CORRECT):**
```go
if timestamp != nil && timestamp.Before(someTime) {
}
```

### Pattern 6: Context conversion for Logger

**OLD (WRONG):**
```go
k.Logger(ctx)  // where ctx is context.Context
```

**NEW (CORRECT):**
```go
sdkCtx := sdk.UnwrapSDKContext(ctx)
k.Logger(sdkCtx)
```

## Files Requiring Fixes

### Completed:
- ✅ x/vcregistry/keeper/keeper.go
- ✅ x/vcregistry/keeper/invariants.go
- ✅ x/auth/types/types.go
- ✅ x/cryptography/keeper/advanced_crypto.go
- ✅ x/cryptography/keeper/cert_pinning.go
- ✅ x/governance/keeper/msg_server.go (partial)

### Remaining:
- ❌ x/vcregistry/keeper/minting.go (partial)
- ❌ x/vcregistry/keeper/msg_server.go
- ❌ x/cryptography/keeper/msg_server.go
- ❌ x/economicsecurity/keeper/governance.go
- ❌ x/identity/types/types.go
- ❌ x/incidentresponse/keeper/msg_server_proto.go
- ❌ x/incidentresponse/keeper/query_server_proto.go
- ❌ x/contractregistry/keeper/invariants.go
- ❌ x/contractregistry/keeper/keeper.go

## Automated Fix Script

For each file:

1. Add import: `gogotypes "github.com/cosmos/gogoproto/types"`
2. Remove import: `"google.golang.org/protobuf/types/known/timestamppb"`
3. Replace all `timestamppb.Now()` with appropriate gogotypes.Timestamp creation
4. Replace all `timestamppb.New(...)` with gogotypes.Timestamp creation
5. Replace all `.AsTime()` calls with manual conversion
6. Replace all `.AsDuration()` calls with manual conversion
7. Fix nil checks on time.Time fields to use `.IsZero()`
8. Fix Logger calls with context.Context to unwrap first

## Example Complete Fix

```go
// BEFORE
func (k *Keeper) CreateRecord(ctx context.Context) error {
    record := &types.Record{
        CreatedAt: timestamppb.Now(),
        ExpiresAt: timestamppb.New(time.Now().Add(24*time.Hour)),
    }

    if record.ExpiresAt != nil && record.CreatedAt.AsTime().Before(record.ExpiresAt.AsTime()) {
        k.Logger(ctx).Info("valid record")
    }

    return nil
}

// AFTER
func (k *Keeper) CreateRecord(ctx context.Context) error {
    now := time.Now()
    expiryTime := now.Add(24*time.Hour)

    record := &types.Record{
        CreatedAt: &gogotypes.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())},
        ExpiresAt: &gogotypes.Timestamp{Seconds: expiryTime.Unix(), Nanos: int32(expiryTime.Nanosecond())},
    }

    createdTime := time.Unix(record.CreatedAt.Seconds, int64(record.CreatedAt.Nanos))
    expiresTime := time.Unix(record.ExpiresAt.Seconds, int64(record.ExpiresAt.Nanos))

    if record.ExpiresAt != nil && createdTime.Before(expiresTime) {
        sdkCtx := sdk.UnwrapSDKContext(ctx)
        k.Logger(sdkCtx).Info("valid record")
    }

    return nil
}
```

## Testing

After fixes, run:
```bash
go build ./...
```

All timestamp-related errors should be resolved.
