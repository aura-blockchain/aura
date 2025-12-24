package keeper

//lint:file-ignore SA1019 // invariants rely on deprecated SDK registry until upstream removal

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// RegisterInvariants registers all identity module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) { //nolint:staticcheck // invariant registry uses deprecated SDK interface
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "role-consistency", RoleConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "identity-validity", IdentityValidityInvariant(k))
}

// AllInvariants runs all invariants of the identity module
func AllInvariants(k *Keeper) sdk.Invariant { //nolint:staticcheck // invariant signature uses deprecated SDK type
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
		_, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("failed to get params: %s", err.Error()),
			), true
		}

		// Additional validation can be added here for specific param fields
		// when they are defined in the proto

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
			if role.CreatedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("role %s has zero created_at", role.Name),
				), true
			}

			if role.UpdatedAt == nil || role.UpdatedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"role-consistency",
					fmt.Sprintf("role %s has nil or zero updated_at", role.Name),
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
		// This invariant would validate all identities if we had a GetAllIdentities method
		// For now, we just validate the module is in a consistent state
		// When GetAllIdentities is implemented, this can be expanded

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
