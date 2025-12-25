// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// InitGenesis initializes the module state from genesis data
//nolint:govet // copylocks warning acceptable for keeper value usage in tests
func (k Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	if err := types.ValidateGenesis(data); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	logger := sdk.UnwrapSDKContext(ctx).Logger()

	// Set params
	if err := k.SetParams(sdkCtx, &data.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Import roles
	for i := range data.Roles {
		if err := k.SetRole(sdkCtx, &data.Roles[i]); err != nil {
			return fmt.Errorf("failed to set role: %w", err)
		}
	}

	// Import role assignments
	for i := range data.RoleAssignments {
		if err := k.SetRoleAssignment(sdkCtx, &data.RoleAssignments[i]); err != nil {
			return fmt.Errorf("failed to set role assignment: %w", err)
		}
	}

	// Import multisig wallets
	for i := range data.MultisigWallets {
		if err := k.SetMultisigWallet(sdkCtx, &data.MultisigWallets[i]); err != nil {
			return fmt.Errorf("failed to set multisig wallet: %w", err)
		}
	}

	// Import multisig proposals
	for i := range data.MultisigProposals {
		if err := k.SetMultisigProposal(sdkCtx, &data.MultisigProposals[i]); err != nil {
			return fmt.Errorf("failed to set multisig proposal: %w", err)
		}
	}

	// Import time-locked actions
	for i := range data.TimeLockedActions {
		if err := k.SetTimeLockedAction(sdkCtx, &data.TimeLockedActions[i]); err != nil {
			return fmt.Errorf("failed to set time-locked action: %w", err)
		}
	}

	// Import emergency admin records
	for i := range data.EmergencyAdmins {
		if err := k.SetEmergencyAdmin(sdkCtx, &data.EmergencyAdmins[i]); err != nil {
			return fmt.Errorf("failed to set emergency admin: %w", err)
		}
	}

	// Import validator key rotations
	for i := range data.ValidatorKeyRotations {
		if err := k.SetValidatorKeyRotation(sdkCtx, &data.ValidatorKeyRotations[i]); err != nil {
			return fmt.Errorf("failed to set validator key rotation: %w", err)
		}
	}

	// Import sessions
	for i := range data.Sessions {
		if err := k.SetSession(sdkCtx, &data.Sessions[i]); err != nil {
			return fmt.Errorf("failed to set session: %w", err)
		}
	}

	// Import rate limit configs
	for i := range data.RateLimitConfigs {
		if err := k.SetRateLimitConfig(sdkCtx, &data.RateLimitConfigs[i]); err != nil {
			return fmt.Errorf("failed to set rate limit config: %w", err)
		}
	}

	// Import audit logs
	for i := range data.AuditLogs {
		auditBytes, err := k.cdc.Marshal(&data.AuditLogs[i])
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
//nolint:govet // copylocks warning acceptable for keeper value usage in tests
func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Export params
	params, err := k.GetParams(sdkCtx)
	var paramsValue authproto.Params
	if err != nil {
		defaultParams := types.DefaultParams()
		paramsValue = *defaultParams
	} else {
		paramsValue = params
	}

	genesis := &types.GenesisState{
		Params:                paramsValue,
		Roles:                 make([]authproto.Role, 0),
		RoleAssignments:       make([]authproto.RoleAssignment, 0),
		MultisigWallets:       make([]authproto.MultisigWallet, 0),
		MultisigProposals:     make([]authproto.MultisigProposal, 0),
		TimeLockedActions:     make([]authproto.TimeLockedAction, 0),
		EmergencyAdmins:       make([]authproto.EmergencyAdmin, 0),
		ValidatorKeyRotations: make([]authproto.ValidatorKeyRotation, 0),
		Sessions:              make([]authproto.Session, 0),
		RateLimitConfigs:      make([]authproto.RateLimitConfig, 0),
		AuditLogs:             make([]authproto.AuditLog, 0),
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
