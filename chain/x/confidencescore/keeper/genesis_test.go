package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

type GenesisTestSuite struct {
	KeeperTestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	suite.Run("default genesis", func() {
		defaultGenesis := types.DefaultGenesisState()
		err := suite.keeper.InitGenesis(*defaultGenesis)
		suite.NoError(err, "InitGenesis should not error with default state")
	})

	suite.Run("valid genesis with data", func() {
		genesis := confidencescorepb.GenesisState{
			Params: types.DefaultParamsProto(),
			UserRecords: []*confidencescorepb.UserConfidenceRecord{
				{
					WalletAddress:     "aura1test1",
					TotalScore:        100,
					CompletedIrs:      []*confidencescorepb.IRCompletion{},
					HasAnchor:         true,
					LastUpdatedHeight: 1000,
					ArenaScores:       map[string]*confidencescorepb.ArenaScore{},
					Status:            types.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
				},
			},
		}

		err := suite.keeper.InitGenesis(genesis)
		suite.NoError(err, "InitGenesis should not error with valid data")

		// Verify data was stored
		record, found := suite.keeper.GetUserRecord("aura1test1")
		suite.True(found, "User record should be found")
		suite.Equal(uint64(100), record.TotalScore)
	})
}

func (suite *GenesisTestSuite) TestInitGenesisWithInvalidData() {
	suite.Run("nil params", func() {
		genesis := confidencescorepb.GenesisState{
			Params: nil,
		}
		err := suite.keeper.InitGenesis(genesis)
		suite.Error(err, "InitGenesis should error with nil params")
	})

	suite.Run("nil user record", func() {
		genesis := confidencescorepb.GenesisState{
			Params:      types.DefaultParamsProto(),
			UserRecords: []*confidencescorepb.UserConfidenceRecord{nil},
		}
		err := suite.keeper.InitGenesis(genesis)
		suite.NoError(err, "InitGenesis should skip nil records")
	})

	suite.Run("invalid user record - empty wallet address", func() {
		genesis := confidencescorepb.GenesisState{
			Params: types.DefaultParamsProto(),
			UserRecords: []*confidencescorepb.UserConfidenceRecord{
				{
					WalletAddress: "", // Invalid
					TotalScore:    100,
				},
			},
		}
		err := suite.keeper.InitGenesis(genesis)
		// Should skip invalid records gracefully
		suite.NoError(err)
	})
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	suite.Run("export empty state", func() {
		exported := suite.keeper.ExportGenesis()
		suite.NotNil(exported.Params, "Exported params should not be nil")
		suite.NotNil(exported.UserRecords, "Exported user records should not be nil")
		suite.NotNil(exported.Completions, "Exported completions should not be nil")
		suite.NotNil(exported.History, "Exported history should not be nil")
	})

	suite.Run("export with data", func() {
		genesis := confidencescorepb.GenesisState{
			Params: types.DefaultParamsProto(),
			UserRecords: []*confidencescorepb.UserConfidenceRecord{
				{
					WalletAddress:     "aura1test1",
					TotalScore:        100,
					CompletedIrs:      []*confidencescorepb.IRCompletion{},
					HasAnchor:         true,
					LastUpdatedHeight: 1000,
					ArenaScores:       map[string]*confidencescorepb.ArenaScore{},
					Status:            types.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
				},
			},
		}

		err := suite.keeper.InitGenesis(genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis()
		suite.NotNil(exported)
		suite.Len(exported.UserRecords, 1, "Should export 1 user record")
		suite.Equal("aura1test1", exported.UserRecords[0].WalletAddress)
	})
}

func (suite *GenesisTestSuite) TestDefaultGenesis() {
	suite.Run("default genesis is valid", func() {
		defaultGenesis := types.DefaultGenesisState()
		suite.NotNil(defaultGenesis)

		err := types.ValidateGenesis(defaultGenesis)
		suite.NoError(err, "Default genesis should be valid")

		suite.NotNil(defaultGenesis.Params)
		suite.NotNil(defaultGenesis.UserRecords)
		suite.Empty(defaultGenesis.UserRecords)
	})

	suite.Run("default genesis can be initialized", func() {
		defaultGenesis := types.DefaultGenesisState()
		err := suite.keeper.InitGenesis(*defaultGenesis)
		suite.NoError(err, "Should be able to initialize default genesis")
	})
}

func (suite *GenesisTestSuite) TestGenesisRoundTrip() {
	suite.Run("round trip with empty state", func() {
		genesis := types.DefaultGenesisState()

		err := suite.keeper.InitGenesis(*genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis()
		suite.Equal(len(genesis.UserRecords), len(exported.UserRecords))
		suite.Equal(len(genesis.Completions), len(exported.Completions))
	})

	suite.Run("round trip with data", func() {
		genesis := confidencescorepb.GenesisState{
			Params: types.DefaultParamsProto(),
			UserRecords: []*confidencescorepb.UserConfidenceRecord{
				{
					WalletAddress:     "aura1test1",
					TotalScore:        100,
					CompletedIrs:      []*confidencescorepb.IRCompletion{},
					HasAnchor:         true,
					LastUpdatedHeight: 1000,
					ArenaScores:       map[string]*confidencescorepb.ArenaScore{},
					Status:            types.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
				},
				{
					WalletAddress:     "aura1test2",
					TotalScore:        200,
					CompletedIrs:      []*confidencescorepb.IRCompletion{},
					HasAnchor:         false,
					LastUpdatedHeight: 1100,
					ArenaScores:       map[string]*confidencescorepb.ArenaScore{},
					Status:            types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
				},
			},
			SlashRecords: []*confidencescorepb.SlashRecord{
				{
					WalletAddress: "aura1test1",
					SlashedAmount: 10,
					Reason:        types.SlashReason_SLASH_REASON_FRAUD_DETECTED,
					SlashTxHash:   "hash1",
					SlashedAt:     1200,
				},
			},
		}

		err := suite.keeper.InitGenesis(genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis()
		suite.Equal(len(genesis.UserRecords), len(exported.UserRecords))
		suite.Equal(len(genesis.SlashRecords), len(exported.SlashRecords))

		err = suite.keeper.InitGenesis(exported)
		suite.NoError(err)

		exported2 := suite.keeper.ExportGenesis()
		suite.Equal(len(exported.UserRecords), len(exported2.UserRecords))
	})
}

func (suite *GenesisTestSuite) TestGenesisEdgeCases() {
	suite.Run("many user records", func() {
		genesis := confidencescorepb.GenesisState{
			Params:      types.DefaultParamsProto(),
			UserRecords: make([]*confidencescorepb.UserConfidenceRecord, 100),
		}

		for i := 0; i < 100; i++ {
			genesis.UserRecords[i] = &confidencescorepb.UserConfidenceRecord{
				WalletAddress:     "aura1test" + string(rune(i)),
				TotalScore:        uint64(100 + i),
				CompletedIrs:      []*confidencescorepb.IRCompletion{},
				HasAnchor:         i%2 == 0,
				LastUpdatedHeight: uint64(1000 + i),
				ArenaScores:       map[string]*confidencescorepb.ArenaScore{},
				Status:            types.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
			}
		}

		err := suite.keeper.InitGenesis(genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis()
		suite.Len(exported.UserRecords, 100)
	})

	suite.Run("user with completions", func() {
		genesis := confidencescorepb.GenesisState{
			Params: types.DefaultParamsProto(),
			UserRecords: []*confidencescorepb.UserConfidenceRecord{
				{
					WalletAddress: "aura1test1",
					TotalScore:    100,
					CompletedIrs: []*confidencescorepb.IRCompletion{
						{
							IrId:       "ir1",
							BaseScore:  50,
							FinalScore: 60,
						},
					},
					HasAnchor:         true,
					LastUpdatedHeight: 1000,
					ArenaScores:       map[string]*confidencescorepb.ArenaScore{},
					Status:            types.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
				},
			},
		}

		err := suite.keeper.InitGenesis(genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis()
		suite.Len(exported.UserRecords, 1)
	})
}

func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name      string
		genesis   *confidencescorepb.GenesisState
		expectErr bool
	}{
		{
			name:      "nil genesis",
			genesis:   nil,
			expectErr: true,
		},
		{
			name: "nil params",
			genesis: &confidencescorepb.GenesisState{
				Params: nil,
			},
			expectErr: false, // Params can be nil, will use defaults
		},
		{
			name:      "valid default genesis",
			genesis:   types.DefaultGenesisState(),
			expectErr: false,
		},
		{
			name: "duplicate wallet address",
			genesis: &confidencescorepb.GenesisState{
				Params: types.DefaultParamsProto(),
				UserRecords: []*confidencescorepb.UserConfidenceRecord{
					{
						WalletAddress: "aura1test",
						TotalScore:    100,
					},
					{
						WalletAddress: "aura1test", // Duplicate
						TotalScore:    200,
					},
				},
			},
			expectErr: true,
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
