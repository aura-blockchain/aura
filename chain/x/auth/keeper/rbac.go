package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// CreateRole creates a new role
func (k *Keeper) CreateRole(ctx sdk.Context, creator, name string, permissions []string, description string) (*authproto.Role, error) {
	// Validate creator has permission
	if err := k.RequirePermission(ctx, creator, types.PermissionCreateRole); err != nil {
		return nil, err
	}

	// Check if role already exists
	if _, err := k.GetRoleFromStore(ctx, name); err == nil {
		k.LogAudit(ctx, creator, "create_role", name, "failed", nil, "role already exists")
		return nil, types.ErrRoleAlreadyExists
	}

	now := time.Now()
	role := &authproto.Role{
		Name:        name,
		Permissions: permissions,
		Description: description,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	// Validate role
	if err := types.ValidateRole(role); err != nil {
		k.LogAudit(ctx, creator, "create_role", name, "failed", nil, err.Error())
		return nil, fmt.Errorf("%w: %v", types.ErrInvalidRole, err)
	}

	if err := k.SetRole(ctx, role); err != nil {
		return nil, err
	}

	k.LogAudit(ctx, creator, "create_role", name, "success", map[string]string{
		"permissions": fmt.Sprintf("%v", permissions),
		"description": description,
	}, "")

	return role, nil
}

// GetRole retrieves a role by name
func (k *Keeper) GetRole(ctx sdk.Context, name string) (*authproto.Role, error) {
	return k.GetRoleFromStore(ctx, name)
}

// ListRoles returns all roles
func (k *Keeper) ListRoles(ctx sdk.Context) ([]*authproto.Role, error) {
	return k.GetAllRoles(ctx)
}

// UpdateRole updates an existing role
func (k *Keeper) UpdateRole(ctx sdk.Context, updater, name string, permissions []string, description string) (*authproto.Role, error) {
	// Validate updater has permission
	if err := k.RequirePermission(ctx, updater, types.PermissionCreateRole); err != nil {
		return nil, err
	}

	// Check if role exists
	role, err := k.GetRoleFromStore(ctx, name)
	if err != nil {
		k.LogAudit(ctx, updater, "update_role", name, "failed", nil, "role not found")
		return nil, types.ErrRoleNotFound
	}

	// Prevent modifying predefined system roles
	if name == types.RoleAdmin && updater != "system" {
		k.LogAudit(ctx, updater, "update_role", name, "failed", nil, "cannot modify system role")
		return nil, fmt.Errorf("cannot modify system role %s", name)
	}

	now := time.Now()
	role.Permissions = permissions
	role.Description = description
	role.UpdatedAt = &now

	if err := k.SetRole(ctx, role); err != nil {
		return nil, err
	}

	k.LogAudit(ctx, updater, "update_role", name, "success", map[string]string{
		"permissions": fmt.Sprintf("%v", permissions),
		"description": description,
	}, "")

	return role, nil
}

// DeleteRoleByName deletes a role
func (k *Keeper) DeleteRoleByName(ctx sdk.Context, deleter, name string) error {
	// Validate deleter has permission
	if err := k.RequirePermission(ctx, deleter, types.PermissionCreateRole); err != nil {
		return err
	}

	// Check if role exists
	if _, err := k.GetRoleFromStore(ctx, name); err != nil {
		k.LogAudit(ctx, deleter, "delete_role", name, "failed", nil, "role not found")
		return types.ErrRoleNotFound
	}

	// Prevent deleting predefined system roles
	if name == types.RoleAdmin || name == types.RoleModerator || name == types.RoleValidator || name == types.RoleUser {
		k.LogAudit(ctx, deleter, "delete_role", name, "failed", nil, "cannot delete system role")
		return fmt.Errorf("cannot delete system role %s", name)
	}

	k.DeleteRole(ctx, name)

	// Remove all assignments of this role
	allAssignments, err := k.GetAllRoleAssignments(ctx)
	if err == nil {
		for _, assignment := range allAssignments {
			if assignment.RoleName == name {
				k.DeleteRoleAssignment(ctx, assignment.Address, name)
			}
		}
	}

	k.LogAudit(ctx, deleter, "delete_role", name, "success", nil, "")

	return nil
}

// AssignRole assigns a role to an address
func (k *Keeper) AssignRole(ctx sdk.Context, assigner, address, roleName string, expiresInSeconds int64) (*authproto.RoleAssignment, error) {
	// Validate assigner has permission
	if err := k.RequirePermission(ctx, assigner, types.PermissionAssignRole); err != nil {
		return nil, err
	}

	// Verify role exists
	if _, err := k.GetRoleFromStore(ctx, roleName); err != nil {
		k.LogAudit(ctx, assigner, "assign_role", address, "failed", map[string]string{
			"role": roleName,
		}, "role not found")
		return nil, types.ErrRoleNotFound
	}

	now := time.Now()
	var expiresAt *time.Time
	if expiresInSeconds > 0 {
		expiry := now.Add(time.Duration(expiresInSeconds) * time.Second)
		expiresAt = &expiry
	}

	assignment := &authproto.RoleAssignment{
		Address:    address,
		RoleName:   roleName,
		AssignedBy: assigner,
		AssignedAt: &now,
		ExpiresAt:  expiresAt,
	}

	// Validate assignment
	if err := types.ValidateRoleAssignment(assignment); err != nil {
		k.LogAudit(ctx, assigner, "assign_role", address, "failed", map[string]string{
			"role": roleName,
		}, err.Error())
		return nil, fmt.Errorf("%w: %v", types.ErrInvalidRoleAssignment, err)
	}

	// Store the assignment (SetRoleAssignment handles updating existing assignments)
	if err := k.SetRoleAssignment(ctx, assignment); err != nil {
		return nil, err
	}

	k.LogAudit(ctx, assigner, "assign_role", address, "success", map[string]string{
		"role": roleName,
	}, "")

	return assignment, nil
}

// RevokeRole revokes a role from an address
func (k *Keeper) RevokeRole(ctx sdk.Context, revoker, address, roleName string) error {
	// Validate revoker has permission
	if err := k.RequirePermission(ctx, revoker, types.PermissionRevokeRole); err != nil {
		return err
	}

	assignments, err := k.GetRoleAssignmentsForAddress(ctx, address)
	if err != nil || len(assignments) == 0 {
		k.LogAudit(ctx, revoker, "revoke_role", address, "failed", map[string]string{
			"role": roleName,
		}, "no assignments found")
		return types.ErrRoleAssignmentNotFound
	}

	found := false
	for _, assignment := range assignments {
		if assignment.RoleName == roleName {
			found = true
			break
		}
	}

	if !found {
		k.LogAudit(ctx, revoker, "revoke_role", address, "failed", map[string]string{
			"role": roleName,
		}, "assignment not found")
		return types.ErrRoleAssignmentNotFound
	}

	if err := k.DeleteRoleAssignment(ctx, address, roleName); err != nil {
		return err
	}

	k.LogAudit(ctx, revoker, "revoke_role", address, "success", map[string]string{
		"role": roleName,
	}, "")

	return nil
}

// GetRoleAssignments returns all role assignments for an address
func (k *Keeper) GetRoleAssignments(ctx sdk.Context, address string) []*authproto.RoleAssignment {
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, address)
	if err != nil {
		return []*authproto.RoleAssignment{}
	}

	// Filter out expired assignments
	now := time.Now()
	active := make([]*authproto.RoleAssignment, 0)
	for _, assignment := range assignments {
		if assignment.ExpiresAt == nil || now.Before(*assignment.ExpiresAt) {
			active = append(active, assignment)
		}
	}

	return active
}

// GetPermissionsForAddress returns all permissions for an address
func (k *Keeper) GetPermissionsForAddress(ctx sdk.Context, address string) []string {
	permMap := make(map[string]bool)

	// Get permissions from role assignments
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, address)
	if err == nil {
		now := time.Now()
		for _, assignment := range assignments {
			// Skip expired assignments
			if assignment.ExpiresAt != nil && now.After(*assignment.ExpiresAt) {
				continue
			}

			// Get role permissions
			if role, err := k.GetRoleFromStore(ctx, assignment.RoleName); err == nil {
				for _, perm := range role.Permissions {
					permMap[perm] = true
				}
			}
		}
	}

	// Get permissions from emergency admin
	if admin, err := k.GetEmergencyAdmin(ctx, address); err == nil {
		if types.IsEmergencyAdminActive(admin) {
			for _, priv := range admin.Privileges {
				permMap[priv] = true
			}
		}
	}

	// Convert to slice
	permissions := make([]string, 0, len(permMap))
	for perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions
}

// CleanupExpiredRoleAssignments removes expired role assignments
func (k *Keeper) CleanupExpiredRoleAssignments(ctx sdk.Context) int {
	allAssignments, err := k.GetAllRoleAssignments(ctx)
	if err != nil {
		return 0
	}

	count := 0
	now := time.Now()

	// Group assignments by address
	assignmentsByAddress := make(map[string][]*authproto.RoleAssignment)
	for _, assignment := range allAssignments {
		assignmentsByAddress[assignment.Address] = append(assignmentsByAddress[assignment.Address], assignment)
	}

	// Clean up expired assignments for each address
	for address, assignments := range assignmentsByAddress {
		for _, assignment := range assignments {
			if assignment.ExpiresAt != nil && now.After(*assignment.ExpiresAt) {
				if err := k.DeleteRoleAssignment(ctx, address, assignment.RoleName); err == nil {
					count++
				}
			}
		}
	}

	return count
}
