// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	economicstypes "github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// TestDefaultParams verifies default parameters are valid and properly structured
func TestDefaultParams(t *testing.T) {
	params := economicstypes.DefaultParams()

	require.NotNil(t, params)

	// Validate default params pass validation
	err := economicstypes.ValidateParams(params)
	require.NoError(t, err, "default params should be valid")

	// Verify sub-params are not nil
	require.NotNil(t, params.Fees)
	require.NotNil(t, params.Vesting)
	require.NotNil(t, params.Treasury)
	require.NotNil(t, params.Governance)
	require.NotNil(t, params.Mev)
	require.NotNil(t, params.WhaleProtection)
	require.NotNil(t, params.LiquidityMining)
	require.NotNil(t, params.Tokenomics)
}

// TestDefaultFeeParams verifies default fee parameters
func TestDefaultFeeParams(t *testing.T) {
	params := economicstypes.DefaultFeeParams()

	require.NotNil(t, params)

	// Verify expected default values
	require.True(t, params.BaseFee.IsPositive(), "base fee should be positive")
	require.True(t, params.MinGasPrice.IsPositive(), "min gas price should be positive")
	require.True(t, params.DynamicFeesEnabled, "dynamic fees should be enabled by default")
	require.True(t, params.FeeBurnPercentage <= 10000, "fee burn percentage should be <= 100%")
	require.True(t, params.MinFeeMultiplier <= params.MaxFeeMultiplier, "min multiplier should be <= max")
	require.True(t, params.TargetBlockUtilization <= 10000, "target utilization should be <= 100%")
	require.True(t, params.FeeAdjustmentSpeed <= 10000, "adjustment speed should be <= 100%")

	// Validate parameters
	err := economicstypes.ValidateFeeParams(params)
	require.NoError(t, err, "default fee params should be valid")
}

// TestDefaultVestingParams verifies default vesting parameters
func TestDefaultVestingParams(t *testing.T) {
	params := economicstypes.DefaultVestingParams()

	require.NotNil(t, params)

	// Verify expected default values
	require.True(t, params.MinVestingDuration > 0, "min vesting duration should be positive")
	require.True(t, params.MaxVestingDuration >= params.MinVestingDuration, "max duration >= min duration")
	require.True(t, params.EarlyUnlockPenalty <= 10000, "early unlock penalty should be <= 100%")
	require.True(t, params.MinCliffDuration >= 0, "min cliff duration should be >= 0")

	// Validate parameters
	err := economicstypes.ValidateVestingParams(params)
	require.NoError(t, err, "default vesting params should be valid")
}

// TestDefaultTreasuryParams verifies default treasury parameters
func TestDefaultTreasuryParams(t *testing.T) {
	params := economicstypes.DefaultTreasuryParams()

	require.NotNil(t, params)

	// Verify expected default values
	require.True(t, params.CommunityPoolPercentage <= 10000, "community pool percentage should be <= 100%")
	require.True(t, params.BurnPercentage <= 10000, "burn percentage should be <= 100%")
	require.True(t, params.MultisigThreshold > 0, "multisig threshold should be positive")
	require.NotNil(t, params.AuthorizedSigners, "authorized signers should not be nil")
	require.True(t, params.SpendingLimit.IsPositive(), "spending limit should be positive")
	require.True(t, params.TimelockDuration > 0, "timelock duration should be positive")

	// Validate parameters
	err := economicstypes.ValidateTreasuryParams(params)
	require.NoError(t, err, "default treasury params should be valid")
}

// TestDefaultGovernanceParams verifies default governance parameters
func TestDefaultGovernanceParams(t *testing.T) {
	params := economicstypes.DefaultGovernanceParams()

	require.NotNil(t, params)

	// Verify expected default values
	require.True(t, params.MaxDepositPeriod > 0, "max deposit period should be positive")
	require.True(t, params.VotingPeriod > 0, "voting period should be positive")
	require.True(t, params.Quorum <= 10000, "quorum should be <= 100%")
	require.True(t, params.Threshold <= 10000, "threshold should be <= 100%")
	require.True(t, params.VetoThreshold <= 10000, "veto threshold should be <= 100%")
	require.True(t, params.ExecutionDelay >= 0, "execution delay should be >= 0")
	require.True(t, params.EmergencyVotingPeriod > 0, "emergency voting period should be positive")
	require.True(t, params.EmergencyQuorum <= 10000, "emergency quorum should be <= 100%")
	require.True(t, params.EmergencyThreshold <= 10000, "emergency threshold should be <= 100%")
	require.True(t, params.VoteLockingEnabled, "vote locking should be enabled by default")
	require.True(t, params.MinLockDuration > 0, "min lock duration should be positive")
	require.True(t, params.MaxLockDuration >= params.MinLockDuration, "max lock >= min lock")
	require.True(t, params.LockMultiplierPerYear > 0, "lock multiplier should be positive")
	require.True(t, params.SnapshotLookbackBlocks > 0, "snapshot lookback should be positive")
	require.True(t, params.RevealPeriod > 0, "reveal period should be positive")

	// Validate parameters
	err := economicstypes.ValidateGovernanceParams(params)
	require.NoError(t, err, "default governance params should be valid")
}

// TestDefaultMEVParams verifies default MEV protection parameters
func TestDefaultMEVParams(t *testing.T) {
	params := economicstypes.DefaultMEVParams()

	require.NotNil(t, params)

	// Verify expected default values
	require.True(t, params.Enabled, "MEV protection should be enabled by default")
	require.True(t, params.MaxFrontrunPenalty.IsPositive(), "max frontrun penalty should be positive")
	require.True(t, params.AuctionDuration > 0, "auction duration should be positive")
	require.True(t, params.UserRedistributionPercentage <= 10000, "user redistribution should be <= 100%")
	require.True(t, params.ValidatorPercentage <= 10000, "validator percentage should be <= 100%")
	require.True(t, params.TreasuryPercentage <= 10000, "treasury percentage should be <= 100%")
	require.True(t, params.BurnPercentage <= 10000, "burn percentage should be <= 100%")

	// Verify total redistribution equals 100%
	total := params.UserRedistributionPercentage + params.ValidatorPercentage +
		params.TreasuryPercentage + params.BurnPercentage
	require.Equal(t, uint64(10000), total, "MEV redistribution should sum to 100%")

	// Verify strategy is valid
	require.NotEqual(t, economicspb.MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED,
		params.Strategy, "MEV strategy should be specified")

	// Validate parameters
	err := economicstypes.ValidateMEVParams(params)
	require.NoError(t, err, "default MEV params should be valid")
}

// TestDefaultWhaleProtectionParams verifies default whale protection parameters
func TestDefaultWhaleProtectionParams(t *testing.T) {
	params := economicstypes.DefaultWhaleProtectionParams()

	require.NotNil(t, params)

	// Verify expected default values
	require.True(t, params.Enabled, "whale protection should be enabled by default")
	require.True(t, params.MaxSingleTransfer.IsPositive(), "max single transfer should be positive")
	require.True(t, params.DailyTransferLimit.IsPositive(), "daily transfer limit should be positive")
	require.True(t, params.CooldownPeriod > 0, "cooldown period should be positive")
	require.True(t, params.MaxHoldingPercentage <= 10000, "max holding percentage should be <= 100%")
	require.True(t, params.LargeTxThreshold <= 10000, "large tx threshold should be <= 100%")
	require.NotNil(t, params.ExemptedAddresses, "exempted addresses should not be nil")

	// Validate parameters
	err := economicstypes.ValidateWhaleProtectionParams(params)
	require.NoError(t, err, "default whale protection params should be valid")
}

// TestDefaultLiquidityMiningParams verifies default liquidity mining parameters
func TestDefaultLiquidityMiningParams(t *testing.T) {
	params := economicstypes.DefaultLiquidityMiningParams()

	require.NotNil(t, params)

	// Verify expected default values
	require.True(t, params.Enabled, "liquidity mining should be enabled by default")
	require.True(t, params.TotalRewardsAllocated.IsPositive(), "total rewards should be positive")
	require.True(t, params.MaxRewardsPerEpoch.IsPositive(), "max rewards per epoch should be positive")
	require.True(t, params.EpochDurationBlocks > 0, "epoch duration should be positive")
	require.True(t, params.IrVerifiedMultiplier > 10000, "IR verified multiplier should be > 1x")

	// Validate max rewards per epoch doesn't exceed total
	require.True(t, params.MaxRewardsPerEpoch.LTE(params.TotalRewardsAllocated),
		"max rewards per epoch should be <= total rewards")

	// Validate parameters
	err := economicstypes.ValidateLiquidityMiningParams(params)
	require.NoError(t, err, "default liquidity mining params should be valid")
}

// TestDefaultTokenomicsParams verifies default tokenomics parameters
func TestDefaultTokenomicsParams(t *testing.T) {
	params := economicstypes.DefaultTokenomicsParams()

	require.NotNil(t, params)

	// Verify expected default values
	require.True(t, params.MaxSupply.IsPositive(), "max supply should be positive")
	require.True(t, params.TargetInflationRate <= 10000, "target inflation should be <= 100%")
	require.True(t, params.MinInflationRate <= 10000, "min inflation should be <= 100%")
	require.True(t, params.MaxInflationRate <= 10000, "max inflation should be <= 100%")
	require.True(t, params.InflationCheckInterval > 0, "inflation check interval should be positive")
	require.True(t, params.InflationAlertThreshold <= 10000, "inflation alert threshold should be <= 100%")

	// Verify inflation rate ordering
	require.True(t, params.MinInflationRate <= params.TargetInflationRate,
		"min inflation should be <= target inflation")
	require.True(t, params.TargetInflationRate <= params.MaxInflationRate,
		"target inflation should be <= max inflation")

	// Validate parameters
	err := economicstypes.ValidateTokenomicsParams(params)
	require.NoError(t, err, "default tokenomics params should be valid")
}

// TestAllDefaultParamsCombined verifies all default params work together
func TestAllDefaultParamsCombined(t *testing.T) {
	// Get default params from all sources
	fullParams := economicstypes.DefaultParams()
	feeParams := economicstypes.DefaultFeeParams()
	vestingParams := economicstypes.DefaultVestingParams()
	treasuryParams := economicstypes.DefaultTreasuryParams()
	govParams := economicstypes.DefaultGovernanceParams()
	mevParams := economicstypes.DefaultMEVParams()
	whaleParams := economicstypes.DefaultWhaleProtectionParams()
	liquidityParams := economicstypes.DefaultLiquidityMiningParams()
	tokenomicsParams := economicstypes.DefaultTokenomicsParams()

	// Verify individual params match combined params
	require.Equal(t, *feeParams, fullParams.Fees, "fee params should match")
	require.Equal(t, *vestingParams, fullParams.Vesting, "vesting params should match")
	require.Equal(t, *treasuryParams, fullParams.Treasury, "treasury params should match")
	require.Equal(t, *govParams, fullParams.Governance, "governance params should match")
	require.Equal(t, *mevParams, fullParams.Mev, "MEV params should match")
	require.Equal(t, *whaleParams, fullParams.WhaleProtection, "whale protection params should match")
	require.Equal(t, *liquidityParams, fullParams.LiquidityMining, "liquidity mining params should match")
	require.Equal(t, *tokenomicsParams, fullParams.Tokenomics, "tokenomics params should match")

	// Verify combined params are valid
	err := economicstypes.ValidateParams(fullParams)
	require.NoError(t, err, "combined default params should be valid")
}

// TestDefaultParamsConsistency verifies defaults are consistent across calls
func TestDefaultParamsConsistency(t *testing.T) {
	// Call default params multiple times
	params1 := economicstypes.DefaultParams()
	params2 := economicstypes.DefaultParams()

	// Verify they are equal
	require.Equal(t, params1.Fees, params2.Fees, "fee params should be consistent")
	require.Equal(t, params1.Vesting, params2.Vesting, "vesting params should be consistent")
	require.Equal(t, params1.Treasury, params2.Treasury, "treasury params should be consistent")
	require.Equal(t, params1.Governance, params2.Governance, "governance params should be consistent")
	require.Equal(t, params1.Mev, params2.Mev, "MEV params should be consistent")
	require.Equal(t, params1.WhaleProtection, params2.WhaleProtection, "whale protection params should be consistent")
	require.Equal(t, params1.LiquidityMining, params2.LiquidityMining, "liquidity mining params should be consistent")
	require.Equal(t, params1.Tokenomics, params2.Tokenomics, "tokenomics params should be consistent")
}

// TestDefaultsAreProductionReady verifies default parameters are sensible for production
func TestDefaultsAreProductionReady(t *testing.T) {
	params := economicstypes.DefaultParams()

	// Fee params should be reasonable
	require.True(t, params.Fees.MinFeeMultiplier > 0, "min fee multiplier should prevent free transactions")
	require.True(t, params.Fees.FeeBurnPercentage > 0, "should burn some fees to prevent inflation")

	// Vesting params should enforce reasonable lockups
	require.True(t, params.Vesting.MinVestingDuration > 0, "should enforce minimum vesting period")
	require.True(t, params.Vesting.EarlyUnlockPenalty > 0 || !params.Vesting.AllowEarlyUnlock,
		"should penalize early unlock or disallow it")

	// Treasury should have multisig security
	require.True(t, params.Treasury.MultisigThreshold >= 2, "should require multiple signatures")
	require.True(t, params.Treasury.TimelockDuration > 0, "should enforce timelock for safety")

	// Governance should be democratic but efficient
	require.True(t, params.Governance.Quorum >= 1000, "should require meaningful quorum (>= 10%)")
	require.True(t, params.Governance.Threshold >= 5000, "should require majority approval")
	require.True(t, params.Governance.VetoThreshold >= 3333, "should require meaningful minority for veto")

	// MEV protection should redistribute fairly
	require.True(t, params.Mev.Enabled, "should protect users from MEV")
	total := params.Mev.UserRedistributionPercentage + params.Mev.ValidatorPercentage +
		params.Mev.TreasuryPercentage + params.Mev.BurnPercentage
	require.Equal(t, uint64(10000), total, "MEV redistribution must sum to 100%")

	// Whale protection should prevent market manipulation
	require.True(t, params.WhaleProtection.Enabled, "should protect against whales")
	require.True(t, params.WhaleProtection.MaxHoldingPercentage < 10000, "should limit individual holdings")
	require.True(t, params.WhaleProtection.CooldownPeriod > 0, "should enforce cooldown on large txs")

	// Tokenomics should be sustainable
	require.True(t, params.Tokenomics.MaxInflationRate < 2000, "inflation should be reasonable (< 20%)")
	require.True(t, params.Tokenomics.MinInflationRate <= params.Tokenomics.TargetInflationRate,
		"min inflation should not exceed target rate")
	require.True(t, params.Tokenomics.MaxSupply.IsPositive(), "should have a max supply cap")
}
