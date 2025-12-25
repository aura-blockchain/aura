// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// CreateRole creates a new role
func (k *Keeper) CreateRole(ctx context.Context, creator string, name string, permissions []string, description string) (*authproto.Role, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if creator has permission
	if err := k.RequirePermission(sdkCtx, creator, types.PermissionCreateRole); err != nil {
		return nil, err
	}

	// Check if role already exists
	if _, err := k.GetRoleFromStore(sdkCtx, name); err == nil {
		return nil, fmt.Errorf("role already exists: %s", name)
	}

	now := sdkCtx.BlockTime()
	role := &authproto.Role{
		Name:        name,
		Permissions: permissions,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := k.SetRole(sdkCtx, role); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, creator, "create_role", name, "success", nil, "")
	return role, nil
}

// AssignRole assigns a role to an address
func (k *Keeper) AssignRole(ctx context.Context, assigner string, address string, roleName string, expiresInSeconds uint64) (*authproto.RoleAssignment, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate address is not empty
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	// Check if assigner has permission
	if err := k.RequirePermission(sdkCtx, assigner, types.PermissionAssignRole); err != nil {
		return nil, err
	}

	// Verify role exists
	if _, err := k.GetRoleFromStore(sdkCtx, roleName); err != nil {
		return nil, fmt.Errorf("role not found: %s", roleName)
	}

	now := sdkCtx.BlockTime()
	assignment := &authproto.RoleAssignment{
		Address:    address,
		RoleName:   roleName,
		AssignedBy: assigner,
		AssignedAt: now,
	}

	if expiresInSeconds > 0 {
		expiresAt := now.Add(time.Duration(expiresInSeconds) * time.Second)
		assignment.ExpiresAt = &expiresAt
	}

	if err := k.SetRoleAssignment(sdkCtx, assignment); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, assigner, "assign_role", fmt.Sprintf("%s->%s", address, roleName), "success", nil, "")
	return assignment, nil
}

// RevokeRole revokes a role from an address
func (k *Keeper) RevokeRole(ctx context.Context, revoker string, address string, roleName string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if revoker has permission
	if err := k.RequirePermission(sdkCtx, revoker, types.PermissionRevokeRole); err != nil {
		return fmt.Errorf("error in RevokeRole: %w", err)
	}

	if err := k.DeleteRoleAssignment(sdkCtx, address, roleName); err != nil {
		return fmt.Errorf("error in RevokeRole: %w", err)
	}

	k.LogAudit(sdkCtx, revoker, "revoke_role", fmt.Sprintf("%s->%s", address, roleName), "success", nil, "")
	return nil
}

// CreateMultisigWallet creates a new multisig wallet
func (k *Keeper) CreateMultisigWallet(ctx context.Context, creator string, signers []string, threshold uint32, walletType string) (*authproto.MultisigWallet, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if creator has permission
	if err := k.RequirePermission(sdkCtx, creator, types.PermissionManageMultisig); err != nil {
		return nil, err
	}

	// Validate threshold
	if threshold == 0 || int(threshold) > len(signers) {
		return nil, fmt.Errorf("invalid threshold: must be between 1 and %d", len(signers))
	}

	// Generate wallet ID
	walletID := fmt.Sprintf("msig_%d", sdkCtx.BlockTime().UnixNano())

	now := sdkCtx.BlockTime()
	wallet := &authproto.MultisigWallet{
		Id:         walletID,
		Signers:    signers,
		Threshold:  threshold,
		WalletType: authproto.WalletType_WALLET_TYPE_CUSTOM,
		CreatedAt:  now,
		CreatedBy:  creator,
	}

	if err := k.SetMultisigWallet(sdkCtx, wallet); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, creator, "create_multisig_wallet", walletID, "success", nil, "")
	return wallet, nil
}

// CreateMultisigProposal creates a new multisig proposal
func (k *Keeper) CreateMultisigProposal(ctx context.Context, proposer string, walletID string, title string, description string, payload []byte, expiresInSeconds uint64) (*authproto.MultisigProposal, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get wallet
	wallet, err := k.GetMultisigWallet(sdkCtx, walletID)
	if err != nil {
		return nil, err
	}

	// Verify proposer is a signer
	isSigner := false
	for _, signer := range wallet.Signers {
		if signer == proposer {
			isSigner = true
			break
		}
	}
	if !isSigner {
		return nil, fmt.Errorf("proposer is not a signer of wallet %s", walletID)
	}

	// Generate proposal ID
	proposalID := fmt.Sprintf("msp_%d", sdkCtx.BlockTime().UnixNano())

	now := sdkCtx.BlockTime()
	proposal := &authproto.MultisigProposal{
		Id:          proposalID,
		WalletId:    walletID,
		Title:       title,
		Description: description,
		Payload:     payload,
		CreatedAt:   now,
		Status:      authproto.ProposalStatus_PROPOSAL_STATUS_PENDING,
		Signatures:  []string{proposer}, // Proposer auto-signs
	}

	if expiresInSeconds > 0 {
		proposal.ExpiresAt = now.Add(time.Duration(expiresInSeconds) * time.Second)
	}

	if err := k.SetMultisigProposal(sdkCtx, proposal); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, proposer, "create_multisig_proposal", proposalID, "success", nil, "")
	return proposal, nil
}

// SignMultisigProposal signs a multisig proposal
func (k *Keeper) SignMultisigProposal(ctx context.Context, signer string, proposalID string) (*authproto.MultisigProposal, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get proposal
	proposal, err := k.GetMultisigProposal(sdkCtx, proposalID)
	if err != nil {
		return nil, err
	}

	// Get wallet
	wallet, err := k.GetMultisigWallet(sdkCtx, proposal.WalletId)
	if err != nil {
		return nil, err
	}

	// Verify signer is a wallet signer
	isSigner := false
	for _, s := range wallet.Signers {
		if s == signer {
			isSigner = true
			break
		}
	}
	if !isSigner {
		return nil, fmt.Errorf("signer is not authorized for this wallet")
	}

	// Check if already signed
	for _, sig := range proposal.Signatures {
		if sig == signer {
			return nil, fmt.Errorf("signer has already signed this proposal")
		}
	}

	// Add signature
	proposal.Signatures = append(proposal.Signatures, signer)

	// Update status if threshold reached
	if uint32(len(proposal.Signatures)) >= wallet.Threshold {
		proposal.Status = authproto.ProposalStatus_PROPOSAL_STATUS_APPROVED
	}

	if err := k.SetMultisigProposal(sdkCtx, proposal); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, signer, "sign_multisig_proposal", proposalID, "success", nil, "")
	return proposal, nil
}

// ExecuteMultisigProposal executes an approved multisig proposal
func (k *Keeper) ExecuteMultisigProposal(ctx context.Context, executor string, proposalID string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get proposal
	proposal, err := k.GetMultisigProposal(sdkCtx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Verify proposal is approved
	if proposal.Status != authproto.ProposalStatus_PROPOSAL_STATUS_APPROVED {
		return fmt.Errorf("proposal is not approved")
	}

	// Mark as executed
	proposal.Status = authproto.ProposalStatus_PROPOSAL_STATUS_EXECUTED
	executedAt := sdkCtx.BlockTime()
	proposal.ExecutedAt = &executedAt

	if err := k.SetMultisigProposal(sdkCtx, proposal); err != nil {
		return fmt.Errorf("error in ExecuteMultisigProposal: %w", err)
	}

	k.LogAudit(sdkCtx, executor, "execute_multisig_proposal", proposalID, "success", nil, "")
	return nil
}

// ProposeTimeLockedAction proposes a time-locked admin action
func (k *Keeper) ProposeTimeLockedAction(ctx context.Context, proposer string, actionType string, payload []byte, delaySeconds uint64) (*authproto.TimeLockedAction, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if proposer has permission
	if err := k.RequirePermission(sdkCtx, proposer, types.PermissionManageTimeLock); err != nil {
		return nil, err
	}

	// Generate action ID
	actionID := fmt.Sprintf("tla_%d", sdkCtx.BlockTime().UnixNano())

	now := sdkCtx.BlockTime()
	readyAt := now.Add(time.Duration(delaySeconds) * time.Second)

	action := &authproto.TimeLockedAction{
		Id:            actionID,
		ActionType:    actionType,
		Payload:       payload,
		Proposer:      proposer,
		ProposedAt:    now,
		ExecutableAt:  readyAt,
		Status:        authproto.ActionStatus_ACTION_STATUS_PENDING,
		DelaySeconds:  delaySeconds,
	}

	if err := k.SetTimeLockedAction(sdkCtx, action); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, proposer, "propose_timelock_action", actionID, "success", nil, "")
	return action, nil
}

// ExecuteTimeLockedAction executes a ready time-locked action
func (k *Keeper) ExecuteTimeLockedAction(ctx context.Context, executor string, actionID string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if executor has permission
	if err := k.RequirePermission(sdkCtx, executor, types.PermissionManageTimeLock); err != nil {
		return fmt.Errorf("error in ExecuteTimeLockedAction: %w", err)
	}

	// Get action
	action, err := k.GetTimeLockedAction(sdkCtx, actionID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Verify action is ready
	now := sdkCtx.BlockTime()
	if now.Before(action.ExecutableAt) {
		return types.ErrActionNotReady
	}

	// Mark as executed
	action.Status = authproto.ActionStatus_ACTION_STATUS_EXECUTED
	action.ExecutedAt = &now

	if err := k.SetTimeLockedAction(sdkCtx, action); err != nil {
		return fmt.Errorf("error in ExecuteTimeLockedAction: %w", err)
	}

	k.LogAudit(sdkCtx, executor, "execute_timelock_action", actionID, "success", nil, "")
	return nil
}

// CancelTimeLockedAction cancels a pending time-locked action
func (k *Keeper) CancelTimeLockedAction(ctx context.Context, canceller string, actionID string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if canceller has permission
	if err := k.RequirePermission(sdkCtx, canceller, types.PermissionManageTimeLock); err != nil {
		return fmt.Errorf("error in CancelTimeLockedAction: %w", err)
	}

	// Get action
	action, err := k.GetTimeLockedAction(sdkCtx, actionID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Mark as cancelled
	action.Status = authproto.ActionStatus_ACTION_STATUS_CANCELLED

	if err := k.SetTimeLockedAction(sdkCtx, action); err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	k.LogAudit(sdkCtx, canceller, "cancel_timelock_action", actionID, "success", nil, "")
	return nil
}

// ActivateEmergencyAdmin activates an emergency admin
func (k *Keeper) ActivateEmergencyAdmin(ctx context.Context, activator string, adminAddress string, privileges []string, expiresInSeconds uint64) (*authproto.EmergencyAdmin, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if activator has permission
	if err := k.RequirePermission(sdkCtx, activator, types.PermissionManageEmergency); err != nil {
		return nil, err
	}

	now := sdkCtx.BlockTime()
	expiresAt := now.Add(time.Duration(expiresInSeconds) * time.Second)

	admin := &authproto.EmergencyAdmin{
		Address:    adminAddress,
		Privileges: privileges,
		ActivatedAt: now,
		ExpiresAt:  &expiresAt,
		ActivatedBy: activator,
		IsActive:   true,
	}

	if err := k.SetEmergencyAdmin(sdkCtx, admin); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, activator, "activate_emergency_admin", adminAddress, "success", nil, "")
	return admin, nil
}

// DeactivateEmergencyAdmin deactivates an emergency admin
func (k *Keeper) DeactivateEmergencyAdmin(ctx context.Context, deactivator string, adminAddress string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if deactivator has permission
	if err := k.RequirePermission(sdkCtx, deactivator, types.PermissionManageEmergency); err != nil {
		return fmt.Errorf("error in DeactivateEmergencyAdmin: %w", err)
	}

	// Get the emergency admin
	admin, err := k.GetEmergencyAdmin(sdkCtx, adminAddress)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Set IsActive to false instead of deleting
	admin.IsActive = false
	if err := k.SetEmergencyAdmin(sdkCtx, admin); err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	k.LogAudit(sdkCtx, deactivator, "deactivate_emergency_admin", adminAddress, "success", nil, "")
	return nil
}

// RotateValidatorKey rotates a validator's consensus key
func (k *Keeper) RotateValidatorKey(ctx context.Context, validator string, newPubKey []byte) (*authproto.ValidatorKeyRotation, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if validator has permission
	if err := k.RequirePermission(sdkCtx, validator, types.PermissionRotateValidatorKey); err != nil {
		return nil, err
	}

	now := sdkCtx.BlockTime()
	rotation := &authproto.ValidatorKeyRotation{
		ValidatorAddress:    validator,
		OldConsensusPubkey:  "", // Would be retrieved from staking module
		NewConsensusPubkey:  string(newPubKey),
		RotationTime:        now,
		InitiatedBy:         validator,
		RotationStatus:      authproto.RotationStatus_ROTATION_STATUS_PENDING,
	}

	if err := k.SetValidatorKeyRotation(sdkCtx, rotation); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, validator, "rotate_validator_key", validator, "success", nil, "")
	return rotation, nil
}

// CreateSession creates a new session
func (k *Keeper) CreateSession(ctx context.Context, userAddress string, expiresInSeconds uint64) (*authproto.Session, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Generate session ID
	sessionID := fmt.Sprintf("sess_%d", sdkCtx.BlockTime().UnixNano())

	now := sdkCtx.BlockTime()
	expiresAt := now.Add(time.Duration(expiresInSeconds) * time.Second)

	session := &authproto.Session{
		SessionId:   sessionID,
		UserAddress: userAddress,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		IpAddress:   k.getIPFromContext(ctx),
	}

	if err := k.SetSession(sdkCtx, session); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, userAddress, "create_session", sessionID, "success", nil, "")
	return session, nil
}

// InvalidateSession invalidates a session
func (k *Keeper) InvalidateSession(ctx context.Context, userAddress string, sessionID string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.DeleteSession(sdkCtx, sessionID); err != nil {
		return fmt.Errorf("error in InvalidateSession for InvalidateSession: %w", err)
	}

	k.LogAudit(sdkCtx, userAddress, "invalidate_session", sessionID, "success", nil, "")
	return nil
}

// InitiateValidatorKeyRotation initiates a validator key rotation
func (k *Keeper) InitiateValidatorKeyRotation(ctx context.Context, initiator string, validatorAddress string, newConsensusPubkey []byte) (*authproto.ValidatorKeyRotation, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if initiator has permission
	if err := k.RequirePermission(sdkCtx, initiator, types.PermissionRotateValidatorKey); err != nil {
		return nil, err
	}

	now := sdkCtx.BlockTime()
	rotation := &authproto.ValidatorKeyRotation{
		ValidatorAddress:    validatorAddress,
		OldConsensusPubkey:  "", // Would be retrieved from staking module
		NewConsensusPubkey:  string(newConsensusPubkey),
		RotationTime:        now,
		InitiatedBy:         initiator,
		RotationStatus:      authproto.RotationStatus_ROTATION_STATUS_PENDING,
	}

	if err := k.SetValidatorKeyRotation(sdkCtx, rotation); err != nil {
		return nil, err
	}

	k.LogAudit(sdkCtx, initiator, "initiate_validator_key_rotation", validatorAddress, "success", nil, "")
	return rotation, nil
}

// CompleteValidatorKeyRotation completes a validator key rotation
func (k *Keeper) CompleteValidatorKeyRotation(ctx context.Context, completer string, validatorAddress string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if completer has permission
	if err := k.RequirePermission(sdkCtx, completer, types.PermissionRotateValidatorKey); err != nil {
		return fmt.Errorf("error in CompleteValidatorKeyRotation for initiate_validator_key_rotation: %w", err)
	}

	// Get rotation record
	rotation, err := k.GetValidatorKeyRotation(sdkCtx, validatorAddress)
	if err != nil {
		return fmt.Errorf("rotation record not found: %w", err)
	}

	// Mark as completed
	rotation.RotationStatus = authproto.RotationStatus_ROTATION_STATUS_COMPLETED
	rotation.RotationTime = sdkCtx.BlockTime()

	if err := k.SetValidatorKeyRotation(sdkCtx, rotation); err != nil {
		return fmt.Errorf("failed to get for GetValidatorKeyRotation: %w", err)
	}

	k.LogAudit(sdkCtx, completer, "complete_validator_key_rotation", validatorAddress, "success", nil, "")
	return nil
}

// RevokeSession revokes a session (alias for InvalidateSession)
func (k *Keeper) RevokeSession(ctx context.Context, userAddress string, sessionID string) error {
	return k.InvalidateSession(ctx, userAddress, sessionID)
}

// NOTE: The following helper functions have been REMOVED as they were stubs without context.
// All query functionality is properly implemented in query_server.go with proper SDK context handling.
// These functions were non-functional placeholders that returned empty data or errors.
//
// Removed functions (now properly implemented in query_server.go):
// - GetRole(name string) - use queryServer.GetRole() instead
// - ListRoles() - use queryServer.ListRoles() instead
// - ListMultisigWallets() - use queryServer.ListMultisigWallets() instead
// - ListMultisigProposals() - use queryServer.ListMultisigProposals() instead
// - ListTimeLockedActions() - use queryServer.ListTimeLockedActions() instead
// - ListEmergencyAdmins() - use queryServer.ListEmergencyAdmins() instead
// - ListValidatorKeyRotations() - use queryServer.ListValidatorKeyRotations() instead
// - ListSessions() - use queryServer.ListSessions() instead
// - GetRateLimitStatus() - use queryServer.GetRateLimitStatus() instead
//
// All of these are properly implemented with full KVStore access and SDK context in:
// /home/decri/blockchain-projects/aura/chain/x/auth/keeper/query_server.go
// /home/decri/blockchain-projects/aura/chain/x/auth/keeper/keeper.go (KV storage methods)

