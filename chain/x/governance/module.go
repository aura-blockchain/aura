package governance

import (
    "github.com/cosmos/cosmos-sdk/types/module"
    "google.golang.org/grpc"

    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/spf13/cobra"

    "github.com/aequitas/aura/chain/x/governance/client/cli"
    "github.com/aequitas/aura/chain/x/governance/keeper"
    "github.com/aequitas/aura/chain/x/governance/types"
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

// RegisterServices registers the module's gRPC services with the configurator
func (am AppModule) RegisterServices(cfg module.Configurator) {
    // Register message and query servers
}

// RegisterGRPCServices registers the module's gRPC services
func (am AppModule) RegisterGRPCServices(server *grpc.Server) {
    // Register message server
}

// Name returns the module name
func (am AppModule) Name() string {
    return "governance"
}

// GetTxCmd returns the transaction commands for this module
func (AppModule) GetTxCmd() *cobra.Command {
    return cli.GetTxCmd()
}

// GetQueryCmd returns the query commands for this module
func (AppModule) GetQueryCmd() *cobra.Command {
    return cli.GetQueryCmd()
}

// DefaultGenesis returns the default genesis state for the governance module.
func (AppModule) DefaultGenesis() types.GenesisState {
    return *types.DefaultGenesis()
}

// ValidateGenesis validates the supplied genesis state.
func (AppModule) ValidateGenesis(genesis types.GenesisState) error {
    return genesis.Validate()
}

// RegisterInvariants registers the module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
    // keeper.RegisterInvariants(ir, am.keeper) // TODO: Implement RegisterInvariants
}
