package types

import (
	"errors"
	"math/big"

	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DefaultParams returns default module parameters
func DefaultParams() *Params {
	now := timestamppb.Now()

	return &Params{
		Tokenomics: &TokenomicsConfig{
			MaxSupply:               "1000000000000000", // 1 billion tokens (with 6 decimals)
			CirculatingSupply:       "0",
			InflationRate:           500,  // 5%
			TargetInflationRate:     500,  // 5%
			MinInflationRate:        100,  // 1%
			MaxInflationRate:        1000, // 10%
			LastInflationAdjustment: now,
			LastInflationCheck:      now,
		},
		WhaleProtection: &WhaleProtection{
			Enabled:              true,
			MaxHoldingPercentage: 500,  // 5% of total supply
			MaxTxPercentage:      100,  // 1% of total supply per tx
			LargeTxCooldown:      3600, // 1 hour
			LargeTxThreshold:     50,   // 0.5% of total supply
			ExemptedAddresses:    []string{},
		},
		TransferTax: &TransferTaxConfig{
			Enabled:                  false,
			BaseTaxRate:              0,
			TaxRecipient:             "",
			BurnPercentage:           5000, // 50% of tax
			TreasuryPercentage:       3000, // 30% of tax
			RedistributePercentage:   2000, // 20% of tax
			ExemptedAddresses:        []string{},
			DynamicAdjustmentEnabled: false,
			MaxTaxRate:               500, // 5% max
			MinTaxRate:               0,
		},
		LiquidityMining: &LiquidityMiningConfig{
			Enabled:                 true,
			TotalRewardsAllocated:   "100000000000", // 100M tokens
			TotalRewardsDistributed: "0",
			MaxRewardsPerEpoch:      "1000000000", // 1M tokens per epoch
			CurrentEpoch:            0,
			EpochDurationBlocks:     100000, // ~1 week at 6s blocks
			LastDistributionHeight:  0,
			IrVerifiedMultiplier:    12000, // 1.2x for IR-verified users
		},
		Governance: &GovernanceConfig{
			MinProposalStake:        "10000000000", // 10k tokens
			QuadraticVotingEnabled:  true,
			VoteLockingEnabled:      true,
			MinLockDuration:         2592000,      // 30 days
			MaxLockDuration:         126144000,    // 4 years
			LockMultiplierPerYear:   10000,        // 1x per year (basis points)
			ProposalDeposit:         "1000000000", // 1k tokens
			QuorumPercentage:        3000,         // 30%
			PassThresholdPercentage: 5000,         // 50%
		},
		TreasuryMultisig: &TreasuryMultisig{
			TreasuryAddress:  "",
			Threshold:        3,
			Signers:          []string{},
			SpendingLimit:    "1000000000", // 1k tokens can be spent without multisig
			TimelockDuration: 86400,        // 24 hours
		},
		DynamicFees: &DynamicFeeConfig{
			Enabled:           true,
			BaseFee:           "1000", // 0.001 tokens
			CurrentMultiplier: 10000,  // 1x
			MinMultiplier:     5000,   // 0.5x
			MaxMultiplier:     50000,  // 5x
			TargetUtilization: 7500,   // 75%
			AdjustmentSpeed:   125,    // 1.25% per block
			RecentUtilization: []uint64{},
			UtilizationWindow: 100, // 100 blocks
		},
		Mev: &MEVConfig{
			Enabled:                      true,
			UserRedistributionPercentage: 4000, // 40% to users
			ValidatorPercentage:          3000, // 30% to validators
			TreasuryPercentage:           2000, // 20% to treasury
			BurnPercentage:               1000, // 10% burned
			TotalMevCaptured:             "0",
			TotalMevRedistributed:        "0",
			Strategy:                     MEVStrategyIRWeighted,
		},
		InflationAlertThreshold: 200,   // 2% deviation triggers alert
		InflationCheckInterval:  43200, // Check every ~3 days (at 6s blocks)
	}
}

// ValidateParams validates module parameters
func ValidateParams(params *Params) error {
	if params.Tokenomics == nil {
		return errors.New("tokenomics config cannot be nil")
	}
	if err := validateTokenomics(params.Tokenomics); err != nil {
		return err
	}

	if params.WhaleProtection == nil {
		return errors.New("whale protection config cannot be nil")
	}
	if err := validateWhaleProtection(params.WhaleProtection); err != nil {
		return err
	}

	if params.TransferTax == nil {
		return errors.New("transfer tax config cannot be nil")
	}
	if err := validateTransferTax(params.TransferTax); err != nil {
		return err
	}

	if params.LiquidityMining == nil {
		return errors.New("liquidity mining config cannot be nil")
	}
	if err := validateLiquidityMining(params.LiquidityMining); err != nil {
		return err
	}

	if params.Governance == nil {
		return errors.New("governance config cannot be nil")
	}
	if err := validateGovernance(params.Governance); err != nil {
		return err
	}

	if params.TreasuryMultisig == nil {
		return errors.New("treasury multisig config cannot be nil")
	}
	if err := validateTreasuryMultisig(params.TreasuryMultisig); err != nil {
		return err
	}

	if params.DynamicFees == nil {
		return errors.New("dynamic fees config cannot be nil")
	}
	if err := validateDynamicFees(params.DynamicFees); err != nil {
		return err
	}

	if params.Mev == nil {
		return errors.New("MEV config cannot be nil")
	}
	if err := validateMEV(params.Mev); err != nil {
		return err
	}

	return nil
}

func validateTokenomics(config *TokenomicsConfig) error {
	// Validate max supply
	maxSupply := new(big.Int)
	if _, ok := maxSupply.SetString(config.MaxSupply, 10); !ok {
		return ErrInvalidSupplyCap
	}
	if maxSupply.Cmp(big.NewInt(0)) <= 0 {
		return ErrInvalidSupplyCap
	}

	// Validate circulating supply
	circSupply := new(big.Int)
	if _, ok := circSupply.SetString(config.CirculatingSupply, 10); !ok {
		return errors.New("invalid circulating supply")
	}
	if circSupply.Cmp(maxSupply) > 0 {
		return ErrMaxSupplyExceeded
	}

	// Validate inflation rates
	if config.InflationRate > config.MaxInflationRate {
		return ErrInflationRateTooHigh
	}
	if config.InflationRate < config.MinInflationRate {
		return ErrInflationRateTooLow
	}
	if config.MinInflationRate > config.MaxInflationRate {
		return errors.New("min inflation rate cannot exceed max inflation rate")
	}

	return nil
}

func validateWhaleProtection(config *WhaleProtection) error {
	if config.MaxHoldingPercentage > BasisPoints {
		return ErrInvalidWhaleConfig
	}
	if config.MaxTxPercentage > BasisPoints {
		return ErrInvalidWhaleConfig
	}
	if config.LargeTxThreshold > BasisPoints {
		return ErrInvalidWhaleConfig
	}
	return nil
}

func validateTransferTax(config *TransferTaxConfig) error {
	if config.BaseTaxRate > config.MaxTaxRate {
		return ErrTaxRateTooHigh
	}
	if config.MaxTaxRate > BasisPoints {
		return ErrTaxRateTooHigh
	}

	// Validate tax distribution adds up to 100%
	total := config.BurnPercentage + config.TreasuryPercentage + config.RedistributePercentage
	if total != BasisPoints {
		return errors.New("tax distribution percentages must sum to 100%")
	}

	return nil
}

func validateLiquidityMining(config *LiquidityMiningConfig) error {
	totalAllocated := new(big.Int)
	if _, ok := totalAllocated.SetString(config.TotalRewardsAllocated, 10); !ok {
		return errors.New("invalid total rewards allocated")
	}

	totalDistributed := new(big.Int)
	if _, ok := totalDistributed.SetString(config.TotalRewardsDistributed, 10); !ok {
		return errors.New("invalid total rewards distributed")
	}

	if totalDistributed.Cmp(totalAllocated) > 0 {
		return errors.New("distributed rewards cannot exceed allocated rewards")
	}

	maxPerEpoch := new(big.Int)
	if _, ok := maxPerEpoch.SetString(config.MaxRewardsPerEpoch, 10); !ok {
		return errors.New("invalid max rewards per epoch")
	}

	if config.EpochDurationBlocks == 0 {
		return ErrInvalidEpoch
	}

	return nil
}

func validateGovernance(config *GovernanceConfig) error {
	minStake := new(big.Int)
	if _, ok := minStake.SetString(config.MinProposalStake, 10); !ok {
		return errors.New("invalid min proposal stake")
	}

	deposit := new(big.Int)
	if _, ok := deposit.SetString(config.ProposalDeposit, 10); !ok {
		return errors.New("invalid proposal deposit")
	}

	if config.MinLockDuration > config.MaxLockDuration {
		return ErrInvalidLockDuration
	}

	if config.QuorumPercentage > BasisPoints {
		return ErrInvalidQuorum
	}

	if config.PassThresholdPercentage > BasisPoints {
		return ErrInvalidThreshold
	}

	return nil
}

func validateTreasuryMultisig(config *TreasuryMultisig) error {
	if config.Threshold == 0 {
		return ErrInvalidThresholdValue
	}

	if uint32(len(config.Signers)) < config.Threshold {
		return errors.New("threshold cannot exceed number of signers")
	}

	spendingLimit := new(big.Int)
	if _, ok := spendingLimit.SetString(config.SpendingLimit, 10); !ok {
		return errors.New("invalid spending limit")
	}

	return nil
}

func validateDynamicFees(config *DynamicFeeConfig) error {
	baseFee := new(big.Int)
	if _, ok := baseFee.SetString(config.BaseFee, 10); !ok {
		return errors.New("invalid base fee")
	}

	if config.MinMultiplier > config.MaxMultiplier {
		return ErrInvalidFeeMultiplier
	}

	if config.TargetUtilization > BasisPoints {
		return ErrInvalidTargetUtilization
	}

	if config.UtilizationWindow == 0 {
		return errors.New("utilization window cannot be zero")
	}

	return nil
}

func validateMEV(config *MEVConfig) error {
	total := config.UserRedistributionPercentage +
		config.ValidatorPercentage +
		config.TreasuryPercentage +
		config.BurnPercentage

	if total != BasisPoints {
		return errors.New("MEV distribution percentages must sum to 100%")
	}

	totalCaptured := new(big.Int)
	if _, ok := totalCaptured.SetString(config.TotalMevCaptured, 10); !ok {
		return errors.New("invalid total MEV captured")
	}

	totalRedistributed := new(big.Int)
	if _, ok := totalRedistributed.SetString(config.TotalMevRedistributed, 10); !ok {
		return errors.New("invalid total MEV redistributed")
	}

	if totalRedistributed.Cmp(totalCaptured) > 0 {
		return errors.New("redistributed MEV cannot exceed captured MEV")
	}

	return nil
}

// ParamsFromProto converts proto Params to internal type
func ParamsFromProto(pb *economicsecuritypb.Params) *Params {
	if pb == nil {
		return DefaultParams()
	}
	return pb
}

// ParamsToProto converts internal Params to proto type
func ParamsToProto(params Params) *economicsecuritypb.Params {
	return &params
}
