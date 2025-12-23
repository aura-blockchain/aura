package monitoring

//lint:file-ignore SA1019 -- monitoring module still registers legacy invariants until SDK migration

import (
	"bytes"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/monitoring/keeper"
	"github.com/aequitas/aura/chain/x/monitoring/types"
	monitoringpb "github.com/aequitas/aura/proto/aura/monitoring/v1beta1"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic implements the AppModuleBasic interface
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
	genesis := types.DefaultGenesisState()
	data, err := json.Marshal(genesis)
	if err != nil {
		panic(err)
	}
	return data
}

// ValidateGenesis performs genesis state validation
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bytes.TrimSpace(bz)) == 0 {
		return nil
	}
	var genesis types.GenesisState
	if err := json.Unmarshal(bz, &genesis); err != nil {
		return err
	}
	return genesis.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC Gateway routes if needed
}

// AppModule implements the AppModule interface
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

// NewAppModule creates a new monitoring module
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
	monitoringpb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	monitoringpb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// InitGenesis initializes module state from genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesis types.GenesisState
	if err := json.Unmarshal(data, &genesis); err != nil {
		panic(err)
	}

	if err := am.keeper.InitGenesis(ctx, &genesis); err != nil {
		panic(err)
	}
}

// ExportGenesis exports module state for genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := am.keeper.ExportGenesis(ctx)
	data, err := json.Marshal(*genesis)
	if err != nil {
		panic(err)
	}
	return data
}

// ConsensusVersion returns the consensus state-breaking version
func (AppModule) ConsensusVersion() uint64 { return 1 }

// RegisterInvariants registers the module's invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Invariants registration commented out - needs implementation
	// keeper.RegisterInvariants(ir, am.keeper)
}

// GetKeeper returns the module's keeper
func (am AppModule) GetKeeper() *keeper.Keeper {
	return am.keeper
}

// BeginBlock is called at the start of each block
func (am AppModule) BeginBlock(ctx sdk.Context) error {
	return am.keeper.BeginBlocker(ctx)
}
