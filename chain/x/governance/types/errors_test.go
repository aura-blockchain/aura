// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"errors"
	"fmt"
	"testing"
)

func TestGovernanceErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrInvalidProposal", ErrInvalidProposal, "invalid proposal"},
		{"ErrInsufficientDeposit", ErrInsufficientDeposit, "insufficient deposit"},
		{"ErrInvalidVote", ErrInvalidVote, "invalid vote"},
		{"ErrProposalNotFound", ErrProposalNotFound, "proposal not found"},
		{"ErrInvalidProposalStatus", ErrInvalidProposalStatus, "invalid proposal status"},
		{"ErrVotingPeriodEnded", ErrVotingPeriodEnded, "voting period has ended"},
		{"ErrVotingPeriodNotStarted", ErrVotingPeriodNotStarted, "voting period has not started"},
		{"ErrAlreadyVoted", ErrAlreadyVoted, "already voted"},
		{"ErrInvalidDeposit", ErrInvalidDeposit, "invalid deposit"},
		{"ErrDepositPeriodEnded", ErrDepositPeriodEnded, "deposit period has ended"},
		{"ErrUnauthorizedVeto", ErrUnauthorizedVeto, "unauthorized veto"},
		{"ErrInsufficientVetoCosigners", ErrInsufficientVetoCosigners, "insufficient veto cosigners"},
		{"ErrDuplicateCosigner", ErrDuplicateCosigner, "cosigner has already signed this veto"},
		{"ErrExecutionDelayNotPassed", ErrExecutionDelayNotPassed, "execution delay has not passed"},
		{"ErrProposalNotPassed", ErrProposalNotPassed, "proposal has not passed"},
		{"ErrAlreadyExecuted", ErrAlreadyExecuted, "proposal already executed"},
		{"ErrInvalidDelegation", ErrInvalidDelegation, "invalid vote delegation"},
		{"ErrDelegationNotFound", ErrDelegationNotFound, "vote delegation not found"},
		{"ErrInsufficientTokens", ErrInsufficientTokens, "insufficient tokens for lock"},
		{"ErrTokensLocked", ErrTokensLocked, "tokens are locked"},
		{"ErrInvalidSnapshot", ErrInvalidSnapshot, "invalid snapshot vote"},
		{"ErrInvalidReveal", ErrInvalidReveal, "invalid vote reveal"},
		{"ErrRevealPeriodNotStarted", ErrRevealPeriodNotStarted, "reveal period has not started"},
		{"ErrRevealPeriodEnded", ErrRevealPeriodEnded, "reveal period has ended"},
		{"ErrInvalidCommitment", ErrInvalidCommitment, "invalid vote commitment"},
		{"ErrWeightedVoteNotEnabled", ErrWeightedVoteNotEnabled, "weighted voting not enabled"},
		{"ErrInvalidWeight", ErrInvalidWeight, "invalid vote weight"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected error message %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestErrorsAreErrors(t *testing.T) {
	var _ error = ErrInvalidProposal
	var _ error = ErrInsufficientDeposit
	var _ error = ErrInvalidVote
	var _ error = ErrProposalNotFound
	var _ error = ErrUnauthorizedVeto
	var _ error = ErrTokensLocked
}

func TestErrorComparison(t *testing.T) {
	err := ErrProposalNotFound
	if !errors.Is(err, ErrProposalNotFound) {
		t.Error("errors.Is should return true for same error")
	}

	if errors.Is(err, ErrInvalidProposal) {
		t.Error("errors.Is should return false for different error")
	}
}

func TestErrorWrapping(t *testing.T) {
	baseErr := ErrProposalNotFound
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	if !errors.Is(wrappedErr, ErrProposalNotFound) {
		t.Error("wrapped error should be detectable with errors.Is")
	}
}
