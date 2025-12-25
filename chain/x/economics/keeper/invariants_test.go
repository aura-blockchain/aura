// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/testutil"
	"github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

type InvariantsTestSuite struct {
	suite.Suite

	keeper *Keeper
	ctx    sdk.Context
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) SetupTest() {
	suite.keeper, suite.ctx = setupKeeperForTest(suite.T())

	// Set default valid params
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Set default valid fee multiplier
	err = suite.keeper.SetFeeMultiplier(suite.ctx, "1.0")
	suite.Require().NoError(err)

	// Set default valid transfer tax configuration
	err = suite.keeper.SetTransferTaxEnabled(suite.ctx, false)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "0.0")
	suite.Require().NoError(err)
}

// setParamsUnchecked bypasses validation and writes params directly to the store.
// This is used in tests that need to inject invalid params to test invariants.
func (suite *InvariantsTestSuite) setParamsUnchecked(params *economicspb.Params) {
	store := suite.keeper.storeService.OpenKVStore(suite.ctx)
	bz, err := suite.keeper.cdc.Marshal(params)
	suite.Require().NoError(err)
	err = store.Set(types.ParamsKey, bz)
	suite.Require().NoError(err)
}

// ============================
// ALL INVARIANTS TESTS
// ============================

func (suite *InvariantsTestSuite) TestAllInvariants_EmptyStore() {
	// Test: All invariants should pass on a fresh store with default params
	inv := AllInvariants(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "all invariants should pass on empty store with defaults")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestAllInvariants_ValidState() {
	// Set up valid state
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	err = suite.keeper.SetFeeMultiplier(suite.ctx, "1.5")
	suite.Require().NoError(err)

	err = suite.keeper.SetTransferTaxEnabled(suite.ctx, true)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "0.05")
	suite.Require().NoError(err)

	// Test: All invariants should pass with valid state
	inv := AllInvariants(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "all invariants should pass with valid state")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestAllInvariants_DetectsInvalidParams() {
	// Inject invalid params (min > max fee multiplier) directly to bypass validation
	params := types.DefaultParams()
	params.Fees.MinFeeMultiplier = 60000 // 6x
	params.Fees.MaxFeeMultiplier = 50000 // 5x (less than min)
	suite.setParamsUnchecked(params)

	// Test: AllInvariants should detect invalid params
	inv := AllInvariants(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "all invariants should detect invalid params")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Verify each invariant doesn't panic when called
	invariants := []func(sdk.Context) (string, bool){
		ParamsInvariant(suite.keeper),
		FeeMultiplierInvariant(suite.keeper),
		TransferTaxConsistencyInvariant(suite.keeper),
	}

	for _, inv := range invariants {
		msg, broken := inv(suite.ctx)
		suite.False(broken, "invariant should not be broken with valid default state")
		suite.Empty(msg)
	}
}

// ============================
// PARAMS INVARIANT TESTS
// ============================

func (suite *InvariantsTestSuite) TestParamsInvariant_ValidParams() {
	// Set valid params
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Test: Params invariant should pass with valid params
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "params invariant should pass with valid params")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidFeeParams_MinGreaterThanMax() {
	// Set params with min > max fee multiplier
	params := types.DefaultParams()
	params.Fees.MinFeeMultiplier = 60000 // 6x
	params.Fees.MaxFeeMultiplier = 50000 // 5x
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when min > max fee multiplier")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidFeeParams_TargetUtilizationTooHigh() {
	// Set params with target utilization > 100%
	params := types.DefaultParams()
	params.Fees.TargetBlockUtilization = 12000 // 120% (invalid)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when target utilization > 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidVestingParams_MaxLessThanMin() {
	// Set params with max < min vesting duration
	params := types.DefaultParams()
	params.Vesting.MinVestingDuration = testutil.Days(365) // 1 year
	params.Vesting.MaxVestingDuration = testutil.Days(30)  // 30 days (less than min)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when max vesting < min vesting")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidVestingParams_PenaltyTooHigh() {
	// Set params with early unlock penalty > 100%
	params := types.DefaultParams()
	params.Vesting.EarlyUnlockPenalty = 12000 // 120% (invalid)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when early unlock penalty > 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidTreasuryParams_CommunityPoolTooHigh() {
	// Set params with community pool percentage > 100%
	params := types.DefaultParams()
	params.Treasury.CommunityPoolPercentage = 15000 // 150% (invalid)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when community pool > 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidTreasuryParams_BurnPercentageTooHigh() {
	// Set params with burn percentage > 100%
	params := types.DefaultParams()
	params.Treasury.BurnPercentage = 10500 // 105% (invalid)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when burn percentage > 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidTreasuryParams_ZeroThreshold() {
	// Set params with zero multisig threshold
	params := types.DefaultParams()
	params.Treasury.MultisigThreshold = 0 // Invalid - must be at least 1
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when multisig threshold is 0")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidGovernanceParams_QuorumTooHigh() {
	// Set params with quorum > 100%
	params := types.DefaultParams()
	params.Governance.Quorum = 11000 // 110% (invalid)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when quorum > 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidGovernanceParams_ThresholdTooHigh() {
	// Set params with threshold > 100%
	params := types.DefaultParams()
	params.Governance.Threshold = 10001 // 100.01% (invalid)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when threshold > 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidGovernanceParams_VetoThresholdTooHigh() {
	// Set params with veto threshold > 100%
	params := types.DefaultParams()
	params.Governance.VetoThreshold = 20000 // 200% (invalid)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when veto threshold > 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidMEVParams_PercentagesMismatch() {
	// Set params with MEV redistribution percentages not summing to 100%
	params := types.DefaultParams()
	params.Mev.UserRedistributionPercentage = 5000  // 50%
	params.Mev.ValidatorPercentage = 3000           // 30%
	params.Mev.TreasuryPercentage = 1000            // 10%
	params.Mev.BurnPercentage = 500                 // 5%
	// Total = 95%, not 100%
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when MEV percentages don't sum to 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidWhaleProtectionParams_MaxHoldingTooHigh() {
	// Set params with max holding percentage > 100%
	params := types.DefaultParams()
	params.WhaleProtection.MaxHoldingPercentage = 10500 // 105% (invalid)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when max holding > 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidWhaleProtectionParams_LargeTxThresholdTooHigh() {
	// Set params with large tx threshold > 100%
	params := types.DefaultParams()
	params.WhaleProtection.LargeTxThreshold = 15000 // 150% (invalid)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when large tx threshold > 100%")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidLiquidityMiningParams_ZeroEpochDuration() {
	// Set params with zero epoch duration blocks
	params := types.DefaultParams()
	params.LiquidityMining.EpochDurationBlocks = 0 // Invalid - must be positive
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when epoch duration is 0")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidTokenomicsParams_MinGreaterThanTarget() {
	// Set params with min inflation > target inflation
	params := types.DefaultParams()
	params.Tokenomics.MinInflationRate = 800  // 8%
	params.Tokenomics.TargetInflationRate = 500 // 5% (less than min)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when min inflation > target inflation")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidTokenomicsParams_TargetGreaterThanMax() {
	// Set params with target inflation > max inflation
	params := types.DefaultParams()
	params.Tokenomics.TargetInflationRate = 1200 // 12%
	params.Tokenomics.MaxInflationRate = 1000    // 10% (less than target)
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when target inflation > max inflation")
	suite.Contains(msg, "invalid module params")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidTokenomicsParams_ZeroCheckInterval() {
	// Set params with zero inflation check interval
	params := types.DefaultParams()
	params.Tokenomics.InflationCheckInterval = 0 // Invalid - must be positive
	suite.setParamsUnchecked(params)

	// Test: Params invariant should fail
	inv := ParamsInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "params invariant should fail when inflation check interval is 0")
	suite.Contains(msg, "invalid module params")
}

// ============================
// FEE MULTIPLIER INVARIANT TESTS
// ============================

func (suite *InvariantsTestSuite) TestFeeMultiplierInvariant_ValidMultiplier() {
	// Set valid fee multiplier
	err := suite.keeper.SetFeeMultiplier(suite.ctx, "1.5")
	suite.Require().NoError(err)

	// Test: Fee multiplier invariant should pass
	inv := FeeMultiplierInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "fee multiplier invariant should pass with valid multiplier")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestFeeMultiplierInvariant_EmptyMultiplier() {
	// Set empty fee multiplier
	err := suite.keeper.SetFeeMultiplier(suite.ctx, "")
	suite.Require().NoError(err)

	// Test: Fee multiplier invariant should fail
	inv := FeeMultiplierInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "fee multiplier invariant should fail with empty multiplier")
	suite.Contains(msg, "fee multiplier is empty")
}

func (suite *InvariantsTestSuite) TestFeeMultiplierInvariant_NegativeMultiplier() {
	// Set negative fee multiplier
	err := suite.keeper.SetFeeMultiplier(suite.ctx, "-0.5")
	suite.Require().NoError(err)

	// Test: Fee multiplier invariant should fail
	inv := FeeMultiplierInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "fee multiplier invariant should fail with negative multiplier")
	suite.Contains(msg, "fee multiplier cannot be negative")
}

func (suite *InvariantsTestSuite) TestFeeMultiplierInvariant_ZeroMultiplier() {
	// Set zero fee multiplier (should be valid)
	err := suite.keeper.SetFeeMultiplier(suite.ctx, "0")
	suite.Require().NoError(err)

	// Test: Fee multiplier invariant should pass (zero is non-negative)
	inv := FeeMultiplierInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "fee multiplier invariant should pass with zero multiplier")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestFeeMultiplierInvariant_VeryLargeMultiplier() {
	// Set very large fee multiplier
	err := suite.keeper.SetFeeMultiplier(suite.ctx, "1000000.999")
	suite.Require().NoError(err)

	// Test: Fee multiplier invariant should pass (large values are valid)
	inv := FeeMultiplierInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "fee multiplier invariant should pass with large multiplier")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestFeeMultiplierInvariant_DecimalMultiplier() {
	// Set decimal fee multiplier
	err := suite.keeper.SetFeeMultiplier(suite.ctx, "0.123456789")
	suite.Require().NoError(err)

	// Test: Fee multiplier invariant should pass
	inv := FeeMultiplierInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "fee multiplier invariant should pass with decimal multiplier")
	suite.Empty(msg)
}

// ============================
// TRANSFER TAX CONSISTENCY INVARIANT TESTS
// ============================

func (suite *InvariantsTestSuite) TestTransferTaxConsistencyInvariant_Disabled() {
	// Set transfer tax disabled
	err := suite.keeper.SetTransferTaxEnabled(suite.ctx, false)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "0")
	suite.Require().NoError(err)

	// Test: Transfer tax invariant should pass when disabled
	inv := TransferTaxConsistencyInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "transfer tax invariant should pass when disabled")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestTransferTaxConsistencyInvariant_EnabledWithValidRate() {
	// Set transfer tax enabled with valid rate
	err := suite.keeper.SetTransferTaxEnabled(suite.ctx, true)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "0.05")
	suite.Require().NoError(err)

	// Test: Transfer tax invariant should pass
	inv := TransferTaxConsistencyInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "transfer tax invariant should pass with valid rate")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestTransferTaxConsistencyInvariant_EnabledWithEmptyRate() {
	// Set transfer tax enabled but with empty rate
	err := suite.keeper.SetTransferTaxEnabled(suite.ctx, true)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "")
	suite.Require().NoError(err)

	// Test: Transfer tax invariant should fail
	inv := TransferTaxConsistencyInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "transfer tax invariant should fail when enabled with empty rate")
	suite.Contains(msg, "transfer tax enabled but rate is empty")
}

func (suite *InvariantsTestSuite) TestTransferTaxConsistencyInvariant_EnabledWithNegativeRate() {
	// Set transfer tax enabled with negative rate
	err := suite.keeper.SetTransferTaxEnabled(suite.ctx, true)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "-0.01")
	suite.Require().NoError(err)

	// Test: Transfer tax invariant should fail
	inv := TransferTaxConsistencyInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "transfer tax invariant should fail with negative rate")
	suite.Contains(msg, "transfer tax rate cannot be negative")
}

func (suite *InvariantsTestSuite) TestTransferTaxConsistencyInvariant_EnabledWithZeroRate() {
	// Set transfer tax enabled with zero rate (should be valid, just ineffective)
	err := suite.keeper.SetTransferTaxEnabled(suite.ctx, true)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "0")
	suite.Require().NoError(err)

	// Test: Transfer tax invariant should pass (zero is non-negative)
	inv := TransferTaxConsistencyInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "transfer tax invariant should pass with zero rate")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestTransferTaxConsistencyInvariant_EnabledWithHighRate() {
	// Set transfer tax enabled with high but valid rate
	err := suite.keeper.SetTransferTaxEnabled(suite.ctx, true)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "0.95") // 95% tax
	suite.Require().NoError(err)

	// Test: Transfer tax invariant should pass (no upper bound in invariant)
	inv := TransferTaxConsistencyInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "transfer tax invariant should pass with high rate")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestTransferTaxConsistencyInvariant_DisabledWithNonZeroRate() {
	// Set transfer tax disabled with non-zero rate (should still be valid)
	err := suite.keeper.SetTransferTaxEnabled(suite.ctx, false)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "0.05")
	suite.Require().NoError(err)

	// Test: Transfer tax invariant should pass (rate can be set even when disabled)
	inv := TransferTaxConsistencyInvariant(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "transfer tax invariant should pass when disabled regardless of rate")
	suite.Empty(msg)
}

// ============================
// EDGE CASE AND STRESS TESTS
// ============================

func (suite *InvariantsTestSuite) TestInvariants_NotPanic() {
	// Test that all invariant functions don't panic
	suite.NotPanics(func() {
		ParamsInvariant(suite.keeper)(suite.ctx)
	}, "ParamsInvariant should not panic")

	suite.NotPanics(func() {
		FeeMultiplierInvariant(suite.keeper)(suite.ctx)
	}, "FeeMultiplierInvariant should not panic")

	suite.NotPanics(func() {
		TransferTaxConsistencyInvariant(suite.keeper)(suite.ctx)
	}, "TransferTaxConsistencyInvariant should not panic")

	suite.NotPanics(func() {
		AllInvariants(suite.keeper)(suite.ctx)
	}, "AllInvariants should not panic")
}

func (suite *InvariantsTestSuite) TestInvariants_MultipleViolations() {
	// Set up state with multiple invariant violations
	params := types.DefaultParams()

	// Violation 1: Invalid fee params
	params.Fees.MinFeeMultiplier = 60000
	params.Fees.MaxFeeMultiplier = 50000

	// Violation 2: Invalid vesting params
	params.Vesting.EarlyUnlockPenalty = 15000

	// Violation 3: Invalid MEV params
	params.Mev.UserRedistributionPercentage = 5000
	params.Mev.ValidatorPercentage = 3000
	params.Mev.TreasuryPercentage = 1000
	params.Mev.BurnPercentage = 500

	suite.setParamsUnchecked(params)

	// Set invalid fee multiplier
	err := suite.keeper.SetFeeMultiplier(suite.ctx, "-1.0")
	suite.Require().NoError(err)

	// Set invalid transfer tax
	err = suite.keeper.SetTransferTaxEnabled(suite.ctx, true)
	suite.Require().NoError(err)
	err = suite.keeper.SetTransferTaxRate(suite.ctx, "-0.05")
	suite.Require().NoError(err)

	// Test: AllInvariants should detect at least one violation (stops on first)
	inv := AllInvariants(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.True(broken, "all invariants should detect violations")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestInvariants_BoundaryValues() {
	// Test boundary values that should be valid
	params := types.DefaultParams()

	// Boundary 1: Min equals max (should be valid)
	params.Fees.MinFeeMultiplier = 10000
	params.Fees.MaxFeeMultiplier = 10000

	// Boundary 2: Exactly 100% utilization
	params.Fees.TargetBlockUtilization = 10000

	// Boundary 3: Exactly 100% percentages
	params.Treasury.CommunityPoolPercentage = 10000
	params.Governance.Quorum = 10000
	params.WhaleProtection.MaxHoldingPercentage = 10000

	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Test: All invariants should pass with boundary values
	inv := AllInvariants(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "all invariants should pass with boundary values")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestInvariants_ExtremelyLargeValues() {
	// Test with extremely large math.Int values
	params := types.DefaultParams()
	params.Treasury.SpendingLimit = sdkmath.NewIntFromUint64(^uint64(0)) // Max uint64
	params.WhaleProtection.MaxSingleTransfer = sdkmath.NewIntFromUint64(^uint64(0))
	params.LiquidityMining.TotalRewardsAllocated = sdkmath.NewIntFromUint64(^uint64(0))
	params.Tokenomics.MaxSupply = sdkmath.NewIntFromUint64(^uint64(0))

	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Test: Invariants should handle extremely large values
	inv := AllInvariants(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "invariants should handle extremely large values")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestInvariants_AllZeroValues() {
	// Test with all zero/minimum values where applicable
	params := types.DefaultParams()
	params.Fees.BaseFee = sdkmath.ZeroInt()
	params.Fees.MinGasPrice = sdkmath.LegacyZeroDec()
	params.Fees.FeeBurnPercentage = 0
	params.Treasury.CommunityPoolPercentage = 0
	params.Treasury.BurnPercentage = 0
	params.Governance.Quorum = 0
	params.Governance.Threshold = 0
	params.Governance.VetoThreshold = 0
	params.Mev.UserRedistributionPercentage = 0
	params.Mev.ValidatorPercentage = 0
	params.Mev.TreasuryPercentage = 0
	params.Mev.BurnPercentage = 10000 // All to burn to sum to 100%
	params.WhaleProtection.MaxHoldingPercentage = 0
	params.WhaleProtection.LargeTxThreshold = 0
	params.Tokenomics.MinInflationRate = 0
	params.Tokenomics.TargetInflationRate = 0
	params.Tokenomics.MaxInflationRate = 0

	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Test: Invariants should handle all-zero values
	inv := AllInvariants(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "invariants should handle all-zero values")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestInvariants_MinimalValidState() {
	// Create minimal valid state
	params := types.DefaultParams()
	params.Treasury.MultisigThreshold = 1 // Minimum valid threshold
	params.LiquidityMining.EpochDurationBlocks = 1 // Minimum valid epoch
	params.Tokenomics.InflationCheckInterval = 1 // Minimum valid interval

	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	err = suite.keeper.SetFeeMultiplier(suite.ctx, "0.01")
	suite.Require().NoError(err)

	err = suite.keeper.SetTransferTaxEnabled(suite.ctx, false)
	suite.Require().NoError(err)

	// Test: All invariants should pass with minimal valid state
	inv := AllInvariants(suite.keeper)
	msg, broken := inv(suite.ctx)
	suite.False(broken, "all invariants should pass with minimal valid state")
	suite.Empty(msg)
}
