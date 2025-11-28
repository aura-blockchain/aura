package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

type InvariantsTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

// Test params are valid
func (suite *InvariantsTestSuite) TestParamsValid() {
	params := suite.Keeper.GetParams(suite.SdkCtx)
	suite.Require().NotNil(params)

	// Validate params
	err := types.ValidateParams(params)
	suite.Require().NoError(err, "default params should be valid")
}

// Test params validation catches invalid values
func (suite *InvariantsTestSuite) TestParamsInvalid() {
	// Test with invalid params
	invalidParams := types.DefaultParams()
	invalidParams.MaxCacheSize = 0 // Invalid - should be > 0

	err := types.ValidateParams(invalidParams)
	suite.Require().Error(err, "invalid params should fail validation")
}

// Test params can be set and retrieved
func (suite *InvariantsTestSuite) TestParamsSetGet() {
	newParams := types.DefaultParams()
	newParams.Enabled = false
	newParams.MaxCacheSize = 5000

	err := suite.Keeper.SetParams(suite.SdkCtx, newParams)
	suite.Require().NoError(err)

	retrieved := suite.Keeper.GetParams(suite.SdkCtx)
	suite.Require().Equal(newParams.Enabled, retrieved.Enabled)
	suite.Require().Equal(newParams.MaxCacheSize, retrieved.MaxCacheSize)
}
