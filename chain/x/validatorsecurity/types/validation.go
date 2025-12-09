package types

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
)

// DefaultParams returns default validator security parameters
func DefaultParams() *ValidatorSecurityParams {
	return &ValidatorSecurityParams{
		DoubleSignSlashFraction: sdkmath.LegacyNewDecWithPrec(5, 2),  // 5% slash for double signing (5/100 = 0.05)
		DowntimeSlashFraction:   sdkmath.LegacyNewDecWithPrec(1, 2),  // 1% slash for downtime (1/100 = 0.01)
		SignedBlocksWindow:      1000,                                // Track last 1000 blocks
		MinSignedPerWindow:      sdkmath.LegacyNewDecWithPrec(5, 1),  // Must sign 50% of blocks in window (5/10 = 0.5)
		DowntimeJailDuration:    1 * time.Hour,                       // 1 hour jail by default
		MinimumStakeAmount:      sdkmath.NewInt(1000000),             // Minimum stake amount
		EnableGeoDistribution:   true,
		MaxValidatorsPerRegion:  50,
		RequireSentryNodes:      true,
		MinSentryNodes:          2,
		MonitoringInterval:      1 * time.Minute,
		EnableAutoFailover:      true,
		FailoverTimeout:         5 * time.Minute,
	}
}

// ValidateParams validates validator security parameters
func ValidateParams(params *ValidatorSecurityParams) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	// Slash fractions must be decimals between 0 and 1 (inclusive of 0, exclusive of 1)
	// Fields are already LegacyDec type, validate directly
	if params.DoubleSignSlashFraction.IsNegative() || params.DoubleSignSlashFraction.GTE(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("double_sign_slash_fraction must be between 0 and 1")
	}

	if params.DowntimeSlashFraction.IsNegative() || params.DowntimeSlashFraction.GTE(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("downtime_slash_fraction must be between 0 and 1")
	}

	if params.SignedBlocksWindow <= 0 {
		return fmt.Errorf("signed_blocks_window must be positive")
	}

	if params.MinSignedPerWindow.IsNegative() || params.MinSignedPerWindow.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("min_signed_per_window must be between 0 and 1")
	}

	// Duration fields are time.Duration, not *durationpb.Duration
	if params.DowntimeJailDuration <= 0 {
		return fmt.Errorf("downtime_jail_duration must be positive")
	}

	// MinimumStakeAmount is math.Int, not string
	if !params.MinimumStakeAmount.IsPositive() {
		return fmt.Errorf("minimum_stake_amount must be a positive integer")
	}

	if params.EnableGeoDistribution && params.MaxValidatorsPerRegion <= 0 {
		return fmt.Errorf("max_validators_per_region must be positive when geo distribution is enabled")
	}

	if params.RequireSentryNodes && params.MinSentryNodes <= 0 {
		return fmt.Errorf("min_sentry_nodes must be positive when sentry nodes are required")
	}

	if params.MonitoringInterval <= 0 {
		return fmt.Errorf("monitoring_interval must be positive")
	}

	if params.EnableAutoFailover {
		if params.FailoverTimeout <= 0 {
			return fmt.Errorf("failover_timeout must be positive when auto failover is enabled")
		}
	}

	return nil
}
