package types

import (
	"testing"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	if params == nil {
		t.Fatal("DefaultParams should not return nil")
	}

	// Validate tokenomics
	if params.Tokenomics == nil {
		t.Fatal("expected Tokenomics to be set")
	}

	if params.Tokenomics.MaxSupply != "1000000000" {
		t.Errorf("expected MaxSupply to be 1000000000, got %s", params.Tokenomics.MaxSupply)
	}

	if params.Tokenomics.CirculatingSupply != "100000000" {
		t.Errorf("expected CirculatingSupply to be 100000000, got %s", params.Tokenomics.CirculatingSupply)
	}

	// Validate whale protection
	if params.WhaleProtection == nil {
		t.Fatal("expected WhaleProtection to be set")
	}

	if !params.WhaleProtection.Enabled {
		t.Error("expected WhaleProtection to be enabled")
	}

	// Validate governance
	if params.Governance == nil {
		t.Fatal("expected Governance to be set")
	}

	if params.Governance.VoteLockingEnabled != true {
		t.Error("expected VoteLockingEnabled to be true")
	}

	// Validate MEV
	if params.Mev == nil {
		t.Fatal("expected Mev to be set")
	}

	if !params.Mev.Enabled {
		t.Error("expected Mev to be enabled")
	}
}

func TestValidateParams_Valid(t *testing.T) {
	params := DefaultParams()

	err := ValidateParams(params)
	if err != nil {
		t.Errorf("ValidateParams should not return error for default params: %v", err)
	}
}

func TestValidateParams_NilParams(t *testing.T) {
	err := ValidateParams(nil)
	if err == nil {
		t.Error("ValidateParams should return error for nil params")
	}
}

func TestValidateParams_EmptyMaxSupply(t *testing.T) {
	params := DefaultParams()
	params.Tokenomics.MaxSupply = ""

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for empty max_supply")
	}
}

func TestValidateParams_EmptyCirculatingSupply(t *testing.T) {
	params := DefaultParams()
	params.Tokenomics.CirculatingSupply = ""

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for empty circulating_supply")
	}
}

func TestValidateParams_InvalidInflationRange(t *testing.T) {
	params := DefaultParams()
	params.Tokenomics.MaxInflationRate = 100
	params.Tokenomics.MinInflationRate = 200

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when max < min inflation rate")
	}
}

func TestValidateParams_WhaleProtectionExceedsMax(t *testing.T) {
	params := DefaultParams()
	params.WhaleProtection.MaxHoldingPercentage = 15000 // > 10000 (100%)

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when max_holding_percentage exceeds 100%")
	}
}

func TestValidateParams_WhaleProtectionTxExceedsMax(t *testing.T) {
	params := DefaultParams()
	params.WhaleProtection.MaxTxPercentage = 15000 // > 10000 (100%)

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when max_tx_percentage exceeds 100%")
	}
}

func TestValidateParams_TransferTaxExceedsMax(t *testing.T) {
	params := DefaultParams()
	params.TransferTax.Enabled = true
	params.TransferTax.BaseTaxRate = 15000 // > 10000 (100%)

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when base_tax_rate exceeds 100%")
	}
}

func TestValidateParams_TransferTaxPercentageSum(t *testing.T) {
	params := DefaultParams()
	params.TransferTax.Enabled = true
	params.TransferTax.BurnPercentage = 3000
	params.TransferTax.TreasuryPercentage = 3000
	params.TransferTax.RedistributePercentage = 3000 // Sum = 9000 != 10000

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when percentages don't sum to 100%")
	}
}

func TestValidateParams_LiquidityMiningEmpty(t *testing.T) {
	params := DefaultParams()
	params.LiquidityMining.Enabled = true
	params.LiquidityMining.TotalRewardsAllocated = ""

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when total_rewards_allocated is empty")
	}
}

func TestValidateParams_EmptyProposalStake(t *testing.T) {
	params := DefaultParams()
	params.Governance.MinProposalStake = ""

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when min_proposal_stake is empty")
	}
}

func TestValidateParams_ZeroThreshold(t *testing.T) {
	params := DefaultParams()
	params.TreasuryMultisig.Threshold = 0

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when threshold is 0")
	}
}

func TestValidateParams_ThresholdExceedsSigners(t *testing.T) {
	params := DefaultParams()
	params.TreasuryMultisig.Threshold = 5
	params.TreasuryMultisig.Signers = []string{"signer1", "signer2"} // Only 2 signers

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when threshold > signers")
	}
}

func TestValidateParams_DynamicFeesEmpty(t *testing.T) {
	params := DefaultParams()
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = ""

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when base_fee is empty")
	}
}

func TestValidateParams_DynamicFeesInvalidRange(t *testing.T) {
	params := DefaultParams()
	params.DynamicFees.Enabled = true
	params.DynamicFees.MinMultiplier = 20000
	params.DynamicFees.MaxMultiplier = 10000

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when min_multiplier > max_multiplier")
	}
}

func TestValidateParams_DynamicFeesUtilizationExceeds(t *testing.T) {
	params := DefaultParams()
	params.DynamicFees.Enabled = true
	params.DynamicFees.TargetUtilization = 15000 // > 10000 (100%)

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when target_utilization exceeds 100%")
	}
}

func TestValidateParams_MEVPercentageSum(t *testing.T) {
	params := DefaultParams()
	params.Mev.Enabled = true
	params.Mev.UserRedistributionPercentage = 3000
	params.Mev.ValidatorPercentage = 3000
	params.Mev.TreasuryPercentage = 3000
	params.Mev.BurnPercentage = 0 // Sum = 9000 != 10000

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when MEV percentages don't sum to 100%")
	}
}

func TestValidateParams_ZeroInflationAlertThreshold(t *testing.T) {
	params := DefaultParams()
	params.InflationAlertThreshold = 0

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when inflation_alert_threshold is 0")
	}
}

func TestValidateParams_ZeroInflationCheckInterval(t *testing.T) {
	params := DefaultParams()
	params.InflationCheckInterval = 0

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error when inflation_check_interval is 0")
	}
}

func TestModuleName(t *testing.T) {
	if ModuleName != "economicsecurity" {
		t.Errorf("expected ModuleName to be 'economicsecurity', got %s", ModuleName)
	}
}

func TestBasisPointsConstant(t *testing.T) {
	if BasisPoints != 10000 {
		t.Errorf("expected BasisPoints to be 10000, got %d", BasisPoints)
	}
}

func TestEnumAliases(t *testing.T) {
	// Test InflationAlertType aliases
	if InflationAlertTypeUnspecified != InflationAlertType_INFLATION_ALERT_TYPE_UNSPECIFIED {
		t.Error("InflationAlertTypeUnspecified alias mismatch")
	}

	// Test AlertSeverity aliases
	if AlertSeverityInfo != AlertSeverity_ALERT_SEVERITY_INFO {
		t.Error("AlertSeverityInfo alias mismatch")
	}

	// Test MEVRedistributionStrategy aliases
	if MEVStrategyProportionalToStake != MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE {
		t.Error("MEVStrategyProportionalToStake alias mismatch")
	}

	// Test VestingType aliases
	if VestingTypeLinear != VestingType_VESTING_TYPE_LINEAR {
		t.Error("VestingTypeLinear alias mismatch")
	}
}
