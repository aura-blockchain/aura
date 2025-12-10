# Gogoproto Type Mappings

**CRITICAL: Read this before writing ANY test code that uses proto-generated types.**

This project uses gogoproto annotations in `.proto` files that transform generated Go types. Using the wrong types in tests causes build failures that are time-consuming to fix.

## Quick Reference

| Proto Type | Gogoproto Annotation | Generated Go Type | WRONG Usage | CORRECT Usage |
|------------|---------------------|-------------------|-------------|---------------|
| `google.protobuf.Timestamp` | `(gogoproto.stdtime) = true` | `time.Time` | `timestamppb.New(t)` | `time.Now()` |
| `google.protobuf.Timestamp` | `(gogoproto.stdtime) = true` (nullable) | `*time.Time` | `timestamppb.Now()` | `&now` where `now := time.Now()` |
| `google.protobuf.Duration` | `(gogoproto.stdduration) = true` | `time.Duration` | `durationpb.New(d)` | `30 * time.Second` |
| `string` | `(gogoproto.customtype) = "cosmossdk.io/math.Int"` | `math.Int` | `"1000"` | `sdkmath.NewInt(1000)` |
| `string` | `(gogoproto.customtype) = "cosmossdk.io/math.LegacyDec"` | `math.LegacyDec` | `"0.5"` | `sdkmath.LegacyMustNewDecFromStr("0.5")` |
| `repeated Message` | `(gogoproto.nullable) = false` | `[]Message` | `[]*Message{}` | `[]Message{}` |
| `Message field` | `(gogoproto.nullable) = false` | `Message` | `&Message{}` | `Message{}` |
| `cosmos.base.v1beta1.Coin` | `(gogoproto.nullable) = false` | `sdk.Coin` | `&sdk.Coin{}` or `nil` | `sdk.Coin{}` |

## Common Mistakes

### 1. Timestamps
```go
// WRONG - timestamppb types
import "google.golang.org/protobuf/types/known/timestamppb"
CreatedAt: timestamppb.New(time.Now()),  // ERROR!
ExpiresAt: timestamppb.Now(),            // ERROR!

// CORRECT - native Go time
CreatedAt: time.Now(),                   // For non-nullable fields
expiresAt := time.Now().Add(24*time.Hour)
ExpiresAt: &expiresAt,                   // For nullable fields (pointer)
```

### 2. Math Types
```go
// WRONG - string values
Amount: "1000000",                       // ERROR!
Fee: "0.05",                             // ERROR!

// CORRECT - math types
import sdkmath "cosmossdk.io/math"
Amount: sdkmath.NewInt(1000000),
Fee: sdkmath.LegacyMustNewDecFromStr("0.05"),
```

### 3. Pointer vs Value
```go
// WRONG - pointer when value expected
Params: &types.Params{...},              // ERROR if nullable=false
Coins: []*sdk.Coin{...},                 // ERROR if nullable=false

// CORRECT - value types
Params: types.Params{...},               // Value, not pointer
Coins: []sdk.Coin{...},                  // Value slice, not pointer slice
```

### 4. Nil vs Zero Value
```go
// WRONG - nil for value types
Amount: nil,                             // ERROR! Can't be nil
CreatedAt: nil,                          // ERROR! Can't be nil

// CORRECT - zero values
Amount: sdk.Coin{},                      // Empty/zero value
CreatedAt: time.Time{},                  // Zero time
```

### 5. Comparing Math Types
```go
// WRONG - direct comparison
if fee != expectedFee { ... }            // ERROR! Struct comparison fails

// CORRECT - use Equal method
if !fee.Equal(expectedFee) { ... }       // Proper comparison
```

## How to Check Proto Annotations

Before writing tests, check the `.proto` file:

```bash
# Find the proto file
find proto/ -name "*.proto" | xargs grep -l "YourMessageType"

# Check annotations
grep -A2 "field_name" proto/aura/module/v1beta1/types.proto
```

Look for these annotations:
- `(gogoproto.stdtime) = true` → use `time.Time`
- `(gogoproto.stdduration) = true` → use `time.Duration`
- `(gogoproto.nullable) = false` → use value type, not pointer
- `(gogoproto.customtype) = "..."` → use the specified Go type

## Test Helpers

Use the helpers in `chain/testutil/proto_helpers.go`:

```go
import "github.com/aequitas/aura/chain/testutil"

// Create properly-typed timestamps
createdAt := testutil.Now()           // time.Time
expiresAt := testutil.NowPtr()        // *time.Time
pastTime := testutil.TimeAgo(1*time.Hour)

// Create properly-typed math values
amount := testutil.NewInt(1000000)    // math.Int
fee := testutil.NewDec("0.05")        // math.LegacyDec

// Create coins
coin := testutil.NewCoin("uaura", 1000000)  // sdk.Coin (value, not pointer)
```

## Import Aliases

Use consistent import aliases:
```go
import (
    sdkmath "cosmossdk.io/math"
    sdk "github.com/cosmos/cosmos-sdk/types"
)
```

**DO NOT import** (unless specifically needed for non-proto work):
```go
// Avoid these for proto-generated types
"google.golang.org/protobuf/types/known/timestamppb"
"google.golang.org/protobuf/types/known/durationpb"
```
