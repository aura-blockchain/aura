// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Economics module error codes - consolidated from economicsecurity and governance
var (
	// ============================
	// General errors (1-9)
	// ============================
	ErrUnauthorized      = errorsmod.Register(ModuleName, 1, "unauthorized")
	ErrInvalidAmount     = errorsmod.Register(ModuleName, 2, "invalid amount")
	ErrInvalidAddress    = errorsmod.Register(ModuleName, 3, "invalid address")
	ErrInvalidDuration   = errorsmod.Register(ModuleName, 4, "invalid duration")
	ErrInvalidScheduleID = errorsmod.Register(ModuleName, 5, "invalid schedule ID")

	// ============================
	// Fee management errors (10-19)
	// ============================
	ErrInvalidFeeMultiplier     = errorsmod.Register(ModuleName, 10, "invalid fee multiplier")
	ErrInvalidTargetUtilization = errorsmod.Register(ModuleName, 11, "invalid target utilization")
	ErrInvalidAdjustmentSpeed   = errorsmod.Register(ModuleName, 12, "invalid adjustment speed")
	ErrInvalidTaxConfig         = errorsmod.Register(ModuleName, 13, "invalid transfer tax configuration")
	ErrTaxRateTooHigh           = errorsmod.Register(ModuleName, 14, "tax rate exceeds maximum")
	ErrInvalidTaxRecipient      = errorsmod.Register(ModuleName, 15, "invalid tax recipient address")

	// ============================
	// Vesting errors (20-49)
	// ============================
	ErrVestingScheduleNotFound  = errorsmod.Register(ModuleName, 20, "vesting schedule not found")
	ErrScheduleNotFound         = errorsmod.Register(ModuleName, 21, "schedule not found")
	ErrScheduleRevoked          = errorsmod.Register(ModuleName, 22, "schedule already revoked")
	ErrVestingAlreadyRevoked    = errorsmod.Register(ModuleName, 23, "vesting schedule already revoked")
	ErrNoVestedTokens           = errorsmod.Register(ModuleName, 24, "no tokens available to vest")
	ErrCliffNotReached          = errorsmod.Register(ModuleName, 25, "cliff period not yet reached")
	ErrInvalidBeneficiary       = errorsmod.Register(ModuleName, 26, "invalid beneficiary address")
	ErrInsufficientVestedAmount = errorsmod.Register(ModuleName, 27, "insufficient vested amount")
	ErrVoteLockNotFound         = errorsmod.Register(ModuleName, 28, "vote lock not found")
	ErrLockNotFound             = errorsmod.Register(ModuleName, 29, "lock not found")
	ErrLockWithdrawn            = errorsmod.Register(ModuleName, 30, "lock already withdrawn")
	ErrLockNotEnded             = errorsmod.Register(ModuleName, 31, "lock period not ended")
	ErrVoteLockNotExpired       = errorsmod.Register(ModuleName, 32, "vote lock has not expired yet")
	ErrInvalidLockDuration      = errorsmod.Register(ModuleName, 33, "invalid lock duration")
	ErrLockDurationTooShort     = errorsmod.Register(ModuleName, 34, "lock duration below minimum")
	ErrLockDurationTooLong      = errorsmod.Register(ModuleName, 35, "lock duration exceeds maximum")
	ErrVoteLockAlreadyWithdrawn = errorsmod.Register(ModuleName, 36, "vote lock already withdrawn")

	// ============================
	// Treasury errors (50-69)
	// ============================
	ErrInvalidTreasuryAddress      = errorsmod.Register(ModuleName, 50, "invalid treasury address")
	ErrInvalidThresholdValue       = errorsmod.Register(ModuleName, 51, "invalid threshold value")
	ErrInsufficientSignatures      = errorsmod.Register(ModuleName, 52, "insufficient signatures")
	ErrTxNotFound                  = errorsmod.Register(ModuleName, 53, "treasury transaction not found")
	ErrTreasuryTxNotFound          = errorsmod.Register(ModuleName, 54, "treasury transaction not found")
	ErrTreasuryTxExecuted          = errorsmod.Register(ModuleName, 55, "treasury transaction already executed")
	ErrTreasuryTxRejected          = errorsmod.Register(ModuleName, 56, "treasury transaction rejected")
	ErrTxAlreadyExecuted           = errorsmod.Register(ModuleName, 57, "transaction already executed")
	ErrTxAlreadyRejected           = errorsmod.Register(ModuleName, 58, "transaction already rejected")
	ErrTimelockNotExpired          = errorsmod.Register(ModuleName, 59, "timelock period not expired")
	ErrTimelockNotMet              = errorsmod.Register(ModuleName, 60, "timelock not met")
	ErrInvalidSigner               = errorsmod.Register(ModuleName, 61, "invalid signer")
	ErrAlreadySigned               = errorsmod.Register(ModuleName, 62, "already signed by this address")
	ErrInsufficientTreasuryBalance = errorsmod.Register(ModuleName, 63, "insufficient treasury balance")

	// ============================
	// Governance proposal errors (70-89)
	// ============================
	ErrInvalidRequest         = errorsmod.Register(ModuleName, 70, "invalid request")
	ErrInvalidProposal        = errorsmod.Register(ModuleName, 71, "invalid proposal")
	ErrProposalNotFound       = errorsmod.Register(ModuleName, 72, "proposal not found")
	ErrInvalidProposalStatus  = errorsmod.Register(ModuleName, 73, "invalid proposal status")
	ErrProposalNotPassed      = errorsmod.Register(ModuleName, 74, "proposal has not passed")
	ErrExecutionDelayNotMet   = errorsmod.Register(ModuleName, 75, "execution delay not met")
	ErrAlreadyExecuted        = errorsmod.Register(ModuleName, 76, "proposal already executed")
	ErrInsufficientStake      = errorsmod.Register(ModuleName, 77, "insufficient stake for governance proposal")
	ErrInvalidProposalDeposit = errorsmod.Register(ModuleName, 78, "invalid proposal deposit")

	// ============================
	// Governance voting errors (90-119)
	// ============================
	ErrInvalidVote             = errorsmod.Register(ModuleName, 90, "invalid vote")
	ErrVoteNotFound            = errorsmod.Register(ModuleName, 91, "vote not found")
	ErrAlreadyVoted            = errorsmod.Register(ModuleName, 92, "already voted")
	ErrVotingPeriodEnded       = errorsmod.Register(ModuleName, 93, "voting period has ended")
	ErrVotingPeriodNotStarted  = errorsmod.Register(ModuleName, 94, "voting period has not started")
	ErrInvalidQuorum           = errorsmod.Register(ModuleName, 95, "invalid quorum percentage")
	ErrInvalidThreshold        = errorsmod.Register(ModuleName, 96, "invalid threshold percentage")
	ErrInvalidWeight           = errorsmod.Register(ModuleName, 97, "invalid vote weight")
	ErrWeightedVoteNotEnabled  = errorsmod.Register(ModuleName, 98, "weighted voting not enabled")
	ErrInsufficientVotingPower = errorsmod.Register(ModuleName, 99, "insufficient voting power")

	// ============================
	// Governance deposit errors (120-129)
	// ============================
	ErrInsufficientDeposit = errorsmod.Register(ModuleName, 120, "insufficient deposit")
	ErrInvalidDeposit      = errorsmod.Register(ModuleName, 121, "invalid deposit")
	ErrDepositPeriodEnded  = errorsmod.Register(ModuleName, 122, "deposit period has ended")

	// ============================
	// Governance delegation errors (130-139)
	// ============================
	ErrInvalidDelegation  = errorsmod.Register(ModuleName, 130, "invalid vote delegation")
	ErrDelegationNotFound = errorsmod.Register(ModuleName, 131, "vote delegation not found")

	// ============================
	// Governance veto errors (140-149)
	// ============================
	ErrUnauthorizedVeto          = errorsmod.Register(ModuleName, 140, "unauthorized veto")
	ErrInsufficientVetoCosigners = errorsmod.Register(ModuleName, 141, "insufficient veto cosigners")
	ErrInvalidVeto               = errorsmod.Register(ModuleName, 142, "invalid veto")
	ErrExecutionDelayNotPassed   = errorsmod.Register(ModuleName, 143, "execution delay has not passed")

	// ============================
	// Secret ballot voting errors (150-169)
	// ============================
	ErrInvalidSnapshot        = errorsmod.Register(ModuleName, 150, "invalid snapshot vote")
	ErrInvalidReveal          = errorsmod.Register(ModuleName, 151, "invalid vote reveal")
	ErrRevealPeriodNotStarted = errorsmod.Register(ModuleName, 152, "reveal period has not started")
	ErrRevealPeriodEnded      = errorsmod.Register(ModuleName, 153, "reveal period has ended")
	ErrInvalidCommitment      = errorsmod.Register(ModuleName, 154, "invalid vote commitment")
	ErrSecretBallotDisabled   = errorsmod.Register(ModuleName, 155, "secret ballot voting is disabled")
	ErrVoteAlreadyRevealed    = errorsmod.Register(ModuleName, 156, "vote has already been revealed")
	ErrInvalidVoteReveal      = errorsmod.Register(ModuleName, 157, "invalid vote reveal")
	ErrNoVoteCommitment       = errorsmod.Register(ModuleName, 158, "no vote commitment found")

	// ============================
	// Quadratic voting errors (170-179)
	// ============================
	ErrQuadraticVotingDisabled = errorsmod.Register(ModuleName, 170, "quadratic voting is disabled")
	ErrInsufficientVoteCredits = errorsmod.Register(ModuleName, 171, "insufficient vote credits")

	// ============================
	// Token lock errors (180-189)
	// ============================
	ErrInsufficientTokens = errorsmod.Register(ModuleName, 180, "insufficient tokens for lock")
	ErrTokensLocked       = errorsmod.Register(ModuleName, 181, "tokens are locked")

	// ============================
	// Inflation errors (190-199)
	// ============================
	ErrInflationRateTooHigh = errorsmod.Register(ModuleName, 190, "inflation rate exceeds maximum")
	ErrInflationRateTooLow  = errorsmod.Register(ModuleName, 191, "inflation rate below minimum")
	ErrInvalidInflationRate = errorsmod.Register(ModuleName, 192, "invalid inflation rate")

	// ============================
	// Whale protection errors (200-209)
	// ============================
	ErrWhaleHoldingLimitExceeded = errorsmod.Register(ModuleName, 200, "whale holding limit exceeded")
	ErrWhaleTxLimitExceeded      = errorsmod.Register(ModuleName, 201, "whale transaction limit exceeded")
	ErrLargeTxCooldownActive     = errorsmod.Register(ModuleName, 202, "large transaction cooldown period active")
	ErrInvalidWhaleConfig        = errorsmod.Register(ModuleName, 203, "invalid whale protection configuration")

	// ============================
	// Liquidity mining errors (210-219)
	// ============================
	ErrLiquidityRewardCapExceeded = errorsmod.Register(ModuleName, 210, "liquidity mining reward cap exceeded")
	ErrInvalidEpoch               = errorsmod.Register(ModuleName, 211, "invalid epoch")
	ErrInsufficientRewards        = errorsmod.Register(ModuleName, 212, "insufficient rewards available")
	ErrLiquidityMiningDisabled    = errorsmod.Register(ModuleName, 213, "liquidity mining disabled")

	// ============================
	// MEV errors (220-239)
	// ============================
	ErrMEVRedistributionDisabled     = errorsmod.Register(ModuleName, 220, "MEV redistribution disabled")
	ErrInvalidMEVConfig              = errorsmod.Register(ModuleName, 221, "invalid MEV configuration")
	ErrInvalidRedistributionStrategy = errorsmod.Register(ModuleName, 222, "invalid redistribution strategy")
	ErrInsufficientMEVBalance        = errorsmod.Register(ModuleName, 223, "insufficient MEV balance")
	ErrMEVAuctionDisabled            = errorsmod.Register(ModuleName, 224, "MEV auction disabled")
	ErrAuctionNotFound               = errorsmod.Register(ModuleName, 225, "auction not found")
	ErrAuctionClosed                 = errorsmod.Register(ModuleName, 226, "auction is closed")
	ErrAuctionAlreadyClosed          = errorsmod.Register(ModuleName, 227, "auction already closed")
	ErrBidTooLow                     = errorsmod.Register(ModuleName, 228, "bid amount below minimum")
	ErrBidderAlreadyBid              = errorsmod.Register(ModuleName, 229, "bidder already placed a bid")

	// ============================
	// Supply cap errors (240-249)
	// ============================
	ErrMaxSupplyExceeded   = errorsmod.Register(ModuleName, 240, "maximum supply cap exceeded")
	ErrSupplyCapAlreadySet = errorsmod.Register(ModuleName, 241, "supply cap already set and is immutable")
	ErrInvalidSupplyCap    = errorsmod.Register(ModuleName, 242, "invalid supply cap")

	// ============================
	// Circuit breaker errors (250-259)
	// ============================
	ErrCircuitBreakerNotFound = errorsmod.Register(ModuleName, 250, "circuit breaker not found")

	// ============================
	// Gas prediction errors (260-269)
	// ============================
	ErrInvalidPriority = errorsmod.Register(ModuleName, 260, "invalid priority level")

	// ============================
	// Transaction batching errors (270-279)
	// ============================
	ErrBatchingDisabled = errorsmod.Register(ModuleName, 270, "transaction batching disabled")
	ErrBatchTooSmall    = errorsmod.Register(ModuleName, 271, "batch size below minimum threshold")
	ErrBatchNotFound    = errorsmod.Register(ModuleName, 272, "no pending batch found")
)
