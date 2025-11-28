package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

type GenesisTestSuite struct {
	KeeperTestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	ctx := suite.SdkCtx

	// Test: InitGenesis with default/empty state should not panic
	suite.Run("default genesis", func() {
		defaultGenesis := types.DefaultGenesis()
		err := suite.keeper.InitGenesis(ctx, defaultGenesis)
		suite.NoError(err, "InitGenesis should not error with default state")
	})

	// Test: InitGenesis with valid data
	suite.Run("valid genesis with data", func() {
		genesis := &authproto.GenesisState{
			Params: types.DefaultParams(),
			EmergencyAdmins: []*authproto.EmergencyAdmin{
				{
					Address:      "aura1test1",
					GrantedBy:    "aura1granter",
					GrantedAt:    1000,
					ExpiresAt:    2000,
					IsActive:     true,
					Permissions:  []string{"admin.emergency.pause"},
				},
			},
			EmergencyActions: []*authproto.EmergencyAction{
				{
					ActionId:      "action1",
					ActionType:    "pause",
					InitiatedBy:   "aura1test1",
					InitiatedAt:   1100,
					Status:        "pending",
					Description:   "Test action",
				},
			},
			PermissionGrants: []*authproto.PermissionGrant{
				{
					Grantee:    "aura1test2",
					Permission: "test.permission",
					GrantedBy:  "aura1granter",
					GrantedAt:  1200,
					ExpiresAt:  2200,
				},
			},
			Roles: []*authproto.Role{
				{
					RoleId:      "role1",
					Name:        "Test Role",
					Permissions: []string{"read", "write"},
					CreatedAt:   1000,
				},
			},
			RoleAssignments: []*authproto.RoleAssignment{
				{
					Address:    "aura1test3",
					RoleId:     "role1",
					AssignedBy: "aura1granter",
					AssignedAt: 1300,
					ExpiresAt:  2300,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should not error with valid data")

		// Verify data was stored
		admin, found := suite.keeper.GetEmergencyAdmin(ctx, "aura1test1")
		suite.True(found, "Emergency admin should be found")
		suite.Equal("aura1test1", admin.Address)
		suite.True(admin.IsActive)
	})
}

func (suite *GenesisTestSuite) TestInitGenesisWithInvalidData() {
	ctx := suite.SdkCtx

	suite.Run("nil genesis", func() {
		err := suite.keeper.InitGenesis(ctx, nil)
		suite.Error(err, "InitGenesis should error with nil state")
	})

	suite.Run("nil params", func() {
		genesis := &authproto.GenesisState{
			Params: nil,
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with nil params")
	})

	suite.Run("nil emergency admin", func() {
		genesis := &authproto.GenesisState{
			Params: types.DefaultParams(),
			EmergencyAdmins: []*authproto.EmergencyAdmin{
				nil,
			},
		}
		// Should skip nil admins without error
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should skip nil admins")
	})

	suite.Run("nil emergency action", func() {
		genesis := &authproto.GenesisState{
			Params: types.DefaultParams(),
			EmergencyActions: []*authproto.EmergencyAction{
				nil,
			},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should skip nil actions")
	})

	suite.Run("nil permission grant", func() {
		genesis := &authproto.GenesisState{
			Params: types.DefaultParams(),
			PermissionGrants: []*authproto.PermissionGrant{
				nil,
			},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should skip nil grants")
	})
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	ctx := suite.SdkCtx

	suite.Run("export empty state", func() {
		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported, "ExportGenesis should not return nil")
		suite.NotNil(exported.Params, "Exported params should not be nil")
		suite.NotNil(exported.EmergencyAdmins, "Exported admins should not be nil")
		suite.NotNil(exported.EmergencyActions, "Exported actions should not be nil")
		suite.NotNil(exported.PermissionGrants, "Exported grants should not be nil")
	})

	suite.Run("export with data", func() {
		// Initialize with some data
		genesis := &authproto.GenesisState{
			Params: types.DefaultParams(),
			EmergencyAdmins: []*authproto.EmergencyAdmin{
				{
					Address:     "aura1test1",
					GrantedBy:   "aura1granter",
					GrantedAt:   1000,
					ExpiresAt:   2000,
					IsActive:    true,
					Permissions: []string{"admin.emergency.pause"},
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Export and verify
		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported)
		suite.Len(exported.EmergencyAdmins, 1, "Should export 1 admin")
		suite.Equal("aura1test1", exported.EmergencyAdmins[0].Address)
	})
}

func (suite *GenesisTestSuite) TestDefaultGenesis() {
	suite.Run("default genesis is valid", func() {
		defaultGenesis := types.DefaultGenesis()
		suite.NotNil(defaultGenesis, "DefaultGenesis should not return nil")

		err := types.ValidateGenesis(defaultGenesis)
		suite.NoError(err, "Default genesis should be valid")

		suite.NotNil(defaultGenesis.Params, "Default params should not be nil")
		suite.NotNil(defaultGenesis.EmergencyAdmins, "Default admins should not be nil")
		suite.Empty(defaultGenesis.EmergencyAdmins, "Default admins should be empty")
		suite.NotNil(defaultGenesis.EmergencyActions, "Default actions should not be nil")
		suite.Empty(defaultGenesis.EmergencyActions, "Default actions should be empty")
	})

	suite.Run("default genesis can be initialized", func() {
		ctx := suite.SdkCtx
		defaultGenesis := types.DefaultGenesis()

		err := suite.keeper.InitGenesis(ctx, defaultGenesis)
		suite.NoError(err, "Should be able to initialize default genesis")
	})
}

func (suite *GenesisTestSuite) TestGenesisRoundTrip() {
	ctx := suite.SdkCtx

	suite.Run("round trip with empty state", func() {
		genesis := types.DefaultGenesis()

		// Initialize
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Export
		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported)

		// Verify consistency
		suite.Equal(len(genesis.EmergencyAdmins), len(exported.EmergencyAdmins))
		suite.Equal(len(genesis.EmergencyActions), len(exported.EmergencyActions))
		suite.Equal(len(genesis.PermissionGrants), len(exported.PermissionGrants))
	})

	suite.Run("round trip with data", func() {
		genesis := &authproto.GenesisState{
			Params: types.DefaultParams(),
			EmergencyAdmins: []*authproto.EmergencyAdmin{
				{
					Address:     "aura1test1",
					GrantedBy:   "aura1granter",
					GrantedAt:   1000,
					ExpiresAt:   2000,
					IsActive:    true,
					Permissions: []string{"admin.emergency.pause"},
				},
				{
					Address:     "aura1test2",
					GrantedBy:   "aura1granter",
					GrantedAt:   1100,
					ExpiresAt:   2100,
					IsActive:    false,
					Permissions: []string{"admin.emergency.halt"},
				},
			},
			EmergencyActions: []*authproto.EmergencyAction{
				{
					ActionId:    "action1",
					ActionType:  "pause",
					InitiatedBy: "aura1test1",
					InitiatedAt: 1200,
					Status:      "pending",
				},
			},
			PermissionGrants: []*authproto.PermissionGrant{
				{
					Grantee:    "aura1test3",
					Permission: "test.permission",
					GrantedBy:  "aura1granter",
					GrantedAt:  1300,
					ExpiresAt:  2300,
				},
			},
		}

		// Initialize
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Export
		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported)

		// Verify all data was preserved
		suite.Equal(len(genesis.EmergencyAdmins), len(exported.EmergencyAdmins))
		suite.Equal(len(genesis.EmergencyActions), len(exported.EmergencyActions))
		suite.Equal(len(genesis.PermissionGrants), len(exported.PermissionGrants))

		// Re-initialize with exported data
		ctx2 := suite.SdkCtx
		err = suite.keeper.InitGenesis(ctx2, exported)
		suite.NoError(err)

		// Export again
		exported2 := suite.keeper.ExportGenesis(ctx2)
		suite.NotNil(exported2)

		// Should be identical
		suite.Equal(len(exported.EmergencyAdmins), len(exported2.EmergencyAdmins))
		suite.Equal(len(exported.EmergencyActions), len(exported2.EmergencyActions))
		suite.Equal(len(exported.PermissionGrants), len(exported2.PermissionGrants))
	})
}

func (suite *GenesisTestSuite) TestGenesisEdgeCases() {
	ctx := suite.SdkCtx

	suite.Run("empty lists", func() {
		genesis := &authproto.GenesisState{
			Params:           types.DefaultParams(),
			EmergencyAdmins:  []*authproto.EmergencyAdmin{},
			EmergencyActions: []*authproto.EmergencyAction{},
			PermissionGrants: []*authproto.PermissionGrant{},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported)
		suite.NotNil(exported.EmergencyAdmins)
		suite.NotNil(exported.EmergencyActions)
		suite.NotNil(exported.PermissionGrants)
	})

	suite.Run("many admins", func() {
		genesis := &authproto.GenesisState{
			Params:          types.DefaultParams(),
			EmergencyAdmins: make([]*authproto.EmergencyAdmin, 100),
		}

		for i := 0; i < 100; i++ {
			genesis.EmergencyAdmins[i] = &authproto.EmergencyAdmin{
				Address:     "aura1test" + string(rune(i)),
				GrantedBy:   "aura1granter",
				GrantedAt:   uint64(1000 + i),
				ExpiresAt:   uint64(2000 + i),
				IsActive:    i%2 == 0,
				Permissions: []string{"perm"},
			}
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.Len(exported.EmergencyAdmins, 100)
	})
}

// Test helper function
func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name      string
		genesis   *authproto.GenesisState
		expectErr bool
	}{
		{
			name:      "nil genesis",
			genesis:   nil,
			expectErr: true,
		},
		{
			name: "nil params",
			genesis: &authproto.GenesisState{
				Params: nil,
			},
			expectErr: true,
		},
		{
			name:      "valid default genesis",
			genesis:   types.DefaultGenesis(),
			expectErr: false,
		},
		{
			name: "valid genesis with data",
			genesis: &authproto.GenesisState{
				Params: types.DefaultParams(),
				EmergencyAdmins: []*authproto.EmergencyAdmin{
					{
						Address:     "aura1test",
						GrantedBy:   "aura1granter",
						GrantedAt:   1000,
						ExpiresAt:   2000,
						IsActive:    true,
						Permissions: []string{"admin.emergency.pause"},
					},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateGenesis(tt.genesis)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
