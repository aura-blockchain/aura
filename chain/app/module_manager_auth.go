package app

import (
	"github.com/aequitas/aura/chain/x/auth"
	"github.com/aequitas/aura/chain/x/bridge"
	"github.com/aequitas/aura/chain/x/compliance"
	"github.com/aequitas/aura/chain/x/confidencescore"
	"github.com/aequitas/aura/chain/x/dataregistry"
	"github.com/aequitas/aura/chain/x/dex"
	"github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	"github.com/aequitas/aura/chain/x/vcregistry"
	identitypb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
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
	complianceModules        []compliance.AppModule
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
	compModules []compliance.AppModule,
	dexModules []dex.AppModule,
	bridgeModules []bridge.AppModule,
) ModuleManagerWithAuth {
	return ModuleManagerWithAuth{
		authModules:              authModules,
		identityChangeModules:    idModules,
		inclusionRoutinesModules: irModules,
		confidenceScoreModules:   csModules,
		vcRegistryModules:        vcModules,
		dataRegistryModules:      drModules,
		complianceModules:        compModules,
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
	for _, module := range m.authModules {
		module.RegisterGRPCServer(server.(*grpc.Server))
	}

	// Register identitychange modules
	idServices := &identityChangeServicesAuth{registrar: server}
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

	// Register compliance modules
	compServices := &complianceServices{registrar: server}
	for _, module := range m.complianceModules {
		module.RegisterServices(compServices)
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

// identityChangeServicesAuth implements identitychange.ModuleServices for auth module manager
type identityChangeServicesAuth struct {
	registrar grpc.ServiceRegistrar
}

func (s *identityChangeServicesAuth) RegisterMsgServer(server identitypb.MsgServer) {
	identitypb.RegisterMsgServer(s.registrar, server)
}

func (s *identityChangeServicesAuth) RegisterQueryServer(server identitypb.QueryServer) {
	identitypb.RegisterQueryServer(s.registrar, server)
}

// Note: Service type definitions (inclusionRoutinesServices, confidenceScoreServices,
// vcRegistryServices, dataRegistryServices, complianceServices, dexServices, bridgeServices)
// are defined in module_manager.go and shared across both ModuleManager and ModuleManagerWithAuth
// to avoid redeclaration.
