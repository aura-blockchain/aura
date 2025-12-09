package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// TestGetAuthority tests the GetAuthority function
func TestGetAuthority(t *testing.T) {
	keeper, _ := setupKeeperForTest(t)

	authority := keeper.GetAuthority()
	require.Equal(t, "authority", authority, "authority should match initialized value")
}

// TestLogger tests the Logger function
func TestLogger(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	logger := keeper.Logger(ctx)
	require.NotNil(t, logger, "logger should not be nil")
}

// TestGetAllDataForPrefix_Empty tests GetAllDataForPrefix with no data
func TestGetAllDataForPrefix_Empty(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	items, err := keeper.GetAllDataForPrefix(ctx, types.RolePrefix)
	require.NoError(t, err)
	require.Empty(t, items, "should return empty slice when no data exists")
}

// TestGetAllDataForPrefix_WithData tests GetAllDataForPrefix with data
func TestGetAllDataForPrefix_WithData(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create multiple roles
	roles := []*types.Role{
		{
			Name:         "role1",
			Permissions:  []string{types.PermissionAdmin},
			Description:  "First role",
			CreatedAt:    ctx.BlockTime(),
			IsSystemRole: false,
			UpdatedAt:    func() *time.Time { t := ctx.BlockTime(); return &t }(),
		},
		{
			Name:         "role2",
			Permissions:  []string{types.PermissionViewAuditLogs},
			Description:  "Second role",
			CreatedAt:    ctx.BlockTime(),
			IsSystemRole: false,
			UpdatedAt:    func() *time.Time { t := ctx.BlockTime(); return &t }(),
		},
		{
			Name:         "role3",
			Permissions:  []string{types.PermissionManageIdentity},
			Description:  "Third role",
			CreatedAt:    ctx.BlockTime(),
			IsSystemRole: false,
			UpdatedAt:    func() *time.Time { t := ctx.BlockTime(); return &t }(),
		},
	}

	// Store roles
	for _, role := range roles {
		err := keeper.SetRole(ctx, role)
		require.NoError(t, err)
	}

	// Retrieve all roles using prefix
	items, err := keeper.GetAllDataForPrefix(ctx, types.RolePrefix)
	require.NoError(t, err)
	require.Len(t, items, 3, "should retrieve all three roles")

	// Verify each item is valid protobuf data
	for _, item := range items {
		var role types.Role
		err := keeper.cdc.Unmarshal(item, &role)
		require.NoError(t, err, "should be able to unmarshal role data")
		require.NotEmpty(t, role.Name, "role name should not be empty")
	}
}

// TestGetAllDataForPrefix_DifferentPrefixes tests that prefix filtering works correctly
func TestGetAllDataForPrefix_DifferentPrefixes(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create a role
	role := &types.Role{
		Name:         "test_role",
		Permissions:  []string{types.PermissionAdmin},
		Description:  "Test role",
		CreatedAt:    ctx.BlockTime(),
		IsSystemRole: false,
		UpdatedAt:    func() *time.Time { t := ctx.BlockTime(); return &t }(),
	}
	err := keeper.SetRole(ctx, role)
	require.NoError(t, err)

	// Create an identity record
	identity := &types.IdentityRecord{
		Did:       "did:aura:test123",
		Address:   "aura1test",
		Status:    types.IdentityStatusActive,
		CreatedAt: ctx.BlockTime(),
		UpdatedAt: func() *time.Time { t := ctx.BlockTime(); return &t }(),
	}
	err = keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Get roles - should only return role, not identity
	roleItems, err := keeper.GetAllDataForPrefix(ctx, types.RolePrefix)
	require.NoError(t, err)
	require.Len(t, roleItems, 1, "should only retrieve roles")

	// Get identities - should only return identity, not role
	identityItems, err := keeper.GetAllDataForPrefix(ctx, types.IdentityRecordPrefix)
	require.NoError(t, err)
	require.Len(t, identityItems, 1, "should only retrieve identities")
}

// TestGetParams_Default tests GetParams when no params are set
func TestGetParams_Default(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	require.NotNil(t, params, "should return default params when none set")

	// Verify default params structure
	defaultParams := types.DefaultParams()
	require.NotNil(t, params.Auth)
	require.NotNil(t, params.Change)
	require.Equal(t, defaultParams.Auth.EnableRbac, params.Auth.EnableRbac)
	require.Equal(t, defaultParams.Auth.MaxRolesPerAccount, params.Auth.MaxRolesPerAccount)
}

// TestGetParams_AfterSet tests GetParams after SetParams
func TestGetParams_AfterSet(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create custom params
	customParams := types.DefaultParams()
	customParams.Auth.EnableRbac = false
	customParams.Auth.MaxRolesPerAccount = 20
	customParams.Change.MaxRequestsPerWalletPerMonth = 50
	customParams.Change.MinConfidenceAfterChange = 75

	// Set params
	err := keeper.SetParams(ctx, customParams)
	require.NoError(t, err)

	// Get params
	retrievedParams, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	require.NotNil(t, retrievedParams)
	require.Equal(t, customParams.Auth.EnableRbac, retrievedParams.Auth.EnableRbac)
	require.Equal(t, customParams.Auth.MaxRolesPerAccount, retrievedParams.Auth.MaxRolesPerAccount)
	require.Equal(t, customParams.Change.MaxRequestsPerWalletPerMonth, retrievedParams.Change.MaxRequestsPerWalletPerMonth)
	require.Equal(t, customParams.Change.MinConfidenceAfterChange, retrievedParams.Change.MinConfidenceAfterChange)
}

// TestSetParams_Nil tests SetParams with nil params
func TestSetParams_Nil(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	err := keeper.SetParams(ctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "params cannot be nil")
}

// TestSetParams_Update tests updating params
func TestSetParams_Update(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set initial params
	initialParams := types.DefaultParams()
	err := keeper.SetParams(ctx, initialParams)
	require.NoError(t, err)

	// Update params
	updatedParams := types.DefaultParams()
	updatedParams.Auth.EnableRbac = false
	updatedParams.Auth.DefaultTimelockDelaySeconds = 7200
	updatedParams.Change.MaxRequestsPerWalletPerMonth = 100
	updatedParams.Change.KeyRotationGracePeriodSeconds = 86400

	err = keeper.SetParams(ctx, updatedParams)
	require.NoError(t, err)

	// Verify update
	retrievedParams, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, updatedParams.Auth.EnableRbac, retrievedParams.Auth.EnableRbac)
	require.Equal(t, updatedParams.Auth.DefaultTimelockDelaySeconds, retrievedParams.Auth.DefaultTimelockDelaySeconds)
	require.Equal(t, updatedParams.Change.MaxRequestsPerWalletPerMonth, retrievedParams.Change.MaxRequestsPerWalletPerMonth)
	require.Equal(t, updatedParams.Change.KeyRotationGracePeriodSeconds, retrievedParams.Change.KeyRotationGracePeriodSeconds)
}
