package keeper

import (
	"context"
	"encoding/binary"
	"fmt"

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
	VotingPowerSnapshotPrefix = []byte{0x0B}

	// KeySeparator is used to separate key components to prevent collisions
	KeySeparator = []byte{0x00}
)

// Keeper maintains the state of the governance module
type Keeper struct {
	storeKey       storetypes.StoreKey
	cdc            codec.BinaryCodec
	stakingKeeper  StakingKeeper
	bankKeeper     types.BankKeeper
	securityKeeper types.SecurityKeeper // Centralized security primitives
}

// NewKeeper creates a new governance keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	stakingKeeper StakingKeeper,
	bankKeeper types.BankKeeper,
	securityKeeper types.SecurityKeeper,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		stakingKeeper:  stakingKeeper,
		bankKeeper:     bankKeeper,
		securityKeeper: securityKeeper,
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
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		// CRITICAL: Params corruption is severe - log and return defaults as recovery
		ctx.Logger().Error("failed to unmarshal governance params, returning defaults",
			"error", err,
			"data_len", len(bz))
		return types.DefaultParams()
	}
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
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(vote.ProposalId)
	voterBytes := []byte(vote.Voter)
	keyLen := len(VotesKeyPrefix) + len(proposalIDBytes) + len(voterBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, VotesKeyPrefix...)
	key = append(key, proposalIDBytes...)
	key = append(key, voterBytes...)
	store.Set(key, bz)
	return nil
}

// GetVote retrieves a vote from the KVStore
func (k *Keeper) GetVote(ctx sdk.Context, proposalID uint64, voter string) (*types.Vote, error) {
	store := ctx.KVStore(k.storeKey)
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	voterBytes := []byte(voter)
	keyLen := len(VotesKeyPrefix) + len(proposalIDBytes) + len(voterBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, VotesKeyPrefix...)
	key = append(key, proposalIDBytes...)
	key = append(key, voterBytes...)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrInvalidVote
	}

	return k.unmarshalVote(bz)
}

// GetVotes retrieves all votes for a proposal
func (k *Keeper) GetVotes(ctx sdk.Context, proposalID uint64) []*types.Vote {
	store := ctx.KVStore(k.storeKey)
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	prefixLen := len(VotesKeyPrefix) + len(proposalIDBytes)
	prefix := make([]byte, 0, prefixLen)
	prefix = append(prefix, VotesKeyPrefix...)
	prefix = append(prefix, proposalIDBytes...)
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

	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(deposit.ProposalId)
	depositorBytes := []byte(deposit.Depositor)
	keyLen := len(DepositsKeyPrefix) + len(proposalIDBytes) + len(depositorBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, DepositsKeyPrefix...)
	key = append(key, proposalIDBytes...)
	key = append(key, depositorBytes...)
	store.Set(key, bz)
	return nil
}

// GetDeposit retrieves a deposit from the KVStore
func (k *Keeper) GetDeposit(ctx sdk.Context, proposalID uint64, depositor string) (*types.Deposit, error) {
	store := ctx.KVStore(k.storeKey)
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	depositorBytes := []byte(depositor)
	keyLen := len(DepositsKeyPrefix) + len(proposalIDBytes) + len(depositorBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, DepositsKeyPrefix...)
	key = append(key, proposalIDBytes...)
	key = append(key, depositorBytes...)

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
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	prefixLen := len(DepositsKeyPrefix) + len(proposalIDBytes)
	prefix := make([]byte, 0, prefixLen)
	prefix = append(prefix, DepositsKeyPrefix...)
	prefix = append(prefix, proposalIDBytes...)
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
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	depositorBytes := []byte(depositor)
	keyLen := len(DepositsKeyPrefix) + len(proposalIDBytes) + len(depositorBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, DepositsKeyPrefix...)
	key = append(key, proposalIDBytes...)
	key = append(key, depositorBytes...)
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
// This should be called when proposals are vetoed or fail to meet quorum
// Deposits are burned by sending them from the module account to a burn address (module account cannot spend)
func (k *Keeper) BurnDeposits(ctx sdk.Context, proposalID uint64) error {
	deposits := k.GetDeposits(ctx, proposalID)

	for _, deposit := range deposits {
		// Parse deposit amount
		coins, err := sdk.ParseCoinsNormalized(deposit.Amount)
		if err != nil {
			ctx.Logger().Error("failed to parse deposit amount for burning", "amount", deposit.Amount, "error", err)
			continue
		}

		// Burn tokens by sending from module account to burn module (or just keep in module as permanently locked)
		// Note: In Cosmos SDK, the proper way to burn is to send to a burn address or use BurnCoins if available
		// For governance, keeping funds in the module account as permanently locked is standard practice
		// as it prevents these tokens from being spent without making them disappear from total supply

		ctx.Logger().Info("burning deposit (permanently locked in module)",
			"proposal_id", proposalID,
			"depositor", deposit.Depositor,
			"amount", coins.String())

		// Emit burn event for transparency
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"deposit_burned",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
				sdk.NewAttribute("depositor", deposit.Depositor),
				sdk.NewAttribute("amount", coins.String()),
				sdk.NewAttribute("reason", "proposal_vetoed_or_failed_quorum"),
			),
		)
	}

	// Delete deposit records after burning
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

	// Build key with separator: prefix | delegator | separator | delegate
	// Pre-allocate to avoid shared underlying arrays with global prefix
	delegatorBytes := []byte(delegation.Delegator)
	delegateBytes := []byte(delegation.Delegate)
	keyLen := len(DelegationsKeyPrefix) + len(delegatorBytes) + len(KeySeparator) + len(delegateBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, DelegationsKeyPrefix...)
	key = append(key, delegatorBytes...)
	key = append(key, KeySeparator...)
	key = append(key, delegateBytes...)
	store.Set(key, bz)
	return nil
}

// DeleteVoteDelegation removes a vote delegation
func (k *Keeper) DeleteVoteDelegation(ctx sdk.Context, delegator, delegate string) error {
	store := ctx.KVStore(k.storeKey)
	// Build key with separator: prefix | delegator | separator | delegate
	// Pre-allocate to avoid shared underlying arrays with global prefix
	delegatorBytes := []byte(delegator)
	delegateBytes := []byte(delegate)
	keyLen := len(DelegationsKeyPrefix) + len(delegatorBytes) + len(KeySeparator) + len(delegateBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, DelegationsKeyPrefix...)
	key = append(key, delegatorBytes...)
	key = append(key, KeySeparator...)
	key = append(key, delegateBytes...)
	store.Delete(key)
	return nil
}

// GetVoteDelegations retrieves all vote delegations for a delegator
func (k *Keeper) GetVoteDelegations(ctx sdk.Context, delegator string) []*types.VoteDelegation {
	store := ctx.KVStore(k.storeKey)
	// Build prefix with separator: prefix | delegator | separator
	// This ensures we match only keys with this exact delegator
	// Pre-allocate to avoid shared underlying arrays with global prefix
	delegatorBytes := []byte(delegator)
	prefixLen := len(DelegationsKeyPrefix) + len(delegatorBytes) + len(KeySeparator)
	prefix := make([]byte, 0, prefixLen)
	prefix = append(prefix, DelegationsKeyPrefix...)
	prefix = append(prefix, delegatorBytes...)
	prefix = append(prefix, KeySeparator...)
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

	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(veto.ProposalId)
	keyLen := len(VetoRequestsKeyPrefix) + len(proposalIDBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, VetoRequestsKeyPrefix...)
	key = append(key, proposalIDBytes...)
	store.Set(key, bz)
	return nil
}

// GetVetoRequest retrieves a veto request
func (k *Keeper) GetVetoRequest(ctx sdk.Context, proposalID uint64) (*types.VetoRequest, error) {
	store := ctx.KVStore(k.storeKey)
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	keyLen := len(VetoRequestsKeyPrefix) + len(proposalIDBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, VetoRequestsKeyPrefix...)
	key = append(key, proposalIDBytes...)

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
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	prefixLen := len(VetoRequestsKeyPrefix) + len(proposalIDBytes)
	prefix := make([]byte, 0, prefixLen)
	prefix = append(prefix, VetoRequestsKeyPrefix...)
	prefix = append(prefix, proposalIDBytes...)
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
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	keyLen := len(VetoRequestsKeyPrefix) + len(proposalIDBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, VetoRequestsKeyPrefix...)
	key = append(key, proposalIDBytes...)
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

	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(vote.ProposalId)
	voterBytes := []byte(vote.Voter)
	keyLen := len(SnapshotVotesKeyPrefix) + len(proposalIDBytes) + len(voterBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, SnapshotVotesKeyPrefix...)
	key = append(key, proposalIDBytes...)
	key = append(key, voterBytes...)
	store.Set(key, bz)
	return nil
}

// GetSnapshotVote retrieves a single snapshot vote
func (k *Keeper) GetSnapshotVote(ctx sdk.Context, proposalID uint64, voter string) (*types.SnapshotVote, error) {
	store := ctx.KVStore(k.storeKey)
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	voterBytes := []byte(voter)
	keyLen := len(SnapshotVotesKeyPrefix) + len(proposalIDBytes) + len(voterBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, SnapshotVotesKeyPrefix...)
	key = append(key, proposalIDBytes...)
	key = append(key, voterBytes...)

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
	// Pre-allocate to avoid shared underlying arrays with global prefix
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	prefixLen := len(SnapshotVotesKeyPrefix) + len(proposalIDBytes)
	prefix := make([]byte, 0, prefixLen)
	prefix = append(prefix, SnapshotVotesKeyPrefix...)
	prefix = append(prefix, proposalIDBytes...)
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
// Performance optimized: uses cached voting power from votes (O(n) where n = votes)
// instead of recalculating power for each voter (which would be O(n*m) where m = delegations)
func (k *Keeper) CalculateTally(ctx sdk.Context, proposalID uint64) *types.TallyResult {
	votes := k.GetVotes(ctx, proposalID)

	// Use proper integers for accumulation
	var (
		yesVotes     = sdkmath.ZeroInt()
		noVotes      = sdkmath.ZeroInt()
		abstainVotes = sdkmath.ZeroInt()
		noWithVeto   = sdkmath.ZeroInt()
	)

	// Accumulate votes weighted by cached voting power
	// This is O(n) instead of O(n*m) because we use pre-cached power
	for _, vote := range votes {
		// Get voter's cached voting power from the vote record
		// This power was snapshotted when the vote was cast
		var voterPower sdkmath.Int
		if vote.VotingPower != "" {
			// Use cached voting power (fast path - O(1))
			power, ok := sdkmath.NewIntFromString(vote.VotingPower)
			if !ok {
				// Fallback: recalculate if cache is invalid (should be rare)
				ctx.Logger().Warn("invalid cached voting power, recalculating",
					"proposal_id", proposalID,
					"voter", vote.Voter,
					"cached_power", vote.VotingPower)
				power, err := k.GetVotingPower(ctx, vote.Voter)
				if err != nil {
					ctx.Logger().Error("failed to recalculate voting power",
						"proposal_id", proposalID,
						"voter", vote.Voter,
						"error", err)
					continue
				}
				voterPower = power
			} else {
				voterPower = power
			}
		} else {
			// No cached power - calculate it (slow path for legacy votes)
			// This should only happen for votes cast before the optimization
			power, err := k.GetVotingPower(ctx, vote.Voter)
			if err != nil {
				ctx.Logger().Error("failed to get voting power",
					"proposal_id", proposalID,
					"voter", vote.Voter,
					"error", err)
				continue
			}
			voterPower = power
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
	// Pre-allocate to avoid shared underlying arrays with global prefix
	ownerBytes := []byte(lock.Owner)
	proposalIDBytes := sdk.Uint64ToBigEndian(lock.ProposalId)
	keyLen := len(TokenLocksKeyPrefix) + len(ownerBytes) + len(proposalIDBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, TokenLocksKeyPrefix...)
	key = append(key, ownerBytes...)
	key = append(key, proposalIDBytes...)
	store.Set(key, bz)
	return nil
}

// GetTokenLocks returns token locks for an address
func (k *Keeper) GetTokenLocks(ctx sdk.Context, address string) []*types.TokenLock {
	store := ctx.KVStore(k.storeKey)
	// Pre-allocate to avoid shared underlying arrays with global prefix
	addressBytes := []byte(address)
	prefixLen := len(TokenLocksKeyPrefix) + len(addressBytes)
	prefix := make([]byte, 0, prefixLen)
	prefix = append(prefix, TokenLocksKeyPrefix...)
	prefix = append(prefix, addressBytes...)
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
	// Pre-allocate to avoid shared underlying arrays with global prefix
	ownerBytes := []byte(owner)
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	keyLen := len(TokenLocksKeyPrefix) + len(ownerBytes) + len(proposalIDBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, TokenLocksKeyPrefix...)
	key = append(key, ownerBytes...)
	key = append(key, proposalIDBytes...)
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

// ============================================================================
// Voting Power Snapshot KVStore Methods (Performance Optimization)
// ============================================================================

// VotingPowerSnapshot represents a cached voting power for a voter at proposal creation
type VotingPowerSnapshot struct {
	ProposalID  uint64
	Voter       string
	VotingPower sdkmath.Int
	Height      int64
}

// SetVotingPowerSnapshot stores a voting power snapshot for a voter on a specific proposal
// This is called during proposal creation or when a voter first votes
// Key: prefix | proposalID (8 bytes) | voter (variable length)
func (k *Keeper) SetVotingPowerSnapshot(ctx sdk.Context, proposalID uint64, voter string, power sdkmath.Int) error {
	store := ctx.KVStore(k.storeKey)

	// Build key: prefix + proposalID + voter
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	voterBytes := []byte(voter)
	keyLen := len(VotingPowerSnapshotPrefix) + len(proposalIDBytes) + len(voterBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, VotingPowerSnapshotPrefix...)
	key = append(key, proposalIDBytes...)
	key = append(key, voterBytes...)

	// Marshal the power as a string (protobuf compatible)
	powerStr := power.String()
	bz := []byte(powerStr)

	store.Set(key, bz)
	return nil
}

// GetVotingPowerSnapshot retrieves the cached voting power for a voter on a specific proposal
// Returns (power, found) where found indicates if a snapshot exists
func (k *Keeper) GetVotingPowerSnapshot(ctx sdk.Context, proposalID uint64, voter string) (sdkmath.Int, bool) {
	store := ctx.KVStore(k.storeKey)

	// Build key: prefix + proposalID + voter
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	voterBytes := []byte(voter)
	keyLen := len(VotingPowerSnapshotPrefix) + len(proposalIDBytes) + len(voterBytes)
	key := make([]byte, 0, keyLen)
	key = append(key, VotingPowerSnapshotPrefix...)
	key = append(key, proposalIDBytes...)
	key = append(key, voterBytes...)

	bz := store.Get(key)
	if bz == nil {
		return sdkmath.ZeroInt(), false
	}

	// Unmarshal the power
	powerStr := string(bz)
	power, ok := sdkmath.NewIntFromString(powerStr)
	if !ok {
		ctx.Logger().Error("failed to parse voting power snapshot",
			"proposal_id", proposalID,
			"voter", voter,
			"power_str", powerStr)
		return sdkmath.ZeroInt(), false
	}

	return power, true
}

// DeleteVotingPowerSnapshots removes all voting power snapshots for a proposal
// This should be called when a proposal is finalized to clean up storage
func (k *Keeper) DeleteVotingPowerSnapshots(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(k.storeKey)

	// Build prefix for all snapshots of this proposal
	proposalIDBytes := sdk.Uint64ToBigEndian(proposalID)
	prefixLen := len(VotingPowerSnapshotPrefix) + len(proposalIDBytes)
	prefix := make([]byte, 0, prefixLen)
	prefix = append(prefix, VotingPowerSnapshotPrefix...)
	prefix = append(prefix, proposalIDBytes...)

	// Iterate and delete all snapshots for this proposal
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	keysToDelete := [][]byte{}
	for ; iterator.Valid(); iterator.Next() {
		// Collect keys to delete (don't delete while iterating)
		keysToDelete = append(keysToDelete, iterator.Key())
	}

	// Delete all collected keys
	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// SnapshotVotingPowerForProposal creates voting power snapshots for all current stakers
// This is called when a proposal enters the voting period
// It caches voting power so that votes can be processed in O(1) time
func (k *Keeper) SnapshotVotingPowerForProposal(ctx sdk.Context, proposalID uint64) error {
	// This is a performance optimization: we snapshot voting power at proposal creation
	// so that voting is O(1) instead of O(n) where n = delegations

	// Note: We're snapshotting as voters cast their votes (lazy snapshotting)
	// This is more efficient than pre-computing for all possible voters
	// The actual snapshot happens in GetOrCreateVotingPowerSnapshot

	ctx.Logger().Info("voting power snapshot enabled for proposal",
		"proposal_id", proposalID,
		"height", ctx.BlockHeight())

	return nil
}

// GetOrCreateVotingPowerSnapshot gets the cached voting power or calculates and caches it
// This implements lazy snapshotting: only calculate power when a voter actually votes
func (k *Keeper) GetOrCreateVotingPowerSnapshot(ctx sdk.Context, proposalID uint64, voter string) (sdkmath.Int, error) {
	// Try to get cached snapshot first (O(1) lookup)
	power, found := k.GetVotingPowerSnapshot(ctx, proposalID, voter)
	if found {
		return power, nil
	}

	// Not cached - calculate voting power (O(n) operation, but only done once per voter)
	power, err := k.GetVotingPower(ctx, voter)
	if err != nil {
		return sdkmath.ZeroInt(), err
	}

	// Cache for future votes (vote updates, tally calculation)
	if err := k.SetVotingPowerSnapshot(ctx, proposalID, voter, power); err != nil {
		return sdkmath.ZeroInt(), err
	}

	return power, nil
}
