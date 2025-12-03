package vcregistry

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/aequitas/aura/chain/x/vcregistry/client/cli"
	"github.com/aequitas/aura/chain/x/vcregistry/keeper"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// AppModuleBasic defines the basic application module
type AppModuleBasic struct{}

// Name returns the module name
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterGRPCGatewayRoutes is a no-op placeholder to satisfy the AppModuleBasic interface.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// RegisterInterfaces wires the module messages into the interface registry so Msg servers can be registered safely.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec satisfies the module.AppModuleBasic interface (proto-only module).
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// DefaultGenesis returns the module's default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return types.DefaultGenesis(cdc)
}

// ValidateGenesis validates the provided genesis state for the module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bytes.TrimSpace(bz)) == 0 {
		return nil
	}
	var gen types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gen); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return types.ValidateGenesisState(&gen)
}

// GetTxCmd returns the root tx command for the vcregistry module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the root query command for the vcregistry module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements the app module interface
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule instance
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// Name returns the module name
func (AppModule) Name() string { return types.ModuleName }

// RegisterServices registers the module's message and query servers
func (m AppModule) RegisterServices(cfg module.Configurator) {
	vcregistrypb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(m.keeper))
	vcregistrypb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(m.keeper))
}

// InitGenesis initializes module state from genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var gen types.GenesisState
	if len(bytes.TrimSpace(data)) == 0 {
		bz := types.DefaultGenesis(cdc)
		cdc.MustUnmarshalJSON(bz, &gen)
	} else if err := cdc.UnmarshalJSON(data, &gen); err != nil {
		panic(fmt.Errorf("failed to unmarshal vcregistry genesis: %w", err))
	}
	// Convert sdk.Context to context.Context for keeper
	if err := am.keeper.InitGenesis(sdk.WrapSDKContext(ctx), gen); err != nil {
		panic(fmt.Errorf("vcregistry InitGenesis: %w", err))
	}
}

// ExportGenesis exports module state for genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// Convert sdk.Context to context.Context for keeper
	gen := am.keeper.ExportGenesis(sdk.WrapSDKContext(ctx))
	return cdc.MustMarshalJSON(&gen)
}

// ConsensusVersion returns the module consensus version
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes all ABCI BeginBlock logic
func (m AppModule) BeginBlock(ctx sdk.Context) {
	// Periodically cleanup expired mint rate limit counters
	m.keeper.CleanupOldMintCounts(sdk.WrapSDKContext(ctx))
}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock(ctx sdk.Context) {}

// RegisterInvariants registers vcregistry invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// IsOnePerModuleType tags this module for depinject one-per-module compatibility.
func (AppModule) IsOnePerModuleType() {}

// IsAppModule tags this module for Cosmos SDK module manager compatibility.
func (AppModule) IsAppModule() {}
