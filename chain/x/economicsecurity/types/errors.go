package types

import "errors"

var (
	// General errors
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidAddress    = errors.New("invalid address")
	ErrInvalidDuration   = errors.New("invalid duration")
	ErrInvalidScheduleID = errors.New("invalid schedule ID")

	// Supply cap errors
	ErrMaxSupplyExceeded   = errors.New("maximum supply cap exceeded")
	ErrSupplyCapAlreadySet = errors.New("supply cap already set and is immutable")
	ErrInvalidSupplyCap    = errors.New("invalid supply cap")

	// Inflation errors
	ErrInflationRateTooHigh = errors.New("inflation rate exceeds maximum")
	ErrInflationRateTooLow  = errors.New("inflation rate below minimum")
	ErrInvalidInflationRate = errors.New("invalid inflation rate")

	// Vesting errors
	ErrVestingScheduleNotFound  = errors.New("vesting schedule not found")
	ErrVestingAlreadyRevoked    = errors.New("vesting schedule already revoked")
	ErrNoVestedTokens           = errors.New("no tokens available to vest")
	ErrCliffNotReached          = errors.New("cliff period not yet reached")
	ErrInvalidBeneficiary       = errors.New("invalid beneficiary address")
	ErrInsufficientVestedAmount = errors.New("insufficient vested amount")

	// Whale protection errors
	ErrWhaleHoldingLimitExceeded = errors.New("whale holding limit exceeded")
	ErrWhaleTxLimitExceeded      = errors.New("whale transaction limit exceeded")
	ErrLargeTxCooldownActive     = errors.New("large transaction cooldown period active")
	ErrInvalidWhaleConfig        = errors.New("invalid whale protection configuration")

	// Transfer tax errors
	ErrInvalidTaxConfig    = errors.New("invalid transfer tax configuration")
	ErrTaxRateTooHigh      = errors.New("tax rate exceeds maximum")
	ErrInvalidTaxRecipient = errors.New("invalid tax recipient address")

	// Liquidity mining errors
	ErrLiquidityRewardCapExceeded = errors.New("liquidity mining reward cap exceeded")
	ErrInvalidEpoch               = errors.New("invalid epoch")
	ErrInsufficientRewards        = errors.New("insufficient rewards available")
	ErrLiquidityMiningDisabled    = errors.New("liquidity mining disabled")

	// Governance errors
	ErrInsufficientStake      = errors.New("insufficient stake for governance proposal")
	ErrInvalidProposalDeposit = errors.New("invalid proposal deposit")
	ErrInvalidQuorum          = errors.New("invalid quorum percentage")
	ErrInvalidThreshold       = errors.New("invalid threshold percentage")

	// Vote locking errors
	ErrVoteLockNotFound         = errors.New("vote lock not found")
	ErrVoteLockNotExpired       = errors.New("vote lock has not expired yet")
	ErrInvalidLockDuration      = errors.New("invalid lock duration")
	ErrLockDurationTooShort     = errors.New("lock duration below minimum")
	ErrLockDurationTooLong      = errors.New("lock duration exceeds maximum")
	ErrVoteLockAlreadyWithdrawn = errors.New("vote lock already withdrawn")

	// Treasury errors
	ErrInvalidTreasuryAddress      = errors.New("invalid treasury address")
	ErrInvalidThresholdValue       = errors.New("invalid threshold value")
	ErrInsufficientSignatures      = errors.New("insufficient signatures")
	ErrTxNotFound                  = errors.New("treasury transaction not found")
	ErrTxAlreadyExecuted           = errors.New("transaction already executed")
	ErrTxAlreadyRejected           = errors.New("transaction already rejected")
	ErrTimelockNotExpired          = errors.New("timelock period not expired")
	ErrInvalidSigner               = errors.New("invalid signer")
	ErrAlreadySigned               = errors.New("already signed by this address")
	ErrInsufficientTreasuryBalance = errors.New("insufficient treasury balance")

	// Dynamic fees errors
	ErrInvalidFeeMultiplier     = errors.New("invalid fee multiplier")
	ErrInvalidTargetUtilization = errors.New("invalid target utilization")
	ErrInvalidAdjustmentSpeed   = errors.New("invalid adjustment speed")

	// MEV errors
	ErrMEVRedistributionDisabled     = errors.New("MEV redistribution disabled")
	ErrInvalidMEVConfig              = errors.New("invalid MEV configuration")
	ErrInvalidRedistributionStrategy = errors.New("invalid redistribution strategy")
	ErrInsufficientMEVBalance        = errors.New("insufficient MEV balance")
)
