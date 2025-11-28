package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/auth/types"
)

// InitGenesis initializes the module state from genesis data
func (k Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	if err := types.ValidateGenesis(data); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	logger := sdk.UnwrapSDKContext(ctx).Logger()

	// Set params
	if data.Params != nil {
		if err := k.SetParams(sdkCtx, data.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Import roles
	for _, role := range data.Roles {
		if role == nil {
			continue
		}
		if err := k.SetRole(sdkCtx, role); err != nil {
			return fmt.Errorf("failed to set role: %w", err)
		}
	}

	// Import role assignments
	for _, assignment := range data.RoleAssignments {
		if assignment == nil {
			continue
		}
		if err := k.SetRoleAssignment(sdkCtx, assignment); err != nil {
			return fmt.Errorf("failed to set role assignment: %w", err)
		}
	}

	// Import multisig wallets
	for _, wallet := range data.MultisigWallets {
		if wallet == nil {
			continue
		}
		if err := k.SetMultisigWallet(sdkCtx, wallet); err != nil {
			return fmt.Errorf("failed to set multisig wallet: %w", err)
		}
	}

	// Import multisig proposals
	for _, proposal := range data.MultisigProposals {
		if proposal == nil {
			continue
		}
		if err := k.SetMultisigProposal(sdkCtx, proposal); err != nil {
			return fmt.Errorf("failed to set multisig proposal: %w", err)
		}
	}

	// Import time-locked actions
	for _, action := range data.TimeLockedActions {
		if action == nil {
			continue
		}
		if err := k.SetTimeLockedAction(sdkCtx, action); err != nil {
			return fmt.Errorf("failed to set time-locked action: %w", err)
		}
	}

	// Import emergency admin records
	for _, admin := range data.EmergencyAdmins {
		if admin == nil {
			continue
		}
		if err := k.SetEmergencyAdmin(sdkCtx, admin); err != nil {
			return fmt.Errorf("failed to set emergency admin: %w", err)
		}
	}

	// Import validator key rotations
	for _, rotation := range data.ValidatorKeyRotations {
		if rotation == nil {
			continue
		}
		if err := k.SetValidatorKeyRotation(sdkCtx, rotation); err != nil {
			return fmt.Errorf("failed to set validator key rotation: %w", err)
		}
	}

	// Import sessions
	for _, session := range data.Sessions {
		if session == nil {
			continue
		}
		if err := k.SetSession(sdkCtx, session); err != nil {
			return fmt.Errorf("failed to set session: %w", err)
		}
	}

	// Import rate limit configs
	for _, config := range data.RateLimitConfigs {
		if config == nil {
			continue
		}
		if err := k.SetRateLimitConfig(sdkCtx, config); err != nil {
			return fmt.Errorf("failed to set rate limit config: %w", err)
		}
	}

	// Import audit logs
	for i, auditLog := range data.AuditLogs {
		if auditLog == nil {
			continue
		}
		auditBytes, err := k.cdc.Marshal(auditLog)
		if err != nil {
			return fmt.Errorf("failed to marshal audit log: %w", err)
		}
		key := append(AuditLogsKeyPrefix, sdk.Uint64ToBigEndian(uint64(i))...)
		store.Set(key, auditBytes)
	}

	logger.Info("Auth genesis imported",
		"roles", len(data.Roles),
		"role_assignments", len(data.RoleAssignments),
		"multisig_wallets", len(data.MultisigWallets),
		"multisig_proposals", len(data.MultisigProposals),
		"time_locked_actions", len(data.TimeLockedActions),
		"emergency_admins", len(data.EmergencyAdmins),
		"validator_key_rotations", len(data.ValidatorKeyRotations),
		"sessions", len(data.Sessions),
		"rate_limit_configs", len(data.RateLimitConfigs),
		"audit_logs", len(data.AuditLogs))

	return nil
}

// ExportGenesis exports the current module state to genesis
func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Export params
	params, err := k.GetParams(sdkCtx)
	if err != nil || params == nil {
		defaultParams := types.DefaultParams()
		params = defaultParams
	}

	genesis := &types.GenesisState{
		Params:                params,
		Roles:                 []*types.Role{},
		RoleAssignments:       []*types.RoleAssignment{},
		MultisigWallets:       []*types.MultisigWallet{},
		MultisigProposals:     []*types.MultisigProposal{},
		TimeLockedActions:     []*types.TimeLockedAction{},
		EmergencyAdmins:       []*types.EmergencyAdmin{},
		ValidatorKeyRotations: []*types.ValidatorKeyRotation{},
		Sessions:              []*types.Session{},
		RateLimitConfigs:      []*types.RateLimitConfig{},
		AuditLogs:             []*types.AuditLog{},
	}

	// Export all roles
	if roles, err := k.GetAllRoles(sdkCtx); err == nil {
		genesis.Roles = roles
	}

	// Export all role assignments
	if assignments, err := k.GetAllRoleAssignments(sdkCtx); err == nil {
		genesis.RoleAssignments = assignments
	}

	// Export all multisig wallets
	if wallets, err := k.GetAllMultisigWallets(sdkCtx); err == nil {
		genesis.MultisigWallets = wallets
	}

	// Export all multisig proposals
	if proposals, err := k.GetAllMultisigProposals(sdkCtx); err == nil {
		genesis.MultisigProposals = proposals
	}

	// Export all time-locked actions
	if actions, err := k.GetAllTimeLockedActions(sdkCtx); err == nil {
		genesis.TimeLockedActions = actions
	}

	// Export all emergency admins
	if admins, err := k.GetAllEmergencyAdmins(sdkCtx); err == nil {
		genesis.EmergencyAdmins = admins
	}

	// Export all validator key rotations
	if rotations, err := k.GetAllValidatorKeyRotations(sdkCtx); err == nil {
		genesis.ValidatorKeyRotations = rotations
	}

	// Export all sessions
	if sessions, err := k.GetAllSessions(sdkCtx); err == nil {
		genesis.Sessions = sessions
	}

	// Export all rate limit configs
	if configs, err := k.GetAllRateLimitConfigs(sdkCtx); err == nil {
		genesis.RateLimitConfigs = configs
	}

	// Export all audit logs
	if logs, err := k.GetAllAuditLogs(sdkCtx); err == nil {
		genesis.AuditLogs = logs
	}

	return genesis
}
