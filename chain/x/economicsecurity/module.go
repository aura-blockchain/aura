package economicsecurity

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/economicsecurity/keeper"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the module's types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesis); err != nil {
		return err
	}
	return types.ValidateGenesis(&genesis)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC Gateway routes if needed
}

// AppModule implements the app module interface
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule instance
func NewAppModule(cdc codec.Codec, k *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface
func (am AppModule) IsAppModule() {}

// RegisterServices registers the module's message and query servers
func (am AppModule) RegisterServices(cfg module.Configurator) {
	economicsecuritypb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	economicsecuritypb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// InitGenesis initializes module state from genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesis)

	// Initialize keeper state from genesis
	if err := am.keeper.InitGenesis(ctx, genesis); err != nil {
		panic(err)
	}
}

// ExportGenesis exports module state for genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return cdc.MustMarshalJSON(&genesis)
}

// ConsensusVersion returns the consensus state-breaking version
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes all ABCI BeginBlock logic
// CONSENSUS-SAFE: Uses deterministic block time from context
func (am AppModule) BeginBlock(ctx context.Context) error {
	// Unwrap SDK context to get deterministic block time and height
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := uint64(sdkCtx.BlockHeight())
	blockTime := sdkCtx.BlockTime().Unix()

	// Set deterministic block state
	if err := am.keeper.SetCurrentHeight(sdkCtx, height); err != nil {
		return fmt.Errorf("failed to record block height: %w", err)
	}
	if err := am.keeper.SetCurrentTime(sdkCtx, blockTime); err != nil {
		return fmt.Errorf("failed to record block time: %w", err)
	}

	// Check inflation periodically - commented out as method not available
	// params := am.keeper.GetParams()
	// if height%params.InflationCheckInterval == 0 {
	//     am.keeper.CheckInflation()
	// }

	// Adjust dynamic fees based on recent utilization - commented out as method not available
	// am.keeper.AdjustDynamicFees()

	return nil
}

// EndBlock executes all ABCI EndBlock logic
func (am AppModule) EndBlock(ctx context.Context) error {
	return nil
}
