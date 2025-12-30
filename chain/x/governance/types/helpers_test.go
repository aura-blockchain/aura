// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	durationpb "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
)

func TestGetVotingPeriodSeconds(t *testing.T) {
	tests := []struct {
		name     string
		params   *GovernanceParams
		expected uint64
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			expected: 604800,
		},
		{
			name: "nil voting period returns default",
			params: &GovernanceParams{
				VotingPeriod: nil,
			},
			expected: 604800,
		},
		{
			name: "valid voting period returns value",
			params: &GovernanceParams{
				VotingPeriod: &durationpb.Duration{Seconds: 86400},
			},
			expected: 86400,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetVotingPeriodSeconds(tc.params)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestGetDepositPeriodSeconds(t *testing.T) {
	tests := []struct {
		name     string
		params   *GovernanceParams
		expected uint64
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			expected: 604800,
		},
		{
			name: "nil deposit period returns default",
			params: &GovernanceParams{
				MaxDepositPeriod: nil,
			},
			expected: 604800,
		},
		{
			name: "valid deposit period returns value",
			params: &GovernanceParams{
				MaxDepositPeriod: &durationpb.Duration{Seconds: 172800},
			},
			expected: 172800,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetDepositPeriodSeconds(tc.params)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestGetExecutionDelaySeconds(t *testing.T) {
	tests := []struct {
		name     string
		params   *GovernanceParams
		expected uint64
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			expected: 172800,
		},
		{
			name: "nil execution delay returns default",
			params: &GovernanceParams{
				ExecutionDelay: nil,
			},
			expected: 172800,
		},
		{
			name: "valid execution delay returns value",
			params: &GovernanceParams{
				ExecutionDelay: &durationpb.Duration{Seconds: 3600},
			},
			expected: 3600,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetExecutionDelaySeconds(tc.params)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestGetMaxDelegationsPerUser(t *testing.T) {
	result := GetMaxDelegationsPerUser(nil)
	require.Equal(t, uint64(10), result)

	result = GetMaxDelegationsPerUser(&GovernanceParams{})
	require.Equal(t, uint64(10), result)
}

func TestGetRefundFailedProposals(t *testing.T) {
	result := GetRefundFailedProposals(nil)
	require.False(t, result)

	result = GetRefundFailedProposals(&GovernanceParams{})
	require.False(t, result)
}

func TestGetFailedProposalRefundPercentage(t *testing.T) {
	result := GetFailedProposalRefundPercentage(nil)
	require.Equal(t, uint64(5000), result)

	result = GetFailedProposalRefundPercentage(&GovernanceParams{})
	require.Equal(t, uint64(5000), result)
}

func TestValidateGovernanceParams(t *testing.T) {
	tests := []struct {
		name    string
		params  *GovernanceParams
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
			errMsg:  "governance params cannot be nil",
		},
		{
			name:    "empty min_deposit",
			params:  &GovernanceParams{MinDeposit: "", Quorum: "0.4", Threshold: "0.5"},
			wantErr: true,
			errMsg:  "min_deposit cannot be empty",
		},
		{
			name:    "empty quorum",
			params:  &GovernanceParams{MinDeposit: "1000", Quorum: "", Threshold: "0.5"},
			wantErr: true,
			errMsg:  "quorum cannot be empty",
		},
		{
			name:    "empty threshold",
			params:  &GovernanceParams{MinDeposit: "1000", Quorum: "0.4", Threshold: ""},
			wantErr: true,
			errMsg:  "threshold cannot be empty",
		},
		{
			name:    "valid params",
			params:  &GovernanceParams{MinDeposit: "1000", Quorum: "0.4", Threshold: "0.5"},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGovernanceParams(tc.params)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
