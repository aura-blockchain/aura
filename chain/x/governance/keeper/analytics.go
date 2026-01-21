// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// GOVERNANCE ANALYTICS (Feature 9)
// ============================

// GetGovernanceAnalytics returns comprehensive governance analytics
func (k *Keeper) GetGovernanceAnalytics(ctx sdk.Context) *types.GovernanceAnalytics {
	allProposals := k.GetAllProposals(ctx)

	analytics := &types.GovernanceAnalytics{
		TotalProposals:          uint64(len(allProposals)),
		ProposalsByStatus:       k.countProposalsByStatus(allProposals),
		ProposalsByType:         k.countProposalsByType(allProposals),
		AverageVotingPower:      k.calculateAverageVotingPower(ctx, allProposals),
		ParticipationRate:       k.calculateParticipationRate(ctx, allProposals),
		AverageProposalDuration: k.calculateAverageProposalDuration(allProposals),
		PassRate:                k.calculatePassRate(allProposals),
		VetoRate:                k.calculateVetoRate(allProposals),
	}

	return analytics
}

// countProposalsByStatus counts proposals by status
func (k *Keeper) countProposalsByStatus(proposals []*types.Proposal) map[string]uint64 {
	counts := make(map[string]uint64)

	for _, proposal := range proposals {
		statusStr := proposal.Status.String()
		counts[statusStr]++
	}

	return counts
}

// countProposalsByType counts proposals by type (using metadata or category)
func (k *Keeper) countProposalsByType(proposals []*types.Proposal) map[string]uint64 {
	counts := make(map[string]uint64)

	for _, proposal := range proposals {
		// Count by category
		categoryStr := proposal.Category.String()
		counts[categoryStr]++
	}

	return counts
}

// calculateAverageVotingPower calculates average voting power per proposal
func (k *Keeper) calculateAverageVotingPower(ctx sdk.Context, proposals []*types.Proposal) string {
	totalPower := big.NewInt(0)
	proposalCount := 0

	for _, proposal := range proposals {
		if proposal.FinalTallyResult != nil {
			yes := new(big.Int)
			yes.SetString(proposal.FinalTallyResult.Yes, 10)

			no := new(big.Int)
			no.SetString(proposal.FinalTallyResult.No, 10)

			abstain := new(big.Int)
			abstain.SetString(proposal.FinalTallyResult.Abstain, 10)

			veto := new(big.Int)
			veto.SetString(proposal.FinalTallyResult.NoWithVeto, 10)

			proposalTotal := new(big.Int)
			proposalTotal.Add(yes, no)
			proposalTotal.Add(proposalTotal, abstain)
			proposalTotal.Add(proposalTotal, veto)

			totalPower.Add(totalPower, proposalTotal)
			proposalCount++
		}
	}

	if proposalCount == 0 {
		return "0"
	}

	avgPower := new(big.Int).Div(totalPower, big.NewInt(int64(proposalCount)))
	return avgPower.String()
}

// calculateParticipationRate calculates voter participation rate using deterministic decimal math.
// Returns sdkmath.LegacyDec for consensus-safe cross-platform calculations.
// Note: Uses sdkmath.LegacyDec instead of float64 to prevent non-deterministic floating point
// operations that could cause state divergence across different validator hardware/platforms.
func (k *Keeper) calculateParticipationRate(ctx sdk.Context, proposals []*types.Proposal) sdkmath.LegacyDec {
	if len(proposals) == 0 {
		return sdkmath.LegacyZeroDec()
	}

	totalParticipation := sdkmath.LegacyZeroDec()
	votedProposals := int64(0)

	for _, proposal := range proposals {
		if proposal.Status == types.StatusVotingPeriod ||
			proposal.Status == types.StatusPassed ||
			proposal.Status == types.StatusFailed ||
			proposal.Status == types.StatusRejected ||
			proposal.Status == types.StatusExecuted {

			votes := k.GetVotes(ctx, proposal.Id)
			if len(votes) > 0 {
				// Simplified - would calculate actual participation rate
				totalParticipation = totalParticipation.Add(sdkmath.LegacyNewDec(int64(len(votes))))
				votedProposals++
			}
		}
	}

	if votedProposals == 0 {
		return sdkmath.LegacyZeroDec()
	}

	// Calculate: (totalParticipation / votedProposals) * 10
	return totalParticipation.Quo(sdkmath.LegacyNewDec(votedProposals)).MulInt64(10)
}

// calculateAverageProposalDuration calculates average proposal duration
func (k *Keeper) calculateAverageProposalDuration(proposals []*types.Proposal) uint64 {
	totalDuration := uint64(0)
	completedProposals := 0

	for _, proposal := range proposals {
		if proposal.SubmitTime != nil {
			var endTime int64
			if proposal.ExecutionTime != nil {
				endTime = proposal.ExecutionTime.Seconds
			} else if proposal.VotingEndTime != nil {
				endTime = proposal.VotingEndTime.Seconds
			} else {
				continue
			}

			duration := uint64(endTime - proposal.SubmitTime.Seconds)
			totalDuration += duration
			completedProposals++
		}
	}

	if completedProposals == 0 {
		return 0
	}

	return totalDuration / uint64(completedProposals)
}

// calculatePassRate calculates proposal pass rate using deterministic decimal math.
// Returns sdkmath.LegacyDec for consensus-safe cross-platform calculations.
// Note: Uses sdkmath.LegacyDec instead of float64 to prevent non-deterministic floating point
// operations that could cause state divergence across different validator hardware/platforms.
func (k *Keeper) calculatePassRate(proposals []*types.Proposal) sdkmath.LegacyDec {
	if len(proposals) == 0 {
		return sdkmath.LegacyZeroDec()
	}

	passed := int64(0)
	total := int64(0)

	for _, proposal := range proposals {
		if proposal.Status == types.StatusPassed ||
			proposal.Status == types.StatusExecuted ||
			proposal.Status == types.StatusFailed ||
			proposal.Status == types.StatusRejected {

			total++
			if proposal.Status == types.StatusPassed || proposal.Status == types.StatusExecuted {
				passed++
			}
		}
	}

	if total == 0 {
		return sdkmath.LegacyZeroDec()
	}

	// Calculate: (passed / total) * 100
	return sdkmath.LegacyNewDec(passed).Quo(sdkmath.LegacyNewDec(total)).MulInt64(100)
}

// calculateVetoRate calculates proposal veto rate using deterministic decimal math.
// Returns sdkmath.LegacyDec for consensus-safe cross-platform calculations.
// Note: Uses sdkmath.LegacyDec instead of float64 to prevent non-deterministic floating point
// operations that could cause state divergence across different validator hardware/platforms.
func (k *Keeper) calculateVetoRate(proposals []*types.Proposal) sdkmath.LegacyDec {
	if len(proposals) == 0 {
		return sdkmath.LegacyZeroDec()
	}

	vetoed := int64(0)
	total := int64(0)

	for _, proposal := range proposals {
		if proposal.Status == types.StatusPassed ||
			proposal.Status == types.StatusExecuted ||
			proposal.Status == types.StatusFailed ||
			proposal.Status == types.StatusRejected {

			total++
			if proposal.Status == types.StatusRejected {
				vetoed++
			}
		}
	}

	if total == 0 {
		return sdkmath.LegacyZeroDec()
	}

	// Calculate: (vetoed / total) * 100
	return sdkmath.LegacyNewDec(vetoed).Quo(sdkmath.LegacyNewDec(total)).MulInt64(100)
}

// GetProposerAnalytics returns analytics for a specific proposer
func (k *Keeper) GetProposerAnalytics(ctx sdk.Context, proposer string) *types.ProposerAnalytics {
	allProposals := k.GetAllProposals(ctx)
	proposerProposals := []*types.Proposal{}

	for _, proposal := range allProposals {
		if proposal.Proposer == proposer {
			proposerProposals = append(proposerProposals, proposal)
		}
	}

	return &types.ProposerAnalytics{
		Proposer:          proposer,
		TotalProposals:    uint64(len(proposerProposals)),
		PassedProposals:   k.countByStatus(proposerProposals, types.StatusPassed),
		FailedProposals:   k.countByStatus(proposerProposals, types.StatusFailed),
		RejectedProposals: k.countByStatus(proposerProposals, types.StatusRejected),
		ExecutedProposals: k.countByStatus(proposerProposals, types.StatusExecuted),
		SuccessRate:       k.calculatePassRate(proposerProposals),
	}
}

// countByStatus counts proposals with specific status
func (k *Keeper) countByStatus(proposals []*types.Proposal, status types.ProposalStatus) uint64 {
	count := uint64(0)
	for _, proposal := range proposals {
		if proposal.Status == status {
			count++
		}
	}
	return count
}
