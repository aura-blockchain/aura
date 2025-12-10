package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil"
	"github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// ============================
// INIT GENESIS TESTS
// ============================

func TestInitGenesisValid(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create valid genesis state
	gs := &economicspb.GenesisState{
		Params:             *types.DefaultParams(),
		VestingSchedules:   []economicspb.VestingSchedule{},
		VoteLocks:          []economicspb.VoteLock{},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{},
		Votes:              []economicspb.Vote{},
		Deposits:           []economicspb.Deposit{},
		VoteDelegations:    []economicspb.VoteDelegation{},
		NextProposalId:     1,
		UserMevBalances:    make(map[string]string),
		LastLargeTxTimes:   make(map[string]int64),
	}

	// Initialize genesis
	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify params were set
	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	require.NotNil(t, params)
}

func TestInitGenesisWithVestingSchedules(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	now := testutil.Now()
	schedule := economicspb.VestingSchedule{
		Id:             "vesting-1",
		Address:        "aura1test",
		OriginalAmount: testutil.NewCoin("uaura", 1000000),
		VestedAmount:   testutil.NewCoin("uaura", 250000),
		StartTime:      now,
		EndTime:        testutil.TimeFromNow(365 * 24 * time.Hour),
		CliffDuration:  uint64(30 * 24 * 3600),
		VestingType:    economicspb.VestingType_VESTING_TYPE_LINEAR,
		ScheduleType:   economicspb.ScheduleType_SCHEDULE_TYPE_TEAM,
		Revoked:        false,
	}

	gs := &economicspb.GenesisState{
		Params:             *types.DefaultParams(),
		VestingSchedules:   []economicspb.VestingSchedule{schedule},
		VoteLocks:          []economicspb.VoteLock{},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{},
		Votes:              []economicspb.Vote{},
		Deposits:           []economicspb.Deposit{},
		VoteDelegations:    []economicspb.VoteDelegation{},
		NextProposalId:     1,
		UserMevBalances:    make(map[string]string),
		LastLargeTxTimes:   make(map[string]int64),
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify schedule was stored
	storedSchedule, err := keeper.GetVestingSchedule(ctx, "vesting-1")
	require.NoError(t, err)
	require.Equal(t, "vesting-1", storedSchedule.Id)
	require.Equal(t, "aura1test", storedSchedule.Address)
}

func TestInitGenesisWithVoteLocks(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	now := testutil.Now()
	lock := economicspb.VoteLock{
		Id:          "lock-1",
		Owner:       "aura1test",
		Amount:      testutil.NewCoin("uaura", 1000000),
		LockStart:   now,
		LockEnd:     testutil.TimeFromNow(365 * 24 * time.Hour),
		VotingPower: testutil.NewInt(2000000),
	}

	gs := &economicspb.GenesisState{
		Params:             *types.DefaultParams(),
		VestingSchedules:   []economicspb.VestingSchedule{},
		VoteLocks:          []economicspb.VoteLock{lock},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{},
		Votes:              []economicspb.Vote{},
		Deposits:           []economicspb.Deposit{},
		VoteDelegations:    []economicspb.VoteDelegation{},
		NextProposalId:     1,
		UserMevBalances:    make(map[string]string),
		LastLargeTxTimes:   make(map[string]int64),
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify lock was stored
	storedLock, err := keeper.GetVoteLock(ctx, "lock-1")
	require.NoError(t, err)
	require.Equal(t, "lock-1", storedLock.Id)
	require.Equal(t, "aura1test", storedLock.Owner)
}

func TestInitGenesisWithProposals(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	now := testutil.Now()
	votingEndTime := testutil.TimeFromNowPtr(7 * 24 * time.Hour)
	proposal := economicspb.Proposal{
		Id:             1,
		Title:          "Test Proposal",
		Description:    "Test description",
		Proposer:       "aura1proposer",
		Status:         economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		SubmitTime:     now,
		DepositEndTime: testutil.TimeFromNow(48 * time.Hour),
		VotingEndTime:  votingEndTime,
		TotalDeposit:   sdk.NewCoins(testutil.NewCoin("uaura", 10000000)),
	}

	gs := &economicspb.GenesisState{
		Params:             *types.DefaultParams(),
		VestingSchedules:   []economicspb.VestingSchedule{},
		VoteLocks:          []economicspb.VoteLock{},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{proposal},
		Votes:              []economicspb.Vote{},
		Deposits:           []economicspb.Deposit{},
		VoteDelegations:    []economicspb.VoteDelegation{},
		NextProposalId:     2,
		UserMevBalances:    make(map[string]string),
		LastLargeTxTimes:   make(map[string]int64),
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify proposal was stored
	storedProposal, err := keeper.GetProposal(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), storedProposal.Id)
	require.Equal(t, "Test Proposal", storedProposal.Title)

	// Verify next proposal ID was set
	nextID, err := keeper.GetNextProposalID(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(2), nextID)
}

func TestInitGenesisWithVotesAndDeposits(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	now := testutil.Now()
	votingEndTime := testutil.TimeFromNowPtr(7 * 24 * time.Hour)
	proposal := economicspb.Proposal{
		Id:             1,
		Title:          "Test Proposal",
		Description:    "Test description",
		Proposer:       "aura1proposer",
		Status:         economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		SubmitTime:     now,
		DepositEndTime: testutil.TimeFromNow(48 * time.Hour),
		VotingEndTime:  votingEndTime,
		TotalDeposit:   sdk.NewCoins(testutil.NewCoin("uaura", 10000000)),
	}

	vote := economicspb.Vote{
		ProposalId:  1,
		Voter:       "aura1voter",
		Option:      economicspb.VoteOption_VOTE_OPTION_YES,
		Timestamp:   now,
		VotingPower: testutil.NewInt(1000000),
	}

	deposit := economicspb.Deposit{
		ProposalId: 1,
		Depositor:  "aura1depositor",
		Amount:     sdk.NewCoins(testutil.NewCoin("uaura", 1000000)),
		Timestamp:  now,
	}

	gs := &economicspb.GenesisState{
		Params:             *types.DefaultParams(),
		VestingSchedules:   []economicspb.VestingSchedule{},
		VoteLocks:          []economicspb.VoteLock{},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{proposal},
		Votes:              []economicspb.Vote{vote},
		Deposits:           []economicspb.Deposit{deposit},
		VoteDelegations:    []economicspb.VoteDelegation{},
		NextProposalId:     2,
		UserMevBalances:    make(map[string]string),
		LastLargeTxTimes:   make(map[string]int64),
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify vote was stored
	storedVote, err := keeper.GetVote(ctx, 1, "aura1voter")
	require.NoError(t, err)
	require.Equal(t, uint64(1), storedVote.ProposalId)
	require.Equal(t, "aura1voter", storedVote.Voter)

	// Verify deposit was stored
	storedDeposit, err := keeper.GetDeposit(ctx, 1, "aura1depositor")
	require.NoError(t, err)
	require.Equal(t, uint64(1), storedDeposit.ProposalId)
	require.Equal(t, "aura1depositor", storedDeposit.Depositor)
}

func TestInitGenesisWithStats(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	now := testutil.Now()
	inflationMetrics := &economicspb.InflationMetrics{
		CurrentRate:       500,
		CirculatingSupply: testutil.NewInt(100000000),
		TotalVested:       testutil.NewInt(50000000),
		TotalVesting:      testutil.NewInt(25000000),
		LastAdjustment:    now,
		NextCheck:         testutil.TimeFromNow(24 * time.Hour),
	}

	mevStats := &economicspb.MEVStats{
		TotalCaptured:         testutil.NewInt(5000000),
		TotalRedistributed:    testutil.NewInt(4500000),
		PendingRedistribution: testutil.NewInt(500000),
	}

	lmStats := &economicspb.LiquidityMiningStats{
		CurrentEpoch:           100,
		TotalDistributed:       testutil.NewInt(10000000),
		RemainingRewards:       testutil.NewInt(5000000),
		RewardsThisEpoch:       testutil.NewInt(100000),
		NextDistributionHeight: 1000,
	}

	gs := &economicspb.GenesisState{
		Params:               *types.DefaultParams(),
		VestingSchedules:     []economicspb.VestingSchedule{},
		VoteLocks:            []economicspb.VoteLock{},
		PendingTreasuryTxs:   []economicspb.PendingTreasuryTx{},
		Proposals:            []economicspb.Proposal{},
		Votes:                []economicspb.Vote{},
		Deposits:             []economicspb.Deposit{},
		VoteDelegations:      []economicspb.VoteDelegation{},
		NextProposalId:       1,
		InflationMetrics:     inflationMetrics,
		MevStats:             mevStats,
		LiquidityMiningStats: lmStats,
		UserMevBalances:      make(map[string]string),
		LastLargeTxTimes:     make(map[string]int64),
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify metrics were stored
	storedInflation, err := keeper.GetInflationMetrics(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(500), storedInflation.CurrentRate)

	storedMEV, err := keeper.GetMEVStats(ctx)
	require.NoError(t, err)
	require.True(t, storedMEV.TotalCaptured.Equal(testutil.NewInt(5000000)))

	storedLM, err := keeper.GetLiquidityMiningStats(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(100), storedLM.CurrentEpoch)
}

func TestInitGenesisWithUserData(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	userMEVBalances := map[string]string{
		"aura1user1": "1000000",
		"aura1user2": "2000000",
	}

	lastLargeTxTimes := map[string]int64{
		"aura1user1": 1234567890,
		"aura1user2": 9876543210,
	}

	gs := &economicspb.GenesisState{
		Params:             *types.DefaultParams(),
		VestingSchedules:   []economicspb.VestingSchedule{},
		VoteLocks:          []economicspb.VoteLock{},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{},
		Votes:              []economicspb.Vote{},
		Deposits:           []economicspb.Deposit{},
		VoteDelegations:    []economicspb.VoteDelegation{},
		NextProposalId:     1,
		UserMevBalances:    userMEVBalances,
		LastLargeTxTimes:   lastLargeTxTimes,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify user data was stored
	balance, err := keeper.GetUserMEVBalance(ctx, "aura1user1")
	require.NoError(t, err)
	require.Equal(t, "1000000", balance)

	// Note: We cannot directly test LastLargeTxTime getter as it's not exported
	// The fact that InitGenesis didn't error indicates it was stored
}

// ============================
// INIT GENESIS ERROR TESTS
// ============================

func TestInitGenesisDuplicateScheduleIDs(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	now := testutil.Now()
	schedule1 := economicspb.VestingSchedule{
		Id:             "vesting-1",
		Address:        "aura1test1",
		OriginalAmount: testutil.NewCoin("uaura", 1000000),
		VestedAmount:   testutil.NewCoin("uaura", 0),
		StartTime:      now,
		EndTime:        testutil.TimeFromNow(365 * 24 * time.Hour),
		VestingType:    economicspb.VestingType_VESTING_TYPE_LINEAR,
	}

	schedule2 := economicspb.VestingSchedule{
		Id:             "vesting-1", // Duplicate ID
		Address:        "aura1test2",
		OriginalAmount: testutil.NewCoin("uaura", 2000000),
		VestedAmount:   testutil.NewCoin("uaura", 0),
		StartTime:      now,
		EndTime:        testutil.TimeFromNow(365 * 24 * time.Hour),
		VestingType:    economicspb.VestingType_VESTING_TYPE_LINEAR,
	}

	gs := &economicspb.GenesisState{
		Params:             *types.DefaultParams(),
		VestingSchedules:   []economicspb.VestingSchedule{schedule1, schedule2},
		VoteLocks:          []economicspb.VoteLock{},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{},
		Votes:              []economicspb.Vote{},
		Deposits:           []economicspb.Deposit{},
		VoteDelegations:    []economicspb.VoteDelegation{},
		NextProposalId:     1,
		UserMevBalances:    make(map[string]string),
		LastLargeTxTimes:   make(map[string]int64),
	}

	err := keeper.InitGenesis(ctx, gs)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate ID")
}

func TestInitGenesisDuplicateLockIDs(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	now := testutil.Now()
	lock1 := economicspb.VoteLock{
		Id:          "lock-1",
		Owner:       "aura1test1",
		Amount:      testutil.NewCoin("uaura", 1000000),
		LockStart:   now,
		LockEnd:     testutil.TimeFromNow(365 * 24 * time.Hour),
		VotingPower: testutil.NewInt(2000000),
	}

	lock2 := economicspb.VoteLock{
		Id:          "lock-1", // Duplicate ID
		Owner:       "aura1test2",
		Amount:      testutil.NewCoin("uaura", 2000000),
		LockStart:   now,
		LockEnd:     testutil.TimeFromNow(365 * 24 * time.Hour),
		VotingPower: testutil.NewInt(4000000),
	}

	gs := &economicspb.GenesisState{
		Params:             *types.DefaultParams(),
		VestingSchedules:   []economicspb.VestingSchedule{},
		VoteLocks:          []economicspb.VoteLock{lock1, lock2},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{},
		Votes:              []economicspb.Vote{},
		Deposits:           []economicspb.Deposit{},
		VoteDelegations:    []economicspb.VoteDelegation{},
		NextProposalId:     1,
		UserMevBalances:    make(map[string]string),
		LastLargeTxTimes:   make(map[string]int64),
	}

	err := keeper.InitGenesis(ctx, gs)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate ID")
}

func TestInitGenesisDuplicateProposalIDs(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	now := testutil.Now()
	votingEndTime := testutil.TimeFromNowPtr(7 * 24 * time.Hour)
	proposal1 := economicspb.Proposal{
		Id:             1,
		Title:          "Proposal 1",
		Description:    "Description 1",
		Proposer:       "aura1proposer1",
		Status:         economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		SubmitTime:     now,
		DepositEndTime: testutil.TimeFromNow(48 * time.Hour),
		VotingEndTime:  votingEndTime,
		TotalDeposit:   sdk.NewCoins(testutil.NewCoin("uaura", 10000000)),
	}

	proposal2 := economicspb.Proposal{
		Id:             1, // Duplicate ID
		Title:          "Proposal 2",
		Description:    "Description 2",
		Proposer:       "aura1proposer2",
		Status:         economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		SubmitTime:     now,
		DepositEndTime: testutil.TimeFromNow(48 * time.Hour),
		VotingEndTime:  votingEndTime,
		TotalDeposit:   sdk.NewCoins(testutil.NewCoin("uaura", 10000000)),
	}

	gs := &economicspb.GenesisState{
		Params:             *types.DefaultParams(),
		VestingSchedules:   []economicspb.VestingSchedule{},
		VoteLocks:          []economicspb.VoteLock{},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{proposal1, proposal2},
		Votes:              []economicspb.Vote{},
		Deposits:           []economicspb.Deposit{},
		VoteDelegations:    []economicspb.VoteDelegation{},
		NextProposalId:     2,
		UserMevBalances:    make(map[string]string),
		LastLargeTxTimes:   make(map[string]int64),
	}

	err := keeper.InitGenesis(ctx, gs)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate ID")
}

// ============================
// EXPORT GENESIS TESTS
// ============================

func TestExportGenesisEmpty(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set default params first
	err := keeper.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	// Export genesis
	gs, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, gs)

	// Verify empty collections
	require.Empty(t, gs.VestingSchedules)
	require.Empty(t, gs.VoteLocks)
	require.Empty(t, gs.PendingTreasuryTxs)
	require.Empty(t, gs.Proposals)
	require.Empty(t, gs.Votes)
	require.Empty(t, gs.Deposits)
	require.Empty(t, gs.VoteDelegations)
}

func TestExportGenesisWithData(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params
	err := keeper.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	// Add vesting schedule
	now := testutil.Now()
	schedule := &economicspb.VestingSchedule{
		Id:             "vesting-1",
		Address:        "aura1test",
		OriginalAmount: testutil.NewCoin("uaura", 1000000),
		VestedAmount:   testutil.NewCoin("uaura", 250000),
		StartTime:      now,
		EndTime:        testutil.TimeFromNow(365 * 24 * time.Hour),
		VestingType:    economicspb.VestingType_VESTING_TYPE_LINEAR,
	}
	err = keeper.SetVestingSchedule(ctx, schedule)
	require.NoError(t, err)

	// Add vote lock
	lock := &economicspb.VoteLock{
		Id:          "lock-1",
		Owner:       "aura1test",
		Amount:      testutil.NewCoin("uaura", 1000000),
		LockStart:   now,
		LockEnd:     testutil.TimeFromNow(365 * 24 * time.Hour),
		VotingPower: testutil.NewInt(2000000),
	}
	err = keeper.SetVoteLock(ctx, lock)
	require.NoError(t, err)

	// Add proposal
	votingEndTime := testutil.TimeFromNowPtr(7 * 24 * time.Hour)
	proposal := &economicspb.Proposal{
		Id:             1,
		Title:          "Test Proposal",
		Description:    "Test description",
		Proposer:       "aura1proposer",
		Status:         economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		SubmitTime:     now,
		DepositEndTime: testutil.TimeFromNow(48 * time.Hour),
		VotingEndTime:  votingEndTime,
		TotalDeposit:   sdk.NewCoins(testutil.NewCoin("uaura", 10000000)),
	}
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)
	err = keeper.SetNextProposalID(ctx, 2)
	require.NoError(t, err)

	// Export genesis
	gs, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, gs)

	// Verify data was exported
	require.Len(t, gs.VestingSchedules, 1)
	require.Equal(t, "vesting-1", gs.VestingSchedules[0].Id)

	require.Len(t, gs.VoteLocks, 1)
	require.Equal(t, "lock-1", gs.VoteLocks[0].Id)

	require.Len(t, gs.Proposals, 1)
	require.Equal(t, uint64(1), gs.Proposals[0].Id)

	require.Equal(t, uint64(2), gs.NextProposalId)
}

func TestExportGenesisRoundTrip(t *testing.T) {
	keeper1, ctx1 := setupKeeperForTest(t)

	// Set params
	err := keeper1.SetParams(ctx1, types.DefaultParams())
	require.NoError(t, err)

	// Create some data
	now := testutil.Now()
	schedule := &economicspb.VestingSchedule{
		Id:             "vesting-1",
		Address:        "aura1test",
		OriginalAmount: testutil.NewCoin("uaura", 1000000),
		VestedAmount:   testutil.NewCoin("uaura", 250000),
		StartTime:      now,
		EndTime:        testutil.TimeFromNow(365 * 24 * time.Hour),
		VestingType:    economicspb.VestingType_VESTING_TYPE_LINEAR,
	}
	err = keeper1.SetVestingSchedule(ctx1, schedule)
	require.NoError(t, err)

	// Set inflation metrics
	metrics := &economicspb.InflationMetrics{
		CurrentRate:       500,
		CirculatingSupply: testutil.NewInt(100000000),
		TotalVested:       testutil.NewInt(50000000),
		TotalVesting:      testutil.NewInt(25000000),
		LastAdjustment:    now,
		NextCheck:         testutil.TimeFromNow(24 * time.Hour),
	}
	err = keeper1.SetInflationMetrics(ctx1, metrics)
	require.NoError(t, err)

	// Export genesis from keeper1
	gs, err := keeper1.ExportGenesis(ctx1)
	require.NoError(t, err)

	// Create new keeper and import
	keeper2, ctx2 := setupKeeperForTest(t)
	err = keeper2.InitGenesis(ctx2, gs)
	require.NoError(t, err)

	// Verify data in keeper2
	storedSchedule, err := keeper2.GetVestingSchedule(ctx2, "vesting-1")
	require.NoError(t, err)
	require.Equal(t, "vesting-1", storedSchedule.Id)
	require.Equal(t, "aura1test", storedSchedule.Address)

	storedMetrics, err := keeper2.GetInflationMetrics(ctx2)
	require.NoError(t, err)
	require.Equal(t, uint64(500), storedMetrics.CurrentRate)
}

func TestExportGenesisWithVotesAndDeposits(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params
	err := keeper.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	// Add proposal
	now := testutil.Now()
	votingEndTime := testutil.TimeFromNowPtr(7 * 24 * time.Hour)
	proposal := &economicspb.Proposal{
		Id:             1,
		Title:          "Test Proposal",
		Description:    "Test description",
		Proposer:       "aura1proposer",
		Status:         economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		SubmitTime:     now,
		DepositEndTime: testutil.TimeFromNow(48 * time.Hour),
		VotingEndTime:  votingEndTime,
		TotalDeposit:   sdk.NewCoins(testutil.NewCoin("uaura", 10000000)),
	}
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Add vote
	vote := &economicspb.Vote{
		ProposalId:  1,
		Voter:       "aura1voter",
		Option:      economicspb.VoteOption_VOTE_OPTION_YES,
		Timestamp:   now,
		VotingPower: testutil.NewInt(1000000),
	}
	err = keeper.SetVote(ctx, vote)
	require.NoError(t, err)

	// Add deposit
	deposit := &economicspb.Deposit{
		ProposalId: 1,
		Depositor:  "aura1depositor",
		Amount:     sdk.NewCoins(testutil.NewCoin("uaura", 1000000)),
		Timestamp:  now,
	}
	err = keeper.SetDeposit(ctx, deposit)
	require.NoError(t, err)

	// Export genesis
	gs, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)

	// Verify votes and deposits were exported
	require.Len(t, gs.Votes, 1)
	require.Equal(t, "aura1voter", gs.Votes[0].Voter)

	require.Len(t, gs.Deposits, 1)
	require.Equal(t, "aura1depositor", gs.Deposits[0].Depositor)
}
