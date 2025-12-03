package keeper

import (
	"fmt"
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// QUADRATIC VOTING (Feature 3)
// ============================

// Helper function to get VoteCreditsPerToken (can be moved to params in future)
func (k *Keeper) getVoteCreditsPerToken() uint64 {
	// Default: 1 token = 100 credits
	return 100
}

// Helper function to check if quadratic voting is enabled
func (k *Keeper) isQuadraticVotingEnabled(ctx sdk.Context) bool {
	// For now, return true - can be made configurable via params
	return true
}

// CastQuadraticVote allows casting a vote with quadratic cost
func (k *Keeper) CastQuadraticVote(
	ctx sdk.Context,
	proposalID uint64,
	voter string,
	option types.VoteOption,
	voteCredits uint64,
) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	// Check if quadratic voting is enabled
	if !k.isQuadraticVotingEnabled(ctx) {
		return types.ErrQuadraticVotingDisabled
	}

	// Verify proposal is in voting period
	if proposal.Status != types.StatusVotingPeriod {
		return types.ErrInvalidProposalStatus
	}

	// Get voter's available vote credits
	availableCredits := k.GetVoteCredits(ctx, voter)

	if voteCredits > availableCredits {
		return types.ErrInsufficientVoteCredits
	}

	// Calculate quadratic voting power
	// Voting power = sqrt(credits spent)
	votingPower := k.calculateQuadraticVotingPower(voteCredits)

	// Create vote
	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter,
		Option:      option,
		VotingPower: fmt.Sprintf("%d", votingPower),
		Timestamp:   nil,
		IsSecret:    false,
	}

	// Store vote with quadratic metadata in vote commitment field (reusing field)
	vote.VoteCommitment = fmt.Sprintf("quadratic:%d:%d", voteCredits, votingPower)

	// Deduct vote credits (implicitly tracked via votes)
	if err := k.DeductVoteCredits(ctx, voter, voteCredits); err != nil {
		return err
	}

	// Store vote
	if err := k.SetVote(ctx, vote); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"quadratic_vote_cast",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("credits_spent", fmt.Sprintf("%d", voteCredits)),
			sdk.NewAttribute("voting_power", fmt.Sprintf("%d", votingPower)),
		),
	)

	return nil
}

// calculateQuadraticVotingPower calculates voting power from credits (square root)
func (k *Keeper) calculateQuadraticVotingPower(credits uint64) uint64 {
	// Voting power = sqrt(credits) * 10000 (to preserve precision)
	sqrtCredits := math.Sqrt(float64(credits))
	return uint64(sqrtCredits * 10000)
}

// GetVoteCredits returns available vote credits for a voter
func (k *Keeper) GetVoteCredits(ctx sdk.Context, voter string) uint64 {
	// Base credits from token holdings
	votingPower, err := k.GetVotingPower(ctx, voter)
	if err != nil {
		return 0
	}
	power := votingPower.BigInt()

	// Convert to credits (1 token = X credits based on params)
	voteCreditsPerToken := k.getVoteCreditsPerToken()
	credits := new(big.Int).Mul(power, big.NewInt(int64(voteCreditsPerToken)))
	credits.Div(credits, big.NewInt(1000000000000000000)) // Adjust for token decimals

	// Get already spent credits
	spent := k.GetSpentVoteCredits(ctx, voter)

	if credits.Cmp(big.NewInt(int64(spent))) <= 0 {
		return 0
	}

	available := new(big.Int).Sub(credits, big.NewInt(int64(spent)))
	return available.Uint64()
}

// GetSpentVoteCredits returns total credits spent by voter
func (k *Keeper) GetSpentVoteCredits(ctx sdk.Context, voter string) uint64 {
	// Iterate through all proposals and sum spent credits
	totalSpent := uint64(0)

	allProposals := k.GetAllProposals(ctx)
	for _, proposal := range allProposals {
		vote, err := k.GetVote(ctx, proposal.Id, voter)
		if err == nil && vote.VoteCommitment != "" {
			// Parse quadratic metadata from vote commitment field
			var creditsSpent, votingPower uint64
			_, parseErr := fmt.Sscanf(vote.VoteCommitment, "quadratic:%d:%d", &creditsSpent, &votingPower)
			if parseErr == nil && creditsSpent > 0 {
				totalSpent += creditsSpent
			}
		}
	}

	return totalSpent
}

// DeductVoteCredits deducts vote credits from voter's balance
func (k *Keeper) DeductVoteCredits(ctx sdk.Context, voter string, credits uint64) error {
	// Credits are implicitly deducted by tracking spent amount in votes
	// No explicit storage needed as we calculate from votes
	return nil
}

// CalculateQuadraticTally calculates tally with quadratic voting power
func (k *Keeper) CalculateQuadraticTally(ctx sdk.Context, proposalID uint64) *types.QuadraticTallyResult {
	votes := k.GetVotes(ctx, proposalID)

	yesVotes := uint64(0)
	noVotes := uint64(0)
	abstainVotes := uint64(0)
	vetoVotes := uint64(0)

	totalCreditsSpent := uint64(0)
	uniqueVoters := uint64(0)

	for _, vote := range votes {
		// Parse quadratic metadata from vote commitment field
		var creditsSpent, votingPower uint64
		_, err := fmt.Sscanf(vote.VoteCommitment, "quadratic:%d:%d", &creditsSpent, &votingPower)

		if err == nil && creditsSpent > 0 {
			totalCreditsSpent += creditsSpent
			uniqueVoters++

			switch vote.Option {
			case types.OptionYes:
				yesVotes += votingPower
			case types.OptionNo:
				noVotes += votingPower
			case types.OptionAbstain:
				abstainVotes += votingPower
			case types.OptionNoWithVeto:
				vetoVotes += votingPower
			}
		}
	}

	return &types.QuadraticTallyResult{
		YesPower:          yesVotes,
		NoPower:           noVotes,
		AbstainPower:      abstainVotes,
		VetoPower:         vetoVotes,
		TotalCreditsSpent: totalCreditsSpent,
		UniqueVoters:      uniqueVoters,
	}
}

// GetQuadraticVotingStats returns quadratic voting statistics
func (k *Keeper) GetQuadraticVotingStats(ctx sdk.Context, proposalID uint64) *types.QuadraticVotingStats {
	tally := k.CalculateQuadraticTally(ctx, proposalID)

	totalPower := tally.YesPower + tally.NoPower + tally.AbstainPower + tally.VetoPower
	avgCreditsPerVoter := uint64(0)
	if tally.UniqueVoters > 0 {
		avgCreditsPerVoter = tally.TotalCreditsSpent / tally.UniqueVoters
	}

	// Calculate cost efficiency (power per credit)
	costEfficiency := float64(0)
	if tally.TotalCreditsSpent > 0 {
		costEfficiency = float64(totalPower) / float64(tally.TotalCreditsSpent)
	}

	return &types.QuadraticVotingStats{
		ProposalId:             proposalID,
		TotalVotingPower:       totalPower,
		TotalCreditsSpent:      tally.TotalCreditsSpent,
		UniqueVoters:           tally.UniqueVoters,
		AverageCreditsPerVoter: avgCreditsPerVoter,
		CostEfficiency:         costEfficiency,
		QuadraticAdvantage:     k.calculateQuadraticAdvantage(tally),
	}
}

// calculateQuadraticAdvantage calculates how much quadratic voting reduced influence concentration
func (k *Keeper) calculateQuadraticAdvantage(tally *types.QuadraticTallyResult) float64 {
	// Compare to what it would be with linear voting
	// Advantage = reduction in large holder influence
	// Simplified calculation
	if tally.TotalCreditsSpent == 0 {
		return 0
	}

	// If all credits were used linearly, total power = total credits
	linearPower := tally.TotalCreditsSpent
	quadraticPower := tally.YesPower + tally.NoPower + tally.AbstainPower + tally.VetoPower

	// Advantage = how much we reduced from linear (as percentage)
	if linearPower == 0 {
		return 0
	}

	reduction := float64(linearPower-quadraticPower) / float64(linearPower)
	return reduction * 100 // As percentage
}
