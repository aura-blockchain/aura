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
	require.Equal(t, *params, resp.Params)
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
		Permission: types.PermissionManageRoles,
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
		Id:        walletID,
		Signers:   []string{"aura1owner1", "aura1owner2"},
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
	require.Equal(t, walletID, resp.Wallet.Id)
	require.Equal(t, uint32(2), resp.Wallet.Threshold)
}

func TestQueryServerAllMultisigWallets(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create multiple multisig wallets
	now := ctx.BlockTime()
	for i := 1; i <= 3; i++ {
		walletID := "wallet" + string(rune('0'+i))
		wallet := &types.MultisigWallet{
			Id:        walletID,
			Signers:   []string{"aura1owner1", "aura1owner2"},
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
			IsActive:     i%2 == 0,
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
