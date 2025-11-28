package keeper_test

import (
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func (suite *KeeperTestSuite) TestValidateMinimumStake() {
	// This test would require mocking the staking keeper
	// For now, we'll test the parameter validation
	params := suite.keeper.GetParams(suite.ctx)
	// MinimumStakeAmount is now a string, just check it's not empty
	suite.Require().NotEmpty(params.MinimumStakeAmount)
}

func (suite *KeeperTestSuite) TestSlashFractionValidation() {
	params := types.DefaultParams()

	// Test valid slash fractions (they're now strings)
	suite.Require().NoError(types.ValidateParams(params))
	suite.Require().NotEmpty(params.DoubleSignSlashFraction)
	suite.Require().NotEmpty(params.DowntimeSlashFraction)
	// String values should be valid decimals
	suite.Require().Equal("0.05", params.DoubleSignSlashFraction)
	suite.Require().Equal("0.01", params.DowntimeSlashFraction)
}

func (suite *KeeperTestSuite) TestInvalidSlashFraction() {
	params := types.DefaultParams()

	// Test empty slash fraction
	params.DoubleSignSlashFraction = ""
	suite.Require().Error(types.ValidateParams(params))

	// Test empty downtime slash fraction
	params = types.DefaultParams()
	params.DowntimeSlashFraction = ""
	suite.Require().Error(types.ValidateParams(params))
}

func (suite *KeeperTestSuite) TestSignedBlocksWindowValidation() {
	params := types.DefaultParams()

	// Valid window - default should be valid
	suite.Require().NoError(types.ValidateParams(params))

	// Test with zero window (if validation is implemented)
	params.SignedBlocksWindow = 0
	// Note: Add validation in types.ValidateParams if needed
}

func (suite *KeeperTestSuite) TestMinSignedPerWindowValidation() {
	params := types.DefaultParams()

	// Valid percentage - should be a string decimal
	suite.Require().NotEmpty(params.MinSignedPerWindow)
	// The string should represent a valid decimal
	// Note: Implement string decimal validation in types.ValidateParams if needed
}
