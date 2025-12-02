package app

import (
	"encoding/json"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/protobuf/types/known/emptypb"
)

// adapterModule wraps an Aura module that exposes InitGenesis/BeginBlock/EndBlock/RegisterServices
// so it can satisfy the sdkmodule.AppModule interface expected by the SDK module.Manager.
type adapterModule struct {
	name   string
	module interface{}
}

// Name implements sdkmodule.AppModuleBasic.
func (a adapterModule) Name() string { return a.name }

// RegisterGRPCGatewayRoutes is a no-op; Aura modules handle GRPC registration via RegisterServices.
func (adapterModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// RegisterInterfaces is a no-op; Aura modules self-register interfaces elsewhere.
func (adapterModule) RegisterInterfaces(codectypes.InterfaceRegistry) {}

// RegisterLegacyAminoCodec is a no-op; Aura modules use protobuf.
func (adapterModule) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// DefaultGenesis returns the wrapped module's default genesis if available; otherwise empty genesis.
func (a adapterModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	if g, ok := a.module.(interface {
		DefaultGenesis(cdc codec.JSONCodec) json.RawMessage
	}); ok {
		return g.DefaultGenesis(cdc)
	}
	return cdc.MustMarshalJSON(&emptypb.Empty{})
}

// ValidateGenesis is a no-op.
func (adapterModule) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, _ json.RawMessage) error {
	return nil
}

// RegisterServices delegates to the Aura module if it supports appmodule.HasServices.
func (a adapterModule) RegisterServices(cfg sdkmodule.Configurator) {
	if srv, ok := a.module.(interface{ RegisterServices(sdkmodule.Configurator) }); ok {
		srv.RegisterServices(cfg)
		return
	}
	if legacy, ok := a.module.(interface{ RegisterServices(interface{}) error }); ok {
		_ = legacy.RegisterServices(cfg)
	}
}

// InitGenesis delegates to the Aura module if it supports appmodule.HasGenesis.
func (a adapterModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) ([]abci.ValidatorUpdate, error) {
	switch mod := a.module.(type) {
	case interface {
		InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage)
	}:
		mod.InitGenesis(ctx, cdc, data)
		return []abci.ValidatorUpdate{}, nil
	case interface {
		InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) error
	}:
		return []abci.ValidatorUpdate{}, mod.InitGenesis(ctx, cdc, data)
	case interface {
		InitGenesis(ctx sdk.Context, data json.RawMessage)
	}:
		mod.InitGenesis(ctx, data)
		return []abci.ValidatorUpdate{}, nil
	default:
		return nil, nil
	}
}

// ExportGenesis delegates to the Aura module if it supports appmodule.HasGenesis.
func (a adapterModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) (json.RawMessage, error) {
	switch mod := a.module.(type) {
	case interface {
		ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage
	}:
		return mod.ExportGenesis(ctx, cdc), nil
	case interface {
		ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) (json.RawMessage, error)
	}:
		return mod.ExportGenesis(ctx, cdc)
	default:
		return json.RawMessage(`{}`), nil
	}
}

// BeginBlock delegates if supported.
func (a adapterModule) BeginBlock(ctx sdk.Context) error {
	switch mod := a.module.(type) {
	case interface {
		BeginBlock(ctx sdk.Context) error
	}:
		return mod.BeginBlock(ctx)
	case interface {
		BeginBlock(ctx sdk.Context)
	}:
		mod.BeginBlock(ctx)
		return nil
	default:
		return nil
	}
}

// EndBlock delegates if supported.
func (a adapterModule) EndBlock(ctx sdk.Context) error {
	switch mod := a.module.(type) {
	case interface {
		EndBlock(ctx sdk.Context) error
	}:
		return mod.EndBlock(ctx)
	case interface {
		EndBlock(ctx sdk.Context)
	}:
		mod.EndBlock(ctx)
		return nil
	default:
		return nil
	}
}

// ConsensusVersion is static.
func (adapterModule) ConsensusVersion() uint64 { return 1 }

// IsOnePerModuleType tags adapter as one-per-module.
func (adapterModule) IsOnePerModuleType() {}

// IsAppModule tags adapter as an appmodule.AppModule.
func (adapterModule) IsAppModule() {}

// wrapAdapters converts Aura modules to sdkmodule.AppModule via adapterModule.
func wrapAdapters(mods map[string]interface{}) []sdkmodule.AppModule {
	out := make([]sdkmodule.AppModule, 0, len(mods))
	for name, mod := range mods {
		out = append(out, adapterModule{name: name, module: mod})
	}
	return out
}
