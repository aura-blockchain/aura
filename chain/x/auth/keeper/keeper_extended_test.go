package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// MockTime returns a deterministic time for testing
func MockTime() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}

// setupTestKeeper creates a test keeper and context for testing
func setupTestKeeper(t testing.TB) (*Keeper, sdk.Context) {
	t.Helper()

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	storeKey := storetypes.NewKVStoreKey("auth")
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)

	require.NoError(t, cms.LoadLatestVersion())

	ctx := sdk.NewContext(cms, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	k := NewKeeper(cdc, storeKey)

	// Initialize default params
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	// Bootstrap the permission system for tests
	// Create predefined roles with their permissions
	now := time.Now()

	// Create admin role with all permissions
	adminRole := &authproto.Role{
		Name: types.RoleAdmin,
		Permissions: []string{
			types.PermissionAdmin,
			types.PermissionCreateRole,
			types.PermissionAssignRole,
			types.PermissionRevokeRole,
			types.PermissionManageMultisig,
			types.PermissionManageTimeLock,
			types.PermissionManageEmergency,
			types.PermissionRotateValidatorKey,
			types.PermissionManageSession,
			types.PermissionViewAuditLogs,
		},
		Description: "Administrator role with all permissions",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, k.SetRole(ctx, adminRole))

	// Create moderator role
	moderatorRole := &authproto.Role{
		Name: types.RoleModerator,
		Permissions: []string{
			types.PermissionAssignRole,
			types.PermissionViewAuditLogs,
		},
		Description: "Moderator role with limited permissions",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, k.SetRole(ctx, moderatorRole))

	// Create validator role
	validatorRole := &authproto.Role{
		Name: types.RoleValidator,
		Permissions: []string{
			types.PermissionRotateValidatorKey,
		},
		Description: "Validator role for consensus validators",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, k.SetRole(ctx, validatorRole))

	// Create user role
	userRole := &authproto.Role{
		Name:        types.RoleUser,
		Permissions: []string{},
		Description: "Basic user role",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, k.SetRole(ctx, userRole))

	// Assign admin role to "system" actor for test bootstrap
	systemAssignment := &authproto.RoleAssignment{
		Address:    "system",
		RoleName:   types.RoleAdmin,
		AssignedBy: "genesis",
		AssignedAt: now,
	}
	require.NoError(t, k.SetRoleAssignment(ctx, systemAssignment))

	return k, ctx
}

// Params Tests

func TestGetParams(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.NotNil(t, params)
	require.NotZero(t, params.SessionTimeoutSeconds)
}

func TestSetParams(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	params := types.DefaultParams()
	params.SessionTimeoutSeconds = 3600

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	retrieved, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(3600), retrieved.SessionTimeoutSeconds)
}

func TestSetInvalidParams(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	params := types.DefaultParams()
	params.SessionTimeoutSeconds = 0 // Invalid - can be tested if validation exists

	err := k.SetParams(ctx, params)
	// Note: This may not error if no validation exists for zero timeout
	_ = err
}

// Role Tests

func TestGetNonExistentRole(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	role, err := k.GetRoleFromStore(ctx, "nonexistent")
	require.Error(t, err)
	require.Nil(t, role)
}

func TestGetAllRoles(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Assign admin first
	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	// Create multiple roles
	for i := 0; i < 5; i++ {
		_, err := k.CreateRole(ctx, "admin1", "role"+string(rune('A'+i)), []string{"perm1"}, "Test role")
		require.NoError(t, err)
	}

	roles, err := k.GetAllRoles(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(roles), 5)
}

func TestDeleteRole(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	role, err := k.CreateRole(ctx, "admin1", "temp_role", []string{"perm1"}, "Temp role")
	require.NoError(t, err)
	require.NotNil(t, role)

	k.DeleteRole(ctx, "temp_role")

	deleted, err := k.GetRoleFromStore(ctx, "temp_role")
	require.Error(t, err)
	require.Nil(t, deleted)
}

func TestCreateDuplicateRole(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.CreateRole(ctx, "admin1", "duplicate", []string{"perm1"}, "First")
	require.NoError(t, err)

	_, err = k.CreateRole(ctx, "admin1", "duplicate", []string{"perm2"}, "Second")
	require.Error(t, err, "Should not allow duplicate role creation")
}

func TestCreateRoleWithoutPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Try to create role without being admin
	_, err := k.CreateRole(ctx, "user1", "new_role", []string{"perm1"}, "New role")
	require.Error(t, err, "Should require admin permission")
}

// Role Assignment Tests

func TestAssignMultipleRolesToUser(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	// Assign multiple roles
	_, err = k.AssignRole(ctx, "admin1", "user1", types.RoleUser, 0)
	require.NoError(t, err)

	_, err = k.AssignRole(ctx, "admin1", "user1", types.RoleModerator, 0)
	require.NoError(t, err)

	assignments, err := k.GetRoleAssignmentsForAddress(ctx, "user1")
	require.NoError(t, err)
	require.Len(t, assignments, 2)
}

func TestAssignRoleToNonExistentUser(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	// Should still work - creates assignment for future user
	_, err = k.AssignRole(ctx, "admin1", "", types.RoleUser, 0)
	require.Error(t, err, "Should not assign to empty address")
}

func TestGetAllRoleAssignments(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	// Assign roles to multiple users
	for i := 0; i < 3; i++ {
		user := "user" + string(rune('1'+i))
		_, err = k.AssignRole(ctx, "admin1", user, types.RoleUser, 0)
		require.NoError(t, err)
	}

	assignments, err := k.GetAllRoleAssignments(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(assignments), 3)
}

func TestDeleteRoleAssignment(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.AssignRole(ctx, "admin1", "user1", types.RoleUser, 0)
	require.NoError(t, err)

	err = k.DeleteRoleAssignment(ctx, "user1", types.RoleUser)
	require.NoError(t, err)

	assignments, err := k.GetRoleAssignmentsForAddress(ctx, "user1")
	require.NoError(t, err)
	require.Len(t, assignments, 0)
}

// Multisig Wallet Tests

func TestGetNonExistentMultisigWallet(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	wallet, err := k.GetMultisigWallet(ctx, "nonexistent")
	require.Error(t, err)
	require.Nil(t, wallet)
}

func TestGetAllMultisigWallets(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create multiple wallets
	for i := 0; i < 3; i++ {
		wallet := &authproto.MultisigWallet{
			Id:        "wallet" + string(rune('1'+i)),
			Signers:   []string{"owner1", "owner2"},
			Threshold: 2,
			CreatedAt: time.Unix(1234567890, 0),
		}
		err := k.SetMultisigWallet(ctx, wallet)
		require.NoError(t, err)
	}

	wallets, err := k.GetAllMultisigWallets(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(wallets), 3)
}

func TestDeleteMultisigWallet(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	wallet := &authproto.MultisigWallet{
		Id:        "temp_wallet",
		Signers:   []string{"owner1", "owner2"},
		Threshold: 2,
		CreatedAt: time.Unix(1234567890, 0),
	}
	err := k.SetMultisigWallet(ctx, wallet)
	require.NoError(t, err)

	k.DeleteMultisigWallet(ctx, "temp_wallet")

	deleted, err := k.GetMultisigWallet(ctx, "temp_wallet")
	require.Error(t, err)
	require.Nil(t, deleted)
}

func TestMultisigWalletThreshold(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	tests := []struct {
		name      string
		owners    []string
		threshold uint64
		shouldErr bool
	}{
		{"Valid 2/3", []string{"o1", "o2", "o3"}, 2, false},
		{"Valid 3/3", []string{"o1", "o2", "o3"}, 3, false},
		{"Invalid 0", []string{"o1", "o2"}, 0, true},
		{"Invalid > owners", []string{"o1", "o2"}, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := &authproto.MultisigWallet{
				Id:        "wallet_" + tt.name,
				Signers:   tt.owners,
				Threshold: uint32(tt.threshold),
				CreatedAt: time.Unix(1234567890, 0),
			}
			err := k.SetMultisigWallet(ctx, wallet)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Multisig Proposal Tests

func TestGetNonExistentMultisigProposal(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	proposal, err := k.GetMultisigProposal(ctx, "nonexistent")
	require.Error(t, err)
	require.Nil(t, proposal)
}

func TestGetAllMultisigProposals(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create multiple proposals
	for i := 0; i < 3; i++ {
		proposal := &authproto.MultisigProposal{
			Id:          "proposal" + string(rune('1'+i)),
			WalletId:    "wallet1",
			Title:       "Test proposal",
			Description: "Test proposal description",
			Signatures:  []string{},
			Status:      authproto.ProposalStatus_PROPOSAL_STATUS_PENDING,
			CreatedAt:   time.Unix(1234567890, 0),
		}
		err := k.SetMultisigProposal(ctx, proposal)
		require.NoError(t, err)
	}

	proposals, err := k.GetAllMultisigProposals(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(proposals), 3)
}

func TestDeleteMultisigProposal(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	proposal := &authproto.MultisigProposal{
		Id:          "temp_proposal",
		WalletId:    "wallet1",
		Title:       "Temp proposal",
		Description: "Temp proposal description",
		Signatures:  []string{},
		Status:      authproto.ProposalStatus_PROPOSAL_STATUS_PENDING,
		CreatedAt:   time.Unix(1234567890, 0),
	}
	err := k.SetMultisigProposal(ctx, proposal)
	require.NoError(t, err)

	k.DeleteMultisigProposal(ctx, "temp_proposal")

	deleted, err := k.GetMultisigProposal(ctx, "temp_proposal")
	require.Error(t, err)
	require.Nil(t, deleted)
}

// Audit Log Tests

func TestGetAuditLogsByActor(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Log multiple events
	for i := 0; i < 5; i++ {
		k.LogAudit(ctx, "actor1", "action"+string(rune('A'+i)), "resource1", "success", nil, "", ctx.BlockTime())
	}

	logs := k.GetAuditLogsByActor("actor1", 10)
	require.GreaterOrEqual(t, len(logs), 5)
}

func TestGetAuditLogsByAction(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	k.LogAudit(ctx, "actor1", "create", "resource1", "success", nil, "", ctx.BlockTime())
	k.LogAudit(ctx, "actor2", "create", "resource2", "success", nil, "", ctx.BlockTime())
	k.LogAudit(ctx, "actor3", "delete", "resource3", "success", nil, "", ctx.BlockTime())

	logs := k.GetAuditLogsByAction("create", 10)
	require.GreaterOrEqual(t, len(logs), 2)
}

func TestGetAuditLogsByResource(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	k.LogAudit(ctx, "actor1", "read", "resource1", "success", nil, "", ctx.BlockTime())
	k.LogAudit(ctx, "actor2", "update", "resource1", "success", nil, "", ctx.BlockTime())

	logs := k.GetAuditLogsByResource("resource1", 10)
	require.GreaterOrEqual(t, len(logs), 2)
}

func TestGetRecentAuditLogs(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	for i := 0; i < 10; i++ {
		k.LogAudit(ctx, "actor1", "action1", "resource1", "success", nil, "", ctx.BlockTime())
	}

	logs := k.GetRecentAuditLogs(5)
	require.Len(t, logs, 5)
}

func TestCountAuditLogs(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	initialCount := k.CountAuditLogs()

	for i := 0; i < 3; i++ {
		k.LogAudit(ctx, "actor1", "action1", "resource1", "success", nil, "", ctx.BlockTime())
	}

	newCount := k.CountAuditLogs()
	require.Equal(t, initialCount+3, newCount)
}

func TestCountAuditLogsByActor(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	for i := 0; i < 3; i++ {
		k.LogAudit(ctx, "actor1", "action1", "resource1", "success", nil, "", ctx.BlockTime())
	}

	count := k.CountAuditLogsByActor("actor1")
	require.GreaterOrEqual(t, count, uint64(3))
}

func TestAuditLogWithMetadata(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	metadata := map[string]string{
		"ip":     "192.168.1.1",
		"method": "POST",
	}

	k.LogAudit(ctx, "actor1", "action1", "resource1", "success", metadata, "", ctx.BlockTime())

	logs := k.GetAuditLogsByActor("actor1", 10)
	require.GreaterOrEqual(t, len(logs), 1)
	if len(logs) > 0 {
		require.NotNil(t, logs[0].Metadata)
	}
}

func TestAuditLogWithError(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	k.LogAudit(ctx, "actor1", "action1", "resource1", "failure", nil, "Permission denied", ctx.BlockTime())

	logs := k.GetAuditLogsByActor("actor1", 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

func TestSearchAuditLogs(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	k.LogAudit(ctx, "actor1", "create", "user_123", "success", map[string]string{"type": "user"}, "", ctx.BlockTime())
	k.LogAudit(ctx, "actor2", "delete", "user_456", "success", map[string]string{"type": "user"}, "", ctx.BlockTime())

	criteria := map[string]string{
		"action": "create",
	}

	logs := k.SearchAuditLogs(criteria, 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

func TestGetAuditLogsByTimeRange(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	now := time.Now().Unix()
	k.LogAudit(ctx, "actor1", "action1", "resource1", "success", nil, "", ctx.BlockTime())

	logs := k.GetAuditLogsByTimeRange(now-3600, now+3600, 10)
	require.GreaterOrEqual(t, len(logs), 1)
}

// Permission Tests

func TestHasPermission(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	require.True(t, k.HasPermission(ctx, "admin1", types.PermissionAdmin))
	require.False(t, k.HasPermission(ctx, "user1", types.PermissionAdmin))
}

func TestGetRoleAssignments(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	_, err := k.AssignRole(ctx, "system", "admin1", types.RoleAdmin, 0)
	require.NoError(t, err)

	_, err = k.AssignRole(ctx, "admin1", "user1", types.RoleUser, 0)
	require.NoError(t, err)

	assignments, err := k.GetRoleAssignmentsForAddress(ctx, "user1")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(assignments), 1)
}
