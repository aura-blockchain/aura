package prevalidation

import (
	"bytes"
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/prevalidation/keeper"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
	pb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module used by the prevalidation module
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the prevalidation module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the prevalidation module's types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the prevalidation module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC Gateway routes if needed
}

// DefaultGenesis returns default genesis state as raw bytes for the prevalidation module
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the prevalidation module
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bytes.TrimSpace(bz)) == 0 {
		return nil
	}
	var data pb.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return err
	}
	return types.ValidateGenesis(&data)
}

// AppModule implements an application module for the prevalidation module
type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface
func (am AppModule) IsAppModule() {}

// RegisterServices registers the module's message and query servers
func (am AppModule) RegisterServices(cfg module.Configurator) {
	pb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	pb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs genesis initialization for the prevalidation module
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState pb.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.initGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the prevalidation module
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.exportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// initGenesis initializes the module's state from genesis
func (am AppModule) initGenesis(ctx sdk.Context, gs *pb.GenesisState) {
	// Genesis initialization - set default params
	params := types.DefaultParams()
	if err := am.keeper.SetParams(ctx, params); err != nil {
		panic(err)
	}
}

// exportGenesis exports the module's state to genesis
func (am AppModule) exportGenesis(ctx sdk.Context) *pb.GenesisState {
	return &pb.GenesisState{
		Params: nil, // Will use proto params when available
	}
}

// BeginBlock performs module-specific logic at the beginning of each block
func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Run scheduled pre-validations
	if err := am.keeper.RunScheduler(sdkCtx); err != nil {
		am.keeper.Logger(sdkCtx).Error("failed to run scheduler", "error", err)
	}

	// Clean up expired pre-validated transactions
	if err := am.keeper.CleanupExpiredTransactions(sdkCtx); err != nil {
		am.keeper.Logger(sdkCtx).Error("failed to cleanup expired transactions", "error", err)
	}

	// Update metrics
	if err := am.keeper.UpdateMetrics(sdkCtx); err != nil {
		am.keeper.Logger(sdkCtx).Error("failed to update metrics", "error", err)
	}

	return nil
}

// EndBlock performs module-specific logic at the end of each block
func (am AppModule) EndBlock(ctx context.Context) error {
	return nil
}

// RegisterInvariants registers the module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Invariants registration commented out - needs implementation
	// keeper.RegisterInvariants(ir, am.keeper)
}
