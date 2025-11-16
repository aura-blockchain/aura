package bridge

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// ModuleServices defines the interface for registering module services
// Note: Proto generation pending - this is a stub interface
type ModuleServices interface {
	RegisterMsgServer(interface{})
	RegisterQueryServer(interface{})
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
	// TODO: Register msg and query servers when proto generation is complete
	// config.RegisterMsgServer(NewMsgServer(m.keeper))
	// config.RegisterQueryServer(NewQueryServer(m.keeper))
}

// BeginBlock executes all ABCI BeginBlock logic
func (m AppModule) BeginBlock() {
	// No begin block logic needed for Bridge
}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock() {
	// No end block logic needed for Bridge currently
}

// InitGenesis initializes module state from genesis
func (m AppModule) InitGenesis(data interface{}) error {
	// TODO: Implement when genesis types are defined
	return nil
}

// ExportGenesis exports module state for genesis
func (m AppModule) ExportGenesis() interface{} {
	// TODO: Implement when genesis types are defined
	return nil
}
