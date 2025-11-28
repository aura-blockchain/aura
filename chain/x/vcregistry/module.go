package vcregistry

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/vcregistry/keeper"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ModuleServices defines the interface for registering module services
type ModuleServices interface {
	RegisterMsgServer(vcregistrypb.MsgServer)
	RegisterQueryServer(vcregistrypb.QueryServer)
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
	config.RegisterMsgServer(keeper.NewMsgServer(m.keeper))
	config.RegisterQueryServer(keeper.NewQueryServer(m.keeper))
}

// BeginBlock executes all ABCI BeginBlock logic
func (m AppModule) BeginBlock(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Consensus safety: time/height come from context now
	// SetCurrentHeight/SetCurrentTime were removed for determinism
	// The keeper gets block metadata directly from ctx when needed
	_ = ctx // Use context directly in keeper methods

	// Periodically cleanup expired mint rate limit counters
	m.keeper.CleanupOldMintCounts(ctx)
}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock() {}

// InitGenesis initializes module state from genesis
func (m AppModule) InitGenesis(ctx context.Context, genesis types.GenesisState) error {
	return m.keeper.InitGenesis(ctx, genesis)
}

// ExportGenesis exports module state for genesis
func (m AppModule) ExportGenesis(ctx context.Context) types.GenesisState {
	return m.keeper.ExportGenesis(ctx)
}

func unwrapSDKContext(ctx context.Context) (sdk.Context, bool) {
	if ctx == nil {
		return sdk.Context{}, false
	}
	var ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx, ok
}
