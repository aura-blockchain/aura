package bridge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/bridge/client/cli"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

// ModuleServices defines how the application wires bridge gRPC services.
type ModuleServices interface {
	RegisterMsgServer(bridgepb.MsgServer)
	RegisterQueryServer(bridgepb.QueryServer)
}

// AppModuleBasic defines the basic application module
type AppModuleBasic struct{}

// Name returns the module name
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterGRPCGatewayRoutes is a no-op placeholder to satisfy the AppModuleBasic interface.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// RegisterServices registers module services (no-op for basic)
func (AppModuleBasic) RegisterServices(ModuleServices) {}

// RegisterInterfaces registers bridge message types for interface resolution.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &bridgepb.Msg_ServiceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&bridgepb.MsgLockTokens{},
		&bridgepb.MsgMintTokens{},
		&bridgepb.MsgUnlockTokens{},
		&bridgepb.MsgBurnTokens{},
		&bridgepb.MsgLinkAddress{},
		&bridgepb.MsgCrossChainSwap{},
		&bridgepb.MsgRelayTransfer{},
		&bridgepb.MsgFinalizeTransfer{},
		&bridgepb.MsgSubmitFraudProof{},
	)
}

// RegisterLegacyAminoCodec satisfies the module.AppModuleBasic interface (proto-first module).
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// DefaultGenesis returns the module's default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) []byte {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis validates the provided genesis state for the module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz []byte) error {
	var gen types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gen); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return types.ValidateGenesis(&gen)
}

// AppModule implements the app module interface
type AppModule struct {
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule instance
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

// Name returns the module name
func (AppModule) Name() string { return types.ModuleName }

// RegisterServices registers the module's message and query servers
func (m AppModule) RegisterServices(config ModuleServices) {
	if config == nil {
		panic(fmt.Sprintf("%s: nil module services", types.ModuleName))
	}
	config.RegisterMsgServer(keeper.NewMsgServerImpl(m.keeper))
	config.RegisterQueryServer(keeper.NewQueryServerImpl(m.keeper))
}

// BeginBlock executes all ABCI BeginBlock logic
func (m AppModule) BeginBlock(ctx sdk.Context) {
	// SECURITY: Reset mint limits for supply cap enforcement
	// This function cleans up old daily/hourly counters and allows fresh limits

	// Reset daily mint counters (removes counters from previous days)
	m.keeper.ResetDailyMint(ctx)

	// Reset hourly mint counters (removes counters from previous hours)
	m.keeper.ResetHourlyMint(ctx)
}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock(ctx sdk.Context) {
	// SECURITY: Auto-finalize expired pending transfers
	// This automatically completes transfers after the fraud proof window expires
	// to ensure users receive their funds without manual finalization
	m.keeper.ProcessExpiredPendingTransfers(ctx)
}

// InitGenesis initializes module state from genesis
func (m AppModule) InitGenesis(ctx sdk.Context, genesis types.GenesisState) error {
	return m.keeper.InitGenesis(ctx, genesis)
}

// ExportGenesis exports module state for genesis
func (m AppModule) ExportGenesis(ctx sdk.Context) types.GenesisState {
	return m.keeper.ExportGenesis(ctx)
}

// IsAppModule tags this module for Cosmos SDK module manager compatibility.
func (AppModule) IsAppModule() {}

// GetTxCmd returns the transaction commands for this module
func (AppModule) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the query commands for this module
func (AppModule) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}
