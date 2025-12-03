package keeper

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// setupQuorumTestKeeper creates a keeper with extended mock functionality for quorum testing
func setupQuorumTestKeeper(t *testing.T) (*Keeper, sdk.Context, *ExtendedMockStakingKeeper, *ExtendedMockBankKeeper) {
	storeKey := storetypes.NewKVStoreKey("governance")

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create extended mock keepers
	stakingKeeper := &ExtendedMockStakingKeeper{
		delegatorBonded: make(map[string]sdkmath.Int),
		totalBonded:     sdkmath.ZeroInt(),
	}
	bankKeeper := &ExtendedMockBankKeeper{
		refundCallCount: 0,
	}
	securityKeeper := &MockSecurityKeeper{}

	keeper := NewKeeper(cdc, storeKey, stakingKeeper, bankKeeper, securityKeeper)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)
	keeper.SetNextProposalID(ctx, 1)

	return keeper, ctx, stakingKeeper, bankKeeper
}

// ExtendedMockStakingKeeper extends MockStakingKeeper with total bonded tokens tracking
type ExtendedMockStakingKeeper struct {
	delegatorBonded map[string]sdkmath.Int
	totalBonded     sdkmath.Int
}

func (m *ExtendedMockStakingKeeper) GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (sdkmath.Int, error) {
	if amount, ok := m.delegatorBonded[delegator.String()]; ok {
		return amount, nil
	}
	return sdkmath.ZeroInt(), nil
}

func (m *ExtendedMockStakingKeeper) TotalBondedTokens(ctx context.Context) (sdkmath.Int, error) {
	return m.totalBonded, nil
}

func (m *ExtendedMockStakingKeeper) SetTotalBondedTokens(total sdkmath.Int) {
	m.totalBonded = total
}

func (m *ExtendedMockStakingKeeper) SetDelegatorBonded(delegator sdk.AccAddress, amount sdkmath.Int) {
	m.delegatorBonded[delegator.String()] = amount
}

// ExtendedMockBankKeeper extends MockBankKeeper with call tracking
type ExtendedMockBankKeeper struct {
	refundCallCount int
}

func (m *ExtendedMockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *ExtendedMockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	m.refundCallCount++
	return nil
}

func (m *ExtendedMockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}

// MockSecurityKeeper is a mock security keeper for testing
type MockSecurityKeeper struct{}

func (m *MockSecurityKeeper) EnterNoReentrant(ctx sdk.Context, key string) error {
	return nil
}

func (m *MockSecurityKeeper) ExitNoReentrant(ctx sdk.Context, key string) {}

func (m *MockSecurityKeeper) WithReentrancyGuard(ctx sdk.Context, key string, fn func() error) error {
	return fn()
}

// TestQuorumEnforcement_ProposalFailsWithoutQuorum tests that proposals fail when quorum is not met
func TestQuorumEnforcement_ProposalFailsWithoutQuorum(t *testing.T) {
	k, ctx, stakingKeeper, _ := setupQuorumTestKeeper(t)

	// Setup: Mock 1,000,000 total bonded tokens
	// With default 33.4% quorum, need 334,000 votes minimum
	stakingKeeper.SetTotalBondedTokens(sdkmath.NewInt(1_000_000))

	// Create a proposal
	proposalID, err := k.CreateProposal(
		ctx,
		"Test Proposal",
		"Test quorum enforcement",
		"aura1proposer",
		types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		"content",
	)
	require.NoError(t, err)

	// Move proposal to voting period
	proposal, err := k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(time.Now())
	proposal.VotingEndTime = timestamppb.New(time.Now().Add(7 * 24 * time.Hour))
	err = k.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Cast votes totaling only 100,000 (below 334,000 quorum)
	// Even though all votes are YES, proposal should fail due to low quorum
	voters := []struct {
		address string
		power   int64
	}{
		{"aura1voter1", 50_000},
		{"aura1voter2", 30_000},
		{"aura1voter3", 20_000},
	}

	for _, voter := range voters {
		// Mock voting power for each voter
		voterAddr, _ := sdk.AccAddressFromBech32(voter.address)
		stakingKeeper.SetDelegatorBonded(voterAddr, sdkmath.NewInt(voter.power))

		// Cast YES vote
		vote := &types.Vote{
			ProposalId: proposalID,
			Voter:      voter.address,
			Option:     types.VoteOption_VOTE_OPTION_YES,
		}
		err = k.SetVote(ctx, vote)
		require.NoError(t, err)
	}

	// Calculate tally
	tally := k.CalculateTally(ctx, proposalID)
	require.Equal(t, "100000", tally.Yes, "Total yes votes should be 100,000")
	require.Equal(t, "0", tally.No)
	require.Equal(t, "0", tally.Abstain)
	require.Equal(t, "0", tally.NoWithVeto)

	// Process proposal outcome through finalizeProposal (which calls processProposalOutcome internally)
	proposal, _ = k.GetProposal(ctx, proposalID)
	params := k.GetParams(ctx)
	err = k.finalizeProposal(ctx, proposal, params)
	require.NoError(t, err)

	// Verify proposal was REJECTED due to insufficient quorum
	proposal, _ = k.GetProposal(ctx, proposalID)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
		"Proposal should be rejected due to insufficient quorum")
}

// TestQuorumEnforcement_ProposalPassesWithSufficientQuorum tests that proposals pass when quorum is met
func TestQuorumEnforcement_ProposalPassesWithSufficientQuorum(t *testing.T) {
	k, ctx, stakingKeeper, _ := setupQuorumTestKeeper(t)

	// Setup: Mock 1,000,000 total bonded tokens
	// With default 33.4% quorum, need 334,000 votes minimum
	stakingKeeper.SetTotalBondedTokens(sdkmath.NewInt(1_000_000))

	// Create a proposal
	proposalID, err := k.CreateProposal(
		ctx,
		"Test Proposal",
		"Test quorum enforcement",
		"aura1proposer",
		types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		"content",
	)
	require.NoError(t, err)

	// Move proposal to voting period
	proposal, err := k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(time.Now())
	proposal.VotingEndTime = timestamppb.New(time.Now().Add(7 * 24 * time.Hour))
	err = k.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Cast votes totaling 400,000 (above 334,000 quorum)
	// 250,001 YES (50% + 1 of non-abstain votes)
	// 149,999 NO (remaining to meet quorum)
	voters := []struct {
		address string
		power   int64
		option  types.VoteOption
	}{
		{"aura1voter1", 150_000, types.VoteOption_VOTE_OPTION_YES},
		{"aura1voter2", 100_001, types.VoteOption_VOTE_OPTION_YES},
		{"aura1voter3", 149_999, types.VoteOption_VOTE_OPTION_NO},
	}

	for _, voter := range voters {
		// Mock voting power for each voter
		voterAddr, _ := sdk.AccAddressFromBech32(voter.address)
		stakingKeeper.SetDelegatorBonded(voterAddr, sdkmath.NewInt(voter.power))

		// Cast vote
		vote := &types.Vote{
			ProposalId: proposalID,
			Voter:      voter.address,
			Option:     voter.option,
		}
		err = k.SetVote(ctx, vote)
		require.NoError(t, err)
	}

	// Calculate tally
	tally := k.CalculateTally(ctx, proposalID)
	require.Equal(t, "250001", tally.Yes, "Total yes votes should be 250,001")
	require.Equal(t, "149999", tally.No, "Total no votes should be 149,999")

	// Process proposal outcome
	proposal, _ = k.GetProposal(ctx, proposalID)
	params := k.GetParams(ctx)
	err = k.finalizeProposal(ctx, proposal, params)
	require.NoError(t, err)

	// Verify proposal was PASSED (quorum met, threshold met)
	proposal, _ = k.GetProposal(ctx, proposalID)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status,
		"Proposal should pass with sufficient quorum and threshold")
}

// TestQuorumEnforcement_ProposalFailsThresholdDespiteQuorum tests that proposals fail when threshold not met even if quorum is met
func TestQuorumEnforcement_ProposalFailsThresholdDespiteQuorum(t *testing.T) {
	k, ctx, stakingKeeper, _ := setupQuorumTestKeeper(t)

	// Setup: Mock 1,000,000 total bonded tokens
	stakingKeeper.SetTotalBondedTokens(sdkmath.NewInt(1_000_000))

	// Create a proposal
	proposalID, err := k.CreateProposal(
		ctx,
		"Test Proposal",
		"Test threshold enforcement",
		"aura1proposer",
		types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		"content",
	)
	require.NoError(t, err)

	// Move proposal to voting period
	proposal, err := k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(time.Now())
	proposal.VotingEndTime = timestamppb.New(time.Now().Add(7 * 24 * time.Hour))
	err = k.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Cast votes totaling 400,000 (meets quorum)
	// BUT: Only 180,000 YES (45% of total) - below 50% threshold
	// 220,000 NO (55% of total)
	voters := []struct {
		address string
		power   int64
		option  types.VoteOption
	}{
		{"aura1voter1", 100_000, types.VoteOption_VOTE_OPTION_YES},
		{"aura1voter2", 80_000, types.VoteOption_VOTE_OPTION_YES},
		{"aura1voter3", 120_000, types.VoteOption_VOTE_OPTION_NO},
		{"aura1voter4", 100_000, types.VoteOption_VOTE_OPTION_NO},
	}

	for _, voter := range voters {
		// Mock voting power for each voter
		voterAddr, _ := sdk.AccAddressFromBech32(voter.address)
		stakingKeeper.SetDelegatorBonded(voterAddr, sdkmath.NewInt(voter.power))

		// Cast vote
		vote := &types.Vote{
			ProposalId: proposalID,
			Voter:      voter.address,
			Option:     voter.option,
		}
		err = k.SetVote(ctx, vote)
		require.NoError(t, err)
	}

	// Calculate tally
	tally := k.CalculateTally(ctx, proposalID)
	require.Equal(t, "180000", tally.Yes, "Total yes votes should be 180,000")
	require.Equal(t, "220000", tally.No, "Total no votes should be 220,000")

	// Process proposal outcome
	proposal, _ = k.GetProposal(ctx, proposalID)
	params := k.GetParams(ctx)
	err = k.finalizeProposal(ctx, proposal, params)
	require.NoError(t, err)

	// Verify proposal was REJECTED (quorum met, but threshold not met)
	proposal, _ = k.GetProposal(ctx, proposalID)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
		"Proposal should be rejected - quorum met but threshold not met (45% < 50%)")
}

// TestQuorumEnforcement_VetoOverridesQuorumAndThreshold tests that veto overrides even when quorum and threshold are met
func TestQuorumEnforcement_VetoOverridesQuorumAndThreshold(t *testing.T) {
	k, ctx, stakingKeeper, _ := setupQuorumTestKeeper(t)

	// Setup: Mock 1,000,000 total bonded tokens
	stakingKeeper.SetTotalBondedTokens(sdkmath.NewInt(1_000_000))

	// Create a proposal
	proposalID, err := k.CreateProposal(
		ctx,
		"Test Proposal",
		"Test veto enforcement",
		"aura1proposer",
		types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		"content",
	)
	require.NoError(t, err)

	// Move proposal to voting period
	proposal, err := k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(time.Now())
	proposal.VotingEndTime = timestamppb.New(time.Now().Add(7 * 24 * time.Hour))
	err = k.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Cast votes totaling 1,000,000 (100% participation - far above quorum)
	// 600,000 YES (60% - well above threshold)
	// 50,000 NO (5%)
	// 350,000 NO_WITH_VETO (35% - above 33.4% veto threshold)
	voters := []struct {
		address string
		power   int64
		option  types.VoteOption
	}{
		{"aura1voter1", 300_000, types.VoteOption_VOTE_OPTION_YES},
		{"aura1voter2", 300_000, types.VoteOption_VOTE_OPTION_YES},
		{"aura1voter3", 50_000, types.VoteOption_VOTE_OPTION_NO},
		{"aura1voter4", 200_000, types.VoteOption_VOTE_OPTION_NO_WITH_VETO},
		{"aura1voter5", 150_000, types.VoteOption_VOTE_OPTION_NO_WITH_VETO},
	}

	for _, voter := range voters {
		// Mock voting power for each voter
		voterAddr, _ := sdk.AccAddressFromBech32(voter.address)
		stakingKeeper.SetDelegatorBonded(voterAddr, sdkmath.NewInt(voter.power))

		// Cast vote
		vote := &types.Vote{
			ProposalId: proposalID,
			Voter:      voter.address,
			Option:     voter.option,
		}
		err = k.SetVote(ctx, vote)
		require.NoError(t, err)
	}

	// Calculate tally
	tally := k.CalculateTally(ctx, proposalID)
	require.Equal(t, "600000", tally.Yes, "Total yes votes should be 600,000")
	require.Equal(t, "50000", tally.No, "Total no votes should be 50,000")
	require.Equal(t, "350000", tally.NoWithVeto, "Total veto votes should be 350,000")

	// Process proposal outcome
	proposal, _ = k.GetProposal(ctx, proposalID)
	params := k.GetParams(ctx)
	err = k.finalizeProposal(ctx, proposal, params)
	require.NoError(t, err)

	// Verify proposal was VETOED (despite having quorum and threshold)
	proposal, _ = k.GetProposal(ctx, proposalID)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_VETOED, proposal.Status,
		"Proposal should be vetoed when NoWithVeto exceeds veto threshold (35% > 33.4%)")
}

// TestQuorumEnforcement_OnlyAbstainVotes tests that proposals with only abstain votes are rejected
func TestQuorumEnforcement_OnlyAbstainVotes(t *testing.T) {
	k, ctx, stakingKeeper, _ := setupQuorumTestKeeper(t)

	// Setup: Mock 1,000,000 total bonded tokens
	stakingKeeper.SetTotalBondedTokens(sdkmath.NewInt(1_000_000))

	// Create a proposal
	proposalID, err := k.CreateProposal(
		ctx,
		"Test Proposal",
		"Test abstain votes",
		"aura1proposer",
		types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		"content",
	)
	require.NoError(t, err)

	// Move proposal to voting period
	proposal, err := k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(time.Now())
	proposal.VotingEndTime = timestamppb.New(time.Now().Add(7 * 24 * time.Hour))
	err = k.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Cast only abstain votes (400,000 - meets quorum)
	voters := []struct {
		address string
		power   int64
	}{
		{"aura1voter1", 200_000},
		{"aura1voter2", 200_000},
	}

	for _, voter := range voters {
		// Mock voting power for each voter
		voterAddr, _ := sdk.AccAddressFromBech32(voter.address)
		stakingKeeper.SetDelegatorBonded(voterAddr, sdkmath.NewInt(voter.power))

		// Cast ABSTAIN vote
		vote := &types.Vote{
			ProposalId: proposalID,
			Voter:      voter.address,
			Option:     types.VoteOption_VOTE_OPTION_ABSTAIN,
		}
		err = k.SetVote(ctx, vote)
		require.NoError(t, err)
	}

	// Calculate tally
	tally := k.CalculateTally(ctx, proposalID)
	require.Equal(t, "0", tally.Yes)
	require.Equal(t, "0", tally.No)
	require.Equal(t, "400000", tally.Abstain, "Total abstain votes should be 400,000")
	require.Equal(t, "0", tally.NoWithVeto)

	// Process proposal outcome
	proposal, _ = k.GetProposal(ctx, proposalID)
	params := k.GetParams(ctx)
	err = k.finalizeProposal(ctx, proposal, params)
	require.NoError(t, err)

	// Verify proposal was REJECTED (only abstain votes)
	proposal, _ = k.GetProposal(ctx, proposalID)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
		"Proposal should be rejected when only abstain votes are cast")
}

// TestQuorumEnforcement_DepositHandling tests that deposits are handled correctly based on outcome
func TestQuorumEnforcement_DepositHandling(t *testing.T) {
	k, ctx, stakingKeeper, bankKeeper := setupQuorumTestKeeper(t)

	stakingKeeper.SetTotalBondedTokens(sdkmath.NewInt(1_000_000))

	t.Run("DepositsRefundedForPassedProposal", func(t *testing.T) {
		// Create a proposal with deposits
		proposalID, err := k.CreateProposal(
			ctx,
			"Test Proposal",
			"Test deposit refund",
			"aura1proposer",
			types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
			"content",
		)
		require.NoError(t, err)

		// Add deposit
		deposit := &types.Deposit{
			ProposalId: proposalID,
			Depositor:  "aura1depositor",
			Amount:     "1000000",
		}
		err = k.SetDeposit(ctx, deposit)
		require.NoError(t, err)

		// Move to voting period
		proposal, _ := k.GetProposal(ctx, proposalID)
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
		proposal.VotingStartTime = timestamppb.New(time.Now())
		proposal.VotingEndTime = timestamppb.New(time.Now().Add(7 * 24 * time.Hour))
		k.SetProposal(ctx, proposal)

		// Cast passing votes (meets quorum and threshold)
		voters := []struct {
			address string
			power   int64
			option  types.VoteOption
		}{
			{"aura1voter1", 250_001, types.VoteOption_VOTE_OPTION_YES},
			{"aura1voter2", 150_000, types.VoteOption_VOTE_OPTION_NO},
		}

		for _, voter := range voters {
			voterAddr, _ := sdk.AccAddressFromBech32(voter.address)
			stakingKeeper.SetDelegatorBonded(voterAddr, sdkmath.NewInt(voter.power))

			vote := &types.Vote{
				ProposalId: proposalID,
				Voter:      voter.address,
				Option:     voter.option,
			}
			k.SetVote(ctx, vote)
		}

		// Process outcome
		proposal, _ = k.GetProposal(ctx, proposalID)
		params := k.GetParams(ctx)
		err = k.finalizeProposal(ctx, proposal, params)
		require.NoError(t, err)

		// Verify refund was called
		require.Greater(t, bankKeeper.refundCallCount, 0,
			"Deposit refund should be called for passed proposal")

		// Verify proposal passed
		proposal, _ = k.GetProposal(ctx, proposalID)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
	})

	t.Run("DepositsBurnedForVetoedProposal", func(t *testing.T) {
		// Reset call counter
		bankKeeper.refundCallCount = 0

		// Create a proposal with deposits
		proposalID, err := k.CreateProposal(
			ctx,
			"Vetoed Proposal",
			"Test deposit burn",
			"aura1proposer2",
			types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
			"content",
		)
		require.NoError(t, err)

		// Add deposit
		deposit := &types.Deposit{
			ProposalId: proposalID,
			Depositor:  "aura1depositor2",
			Amount:     "1000000",
		}
		err = k.SetDeposit(ctx, deposit)
		require.NoError(t, err)

		// Move to voting period
		proposal, _ := k.GetProposal(ctx, proposalID)
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
		proposal.VotingStartTime = timestamppb.New(time.Now())
		proposal.VotingEndTime = timestamppb.New(time.Now().Add(7 * 24 * time.Hour))
		k.SetProposal(ctx, proposal)

		// Cast votes with veto (meets quorum, veto threshold exceeded)
		voters := []struct {
			address string
			power   int64
			option  types.VoteOption
		}{
			{"aura1voter3", 300_000, types.VoteOption_VOTE_OPTION_YES},
			{"aura1voter4", 350_000, types.VoteOption_VOTE_OPTION_NO_WITH_VETO}, // 35% > 33.4% threshold
		}

		for _, voter := range voters {
			voterAddr, _ := sdk.AccAddressFromBech32(voter.address)
			stakingKeeper.SetDelegatorBonded(voterAddr, sdkmath.NewInt(voter.power))

			vote := &types.Vote{
				ProposalId: proposalID,
				Voter:      voter.address,
				Option:     voter.option,
			}
			k.SetVote(ctx, vote)
		}

		// Process outcome
		proposal, _ = k.GetProposal(ctx, proposalID)
		params := k.GetParams(ctx)
		err = k.finalizeProposal(ctx, proposal, params)
		require.NoError(t, err)

		// Verify no refund was called (deposits burned)
		require.Equal(t, 0, bankKeeper.refundCallCount,
			"Deposit should NOT be refunded for vetoed proposal (burned instead)")

		// Verify proposal status
		proposal, _ = k.GetProposal(ctx, proposalID)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_VETOED, proposal.Status,
			"Proposal should be vetoed")
	})
}
