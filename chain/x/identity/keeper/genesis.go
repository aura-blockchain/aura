package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// InitGenesis initializes the module state from genesis
func (k *Keeper) InitGenesis(ctx sdk.Context, gs *types.GenesisState) error {
	// Validate genesis state
	if err := types.ValidateGenesisState(gs); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	store := k.storeService.OpenKVStore(ctx)

	// Set params
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Initialize roles
	seenRoleNames := make(map[string]bool)
	for _, role := range gs.Roles {
		// CRITICAL SECURITY: Detect duplicate role names during import
		if seenRoleNames[role.Name] {
			return fmt.Errorf("duplicate role name in genesis: %s", role.Name)
		}
		seenRoleNames[role.Name] = true

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
	seenSessionIDs := make(map[string]bool)
	for _, session := range gs.Sessions {
		// CRITICAL SECURITY: Detect duplicate session IDs during import
		if seenSessionIDs[session.Id] {
			return fmt.Errorf("duplicate session ID in genesis: %s", session.Id)
		}
		seenSessionIDs[session.Id] = true

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
	seenWalletIDs := make(map[string]bool)
	for _, wallet := range gs.MultisigWallets {
		// CRITICAL SECURITY: Detect duplicate wallet IDs during import
		if seenWalletIDs[wallet.Id] {
			return fmt.Errorf("duplicate multisig wallet ID in genesis: %s", wallet.Id)
		}
		seenWalletIDs[wallet.Id] = true

		if err := k.SetMultisigWallet(ctx, wallet); err != nil {
			return fmt.Errorf("failed to set multisig wallet %s: %w", wallet.Id, err)
		}
	}

	// Initialize multisig proposals
	seenProposalIDs := make(map[string]bool)
	for _, proposal := range gs.MultisigProposals {
		// CRITICAL SECURITY: Detect duplicate proposal IDs during import
		if seenProposalIDs[proposal.Id] {
			return fmt.Errorf("duplicate multisig proposal ID in genesis: %s", proposal.Id)
		}
		seenProposalIDs[proposal.Id] = true

		if err := k.SetMultisigProposal(ctx, proposal); err != nil {
			return fmt.Errorf("failed to set multisig proposal %s: %w", proposal.Id, err)
		}
	}

	// Initialize time-locked actions
	seenActionIDs := make(map[string]bool)
	for _, action := range gs.TimeLockedActions {
		// CRITICAL SECURITY: Detect duplicate action IDs during import
		if seenActionIDs[action.Id] {
			return fmt.Errorf("duplicate time-locked action ID in genesis: %s", action.Id)
		}
		seenActionIDs[action.Id] = true

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
	seenDIDs := make(map[string]bool)
	for _, record := range gs.IdentityRecords {
		// CRITICAL SECURITY: Detect duplicate DIDs during import
		if seenDIDs[record.Did] {
			return fmt.Errorf("duplicate DID in genesis: %s", record.Did)
		}
		seenDIDs[record.Did] = true

		if err := k.SetIdentityRecord(ctx, record); err != nil {
			return fmt.Errorf("failed to set identity record %s: %w", record.Did, err)
		}
	}

	// Initialize credential revocations
	seenCredentialIDs := make(map[string]bool)
	for _, revocation := range gs.CredentialRevocations {
		// CRITICAL SECURITY: Detect duplicate credential IDs during import
		if seenCredentialIDs[revocation.CredentialId] {
			return fmt.Errorf("duplicate credential revocation in genesis: %s", revocation.CredentialId)
		}
		seenCredentialIDs[revocation.CredentialId] = true

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
	seenChangeRequestIDs := make(map[string]bool)
	for _, request := range gs.ChangeRequests {
		// CRITICAL SECURITY: Detect duplicate change request IDs during import
		if seenChangeRequestIDs[request.Id] {
			return fmt.Errorf("duplicate change request ID in genesis: %s", request.Id)
		}
		seenChangeRequestIDs[request.Id] = true

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

	// Set identity changes suspended flag
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

	// Get identity changes suspended flag
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
		DidKeyRotations:          didKeyRotations,
		DidKeyHistories:          didKeyHistories,
		ChangeRequests:           changeRequests,
		ChangeHistory:            changeHistory,
		IdentityChangesSuspended: suspended,
		NextAuditLogId:           nextAuditLogID,
	}, nil
}
