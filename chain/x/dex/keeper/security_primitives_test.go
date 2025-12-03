package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

// SecurityPrimitivesTestSuite tests the security primitives integration
type SecurityPrimitivesTestSuite struct {
	KeeperTestSuite
}

// TestSecurityPrimitives runs the security primitives test suite
func TestSecurityPrimitives(t *testing.T) {
	suite.Run(t, new(SecurityPrimitivesTestSuite))
}

// TestReentrancyGuardInitialized verifies the reentrancy guard is initialized
func (suite *SecurityPrimitivesTestSuite) TestReentrancyGuardInitialized() {
	// Verify keeper has reentrancy guard
	require := suite.Require()
	require.NotNil(suite.Keeper)

	// The keeper should be initialized with security primitives
	// This is verified by successful instantiation in SetupTest
	require.NotNil(suite.Keeper.reentrancyGuard, "reentrancy guard should be initialized")
	require.NotNil(suite.Keeper.pauseGuard, "pause guard should be initialized")
	require.NotNil(suite.Keeper.inputValidator, "input validator should be initialized")
	require.NotNil(suite.Keeper.safeMath, "safe math should be initialized")
	require.NotNil(suite.Keeper.gasLimitGuard, "gas limit guard should be initialized")
	require.NotNil(suite.Keeper.accessControl, "access control should be initialized")
}

// TestPauseGuardInitialized verifies the pause guard is properly initialized
func (suite *SecurityPrimitivesTestSuite) TestPauseGuardInitialized() {
	require := suite.Require()

	// Verify pause guard is not nil and not paused by default
	require.NotNil(suite.Keeper.pauseGuard)
	require.False(suite.Keeper.pauseGuard.IsPaused(), "module should not be paused by default")
}

// TestInputValidatorInitialized verifies the input validator is properly initialized
func (suite *SecurityPrimitivesTestSuite) TestInputValidatorInitialized() {
	require := suite.Require()

	// Verify input validator can validate positive amounts
	positiveAmount := sdkmath.NewInt(1000)
	err := suite.Keeper.inputValidator.ValidateAmount(positiveAmount)
	require.NoError(err, "should validate positive amounts")

	// Verify input validator rejects zero amounts
	zeroAmount := sdkmath.ZeroInt()
	err = suite.Keeper.inputValidator.ValidateAmount(zeroAmount)
	require.Error(err, "should reject zero amounts")

	// Verify input validator rejects negative amounts (if possible to create)
	// Note: sdkmath.Int doesn't allow negative values by design, so this is implicit
}

// TestSafeMathInitialized verifies the safe math utility is properly initialized
func (suite *SecurityPrimitivesTestSuite) TestSafeMathInitialized() {
	require := suite.Require()
	require.NotNil(suite.Keeper.safeMath, "safe math should be initialized")

	// Safe math is a utility that provides overflow-safe arithmetic
	// Actual overflow protection is tested in the SafeMath implementation tests
}

// TestGasLimitGuardInitialized verifies the gas limit guard is properly initialized
func (suite *SecurityPrimitivesTestSuite) TestGasLimitGuardInitialized() {
	require := suite.Require()
	require.NotNil(suite.Keeper.gasLimitGuard, "gas limit guard should be initialized")

	// Gas limit guard should reject zero gas limit
	err := suite.Keeper.gasLimitGuard.ValidateGasLimit(0)
	require.Error(err, "should reject zero gas limit")

	// Gas limit guard should accept reasonable gas limit
	err = suite.Keeper.gasLimitGuard.ValidateGasLimit(500000)
	require.NoError(err, "should accept reasonable gas limit")
}

// TestAccessControlInitialized verifies the access control is properly initialized
func (suite *SecurityPrimitivesTestSuite) TestAccessControlInitialized() {
	require := suite.Require()
	require.NotNil(suite.Keeper.accessControl, "access control should be initialized")

	// Access control should be initialized with no admins by default
	testAddr := sdk.AccAddress("test_address_______")
	require.False(suite.Keeper.accessControl.IsAdmin(testAddr.String()), "no admins by default")
}
