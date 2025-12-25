// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ authproto.MsgServer = msgServer{}

type msgServer struct {
	authproto.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl creates a new message server implementation
func NewMsgServerImpl(keeper *Keeper) authproto.MsgServer {
	return &msgServer{Keeper: keeper}
}

// CreateRole creates a new role
func (ms msgServer) CreateRole(goCtx context.Context, msg *authproto.MsgCreateRole) (*authproto.MsgCreateRoleResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Creator == "" {
		return nil, status.Error(codes.InvalidArgument, "creator cannot be empty")
	}

	if msg.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "role name cannot be empty")
	}

	if len(msg.Permissions) == 0 {
		return nil, status.Error(codes.InvalidArgument, "permissions cannot be empty")
	}

	role, err := ms.Keeper.CreateRole(goCtx, msg.Creator, msg.Name, msg.Permissions, msg.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgCreateRoleResponse{Role: role}, nil
}

// AssignRole assigns a role to an address
func (ms msgServer) AssignRole(goCtx context.Context, msg *authproto.MsgAssignRole) (*authproto.MsgAssignRoleResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Assigner == "" {
		return nil, status.Error(codes.InvalidArgument, "assigner cannot be empty")
	}

	if msg.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	if msg.RoleName == "" {
		return nil, status.Error(codes.InvalidArgument, "role name cannot be empty")
	}

	expiresInSeconds := uint64(0)
	if msg.ExpiresInSeconds > 0 {
		expiresInSeconds = uint64(msg.ExpiresInSeconds)
	}

	assignment, err := ms.Keeper.AssignRole(goCtx, msg.Assigner, msg.Address, msg.RoleName, expiresInSeconds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgAssignRoleResponse{Assignment: assignment}, nil
}

// RevokeRole revokes a role from an address
func (ms msgServer) RevokeRole(goCtx context.Context, msg *authproto.MsgRevokeRole) (*authproto.MsgRevokeRoleResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Revoker == "" {
		return nil, status.Error(codes.InvalidArgument, "revoker cannot be empty")
	}

	if msg.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	if msg.RoleName == "" {
		return nil, status.Error(codes.InvalidArgument, "role name cannot be empty")
	}

	err := ms.Keeper.RevokeRole(goCtx, msg.Revoker, msg.Address, msg.RoleName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgRevokeRoleResponse{Success: true}, nil
}

// CreateMultisigWallet creates a new multisig wallet
func (ms msgServer) CreateMultisigWallet(goCtx context.Context, msg *authproto.MsgCreateMultisigWallet) (*authproto.MsgCreateMultisigWalletResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Creator == "" {
		return nil, status.Error(codes.InvalidArgument, "creator cannot be empty")
	}

	if len(msg.Signers) == 0 {
		return nil, status.Error(codes.InvalidArgument, "signers cannot be empty")
	}

	if msg.Threshold == 0 {
		return nil, status.Error(codes.InvalidArgument, "threshold must be greater than 0")
	}

	walletType := msg.WalletType.String()
	wallet, err := ms.Keeper.CreateMultisigWallet(goCtx, msg.Creator, msg.Signers, msg.Threshold, walletType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgCreateMultisigWalletResponse{Wallet: wallet}, nil
}

// CreateMultisigProposal creates a new multisig proposal
func (ms msgServer) CreateMultisigProposal(goCtx context.Context, msg *authproto.MsgCreateMultisigProposal) (*authproto.MsgCreateMultisigProposalResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Proposer == "" {
		return nil, status.Error(codes.InvalidArgument, "proposer cannot be empty")
	}

	if msg.WalletId == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet id cannot be empty")
	}

	if msg.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title cannot be empty")
	}

	if len(msg.Payload) == 0 {
		return nil, status.Error(codes.InvalidArgument, "payload cannot be empty")
	}

	expiresInSeconds := uint64(0)
	if msg.ExpiresInSeconds > 0 {
		expiresInSeconds = uint64(msg.ExpiresInSeconds)
	}

	proposal, err := ms.Keeper.CreateMultisigProposal(goCtx, msg.Proposer, msg.WalletId, msg.Title, msg.Description, msg.Payload, expiresInSeconds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgCreateMultisigProposalResponse{Proposal: proposal}, nil
}

// SignMultisigProposal signs a multisig proposal
func (ms msgServer) SignMultisigProposal(goCtx context.Context, msg *authproto.MsgSignMultisigProposal) (*authproto.MsgSignMultisigProposalResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Signer == "" {
		return nil, status.Error(codes.InvalidArgument, "signer cannot be empty")
	}

	if msg.ProposalId == "" {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be empty")
	}

	proposal, err := ms.Keeper.SignMultisigProposal(goCtx, msg.Signer, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgSignMultisigProposalResponse{Proposal: proposal}, nil
}

// ExecuteMultisigProposal executes an approved multisig proposal
func (ms msgServer) ExecuteMultisigProposal(goCtx context.Context, msg *authproto.MsgExecuteMultisigProposal) (*authproto.MsgExecuteMultisigProposalResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Executor == "" {
		return nil, status.Error(codes.InvalidArgument, "executor cannot be empty")
	}

	if msg.ProposalId == "" {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be empty")
	}

	err := ms.Keeper.ExecuteMultisigProposal(goCtx, msg.Executor, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgExecuteMultisigProposalResponse{Success: true}, nil
}

// ProposeTimeLockedAction proposes a time-locked admin action
func (ms msgServer) ProposeTimeLockedAction(goCtx context.Context, msg *authproto.MsgProposeTimeLockedAction) (*authproto.MsgProposeTimeLockedActionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Proposer == "" {
		return nil, status.Error(codes.InvalidArgument, "proposer cannot be empty")
	}

	if msg.ActionType == "" {
		return nil, status.Error(codes.InvalidArgument, "action type cannot be empty")
	}

	if len(msg.Payload) == 0 {
		return nil, status.Error(codes.InvalidArgument, "payload cannot be empty")
	}

	if msg.DelaySeconds == 0 {
		return nil, status.Error(codes.InvalidArgument, "delay seconds must be greater than 0")
	}

	action, err := ms.Keeper.ProposeTimeLockedAction(goCtx, msg.Proposer, msg.ActionType, msg.Payload, msg.DelaySeconds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgProposeTimeLockedActionResponse{Action: action}, nil
}

// ExecuteTimeLockedAction executes a ready time-locked action
func (ms msgServer) ExecuteTimeLockedAction(goCtx context.Context, msg *authproto.MsgExecuteTimeLockedAction) (*authproto.MsgExecuteTimeLockedActionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Executor == "" {
		return nil, status.Error(codes.InvalidArgument, "executor cannot be empty")
	}

	if msg.ActionId == "" {
		return nil, status.Error(codes.InvalidArgument, "action id cannot be empty")
	}

	err := ms.Keeper.ExecuteTimeLockedAction(goCtx, msg.Executor, msg.ActionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgExecuteTimeLockedActionResponse{Success: true}, nil
}

// CancelTimeLockedAction cancels a pending time-locked action
func (ms msgServer) CancelTimeLockedAction(goCtx context.Context, msg *authproto.MsgCancelTimeLockedAction) (*authproto.MsgCancelTimeLockedActionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Canceller == "" {
		return nil, status.Error(codes.InvalidArgument, "canceler cannot be empty")
	}

	if msg.ActionId == "" {
		return nil, status.Error(codes.InvalidArgument, "action id cannot be empty")
	}

	err := ms.Keeper.CancelTimeLockedAction(goCtx, msg.Canceller, msg.ActionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgCancelTimeLockedActionResponse{Success: true}, nil
}

// ActivateEmergencyAdmin activates an emergency admin
func (ms msgServer) ActivateEmergencyAdmin(goCtx context.Context, msg *authproto.MsgActivateEmergencyAdmin) (*authproto.MsgActivateEmergencyAdminResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Activator == "" {
		return nil, status.Error(codes.InvalidArgument, "activator cannot be empty")
	}

	if msg.AdminAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "admin address cannot be empty")
	}

	if len(msg.Privileges) == 0 {
		return nil, status.Error(codes.InvalidArgument, "privileges cannot be empty")
	}

	expiresInSeconds := uint64(0)
	if msg.ExpiresInSeconds > 0 {
		expiresInSeconds = uint64(msg.ExpiresInSeconds)
	}

	admin, err := ms.Keeper.ActivateEmergencyAdmin(goCtx, msg.Activator, msg.AdminAddress, msg.Privileges, expiresInSeconds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgActivateEmergencyAdminResponse{Admin: admin}, nil
}

// DeactivateEmergencyAdmin deactivates an emergency admin
func (ms msgServer) DeactivateEmergencyAdmin(goCtx context.Context, msg *authproto.MsgDeactivateEmergencyAdmin) (*authproto.MsgDeactivateEmergencyAdminResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Deactivator == "" {
		return nil, status.Error(codes.InvalidArgument, "deactivator cannot be empty")
	}

	if msg.AdminAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "admin address cannot be empty")
	}

	err := ms.Keeper.DeactivateEmergencyAdmin(goCtx, msg.Deactivator, msg.AdminAddress)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgDeactivateEmergencyAdminResponse{Success: true}, nil
}

// InitiateValidatorKeyRotation initiates validator key rotation
func (ms msgServer) InitiateValidatorKeyRotation(goCtx context.Context, msg *authproto.MsgInitiateValidatorKeyRotation) (*authproto.MsgInitiateValidatorKeyRotationResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Initiator == "" {
		return nil, status.Error(codes.InvalidArgument, "initiator cannot be empty")
	}

	if msg.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	if msg.NewConsensusPubkey == "" {
		return nil, status.Error(codes.InvalidArgument, "new public key cannot be empty")
	}

	rotation, err := ms.Keeper.InitiateValidatorKeyRotation(goCtx, msg.Initiator, msg.ValidatorAddress, []byte(msg.NewConsensusPubkey))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgInitiateValidatorKeyRotationResponse{Rotation: rotation}, nil
}

// CompleteValidatorKeyRotation completes validator key rotation
func (ms msgServer) CompleteValidatorKeyRotation(goCtx context.Context, msg *authproto.MsgCompleteValidatorKeyRotation) (*authproto.MsgCompleteValidatorKeyRotationResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Completer == "" {
		return nil, status.Error(codes.InvalidArgument, "completer cannot be empty")
	}

	if msg.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	err := ms.Keeper.CompleteValidatorKeyRotation(goCtx, msg.Completer, msg.ValidatorAddress)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgCompleteValidatorKeyRotationResponse{Success: true}, nil
}

// CreateSession creates a new API session
func (ms msgServer) CreateSession(goCtx context.Context, msg *authproto.MsgCreateSession) (*authproto.MsgCreateSessionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.UserAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "user address cannot be empty")
	}

	// Default session duration (e.g., 24 hours)
	expiresInSeconds := uint64(86400)

	session, err := ms.Keeper.CreateSession(goCtx, msg.UserAddress, expiresInSeconds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgCreateSessionResponse{Session: session}, nil
}

// RevokeSession revokes an active session
func (ms msgServer) RevokeSession(goCtx context.Context, msg *authproto.MsgRevokeSession) (*authproto.MsgRevokeSessionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.UserAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "user address cannot be empty")
	}

	if msg.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session id cannot be empty")
	}

	err := ms.Keeper.RevokeSession(goCtx, msg.UserAddress, msg.SessionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.MsgRevokeSessionResponse{Success: true}, nil
}
