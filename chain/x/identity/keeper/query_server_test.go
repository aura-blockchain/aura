package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

func TestQueryServerParams(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Set params
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Query params
	resp, err := queryServer.Params(ctx, &identitypb.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, params, resp.Params)
}

func TestQueryServerIdentityRecord_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create identity record
	did := "did:aura:test123"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   "aura1test",
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
		UpdatedAt: &now,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Query identity record
	resp, err := queryServer.IdentityRecord(ctx, &identitypb.QueryIdentityRecordRequest{
		Did: did,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, did, resp.Record.Did)
	require.Equal(t, "aura1test", resp.Record.Address)
}

func TestQueryServerIdentityRecord_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent identity
	_, err := queryServer.IdentityRecord(ctx, &identitypb.QueryIdentityRecordRequest{
		Did: "did:aura:nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerIdentityRecordByAddress_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create identity record
	did := "did:aura:test123"
	address := "aura1test"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   address,
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
		UpdatedAt: &now,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Query by address
	resp, err := queryServer.IdentityRecordByAddress(ctx, &identitypb.QueryIdentityRecordByAddressRequest{
		Address: address,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, did, resp.Record.Did)
	require.Equal(t, address, resp.Record.Address)
}

func TestQueryServerIdentityRecordByAddress_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent address
	_, err := queryServer.IdentityRecordByAddress(ctx, &identitypb.QueryIdentityRecordByAddressRequest{
		Address: "aura1nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerAllIdentityRecords(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multiple identity records
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		did := "did:aura:test" + string(rune('0'+i))
		address := "aura1test" + string(rune('0'+i))
		record := &types.IdentityRecord{
			Did:       did,
			Address:   address,
			Status:    types.IdentityStatusActive,
			CreatedAt: now,
			UpdatedAt: &now,
		}
		err := keeper.SetIdentityRecord(ctx, record)
		require.NoError(t, err)
	}

	// Query all records
	resp, err := queryServer.AllIdentityRecords(ctx, &identitypb.QueryAllIdentityRecordsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Records, 3)
}

func TestQueryServerChangeRequest_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create change request
	requestID := "req123"
	now := ctx.BlockTime()
	request := &types.ChangeRequest{
		RequestId: requestID,
		Did:       "did:aura:test",
		Requester: "aura1requester",
		Status:    types.ChangeRequestStatusPending,
		CreatedAt: now,
	}

	err := keeper.SetChangeRequest(ctx, request)
	require.NoError(t, err)

	// Query change request
	resp, err := queryServer.ChangeRequest(ctx, &identitypb.QueryChangeRequestRequest{
		RequestId: requestID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, requestID, resp.Request.RequestId)
}

func TestQueryServerChangeRequest_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent request
	_, err := queryServer.ChangeRequest(ctx, &identitypb.QueryChangeRequestRequest{
		RequestId: "nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerChangeRequestsByDID(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multiple change requests for same DID
	did := "did:aura:test"
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		requestID := "req" + string(rune('0'+i))
		request := &types.ChangeRequest{
			RequestId: requestID,
			Did:       did,
			Requester: "aura1requester",
			Status:    types.ChangeRequestStatusPending,
			CreatedAt: now,
		}
		err := keeper.SetChangeRequest(ctx, request)
		require.NoError(t, err)
	}

	// Query requests by DID
	resp, err := queryServer.ChangeRequestsByDID(ctx, &identitypb.QueryChangeRequestsByDIDRequest{
		Did:        did,
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Requests, 3)
}

func TestQueryServerChangeHistory(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create change history entries
	did := "did:aura:test"
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		entry := &types.ChangeHistoryEntry{
			Did:       did,
			ChangedAt: now,
			ChangedBy: "aura1changer",
			Action:    "test_action",
		}
		err := keeper.AppendChangeHistory(ctx, entry)
		require.NoError(t, err)
	}

	// Query change history
	resp, err := queryServer.ChangeHistory(ctx, &identitypb.QueryChangeHistoryRequest{
		Did:        did,
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.History, 3)
}

func TestQueryServerRole_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create role
	roleName := "admin"
	role := &types.Role{
		Name:         roleName,
		Permissions:  []string{types.PermissionAdmin},
		Description:  "Admin role",
		CreatedAt:    ctx.BlockTime(),
		IsSystemRole: false,
		UpdatedAt:    func() *time.Time { t := ctx.BlockTime(); return &t }(),
	}

	err := keeper.SetRole(ctx, role)
	require.NoError(t, err)

	// Query role
	resp, err := queryServer.Role(ctx, &identitypb.QueryRoleRequest{
		Name: roleName,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, roleName, resp.Role.Name)
}

func TestQueryServerRole_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent role
	_, err := queryServer.Role(ctx, &identitypb.QueryRoleRequest{
		Name: "nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerAllRoles(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multiple roles
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		roleName := "role" + string(rune('0'+i))
		role := &types.Role{
			Name:         roleName,
			Permissions:  []string{types.PermissionAdmin},
			Description:  "Test role",
			CreatedAt:    now,
			IsSystemRole: false,
			UpdatedAt:    &now,
		}
		err := keeper.SetRole(ctx, role)
		require.NoError(t, err)
	}

	// Query all roles
	resp, err := queryServer.AllRoles(ctx, &identitypb.QueryAllRolesRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.GreaterOrEqual(t, len(resp.Roles), 3)
}

func TestQueryServerRoleAssignments(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create role assignments
	address := "aura1test"
	now := ctx.BlockTime()
	for i := 1; i <= 2; i++ {
		roleName := "role" + string(rune('0'+i))
		assignment := &types.RoleAssignment{
			Address:    address,
			Role:       roleName,
			AssignedAt: now,
			AssignedBy: "aura1admin",
		}
		err := keeper.SetRoleAssignment(ctx, assignment)
		require.NoError(t, err)
	}

	// Query role assignments
	resp, err := queryServer.RoleAssignments(ctx, &identitypb.QueryRoleAssignmentsRequest{
		Address:    address,
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Assignments, 2)
}

func TestQueryServerHasPermission(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create role with permissions
	roleName := "admin"
	role := &types.Role{
		Name:         roleName,
		Permissions:  []string{types.PermissionAdmin},
		Description:  "Admin role",
		CreatedAt:    ctx.BlockTime(),
		IsSystemRole: false,
		UpdatedAt:    func() *time.Time { t := ctx.BlockTime(); return &t }(),
	}
	err := keeper.SetRole(ctx, role)
	require.NoError(t, err)

	// Assign role to address
	address := "aura1test"
	assignment := &types.RoleAssignment{
		Address:    address,
		Role:       roleName,
		AssignedAt: ctx.BlockTime(),
		AssignedBy: "aura1admin",
	}
	err = keeper.SetRoleAssignment(ctx, assignment)
	require.NoError(t, err)

	// Query permission check - should have permission
	resp, err := queryServer.HasPermission(ctx, &identitypb.QueryHasPermissionRequest{
		Address:    address,
		Permission: types.PermissionAdmin,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.HasPermission)

	// Query permission check - should not have permission
	resp2, err := queryServer.HasPermission(ctx, &identitypb.QueryHasPermissionRequest{
		Address:    address,
		Permission: types.PermissionManageRoles, // Different permission
	})
	require.NoError(t, err)
	require.NotNil(t, resp2)
	require.False(t, resp2.HasPermission)
}

func TestQueryServerMultisigWallet_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multisig wallet
	walletID := "wallet123"
	now := ctx.BlockTime()
	wallet := &types.MultisigWallet{
		WalletId:  walletID,
		Owners:    []string{"aura1owner1", "aura1owner2"},
		Threshold: 2,
		CreatedAt: now,
	}

	err := keeper.SetMultisigWallet(ctx, wallet)
	require.NoError(t, err)

	// Query multisig wallet
	resp, err := queryServer.MultisigWallet(ctx, &identitypb.QueryMultisigWalletRequest{
		WalletId: walletID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, walletID, resp.Wallet.WalletId)
	require.Equal(t, uint32(2), resp.Wallet.Threshold)
}

func TestQueryServerMultisigWallet_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent wallet
	_, err := queryServer.MultisigWallet(ctx, &identitypb.QueryMultisigWalletRequest{
		WalletId: "nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerAllMultisigWallets(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multiple multisig wallets
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		walletID := "wallet" + string(rune('0'+i))
		wallet := &types.MultisigWallet{
			WalletId:  walletID,
			Owners:    []string{"aura1owner1", "aura1owner2"},
			Threshold: 2,
			CreatedAt: now,
		}
		err := keeper.SetMultisigWallet(ctx, wallet)
		require.NoError(t, err)
	}

	// Query all wallets
	resp, err := queryServer.AllMultisigWallets(ctx, &identitypb.QueryAllMultisigWalletsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Wallets, 3)
}

func TestQueryServerMultisigProposal_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multisig proposal
	proposalID := "proposal123"
	now := ctx.BlockTime()
	proposal := &types.MultisigProposal{
		ProposalId: proposalID,
		WalletId:   "wallet1",
		Proposer:   "aura1proposer",
		Status:     types.ProposalStatusPending,
		CreatedAt:  now,
	}

	err := keeper.SetMultisigProposal(ctx, proposal)
	require.NoError(t, err)

	// Query proposal
	resp, err := queryServer.MultisigProposal(ctx, &identitypb.QueryMultisigProposalRequest{
		ProposalId: proposalID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, proposalID, resp.Proposal.ProposalId)
}

func TestQueryServerMultisigProposal_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent proposal
	_, err := queryServer.MultisigProposal(ctx, &identitypb.QueryMultisigProposalRequest{
		ProposalId: "nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerMultisigProposalsByWallet(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multiple proposals for same wallet
	walletID := "wallet1"
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		proposalID := "proposal" + string(rune('0'+i))
		proposal := &types.MultisigProposal{
			ProposalId: proposalID,
			WalletId:   walletID,
			Proposer:   "aura1proposer",
			Status:     types.ProposalStatusPending,
			CreatedAt:  now,
		}
		err := keeper.SetMultisigProposal(ctx, proposal)
		require.NoError(t, err)
	}

	// Query proposals by wallet
	resp, err := queryServer.MultisigProposalsByWallet(ctx, &identitypb.QueryMultisigProposalsByWalletRequest{
		WalletId:   walletID,
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Proposals, 3)
}

func TestQueryServerTimeLockedAction_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create time-locked action
	actionID := "action123"
	now := ctx.BlockTime()
	action := &types.TimeLockedAction{
		ActionId:     actionID,
		Proposer:     "aura1proposer",
		Status:       types.ActionStatusPending,
		ProposedAt:   now,
		ExecutableAt: now.Add(24 * time.Hour),
	}

	err := keeper.SetTimeLockedAction(ctx, action)
	require.NoError(t, err)

	// Query action
	resp, err := queryServer.TimeLockedAction(ctx, &identitypb.QueryTimeLockedActionRequest{
		ActionId: actionID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, actionID, resp.Action.ActionId)
}

func TestQueryServerTimeLockedAction_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent action
	_, err := queryServer.TimeLockedAction(ctx, &identitypb.QueryTimeLockedActionRequest{
		ActionId: "nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerAllTimeLockedActions(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multiple time-locked actions
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		actionID := "action" + string(rune('0'+i))
		action := &types.TimeLockedAction{
			ActionId:     actionID,
			Proposer:     "aura1proposer",
			Status:       types.ActionStatusPending,
			ProposedAt:   now,
			ExecutableAt: now.Add(24 * time.Hour),
		}
		err := keeper.SetTimeLockedAction(ctx, action)
		require.NoError(t, err)
	}

	// Query all actions
	resp, err := queryServer.AllTimeLockedActions(ctx, &identitypb.QueryAllTimeLockedActionsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Actions, 3)
}

func TestQueryServerEmergencyAdmin_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create emergency admin
	adminID := "admin123"
	now := ctx.BlockTime()
	admin := &types.EmergencyAdmin{
		AdminId:      adminID,
		Address:      "aura1admin",
		ActivatedAt:  now,
		IsActive:     true,
		ActivatedBy:  "aura1authority",
		DeactivatedBy: "",
	}

	err := keeper.SetEmergencyAdmin(ctx, admin)
	require.NoError(t, err)

	// Query emergency admin
	resp, err := queryServer.EmergencyAdmin(ctx, &identitypb.QueryEmergencyAdminRequest{
		AdminId: adminID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, adminID, resp.Admin.AdminId)
	require.True(t, resp.Admin.IsActive)
}

func TestQueryServerEmergencyAdmin_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent admin
	_, err := queryServer.EmergencyAdmin(ctx, &identitypb.QueryEmergencyAdminRequest{
		AdminId: "nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerAllEmergencyAdmins(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multiple emergency admins
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		adminID := "admin" + string(rune('0'+i))
		admin := &types.EmergencyAdmin{
			AdminId:      adminID,
			Address:      "aura1admin",
			ActivatedAt:  now,
			IsActive:     i%2 == 0, // Alternate active/inactive
			ActivatedBy:  "aura1authority",
			DeactivatedBy: "",
		}
		err := keeper.SetEmergencyAdmin(ctx, admin)
		require.NoError(t, err)
	}

	// Query all admins
	resp, err := queryServer.AllEmergencyAdmins(ctx, &identitypb.QueryAllEmergencyAdminsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Admins, 3)
}

func TestQueryServerValidatorRotation_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create validator rotation
	rotationID := "rotation123"
	now := ctx.BlockTime()
	rotation := &types.ValidatorRotation{
		RotationId:    rotationID,
		ValidatorAddr: sdk.ValAddress("validator1").String(),
		OldPubKey:     "old_pubkey",
		NewPubKey:     "new_pubkey",
		RotatedAt:     now,
	}

	err := keeper.SetValidatorRotation(ctx, rotation)
	require.NoError(t, err)

	// Query rotation
	resp, err := queryServer.ValidatorRotation(ctx, &identitypb.QueryValidatorRotationRequest{
		RotationId: rotationID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, rotationID, resp.Rotation.RotationId)
}

func TestQueryServerValidatorRotation_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent rotation
	_, err := queryServer.ValidatorRotation(ctx, &identitypb.QueryValidatorRotationRequest{
		RotationId: "nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerSession_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create session
	sessionID := "session123"
	now := ctx.BlockTime()
	session := &types.Session{
		SessionId: sessionID,
		Address:   "aura1test",
		CreatedAt: now,
		ExpiresAt: now.Add(1 * time.Hour),
		IsActive:  true,
	}

	err := keeper.SetSession(ctx, session)
	require.NoError(t, err)

	// Query session
	resp, err := queryServer.Session(ctx, &identitypb.QuerySessionRequest{
		SessionId: sessionID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, sessionID, resp.Session.SessionId)
	require.True(t, resp.Session.IsActive)
}

func TestQueryServerSession_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent session
	_, err := queryServer.Session(ctx, &identitypb.QuerySessionRequest{
		SessionId: "nonexistent",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerSessionsByAddress(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multiple sessions for same address
	address := "aura1test"
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		sessionID := "session" + string(rune('0'+i))
		session := &types.Session{
			SessionId: sessionID,
			Address:   address,
			CreatedAt: now,
			ExpiresAt: now.Add(1 * time.Hour),
			IsActive:  true,
		}
		err := keeper.SetSession(ctx, session)
		require.NoError(t, err)
	}

	// Query sessions by address
	resp, err := queryServer.SessionsByAddress(ctx, &identitypb.QuerySessionsByAddressRequest{
		Address:    address,
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Sessions, 3)
}

func TestQueryServerRateLimit_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create rate limit
	address := "aura1test"
	now := ctx.BlockTime()
	rateLimit := &types.RateLimit{
		Address:       address,
		Action:        "test_action",
		Count:         5,
		WindowStart:   now,
		WindowEnd:     now.Add(1 * time.Hour),
		MaxAllowed:    10,
		WindowSeconds: 3600,
	}

	err := keeper.SetRateLimit(ctx, rateLimit)
	require.NoError(t, err)

	// Query rate limit
	resp, err := queryServer.RateLimit(ctx, &identitypb.QueryRateLimitRequest{
		Address: address,
		Action:  "test_action",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(5), resp.RateLimit.Count)
}

func TestQueryServerRateLimit_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Query non-existent rate limit
	_, err := queryServer.RateLimit(ctx, &identitypb.QueryRateLimitRequest{
		Address: "aura1nonexistent",
		Action:  "test_action",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryServerAuditLogs(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create audit logs
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		log := &types.AuditLog{
			Action:    "test_action",
			Actor:     "aura1actor",
			Timestamp: now,
			Details:   "Test details",
		}
		err := keeper.AppendAuditLog(ctx, log)
		require.NoError(t, err)
	}

	// Query audit logs
	resp, err := queryServer.AuditLogs(ctx, &identitypb.QueryAuditLogsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Logs, 3)
}

func TestQueryServerAuditLogsByActor(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create audit logs for specific actor
	actor := "aura1actor"
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		log := &types.AuditLog{
			Action:    "test_action",
			Actor:     actor,
			Timestamp: now,
			Details:   "Test details",
		}
		err := keeper.AppendAuditLog(ctx, log)
		require.NoError(t, err)
	}

	// Create log for different actor (should not be returned)
	otherLog := &types.AuditLog{
		Action:    "test_action",
		Actor:     "aura1other",
		Timestamp: now,
		Details:   "Other details",
	}
	err := keeper.AppendAuditLog(ctx, otherLog)
	require.NoError(t, err)

	// Query audit logs by actor
	resp, err := queryServer.AuditLogsByActor(ctx, &identitypb.QueryAuditLogsByActorRequest{
		Actor:      actor,
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Logs, 3)
	for _, log := range resp.Logs {
		require.Equal(t, actor, log.Actor)
	}
}
