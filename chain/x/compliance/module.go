package compliance

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

	"github.com/aequitas/aura/chain/x/compliance/client/cli"
	"github.com/aequitas/aura/chain/x/compliance/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

// AppModuleBasic implements the AppModuleBasic interface for compliance.
type AppModuleBasic struct{}

// Name returns compliance module name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers legacy amino types.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers protobuf interfaces.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterGRPCGatewayRoutes registers REST routes.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	_ = clientCtx
	_ = mux
}

// DefaultGenesis returns default genesis.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis validates the module genesis data.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bytes.TrimSpace(bz)) == 0 {
		return nil
	}
	var gen types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gen); err != nil {
		return err
	}
	return types.ValidateGenesis(&gen)
}

// GetTxCmd returns the tx command.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the query command.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements the compliance module interface.
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

// NewAppModule returns a new AppModule instance.
func NewAppModule(keeper *keeper.Keeper) AppModule {
	return AppModule{AppModuleBasic: AppModuleBasic{}, keeper: keeper}
}

// Name returns the module name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterServices registers gRPC services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	compliancepb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	compliancepb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// InitGenesis initializes genesis state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesis *types.GenesisState
	if len(bytes.TrimSpace(data)) == 0 {
		genesis = types.DefaultGenesis()
	} else {
		var decoded types.GenesisState
		if err := cdc.UnmarshalJSON(data, &decoded); err != nil {
			panic(fmt.Errorf("compliance module failed to unmarshal genesis: %w", err))
		}
		genesis = &decoded
	}
	if err := am.keeper.InitGenesis(ctx, genesis); err != nil {
		panic(fmt.Errorf("compliance InitGenesis: %w", err))
	}
}

// ExportGenesis exports module state.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gen := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gen)
}

// ConsensusVersion returns the module consensus version.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// RegisterInvariants registers compliance invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// TODO: add invariants if needed
}

// BeginBlock processes expired KYC records at the start of each block.
// This automatically marks KYC records as expired and emits events for monitoring.
//
// Processing includes:
//   - Iterate through all KYC records
//   - Check if expired (BlockTime > ExpiresAt)
//   - Emit EventTypeKYCExpired for each expired record
//
// Security considerations:
//   - Read-only iteration (does not modify state)
//   - Gas-bounded: Limited processing per block
//   - Time-based: Uses blockchain time (trustworthy)
//   - Events provide audit trail for compliance
//
// Compliance:
//   - FinCEN: Continuous monitoring of verification status
//   - FATF: Ongoing due diligence enforcement
//   - Provides immutable audit trail of expiry events
//
// Performance notes:
//   - O(n) where n is total KYC records
//   - Consider pagination if records exceed gas limit
//   - Events are indexed for efficient queries
func (am AppModule) BeginBlock(ctx sdk.Context) {
	am.keeper.BeginBlocker(ctx)
}

// EndBlock is a no-op.
func (AppModule) EndBlock(ctx sdk.Context) {}

// IsAppModule tags module for compatibility.
func (AppModule) IsAppModule() {}
