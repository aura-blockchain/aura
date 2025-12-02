package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/aequitas/aura/chain/x/identity/keeper"
	"github.com/aequitas/aura/chain/x/identity/types"
)

var (
	_ module.AppModuleBasic = AppModule{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module used by the identity module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the identity module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the identity module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register types here if using legacy amino
}

// RegisterInterfaces registers the identity module's interface types
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// Register interface types here
}

// DefaultGenesis returns default genesis state as raw bytes for the identity module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the identity module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bytes.TrimSpace(bz)) == 0 {
		return nil
	}
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return types.ValidateGenesisState(&data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the identity module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes here
}

// GetTxCmd returns the root tx command for the identity module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	// Return transaction commands
	return nil
}

// GetQueryCmd returns no root query command for the identity module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	// Return query commands
	return nil
}

// ============================================================================
// AppModule
// ============================================================================

// AppModule implements an application module for the identity module.
type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper *keeper.Keeper,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the identity module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Register Msg service
	// types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))

	// Register Query service
	// types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// RegisterInvariants registers the identity module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Register invariants here
}

// InitGenesis performs genesis initialization for the identity module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	if err := am.keeper.InitGenesis(ctx, &genesisState); err != nil {
		panic(fmt.Sprintf("failed to initialize genesis state: %s", err))
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the identity module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to export genesis state: %s", err))
	}
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ============================================================================
// App Wiring Setup
// ============================================================================

// Module is the app module type for the identity module.
type Module struct {
	cdc          codec.Codec
	storeService store.KVStoreService
}

// NewModule creates a new Module instance
func NewModule(cdc codec.Codec, storeService store.KVStoreService) *Module {
	return &Module{
		cdc:          cdc,
		storeService: storeService,
	}
}

// ProvideModule provides the identity module to the app wiring system
func ProvideModule(
	cdc codec.Codec,
	storeService store.KVStoreService,
	authority string,
	logger log.Logger,
) (AppModule, *keeper.Keeper) {
	k := keeper.NewKeeper(
		storeService,
		cdc,
		authority,
		logger,
	)

	return NewAppModule(cdc, k), k
}

// ============================================================================
// AutoCLI Support
// ============================================================================

// AutoCLIOptions returns the autocli options for the identity module.
func (am AppModule) AutoCLIOptions() *AutoCLIOptions {
	return &AutoCLIOptions{
		TxCommands: []*AutoCLICommand{
			{
				RpcMethod: "CreateRole",
				Use:       "create-role [name] [permissions] [description]",
				Short:     "Create a new role with specified permissions",
			},
			{
				RpcMethod: "AssignRole",
				Use:       "assign-role [address] [role-name]",
				Short:     "Assign a role to an address",
			},
			{
				RpcMethod: "RevokeRole",
				Use:       "revoke-role [address] [role-name]",
				Short:     "Revoke a role from an address",
			},
			{
				RpcMethod: "CreateChangeRequest",
				Use:       "create-change-request [did] [ir-id] [metadata-hash]",
				Short:     "Create an identity change request",
			},
			{
				RpcMethod: "SubmitVerification",
				Use:       "submit-verification [request-id] [approved]",
				Short:     "Submit verification for a change request",
			},
			{
				RpcMethod: "ApplyChange",
				Use:       "apply-change [request-id]",
				Short:     "Apply an approved identity change",
			},
		},
		QueryCommands: []*AutoCLICommand{
			{
				RpcMethod: "Params",
				Use:       "params",
				Short:     "Query identity module parameters",
			},
			{
				RpcMethod: "Role",
				Use:       "role [name]",
				Short:     "Query a role by name",
			},
			{
				RpcMethod: "IdentityRecord",
				Use:       "identity [did]",
				Short:     "Query an identity record by DID",
			},
			{
				RpcMethod: "ChangeRequest",
				Use:       "change-request [request-id]",
				Short:     "Query a change request by ID",
			},
			{
				RpcMethod: "ChangeHistory",
				Use:       "change-history [did]",
				Short:     "Query change history for a DID",
			},
		},
	}
}

// AutoCLIOptions defines the autocli options for a module.
type AutoCLIOptions struct {
	TxCommands    []*AutoCLICommand
	QueryCommands []*AutoCLICommand
}

// AutoCLICommand defines an autocli command.
type AutoCLICommand struct {
	RpcMethod string
	Use       string
	Short     string
	Long      string
}

// ============================================================================
// BeginBlocker and EndBlocker
// ============================================================================

// BeginBlocker is called at the beginning of every block
func (am AppModule) BeginBlock(ctx context.Context) error {
	// Perform any begin block logic here
	return nil
}

// EndBlocker is called at the end of every block
func (am AppModule) EndBlock(ctx context.Context) error {
	// Perform any end block logic here
	// For example: cleanup expired sessions, proposals, etc.
	return nil
}
