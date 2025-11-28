# Module Development Guide

This guide explains how to develop custom modules for the AURA blockchain following Cosmos SDK best practices.

## Table of Contents

- [Module Structure](#module-structure)
- [Creating a New Module](#creating-a-new-module)
- [Keeper Pattern](#keeper-pattern)
- [Message Handlers](#message-handlers)
- [Query Handlers](#query-handlers)
- [Events](#events)
- [Testing](#testing)
- [Best Practices](#best-practices)

## Module Structure

Standard AURA module structure:

```
x/mymodule/
├── client/
│   └── cli/
│       ├── query.go      # CLI query commands
│       └── tx.go         # CLI transaction commands
├── keeper/
│   ├── keeper.go         # Core business logic
│   ├── msg_server.go     # Message handler implementation
│   ├── query_server.go   # Query handler implementation
│   └── keeper_test.go    # Unit tests
├── types/
│   ├── codec.go          # Amino/protobuf codecs
│   ├── errors.go         # Module-specific errors
│   ├── events.go         # Event type definitions
│   ├── genesis.go        # Genesis state
│   ├── keys.go           # Store keys and prefixes
│   ├── msgs.go           # Message type definitions
│   ├── params.go         # Module parameters
│   └── types.go          # Core type definitions
├── doc.go                # Package documentation
├── module.go             # Module implementation
└── README.md             # Module documentation
```

## Creating a New Module

### Step 1: Define Protocol Buffers

Create `proto/aura/mymodule/v1beta1/mymodule.proto`:

```protobuf
syntax = "proto3";
package aura.mymodule.v1beta1;

import "gogoproto/gogo.proto";
import "google/protobuf/timestamp.proto";

option go_package = "github.com/aequitas/aura/proto/aura/mymodule/v1beta1";

// MyData represents the core data structure
message MyData {
  string id = 1;
  string owner = 2;
  string value = 3;
  google.protobuf.Timestamp created_at = 4 [(gogoproto.stdtime) = true];
}

// Params defines module parameters
message Params {
  option (gogoproto.goproto_stringer) = false;

  uint64 max_data_size = 1;
  string fee_multiplier = 2 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable) = false
  ];
}

// GenesisState represents the initial state
message GenesisState {
  Params params = 1 [(gogoproto.nullable) = false];
  repeated MyData data_list = 2;
}
```

### Step 2: Define Messages

Create `proto/aura/mymodule/v1beta1/tx.proto`:

```protobuf
syntax = "proto3";
package aura.mymodule.v1beta1;

import "gogoproto/gogo.proto";
import "cosmos/msg/v1/msg.proto";

option go_package = "github.com/aequitas/aura/proto/aura/mymodule/v1beta1";

service Msg {
  rpc CreateData(MsgCreateData) returns (MsgCreateDataResponse);
  rpc UpdateData(MsgUpdateData) returns (MsgUpdateDataResponse);
  rpc DeleteData(MsgDeleteData) returns (MsgDeleteDataResponse);
}

message MsgCreateData {
  option (cosmos.msg.v1.signer) = "creator";

  string creator = 1;
  string value = 2;
}

message MsgCreateDataResponse {
  string id = 1;
}

message MsgUpdateData {
  option (cosmos.msg.v1.signer) = "creator";

  string creator = 1;
  string id = 2;
  string value = 3;
}

message MsgUpdateDataResponse {}

message MsgDeleteData {
  option (cosmos.msg.v1.signer) = "creator";

  string creator = 1;
  string id = 2;
}

message MsgDeleteDataResponse {}
```

### Step 3: Define Queries

Create `proto/aura/mymodule/v1beta1/query.proto`:

```protobuf
syntax = "proto3";
package aura.mymodule.v1beta1;

import "google/api/annotations.proto";
import "cosmos/base/query/v1beta1/pagination.proto";
import "aura/mymodule/v1beta1/mymodule.proto";

option go_package = "github.com/aequitas/aura/proto/aura/mymodule/v1beta1";

service Query {
  rpc Params(QueryParamsRequest) returns (QueryParamsResponse) {
    option (google.api.http).get = "/aura/mymodule/v1beta1/params";
  }

  rpc GetData(QueryGetDataRequest) returns (QueryGetDataResponse) {
    option (google.api.http).get = "/aura/mymodule/v1beta1/data/{id}";
  }

  rpc ListData(QueryListDataRequest) returns (QueryListDataResponse) {
    option (google.api.http).get = "/aura/mymodule/v1beta1/data";
  }
}

message QueryParamsRequest {}
message QueryParamsResponse {
  Params params = 1;
}

message QueryGetDataRequest {
  string id = 1;
}
message QueryGetDataResponse {
  MyData data = 1;
}

message QueryListDataRequest {
  cosmos.base.query.v1beta1.PageRequest pagination = 1;
}
message QueryListDataResponse {
  repeated MyData data = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

### Step 4: Generate Code

```bash
cd proto
buf generate
```

### Step 5: Implement the Keeper

Create `x/mymodule/keeper/keeper.go`:

```go
package keeper

import (
    "fmt"

    "github.com/cosmos/cosmos-sdk/codec"
    storetypes "cosmossdk.io/store/types"
    sdk "github.com/cosmos/cosmos-sdk/types"

    "github.com/aequitas/aura/chain/x/mymodule/types"
)

// Keeper maintains the state for the mymodule module
type Keeper struct {
    cdc      codec.BinaryCodec
    storeKey storetypes.StoreKey
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
    cdc codec.BinaryCodec,
    storeKey storetypes.StoreKey,
) *Keeper {
    return &Keeper{
        cdc:      cdc,
        storeKey: storeKey,
    }
}

// SetData stores MyData
func (k Keeper) SetData(ctx sdk.Context, data *types.MyData) error {
    if data == nil || data.Id == "" {
        return fmt.Errorf("invalid data")
    }

    store := ctx.KVStore(k.storeKey)
    bz := k.cdc.MustMarshal(data)
    store.Set(types.DataKey(data.Id), bz)
    return nil
}

// GetData retrieves MyData by ID
func (k Keeper) GetData(ctx sdk.Context, id string) (*types.MyData, error) {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get(types.DataKey(id))
    if bz == nil {
        return nil, fmt.Errorf("data not found: %s", id)
    }

    var data types.MyData
    k.cdc.MustUnmarshal(bz, &data)
    return &data, nil
}

// DeleteData removes MyData
func (k Keeper) DeleteData(ctx sdk.Context, id string) {
    store := ctx.KVStore(k.storeKey)
    store.Delete(types.DataKey(id))
}
```

### Step 6: Implement Message Server

Create `x/mymodule/keeper/msg_server.go`:

```go
package keeper

import (
    "context"
    "time"

    sdk "github.com/cosmos/cosmos-sdk/types"
    "google.golang.org/protobuf/types/known/timestamppb"

    "github.com/aequitas/aura/chain/x/mymodule/types"
)

type msgServer struct {
    Keeper
}

// NewMsgServerImpl returns an implementation of the module's MsgServer
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
    return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateData handles MsgCreateData
func (m msgServer) CreateData(goCtx context.Context, msg *types.MsgCreateData) (*types.MsgCreateDataResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Validate message
    if err := msg.ValidateBasic(); err != nil {
        return nil, err
    }

    // Generate ID
    id := fmt.Sprintf("data-%d", ctx.BlockHeight())

    // Create data
    data := &types.MyData{
        Id:        id,
        Owner:     msg.Creator,
        Value:     msg.Value,
        CreatedAt: timestamppb.New(ctx.BlockTime()),
    }

    // Store data
    if err := m.SetData(ctx, data); err != nil {
        return nil, err
    }

    // Emit event
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeDataCreated,
            sdk.NewAttribute(types.AttributeKeyID, id),
            sdk.NewAttribute(types.AttributeKeyOwner, msg.Creator),
        ),
    )

    return &types.MsgCreateDataResponse{Id: id}, nil
}
```

### Step 7: Implement Query Server

Create `x/mymodule/keeper/query_server.go`:

```go
package keeper

import (
    "context"

    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/query"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/aequitas/aura/chain/x/mymodule/types"
)

type queryServer struct {
    Keeper
}

// NewQueryServerImpl returns an implementation of the module's QueryServer
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
    return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Params queries module parameters
func (q queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

    ctx := sdk.UnwrapSDKContext(goCtx)
    params := q.GetParams(ctx)

    return &types.QueryParamsResponse{Params: params}, nil
}

// GetData queries a single data item by ID
func (q queryServer) GetData(goCtx context.Context, req *types.QueryGetDataRequest) (*types.QueryGetDataResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

    ctx := sdk.UnwrapSDKContext(goCtx)
    data, err := q.Keeper.GetData(ctx, req.Id)
    if err != nil {
        return nil, status.Error(codes.NotFound, err.Error())
    }

    return &types.QueryGetDataResponse{Data: data}, nil
}
```

### Step 8: Implement Module Interface

Create `x/mymodule/module.go`:

```go
package mymodule

import (
    "context"
    "encoding/json"

    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/spf13/cobra"

    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/codec"
    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/module"

    "github.com/aequitas/aura/chain/x/mymodule/client/cli"
    "github.com/aequitas/aura/chain/x/mymodule/keeper"
    "github.com/aequitas/aura/chain/x/mymodule/types"
)

var (
    _ module.AppModuleBasic = AppModuleBasic{}
    _ module.AppModule      = AppModule{}
)

// AppModuleBasic defines the basic application module
type AppModuleBasic struct {
    cdc codec.Codec
}

// Name returns the module's name
func (AppModuleBasic) Name() string {
    return types.ModuleName
}

// RegisterLegacyAminoCodec registers the module's types with the given codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
    types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (a AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
    types.RegisterInterfaces(reg)
}

// DefaultGenesis returns default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
    return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
    var genState types.GenesisState
    if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
        return err
    }
    return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
    types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetTxCmd returns the root tx command
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
    return cli.GetTxCmd()
}

// GetQueryCmd returns the root query command
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
    return cli.GetQueryCmd()
}

// AppModule implements the AppModule interface
type AppModule struct {
    AppModuleBasic

    keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
    return AppModule{
        AppModuleBasic: AppModuleBasic{cdc: cdc},
        keeper:         keeper,
    }
}

// Name returns the module's name
func (am AppModule) Name() string {
    return am.AppModuleBasic.Name()
}

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
    types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
    types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs genesis initialization
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
    var genState types.GenesisState
    cdc.MustUnmarshalJSON(data, &genState)
    InitGenesis(ctx, am.keeper, genState)
    return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
    genState := ExportGenesis(ctx, am.keeper)
    return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }
```

## Keeper Pattern

The Keeper pattern is central to Cosmos SDK module development:

### Key Principles

1. **Single Responsibility**: Each keeper manages one module's state
2. **Dependency Injection**: Keepers receive dependencies via constructor
3. **Interface Segregation**: Expose only necessary methods
4. **Immutability**: Context is read-only in queries

### Example Keeper Structure

```go
type Keeper struct {
    // Dependencies
    cdc           codec.BinaryCodec
    storeKey      storetypes.StoreKey
    paramSpace    paramtypes.Subspace
    bankKeeper    types.BankKeeper
    accountKeeper types.AccountKeeper

    // Security
    reentrancyGuard *security.ReentrancyGuard
    pauseGuard      *security.PauseGuard
}
```

## Message Handlers

### ValidateBasic Implementation

```go
func (msg *MsgCreateData) ValidateBasic() error {
    // Validate creator address
    if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
        return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address: %s", err)
    }

    // Validate value
    if len(msg.Value) == 0 {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "value cannot be empty")
    }

    if len(msg.Value) > 1024 {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "value too long")
    }

    return nil
}
```

### Message Handler Best Practices

```go
func (m msgServer) CreateData(goCtx context.Context, msg *MsgCreateData) (*types.MsgCreateDataResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // 1. Security checks
    if err := m.pauseGuard.RequireNotPaused(ctx); err != nil {
        return nil, err
    }

    if err := m.reentrancyGuard.Enter(ctx); err != nil {
        return nil, err
    }
    defer m.reentrancyGuard.Exit(ctx)

    // 2. Input validation
    if err := msg.ValidateBasic(); err != nil {
        return nil, err
    }

    // 3. Permission checks
    if !m.HasPermission(ctx, msg.Creator, "create_data") {
        return nil, types.ErrUnauthorized
    }

    // 4. Business logic
    data := &types.MyData{
        Id:    generateID(),
        Owner: msg.Creator,
        Value: msg.Value,
    }

    // 5. State changes
    if err := m.SetData(ctx, data); err != nil {
        return nil, err
    }

    // 6. Emit events
    ctx.EventManager().EmitEvents(sdk.Events{
        sdk.NewEvent(
            types.EventTypeDataCreated,
            sdk.NewAttribute(types.AttributeKeyID, data.Id),
        ),
    })

    return &types.MsgCreateDataResponse{Id: data.Id}, nil
}
```

## Events

Define events in `types/events.go`:

```go
const (
    EventTypeDataCreated = "data_created"
    EventTypeDataUpdated = "data_updated"
    EventTypeDataDeleted = "data_deleted"

    AttributeKeyID    = "id"
    AttributeKeyOwner = "owner"
    AttributeKeyValue = "value"
)
```

Emit events:

```go
ctx.EventManager().EmitEvent(
    sdk.NewEvent(
        EventTypeDataCreated,
        sdk.NewAttribute(AttributeKeyID, id),
        sdk.NewAttribute(AttributeKeyOwner, owner),
        sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
    ),
)
```

## Testing

### Unit Test Example

```go
func TestCreateData(t *testing.T) {
    // Setup
    k, ctx := setupKeeper(t)
    msgServer := keeper.NewMsgServerImpl(k)

    // Test case
    msg := &types.MsgCreateData{
        Creator: "aura1test...",
        Value:   "test value",
    }

    // Execute
    resp, err := msgServer.CreateData(sdk.WrapSDKContext(ctx), msg)

    // Assert
    require.NoError(t, err)
    require.NotEmpty(t, resp.Id)

    // Verify state
    data, err := k.GetData(ctx, resp.Id)
    require.NoError(t, err)
    require.Equal(t, msg.Value, data.Value)
}
```

## Best Practices

1. **Always validate inputs** in both ValidateBasic and handlers
2. **Use deterministic operations** - avoid randomness in state machines
3. **Emit comprehensive events** for indexing and monitoring
4. **Write extensive tests** - aim for >80% coverage
5. **Document all exported functions** with godoc comments
6. **Handle errors gracefully** - never panic in production code
7. **Use security guards** - reentrancy, pause, access control
8. **Optimize gas usage** - minimize storage operations
9. **Version your APIs** - use semantic versioning
10. **Follow Cosmos SDK conventions** - consistency matters

## Next Steps

- [Testing Guide](../testing/TESTING_GUIDE.md)
- [Integration Guide](INTEGRATION_GUIDE.md)
- [Security Best Practices](../SECURITY_BEST_PRACTICES.md)
- [Performance Optimization](PERFORMANCE_OPTIMIZATION.md)
