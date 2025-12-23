package keeper

import (
	"testing"
	"time"

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

func (suite *GenesisTestSuite) TestInitGenesis_Default() {
	ctx := suite.SdkCtx

	// Default genesis should initialize cleanly
	defaultGenesis := types.DefaultGenesis()
	err := suite.Keeper.InitGenesis(ctx, defaultGenesis)
	suite.NoError(err)

	// Verify params were set
	params, err := suite.Keeper.GetParams(ctx)
	suite.NoError(err)
	suite.NotNil(params)
}

func (suite *GenesisTestSuite) TestInitGenesis_WithData() {
	ctx := suite.SdkCtx

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	genesis := &types.GenesisState{
		Params: *types.DefaultParams(),
		Roles: []authproto.Role{
			{
				Name:        "admin",
				Permissions: []string{"*"},
				Description: "Administrator role",
			},
			{
				Name:        "user",
				Permissions: []string{"read"},
				Description: "User role",
			},
		},
		RoleAssignments: []authproto.RoleAssignment{
			{
				Address:  "aura1test1",
				RoleName: "admin",
			},
		},
		MultisigWallets: []authproto.MultisigWallet{
			{
				Id:        "wallet-1",
				Signers:   []string{"aura1test1", "aura1test2"},
				Threshold: 2,
			},
		},
		Sessions: []authproto.Session{
			{
				Id:        "session-1",
				Address:   "aura1test1",
				CreatedAt: now,
				ExpiresAt: expiresAt,
			},
		},
	}

	err := suite.Keeper.InitGenesis(ctx, genesis)
	suite.NoError(err)

	// Verify data was imported
	roles, err := suite.Keeper.GetAllRoles(ctx)
	suite.NoError(err)
	suite.Len(roles, 2)

	assignments, err := suite.Keeper.GetAllRoleAssignments(ctx)
	suite.NoError(err)
	suite.Len(assignments, 1)

	wallets, err := suite.Keeper.GetAllMultisigWallets(ctx)
	suite.NoError(err)
	suite.Len(wallets, 1)

	sessions, err := suite.Keeper.GetAllSessions(ctx)
	suite.NoError(err)
	suite.Len(sessions, 1)
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	ctx := suite.SdkCtx

	// Initialize with default genesis
	defaultGenesis := types.DefaultGenesis()
	err := suite.Keeper.InitGenesis(ctx, defaultGenesis)
	suite.NoError(err)

	// Export genesis
	exported := suite.Keeper.ExportGenesis(ctx)
	suite.NotNil(exported)
	suite.NotNil(exported.Params)
}

func (suite *GenesisTestSuite) TestGenesisRoundTrip() {
	ctx := suite.SdkCtx

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	executableAt := now.Add(48 * time.Hour)

	// Create a comprehensive genesis state
	genesis := &types.GenesisState{
		Params: *types.DefaultParams(),
		Roles: []authproto.Role{
			{
				Name:        "admin",
				Permissions: []string{"*"},
				Description: "Administrator role",
			},
			{
				Name:        "moderator",
				Permissions: []string{"read", "write"},
				Description: "Moderator role",
			},
			{
				Name:        "user",
				Permissions: []string{"read"},
				Description: "Basic user role",
			},
		},
		RoleAssignments: []authproto.RoleAssignment{
			{
				Address:  "aura1test1",
				RoleName: "admin",
			},
			{
				Address:  "aura1test2",
				RoleName: "moderator",
			},
		},
		MultisigWallets: []authproto.MultisigWallet{
			{
				Id:        "wallet-1",
				Signers:   []string{"aura1test1", "aura1test2", "aura1test3"},
				Threshold: 2,
			},
			{
				Id:        "wallet-2",
				Signers:   []string{"aura1test4", "aura1test5"},
				Threshold: 2,
			},
		},
		MultisigProposals: []authproto.MultisigProposal{
			{
				Id:       "proposal-1",
				WalletId: "wallet-1",
				Title:    "Test Proposal",
			},
		},
		TimeLockedActions: []authproto.TimeLockedAction{
			{
				Id:           "action-1",
				Proposer:     "aura1test1",
				ProposedAt:   now,
				ExecutableAt: executableAt,
			},
		},
		EmergencyAdmins: []authproto.EmergencyAdmin{
			{
				Address: "aura1emergency",
			},
		},
		Sessions: []authproto.Session{
			{
				Id:        "session-1",
				Address:   "aura1test1",
				CreatedAt: now,
				ExpiresAt: expiresAt,
			},
			{
				Id:        "session-2",
				Address:   "aura1test2",
				CreatedAt: now,
				ExpiresAt: expiresAt,
			},
		},
		RateLimitConfigs: []authproto.RateLimitConfig{
			{
				UserAddress:      "aura1test1",
				RequestsPerMinute: 100,
			},
		},
	}

	// Import genesis
	err := suite.Keeper.InitGenesis(ctx, genesis)
	suite.NoError(err)

	// Export genesis (first export)
	exported1 := suite.Keeper.ExportGenesis(ctx)
	suite.NotNil(exported1)

	// Verify exported data matches original counts
	suite.Equal(len(genesis.Roles), len(exported1.Roles))
	suite.Equal(len(genesis.RoleAssignments), len(exported1.RoleAssignments))
	suite.Equal(len(genesis.MultisigWallets), len(exported1.MultisigWallets))
	suite.Equal(len(genesis.MultisigProposals), len(exported1.MultisigProposals))
	suite.Equal(len(genesis.TimeLockedActions), len(exported1.TimeLockedActions))
	suite.Equal(len(genesis.EmergencyAdmins), len(exported1.EmergencyAdmins))
	suite.Equal(len(genesis.Sessions), len(exported1.Sessions))
	suite.Equal(len(genesis.RateLimitConfigs), len(exported1.RateLimitConfigs))

	// Create a fresh test suite for re-import
	suite.SetupTest()
	ctx2 := suite.SdkCtx

	// Re-import the exported genesis
	err = suite.Keeper.InitGenesis(ctx2, exported1)
	suite.NoError(err)

	// Export again (second export)
	exported2 := suite.Keeper.ExportGenesis(ctx2)
	suite.NotNil(exported2)

	// The two exports should be identical
	suite.Equal(len(exported1.Roles), len(exported2.Roles))
	suite.Equal(len(exported1.RoleAssignments), len(exported2.RoleAssignments))
	suite.Equal(len(exported1.MultisigWallets), len(exported2.MultisigWallets))
	suite.Equal(len(exported1.MultisigProposals), len(exported2.MultisigProposals))
	suite.Equal(len(exported1.TimeLockedActions), len(exported2.TimeLockedActions))
	suite.Equal(len(exported1.EmergencyAdmins), len(exported2.EmergencyAdmins))
	suite.Equal(len(exported1.Sessions), len(exported2.Sessions))
	suite.Equal(len(exported1.RateLimitConfigs), len(exported2.RateLimitConfigs))

	// Verify individual records match
	for i := range exported1.Roles {
		suite.Equal(exported1.Roles[i].Name, exported2.Roles[i].Name)
		suite.Equal(exported1.Roles[i].Description, exported2.Roles[i].Description)
		suite.ElementsMatch(exported1.Roles[i].Permissions, exported2.Roles[i].Permissions)
	}

	for i := range exported1.RoleAssignments {
		suite.Equal(exported1.RoleAssignments[i].Address, exported2.RoleAssignments[i].Address)
		suite.Equal(exported1.RoleAssignments[i].RoleName, exported2.RoleAssignments[i].RoleName)
	}

	for i := range exported1.MultisigWallets {
		suite.Equal(exported1.MultisigWallets[i].Id, exported2.MultisigWallets[i].Id)
		suite.Equal(exported1.MultisigWallets[i].Threshold, exported2.MultisigWallets[i].Threshold)
		suite.ElementsMatch(exported1.MultisigWallets[i].Signers, exported2.MultisigWallets[i].Signers)
	}

	for i := range exported1.Sessions {
		suite.Equal(exported1.Sessions[i].Id, exported2.Sessions[i].Id)
		suite.Equal(exported1.Sessions[i].Address, exported2.Sessions[i].Address)
	}
}

func (suite *GenesisTestSuite) TestGenesisValidation() {
	// Valid default genesis
	suite.NoError(types.ValidateGenesis(types.DefaultGenesis()))

	// Nil genesis should error
	suite.Error(types.ValidateGenesis(nil))
}
