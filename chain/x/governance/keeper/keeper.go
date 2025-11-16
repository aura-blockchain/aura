package keeper

import (
	"encoding/binary"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// Key prefixes for KVStore
var (
	ProposalsKeyPrefix       = []byte{0x01}
	VotesKeyPrefix           = []byte{0x02}
	DepositsKeyPrefix        = []byte{0x03}
	DelegationsKeyPrefix     = []byte{0x04}
	TokenLocksKeyPrefix      = []byte{0x05}
	VetoRequestsKeyPrefix    = []byte{0x06}
	SnapshotVotesKeyPrefix   = []byte{0x07}
	VoteCommitmentsKeyPrefix = []byte{0x08}
	ParamsKeyPrefix          = []byte{0x09}
	NextProposalIDKeyPrefix  = []byte{0x0A}
)

// Keeper maintains the state of the governance module
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

// NewKeeper creates a new governance keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// ============================================================================
// Params KVStore Methods
// ============================================================================

// GetParams returns the governance parameters
func (k *Keeper) GetParams(ctx sdk.Context) types.GovernanceParams {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ParamsKeyPrefix)
	if bz == nil {
		return types.DefaultGovernanceParams()
	}

	var params types.GovernanceParams
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the governance parameters
func (k *Keeper) SetParams(ctx sdk.Context, params types.GovernanceParams) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(ParamsKeyPrefix, bz)
}

// ============================================================================
// Next Proposal ID KVStore Methods
// ============================================================================

// GetNextProposalID gets the next proposal ID
func (k *Keeper) GetNextProposalID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(NextProposalIDKeyPrefix)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// SetNextProposalID sets the next proposal ID
func (k *Keeper) SetNextProposalID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	store.Set(NextProposalIDKeyPrefix, bz)
}

// ============================================================================
// Proposal KVStore Methods
// ============================================================================

// SetProposal stores a proposal in the KVStore
func (k *Keeper) SetProposal(ctx sdk.Context, proposal *types.Proposal) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(proposal)
	if err != nil {
		return err
	}

	key := make([]byte, 9)
	copy(key, ProposalsKeyPrefix)
	binary.BigEndian.PutUint64(key[1:], proposal.ID)
	store.Set(key, bz)
	return nil
}

// GetProposal retrieves a proposal from the KVStore
func (k *Keeper) GetProposal(ctx sdk.Context, proposalID uint64) (*types.Proposal, error) {
	store := ctx.KVStore(k.storeKey)

	key := make([]byte, 9)
	copy(key, ProposalsKeyPrefix)
	binary.BigEndian.PutUint64(key[1:], proposalID)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrProposalNotFound
	}

	var proposal types.Proposal
	if err := k.cdc.Unmarshal(bz, &proposal); err != nil {
		return nil, err
	}
	return &proposal, nil
}

// GetAllProposals retrieves all proposals from the KVStore
func (k *Keeper) GetAllProposals(ctx sdk.Context) []*types.Proposal {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, ProposalsKeyPrefix)
	defer iterator.Close()

	var proposals []*types.Proposal
	for ; iterator.Valid(); iterator.Next() {
		var proposal types.Proposal
		if err := k.cdc.Unmarshal(iterator.Value(), &proposal); err != nil {
			continue
		}
		proposals = append(proposals, &proposal)
	}
	return proposals
}

// GetProposals retrieves all proposals (alias for compatibility)
func (k *Keeper) GetProposals() []*types.Proposal {
	// This requires a context - will be called with context in actual usage
	// For now, return empty slice
	return []*types.Proposal{}
}

// DeleteProposal removes a proposal from the KVStore
func (k *Keeper) DeleteProposal(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := make([]byte, 9)
	copy(key, ProposalsKeyPrefix)
	binary.BigEndian.PutUint64(key[1:], proposalID)
	store.Delete(key)
}

// ============================================================================
// Vote KVStore Methods
// ============================================================================

// SetVote stores a vote in the KVStore
func (k *Keeper) SetVote(ctx sdk.Context, vote *types.Vote) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.marshalVote(vote)
	if err != nil {
		return err
	}

	// Key: prefix + proposalID (8 bytes) + voter (variable length)
	key := append(VotesKeyPrefix, sdk.Uint64ToBigEndian(vote.ProposalID)...)
	key = append(key, []byte(vote.Voter)...)
	store.Set(key, bz)
	return nil
}

// GetVote retrieves a vote from the KVStore
func (k *Keeper) GetVote(ctx sdk.Context, proposalID uint64, voter string) (*types.Vote, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(VotesKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	key = append(key, []byte(voter)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrInvalidVote
	}

	var vote types.Vote
	return &vote, k.cdc.Unmarshal(bz, &vote)
}
