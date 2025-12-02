package dataregistry

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	pb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

// ModuleServices defines the interface for registering module services
type ModuleServices interface {
	RegisterMsgServer(pb.MsgServer)
	RegisterQueryServer(pb.QueryServer)
}

// AppModuleBasic defines the basic application module
type AppModuleBasic struct{}

// Name returns the module name
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers amino types (no-op for proto-only module).
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterGRPCGatewayRoutes is a no-op placeholder to satisfy the AppModuleBasic interface.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// RegisterServices registers module services (no-op for basic)
func (AppModuleBasic) RegisterServices(ModuleServices) {}

// RegisterInterfaces wires the data registry msg service into the interface registry.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &pb.Msg_ServiceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&pb.MsgStoreDataItem{},
		&pb.MsgUpdateDataItem{},
		&pb.MsgDeleteDataItem{},
		&pb.MsgVerifyDataItem{},
		&pb.MsgRevokeDataItem{},
	)
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
func (m AppModule) RegisterServices(config ModuleServices) {
	if config == nil {
		panic(fmt.Sprintf("%s: nil module services", types.ModuleName))
	}
	// TODO: Re-enable once msg_server.go and query_server.go are fixed
	// config.RegisterMsgServer(keeper.NewMsgServer(m.keeper))
	// config.RegisterQueryServer(keeper.NewQueryServer(m.keeper))
}

// BeginBlock executes all ABCI BeginBlock logic
func (m AppModule) BeginBlock() {
	// Periodically check for expired data items
	// In production, implement expiration logic
}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock() {}

// InitGenesis initializes module state from genesis
func (m AppModule) InitGenesis(ctx context.Context, genesis types.GenesisState) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return m.keeper.InitGenesis(sdkCtx, genesis)
}

// ExportGenesis exports module state for genesis
func (m AppModule) ExportGenesis(ctx context.Context) types.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return m.keeper.ExportGenesis(sdkCtx)
}

// IsAppModule tags this module for Cosmos SDK module manager compatibility.
func (AppModule) IsAppModule() {}
