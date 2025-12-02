package confidencescore

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/confidencescore/keeper"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module
type AppModuleBasic struct{}

// Name returns the module name
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterGRPCGatewayRoutes is a no-op placeholder to satisfy the AppModuleBasic interface.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// RegisterLegacyAminoCodec registers legacy interfaces (none).
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterInterfaces registers protobuf interfaces.
func (AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis validates supplied genesis.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gen confidencescorepb.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gen); err != nil {
		return err
	}
	return types.ValidateGenesisState(&gen)
}

// AppModule implements the app module interface
type AppModule struct {
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule instance
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

// Name returns the module name
func (AppModule) Name() string { return types.ModuleName }

// RegisterServices registers the module's message and query servers
func (m AppModule) RegisterServices(cfg module.Configurator) {
	if cfg == nil {
		panic(fmt.Sprintf("%s: nil module configurator", types.ModuleName))
	}
	confidencescorepb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(m.keeper))
	confidencescorepb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(m.keeper))
}

// BeginBlock executes all ABCI BeginBlock logic
func (m AppModule) BeginBlock(ctx context.Context) error {
	// CleanupExpiredRateLimits is safe to ignore errors (no-op if not implemented)
	m.keeper.CleanupExpiredRateLimits(sdk.UnwrapSDKContext(ctx))
	return nil
}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock(ctx context.Context) error {
	_ = sdk.UnwrapSDKContext(ctx)
	return nil
}

// IsAppModule tags this module for Cosmos SDK module manager compatibility.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType marks depinject singleton.
func (AppModule) IsOnePerModuleType() {}

// RegisterLegacyAminoCodec is a no-op for module.AppModule compatibility.
func (AppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterInterfaces wires protobuf interfaces for module.AppModule compatibility.
func (AppModule) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// RegisterGRPCGatewayRoutes is a no-op for module.AppModule compatibility.
func (AppModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// DefaultGenesis returns default genesis for legacy module.HasGenesis compatibility.
func (m AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis validates supplied genesis for legacy module.HasGenesis compatibility.
func (m AppModule) ValidateGenesis(cdc codec.JSONCodec, cfg client.TxEncodingConfig, bz json.RawMessage) error {
	var gen confidencescorepb.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gen); err != nil {
		return err
	}
	return types.ValidateGenesisState(&gen)
}

// InitGenesis satisfies module.HasGenesis using sdk.Context.
func (m AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	var gen confidencescorepb.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gen); err != nil {
		panic(err)
	}
	if err := m.keeper.InitGenesis(ctx, gen); err != nil {
		panic(err)
	}
}

// ExportGenesis satisfies module.HasGenesis using sdk.Context.
func (m AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state := m.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(&state)
}

// Ensure AppModule implements module.HasGenesis.
var _ module.HasGenesis = AppModule{}
