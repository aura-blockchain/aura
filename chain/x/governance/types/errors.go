package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Governance module error codes
var (
	// Proposal errors (1-19)
	ErrInvalidProposal       = errorsmod.Register(ModuleName, 1, "invalid proposal")
	ErrProposalNotFound      = errorsmod.Register(ModuleName, 2, "proposal not found")
	ErrInvalidProposalStatus = errorsmod.Register(ModuleName, 3, "invalid proposal status")
	ErrProposalNotPassed     = errorsmod.Register(ModuleName, 4, "proposal has not passed")
	ErrAlreadyExecuted       = errorsmod.Register(ModuleName, 5, "proposal already executed")

	// Deposit errors (20-29)
	ErrInsufficientDeposit = errorsmod.Register(ModuleName, 20, "insufficient deposit")
	ErrInvalidDeposit      = errorsmod.Register(ModuleName, 21, "invalid deposit")
	ErrDepositPeriodEnded  = errorsmod.Register(ModuleName, 22, "deposit period has ended")

	// Voting errors (30-49)
	ErrInvalidVote            = errorsmod.Register(ModuleName, 30, "invalid vote")
	ErrVotingPeriodEnded      = errorsmod.Register(ModuleName, 31, "voting period has ended")
	ErrVotingPeriodNotStarted = errorsmod.Register(ModuleName, 32, "voting period has not started")
	ErrAlreadyVoted           = errorsmod.Register(ModuleName, 33, "already voted")
	ErrWeightedVoteNotEnabled = errorsmod.Register(ModuleName, 34, "weighted voting not enabled")
	ErrInvalidWeight          = errorsmod.Register(ModuleName, 35, "invalid vote weight")

	// Veto errors (50-59)
	ErrUnauthorizedVeto          = errorsmod.Register(ModuleName, 50, "unauthorized veto")
	ErrInsufficientVetoCosigners = errorsmod.Register(ModuleName, 51, "insufficient veto cosigners")
	ErrInvalidVeto               = errorsmod.Register(ModuleName, 52, "invalid veto")
	ErrExecutionDelayNotPassed   = errorsmod.Register(ModuleName, 53, "execution delay has not passed")

	// Delegation errors (60-69)
	ErrInvalidDelegation  = errorsmod.Register(ModuleName, 60, "invalid vote delegation")
	ErrDelegationNotFound = errorsmod.Register(ModuleName, 61, "vote delegation not found")

	// Token lock errors (70-79)
	ErrInsufficientTokens = errorsmod.Register(ModuleName, 70, "insufficient tokens for lock")
	ErrTokensLocked       = errorsmod.Register(ModuleName, 71, "tokens are locked")

	// Secret ballot errors (80-99)
	ErrInvalidSnapshot        = errorsmod.Register(ModuleName, 80, "invalid snapshot vote")
	ErrInvalidReveal          = errorsmod.Register(ModuleName, 81, "invalid vote reveal")
	ErrRevealPeriodNotStarted = errorsmod.Register(ModuleName, 82, "reveal period has not started")
	ErrRevealPeriodEnded      = errorsmod.Register(ModuleName, 83, "reveal period has ended")
	ErrInvalidCommitment      = errorsmod.Register(ModuleName, 84, "invalid vote commitment")
	ErrSecretBallotDisabled   = errorsmod.Register(ModuleName, 85, "secret ballot voting is disabled")
	ErrVoteAlreadyRevealed    = errorsmod.Register(ModuleName, 86, "vote has already been revealed")
	ErrInvalidVoteReveal      = errorsmod.Register(ModuleName, 87, "invalid vote reveal")
	ErrNoVoteCommitment       = errorsmod.Register(ModuleName, 88, "no vote commitment found")

	// Quadratic voting errors (100-109)
	ErrQuadraticVotingDisabled = errorsmod.Register(ModuleName, 100, "quadratic voting is disabled")
	ErrInsufficientVoteCredits = errorsmod.Register(ModuleName, 101, "insufficient vote credits")

	// Storage errors (900-909)
	ErrCorruptedData = errorsmod.Register(ModuleName, 900, "corrupted data in storage")
)
