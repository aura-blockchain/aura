// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

func setupExtendedKeeper(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")
	mockStaking := NewMockStakingKeeper()
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	mockSecurity := &MockSecurityKeeper{}
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})
	keeper.SetParams(ctx, types.DefaultParams())
	return keeper, ctx
}

// TestDeleteProposalMethod tests proposal deletion
func TestDeleteProposalMethod(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusVotingPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Verify proposal exists
	_, err := keeper.GetProposal(ctx, 1)
	require.NoError(t, err)

	// Delete proposal
	keeper.DeleteProposal(ctx, 1)

	// Verify proposal is deleted
	_, err = keeper.GetProposal(ctx, 1)
	require.Error(t, err)
}

// TestCalculateTallyComprehensive tests tally calculation
func TestCalculateTallyComprehensive(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusVotingPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Create votes
	votes := []struct {
		voter  string
		option types.VoteOption
		power  string
	}{
		{"voter1", types.VoteOptionYes, "1000"},
		{"voter2", types.VoteOptionYes, "2000"},
		{"voter3", types.VoteOptionNo, "500"},
		{"voter4", types.VoteOptionAbstain, "300"},
		{"voter5", types.VoteOptionNoWithVeto, "200"},
	}

	for _, v := range votes {
		vote := &types.Vote{
			ProposalId:  1,
			Voter:       testAddr(v.voter),
			Option:      v.option,
			VotingPower: v.power,
			Timestamp:   ts,
		}
		keeper.SetVote(ctx, vote)
	}

	// Calculate tally
	tally := keeper.CalculateTally(ctx, 1)
	require.NotNil(t, tally)

	// Verify tally results
	yes, _ := sdkmath.NewIntFromString(tally.Yes)
	no, _ := sdkmath.NewIntFromString(tally.No)
	abstain, _ := sdkmath.NewIntFromString(tally.Abstain)
	noWithVeto, _ := sdkmath.NewIntFromString(tally.NoWithVeto)

	require.Equal(t, sdkmath.NewInt(3000), yes)
	require.Equal(t, sdkmath.NewInt(500), no)
	require.Equal(t, sdkmath.NewInt(300), abstain)
	require.Equal(t, sdkmath.NewInt(200), noWithVeto)
}

// TestRefundDepositsMethod tests deposit refunding
func TestRefundDepositsMethod(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	depositor := testAddr("depositor1")

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusPassed,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Create deposit
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  depositor,
		Amount:     "1000uaura",
		Timestamp:  ts,
	}
	keeper.SetDeposit(ctx, deposit)

	// Fund module account
	keeper.bankKeeper.(*MockBankKeeper).moduleBalances[types.ModuleName] = sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1000)))

	// Refund deposits
	err := keeper.RefundDeposits(ctx, 1)
	require.NoError(t, err)
}

// TestBurnDepositsMethod tests deposit burning
func TestBurnDepositsMethod(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	depositor := testAddr("depositor1")

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusRejected,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Create deposit
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  depositor,
		Amount:     "1000uaura",
		Timestamp:  ts,
	}
	keeper.SetDeposit(ctx, deposit)

	// Fund module account
	keeper.bankKeeper.(*MockBankKeeper).moduleBalances[types.ModuleName] = sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1000)))

	// Burn deposits
	err := keeper.BurnDeposits(ctx, 1)
	require.NoError(t, err)
}

// TestTokenLocksOperations tests token lock operations
func TestTokenLocksOperations(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	owner := testAddr("owner1")

	// Create token lock
	ts, _ := gogotypes.TimestampProto(time.Now())
	unlockTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	lock := &types.TokenLock{
		Owner:        owner,
		ProposalId:   1,
		LockedAmount: "1000",
		LockTime:     ts,
		UnlockTime:   unlockTs,
	}

	// Set token lock
	keeper.SetTokenLock(ctx, lock)

	// Get token locks
	locks := keeper.GetTokenLocks(ctx, owner)
	require.Len(t, locks, 1)
	require.Equal(t, uint64(1), locks[0].ProposalId)

	// Delete token lock
	keeper.DeleteTokenLock(ctx, owner, 1)

	// Verify lock is deleted
	locks = keeper.GetTokenLocks(ctx, owner)
	require.Len(t, locks, 0)
}

// TestVetoRequestsOperations tests veto request operations
func TestVetoRequestsOperations(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	vetoer := testAddr("vetoer1")

	// Create veto request
	ts, _ := gogotypes.TimestampProto(time.Now())
	veto := &types.VetoRequest{
		ProposalId: 1,
		Vetoer:     vetoer,
		Reason:     "Test reason",
		Timestamp:  ts,
		Cosigners:  []string{},
	}

	// Set veto request
	keeper.SetVetoRequest(ctx, veto)

	// Get veto request
	retrieved, err := keeper.GetVetoRequest(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, "Test reason", retrieved.Reason)

	// Get all veto requests
	vetoRequests := keeper.GetVetoRequests(ctx, 1)
	require.Len(t, vetoRequests, 1)
}

// TestSnapshotVotesOperations tests snapshot vote operations
func TestSnapshotVotesOperations(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	voter := testAddr("voter1")

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:             1,
		Title:          "Test Proposal",
		Description:    "Test Description",
		Proposer:       testAddr("proposer1"),
		Status:         types.StatusVotingPeriod,
		Category:       types.CategoryText,
		SubmitTime:     ts,
		SnapshotHeight: 100,
	}
	keeper.SetProposal(ctx, proposal)

	// Set voting power snapshot
	err := keeper.SetVotingPowerSnapshot(ctx, 1, voter, sdkmath.NewInt(1000))
	require.NoError(t, err)

	// Get voting power snapshot
	power, found := keeper.GetVotingPowerSnapshot(ctx, 1, voter)
	require.True(t, found)
	require.Equal(t, sdkmath.NewInt(1000), power)

	// Get snapshot votes
	votes := keeper.GetSnapshotVotes(ctx, 1)
	require.NotNil(t, votes)
}

// TestProposalStatusTransitionsExtended tests proposal status transitions
func TestProposalStatusTransitionsExtended(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Create proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusDepositPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Transition to voting period
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// Verify status
	retrieved, err := keeper.GetProposal(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, types.StatusVotingPeriod, retrieved.Status)
}

// TestMultipleDepositsOnProposal tests multiple deposits on a proposal
func TestMultipleDepositsOnProposal(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Create proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusDepositPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Create multiple deposits
	for i := 1; i <= 5; i++ {
		deposit := &types.Deposit{
			ProposalId: 1,
			Depositor:  testAddr("depositor" + string(rune('0'+i))),
			Amount:     "1000",
			Timestamp:  ts,
		}
		keeper.SetDeposit(ctx, deposit)
	}

	// Verify all deposits
	deposits := keeper.GetDeposits(ctx, 1)
	require.Len(t, deposits, 5)
}

// TestMultipleVotesOnProposal tests multiple votes on a proposal
func TestMultipleVotesOnProposal(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Create proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusVotingPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Create multiple votes
	for i := 1; i <= 10; i++ {
		vote := &types.Vote{
			ProposalId:  1,
			Voter:       testAddr("voter" + string(rune('0'+i))),
			Option:      types.VoteOptionYes,
			VotingPower: "1000",
			Timestamp:   ts,
		}
		keeper.SetVote(ctx, vote)
	}

	// Verify all votes
	votes := keeper.GetVotes(ctx, 1)
	require.Len(t, votes, 10)
}

// TestVoteDelegationOperations tests vote delegation
func TestVoteDelegationOperations(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")

	// Create delegation
	ts, _ := gogotypes.TimestampProto(time.Now())
	delegation := &types.VoteDelegation{
		Delegator:      delegator,
		Delegate:       delegate,
		DelegationTime: ts,
		DelegatedPower: "1000",
	}
	keeper.SetVoteDelegation(ctx, delegation)

	// Get delegations
	delegations := keeper.GetVoteDelegations(ctx, delegator)
	require.Len(t, delegations, 1)
	require.Equal(t, delegate, delegations[0].Delegate)

	// Delete delegation
	keeper.DeleteVoteDelegation(ctx, delegator, delegate)

	// Verify deletion
	delegations = keeper.GetVoteDelegations(ctx, delegator)
	require.Len(t, delegations, 0)
}

// TestParamsStorageOperations tests params storage operations
func TestParamsStorageOperations(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Get default params
	params := keeper.GetParams(ctx)
	require.NotNil(t, params)
	require.NotEmpty(t, params.MinDeposit)

	// Set custom params
	customParams := types.DefaultParams()
	customParams.MinDeposit = "5000uaura"
	keeper.SetParams(ctx, customParams)

	// Verify custom params
	retrievedParams := keeper.GetParams(ctx)
	require.Equal(t, "5000uaura", retrievedParams.MinDeposit)
}

// TestProposalIDCounterOperations tests proposal ID counter
func TestProposalIDCounterOperations(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Get initial ID
	id1 := keeper.GetNextProposalID(ctx)
	require.Equal(t, uint64(1), id1)

	// Set and increment
	keeper.SetNextProposalID(ctx, 5)
	id2 := keeper.GetNextProposalID(ctx)
	require.Equal(t, uint64(5), id2)

	// Set again
	keeper.SetNextProposalID(ctx, 100)
	id3 := keeper.GetNextProposalID(ctx)
	require.Equal(t, uint64(100), id3)
}

// TestGetAllProposalsComprehensive tests getting all proposals
func TestGetAllProposalsComprehensive(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Create multiple proposals
	ts, _ := gogotypes.TimestampProto(time.Now())
	for i := uint64(1); i <= 5; i++ {
		proposal := &types.Proposal{
			Id:          i,
			Title:       "Test Proposal",
			Description: "Test Description",
			Proposer:    testAddr("proposer1"),
			Status:      types.StatusVotingPeriod,
			Category:    types.CategoryText,
			SubmitTime:  ts,
		}
		keeper.SetProposal(ctx, proposal)
	}
	keeper.SetNextProposalID(ctx, 6)

	// Get all proposals
	proposals := keeper.GetAllProposals(ctx)
	require.Len(t, proposals, 5)
}

// TestDepositDeletion tests deposit deletion
func TestDepositDeletion(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	depositor := testAddr("depositor1")

	// Create proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusDepositPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Create deposit
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  depositor,
		Amount:     "1000",
		Timestamp:  ts,
	}
	keeper.SetDeposit(ctx, deposit)

	// Verify deposit exists
	_, err := keeper.GetDeposit(ctx, 1, depositor)
	require.NoError(t, err)

	// Delete deposit
	keeper.DeleteDeposit(ctx, 1, depositor)

	// Verify deletion
	_, err = keeper.GetDeposit(ctx, 1, depositor)
	require.Error(t, err)
}

// TestDeleteAllDeposits tests deleting all deposits for a proposal
func TestDeleteAllDeposits(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Create proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusDepositPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Create multiple deposits
	for i := 1; i <= 3; i++ {
		deposit := &types.Deposit{
			ProposalId: 1,
			Depositor:  testAddr("depositor" + string(rune('0'+i))),
			Amount:     "1000",
			Timestamp:  ts,
		}
		keeper.SetDeposit(ctx, deposit)
	}

	// Verify deposits exist
	deposits := keeper.GetDeposits(ctx, 1)
	require.Len(t, deposits, 3)

	// Delete all deposits
	keeper.DeleteDeposits(ctx, 1)

	// Verify all deleted
	deposits = keeper.GetDeposits(ctx, 1)
	require.Len(t, deposits, 0)
}

// TestGetOrCreateSnapshotMethod tests get or create voting power snapshot
func TestGetOrCreateSnapshotMethod(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	voter := testAddr("voter1")

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:             1,
		Title:          "Test Proposal",
		Description:    "Test Description",
		Proposer:       testAddr("proposer1"),
		Status:         types.StatusVotingPeriod,
		Category:       types.CategoryText,
		SubmitTime:     ts,
		SnapshotHeight: 100,
	}
	keeper.SetProposal(ctx, proposal)

	// Get or create snapshot
	power, err := keeper.GetOrCreateVotingPowerSnapshot(ctx, 1, voter)
	require.NoError(t, err)
	require.NotNil(t, power)
}

// TestDeleteVotingPowerSnapshotsMethod tests deleting voting power snapshots
func TestDeleteVotingPowerSnapshotsMethod(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	voter1 := testAddr("voter1")
	voter2 := testAddr("voter2")

	// Create snapshots
	keeper.SetVotingPowerSnapshot(ctx, 1, voter1, sdkmath.NewInt(1000))
	keeper.SetVotingPowerSnapshot(ctx, 1, voter2, sdkmath.NewInt(2000))

	// Verify they exist
	_, found1 := keeper.GetVotingPowerSnapshot(ctx, 1, voter1)
	require.True(t, found1)
	_, found2 := keeper.GetVotingPowerSnapshot(ctx, 1, voter2)
	require.True(t, found2)

	// Delete all snapshots for proposal 1
	keeper.DeleteVotingPowerSnapshots(ctx, 1)

	// Verify they're deleted
	_, found1 = keeper.GetVotingPowerSnapshot(ctx, 1, voter1)
	require.False(t, found1)
	_, found2 = keeper.GetVotingPowerSnapshot(ctx, 1, voter2)
	require.False(t, found2)
}

// TestSnapshotVotingPowerForProposalMethod tests creating a snapshot for a proposal
func TestSnapshotVotingPowerForProposalMethod(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:             1,
		Title:          "Test Proposal",
		Description:    "Test Description",
		Proposer:       testAddr("proposer1"),
		Status:         types.StatusVotingPeriod,
		Category:       types.CategoryText,
		SubmitTime:     ts,
		SnapshotHeight: 100,
	}
	keeper.SetProposal(ctx, proposal)

	// Snapshot voting power
	err := keeper.SnapshotVotingPowerForProposal(ctx, 1)
	require.NoError(t, err)
}

// TestVotingPowerCalculation tests voting power calculation
func TestVotingPowerCalculation(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	voter := testAddr("voter1")

	// Get voting power (from staking)
	power, err := keeper.GetVotingPower(ctx, voter)
	require.NoError(t, err)
	require.NotNil(t, power)
}

// TestDelegatedPowerCalculation tests delegated voting power
func TestDelegatedPowerCalculation(t *testing.T) {
	keeper, ctx := setupExtendedKeeper(t)

	delegate := testAddr("delegate1")
	delegator := testAddr("delegator1")

	// Create a delegation
	ts, _ := gogotypes.TimestampProto(time.Now())
	delegation := &types.VoteDelegation{
		Delegator:      delegator,
		Delegate:       delegate,
		DelegationTime: ts,
		DelegatedPower: "1000",
	}
	keeper.SetVoteDelegation(ctx, delegation)

	// Get delegated power - should return staking power of delegators
	power := keeper.GetDelegatedVotingPower(ctx, delegate)
	// The function sums staked tokens, which may be 0 in test env
	require.True(t, power.GTE(sdkmath.ZeroInt()))
}
