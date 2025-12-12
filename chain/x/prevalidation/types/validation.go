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
		// Add scheduler validation if needed
	}

	// Validate auto-scaling config
	if params.AutoScalingConfig != nil {
		// Add auto-scaling validation if needed
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
