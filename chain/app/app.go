package app

import (
	"github.com/aequitas/aura/chain/x/bridge"
	"github.com/aequitas/aura/chain/x/confidencescore"
	cskeeper "github.com/aequitas/aura/chain/x/confidencescore/keeper"
	csparams "github.com/aequitas/aura/chain/x/confidencescore/params"
	cstypes "github.com/aequitas/aura/chain/x/confidencescore/types"
	"github.com/aequitas/aura/chain/x/dataregistry"
	drkeeper "github.com/aequitas/aura/chain/x/dataregistry/keeper"
	drparams "github.com/aequitas/aura/chain/x/dataregistry/params"
	drtypes "github.com/aequitas/aura/chain/x/dataregistry/types"
	"github.com/aequitas/aura/chain/x/dex"
	"github.com/aequitas/aura/chain/x/governance"
	govkeeper "github.com/aequitas/aura/chain/x/governance/keeper"
	govtypes "github.com/aequitas/aura/chain/x/governance/types"
	"github.com/aequitas/aura/chain/x/identitychange"
	idkeeper "github.com/aequitas/aura/chain/x/identitychange/keeper"
	idparams "github.com/aequitas/aura/chain/x/identitychange/params"
	idtypes "github.com/aequitas/aura/chain/x/identitychange/types"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	irkeeper "github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	irparams "github.com/aequitas/aura/chain/x/inclusionroutines/params"
	irtypes "github.com/aequitas/aura/chain/x/inclusionroutines/types"
	"github.com/aequitas/aura/chain/x/vcregistry"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
	vcparams "github.com/aequitas/aura/chain/x/vcregistry/params"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	"google.golang.org/grpc"
)

// App wires the identitychange, inclusionroutines, confidencescore, vcregistry, dataregistry, governance, dex, and bridge modules into an application-level shell.
type App struct {
	moduleManager ModuleManager
	grpcServer    *grpc.Server
}

// NewApp builds the shell with the identitychange, inclusionroutines, confidencescore, vcregistry, dataregistry, governance, dex, and bridge keepers, modules, and ModuleManager ready to register gRPC services.
// Note: DEX and Bridge modules require full Cosmos SDK keeper dependencies (BankKeeper, AccountKeeper) which are not yet initialized in this simplified app structure.
func NewApp() *App {
	// Initialize identitychange module
	idParamsStore := idparams.NewStore(idtypes.DefaultParams())
	idKeeper := idkeeper.NewKeeper(idParamsStore)
	idModule := identitychange.NewAppModule(idKeeper)

	// Initialize inclusionroutines module
	irParamsStore := irparams.NewStore(irtypes.DefaultParams())
	irKeeper := irkeeper.NewKeeper(irParamsStore)
	irModule := inclusionroutines.NewAppModule(irKeeper)

	// Initialize confidencescore module (needs IR keeper as dependency)
	csParamsStore := csparams.NewStore(cstypes.DefaultParams())
	csKeeper := cskeeper.NewKeeper(csParamsStore)
	csKeeper.SetIRRegistry(irKeeper)
	csModule := confidencescore.NewAppModule(csKeeper)

	// Initialize vcregistry module (needs CS keeper as dependency)
	vcParamsStore := vcparams.NewStore(vctypes.DefaultParams())
	vcKeeper := vckeeper.NewKeeper(vcParamsStore)
	vcKeeper.SetConfidenceScoreKeeper(csKeeper)
	vcModule := vcregistry.NewAppModule(vcKeeper)

	// Initialize dataregistry module
	drParamsStore := drparams.NewStore(drtypes.DefaultParams())
	drKeeper := drkeeper.NewKeeper(drParamsStore)
	drModule := dataregistry.NewAppModule(drKeeper)

	// Initialize governance module with comprehensive security features
	govKeeper := govkeeper.NewKeeper(govtypes.DefaultParams())
	govModule := governance.NewAppModule(govKeeper)

	// TODO: Initialize DEX module when Cosmos SDK keepers are available
	// DEX requires: BankKeeper, AccountKeeper, and VCRegistryKeeper for IR boost feature
	// dexKeeper := dexkeeper.NewKeeper(cdc, storeKey, paramSpace, bankKeeper, accountKeeper, vcKeeper)
	// dexModule := dex.NewAppModule(dexKeeper)
	var dexModules []dex.AppModule

	// TODO: Initialize Bridge module when Cosmos SDK keepers are available
	// Bridge requires: BankKeeper, AccountKeeper, and VCRegistryKeeper for shared identity
	// bridgeKeeper := bridgekeeper.NewKeeper(cdc, storeKey, paramSpace, bankKeeper, accountKeeper, vcKeeper)
	// bridgeModule := bridge.NewAppModule(bridgeKeeper)
	var bridgeModules []bridge.AppModule

	// Create module manager with all modules
	manager := NewModuleManager(
		[]identitychange.AppModule{idModule},
		[]inclusionroutines.AppModule{irModule},
		[]confidencescore.AppModule{csModule},
		[]vcregistry.AppModule{vcModule},
		[]dataregistry.AppModule{drModule},
		[]governance.AppModule{govModule},
		dexModules,
		bridgeModules,
	)
	server := grpc.NewServer()
	app := &App{
		moduleManager: manager,
		grpcServer:    server,
	}
	app.RegisterGRPCServices()
	return app
}

// RegisterGRPCServices wires the module manager into the underlying gRPC server registrar.
func (a *App) RegisterGRPCServices() {
	a.moduleManager.RegisterGRPCServices(a.grpcServer)
}

// GRPCServer exposes the underlying gRPC server so callers can hook HTTP/gRPC transport.
func (a *App) GRPCServer() *grpc.Server {
	return a.grpcServer
}
