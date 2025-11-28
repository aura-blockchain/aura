package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// VOTE PRIVACY (SECRET BALLOT) (Feature 8)
// ============================

// CommitVote commits a vote hash for secret ballot voting
func (k *Keeper) CommitVote(
	ctx sdk.Context,
	proposalID uint64,
	voter string,
	voteHash string,
) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)

	if !params.SecretBallotEnabled {
		return types.ErrSecretBallotDisabled
	}

	if proposal.Status != types.StatusVotingPeriod {
		return types.ErrInvalidProposalStatus
	}

	// Create vote commitment
	commitment := &types.VoteCommitment{
		ProposalId:  proposalID,
		Voter:       voter,
		VoteHash:    voteHash,
		CommittedAt: timestamppb.Now(),
		Revealed:    false,
	}

	// Store commitment
	if err := k.setVoteCommitment(ctx, commitment); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_committed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("voter", voter),
		),
	)

	return nil
}

// RevealVote reveals a committed vote
func (k *Keeper) RevealVote(
	ctx sdk.Context,
	proposalID uint64,
	voter string,
	option types.VoteOption,
	salt string,
) error {
	// Get commitment
	commitment, err := k.getVoteCommitment(ctx, proposalID, voter)
	if err != nil {
		return err
	}

	if commitment.Revealed {
		return types.ErrVoteAlreadyRevealed
	}

	// Verify hash
	expectedHash := k.calculateVoteHash(proposalID, voter, option, salt)
	if expectedHash != commitment.VoteHash {
		return types.ErrInvalidVoteReveal
	}

	// Create actual vote
	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter,
		Option:      option,
		Timestamp:   timestamppb.Now(),
		IsSecret:    true,
		VotingPower: k.GetVotingPower(ctx, voter),
	}

	// Store vote
	if err := k.SetVote(ctx, vote); err != nil {
		return err
	}

	// Mark commitment as revealed
	commitment.Revealed = true
	k.setVoteCommitment(ctx, commitment)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_revealed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("option", fmt.Sprintf("%d", option)),
		),
	)

	return nil
}

// calculateVoteHash calculates hash for vote commitment
func (k *Keeper) calculateVoteHash(proposalID uint64, voter string, option types.VoteOption, salt string) string {
	data := fmt.Sprintf("%d:%s:%d:%s", proposalID, voter, option, salt)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// setVoteCommitment stores a vote commitment
func (k *Keeper) setVoteCommitment(ctx sdk.Context, commitment *types.VoteCommitment) error {
	store := ctx.KVStore(k.storeKey)
	key := append(VoteCommitmentsKeyPrefix, sdk.Uint64ToBigEndian(commitment.ProposalId)...)
	key = append(key, []byte(commitment.Voter)...)

	// Use JSON marshaling for custom types
	bz, err := json.Marshal(commitment)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// getVoteCommitment retrieves a vote commitment
func (k *Keeper) getVoteCommitment(ctx sdk.Context, proposalID uint64, voter string) (*types.VoteCommitment, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(VoteCommitmentsKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	key = append(key, []byte(voter)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrNoVoteCommitment
	}

	var commitment types.VoteCommitment
	if err := json.Unmarshal(bz, &commitment); err != nil {
		return nil, err
	}

	return &commitment, nil
}
