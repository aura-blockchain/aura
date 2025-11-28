package types

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"google.golang.org/protobuf/types/known/durationpb"
)

// DefaultParams returns default validator security parameters
func DefaultParams() *ValidatorSecurityParams {
	return &ValidatorSecurityParams{
		DoubleSignSlashFraction: "0.05",                        // 5% slash for double signing
		DowntimeSlashFraction:   "0.01",                        // 1% slash for downtime
		SignedBlocksWindow:      1000,                          // Track last 1000 blocks
		MinSignedPerWindow:      "0.5",                         // Must sign 50% of blocks in window
		DowntimeJailDuration:    durationpb.New(1 * time.Hour), // 1 hour jail by default
		MinimumStakeAmount:      "1000000",                     // Minimum stake amount
		EnableGeoDistribution:   true,
		MaxValidatorsPerRegion:  50,
		RequireSentryNodes:      true,
		MinSentryNodes:          2,
		MonitoringInterval:      durationpb.New(1 * time.Minute),
		EnableAutoFailover:      true,
		FailoverTimeout:         durationpb.New(5 * time.Minute),
	}
}

// ValidateParams validates validator security parameters
func ValidateParams(params *ValidatorSecurityParams) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	// Slash fractions must be decimals between 0 and 1 (inclusive of 0, exclusive of 1)
	doubleSlash, err := sdkmath.LegacyNewDecFromStr(params.DoubleSignSlashFraction)
	if err != nil {
		return fmt.Errorf("invalid double_sign_slash_fraction: %w", err)
	}
	if doubleSlash.IsNegative() || doubleSlash.GTE(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("double_sign_slash_fraction must be between 0 and 1")
	}

	downtimeSlash, err := sdkmath.LegacyNewDecFromStr(params.DowntimeSlashFraction)
	if err != nil {
		return fmt.Errorf("invalid downtime_slash_fraction: %w", err)
	}
	if downtimeSlash.IsNegative() || downtimeSlash.GTE(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("downtime_slash_fraction must be between 0 and 1")
	}

	if params.SignedBlocksWindow <= 0 {
		return fmt.Errorf("signed_blocks_window must be positive")
	}

	minSigned, err := sdkmath.LegacyNewDecFromStr(params.MinSignedPerWindow)
	if err != nil {
		return fmt.Errorf("invalid min_signed_per_window: %w", err)
	}
	if minSigned.IsNegative() || minSigned.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("min_signed_per_window must be between 0 and 1")
	}

	if params.DowntimeJailDuration == nil || params.DowntimeJailDuration.AsDuration() <= 0 {
		return fmt.Errorf("downtime_jail_duration must be positive")
	}

	if params.MinimumStakeAmount == "" {
		return fmt.Errorf("minimum_stake_amount cannot be empty")
	}
	minStake, ok := sdkmath.NewIntFromString(params.MinimumStakeAmount)
	if !ok || !minStake.IsPositive() {
		return fmt.Errorf("minimum_stake_amount must be a positive integer")
	}

	if params.EnableGeoDistribution && params.MaxValidatorsPerRegion <= 0 {
		return fmt.Errorf("max_validators_per_region must be positive when geo distribution is enabled")
	}

	if params.RequireSentryNodes && params.MinSentryNodes <= 0 {
		return fmt.Errorf("min_sentry_nodes must be positive when sentry nodes are required")
	}

	if params.MonitoringInterval == nil || params.MonitoringInterval.AsDuration() <= 0 {
		return fmt.Errorf("monitoring_interval must be positive")
	}

	if params.EnableAutoFailover {
		if params.FailoverTimeout == nil || params.FailoverTimeout.AsDuration() <= 0 {
			return fmt.Errorf("failover_timeout must be positive when auto failover is enabled")
		}
	}

	return nil
}
