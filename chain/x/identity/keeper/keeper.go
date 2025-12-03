package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// Keeper manages the identity module state
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	authority    string
	logger       log.Logger
}

// NewKeeper creates a new identity keeper
func NewKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	authority string,
	logger log.Logger,
) *Keeper {
	return &Keeper{
		storeService: storeService,
		cdc:          cdc,
		authority:    authority,
		logger:       logger,
	}
}

// GetAuthority returns the module authority address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns the module logger
func (k *Keeper) Logger() log.Logger {
	return k.logger
}

// ============================================================================
// Genesis Methods
// ============================================================================

// InitGenesis initializes the module state from genesis
func (k *Keeper) InitGenesis(ctx sdk.Context, gs *types.GenesisState) error {
	// Validate genesis state
	if gs == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)

	// Set params
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Initialize roles
	for _, role := range gs.Roles {
		if err := k.SetRole(ctx, role); err != nil {
			return fmt.Errorf("failed to set role %s: %w", role.Name, err)
		}
	}

	// Initialize default roles if none provided
	if len(gs.Roles) == 0 {
		if err := k.initializeDefaultRoles(ctx); err != nil {
			return fmt.Errorf("failed to initialize default roles: %w", err)
		}
	}

	// Initialize role assignments
	for _, assignment := range gs.RoleAssignments {
		if err := k.SetRoleAssignment(ctx, assignment); err != nil {
			return fmt.Errorf("failed to set role assignment for %s: %w", assignment.Address, err)
		}
	}

	// Initialize audit logs
	for _, log := range gs.AuditLogs {
		if err := k.SetAuditLog(ctx, log); err != nil {
			return fmt.Errorf("failed to set audit log: %w", err)
		}
	}

	// Initialize sessions
	for _, session := range gs.Sessions {
		if err := k.SetSession(ctx, session); err != nil {
			return fmt.Errorf("failed to set session %s: %w", session.Id, err)
		}
	}

	// Initialize rate limit configs
	for _, config := range gs.RateLimits {
		if err := k.SetRateLimitConfig(ctx, config); err != nil {
			return fmt.Errorf("failed to set rate limit config for %s: %w", config.UserAddress, err)
		}
	}

	// Initialize multisig wallets
	for _, wallet := range gs.MultisigWallets {
		if err := k.SetMultisigWallet(ctx, wallet); err != nil {
			return fmt.Errorf("failed to set multisig wallet %s: %w", wallet.Id, err)
		}
	}

	// Initialize multisig proposals
	for _, proposal := range gs.MultisigProposals {
		if err := k.SetMultisigProposal(ctx, proposal); err != nil {
			return fmt.Errorf("failed to set multisig proposal %s: %w", proposal.Id, err)
		}
	}

	// Initialize time-locked actions
	for _, action := range gs.TimeLockedActions {
		if err := k.SetTimeLockedAction(ctx, action); err != nil {
			return fmt.Errorf("failed to set time-locked action %s: %w", action.Id, err)
		}
	}

	// Initialize emergency admins
	for _, admin := range gs.EmergencyAdmins {
		if err := k.SetEmergencyAdmin(ctx, admin); err != nil {
			return fmt.Errorf("failed to set emergency admin %s: %w", admin.Address, err)
		}
	}

	// Initialize validator rotations
	for _, rotation := range gs.ValidatorRotations {
		if err := k.SetValidatorRotation(ctx, rotation); err != nil {
			return fmt.Errorf("failed to set validator rotation for %s: %w", rotation.ValidatorAddress, err)
		}
	}

	// Initialize identity records
	for _, record := range gs.IdentityRecords {
		if err := k.SetIdentityRecord(ctx, record); err != nil {
			return fmt.Errorf("failed to set identity record %s: %w", record.Did, err)
		}
	}

	// Initialize credential revocations
	for _, revocation := range gs.CredentialRevocations {
		store := k.storeService.OpenKVStore(ctx)
		key := types.GetCredentialRevocationKey(revocation.CredentialId)
		bz, err := k.cdc.Marshal(revocation)
		if err != nil {
			return fmt.Errorf("failed to marshal credential revocation %s: %w", revocation.CredentialId, err)
		}
		if err := store.Set(key, bz); err != nil {
			return fmt.Errorf("failed to set credential revocation %s: %w", revocation.CredentialId, err)
		}
	}

	// Initialize DID key rotations
	for _, rotation := range gs.DidKeyRotations {
		if err := k.SetDIDKeyRotation(ctx, rotation); err != nil {
			return fmt.Errorf("failed to set DID key rotation for %s: %w", rotation.Did, err)
		}
	}

	// Initialize DID key histories
	for _, history := range gs.DidKeyHistories {
		if err := k.SetDIDKeyHistory(ctx, history); err != nil {
			return fmt.Errorf("failed to set DID key history for %s: %w", history.Did, err)
		}
	}

	// Initialize change requests
	for _, request := range gs.ChangeRequests {
		if err := k.SetChangeRequest(ctx, request); err != nil {
			return fmt.Errorf("failed to set change request %s: %w", request.Id, err)
		}
	}

	// Initialize change history
	for _, history := range gs.ChangeHistory {
		if err := k.SetChangeHistory(ctx, history); err != nil {
			return fmt.Errorf("failed to set change history: %w", err)
		}
	}

	suspendedBz := []byte{0x00}
	if gs.IdentityChangesSuspended {
		suspendedBz = []byte{0x01}
	}
	if err := store.Set(types.SuspendedKey, suspendedBz); err != nil {
		return fmt.Errorf("failed to set suspended flag: %w", err)
	}

	// Set counters
	if err := store.Set(types.AuditLogCounterPrefix, sdk.Uint64ToBigEndian(gs.NextAuditLogId)); err != nil {
		return fmt.Errorf("failed to set audit log counter: %w", err)
	}
	// Note: NextChangeRequestId not in proto yet, defaulting to 1
	if err := store.Set(types.ChangeRequestCounterPrefix, sdk.Uint64ToBigEndian(1)); err != nil {
		return fmt.Errorf("failed to set change request counter: %w", err)
	}

	return nil
}

// ExportGenesis exports the module state to genesis
func (k *Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	store := k.storeService.OpenKVStore(ctx)

	// Get params
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	// Get all roles
	roles, err := k.GetAllRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	// Get all role assignments
	roleAssignments, err := k.GetAllRoleAssignments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get role assignments: %w", err)
	}

	// Get all audit logs
	auditLogs, err := k.GetAllAuditLogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	// Get all sessions
	sessions, err := k.GetAllSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Get all rate limit configs
	rateLimitConfigs, err := k.GetAllRateLimitConfigs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit configs: %w", err)
	}

	// Get all multisig wallets
	multisigWallets, err := k.GetAllMultisigWallets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get multisig wallets: %w", err)
	}

	// Get all multisig proposals
	multisigProposals, err := k.GetAllMultisigProposals(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get multisig proposals: %w", err)
	}

	// Get all time-locked actions
	timeLockedActions, err := k.GetAllTimeLockedActions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get time-locked actions: %w", err)
	}

	// Get all emergency admins
	emergencyAdmins, err := k.GetAllEmergencyAdmins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get emergency admins: %w", err)
	}

	// Get all validator rotations
	validatorRotations, err := k.GetAllValidatorRotations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator rotations: %w", err)
	}

	// Get all identity records
	identityRecords, err := k.GetAllIdentityRecords(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity records: %w", err)
	}

	// Get all credential revocations
	credentialRevocations, err := k.GetAllCredentialRevocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential revocations: %w", err)
	}

	// Get all DID key rotations
	didKeyRotations, err := k.GetAllDIDKeyRotations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DID key rotations: %w", err)
	}

	// Get all DID key histories
	didKeyHistories, err := k.GetAllDIDKeyHistories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DID key histories: %w", err)
	}

	// Get all change requests
	changeRequests, err := k.GetAllChangeRequests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get change requests: %w", err)
	}

	// Get all change history
	changeHistory, err := k.GetAllChangeHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get change history: %w", err)
	}

	suspendedBz, err := store.Get(types.SuspendedKey)
	suspended := false
	if err == nil && suspendedBz != nil && len(suspendedBz) > 0 {
		suspended = suspendedBz[0] == 0x01
	}

	// Get counters
	nextAuditLogID := uint64(1)
	if bz, err := store.Get(types.AuditLogCounterPrefix); err == nil && bz != nil {
		nextAuditLogID = sdk.BigEndianToUint64(bz)
	}

	// nextChangeRequestID counter skipped - not in proto yet

	return &types.GenesisState{
		Params:                   params,
		Roles:                    roles,
		RoleAssignments:          roleAssignments,
		AuditLogs:                auditLogs,
		Sessions:                 sessions,
		RateLimits:               rateLimitConfigs,
		MultisigWallets:          multisigWallets,
		MultisigProposals:        multisigProposals,
		TimeLockedActions:        timeLockedActions,
		EmergencyAdmins:          emergencyAdmins,
		ValidatorRotations:       validatorRotations,
		IdentityRecords:          identityRecords,
		CredentialRevocations:    credentialRevocations,
		ChangeRequests:           changeRequests,
		ChangeHistory:            changeHistory,
		IdentityChangesSuspended: suspended,
		NextAuditLogId:           nextAuditLogID,
	}, nil
}

// ============================================================================
// Params Methods
// ============================================================================

// GetParams retrieves the module parameters
func (k *Keeper) GetParams(ctx sdk.Context) (*types.Params, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return types.DefaultParams(), nil
	}

	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return nil, err
	}
	return &params, nil
}

// SetParams sets the module parameters
func (k *Keeper) SetParams(ctx sdk.Context, params *types.Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(params)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsKey, bz)
}

// ============================================================================
// Helper Methods
// ============================================================================

// initializeDefaultRoles creates the default system roles
func (k *Keeper) initializeDefaultRoles(ctx sdk.Context) error {
	now := ctx.BlockTime()

	// Admin role with all permissions
	adminRole := &types.Role{
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
			types.PermissionManageIdentity,
			types.PermissionVerifyIdentity,
			types.PermissionApproveChangeRequest,
		},
		Description: "Full administrative access",
		CreatedAt: timestamppb.New(now),
		CreatedBy:   "",
		IsSystemRole: true,
		UpdatedAt: timestamppb.New(now),
	}
	if err := k.SetRole(ctx, adminRole); err != nil {
		return err
	}

	// Moderator role
	moderatorRole := &types.Role{
		Name: types.RoleModerator,
		Permissions: []string{
			types.PermissionAssignRole,
			types.PermissionRevokeRole,
			types.PermissionManageSession,
			types.PermissionViewAuditLogs,
			types.PermissionVerifyIdentity,
		},
		Description: "Moderate user permissions and verify identities",
		CreatedAt: timestamppb.New(now),
		CreatedBy:   "",
		IsSystemRole: true,
		UpdatedAt: timestamppb.New(now),
	}
	if err := k.SetRole(ctx, moderatorRole); err != nil {
		return err
	}

	// Validator role
	validatorRole := &types.Role{
		Name: types.RoleValidator,
		Permissions: []string{
			types.PermissionRotateValidatorKey,
		},
		Description: "Validator-specific permissions",
		CreatedAt: timestamppb.New(now),
		CreatedBy:   "",
		IsSystemRole: true,
		UpdatedAt: timestamppb.New(now),
	}
	if err := k.SetRole(ctx, validatorRole); err != nil {
		return err
	}

	// User role (basic)
	userRole := &types.Role{
		Name:        types.RoleUser,
		Permissions: []string{},
		Description: "Basic user permissions",
		CreatedAt: timestamppb.New(now),
		CreatedBy:   "",
		IsSystemRole: true,
		UpdatedAt: timestamppb.New(now),
	}
	return k.SetRole(ctx, userRole)
}

// GetAllDataForPrefix retrieves all items with a given prefix
func (k *Keeper) GetAllDataForPrefix(ctx sdk.Context, prefix []byte) ([][]byte, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var items [][]byte
	for ; iterator.Valid(); iterator.Next() {
		items = append(items, iterator.Value())
	}
	return items, nil
}
