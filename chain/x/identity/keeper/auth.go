// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// ============================================================================
// Role Management
// ============================================================================

// SetRole stores a role in the KVStore
func (k *Keeper) SetRole(ctx sdk.Context, role *types.Role) error {
	if role == nil || role.Name == "" {
		return types.ErrInvalidRole.Wrap("role name cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(role)
	if err != nil {
		return fmt.Errorf("failed to marshal role: %w", err)
	}

	key := types.GetRoleKey(role.Name)
	return store.Set(key, bz)
}

// GetRole retrieves a role from the KVStore
func (k *Keeper) GetRole(ctx sdk.Context, name string) (*types.Role, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetRoleKey(name)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrRoleNotFound.Wrapf("role not found: %s", name)
	}

	var role types.Role
	if err := k.cdc.Unmarshal(bz, &role); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role: %w", err)
	}
	return &role, nil
}

// GetAllRoles retrieves all roles from the KVStore
func (k *Keeper) GetAllRoles(ctx sdk.Context) ([]*types.Role, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.RolePrefix, storetypes.PrefixEndBytes(types.RolePrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	roles := make([]*types.Role, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var role types.Role
		if err := k.cdc.Unmarshal(iterator.Value(), &role); err != nil {
			return nil, fmt.Errorf("failed to unmarshal role: %w", err)
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

// DeleteRole removes a role from the KVStore
func (k *Keeper) DeleteRole(ctx sdk.Context, name string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetRoleKey(name)
	return store.Delete(key)
}

// CreateRole creates a new role
func (k *Keeper) CreateRole(ctx sdk.Context, creator, name string, permissions []string, description string) (*types.Role, error) {
	// Check if role already exists
	_, err := k.GetRole(ctx, name)
	if err == nil {
		return nil, types.ErrRoleAlreadyExists.Wrapf("role already exists: %s", name)
	}

	// Check creator has permission
	if err := k.RequirePermission(ctx, creator, types.PermissionCreateRole); err != nil {
		return nil, err
	}

	// Validate permissions count
	params, _ := k.GetParams(ctx)
	if err == nil && uint32(len(permissions)) > params.Auth.MaxRolesPerAccount {
		return nil, types.ErrInvalidRole.Wrap("exceeds maximum permissions per role")
	}

	now := ctx.BlockTime()
	role := &types.Role{
		Name:         name,
		Permissions:  permissions,
		Description:  description,
		CreatedAt:    now,
		CreatedBy:    creator,
		IsSystemRole: false,
		UpdatedAt:    &now,
	}

	if err := k.SetRole(ctx, role); err != nil {
		return nil, err
	}

	// Log audit trail
	k.LogAudit(ctx, creator, "create_role", name, "success", map[string]string{
		"permissions": fmt.Sprintf("%v", permissions),
	}, "")

	return role, nil
}

// UpdateRole updates an existing role
func (k *Keeper) UpdateRole(ctx sdk.Context, updater, name string, permissions []string, description string) (types.Role, error) {
	// Check updater has permission
	if err := k.RequirePermission(ctx, updater, types.PermissionCreateRole); err != nil {
		return types.Role{}, err
	}

	// Get existing role
	role, err := k.GetRole(ctx, name)
	if err != nil {
		return types.Role{}, err
	}

	// Update fields
	role.Permissions = permissions
	role.Description = description
	updatedAt := ctx.BlockTime()
	role.UpdatedAt = &updatedAt

	if err := k.SetRole(ctx, role); err != nil {
		return types.Role{}, err
	}

	// Log audit trail
	k.LogAudit(ctx, updater, "update_role", name, "success", map[string]string{
		"permissions": fmt.Sprintf("%v", permissions),
	}, "")

	return *role, nil
}

// ============================================================================
// Role Assignment Management
// ============================================================================

// SetRoleAssignment stores a role assignment in the KVStore
func (k *Keeper) SetRoleAssignment(ctx sdk.Context, assignment *types.RoleAssignment) error {
	if assignment.Address == "" {
		return types.ErrInvalidRoleAssignment.Wrap("address cannot be empty")
	}
	if assignment.RoleName == "" {
		return types.ErrInvalidRoleAssignment.Wrap("role name cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)

	// Get existing assignments for this address
	assignments, _ := k.GetRoleAssignments(ctx, assignment.Address)

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

	// Marshal the assignments list
	bz, err := json.Marshal(types.RoleAssignmentList{Assignments: assignments})
	if err != nil {
		return fmt.Errorf("failed to marshal role assignments: %w", err)
	}

	key := types.GetRoleAssignmentKey(assignment.Address)
	return store.Set(key, bz)
}

// GetRoleAssignments retrieves all role assignments for an address
func (k *Keeper) GetRoleAssignments(ctx sdk.Context, address string) ([]*types.RoleAssignment, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetRoleAssignmentKey(address)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return []*types.RoleAssignment{}, nil
	}

	var assignmentList types.RoleAssignmentList
	if err := json.Unmarshal(bz, &assignmentList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role assignments: %w", err)
	}

	// Filter out expired assignments
	now := ctx.BlockTime()
	filtered := make([]*types.RoleAssignment, 0, 64)
	for _, assignment := range assignmentList.Assignments {
		if assignment.ExpiresAt != nil && now.After(*assignment.ExpiresAt) {
			continue
		}
		filtered = append(filtered, assignment)
	}

	return filtered, nil
}

// GetAllRoleAssignments retrieves all role assignments
func (k *Keeper) GetAllRoleAssignments(ctx sdk.Context) ([]*types.RoleAssignment, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.RoleAssignmentPrefix, storetypes.PrefixEndBytes(types.RoleAssignmentPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	allAssignments := make([]*types.RoleAssignment, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var assignmentList types.RoleAssignmentList
		if err := json.Unmarshal(iterator.Value(), &assignmentList); err != nil {
			return nil, fmt.Errorf("failed to unmarshal role assignments: %w", err)
		}
		allAssignments = append(allAssignments, assignmentList.Assignments...)
	}
	return allAssignments, nil
}

// DeleteRoleAssignment removes a specific role assignment
func (k *Keeper) DeleteRoleAssignment(ctx sdk.Context, address, roleName string) error {
	assignments, err := k.GetRoleAssignments(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	filtered := make([]*types.RoleAssignment, 0, 64)
	for _, assignment := range assignments {
		if assignment.RoleName != roleName {
			filtered = append(filtered, assignment)
		}
	}

	store := k.storeService.OpenKVStore(ctx)
	key := types.GetRoleAssignmentKey(address)

	if len(filtered) == 0 {
		return store.Delete(key)
	}

	bz, err := json.Marshal(types.RoleAssignmentList{Assignments: filtered})
	if err != nil {
		return fmt.Errorf("failed to marshal role assignments: %w", err)
	}
	return store.Set(key, bz)
}

// AssignRole assigns a role to an address
func (k *Keeper) AssignRole(ctx sdk.Context, assigner, address, roleName string, expirySeconds uint64) (types.RoleAssignment, error) {
	// Check assigner has permission
	if err := k.RequirePermission(ctx, assigner, types.PermissionAssignRole); err != nil {
		return types.RoleAssignment{}, err
	}

	// Verify role exists
	if _, err := k.GetRole(ctx, roleName); err != nil {
		return types.RoleAssignment{}, err
	}

	// Check max roles per address
	params, err := k.GetParams(ctx)
	assignments, _ := k.GetRoleAssignments(ctx, address)
	if err == nil && uint32(len(assignments)) >= params.Auth.MaxRolesPerAccount {
		return types.RoleAssignment{}, types.ErrInvalidRoleAssignment.Wrap("exceeds maximum roles per address")
	}

	now := ctx.BlockTime()
	var expiresAt *time.Time
	if expirySeconds > 0 {
		expiry := now.Add(time.Duration(expirySeconds) * time.Second)
		expiresAt = &expiry
	}

	assignment := &types.RoleAssignment{
		Address:    address,
		RoleName:   roleName,
		AssignedBy: assigner,
		AssignedAt: now,
		ExpiresAt:  expiresAt,
	}

	if err := k.SetRoleAssignment(ctx, assignment); err != nil {
		return types.RoleAssignment{}, err
	}

	// Log audit trail
	k.LogAudit(ctx, assigner, "assign_role", address, "success", map[string]string{
		"role": roleName,
	}, "")

	return *assignment, nil
}

// RevokeRole revokes a role from an address
func (k *Keeper) RevokeRole(ctx sdk.Context, revoker, address, roleName string) error {
	// Check revoker has permission
	if err := k.RequirePermission(ctx, revoker, types.PermissionRevokeRole); err != nil {
		return fmt.Errorf("error in RevokeRole: %w", err)
	}

	if err := k.DeleteRoleAssignment(ctx, address, roleName); err != nil {
		return fmt.Errorf("error in RevokeRole: %w", err)
	}

	// Log audit trail
	k.LogAudit(ctx, revoker, "revoke_role", address, "success", map[string]string{
		"role": roleName,
	}, "")

	return nil
}

// ============================================================================
// Permission Checking
// ============================================================================

// HasPermission checks if an address has a specific permission
func (k *Keeper) HasPermission(ctx sdk.Context, address, permission string) bool {
	// Get role assignments for address
	assignments, err := k.GetRoleAssignments(ctx, address)
	if err != nil {
		return false
	}

	// Check each assigned role for the permission
	for _, assignment := range assignments {
		role, err := k.GetRole(ctx, assignment.RoleName)
		if err != nil {
			continue
		}

		// Check if role has admin permission (grants all)
		for _, perm := range role.Permissions {
			if perm == types.PermissionAdmin || perm == permission {
				return true
			}
		}
	}

	// Check emergency admin privileges
	admin, err := k.GetEmergencyAdmin(ctx, address)
	if err == nil && admin.IsActive {
		now := ctx.BlockTime()
		if admin.ExpiresAt != nil && !now.After(*admin.ExpiresAt) {
			for _, priv := range admin.Privileges {
				if priv == types.PermissionAdmin || priv == permission {
					return true
				}
			}
		}
	}

	return false
}

// RequirePermission checks permission and returns error if not authorized
func (k *Keeper) RequirePermission(ctx sdk.Context, address, permission string) error {
	if !k.HasPermission(ctx, address, permission) {
		k.LogAudit(ctx, address, "permission_check", permission, "denied", nil, "insufficient permissions")
		return types.ErrInsufficientPermissions.Wrapf("%s does not have permission %s", address, permission)
	}
	return nil
}

// ============================================================================
// Audit Logging
// ============================================================================

// LogAudit creates an audit log entry
func (k *Keeper) LogAudit(ctx sdk.Context, actor, action, target, result string, metadata map[string]string, errorDetail string) {
	params, err := k.GetParams(ctx)
	if err != nil || !params.Auth.EnableAuditLogging {
		return
	}

	store := k.storeService.OpenKVStore(ctx)

	// Get next audit log ID
	counterBz, _ := store.Get(types.AuditLogCounterPrefix)
	var logID uint64 = 1
	if counterBz != nil {
		logID = sdk.BigEndianToUint64(counterBz)
	}

	// Convert result string to AuditResult enum
	auditResult := types.AuditResultUnspecified
	switch result {
	case "success":
		auditResult = types.AuditResultSuccess
	case "failure":
		auditResult = types.AuditResultFailure
	case "denied":
		auditResult = types.AuditResultDenied
	}

	timestamp := ctx.BlockTime()
	auditLog := &types.AuditLog{
		Id:        fmt.Sprintf("%d", logID),
		Actor:     actor,
		Action:    action,
		Target:    target,
		Timestamp: timestamp,
		Result:    auditResult,
		Details:   errorDetail,
	}

	if err := k.SetAuditLog(ctx, auditLog); err != nil {
		k.logger.Error("failed to set audit log", "error", err)
		return
	}

	// Increment counter
	if err := store.Set(types.AuditLogCounterPrefix, sdk.Uint64ToBigEndian(logID+1)); err != nil {
		k.logger.Error("failed to update audit log counter", "error", err)
	}
}

// SetAuditLog stores an audit log
func (k *Keeper) SetAuditLog(ctx sdk.Context, log *types.AuditLog) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	// Convert string ID to uint64 for key generation
	var logID uint64
	if _, err := fmt.Sscanf(log.Id, "%d", &logID); err != nil {
		return fmt.Errorf("invalid audit log id %s: %w", log.Id, err)
	}
	key := types.GetAuditLogKey(logID)
	return store.Set(key, bz)
}

// GetAllAuditLogs retrieves all audit logs
func (k *Keeper) GetAllAuditLogs(ctx sdk.Context) ([]*types.AuditLog, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.AuditLogPrefix, storetypes.PrefixEndBytes(types.AuditLogPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	logs := make([]*types.AuditLog, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var log types.AuditLog
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			return nil, fmt.Errorf("failed to unmarshal audit log: %w", err)
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

// cleanupOldAuditLogs removes old audit logs beyond the retention limit
func (k *Keeper) cleanupOldAuditLogs(ctx sdk.Context, maxRetained uint64) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.AuditLogPrefix, storetypes.PrefixEndBytes(types.AuditLogPrefix))
	if err != nil {
		return
	}
	defer iterator.Close()

	keys := make([][]byte, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, iterator.Key())
	}

	// If we have more than the max, delete the oldest ones
	if uint64(len(keys)) > maxRetained {
		toDelete := uint64(len(keys)) - maxRetained
		for i := uint64(0); i < toDelete; i++ {
			if err := store.Delete(keys[i]); err != nil {
				k.logger.Error("failed to delete old audit log", "error", err)
			}
		}
	}
}
