package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
	govpb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

// MockStakingKeeperTest for tests
type MockStakingKeeperTest struct {
	delegatorBonded map[string]sdkmath.Int
}

func (m *MockStakingKeeperTest) GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) sdkmath.Int {
	if amount, ok := m.delegatorBonded[delegator.String()]; ok {
		return amount
	}
	return sdkmath.ZeroInt()
}

func (m *MockStakingKeeperTest) TotalBondedTokens(ctx sdk.Context) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, amount := range m.delegatorBonded {
		total = total.Add(amount)
	}
	return total
}

// MockBankKeeperTest for tests (minimal implementation)
type MockBankKeeperTest struct{}

func (m *MockBankKeeperTest) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeperTest) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeperTest) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}

// Helper function to setup keeper for testing
func setupKeeperForTest(t *testing.T) (keepertest.TestInput, *keeper.Keeper) {
	t.Helper()

	input := keepertest.CreateTestInput(t)
	mockStaking := &MockStakingKeeperTest{delegatorBonded: make(map[string]sdkmath.Int)}
	mockBank := &MockBankKeeperTest{}
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank)
	require.NotNil(t, k)

	return input, k
}

// TestDoubleVotePrevention verifies that a voter cannot cast multiple votes that all count
// Each voter should only have ONE vote per proposal in storage
func TestDoubleVotePrevention(t *testing.T) {
	// Setup test environment
	input, k := setupKeeperForTest(t)

	// Create a proposal in voting period
	proposalID := uint64(1)
	proposal := &types.Proposal{
		Id:          proposalID,
		Title:       "Test Proposal",
		Description: "Test double vote prevention",
		Status:      govpb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Proposer:    "cosmos1test",
		SubmitTime:  timestamppb.Now(),
	}
	err := k.SetProposal(input.Ctx, proposal)
	require.NoError(t, err)

	voter := "cosmos1voter"

	// First vote: YES
	vote1 := &types.Vote{
		ProposalId: proposalID,
		Voter:      voter,
		Option:     1, // YES
		Timestamp:  timestamppb.Now(),
	}
	err = k.SetVote(input.Ctx, vote1)
	require.NoError(t, err)

	// Verify first vote is stored
	retrievedVote, err := k.GetVote(input.Ctx, proposalID, voter)
	require.NoError(t, err)
	require.NotNil(t, retrievedVote)
	require.Equal(t, int32(1), retrievedVote.Option) // YES

	// Attempt second vote: NO (should overwrite, not append)
	vote2 := &types.Vote{
		ProposalId: proposalID,
		Voter:      voter,
		Option:     3, // NO
		Timestamp:  timestamppb.Now(),
	}
	err = k.SetVote(input.Ctx, vote2)
	require.NoError(t, err)

	// Verify that only ONE vote exists for this voter
	retrievedVote, err = k.GetVote(input.Ctx, proposalID, voter)
	require.NoError(t, err)
	require.NotNil(t, retrievedVote)
	require.Equal(t, int32(3), retrievedVote.Option) // Should be NO (updated)

	// Verify GetVotes returns only ONE vote for this voter
	allVotes := k.GetVotes(input.Ctx, proposalID)
	require.Equal(t, 1, len(allVotes), "Should only have 1 vote, not multiple")
	require.Equal(t, voter, allVotes[0].Voter)
	require.Equal(t, int32(3), allVotes[0].Option) // Latest vote
}

// TestTallyWithoutDoubleVoting verifies that tally counts each voter only once
func TestTallyWithoutDoubleVoting(t *testing.T) {
	input, k := setupKeeperForTest(t)

	// Create proposal
	proposalID := uint64(1)
	proposal := &types.Proposal{
		Id:          proposalID,
		Title:       "Test Proposal",
		Description: "Test tally without double voting",
		Status:      govpb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Proposer:    "cosmos1test",
		SubmitTime:  timestamppb.Now(),
	}
	err := k.SetProposal(input.Ctx, proposal)
	require.NoError(t, err)

	voter := "cosmos1voter"

	// Cast initial vote: YES
	vote1 := &types.Vote{
		ProposalId: proposalID,
		Voter:      voter,
		Option:     1, // YES
		Timestamp:  timestamppb.Now(),
	}
	err = k.SetVote(input.Ctx, vote1)
	require.NoError(t, err)

	// Update to NO
	vote2 := &types.Vote{
		ProposalId: proposalID,
		Voter:      voter,
		Option:     3, // NO
		Timestamp:  timestamppb.Now(),
	}
	err = k.SetVote(input.Ctx, vote2)
	require.NoError(t, err)

	// Calculate tally
	tally := k.CalculateTally(input.Ctx, proposalID)
	require.NotNil(t, tally)

	// The voter should only be counted ONCE (in NO, not in both YES and NO)
	require.NotNil(t, tally.Yes)
	require.NotNil(t, tally.No)

	// Verify only one vote exists in storage
	allVotes := k.GetVotes(input.Ctx, proposalID)
	require.Equal(t, 1, len(allVotes))
	require.Equal(t, int32(3), allVotes[0].Option) // Should be NO
}

// TestSnapshotVoteNoDuplication verifies snapshot votes cannot be duplicated
func TestSnapshotVoteNoDuplication(t *testing.T) {
	input, k := setupKeeperForTest(t)

	proposalID := uint64(1)
	voter := "cosmos1voter"

	// First snapshot vote
	vote1 := &types.SnapshotVote{
		ProposalId: proposalID,
		Voter:      voter,
		Option:     1, // YES
		Signature:  "sig1",
		Timestamp:  timestamppb.Now(),
	}
	err := k.SetSnapshotVote(input.Ctx, vote1)
	require.NoError(t, err)

	// Update snapshot vote
	vote2 := &types.SnapshotVote{
		ProposalId: proposalID,
		Voter:      voter,
		Option:     3, // NO
		Signature:  "sig2",
		Timestamp:  timestamppb.Now(),
	}
	err = k.SetSnapshotVote(input.Ctx, vote2)
	require.NoError(t, err)

	// Verify only one snapshot vote exists
	allVotes := k.GetSnapshotVotes(input.Ctx, proposalID)
	require.Equal(t, 1, len(allVotes))
	require.Equal(t, voter, allVotes[0].Voter)
	require.Equal(t, int32(3), allVotes[0].Option) // Updated to NO
	require.Equal(t, "sig2", allVotes[0].Signature)
}

// TestMultipleVotersNoInterference verifies that different voters don't interfere
func TestMultipleVotersNoInterference(t *testing.T) {
	input, k := setupKeeperForTest(t)

	proposalID := uint64(1)
	proposal := &types.Proposal{
		Id:          proposalID,
		Title:       "Test Proposal",
		Description: "Test multiple voters",
		Status:      govpb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Proposer:    "cosmos1test",
		SubmitTime:  timestamppb.Now(),
	}
	err := k.SetProposal(input.Ctx, proposal)
	require.NoError(t, err)

	// Voter 1 votes YES
	vote1 := &types.Vote{
		ProposalId: proposalID,
		Voter:      "cosmos1voter1",
		Option:     1, // YES
		Timestamp:  timestamppb.Now(),
	}
	err = k.SetVote(input.Ctx, vote1)
	require.NoError(t, err)

	// Voter 2 votes NO
	vote2 := &types.Vote{
		ProposalId: proposalID,
		Voter:      "cosmos1voter2",
		Option:     3, // NO
		Timestamp:  timestamppb.Now(),
	}
	err = k.SetVote(input.Ctx, vote2)
	require.NoError(t, err)

	// Voter 1 updates to ABSTAIN
	vote3 := &types.Vote{
		ProposalId: proposalID,
		Voter:      "cosmos1voter1",
		Option:     2, // ABSTAIN
		Timestamp:  timestamppb.Now(),
	}
	err = k.SetVote(input.Ctx, vote3)
	require.NoError(t, err)

	// Verify both voters have exactly one vote each
	allVotes := k.GetVotes(input.Ctx, proposalID)
	require.Equal(t, 2, len(allVotes), "Should have exactly 2 votes (one per voter)")

	// Verify voter 1's vote was updated
	voter1Vote, err := k.GetVote(input.Ctx, proposalID, "cosmos1voter1")
	require.NoError(t, err)
	require.Equal(t, int32(2), voter1Vote.Option) // ABSTAIN

	// Verify voter 2's vote unchanged
	voter2Vote, err := k.GetVote(input.Ctx, proposalID, "cosmos1voter2")
	require.NoError(t, err)
	require.Equal(t, int32(3), voter2Vote.Option) // NO
}
