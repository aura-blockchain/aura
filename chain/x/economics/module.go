package economics

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

	"github.com/aequitas/aura/chain/x/economics/keeper"
	"github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
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
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bytes.TrimSpace(bz)) == 0 {
		return nil
	}
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesis); err != nil {
		return err
	}
	return types.ValidateGenesisState(&genesis)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Gateway registration will be added after proto generation with grpc-gateway support
	// if err := economicspb.RegisterQueryHandlerClient(context.Background(), mux, economicspb.NewQueryClient(clientCtx)); err != nil {
	// 	panic(err)
	// }
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
	// Register query server
	economicspb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
	// Register message server
	economicspb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
}

// InitGenesis initializes module state from genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesis)

	// Initialize the keeper with genesis state
	if err := am.keeper.InitGenesis(ctx, &genesis); err != nil {
		panic(err)
	}
}

// ExportGenesis exports module state for genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return cdc.MustMarshalJSON(genesis)
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
		return err
	}
	if err := am.keeper.SetCurrentTime(sdkCtx, blockTime); err != nil {
		return err
	}

	// Process block utilization for dynamic fees
	// In production, would calculate actual block utilization
	utilization := uint64(50) // Placeholder
	if err := am.keeper.RecordBlockUtilization(sdkCtx, utilization); err != nil {
		return err
	}

	// Check and update proposal statuses
	// In production, would iterate through active proposals
	// For now, this is a placeholder for the logic
	params, err := am.keeper.GetParams(sdkCtx)
	if err != nil {
		return err
	}

	// Check inflation periodically
	if params.Tokenomics != nil && params.Tokenomics.InflationCheckInterval > 0 && height%params.Tokenomics.InflationCheckInterval == 0 {
		// Would implement inflation monitoring here
		// currentInflation := calculateInflation()
		// if abs(currentInflation - targetInflation) > threshold {
		//     createInflationAlert()
		// }
	}

	// Process vesting schedules that need release
	// In production, would check time-based vesting releases

	// Process expired vote locks
	// In production, would unlock tokens for expired locks

	return nil
}

// EndBlock executes all ABCI EndBlock logic
func (am AppModule) EndBlock(ctx context.Context) error {
	// Unwrap SDK context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Process any end-of-block operations
	// For example: finalize proposals, distribute rewards, etc.

	currentTime, err := am.keeper.GetCurrentTime(sdkCtx)
	if err != nil {
		return err
	}

	// Check for proposals that need status updates
	err = am.keeper.IterateProposals(sdkCtx, func(proposal *economicspb.Proposal) bool {
		// Update proposal status based on time
		if proposal.VotingEndTime != nil && proposal.Status == economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD && currentTime >= proposal.VotingEndTime.AsTime().Unix() {
			// Would call UpdateProposalStatus here
			_ = am.keeper.UpdateProposalStatus(sdkCtx, proposal.Id)
		}
		return false
	})

	if err != nil {
		return err
	}

	return nil
}
