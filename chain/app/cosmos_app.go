package app

import (
	tmlog "cosmossdk.io/log"
	"github.com/aequitas/aura/chain/x/bridge"
	"github.com/aequitas/aura/chain/x/confidencescore"
	cskeeper "github.com/aequitas/aura/chain/x/confidencescore/keeper"
	csparams "github.com/aequitas/aura/chain/x/confidencescore/params"
	cstypes "github.com/aequitas/aura/chain/x/confidencescore/types"
	"github.com/aequitas/aura/chain/x/dataregistry"
	"github.com/aequitas/aura/chain/x/dex"
	"github.com/aequitas/aura/chain/x/identitychange"
	idkeeper "github.com/aequitas/aura/chain/x/identitychange/keeper"
	idparams "github.com/aequitas/aura/chain/x/identitychange/params"
	idtypes "github.com/aequitas/aura/chain/x/identitychange/types"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	irkeeper "github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	irparams "github.com/aequitas/aura/chain/x/inclusionroutines/params"
	irtypes "github.com/aequitas/aura/chain/x/inclusionroutines/types"
	"github.com/aequitas/aura/chain/x/vcregistry"
	"github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
)

// CosmosApp represents a minimal Cosmos SDK application shell that wires the identitychange and inclusionroutines modules into baseapp.
type CosmosApp struct {
	*baseapp.BaseApp
	moduleManager ModuleManager
	grpcServer    *grpc.Server
	encoding      EncodingConfig
}

// NewCosmosApp builds the shell with the identitychange, inclusionroutines, and confidencescore keeper/module manager wired into a baseapp instance.
func NewCosmosApp(logger tmlog.Logger) *CosmosApp {
	if logger == nil {
		logger = tmlog.NewNopLogger()
	}
	encoding := MakeEncodingConfig()
	memDB := db.NewMemDB()
	bApp := baseapp.NewBaseApp(
		"aura",
		logger,
		memDB,
		dummyTxDecoder,
	)
	bApp.SetInterfaceRegistry(encoding.InterfaceRegistry)

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

	// Create module manager with all modules
	manager := NewModuleManager(
		[]identitychange.AppModule{idModule},
		[]inclusionroutines.AppModule{irModule},
		[]confidencescore.AppModule{csModule},
		[]vcregistry.AppModule{},
		[]dataregistry.AppModule{}, // Empty for now
		[]dex.AppModule{},          // Empty - requires BankKeeper, AccountKeeper
		[]bridge.AppModule{},       // Empty - requires BankKeeper, AccountKeeper
	)
	grpcServer := grpc.NewServer()
	app := &CosmosApp{
		BaseApp:       bApp,
		moduleManager: manager,
		grpcServer:    grpcServer,
		encoding:      encoding,
	}
	app.RegisterGRPCServices()
	return app
}

// EncodingConfig defines the concrete types required by the Cosmos app shell.
type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
}

// RegisterGRPCServices wires the gRPC server through the module manager.
func (c *CosmosApp) RegisterGRPCServices() {
	c.moduleManager.RegisterGRPCServices(c.grpcServer)
}

// Encoding returns the encoding configuration used by this app.
func (c *CosmosApp) Encoding() EncodingConfig {
	return c.encoding
}

// GRPCServer exposes the gRPC server so transport wiring can reuse it.
func (c *CosmosApp) GRPCServer() *grpc.Server {
	return c.grpcServer
}

// MakeEncodingConfig builds a minimal encoding configuration using protobuf codecs.
func MakeEncodingConfig() EncodingConfig {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
	}
}

func dummyTxDecoder([]byte) (sdk.Tx, error) {
	return nil, nil
}
