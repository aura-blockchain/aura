// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

type testAddresses struct {
	requester      string
	assistant      string
	admin          string
	user           string
	signer1        string
	signer2        string
	signerOutsider string
	proposer       string
	executor       string
	initiator      string
	emergencyAdmin string
	validator      string
	creator        string
	canceller      string
	deactivator    string
	owner          string
}

func newTestAddresses() testAddresses {
	return testAddresses{
		requester:      testAddr(40),
		assistant:      testAddr(41),
		admin:          testAddr(42),
		user:           testAddr(43),
		signer1:        testAddr(44),
		signer2:        testAddr(45),
		signerOutsider: testAddr(46),
		proposer:       testAddr(44),
		executor:       testAddr(45),
		initiator:      testAddr(47),
		emergencyAdmin: testAddr(48),
		validator:      testAddr(49),
		creator:        testAddr(50),
		canceller:      testAddr(51),
		deactivator:    testAddr(52),
		owner:          testAddr(53),
	}
}

// ========================
// REQUEST IDENTITY CHANGE
// ========================

func TestMsgServerRequestIdentityChange_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msg := &identitypb.MsgRequestIdentityChange{
		Requester:    addrs.requester,
		TargetDid:    "did:aura:target123",
		IrId:         "ir-001",
		MetadataHash: "hash123",
	}

	resp, err := msgServer.RequestIdentityChange(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.RequestId)
}

func TestMsgServerRequestIdentityChange_NilRequest(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	_, err := msgServer.RequestIdentityChange(ctx, nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

// ========================
// SUBMIT ASSISTANT PROOF
// ========================

func TestMsgServerSubmitAssistantProof_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))
	addrs := newTestAddresses()

	// Set up assistant with verify_identity permission
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "verifier",
		Permissions: []string{types.PermissionVerifyIdentity},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.assistant,
		RoleName: "verifier",
	}))

	// First create a change request
	createMsg := &identitypb.MsgRequestIdentityChange{
		Requester:    addrs.requester,
		TargetDid:    "did:aura:target123",
		IrId:         "ir-001",
		MetadataHash: "hash123",
	}
	createResp, err := msgServer.RequestIdentityChange(ctx, createMsg)
	require.NoError(t, err)

	// Submit assistant proof
	msg := &identitypb.MsgSubmitAssistantProof{
		RequestId: createResp.RequestId,
		Assistant: addrs.assistant,
		Success:   true,
	}

	resp, err := msgServer.SubmitAssistantProof(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerSubmitAssistantProof_NilRequest(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	_, err := msgServer.SubmitAssistantProof(ctx, nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

// ========================
// SUSPEND IDENTITY CHANGES
// ========================

func TestMsgServerSuspendIdentityChanges_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msg := &identitypb.MsgSuspendIdentityChanges{
		Authority: "authority",
		Reason:    "security incident",
	}

	resp, err := msgServer.SuspendIdentityChanges(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify suspended state
	suspended := keeper.IsSuspended(ctx)
	require.True(t, suspended)
}

func TestMsgServerSuspendIdentityChanges_InvalidAuthority(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msg := &identitypb.MsgSuspendIdentityChanges{
		Authority: "invalid",
		Reason:    "test",
	}

	_, err := msgServer.SuspendIdentityChanges(ctx, msg)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}

// ========================
// CREATE ROLE
// ========================

func TestMsgServerCreateRole_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Set up creator with permission
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "admin",
		Permissions: []string{types.PermissionAdmin},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.admin,
		RoleName: "admin",
	}))

	msg := &identitypb.MsgCreateRole{
		Creator:     addrs.admin,
		RoleName:    "test_role",
		Permissions: []string{types.PermissionViewAuditLogs},
		Description: "Test role description",
	}

	resp, err := msgServer.CreateRole(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify role was created
	role, err := keeper.GetRole(ctx, "test_role")
	require.NoError(t, err)
	require.Equal(t, "test_role", role.Name)
}

func TestMsgServerCreateRole_NilRequest(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	_, err := msgServer.CreateRole(ctx, nil)
	require.Error(t, err)
}

// ========================
// ASSIGN ROLE
// ========================

func TestMsgServerAssignRole_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Set up role and assigner with permission
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "admin",
		Permissions: []string{types.PermissionAdmin},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.admin,
		RoleName: "admin",
	}))

	// Create role to assign
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "viewer",
		Permissions: []string{types.PermissionViewAuditLogs},
	}))

	msg := &identitypb.MsgAssignRole{
		Assigner: addrs.admin,
		Address:  addrs.user,
		RoleName: "viewer",
	}

	resp, err := msgServer.AssignRole(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerAssignRole_NilRequest(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	_, err := msgServer.AssignRole(ctx, nil)
	require.Error(t, err)
}

// ========================
// REVOKE ROLE
// ========================

func TestMsgServerRevokeRole_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Set up role and admin
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "admin",
		Permissions: []string{types.PermissionAdmin},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.admin,
		RoleName: "admin",
	}))

	// Assign role to user
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "viewer",
		Permissions: []string{types.PermissionViewAuditLogs},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.user,
		RoleName: "viewer",
	}))

	msg := &identitypb.MsgRevokeRole{
		Revoker:  addrs.admin,
		Address:  addrs.user,
		RoleName: "viewer",
	}

	resp, err := msgServer.RevokeRole(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// ========================
// CREATE MULTISIG PROPOSAL
// ========================

func TestMsgServerCreateMultisigProposal_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create wallet first
	now := ctx.BlockTime()
	wallet := &types.MultisigWallet{
		Id:        "mswallet-test",
		Signers:   []string{addrs.signer1, addrs.signer2},
		Threshold: 2,
		CreatedAt: now,
		CreatedBy: addrs.creator,
	}
	require.NoError(t, keeper.SetMultisigWallet(ctx, wallet))

	msg := &identitypb.MsgCreateMultisigProposal{
		WalletId:    "mswallet-test",
		Proposer:    addrs.signer1,
		Title:       "Test Proposal",
		Description: "A test proposal",
		Payload:     []byte("test payload"),
	}

	resp, err := msgServer.CreateMultisigProposal(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ProposalId)
}

func TestMsgServerCreateMultisigProposal_NotSigner(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create wallet
	now := ctx.BlockTime()
	wallet := &types.MultisigWallet{
		Id:        "mswallet-test",
		Signers:   []string{addrs.signer1, addrs.signer2},
		Threshold: 2,
		CreatedAt: now,
		CreatedBy: addrs.creator,
	}
	require.NoError(t, keeper.SetMultisigWallet(ctx, wallet))

	msg := &identitypb.MsgCreateMultisigProposal{
		WalletId:    "mswallet-test",
		Proposer:    addrs.signerOutsider,
		Title:       "Test",
		Description: "Unauthorized",
	}

	_, err := msgServer.CreateMultisigProposal(ctx, msg)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}

// ========================
// SIGN MULTISIG PROPOSAL
// ========================

func TestMsgServerSignMultisigProposal_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create wallet
	now := ctx.BlockTime()
	wallet := &types.MultisigWallet{
		Id:        "mswallet-test",
		Signers:   []string{addrs.signer1, addrs.signer2},
		Threshold: 2,
		CreatedAt: now,
	}
	require.NoError(t, keeper.SetMultisigWallet(ctx, wallet))

	// Create proposal
	proposal := &types.MultisigProposal{
		Id:         "msprop-test",
		WalletId:   "mswallet-test",
		Status:     types.ProposalStatusPending,
		Signatures: []string{},
	}
	require.NoError(t, keeper.SetMultisigProposal(ctx, proposal))

	msg := &identitypb.MsgSignMultisigProposal{
		ProposalId: "msprop-test",
		Signer:     addrs.signer1,
	}

	resp, err := msgServer.SignMultisigProposal(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify signature was added
	updated, err := keeper.GetMultisigProposal(ctx, "msprop-test")
	require.NoError(t, err)
	require.Contains(t, updated.Signatures, addrs.signer1)
}

func TestMsgServerSignMultisigProposal_DuplicateSignature(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create wallet
	now := ctx.BlockTime()
	wallet := &types.MultisigWallet{
		Id:        "mswallet-test",
		Signers:   []string{addrs.signer1, addrs.signer2},
		Threshold: 2,
		CreatedAt: now,
	}
	require.NoError(t, keeper.SetMultisigWallet(ctx, wallet))

	// Create proposal with existing signature
	proposal := &types.MultisigProposal{
		Id:         "msprop-test",
		WalletId:   "mswallet-test",
		Status:     types.ProposalStatusPending,
		Signatures: []string{addrs.signer1},
	}
	require.NoError(t, keeper.SetMultisigProposal(ctx, proposal))

	msg := &identitypb.MsgSignMultisigProposal{
		ProposalId: "msprop-test",
		Signer:     addrs.signer1,
	}

	_, err := msgServer.SignMultisigProposal(ctx, msg)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
}

// ========================
// EXECUTE MULTISIG PROPOSAL
// ========================

func TestMsgServerExecuteMultisigProposal_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create approved proposal
	proposal := &types.MultisigProposal{
		Id:         "msprop-test",
		WalletId:   "mswallet-test",
		Status:     types.ProposalStatusApproved,
		Signatures: []string{addrs.signer1, addrs.signer2},
	}
	require.NoError(t, keeper.SetMultisigProposal(ctx, proposal))

	msg := &identitypb.MsgExecuteMultisigProposal{
		ProposalId: "msprop-test",
		Executor:   addrs.executor,
	}

	resp, err := msgServer.ExecuteMultisigProposal(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify status changed
	updated, err := keeper.GetMultisigProposal(ctx, "msprop-test")
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatusExecuted, updated.Status)
}

func TestMsgServerExecuteMultisigProposal_NotApproved(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create pending proposal
	proposal := &types.MultisigProposal{
		Id:       "msprop-test",
		WalletId: "mswallet-test",
		Status:   types.ProposalStatusPending,
	}
	require.NoError(t, keeper.SetMultisigProposal(ctx, proposal))

	msg := &identitypb.MsgExecuteMultisigProposal{
		ProposalId: "msprop-test",
		Executor:   addrs.executor,
	}

	_, err := msgServer.ExecuteMultisigProposal(ctx, msg)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

// ========================
// PROPOSE TIME LOCKED ACTION
// ========================

func TestMsgServerProposeTimeLockedAction_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Set up proposer with permission
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "timelock_manager",
		Permissions: []string{types.PermissionManageTimeLock},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.proposer,
		RoleName: "timelock_manager",
	}))

	msg := &identitypb.MsgProposeTimeLockedAction{
		Proposer:     addrs.proposer,
		ActionType:   "upgrade",
		Payload:      []byte("upgrade payload"),
		DelaySeconds: 3600,
	}

	resp, err := msgServer.ProposeTimeLockedAction(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ActionId)
	require.False(t, resp.ExecutableAt.IsZero())
}

// ========================
// EXECUTE TIME LOCKED ACTION
// ========================

func TestMsgServerExecuteTimeLockedAction_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create action that's ready to execute
	now := ctx.BlockTime()
	action := &types.TimeLockedAction{
		Id:           "tlaction-test",
		ActionType:   "upgrade",
		ExecutableAt: now.Add(-time.Hour), // Already past
		Status:       types.ActionStatusPending,
	}
	require.NoError(t, keeper.SetTimeLockedAction(ctx, action))

	msg := &identitypb.MsgExecuteTimeLockedAction{
		ActionId: "tlaction-test",
		Executor: addrs.executor,
	}

	resp, err := msgServer.ExecuteTimeLockedAction(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify status changed
	updated, err := keeper.GetTimeLockedAction(ctx, "tlaction-test")
	require.NoError(t, err)
	require.Equal(t, types.ActionStatusExecuted, updated.Status)
}

func TestMsgServerExecuteTimeLockedAction_DelayNotElapsed(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create action that's not ready
	now := ctx.BlockTime()
	action := &types.TimeLockedAction{
		Id:           "tlaction-test",
		ActionType:   "upgrade",
		ExecutableAt: now.Add(time.Hour), // Future
		Status:       types.ActionStatusPending,
	}
	require.NoError(t, keeper.SetTimeLockedAction(ctx, action))

	msg := &identitypb.MsgExecuteTimeLockedAction{
		ActionId: "tlaction-test",
		Executor: addrs.executor,
	}

	_, err := msgServer.ExecuteTimeLockedAction(ctx, msg)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

// ========================
// CANCEL TIME LOCKED ACTION
// ========================

func TestMsgServerCancelTimeLockedAction_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Set up canceller with permission
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "timelock_manager",
		Permissions: []string{types.PermissionManageTimeLock},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.canceller,
		RoleName: "timelock_manager",
	}))

	// Create pending action
	now := ctx.BlockTime()
	action := &types.TimeLockedAction{
		Id:           "tlaction-test",
		ActionType:   "upgrade",
		ExecutableAt: now.Add(time.Hour),
		Status:       types.ActionStatusPending,
	}
	require.NoError(t, keeper.SetTimeLockedAction(ctx, action))

	msg := &identitypb.MsgCancelTimeLockedAction{
		ActionId:  "tlaction-test",
		Canceller: addrs.canceller,
	}

	resp, err := msgServer.CancelTimeLockedAction(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify status changed
	updated, err := keeper.GetTimeLockedAction(ctx, "tlaction-test")
	require.NoError(t, err)
	require.Equal(t, types.ActionStatusCancelled, updated.Status)
}

// ========================
// ACTIVATE EMERGENCY ADMIN
// ========================

func TestMsgServerActivateEmergencyAdmin_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Set up activator with permission
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "emergency_manager",
		Permissions: []string{types.PermissionManageEmergency},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.initiator,
		RoleName: "emergency_manager",
	}))

	now := ctx.BlockTime()
	expiresAt := now.Add(24 * time.Hour)
	msg := &identitypb.MsgActivateEmergencyAdmin{
		Activator:    addrs.initiator,
		AdminAddress: addrs.emergencyAdmin,
		Privileges:   []string{"pause_system", "freeze_accounts"},
		ExpiresAt:    &expiresAt,
	}

	resp, err := msgServer.ActivateEmergencyAdmin(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify admin was created
	admin, err := keeper.GetEmergencyAdmin(ctx, addrs.emergencyAdmin)
	require.NoError(t, err)
	require.True(t, admin.IsActive)
}

// ========================
// DEACTIVATE EMERGENCY ADMIN
// ========================

func TestMsgServerDeactivateEmergencyAdmin_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Set up deactivator with permission
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "emergency_manager",
		Permissions: []string{types.PermissionManageEmergency},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.deactivator,
		RoleName: "emergency_manager",
	}))

	// Create active emergency admin
	now := ctx.BlockTime()
	admin := &types.EmergencyAdmin{
		Address:     addrs.emergencyAdmin,
		IsActive:    true,
		ActivatedAt: now,
	}
	require.NoError(t, keeper.SetEmergencyAdmin(ctx, admin))

	msg := &identitypb.MsgDeactivateEmergencyAdmin{
		Deactivator:  addrs.deactivator,
		AdminAddress: addrs.emergencyAdmin,
	}

	resp, err := msgServer.DeactivateEmergencyAdmin(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify admin was deactivated
	updated, err := keeper.GetEmergencyAdmin(ctx, addrs.emergencyAdmin)
	require.NoError(t, err)
	require.False(t, updated.IsActive)
}

// ========================
// ROTATE VALIDATOR KEY
// ========================

func TestMsgServerRotateValidatorKey_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Set up validator with permission
	require.NoError(t, keeper.SetRole(ctx, &types.Role{
		Name:        "validator",
		Permissions: []string{types.PermissionRotateValidatorKey},
	}))
	require.NoError(t, keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  addrs.validator,
		RoleName: "validator",
	}))

	msg := &identitypb.MsgRotateValidatorKey{
		ValidatorAddress:   addrs.validator,
		NewConsensusPubkey: "newpubkey123",
	}

	resp, err := msgServer.RotateValidatorKey(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// ========================
// ROTATE DID KEY
// ========================

func TestMsgServerRotateDIDKey_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create DID record
	did := "did:aura:test123"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   addrs.owner,
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
	}
	require.NoError(t, keeper.SetIdentityRecord(ctx, record))

	msg := &identitypb.MsgRotateDIDKey{
		Did:                   did,
		Initiator:             addrs.owner,
		NewVerificationMethod: "newkey123",
		Reason:                "security rotation",
	}

	resp, err := msgServer.RotateDIDKey(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.RotationTime.IsZero())
	require.False(t, resp.GracePeriodEnd.IsZero())
}

func TestMsgServerRotateDIDKey_EmptyDID(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	addrs := newTestAddresses()
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msg := &identitypb.MsgRotateDIDKey{
		Did:                   "",
		Initiator:             addrs.owner,
		NewVerificationMethod: "newkey123",
	}

	_, err := msgServer.RotateDIDKey(ctx, msg)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}
