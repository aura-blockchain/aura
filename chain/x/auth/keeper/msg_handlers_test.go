package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// ============================================================================
// Message Handler Tests
// ============================================================================

func TestRotateValidatorKey(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign validator role
	_, err := k.AssignRole(ctx, "system", "validator1", types.RoleValidator, 0)
	require.NoError(t, err)

	// Rotate key
	newPubKey := []byte("new_public_key")
	rotation, err := k.RotateValidatorKey(sdk.WrapSDKContext(ctx), "validator1", newPubKey)
	require.NoError(t, err)
	require.NotNil(t, rotation)
	require.Equal(t, "validator1", rotation.ValidatorAddress)
	require.Equal(t, string(newPubKey), rotation.NewConsensusPubkey)
	require.Equal(t, authproto.RotationStatus_ROTATION_STATUS_PENDING, rotation.RotationStatus)
}

func TestRotateValidatorKey_NoPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Try without permission
	_, err := k.RotateValidatorKey(sdk.WrapSDKContext(ctx), "user1", []byte("key"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestInitiateValidatorKeyRotation(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	// Initiate rotation
	newPubKey := []byte("new_consensus_key")
	rotation, err := k.InitiateValidatorKeyRotation(sdk.WrapSDKContext(ctx), "admin1", "validator1", newPubKey)
	require.NoError(t, err)
	require.NotNil(t, rotation)
	require.Equal(t, "validator1", rotation.ValidatorAddress)
	require.Equal(t, "admin1", rotation.InitiatedBy)
	require.Equal(t, authproto.RotationStatus_ROTATION_STATUS_PENDING, rotation.RotationStatus)
}

func TestInitiateValidatorKeyRotation_NoPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.InitiateValidatorKeyRotation(sdk.WrapSDKContext(ctx), "user1", "validator1", []byte("key"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestCompleteValidatorKeyRotation(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	// Create rotation first
	rotation, err := k.InitiateValidatorKeyRotation(sdk.WrapSDKContext(ctx), "admin1", "validator1", []byte("key"))
	require.NoError(t, err)
	require.NotNil(t, rotation)

	// Complete rotation
	err = k.CompleteValidatorKeyRotation(sdk.WrapSDKContext(ctx), "admin1", "validator1")
	require.NoError(t, err)

	// Verify status
	completed, err := k.GetValidatorKeyRotation(ctx, "validator1")
	require.NoError(t, err)
	require.Equal(t, authproto.RotationStatus_ROTATION_STATUS_COMPLETED, completed.RotationStatus)
}

func TestCompleteValidatorKeyRotation_NoPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	err := k.CompleteValidatorKeyRotation(sdk.WrapSDKContext(ctx), "user1", "validator1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestCompleteValidatorKeyRotation_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	err = k.CompleteValidatorKeyRotation(sdk.WrapSDKContext(ctx), "admin1", "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "rotation record not found")
}

func TestRevokeSession(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create session
	session, err := k.CreateSession(sdk.WrapSDKContext(ctx), "user1", 3600)
	require.NoError(t, err)
	require.NotNil(t, session)

	// Revoke session
	err = k.RevokeSession(sdk.WrapSDKContext(ctx), "user1", session.SessionId)
	require.NoError(t, err)

	// Verify session is gone
	_, err = k.GetSession(ctx, session.SessionId)
	require.Error(t, err)
}

// NOTE: The following tests have been REMOVED as they tested stub functions that are now deleted.
// All query functionality is now tested via query_server_test.go with proper SDK context handling.
//
// Removed tests (functionality now in query_server_test.go):
// - TestGetRole - tested via query server
// - TestListRoles - tested via query server
// - TestListMultisigWallets - tested via query server
// - TestListMultisigProposals - tested via query server
// - TestListTimeLockedActions - tested via query server
// - TestListEmergencyAdmins - tested via query server
// - TestListValidatorKeyRotations - tested via query server
// - TestListSessions - tested via query server
// - TestGetRateLimitStatus - tested via query server
//
// These stub functions were non-functional placeholders without SDK context.
// See query_server_test.go for production query tests with full KVStore integration.

// ============================================================================
// Additional Edge Cases for Existing Message Handlers
// ============================================================================

func TestCreateRole_WithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Try to create role without permission
	_, err := k.CreateRole(sdk.WrapSDKContext(ctx), "user1", "new_role", []string{"perm1"}, "Test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestAssignRole_EmptyAddress(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign admin role first
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	// Try to assign role to empty address
	_, err = k.AssignRole(sdk.WrapSDKContext(ctx), "admin1", "", types.RoleUser, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "address cannot be empty")
}

func TestAssignRole_NonexistentRole(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign admin role first
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	// Try to assign nonexistent role
	_, err = k.AssignRole(sdk.WrapSDKContext(ctx), "admin1", "user1", "nonexistent_role", 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "role not found")
}

func TestAssignRole_WithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Try to assign role without permission
	_, err := k.AssignRole(sdk.WrapSDKContext(ctx), "user1", "user2", types.RoleUser, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestRevokeRole_WithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Try to revoke role without permission
	err := k.RevokeRole(sdk.WrapSDKContext(ctx), "user1", "user2", types.RoleUser)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestCreateMultisigWallet_WithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Try to create wallet without permission
	_, err := k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "user1", []string{"s1", "s2"}, 2, "custom")
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestCreateMultisigWallet_InvalidThreshold(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	// Try with threshold 0
	_, err = k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"s1", "s2"}, 0, "custom")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid threshold")

	// Try with threshold > signers
	_, err = k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"s1", "s2"}, 5, "custom")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid threshold")
}

func TestCreateMultisigProposal_WalletNotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Try to create proposal for nonexistent wallet
	_, err := k.CreateMultisigProposal(sdk.WrapSDKContext(ctx), "proposer1", "nonexistent", "Title", "Desc", []byte("data"), 3600)
	require.Error(t, err)
}

func TestCreateMultisigProposal_NotSigner(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign permission and create wallet
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	wallet, err := k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"signer1", "signer2"}, 2, "custom")
	require.NoError(t, err)

	// Try to create proposal with non-signer
	_, err = k.CreateMultisigProposal(sdk.WrapSDKContext(ctx), "not_a_signer", wallet.Id, "Title", "Desc", []byte("data"), 3600)
	require.Error(t, err)
	require.Contains(t, err.Error(), "proposer is not a signer")
}

func TestSignMultisigProposal_ProposalNotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.SignMultisigProposal(sdk.WrapSDKContext(ctx), "signer1", "nonexistent")
	require.Error(t, err)
}

func TestSignMultisigProposal_NotAuthorized(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create wallet and proposal
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	wallet, err := k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"signer1", "signer2"}, 2, "custom")
	require.NoError(t, err)

	proposal, err := k.CreateMultisigProposal(sdk.WrapSDKContext(ctx), "signer1", wallet.Id, "Title", "Desc", []byte("data"), 3600)
	require.NoError(t, err)

	// Try to sign with unauthorized signer
	_, err = k.SignMultisigProposal(sdk.WrapSDKContext(ctx), "not_authorized", proposal.Id)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not authorized")
}

func TestSignMultisigProposal_AlreadySigned(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create wallet and proposal
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	wallet, err := k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"signer1", "signer2"}, 2, "custom")
	require.NoError(t, err)

	proposal, err := k.CreateMultisigProposal(sdk.WrapSDKContext(ctx), "signer1", wallet.Id, "Title", "Desc", []byte("data"), 3600)
	require.NoError(t, err)

	// Signer1 already signed when creating proposal
	_, err = k.SignMultisigProposal(sdk.WrapSDKContext(ctx), "signer1", proposal.Id)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already signed")
}

func TestExecuteMultisigProposal_ProposalNotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	err := k.ExecuteMultisigProposal(sdk.WrapSDKContext(ctx), "executor", "nonexistent")
	require.Error(t, err)
}

func TestExecuteMultisigProposal_NotApproved(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create wallet and proposal
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	wallet, err := k.CreateMultisigWallet(sdk.WrapSDKContext(ctx), "admin1", []string{"s1", "s2", "s3"}, 3, "custom")
	require.NoError(t, err)

	proposal, err := k.CreateMultisigProposal(sdk.WrapSDKContext(ctx), "s1", wallet.Id, "Title", "Desc", []byte("data"), 3600)
	require.NoError(t, err)

	// Try to execute before threshold reached (only 1 signature)
	err = k.ExecuteMultisigProposal(sdk.WrapSDKContext(ctx), "s1", proposal.Id)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not approved")
}

func TestProposeTimeLockedAction_WithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.ProposeTimeLockedAction(sdk.WrapSDKContext(ctx), "user1", "action", []byte("data"), 3600)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestExecuteTimeLockedAction_WithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	err := k.ExecuteTimeLockedAction(sdk.WrapSDKContext(ctx), "user1", "action1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestExecuteTimeLockedAction_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	err = k.ExecuteTimeLockedAction(sdk.WrapSDKContext(ctx), "admin1", "nonexistent")
	require.Error(t, err)
}

func TestCancelTimeLockedAction_WithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	err := k.CancelTimeLockedAction(sdk.WrapSDKContext(ctx), "user1", "action1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestCancelTimeLockedAction_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	err = k.CancelTimeLockedAction(sdk.WrapSDKContext(ctx), "admin1", "nonexistent")
	require.Error(t, err)
}

func TestActivateEmergencyAdmin_WithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.ActivateEmergencyAdmin(sdk.WrapSDKContext(ctx), "user1", "admin1", []string{"perm1"}, 3600)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestDeactivateEmergencyAdmin_WithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	err := k.DeactivateEmergencyAdmin(sdk.WrapSDKContext(ctx), "user1", "admin1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient permissions")
}

func TestDeactivateEmergencyAdmin_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign permission
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	err = k.DeactivateEmergencyAdmin(sdk.WrapSDKContext(ctx), "admin1", "nonexistent")
	require.Error(t, err)
}

func TestInvalidateSession_NotFound(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	err := k.InvalidateSession(sdk.WrapSDKContext(ctx), "user1", "nonexistent")
	require.Error(t, err)
}
