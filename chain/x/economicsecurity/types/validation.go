package types

import (
	"fmt"
)

// Constants
const (
	BasisPoints = 10000 // 100% = 10000 basis points (1 basis point = 0.01%)
)

// Enum constant aliases for backward compatibility
const (
	// InflationAlertType aliases
	InflationAlertTypeUnspecified = InflationAlertType_INFLATION_ALERT_TYPE_UNSPECIFIED
	InflationAlertTypeAboveTarget = InflationAlertType_INFLATION_ALERT_TYPE_ABOVE_TARGET
	InflationAlertTypeBelowTarget = InflationAlertType_INFLATION_ALERT_TYPE_BELOW_TARGET
	InflationAlertTypeAboveMax    = InflationAlertType_INFLATION_ALERT_TYPE_ABOVE_MAX
	InflationAlertTypeBelowMin    = InflationAlertType_INFLATION_ALERT_TYPE_BELOW_MIN
	InflationAlertTypeRapidChange = InflationAlertType_INFLATION_ALERT_TYPE_RAPID_CHANGE

	// AlertSeverity aliases
	AlertSeverityUnspecified = AlertSeverity_ALERT_SEVERITY_UNSPECIFIED
	AlertSeverityInfo        = AlertSeverity_ALERT_SEVERITY_INFO
	AlertSeverityWarning     = AlertSeverity_ALERT_SEVERITY_WARNING
	AlertSeverityCritical    = AlertSeverity_ALERT_SEVERITY_CRITICAL
	AlertSeverityEmergency   = AlertSeverity_ALERT_SEVERITY_EMERGENCY

	// MEVRedistributionStrategy aliases
	MEVStrategyUnspecified            = MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED
	MEVStrategyProportionalToStake    = MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE
	MEVStrategyProportionalToActivity = MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_ACTIVITY
	MEVStrategyEqualDistribution      = MEVRedistributionStrategy_MEV_STRATEGY_EQUAL_DISTRIBUTION
	MEVStrategyIRWeighted             = MEVRedistributionStrategy_MEV_STRATEGY_IR_WEIGHTED

	// VestingType aliases
	VestingTypeUnspecified     = VestingType_VESTING_TYPE_UNSPECIFIED
	VestingTypeLinear          = VestingType_VESTING_TYPE_LINEAR
	VestingTypeMilestone       = VestingType_VESTING_TYPE_MILESTONE
	VestingTypeCliffThenLinear = VestingType_VESTING_TYPE_CLIFF_THEN_LINEAR
)

// DefaultParams returns default economics security parameters
func DefaultParams() *Params {
	return &Params{
		Tokenomics: &TokenomicsConfig{
			MaxSupply:          "1000000000", // 1 billion tokens
			CirculatingSupply:  "100000000",  // 100 million tokens
			InflationRate:      500,          // 5% (500 basis points)
			TargetInflationRate: 500,         // 5%
			MinInflationRate:   100,          // 1%
			MaxInflationRate:   2000,         // 20%
		},
		WhaleProtection: &WhaleProtection{
			Enabled:              true,
			MaxHoldingPercentage: 500,  // 5% (500 basis points)
			MaxTxPercentage:      100,  // 1% (100 basis points)
			LargeTxCooldown:      86400, // 24 hours
			LargeTxThreshold:     100,  // 1%
			ExemptedAddresses:    []string{},
		},
		TransferTax: &TransferTaxConfig{
			Enabled:                  false,
			BaseTaxRate:              10,  // 0.1% (10 basis points)
			TaxRecipient:             "",
			BurnPercentage:           5000, // 50% (5000 basis points)
			TreasuryPercentage:       5000, // 50%
			RedistributePercentage:   0,
			ExemptedAddresses:        []string{},
			DynamicAdjustmentEnabled: false,
			MaxTaxRate:               100, // 1%
			MinTaxRate:               1,   // 0.01%
		},
		LiquidityMining: &LiquidityMiningConfig{
			Enabled:                 true,
			TotalRewardsAllocated:   "1000000",
			TotalRewardsDistributed: "0",
			MaxRewardsPerEpoch:      "1000",
			CurrentEpoch:            0,
			EpochDurationBlocks:     10000,
			LastDistributionHeight:  0,
			IrVerifiedMultiplier:    15000, // 1.5x (15000 basis points)
		},
		Governance: &GovernanceConfig{
			MinProposalStake:         "1000",
			QuadraticVotingEnabled:   false,
			VoteLockingEnabled:       true,
			MinLockDuration:          604800,  // 7 days
			MaxLockDuration:          31536000, // 1 year
			LockMultiplierPerYear:    10000,   // 1x per year (10000 basis points)
			ProposalDeposit:          "1000",
			QuorumPercentage:         3333, // 33.33%
			PassThresholdPercentage:  5000, // 50%
		},
		TreasuryMultisig: &TreasuryMultisig{
			TreasuryAddress:  "",
			Threshold:        3,
			Signers:          []string{},
			SpendingLimit:    "10000",
			TimelockDuration: 604800, // 7 days
		},
		DynamicFees: &DynamicFeeConfig{
			Enabled:           true,
			BaseFee:           "100",
			CurrentMultiplier: 10000, // 1x (10000 basis points)
			MinMultiplier:     5000,  // 0.5x
			MaxMultiplier:     20000, // 2x
			TargetUtilization: 7000,  // 70%
			AdjustmentSpeed:   100,   // 1% adjustment
			RecentUtilization: []uint64{},
			UtilizationWindow: 20,    // Track last 20 blocks
		},
		Mev: &MEVConfig{
			Enabled:                       true,
			UserRedistributionPercentage: 4000, // 40%
			ValidatorPercentage:           5000, // 50%
			TreasuryPercentage:            1000, // 10%
			BurnPercentage:                0,
			TotalMevCaptured:              "0",
			TotalMevRedistributed:         "0",
			Strategy:                      0, // Unspecified
		},
		InflationAlertThreshold: 100,   // 1% deviation (100 basis points)
		InflationCheckInterval:  10000, // Check every 10000 blocks
	}
}

// ValidateParams validates economics security parameters
func ValidateParams(params *Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	// Validate Tokenomics
	if params.Tokenomics != nil {
		if params.Tokenomics.MaxSupply == "" {
			return fmt.Errorf("max_supply cannot be empty")
		}
		if params.Tokenomics.CirculatingSupply == "" {
			return fmt.Errorf("circulating_supply cannot be empty")
		}
		if params.Tokenomics.MaxInflationRate < params.Tokenomics.MinInflationRate {
			return fmt.Errorf("max_inflation_rate must be >= min_inflation_rate")
		}
	}

	// Validate WhaleProtection
	if params.WhaleProtection != nil && params.WhaleProtection.Enabled {
		if params.WhaleProtection.MaxHoldingPercentage > 10000 {
			return fmt.Errorf("max_holding_percentage cannot exceed 100%% (10000 basis points)")
		}
		if params.WhaleProtection.MaxTxPercentage > 10000 {
			return fmt.Errorf("max_tx_percentage cannot exceed 100%% (10000 basis points)")
		}
	}

	// Validate TransferTax
	if params.TransferTax != nil && params.TransferTax.Enabled {
		if params.TransferTax.BaseTaxRate > 10000 {
			return fmt.Errorf("base_tax_rate cannot exceed 100%% (10000 basis points)")
		}
		total := params.TransferTax.BurnPercentage + params.TransferTax.TreasuryPercentage + params.TransferTax.RedistributePercentage
		if total != 10000 {
			return fmt.Errorf("burn + treasury + redistribute percentages must equal 100%% (10000 basis points)")
		}
	}

	// Validate LiquidityMining
	if params.LiquidityMining != nil && params.LiquidityMining.Enabled {
		if params.LiquidityMining.TotalRewardsAllocated == "" {
			return fmt.Errorf("total_rewards_allocated cannot be empty when liquidity mining is enabled")
		}
		if params.LiquidityMining.MaxRewardsPerEpoch == "" {
			return fmt.Errorf("max_rewards_per_epoch cannot be empty")
		}
	}

	// Validate Governance
	if params.Governance != nil {
		if params.Governance.MinProposalStake == "" {
			return fmt.Errorf("min_proposal_stake cannot be empty")
		}
		if params.Governance.ProposalDeposit == "" {
			return fmt.Errorf("proposal_deposit cannot be empty")
		}
	}

	// Validate TreasuryMultisig
	if params.TreasuryMultisig != nil {
		if params.TreasuryMultisig.Threshold == 0 {
			return fmt.Errorf("threshold must be greater than 0")
		}
		if params.TreasuryMultisig.Threshold > uint32(len(params.TreasuryMultisig.Signers)) && len(params.TreasuryMultisig.Signers) > 0 {
			return fmt.Errorf("threshold cannot be greater than number of signers")
		}
	}

	// Validate DynamicFees
	if params.DynamicFees != nil && params.DynamicFees.Enabled {
		if params.DynamicFees.BaseFee == "" {
			return fmt.Errorf("base_fee cannot be empty when dynamic fees is enabled")
		}
		if params.DynamicFees.MinMultiplier > params.DynamicFees.MaxMultiplier {
			return fmt.Errorf("min_multiplier must be <= max_multiplier")
		}
		if params.DynamicFees.TargetUtilization > 10000 {
			return fmt.Errorf("target_utilization cannot exceed 100%% (10000 basis points)")
		}
	}

	// Validate MEV
	if params.Mev != nil && params.Mev.Enabled {
		total := params.Mev.UserRedistributionPercentage + params.Mev.ValidatorPercentage + params.Mev.TreasuryPercentage + params.Mev.BurnPercentage
		if total != 10000 {
			return fmt.Errorf("user + validator + treasury + burn percentages must equal 100%% (10000 basis points)")
		}
	}

	// Validate inflation parameters
	if params.InflationAlertThreshold == 0 {
		return fmt.Errorf("inflation_alert_threshold must be greater than 0")
	}
	if params.InflationCheckInterval == 0 {
		return fmt.Errorf("inflation_check_interval must be greater than 0")
	}

	return nil
}
