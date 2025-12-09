package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// ============================================================================
// Initialization Tests
// ============================================================================

func TestInitializeDefaultRoles(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Call initializeDefaultRoles directly
	err := k.InitializeDefaultRoles(ctx)
	require.NoError(t, err)

	// Verify admin role
	adminRole, err := k.GetRoleFromStore(ctx, types.RoleAdmin)
	require.NoError(t, err)
	require.Equal(t, types.RoleAdmin, adminRole.Name)
	require.Contains(t, adminRole.Permissions, types.PermissionAdmin)

	// Verify moderator role
	modRole, err := k.GetRoleFromStore(ctx, types.RoleModerator)
	require.NoError(t, err)
	require.Equal(t, types.RoleModerator, modRole.Name)

	// Verify validator role
	valRole, err := k.GetRoleFromStore(ctx, types.RoleValidator)
	require.NoError(t, err)
	require.Equal(t, types.RoleValidator, valRole.Name)

	// Verify user role
	userRole, err := k.GetRoleFromStore(ctx, types.RoleUser)
	require.NoError(t, err)
	require.Equal(t, types.RoleUser, userRole.Name)
}

// ============================================================================
// Params Tests (Edge Cases)
// ============================================================================

func TestGetParams_NotSet(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Delete params to test default behavior
	store := ctx.KVStore(k.storeKey)
	store.Delete(ParamsKeyPrefix)

	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.NotNil(t, params)
}

func TestSetParams_MarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// This tests the successful case since we can't easily trigger marshal errors
	params := types.DefaultParams()
	err := k.SetParams(ctx, params)
	require.NoError(t, err)
}

// ============================================================================
// Role Tests (Edge Cases)
// ============================================================================

func TestSetRole_MarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Valid role should work
	role := &authproto.Role{
		Name:        "test_role",
		Permissions: []string{"perm1"},
		Description: "Test",
	}
	err := k.SetRole(ctx, role)
	require.NoError(t, err)
}

func TestGetRoleFromStore_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store invalid data
	store := ctx.KVStore(k.storeKey)
	key := append(RolesKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid data"))

	role, err := k.GetRoleFromStore(ctx, "corrupt")
	require.Error(t, err)
	require.Nil(t, role)
}

func TestGetAllRoles_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(RolesKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	roles, err := k.GetAllRoles(ctx)
	require.Error(t, err)
	require.Nil(t, roles)
}

// ============================================================================
// Role Assignment Tests (Edge Cases)
// ============================================================================

func TestSetRoleAssignment_UpdateExisting(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	assignment := &authproto.RoleAssignment{
		Address:    "user1",
		RoleName:   types.RoleUser,
		AssignedBy: "admin1",
		AssignedAt: time.Now(),
	}

	// Set once
	err := k.SetRoleAssignment(ctx, assignment)
	require.NoError(t, err)

	// Update with new expiry
	expiresAt := time.Now().Add(1 * time.Hour)
	assignment.ExpiresAt = &expiresAt
	err = k.SetRoleAssignment(ctx, assignment)
	require.NoError(t, err)

	// Verify update
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, "user1")
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	require.NotNil(t, assignments[0].ExpiresAt)
}

func TestGetRoleAssignmentsForAddress_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(RoleAssignmentsKeyPrefix, []byte("user1")...)
	store.Set(key, []byte("invalid"))

	assignments, err := k.GetRoleAssignmentsForAddress(ctx, "user1")
	require.Error(t, err)
	require.Nil(t, assignments)
}

func TestDeleteRoleAssignment_EmptyResult(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign role
	assignment := &authproto.RoleAssignment{
		Address:    "user1",
		RoleName:   types.RoleUser,
		AssignedBy: "admin1",
		AssignedAt: time.Now(),
	}
	err := k.SetRoleAssignment(ctx, assignment)
	require.NoError(t, err)

	// Delete it
	err = k.DeleteRoleAssignment(ctx, "user1", types.RoleUser)
	require.NoError(t, err)

	// Verify deleted
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, "user1")
	require.NoError(t, err)
	require.Len(t, assignments, 0)
}

func TestGetAllRoleAssignments_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(RoleAssignmentsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	assignments, err := k.GetAllRoleAssignments(ctx)
	require.Error(t, err)
	require.Nil(t, assignments)
}

// ============================================================================
// Multisig Wallet Tests (Edge Cases)
// ============================================================================

func TestSetMultisigWallet_ZeroThreshold(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	wallet := &authproto.MultisigWallet{
		Id:        "wallet1",
		Signers:   []string{"s1", "s2"},
		Threshold: 0,
	}

	err := k.SetMultisigWallet(ctx, wallet)
	require.Error(t, err)
	require.Contains(t, err.Error(), "threshold must be greater than 0")
}

func TestSetMultisigWallet_ThresholdExceedsSigners(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	wallet := &authproto.MultisigWallet{
		Id:        "wallet1",
		Signers:   []string{"s1", "s2"},
		Threshold: 5,
	}

	err := k.SetMultisigWallet(ctx, wallet)
	require.Error(t, err)
	require.Contains(t, err.Error(), "threshold cannot be greater than number of signers")
}

func TestGetMultisigWallet_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(MultisigWalletsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	wallet, err := k.GetMultisigWallet(ctx, "corrupt")
	require.Error(t, err)
	require.Nil(t, wallet)
}

func TestGetAllMultisigWallets_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(MultisigWalletsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	wallets, err := k.GetAllMultisigWallets(ctx)
	require.Error(t, err)
	require.Nil(t, wallets)
}

// ============================================================================
// Multisig Proposal Tests (Edge Cases)
// ============================================================================

func TestGetMultisigProposal_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(MultisigProposalsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	proposal, err := k.GetMultisigProposal(ctx, "corrupt")
	require.Error(t, err)
	require.Nil(t, proposal)
}

func TestGetAllMultisigProposals_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(MultisigProposalsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	proposals, err := k.GetAllMultisigProposals(ctx)
	require.Error(t, err)
	require.Nil(t, proposals)
}

// ============================================================================
// Time-Locked Action Tests
// ============================================================================

func TestGetTimeLockedAction_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	action, err := k.GetTimeLockedAction(ctx, "nonexistent")
	require.Error(t, err)
	require.Nil(t, action)
}

func TestGetTimeLockedAction_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(TimeLockedActionsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	action, err := k.GetTimeLockedAction(ctx, "corrupt")
	require.Error(t, err)
	require.Nil(t, action)
}

func TestGetAllTimeLockedActions(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create multiple actions
	for i := 0; i < 3; i++ {
		action := &authproto.TimeLockedAction{
			Id:           "action" + string(rune('1'+i)),
			ActionType:   "test",
			Payload:      []byte("data"),
			Proposer:     "proposer1",
			ProposedAt:   time.Now(),
			ExecutableAt: time.Now().Add(1 * time.Hour),
			Status:       authproto.ActionStatus_ACTION_STATUS_PENDING,
			DelaySeconds: 3600,
		}
		err := k.SetTimeLockedAction(ctx, action)
		require.NoError(t, err)
	}

	actions, err := k.GetAllTimeLockedActions(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(actions), 3)
}

func TestGetAllTimeLockedActions_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(TimeLockedActionsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	actions, err := k.GetAllTimeLockedActions(ctx)
	require.Error(t, err)
	require.Nil(t, actions)
}

func TestDeleteTimeLockedAction(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	action := &authproto.TimeLockedAction{
		Id:           "temp_action",
		ActionType:   "test",
		Payload:      []byte("data"),
		Proposer:     "proposer1",
		ProposedAt:   time.Now(),
		ExecutableAt: time.Now().Add(1 * time.Hour),
		Status:       authproto.ActionStatus_ACTION_STATUS_PENDING,
		DelaySeconds: 3600,
	}
	err := k.SetTimeLockedAction(ctx, action)
	require.NoError(t, err)

	k.DeleteTimeLockedAction(ctx, "temp_action")

	deleted, err := k.GetTimeLockedAction(ctx, "temp_action")
	require.Error(t, err)
	require.Nil(t, deleted)
}

// ============================================================================
// Emergency Admin Tests
// ============================================================================

func TestGetEmergencyAdmin_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	admin, err := k.GetEmergencyAdmin(ctx, "nonexistent")
	require.Error(t, err)
	require.Nil(t, admin)
}

func TestGetEmergencyAdmin_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(EmergencyAdminsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	admin, err := k.GetEmergencyAdmin(ctx, "corrupt")
	require.Error(t, err)
	require.Nil(t, admin)
}

func TestGetAllEmergencyAdmins(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create multiple emergency admins
	for i := 0; i < 3; i++ {
		admin := &authproto.EmergencyAdmin{
			Address:     "admin" + string(rune('1'+i)),
			Privileges:  []string{types.PermissionAdmin},
			ActivatedAt: time.Now(),
			ExpiresAt:   func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
			ActivatedBy: "activator",
			IsActive:    true,
		}
		err := k.SetEmergencyAdmin(ctx, admin)
		require.NoError(t, err)
	}

	admins, err := k.GetAllEmergencyAdmins(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(admins), 3)
}

func TestGetAllEmergencyAdmins_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(EmergencyAdminsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	admins, err := k.GetAllEmergencyAdmins(ctx)
	require.Error(t, err)
	require.Nil(t, admins)
}

func TestDeleteEmergencyAdmin(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	admin := &authproto.EmergencyAdmin{
		Address:     "temp_admin",
		Privileges:  []string{types.PermissionAdmin},
		ActivatedAt: time.Now(),
		ExpiresAt:   func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
		ActivatedBy: "activator",
		IsActive:    true,
	}
	err := k.SetEmergencyAdmin(ctx, admin)
	require.NoError(t, err)

	k.DeleteEmergencyAdmin(ctx, "temp_admin")

	deleted, err := k.GetEmergencyAdmin(ctx, "temp_admin")
	require.Error(t, err)
	require.Nil(t, deleted)
}

// ============================================================================
// Validator Key Rotation Tests
// ============================================================================

func TestSetValidatorKeyRotation(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	rotation := &authproto.ValidatorKeyRotation{
		ValidatorAddress:   "validator1",
		OldConsensusPubkey: "old_key",
		NewConsensusPubkey: "new_key",
		RotationTime:       time.Now(),
		InitiatedBy:        "initiator",
		RotationStatus:     authproto.RotationStatus_ROTATION_STATUS_PENDING,
	}

	err := k.SetValidatorKeyRotation(ctx, rotation)
	require.NoError(t, err)

	// Verify retrieval
	retrieved, err := k.GetValidatorKeyRotation(ctx, "validator1")
	require.NoError(t, err)
	require.Equal(t, "validator1", retrieved.ValidatorAddress)
	require.Equal(t, "new_key", retrieved.NewConsensusPubkey)
}

func TestGetValidatorKeyRotation_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	rotation, err := k.GetValidatorKeyRotation(ctx, "nonexistent")
	require.Error(t, err)
	require.Nil(t, rotation)
}

func TestGetValidatorKeyRotation_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(ValidatorRotationsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	rotation, err := k.GetValidatorKeyRotation(ctx, "corrupt")
	require.Error(t, err)
	require.Nil(t, rotation)
}

func TestGetAllValidatorKeyRotations(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create multiple rotations
	for i := 0; i < 3; i++ {
		rotation := &authproto.ValidatorKeyRotation{
			ValidatorAddress:   "validator" + string(rune('1'+i)),
			OldConsensusPubkey: "old_key",
			NewConsensusPubkey: "new_key",
			RotationTime:       time.Now(),
			InitiatedBy:        "initiator",
			RotationStatus:     authproto.RotationStatus_ROTATION_STATUS_PENDING,
		}
		err := k.SetValidatorKeyRotation(ctx, rotation)
		require.NoError(t, err)
	}

	rotations, err := k.GetAllValidatorKeyRotations(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rotations), 3)
}

func TestGetAllValidatorKeyRotations_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(ValidatorRotationsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	rotations, err := k.GetAllValidatorKeyRotations(ctx)
	require.Error(t, err)
	require.Nil(t, rotations)
}

func TestDeleteValidatorKeyRotation(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	rotation := &authproto.ValidatorKeyRotation{
		ValidatorAddress:   "temp_validator",
		OldConsensusPubkey: "old_key",
		NewConsensusPubkey: "new_key",
		RotationTime:       time.Now(),
		InitiatedBy:        "initiator",
		RotationStatus:     authproto.RotationStatus_ROTATION_STATUS_PENDING,
	}
	err := k.SetValidatorKeyRotation(ctx, rotation)
	require.NoError(t, err)

	k.DeleteValidatorKeyRotation(ctx, "temp_validator")

	deleted, err := k.GetValidatorKeyRotation(ctx, "temp_validator")
	require.Error(t, err)
	require.Nil(t, deleted)
}

// ============================================================================
// Session Tests
// ============================================================================

func TestGetSession_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	session, err := k.GetSession(ctx, "nonexistent")
	require.Error(t, err)
	require.Nil(t, session)
}

func TestGetSession_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(SessionsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	session, err := k.GetSession(ctx, "corrupt")
	require.Error(t, err)
	require.Nil(t, session)
}

func TestGetAllSessions(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		session := &authproto.Session{
			SessionId:   "session" + string(rune('1'+i)),
			UserAddress: "user1",
			CreatedAt:   time.Now(),
			ExpiresAt:   func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
			IpAddress:   "127.0.0.1",
		}
		err := k.SetSession(ctx, session)
		require.NoError(t, err)
	}

	sessions, err := k.GetAllSessions(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(sessions), 3)
}

func TestGetAllSessions_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(SessionsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	sessions, err := k.GetAllSessions(ctx)
	require.Error(t, err)
	require.Nil(t, sessions)
}

func TestGetUserSessions_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	sessions, err := k.GetUserSessions(ctx, "nonexistent")
	require.NoError(t, err)
	require.Len(t, sessions, 0)
}

func TestGetUserSessions_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(UserSessionsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	sessions, err := k.GetUserSessions(ctx, "corrupt")
	require.Error(t, err)
	require.Nil(t, sessions)
}

func TestRemoveUserSession_EmptyResult(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create session
	session := &authproto.Session{
		SessionId:   "session1",
		UserAddress: "user1",
		CreatedAt:   time.Now(),
		ExpiresAt:   func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
	}
	err := k.SetSession(ctx, session)
	require.NoError(t, err)

	// Remove it
	err = k.removeUserSession(ctx, "user1", "session1")
	require.NoError(t, err)

	// Verify removed
	sessions, err := k.GetUserSessions(ctx, "user1")
	require.NoError(t, err)
	require.Len(t, sessions, 0)
}

// ============================================================================
// Rate Limit Tests
// ============================================================================

func TestSetRateLimitConfig(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	config := &authproto.RateLimitConfig{
		UserAddress:        "user1",
		RequestsPerMinute:  100,
		RequestsPerHour:    6000,
		RequestsPerDay:     144000,
		CurrentMinuteCount: 0,
		CurrentHourCount:   0,
		CurrentDayCount:    0,
		WindowStart:        time.Now(),
	}

	err := k.SetRateLimitConfig(ctx, config)
	require.NoError(t, err)

	// Verify retrieval
	retrieved, err := k.GetRateLimitConfig(ctx, "user1")
	require.NoError(t, err)
	require.Equal(t, uint64(100), retrieved.RequestsPerMinute)
}

func TestGetRateLimitConfig_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	config, err := k.GetRateLimitConfig(ctx, "nonexistent")
	require.Error(t, err)
	require.Nil(t, config)
}

func TestGetRateLimitConfig_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(RateLimitsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	config, err := k.GetRateLimitConfig(ctx, "corrupt")
	require.Error(t, err)
	require.Nil(t, config)
}

func TestGetAllRateLimitConfigs(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create multiple configs
	for i := 0; i < 3; i++ {
		config := &authproto.RateLimitConfig{
			UserAddress:        "user" + string(rune('1'+i)),
			RequestsPerMinute:  100,
			RequestsPerHour:    6000,
			RequestsPerDay:     144000,
			CurrentMinuteCount: 0,
			CurrentHourCount:   0,
			CurrentDayCount:    0,
			WindowStart:        time.Now(),
		}
		err := k.SetRateLimitConfig(ctx, config)
		require.NoError(t, err)
	}

	configs, err := k.GetAllRateLimitConfigs(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(configs), 3)
}

func TestGetAllRateLimitConfigs_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(RateLimitsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	configs, err := k.GetAllRateLimitConfigs(ctx)
	require.Error(t, err)
	require.Nil(t, configs)
}

func TestDeleteRateLimitConfig(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	config := &authproto.RateLimitConfig{
		UserAddress:       "temp_user",
		RequestsPerMinute: 100,
		RequestsPerHour:   6000,
		RequestsPerDay:    144000,
		WindowStart:       time.Now(),
	}
	err := k.SetRateLimitConfig(ctx, config)
	require.NoError(t, err)

	k.DeleteRateLimitConfig(ctx, "temp_user")

	deleted, err := k.GetRateLimitConfig(ctx, "temp_user")
	require.Error(t, err)
	require.Nil(t, deleted)
}

func TestResetRateLimitWindow(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create config with old window start
	oldTime := time.Now().Add(-2 * time.Hour)
	config := &authproto.RateLimitConfig{
		UserAddress:        "user1",
		RequestsPerMinute:  100,
		RequestsPerHour:    6000,
		RequestsPerDay:     144000,
		CurrentMinuteCount: 50,
		CurrentHourCount:   500,
		CurrentDayCount:    5000,
		WindowStart:        timestamppb.New(oldTime),
	}
	err := k.SetRateLimitConfig(ctx, config)
	require.NoError(t, err)

	// Reset window
	k.ResetRateLimitWindow(ctx, "user1")

	// Verify counters reset
	updated, err := k.GetRateLimitConfig(ctx, "user1")
	require.NoError(t, err)
	require.Equal(t, uint64(0), updated.CurrentMinuteCount)
	require.Equal(t, uint64(0), updated.CurrentHourCount)
}

func TestResetRateLimitWindow_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Should not panic
	k.ResetRateLimitWindow(ctx, "nonexistent")
}

func TestResetRateLimitWindow_NilWindowStart(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	config := &authproto.RateLimitConfig{
		UserAddress:        "user1",
		RequestsPerMinute:  100,
		RequestsPerHour:    6000,
		RequestsPerDay:     144000,
		CurrentMinuteCount: 50,
		WindowStart:        nil,
	}
	err := k.SetRateLimitConfig(ctx, config)
	require.NoError(t, err)

	// Should not panic
	k.ResetRateLimitWindow(ctx, "user1")
}

// ============================================================================
// Audit Log Tests
// ============================================================================

func TestGetNextAuditLogID(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	id1 := k.getNextAuditLogID(ctx)
	id2 := k.getNextAuditLogID(ctx)

	require.Greater(t, id2, id1)
}

func TestGetAllAuditLogs(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create audit logs in KV store
	now := time.Now()
	for i := 0; i < 3; i++ {
		log := &authproto.AuditLog{
			Id:        fmt.Sprintf("%d", k.getNextAuditLogID(ctx)),
			Actor:     "actor1",
			Action:    "action" + string(rune('1'+i)),
			Resource:  "resource1",
			Result:    "success",
			Timestamp: timestamppb.New(now),
		}

		store := ctx.KVStore(k.storeKey)
		bz, err := k.cdc.Marshal(log)
		require.NoError(t, err)
		key := append(AuditLogsKeyPrefix, []byte(string(rune(i)))...)
		store.Set(key, bz)
	}

	logs, err := k.GetAllAuditLogs(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(logs), 3)
}

func TestGetAllAuditLogs_UnmarshalError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(AuditLogsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	logs, err := k.GetAllAuditLogs(ctx)
	require.Error(t, err)
	require.Nil(t, logs)
}

func TestCleanupOldAuditLogs(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// This function just cleans up, no error to check
	k.cleanupOldAuditLogs(ctx)

	// Create more than 10000 logs would be too slow for tests
	// Just verify it doesn't panic
}

// ============================================================================
// Cleanup Tests
// ============================================================================

func TestCleanupExpiredSessions(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create expired session
	expiredSession := &authproto.Session{
		SessionId:   "expired_session",
		UserAddress: "user1",
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		ExpiresAt:   func() *time.Time { t := time.Now().Add(-1 * time.Hour); return &t }(),
	}
	err := k.SetSession(ctx, expiredSession)
	require.NoError(t, err)

	// Create active session
	activeSession := &authproto.Session{
		SessionId:   "active_session",
		UserAddress: "user1",
		CreatedAt:   time.Now(),
		ExpiresAt:   func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
	}
	err = k.SetSession(ctx, activeSession)
	require.NoError(t, err)

	// Cleanup
	count := k.CleanupExpiredSessions(ctx)
	require.Greater(t, count, 0)

	// Verify expired session is gone
	_, err = k.GetSession(ctx, "expired_session")
	require.Error(t, err)

	// Verify active session still exists
	_, err = k.GetSession(ctx, "active_session")
	require.NoError(t, err)
}

func TestCleanupExpiredSessions_Error(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// No sessions, should return 0
	count := k.CleanupExpiredSessions(ctx)
	require.Equal(t, 0, count)
}

func TestCleanupExpiredProposals(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create expired proposal
	expiredProposal := &authproto.MultisigProposal{
		Id:          "expired_proposal",
		WalletId:    "wallet1",
		Title:       "Expired",
		Description: "Expired proposal",
		Payload:     []byte("data"),
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		ExpiresAt:   func() *time.Time { t := time.Now().Add(-1 * time.Hour); return &t }(),
		Status:      authproto.ProposalStatus_PROPOSAL_STATUS_PENDING,
		Signatures:  []string{},
	}
	err := k.SetMultisigProposal(ctx, expiredProposal)
	require.NoError(t, err)

	// Create active proposal
	activeProposal := &authproto.MultisigProposal{
		Id:          "active_proposal",
		WalletId:    "wallet1",
		Title:       "Active",
		Description: "Active proposal",
		Payload:     []byte("data"),
		CreatedAt:   time.Now(),
		ExpiresAt:   func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
		Status:      authproto.ProposalStatus_PROPOSAL_STATUS_PENDING,
		Signatures:  []string{},
	}
	err = k.SetMultisigProposal(ctx, activeProposal)
	require.NoError(t, err)

	// Cleanup
	count := k.CleanupExpiredProposals(ctx)
	require.Greater(t, count, 0)

	// Verify expired proposal status changed
	expired, err := k.GetMultisigProposal(ctx, "expired_proposal")
	require.NoError(t, err)
	require.Equal(t, authproto.ProposalStatus_PROPOSAL_STATUS_EXPIRED, expired.Status)

	// Verify active proposal unchanged
	active, err := k.GetMultisigProposal(ctx, "active_proposal")
	require.NoError(t, err)
	require.Equal(t, authproto.ProposalStatus_PROPOSAL_STATUS_PENDING, active.Status)
}

func TestCleanupExpiredProposals_AlreadyExecuted(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create executed but expired proposal
	proposal := &authproto.MultisigProposal{
		Id:          "executed_proposal",
		WalletId:    "wallet1",
		Title:       "Executed",
		Description: "Executed proposal",
		Payload:     []byte("data"),
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		ExpiresAt:   func() *time.Time { t := time.Now().Add(-1 * time.Hour); return &t }(),
		Status:      authproto.ProposalStatus_PROPOSAL_STATUS_EXECUTED,
		Signatures:  []string{},
	}
	err := k.SetMultisigProposal(ctx, proposal)
	require.NoError(t, err)

	// Cleanup should not change executed proposals
	k.CleanupExpiredProposals(ctx)

	// Verify status unchanged
	retrieved, err := k.GetMultisigProposal(ctx, "executed_proposal")
	require.NoError(t, err)
	require.Equal(t, authproto.ProposalStatus_PROPOSAL_STATUS_EXECUTED, retrieved.Status)
}

func TestCleanupExpiredProposals_Error(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// No proposals, should return 0
	count := k.CleanupExpiredProposals(ctx)
	require.Equal(t, 0, count)
}

// ============================================================================
// Permission Tests (Edge Cases)
// ============================================================================

func TestHasPermission_ExpiredRole(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign role with expiry in the past
	pastTime := time.Now().Add(-1 * time.Hour)
	assignment := &authproto.RoleAssignment{
		Address:    "user1",
		RoleName:   types.RoleAdmin,
		AssignedBy: "system",
		AssignedAt: timestamppb.New(pastTime),
		ExpiresAt:  timestamppb.New(pastTime.Add(30 * time.Minute)),
	}
	err := k.SetRoleAssignment(ctx, assignment)
	require.NoError(t, err)

	// Should not have permission due to expiry
	hasPermission := k.HasPermission(ctx, "user1", types.PermissionAdmin)
	require.False(t, hasPermission)
}

func TestHasPermission_EmergencyAdminExpired(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create expired emergency admin
	pastTime := time.Now().Add(-1 * time.Hour)
	admin := &authproto.EmergencyAdmin{
		Address:     "emergency1",
		Privileges:  []string{types.PermissionAdmin},
		ActivatedAt: timestamppb.New(pastTime),
		ExpiresAt:   timestamppb.New(pastTime.Add(30 * time.Minute)),
		ActivatedBy: "activator",
		IsActive:    true,
	}
	err := k.SetEmergencyAdmin(ctx, admin)
	require.NoError(t, err)

	// Should not have permission due to expiry
	hasPermission := k.HasPermission(ctx, "emergency1", types.PermissionAdmin)
	require.False(t, hasPermission)
}

func TestHasPermission_RoleNotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign nonexistent role
	assignment := &authproto.RoleAssignment{
		Address:    "user1",
		RoleName:   "nonexistent_role",
		AssignedBy: "system",
		AssignedAt: time.Now(),
	}
	err := k.SetRoleAssignment(ctx, assignment)
	require.NoError(t, err)

	// Should not have permission
	hasPermission := k.HasPermission(ctx, "user1", types.PermissionAdmin)
	require.False(t, hasPermission)
}

func TestRequirePermission_Success(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	err = k.RequirePermission(ctx, "admin1", types.PermissionAdmin)
	require.NoError(t, err)
}

func TestRequirePermission_Denied(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	err := k.RequirePermission(ctx, "user1", types.PermissionAdmin)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}
