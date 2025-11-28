package types

import (
	"fmt"
)

// DefaultParams returns default prevalidation parameters
func DefaultParams() *Params {
	return &Params{
		Enabled:                      true,
		SchedulerConfig:              &SchedulerConfig{},
		AutoScalingConfig:            &AutoScalingConfig{},
		CacheStrategy:                CacheStrategy_CACHE_STRATEGY_LRU,
		MaxCacheSize:                 10000,
		ExpiryHours:                  24,
		EncryptionAlgorithm:          "AES-256-GCM",
		ControlGroupPercentage:       10.0,
		MinConfidenceScore:           50,
		EnergyCostPerValidationKwh:   0.001,
		EnergyCostPerExecutionKwh:    0.005,
		MetricsEnabled:               true,
		DetailedLogging:              false,
		MaxValidationAttempts:        3,
		RetryDelaySeconds:            5,
	}
}

// ValidateParams validates prevalidation parameters
func ValidateParams(params *Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
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
	if params.ControlGroupPercentage < 0.0 || params.ControlGroupPercentage > 100.0 {
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

