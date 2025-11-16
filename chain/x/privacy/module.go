package privacy

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/privacy/keeper"
	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module used by the privacy module
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the privacy module's name
func (AppModuleBasic) Name() string {
	return "privacy"
}

// RegisterLegacyAminoCodec registers the privacy module's types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register types here
}

// RegisterInterfaces registers the module's interface types
func (a AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	// Register interfaces here
}

// DefaultGenesis returns default genesis state as raw bytes for the privacy module
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&privacyproto.GenesisState{
		Params: &privacyproto.Params{
			MinAnonymitySetSize:         3,
			MaxMixingPoolSize:           100,
			MixingFee:                   "1000",
			ViewKeyRotationPeriodDays:   90,
			EnableStealthAddresses:      true,
			EnableConfidentialTransfers: true,
			EnableZkProofs:              true,
		},
	})
}

// ValidateGenesis performs genesis state validation for the privacy module
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState privacyproto.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", "privacy", err)
	}
	return ValidateGenesisState(&genState)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes here
}

// AppModule implements an application module for the privacy module
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

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Register services here
}

// InitGenesis performs genesis initialization for the privacy module
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState privacyproto.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)

	// Initialize genesis state via keeper
	if err := am.keeper.InitGenesis(ctx, &genState); err != nil {
		panic(fmt.Sprintf("failed to initialize privacy genesis: %v", err))
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the privacy module
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ValidateGenesisState performs basic validation on privacy genesis state
func ValidateGenesisState(gs *privacyproto.GenesisState) error {
	if gs.Params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	if gs.Params.MinAnonymitySetSize < 2 {
		return fmt.Errorf("min anonymity set size must be at least 2")
	}
	if gs.Params.MaxMixingPoolSize < gs.Params.MinAnonymitySetSize {
		return fmt.Errorf("max mixing pool size must be greater than or equal to min anonymity set size")
	}
	return nil
}
