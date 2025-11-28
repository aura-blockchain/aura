package keeper_test

import (
	"time"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestJailValidator() {
	validatorAddr := newValAddr()

	// Setup validator
	err := suite.keeper.RegisterValidator(
		suite.ctx,
		validatorAddr,
		"hot",
		"cold",
		"region",
		"US",
		37.0,
		-122.0,
		nil,
	)
	suite.Require().NoError(err)

	// Jail validator
	duration := 24 * time.Hour
	err = suite.keeper.JailValidator(suite.ctx, validatorAddr, duration)
	suite.Require().NoError(err)

	// Verify jailed
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().True(info.IsJailed)
	suite.Require().NotNil(info.JailedUntil)

	// Check IsValidatorJailed
	suite.Require().True(suite.keeper.IsValidatorJailed(suite.ctx, validatorAddr))
}

func (suite *KeeperTestSuite) TestJailValidatorAlreadyJailed() {
	validatorAddr := newValAddr()

	// Setup and jail
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	err = suite.keeper.JailValidator(suite.ctx, validatorAddr, time.Hour)
	suite.Require().NoError(err)

	// Jail again - should succeed without error (already jailed)
	err = suite.keeper.JailValidator(suite.ctx, validatorAddr, time.Hour)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestUnjailValidator() {
	validatorAddr := newValAddr()

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Jail
	err = suite.keeper.JailValidator(suite.ctx, validatorAddr, time.Nanosecond)
	suite.Require().NoError(err)

	// Wait for jail period
	time.Sleep(time.Millisecond * 10)

	// Unjail
	err = suite.keeper.UnjailValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Verify unjailed
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().False(info.IsJailed)
	suite.Require().Equal(int64(0), info.MissedBlocksCounter)
}

func (suite *KeeperTestSuite) TestUnjailValidatorNotJailed() {
	validatorAddr := newValAddr()

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Try to unjail when not jailed
	err = suite.keeper.UnjailValidator(suite.ctx, validatorAddr)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrCannotUnjail, err)
}

func (suite *KeeperTestSuite) TestTombstoneValidator() {
	validatorAddr := newValAddr()

	// Setup
	params := types.DefaultParams()
	params.EnableGeoDistribution = true
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region1", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Tombstone
	err = suite.keeper.TombstoneValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Verify tombstoned
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().True(info.IsTombstoned)
	suite.Require().True(info.IsJailed)
	suite.Require().NotNil(info.TombstonedAt)

	// Check IsValidatorTombstoned
	suite.Require().True(suite.keeper.IsValidatorTombstoned(suite.ctx, validatorAddr))
}

func (suite *KeeperTestSuite) TestTombstoneValidatorAlreadyTombstoned() {
	validatorAddr := "auravaloper1tomb2"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Tombstone
	err = suite.keeper.TombstoneValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Tombstone again - should succeed without error
	err = suite.keeper.TombstoneValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetJailedValidators() {
	// Setup multiple validators
	val1 := "auravaloper1jailed1"
	val2 := "auravaloper1jailed2"
	val3 := "auravaloper1notjailed"

	err := suite.keeper.RegisterValidator(suite.ctx, val1, "hot1", "cold1", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, val2, "hot2", "cold2", "region", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, val3, "hot3", "cold3", "region", "US", 39.0, -122.0, nil)
	suite.Require().NoError(err)

	// Jail val1 and val2
	err = suite.keeper.JailValidator(suite.ctx, val1, time.Hour)
	suite.Require().NoError(err)
	err = suite.keeper.JailValidator(suite.ctx, val2, time.Hour)
	suite.Require().NoError(err)

	// Get jailed validators
	jailed := suite.keeper.GetJailedValidators(suite.ctx)
	suite.Require().GreaterOrEqual(len(jailed), 2)
}

func (suite *KeeperTestSuite) TestGetTombstonedValidators() {
	// Setup validators
	val1 := "auravaloper1tomb1"
	val2 := "auravaloper1tomb2"

	err := suite.keeper.RegisterValidator(suite.ctx, val1, "hot1", "cold1", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, val2, "hot2", "cold2", "region", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)

	// Tombstone val1
	err = suite.keeper.TombstoneValidator(suite.ctx, val1)
	suite.Require().NoError(err)

	// Get tombstoned validators
	tombstoned := suite.keeper.GetTombstonedValidators(suite.ctx)
	suite.Require().GreaterOrEqual(len(tombstoned), 1)
}

func (suite *KeeperTestSuite) TestJailTombstonedValidator() {
	validatorAddr := "auravaloper1tombjail"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Tombstone first
	err = suite.keeper.TombstoneValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Try to jail - should fail
	err = suite.keeper.JailValidator(suite.ctx, validatorAddr, time.Hour)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrValidatorTombstoned, err)
}

func (suite *KeeperTestSuite) TestUnjailTombstonedValidator() {
	validatorAddr := "auravaloper1tombunjail"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Tombstone
	err = suite.keeper.TombstoneValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Try to unjail - should fail
	err = suite.keeper.UnjailValidator(suite.ctx, validatorAddr)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrValidatorTombstoned, err)
}

func (suite *KeeperTestSuite) TestIsValidatorJailedNotFound() {
	// Check non-existent validator
	isJailed := suite.keeper.IsValidatorJailed(suite.ctx, "nonexistent")
	suite.Require().False(isJailed)
}

func (suite *KeeperTestSuite) TestIsValidatorTombstonedNotFound() {
	// Check non-existent validator
	isTombstoned := suite.keeper.IsValidatorTombstoned(suite.ctx, "nonexistent")
	suite.Require().False(isTombstoned)
}

func newValAddr() string {
	return sdk.ValAddress(keepertest.GenTestAddr()).String()
}
