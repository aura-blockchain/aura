package governance

import (
	"google.golang.org/grpc"

	"github.com/aequitas/aura/chain/x/governance/keeper"
)

// AppModule represents the governance application module
type AppModule struct {
	keeper *keeper.Keeper
}

// NewAppModule creates a new governance AppModule
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{
		keeper: k,
	}
}

// RegisterGRPCServices registers the module's gRPC services
func (am AppModule) RegisterGRPCServices(server *grpc.Server) {
	// Register message server
}

// Name returns the module name
func (am AppModule) Name() string {
	return "governance"
}
