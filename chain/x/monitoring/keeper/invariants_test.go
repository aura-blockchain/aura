package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

type InvariantsTestSuite struct {
	suite.Suite

	Keeper *Keeper
	SdkCtx sdk.Context
}

func (suite *InvariantsTestSuite) SetupTest() {
	keeper, ctx := NewTestKeeper(suite.T())
	suite.Keeper = &keeper
	suite.SdkCtx = ctx

	// Set default params
	params := types.DefaultParams()
	err := suite.Keeper.SetParams(ctx, params)
	suite.Require().NoError(err)
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

// Test all invariants on empty store
func (suite *InvariantsTestSuite) TestAllInvariantsEmptyStore() {
	inv := AllInvariants(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

// Test invariant registration - skipped as SDK doesn't expose NewInvariantRegistry in newer versions
// The invariants are still tested individually below
func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	suite.T().Skip("SDK InvariantRegistry constructor not available in test context")
}

// ============================================================================
// ParamsInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestParamsInvariantValid() {
	inv := ParamsInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "params invariant should pass with default params")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestParamsInvariantInvalid() {
	// Verify that SetParams rejects invalid params (negative retention period)
	params, err := suite.Keeper.GetParams(suite.SdkCtx)
	suite.Require().NoError(err)
	params.AlertRetentionPeriod = -100 * time.Second // Invalid
	err = suite.Keeper.SetParams(suite.SdkCtx, params)
	suite.Error(err, "SetParams should reject invalid params")

	// Invariant should still pass since invalid params were rejected
	inv := ParamsInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "params invariant should pass since invalid params were rejected")
	suite.Empty(msg)
}

// ============================================================================
// AlertsInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestAlertsInvariantValid() {
	// Create a valid alert
	alert, err := suite.Keeper.CreateAlert(
		suite.SdkCtx,
		types.AlertTypeSystemError,
		types.SeverityMedium,
		"Test alert",
		map[string]interface{}{"test": "data"},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(alert)

	inv := AlertsInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "alerts invariant should pass with valid alert")
	suite.Empty(msg)
}

// ============================================================================
// ValidatorUptimeInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestValidatorUptimeInvariantValid() {
	// Create valid validator uptime
	err := suite.Keeper.UpdateValidatorUptime(
		suite.SdkCtx,
		"auravaloper1test",
		"TestValidator",
		100,
		true,
	)
	suite.Require().NoError(err)

	inv := ValidatorUptimeInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "validator uptime invariant should pass with valid data")
	suite.Empty(msg)
}

// ============================================================================
// All Invariants Integration Test
// ============================================================================

func (suite *InvariantsTestSuite) TestAllInvariantsWithValidData() {
	// Create valid alert
	alert, err := suite.Keeper.CreateAlert(
		suite.SdkCtx,
		types.AlertTypeLargeTransaction,
		types.SeverityHigh,
		"Large transaction detected",
		map[string]interface{}{"amount": 1000000},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(alert)

	// Create valid validator uptime
	err = suite.Keeper.UpdateValidatorUptime(
		suite.SdkCtx,
		"auravaloper1test",
		"TestValidator",
		100,
		true,
	)
	suite.Require().NoError(err)

	// Run all invariants
	inv := AllInvariants(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "all invariants should pass with valid data")
	suite.Empty(msg)
}
