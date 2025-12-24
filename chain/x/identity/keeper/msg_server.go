package keeper

import (
	"context"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	metrics "github.com/hashicorp/go-metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/pkg/log"
	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

var _ identitypb.MsgServer = (*msgServer)(nil)

type msgServer struct {
	identitypb.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl wires the keeper into the protobuf msg server implementation
func NewMsgServerImpl(k *Keeper) identitypb.MsgServer {
	return &msgServer{Keeper: k}
}

// RequestIdentityChange handles identity change requests
func (ms msgServer) RequestIdentityChange(goCtx context.Context, msg *identitypb.MsgRequestIdentityChange) (*identitypb.MsgRequestIdentityChangeResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	log.TxStart(ctx, "MsgRequestIdentityChange", msg.Requester)

	// Create change request
	request, err := ms.Keeper.CreateChangeRequest(ctx, msg.Requester, msg.TargetDid, msg.IrId, msg.MetadataHash)
	if err != nil {
		log.TxError(ctx, "MsgRequestIdentityChange", err, "requester", msg.Requester, "target_did", msg.TargetDid)
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.TxSuccess(ctx, "MsgRequestIdentityChange", "requester", msg.Requester, "request_id", request.Id, "target_did", msg.TargetDid)
	log.StateChange(ctx, "identity_change_request", "created", request.Id)
	return &identitypb.MsgRequestIdentityChangeResponse{
		RequestId: request.Id,
	}, nil
}

// SubmitAssistantProof handles assistant proof submission for identity verification
func (ms msgServer) SubmitAssistantProof(goCtx context.Context, msg *identitypb.MsgSubmitAssistantProof) (*identitypb.MsgSubmitAssistantProofResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Submit verification
	_, err := ms.Keeper.SubmitVerification(ctx, msg.RequestId, msg.Assistant, msg.Success, "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.MsgSubmitAssistantProofResponse{}, nil
}

// ApplyIdentityChange applies an approved identity change
func (ms msgServer) ApplyIdentityChange(goCtx context.Context, msg *identitypb.MsgApplyIdentityChange) (*identitypb.MsgApplyIdentityChangeResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Apply change
	_, err := ms.Keeper.ApplyChange(ctx, msg.RequestId, msg.Requester)
	if err != nil {
		if errorsmod.IsOf(err, types.ErrUnauthorized) || errorsmod.IsOf(err, types.ErrInsufficientPermissions) {
			telemetry.IncrCounterWithLabels(
				[]string{"identity", "auth_denied"},
				1,
				[]metrics.Label{
					{Name: "action", Value: "apply_change"},
				},
			)
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.MsgApplyIdentityChangeResponse{}, nil
}

// RejectIdentityChange rejects an identity change request
func (ms msgServer) RejectIdentityChange(goCtx context.Context, msg *identitypb.MsgRejectIdentityChange) (*identitypb.MsgRejectIdentityChangeResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Reject change
	_, err := ms.Keeper.RejectChange(ctx, msg.RequestId, msg.Actor, msg.Reason)
	if err != nil {
		if errorsmod.IsOf(err, types.ErrUnauthorized) || errorsmod.IsOf(err, types.ErrInsufficientPermissions) {
			telemetry.IncrCounterWithLabels(
				[]string{"identity", "auth_denied"},
				1,
				[]metrics.Label{
					{Name: "action", Value: "reject_change"},
				},
			)
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.MsgRejectIdentityChangeResponse{}, nil
}

// SuspendIdentityChanges suspends all identity changes (admin only)
func (ms msgServer) SuspendIdentityChanges(goCtx context.Context, msg *identitypb.MsgSuspendIdentityChanges) (*identitypb.MsgSuspendIdentityChangesResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.GetAuthority() {
		return nil, status.Error(codes.PermissionDenied, errorsmod.Wrapf(types.ErrUnauthorized, "invalid authority; expected %s, got %s", ms.Keeper.GetAuthority(), msg.Authority).Error())
	}

	// Set suspended flag
	if err := ms.Keeper.SetSuspended(ctx, true); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Authority, "suspend_identity_changes", "", "success", map[string]string{
		"reason": msg.Reason,
	}, "")

	return &identitypb.MsgSuspendIdentityChangesResponse{}, nil
}

// CreateRole creates a new role
func (ms msgServer) CreateRole(goCtx context.Context, msg *identitypb.MsgCreateRole) (*identitypb.MsgCreateRoleResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create role
	_, err := ms.Keeper.CreateRole(ctx, msg.Creator, msg.RoleName, msg.Permissions, msg.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.MsgCreateRoleResponse{}, nil
}

// AssignRole assigns a role to an address
func (ms msgServer) AssignRole(goCtx context.Context, msg *identitypb.MsgAssignRole) (*identitypb.MsgAssignRoleResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Calculate expiry in seconds
	var expirySeconds uint64
	if msg.ExpiresAt != nil {
		now := ctx.BlockTime()
		if msg.ExpiresAt.After(now) {
			expirySeconds = uint64(msg.ExpiresAt.Sub(now).Seconds())
		}
	}

	// Assign role
	_, err := ms.Keeper.AssignRole(ctx, msg.Assigner, msg.Address, msg.RoleName, expirySeconds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.MsgAssignRoleResponse{}, nil
}

// RevokeRole revokes a role from an address
func (ms msgServer) RevokeRole(goCtx context.Context, msg *identitypb.MsgRevokeRole) (*identitypb.MsgRevokeRoleResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Revoke role
	if err := ms.Keeper.RevokeRole(ctx, msg.Revoker, msg.Address, msg.RoleName); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.MsgRevokeRoleResponse{}, nil
}

// CreateMultisigWallet creates a new multisig wallet
func (ms msgServer) CreateMultisigWallet(goCtx context.Context, msg *identitypb.MsgCreateMultisigWallet) (*identitypb.MsgCreateMultisigWalletResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.Threshold == 0 {
		return nil, status.Error(codes.InvalidArgument, "threshold must be greater than 0")
	}
	if uint32(len(msg.Signers)) < msg.Threshold {
		return nil, status.Error(codes.InvalidArgument, "threshold cannot exceed number of signers")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check creator has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Creator, types.PermissionManageMultisig); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// Generate wallet ID
	walletID := fmt.Sprintf("mswallet-%s-%d", msg.Creator, ctx.BlockTime().Unix())

	now := ctx.BlockTime()
	wallet := &types.MultisigWallet{
		Id:         walletID,
		Signers:    msg.Signers,
		Threshold:  msg.Threshold,
		CreatedAt:  now,
		CreatedBy:  msg.Creator,
		WalletType: msg.WalletType,
	}

	if err := ms.Keeper.SetMultisigWallet(ctx, wallet); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Creator, "create_multisig_wallet", walletID, "success", map[string]string{
		"signers":   fmt.Sprintf("%d", len(msg.Signers)),
		"threshold": fmt.Sprintf("%d", msg.Threshold),
	}, "")

	return &identitypb.MsgCreateMultisigWalletResponse{
		WalletId: walletID,
	}, nil
}

// CreateMultisigProposal creates a new multisig proposal
func (ms msgServer) CreateMultisigProposal(goCtx context.Context, msg *identitypb.MsgCreateMultisigProposal) (*identitypb.MsgCreateMultisigProposalResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get wallet and verify proposer is a signer
	wallet, err := ms.Keeper.GetMultisigWallet(ctx, msg.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "wallet not found")
	}

	isSigner := false
	for _, signer := range wallet.Signers {
		if signer == msg.Proposer {
			isSigner = true
			break
		}
	}
	if !isSigner {
		return nil, status.Error(codes.PermissionDenied, "proposer is not a wallet signer")
	}

	// Generate proposal ID
	proposalID := fmt.Sprintf("msprop-%s-%d", msg.WalletId, ctx.BlockTime().Unix())

	now := ctx.BlockTime()
	// Get params for expiry
	params, err := ms.Keeper.GetParams(ctx)
	expirySeconds := uint64(604800) // default 7 days
	if err == nil {
		expirySeconds = params.Auth.MultisigProposalExpirySeconds
	}
	expiresAt := now.Add(time.Duration(expirySeconds) * time.Second)

	proposal := &types.MultisigProposal{
		Id:          proposalID,
		WalletId:    msg.WalletId,
		Title:       msg.Title,
		Description: msg.Description,
		Payload:     msg.Payload,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		Signatures:  []string{},
		Status:      types.ProposalStatusPending,
	}

	if err := ms.Keeper.SetMultisigProposal(ctx, proposal); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Proposer, "create_multisig_proposal", proposalID, "success", map[string]string{
		"wallet_id": msg.WalletId,
		"title":     msg.Title,
	}, "")

	return &identitypb.MsgCreateMultisigProposalResponse{
		ProposalId: proposalID,
	}, nil
}

// SignMultisigProposal adds a signature to a multisig proposal
func (ms msgServer) SignMultisigProposal(goCtx context.Context, msg *identitypb.MsgSignMultisigProposal) (*identitypb.MsgSignMultisigProposalResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get proposal - this reads the latest state from storage
	// ensuring we see any signatures added by previous transactions in this block
	proposal, err := ms.Keeper.GetMultisigProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	// Get wallet
	wallet, err := ms.Keeper.GetMultisigWallet(ctx, proposal.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "wallet not found")
	}

	// Verify signer is a wallet signer
	isSigner := false
	for _, signer := range wallet.Signers {
		if signer == msg.Signer {
			isSigner = true
			break
		}
	}
	if !isSigner {
		return nil, status.Error(codes.PermissionDenied, "signer is not a wallet signer")
	}

	// CRITICAL: Check if already signed - this must happen AFTER reading from storage
	// to catch signatures added by concurrent transactions in the same block.
	// If signer already exists in the signatures list, reject immediately to prevent
	// duplicate signatures and ensure idempotency.
	for _, sig := range proposal.Signatures {
		if sig == msg.Signer {
			return nil, status.Error(codes.AlreadyExists, "signer has already signed this proposal")
		}
	}

	// Add signature - atomic append
	proposal.Signatures = append(proposal.Signatures, msg.Signer)

	// Check if threshold met
	if uint32(len(proposal.Signatures)) >= wallet.Threshold {
		proposal.Status = types.ProposalStatusApproved
	}

	// Write back to storage - this completes the atomic read-modify-write pattern
	// The KVStore ensures that each transaction in a block sees the committed state
	// from previous transactions, preventing signature loss
	if err := ms.Keeper.SetMultisigProposal(ctx, &proposal); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Signer, "sign_multisig_proposal", msg.ProposalId, "success", map[string]string{
		"signatures": fmt.Sprintf("%d/%d", len(proposal.Signatures), wallet.Threshold),
	}, "")

	return &identitypb.MsgSignMultisigProposalResponse{}, nil
}

// ExecuteMultisigProposal executes an approved multisig proposal
func (ms msgServer) ExecuteMultisigProposal(goCtx context.Context, msg *identitypb.MsgExecuteMultisigProposal) (*identitypb.MsgExecuteMultisigProposalResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get proposal
	proposal, err := ms.Keeper.GetMultisigProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	// Verify proposal is approved
	if proposal.Status != types.ProposalStatusApproved {
		return nil, status.Error(codes.FailedPrecondition, "proposal not approved")
	}

	// Mark as executed
	now := ctx.BlockTime()
	proposal.Status = types.ProposalStatusExecuted
	proposal.ExecutedAt = &now

	if err := ms.Keeper.SetMultisigProposal(ctx, &proposal); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Executor, "execute_multisig_proposal", msg.ProposalId, "success", nil, "")

	return &identitypb.MsgExecuteMultisigProposalResponse{}, nil
}

// ProposeTimeLockedAction proposes a time-locked action
func (ms msgServer) ProposeTimeLockedAction(goCtx context.Context, msg *identitypb.MsgProposeTimeLockedAction) (*identitypb.MsgProposeTimeLockedActionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check proposer has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Proposer, types.PermissionManageTimeLock); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// Generate action ID
	actionID := fmt.Sprintf("tlaction-%s-%d", msg.Proposer, ctx.BlockTime().Unix())

	now := ctx.BlockTime()
	executableAt := now.Add(time.Duration(msg.DelaySeconds) * time.Second)

	action := &types.TimeLockedAction{
		Id:           actionID,
		ActionType:   msg.ActionType,
		Payload:      msg.Payload,
		ProposedAt:   now,
		Proposer:     msg.Proposer,
		ExecutableAt: executableAt,
		Status:       types.ActionStatusPending,
		DelaySeconds: msg.DelaySeconds,
	}

	if err := ms.Keeper.SetTimeLockedAction(ctx, action); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Proposer, "propose_time_locked_action", actionID, "success", map[string]string{
		"action_type":   msg.ActionType,
		"delay_seconds": fmt.Sprintf("%d", msg.DelaySeconds),
	}, "")

	return &identitypb.MsgProposeTimeLockedActionResponse{
		ActionId:     actionID,
		ExecutableAt: executableAt,
	}, nil
}

// ExecuteTimeLockedAction executes a time-locked action after delay
func (ms msgServer) ExecuteTimeLockedAction(goCtx context.Context, msg *identitypb.MsgExecuteTimeLockedAction) (*identitypb.MsgExecuteTimeLockedActionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get action
	action, err := ms.Keeper.GetTimeLockedAction(ctx, msg.ActionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "action not found")
	}

	// Verify delay has passed
	now := ctx.BlockTime()
	if now.Before(action.ExecutableAt) {
		return nil, status.Error(codes.FailedPrecondition, "action delay has not elapsed")
	}

	// Verify status
	if action.Status != types.ActionStatusPending {
		return nil, status.Error(codes.FailedPrecondition, "action not pending")
	}

	// Mark as executed
	action.Status = types.ActionStatusExecuted
	action.ExecutedAt = &now

	if err := ms.Keeper.SetTimeLockedAction(ctx, &action); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Executor, "execute_time_locked_action", msg.ActionId, "success", nil, "")

	return &identitypb.MsgExecuteTimeLockedActionResponse{}, nil
}

// CancelTimeLockedAction cancels a pending time-locked action
func (ms msgServer) CancelTimeLockedAction(goCtx context.Context, msg *identitypb.MsgCancelTimeLockedAction) (*identitypb.MsgCancelTimeLockedActionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get action
	action, err := ms.Keeper.GetTimeLockedAction(ctx, msg.ActionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "action not found")
	}

	// Check canceller has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Canceller, types.PermissionManageTimeLock); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// Verify status
	if action.Status != types.ActionStatusPending {
		return nil, status.Error(codes.FailedPrecondition, "only pending actions can be cancelled")
	}

	// Mark as cancelled
	action.Status = types.ActionStatusCancelled

	if err := ms.Keeper.SetTimeLockedAction(ctx, &action); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Canceller, "cancel_time_locked_action", msg.ActionId, "success", nil, "")

	return &identitypb.MsgCancelTimeLockedActionResponse{}, nil
}

// ActivateEmergencyAdmin activates an emergency admin with temporary privileges
func (ms msgServer) ActivateEmergencyAdmin(goCtx context.Context, msg *identitypb.MsgActivateEmergencyAdmin) (*identitypb.MsgActivateEmergencyAdminResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check activator has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Activator, types.PermissionManageEmergency); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	now := ctx.BlockTime()
	admin := &types.EmergencyAdmin{
		Address:     msg.AdminAddress,
		Privileges:  msg.Privileges,
		ActivatedAt: now,
		ActivatedBy: msg.Activator,
		ExpiresAt:   msg.ExpiresAt,
		IsActive:    true,
	}

	if err := ms.Keeper.SetEmergencyAdmin(ctx, admin); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Activator, "activate_emergency_admin", msg.AdminAddress, "success", map[string]string{
		"privileges": fmt.Sprintf("%v", msg.Privileges),
	}, "")

	return &identitypb.MsgActivateEmergencyAdminResponse{}, nil
}

// DeactivateEmergencyAdmin deactivates an emergency admin
func (ms msgServer) DeactivateEmergencyAdmin(goCtx context.Context, msg *identitypb.MsgDeactivateEmergencyAdmin) (*identitypb.MsgDeactivateEmergencyAdminResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check deactivator has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.Deactivator, types.PermissionManageEmergency); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// Get admin
	admin, err := ms.Keeper.GetEmergencyAdmin(ctx, msg.AdminAddress)
	if err != nil {
		return nil, status.Error(codes.NotFound, "emergency admin not found")
	}

	// Deactivate
	admin.IsActive = false

	if err := ms.Keeper.SetEmergencyAdmin(ctx, &admin); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Deactivator, "deactivate_emergency_admin", msg.AdminAddress, "success", nil, "")

	return &identitypb.MsgDeactivateEmergencyAdminResponse{}, nil
}

// RotateValidatorKey rotates a validator consensus public key
func (ms msgServer) RotateValidatorKey(goCtx context.Context, msg *identitypb.MsgRotateValidatorKey) (*identitypb.MsgRotateValidatorKeyResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validator has permission
	if err := ms.Keeper.RequirePermission(ctx, msg.ValidatorAddress, types.PermissionRotateValidatorKey); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	now := ctx.BlockTime()
	rotation := &types.ValidatorKeyRotation{
		ValidatorAddress:   msg.ValidatorAddress,
		NewConsensusPubkey: msg.NewConsensusPubkey,
		RotationTime:       now,
		InitiatedBy:        msg.ValidatorAddress,
		RotationStatus:     types.RotationStatusPending,
	}

	if err := ms.Keeper.SetValidatorRotation(ctx, rotation); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.ValidatorAddress, "rotate_validator_key", msg.ValidatorAddress, "success", nil, "")

	return &identitypb.MsgRotateValidatorKeyResponse{}, nil
}

// CreateSession creates a new user session
func (ms msgServer) CreateSession(goCtx context.Context, msg *identitypb.MsgCreateSession) (*identitypb.MsgCreateSessionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get params for default session duration
	params, err := ms.Keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get params")
	}

	var sessionDuration uint64 = 3600 // default 1 hour
	if err == nil && params.Auth.SessionTimeout > 0 {
		sessionDuration = uint64(params.Auth.SessionTimeout.Seconds())
	}

	// Create session
	session, err := ms.Keeper.CreateSession(ctx, msg.Address, sessionDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.MsgCreateSessionResponse{
		SessionId: session.Id,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// EndSession ends an active session
func (ms msgServer) EndSession(goCtx context.Context, msg *identitypb.MsgEndSession) (*identitypb.MsgEndSessionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Revoke session
	if err := ms.Keeper.RevokeSession(ctx, msg.Address, msg.SessionId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.MsgEndSessionResponse{}, nil
}

// UpdateParams updates module parameters (authority only)
func (ms msgServer) UpdateParams(goCtx context.Context, msg *identitypb.MsgUpdateParams) (*identitypb.MsgUpdateParamsResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.GetAuthority() {
		return nil, status.Error(codes.PermissionDenied, errorsmod.Wrapf(types.ErrUnauthorized, "invalid authority; expected %s, got %s", ms.Keeper.GetAuthority(), msg.Authority).Error())
	}

	// Set params - convert from proto to types
	params := &types.Params{
		Auth:   msg.Params.Auth,
		Change: msg.Params.Change,
	}
	if err := ms.Keeper.SetParams(ctx, params); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Log audit trail
	ms.Keeper.LogAudit(ctx, msg.Authority, "update_params", "", "success", nil, "")

	return &identitypb.MsgUpdateParamsResponse{}, nil
}

// EraseIdentity implements GDPR Right to Erasure
func (ms msgServer) EraseIdentity(goCtx context.Context, msg *identitypb.MsgEraseIdentity) (*identitypb.MsgEraseIdentityResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.Did == "" {
		return nil, status.Error(codes.InvalidArgument, "DID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Erase identity
	if err := ms.Keeper.EraseIdentity(ctx, msg.Did, msg.Requester, msg.Reason); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	now := ctx.BlockTime()
	return &identitypb.MsgEraseIdentityResponse{
		ErasedAt: now,
	}, nil
}

// RotateDIDKey initiates a DID key rotation with grace period
func (ms msgServer) RotateDIDKey(goCtx context.Context, msg *identitypb.MsgRotateDIDKey) (*identitypb.MsgRotateDIDKeyResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.Did == "" {
		return nil, status.Error(codes.InvalidArgument, "DID cannot be empty")
	}
	if msg.NewVerificationMethod == "" {
		return nil, status.Error(codes.InvalidArgument, "new verification method cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Rotate DID key
	rotation, err := ms.Keeper.RotateDIDKey(ctx, msg.Did, msg.Initiator, msg.NewVerificationMethod, msg.Reason)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.MsgRotateDIDKeyResponse{
		RotationTime:   rotation.RotationTime,
		GracePeriodEnd: rotation.GracePeriodEnd,
	}, nil
}
