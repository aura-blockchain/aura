package keeper_test

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

// Test Next Proposal ID

func TestGetNextProposalID(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockStaking := &MockStakingKeeperForExtended{delegatorBonded: make(map[string]sdkmath.Int)}
	mockBank := &MockBankKeeperForExtended{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank)

	// First ID should be 1
	id := k.GetNextProposalID(input.Ctx)
	assert.Equal(t, uint64(1), id)
}

// createTestKeeper is a helper function to create a keeper with mock staking keeper
func createTestKeeper(input keepertest.TestInput) *keeper.Keeper {
	mockStaking := &MockStakingKeeperForExtended{delegatorBonded: make(map[string]sdkmath.Int)}
	mockBank := &MockBankKeeperForExtended{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	return keeper.NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank)
}

func TestSetNextProposalID(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	k.SetNextProposalID(input.Ctx, 42)
	id := k.GetNextProposalID(input.Ctx)
	assert.Equal(t, uint64(42), id)
}

func TestNextProposalIDIncrement(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	// Get initial ID
	id1 := k.GetNextProposalID(input.Ctx)

	// Set next ID
	k.SetNextProposalID(input.Ctx, id1+1)

	// Get next ID
	id2 := k.GetNextProposalID(input.Ctx)
	assert.Equal(t, id1+1, id2)
}

// Test Proposal CRUD Operations

func TestSetAndGetProposal(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "This is a test",
		Proposer:    keepertest.GenTestAddr().String(),
		Status:      types.StatusDepositPeriod,
		Category:    types.CategoryText,
	}

	err := k.SetProposal(input.Ctx, proposal)
	require.NoError(t, err)

	retrieved, err := k.GetProposal(input.Ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, proposal.Title, retrieved.Title)
	assert.Equal(t, proposal.Description, retrieved.Description)
	assert.Equal(t, proposal.Status, retrieved.Status)
}

func TestDeleteProposal(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	proposal := &types.Proposal{
		Id:       1,
		Title:    "Test",
		Proposer: keepertest.GenTestAddr().String(),
		Status:   types.StatusDepositPeriod,
	}

	err := k.SetProposal(input.Ctx, proposal)
	require.NoError(t, err)

	// Verify it exists
	_, err = k.GetProposal(input.Ctx, 1)
	require.NoError(t, err)

	// Delete it
	k.DeleteProposal(input.Ctx, 1)

	// Verify it's gone
	_, err = k.GetProposal(input.Ctx, 1)
	require.Error(t, err)
	assert.Equal(t, types.ErrProposalNotFound, err)
}

func TestGetAllProposalsEmpty(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	proposals := k.GetAllProposals(input.Ctx)
	assert.Empty(t, proposals)
}

func TestGetAllProposalsMultiple(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	// Create multiple proposals
	for i := 1; i <= 5; i++ {
		proposal := &types.Proposal{
			Id:       uint64(i),
			Title:    "Proposal " + string(rune('0'+i)),
			Proposer: keepertest.GenTestAddr().String(),
			Status:   types.StatusDepositPeriod,
		}
		err := k.SetProposal(input.Ctx, proposal)
		require.NoError(t, err)
	}

	proposals := k.GetAllProposals(input.Ctx)
	assert.Len(t, proposals, 5)
}

// Test Vote Operations

func TestSetAndGetVote(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	voter := keepertest.GenTestAddr().String()
	vote := &types.Vote{
		ProposalId:  1,
		Voter:       voter,
		Option:      types.OptionYes,
		VotingPower: "1000000",
	}

	err := k.SetVote(input.Ctx, vote)
	require.NoError(t, err)

	retrieved, err := k.GetVote(input.Ctx, 1, voter)
	require.NoError(t, err)
	assert.Equal(t, vote.Option, retrieved.Option)
	assert.Equal(t, vote.VotingPower, retrieved.VotingPower)
}

func TestGetVoteNotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	_, err := k.GetVote(input.Ctx, 1, "nonexistent")
	require.Error(t, err)
}

func TestGetVotesForProposal(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	proposalID := uint64(1)

	// Create multiple votes with unique addresses
	voters := keepertest.GenTestAddrs(3)
	for _, voter := range voters {
		vote := &types.Vote{
			ProposalId:  proposalID,
			Voter:       voter.String(),
			Option:      types.OptionYes,
			VotingPower: "1000000",
		}
		err := k.SetVote(input.Ctx, vote)
		require.NoError(t, err)
	}

	votes := k.GetVotes(input.Ctx, proposalID)
	assert.Len(t, votes, 3)
}

func TestGetVotesEmpty(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	votes := k.GetVotes(input.Ctx, 999)
	assert.Empty(t, votes)
}

// Test Deposit Operations

func TestSetAndGetDeposit(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	depositor := keepertest.GenTestAddr().String()
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  depositor,
		Amount:     "1000000",
	}

	err := k.SetDeposit(input.Ctx, deposit)
	require.NoError(t, err)

	retrieved, err := k.GetDeposit(input.Ctx, 1, depositor)
	require.NoError(t, err)
	assert.Equal(t, deposit.Amount, retrieved.Amount)
}

func TestGetDepositNotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	_, err := k.GetDeposit(input.Ctx, 1, "nonexistent")
	require.Error(t, err)
	assert.Equal(t, types.ErrInvalidDeposit, err)
}

func TestGetDepositsForProposal(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	proposalID := uint64(1)

	// Create multiple deposits
	for i := 0; i < 3; i++ {
		depositor := keepertest.GenTestAddr().String()
		deposit := &types.Deposit{
			ProposalId: proposalID,
			Depositor:  depositor,
			Amount:     "1000000",
		}
		err := k.SetDeposit(input.Ctx, deposit)
		require.NoError(t, err)
	}

	deposits := k.GetDeposits(input.Ctx, proposalID)
	assert.Len(t, deposits, 3)
}

// Test Vote Delegation

func TestSetVoteDelegation(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	delegator := keepertest.GenTestAddr().String()
	delegate := keepertest.GenTestAddr().String()

	delegation := &types.VoteDelegation{
		Delegator:      delegator,
		Delegate:       delegate,
		DelegatedPower: "1000000",
	}

	err := k.SetVoteDelegation(input.Ctx, delegation)
	require.NoError(t, err)
}

func TestDeleteVoteDelegation(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	delegator := keepertest.GenTestAddr().String()
	delegate := keepertest.GenTestAddr().String()

	delegation := &types.VoteDelegation{
		Delegator:      delegator,
		Delegate:       delegate,
		DelegatedPower: "1000000",
	}

	err := k.SetVoteDelegation(input.Ctx, delegation)
	require.NoError(t, err)

	err = k.DeleteVoteDelegation(input.Ctx, delegator, delegate)
	require.NoError(t, err)
}

func TestGetVoteDelegations(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	delegator := keepertest.GenTestAddr().String()

	// Create multiple delegations
	for i := 0; i < 3; i++ {
		delegate := keepertest.GenTestAddr().String()
		delegation := &types.VoteDelegation{
			Delegator:      delegator,
			Delegate:       delegate,
			DelegatedPower: "1000000",
		}
		err := k.SetVoteDelegation(input.Ctx, delegation)
		require.NoError(t, err)
	}

	delegations := k.GetVoteDelegations(input.Ctx, delegator)
	assert.Len(t, delegations, 3)
}

// Test Veto Requests

func TestSetAndGetVetoRequest(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	vetoer := keepertest.GenTestAddr().String()
	veto := &types.VetoRequest{
		ProposalId: 1,
		Vetoer:     vetoer,
		Reason:     "Security concerns",
		Cosigners:  []string{vetoer},
	}

	err := k.SetVetoRequest(input.Ctx, veto)
	require.NoError(t, err)

	retrieved, err := k.GetVetoRequest(input.Ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, veto.Reason, retrieved.Reason)
	assert.Equal(t, veto.Vetoer, retrieved.Vetoer)
}

func TestGetVetoRequestNotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	_, err := k.GetVetoRequest(input.Ctx, 999)
	require.Error(t, err)
	assert.Equal(t, types.ErrInvalidVeto, err)
}

func TestGetVetoRequests(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	proposalID := uint64(1)
	vetoer := keepertest.GenTestAddr().String()
	veto := &types.VetoRequest{
		ProposalId: proposalID,
		Vetoer:     vetoer,
		Reason:     "Test veto",
		Cosigners:  []string{vetoer},
	}

	err := k.SetVetoRequest(input.Ctx, veto)
	require.NoError(t, err)

	vetos := k.GetVetoRequests(input.Ctx, proposalID)
	assert.NotEmpty(t, vetos)
}

// Test Snapshot Votes

func TestSetSnapshotVote(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	voter := keepertest.GenTestAddr().String()
	vote := &types.SnapshotVote{
		ProposalId:            1,
		Voter:                 voter,
		Option:                types.OptionYes,
		VotingPowerAtSnapshot: "1000000",
		Signature:             "signature_data",
	}

	err := k.SetSnapshotVote(input.Ctx, vote)
	require.NoError(t, err)
}

func TestGetSnapshotVotes(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	proposalID := uint64(1)

	// Create multiple snapshot votes
	for i := 0; i < 3; i++ {
		voter := keepertest.GenTestAddr().String()
		vote := &types.SnapshotVote{
			ProposalId:            proposalID,
			Voter:                 voter,
			Option:                types.OptionYes,
			VotingPowerAtSnapshot: "1000000",
			Signature:             "sig",
		}
		err := k.SetSnapshotVote(input.Ctx, vote)
		require.NoError(t, err)
	}

	votes := k.GetSnapshotVotes(input.Ctx, proposalID)
	assert.Len(t, votes, 3)
}

// Test Tally Calculation

func TestCalculateTallyNoVotes(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	tally := k.CalculateTally(input.Ctx, 1)
	require.NotNil(t, tally)
	assert.Equal(t, "0", tally.Yes)
	assert.Equal(t, "0", tally.No)
	assert.Equal(t, "0", tally.Abstain)
	assert.Equal(t, "0", tally.NoWithVeto)
}

func TestCalculateTallyWithVotes(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockStaking := &MockStakingKeeperForExtended{delegatorBonded: make(map[string]sdkmath.Int)}
	mockBank := &MockBankKeeperForExtended{balances: make(map[string]sdk.Coins), moduleBalances: make(map[string]sdk.Coins)}
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank)

	proposalID := uint64(1)

	// Add different types of votes with staking power
	votesWithPower := []struct {
		option types.VoteOption
		power  sdkmath.Int
	}{
		{types.OptionYes, sdkmath.NewInt(1000000)},
		{types.OptionYes, sdkmath.NewInt(2000000)},
		{types.OptionNo, sdkmath.NewInt(500000)},
		{types.OptionAbstain, sdkmath.NewInt(300000)},
		{types.OptionNoWithVeto, sdkmath.NewInt(100000)},
	}

	// Generate unique addresses for each voter
	voters := keepertest.GenTestAddrs(len(votesWithPower))

	for i, voteInfo := range votesWithPower {
		voter := voters[i]

		// Set up staking power for this voter in the mock
		mockStaking.delegatorBonded[voter.String()] = voteInfo.power

		vote := &types.Vote{
			ProposalId:  proposalID,
			Voter:       voter.String(),
			Option:      voteInfo.option,
			VotingPower: voteInfo.power.String(),
		}
		err := k.SetVote(input.Ctx, vote)
		require.NoError(t, err, "vote %d failed", i)
	}

	tally := k.CalculateTally(input.Ctx, proposalID)
	require.NotNil(t, tally)

	// Verify proper accumulation (not just "1" per vote)
	assert.Equal(t, "3000000", tally.Yes, "Yes votes should total 1M + 2M")
	assert.Equal(t, "500000", tally.No, "No votes should total 500K")
	assert.Equal(t, "300000", tally.Abstain, "Abstain votes should total 300K")
	assert.Equal(t, "100000", tally.NoWithVeto, "NoWithVeto votes should total 100K")
}

// Test Helper Methods

func TestGetVotingPower(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	addr := keepertest.GenTestAddr().String()
	power, err := k.GetVotingPower(input.Ctx, addr)
	require.NoError(t, err)
	assert.NotNil(t, power)
	// Default voting power is 0 since no stake is set in mock
	assert.Equal(t, sdkmath.ZeroInt(), power)
}

func TestGetDelegatedPower(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	addr := keepertest.GenTestAddr().String()
	power := k.GetDelegatedPower(input.Ctx, addr)
	assert.NotNil(t, power)
}

func TestGetTokenLocks(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	addr := keepertest.GenTestAddr().String()
	locks := k.GetTokenLocks(input.Ctx, addr)
	assert.NotNil(t, locks)
}

// Test Edge Cases

func TestSetProposalWithZeroID(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	proposal := &types.Proposal{
		Id:       0,
		Title:    "Zero ID",
		Proposer: keepertest.GenTestAddr().String(),
		Status:   types.StatusDepositPeriod,
	}

	err := k.SetProposal(input.Ctx, proposal)
	require.NoError(t, err)

	retrieved, err := k.GetProposal(input.Ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, "Zero ID", retrieved.Title)
}

func TestSetVoteWithEmptyVoter(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	vote := &types.Vote{
		ProposalId:  1,
		Voter:       "",
		Option:      types.OptionYes,
		VotingPower: "1000000",
	}

	err := k.SetVote(input.Ctx, vote)
	require.NoError(t, err)
}

func TestMultipleProposalCategories(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	categories := []types.ProposalCategory{
		types.CategoryText,
		types.CategoryParameterChange,
		types.CategorySoftwareUpgrade,
		types.CategorySpending,
		types.CategoryEmergency,
		types.CategoryConstitution,
	}

	for i, category := range categories {
		proposal := &types.Proposal{
			Id:       uint64(i + 1),
			Title:    "Proposal",
			Proposer: keepertest.GenTestAddr().String(),
			Status:   types.StatusDepositPeriod,
			Category: category,
		}

		err := k.SetProposal(input.Ctx, proposal)
		require.NoError(t, err)

		retrieved, err := k.GetProposal(input.Ctx, uint64(i+1))
		require.NoError(t, err)
		assert.Equal(t, category, retrieved.Category)
	}
}

func TestProposalStatusTransitions(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	statuses := []types.ProposalStatus{
		types.StatusDepositPeriod,
		types.StatusVotingPeriod,
		types.StatusPassed,
		types.StatusRejected,
		types.StatusFailed,
		types.StatusVetoed,
		types.StatusExecutionDelay,
		types.StatusReadyForExecution,
		types.StatusExecuted,
	}

	for i, status := range statuses {
		proposal := &types.Proposal{
			Id:       uint64(i + 1),
			Title:    "Proposal",
			Proposer: keepertest.GenTestAddr().String(),
			Status:   status,
		}

		err := k.SetProposal(input.Ctx, proposal)
		require.NoError(t, err)

		retrieved, err := k.GetProposal(input.Ctx, uint64(i+1))
		require.NoError(t, err)
		assert.Equal(t, status, retrieved.Status)
	}
}

func TestVoteOptions(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	options := []types.VoteOption{
		types.OptionUnspecified,
		types.OptionYes,
		types.OptionAbstain,
		types.OptionNo,
		types.OptionNoWithVeto,
	}

	for i, option := range options {
		voter := keepertest.GenTestAddr().String()
		vote := &types.Vote{
			ProposalId:  1,
			Voter:       voter + string(rune(i)),
			Option:      option,
			VotingPower: "1000000",
		}

		err := k.SetVote(input.Ctx, vote)
		require.NoError(t, err)
	}

	votes := k.GetVotes(input.Ctx, 1)
	assert.Len(t, votes, len(options))
}

// Test Concurrent Operations

func TestConcurrentProposalCreation(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	// Create multiple proposals with different IDs
	for i := 1; i <= 10; i++ {
		proposal := &types.Proposal{
			Id:       uint64(i),
			Title:    "Concurrent Proposal",
			Proposer: keepertest.GenTestAddr().String(),
			Status:   types.StatusDepositPeriod,
		}

		err := k.SetProposal(input.Ctx, proposal)
		require.NoError(t, err)
	}

	proposals := k.GetAllProposals(input.Ctx)
	assert.Len(t, proposals, 10)
}

func TestBoundaryValues(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := createTestKeeper(input)

	// Test with max uint64 ID
	proposal := &types.Proposal{
		Id:       ^uint64(0), // Max uint64
		Title:    "Max ID",
		Proposer: keepertest.GenTestAddr().String(),
		Status:   types.StatusDepositPeriod,
	}

	err := k.SetProposal(input.Ctx, proposal)
	require.NoError(t, err)

	retrieved, err := k.GetProposal(input.Ctx, ^uint64(0))
	require.NoError(t, err)
	assert.Equal(t, "Max ID", retrieved.Title)
}

// MockStakingKeeperForExtended for tests in keeper_extended_test
type MockStakingKeeperForExtended struct {
	delegatorBonded map[string]sdkmath.Int
}

func (m *MockStakingKeeperForExtended) GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (sdkmath.Int, error) {
	if amount, ok := m.delegatorBonded[delegator.String()]; ok {
		return amount, nil
	}
	return sdkmath.ZeroInt(), nil
}

func (m *MockStakingKeeperForExtended) TotalBondedTokens(ctx context.Context) (sdkmath.Int, error) {
	total := sdkmath.ZeroInt()
	for _, amount := range m.delegatorBonded {
		total = total.Add(amount)
	}
	return total, nil
}

// MockBankKeeperForExtended for tests
type MockBankKeeperForExtended struct {
	balances       map[string]sdk.Coins
	moduleBalances map[string]sdk.Coins
}

func (m *MockBankKeeperForExtended) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeperForExtended) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeperForExtended) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}
