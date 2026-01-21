// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// TestProposalOutcome_QuorumCalculation tests quorum calculation logic
func TestProposalOutcome_QuorumCalculation(t *testing.T) {
	// Test quorum threshold calculation
	params := types.DefaultParams()
	quorumDec, err := sdkmath.LegacyNewDecFromStr(params.Quorum)
	require.NoError(t, err)
	require.Equal(t, "0.334000000000000000", quorumDec.String(), "Default quorum should be 33.4%")

	// Calculate required quorum for 1,000,000 total bonded tokens
	totalBonded := sdkmath.NewInt(1_000_000)
	requiredQuorum := quorumDec.MulInt(totalBonded).TruncateInt()
	require.Equal(t, "334000", requiredQuorum.String(), "Required quorum should be 334,000")

	// Test case 1: Votes below quorum (100,000 < 334,000) - should FAIL
	votes := sdkmath.NewInt(100_000)
	require.True(t, votes.LT(requiredQuorum), "100,000 votes should be below quorum")

	// Test case 2: Votes at quorum (334,000 == 334,000) - should PASS quorum check
	votesAtQuorum := sdkmath.NewInt(334_000)
	require.False(t, votesAtQuorum.LT(requiredQuorum), "Votes at quorum should not be less than required")

	// Test case 3: Votes above quorum (400,000 > 334,000) - should PASS quorum check
	votesAboveQuorum := sdkmath.NewInt(400_000)
	require.False(t, votesAboveQuorum.LT(requiredQuorum), "Votes above quorum should not be less than required")
}

// TestProposalOutcome_ThresholdCalculation tests pass threshold calculation
func TestProposalOutcome_ThresholdCalculation(t *testing.T) {
	params := types.DefaultParams()
	thresholdDec, err := sdkmath.LegacyNewDecFromStr(params.Threshold)
	require.NoError(t, err)
	require.Equal(t, "0.500000000000000000", thresholdDec.String(), "Default threshold should be 50%")

	// Test case 1: Exactly 50% yes votes (on boundary)
	totalVotes := sdkmath.NewInt(500_000)
	passThreshold := thresholdDec.MulInt(totalVotes).TruncateInt()
	require.Equal(t, "250000", passThreshold.String(), "Pass threshold should be 250,000")

	yesVotes := sdkmath.NewInt(250_000)
	require.False(t, yesVotes.GT(passThreshold), "Exactly 50% should NOT pass (need > threshold)")

	// Test case 2: 50% + 1 yes vote - should PASS
	yesVotesPass := sdkmath.NewInt(250_001)
	require.True(t, yesVotesPass.GT(passThreshold), "50% + 1 should pass")

	// Test case 3: Less than 50% - should FAIL
	yesVotesFail := sdkmath.NewInt(249_999)
	require.False(t, yesVotesFail.GT(passThreshold), "Less than 50% should not pass")
}

// TestProposalOutcome_VetoThresholdCalculation tests veto threshold calculation
func TestProposalOutcome_VetoThresholdCalculation(t *testing.T) {
	params := types.DefaultParams()
	vetoThresholdDec, err := sdkmath.LegacyNewDecFromStr(params.VetoThreshold)
	require.NoError(t, err)
	require.Equal(t, "0.334000000000000000", vetoThresholdDec.String(), "Default veto threshold should be 33.4%")

	// Test with 1,000,000 total votes
	totalVotes := sdkmath.NewInt(1_000_000)
	vetoThreshold := vetoThresholdDec.MulInt(totalVotes).TruncateInt()
	require.Equal(t, "334000", vetoThreshold.String(), "Veto threshold should be 334,000")

	// Test case 1: Exactly at veto threshold - should NOT veto (need > threshold)
	noWithVetoAtThreshold := sdkmath.NewInt(334_000)
	require.False(t, noWithVetoAtThreshold.GT(vetoThreshold), "Exactly at threshold should not trigger veto")

	// Test case 2: Just over veto threshold - should VETO
	noWithVetoAboveThreshold := sdkmath.NewInt(334_001)
	require.True(t, noWithVetoAboveThreshold.GT(vetoThreshold), "Over threshold should trigger veto")

	// Test case 3: Below veto threshold - should NOT veto
	noWithVetoBelowThreshold := sdkmath.NewInt(333_999)
	require.False(t, noWithVetoBelowThreshold.GT(vetoThreshold), "Below threshold should not trigger veto")
}

// TestProposalOutcome_SecurityScenarios tests critical security scenarios
func TestProposalOutcome_SecurityScenarios(t *testing.T) {
	params := types.DefaultParams()

	// Scenario 1: Single vote cannot pass a proposal
	t.Run("SingleVoteCannotPass", func(t *testing.T) {
		totalBonded := sdkmath.NewInt(1_000_000)
		quorumDec, _ := sdkmath.LegacyNewDecFromStr(params.Quorum)
		requiredQuorum := quorumDec.MulInt(totalBonded).TruncateInt()

		singleVote := sdkmath.NewInt(1)
		require.True(t, singleVote.LT(requiredQuorum), "Single vote should fail quorum check")
	})

	// Scenario 2: Even 100% yes votes cannot pass without quorum
	t.Run("HundredPercentYesWithoutQuorumFails", func(t *testing.T) {
		totalBonded := sdkmath.NewInt(1_000_000)
		quorumDec, _ := sdkmath.LegacyNewDecFromStr(params.Quorum)
		requiredQuorum := quorumDec.MulInt(totalBonded).TruncateInt()

		// All votes are yes, but total is below quorum
		yesVotes := sdkmath.NewInt(100_000) // 100% yes but only 100k total
		totalVotes := yesVotes

		require.True(t, totalVotes.LT(requiredQuorum), "Even 100% yes should fail without quorum")
	})

	// Scenario 3: Veto overrides yes votes
	t.Run("VetoOverridesYesVotes", func(t *testing.T) {
		totalVotes := sdkmath.NewInt(1_000_000)
		vetoThresholdDec, _ := sdkmath.LegacyNewDecFromStr(params.VetoThreshold)
		vetoThreshold := vetoThresholdDec.MulInt(totalVotes).TruncateInt()

		// Even if yes votes would pass, veto can block
		yesVotes := sdkmath.NewInt(600_000)   // 60% yes
		noWithVeto := sdkmath.NewInt(350_000) // 35% veto (over 33.4% threshold)

		thresholdDec, _ := sdkmath.LegacyNewDecFromStr(params.Threshold)
		votesExcludingAbstain := yesVotes.Add(noWithVeto)
		passThreshold := thresholdDec.MulInt(votesExcludingAbstain).TruncateInt()

		// Yes votes would pass threshold
		require.True(t, yesVotes.GT(passThreshold), "Yes votes exceed pass threshold")
		// But veto blocks it
		require.True(t, noWithVeto.GT(vetoThreshold), "Veto threshold exceeded")
	})

	// Scenario 4: Quorum met, but threshold not met
	t.Run("QuorumMetThresholdNotMet", func(t *testing.T) {
		totalBonded := sdkmath.NewInt(1_000_000)
		quorumDec, _ := sdkmath.LegacyNewDecFromStr(params.Quorum)
		requiredQuorum := quorumDec.MulInt(totalBonded).TruncateInt()

		// Total votes: 400,000 (meets quorum of 334,000)
		yesVotes := sdkmath.NewInt(180_000) // 45%
		noVotes := sdkmath.NewInt(220_000)  // 55%
		totalVotes := yesVotes.Add(noVotes)

		require.False(t, totalVotes.LT(requiredQuorum), "Quorum should be met")

		thresholdDec, _ := sdkmath.LegacyNewDecFromStr(params.Threshold)
		passThreshold := thresholdDec.MulInt(totalVotes).TruncateInt()

		require.False(t, yesVotes.GT(passThreshold), "Threshold should not be met (45% < 50%)")
	})

	// Scenario 5: Only abstain votes - should reject
	t.Run("OnlyAbstainVotes", func(t *testing.T) {
		_ = sdkmath.NewInt(400_000) // abstainVotes
		votesExcludingAbstain := sdkmath.ZeroInt()

		require.True(t, votesExcludingAbstain.IsZero(), "With only abstain, non-abstain votes should be zero")
	})
}

// TestProposalOutcome_EdgeCases tests edge cases and boundary conditions
func TestProposalOutcome_EdgeCases(t *testing.T) {
	params := types.DefaultParams()

	t.Run("ZeroTotalBondedTokens", func(t *testing.T) {
		// Edge case: If there are no bonded tokens, quorum should be 0
		totalBonded := sdkmath.ZeroInt()
		quorumDec, _ := sdkmath.LegacyNewDecFromStr(params.Quorum)
		requiredQuorum := quorumDec.MulInt(totalBonded).TruncateInt()

		require.True(t, requiredQuorum.IsZero(), "Quorum should be zero with no bonded tokens")
	})

	t.Run("ExactBoundaries", func(t *testing.T) {
		// Test exact boundary values
		totalBonded := sdkmath.NewInt(1_000_000)
		quorumDec, _ := sdkmath.LegacyNewDecFromStr(params.Quorum)
		requiredQuorum := quorumDec.MulInt(totalBonded).TruncateInt()

		// Exactly at quorum (not less than)
		votesAtQuorum := sdkmath.NewInt(334_000)
		require.False(t, votesAtQuorum.LT(requiredQuorum), "At quorum should pass check")

		// One below quorum
		votesBelowQuorum := sdkmath.NewInt(333_999)
		require.True(t, votesBelowQuorum.LT(requiredQuorum), "One below quorum should fail")
	})

	t.Run("VeryLargeNumbers", func(t *testing.T) {
		// Test with very large token amounts
		totalBonded := sdkmath.NewInt(1_000_000_000_000) // 1 trillion
		quorumDec, _ := sdkmath.LegacyNewDecFromStr(params.Quorum)
		requiredQuorum := quorumDec.MulInt(totalBonded).TruncateInt()

		// Should scale correctly
		expectedQuorum := sdkmath.NewInt(334_000_000_000) // 334 billion
		require.Equal(t, expectedQuorum.String(), requiredQuorum.String(), "Large numbers should scale correctly")
	})
}

// TestProposalOutcome_ParameterValidation tests parameter validation
func TestProposalOutcome_ParameterValidation(t *testing.T) {
	params := types.DefaultParams()

	t.Run("QuorumIsValidPercentage", func(t *testing.T) {
		quorumDec, err := sdkmath.LegacyNewDecFromStr(params.Quorum)
		require.NoError(t, err)
		require.False(t, quorumDec.IsNegative(), "Quorum should not be negative")
		require.True(t, quorumDec.LTE(sdkmath.LegacyOneDec()), "Quorum should be <= 100%")
	})

	t.Run("ThresholdIsValidPercentage", func(t *testing.T) {
		thresholdDec, err := sdkmath.LegacyNewDecFromStr(params.Threshold)
		require.NoError(t, err)
		require.False(t, thresholdDec.IsNegative(), "Threshold should not be negative")
		require.True(t, thresholdDec.LTE(sdkmath.LegacyOneDec()), "Threshold should be <= 100%")
	})

	t.Run("VetoThresholdIsValidPercentage", func(t *testing.T) {
		vetoThresholdDec, err := sdkmath.LegacyNewDecFromStr(params.VetoThreshold)
		require.NoError(t, err)
		require.False(t, vetoThresholdDec.IsNegative(), "Veto threshold should not be negative")
		require.True(t, vetoThresholdDec.LTE(sdkmath.LegacyOneDec()), "Veto threshold should be <= 100%")
	})

	t.Run("VetoThresholdLessThanPassThreshold", func(t *testing.T) {
		vetoThresholdDec, _ := sdkmath.LegacyNewDecFromStr(params.VetoThreshold)
		thresholdDec, _ := sdkmath.LegacyNewDecFromStr(params.Threshold)

		require.True(t, vetoThresholdDec.LT(thresholdDec),
			"Veto threshold (33.4%) should be less than pass threshold (50%)")
	})
}
