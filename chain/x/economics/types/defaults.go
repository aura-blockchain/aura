package types

import (
	"time"

	"cosmossdk.io/math"

	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// DefaultParams returns default economics module parameters
func DefaultParams() *economicspb.Params {
	return &economicspb.Params{
		Fees:             *DefaultFeeParams(),
		Vesting:          *DefaultVestingParams(),
		Treasury:         *DefaultTreasuryParams(),
		Governance:       *DefaultGovernanceParams(),
		Mev:              *DefaultMEVParams(),
		WhaleProtection:  *DefaultWhaleProtectionParams(),
		LiquidityMining:  *DefaultLiquidityMiningParams(),
		Tokenomics:       *DefaultTokenomicsParams(),
	}
}

// DefaultFeeParams returns default fee parameters
func DefaultFeeParams() *economicspb.FeeParams {
	return &economicspb.FeeParams{
		BaseFee:                math.NewInt(1000),
		MinGasPrice:            math.LegacyMustNewDecFromStr("1"),
		DynamicFeesEnabled:     true,
		FeeBurnPercentage:      1000, // 10%
		MinFeeMultiplier:       5000, // 0.5x
		MaxFeeMultiplier:       50000, // 5x
		TargetBlockUtilization: 5000, // 50%
		FeeAdjustmentSpeed:     1000, // 10%
	}
}

// DefaultVestingParams returns default vesting parameters
func DefaultVestingParams() *economicspb.VestingParams {
	return &economicspb.VestingParams{
		MinVestingDuration:  30 * 24 * time.Hour, // 30 days
		MaxVestingDuration:  4 * 365 * 24 * time.Hour, // 4 years
		AllowEarlyUnlock:    false,
		EarlyUnlockPenalty:  2000, // 20%
		MinCliffDuration:    0,
	}
}

// DefaultTreasuryParams returns default treasury parameters
func DefaultTreasuryParams() *economicspb.TreasuryParams {
	return &economicspb.TreasuryParams{
		TreasuryAddress:         "",
		CommunityPoolPercentage: 2000, // 20%
		BurnPercentage:          0,
		MultisigThreshold:       2,
		AuthorizedSigners:       []string{},
		SpendingLimit:           math.NewInt(1000000),
		TimelockDuration:        24 * time.Hour,
	}
}

// DefaultGovernanceParams returns default governance parameters
func DefaultGovernanceParams() *economicspb.GovernanceParams {
	// Note: MinDeposit is set to nil to avoid protobuf serialization issues
	// with sdk.Coin types. This should be set properly in genesis or via governance.
	return &economicspb.GovernanceParams{
		MinDeposit:              nil, // TODO: Set in genesis - []*sdk.Coin not serializable with protoc-gen-go
		MaxDepositPeriod:        7 * 24 * time.Hour,
		VotingPeriod:            7 * 24 * time.Hour,
		Quorum:                  3333, // 33.33%
		Threshold:               5000, // 50%
		VetoThreshold:           3333, // 33.33%
		ExecutionDelay:          24 * time.Hour,
		EmergencyVotingPeriod:   24 * time.Hour,
		EmergencyQuorum:         6667, // 66.67%
		EmergencyThreshold:      6667, // 66.67%
		QuadraticVotingEnabled:  false,
		VoteLockingEnabled:      true,
		MinLockDuration:         7 * 24 * time.Hour,
		MaxLockDuration:         365 * 24 * time.Hour,
		LockMultiplierPerYear:   10000, // 1x per year
		SnapshotVotingEnabled:   false,
		SnapshotLookbackBlocks:  100,
		SecretBallotEnabled:     false,
		RevealPeriod:            24 * time.Hour,
	}
}

// DefaultMEVParams returns default MEV protection parameters
func DefaultMEVParams() *economicspb.MEVParams {
	return &economicspb.MEVParams{
		Enabled:                     true,
		MaxFrontrunPenalty:          math.NewInt(1000000),
		AuctionDuration:             10 * time.Second,
		UserRedistributionPercentage: 4000, // 40%
		ValidatorPercentage:         4000, // 40%
		TreasuryPercentage:          1000, // 10%
		BurnPercentage:              1000, // 10%
		Strategy:                    economicspb.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE,
	}
}

// DefaultWhaleProtectionParams returns default whale protection parameters
func DefaultWhaleProtectionParams() *economicspb.WhaleProtectionParams {
	return &economicspb.WhaleProtectionParams{
		Enabled:              true,
		MaxSingleTransfer:    math.NewInt(1000000000000),
		DailyTransferLimit:   math.NewInt(5000000000000),
		CooldownPeriod:       1 * time.Hour,
		MaxHoldingPercentage: 500, // 5%
		LargeTxThreshold:     100, // 1%
		ExemptedAddresses:    []string{},
	}
}

// DefaultLiquidityMiningParams returns default liquidity mining parameters
func DefaultLiquidityMiningParams() *economicspb.LiquidityMiningParams {
	return &economicspb.LiquidityMiningParams{
		Enabled:                 true,
		TotalRewardsAllocated:   math.NewInt(100000000000000),
		MaxRewardsPerEpoch:      math.NewInt(1000000000000),
		EpochDurationBlocks:     100000,
		IrVerifiedMultiplier:    15000, // 1.5x
	}
}

// DefaultTokenomicsParams returns default tokenomics parameters
func DefaultTokenomicsParams() *economicspb.TokenomicsParams {
	return &economicspb.TokenomicsParams{
		MaxSupply:                 math.NewInt(1000000000000000),
		TargetInflationRate:       500, // 5%
		MinInflationRate:          200, // 2%
		MaxInflationRate:          1000, // 10%
		InflationCheckInterval:    10000,
		InflationAlertThreshold:   100, // 1%
	}
}
