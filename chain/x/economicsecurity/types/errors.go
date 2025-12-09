package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Economic security module error codes
var (
	// General errors (1-9)
	ErrUnauthorized      = errorsmod.Register(ModuleName, 1, "unauthorized")
	ErrInvalidAmount     = errorsmod.Register(ModuleName, 2, "invalid amount")
	ErrInvalidAddress    = errorsmod.Register(ModuleName, 3, "invalid address")
	ErrInvalidDuration   = errorsmod.Register(ModuleName, 4, "invalid duration")
	ErrInvalidScheduleID = errorsmod.Register(ModuleName, 5, "invalid schedule ID")

	// Supply cap errors (10-19)
	ErrMaxSupplyExceeded   = errorsmod.Register(ModuleName, 10, "maximum supply cap exceeded")
	ErrSupplyCapAlreadySet = errorsmod.Register(ModuleName, 11, "supply cap already set and is immutable")
	ErrInvalidSupplyCap    = errorsmod.Register(ModuleName, 12, "invalid supply cap")

	// Inflation errors (20-29)
	ErrInflationRateTooHigh = errorsmod.Register(ModuleName, 20, "inflation rate exceeds maximum")
	ErrInflationRateTooLow  = errorsmod.Register(ModuleName, 21, "inflation rate below minimum")
	ErrInvalidInflationRate = errorsmod.Register(ModuleName, 22, "invalid inflation rate")

	// Vesting errors (30-49)
	ErrVestingScheduleNotFound  = errorsmod.Register(ModuleName, 30, "vesting schedule not found")
	ErrVestingAlreadyRevoked    = errorsmod.Register(ModuleName, 31, "vesting schedule already revoked")
	ErrNoVestedTokens           = errorsmod.Register(ModuleName, 32, "no tokens available to vest")
	ErrCliffNotReached          = errorsmod.Register(ModuleName, 33, "cliff period not yet reached")
	ErrInvalidBeneficiary       = errorsmod.Register(ModuleName, 34, "invalid beneficiary address")
	ErrInsufficientVestedAmount = errorsmod.Register(ModuleName, 35, "insufficient vested amount")

	// Whale protection errors (50-59)
	ErrWhaleHoldingLimitExceeded = errorsmod.Register(ModuleName, 50, "whale holding limit exceeded")
	ErrWhaleTxLimitExceeded      = errorsmod.Register(ModuleName, 51, "whale transaction limit exceeded")
	ErrLargeTxCooldownActive     = errorsmod.Register(ModuleName, 52, "large transaction cooldown period active")
	ErrInvalidWhaleConfig        = errorsmod.Register(ModuleName, 53, "invalid whale protection configuration")

	// Transfer tax errors (60-69)
	ErrInvalidTaxConfig    = errorsmod.Register(ModuleName, 60, "invalid transfer tax configuration")
	ErrTaxRateTooHigh      = errorsmod.Register(ModuleName, 61, "tax rate exceeds maximum")
	ErrInvalidTaxRecipient = errorsmod.Register(ModuleName, 62, "invalid tax recipient address")

	// Liquidity mining errors (70-79)
	ErrLiquidityRewardCapExceeded = errorsmod.Register(ModuleName, 70, "liquidity mining reward cap exceeded")
	ErrInvalidEpoch               = errorsmod.Register(ModuleName, 71, "invalid epoch")
	ErrInsufficientRewards        = errorsmod.Register(ModuleName, 72, "insufficient rewards available")
	ErrLiquidityMiningDisabled    = errorsmod.Register(ModuleName, 73, "liquidity mining disabled")

	// Governance errors (80-89)
	ErrInsufficientStake      = errorsmod.Register(ModuleName, 80, "insufficient stake for governance proposal")
	ErrInvalidProposalDeposit = errorsmod.Register(ModuleName, 81, "invalid proposal deposit")
	ErrInvalidQuorum          = errorsmod.Register(ModuleName, 82, "invalid quorum percentage")
	ErrInvalidThreshold       = errorsmod.Register(ModuleName, 83, "invalid threshold percentage")

	// Vote locking errors (90-99)
	ErrVoteLockNotFound         = errorsmod.Register(ModuleName, 90, "vote lock not found")
	ErrVoteLockNotExpired       = errorsmod.Register(ModuleName, 91, "vote lock has not expired yet")
	ErrInvalidLockDuration      = errorsmod.Register(ModuleName, 92, "invalid lock duration")
	ErrLockDurationTooShort     = errorsmod.Register(ModuleName, 93, "lock duration below minimum")
	ErrLockDurationTooLong      = errorsmod.Register(ModuleName, 94, "lock duration exceeds maximum")
	ErrVoteLockAlreadyWithdrawn = errorsmod.Register(ModuleName, 95, "vote lock already withdrawn")

	// Treasury errors (100-119)
	ErrInvalidTreasuryAddress      = errorsmod.Register(ModuleName, 100, "invalid treasury address")
	ErrInvalidThresholdValue       = errorsmod.Register(ModuleName, 101, "invalid threshold value")
	ErrInsufficientSignatures      = errorsmod.Register(ModuleName, 102, "insufficient signatures")
	ErrTxNotFound                  = errorsmod.Register(ModuleName, 103, "treasury transaction not found")
	ErrTxAlreadyExecuted           = errorsmod.Register(ModuleName, 104, "transaction already executed")
	ErrTxAlreadyRejected           = errorsmod.Register(ModuleName, 105, "transaction already rejected")
	ErrTimelockNotExpired          = errorsmod.Register(ModuleName, 106, "timelock period not expired")
	ErrInvalidSigner               = errorsmod.Register(ModuleName, 107, "invalid signer")
	ErrAlreadySigned               = errorsmod.Register(ModuleName, 108, "already signed by this address")
	ErrInsufficientTreasuryBalance = errorsmod.Register(ModuleName, 109, "insufficient treasury balance")

	// Dynamic fees errors (120-129)
	ErrInvalidFeeMultiplier     = errorsmod.Register(ModuleName, 120, "invalid fee multiplier")
	ErrInvalidTargetUtilization = errorsmod.Register(ModuleName, 121, "invalid target utilization")
	ErrInvalidAdjustmentSpeed   = errorsmod.Register(ModuleName, 122, "invalid adjustment speed")

	// MEV errors (130-149)
	ErrMEVRedistributionDisabled     = errorsmod.Register(ModuleName, 130, "MEV redistribution disabled")
	ErrInvalidMEVConfig              = errorsmod.Register(ModuleName, 131, "invalid MEV configuration")
	ErrInvalidRedistributionStrategy = errorsmod.Register(ModuleName, 132, "invalid redistribution strategy")
	ErrInsufficientMEVBalance        = errorsmod.Register(ModuleName, 133, "insufficient MEV balance")

	// MEV Auction errors (150-159)
	ErrMEVAuctionDisabled   = errorsmod.Register(ModuleName, 150, "MEV auction disabled")
	ErrAuctionNotFound      = errorsmod.Register(ModuleName, 151, "auction not found")
	ErrAuctionClosed        = errorsmod.Register(ModuleName, 152, "auction is closed")
	ErrAuctionAlreadyClosed = errorsmod.Register(ModuleName, 153, "auction already closed")
	ErrBidTooLow            = errorsmod.Register(ModuleName, 154, "bid amount below minimum")
	ErrBidderAlreadyBid     = errorsmod.Register(ModuleName, 155, "bidder already placed a bid")

	// Circuit breaker errors (160-169)
	ErrCircuitBreakerNotFound = errorsmod.Register(ModuleName, 160, "circuit breaker not found")

	// Gas prediction errors (170-179)
	ErrInvalidPriority = errorsmod.Register(ModuleName, 170, "invalid priority level")

	// Transaction batching errors (180-189)
	ErrBatchingDisabled = errorsmod.Register(ModuleName, 180, "transaction batching disabled")
	ErrBatchTooSmall    = errorsmod.Register(ModuleName, 181, "batch size below minimum threshold")
	ErrBatchNotFound    = errorsmod.Register(ModuleName, 182, "no pending batch found")
)
