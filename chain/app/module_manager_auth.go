package app

import (
	"github.com/aequitas/aura/chain/x/auth"
	"github.com/aequitas/aura/chain/x/bridge"
	"github.com/aequitas/aura/chain/x/confidencescore"
	"github.com/aequitas/aura/chain/x/dataregistry"
	"github.com/aequitas/aura/chain/x/dex"
	"github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	"github.com/aequitas/aura/chain/x/vcregistry"
	authpb "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"google.golang.org/grpc"
)

// ModuleManagerWithAuth extends ModuleManager to include auth module
type ModuleManagerWithAuth struct {
	authModules              []auth.AppModule
	identityChangeModules    []identitychange.AppModule
	inclusionRoutinesModules []inclusionroutines.AppModule
	confidenceScoreModules   []confidencescore.AppModule
	vcRegistryModules        []vcregistry.AppModule
	dataRegistryModules      []dataregistry.AppModule
	dexModules               []dex.AppModule
	bridgeModules            []bridge.AppModule
}

// NewModuleManagerWithAuth initializes a tracker for all modules including auth
func NewModuleManagerWithAuth(
	authModules []auth.AppModule,
	idModules []identitychange.AppModule,
	irModules []inclusionroutines.AppModule,
	csModules []confidencescore.AppModule,
	vcModules []vcregistry.AppModule,
	drModules []dataregistry.AppModule,
	dexModules []dex.AppModule,
	bridgeModules []bridge.AppModule,
) ModuleManager {
	return ModuleManagerWithAuth{
		authModules:              authModules,
		identityChangeModules:    idModules,
		inclusionRoutinesModules: irModules,
		confidenceScoreModules:   csModules,
		vcRegistryModules:        vcModules,
		dataRegistryModules:      drModules,
		dexModules:               dexModules,
		bridgeModules:            bridgeModules,
	}
}

// RegisterGRPCServices registers all module gRPC handlers including auth
func (m ModuleManagerWithAuth) RegisterGRPCServices(server grpc.ServiceRegistrar) {
	if server == nil {
		panic("module manager: nil gRPC server registrar")
	}

	// Register auth modules first (for access control)
	authServices := &authServices{registrar: server}
	for _, module := range m.authModules {
		module.RegisterGRPCServer(server.(*grpc.Server))
	}

	// Register identitychange modules
	idServices := &identityChangeServices{registrar: server}
	for _, module := range m.identityChangeModules {
		module.RegisterServices(idServices)
	}

	// Register inclusionroutines modules
	irServices := &inclusionRoutinesServices{registrar: server}
	for _, module := range m.inclusionRoutinesModules {
		module.RegisterServices(irServices)
	}

	// Register confidencescore modules
	csServices := &confidenceScoreServices{registrar: server}
	for _, module := range m.confidenceScoreModules {
		module.RegisterServices(csServices)
	}

	// Register vcregistry modules
	vcServices := &vcRegistryServices{registrar: server}
	for _, module := range m.vcRegistryModules {
		module.RegisterServices(vcServices)
	}

	// Register dataregistry modules
	drServices := &dataRegistryServices{registrar: server}
	for _, module := range m.dataRegistryModules {
		module.RegisterServices(drServices)
	}

	// Register dex modules
	dexServices := &dexServices{registrar: server}
	for _, module := range m.dexModules {
		module.RegisterServices(dexServices)
	}

	// Register bridge modules
	bridgeServices := &bridgeServices{registrar: server}
	for _, module := range m.bridgeModules {
		module.RegisterServices(bridgeServices)
	}
}

// authServices implements auth.ModuleServices
type authServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *authServices) RegisterMsgServer(server authpb.MsgServer) {
	authpb.RegisterMsgServer(s.registrar, server)
}

func (s *authServices) RegisterQueryServer(server authpb.QueryServer) {
	authpb.RegisterQueryServer(s.registrar, server)
}
