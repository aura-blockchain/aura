// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultParams returns default prevalidation parameters
func DefaultParams() *Params {
	return &Params{
		Enabled:                    true,
		SchedulerConfig:            &SchedulerConfig{},
		AutoScalingConfig:          &AutoScalingConfig{},
		CacheStrategy:              CacheStrategy_CACHE_STRATEGY_LRU,
		MaxCacheSize:               10000,
		ExpiryHours:                24,
		EncryptionAlgorithm:        "AES-256-GCM",
		ControlGroupPercentage:     math.LegacyNewDec(10), // 10.0 as deterministic decimal
		MinConfidenceScore:         50,
		EnergyCostPerValidationKwh: math.LegacyNewDecWithPrec(1, 3), // 0.001 as deterministic decimal
		EnergyCostPerExecutionKwh:  math.LegacyNewDecWithPrec(5, 3), // 0.005 as deterministic decimal
		MetricsEnabled:             true,
		DetailedLogging:            false,
		MaxValidationAttempts:      3,
		RetryDelaySeconds:          5,
	}
}

// ValidateParams validates prevalidation parameters
func ValidateParams(params *Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	// Normalize decimal fields to avoid nil deref
	if params.ControlGroupPercentage.IsNil() {
		params.ControlGroupPercentage = math.LegacyNewDec(0)
	}
	if params.EnergyCostPerValidationKwh.IsNil() {
		params.EnergyCostPerValidationKwh = math.LegacyNewDec(0)
	}
	if params.EnergyCostPerExecutionKwh.IsNil() {
		params.EnergyCostPerExecutionKwh = math.LegacyNewDec(0)
	}

	// Validate scheduler config
	if params.SchedulerConfig != nil {
		if params.SchedulerConfig.Enabled {
			if params.SchedulerConfig.RunIntervalMinutes == 0 {
				return fmt.Errorf("run_interval_minutes must be greater than 0 when scheduler is enabled")
			}
			if params.SchedulerConfig.Timezone == "" {
				return fmt.Errorf("timezone must be provided when scheduler is enabled")
			}
			if params.SchedulerConfig.MaxPerRun == 0 {
				return fmt.Errorf("max_per_run must be greater than 0 when scheduler is enabled")
			}
		}
	}

	// Validate auto-scaling config
	if params.AutoScalingConfig != nil {
		if params.AutoScalingConfig.TargetCacheHitRate.IsNil() {
			params.AutoScalingConfig.TargetCacheHitRate = math.LegacyNewDec(0)
		}
		if params.AutoScalingConfig.MinCacheHitRate.IsNil() {
			params.AutoScalingConfig.MinCacheHitRate = math.LegacyNewDec(0)
		}
		if params.AutoScalingConfig.ScaleUpFactor.IsNil() {
			params.AutoScalingConfig.ScaleUpFactor = math.LegacyNewDec(0)
		}
		if params.AutoScalingConfig.ScaleDownFactor.IsNil() {
			params.AutoScalingConfig.ScaleDownFactor = math.LegacyNewDec(0)
		}

		if params.AutoScalingConfig.Enabled {
			if params.AutoScalingConfig.CooldownMinutes == 0 {
				return fmt.Errorf("cooldown_minutes must be greater than 0 when auto-scaling is enabled")
			}
			if params.AutoScalingConfig.EvaluationPeriodHours == 0 {
				return fmt.Errorf("evaluation_period_hours must be greater than 0 when auto-scaling is enabled")
			}
		}

		if params.AutoScalingConfig.MinCacheHitRate.IsNegative() || params.AutoScalingConfig.MinCacheHitRate.GT(math.LegacyNewDec(100)) {
			return fmt.Errorf("min_cache_hit_rate must be between 0.0 and 100.0")
		}
		if params.AutoScalingConfig.TargetCacheHitRate.IsNegative() || params.AutoScalingConfig.TargetCacheHitRate.GT(math.LegacyNewDec(100)) {
			return fmt.Errorf("target_cache_hit_rate must be between 0.0 and 100.0")
		}
		if params.AutoScalingConfig.ScaleUpFactor.IsNegative() || params.AutoScalingConfig.ScaleDownFactor.IsNegative() {
			return fmt.Errorf("scale factors cannot be negative")
		}
	}

	// Validate control group percentage
	if params.ControlGroupPercentage.IsNegative() || params.ControlGroupPercentage.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("control_group_percentage must be between 0.0 and 100.0")
	}

	// Validate cache size
	if params.MaxCacheSize == 0 {
		return fmt.Errorf("max_cache_size must be greater than 0")
	}

	// Validate expiry hours
	if params.ExpiryHours == 0 {
		return fmt.Errorf("expiry_hours must be greater than 0")
	}

	return nil
}
