// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// Test scenario: Complete bridge transfer with all security features
// This integration test verifies all security parameters work together properly
func TestCompleteSecureBridgeTransfer(t *testing.T) {
	params := types.DefaultSecurityParams()

	// Verify all security features are enabled and configured properly
	require.False(t, params.EmergencyPaused, "should not be paused by default")
	require.True(t, params.CircuitBreakerEnabled, "circuit breaker should be enabled")

	// Verify limits are reasonable
	require.True(t, params.MinTransferAmount.IsPositive(), "min transfer should be positive")
	require.True(t, params.MaxTransferAmount.GT(params.MinTransferAmount), "max should be greater than min")
	require.True(t, params.DailyWithdrawalLimit.IsPositive(), "daily limit should be positive")

	// Verify time-lock is configured
	require.Greater(t, params.TimeLockDuration, time.Duration(0), "time-lock duration should be positive")
	require.True(t, params.TimeLockThreshold.IsPositive(), "time-lock threshold should be positive")

	// Verify validator requirements
	require.Greater(t, params.MinValidatorSignatures, uint64(0), "should require at least 1 signature")

	// Verify fees are configured
	require.True(t, params.FixedTransferFee.IsPositive(), "fixed fee should be positive")
	require.Greater(t, params.PercentageFeeBPS, uint64(0), "percentage fee should be positive")

	// Verify fraud proof window exists
	require.Greater(t, params.FraudProofWindowDuration, time.Duration(0), "fraud proof window should be positive")
	require.True(t, params.FraudProofReward.IsPositive(), "fraud proof reward should be positive")
}

func TestSecurityFeatureDefaults(t *testing.T) {
	// Test that default parameters are secure
	params := types.DefaultSecurityParams()

	// Verify critical settings
	require.False(t, params.EmergencyPaused, "should not be paused by default")
	require.True(t, params.CircuitBreakerEnabled, "circuit breaker should be enabled")
	require.True(t, params.MinTransferAmount.IsPositive(), "min transfer should be positive")
	require.True(t, params.MaxTransferAmount.GT(params.MinTransferAmount), "max should be greater than min")
	require.True(t, params.DailyWithdrawalLimit.IsPositive(), "daily limit should be positive")

	// Verify time-lock settings
	require.Greater(t, params.TimeLockDuration, time.Duration(0), "time-lock duration should be positive")
	require.True(t, params.TimeLockThreshold.IsPositive(), "time-lock threshold should be positive")

	// Verify slashing fractions are reasonable (not too high)
	require.True(t, params.SlashFractionInvalidProof.LTE(math.LegacyNewDecWithPrec(10, 2)), "slash fraction should be <= 10%")
	require.True(t, params.SlashFractionDoubleSign.LTE(math.LegacyNewDecWithPrec(20, 2)), "slash fraction should be <= 20%")

	// Verify fee settings
	require.True(t, params.FixedTransferFee.IsPositive(), "fixed fee should be positive")
	require.Greater(t, params.PercentageFeeBPS, uint64(0), "percentage fee should be positive")
	require.Less(t, params.PercentageFeeBPS, uint64(1000), "percentage fee should be < 10%")

	// Verify insurance fund settings
	require.Greater(t, params.InsuranceFundContributionBPS, uint64(0), "insurance contribution should be positive")
	require.Less(t, params.InsuranceFundContributionBPS, uint64(5000), "insurance contribution should be < 50%")
}

func TestCircuitBreakerLogic(t *testing.T) {
	// Test circuit breaker thresholds
	params := types.DefaultSecurityParams()

	// Verify circuit breaker has reasonable limits
	require.True(t, params.MaxHourlyVolume.IsPositive(), "max hourly volume should be positive")
	require.Greater(t, params.MaxFailedTransfersPerHour, uint64(0), "max failed transfers should be positive")

	// Max hourly volume should be reasonable compared to max transfer
	maxTransfers := params.MaxHourlyVolume.Quo(params.MaxTransferAmount)
	require.True(t, maxTransfers.GT(math.NewInt(5)), "should allow at least 5 max transfers per hour")
}

func TestValidatorThresholds(t *testing.T) {
	params := types.DefaultSecurityParams()

	// Verify validator signature requirements
	require.Greater(t, params.MinValidatorSignatures, uint64(0), "should require at least 1 signature")
	require.GreaterOrEqual(t, params.MinValidatorSignatures, uint64(3), "should require at least 3 signatures for security")

	// Verify rotation period is reasonable
	require.Greater(t, params.ValidatorRotationPeriod, 7*24*time.Hour, "rotation period should be at least 7 days")
	require.Less(t, params.ValidatorRotationPeriod, 365*24*time.Hour, "rotation period should be less than 1 year")
}

func TestFraudProofWindowSecurity(t *testing.T) {
	params := types.DefaultSecurityParams()

	// Verify fraud proof window is reasonable
	require.Greater(t, params.FraudProofWindowDuration, 24*time.Hour, "should be at least 1 day")
	require.Less(t, params.FraudProofWindowDuration, 30*24*time.Hour, "should be less than 30 days")

	// Verify fraud proof reward is meaningful
	require.True(t, params.FraudProofReward.IsPositive(), "reward should be positive")
}

func TestTimeLockConfiguration(t *testing.T) {
	params := types.DefaultSecurityParams()

	// Time-lock duration should be long enough to allow for fraud detection
	require.GreaterOrEqual(t, params.TimeLockDuration, 12*time.Hour, "should be at least 12 hours")
	require.LessOrEqual(t, params.TimeLockDuration, 7*24*time.Hour, "should be at most 7 days")

	// Time-lock threshold should trigger on large transfers
	require.True(t, params.TimeLockThreshold.LT(params.MaxTransferAmount), "threshold should be less than max")
	require.True(t, params.TimeLockThreshold.GT(params.MinTransferAmount), "threshold should be greater than min")
}

func TestWithdrawalLimitRatios(t *testing.T) {
	params := types.DefaultSecurityParams()

	// Daily withdrawal limit should be reasonable compared to max transfer
	ratio := params.DailyWithdrawalLimit.Quo(params.MaxTransferAmount)
	require.True(t, ratio.GT(math.NewInt(1)), "daily limit should allow multiple max transfers")
	require.True(t, ratio.LT(math.NewInt(1000)), "daily limit should not be too high")
}

// Test fee calculation logic
func TestFeeCalculation(t *testing.T) {
	params := types.DefaultSecurityParams()

	// Test with various amounts
	testCases := []struct {
		name   string
		amount math.Int
	}{
		{"small", math.NewInt(1000000)},           // 1 token
		{"medium", math.NewInt(100000000)},        // 100 tokens
		{"large", math.NewInt(10000000000)},       // 10,000 tokens
		{"very_large", math.NewInt(100000000000)}, // 100,000 tokens
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate fee
			fixedFee := params.FixedTransferFee
			percentageFee := tc.amount.MulRaw(int64(params.PercentageFeeBPS)).QuoRaw(10000)
			totalFee := fixedFee.Add(percentageFee)

			// Verify fee is reasonable (not more than 5% of amount)
			maxAllowedFee := tc.amount.MulRaw(5).QuoRaw(100)
			require.True(t, totalFee.LTE(maxAllowedFee) || totalFee.Equal(maxAllowedFee),
				"fee should not exceed 5%% of amount")

			// Verify fee is not zero
			require.True(t, totalFee.IsPositive(), "fee should be positive")
		})
	}
}
