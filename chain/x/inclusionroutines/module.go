package inclusionroutines

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// ModuleServices defines the interface for registering module services
type ModuleServices interface {
	RegisterMsgServer(inclusionroutinespb.MsgServer)
	RegisterQueryServer(inclusionroutinespb.QueryServer)
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
	config.RegisterMsgServer(NewMsgServer(m.keeper))
	config.RegisterQueryServer(NewQueryServer(m.keeper))
}

// BeginBlock executes all ABCI BeginBlock logic
func (m AppModule) BeginBlock() {
	// Periodically cleanup expired rate limit counters
	// This could be enhanced to run less frequently
	m.keeper.CleanupExpiredRateLimits()
}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock() {}
