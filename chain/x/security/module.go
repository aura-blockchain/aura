// Package security implements the consolidated security module for the Aura blockchain.
// This module consolidates: networksecurity, validatorsecurity, walletsecurity,
// incidentresponse, cryptography, and privacy into a unified security layer.
package security

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
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/aequitas/aura/chain/x/security/keeper"
	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

var (
	_ module.AppModuleBasic     = AppModuleBasic{}
	_ module.HasGenesis         = AppModule{}
	_ module.HasServices        = AppModule{}
	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
	_ appmodule.HasEndBlocker   = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic defines the basic application module used by the security module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the security module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the security module's types on the LegacyAmino codec.
// In SDK v0.50+, this is primarily for backwards compatibility.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types if needed for amino JSON serialization
	// Currently empty as we primarily use protobuf
}

// RegisterInterfaces registers the module's interface types with the interface registry.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &securitypb.Msg_ServiceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&securitypb.MsgAddTrustedPeer{},
		&securitypb.MsgRemoveTrustedPeer{},
		&securitypb.MsgBanPeer{},
		&securitypb.MsgUnbanPeer{},
		&securitypb.MsgUpdatePeerReputation{},
		&securitypb.MsgResolveForkAlert{},
		&securitypb.MsgResolvePartitionAlert{},
		&securitypb.MsgRegisterValidatorSecurity{},
		&securitypb.MsgUpdateValidatorSecurity{},
		&securitypb.MsgRegisterSentryNode{},
		&securitypb.MsgRemoveSentryNode{},
		&securitypb.MsgReportDoubleSign{},
		&securitypb.MsgAcknowledgeValidatorAlert{},
		&securitypb.MsgTriggerFailover{},
		&securitypb.MsgRegisterHardwareWallet{},
		&securitypb.MsgCreateMultiSigWallet{},
		&securitypb.MsgProposeMultiSigTransaction{},
		&securitypb.MsgSignMultiSigTransaction{},
		&securitypb.MsgExecuteMultiSigTransaction{},
		&securitypb.MsgConfigureSocialRecovery{},
		&securitypb.MsgInitiateRecovery{},
		&securitypb.MsgApproveRecovery{},
		&securitypb.MsgExecuteRecovery{},
		&securitypb.MsgSetSpendingLimits{},
		&securitypb.MsgRegisterBiometric{},
		&securitypb.MsgCreateIncident{},
		&securitypb.MsgUpdateIncident{},
		&securitypb.MsgResolveIncident{},
		&securitypb.MsgExecuteResponseAction{},
		&securitypb.MsgAddAuditLogEntry{},
		&securitypb.MsgCreateKeyRotationSchedule{},
		&securitypb.MsgRotateKey{},
		&securitypb.MsgCreateThresholdScheme{},
		&securitypb.MsgSubmitThresholdSignatureShare{},
		&securitypb.MsgRegisterZKProofCircuit{},
		&securitypb.MsgSubmitZKProof{},
		&securitypb.MsgGenerateQuantumResistantKey{},
		&securitypb.MsgCreateMixingPool{},
		&securitypb.MsgJoinMixingPool{},
		&securitypb.MsgExecuteMixing{},
		&securitypb.MsgGenerateStealthAddress{},
		&securitypb.MsgCreateRingSignature{},
		&securitypb.MsgCreateConfidentialTransaction{},
		&securitypb.MsgUpdateParams{},
	)
}

// DefaultGenesis returns default genesis state as raw bytes for the security module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the security module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bytes.TrimSpace(bz)) == 0 {
		return nil
	}
	var genState securitypb.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return types.ValidateGenesisState(&genState)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the security module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC Gateway routes when protobuf query service is available
	// Example:
	// if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
	// 	panic(fmt.Sprintf("failed to register gRPC gateway routes: %s", err))
	// }
}

// GetTxCmd returns the root tx command for the security module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	// Return the root tx command when CLI commands are implemented
	// For now, return nil to indicate no CLI commands available yet
	return nil
}

// GetQueryCmd returns the root query command for the security module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	// Return the root query command when CLI commands are implemented
	// For now, return nil to indicate no CLI commands available yet
	return nil
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements an application module for the security module.
type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(cdc codec.Codec, keeper *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the security module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Register message server
	securitypb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))

	// Register query server
	securitypb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// RegisterInvariants registers the security module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Register invariants for runtime checks
	// Example: keeper.RegisterInvariants(ir, am.keeper)
}

// InitGenesis performs genesis initialization for the security module.
// It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState securitypb.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	// Validate genesis state before initialization
	if err := types.ValidateGenesisState(&genesisState); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %v", types.ModuleName, err))
	}

	// Initialize module state
	am.keeper.InitGenesis(ctx, &genesisState)

	am.keeper.Logger(ctx).Info(
		"security module genesis initialized",
		"module", types.ModuleName,
	)
}

// ExportGenesis returns the exported genesis state as raw bytes for the security module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes all ABCI BeginBlock logic for the security module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := am.keeper.Logger(sdkCtx)

	// Process cryptographic key rotations
	if err := am.processKeyRotations(sdkCtx); err != nil {
		logger.Error("failed to process key rotations", "error", err)
		// Don't fail the block, just log the error
	}

	// Update network security metrics
	if err := am.updateNetworkMetrics(sdkCtx); err != nil {
		logger.Error("failed to update network metrics", "error", err)
	}

	// Check validator security health
	if err := am.checkValidatorSecurity(sdkCtx); err != nil {
		logger.Error("failed to check validator security", "error", err)
	}

	// Process wallet security checks
	if err := am.processWalletSecurity(sdkCtx); err != nil {
		logger.Error("failed to process wallet security", "error", err)
	}

	// Update incident response state
	if err := am.updateIncidentState(sdkCtx); err != nil {
		logger.Error("failed to update incident state", "error", err)
	}

	// Refresh privacy pools
	if err := am.refreshPrivacyPools(sdkCtx); err != nil {
		logger.Error("failed to refresh privacy pools", "error", err)
	}

	return nil
}

// EndBlock executes all ABCI EndBlock logic for the security module.
func (am AppModule) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := am.keeper.Logger(sdkCtx)

	// Clean up expired sessions
	if err := am.cleanupExpiredSessions(sdkCtx); err != nil {
		logger.Error("failed to cleanup expired sessions", "error", err)
	}

	// Finalize security metrics for the block
	if err := am.finalizeSecurityMetrics(sdkCtx); err != nil {
		logger.Error("failed to finalize security metrics", "error", err)
	}

	return nil
}

// ----------------------------------------------------------------------------
// BeginBlock/EndBlock helper functions
// ----------------------------------------------------------------------------

// processKeyRotations handles scheduled cryptographic key rotations
func (am AppModule) processKeyRotations(ctx sdk.Context) error {
	// Process any scheduled key rotations for cryptographic operations
	schedules := am.keeper.GetAllKeyRotationSchedules(ctx)
	blockTime := ctx.BlockTime()

	for _, schedule := range schedules {
		if schedule.NextRotationTime != nil && schedule.Enabled {
			nextRotation := schedule.NextRotationTime.AsTime()
			if nextRotation.Before(blockTime) || nextRotation.Equal(blockTime) {
				// Key rotation logic would be implemented here
				// For now, just log that rotation is due
				am.keeper.Logger(ctx).Info(
					"key rotation due",
					"key_id", schedule.KeyId,
					"scheduled_time", nextRotation,
				)
			}
		}
	}
	return nil
}

// updateNetworkMetrics updates network security metrics
func (am AppModule) updateNetworkMetrics(ctx sdk.Context) error {
	// Update peer reputation scores, rate limits, etc.
	// This is a placeholder for actual metric calculation logic
	return nil
}

// checkValidatorSecurity checks validator security health
func (am AppModule) checkValidatorSecurity(ctx sdk.Context) error {
	// Check for double-signing, downtime, and other validator security issues
	// This is a placeholder for actual validator security checks
	return nil
}

// processWalletSecurity processes wallet security checks
func (am AppModule) processWalletSecurity(ctx sdk.Context) error {
	// Process pending multi-sig transactions, recovery requests, etc.
	// This is a placeholder for actual wallet security processing
	return nil
}

// updateIncidentState updates incident response state
func (am AppModule) updateIncidentState(ctx sdk.Context) error {
	// Update incident statuses, check for auto-resolution conditions, etc.
	// This is a placeholder for actual incident state updates
	return nil
}

// refreshPrivacyPools refreshes privacy mixing pools
func (am AppModule) refreshPrivacyPools(ctx sdk.Context) error {
	// Refresh mixing pools, clean up completed mixes, etc.
	// This is a placeholder for actual privacy pool management
	return nil
}

// cleanupExpiredSessions removes expired wallet sessions
func (am AppModule) cleanupExpiredSessions(ctx sdk.Context) error {
	sessions := am.keeper.GetAllSessions(ctx)
	blockTime := ctx.BlockTime()

	for _, session := range sessions {
		if session.ExpiresAt != nil && session.ExpiresAt.Before(blockTime) {
			// Remove expired session
			am.keeper.DeleteSession(ctx, session.Id)
			am.keeper.Logger(ctx).Debug(
				"cleaning up expired session",
				"session_id", session.Id,
				"expired_at", session.ExpiresAt,
			)
		}
	}
	return nil
}

// finalizeSecurityMetrics finalizes security metrics for the block
func (am AppModule) finalizeSecurityMetrics(ctx sdk.Context) error {
	// Aggregate security metrics, emit events, etc.
	// This is a placeholder for actual metrics finalization
	return nil
}
