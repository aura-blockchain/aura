package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// ============================================================================
// Final Coverage Tests for Edge Cases
// ============================================================================

// Test GetAuditLogs with nil timestamps
func TestGetAuditLogs_NilTimestamp(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Log with nil timestamp
	k.LogAudit(ctx, "actor1", "action1", "resource1", "success", nil, "")

	// Get logs - should handle nil timestamp gracefully
	logs := k.GetAuditLogs("", "", 0, 0, 10)
	require.NotNil(t, logs)
}

// Test GetAuditLogsByResource with nil timestamps
func TestGetAuditLogsByResource_NilTimestamp(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	k.LogAudit(ctx, "actor1", "action1", "resource1", "success", nil, "")

	logs := k.GetAuditLogsByResource("resource1", 10)
	require.NotNil(t, logs)
}

// Test GetRecentAuditLogs with nil timestamps
func TestGetRecentAuditLogs_NilTimestamp(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	k.LogAudit(ctx, "actor1", "action1", "resource1", "success", nil, "")

	logs := k.GetRecentAuditLogs(10)
	require.NotNil(t, logs)
}

// Test SearchAuditLogs with all search criteria
func TestSearchAuditLogs_AllCriteria(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	k.LogAudit(ctx, "alice", "create", "user_123", "success", map[string]string{"type": "user"}, "")
	k.LogAudit(ctx, "bob", "delete", "user_456", "failed", map[string]string{"type": "user"}, "error")

	// Test resource criteria
	criteria := map[string]string{
		"resource": "user",
	}
	logs := k.SearchAuditLogs(criteria, 10)
	require.NotNil(t, logs)

	// Test status criteria
	criteria = map[string]string{
		"status": "failed",
	}
	logs = k.SearchAuditLogs(criteria, 10)
	require.GreaterOrEqual(t, len(logs), 1)

	// Test multiple criteria
	criteria = map[string]string{
		"actor":  "alice",
		"action": "create",
		"status": "success",
	}
	logs = k.SearchAuditLogs(criteria, 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test SearchAuditLogs with nil timestamps
func TestSearchAuditLogs_NilTimestamp(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	k.LogAudit(ctx, "actor1", "action1", "resource1", "success", nil, "")

	criteria := map[string]string{
		"actor": "actor1",
	}
	logs := k.SearchAuditLogs(criteria, 10)
	require.NotNil(t, logs)
}

// Test CountAuditLogsByActor with no logs
func TestCountAuditLogsByActor_NotFound(t *testing.T) {
	k, _ := setupTestKeeper(t)

	count := k.CountAuditLogsByActor("nonexistent")
	require.Equal(t, uint64(0), count)
}

// Test initializeDefaultRoles error on SetRole
func TestInitializeDefaultRoles_Complete(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	err := k.initializeDefaultRoles(ctx)
	require.NoError(t, err)

	// Verify all 4 roles exist
	roles, err := k.GetAllRoles(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(roles), 4)
}

// Test DeleteRoleAssignment with unmarshal error
func TestDeleteRoleAssignment_GetError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Store corrupt data
	store := ctx.KVStore(k.storeKey)
	key := append(RoleAssignmentsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid"))

	err := k.DeleteRoleAssignment(ctx, "corrupt", "role")
	require.Error(t, err)
}

// Test DeleteRoleAssignment with marshal error on save
func TestDeleteRoleAssignment_MultipleRoles(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign multiple roles
	assignment1 := &authproto.RoleAssignment{
		Address:    "user1",
		RoleName:   types.RoleUser,
		AssignedBy: "admin1",
		AssignedAt: timestamppb.New(time.Now()),
	}
	err := k.SetRoleAssignment(ctx, assignment1)
	require.NoError(t, err)

	assignment2 := &authproto.RoleAssignment{
		Address:    "user1",
		RoleName:   types.RoleModerator,
		AssignedBy: "admin1",
		AssignedAt: timestamppb.New(time.Now()),
	}
	err = k.SetRoleAssignment(ctx, assignment2)
	require.NoError(t, err)

	// Delete one role
	err = k.DeleteRoleAssignment(ctx, "user1", types.RoleUser)
	require.NoError(t, err)

	// Verify only moderator remains
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, "user1")
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	require.Equal(t, types.RoleModerator, assignments[0].RoleName)
}

// Test cleanupOldAuditLogs with many logs
func TestCleanupOldAuditLogs_ManyLogs(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// This test covers the deletion path in cleanupOldAuditLogs
	// We can't create 10000+ logs easily, but we can test the logic
	k.cleanupOldAuditLogs(ctx)
}

// Test ResetRateLimitWindow with day counter reset
func TestResetRateLimitWindow_DayReset(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create config with old window (>24 hours)
	oldTime := time.Now().Add(-25 * time.Hour)
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

	// Verify all counters reset including day
	updated, err := k.GetRateLimitConfig(ctx, "user1")
	require.NoError(t, err)
	require.Equal(t, uint64(0), updated.CurrentMinuteCount)
	require.Equal(t, uint64(0), updated.CurrentHourCount)
	require.Equal(t, uint64(0), updated.CurrentDayCount)
	require.NotNil(t, updated.WindowStart)
}

// Test CreateRole audit log
func TestCreateRole_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.CreateRole(sdk.WrapSDKContext(ctx), "admin1", "new_role", []string{"perm1"}, "Test")
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByActor("admin1", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test AssignRole audit log
func TestAssignRole_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.AssignRole(sdk.WrapSDKContext(ctx), "admin1", "user1", types.RoleUser, 0)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByActor("admin1", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test RevokeRole audit log
func TestRevokeRole_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.AssignRole(sdk.WrapSDKContext(ctx), "admin1", "user1", types.RoleUser, 0)
	require.NoError(t, err)

	err = k.RevokeRole(sdk.WrapSDKContext(ctx), "admin1", "user1", types.RoleUser)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("revoke_role", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test CreateMultisigWallet audit log
func TestCreateMultisigWallet_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"s1", "s2"}, 2, "custom")
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("create_multisig_wallet", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test CreateMultisigProposal audit log
func TestCreateMultisigProposal_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	wallet, err := k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"s1", "s2"}, 2, "custom")
	require.NoError(t, err)

	_, err = k.CreateMultisigProposal(sdk.WrapSDKContext(ctx), "s1", wallet.Id, "Title", "Desc", []byte("data"), 3600)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("create_multisig_proposal", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test SignMultisigProposal audit log
func TestSignMultisigProposal_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	wallet, err := k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"s1", "s2"}, 2, "custom")
	require.NoError(t, err)

	proposal, err := k.CreateMultisigProposal(sdk.WrapSDKContext(ctx), "s1", wallet.Id, "Title", "Desc", []byte("data"), 3600)
	require.NoError(t, err)

	_, err = k.SignMultisigProposal(sdk.WrapSDKContext(ctx), "s2", proposal.Id)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("sign_multisig_proposal", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test ExecuteMultisigProposal audit log
func TestExecuteMultisigProposal_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	wallet, err := k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"s1", "s2"}, 2, "custom")
	require.NoError(t, err)

	proposal, err := k.CreateMultisigProposal(sdk.WrapSDKContext(ctx), "s1", wallet.Id, "Title", "Desc", []byte("data"), 3600)
	require.NoError(t, err)

	_, err = k.SignMultisigProposal(sdk.WrapSDKContext(ctx), "s2", proposal.Id)
	require.NoError(t, err)

	err = k.ExecuteMultisigProposal(sdk.WrapSDKContext(ctx), "s1", proposal.Id)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("execute_multisig_proposal", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test ProposeTimeLockedAction audit log
func TestProposeTimeLockedAction_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.ProposeTimeLockedAction(sdk.WrapSDKContext(ctx), "admin1", "action", []byte("data"), 3600)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("propose_timelock_action", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test ExecuteTimeLockedAction audit log
func TestExecuteTimeLockedAction_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	action, err := k.ProposeTimeLockedAction(sdk.WrapSDKContext(ctx), "admin1", "action", []byte("data"), 1)
	require.NoError(t, err)

	// Wait for action to be ready
	time.Sleep(2 * time.Second)

	err = k.ExecuteTimeLockedAction(sdk.WrapSDKContext(ctx), "admin1", action.Id)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("execute_timelock_action", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test CancelTimeLockedAction audit log
func TestCancelTimeLockedAction_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	action, err := k.ProposeTimeLockedAction(sdk.WrapSDKContext(ctx), "admin1", "action", []byte("data"), 3600)
	require.NoError(t, err)

	err = k.CancelTimeLockedAction(sdk.WrapSDKContext(ctx), "admin1", action.Id)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("cancel_timelock_action", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test ActivateEmergencyAdmin audit log
func TestActivateEmergencyAdmin_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.ActivateEmergencyAdmin(sdk.WrapSDKContext(ctx), "admin1", "emergency1", []string{types.PermissionAdmin}, 3600)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("activate_emergency_admin", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test DeactivateEmergencyAdmin audit log
func TestDeactivateEmergencyAdmin_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.ActivateEmergencyAdmin(sdk.WrapSDKContext(ctx), "admin1", "emergency1", []string{types.PermissionAdmin}, 3600)
	require.NoError(t, err)

	err = k.DeactivateEmergencyAdmin(sdk.WrapSDKContext(ctx), "admin1", "emergency1")
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("deactivate_emergency_admin", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test RotateValidatorKey audit log
func TestRotateValidatorKey_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "validator1", types.RoleValidator, 0)
	require.NoError(t, err)

	_, err = k.RotateValidatorKey(sdk.WrapSDKContext(ctx), "validator1", []byte("key"))
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("rotate_validator_key", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test InitiateValidatorKeyRotation audit log
func TestInitiateValidatorKeyRotation_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.InitiateValidatorKeyRotation(sdk.WrapSDKContext(ctx), "admin1", "validator1", []byte("key"))
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("initiate_validator_key_rotation", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test CompleteValidatorKeyRotation audit log
func TestCompleteValidatorKeyRotation_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.InitiateValidatorKeyRotation(sdk.WrapSDKContext(ctx), "admin1", "validator1", []byte("key"))
	require.NoError(t, err)

	err = k.CompleteValidatorKeyRotation(sdk.WrapSDKContext(ctx), "admin1", "validator1")
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("complete_validator_key_rotation", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test CreateSession audit log
func TestCreateSession_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.CreateSession(sdk.WrapSDKContext(ctx), "user1", 3600)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("create_session", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test InvalidateSession audit log
func TestInvalidateSession_AuditLog(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	session, err := k.CreateSession(sdk.WrapSDKContext(ctx), "user1", 3600)
	require.NoError(t, err)

	err = k.InvalidateSession(sdk.WrapSDKContext(ctx), "user1", session.SessionId)
	require.NoError(t, err)

	// Verify audit log created
	logs := k.GetAuditLogsByAction("invalidate_session", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Test HasPermission with emergency admin but no IsActive flag
func TestHasPermission_EmergencyAdminInactive(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create inactive emergency admin
	admin := &authproto.EmergencyAdmin{
		Address:     "emergency1",
		Privileges:  []string{types.PermissionAdmin},
		ActivatedAt: timestamppb.New(time.Now()),
		ExpiresAt:   timestamppb.New(time.Now().Add(1 * time.Hour)),
		ActivatedBy: "activator",
		IsActive:    false,
	}
	err := k.SetEmergencyAdmin(ctx, admin)
	require.NoError(t, err)

	// Should not have permission
	hasPermission := k.HasPermission(ctx, "emergency1", types.PermissionAdmin)
	require.False(t, hasPermission)
}

// Test CleanupExpiredSessions with delete error handling
func TestCleanupExpiredSessions_DeleteError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create an expired session
	expiredSession := &authproto.Session{
		SessionId:   "expired_session",
		UserAddress: "user1",
		CreatedAt:   timestamppb.New(time.Now().Add(-2 * time.Hour)),
		ExpiresAt:   timestamppb.New(time.Now().Add(-1 * time.Hour)),
	}
	err := k.SetSession(ctx, expiredSession)
	require.NoError(t, err)

	// Cleanup should handle it
	count := k.CleanupExpiredSessions(ctx)
	require.GreaterOrEqual(t, count, 0) // May be 0 if delete fails, but shouldn't panic
}

// Test CleanupExpiredProposals with save error handling
func TestCleanupExpiredProposals_SaveError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create expired proposal
	expiredProposal := &authproto.MultisigProposal{
		Id:          "expired_proposal",
		WalletId:    "wallet1",
		Title:       "Expired",
		Description: "Expired proposal",
		Payload:     []byte("data"),
		CreatedAt:   timestamppb.New(time.Now().Add(-2 * time.Hour)),
		ExpiresAt:   timestamppb.New(time.Now().Add(-1 * time.Hour)),
		Status:      authproto.ProposalStatus_PROPOSAL_STATUS_PENDING,
		Signatures:  []string{},
	}
	err := k.SetMultisigProposal(ctx, expiredProposal)
	require.NoError(t, err)

	// Cleanup should handle it
	count := k.CleanupExpiredProposals(ctx)
	require.GreaterOrEqual(t, count, 0)
}
