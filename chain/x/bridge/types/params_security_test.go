package types_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestDefaultSecurityParams(t *testing.T) {
	params := types.DefaultSecurityParams()

	require.NotNil(t, params)
	require.False(t, params.EmergencyPaused)
	require.True(t, params.MinTransferAmount.IsPositive())
	require.True(t, params.MaxTransferAmount.GT(params.MinTransferAmount))
	require.Equal(t, 24*time.Hour, params.TimeLockDuration)
	require.True(t, params.CircuitBreakerEnabled)
	require.Equal(t, uint64(3), params.MinValidatorSignatures)
	require.False(t, params.WhitelistEnabled)
}

func TestSecurityParamSetPairs(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	require.NotNil(t, pairs)
	require.NotEmpty(t, pairs)
	// Should have 20 security parameters
	require.Equal(t, 20, len(pairs))
}

func TestValidateInt_Valid(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	// Find a pair that uses validateInt
	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeyMinTransferAmount) {
			// Test valid positive int
			err := pair.ValidatorFn(math.NewInt(1000))
			require.NoError(t, err)

			// Test zero
			err = pair.ValidatorFn(math.ZeroInt())
			require.NoError(t, err)
		}
	}
}

func TestValidateInt_Invalid(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeyMinTransferAmount) {
			// Test negative int
			err := pair.ValidatorFn(math.NewInt(-1000))
			require.Error(t, err)
			require.Contains(t, err.Error(), "cannot be negative")

			// Test invalid type
			err = pair.ValidatorFn("not an int")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid parameter type")

			err = pair.ValidatorFn(123)
			require.Error(t, err)
		}
	}
}

func TestValidateUint64_Valid(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeyMinValidatorSignatures) {
			// Test valid uint64
			err := pair.ValidatorFn(uint64(5))
			require.NoError(t, err)

			// Test zero
			err = pair.ValidatorFn(uint64(0))
			require.NoError(t, err)
		}
	}
}

func TestValidateUint64_Invalid(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeyMinValidatorSignatures) {
			// Test invalid types
			err := pair.ValidatorFn("not a uint64")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid parameter type")

			err = pair.ValidatorFn(int64(5))
			require.Error(t, err)

			err = pair.ValidatorFn(-5)
			require.Error(t, err)
		}
	}
}

func TestValidateDuration_Valid(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeyTimeLockDuration) {
			// Test valid duration
			err := pair.ValidatorFn(24 * time.Hour)
			require.NoError(t, err)

			// Test zero duration
			err = pair.ValidatorFn(time.Duration(0))
			require.NoError(t, err)
		}
	}
}

func TestValidateDuration_Invalid(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeyTimeLockDuration) {
			// Test negative duration
			err := pair.ValidatorFn(-24 * time.Hour)
			require.Error(t, err)
			require.Contains(t, err.Error(), "cannot be negative")

			// Test invalid type
			err = pair.ValidatorFn("not a duration")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid parameter type")

			err = pair.ValidatorFn(123)
			require.Error(t, err)
		}
	}
}

func TestValidateDec_Valid(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeySlashFractionInvalidProof) {
			// Test valid decimal (0.5)
			err := pair.ValidatorFn(math.LegacyNewDecWithPrec(50, 2))
			require.NoError(t, err)

			// Test zero
			err = pair.ValidatorFn(math.LegacyZeroDec())
			require.NoError(t, err)

			// Test one
			err = pair.ValidatorFn(math.LegacyOneDec())
			require.NoError(t, err)
		}
	}
}

func TestValidateDec_Invalid(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeySlashFractionInvalidProof) {
			// Test negative decimal
			err := pair.ValidatorFn(math.LegacyNewDec(-1))
			require.Error(t, err)
			require.Contains(t, err.Error(), "cannot be negative")

			// Test greater than 1
			err = pair.ValidatorFn(math.LegacyNewDec(2))
			require.Error(t, err)
			require.Contains(t, err.Error(), "cannot be greater than 1")

			// Test invalid type
			err = pair.ValidatorFn("not a decimal")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid parameter type")

			err = pair.ValidatorFn(0.5)
			require.Error(t, err)
		}
	}
}

func TestSecurityParamsAllFields(t *testing.T) {
	params := types.SecurityParams{
		EmergencyPaused:              true,
		MinTransferAmount:            math.NewInt(100),
		MaxTransferAmount:            math.NewInt(1000000),
		TimeLockDuration:             12 * time.Hour,
		TimeLockThreshold:            math.NewInt(50000),
		DailyWithdrawalLimit:         math.NewInt(100000),
		CircuitBreakerEnabled:        false,
		MaxHourlyVolume:              math.NewInt(500000),
		MaxFailedTransfersPerHour:    5,
		MinValidatorSignatures:       2,
		ValidatorRotationPeriod:      7 * 24 * time.Hour,
		SlashFractionInvalidProof:    math.LegacyNewDecWithPrec(10, 2),
		SlashFractionDoubleSign:      math.LegacyNewDecWithPrec(20, 2),
		SlashFractionDowntime:        math.LegacyNewDecWithPrec(5, 2),
		FraudProofReward:             math.NewInt(5000),
		FraudProofWindowDuration:     3 * 24 * time.Hour,
		FixedTransferFee:             math.NewInt(100),
		PercentageFeeBPS:             20,
		InsuranceFundContributionBPS: 1000,
		WhitelistEnabled:             true,
	}

	require.True(t, params.EmergencyPaused)
	require.Equal(t, math.NewInt(100), params.MinTransferAmount)
	require.Equal(t, math.NewInt(1000000), params.MaxTransferAmount)
	require.Equal(t, 12*time.Hour, params.TimeLockDuration)
	require.Equal(t, uint64(5), params.MaxFailedTransfersPerHour)
	require.True(t, params.WhitelistEnabled)
}

func TestSecurityParamsKeyConstants(t *testing.T) {
	// Verify all key constants are defined and not empty
	require.NotEmpty(t, types.KeyEmergencyPaused)
	require.NotEmpty(t, types.KeyMinTransferAmount)
	require.NotEmpty(t, types.KeyMaxTransferAmount)
	require.NotEmpty(t, types.KeyTimeLockDuration)
	require.NotEmpty(t, types.KeyTimeLockThreshold)
	require.NotEmpty(t, types.KeyDailyWithdrawalLimit)
	require.NotEmpty(t, types.KeyCircuitBreakerEnabled)
	require.NotEmpty(t, types.KeyMaxHourlyVolume)
	require.NotEmpty(t, types.KeyMaxFailedTransfersPerHour)
	require.NotEmpty(t, types.KeyMinValidatorSignatures)
	require.NotEmpty(t, types.KeyValidatorRotationPeriod)
	require.NotEmpty(t, types.KeySlashFractionInvalidProof)
	require.NotEmpty(t, types.KeySlashFractionDoubleSign)
	require.NotEmpty(t, types.KeySlashFractionDowntime)
	require.NotEmpty(t, types.KeyFraudProofReward)
	require.NotEmpty(t, types.KeyFraudProofWindowDuration)
	require.NotEmpty(t, types.KeyFixedTransferFee)
	require.NotEmpty(t, types.KeyPercentageFeeBPS)
	require.NotEmpty(t, types.KeyInsuranceFundContributionBPS)
	require.NotEmpty(t, types.KeyWhitelistEnabled)
}

func TestValidateBool_SecurityParams(t *testing.T) {
	params := types.DefaultSecurityParams()
	pairs := params.SecurityParamSetPairs()

	// Find the EmergencyPaused pair (uses validateBool)
	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeyEmergencyPaused) {
			err := pair.ValidatorFn(true)
			require.NoError(t, err)

			err = pair.ValidatorFn(false)
			require.NoError(t, err)

			// Invalid type
			err = pair.ValidatorFn("not a bool")
			require.Error(t, err)
		}
	}
}

func TestDefaultSecurityParams_Consistency(t *testing.T) {
	params := types.DefaultSecurityParams()

	// Verify min < max
	require.True(t, params.MinTransferAmount.LT(params.MaxTransferAmount))

	// Verify max transfer < daily limit
	require.True(t, params.MaxTransferAmount.LT(params.DailyWithdrawalLimit))

	// Verify slash fractions are between 0 and 1
	require.True(t, params.SlashFractionInvalidProof.GTE(math.LegacyZeroDec()))
	require.True(t, params.SlashFractionInvalidProof.LTE(math.LegacyOneDec()))
	require.True(t, params.SlashFractionDoubleSign.GTE(math.LegacyZeroDec()))
	require.True(t, params.SlashFractionDoubleSign.LTE(math.LegacyOneDec()))
	require.True(t, params.SlashFractionDowntime.GTE(math.LegacyZeroDec()))
	require.True(t, params.SlashFractionDowntime.LTE(math.LegacyOneDec()))

	// Verify durations are positive
	require.True(t, params.TimeLockDuration > 0)
	require.True(t, params.ValidatorRotationPeriod > 0)
	require.True(t, params.FraudProofWindowDuration > 0)
}

// TestMinConfirmationsSecurity verifies the critical security fix for single validator control
// This addresses TODO-035: Bridge Single Validator Can Drain Funds
func TestMinConfirmationsSecurity(t *testing.T) {
	t.Run("DefaultParams must require 3 confirmations", func(t *testing.T) {
		params := types.DefaultParams()
		require.Equal(t, types.DefaultMinConfirmations, params.MinConfirmations,
			"Default params must require 3 confirmations to prevent single validator control")
		require.Equal(t, uint64(3), params.MinConfirmations,
			"DefaultMinConfirmations constant must be 3")
	})

	t.Run("Validate rejects less than MinAllowedConfirmations", func(t *testing.T) {
		params := types.DefaultParams()

		// Test with 0 confirmations (CRITICAL vulnerability)
		params.MinConfirmations = 0
		err := params.Validate()
		require.Error(t, err, "Must reject 0 confirmations")
		require.Contains(t, err.Error(), "MinConfirmations must be >= 2",
			"Error message must reference minimum requirement")

		// Test with 1 confirmation (CRITICAL vulnerability - single validator control)
		params.MinConfirmations = 1
		err = params.Validate()
		require.Error(t, err, "Must reject single validator confirmation")
		require.Contains(t, err.Error(), "prevents single validator control",
			"Error message must explain security rationale")

		// Test with 2 confirmations (minimum allowed)
		params.MinConfirmations = 2
		err = params.Validate()
		require.NoError(t, err, "Must accept 2 confirmations (minimum allowed)")

		// Test with 3 confirmations (default)
		params.MinConfirmations = 3
		err = params.Validate()
		require.NoError(t, err, "Must accept 3 confirmations (default)")

		// Test with higher values
		params.MinConfirmations = 10
		err = params.Validate()
		require.NoError(t, err, "Must accept higher confirmation requirements")
	})

	t.Run("MinAllowedConfirmations constant is 2", func(t *testing.T) {
		require.Equal(t, uint64(2), types.MinAllowedConfirmations,
			"MinAllowedConfirmations must be 2 to prevent single validator control")
	})

	t.Run("validateMinConfirmations enforces minimum", func(t *testing.T) {
		params := types.DefaultParams()
		pairs := params.ParamSetPairs()

		// Find the MinConfirmations validator function
		var minConfValidator func(interface{}) error
		for _, pair := range pairs {
			if string(pair.Key) == "MinConfirmations" {
				minConfValidator = pair.ValidatorFn
				break
			}
		}

		require.NotNil(t, minConfValidator, "MinConfirmations must have a validator function")

		// Test validator rejects values below minimum
		err := minConfValidator(uint64(0))
		require.Error(t, err, "Validator must reject 0")

		err = minConfValidator(uint64(1))
		require.Error(t, err, "Validator must reject 1")

		// Test validator accepts minimum and above
		err = minConfValidator(uint64(2))
		require.NoError(t, err, "Validator must accept 2")

		err = minConfValidator(uint64(3))
		require.NoError(t, err, "Validator must accept 3")

		err = minConfValidator(uint64(100))
		require.NoError(t, err, "Validator must accept higher values")
	})
}

// TestSecurityRegressionPrevention ensures the vulnerability cannot be reintroduced
func TestSecurityRegressionPrevention(t *testing.T) {
	t.Run("Default params must never allow single validator control", func(t *testing.T) {
		params := types.DefaultParams()

		// This test will fail if someone changes the default back to 1
		require.GreaterOrEqual(t, params.MinConfirmations, types.MinAllowedConfirmations,
			"SECURITY REGRESSION: Default params must require at least MinAllowedConfirmations")

		require.GreaterOrEqual(t, params.MinConfirmations, uint64(3),
			"SECURITY REGRESSION: Default params should require at least 3 confirmations")
	})

	t.Run("MinAllowedConfirmations constant must prevent single validator", func(t *testing.T) {
		// This test will fail if someone changes MinAllowedConfirmations to less than 2
		require.GreaterOrEqual(t, types.MinAllowedConfirmations, uint64(2),
			"SECURITY REGRESSION: MinAllowedConfirmations must be at least 2 to prevent single validator control")
	})

	t.Run("Validation must enforce minimum at protocol level", func(t *testing.T) {
		// Even if someone tries to set params with MinConfirmations=1, validation must reject it
		invalidParams := types.Params{
			BridgeEnabled:                true,
			MinConfirmations:             1, // INVALID
			BridgeFeeBasisPoints:         30,
			MaxTransferAmount:            "1000000000",
			ValidatorThresholdPercentage: 67,
		}

		err := invalidParams.Validate()
		require.Error(t, err,
			"SECURITY REGRESSION: Validation must reject MinConfirmations=1")
		require.Contains(t, err.Error(), "prevents single validator control",
			"Error message must explain security rationale")
	})
}
