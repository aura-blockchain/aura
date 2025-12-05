package types

import "errors"

// Economics module errors - consolidated from economicsecurity and governance
var (
	// ============================
	// General errors
	// ============================
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidAddress    = errors.New("invalid address")
	ErrInvalidDuration   = errors.New("invalid duration")
	ErrInvalidScheduleID = errors.New("invalid schedule ID")

	// ============================
	// Fee management errors
	// ============================
	ErrInvalidFeeMultiplier     = errors.New("invalid fee multiplier")
	ErrInvalidTargetUtilization = errors.New("invalid target utilization")
	ErrInvalidAdjustmentSpeed   = errors.New("invalid adjustment speed")
	ErrInvalidTaxConfig         = errors.New("invalid transfer tax configuration")
	ErrTaxRateTooHigh           = errors.New("tax rate exceeds maximum")
	ErrInvalidTaxRecipient      = errors.New("invalid tax recipient address")

	// ============================
	// Vesting errors
	// ============================
	ErrVestingScheduleNotFound  = errors.New("vesting schedule not found")
	ErrScheduleNotFound         = errors.New("schedule not found")
	ErrScheduleRevoked          = errors.New("schedule already revoked")
	ErrVestingAlreadyRevoked    = errors.New("vesting schedule already revoked")
	ErrNoVestedTokens           = errors.New("no tokens available to vest")
	ErrCliffNotReached          = errors.New("cliff period not yet reached")
	ErrInvalidBeneficiary       = errors.New("invalid beneficiary address")
	ErrInsufficientVestedAmount = errors.New("insufficient vested amount")
	ErrVoteLockNotFound         = errors.New("vote lock not found")
	ErrLockNotFound             = errors.New("lock not found")
	ErrLockWithdrawn            = errors.New("lock already withdrawn")
	ErrLockNotEnded             = errors.New("lock period not ended")
	ErrVoteLockNotExpired       = errors.New("vote lock has not expired yet")
	ErrInvalidLockDuration      = errors.New("invalid lock duration")
	ErrLockDurationTooShort     = errors.New("lock duration below minimum")
	ErrLockDurationTooLong      = errors.New("lock duration exceeds maximum")
	ErrVoteLockAlreadyWithdrawn = errors.New("vote lock already withdrawn")

	// ============================
	// Treasury errors
	// ============================
	ErrInvalidTreasuryAddress      = errors.New("invalid treasury address")
	ErrInvalidThresholdValue       = errors.New("invalid threshold value")
	ErrInsufficientSignatures      = errors.New("insufficient signatures")
	ErrTxNotFound                  = errors.New("treasury transaction not found")
	ErrTreasuryTxNotFound          = errors.New("treasury transaction not found")
	ErrTreasuryTxExecuted          = errors.New("treasury transaction already executed")
	ErrTreasuryTxRejected          = errors.New("treasury transaction rejected")
	ErrTxAlreadyExecuted           = errors.New("transaction already executed")
	ErrTxAlreadyRejected           = errors.New("transaction already rejected")
	ErrTimelockNotExpired          = errors.New("timelock period not expired")
	ErrTimelockNotMet              = errors.New("timelock not met")
	ErrInvalidSigner               = errors.New("invalid signer")
	ErrAlreadySigned               = errors.New("already signed by this address")
	ErrInsufficientTreasuryBalance = errors.New("insufficient treasury balance")

	// ============================
	// Governance proposal errors
	// ============================
	ErrInvalidRequest         = errors.New("invalid request")
	ErrInvalidProposal        = errors.New("invalid proposal")
	ErrProposalNotFound       = errors.New("proposal not found")
	ErrInvalidProposalStatus  = errors.New("invalid proposal status")
	ErrProposalNotPassed      = errors.New("proposal has not passed")
	ErrExecutionDelayNotMet   = errors.New("execution delay not met")
	ErrAlreadyExecuted        = errors.New("proposal already executed")
	ErrInsufficientStake      = errors.New("insufficient stake for governance proposal")
	ErrInvalidProposalDeposit = errors.New("invalid proposal deposit")

	// ============================
	// Governance voting errors
	// ============================
	ErrInvalidVote            = errors.New("invalid vote")
	ErrVoteNotFound           = errors.New("vote not found")
	ErrAlreadyVoted           = errors.New("already voted")
	ErrVotingPeriodEnded      = errors.New("voting period has ended")
	ErrVotingPeriodNotStarted = errors.New("voting period has not started")
	ErrInvalidQuorum          = errors.New("invalid quorum percentage")
	ErrInvalidThreshold       = errors.New("invalid threshold percentage")
	ErrInvalidWeight          = errors.New("invalid vote weight")
	ErrWeightedVoteNotEnabled = errors.New("weighted voting not enabled")

	// ============================
	// Governance deposit errors
	// ============================
	ErrInsufficientDeposit = errors.New("insufficient deposit")
	ErrInvalidDeposit      = errors.New("invalid deposit")
	ErrDepositPeriodEnded  = errors.New("deposit period has ended")

	// ============================
	// Governance delegation errors
	// ============================
	ErrInvalidDelegation  = errors.New("invalid vote delegation")
	ErrDelegationNotFound = errors.New("vote delegation not found")

	// ============================
	// Governance veto errors
	// ============================
	ErrUnauthorizedVeto          = errors.New("unauthorized veto")
	ErrInsufficientVetoCosigners = errors.New("insufficient veto cosigners")
	ErrInvalidVeto               = errors.New("invalid veto")
	ErrExecutionDelayNotPassed   = errors.New("execution delay has not passed")

	// ============================
	// Secret ballot voting errors
	// ============================
	ErrInvalidSnapshot        = errors.New("invalid snapshot vote")
	ErrInvalidReveal          = errors.New("invalid vote reveal")
	ErrRevealPeriodNotStarted = errors.New("reveal period has not started")
	ErrRevealPeriodEnded      = errors.New("reveal period has ended")
	ErrInvalidCommitment      = errors.New("invalid vote commitment")
	ErrSecretBallotDisabled   = errors.New("secret ballot voting is disabled")
	ErrVoteAlreadyRevealed    = errors.New("vote has already been revealed")
	ErrInvalidVoteReveal      = errors.New("invalid vote reveal")
	ErrNoVoteCommitment       = errors.New("no vote commitment found")

	// ============================
	// Quadratic voting errors
	// ============================
	ErrQuadraticVotingDisabled = errors.New("quadratic voting is disabled")
	ErrInsufficientVoteCredits = errors.New("insufficient vote credits")

	// ============================
	// Token lock errors
	// ============================
	ErrInsufficientTokens = errors.New("insufficient tokens for lock")
	ErrTokensLocked       = errors.New("tokens are locked")

	// ============================
	// Inflation errors
	// ============================
	ErrInflationRateTooHigh = errors.New("inflation rate exceeds maximum")
	ErrInflationRateTooLow  = errors.New("inflation rate below minimum")
	ErrInvalidInflationRate = errors.New("invalid inflation rate")

	// ============================
	// Whale protection errors
	// ============================
	ErrWhaleHoldingLimitExceeded = errors.New("whale holding limit exceeded")
	ErrWhaleTxLimitExceeded      = errors.New("whale transaction limit exceeded")
	ErrLargeTxCooldownActive     = errors.New("large transaction cooldown period active")
	ErrInvalidWhaleConfig        = errors.New("invalid whale protection configuration")

	// ============================
	// Liquidity mining errors
	// ============================
	ErrLiquidityRewardCapExceeded = errors.New("liquidity mining reward cap exceeded")
	ErrInvalidEpoch               = errors.New("invalid epoch")
	ErrInsufficientRewards        = errors.New("insufficient rewards available")
	ErrLiquidityMiningDisabled    = errors.New("liquidity mining disabled")

	// ============================
	// MEV errors
	// ============================
	ErrMEVRedistributionDisabled     = errors.New("MEV redistribution disabled")
	ErrInvalidMEVConfig              = errors.New("invalid MEV configuration")
	ErrInvalidRedistributionStrategy = errors.New("invalid redistribution strategy")
	ErrInsufficientMEVBalance        = errors.New("insufficient MEV balance")
	ErrMEVAuctionDisabled            = errors.New("MEV auction disabled")
	ErrAuctionNotFound               = errors.New("auction not found")
	ErrAuctionClosed                 = errors.New("auction is closed")
	ErrAuctionAlreadyClosed          = errors.New("auction already closed")
	ErrBidTooLow                     = errors.New("bid amount below minimum")
	ErrBidderAlreadyBid              = errors.New("bidder already placed a bid")

	// ============================
	// Supply cap errors
	// ============================
	ErrMaxSupplyExceeded   = errors.New("maximum supply cap exceeded")
	ErrSupplyCapAlreadySet = errors.New("supply cap already set and is immutable")
	ErrInvalidSupplyCap    = errors.New("invalid supply cap")

	// ============================
	// Circuit breaker errors
	// ============================
	ErrCircuitBreakerNotFound = errors.New("circuit breaker not found")

	// ============================
	// Gas prediction errors
	// ============================
	ErrInvalidPriority = errors.New("invalid priority level")

	// ============================
	// Transaction batching errors
	// ============================
	ErrBatchingDisabled = errors.New("transaction batching disabled")
	ErrBatchTooSmall    = errors.New("batch size below minimum threshold")
	ErrBatchNotFound    = errors.New("no pending batch found")
)
