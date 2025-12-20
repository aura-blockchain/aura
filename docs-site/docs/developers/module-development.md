---
sidebar_position: 2
---

# Module Development

Learn how to develop custom modules for the Aura blockchain using the Cosmos SDK framework.

## Module Structure

Standard Aura module structure:

```
x/mymodule/
├── client/cli/           # CLI commands
├── keeper/               # Business logic
│   ├── keeper.go
│   ├── msg_server.go
│   └── query_server.go
├── types/                # Type definitions
│   ├── codec.go
│   ├── errors.go
│   ├── msgs.go
│   └── genesis.go
└── module.go             # Module interface
```

## Creating a Module

### 1. Define Protocol Buffers

Create proto definitions in `proto/aura/mymodule/v1beta1/`:

```protobuf
syntax = "proto3";
package aura.mymodule.v1beta1;

message MyData {
  string id = 1;
  string owner = 2;
  string value = 3;
}
```

### 2. Implement Keeper

The Keeper manages state and business logic:

```go
type Keeper struct {
    cdc      codec.BinaryCodec
    storeKey storetypes.StoreKey
}

func (k Keeper) SetData(ctx sdk.Context, data types.MyData) error {
    store := ctx.KVStore(k.storeKey)
    bz := k.cdc.MustMarshal(&data)
    store.Set(types.DataKey(data.Id), bz)
    return nil
}
```

### 3. Define Messages

Messages represent transactions:

```go
type MsgCreateData struct {
    Creator string
    Value   string
}

func (msg *MsgCreateData) ValidateBasic() error {
    _, err := sdk.AccAddressFromBech32(msg.Creator)
    if err != nil {
        return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address")
    }
    return nil
}
```

### 4. Implement Message Server

Handle message execution:

```go
func (k msgServer) CreateData(goCtx context.Context, msg *types.MsgCreateData) (*types.MsgCreateDataResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    data := types.MyData{
        Id:    uuid.New().String(),
        Owner: msg.Creator,
        Value: msg.Value,
    }

    if err := k.SetData(ctx, data); err != nil {
        return nil, err
    }

    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeCreateData,
            sdk.NewAttribute(types.AttributeKeyID, data.Id),
        ),
    )

    return &types.MsgCreateDataResponse{Id: data.Id}, nil
}
```

## Testing

### Unit Tests

```go
func TestKeeper_SetData(t *testing.T) {
    keeper, ctx := setupKeeper(t)

    data := types.MyData{
        Id:    "test-1",
        Owner: "aura1...",
        Value: "test-value",
    }

    err := keeper.SetData(ctx, data)
    require.NoError(t, err)

    retrieved, found := keeper.GetData(ctx, data.Id)
    require.True(t, found)
    require.Equal(t, data, retrieved)
}
```

### Integration Tests

```go
func TestIntegration_CreateData(t *testing.T) {
    app := simapp.Setup(t, false)
    ctx := app.BaseApp.NewContext(false, tmproto.Header{})

    msgServer := keeper.NewMsgServerImpl(app.MyModuleKeeper)

    msg := &types.MsgCreateData{
        Creator: "aura1...",
        Value:   "test",
    }

    res, err := msgServer.CreateData(sdk.WrapSDKContext(ctx), msg)
    require.NoError(t, err)
    require.NotEmpty(t, res.Id)
}
```

## Best Practices

- Use proper error handling with typed errors
- Emit events for all state changes
- Validate all inputs in `ValidateBasic()`
- Write comprehensive tests (>80% coverage)
- Document public APIs
- Follow Cosmos SDK coding standards

## Resources

- [Cosmos SDK Documentation](https://docs.cosmos.network)
- [Aura Module Examples](https://github.com/aura-blockchain/aura/tree/main/x)
- [Proto Style Guide](https://github.com/aura-blockchain/aura/blob/main/docs/proto/STYLE_GUIDE.md)

For complete details, see the [full Module Development Guide](https://github.com/aura-blockchain/aura/blob/main/docs/developers/MODULE_DEVELOPMENT_GUIDE.md) in the repository.
