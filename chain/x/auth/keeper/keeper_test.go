package keeper

import (
	"context"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

func setupKeeper() *Keeper {
	params := types.DefaultParams()
	return NewKeeper(params)
}

// RBAC Tests

func TestCreateRole(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Assign admin role to creator first
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	role, err := k.CreateRole(ctx, "admin1", "custom_role", []string{"permission1", "permission2"}, "Custom role")
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	if role.Name != "custom_role" {
		t.Errorf("Expected role name 'custom_role', got %s", role.Name)
	}

	if len(role.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(role.Permissions))
	}
}

func TestAssignRevokeRole(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Assign admin role to assigner
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	// Assign role to user
	assignment, err := k.AssignRole(ctx, "admin1", "user1", types.RoleUser, 0)
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	if assignment.Address != "user1" {
		t.Errorf("Expected address 'user1', got %s", assignment.Address)
	}

	// Check permission
	if !k.HasPermission("admin1", types.PermissionAdmin) {
		t.Error("Admin should have admin permission")
	}

	// Revoke role
	err = k.RevokeRole(ctx, "admin1", "user1", types.RoleUser)
	if err != nil {
		t.Fatalf("Failed to revoke role: %v", err)
	}

	// Verify revoked
	assignments := k.GetRoleAssignments("user1")
	if len(assignments) != 0 {
		t.Errorf("Expected 0 assignments after revoke, got %d", len(assignments))
	}
}

func TestRoleExpiration(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Assign admin role
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	// Assign role with 1 second expiry
	_, err = k.AssignRole(ctx, "admin1", "user1", types.RoleUser, 1)
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	// Should have role initially
	assignments := k.GetRoleAssignments("user1")
	if len(assignments) != 1 {
		t.Errorf("Expected 1 assignment, got %d", len(assignments))
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Should be filtered out
	assignments = k.GetRoleAssignments("user1")
	if len(assignments) != 0 {
		t.Errorf("Expected 0 assignments after expiry, got %d", len(assignments))
	}
}

// Multisig Tests

func TestCreateMultisigWallet(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	// Create 3-of-5 wallet
	signers := []string{"signer1", "signer2", "signer3", "signer4", "signer5"}
	wallet, err := k.CreateMultisigWallet(ctx, "admin1", signers, 3, authproto.WalletType_WALLET_TYPE_3_OF_5)
	if err != nil {
		t.Fatalf("Failed to create multisig wallet: %v", err)
	}

	if wallet.Threshold != 3 {
		t.Errorf("Expected threshold 3, got %d", wallet.Threshold)
	}

	if len(wallet.Signers) != 5 {
		t.Errorf("Expected 5 signers, got %d", len(wallet.Signers))
	}
}

func TestMultisigProposal(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Setup
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	signers := []string{"signer1", "signer2", "signer3", "signer4", "signer5"}
	wallet, err := k.CreateMultisigWallet(ctx, "admin1", signers, 3, authproto.WalletType_WALLET_TYPE_3_OF_5)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Create proposal
	proposal, err := k.CreateMultisigProposal(ctx, "signer1", wallet.Id, "Test Proposal", "Test description", []byte("payload"), 3600)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Should have 1 signature (proposer auto-signs)
	if len(proposal.Signatures) != 1 {
		t.Errorf("Expected 1 signature, got %d", len(proposal.Signatures))
	}

	// Sign by signer2
	proposal, err = k.SignMultisigProposal(ctx, "signer2", proposal.Id)
	if err != nil {
		t.Fatalf("Failed to sign proposal: %v", err)
	}

	if len(proposal.Signatures) != 2 {
		t.Errorf("Expected 2 signatures, got %d", len(proposal.Signatures))
	}

	// Sign by signer3
	proposal, err = k.SignMultisigProposal(ctx, "signer3", proposal.Id)
	if err != nil {
		t.Fatalf("Failed to sign proposal: %v", err)
	}

	// Should be approved now (3-of-5)
	if proposal.Status != authproto.ProposalStatus_PROPOSAL_STATUS_APPROVED {
		t.Errorf("Expected approved status, got %s", proposal.Status)
	}

	// Execute proposal
	err = k.ExecuteMultisigProposal(ctx, "signer1", proposal.Id)
	if err != nil {
		t.Fatalf("Failed to execute proposal: %v", err)
	}

	// Verify executed
	proposal, err = k.GetMultisigProposal(proposal.Id)
	if err != nil {
		t.Fatalf("Failed to get proposal: %v", err)
	}

	if proposal.Status != authproto.ProposalStatus_PROPOSAL_STATUS_EXECUTED {
		t.Errorf("Expected executed status, got %s", proposal.Status)
	}
}

// Time-locked Actions Tests

func TestTimeLockedAction(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	// Create time-locked action with 2 second delay
	action, err := k.ProposeTimeLockedAction(ctx, "admin1", "UPDATE_PARAMS", []byte("params"), 2)
	if err != nil {
		t.Fatalf("Failed to propose action: %v", err)
	}

	if action.Status != authproto.ActionStatus_ACTION_STATUS_PENDING {
		t.Errorf("Expected pending status, got %s", action.Status)
	}

	// Try to execute immediately (should fail)
	err = k.ExecuteTimeLockedAction(ctx, "admin1", action.Id)
	if err != types.ErrActionNotReady {
		t.Errorf("Expected ErrActionNotReady, got %v", err)
	}

	// Wait for delay
	time.Sleep(3 * time.Second)

	// Should be ready now
	err = k.ExecuteTimeLockedAction(ctx, "admin1", action.Id)
	if err != nil {
		t.Fatalf("Failed to execute action: %v", err)
	}

	// Verify executed
	action, err = k.GetTimeLockedAction(action.Id)
	if err != nil {
		t.Fatalf("Failed to get action: %v", err)
	}

	if action.Status != authproto.ActionStatus_ACTION_STATUS_EXECUTED {
		t.Errorf("Expected executed status, got %s", action.Status)
	}
}

func TestCancelTimeLockedAction(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	// Create action
	action, err := k.ProposeTimeLockedAction(ctx, "admin1", "UPDATE_PARAMS", []byte("params"), 3600)
	if err != nil {
		t.Fatalf("Failed to propose action: %v", err)
	}

	// Cancel action
	err = k.CancelTimeLockedAction(ctx, "admin1", action.Id)
	if err != nil {
		t.Fatalf("Failed to cancel action: %v", err)
	}

	// Verify cancelled
	action, err = k.GetTimeLockedAction(action.Id)
	if err != nil {
		t.Fatalf("Failed to get action: %v", err)
	}

	if action.Status != authproto.ActionStatus_ACTION_STATUS_CANCELLED {
		t.Errorf("Expected cancelled status, got %s", action.Status)
	}
}

// Emergency Admin Tests

func TestEmergencyAdmin(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	// Activate emergency admin
	privileges := []string{types.PermissionAdmin, types.PermissionManageEmergency}
	admin, err := k.ActivateEmergencyAdmin(ctx, "admin1", "emergency1", privileges, 3600)
	if err != nil {
		t.Fatalf("Failed to activate emergency admin: %v", err)
	}

	if !admin.IsActive {
		t.Error("Emergency admin should be active")
	}

	// Check privilege
	if !k.HasEmergencyPrivilege("emergency1", types.PermissionAdmin) {
		t.Error("Emergency admin should have admin privilege")
	}

	// Deactivate
	err = k.DeactivateEmergencyAdmin(ctx, "admin1", "emergency1")
	if err != nil {
		t.Fatalf("Failed to deactivate emergency admin: %v", err)
	}

	// Verify deactivated
	admin, err = k.GetEmergencyAdmin("emergency1")
	if err != nil {
		t.Fatalf("Failed to get emergency admin: %v", err)
	}

	if admin.IsActive {
		t.Error("Emergency admin should be inactive")
	}
}

// Session Tests

func TestSessionManagement(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Create session
	metadata := map[string]string{"user_agent": "test"}
	session, err := k.CreateSession(ctx, "user1", "127.0.0.1", metadata)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if !session.IsActive {
		t.Error("Session should be active")
	}

	// Validate session
	validated, err := k.ValidateSession(ctx, session.SessionId)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}

	if validated.SessionId != session.SessionId {
		t.Error("Session IDs should match")
	}

	// Revoke session
	err = k.RevokeSession(ctx, "user1", session.SessionId)
	if err != nil {
		t.Fatalf("Failed to revoke session: %v", err)
	}

	// Should be inactive
	_, err = k.ValidateSession(ctx, session.SessionId)
	if err != types.ErrSessionInactive {
		t.Errorf("Expected ErrSessionInactive, got %v", err)
	}
}

// Rate Limiting Tests

func TestRateLimit(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Set low limit for testing
	k.params.DefaultRequestsPerMinute = 3

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		err := k.CheckRateLimit(ctx, "user1")
		if err != nil {
			t.Fatalf("Request %d should not be rate limited: %v", i+1, err)
		}
	}

	// 4th request should be limited
	err := k.CheckRateLimit(ctx, "user1")
	if err != types.ErrRateLimitExceeded {
		t.Errorf("Expected ErrRateLimitExceeded, got %v", err)
	}
}

func TestCustomRateLimit(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Assign admin
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	// Set custom limit
	err = k.SetCustomRateLimit(ctx, "admin1", "user1", 100, 6000, 144000)
	if err != nil {
		t.Fatalf("Failed to set custom rate limit: %v", err)
	}

	// Verify custom limit
	config, err := k.GetRateLimitConfig("user1")
	if err != nil {
		t.Fatalf("Failed to get rate limit config: %v", err)
	}

	if config.RequestsPerMinute != 100 {
		t.Errorf("Expected 100 requests per minute, got %d", config.RequestsPerMinute)
	}
}

// Audit Logging Tests

func TestAuditLogging(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Generate some audit logs
	k.LogAudit(ctx, "user1", "test_action", "resource1", "success", nil, "")
	k.LogAudit(ctx, "user2", "test_action", "resource2", "failed", nil, "error")

	// Get logs
	logs := k.GetRecentAuditLogs(10)
	if len(logs) < 2 {
		t.Errorf("Expected at least 2 logs, got %d", len(logs))
	}

	// Get by actor
	logs = k.GetAuditLogsByActor("user1", 10)
	if len(logs) == 0 {
		t.Error("Expected logs for user1")
	}

	// Get failed actions
	logs = k.GetFailedActions(10)
	if len(logs) == 0 {
		t.Error("Expected failed action logs")
	}
}

func TestAuditLogSearch(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Create logs
	k.LogAudit(ctx, "alice", "login", "system", "success", nil, "")
	k.LogAudit(ctx, "bob", "logout", "system", "success", nil, "")
	k.LogAudit(ctx, "alice", "create_role", "role1", "success", nil, "")

	// Search for alice
	logs := k.SearchAuditLogs("alice", 10)
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs for alice, got %d", len(logs))
	}

	// Search for login
	logs = k.SearchAuditLogs("login", 10)
	if len(logs) != 1 {
		t.Errorf("Expected 1 login log, got %d", len(logs))
	}
}

func TestAuditStatistics(t *testing.T) {
	k := setupKeeper()
	ctx := context.Background()

	// Create some logs
	k.LogAudit(ctx, "user1", "action1", "resource1", "success", nil, "")
	k.LogAudit(ctx, "user1", "action2", "resource2", "success", nil, "")
	k.LogAudit(ctx, "user2", "action1", "resource3", "failed", nil, "")

	// Get statistics
	stats := k.GetAuditStatistics()

	totalLogs := stats["total_logs"].(int)
	if totalLogs < 3 {
		t.Errorf("Expected at least 3 logs, got %d", totalLogs)
	}
}
