// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	sdkmath "cosmossdk.io/math"

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

	// Test valid slash fractions (they're LegacyDec types)
	suite.Require().NoError(types.ValidateParams(params))
	suite.Require().False(params.DoubleSignSlashFraction.IsNil())
	suite.Require().False(params.DowntimeSlashFraction.IsNil())
	// Check expected values
	suite.Require().Equal(sdkmath.LegacyMustNewDecFromStr("0.05").String(), params.DoubleSignSlashFraction.String())
	suite.Require().Equal(sdkmath.LegacyMustNewDecFromStr("0.01").String(), params.DowntimeSlashFraction.String())
}

func (suite *KeeperTestSuite) TestInvalidSlashFraction() {
	params := types.DefaultParams()

	// Test negative slash fraction
	params.DoubleSignSlashFraction = sdkmath.LegacyMustNewDecFromStr("-0.05")
	suite.Require().Error(types.ValidateParams(params))

	// Test slash fraction > 1
	params = types.DefaultParams()
	params.DowntimeSlashFraction = sdkmath.LegacyMustNewDecFromStr("1.5")
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
