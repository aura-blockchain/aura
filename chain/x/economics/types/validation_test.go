// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	economicstypes "github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// TestValidateParams tests comprehensive validation of economics module parameters
func TestValidateParams(t *testing.T) {
	tests := []struct {
		name      string
		params    *economicspb.Params
		expectErr bool
		errMsg    string
	}{
		{
			name:      "nil params",
			params:    nil,
			expectErr: true,
			errMsg:    "params cannot be nil",
		},
		{
			name:      "valid default params",
			params:    economicstypes.DefaultParams(),
			expectErr: false,
		},
		{
			name: "invalid fee params - min > max multiplier",
			params: &economicspb.Params{
				Fees: economicspb.FeeParams{
					MinFeeMultiplier: 200,
					MaxFeeMultiplier: 100,
				},
				Vesting:         economicspb.VestingParams{},
				Treasury:        economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "min fee multiplier cannot be greater than max fee multiplier",
		},
		{
			name: "invalid fee params - target utilization > 100%",
			params: &economicspb.Params{
				Fees: economicspb.FeeParams{
					MinFeeMultiplier:        100,
					MaxFeeMultiplier:        200,
					TargetBlockUtilization:  10001,
				},
				Vesting:         economicspb.VestingParams{},
				Treasury:        economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "target block utilization cannot exceed 100%",
		},
		{
			name: "invalid vesting params - max < min duration",
			params: &economicspb.Params{
				Fees: economicspb.FeeParams{},
				Vesting: economicspb.VestingParams{
					MinVestingDuration: 365 * 24 * 3600, // 1 year
					MaxVestingDuration: 30 * 24 * 3600,  // 30 days
				},
				Treasury:        economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "max vesting duration must be greater than or equal to min vesting duration",
		},
		{
			name: "invalid vesting params - early unlock penalty > 100%",
			params: &economicspb.Params{
				Fees: economicspb.FeeParams{},
				Vesting: economicspb.VestingParams{
					MinVestingDuration: 30 * 24 * 3600,
					MaxVestingDuration: 365 * 24 * 3600,
					EarlyUnlockPenalty: 10001,
				},
				Treasury:        economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "early unlock penalty cannot exceed 100%",
		},
		{
			name: "invalid treasury params - community pool > 100%",
			params: &economicspb.Params{
				Fees:    economicspb.FeeParams{},
				Vesting: economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{
					CommunityPoolPercentage: 10001,
					MultisigThreshold:       1,
				},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "community pool percentage cannot exceed 100%",
		},
		{
			name: "invalid treasury params - burn percentage > 100%",
			params: &economicspb.Params{
				Fees:    economicspb.FeeParams{},
				Vesting: economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{
					BurnPercentage:    10001,
					MultisigThreshold: 1,
				},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "burn percentage cannot exceed 100%",
		},
		{
			name: "invalid treasury params - zero multisig threshold",
			params: &economicspb.Params{
				Fees:    economicspb.FeeParams{},
				Vesting: economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{
					MultisigThreshold: 0,
				},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "multisig threshold must be at least 1",
		},
		{
			name: "invalid governance params - quorum > 100%",
			params: &economicspb.Params{
				Fees:     economicspb.FeeParams{},
				Vesting:  economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance: economicspb.GovernanceParams{
					Quorum: 10001,
				},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "quorum cannot exceed 100%",
		},
		{
			name: "invalid governance params - threshold > 100%",
			params: &economicspb.Params{
				Fees:     economicspb.FeeParams{},
				Vesting:  economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance: economicspb.GovernanceParams{
					Threshold: 10001,
				},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "threshold cannot exceed 100%",
		},
		{
			name: "invalid governance params - veto threshold > 100%",
			params: &economicspb.Params{
				Fees:     economicspb.FeeParams{},
				Vesting:  economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance: economicspb.GovernanceParams{
					VetoThreshold: 10001,
				},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "veto threshold cannot exceed 100%",
		},
		{
			name: "invalid MEV params - percentages don't sum to 100%",
			params: &economicspb.Params{
				Fees:     economicspb.FeeParams{},
				Vesting:  economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance: economicspb.GovernanceParams{},
				Mev: economicspb.MEVParams{
					UserRedistributionPercentage: 5000,
					ValidatorPercentage:          3000,
					TreasuryPercentage:           1000,
					BurnPercentage:               500, // Total = 9500, not 10000
				},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "MEV redistribution percentages must sum to 100%",
		},
		{
			name: "invalid whale protection - max holding > 100%",
			params: &economicspb.Params{
				Fees:       economicspb.FeeParams{},
				Vesting:    economicspb.VestingParams{},
				Treasury:   economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance: economicspb.GovernanceParams{},
				Mev:        economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{
					MaxHoldingPercentage: 10001,
				},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "max holding percentage cannot exceed 100%",
		},
		{
			name: "invalid whale protection - large tx threshold > 100%",
			params: &economicspb.Params{
				Fees:       economicspb.FeeParams{},
				Vesting:    economicspb.VestingParams{},
				Treasury:   economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance: economicspb.GovernanceParams{},
				Mev:        economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{
					LargeTxThreshold: 10001,
				},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "large tx threshold cannot exceed 100%",
		},
		{
			name: "invalid liquidity mining - zero epoch duration",
			params: &economicspb.Params{
				Fees:            economicspb.FeeParams{},
				Vesting:         economicspb.VestingParams{},
				Treasury:        economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{
					EpochDurationBlocks: 0,
				},
				Tokenomics: economicspb.TokenomicsParams{InflationCheckInterval: 100},
			},
			expectErr: true,
			errMsg:    "epoch duration blocks must be positive",
		},
		{
			name: "invalid tokenomics - min inflation > target inflation",
			params: &economicspb.Params{
				Fees:            economicspb.FeeParams{},
				Vesting:         economicspb.VestingParams{},
				Treasury:        economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics: economicspb.TokenomicsParams{
					MinInflationRate:        1000,
					TargetInflationRate:     500,
					MaxInflationRate:        1500,
					InflationCheckInterval:  100,
				},
			},
			expectErr: true,
			errMsg:    "min inflation rate cannot exceed target inflation rate",
		},
		{
			name: "invalid tokenomics - target inflation > max inflation",
			params: &economicspb.Params{
				Fees:            economicspb.FeeParams{},
				Vesting:         economicspb.VestingParams{},
				Treasury:        economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics: economicspb.TokenomicsParams{
					MinInflationRate:        500,
					TargetInflationRate:     1500,
					MaxInflationRate:        1000,
					InflationCheckInterval:  100,
				},
			},
			expectErr: true,
			errMsg:    "target inflation rate cannot exceed max inflation rate",
		},
		{
			name: "invalid tokenomics - zero inflation check interval",
			params: &economicspb.Params{
				Fees:            economicspb.FeeParams{},
				Vesting:         economicspb.VestingParams{},
				Treasury:        economicspb.TreasuryParams{MultisigThreshold: 1},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
				Tokenomics: economicspb.TokenomicsParams{
					MinInflationRate:        500,
					TargetInflationRate:     700,
					MaxInflationRate:        1000,
					InflationCheckInterval:  0,
				},
			},
			expectErr: true,
			errMsg:    "inflation check interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := economicstypes.ValidateParams(tt.params)

			if tt.expectErr {
				require.Error(t, err, "expected error for %s", tt.name)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg, "error message mismatch")
				}
			} else {
				require.NoError(t, err, "unexpected error for %s", tt.name)
			}
		})
	}
}

// TestValidateFeeParams tests fee parameter validation edge cases
func TestValidateFeeParams(t *testing.T) {
	tests := []struct {
		name      string
		params    *economicspb.FeeParams
		expectErr bool
	}{
		{
			name:      "nil params",
			params:    nil,
			expectErr: false, // nil is allowed for optional params
		},
		{
			name: "valid params",
			params: &economicspb.FeeParams{
				MinFeeMultiplier:       100,
				MaxFeeMultiplier:       500,
				TargetBlockUtilization: 7000,
			},
			expectErr: false,
		},
		{
			name: "equal min and max multipliers",
			params: &economicspb.FeeParams{
				MinFeeMultiplier: 100,
				MaxFeeMultiplier: 100,
			},
			expectErr: false, // Equal is allowed
		},
		{
			name: "max utilization exactly 100%",
			params: &economicspb.FeeParams{
				MinFeeMultiplier:       100,
				MaxFeeMultiplier:       200,
				TargetBlockUtilization: 10000,
			},
			expectErr: false, // Exactly 100% is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := economicstypes.ValidateFeeParams(tt.params)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateMEVParams tests MEV redistribution parameter validation
func TestValidateMEVParams(t *testing.T) {
	tests := []struct {
		name      string
		params    *economicspb.MEVParams
		expectErr bool
		errMsg    string
	}{
		{
			name:      "nil params",
			params:    nil,
			expectErr: false,
		},
		{
			name: "valid redistribution - 50/30/10/10",
			params: &economicspb.MEVParams{
				UserRedistributionPercentage: 5000,
				ValidatorPercentage:          3000,
				TreasuryPercentage:           1000,
				BurnPercentage:               1000,
			},
			expectErr: false,
		},
		{
			name: "valid redistribution - all to users",
			params: &economicspb.MEVParams{
				UserRedistributionPercentage: 10000,
				ValidatorPercentage:          0,
				TreasuryPercentage:           0,
				BurnPercentage:               0,
			},
			expectErr: false,
		},
		{
			name: "valid redistribution - all to validators",
			params: &economicspb.MEVParams{
				UserRedistributionPercentage: 0,
				ValidatorPercentage:          10000,
				TreasuryPercentage:           0,
				BurnPercentage:               0,
			},
			expectErr: false,
		},
		{
			name: "invalid redistribution - sum less than 100%",
			params: &economicspb.MEVParams{
				UserRedistributionPercentage: 5000,
				ValidatorPercentage:          3000,
				TreasuryPercentage:           1000,
				BurnPercentage:               500,
			},
			expectErr: true,
			errMsg:    "must sum to 100%",
		},
		{
			name: "invalid redistribution - sum greater than 100%",
			params: &economicspb.MEVParams{
				UserRedistributionPercentage: 5000,
				ValidatorPercentage:          4000,
				TreasuryPercentage:           2000,
				BurnPercentage:               2000,
			},
			expectErr: true,
			errMsg:    "must sum to 100%",
		},
		{
			name: "invalid redistribution - all zeros",
			params: &economicspb.MEVParams{
				UserRedistributionPercentage: 0,
				ValidatorPercentage:          0,
				TreasuryPercentage:           0,
				BurnPercentage:               0,
			},
			expectErr: true,
			errMsg:    "must sum to 100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := economicstypes.ValidateMEVParams(tt.params)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateTokenomicsParams tests tokenomics parameter validation edge cases
func TestValidateTokenomicsParams(t *testing.T) {
	tests := []struct {
		name      string
		params    *economicspb.TokenomicsParams
		expectErr bool
		errMsg    string
	}{
		{
			name:      "nil params",
			params:    nil,
			expectErr: false,
		},
		{
			name: "valid params",
			params: &economicspb.TokenomicsParams{
				MinInflationRate:       500,
				TargetInflationRate:    700,
				MaxInflationRate:       1000,
				InflationCheckInterval: 100,
			},
			expectErr: false,
		},
		{
			name: "equal min and target",
			params: &economicspb.TokenomicsParams{
				MinInflationRate:       500,
				TargetInflationRate:    500,
				MaxInflationRate:       1000,
				InflationCheckInterval: 100,
			},
			expectErr: false,
		},
		{
			name: "equal target and max",
			params: &economicspb.TokenomicsParams{
				MinInflationRate:       500,
				TargetInflationRate:    1000,
				MaxInflationRate:       1000,
				InflationCheckInterval: 100,
			},
			expectErr: false,
		},
		{
			name: "all equal",
			params: &economicspb.TokenomicsParams{
				MinInflationRate:       500,
				TargetInflationRate:    500,
				MaxInflationRate:       500,
				InflationCheckInterval: 100,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := economicstypes.ValidateTokenomicsParams(tt.params)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
