package app

import (
	"github.com/aequitas/aura/chain/x/bridge"
	"github.com/aequitas/aura/chain/x/confidencescore"
	"github.com/aequitas/aura/chain/x/dataregistry"
	"github.com/aequitas/aura/chain/x/dex"
	"github.com/aequitas/aura/chain/x/governance"
	"github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	"github.com/aequitas/aura/chain/x/vcregistry"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"google.golang.org/grpc"
)

// ModuleManager wires Cosmos SDK modules (eventually) into the application lifecycle.
type ModuleManager struct {
	identityChangeModules    []identitychange.AppModule
	inclusionRoutinesModules []inclusionroutines.AppModule
	confidenceScoreModules   []confidencescore.AppModule
	vcRegistryModules        []vcregistry.AppModule
	dataRegistryModules      []dataregistry.AppModule
	governanceModules        []governance.AppModule
	dexModules               []dex.AppModule
	bridgeModules            []bridge.AppModule
}

// NewModuleManager initializes a tracker for the provided modules.
func NewModuleManager(idModules []identitychange.AppModule, irModules []inclusionroutines.AppModule, csModules []confidencescore.AppModule, vcModules []vcregistry.AppModule, drModules []dataregistry.AppModule, govModules []governance.AppModule, dexModules []dex.AppModule, bridgeModules []bridge.AppModule) ModuleManager {
	return ModuleManager{
		identityChangeModules:    idModules,
		inclusionRoutinesModules: irModules,
		confidenceScoreModules:   csModules,
		vcRegistryModules:        vcModules,
		dataRegistryModules:      drModules,
		governanceModules:        govModules,
		dexModules:               dexModules,
		bridgeModules:            bridgeModules,
	}
}

// RegisterGRPCServices registers each module's gRPC handlers with the provided registrar.
func (m ModuleManager) RegisterGRPCServices(server grpc.ServiceRegistrar) {
	if server == nil {
		panic("module manager: nil gRPC server registrar")
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

	// Register governance modules
	govServices := &governanceServices{registrar: server}
	for _, module := range m.governanceModules {
		module.RegisterGRPCServices(govServices)
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

// identityChangeServices implements identitychange.ModuleServices
type identityChangeServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *identityChangeServices) RegisterMsgServer(server identitychangepb.MsgServer) {
	identitychangepb.RegisterMsgServer(s.registrar, server)
}

func (s *identityChangeServices) RegisterQueryServer(server identitychangepb.QueryServer) {
	identitychangepb.RegisterQueryServer(s.registrar, server)
}

// inclusionRoutinesServices implements inclusionroutines.ModuleServices
type inclusionRoutinesServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *inclusionRoutinesServices) RegisterMsgServer(server inclusionroutinespb.MsgServer) {
	inclusionroutinespb.RegisterMsgServer(s.registrar, server)
}

func (s *inclusionRoutinesServices) RegisterQueryServer(server inclusionroutinespb.QueryServer) {
	inclusionroutinespb.RegisterQueryServer(s.registrar, server)
}

// confidenceScoreServices implements confidencescore.ModuleServices
type confidenceScoreServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *confidenceScoreServices) RegisterMsgServer(server confidencescorepb.MsgServer) {
	confidencescorepb.RegisterMsgServer(s.registrar, server)
}

func (s *confidenceScoreServices) RegisterQueryServer(server confidencescorepb.QueryServer) {
	confidencescorepb.RegisterQueryServer(s.registrar, server)
}

// vcRegistryServices implements vcregistry.ModuleServices
type vcRegistryServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *vcRegistryServices) RegisterMsgServer(server vcregistrypb.MsgServer) {
	vcregistrypb.RegisterMsgServer(s.registrar, server)
}

func (s *vcRegistryServices) RegisterQueryServer(server vcregistrypb.QueryServer) {
	vcregistrypb.RegisterQueryServer(s.registrar, server)
}

// dataRegistryServices implements dataregistry.ModuleServices
type dataRegistryServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *dataRegistryServices) RegisterMsgServer(server dataregistry.MsgServer) {
	// Note: In production, this would register with the gRPC server
	// For now, the interface is defined but not registered
}

func (s *dataRegistryServices) RegisterQueryServer(server dataregistry.QueryServer) {
	// Note: In production, this would register with the gRPC server
	// For now, the interface is defined but not registered
}

// governanceServices implements governance.ModuleServices
type governanceServices struct {
	registrar grpc.ServiceRegistrar
}

// RegisterGRPCServices is a pass-through to satisfy the interface
// The governance module handles its own gRPC registration
func (s *governanceServices) RegisterGRPCServices(server *grpc.Server) {
	// Governance module handles its own registration
	// This is just to satisfy the interface
}

// dexServices implements dex.ModuleServices
type dexServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *dexServices) RegisterMsgServer(server interface{}) {
	// TODO: Register with gRPC server when proto files are generated
	// dexpb.RegisterMsgServer(s.registrar, server.(dexpb.MsgServer))
}

func (s *dexServices) RegisterQueryServer(server interface{}) {
	// TODO: Register with gRPC server when proto files are generated
	// dexpb.RegisterQueryServer(s.registrar, server.(dexpb.QueryServer))
}

// bridgeServices implements bridge.ModuleServices
type bridgeServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *bridgeServices) RegisterMsgServer(server interface{}) {
	// TODO: Register with gRPC server when proto files are generated
	// bridgepb.RegisterMsgServer(s.registrar, server.(bridgepb.MsgServer))
}

func (s *bridgeServices) RegisterQueryServer(server interface{}) {
	// TODO: Register with gRPC server when proto files are generated
	// bridgepb.RegisterQueryServer(s.registrar, server.(bridgepb.QueryServer))
}
