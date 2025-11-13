package app

import (
	"github.com/aequitas/aura/chain/x/identitychange"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
	"google.golang.org/grpc"
)

// ModuleManager wires Cosmos SDK modules (eventually) into the application lifecycle.
type ModuleManager struct {
	modules []identitychange.AppModule
}

// NewModuleManager initializes a tracker for the provided modules.
func NewModuleManager(modules ...identitychange.AppModule) ModuleManager {
	return ModuleManager{modules: modules}
}

// RegisterGRPCServices registers each module's gRPC handlers with the provided registrar.
func (m ModuleManager) RegisterGRPCServices(server grpc.ServiceRegistrar) {
	if server == nil {
		panic("module manager: nil gRPC server registrar")
	}
	services := &grpcModuleServices{registrar: server}
	for _, module := range m.modules {
		module.RegisterServices(services)
	}
}

type grpcModuleServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *grpcModuleServices) RegisterMsgServer(server identitychangepb.MsgServer) {
	identitychangepb.RegisterMsgServer(s.registrar, server)
}

func (s *grpcModuleServices) RegisterQueryServer(server identitychangepb.QueryServer) {
	identitychangepb.RegisterQueryServer(s.registrar, server)
}
