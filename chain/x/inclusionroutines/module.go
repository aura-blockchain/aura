package inclusionroutines

import (
	"bytes"
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
	"github.com/spf13/cobra"

	"github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic module functions.
type AppModuleBasic struct{}

// Name implements AppModuleBasic.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers legacy Amino codecs.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers proto interfaces.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterGRPCGatewayRoutes registers REST routes.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	_ = clientCtx
	_ = mux
}

// DefaultGenesis returns the default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis validates the genesis state.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bytes.TrimSpace(bz)) == 0 {
		return nil
	}
	var gen types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gen); err != nil {
		return fmt.Errorf("failed to unmarshal inclusionroutines genesis: %w", err)
	}
	return types.ValidateGenesisState(&gen)
}

// GetTxCmd returns nil.
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// GetQueryCmd returns nil.
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }

// AppModule implements the module interface.
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

// NewAppModule returns a new AppModule instance.
func NewAppModule(keeper *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// Name returns module name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	inclusionroutinespb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	inclusionroutinespb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// InitGenesis initializes module genesis state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesis types.GenesisState
	if len(bytes.TrimSpace(data)) == 0 {
		genesis = *types.DefaultGenesisState()
	} else if err := cdc.UnmarshalJSON(data, &genesis); err != nil {
		panic(fmt.Errorf("inclusionroutines InitGenesis unmarshal: %w", err))
	}
	if err := am.keeper.InitGenesis(ctx, genesis); err != nil {
		panic(fmt.Errorf("inclusionroutines InitGenesis: %w", err))
	}
}

// ExportGenesis returns module genesis.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gen := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(&gen)
}

// ConsensusVersion returns module version.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock is no-op.
func (AppModule) BeginBlock(ctx context.Context) error {
	_ = sdk.UnwrapSDKContext(ctx)
	return nil
}

// EndBlock returns no-op.
func (AppModule) EndBlock(ctx context.Context) error {
	_ = sdk.UnwrapSDKContext(ctx)
	return nil
}

// RegisterInvariants registers invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// IsAppModule makes it compatible.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType marks the module as a depinject singleton.
func (AppModule) IsOnePerModuleType() {}
