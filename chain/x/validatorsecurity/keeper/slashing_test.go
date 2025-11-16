package keeper_test

import (
	"cosmossdk.io/math"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func (suite *KeeperTestSuite) TestValidateMinimumStake() {
	// This test would require mocking the staking keeper
	// For now, we'll test the parameter validation
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().False(params.MinimumStakeAmount.IsNegative())
}

func (suite *KeeperTestSuite) TestSlashFractionValidation() {
	params := types.DefaultParams()

	// Test valid slash fractions
	suite.Require().NoError(params.Validate())
	suite.Require().False(params.DoubleSignSlashFraction.IsNegative())
	suite.Require().False(params.DowntimeSlashFraction.IsNegative())
	suite.Require().True(params.DoubleSignSlashFraction.LTE(math.LegacyOneDec()))
	suite.Require().True(params.DowntimeSlashFraction.LTE(math.LegacyOneDec()))
}

func (suite *KeeperTestSuite) TestInvalidSlashFraction() {
	params := types.DefaultParams()

	// Test negative slash fraction
	params.DoubleSignSlashFraction = math.LegacyNewDec(-1)
	suite.Require().Error(params.Validate())

	// Test slash fraction > 1
	params = types.DefaultParams()
	params.DoubleSignSlashFraction = math.LegacyNewDec(2)
	suite.Require().Error(params.Validate())
}

func (suite *KeeperTestSuite) TestSignedBlocksWindowValidation() {
	params := types.DefaultParams()

	// Valid window
	suite.Require().True(params.SignedBlocksWindow > 0)

	// Invalid window
	params.SignedBlocksWindow = -100
	suite.Require().Error(params.Validate())

	params.SignedBlocksWindow = 0
	suite.Require().Error(params.Validate())
}

func (suite *KeeperTestSuite) TestMinSignedPerWindowValidation() {
	params := types.DefaultParams()

	// Valid percentage
	suite.Require().False(params.MinSignedPerWindow.IsNegative())
	suite.Require().True(params.MinSignedPerWindow.LTE(math.LegacyOneDec()))

	// Invalid percentage
	params.MinSignedPerWindow = math.LegacyNewDec(-1)
	suite.Require().Error(params.Validate())

	params = types.DefaultParams()
	params.MinSignedPerWindow = math.LegacyNewDec(2)
	suite.Require().Error(params.Validate())
}
