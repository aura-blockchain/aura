package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	pb "github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Default parameter values
var (
	DefaultDoubleSignSlashFraction = math.LegacyNewDecWithPrec(5, 2) // 5%
	DefaultDowntimeSlashFraction   = math.LegacyNewDecWithPrec(1, 4) // 0.01%
	DefaultSignedBlocksWindow      = int64(10000)
	DefaultMinSignedPerWindow      = math.LegacyNewDecWithPrec(5, 1) // 50%
	DefaultDowntimeJailDuration    = time.Hour * 24                  // 24 hours
	DefaultMinimumStakeAmount      = math.NewInt(1000000000000)      // 1000 tokens with 9 decimals
	DefaultEnableGeoDistribution   = true
	DefaultMaxValidatorsPerRegion  = int32(10)
	DefaultRequireSentryNodes      = true
	DefaultMinSentryNodes          = int32(2)
	DefaultMonitoringInterval      = time.Minute * 5
	DefaultEnableAutoFailover      = true
	DefaultFailoverTimeout         = time.Minute * 10
)

// DefaultParams returns default validator security parameters
func DefaultParams() *pb.ValidatorSecurityParams {
	return &pb.ValidatorSecurityParams{
		DoubleSignSlashFraction: DefaultDoubleSignSlashFraction.String(),
		DowntimeSlashFraction:   DefaultDowntimeSlashFraction.String(),
		SignedBlocksWindow:      DefaultSignedBlocksWindow,
		MinSignedPerWindow:      DefaultMinSignedPerWindow.String(),
		DowntimeJailDuration:    durationpb.New(DefaultDowntimeJailDuration),
		MinimumStakeAmount:      DefaultMinimumStakeAmount.String(),
		EnableGeoDistribution:   DefaultEnableGeoDistribution,
		MaxValidatorsPerRegion:  DefaultMaxValidatorsPerRegion,
		RequireSentryNodes:      DefaultRequireSentryNodes,
		MinSentryNodes:          DefaultMinSentryNodes,
		MonitoringInterval:      durationpb.New(DefaultMonitoringInterval),
		EnableAutoFailover:      DefaultEnableAutoFailover,
		FailoverTimeout:         durationpb.New(DefaultFailoverTimeout),
	}
}

// ValidateParams performs validation of validator security parameters
func ValidateParams(params *pb.ValidatorSecurityParams) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	// Validate slash fractions
	doubleSignSlash, err := math.LegacyNewDecFromStr(params.DoubleSignSlashFraction)
	if err != nil {
		return fmt.Errorf("invalid double_sign_slash_fraction: %w", err)
	}
	if doubleSignSlash.IsNegative() || doubleSignSlash.GT(math.LegacyOneDec()) {
		return fmt.Errorf("double_sign_slash_fraction must be between 0 and 1")
	}

	downtimeSlash, err := math.LegacyNewDecFromStr(params.DowntimeSlashFraction)
	if err != nil {
		return fmt.Errorf("invalid downtime_slash_fraction: %w", err)
	}
	if downtimeSlash.IsNegative() || downtimeSlash.GT(math.LegacyOneDec()) {
		return fmt.Errorf("downtime_slash_fraction must be between 0 and 1")
	}

	// Validate signed blocks window
	if params.SignedBlocksWindow <= 0 {
		return fmt.Errorf("signed_blocks_window must be positive")
	}

	// Validate min signed per window
	minSigned, err := math.LegacyNewDecFromStr(params.MinSignedPerWindow)
	if err != nil {
		return fmt.Errorf("invalid min_signed_per_window: %w", err)
	}
	if minSigned.IsNegative() || minSigned.GT(math.LegacyOneDec()) {
		return fmt.Errorf("min_signed_per_window must be between 0 and 1")
	}

	// Validate downtime jail duration
	if params.DowntimeJailDuration == nil || params.DowntimeJailDuration.AsDuration() <= 0 {
		return fmt.Errorf("downtime_jail_duration must be positive")
	}

	// Validate minimum stake amount
	minStake, ok := math.NewIntFromString(params.MinimumStakeAmount)
	if !ok || minStake.IsNegative() {
		return fmt.Errorf("invalid minimum_stake_amount")
	}

	// Validate geo distribution settings
	if params.EnableGeoDistribution {
		if params.MaxValidatorsPerRegion <= 0 {
			return fmt.Errorf("max_validators_per_region must be positive when geo distribution is enabled")
		}
	}

	// Validate sentry node settings
	if params.RequireSentryNodes {
		if params.MinSentryNodes <= 0 {
			return fmt.Errorf("min_sentry_nodes must be positive when sentry nodes are required")
		}
	}

	// Validate monitoring interval
	if params.MonitoringInterval == nil || params.MonitoringInterval.AsDuration() <= 0 {
		return fmt.Errorf("monitoring_interval must be positive")
	}

	// Validate failover settings
	if params.EnableAutoFailover {
		if params.FailoverTimeout == nil || params.FailoverTimeout.AsDuration() <= 0 {
			return fmt.Errorf("failover_timeout must be positive when auto failover is enabled")
		}
	}

	return nil
}
