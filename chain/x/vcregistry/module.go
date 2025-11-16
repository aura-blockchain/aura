package vcregistry

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/vcregistry/keeper"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
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
	config.RegisterMsgServer(NewMsgServer(m.keeper))
	config.RegisterQueryServer(NewQueryServer(m.keeper))
}

// BeginBlock executes all ABCI BeginBlock logic
// Note: SetCurrentHeight and SetCurrentTime should be called from the app wiring layer
// where the block context is available, before calling this method
func (m AppModule) BeginBlock() {
	// Periodically cleanup expired mint rate limit counters
	// Remove counters older than 7 days
	m.keeper.CleanupOldMintCounts()
}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock() {}

// InitGenesis initializes module state from genesis
func (m AppModule) InitGenesis(genesis types.GenesisState) error {
	return m.keeper.InitGenesis(genesis)
}

// ExportGenesis exports module state for genesis
func (m AppModule) ExportGenesis() types.GenesisState {
	return m.keeper.ExportGenesis()
}
