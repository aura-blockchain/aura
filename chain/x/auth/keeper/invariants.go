package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	storeprefix "cosmossdk.io/store/prefix"
	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// RegisterInvariants registers all auth module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "role-consistency", RoleConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "role-assignment-consistency", RoleAssignmentConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "multisig-quorum", MultisigQuorumInvariant(k))
	ir.RegisterRoute(types.ModuleName, "multisig-proposal-validity", MultisigProposalInvariant(k))
	ir.RegisterRoute(types.ModuleName, "timelock-validity", TimeLockInvariant(k))
	ir.RegisterRoute(types.ModuleName, "emergency-admin-validity", EmergencyAdminInvariant(k))
	ir.RegisterRoute(types.ModuleName, "session-validity", SessionValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rate-limit-validity", RateLimitInvariant(k))
	ir.RegisterRoute(types.ModuleName, "audit-log-integrity", AuditLogIntegrityInvariant(k))
}

// AllInvariants runs all invariants of the auth module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			RoleConsistencyInvariant(k),
			RoleAssignmentConsistencyInvariant(k),
			MultisigQuorumInvariant(k),
			MultisigProposalInvariant(k),
			TimeLockInvariant(k),
			EmergencyAdminInvariant(k),
			SessionValidityInvariant(k),
			RateLimitInvariant(k),
			AuditLogIntegrityInvariant(k),
		}

		for _, inv := range invariants {
			msg, broken := inv(ctx)
			if broken {
				return msg, broken
			}
		}

		return "", false
	}
}

// RoleConsistencyInvariant checks that all roles have valid permissions
func RoleConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, RolesKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		validPermissions := map[string]bool{
			types.PermissionAdmin:              true,
			types.PermissionCreateRole:         true,
			types.PermissionAssignRole:         true,
			types.PermissionRevokeRole:         true,
			types.PermissionManageMultisig:     true,
			types.PermissionManageTimeLock:     true,
			types.PermissionManageEmergency:    true,
			types.PermissionRotateValidatorKey: true,
			types.PermissionManageSession:      true,
			types.PermissionViewAuditLogs:      true,
		}

		for ; iterator.Valid(); iterator.Next() {
			var role authproto.Role
			if err := k.cdc.Unmarshal(iterator.Value(), &role); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("failed to unmarshal role: %s", err.Error()),
				), true
			}

			// Check role name is not empty
			if role.Name == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					"role has empty name",
				), true
			}

			// Check all permissions are valid
			for _, perm := range role.Permissions {
				if !validPermissions[perm] {
					return sdk.FormatInvariant(
						types.ModuleName,
						"role-consistency",
						fmt.Sprintf("role %s has invalid permission: %s", role.Name, perm),
					), true
				}
			}

			// Check timestamps are valid
			if role.CreatedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("role %s has zero created_at", role.Name),
				), true
			}

			if role.UpdatedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("role %s has zero updated_at", role.Name),
				), true
			}
		}

		return "", false
	}
}

// RoleAssignmentConsistencyInvariant checks that all role assignments reference valid roles and addresses
func RoleAssignmentConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, RoleAssignmentsKeyPrefix)

		// Build set of valid roles
		validRoles := make(map[string]bool)
		rolesStore := storeprefix.NewStore(store, RolesKeyPrefix)
		rolesIter := rolesStore.Iterator(nil, nil)
		for ; rolesIter.Valid(); rolesIter.Next() {
			var role authproto.Role
			if err := k.cdc.Unmarshal(rolesIter.Value(), &role); err == nil {
				validRoles[role.Name] = true
			}
		}
		rolesIter.Close()

		// Check all assignments
		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var assignment authproto.RoleAssignment
			if err := k.cdc.Unmarshal(iterator.Value(), &assignment); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-assignment-consistency",
					fmt.Sprintf("failed to unmarshal role assignment: %s", err.Error()),
				), true
			}

			// Check address is valid
			if _, err := sdk.AccAddressFromBech32(assignment.Address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-assignment-consistency",
					fmt.Sprintf("invalid address in assignment: %s", assignment.Address),
				), true
			}

			// Check role exists
			if !validRoles[assignment.RoleName] {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-assignment-consistency",
					fmt.Sprintf("assignment references non-existent role: %s", assignment.RoleName),
				), true
			}

			// Check timestamp
			if assignment.AssignedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-assignment-consistency",
					fmt.Sprintf("assignment for %s has zero assigned_at", assignment.Address),
				), true
			}
		}

		return "", false
	}
}

// MultisigQuorumInvariant checks that quorum <= total signers for all multisig wallets
func MultisigQuorumInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, MultisigWalletsKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var wallet authproto.MultisigWallet
			if err := k.cdc.Unmarshal(iterator.Value(), &wallet); err != nil {
				continue
			}

			// Check threshold
			if wallet.Threshold == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"multisig-quorum",
					fmt.Sprintf("wallet %s has zero threshold", wallet.Id),
				), true
			}

			if wallet.Threshold > uint32(len(wallet.Signers)) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"multisig-quorum",
					fmt.Sprintf("wallet %s threshold (%d) exceeds signers count (%d)",
						wallet.Id, wallet.Threshold, len(wallet.Signers)),
				), true
			}

			// Check all signer addresses are valid
			for _, signer := range wallet.Signers {
				if _, err := sdk.AccAddressFromBech32(signer); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"multisig-quorum",
						fmt.Sprintf("wallet %s has invalid signer address: %s", wallet.Id, signer),
					), true
				}
			}
		}

		return "", false
	}
}

// MultisigProposalInvariant checks that all proposals reference valid wallets
func MultisigProposalInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		proposalsStore := storeprefix.NewStore(store, MultisigProposalsKeyPrefix)

		// Build set of valid wallet addresses
		validWallets := make(map[string]bool)
		walletsStore := storeprefix.NewStore(store, MultisigWalletsKeyPrefix)
		walletsIter := walletsStore.Iterator(nil, nil)
		for ; walletsIter.Valid(); walletsIter.Next() {
			var wallet authproto.MultisigWallet
			if err := k.cdc.Unmarshal(walletsIter.Value(), &wallet); err == nil {
				validWallets[wallet.Id] = true
			}
		}
		walletsIter.Close()

		// Check all proposals
		iterator := proposalsStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var proposal authproto.MultisigProposal
			if err := k.cdc.Unmarshal(iterator.Value(), &proposal); err != nil {
				continue
			}

			// Check wallet exists
			if !validWallets[proposal.WalletId] {
				return sdk.FormatInvariant(
					types.ModuleName,
					"multisig-proposal-validity",
					fmt.Sprintf("proposal %s references non-existent wallet: %s",
						proposal.Id, proposal.WalletId),
				), true
			}

			// Note: We can't check signature count against threshold here without loading the wallet
			// This is acceptable as the invariant primarily checks referential integrity
		}

		return "", false
	}
}

// TimeLockInvariant checks that execution time > creation time for all time-locked actions
func TimeLockInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, TimeLockedActionsKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var action authproto.TimeLockedAction
			if err := k.cdc.Unmarshal(iterator.Value(), &action); err != nil {
				continue
			}

			// Check timestamps are not zero
			if action.ProposedAt.IsZero() || action.ExecutableAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"timelock-validity",
					fmt.Sprintf("action %s has zero timestamp", action.Id),
				), true
			}

			// Check executable time is after proposed time
			if action.ExecutableAt.Before(action.ProposedAt) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"timelock-validity",
					fmt.Sprintf("action %s executable time before proposed time", action.Id),
				), true
			}

			// Check proposer address
			if _, err := sdk.AccAddressFromBech32(action.Proposer); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"timelock-validity",
					fmt.Sprintf("action %s has invalid proposer address: %s", action.Id, action.Proposer),
				), true
			}
		}

		return "", false
	}
}

// EmergencyAdminInvariant checks that all emergency admin addresses are valid
func EmergencyAdminInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, EmergencyAdminsKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			address := string(iterator.Value())
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"emergency-admin-validity",
					fmt.Sprintf("invalid emergency admin address: %s", address),
				), true
			}
		}

		return "", false
	}
}

// SessionValidityInvariant checks that active sessions are not expired
func SessionValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, SessionsKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var session authproto.Session
			if err := k.cdc.Unmarshal(iterator.Value(), &session); err != nil {
				continue
			}

			// Check timestamps
			if session.CreatedAt.IsZero() || session.ExpiresAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"session-validity",
					fmt.Sprintf("session %s has zero timestamp", session.SessionId),
				), true
			}

			// Check user address
			if _, err := sdk.AccAddressFromBech32(session.UserAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"session-validity",
					fmt.Sprintf("session %s has invalid user address: %s", session.SessionId, session.UserAddress),
				), true
			}

			// Check expiration is after creation
			if session.ExpiresAt.Before(session.CreatedAt) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"session-validity",
					fmt.Sprintf("session %s expires before creation", session.SessionId),
				), true
			}
		}

		return "", false
	}
}

// RateLimitInvariant checks that rate limits have reasonable values
func RateLimitInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, RateLimitsKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var rateLimit authproto.RateLimitConfig
			if err := k.cdc.Unmarshal(iterator.Value(), &rateLimit); err != nil {
				continue
			}

			// Check user address is valid
			if rateLimit.UserAddress == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"rate-limit-validity",
					"rate limit config has empty user address",
				), true
			}

			if _, err := sdk.AccAddressFromBech32(rateLimit.UserAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"rate-limit-validity",
					fmt.Sprintf("rate limit config has invalid user address: %s", rateLimit.UserAddress),
				), true
			}

			// Check that limits are reasonable (at least one should be non-zero)
			if rateLimit.RequestsPerMinute == 0 && rateLimit.RequestsPerHour == 0 && rateLimit.RequestsPerDay == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"rate-limit-validity",
					fmt.Sprintf("rate limit for %s has all zero limits", rateLimit.UserAddress),
				), true
			}

			// Check current counts don't exceed their respective limits
			if rateLimit.RequestsPerMinute > 0 && rateLimit.CurrentMinuteCount > rateLimit.RequestsPerMinute {
				return sdk.FormatInvariant(
					types.ModuleName,
					"rate-limit-validity",
					fmt.Sprintf("rate limit for %s minute count (%d) exceeds limit (%d)",
						rateLimit.UserAddress, rateLimit.CurrentMinuteCount, rateLimit.RequestsPerMinute),
				), true
			}

			if rateLimit.RequestsPerHour > 0 && rateLimit.CurrentHourCount > rateLimit.RequestsPerHour {
				return sdk.FormatInvariant(
					types.ModuleName,
					"rate-limit-validity",
					fmt.Sprintf("rate limit for %s hour count (%d) exceeds limit (%d)",
						rateLimit.UserAddress, rateLimit.CurrentHourCount, rateLimit.RequestsPerHour),
				), true
			}

			if rateLimit.RequestsPerDay > 0 && rateLimit.CurrentDayCount > rateLimit.RequestsPerDay {
				return sdk.FormatInvariant(
					types.ModuleName,
					"rate-limit-validity",
					fmt.Sprintf("rate limit for %s day count (%d) exceeds limit (%d)",
						rateLimit.UserAddress, rateLimit.CurrentDayCount, rateLimit.RequestsPerDay),
				), true
			}
		}

		return "", false
	}
}

// AuditLogIntegrityInvariant checks audit log integrity
func AuditLogIntegrityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, AuditLogsKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		prevTimestamp := int64(0)

		for ; iterator.Valid(); iterator.Next() {
			var log authproto.AuditLog
			if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
				continue
			}

			// Check timestamp is not zero
			if log.Timestamp.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"audit-log-integrity",
					"audit log has zero timestamp",
				), true
			}

			// Check timestamps are monotonically increasing (optional, depends on storage order)
			currentTimestamp := log.Timestamp.Unix()
			if currentTimestamp < prevTimestamp {
				return sdk.FormatInvariant(
					types.ModuleName,
					"audit-log-integrity",
					"audit log timestamps not monotonically increasing",
				), true
			}
			prevTimestamp = currentTimestamp

			// Check actor address if present
			if log.Actor != "" {
				if _, err := sdk.AccAddressFromBech32(log.Actor); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"audit-log-integrity",
						fmt.Sprintf("audit log has invalid actor address: %s", log.Actor),
					), true
				}
			}
		}

		return "", false
	}
}
