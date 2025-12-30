// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewProposalSubmittedEvent(t *testing.T) {
	event := NewProposalSubmittedEvent(1, "aura1proposer", "Test Proposal", 100, "2024-01-01T00:00:00Z")

	require.Equal(t, "1", event[AttributeKeyProposalID])
	require.Equal(t, "aura1proposer", event[AttributeKeyProposer])
	require.Equal(t, "Test Proposal", event[AttributeKeyTitle])
	require.Equal(t, "100", event[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-01T00:00:00Z", event[AttributeKeyBlockTime])
}

func TestNewProposalSubmittedEvent_LargeValues(t *testing.T) {
	event := NewProposalSubmittedEvent(999999999, "aura1verylongaddress", "Very Long Title That Is Very Long", 9999999999, "2024-12-31T23:59:59Z")

	require.Equal(t, "999999999", event[AttributeKeyProposalID])
	require.Equal(t, "aura1verylongaddress", event[AttributeKeyProposer])
	require.Equal(t, "Very Long Title That Is Very Long", event[AttributeKeyTitle])
	require.Equal(t, "9999999999", event[AttributeKeyBlockHeight])
}

func TestEventConstants(t *testing.T) {
	// Verify event type constants are defined
	require.NotEmpty(t, EventTypeProposalSubmitted)
	require.NotEmpty(t, EventTypeProposalDeposit)
	require.NotEmpty(t, EventTypeVoteCast)
	require.NotEmpty(t, EventTypeProposalPassed)
	require.NotEmpty(t, EventTypeProposalRejected)
	require.NotEmpty(t, EventTypeProposalExpired)
	require.NotEmpty(t, EventTypeParamsUpdated)
}

func TestAttributeKeyConstants(t *testing.T) {
	// Verify attribute key constants are defined
	require.NotEmpty(t, AttributeKeyProposalID)
	require.NotEmpty(t, AttributeKeyProposer)
	require.NotEmpty(t, AttributeKeyTitle)
	require.NotEmpty(t, AttributeKeyDepositor)
	require.NotEmpty(t, AttributeKeyDepositAmount)
	require.NotEmpty(t, AttributeKeyTotalDeposit)
	require.NotEmpty(t, AttributeKeyVoter)
	require.NotEmpty(t, AttributeKeyOption)
	require.NotEmpty(t, AttributeKeyVotingPower)
	require.NotEmpty(t, AttributeKeyYesVotes)
	require.NotEmpty(t, AttributeKeyNoVotes)
	require.NotEmpty(t, AttributeKeyAbstainVotes)
	require.NotEmpty(t, AttributeKeyVetoVotes)
	require.NotEmpty(t, AttributeKeyTotalVotes)
	require.NotEmpty(t, AttributeKeyQuorum)
	require.NotEmpty(t, AttributeKeyThreshold)
	require.NotEmpty(t, AttributeKeyBlockHeight)
	require.NotEmpty(t, AttributeKeyBlockTime)
}
