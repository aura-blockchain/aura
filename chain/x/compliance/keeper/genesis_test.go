package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/compliance/types"
	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

type GenesisTestSuite struct {
	KeeperTestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	ctx := suite.SdkCtx

	suite.Run("default genesis", func() {
		defaultGenesis := types.DefaultGenesis()
		err := suite.keeper.InitGenesis(ctx, defaultGenesis)
		suite.NoError(err, "InitGenesis should not error with default state")
	})

	suite.Run("valid genesis with data", func() {
		genesis := &compliancepb.GenesisState{
			Params: types.DefaultParams(),
			KycRecords: []*compliancepb.KYCRecord{
				{
					Address:        "aura1test1",
					Level:          "LEVEL_BASIC",
					Verified:       true,
					VerifiedAt:     1000,
					ExpiresAt:      2000,
					Jurisdiction:   "US",
				},
			},
			AmlProfiles: []*compliancepb.AMLProfile{
				{
					Address:    "aura1test1",
					RiskScore:  50,
					RiskLevel:  "MEDIUM",
					LastUpdate: 1000,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should not error with valid data")
	})
}

func (suite *GenesisTestSuite) TestInitGenesisWithInvalidData() {
	ctx := suite.SdkCtx

	suite.Run("nil genesis", func() {
		err := suite.keeper.InitGenesis(ctx, nil)
		suite.Error(err, "InitGenesis should error with nil state")
	})

	suite.Run("nil params", func() {
		genesis := &compliancepb.GenesisState{
			Params: nil,
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with nil params")
	})

	suite.Run("nil kyc record", func() {
		genesis := &compliancepb.GenesisState{
			Params:     types.DefaultParams(),
			KycRecords: []*compliancepb.KYCRecord{nil},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should skip nil records")
	})
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	ctx := suite.SdkCtx

	suite.Run("export empty state", func() {
		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported, "ExportGenesis should not return nil")
		suite.NotNil(exported.Params, "Exported params should not be nil")
		suite.NotNil(exported.KycRecords, "Exported KYC records should not be nil")
	})

	suite.Run("export with data", func() {
		genesis := &compliancepb.GenesisState{
			Params: types.DefaultParams(),
			KycRecords: []*compliancepb.KYCRecord{
				{
					Address:    "aura1test1",
					Level:      "LEVEL_BASIC",
					Verified:   true,
					VerifiedAt: 1000,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported)
		suite.Len(exported.KycRecords, 1)
	})
}

func (suite *GenesisTestSuite) TestDefaultGenesis() {
	suite.Run("default genesis is valid", func() {
		defaultGenesis := types.DefaultGenesis()
		suite.NotNil(defaultGenesis)

		err := types.ValidateGenesis(defaultGenesis)
		suite.NoError(err, "Default genesis should be valid")
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

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported)
		suite.Equal(len(genesis.KycRecords), len(exported.KycRecords))
	})

	suite.Run("round trip with data", func() {
		genesis := &compliancepb.GenesisState{
			Params: types.DefaultParams(),
			KycRecords: []*compliancepb.KYCRecord{
				{
					Address:    "aura1test1",
					Level:      "LEVEL_BASIC",
					Verified:   true,
					VerifiedAt: 1000,
				},
				{
					Address:    "aura1test2",
					Level:      "LEVEL_ENHANCED",
					Verified:   true,
					VerifiedAt: 1100,
				},
			},
			AmlProfiles: []*compliancepb.AMLProfile{
				{
					Address:    "aura1test1",
					RiskScore:  30,
					RiskLevel:  "LOW",
					LastUpdate: 1000,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.Equal(len(genesis.KycRecords), len(exported.KycRecords))
		suite.Equal(len(genesis.AmlProfiles), len(exported.AmlProfiles))

		err = suite.keeper.InitGenesis(ctx, exported)
		suite.NoError(err)

		exported2 := suite.keeper.ExportGenesis(ctx)
		suite.Equal(len(exported.KycRecords), len(exported2.KycRecords))
	})
}

func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name      string
		genesis   *compliancepb.GenesisState
		expectErr bool
	}{
		{
			name:      "nil genesis",
			genesis:   nil,
			expectErr: true,
		},
		{
			name: "nil params",
			genesis: &compliancepb.GenesisState{
				Params: nil,
			},
			expectErr: true,
		},
		{
			name:      "valid default genesis",
			genesis:   types.DefaultGenesis(),
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
