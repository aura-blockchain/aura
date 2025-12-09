package dex

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

	"github.com/aequitas/aura/chain/x/dex/client/cli"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// AppModuleBasic defines the basic application module
type AppModuleBasic struct{}

// Name returns the module name
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers Amino types (DEX is protobuf-native).
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterGRPCGatewayRoutes is a no-op placeholder to satisfy the AppModuleBasic interface.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// RegisterInterfaces registers DEX message types for the interface registry.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns the module's default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
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
	return types.ValidateGenesis(&gen)
}

// GetTxCmd returns the root tx command for the dex module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the root query command for the dex module.
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
	dexpb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(m.keeper))
	dexpb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(m.keeper))
}

// InitGenesis initializes module state from genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var gen types.GenesisState
	if len(bytes.TrimSpace(data)) == 0 {
		gen = *types.DefaultGenesis()
	} else if err := cdc.UnmarshalJSON(data, &gen); err != nil {
		panic(fmt.Errorf("failed to unmarshal dex genesis: %w", err))
	}
	if err := am.keeper.InitGenesis(ctx, gen); err != nil {
		panic(fmt.Errorf("dex InitGenesis: %w", err))
	}
}

// ExportGenesis exports module state for genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gen := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gen)
}

// ConsensusVersion returns the module consensus version
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes any BeginBlock logic. Currently no-op.
func (m AppModule) BeginBlock(_ sdk.Context) {}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock(ctx sdk.Context) {
	// Run housekeeping such as pruning expired orders
	m.keeper.CleanupExpiredOrders(ctx)
	m.keeper.CleanupExpiredHTLCs(ctx)

	// Record TWAP price observations for all active pools
	// This provides flash loan attack protection
	m.keeper.RecordAllPoolPrices(ctx)

	// Cleanup expired order commitments (commit-reveal scheme)
	m.keeper.CleanupExpiredCommitments(ctx)

	// Execute batch of revealed orders (front-running protection)
	params := m.keeper.GetParams(ctx)
	if params.BatchExecutionEnabled {
		// Execute batch every N blocks
		if ctx.BlockHeight()%int64(params.BatchExecutionInterval) == 0 {
			if err := m.keeper.ExecuteBatch(ctx); err != nil {
				// Log error but don't panic (batch execution is non-critical)
				ctx.Logger().Error("failed to execute batch", "error", err)
			}
		}
	}
}

// RegisterInvariants registers dex invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// IsOnePerModuleType tags this module for depinject one-per-module compatibility.
func (AppModule) IsOnePerModuleType() {}

// IsAppModule tags this module for Cosmos SDK module manager compatibility.
func (AppModule) IsAppModule() {}
