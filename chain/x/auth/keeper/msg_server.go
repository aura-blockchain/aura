package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ authproto.MsgServer = msgServer{}

// msgServer implements the MsgServer interface
type msgServer struct {
	authproto.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) authproto.MsgServer {
	return &msgServer{Keeper: keeper}
}

// CreateRole creates a new role
func (ms msgServer) CreateRole(goCtx context.Context, msg *authproto.MsgCreateRole) (*authproto.MsgCreateRoleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Creator, types.PermissionCreateRole); err != nil {
		return nil, err
	}

	// Validate inputs
	if msg.Name == "" {
		return nil, types.ErrInvalidInput.Wrap("role name cannot be empty")
	}
	if len(msg.Permissions) == 0 {
		return nil, types.ErrInvalidInput.Wrap("role must have at least one permission")
	}

	// Create role
	role, err := ms.Keeper.CreateRole(ctx, msg.Name, msg.Permissions, msg.Description)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Creator, "create_role", msg.Name, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Creator, "create_role", msg.Name, "success", map[string]string{
		"permissions": fmt.Sprintf("%v", msg.Permissions),
	}, "")

	return &authproto.MsgCreateRoleResponse{Role: role}, nil
}

// AssignRole assigns a role to an address
func (ms msgServer) AssignRole(goCtx context.Context, msg *authproto.MsgAssignRole) (*authproto.MsgAssignRoleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate assigner has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Assigner, types.PermissionAssignRole); err != nil {
		return nil, err
	}

	// Validate inputs
	if msg.Address == "" || msg.RoleName == "" {
		return nil, types.ErrInvalidInput.Wrap("address and role name are required")
	}

	// Assign role
	assignment, err := ms.Keeper.AssignRole(ctx, msg.Address, msg.RoleName, msg.ExpiresInSeconds)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Assigner, "assign_role", msg.Address, "failure", map[string]string{
			"role": msg.RoleName,
		}, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Assigner, "assign_role", msg.Address, "success", map[string]string{
		"role": msg.RoleName,
	}, "")

	return &authproto.MsgAssignRoleResponse{Assignment: assignment}, nil
}

// RevokeRole revokes a role from an address
func (ms msgServer) RevokeRole(goCtx context.Context, msg *authproto.MsgRevokeRole) (*authproto.MsgRevokeRoleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate revoker has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Revoker, types.PermissionRevokeRole); err != nil {
		return nil, err
	}

	// Revoke role
	err := ms.Keeper.RevokeRole(ctx, msg.Address, msg.RoleName)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Revoker, "revoke_role", msg.Address, "failure", map[string]string{
			"role": msg.RoleName,
		}, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Revoker, "revoke_role", msg.Address, "success", map[string]string{
		"role": msg.RoleName,
	}, "")

	return &authproto.MsgRevokeRoleResponse{Success: true}, nil
}

// CreateMultisigWallet creates a new multisig wallet
func (ms msgServer) CreateMultisigWallet(goCtx context.Context, msg *authproto.MsgCreateMultisigWallet) (*authproto.MsgCreateMultisigWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Creator, types.PermissionManageMultisig); err != nil {
		return nil, err
	}

	// Validate inputs
	if len(msg.Signers) < 2 {
		return nil, types.ErrInvalidInput.Wrap("multisig requires at least 2 signers")
	}
	if msg.Threshold < 1 || msg.Threshold > uint32(len(msg.Signers)) {
		return nil, types.ErrInvalidInput.Wrap("invalid threshold")
	}

	// Create wallet
	wallet, err := ms.Keeper.CreateMultisigWallet(ctx, msg.Signers, msg.Threshold, msg.WalletType)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Creator, "create_multisig", "", "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Creator, "create_multisig", wallet.Id, "success", map[string]string{
		"signers":   fmt.Sprintf("%v", msg.Signers),
		"threshold": fmt.Sprintf("%d", msg.Threshold),
	}, "")

	return &authproto.MsgCreateMultisigWalletResponse{Wallet: wallet}, nil
}

// CreateMultisigProposal creates a new multisig proposal
func (ms msgServer) CreateMultisigProposal(goCtx context.Context, msg *authproto.MsgCreateMultisigProposal) (*authproto.MsgCreateMultisigProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.WalletId == "" || msg.Title == "" {
		return nil, types.ErrInvalidInput.Wrap("wallet_id and title are required")
	}

	// Create proposal
	proposal, err := ms.Keeper.CreateMultisigProposal(ctx, msg.Proposer, msg.WalletId, msg.Title, msg.Description, msg.Payload, msg.ExpiresInSeconds)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Proposer, "create_multisig_proposal", msg.WalletId, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Proposer, "create_multisig_proposal", proposal.Id, "success", map[string]string{
		"wallet": msg.WalletId,
		"title":  msg.Title,
	}, "")

	return &authproto.MsgCreateMultisigProposalResponse{Proposal: proposal}, nil
}

// SignMultisigProposal signs a multisig proposal
func (ms msgServer) SignMultisigProposal(goCtx context.Context, msg *authproto.MsgSignMultisigProposal) (*authproto.MsgSignMultisigProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Sign proposal
	proposal, err := ms.Keeper.SignMultisigProposal(ctx, msg.Signer, msg.ProposalId)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Signer, "sign_multisig_proposal", msg.ProposalId, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Signer, "sign_multisig_proposal", msg.ProposalId, "success", nil, "")

	return &authproto.MsgSignMultisigProposalResponse{Proposal: proposal}, nil
}

// ExecuteMultisigProposal executes an approved multisig proposal
func (ms msgServer) ExecuteMultisigProposal(goCtx context.Context, msg *authproto.MsgExecuteMultisigProposal) (*authproto.MsgExecuteMultisigProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Execute proposal
	err := ms.Keeper.ExecuteMultisigProposal(ctx, msg.Executor, msg.ProposalId)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Executor, "execute_multisig_proposal", msg.ProposalId, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Executor, "execute_multisig_proposal", msg.ProposalId, "success", nil, "")

	return &authproto.MsgExecuteMultisigProposalResponse{Success: true}, nil
}

// ProposeTimeLockedAction proposes a time-locked admin action
func (ms msgServer) ProposeTimeLockedAction(goCtx context.Context, msg *authproto.MsgProposeTimeLockedAction) (*authproto.MsgProposeTimeLockedActionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate proposer has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Proposer, types.PermissionManageTimeLock); err != nil {
		return nil, err
	}

	// Validate inputs
	if msg.ActionType == "" {
		return nil, types.ErrInvalidInput.Wrap("action_type is required")
	}

	// Create time-locked action
	action, err := ms.Keeper.ProposeTimeLockedAction(ctx, msg.Proposer, msg.ActionType, msg.Payload, msg.DelaySeconds)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Proposer, "propose_timelock", "", "failure", map[string]string{
			"action_type": msg.ActionType,
		}, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Proposer, "propose_timelock", action.Id, "success", map[string]string{
		"action_type": msg.ActionType,
		"delay":       fmt.Sprintf("%d", msg.DelaySeconds),
	}, "")

	return &authproto.MsgProposeTimeLockedActionResponse{Action: action}, nil
}

// ExecuteTimeLockedAction executes a ready time-locked action
func (ms msgServer) ExecuteTimeLockedAction(goCtx context.Context, msg *authproto.MsgExecuteTimeLockedAction) (*authproto.MsgExecuteTimeLockedActionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Execute action
	err := ms.Keeper.ExecuteTimeLockedAction(ctx, msg.Executor, msg.ActionId)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Executor, "execute_timelock", msg.ActionId, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Executor, "execute_timelock", msg.ActionId, "success", nil, "")

	return &authproto.MsgExecuteTimeLockedActionResponse{Success: true}, nil
}

// CancelTimeLockedAction cancels a pending time-locked action
func (ms msgServer) CancelTimeLockedAction(goCtx context.Context, msg *authproto.MsgCancelTimeLockedAction) (*authproto.MsgCancelTimeLockedActionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Cancel action
	err := ms.Keeper.CancelTimeLockedAction(ctx, msg.Canceller, msg.ActionId)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Canceller, "cancel_timelock", msg.ActionId, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Canceller, "cancel_timelock", msg.ActionId, "success", nil, "")

	return &authproto.MsgCancelTimeLockedActionResponse{Success: true}, nil
}

// ActivateEmergencyAdmin activates an emergency admin
func (ms msgServer) ActivateEmergencyAdmin(goCtx context.Context, msg *authproto.MsgActivateEmergencyAdmin) (*authproto.MsgActivateEmergencyAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate activator has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Activator, types.PermissionManageEmergency); err != nil {
		return nil, err
	}

	// Activate emergency admin
	admin, err := ms.Keeper.ActivateEmergencyAdmin(ctx, msg.AdminAddress, msg.Privileges, msg.ExpiresInSeconds)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Activator, "activate_emergency_admin", msg.AdminAddress, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Activator, "activate_emergency_admin", msg.AdminAddress, "success", map[string]string{
		"privileges": fmt.Sprintf("%v", msg.Privileges),
	}, "")

	return &authproto.MsgActivateEmergencyAdminResponse{Admin: admin}, nil
}

// DeactivateEmergencyAdmin deactivates an emergency admin
func (ms msgServer) DeactivateEmergencyAdmin(goCtx context.Context, msg *authproto.MsgDeactivateEmergencyAdmin) (*authproto.MsgDeactivateEmergencyAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate deactivator has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Deactivator, types.PermissionManageEmergency); err != nil {
		return nil, err
	}

	// Deactivate emergency admin
	err := ms.Keeper.DeactivateEmergencyAdmin(ctx, msg.AdminAddress)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Deactivator, "deactivate_emergency_admin", msg.AdminAddress, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Deactivator, "deactivate_emergency_admin", msg.AdminAddress, "success", nil, "")

	return &authproto.MsgDeactivateEmergencyAdminResponse{Success: true}, nil
}

// InitiateValidatorKeyRotation initiates validator key rotation
func (ms msgServer) InitiateValidatorKeyRotation(goCtx context.Context, msg *authproto.MsgInitiateValidatorKeyRotation) (*authproto.MsgInitiateValidatorKeyRotationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate initiator has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Initiator, types.PermissionRotateValidatorKey); err != nil {
		return nil, err
	}

	// Initiate rotation
	rotation, err := ms.Keeper.InitiateValidatorKeyRotation(ctx, msg.ValidatorAddress, msg.NewConsensusPubkey)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Initiator, "initiate_key_rotation", msg.ValidatorAddress, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Initiator, "initiate_key_rotation", msg.ValidatorAddress, "success", nil, "")

	return &authproto.MsgInitiateValidatorKeyRotationResponse{Rotation: rotation}, nil
}

// CompleteValidatorKeyRotation completes validator key rotation
func (ms msgServer) CompleteValidatorKeyRotation(goCtx context.Context, msg *authproto.MsgCompleteValidatorKeyRotation) (*authproto.MsgCompleteValidatorKeyRotationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Complete rotation
	err := ms.Keeper.CompleteValidatorKeyRotation(ctx, msg.ValidatorAddress)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.Completer, "complete_key_rotation", msg.ValidatorAddress, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.Completer, "complete_key_rotation", msg.ValidatorAddress, "success", nil, "")

	return &authproto.MsgCompleteValidatorKeyRotationResponse{Success: true}, nil
}

// CreateSession creates a new API session
func (ms msgServer) CreateSession(goCtx context.Context, msg *authproto.MsgCreateSession) (*authproto.MsgCreateSessionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create session
	session, err := ms.Keeper.CreateSession(ctx, msg.UserAddress, msg.IpAddress, msg.Metadata)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.UserAddress, "create_session", "", "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.UserAddress, "create_session", session.Id, "success", map[string]string{
		"ip": msg.IpAddress,
	}, "")

	return &authproto.MsgCreateSessionResponse{Session: session}, nil
}

// RevokeSession revokes an active session
func (ms msgServer) RevokeSession(goCtx context.Context, msg *authproto.MsgRevokeSession) (*authproto.MsgRevokeSessionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Revoke session
	err := ms.Keeper.RevokeSession(ctx, msg.UserAddress, msg.SessionId)
	if err != nil {
		ms.Keeper.LogAudit(ctx, msg.UserAddress, "revoke_session", msg.SessionId, "failure", nil, err.Error())
		return nil, err
	}

	ms.Keeper.LogAudit(ctx, msg.UserAddress, "revoke_session", msg.SessionId, "success", nil, "")

	return &authproto.MsgRevokeSessionResponse{Success: true}, nil
}
