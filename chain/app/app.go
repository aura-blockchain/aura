package app

import (
	"github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
	"google.golang.org/grpc"
)

// App wires the identitychange module into an application-level shell.
type App struct {
	moduleManager ModuleManager
	grpcServer    *grpc.Server
}

// NewApp builds the shell with the identitychange keeper, module, and ModuleManager ready to register gRPC services.
func NewApp() *App {
	paramsStore := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(paramsStore)
	module := identitychange.NewAppModule(k)
	manager := NewModuleManager(module)
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
