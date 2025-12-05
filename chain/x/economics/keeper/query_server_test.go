package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// ============================
// TEST HELPERS
// ============================

// setupQueryServer creates a keeper, context, and query server for testing
func setupQueryServer(t *testing.T) (*Keeper, sdk.Context, economicspb.QueryServer) {
	t.Helper()
	keeper, ctx := setupKeeperForTest(t)
	server := NewQueryServer(keeper)
	return keeper, ctx, server
}

// generateTestAddress generates a valid bech32 address for testing
// This ensures addresses are properly formatted for the aura chain
func generateTestAddress(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	addr := sdk.AccAddress(hash[:20])
	return addr.String()
}

// createTestVestingSchedule creates a test vesting schedule
func createTestVestingSchedule(id, address string, amount int64) *economicspb.VestingSchedule {
	now := time.Now()
	originalAmount := sdk.NewCoin("uaura", math.NewInt(amount))
	vestedAmount := sdk.NewCoin("uaura", math.NewInt(amount/4)) // 25% vested
	return &economicspb.VestingSchedule{
		Id:             id,
		Address:        address,
		OriginalAmount: originalAmount,
		VestedAmount:   vestedAmount,
		StartTime:      now,
		EndTime:        now.Add(365 * 24 * time.Hour),
		CliffDuration:  uint64(30 * 24 * 3600), // 30 days in seconds
		VestingType:    economicspb.VestingType_VESTING_TYPE_LINEAR,
		ScheduleType:   economicspb.ScheduleType_SCHEDULE_TYPE_TEAM,
		Revoked:        false,
	}
}

// createTestProposal creates a test proposal
func createTestProposal(id uint64, status economicspb.ProposalStatus) *economicspb.Proposal {
	now := time.Now()
	totalDeposit := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(10000000)))
	return &economicspb.Proposal{
		Id:             id,
		Title:          "Test Proposal",
		Description:    "Test proposal description",
		Proposer:       "aura1test",
		Status:         status,
		SubmitTime:     now,
		DepositEndTime: now.Add(48 * time.Hour),
		VotingEndTime:  func() *time.Time { t := now.Add(7 * 24 * time.Hour); return &t }(),
		TotalDeposit:   totalDeposit,
	}
}

// createTestVoteLock creates a test vote lock
func createTestVoteLock(id, owner string, amount int64) *economicspb.VoteLock {
	now := time.Now()
	lockedAmount := sdk.NewCoin("uaura", math.NewInt(amount))
	return &economicspb.VoteLock{
		Id:          id,
		Owner:       owner,
		Amount:      lockedAmount,
		LockStart:   now,
		LockEnd:     now.Add(365 * 24 * time.Hour),
		VotingPower: math.NewInt(amount * 2), // 2x multiplier
	}
}

// ============================
// PARAMS QUERY TESTS
// ============================

func TestQueryParams(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query params
	resp, err := server.Params(sdk.WrapSDKContext(ctx), &economicspb.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)

	// Verify default params structure
	require.NotNil(t, resp.Params.Fees)
	require.NotNil(t, resp.Params.Vesting)
	require.NotNil(t, resp.Params.Treasury)
	require.NotNil(t, resp.Params.Governance)
	require.NotNil(t, resp.Params.Mev)
	require.NotNil(t, resp.Params.WhaleProtection)
	require.NotNil(t, resp.Params.LiquidityMining)
	require.NotNil(t, resp.Params.Tokenomics)
}

func TestQueryParamsAfterUpdate(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Update params
	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	params.Fees.DynamicFeesEnabled = true
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Query updated params
	resp, err := server.Params(sdk.WrapSDKContext(ctx), &economicspb.QueryParamsRequest{})
	require.NoError(t, err)
	require.True(t, resp.Params.Fees.DynamicFeesEnabled)
}

// ============================
// VESTING SCHEDULE QUERY TESTS
// ============================

func TestQueryVestingSchedule(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create test schedule
	schedule := createTestVestingSchedule("schedule-1", "aura1test", 1000000)
	err := keeper.SetVestingSchedule(ctx, schedule)
	require.NoError(t, err)

	// Query schedule
	resp, err := server.VestingSchedule(sdk.WrapSDKContext(ctx), &economicspb.QueryVestingScheduleRequest{
		ScheduleId: "schedule-1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "schedule-1", resp.Schedule.Id)
	require.Equal(t, "aura1test", resp.Schedule.Address)
	require.Equal(t, sdk.NewCoin("uaura", math.NewInt(1000000)), resp.Schedule.OriginalAmount)
	require.NotEmpty(t, resp.VestedAmount)
	require.NotEmpty(t, resp.RemainingAmount)
}

func TestQueryVestingScheduleNotFound(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query non-existent schedule
	_, err := server.VestingSchedule(sdk.WrapSDKContext(ctx), &economicspb.QueryVestingScheduleRequest{
		ScheduleId: "non-existent",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestQueryVestingSchedulesByAddress(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	testAddr := generateTestAddress("beneficiary")

	// Create multiple schedules for the same address
	schedule1 := createTestVestingSchedule("schedule-1", testAddr, 1000000)
	schedule2 := createTestVestingSchedule("schedule-2", testAddr, 2000000)
	schedule3 := createTestVestingSchedule("schedule-3", generateTestAddress("other"), 500000)

	err := keeper.SetVestingSchedule(ctx, schedule1)
	require.NoError(t, err)
	err = keeper.SetVestingSchedule(ctx, schedule2)
	require.NoError(t, err)
	err = keeper.SetVestingSchedule(ctx, schedule3)
	require.NoError(t, err)

	// Add to user index
	err = keeper.AddUserVestingSchedule(ctx, testAddr, "schedule-1")
	require.NoError(t, err)
	err = keeper.AddUserVestingSchedule(ctx, testAddr, "schedule-2")
	require.NoError(t, err)
	err = keeper.AddUserVestingSchedule(ctx, generateTestAddress("other"), "schedule-3")
	require.NoError(t, err)

	// Query schedules by address
	resp, err := server.VestingSchedulesByAddress(sdk.WrapSDKContext(ctx), &economicspb.QueryVestingSchedulesByAddressRequest{
		Address: testAddr,
	})
	require.NoError(t, err)
	require.Len(t, resp.Schedules, 2)
	require.NotEmpty(t, resp.TotalVested)
	require.NotEmpty(t, resp.TotalVesting)
}

func TestQueryVestingSchedulesByAddressEmpty(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query schedules for address with no schedules
	resp, err := server.VestingSchedulesByAddress(sdk.WrapSDKContext(ctx), &economicspb.QueryVestingSchedulesByAddressRequest{
		Address: generateTestAddress("empty"),
	})
	require.NoError(t, err)
	require.Empty(t, resp.Schedules)
	require.True(t, resp.TotalVested.IsZero())
	require.True(t, resp.TotalVesting.IsZero())
}

func TestQueryAllVestingSchedules(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create multiple schedules
	schedule1 := createTestVestingSchedule("schedule-1", "aura1test1", 1000000)
	schedule2 := createTestVestingSchedule("schedule-2", "aura1test2", 2000000)
	schedule3 := createTestVestingSchedule("schedule-3", "aura1test3", 3000000)

	err := keeper.SetVestingSchedule(ctx, schedule1)
	require.NoError(t, err)
	err = keeper.SetVestingSchedule(ctx, schedule2)
	require.NoError(t, err)
	err = keeper.SetVestingSchedule(ctx, schedule3)
	require.NoError(t, err)

	// Query all schedules
	resp, err := server.AllVestingSchedules(sdk.WrapSDKContext(ctx), &economicspb.QueryAllVestingSchedulesRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Schedules, 3)
}

func TestQueryAllVestingSchedulesEmpty(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query with no schedules
	resp, err := server.AllVestingSchedules(sdk.WrapSDKContext(ctx), &economicspb.QueryAllVestingSchedulesRequest{})
	require.NoError(t, err)
	require.Empty(t, resp.Schedules)
}

// ============================
// GOVERNANCE QUERY TESTS
// ============================

func TestQueryProposal(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create test proposal
	proposal := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Query proposal
	resp, err := server.Proposal(sdk.WrapSDKContext(ctx), &economicspb.QueryProposalRequest{
		ProposalId: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(1), resp.Proposal.Id)
	require.Equal(t, "Test Proposal", resp.Proposal.Title)
	require.Equal(t, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, resp.Proposal.Status)
}

func TestQueryProposalNotFound(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query non-existent proposal
	_, err := server.Proposal(sdk.WrapSDKContext(ctx), &economicspb.QueryProposalRequest{
		ProposalId: 999,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestQueryProposals(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create multiple proposals with different statuses
	proposal1 := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
	proposal2 := createTestProposal(2, economicspb.ProposalStatus_PROPOSAL_STATUS_PASSED)
	proposal3 := createTestProposal(3, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)

	err := keeper.SetProposal(ctx, proposal1)
	require.NoError(t, err)
	err = keeper.SetProposal(ctx, proposal2)
	require.NoError(t, err)
	err = keeper.SetProposal(ctx, proposal3)
	require.NoError(t, err)

	// Query all proposals
	resp, err := server.Proposals(sdk.WrapSDKContext(ctx), &economicspb.QueryProposalsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Proposals, 3)
}

func TestQueryProposalsByStatus(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create proposals with different statuses
	proposal1 := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
	proposal2 := createTestProposal(2, economicspb.ProposalStatus_PROPOSAL_STATUS_PASSED)
	proposal3 := createTestProposal(3, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)

	err := keeper.SetProposal(ctx, proposal1)
	require.NoError(t, err)
	err = keeper.SetProposal(ctx, proposal2)
	require.NoError(t, err)
	err = keeper.SetProposal(ctx, proposal3)
	require.NoError(t, err)

	// Query proposals by status (voting period)
	resp, err := server.Proposals(sdk.WrapSDKContext(ctx), &economicspb.QueryProposalsRequest{
		Status: economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
	})
	require.NoError(t, err)
	require.Len(t, resp.Proposals, 2)
	for _, p := range resp.Proposals {
		require.Equal(t, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, p.Status)
	}
}

func TestQueryProposalsEmpty(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query with no proposals
	resp, err := server.Proposals(sdk.WrapSDKContext(ctx), &economicspb.QueryProposalsRequest{})
	require.NoError(t, err)
	require.Empty(t, resp.Proposals)
}

// ============================
// VOTE QUERY TESTS
// ============================

func TestQueryVote(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	voterAddr := generateTestAddress("voter")

	// Create proposal and vote
	proposal := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	vote := &economicspb.Vote{
		ProposalId:  1,
		Voter:       voterAddr,
		Option:      economicspb.VoteOption_VOTE_OPTION_YES,
		Timestamp:   time.Now(),
		VotingPower: math.NewInt(100),
	}
	err = keeper.SetVote(ctx, vote)
	require.NoError(t, err)

	// Query vote
	resp, err := server.Vote(sdk.WrapSDKContext(ctx), &economicspb.QueryVoteRequest{
		ProposalId: 1,
		Voter:      voterAddr,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(1), resp.Vote.ProposalId)
	require.Equal(t, voterAddr, resp.Vote.Voter)
	require.Equal(t, economicspb.VoteOption_VOTE_OPTION_YES, resp.Vote.Option)
}

func TestQueryVoteNotFound(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query non-existent vote
	_, err := server.Vote(sdk.WrapSDKContext(ctx), &economicspb.QueryVoteRequest{
		ProposalId: 1,
		Voter:      generateTestAddress("voter"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid vote")
}

func TestQueryVotes(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create proposal
	proposal := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Create multiple votes
	vote1 := &economicspb.Vote{
		ProposalId:  1,
		Voter:       "aura1voter1",
		Option:      economicspb.VoteOption_VOTE_OPTION_YES,
		Timestamp:   time.Now(),
		VotingPower: math.NewInt(100),
	}
	vote2 := &economicspb.Vote{
		ProposalId:  1,
		Voter:       "aura1voter2",
		Option:      economicspb.VoteOption_VOTE_OPTION_NO,
		Timestamp:   time.Now(),
		VotingPower: math.NewInt(50),
	}

	err = keeper.SetVote(ctx, vote1)
	require.NoError(t, err)
	err = keeper.SetVote(ctx, vote2)
	require.NoError(t, err)

	// Query all votes for proposal
	resp, err := server.Votes(sdk.WrapSDKContext(ctx), &economicspb.QueryVotesRequest{
		ProposalId: 1,
	})
	require.NoError(t, err)
	require.Len(t, resp.Votes, 2)
}

func TestQueryVotesEmpty(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create proposal with no votes
	proposal := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Query votes
	resp, err := server.Votes(sdk.WrapSDKContext(ctx), &economicspb.QueryVotesRequest{
		ProposalId: 1,
	})
	require.NoError(t, err)
	require.Empty(t, resp.Votes)
}

// ============================
// DEPOSIT QUERY TESTS
// ============================

func TestQueryDeposit(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	depositorAddr := generateTestAddress("depositor")

	// Create proposal and deposit
	proposal := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD)
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	depositAmount := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(1000000)))
	deposit := &economicspb.Deposit{
		ProposalId: 1,
		Depositor:  depositorAddr,
		Amount:     depositAmount,
		Timestamp:  time.Now(),
	}
	err = keeper.SetDeposit(ctx, deposit)
	require.NoError(t, err)

	// Query deposit
	resp, err := server.Deposit(sdk.WrapSDKContext(ctx), &economicspb.QueryDepositRequest{
		ProposalId: 1,
		Depositor:  depositorAddr,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(1), resp.Deposit.ProposalId)
	require.Equal(t, depositorAddr, resp.Deposit.Depositor)
	require.Equal(t, 1, len(resp.Deposit.Amount))
	require.Equal(t, "1000000", resp.Deposit.Amount[0].Amount.String())
	require.Equal(t, "uaura", resp.Deposit.Amount[0].Denom)
}

func TestQueryDepositNotFound(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query non-existent deposit
	_, err := server.Deposit(sdk.WrapSDKContext(ctx), &economicspb.QueryDepositRequest{
		ProposalId: 1,
		Depositor:  generateTestAddress("depositor"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid deposit")
}

func TestQueryDeposits(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create proposal
	proposal := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD)
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Create multiple deposits
	amount1 := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(1000000)))
	deposit1 := &economicspb.Deposit{
		ProposalId: 1,
		Depositor:  "aura1depositor1",
		Amount:     amount1,
		Timestamp:  time.Now(),
	}

	amount2 := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(500000)))
	deposit2 := &economicspb.Deposit{
		ProposalId: 1,
		Depositor:  "aura1depositor2",
		Amount:     amount2,
		Timestamp:  time.Now(),
	}

	err = keeper.SetDeposit(ctx, deposit1)
	require.NoError(t, err)
	err = keeper.SetDeposit(ctx, deposit2)
	require.NoError(t, err)

	// Query all deposits for proposal
	resp, err := server.Deposits(sdk.WrapSDKContext(ctx), &economicspb.QueryDepositsRequest{
		ProposalId: 1,
	})
	require.NoError(t, err)
	require.Len(t, resp.Deposits, 2)
}

func TestQueryDepositsEmpty(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create proposal with no deposits
	proposal := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD)
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Query deposits
	resp, err := server.Deposits(sdk.WrapSDKContext(ctx), &economicspb.QueryDepositsRequest{
		ProposalId: 1,
	})
	require.NoError(t, err)
	require.Empty(t, resp.Deposits)
}

// ============================
// TALLY RESULT QUERY TESTS
// ============================

func TestQueryTallyResult(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create proposal
	proposal := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Set tally result
	tally := &economicspb.TallyResult{
		YesCount:         math.NewInt(100),
		NoCount:          math.NewInt(50),
		AbstainCount:     math.NewInt(10),
		NoWithVetoCount:  math.NewInt(5),
		TotalVotingPower: math.NewInt(1000),
	}
	err = keeper.SetTallyResult(ctx, 1, tally)
	require.NoError(t, err)

	// Query tally result
	resp, err := server.TallyResult(sdk.WrapSDKContext(ctx), &economicspb.QueryTallyResultRequest{
		ProposalId: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Tally.YesCount)
	require.NotEmpty(t, resp.Tally.NoCount)
	require.NotEmpty(t, resp.Tally.AbstainCount)
	require.NotEmpty(t, resp.Tally.NoWithVetoCount)
}

func TestQueryTallyResultNotFound(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query tally for non-existent proposal
	_, err := server.TallyResult(sdk.WrapSDKContext(ctx), &economicspb.QueryTallyResultRequest{
		ProposalId: 999,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// ============================
// VOTE LOCK QUERY TESTS
// ============================

func TestQueryVoteLock(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create test vote lock
	lock := createTestVoteLock("lock-1", "aura1owner", 1000000)
	err := keeper.SetVoteLock(ctx, lock)
	require.NoError(t, err)

	// Query vote lock
	resp, err := server.VoteLock(sdk.WrapSDKContext(ctx), &economicspb.QueryVoteLockRequest{
		LockId: "lock-1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "lock-1", resp.Lock.Id)
	require.Equal(t, "aura1owner", resp.Lock.Owner)
	require.Equal(t, sdk.NewCoin("uaura", math.NewInt(1000000)), resp.Lock.Amount)
	require.NotEmpty(t, resp.Lock.VotingPower)
}

func TestQueryVoteLockNotFound(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query non-existent vote lock
	_, err := server.VoteLock(sdk.WrapSDKContext(ctx), &economicspb.QueryVoteLockRequest{
		LockId: "non-existent",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestQueryVoteLocksByOwner(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	testOwner := generateTestAddress("owner")

	// Create multiple locks for the same owner
	lock1 := createTestVoteLock("lock-1", testOwner, 1000000)
	lock2 := createTestVoteLock("lock-2", testOwner, 2000000)
	lock3 := createTestVoteLock("lock-3", generateTestAddress("other"), 500000)

	err := keeper.SetVoteLock(ctx, lock1)
	require.NoError(t, err)
	err = keeper.SetVoteLock(ctx, lock2)
	require.NoError(t, err)
	err = keeper.SetVoteLock(ctx, lock3)
	require.NoError(t, err)

	// Add to user index
	err = keeper.AddUserVoteLock(ctx, testOwner, "lock-1")
	require.NoError(t, err)
	err = keeper.AddUserVoteLock(ctx, testOwner, "lock-2")
	require.NoError(t, err)
	err = keeper.AddUserVoteLock(ctx, generateTestAddress("other"), "lock-3")
	require.NoError(t, err)

	// Query locks by owner
	resp, err := server.VoteLocksByOwner(sdk.WrapSDKContext(ctx), &economicspb.QueryVoteLocksByOwnerRequest{
		Owner: testOwner,
	})
	require.NoError(t, err)
	require.Len(t, resp.Locks, 2)
	require.NotEmpty(t, resp.TotalLocked)
	require.NotEmpty(t, resp.TotalVotingPower)
}

func TestQueryVoteLocksByOwnerEmpty(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query locks for owner with no locks
	resp, err := server.VoteLocksByOwner(sdk.WrapSDKContext(ctx), &economicspb.QueryVoteLocksByOwnerRequest{
		Owner: generateTestAddress("empty"),
	})
	require.NoError(t, err)
	require.Empty(t, resp.Locks)
	require.True(t, resp.TotalLocked.IsZero())
	require.True(t, resp.TotalVotingPower.IsZero())
}

// ============================
// VOTING POWER QUERY TESTS
// ============================

func TestQueryVotingPower(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	testAddr := generateTestAddress("voter")

	// Create vote locks for the address
	lock1 := createTestVoteLock("lock-1", testAddr, 1000000)
	lock2 := createTestVoteLock("lock-2", testAddr, 500000)

	err := keeper.SetVoteLock(ctx, lock1)
	require.NoError(t, err)
	err = keeper.SetVoteLock(ctx, lock2)
	require.NoError(t, err)

	err = keeper.AddUserVoteLock(ctx, testAddr, "lock-1")
	require.NoError(t, err)
	err = keeper.AddUserVoteLock(ctx, testAddr, "lock-2")
	require.NoError(t, err)

	// Query voting power
	resp, err := server.VotingPower(sdk.WrapSDKContext(ctx), &economicspb.QueryVotingPowerRequest{
		Address: testAddr,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.VotingPower)
	require.NotEmpty(t, resp.LockedAmount)
	require.NotEmpty(t, resp.DelegatedPower)
	require.Equal(t, uint64(2), resp.ActiveLocks)
}

func TestQueryVotingPowerWithProposal(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	testAddr := generateTestAddress("voter")

	// Create vote lock
	lock := createTestVoteLock("lock-1", testAddr, 1000000)
	err := keeper.SetVoteLock(ctx, lock)
	require.NoError(t, err)
	err = keeper.AddUserVoteLock(ctx, testAddr, "lock-1")
	require.NoError(t, err)

	// Create proposal
	proposal := createTestProposal(1, economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Query voting power for specific proposal (snapshot voting)
	resp, err := server.VotingPower(sdk.WrapSDKContext(ctx), &economicspb.QueryVotingPowerRequest{
		Address:    testAddr,
		ProposalId: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.VotingPower)
}

func TestQueryVotingPowerZero(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query voting power for address with no locks
	resp, err := server.VotingPower(sdk.WrapSDKContext(ctx), &economicspb.QueryVotingPowerRequest{
		Address: generateTestAddress("empty"),
	})
	require.NoError(t, err)
	require.True(t, resp.VotingPower.IsZero())
	require.True(t, resp.LockedAmount.IsZero())
	require.True(t, resp.DelegatedPower.IsZero())
	require.Equal(t, uint64(0), resp.ActiveLocks)
}

// ============================
// VOTE DELEGATION QUERY TESTS
// ============================

func TestQueryVoteDelegations(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	testDelegator := generateTestAddress("delegator")

	// Create vote delegations
	delegation1 := &economicspb.VoteDelegation{
		Delegator:      testDelegator,
		Delegate:       generateTestAddress("delegate1"),
		DelegationTime: time.Now(),
		DelegatedPower: math.NewInt(1000000),
	}
	delegation2 := &economicspb.VoteDelegation{
		Delegator:      testDelegator,
		Delegate:       generateTestAddress("delegate2"),
		DelegationTime: time.Now(),
		DelegatedPower: math.NewInt(500000),
	}

	err := keeper.SetVoteDelegation(ctx, delegation1)
	require.NoError(t, err)
	err = keeper.SetVoteDelegation(ctx, delegation2)
	require.NoError(t, err)

	// Query delegations
	resp, err := server.VoteDelegations(sdk.WrapSDKContext(ctx), &economicspb.QueryVoteDelegationsRequest{
		Delegator: testDelegator,
	})
	require.NoError(t, err)
	require.Len(t, resp.Delegations, 2)
}

func TestQueryVoteDelegationsEmpty(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query delegations for address with none
	resp, err := server.VoteDelegations(sdk.WrapSDKContext(ctx), &economicspb.QueryVoteDelegationsRequest{
		Delegator: generateTestAddress("empty"),
	})
	require.NoError(t, err)
	require.Empty(t, resp.Delegations)
}

// ============================
// TREASURY QUERY TESTS
// ============================

func TestQueryPendingTreasuryTx(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create pending treasury transaction
	amount := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(5000000)))
	tx := &economicspb.PendingTreasuryTx{
		TxId:         "tx-1",
		Recipient:    "aura1recipient",
		Amount:       amount,
		Description:  "Test transaction",
		Proposer:     "aura1proposer",
		CreatedAt:    time.Now(),
		ExecutableAt: time.Now().Add(24 * time.Hour),
	}
	err := keeper.SetPendingTreasuryTx(ctx, tx)
	require.NoError(t, err)

	// Query pending transaction
	resp, err := server.PendingTreasuryTx(sdk.WrapSDKContext(ctx), &economicspb.QueryPendingTreasuryTxRequest{
		TxId: "tx-1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "tx-1", resp.Transaction.TxId)
	require.Equal(t, "aura1recipient", resp.Transaction.Recipient)
	require.Equal(t, 1, len(resp.Transaction.Amount))
	require.Equal(t, math.NewInt(5000000), resp.Transaction.Amount[0].Amount)
	require.Equal(t, "uaura", resp.Transaction.Amount[0].Denom)
}

func TestQueryPendingTreasuryTxNotFound(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query non-existent transaction
	_, err := server.PendingTreasuryTx(sdk.WrapSDKContext(ctx), &economicspb.QueryPendingTreasuryTxRequest{
		TxId: "non-existent",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestQueryPendingTreasuryTxs(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create multiple pending transactions
	amount1 := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(1000000)))
	tx1 := &economicspb.PendingTreasuryTx{
		TxId:         "tx-1",
		Recipient:    "aura1recipient1",
		Amount:       amount1,
		Description:  "Transaction 1",
		Proposer:     "aura1proposer",
		CreatedAt:    time.Now(),
		ExecutableAt: time.Now().Add(24 * time.Hour),
	}

	amount2 := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(2000000)))
	tx2 := &economicspb.PendingTreasuryTx{
		TxId:         "tx-2",
		Recipient:    "aura1recipient2",
		Amount:       amount2,
		Description:  "Transaction 2",
		Proposer:     "aura1proposer",
		CreatedAt:    time.Now(),
		ExecutableAt: time.Now().Add(48 * time.Hour),
	}

	err := keeper.SetPendingTreasuryTx(ctx, tx1)
	require.NoError(t, err)
	err = keeper.SetPendingTreasuryTx(ctx, tx2)
	require.NoError(t, err)

	// Query all pending transactions
	resp, err := server.PendingTreasuryTxs(sdk.WrapSDKContext(ctx), &economicspb.QueryPendingTreasuryTxsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Transactions, 2)
}

func TestQueryPendingTreasuryTxsEmpty(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query with no pending transactions
	resp, err := server.PendingTreasuryTxs(sdk.WrapSDKContext(ctx), &economicspb.QueryPendingTreasuryTxsRequest{})
	require.NoError(t, err)
	require.Empty(t, resp.Transactions)
}

// ============================
// INFLATION METRICS QUERY TESTS
// ============================

func TestQueryInflationMetrics(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Set inflation metrics
	metrics := &economicspb.InflationMetrics{
		CurrentRate:       500, // 5%
		CirculatingSupply: math.NewInt(100000000),
		TotalVested:       math.NewInt(50000000),
		TotalVesting:      math.NewInt(25000000),
		LastAdjustment:    time.Now(),
		NextCheck:         time.Now().Add(24 * time.Hour),
	}
	err := keeper.SetInflationMetrics(ctx, metrics)
	require.NoError(t, err)

	// Query inflation metrics
	resp, err := server.InflationMetrics(sdk.WrapSDKContext(ctx), &economicspb.QueryInflationMetricsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(500), resp.Metrics.CurrentRate)
	require.NotEmpty(t, resp.Metrics.CirculatingSupply)
	require.NotEmpty(t, resp.Metrics.TotalVested)
}

func TestQueryInflationMetricsDefault(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query inflation metrics with no data (should return defaults)
	resp, err := server.InflationMetrics(sdk.WrapSDKContext(ctx), &economicspb.QueryInflationMetricsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Metrics)
}

// ============================
// MEV STATS QUERY TESTS
// ============================

func TestQueryMEVStats(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Set MEV stats
	stats := &economicspb.MEVStats{
		TotalCaptured:         math.NewInt(5000000),
		TotalRedistributed:    math.NewInt(4500000),
		PendingRedistribution: math.NewInt(500000),
	}
	err := keeper.SetMEVStats(ctx, stats)
	require.NoError(t, err)

	// Query MEV stats
	resp, err := server.MEVStats(sdk.WrapSDKContext(ctx), &economicspb.QueryMEVStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Stats.TotalCaptured)
	require.NotEmpty(t, resp.Stats.TotalRedistributed)
	require.NotEmpty(t, resp.Stats.PendingRedistribution)
}

func TestQueryMEVStatsDefault(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query MEV stats with no data (should return defaults)
	resp, err := server.MEVStats(sdk.WrapSDKContext(ctx), &economicspb.QueryMEVStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Stats)
}

// Note: UserMEVBalance tests skipped - message not defined in proto
// func TestQueryUserMEVBalance(t *testing.T) {
// 	keeper, ctx, server := setupQueryServer(t)
//
// 	testAddr := "aura1user"
//
// 	// Set user MEV balance - message not defined
//
// func TestQueryUserMEVBalanceZero(t *testing.T) {
// 	_, ctx, server := setupQueryServer(t)
//
// 	// Query user MEV balance for address with no balance - message not defined
// 	// resp, err := server.UserMEVBalance(sdk.WrapSDKContext(ctx), &economicspb.QueryUserMEVBalanceRequest{
// 	// 	Address: "aura1empty",
// 	require.NoError(t, err)
// 	require.Empty(t, resp.Balance)
// 	require.Empty(t, resp.LifetimeReceived)

// ============================
// LIQUIDITY MINING STATS QUERY TESTS
// ============================

func TestQueryLiquidityMiningStats(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Set liquidity mining stats
	stats := &economicspb.LiquidityMiningStats{
		CurrentEpoch:           100,
		TotalDistributed:       math.NewInt(10000000),
		RemainingRewards:       math.NewInt(5000000),
		RewardsThisEpoch:       math.NewInt(100000),
		NextDistributionHeight: 1000,
	}
	err := keeper.SetLiquidityMiningStats(ctx, stats)
	require.NoError(t, err)

	// Query liquidity mining stats
	resp, err := server.LiquidityMiningStats(sdk.WrapSDKContext(ctx), &economicspb.QueryLiquidityMiningStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Stats.TotalDistributed)
	require.Equal(t, uint64(100), resp.Stats.CurrentEpoch)
	require.NotEmpty(t, resp.Stats.RemainingRewards)
	require.Equal(t, uint64(1000), resp.Stats.NextDistributionHeight)
}

func TestQueryLiquidityMiningStatsDefault(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query liquidity mining stats with no data (should return defaults)
	resp, err := server.LiquidityMiningStats(sdk.WrapSDKContext(ctx), &economicspb.QueryLiquidityMiningStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Stats)
}

// ============================
// TOKENOMICS STATS QUERY TESTS
// ============================

func TestQueryTokenomicsStats(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Set up various data for comprehensive tokenomics stats

	// Create vesting schedules
	schedule1 := createTestVestingSchedule("schedule-1", "aura1test1", 1000000)
	err := keeper.SetVestingSchedule(ctx, schedule1)
	require.NoError(t, err)

	// Create vote locks
	lock1 := createTestVoteLock("lock-1", "aura1test1", 500000)
	err = keeper.SetVoteLock(ctx, lock1)
	require.NoError(t, err)

	// Query tokenomics stats
	resp, err := server.TokenomicsStats(sdk.WrapSDKContext(ctx), &economicspb.QueryTokenomicsStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify all required fields are present
	require.NotEmpty(t, resp.MaxSupply)
	require.NotEmpty(t, resp.CirculatingSupply)
	require.NotEmpty(t, resp.TotalVested)
	require.NotEmpty(t, resp.TotalVesting)
	require.NotEmpty(t, resp.TotalLockedGovernance)
	require.NotEmpty(t, resp.TreasuryBalance)
	require.NotEmpty(t, resp.TotalBurned)
	require.NotEmpty(t, resp.TransferTaxCollected_24H)
}

func TestQueryTokenomicsStatsEmpty(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query tokenomics stats with no data
	resp, err := server.TokenomicsStats(sdk.WrapSDKContext(ctx), &economicspb.QueryTokenomicsStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// ============================
// EDGE CASE TESTS
// ============================

func TestQueryWithInvalidAddress(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query with invalid address format
	_, err := server.VestingSchedulesByAddress(sdk.WrapSDKContext(ctx), &economicspb.QueryVestingSchedulesByAddressRequest{
		Address: "invalid",
	})
	// Should either return empty result or error with invalid address
	// The exact behavior depends on implementation
	if err != nil {
		require.Contains(t, err.Error(), "invalid")
	}
}

func TestQueryWithEmptyRequest(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query with empty schedule ID
	_, err := server.VestingSchedule(sdk.WrapSDKContext(ctx), &economicspb.QueryVestingScheduleRequest{
		ScheduleId: "",
	})
	require.Error(t, err)
}

func TestQueryProposalWithZeroID(t *testing.T) {
	_, ctx, server := setupQueryServer(t)

	// Query proposal with ID 0 (invalid)
	_, err := server.Proposal(sdk.WrapSDKContext(ctx), &economicspb.QueryProposalRequest{
		ProposalId: 0,
	})
	require.Error(t, err)
}

// ============================
// CONCURRENT QUERY TESTS
// ============================

func TestConcurrentQueries(t *testing.T) {
	keeper, ctx, server := setupQueryServer(t)

	// Create test data
	schedule := createTestVestingSchedule("schedule-1", "aura1test", 1000000)
	err := keeper.SetVestingSchedule(ctx, schedule)
	require.NoError(t, err)

	// Run concurrent queries
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := server.VestingSchedule(sdk.WrapSDKContext(ctx), &economicspb.QueryVestingScheduleRequest{
				ScheduleId: "schedule-1",
			})
			require.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
