// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"

	governancev1beta1 "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

// ============================================================================
// parseProposalStatus Tests
// ============================================================================

func TestParseProposalStatus(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    governancev1beta1.ProposalStatus
		expectError bool
	}{
		// Deposit period variants
		{
			name:        "deposit-period",
			input:       "deposit-period",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
			expectError: false,
		},
		{
			name:        "deposit",
			input:       "deposit",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
			expectError: false,
		},
		{
			name:        "deposit with underscore",
			input:       "deposit_period",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
			expectError: false,
		},
		// Voting period variants
		{
			name:        "voting-period",
			input:       "voting-period",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
			expectError: false,
		},
		{
			name:        "voting",
			input:       "voting",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
			expectError: false,
		},
		{
			name:        "voting with underscore",
			input:       "voting_period",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
			expectError: false,
		},
		// Passed variants
		{
			name:        "passed",
			input:       "passed",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_PASSED,
			expectError: false,
		},
		{
			name:        "pass",
			input:       "pass",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_PASSED,
			expectError: false,
		},
		// Rejected variants
		{
			name:        "rejected",
			input:       "rejected",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_REJECTED,
			expectError: false,
		},
		{
			name:        "reject",
			input:       "reject",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_REJECTED,
			expectError: false,
		},
		// Failed variants
		{
			name:        "failed",
			input:       "failed",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_FAILED,
			expectError: false,
		},
		{
			name:        "fail",
			input:       "fail",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_FAILED,
			expectError: false,
		},
		// Vetoed variants
		{
			name:        "vetoed",
			input:       "vetoed",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VETOED,
			expectError: false,
		},
		{
			name:        "veto",
			input:       "veto",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VETOED,
			expectError: false,
		},
		// Execution delay variants
		{
			name:        "execution-delay",
			input:       "execution-delay",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY,
			expectError: false,
		},
		{
			name:        "timelock",
			input:       "timelock",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY,
			expectError: false,
		},
		{
			name:        "execution_delay with underscore",
			input:       "execution_delay",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY,
			expectError: false,
		},
		// Ready for execution variants
		{
			name:        "ready-for-execution",
			input:       "ready-for-execution",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION,
			expectError: false,
		},
		{
			name:        "ready",
			input:       "ready",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION,
			expectError: false,
		},
		{
			name:        "ready_for_execution with underscores",
			input:       "ready_for_execution",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION,
			expectError: false,
		},
		// Executed variants
		{
			name:        "executed",
			input:       "executed",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTED,
			expectError: false,
		},
		{
			name:        "exec",
			input:       "exec",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTED,
			expectError: false,
		},
		// Case insensitivity tests
		{
			name:        "uppercase PASSED",
			input:       "PASSED",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_PASSED,
			expectError: false,
		},
		{
			name:        "mixed case Voting",
			input:       "Voting",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
			expectError: false,
		},
		// Whitespace handling
		{
			name:        "with leading/trailing whitespace",
			input:       "  passed  ",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_PASSED,
			expectError: false,
		},
		// Invalid inputs
		{
			name:        "invalid status",
			input:       "invalid",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			expectError: true,
		},
		{
			name:        "random garbage",
			input:       "xyz123",
			expected:    governancev1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseProposalStatus(tt.input)
			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid status")
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// FormatProposalStatus Tests
// ============================================================================

func TestFormatProposalStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   governancev1beta1.ProposalStatus
		expected string
	}{
		{
			name:     "deposit period",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
			expected: "DEPOSIT_PERIOD",
		},
		{
			name:     "voting period",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
			expected: "VOTING_PERIOD",
		},
		{
			name:     "passed",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_PASSED,
			expected: "PASSED",
		},
		{
			name:     "rejected",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_REJECTED,
			expected: "REJECTED",
		},
		{
			name:     "failed",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_FAILED,
			expected: "FAILED",
		},
		{
			name:     "vetoed",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VETOED,
			expected: "VETOED",
		},
		{
			name:     "execution delay",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY,
			expected: "EXECUTION_DELAY",
		},
		{
			name:     "ready for execution",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION,
			expected: "READY_FOR_EXECUTION",
		},
		{
			name:     "executed",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTED,
			expected: "EXECUTED",
		},
		{
			name:     "unspecified",
			status:   governancev1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			expected: "UNSPECIFIED",
		},
		{
			name:     "unknown status (default case)",
			status:   governancev1beta1.ProposalStatus(999),
			expected: "UNSPECIFIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatProposalStatus(tt.status)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// FormatVoteOption Tests
// ============================================================================

func TestFormatVoteOption(t *testing.T) {
	tests := []struct {
		name     string
		option   governancev1beta1.VoteOption
		expected string
	}{
		{
			name:     "yes",
			option:   governancev1beta1.VoteOption_VOTE_OPTION_YES,
			expected: "YES",
		},
		{
			name:     "no",
			option:   governancev1beta1.VoteOption_VOTE_OPTION_NO,
			expected: "NO",
		},
		{
			name:     "abstain",
			option:   governancev1beta1.VoteOption_VOTE_OPTION_ABSTAIN,
			expected: "ABSTAIN",
		},
		{
			name:     "no with veto",
			option:   governancev1beta1.VoteOption_VOTE_OPTION_NO_WITH_VETO,
			expected: "NO_WITH_VETO",
		},
		{
			name:     "unspecified",
			option:   governancev1beta1.VoteOption_VOTE_OPTION_UNSPECIFIED,
			expected: "UNSPECIFIED",
		},
		{
			name:     "unknown option (default case)",
			option:   governancev1beta1.VoteOption(999),
			expected: "UNSPECIFIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatVoteOption(tt.option)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// FormatProposalCategory Tests
// ============================================================================

func TestFormatProposalCategory(t *testing.T) {
	tests := []struct {
		name     string
		category governancev1beta1.ProposalCategory
		expected string
	}{
		{
			name:     "text",
			category: governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
			expected: "TEXT",
		},
		{
			name:     "parameter change",
			category: governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE,
			expected: "PARAMETER_CHANGE",
		},
		{
			name:     "software upgrade",
			category: governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE,
			expected: "SOFTWARE_UPGRADE",
		},
		{
			name:     "spending",
			category: governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_SPENDING,
			expected: "SPENDING",
		},
		{
			name:     "emergency",
			category: governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY,
			expected: "EMERGENCY",
		},
		{
			name:     "constitution",
			category: governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION,
			expected: "CONSTITUTION",
		},
		{
			name:     "unspecified",
			category: governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_UNSPECIFIED,
			expected: "UNSPECIFIED",
		},
		{
			name:     "unknown category (default case)",
			category: governancev1beta1.ProposalCategory(999),
			expected: "UNSPECIFIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatProposalCategory(tt.category)
			require.Equal(t, tt.expected, result)
		})
	}
}
