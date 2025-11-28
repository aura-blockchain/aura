package dex

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// ModuleServices defines the interface for registering module services.
// The ModuleManager provides a concrete implementation backed by gRPC registration.
type ModuleServices interface {
	RegisterMsgServer(dexpb.MsgServer)
	RegisterQueryServer(dexpb.QueryServer)
}

// AppModuleBasic defines the basic application module
type AppModuleBasic struct{}

// Name returns the module name
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterServices registers module services (no-op for basic)
func (AppModuleBasic) RegisterServices(ModuleServices) {}

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

	config.RegisterMsgServer(keeper.NewMsgServerImpl(m.keeper))
	config.RegisterQueryServer(keeper.NewQueryServerImpl(m.keeper))
}

// BeginBlock executes any BeginBlock logic. Currently no-op.
func (m AppModule) BeginBlock(_ context.Context) {}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock(ctx context.Context) {
	if ctx == nil || m.keeper == nil {
		return
	}

	// Use the SDK context to run housekeeping such as pruning expired orders.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	m.keeper.CleanupExpiredOrders(sdkCtx)
	m.keeper.CleanupExpiredHTLCs(sdkCtx)
}

// InitGenesis initializes module state from genesis
func (m AppModule) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	if m.keeper == nil {
		return fmt.Errorf("dex keeper not initialized")
	}
	return m.keeper.InitGenesis(ctx, data)
}

// ExportGenesis exports module state for genesis
func (m AppModule) ExportGenesis(ctx sdk.Context) types.GenesisState {
	if m.keeper == nil {
		return *types.DefaultGenesis()
	}
	return m.keeper.ExportGenesis(ctx)
}
