package app

import (
	tmlog "cosmossdk.io/log"
	"github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
	"github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
)

// CosmosApp represents a minimal Cosmos SDK application shell that wires the identitychange module into baseapp.
type CosmosApp struct {
	*baseapp.BaseApp
	moduleManager ModuleManager
	grpcServer    *grpc.Server
	encoding      EncodingConfig
}

// NewCosmosApp builds the shell with the identitychange keeper/module manager wired into a baseapp instance.
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
	paramsStore := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(paramsStore)
	module := identitychange.NewAppModule(k)
	manager := NewModuleManager(module)
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
