// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

func TestParseProposalStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected v1beta1.ProposalStatus
	}{
		{
			name:     "pending status",
			input:    "pending",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_PENDING,
		},
		{
			name:     "approved status",
			input:    "approved",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_APPROVED,
		},
		{
			name:     "executed status",
			input:    "executed",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTED,
		},
		{
			name:     "rejected status",
			input:    "rejected",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_REJECTED,
		},
		{
			name:     "expired status",
			input:    "expired",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_EXPIRED,
		},
		{
			name:     "unspecified for unknown input",
			input:    "unknown",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
		},
		{
			name:     "unspecified for empty string",
			input:    "",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
		},
		{
			name:     "case sensitive - uppercase PENDING returns unspecified",
			input:    "PENDING",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
		},
		{
			name:     "case sensitive - mixed case Pending returns unspecified",
			input:    "Pending",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
		},
		{
			name:     "invalid with whitespace",
			input:    " pending ",
			expected: v1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseProposalStatus(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseActionStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected v1beta1.ActionStatus
	}{
		{
			name:     "pending status",
			input:    "pending",
			expected: v1beta1.ActionStatus_ACTION_STATUS_PENDING,
		},
		{
			name:     "ready status",
			input:    "ready",
			expected: v1beta1.ActionStatus_ACTION_STATUS_READY,
		},
		{
			name:     "executed status",
			input:    "executed",
			expected: v1beta1.ActionStatus_ACTION_STATUS_EXECUTED,
		},
		{
			name:     "cancelled status",
			input:    "cancelled",
			expected: v1beta1.ActionStatus_ACTION_STATUS_CANCELLED,
		},
		{
			name:     "unspecified for unknown input",
			input:    "unknown",
			expected: v1beta1.ActionStatus_ACTION_STATUS_UNSPECIFIED,
		},
		{
			name:     "unspecified for empty string",
			input:    "",
			expected: v1beta1.ActionStatus_ACTION_STATUS_UNSPECIFIED,
		},
		{
			name:     "case sensitive - uppercase READY returns unspecified",
			input:    "READY",
			expected: v1beta1.ActionStatus_ACTION_STATUS_UNSPECIFIED,
		},
		{
			name:     "case sensitive - mixed case Ready returns unspecified",
			input:    "Ready",
			expected: v1beta1.ActionStatus_ACTION_STATUS_UNSPECIFIED,
		},
		{
			name:     "invalid with whitespace",
			input:    " ready ",
			expected: v1beta1.ActionStatus_ACTION_STATUS_UNSPECIFIED,
		},
		{
			name:     "american spelling canceled returns unspecified",
			input:    "canceled",
			expected: v1beta1.ActionStatus_ACTION_STATUS_UNSPECIFIED,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseActionStatus(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseWalletType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected v1beta1.WalletType
	}{
		{
			name:     "3-of-5 wallet type lowercase",
			input:    "3-of-5",
			expected: v1beta1.WalletType_WALLET_TYPE_3_OF_5,
		},
		{
			name:     "5-of-7 wallet type lowercase",
			input:    "5-of-7",
			expected: v1beta1.WalletType_WALLET_TYPE_5_OF_7,
		},
		{
			name:     "custom wallet type lowercase",
			input:    "custom",
			expected: v1beta1.WalletType_WALLET_TYPE_CUSTOM,
		},
		{
			name:     "3-of-5 uppercase",
			input:    "3-OF-5",
			expected: v1beta1.WalletType_WALLET_TYPE_3_OF_5,
		},
		{
			name:     "5-of-7 uppercase",
			input:    "5-OF-7",
			expected: v1beta1.WalletType_WALLET_TYPE_5_OF_7,
		},
		{
			name:     "CUSTOM uppercase",
			input:    "CUSTOM",
			expected: v1beta1.WalletType_WALLET_TYPE_CUSTOM,
		},
		{
			name:     "Custom mixed case",
			input:    "Custom",
			expected: v1beta1.WalletType_WALLET_TYPE_CUSTOM,
		},
		{
			name:     "unspecified for unknown input",
			input:    "unknown",
			expected: v1beta1.WalletType_WALLET_TYPE_UNSPECIFIED,
		},
		{
			name:     "unspecified for empty string",
			input:    "",
			expected: v1beta1.WalletType_WALLET_TYPE_UNSPECIFIED,
		},
		{
			name:     "invalid wallet type 2-of-3",
			input:    "2-of-3",
			expected: v1beta1.WalletType_WALLET_TYPE_UNSPECIFIED,
		},
		{
			name:     "invalid wallet type 7-of-9",
			input:    "7-of-9",
			expected: v1beta1.WalletType_WALLET_TYPE_UNSPECIFIED,
		},
		{
			name:     "invalid with whitespace",
			input:    " custom ",
			expected: v1beta1.WalletType_WALLET_TYPE_UNSPECIFIED,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseWalletType(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
