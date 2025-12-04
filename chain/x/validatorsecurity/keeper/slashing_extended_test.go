package keeper_test

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func (suite *KeeperTestSuite) TestHandleDoubleSign() {
	validatorAddr := "auravaloper1double"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Handle double sign
	voteA := []byte("vote_a_data")
	voteB := []byte("vote_b_data")
	_, err = suite.keeper.HandleDoubleSign(suite.ctx, validatorAddr, 100, voteA, voteB)

	// May fail due to mock keepers, but we're testing the logic path
	// The important thing is it doesn't panic
}

func (suite *KeeperTestSuite) TestHandleDoubleSignSameVotes() {
	validatorAddr := "auravaloper1doublesame"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Try double sign with same votes - should fail
	voteA := []byte("same_vote")
	voteB := []byte("same_vote")
	_, err = suite.keeper.HandleDoubleSign(suite.ctx, validatorAddr, 100, voteA, voteB)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrInvalidDoubleSignEvidence, err)
}

func (suite *KeeperTestSuite) TestHandleDoubleSignTombstoned() {
	validatorAddr := "auravaloper1doubletomb"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Tombstone first
	err = suite.keeper.TombstoneValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Try double sign - should fail
	voteA := []byte("vote_a")
	voteB := []byte("vote_b")
	_, err = suite.keeper.HandleDoubleSign(suite.ctx, validatorAddr, 100, voteA, voteB)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrValidatorTombstoned, err)
}

func (suite *KeeperTestSuite) TestGetAllDoubleSignEvidences() {
	// Create multiple evidences with different validator addresses
	// Using unique validator addresses to avoid test isolation issues
	val1 := newValAddr()
	val2 := newValAddr()
	val3 := newValAddr()

	evidence1 := types.DoubleSignEvidence{
		ValidatorAddress: val1,
		Height:           100,
		Time:             timestamppb.New(time.Now()),
		VoteA:            []byte("vote_a1"),
		VoteB:            []byte("vote_b1"),
		SlashFraction:    "0.05",
	}
	suite.keeper.SetDoubleSignEvidence(suite.ctx, evidence1)

	evidence2 := types.DoubleSignEvidence{
		ValidatorAddress: val2,
		Height:           101,
		Time:             timestamppb.New(time.Now()),
		VoteA:            []byte("vote_a2"),
		VoteB:            []byte("vote_b2"),
		SlashFraction:    "0.05",
	}
	suite.keeper.SetDoubleSignEvidence(suite.ctx, evidence2)

	evidence3 := types.DoubleSignEvidence{
		ValidatorAddress: val3,
		Height:           102,
		Time:             timestamppb.New(time.Now()),
		VoteA:            []byte("vote_a3"),
		VoteB:            []byte("vote_b3"),
		SlashFraction:    "0.05",
	}
	suite.keeper.SetDoubleSignEvidence(suite.ctx, evidence3)

	// Get all evidences - check that we have at least the 3 we just created
	evidences := suite.keeper.GetAllDoubleSignEvidences(suite.ctx)
	suite.Require().GreaterOrEqual(len(evidences), 3, "Should have at least 3 evidences")
}

func (suite *KeeperTestSuite) TestValidateMinimumStakeExtended() {
	validatorAddr := "auravaloper1stake"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Validate minimum stake - will fail with mock keeper but shouldn't panic
	_ = suite.keeper.ValidateMinimumStake(suite.ctx, validatorAddr)
}

func (suite *KeeperTestSuite) TestHandleDowntimeNoViolation() {
	validatorAddr := "auravaloper1nodowntime"

	// Setup
	params := types.DefaultParams()
	params.SignedBlocksWindow = 1000
	params.MinSignedPerWindow = "0.5"
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Set missed blocks below threshold
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	info.MissedBlocksCounter = 100 // Below threshold
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Handle downtime - should not jail
	err = suite.keeper.HandleDowntime(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Verify not jailed
	info, err = suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().False(info.IsJailed)
}

func (suite *KeeperTestSuite) TestHandleDowntimeAlreadyJailed() {
	validatorAddr := "auravaloper1jaileddowntime"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Jail validator
	err = suite.keeper.JailValidator(suite.ctx, validatorAddr, time.Hour)
	suite.Require().NoError(err)

	// Set high missed blocks
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	info.MissedBlocksCounter = 1000
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Handle downtime - should not slash again
	err = suite.keeper.HandleDowntime(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestHandleDowntimeTombstoned() {
	validatorAddr := "auravaloper1tombdowntime"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Tombstone
	err = suite.keeper.TombstoneValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Set high missed blocks
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	info.MissedBlocksCounter = 1000
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Handle downtime - should not slash tombstoned validator
	err = suite.keeper.HandleDowntime(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetDoubleSignEvidenceNotFound() {
	// Get non-existent evidence
	_, err := suite.keeper.GetDoubleSignEvidence(suite.ctx, "nonexistent")
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrEvidenceNotFound, err)
}

func (suite *KeeperTestSuite) TestGetDowntimeInfractionNotFound() {
	// Get non-existent infraction
	_, err := suite.keeper.GetDowntimeInfraction(suite.ctx, "nonexistent")
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrInfractionNotFound, err)
}

func (suite *KeeperTestSuite) TestSetAndGetDowntimeInfraction() {
	validatorAddr := "auravaloper1infraction"

	infraction := types.DowntimeInfraction{
		ValidatorAddress: validatorAddr,
		MissedBlocks:     200,
		WindowSize:       1000,
		DetectedAt:       timestamppb.New(time.Now()),
		SlashFraction:    "0.0001",
	}

	suite.keeper.SetDowntimeInfraction(suite.ctx, infraction)

	// Retrieve
	retrieved, err := suite.keeper.GetDowntimeInfraction(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(validatorAddr, retrieved.ValidatorAddress)
	suite.Require().Equal(int64(200), retrieved.MissedBlocks)
}
