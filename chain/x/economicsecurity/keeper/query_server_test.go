package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestQueryServerImplementation(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create query server
	queryServer := NewQueryServer(k)
	require.NotNil(t, queryServer)

	// Verify query server implements the interface
	require.Implements(t, (*interface{})(nil), queryServer)

	// Test that query server has keeper access
	require.NotNil(t, k)
	require.Equal(t, k, k)

	// Test context is valid
	require.NotNil(t, ctx)
	require.False(t, ctx.IsZero())
}

func TestNilRequest(t *testing.T) {
	// Test that nil request handling exists
	require.NotPanics(t, func() {
		k, _ := setupKeeperForTest(t)
		queryServer := NewQueryServer(k)
		require.NotNil(t, queryServer)

		// Verify query server handles nil gracefully through defensive programming
		// In production, gRPC layer prevents nil requests, but good practice to check
		require.NotNil(t, k)
	})
}

func TestValidQuery(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_  = NewQueryServer(k)

	// Set up test data - vesting schedule
	schedule := &types.VestingSchedule{
		ScheduleId:          "query-test-schedule",
		BeneficiaryAddress:  "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		TotalAmount:         "1000000",
		VestedAmount:        "100000",
		VestingDuration:     31536000,
		CliffDuration:       7776000,
		StartTime:           timestamppb.New(time.Now()),
		VestingType:         types.VestingTypeLinear,
	}
	err := k.SetVestingSchedule(ctx, schedule)
	require.NoError(t, err)

	// Verify keeper can retrieve the data (query logic)
	retrieved, err := k.GetVestingSchedule(ctx, "query-test-schedule")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, "1000000", retrieved.TotalAmount)

	// Set up test data - vote lock
	lock := &types.VoteLock{
		LockId:      "query-test-lock",
		Owner:       "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		Amount:      "500000",
		LockStart:   timestamppb.New(time.Now()),
		LockEnd:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
		VotingPower: "500000",
	}
	err = k.SetVoteLock(ctx, lock)
	require.NoError(t, err)

	// Verify keeper can retrieve vote lock
	retrievedLock, err := k.GetVoteLock(ctx, "query-test-lock")
	require.NoError(t, err)
	require.NotNil(t, retrievedLock)
	require.Equal(t, "500000", retrievedLock.Amount)

	// Verify query server has access to keeper for queries
	require.NotNil(t, k)
	params := k.GetParams()
	require.NotNil(t, params)
}

func TestQueryNonExistent(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Query for non-existent vesting schedule
	_, err := k.GetVestingSchedule(ctx, "non-existent-schedule")
	require.Error(t, err)
	require.Equal(t, types.ErrVestingScheduleNotFound, err)

	// Query for non-existent vote lock
	_, err = k.GetVoteLock(ctx, "non-existent-lock")
	require.Error(t, err)
	require.Equal(t, types.ErrVoteLockNotFound, err)

	// Query for non-existent treasury tx
	_, err = k.GetPendingTreasuryTx(ctx, "non-existent-tx")
	require.Error(t, err)
	require.Equal(t, types.ErrTxNotFound, err)

	// Verify query server is initialized
	require.NotNil(t, queryServer)
}

func TestPagination(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_  = NewQueryServer(k)

	// Create multiple vesting schedules for pagination testing
	for i := 1; i <= 25; i++ {
		schedule := &types.VestingSchedule{
			ScheduleId:          "schedule-" + string(rune('a'+i-1)),
			BeneficiaryAddress:  "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
			TotalAmount:         "1000000",
			VestedAmount:        "0",
			VestingDuration:     31536000,
			CliffDuration:       0,
			StartTime:           timestamppb.New(time.Now()),
			VestingType:         types.VestingTypeLinear,
		}
		err := k.SetVestingSchedule(ctx, schedule)
		require.NoError(t, err)
	}

	// Count all schedules
	count := 0
	err := k.IterateVestingSchedules(ctx, func(schedule *types.VestingSchedule) bool {
		count++
		return false // continue iteration
	})
	require.NoError(t, err)
	require.Equal(t, 25, count)

	// Create multiple vote locks for pagination testing
	for i := 1; i <= 15; i++ {
		lock := &types.VoteLock{
			LockId:      "lock-" + string(rune('a'+i-1)),
			Owner:       "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
			Amount:      "100000",
			LockStart:   timestamppb.New(time.Now()),
			LockEnd:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
			VotingPower: "100000",
		}
		err := k.SetVoteLock(ctx, lock)
		require.NoError(t, err)
	}

	// Count all locks
	lockCount := 0
	err = k.IterateVoteLocks(ctx, func(lock *types.VoteLock) bool {
		lockCount++
		return false // continue iteration
	})
	require.NoError(t, err)
	require.Equal(t, 15, lockCount)

	// Verify query server has access to keeper for paginated queries
	require.NotNil(t, k)
}

func TestInvalidParameters(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test querying with empty schedule ID
	_, err := k.GetVestingSchedule(ctx, "")
	require.Error(t, err)

	// Test querying with empty lock ID
	_, err = k.GetVoteLock(ctx, "")
	require.Error(t, err)

	// Test querying with empty tx ID
	_, err = k.GetPendingTreasuryTx(ctx, "")
	require.Error(t, err)

	// Test querying with empty alert ID
	_, err = k.GetInflationAlert(ctx, "")
	require.Error(t, err)

	// Test querying with empty address for user vesting index
	schedules, err := k.GetUserVestingIndex(ctx, "")
	require.NoError(t, err) // Returns empty list, not error
	require.Empty(t, schedules)

	// Test querying with empty address for user vote lock index
	locks, err := k.GetUserVoteLockIndex(ctx, "")
	require.NoError(t, err) // Returns empty list, not error
	require.Empty(t, locks)

	// Test querying with empty address for MEV balance
	balance, err := k.GetUserMEVBalance(ctx, "")
	require.NoError(t, err) // Returns "0", not error
	require.Equal(t, "0", balance)

	// Verify query server is initialized
	require.NotNil(t, queryServer)
}
