// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/auth/keeper"
	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module used by the auth module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the auth module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the auth module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the auth module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the auth module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesis); err != nil {
		return err
	}
	return types.ValidateGenesis(&genesis)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the auth module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC Gateway routes if needed
	// authproto.RegisterQueryHandlerClient(context.Background(), mux, authproto.NewQueryClient(clientCtx))
}

// AppModule implements the AppModule interface for the auth module.
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

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services with the configurator.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	authproto.RegisterMsgServer(cfg.MsgServer(), NewMsgServer(am.keeper))
	authproto.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs genesis initialization for the auth module.
// It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesis)

	// Initialize keeper state from genesis
	if err := am.keeper.InitGenesis(ctx, &genesis); err != nil {
		panic(err)
	}

	// Initialize default roles if no roles exist
	if len(genesis.Roles) == 0 {
		if err := am.keeper.InitializeDefaultRoles(ctx); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the auth module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genesis)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ============================================================================
// Message Server Implementation
// ============================================================================

// MsgServer implements the auth Msg service
type MsgServer struct {
	keeper *keeper.Keeper
	authproto.UnimplementedMsgServer
}

// NewMsgServer creates a new MsgServer
func NewMsgServer(keeper *keeper.Keeper) authproto.MsgServer {
	return &MsgServer{keeper: keeper}
}

// CreateRole creates a new role
func (s *MsgServer) CreateRole(ctx context.Context, req *authproto.MsgCreateRole) (*authproto.MsgCreateRoleResponse, error) {
	role, err := s.keeper.CreateRole(ctx, req.Creator, req.Name, req.Permissions, req.Description)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCreateRoleResponse{Role: role}, nil
}

// AssignRole assigns a role to an address
func (s *MsgServer) AssignRole(ctx context.Context, req *authproto.MsgAssignRole) (*authproto.MsgAssignRoleResponse, error) {
	assignment, err := s.keeper.AssignRole(ctx, req.Assigner, req.Address, req.RoleName, uint64(req.ExpiresInSeconds))
	if err != nil {
		return nil, err
	}
	return &authproto.MsgAssignRoleResponse{Assignment: assignment}, nil
}

// RevokeRole revokes a role from an address
func (s *MsgServer) RevokeRole(ctx context.Context, req *authproto.MsgRevokeRole) (*authproto.MsgRevokeRoleResponse, error) {
	err := s.keeper.RevokeRole(ctx, req.Revoker, req.Address, req.RoleName)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgRevokeRoleResponse{Success: true}, nil
}

// CreateMultisigWallet creates a new multisig wallet
func (s *MsgServer) CreateMultisigWallet(ctx context.Context, req *authproto.MsgCreateMultisigWallet) (*authproto.MsgCreateMultisigWalletResponse, error) {
	wallet, err := s.keeper.CreateMultisigWallet(ctx, req.Creator, req.Signers, req.Threshold, req.WalletType.String())
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCreateMultisigWalletResponse{Wallet: wallet}, nil
}

// CreateMultisigProposal creates a new multisig proposal
func (s *MsgServer) CreateMultisigProposal(ctx context.Context, req *authproto.MsgCreateMultisigProposal) (*authproto.MsgCreateMultisigProposalResponse, error) {
	proposal, err := s.keeper.CreateMultisigProposal(ctx, req.Proposer, req.WalletId, req.Title, req.Description, req.Payload, uint64(req.ExpiresInSeconds))
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCreateMultisigProposalResponse{Proposal: proposal}, nil
}

// SignMultisigProposal signs a multisig proposal
func (s *MsgServer) SignMultisigProposal(ctx context.Context, req *authproto.MsgSignMultisigProposal) (*authproto.MsgSignMultisigProposalResponse, error) {
	proposal, err := s.keeper.SignMultisigProposal(ctx, req.Signer, req.ProposalId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgSignMultisigProposalResponse{Proposal: proposal}, nil
}

// ExecuteMultisigProposal executes an approved multisig proposal
func (s *MsgServer) ExecuteMultisigProposal(ctx context.Context, req *authproto.MsgExecuteMultisigProposal) (*authproto.MsgExecuteMultisigProposalResponse, error) {
	err := s.keeper.ExecuteMultisigProposal(ctx, req.Executor, req.ProposalId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgExecuteMultisigProposalResponse{Success: true}, nil
}

// ProposeTimeLockedAction proposes a time-locked admin action
func (s *MsgServer) ProposeTimeLockedAction(ctx context.Context, req *authproto.MsgProposeTimeLockedAction) (*authproto.MsgProposeTimeLockedActionResponse, error) {
	action, err := s.keeper.ProposeTimeLockedAction(ctx, req.Proposer, req.ActionType, req.Payload, req.DelaySeconds)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgProposeTimeLockedActionResponse{Action: action}, nil
}

// ExecuteTimeLockedAction executes a ready time-locked action
func (s *MsgServer) ExecuteTimeLockedAction(ctx context.Context, req *authproto.MsgExecuteTimeLockedAction) (*authproto.MsgExecuteTimeLockedActionResponse, error) {
	err := s.keeper.ExecuteTimeLockedAction(ctx, req.Executor, req.ActionId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgExecuteTimeLockedActionResponse{Success: true}, nil
}

// CancelTimeLockedAction cancels a pending time-locked action
func (s *MsgServer) CancelTimeLockedAction(ctx context.Context, req *authproto.MsgCancelTimeLockedAction) (*authproto.MsgCancelTimeLockedActionResponse, error) {
	err := s.keeper.CancelTimeLockedAction(ctx, req.Canceller, req.ActionId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCancelTimeLockedActionResponse{Success: true}, nil
}

// ActivateEmergencyAdmin activates an emergency admin
func (s *MsgServer) ActivateEmergencyAdmin(ctx context.Context, req *authproto.MsgActivateEmergencyAdmin) (*authproto.MsgActivateEmergencyAdminResponse, error) {
	admin, err := s.keeper.ActivateEmergencyAdmin(ctx, req.Activator, req.AdminAddress, req.Privileges, uint64(req.ExpiresInSeconds))
	if err != nil {
		return nil, err
	}
	return &authproto.MsgActivateEmergencyAdminResponse{Admin: admin}, nil
}

// DeactivateEmergencyAdmin deactivates an emergency admin
func (s *MsgServer) DeactivateEmergencyAdmin(ctx context.Context, req *authproto.MsgDeactivateEmergencyAdmin) (*authproto.MsgDeactivateEmergencyAdminResponse, error) {
	err := s.keeper.DeactivateEmergencyAdmin(ctx, req.Deactivator, req.AdminAddress)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgDeactivateEmergencyAdminResponse{Success: true}, nil
}

// InitiateValidatorKeyRotation initiates validator key rotation
func (s *MsgServer) InitiateValidatorKeyRotation(ctx context.Context, req *authproto.MsgInitiateValidatorKeyRotation) (*authproto.MsgInitiateValidatorKeyRotationResponse, error) {
	rotation, err := s.keeper.InitiateValidatorKeyRotation(ctx, req.Initiator, req.ValidatorAddress, []byte(req.NewConsensusPubkey))
	if err != nil {
		return nil, err
	}
	return &authproto.MsgInitiateValidatorKeyRotationResponse{Rotation: rotation}, nil
}

// CompleteValidatorKeyRotation completes validator key rotation
func (s *MsgServer) CompleteValidatorKeyRotation(ctx context.Context, req *authproto.MsgCompleteValidatorKeyRotation) (*authproto.MsgCompleteValidatorKeyRotationResponse, error) {
	err := s.keeper.CompleteValidatorKeyRotation(ctx, req.Completer, req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCompleteValidatorKeyRotationResponse{Success: true}, nil
}

// CreateSession creates a new API session
func (s *MsgServer) CreateSession(ctx context.Context, req *authproto.MsgCreateSession) (*authproto.MsgCreateSessionResponse, error) {
	// Use default session expiry of 24 hours
	session, err := s.keeper.CreateSession(ctx, req.UserAddress, 86400)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCreateSessionResponse{Session: session}, nil
}

// RevokeSession revokes an active session
func (s *MsgServer) RevokeSession(ctx context.Context, req *authproto.MsgRevokeSession) (*authproto.MsgRevokeSessionResponse, error) {
	err := s.keeper.RevokeSession(ctx, req.UserAddress, req.SessionId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgRevokeSessionResponse{Success: true}, nil
}

// NOTE: QueryServer implementation has been moved to keeper/query_server.go
// This eliminates duplicate query server code and uses the production-ready
// implementation with proper KVStore access and SDK context handling.
// All query methods are now properly implemented in:
// /home/decri/blockchain-projects/aura/chain/x/auth/keeper/query_server.go
