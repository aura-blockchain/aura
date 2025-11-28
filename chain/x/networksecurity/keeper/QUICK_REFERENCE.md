# NetworkSecurity Keeper Test Quick Reference

## Fixed Build Errors

### Error 1: undefined: KeeperTestSuite
**Files affected:** `invariants_test.go`, `query_server_test.go`

**Solution:** Created `/home/decri/blockchain-projects/aura/chain/x/networksecurity/keeper/test_suite.go` with `KeeperTestSuite` definition for `package keeper` tests.

### Error 2: Params type mismatch
**File:** `genesis_test.go:30`
```go
// Before (ERROR):
Params: *types.DefaultParams()  // DefaultParams() already returns *Params

// After (FIXED):
Params: types.DefaultParams()   // Use pointer directly
```

### Error 3: TrustedPeers slice type mismatch
**File:** `genesis_test.go:31`
```go
// Before (ERROR):
TrustedPeers: []types.TrustedPeer{...}

// After (FIXED):
TrustedPeers: []*types.TrustedPeer{...}  // Pointer slice
```

### Error 4: PublicKey type mismatch
**File:** `genesis_test.go:34`
```go
// Before (ERROR):
PublicKey: "pubkey1"  // String

// After (FIXED):
PublicKey: []byte("pubkey1")  // []byte
```

### Error 5: AddedAt type mismatch
**File:** `genesis_test.go:36`
```go
// Before (ERROR):
AddedAt: 1000  // int

// After (FIXED):
AddedAt: timestamppb.New(time.Unix(1000, 0))  // *timestamppb.Timestamp
```

### Error 6: Unknown field Active
**File:** `genesis_test.go:37`
```go
// Before (ERROR):
Active: true  // Field doesn't exist in proto

// After (FIXED):
// Removed - field not in proto definition
```

## Test Suite Structure

### For `package keeper` tests:
```go
import "github.com/aequitas/aura/chain/x/networksecurity/keeper"

type MyTestSuite struct {
    keeper.KeeperTestSuite  // Embed the base test suite
}

func (suite *MyTestSuite) TestSomething() {
    // Access keeper: suite.Keeper
    // Access context: suite.SdkCtx
    // Access codec: suite.Cdc
}
```

### For `package keeper_test` tests:
```go
// Use the existing KeeperTestSuite in keeper_test.go
type MyTestSuite struct {
    keeper_test.KeeperTestSuite
}
```

## Proto Types Cheat Sheet

### TrustedPeer
```go
&types.TrustedPeer{
    PeerId:      "peer1",
    Address:     "192.168.1.1",              // Required
    PublicKey:   []byte("key"),              // []byte, not string
    Description: "Description",
    AddedAt:     timestamppb.New(time.Now()), // *timestamppb.Timestamp
}
```

### NodeReputation
```go
&types.NodeReputation{
    PeerId:            "peer1",
    Score:             75,
    LastUpdatedHeight: 1000,              // Not LastUpdateTime
    MessagesReceived:  100,               // Not MessagesSent
    ValidMessages:     95,                // Not MessagesValid
    InvalidMessages:   5,                 // Required
    Uptime:            3600,
    MisbehaviorCount:  0,                 // Required
}
```

### RateLimitEntry
```go
&types.RateLimitEntry{
    PeerId:        "peer1",
    RequestCount:  50,                           // Not TokensUsed
    WindowStart:   timestamppb.New(time.Now()),  // *timestamppb.Timestamp
    IsBanned:      false,                        // Not IsBlocked
    BanExpiresAt:  timestamppb.New(time.Now()),  // Optional
    BytesSent:     1024,
    BytesReceived: 2048,
}
```

### ForkAlert
```go
&types.ForkAlert{
    AlertId:           "fork1",
    BlockHeight:       100,                         // Not ForkHeight
    ChainAHash:        []byte("hash_a"),
    ChainBHash:        []byte("hash_b"),
    DetectedAt:        timestamppb.New(time.Now()),
    Resolved:          false,
    ResolutionDetails: "",
}
```

### PartitionAlert
```go
&types.PartitionAlert{
    AlertId:        "partition1",
    ConnectedPeers: 2,
    ExpectedPeers:  10,
    MissingPeerIds: []string{"peer3", "peer4"},
    DetectedAt:     timestamppb.New(time.Now()),
    Resolved:       false,
}
```

## Running Tests

```bash
# Run all keeper tests
go test ./x/networksecurity/keeper/...

# Run specific test
go test ./x/networksecurity/keeper -run TestInitGenesis

# Run with verbose output
go test ./x/networksecurity/keeper/... -v

# Build tests only (no run)
go test -c ./x/networksecurity/keeper
```

## Common Patterns

### Creating test genesis state
```go
genesis := &types.GenesisState{
    Params:          types.DefaultParams(),
    TrustedPeers:    []*types.TrustedPeer{},
    Reputations:     []*types.NodeReputation{},
    RateLimits:      []*types.RateLimitEntry{},
    ForkAlerts:      []*types.ForkAlert{},
    PartitionAlerts: []*types.PartitionAlert{},
}
```

### Setting params
```go
// DefaultParams returns a pointer, but SetParams takes a value
params := types.DefaultParams()
err := keeper.SetParams(ctx, *params)
```

### Getting params
```go
params, err := keeper.GetParams(ctx)
// params is a value, not a pointer
```
