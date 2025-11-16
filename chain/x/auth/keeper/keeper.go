package keeper

import (
	"context"
	"fmt"
	"sync"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	mu        sync.RWMutex
	auditLogs map[string][]*authproto.AuditLog
	storeKey  storetypes.StoreKey
	cdc       codec.BinaryCodec
}

// NewKeeper creates a new auth keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{
		auditLogs: make(map[string][]*authproto.AuditLog),
		cdc:       cdc,
		storeKey:  storeKey,
	}
}

// initializeDefaultRoles creates predefined roles
func (k *Keeper) initializeDefaultRoles(ctx sdk.Context) error {
	now := time.Now()
	nowProto := timestamppb.New(now)

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
		CreatedAt:   nowProto,
		UpdatedAt:   nowProto,
	}
	if err := k.SetRole(ctx, adminRole); err != nil {
		return err
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
		CreatedAt:   nowProto,
		UpdatedAt:   nowProto,
	}
	if err := k.SetRole(ctx, moderatorRole); err != nil {
		return err
	}

	// Validator role
	validatorRole := &authproto.Role{
		Name: types.RoleValidator,
		Permissions: []string{
			types.PermissionRotateValidatorKey,
		},
		Description: "Validator-specific permissions",
		CreatedAt:   nowProto,
		UpdatedAt:   nowProto,
	}
	if err := k.SetRole(ctx, validatorRole); err != nil {
		return err
	}

	// User role
	userRole := &authproto.Role{
		Name:        types.RoleUser,
		Permissions: []string{},
		Description: "Basic user permissions",
		CreatedAt:   nowProto,
		UpdatedAt:   nowProto,
	}
	return k.SetRole(ctx, userRole)
}

// ============================================================================
// Params KVStore Methods
// ============================================================================

// GetParams returns the module parameters
func (k *Keeper) GetParams(ctx sdk.Context) (*authproto.Params, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ParamsKeyPrefix)
	if bz == nil {
		// Return default params if not set
		return &authproto.Params{}, nil
	}

	var params authproto.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return nil, err
	}
	return &params, nil
}

// SetParams updates the module parameters
func (k *Keeper) SetParams(ctx sdk.Context, params *authproto.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(params)
	if err != nil {
		return err
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
		return err
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
func (k *Keeper) GetAllRoles(ctx sdk.Context) ([]*authproto.Role, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, RolesKeyPrefix)
	defer iterator.Close()

	var roles []*authproto.Role
	for ; iterator.Valid(); iterator.Next() {
		var role authproto.Role
		if err := k.cdc.Unmarshal(iterator.Value(), &role); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
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
		return err
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
	return assignmentList.Assignments, nil
}

// DeleteRoleAssignment removes a specific role assignment
func (k *Keeper) DeleteRoleAssignment(ctx sdk.Context, address, roleName string) error {
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, address)
	if err != nil {
		return err
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
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetAllRoleAssignments retrieves all role assignments
func (k *Keeper) GetAllRoleAssignments(ctx sdk.Context) ([]*authproto.RoleAssignment, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, RoleAssignmentsKeyPrefix)
	defer iterator.Close()

	var allAssignments []*authproto.RoleAssignment
	for ; iterator.Valid(); iterator.Next() {
		var assignmentList authproto.RoleAssignmentList
		if err := k.cdc.Unmarshal(iterator.Value(), &assignmentList); err != nil {
			return nil, err
		}
		allAssignments = append(allAssignments, assignmentList.Assignments...)
	}
	return allAssignments, nil
}

// ============================================================================
// Multisig Wallet KVStore Methods
// ============================================================================

// SetMultisigWallet stores a multisig wallet
func (k *Keeper) SetMultisigWallet(ctx sdk.Context, wallet *authproto.MultisigWallet) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(wallet)
	if err != nil {
		return err
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
func (k *Keeper) GetAllMultisigWallets(ctx sdk.Context) ([]*authproto.MultisigWallet, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, MultisigWalletsKeyPrefix)
	defer iterator.Close()

	var wallets []*authproto.MultisigWallet
	for ; iterator.Valid(); iterator.Next() {
		var wallet authproto.MultisigWallet
		if err := k.cdc.Unmarshal(iterator.Value(), &wallet); err != nil {
			return nil, err
		}
		wallets = append(wallets, &wallet)
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
		return err
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
func (k *Keeper) GetAllMultisigProposals(ctx sdk.Context) ([]*authproto.MultisigProposal, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, MultisigProposalsKeyPrefix)
	defer iterator.Close()

	var proposals []*authproto.MultisigProposal
	for ; iterator.Valid(); iterator.Next() {
		var proposal authproto.MultisigProposal
		if err := k.cdc.Unmarshal(iterator.Value(), &proposal); err != nil {
			return nil, err
		}
		proposals = append(proposals, &proposal)
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
		return err
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
func (k *Keeper) GetAllTimeLockedActions(ctx sdk.Context) ([]*authproto.TimeLockedAction, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, TimeLockedActionsKeyPrefix)
	defer iterator.Close()

	var actions []*authproto.TimeLockedAction
	for ; iterator.Valid(); iterator.Next() {
		var action authproto.TimeLockedAction
		if err := k.cdc.Unmarshal(iterator.Value(), &action); err != nil {
			return nil, err
		}
		actions = append(actions, &action)
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
		return err
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
func (k *Keeper) GetAllEmergencyAdmins(ctx sdk.Context) ([]*authproto.EmergencyAdmin, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, EmergencyAdminsKeyPrefix)
	defer iterator.Close()

	var admins []*authproto.EmergencyAdmin
	for ; iterator.Valid(); iterator.Next() {
		var admin authproto.EmergencyAdmin
		if err := k.cdc.Unmarshal(iterator.Value(), &admin); err != nil {
			return nil, err
		}
		admins = append(admins, &admin)
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
		return err
	}
	key := append(ValidatorRotationsKeyPrefix, []byte(rotation.ValidatorAddress)...)
	store.Set(key, bz)
	return nil
}

// GetAllValidatorKeyRotations retrieves all validator key rotations
func (k *Keeper) GetAllValidatorKeyRotations(ctx sdk.Context) ([]*authproto.ValidatorKeyRotation, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, ValidatorRotationsKeyPrefix)
	defer iterator.Close()

	var rotations []*authproto.ValidatorKeyRotation
	for ; iterator.Valid(); iterator.Next() {
		var rotation authproto.ValidatorKeyRotation
		if err := k.cdc.Unmarshal(iterator.Value(), &rotation); err != nil {
			return nil, err
		}
		rotations = append(rotations, &rotation)
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
		return err
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
func (k *Keeper) GetAllSessions(ctx sdk.Context) ([]*authproto.Session, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, SessionsKeyPrefix)
	defer iterator.Close()

	var sessions []*authproto.Session
	for ; iterator.Valid(); iterator.Next() {
		var session authproto.Session
		if err := k.cdc.Unmarshal(iterator.Value(), &session); err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}
	return sessions, nil
}

// DeleteSession removes a session
func (k *Keeper) DeleteSession(ctx sdk.Context, sessionID string) error {
	// Get session to find user address
	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return err
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
		return err
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
		return err
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
		return err
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
func (k *Keeper) GetAllRateLimitConfigs(ctx sdk.Context) ([]*authproto.RateLimitConfig, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, RateLimitsKeyPrefix)
	defer iterator.Close()

	var configs []*authproto.RateLimitConfig
	for ; iterator.Valid(); iterator.Next() {
		var config authproto.RateLimitConfig
		if err := k.cdc.Unmarshal(iterator.Value(), &config); err != nil {
			return nil, err
		}
		configs = append(configs, &config)
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

// cleanupOldAuditLogs keeps only the last 10000 audit logs
func (k *Keeper) cleanupOldAuditLogs(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	// Count total logs
	var count int
	var keys [][]byte
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
func (k *Keeper) GetAllAuditLogs(ctx sdk.Context) ([]*authproto.AuditLog, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	var logs []*authproto.AuditLog
	for ; iterator.Valid(); iterator.Next() {
		var log authproto.AuditLog
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
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

// HasPermission checks if an address has a specific permission
func (k *Keeper) HasPermission(ctx sdk.Context, address, permission string) bool {
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, address)
	if err != nil || len(assignments) == 0 {
		return false
	}

	now := time.Now()
	for _, assignment := range assignments {
		// Check if assignment is still valid
		if assignment.ExpiresAt != nil && now.After(assignment.ExpiresAt.AsTime()) {
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

	// Check emergency admin privileges
	admin, err := k.GetEmergencyAdmin(ctx, address)
	if err == nil && types.IsEmergencyAdminActive(admin) {
		for _, priv := range admin.Privileges {
			if priv == permission || priv == types.PermissionAdmin {
				return true
			}
		}
	}

	return false
}

// RequirePermission checks permission and returns error if not authorized
func (k *Keeper) RequirePermission(ctx sdk.Context, address, permission string) error {
	if !k.HasPermission(ctx, address, permission) {
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
	now := time.Now()

	for _, session := range sessions {
		if session.ExpiresAt != nil && now.After(session.ExpiresAt.AsTime()) {
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
	for _, proposal := range proposals {
		if types.IsProposalExpired(proposal) &&
			proposal.Status != authproto.ProposalStatus_PROPOSAL_STATUS_EXECUTED {
			proposal.Status = authproto.ProposalStatus_PROPOSAL_STATUS_EXPIRED
			if err := k.SetMultisigProposal(ctx, proposal); err == nil {
				count++
			}
		}
	}

	return count
}

// ResetRateLimitWindow resets rate limit counters if window has passed
func (k *Keeper) ResetRateLimitWindow(ctx sdk.Context, userAddress string) {
	config, err := k.GetRateLimitConfig(ctx, userAddress)
	if err != nil {
		return
	}

	now := time.Now()
	windowStart := config.WindowStart

	if windowStart == nil {
		return
	}

	// Reset minute counter
	if now.Sub(windowStart.AsTime()) >= time.Minute {
		config.CurrentMinuteCount = 0
	}

	// Reset hour counter
	if now.Sub(windowStart.AsTime()) >= time.Hour {
		config.CurrentHourCount = 0
	}

	// Reset day counter
	if now.Sub(windowStart.AsTime()) >= 24*time.Hour {
		config.CurrentDayCount = 0
		config.WindowStart = timestamppb.New(now)
	}

	k.SetRateLimitConfig(ctx, config)
}
