package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// RegisterInvariants registers all identity module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "role-consistency", RoleConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "identity-validity", IdentityValidityInvariant(k))
}

// AllInvariants runs all invariants of the identity module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			RoleConsistencyInvariant(k),
			IdentityValidityInvariant(k),
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

// ParamsInvariant checks that module parameters are valid
func ParamsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("failed to get params: %s", err.Error()),
			), true
		}

		if params == nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				"params are nil",
			), true
		}

		// Validate identity cost is non-negative
		if params.IdentityCreationCost != nil && params.IdentityCreationCost.IsNegative() {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("identity creation cost cannot be negative: %s", params.IdentityCreationCost.String()),
			), true
		}

		// Validate verification expiry is positive if set
		if params.VerificationExpirySeconds > 0 && params.VerificationExpirySeconds < 3600 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("verification expiry too short (min 1 hour): %d seconds", params.VerificationExpirySeconds),
			), true
		}

		// Validate max identities per address
		if params.MaxIdentitiesPerAddress == 0 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				"max identities per address cannot be zero",
			), true
		}

		return "", false
	}
}

// RoleConsistencyInvariant checks that all roles are valid and consistent
func RoleConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Get all roles
		roles, err := k.GetAllRoles(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"role-consistency",
				fmt.Sprintf("failed to get roles: %s", err.Error()),
			), true
		}

		// Track role names to detect duplicates
		roleNames := make(map[string]bool)

		for _, role := range roles {
			// Check role name is not empty
			if role.Name == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					"role has empty name",
				), true
			}

			// Check for duplicate role names
			if roleNames[role.Name] {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("duplicate role name: %s", role.Name),
				), true
			}
			roleNames[role.Name] = true

			// Validate permissions are not empty for non-basic roles
			if role.Name != types.RoleUser && len(role.Permissions) == 0 && !role.IsSystemRole {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("role %s has no permissions", role.Name),
				), true
			}

			// Validate timestamps are set
			if role.CreatedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("role %s has nil created_at", role.Name),
				), true
			}

			if role.UpdatedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("role %s has nil updated_at", role.Name),
				), true
			}

			// System roles cannot be deleted/modified
			if role.IsSystemRole {
				// Verify system roles have the expected permissions
				switch role.Name {
				case types.RoleAdmin:
					if !containsPermission(role.Permissions, types.PermissionAdmin) {
						return sdk.FormatInvariant(
							types.ModuleName,
							"role-consistency",
							fmt.Sprintf("system role %s missing admin permission", role.Name),
						), true
					}
				case types.RoleValidator:
					if !containsPermission(role.Permissions, types.PermissionRotateValidatorKey) {
						return sdk.FormatInvariant(
							types.ModuleName,
							"role-consistency",
							fmt.Sprintf("system role %s missing validator permission", role.Name),
						), true
					}
				}
			}
		}

		// Verify required system roles exist
		requiredSystemRoles := []string{types.RoleAdmin, types.RoleUser, types.RoleValidator, types.RoleModerator}
		for _, requiredRole := range requiredSystemRoles {
			if !roleNames[requiredRole] {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("required system role missing: %s", requiredRole),
				), true
			}
		}

		return "", false
	}
}

// IdentityValidityInvariant checks that all identities are valid
func IdentityValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Get all identities
		identities, err := k.GetAllIdentities(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"identity-validity",
				fmt.Sprintf("failed to get identities: %s", err.Error()),
			), true
		}

		// Track identity IDs to detect duplicates
		identityIDs := make(map[string]bool)

		for _, identity := range identities {
			// Check identity ID is not empty
			if identity.Id == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"identity-validity",
					"identity has empty ID",
				), true
			}

			// Check for duplicate identity IDs
			if identityIDs[identity.Id] {
				return sdk.FormatInvariant(
					types.ModuleName,
					"identity-validity",
					fmt.Sprintf("duplicate identity ID: %s", identity.Id),
				), true
			}
			identityIDs[identity.Id] = true

			// Validate owner address
			if identity.Owner == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"identity-validity",
					fmt.Sprintf("identity %s has empty owner", identity.Id),
				), true
			}

			// Validate address format
			if _, err := sdk.AccAddressFromBech32(identity.Owner); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"identity-validity",
					fmt.Sprintf("identity %s has invalid owner address: %s", identity.Id, identity.Owner),
				), true
			}

			// Validate created timestamp
			if identity.CreatedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"identity-validity",
					fmt.Sprintf("identity %s has nil created_at", identity.Id),
				), true
			}

			// Validate updated timestamp
			if identity.UpdatedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"identity-validity",
					fmt.Sprintf("identity %s has nil updated_at", identity.Id),
				), true
			}

			// Validate status is valid
			if identity.Status == types.IdentityStatus_IDENTITY_STATUS_UNSPECIFIED {
				return sdk.FormatInvariant(
					types.ModuleName,
					"identity-validity",
					fmt.Sprintf("identity %s has unspecified status", identity.Id),
				), true
			}

			// If identity is verified, must have verified_at timestamp
			if identity.Verified && identity.VerifiedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"identity-validity",
					fmt.Sprintf("identity %s is verified but has nil verified_at", identity.Id),
				), true
			}

			// If identity is revoked, must have revoked_at timestamp
			if identity.Status == types.IdentityStatus_REVOKED && identity.RevokedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"identity-validity",
					fmt.Sprintf("identity %s is revoked but has nil revoked_at", identity.Id),
				), true
			}
		}

		return "", false
	}
}

// Helper function to check if a permission exists in a slice
func containsPermission(permissions []string, target string) bool {
	for _, perm := range permissions {
		if perm == target {
			return true
		}
	}
	return false
}
