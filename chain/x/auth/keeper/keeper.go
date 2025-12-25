// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// Key prefixes for KVStore
var (
	RolesKeyPrefix              = []byte{0x01}
	RoleAssignmentsKeyPrefix    = []byte{0x02}
	MultisigWalletsKeyPrefix    = []byte{0x03}
	MultisigProposalsKeyPrefix  = []byte{0x04}
	TimeLockedActionsKeyPrefix  = []byte{0x05}
	EmergencyAdminsKeyPrefix    = []byte{0x06}
	ValidatorRotationsKeyPrefix = []byte{0x07}
	SessionsKeyPrefix           = []byte{0x08}
	UserSessionsKeyPrefix       = []byte{0x09}
	RateLimitsKeyPrefix         = []byte{0x0A}
	AuditLogsKeyPrefix          = []byte{0x0B}
	ParamsKeyPrefix             = []byte{0x0C}
	AuditLogCounterKeyPrefix    = []byte{0x0D}
)

// Keeper manages authentication and authorization state
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

// NewKeeper creates a new auth keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// InitializeDefaultRoles creates predefined roles
func (k *Keeper) InitializeDefaultRoles(ctx sdk.Context) error {
	now := ctx.BlockTime()

	// Admin role with all permissions
	adminRole := &authproto.Role{
		Name: types.RoleAdmin,
		Permissions: []string{
			types.PermissionAdmin,
			types.PermissionCreateRole,
			types.PermissionAssignRole,
			types.PermissionRevokeRole,
			types.PermissionManageMultisig,
			types.PermissionManageTimeLock,
			types.PermissionManageEmergency,
			types.PermissionRotateValidatorKey,
			types.PermissionManageSession,
			types.PermissionViewAuditLogs,
		},
		Description: "Full administrative access",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := k.SetRole(ctx, adminRole); err != nil {
		return fmt.Errorf("error in InitializeDefaultRoles for PermissionRotateValidatorKey: %w", err)
	}

	// Moderator role
	moderatorRole := &authproto.Role{
		Name: types.RoleModerator,
		Permissions: []string{
			types.PermissionAssignRole,
			types.PermissionRevokeRole,
			types.PermissionManageSession,
			types.PermissionViewAuditLogs,
		},
		Description: "Moderate user permissions",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := k.SetRole(ctx, moderatorRole); err != nil {
		return fmt.Errorf("error in InitializeDefaultRoles: %w", err)
	}

	// Validator role
	validatorRole := &authproto.Role{
		Name: types.RoleValidator,
		Permissions: []string{
			types.PermissionRotateValidatorKey,
		},
		Description: "Validator-specific permissions",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := k.SetRole(ctx, validatorRole); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// User role
	userRole := &authproto.Role{
		Name:        types.RoleUser,
		Permissions: []string{},
		Description: "Basic user permissions",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return k.SetRole(ctx, userRole)
}

// ============================================================================
// Params KVStore Methods
// ============================================================================

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) (authproto.Params, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ParamsKeyPrefix)
	if bz == nil {
		// Return default params if not set
		return authproto.Params{}, nil
	}

	var params authproto.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return authproto.Params{}, err
	}
	return params, nil
}

// SetParams updates the module parameters
func (k *Keeper) SetParams(ctx sdk.Context, params *authproto.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	store.Set(ParamsKeyPrefix, bz)
	return nil
}

// ============================================================================
// Role KVStore Methods
// ============================================================================

// SetRole stores a role in the KVStore
func (k *Keeper) SetRole(ctx sdk.Context, role *authproto.Role) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(role)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(RolesKeyPrefix, []byte(role.Name)...)
	store.Set(key, bz)
	return nil
}

// GetRole retrieves a role from the KVStore
func (k *Keeper) GetRoleFromStore(ctx sdk.Context, name string) (*authproto.Role, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(RolesKeyPrefix, []byte(name)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("role not found: %s", name)
	}

	var role authproto.Role
	if err := k.cdc.Unmarshal(bz, &role); err != nil {
		return nil, err
	}
	return &role, nil
}

// GetAllRoles retrieves all roles from the KVStore
func (k *Keeper) GetAllRoles(ctx sdk.Context) ([]authproto.Role, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, RolesKeyPrefix)
	defer iterator.Close()

	var roles []authproto.Role
	for ; iterator.Valid(); iterator.Next() {
		var role authproto.Role
		if err := k.cdc.Unmarshal(iterator.Value(), &role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// DeleteRole removes a role from the KVStore
func (k *Keeper) DeleteRole(ctx sdk.Context, name string) {
	store := ctx.KVStore(k.storeKey)
	key := append(RolesKeyPrefix, []byte(name)...)
	store.Delete(key)
}

// ============================================================================
// RoleAssignment KVStore Methods
// ============================================================================

// SetRoleAssignment stores a role assignment in the KVStore
func (k *Keeper) SetRoleAssignment(ctx sdk.Context, assignment *authproto.RoleAssignment) error {
	store := ctx.KVStore(k.storeKey)

	// Get existing assignments for this address
	assignments, _ := k.GetRoleAssignmentsForAddress(ctx, assignment.Address)

	// Check if assignment already exists and update it
	found := false
	for i, existing := range assignments {
		if existing.RoleName == assignment.RoleName {
			assignments[i] = assignment
			found = true
			break
		}
	}

	if !found {
		assignments = append(assignments, assignment)
	}

	// Store all assignments for this address
	bz, err := k.cdc.Marshal(&authproto.RoleAssignmentList{Assignments: assignments})
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	key := append(RoleAssignmentsKeyPrefix, []byte(assignment.Address)...)
	store.Set(key, bz)
	return nil
}

// GetRoleAssignmentsForAddress retrieves all role assignments for an address
func (k *Keeper) GetRoleAssignmentsForAddress(ctx sdk.Context, address string) ([]*authproto.RoleAssignment, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(RoleAssignmentsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return []*authproto.RoleAssignment{}, nil
	}

	var assignmentList authproto.RoleAssignmentList
	if err := k.cdc.Unmarshal(bz, &assignmentList); err != nil {
		return nil, err
	}

	// Filter out expired assignments
	now := ctx.BlockTime()
	filtered := make([]*authproto.RoleAssignment, 0)
	for _, assignment := range assignmentList.Assignments {
		if assignment.ExpiresAt != nil && now.After(*assignment.ExpiresAt) {
			continue
		}
		filtered = append(filtered, assignment)
	}

	return filtered, nil
}

// DeleteRoleAssignment removes a specific role assignment
func (k *Keeper) DeleteRoleAssignment(ctx sdk.Context, address, roleName string) error {
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	filtered := make([]*authproto.RoleAssignment, 0)
	for _, assignment := range assignments {
		if assignment.RoleName != roleName {
			filtered = append(filtered, assignment)
		}
	}

	store := ctx.KVStore(k.storeKey)
	key := append(RoleAssignmentsKeyPrefix, []byte(address)...)

	if len(filtered) == 0 {
		store.Delete(key)
		return nil
	}

	bz, err := k.cdc.Marshal(&authproto.RoleAssignmentList{Assignments: filtered})
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// GetAllRoleAssignments retrieves all role assignments
func (k *Keeper) GetAllRoleAssignments(ctx sdk.Context) ([]authproto.RoleAssignment, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, RoleAssignmentsKeyPrefix)
	defer iterator.Close()

	var allAssignments []authproto.RoleAssignment
	for ; iterator.Valid(); iterator.Next() {
		var assignmentList authproto.RoleAssignmentList
		if err := k.cdc.Unmarshal(iterator.Value(), &assignmentList); err != nil {
			return nil, err
		}
		// Convert pointer slice to value slice
		for _, assignment := range assignmentList.Assignments {
			if assignment != nil {
				allAssignments = append(allAssignments, *assignment)
			}
		}
	}
	return allAssignments, nil
}

// ============================================================================
// Multisig Wallet KVStore Methods
// ============================================================================

// SetMultisigWallet stores a multisig wallet
func (k *Keeper) SetMultisigWallet(ctx sdk.Context, wallet *authproto.MultisigWallet) error {
	// Validate threshold is not zero
	if wallet.Threshold == 0 {
		return fmt.Errorf("threshold must be greater than 0")
	}

	// Validate threshold does not exceed number of signers
	if int(wallet.Threshold) > len(wallet.Signers) {
		return fmt.Errorf("threshold cannot be greater than number of signers")
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(wallet)
	if err != nil {
		return fmt.Errorf("failed to marshal for Validate: %w", err)
	}
	key := append(MultisigWalletsKeyPrefix, []byte(wallet.Id)...)
	store.Set(key, bz)
	return nil
}

// GetMultisigWallet retrieves a multisig wallet by ID
func (k *Keeper) GetMultisigWallet(ctx sdk.Context, walletID string) (*authproto.MultisigWallet, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(MultisigWalletsKeyPrefix, []byte(walletID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("multisig wallet not found: %s", walletID)
	}

	var wallet authproto.MultisigWallet
	if err := k.cdc.Unmarshal(bz, &wallet); err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetAllMultisigWallets retrieves all multisig wallets
func (k *Keeper) GetAllMultisigWallets(ctx sdk.Context) ([]authproto.MultisigWallet, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, MultisigWalletsKeyPrefix)
	defer iterator.Close()

	var wallets []authproto.MultisigWallet
	for ; iterator.Valid(); iterator.Next() {
		var wallet authproto.MultisigWallet
		if err := k.cdc.Unmarshal(iterator.Value(), &wallet); err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}
	return wallets, nil
}

// DeleteMultisigWallet removes a multisig wallet
func (k *Keeper) DeleteMultisigWallet(ctx sdk.Context, walletID string) {
	store := ctx.KVStore(k.storeKey)
	key := append(MultisigWalletsKeyPrefix, []byte(walletID)...)
	store.Delete(key)
}

// ============================================================================
// Multisig Proposal KVStore Methods
// ============================================================================

// SetMultisigProposal stores a multisig proposal
func (k *Keeper) SetMultisigProposal(ctx sdk.Context, proposal *authproto.MultisigProposal) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(proposal)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(MultisigProposalsKeyPrefix, []byte(proposal.Id)...)
	store.Set(key, bz)
	return nil
}

// GetMultisigProposal retrieves a multisig proposal by ID
func (k *Keeper) GetMultisigProposal(ctx sdk.Context, proposalID string) (*authproto.MultisigProposal, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(MultisigProposalsKeyPrefix, []byte(proposalID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("multisig proposal not found: %s", proposalID)
	}

	var proposal authproto.MultisigProposal
	if err := k.cdc.Unmarshal(bz, &proposal); err != nil {
		return nil, err
	}
	return &proposal, nil
}

// GetAllMultisigProposals retrieves all multisig proposals
func (k *Keeper) GetAllMultisigProposals(ctx sdk.Context) ([]authproto.MultisigProposal, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, MultisigProposalsKeyPrefix)
	defer iterator.Close()

	var proposals []authproto.MultisigProposal
	for ; iterator.Valid(); iterator.Next() {
		var proposal authproto.MultisigProposal
		if err := k.cdc.Unmarshal(iterator.Value(), &proposal); err != nil {
			return nil, err
		}
		proposals = append(proposals, proposal)
	}
	return proposals, nil
}

// DeleteMultisigProposal removes a multisig proposal
func (k *Keeper) DeleteMultisigProposal(ctx sdk.Context, proposalID string) {
	store := ctx.KVStore(k.storeKey)
	key := append(MultisigProposalsKeyPrefix, []byte(proposalID)...)
	store.Delete(key)
}

// ============================================================================
// Time-Locked Action KVStore Methods
// ============================================================================

// SetTimeLockedAction stores a time-locked action
func (k *Keeper) SetTimeLockedAction(ctx sdk.Context, action *authproto.TimeLockedAction) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(action)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(TimeLockedActionsKeyPrefix, []byte(action.Id)...)
	store.Set(key, bz)
	return nil
}

// GetTimeLockedAction retrieves a time-locked action by ID
func (k *Keeper) GetTimeLockedAction(ctx sdk.Context, actionID string) (*authproto.TimeLockedAction, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(TimeLockedActionsKeyPrefix, []byte(actionID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("time-locked action not found: %s", actionID)
	}

	var action authproto.TimeLockedAction
	if err := k.cdc.Unmarshal(bz, &action); err != nil {
		return nil, err
	}
	return &action, nil
}

// GetAllTimeLockedActions retrieves all time-locked actions
func (k *Keeper) GetAllTimeLockedActions(ctx sdk.Context) ([]authproto.TimeLockedAction, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, TimeLockedActionsKeyPrefix)
	defer iterator.Close()

	var actions []authproto.TimeLockedAction
	for ; iterator.Valid(); iterator.Next() {
		var action authproto.TimeLockedAction
		if err := k.cdc.Unmarshal(iterator.Value(), &action); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, nil
}

// DeleteTimeLockedAction removes a time-locked action
func (k *Keeper) DeleteTimeLockedAction(ctx sdk.Context, actionID string) {
	store := ctx.KVStore(k.storeKey)
	key := append(TimeLockedActionsKeyPrefix, []byte(actionID)...)
	store.Delete(key)
}

// ============================================================================
// Emergency Admin KVStore Methods
// ============================================================================

// SetEmergencyAdmin stores an emergency admin
func (k *Keeper) SetEmergencyAdmin(ctx sdk.Context, admin *authproto.EmergencyAdmin) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(admin)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(EmergencyAdminsKeyPrefix, []byte(admin.Address)...)
	store.Set(key, bz)
	return nil
}

// GetEmergencyAdmin retrieves an emergency admin
func (k *Keeper) GetEmergencyAdmin(ctx sdk.Context, address string) (*authproto.EmergencyAdmin, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(EmergencyAdminsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("emergency admin not found: %s", address)
	}

	var admin authproto.EmergencyAdmin
	if err := k.cdc.Unmarshal(bz, &admin); err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetAllEmergencyAdmins retrieves all emergency admins
func (k *Keeper) GetAllEmergencyAdmins(ctx sdk.Context) ([]authproto.EmergencyAdmin, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, EmergencyAdminsKeyPrefix)
	defer iterator.Close()

	var admins []authproto.EmergencyAdmin
	for ; iterator.Valid(); iterator.Next() {
		var admin authproto.EmergencyAdmin
		if err := k.cdc.Unmarshal(iterator.Value(), &admin); err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}
	return admins, nil
}

// DeleteEmergencyAdmin removes an emergency admin
func (k *Keeper) DeleteEmergencyAdmin(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	key := append(EmergencyAdminsKeyPrefix, []byte(address)...)
	store.Delete(key)
}

// ============================================================================
// Validator Key Rotation KVStore Methods
// ============================================================================

// SetValidatorKeyRotation stores a validator key rotation
func (k *Keeper) SetValidatorKeyRotation(ctx sdk.Context, rotation *authproto.ValidatorKeyRotation) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(rotation)
	if err != nil {
		return fmt.Errorf("failed to marshal for Validator: %w", err)
	}
	key := append(ValidatorRotationsKeyPrefix, []byte(rotation.ValidatorAddress)...)
	store.Set(key, bz)
	return nil
}

// GetValidatorKeyRotation retrieves a validator key rotation by address
func (k *Keeper) GetValidatorKeyRotation(ctx sdk.Context, validatorAddress string) (*authproto.ValidatorKeyRotation, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(ValidatorRotationsKeyPrefix, []byte(validatorAddress)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("validator key rotation not found: %s", validatorAddress)
	}

	var rotation authproto.ValidatorKeyRotation
	if err := k.cdc.Unmarshal(bz, &rotation); err != nil {
		return nil, err
	}
	return &rotation, nil
}

// GetAllValidatorKeyRotations retrieves all validator key rotations
func (k *Keeper) GetAllValidatorKeyRotations(ctx sdk.Context) ([]authproto.ValidatorKeyRotation, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, ValidatorRotationsKeyPrefix)
	defer iterator.Close()

	var rotations []authproto.ValidatorKeyRotation
	for ; iterator.Valid(); iterator.Next() {
		var rotation authproto.ValidatorKeyRotation
		if err := k.cdc.Unmarshal(iterator.Value(), &rotation); err != nil {
			return nil, err
		}
		rotations = append(rotations, rotation)
	}
	return rotations, nil
}

// DeleteValidatorKeyRotation removes a validator key rotation
func (k *Keeper) DeleteValidatorKeyRotation(ctx sdk.Context, validatorAddress string) {
	store := ctx.KVStore(k.storeKey)
	key := append(ValidatorRotationsKeyPrefix, []byte(validatorAddress)...)
	store.Delete(key)
}

// ============================================================================
// Session KVStore Methods
// ============================================================================

// SetSession stores a session
func (k *Keeper) SetSession(ctx sdk.Context, session *authproto.Session) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(SessionsKeyPrefix, []byte(session.SessionId)...)
	store.Set(key, bz)

	// Also update user sessions index
	return k.addUserSession(ctx, session.UserAddress, session.SessionId)
}

// GetSession retrieves a session by ID
func (k *Keeper) GetSession(ctx sdk.Context, sessionID string) (*authproto.Session, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(SessionsKeyPrefix, []byte(sessionID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	var session authproto.Session
	if err := k.cdc.Unmarshal(bz, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// GetAllSessions retrieves all sessions
func (k *Keeper) GetAllSessions(ctx sdk.Context) ([]authproto.Session, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, SessionsKeyPrefix)
	defer iterator.Close()

	var sessions []authproto.Session
	for ; iterator.Valid(); iterator.Next() {
		var session authproto.Session
		if err := k.cdc.Unmarshal(iterator.Value(), &session); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// DeleteSession removes a session
func (k *Keeper) DeleteSession(ctx sdk.Context, sessionID string) error {
	// Get session to find user address
	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	store := ctx.KVStore(k.storeKey)
	key := append(SessionsKeyPrefix, []byte(sessionID)...)
	store.Delete(key)

	// Also update user sessions index
	return k.removeUserSession(ctx, session.UserAddress, sessionID)
}

// addUserSession adds a session ID to a user's session list
func (k *Keeper) addUserSession(ctx sdk.Context, userAddress, sessionID string) error {
	sessions, _ := k.GetUserSessions(ctx, userAddress)
	sessions = append(sessions, sessionID)

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&authproto.SessionIDList{SessionIds: sessions})
	if err != nil {
		return fmt.Errorf("failed to marshal for SessionIds: %w", err)
	}
	key := append(UserSessionsKeyPrefix, []byte(userAddress)...)
	store.Set(key, bz)
	return nil
}

// removeUserSession removes a session ID from a user's session list
func (k *Keeper) removeUserSession(ctx sdk.Context, userAddress, sessionID string) error {
	sessions, _ := k.GetUserSessions(ctx, userAddress)
	filtered := make([]string, 0)
	for _, sid := range sessions {
		if sid != sessionID {
			filtered = append(filtered, sid)
		}
	}

	store := ctx.KVStore(k.storeKey)
	key := append(UserSessionsKeyPrefix, []byte(userAddress)...)

	if len(filtered) == 0 {
		store.Delete(key)
		return nil
	}

	bz, err := k.cdc.Marshal(&authproto.SessionIDList{SessionIds: filtered})
	if err != nil {
		return fmt.Errorf("failed to marshal for SessionIds: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// GetUserSessions retrieves all session IDs for a user
func (k *Keeper) GetUserSessions(ctx sdk.Context, userAddress string) ([]string, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(UserSessionsKeyPrefix, []byte(userAddress)...)
	bz := store.Get(key)
	if bz == nil {
		return []string{}, nil
	}

	var sessionList authproto.SessionIDList
	if err := k.cdc.Unmarshal(bz, &sessionList); err != nil {
		return nil, err
	}
	return sessionList.SessionIds, nil
}

// ============================================================================
// Rate Limit Config KVStore Methods
// ============================================================================

// SetRateLimitConfig stores a rate limit config
func (k *Keeper) SetRateLimitConfig(ctx sdk.Context, config *authproto.RateLimitConfig) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(RateLimitsKeyPrefix, []byte(config.UserAddress)...)
	store.Set(key, bz)
	return nil
}

// GetRateLimitConfig retrieves a rate limit config by user address
func (k *Keeper) GetRateLimitConfig(ctx sdk.Context, userAddress string) (*authproto.RateLimitConfig, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(RateLimitsKeyPrefix, []byte(userAddress)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("rate limit config not found: %s", userAddress)
	}

	var config authproto.RateLimitConfig
	if err := k.cdc.Unmarshal(bz, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetAllRateLimitConfigs retrieves all rate limit configs
func (k *Keeper) GetAllRateLimitConfigs(ctx sdk.Context) ([]authproto.RateLimitConfig, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, RateLimitsKeyPrefix)
	defer iterator.Close()

	var configs []authproto.RateLimitConfig
	for ; iterator.Valid(); iterator.Next() {
		var config authproto.RateLimitConfig
		if err := k.cdc.Unmarshal(iterator.Value(), &config); err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// DeleteRateLimitConfig removes a rate limit config
func (k *Keeper) DeleteRateLimitConfig(ctx sdk.Context, userAddress string) {
	store := ctx.KVStore(k.storeKey)
	key := append(RateLimitsKeyPrefix, []byte(userAddress)...)
	store.Delete(key)
}

// ============================================================================
// Audit Log KVStore Methods
// ============================================================================

// getNextAuditLogID generates the next audit log ID
func (k *Keeper) getNextAuditLogID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(AuditLogCounterKeyPrefix)
	var counter uint64 = 0
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}
	counter++
	store.Set(AuditLogCounterKeyPrefix, sdk.Uint64ToBigEndian(counter))
	return counter
}

// SetAuditLog stores an audit log entry in the KVStore
func (k *Keeper) SetAuditLog(ctx sdk.Context, log *authproto.AuditLog) error {
	store := ctx.KVStore(k.storeKey)

	// Generate unique ID if not set
	if log.Id == "" {
		logID := k.getNextAuditLogID(ctx)
		log.Id = fmt.Sprintf("%d", logID)
	}

	bz, err := k.cdc.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	key := append(AuditLogsKeyPrefix, []byte(log.Id)...)
	store.Set(key, bz)

	// Cleanup old logs periodically (every 100 new logs)
	if logID := k.getNextAuditLogID(ctx); logID%100 == 0 {
		k.cleanupOldAuditLogs(ctx)
	}

	return nil
}

// GetAuditLog retrieves a single audit log by ID
func (k *Keeper) GetAuditLog(ctx sdk.Context, logID string) (*authproto.AuditLog, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(AuditLogsKeyPrefix, []byte(logID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("audit log not found: %s", logID)
	}

	var log authproto.AuditLog
	if err := k.cdc.Unmarshal(bz, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// cleanupOldAuditLogs keeps only the last 10000 audit logs
func (k *Keeper) cleanupOldAuditLogs(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	// Count total logs
	var count int
	keys := make([][]byte, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, iterator.Key())
		count++
	}

	// If more than 10000, delete the oldest ones
	if count > 10000 {
		toDelete := count - 10000
		for i := 0; i < toDelete; i++ {
			store.Delete(keys[i])
		}
	}
}

// GetAllAuditLogs retrieves all audit logs
func (k *Keeper) GetAllAuditLogs(ctx sdk.Context) ([]authproto.AuditLog, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	var logs []authproto.AuditLog
	for ; iterator.Valid(); iterator.Next() {
		var log authproto.AuditLog
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// getIPFromContext extracts IP address from context
func (k *Keeper) getIPFromContext(ctx context.Context) string {
	// In a real implementation, extract from gRPC metadata
	return "unknown"
}

// ============================================================================
// Permission Helper Methods
// ============================================================================

// HasPermission checks if an address has a specific permission (context version)
func (k *Keeper) HasPermission(ctx sdk.Context, address, permission string) bool {
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, address)
	if err == nil && len(assignments) > 0 {
		now := ctx.BlockTime()
		for _, assignment := range assignments {
			// Check if assignment is still valid
			if assignment.ExpiresAt != nil && now.After(*assignment.ExpiresAt) {
				continue
			}

			// Get the role
			role, err := k.GetRoleFromStore(ctx, assignment.RoleName)
			if err != nil {
				continue
			}

			// Check if role has the permission
			for _, perm := range role.Permissions {
				if perm == permission || perm == types.PermissionAdmin {
					return true
				}
			}
		}
	}

	// Check emergency admin privileges
	// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
	admin, err := k.GetEmergencyAdmin(ctx, address)
	if err == nil && types.IsEmergencyAdminActive(admin, ctx.BlockTime()) {
		for _, priv := range admin.Privileges {
			if priv == permission || priv == types.PermissionAdmin {
				return true
			}
		}
	}

	return false
}

// GetRoleAssignments retrieves role assignments for an address (query helper)
func (k *Keeper) GetRoleAssignments(address string) []*authproto.RoleAssignment {
	// Stub - would need context passed from query server
	return []*authproto.RoleAssignment{}
}

// RequirePermission checks permission and returns error if not authorized
func (k *Keeper) RequirePermission(ctx sdk.Context, address, permission string) error {
	if !k.HasPermission(ctx, address, permission) {
		// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
		k.LogAudit(ctx, address, "permission_check", permission, "denied", nil, "insufficient permissions")
		return fmt.Errorf("%w: %s does not have permission %s", types.ErrInsufficientPermissions, address, permission)
	}
	return nil
}

// ============================================================================
// Cleanup Helper Methods
// ============================================================================

// CleanupExpiredSessions removes expired sessions
func (k *Keeper) CleanupExpiredSessions(ctx sdk.Context) int {
	sessions, err := k.GetAllSessions(ctx)
	if err != nil {
		return 0
	}

	count := 0
	now := ctx.BlockTime()

	for _, session := range sessions {
		if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
			if err := k.DeleteSession(ctx, session.SessionId); err == nil {
				count++
			}
		}
	}

	return count
}

// CleanupExpiredProposals removes expired multisig proposals
func (k *Keeper) CleanupExpiredProposals(ctx sdk.Context) int {
	proposals, err := k.GetAllMultisigProposals(ctx)
	if err != nil {
		return 0
	}

	count := 0
	// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
	blockTime := ctx.BlockTime()
	for _, proposal := range proposals {
		if types.IsProposalExpired(&proposal, blockTime) &&
			proposal.Status != authproto.ProposalStatus_PROPOSAL_STATUS_EXECUTED {
			proposal.Status = authproto.ProposalStatus_PROPOSAL_STATUS_EXPIRED
			if err := k.SetMultisigProposal(ctx, &proposal); err == nil {
				count++
			}
		}
	}

	return count
}

// ResetRateLimitWindow resets rate limit counters if window has passed
func (k *Keeper) ResetRateLimitWindow(ctx sdk.Context, userAddress string) error {
	config, err := k.GetRateLimitConfig(ctx, userAddress)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	now := ctx.BlockTime()
	windowStart := config.WindowStart

	if windowStart.IsZero() {
		return nil
	}

	// Reset minute counter
	if now.Sub(windowStart) >= time.Minute {
		config.CurrentMinuteCount = 0
	}

	// Reset hour counter
	if now.Sub(windowStart) >= time.Hour {
		config.CurrentHourCount = 0
	}

	// Reset day counter
	if now.Sub(windowStart) >= 24*time.Hour {
		config.CurrentDayCount = 0
		config.WindowStart = now
	}

	if err := k.SetRateLimitConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to persist rate limit config: %w", err)
	}
	return nil
}

// CheckRateLimit checks if a user has exceeded their rate limit and increments counters.
//
// This method retrieves or creates a rate limit configuration for the user, resets
// expired time windows, checks against minute/hour/day limits, and increments counters.
//
// Parameters:
//   - ctx: SDK context for state access
//   - userAddress: Address of the user to check
//
// Returns:
//   - error: ErrRateLimitExceeded if any limit is exceeded, or other errors
//
// Security considerations:
//   - Creates default config from params if user has no custom limit
//   - Automatically resets counters when time windows expire
//   - Checks all three time windows (minute, hour, day)
//   - Increments counters atomically after successful check
func (k *Keeper) CheckRateLimit(ctx sdk.Context, userAddress string) error {
	if userAddress == "" {
		return fmt.Errorf("user address required")
	}

	// Try to get existing config, or create default from params
	config, err := k.GetRateLimitConfig(ctx, userAddress)
	if err != nil {
		// No custom config, create default from params
		params, err := k.GetParams(ctx)
		if err != nil {
			return fmt.Errorf("failed to get: %w", err)
		}

		config = &authproto.RateLimitConfig{
			UserAddress:        userAddress,
			RequestsPerMinute:  params.DefaultRequestsPerMinute,
			RequestsPerHour:    params.DefaultRequestsPerHour,
			RequestsPerDay:     params.DefaultRequestsPerDay,
			CurrentMinuteCount: 0,
			CurrentHourCount:   0,
			CurrentDayCount:    0,
			WindowStart:        ctx.BlockTime(),
		}
	}

	// Reset counters if time windows have passed
	now := ctx.BlockTime()
	windowStart := config.WindowStart

	if now.Sub(windowStart) >= time.Minute {
		config.CurrentMinuteCount = 0
	}
	if now.Sub(windowStart) >= time.Hour {
		config.CurrentHourCount = 0
	}
	if now.Sub(windowStart) >= 24*time.Hour {
		config.CurrentDayCount = 0
		config.WindowStart = now
	}

	// Check limits (only if limits are configured)
	if config.RequestsPerMinute > 0 && config.CurrentMinuteCount >= config.RequestsPerMinute {
		return types.ErrRateLimitExceeded
	}
	if config.RequestsPerHour > 0 && config.CurrentHourCount >= config.RequestsPerHour {
		return types.ErrRateLimitExceeded
	}
	if config.RequestsPerDay > 0 && config.CurrentDayCount >= config.RequestsPerDay {
		return types.ErrRateLimitExceeded
	}

	// Increment counters
	config.CurrentMinuteCount++
	config.CurrentHourCount++
	config.CurrentDayCount++

	// Save updated config
	return k.SetRateLimitConfig(ctx, config)
}

// SetCustomRateLimit sets a custom rate limit configuration for a user.
//
// This method allows an admin to override the default rate limits for a specific user.
// The admin must have the RoleAdmin permission to perform this operation.
//
// Parameters:
//   - ctx: SDK context for state access
//   - adminAddress: Address of the admin setting the limit (must have RoleAdmin)
//   - userAddress: Address of the user whose limit is being set
//   - requestsPerMinute: Maximum requests allowed per minute (0 = no limit)
//   - requestsPerHour: Maximum requests allowed per hour (0 = no limit)
//   - requestsPerDay: Maximum requests allowed per day (0 = no limit)
//
// Returns:
//   - error: Permission denied if admin lacks RoleAdmin, or validation errors
//
// Security considerations:
//   - Requires admin to have RoleAdmin permission
//   - Validates all rate limit values are non-negative
//   - Creates new config or updates existing one
//   - Preserves current counters if config exists
func (k *Keeper) SetCustomRateLimit(ctx sdk.Context, adminAddress, userAddress string, requestsPerMinute, requestsPerHour, requestsPerDay uint64) error {
	if adminAddress == "" {
		return fmt.Errorf("admin address required")
	}
	if userAddress == "" {
		return fmt.Errorf("user address required")
	}

	// Verify admin has RoleAdmin permission
	roleAssignments, err := k.GetRoleAssignmentsForAddress(ctx, adminAddress)
	if err != nil {
		return fmt.Errorf("failed to get admin roles: %w", err)
	}

	hasAdminRole := false
	for _, assignment := range roleAssignments {
		if assignment.RoleName == types.RoleAdmin {
			hasAdminRole = true
			break
		}
	}
	if !hasAdminRole {
		return types.ErrInsufficientPermissions
	}

	// Try to get existing config to preserve current counters
	existingConfig, err := k.GetRateLimitConfig(ctx, userAddress)

	var config *authproto.RateLimitConfig
	if err != nil {
		// No existing config, create new one
		config = &authproto.RateLimitConfig{
			UserAddress:        userAddress,
			RequestsPerMinute:  requestsPerMinute,
			RequestsPerHour:    requestsPerHour,
			RequestsPerDay:     requestsPerDay,
			CurrentMinuteCount: 0,
			CurrentHourCount:   0,
			CurrentDayCount:    0,
			WindowStart:        ctx.BlockTime(),
		}
	} else {
		// Update existing config, preserve counters
		config = existingConfig
		config.RequestsPerMinute = requestsPerMinute
		config.RequestsPerHour = requestsPerHour
		config.RequestsPerDay = requestsPerDay
	}

	// Validate config
	if err := types.ValidateRateLimitConfig(config); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return k.SetRateLimitConfig(ctx, config)
}
