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
func (k *Keeper) GetParams(ctx sdk.Context) *types.GovernanceParams {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ParamsKeyPrefix)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.GovernanceParams
	k.cdc.MustUnmarshal(bz, &params)
	return &params
}

// SetParams sets the governance parameters
func (k *Keeper) SetParams(ctx sdk.Context, params *types.GovernanceParams) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(params)
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
	binary.BigEndian.PutUint64(key[1:], proposal.Id)
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
	key := append(VotesKeyPrefix, sdk.Uint64ToBigEndian(vote.ProposalId)...)
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

// GetVotes retrieves all votes for a proposal
func (k *Keeper) GetVotes(ctx sdk.Context, proposalID uint64) []*types.Vote {
	store := ctx.KVStore(k.storeKey)
	prefix := append(VotesKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var votes []*types.Vote
	for ; iterator.Valid(); iterator.Next() {
		var vote types.Vote
		if err := k.cdc.Unmarshal(iterator.Value(), &vote); err != nil {
			continue
		}
		votes = append(votes, &vote)
	}
	return votes
}

// ============================================================================
// Deposit KVStore Methods
// ============================================================================

// SetDeposit stores a deposit in the KVStore
func (k *Keeper) SetDeposit(ctx sdk.Context, deposit *types.Deposit) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(deposit)
	if err != nil {
		return err
	}

	key := append(DepositsKeyPrefix, sdk.Uint64ToBigEndian(deposit.ProposalId)...)
	key = append(key, []byte(deposit.Depositor)...)
	store.Set(key, bz)
	return nil
}

// GetDeposit retrieves a deposit from the KVStore
func (k *Keeper) GetDeposit(ctx sdk.Context, proposalID uint64, depositor string) (*types.Deposit, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(DepositsKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	key = append(key, []byte(depositor)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrInvalidDeposit
	}

	var deposit types.Deposit
	return &deposit, k.cdc.Unmarshal(bz, &deposit)
}

// GetDeposits retrieves all deposits for a proposal
func (k *Keeper) GetDeposits(ctx sdk.Context, proposalID uint64) []*types.Deposit {
	store := ctx.KVStore(k.storeKey)
	prefix := append(DepositsKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var deposits []*types.Deposit
	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		if err := k.cdc.Unmarshal(iterator.Value(), &deposit); err != nil {
			continue
		}
		deposits = append(deposits, &deposit)
	}
	return deposits
}

// ============================================================================
// Vote Delegation KVStore Methods
// ============================================================================

// SetVoteDelegation stores a vote delegation
func (k *Keeper) SetVoteDelegation(ctx sdk.Context, delegation *types.VoteDelegation) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(delegation)
	if err != nil {
		return err
	}

	key := append(DelegationsKeyPrefix, []byte(delegation.Delegator)...)
	key = append(key, []byte(delegation.Delegate)...)
	store.Set(key, bz)
	return nil
}

// DeleteVoteDelegation removes a vote delegation
func (k *Keeper) DeleteVoteDelegation(ctx sdk.Context, delegator, delegate string) error {
	store := ctx.KVStore(k.storeKey)
	key := append(DelegationsKeyPrefix, []byte(delegator)...)
	key = append(key, []byte(delegate)...)
	store.Delete(key)
	return nil
}

// GetVoteDelegations retrieves all vote delegations for a delegator
func (k *Keeper) GetVoteDelegations(ctx sdk.Context, delegator string) []*types.VoteDelegation {
	store := ctx.KVStore(k.storeKey)
	prefix := append(DelegationsKeyPrefix, []byte(delegator)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var delegations []*types.VoteDelegation
	for ; iterator.Valid(); iterator.Next() {
		var delegation types.VoteDelegation
		if err := k.cdc.Unmarshal(iterator.Value(), &delegation); err != nil {
			continue
		}
		delegations = append(delegations, &delegation)
	}
	return delegations
}

// ============================================================================
// Veto Request KVStore Methods
// ============================================================================

// SetVetoRequest stores a veto request
func (k *Keeper) SetVetoRequest(ctx sdk.Context, veto *types.VetoRequest) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(veto)
	if err != nil {
		return err
	}

	key := append(VetoRequestsKeyPrefix, sdk.Uint64ToBigEndian(veto.ProposalId)...)
	store.Set(key, bz)
	return nil
}

// GetVetoRequest retrieves a veto request
func (k *Keeper) GetVetoRequest(ctx sdk.Context, proposalID uint64) (*types.VetoRequest, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(VetoRequestsKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrInvalidVeto
	}

	var veto types.VetoRequest
	return &veto, k.cdc.Unmarshal(bz, &veto)
}

// GetVetoRequests retrieves all veto requests for a proposal
func (k *Keeper) GetVetoRequests(ctx sdk.Context, proposalID uint64) []*types.VetoRequest {
	store := ctx.KVStore(k.storeKey)
	prefix := append(VetoRequestsKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var vetos []*types.VetoRequest
	for ; iterator.Valid(); iterator.Next() {
		var veto types.VetoRequest
		if err := k.cdc.Unmarshal(iterator.Value(), &veto); err != nil {
			continue
		}
		vetos = append(vetos, &veto)
	}
	return vetos
}

// ============================================================================
// Snapshot Vote KVStore Methods
// ============================================================================

// SetSnapshotVote stores a snapshot vote
func (k *Keeper) SetSnapshotVote(ctx sdk.Context, vote *types.SnapshotVote) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(vote)
	if err != nil {
		return err
	}

	key := append(SnapshotVotesKeyPrefix, sdk.Uint64ToBigEndian(vote.ProposalId)...)
	key = append(key, []byte(vote.Voter)...)
	store.Set(key, bz)
	return nil
}

// GetSnapshotVotes retrieves all snapshot votes for a proposal
func (k *Keeper) GetSnapshotVotes(ctx sdk.Context, proposalID uint64) []*types.SnapshotVote {
	store := ctx.KVStore(k.storeKey)
	prefix := append(SnapshotVotesKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var votes []*types.SnapshotVote
	for ; iterator.Valid(); iterator.Next() {
		var vote types.SnapshotVote
		if err := k.cdc.Unmarshal(iterator.Value(), &vote); err != nil {
			continue
		}
		votes = append(votes, &vote)
	}
	return votes
}

// ============================================================================
// Helper Methods
// ============================================================================

// CalculateTally calculates the tally for a proposal
func (k *Keeper) CalculateTally(ctx sdk.Context, proposalID uint64) *types.TallyResult {
	votes := k.GetVotes(ctx, proposalID)

	tally := &types.TallyResult{
		Yes:        "0",
		Abstain:    "0",
		No:         "0",
		NoWithVeto: "0",
	}

	// Simple counting (in production, would weight by voting power)
	for _, vote := range votes {
		switch vote.Option {
		case 1: // Yes
			tally.Yes = "1" // Simplified
		case 2: // Abstain
			tally.Abstain = "1"
		case 3: // No
			tally.No = "1"
		case 4: // NoWithVeto
			tally.NoWithVeto = "1"
		}
	}

	return tally
}

// GetVotingPower returns the voting power for an address
func (k *Keeper) GetVotingPower(ctx sdk.Context, address string) string {
	// Simplified: return "1000" for testing
	return "1000"
}

// GetDelegatedPower returns the delegated voting power for an address
func (k *Keeper) GetDelegatedPower(ctx sdk.Context, address string) string {
	// Simplified: return "0" for testing
	return "0"
}

// GetTokenLocks returns token locks for an address
func (k *Keeper) GetTokenLocks(ctx sdk.Context, address string) []*types.TokenLock {
	// Simplified: return empty slice
	return []*types.TokenLock{}
}
