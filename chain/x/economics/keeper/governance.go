// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	"context"
	"encoding/binary"
	"math/big"

	"cosmossdk.io/math"

	"github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// ============================
// PROPOSAL OPERATIONS
// ============================

// GetNextProposalID gets the next proposal ID
func (k Keeper) GetNextProposalID(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.NextProposalIDKey)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 1, nil
	}
	return binary.BigEndian.Uint64(bz), nil
}

// SetNextProposalID sets the next proposal ID
func (k Keeper) SetNextProposalID(ctx context.Context, id uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return store.Set(types.NextProposalIDKey, bz)
}

// SetProposal stores a proposal
func (k Keeper) SetProposal(ctx context.Context, proposal *economicspb.Proposal) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetProposalKey(proposal.Id)
	bz, err := k.cdc.Marshal(proposal)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return store.Set(key, bz)
}

// GetProposal retrieves a proposal
func (k Keeper) GetProposal(ctx context.Context, proposalID uint64) (*economicspb.Proposal, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetProposalKey(proposalID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrProposalNotFound
	}

	var proposal economicspb.Proposal
	if err := k.cdc.Unmarshal(bz, &proposal); err != nil {
		return nil, err
	}
	return &proposal, nil
}

// DeleteProposal removes a proposal
func (k Keeper) DeleteProposal(ctx context.Context, proposalID uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetProposalKey(proposalID)
	return store.Delete(key)
}

// IterateProposals iterates over all proposals
func (k Keeper) IterateProposals(ctx context.Context, cb func(proposal *economicspb.Proposal) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ProposalPrefix, storeprefixend(types.ProposalPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var proposal economicspb.Proposal
		if err := k.cdc.Unmarshal(iterator.Value(), &proposal); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&proposal) {
			break
		}
	}
	return nil
}

// ============================
// VOTE OPERATIONS
// ============================

// SetVote stores a vote
func (k Keeper) SetVote(ctx context.Context, vote *economicspb.Vote) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteKey(vote.ProposalId, vote.Voter)
	bz, err := k.cdc.Marshal(vote)
	if err != nil {
		return fmt.Errorf("failed to marshal for ProposalId: %w", err)
	}
	return store.Set(key, bz)
}

// GetVote retrieves a vote
func (k Keeper) GetVote(ctx context.Context, proposalID uint64, voter string) (*economicspb.Vote, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteKey(proposalID, voter)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrInvalidVote
	}

	var vote economicspb.Vote
	if err := k.cdc.Unmarshal(bz, &vote); err != nil {
		return nil, err
	}
	return &vote, nil
}

// IterateVotes iterates over all votes for a proposal
func (k Keeper) IterateVotes(ctx context.Context, proposalID uint64, cb func(vote *economicspb.Vote) bool) error {
	store := k.storeService.OpenKVStore(ctx)

	// Create prefix for this proposal's votes
	proposalPrefix := make([]byte, 9)
	copy(proposalPrefix, types.VotePrefix)
	binary.BigEndian.PutUint64(proposalPrefix[1:], proposalID)

	iterator, err := store.Iterator(proposalPrefix, storeprefixend(proposalPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var vote economicspb.Vote
		if err := k.cdc.Unmarshal(iterator.Value(), &vote); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&vote) {
			break
		}
	}
	return nil
}

// ============================
// DEPOSIT OPERATIONS
// ============================

// SetDeposit stores a deposit
func (k Keeper) SetDeposit(ctx context.Context, deposit *economicspb.Deposit) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetDepositKey(deposit.ProposalId, deposit.Depositor)
	bz, err := k.cdc.Marshal(deposit)
	if err != nil {
		return fmt.Errorf("failed to marshal for ProposalId: %w", err)
	}
	return store.Set(key, bz)
}

// GetDeposit retrieves a deposit
func (k Keeper) GetDeposit(ctx context.Context, proposalID uint64, depositor string) (*economicspb.Deposit, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetDepositKey(proposalID, depositor)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrInvalidDeposit
	}

	var deposit economicspb.Deposit
	if err := k.cdc.Unmarshal(bz, &deposit); err != nil {
		return nil, err
	}
	return &deposit, nil
}

// IterateDeposits iterates over all deposits for a proposal
func (k Keeper) IterateDeposits(ctx context.Context, proposalID uint64, cb func(deposit *economicspb.Deposit) bool) error {
	store := k.storeService.OpenKVStore(ctx)

	// Create prefix for this proposal's deposits
	proposalPrefix := make([]byte, 9)
	copy(proposalPrefix, types.DepositPrefix)
	binary.BigEndian.PutUint64(proposalPrefix[1:], proposalID)

	iterator, err := store.Iterator(proposalPrefix, storeprefixend(proposalPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit economicspb.Deposit
		if err := k.cdc.Unmarshal(iterator.Value(), &deposit); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&deposit) {
			break
		}
	}
	return nil
}

// ============================
// VOTE DELEGATION OPERATIONS
// ============================

// SetVoteDelegation stores a vote delegation
func (k Keeper) SetVoteDelegation(ctx context.Context, delegation *economicspb.VoteDelegation) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteDelegationKey(delegation.Delegator, delegation.Delegate)
	bz, err := k.cdc.Marshal(delegation)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return store.Set(key, bz)
}

// GetVoteDelegation retrieves a vote delegation
func (k Keeper) GetVoteDelegation(ctx context.Context, delegator, delegate string) (*economicspb.VoteDelegation, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteDelegationKey(delegator, delegate)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrDelegationNotFound
	}

	var delegation economicspb.VoteDelegation
	if err := k.cdc.Unmarshal(bz, &delegation); err != nil {
		return nil, err
	}
	return &delegation, nil
}

// DeleteVoteDelegation removes a vote delegation
func (k Keeper) DeleteVoteDelegation(ctx context.Context, delegator, delegate string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteDelegationKey(delegator, delegate)
	return store.Delete(key)
}

// ============================
// VETO REQUEST OPERATIONS
// ============================

// SetVetoRequest stores a veto request (deprecated - kept for backward compatibility)
func (k Keeper) SetVetoRequest(ctx context.Context, veto interface{}) error {
	// Veto requests are no longer stored separately in the proto schema
	// They are part of the proposal voting process
	return nil
}

// GetVetoRequest retrieves a veto request (deprecated - kept for backward compatibility)
func (k Keeper) GetVetoRequest(ctx context.Context, proposalID uint64) (interface{}, error) {
	// Veto requests are no longer stored separately in the proto schema
	return nil, nil
}

// ============================
// SNAPSHOT VOTE OPERATIONS
// ============================

// SetSnapshotVote stores a snapshot vote (deprecated - now part of Vote)
func (k Keeper) SetSnapshotVote(ctx context.Context, vote interface{}) error {
	// Snapshot voting is now handled through the regular Vote type
	return nil
}

// GetSnapshotVote retrieves a snapshot vote (deprecated - now part of Vote)
func (k Keeper) GetSnapshotVote(ctx context.Context, proposalID uint64, voter string) (interface{}, error) {
	// Snapshot voting is now handled through the regular Vote type
	return nil, nil
}

// ============================
// VOTE COMMITMENT OPERATIONS (Secret Ballot)
// ============================

// SetVoteCommitment stores a vote commitment (deprecated - now part of Vote)
func (k Keeper) SetVoteCommitment(ctx context.Context, commitment interface{}) error {
	// Vote commitments are now stored as part of the Vote type
	return nil
}

// GetVoteCommitment retrieves a vote commitment (deprecated - now part of Vote)
func (k Keeper) GetVoteCommitment(ctx context.Context, proposalID uint64, voter string) (interface{}, error) {
	// Vote commitments are now stored as part of the Vote type
	return nil, nil
}

// ============================
// TOKEN LOCK OPERATIONS
// ============================

// SetTokenLock stores a token lock (deprecated - use VoteLock)
func (k Keeper) SetTokenLock(ctx context.Context, lock interface{}) error {
	// Token locks are now handled through VoteLock
	return nil
}

// GetTokenLock retrieves a token lock (deprecated - use VoteLock)
func (k Keeper) GetTokenLock(ctx context.Context, owner, lockID string) (interface{}, error) {
	// Token locks are now handled through VoteLock
	return nil, nil
}

// ============================
// TALLY OPERATIONS
// ============================

// CalculateTally calculates the tally for a proposal
func (k Keeper) CalculateTally(ctx context.Context, proposalID uint64) (*economicspb.TallyResult, error) {
	yesVotes := big.NewInt(0)
	noVotes := big.NewInt(0)
	abstainVotes := big.NewInt(0)
	vetoVotes := big.NewInt(0)

	err := k.IterateVotes(ctx, proposalID, func(vote *economicspb.Vote) bool {
		// VotingPower is math.Int, convert to big.Int
		weight := vote.VotingPower.BigInt()

		switch vote.Option {
		case economicspb.VoteOption_VOTE_OPTION_YES:
			yesVotes.Add(yesVotes, weight)
		case economicspb.VoteOption_VOTE_OPTION_NO:
			noVotes.Add(noVotes, weight)
		case economicspb.VoteOption_VOTE_OPTION_ABSTAIN:
			abstainVotes.Add(abstainVotes, weight)
		case economicspb.VoteOption_VOTE_OPTION_NO_WITH_VETO:
			vetoVotes.Add(vetoVotes, weight)
		}
		return false
	})

	if err != nil {
		return nil, err
	}

	totalVotes := new(big.Int)
	totalVotes.Add(totalVotes, yesVotes)
	totalVotes.Add(totalVotes, noVotes)
	totalVotes.Add(totalVotes, abstainVotes)
	totalVotes.Add(totalVotes, vetoVotes)

	return &economicspb.TallyResult{
		YesCount:         math.NewIntFromBigInt(yesVotes),
		NoCount:          math.NewIntFromBigInt(noVotes),
		AbstainCount:     math.NewIntFromBigInt(abstainVotes),
		NoWithVetoCount:  math.NewIntFromBigInt(vetoVotes),
		TotalVotingPower: math.NewIntFromBigInt(totalVotes),
	}, nil
}

// UpdateProposalStatus updates a proposal's status based on votes and time
func (k Keeper) UpdateProposalStatus(ctx context.Context, proposalID uint64) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Note: Time-based checks would need to use proposal timestamp fields
	// This is a simplified version that would need proper time handling

	// Calculate tally
	tally, err := k.CalculateTally(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to CalculateTally: %w", err)
	}

	// Update proposal with tally
	proposal.FinalTallyResult = tally

	// Determine if proposal passed based on tally
	totalVotes := tally.TotalVotingPower.BigInt()

	if totalVotes.Sign() > 0 {
		// Simple majority check (would need proper quorum/threshold logic from params)
		proposal.Status = economicspb.ProposalStatus_PROPOSAL_STATUS_PASSED
	} else {
		proposal.Status = economicspb.ProposalStatus_PROPOSAL_STATUS_REJECTED
	}

	return k.SetProposal(ctx, proposal)
}

// ============================
// TREASURY OPERATIONS
// ============================

// SetTreasuryMultisig stores the treasury multisig configuration (deprecated - now in params)
func (k Keeper) SetTreasuryMultisig(ctx context.Context, multisig interface{}) error {
	// Treasury multisig is now part of the module params
	return nil
}

// GetTreasuryMultisig retrieves the treasury multisig configuration (deprecated - now in params)
func (k Keeper) GetTreasuryMultisig(ctx context.Context) (interface{}, error) {
	// Treasury multisig is now part of the module params
	return nil, nil
}

// SetPendingTreasuryTx stores a pending treasury transaction
func (k Keeper) SetPendingTreasuryTx(ctx context.Context, tx *economicspb.PendingTreasuryTx) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetPendingTreasuryTxKey(tx.TxId)
	bz, err := k.cdc.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal for TxId: %w", err)
	}
	return store.Set(key, bz)
}

// GetPendingTreasuryTx retrieves a pending treasury transaction
func (k Keeper) GetPendingTreasuryTx(ctx context.Context, txID string) (*economicspb.PendingTreasuryTx, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetPendingTreasuryTxKey(txID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrTxNotFound
	}

	var tx economicspb.PendingTreasuryTx
	if err := k.cdc.Unmarshal(bz, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

// DeletePendingTreasuryTx removes a pending treasury transaction
func (k Keeper) DeletePendingTreasuryTx(ctx context.Context, txID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetPendingTreasuryTxKey(txID)
	return store.Delete(key)
}

// IteratePendingTreasuryTxs iterates over all pending treasury transactions
func (k Keeper) IteratePendingTreasuryTxs(ctx context.Context, cb func(tx *economicspb.PendingTreasuryTx) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.PendingTreasuryTxPrefix, storeprefixend(types.PendingTreasuryTxPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tx economicspb.PendingTreasuryTx
		if err := k.cdc.Unmarshal(iterator.Value(), &tx); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&tx) {
			break
		}
	}
	return nil
}

// SetTreasuryBalance stores the treasury balance
func (k Keeper) SetTreasuryBalance(ctx context.Context, balance string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.TreasuryBalanceKey, []byte(balance))
}

// GetTreasuryBalance retrieves the treasury balance
func (k Keeper) GetTreasuryBalance(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.TreasuryBalanceKey)
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}
	return string(bz), nil
}

// ============================
// ECONOMIC MONITORING OPERATIONS
// ============================

// SetInflationAlert stores an inflation alert (deprecated - monitoring moved to separate module)
func (k Keeper) SetInflationAlert(ctx context.Context, alert interface{}) error {
	// Inflation alerts are now handled by the monitoring module
	return nil
}

// GetInflationAlert retrieves an inflation alert (deprecated - monitoring moved to separate module)
func (k Keeper) GetInflationAlert(ctx context.Context, alertID string) (interface{}, error) {
	// Inflation alerts are now handled by the monitoring module
	return nil, nil
}

// IterateInflationAlerts iterates over all inflation alerts (deprecated - monitoring moved to separate module)
func (k Keeper) IterateInflationAlerts(ctx context.Context, cb func(alert interface{}) bool) error {
	// Inflation alerts are now handled by the monitoring module
	return nil
}

// SetLargeTxRecord stores a large transaction record (deprecated - monitoring moved to separate module)
func (k Keeper) SetLargeTxRecord(ctx context.Context, record interface{}) error {
	// Large tx records are now handled by the monitoring module
	return nil
}

// SetLastLargeTxTime stores the last large transaction time for an address
func (k Keeper) SetLastLargeTxTime(ctx context.Context, address string, timestamp int64) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetLastLargeTxTimeKey(address)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(timestamp))
	return store.Set(key, bz)
}

// SetAddressHolding stores the holding amount for an address
func (k Keeper) SetAddressHolding(ctx context.Context, address string, amount string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAddressHoldingKey(address)
	return store.Set(key, []byte(amount))
}

// SetUserMEVBalance stores the MEV balance for a user
func (k Keeper) SetUserMEVBalance(ctx context.Context, address string, balance string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserMEVBalanceKey(address)
	return store.Set(key, []byte(balance))
}

// GetUserMEVBalance retrieves the MEV balance for a user
func (k Keeper) GetUserMEVBalance(ctx context.Context, address string) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserMEVBalanceKey(address)
	bz, err := store.Get(key)
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}
	return string(bz), nil
}

// SetTotalMEVPending stores the total pending MEV amount
func (k Keeper) SetTotalMEVPending(ctx context.Context, amount string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.TotalMEVPendingKey, []byte(amount))
}

// GetTotalMEVPending retrieves the total pending MEV amount
func (k Keeper) GetTotalMEVPending(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.TotalMEVPendingKey)
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}
	return string(bz), nil
}

// SetTotalBurned stores the total burned amount
func (k Keeper) SetTotalBurned(ctx context.Context, amount string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.TotalBurnedKey, []byte(amount))
}

// GetTotalBurned retrieves the total burned amount
func (k Keeper) GetTotalBurned(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.TotalBurnedKey)
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}
	return string(bz), nil
}

// SetPreviousInflation stores the previous inflation rate
func (k Keeper) SetPreviousInflation(ctx context.Context, rate uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, rate)
	return store.Set(types.PreviousInflationKey, bz)
}

// GetPreviousInflation retrieves the previous inflation rate
func (k Keeper) GetPreviousInflation(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.PreviousInflationKey)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil
	}
	return binary.BigEndian.Uint64(bz), nil
}
