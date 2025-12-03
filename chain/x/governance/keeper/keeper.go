package keeper

import (
	"context"
	"encoding/binary"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// StakingKeeper defines the expected staking keeper interface
type StakingKeeper interface {
	GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (sdkmath.Int, error)
	TotalBondedTokens(ctx context.Context) (sdkmath.Int, error)
}

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
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	stakingKeeper StakingKeeper
	bankKeeper    types.BankKeeper
}

// NewKeeper creates a new governance keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, stakingKeeper StakingKeeper, bankKeeper types.BankKeeper) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		stakingKeeper: stakingKeeper,
		bankKeeper:    bankKeeper,
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

	return k.unmarshalVote(bz)
}

// GetVotes retrieves all votes for a proposal
func (k *Keeper) GetVotes(ctx sdk.Context, proposalID uint64) []*types.Vote {
	store := ctx.KVStore(k.storeKey)
	prefix := append(VotesKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var votes []*types.Vote
	for ; iterator.Valid(); iterator.Next() {
		vote, err := k.unmarshalVote(iterator.Value())
		if err != nil {
			continue
		}
		votes = append(votes, vote)
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

// DeleteDeposit removes a specific deposit from the KVStore
func (k *Keeper) DeleteDeposit(ctx sdk.Context, proposalID uint64, depositor string) {
	store := ctx.KVStore(k.storeKey)
	key := append(DepositsKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	key = append(key, []byte(depositor)...)
	store.Delete(key)
}

// DeleteDeposits removes all deposits for a proposal
func (k *Keeper) DeleteDeposits(ctx sdk.Context, proposalID uint64) error {
	deposits := k.GetDeposits(ctx, proposalID)
	for _, deposit := range deposits {
		k.DeleteDeposit(ctx, proposalID, deposit.Depositor)
	}
	return nil
}

// RefundDeposits refunds all deposits for a proposal
// This should be called when proposals pass or are rejected (but not vetoed)
func (k *Keeper) RefundDeposits(ctx sdk.Context, proposalID uint64) error {
	deposits := k.GetDeposits(ctx, proposalID)

	for _, deposit := range deposits {
		// Parse depositor address
		depositorAddr, err := sdk.AccAddressFromBech32(deposit.Depositor)
		if err != nil {
			// Log error but continue with other deposits
			ctx.Logger().Error("failed to parse depositor address", "depositor", deposit.Depositor, "error", err)
			continue
		}

		// Parse deposit amount
		coins, err := sdk.ParseCoinsNormalized(deposit.Amount)
		if err != nil {
			ctx.Logger().Error("failed to parse deposit amount", "amount", deposit.Amount, "error", err)
			continue
		}

		// Transfer coins from module account back to depositor
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositorAddr, coins)
		if err != nil {
			ctx.Logger().Error("failed to refund deposit", "depositor", deposit.Depositor, "amount", coins, "error", err)
			continue
		}

		ctx.Logger().Info("refunded deposit", "proposal_id", proposalID, "depositor", deposit.Depositor, "amount", coins)
	}

	// Delete all deposits after refunding
	return k.DeleteDeposits(ctx, proposalID)
}

// BurnDeposits burns all deposits for a proposal
// This should be called when proposals are vetoed
func (k *Keeper) BurnDeposits(ctx sdk.Context, proposalID uint64) error {
	// For now, we just delete the deposits without refunding
	// In a production system, you would actually burn the tokens by sending them
	// to a burner address or using the bank keeper's BurnCoins method
	deposits := k.GetDeposits(ctx, proposalID)

	for _, deposit := range deposits {
		ctx.Logger().Info("burning deposit (no refund)", "proposal_id", proposalID, "depositor", deposit.Depositor, "amount", deposit.Amount)
	}

	return k.DeleteDeposits(ctx, proposalID)
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

// DeleteVetoRequest removes a veto request from the KVStore
func (k *Keeper) DeleteVetoRequest(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := append(VetoRequestsKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	store.Delete(key)
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

// GetSnapshotVote retrieves a single snapshot vote
func (k *Keeper) GetSnapshotVote(ctx sdk.Context, proposalID uint64, voter string) (*types.SnapshotVote, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(SnapshotVotesKeyPrefix, sdk.Uint64ToBigEndian(proposalID)...)
	key = append(key, []byte(voter)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrInvalidSnapshot
	}

	var vote types.SnapshotVote
	return &vote, k.cdc.Unmarshal(bz, &vote)
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
// Vote counts are weighted by the voter's staking power and vote delegations
func (k *Keeper) CalculateTally(ctx sdk.Context, proposalID uint64) *types.TallyResult {
	votes := k.GetVotes(ctx, proposalID)

	// Use proper integers for accumulation
	var (
		yesVotes     = sdkmath.ZeroInt()
		noVotes      = sdkmath.ZeroInt()
		abstainVotes = sdkmath.ZeroInt()
		noWithVeto   = sdkmath.ZeroInt()
	)

	// Accumulate votes weighted by voting power
	for _, vote := range votes {
		// Get voter's actual voting power (includes staked tokens and delegations)
		voterPower, err := k.GetVotingPower(ctx, vote.Voter)
		if err != nil {
			// Skip votes from invalid voters
			continue
		}

		// Add the voter's power to the appropriate tally category
		switch vote.Option {
		case 1: // Yes
			yesVotes = yesVotes.Add(voterPower)
		case 2: // Abstain
			abstainVotes = abstainVotes.Add(voterPower)
		case 3: // No
			noVotes = noVotes.Add(voterPower)
		case 4: // NoWithVeto
			noWithVeto = noWithVeto.Add(voterPower)
		}
	}

	// Convert accumulated integers to strings for TallyResult
	return &types.TallyResult{
		Yes:        yesVotes.String(),
		Abstain:    abstainVotes.String(),
		No:         noVotes.String(),
		NoWithVeto: noWithVeto.String(),
	}
}

// GetVotingPower returns the voting power for an address based on staked tokens and delegations
// Returns the total voting power as sdkmath.Int and any error encountered
func (k *Keeper) GetVotingPower(ctx sdk.Context, address string) (sdkmath.Int, error) {
	totalPower := sdkmath.ZeroInt()

	// Parse the address
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return sdkmath.ZeroInt(), err
	}

	// 1. Get direct staked tokens from staking module
	// This is the validator's own bonded stake
	stakedAmount, err := k.stakingKeeper.GetDelegatorBonded(ctx, addr)
	if err != nil {
		// On error, continue with zero staked amount
		stakedAmount = sdkmath.ZeroInt()
	}
	totalPower = totalPower.Add(stakedAmount)

	// 2. Add voting power delegated TO this address (from others)
	// When someone delegates their vote to this address, this address gains their voting power
	delegatedPower := k.GetDelegatedVotingPower(ctx, address)
	totalPower = totalPower.Add(delegatedPower)

	// 3. Subtract voting power delegated AWAY from this address
	// When this address delegates their vote to someone else, they lose their voting power
	powerDelegatedAway := k.GetPowerDelegatedAway(ctx, address)
	totalPower = totalPower.Sub(powerDelegatedAway)

	// Ensure voting power is never negative (defensive programming)
	if totalPower.IsNegative() {
		totalPower = sdkmath.ZeroInt()
	}

	return totalPower, nil
}

// GetDelegatedVotingPower calculates the total voting power delegated TO this address
// This iterates all vote delegations and sums up the staked tokens of delegators
// who have delegated their voting power to this address
func (k *Keeper) GetDelegatedVotingPower(ctx sdk.Context, delegate string) sdkmath.Int {
	totalDelegated := sdkmath.ZeroInt()

	// Iterate all vote delegations in the store
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, DelegationsKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var delegation types.VoteDelegation
		if err := k.cdc.Unmarshal(iterator.Value(), &delegation); err != nil {
			continue // Skip invalid delegations
		}

		// Check if this delegation is TO the target address
		if delegation.Delegate == delegate {
			// Get the delegator's staked tokens
			delegatorAddr, err := sdk.AccAddressFromBech32(delegation.Delegator)
			if err != nil {
				continue // Skip invalid addresses
			}

			// Add the delegator's staked amount to the delegate's voting power
			delegatorStake, err := k.stakingKeeper.GetDelegatorBonded(ctx, delegatorAddr)
			if err != nil {
				continue // Skip on error
			}
			totalDelegated = totalDelegated.Add(delegatorStake)
		}
	}

	return totalDelegated
}

// GetPowerDelegatedAway calculates the voting power this address has delegated away
// When an address delegates their vote to someone else, they lose their voting power
func (k *Keeper) GetPowerDelegatedAway(ctx sdk.Context, delegator string) sdkmath.Int {
	// Get all delegations FROM this address
	delegations := k.GetVoteDelegations(ctx, delegator)

	// If this address has delegated their vote, they have delegated away their full staked power
	// Note: In this implementation, we assume all-or-nothing delegation
	// If the user has ANY active delegation, their entire stake is delegated away
	if len(delegations) > 0 {
		addr, err := sdk.AccAddressFromBech32(delegator)
		if err != nil {
			return sdkmath.ZeroInt()
		}

		// Return the full staked amount since it's delegated away
		stakedAmount, err := k.stakingKeeper.GetDelegatorBonded(ctx, addr)
		if err != nil {
			return sdkmath.ZeroInt()
		}
		return stakedAmount
	}

	return sdkmath.ZeroInt()
}

// GetDelegatedPower returns the delegated voting power for an address (legacy compatibility)
// Deprecated: Use GetDelegatedVotingPower instead
func (k *Keeper) GetDelegatedPower(ctx sdk.Context, address string) string {
	power := k.GetDelegatedVotingPower(ctx, address)
	return power.String()
}

// ============================================================================
// Token Lock KVStore Methods
// ============================================================================

// SetTokenLock stores a token lock in the KVStore
func (k *Keeper) SetTokenLock(ctx sdk.Context, lock *types.TokenLock) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(lock)
	if err != nil {
		return err
	}

	// Key: prefix + owner address + proposal ID
	key := append(TokenLocksKeyPrefix, []byte(lock.Owner)...)
	key = append(key, sdk.Uint64ToBigEndian(lock.ProposalId)...)
	store.Set(key, bz)
	return nil
}

// GetTokenLocks returns token locks for an address
func (k *Keeper) GetTokenLocks(ctx sdk.Context, address string) []*types.TokenLock {
	store := ctx.KVStore(k.storeKey)
	prefix := append(TokenLocksKeyPrefix, []byte(address)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var locks []*types.TokenLock
	for ; iterator.Valid(); iterator.Next() {
		var lock types.TokenLock
		if err := k.cdc.Unmarshal(iterator.Value(), &lock); err != nil {
			continue
		}
		locks = append(locks, &lock)
	}
	return locks
}

// DeleteTokenLock removes a token lock from the KVStore
func (k *Keeper) DeleteTokenLock(ctx sdk.Context, owner string, proposalID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := append(TokenLocksKeyPrefix, []byte(owner)...)
	key = append(key, sdk.Uint64ToBigEndian(proposalID)...)
	store.Delete(key)
}

// GetAllTokenLocks retrieves all token locks from the KVStore
func (k *Keeper) GetAllTokenLocks(ctx sdk.Context) []*types.TokenLock {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, TokenLocksKeyPrefix)
	defer iterator.Close()

	var locks []*types.TokenLock
	for ; iterator.Valid(); iterator.Next() {
		var lock types.TokenLock
		if err := k.cdc.Unmarshal(iterator.Value(), &lock); err != nil {
			ctx.Logger().Error("failed to unmarshal token lock", "error", err)
			continue
		}
		locks = append(locks, &lock)
	}
	return locks
}

// GetAllVoteDelegations retrieves all vote delegations from the KVStore
func (k *Keeper) GetAllVoteDelegations(ctx sdk.Context) []*types.VoteDelegation {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, DelegationsKeyPrefix)
	defer iterator.Close()

	var delegations []*types.VoteDelegation
	for ; iterator.Valid(); iterator.Next() {
		var delegation types.VoteDelegation
		if err := k.cdc.Unmarshal(iterator.Value(), &delegation); err != nil {
			ctx.Logger().Error("failed to unmarshal vote delegation", "error", err)
			continue
		}
		delegations = append(delegations, &delegation)
	}
	return delegations
}

// GetAllVotes retrieves all votes from the KVStore
func (k *Keeper) GetAllVotes(ctx sdk.Context) []*types.Vote {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, VotesKeyPrefix)
	defer iterator.Close()

	var votes []*types.Vote
	for ; iterator.Valid(); iterator.Next() {
		vote, err := k.unmarshalVote(iterator.Value())
		if err != nil {
			ctx.Logger().Error("failed to unmarshal vote", "error", err)
			continue
		}
		votes = append(votes, vote)
	}
	return votes
}

// GetAllDeposits retrieves all deposits from the KVStore
func (k *Keeper) GetAllDeposits(ctx sdk.Context) []*types.Deposit {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, DepositsKeyPrefix)
	defer iterator.Close()

	var deposits []*types.Deposit
	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		if err := k.cdc.Unmarshal(iterator.Value(), &deposit); err != nil {
			ctx.Logger().Error("failed to unmarshal deposit", "error", err)
			continue
		}
		deposits = append(deposits, &deposit)
	}
	return deposits
}

// GetAllVetoRequests retrieves all veto requests from the KVStore
func (k *Keeper) GetAllVetoRequests(ctx sdk.Context) []*types.VetoRequest {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, VetoRequestsKeyPrefix)
	defer iterator.Close()

	var vetos []*types.VetoRequest
	for ; iterator.Valid(); iterator.Next() {
		var veto types.VetoRequest
		if err := k.cdc.Unmarshal(iterator.Value(), &veto); err != nil {
			ctx.Logger().Error("failed to unmarshal veto request", "error", err)
			continue
		}
		vetos = append(vetos, &veto)
	}
	return vetos
}
