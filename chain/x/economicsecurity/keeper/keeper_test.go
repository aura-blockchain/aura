package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestKeeperFunctionality(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Test keeper initialization
	require.NotNil(t, k)
	require.NotNil(t, ctx)

	// Test authority
	authority := k.GetAuthority()
	require.NotEmpty(t, authority)
	require.Equal(t, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", authority)

	// Test params operations
	params := k.GetParams()
	require.NotNil(t, params)
	require.NotNil(t, params.Tokenomics)
	require.NotNil(t, params.WhaleProtection)

	// Modify and set params
	newParams := params
	newParams.Tokenomics.InflationRate = 1500 // 15%
	err := k.SetParams(newParams)
	require.NoError(t, err)

	// Verify params were updated
	updatedParams := k.GetParams()
	require.Equal(t, uint64(1500), updatedParams.Tokenomics.InflationRate)

	// Test vesting schedule operations
	schedule := &types.VestingSchedule{
		ScheduleId:          "test-schedule-1",
		BeneficiaryAddress:  "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		TotalAmount:         "10000000",
		VestedAmount:        "1000000",
		VestingDuration:     31536000,
		CliffDuration:       7776000,
		StartTime:           timestamppb.New(time.Now()),
		VestingType:         types.VestingTypeLinear,
	}

	// Set vesting schedule
	err = k.SetVestingSchedule(ctx, schedule)
	require.NoError(t, err)

	// Get vesting schedule
	retrieved, err := k.GetVestingSchedule(ctx, "test-schedule-1")
	require.NoError(t, err)
	require.Equal(t, schedule.ScheduleId, retrieved.ScheduleId)
	require.Equal(t, schedule.TotalAmount, retrieved.TotalAmount)

	// Delete vesting schedule
	err = k.DeleteVestingSchedule(ctx, "test-schedule-1")
	require.NoError(t, err)

	// Verify deletion
	_, err = k.GetVestingSchedule(ctx, "test-schedule-1")
	require.Error(t, err)
	require.Equal(t, types.ErrVestingScheduleNotFound, err)

	// Test vote lock operations
	lock := &types.VoteLock{
		LockId:      "test-lock-1",
		Owner:       "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		Amount:      "1000000",
		LockStart:   timestamppb.New(time.Now()),
		LockEnd:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
		VotingPower: "1500000",
	}

	// Set vote lock
	err = k.SetVoteLock(ctx, lock)
	require.NoError(t, err)

	// Get vote lock
	retrievedLock, err := k.GetVoteLock(ctx, "test-lock-1")
	require.NoError(t, err)
	require.Equal(t, lock.LockId, retrievedLock.LockId)
	require.Equal(t, lock.Amount, retrievedLock.Amount)

	// Delete vote lock
	err = k.DeleteVoteLock(ctx, "test-lock-1")
	require.NoError(t, err)

	// Verify deletion
	_, err = k.GetVoteLock(ctx, "test-lock-1")
	require.Error(t, err)
	require.Equal(t, types.ErrVoteLockNotFound, err)

	// Test pending treasury tx operations
	tx := &types.PendingTreasuryTx{
		TxId:         "test-tx-1",
		Recipient:    "aura1w3jhxapjta047h6lta047h6lta047h6l42n9lg",
		Amount:       "500000",
		Proposer:     "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		Signatures:   []string{"sig1", "sig2"},
		CreatedAt:    timestamppb.New(time.Now()),
		ExecutableAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
	}

	// Set pending treasury tx
	err = k.SetPendingTreasuryTx(ctx, tx)
	require.NoError(t, err)

	// Get pending treasury tx
	retrievedTx, err := k.GetPendingTreasuryTx(ctx, "test-tx-1")
	require.NoError(t, err)
	require.Equal(t, tx.TxId, retrievedTx.TxId)
	require.Equal(t, tx.Amount, retrievedTx.Amount)

	// Delete pending treasury tx
	err = k.DeletePendingTreasuryTx(ctx, "test-tx-1")
	require.NoError(t, err)

	// Verify deletion
	_, err = k.GetPendingTreasuryTx(ctx, "test-tx-1")
	require.Error(t, err)
	require.Equal(t, types.ErrTxNotFound, err)

	// Test inflation alert operations
	alert := &types.InflationAlert{
		AlertId:              "test-alert-1",
		AlertType:            types.InflationAlertTypeRapidChange,
		Severity:             types.AlertSeverityCritical,
		CurrentInflationRate: 2000,
		TriggeredAt:          timestamppb.New(time.Now()),
		Message:              "Critical inflation spike detected",
	}

	// Set inflation alert
	err = k.SetInflationAlert(ctx, alert)
	require.NoError(t, err)

	// Get inflation alert
	retrievedAlert, err := k.GetInflationAlert(ctx, "test-alert-1")
	require.NoError(t, err)
	require.Equal(t, alert.AlertId, retrievedAlert.AlertId)
	require.Equal(t, alert.CurrentInflationRate, retrievedAlert.CurrentInflationRate)

	// Test large tx record operations
	record := &types.LargeTxRecord{
		TxHash:             "test-hash-1",
		Sender:             "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		Recipient:          "aura1w3jhxapjta047h6lta047h6lta047h6l42n9lg",
		Amount:             "5000000",
		Timestamp:          timestamppb.New(time.Now()),
		BlockHeight:        12345,
		PercentageOfSupply: 500,
	}

	// Set large tx record
	err = k.SetLargeTxRecord(ctx, record)
	require.NoError(t, err)

	// Get large tx record
	retrievedRecord, err := k.GetLargeTxRecord(ctx, "test-hash-1")
	require.NoError(t, err)
	require.Equal(t, record.TxHash, retrievedRecord.TxHash)
	require.Equal(t, record.Amount, retrievedRecord.Amount)

	// Test user MEV balance operations
	err = k.SetUserMEVBalance(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", "75000")
	require.NoError(t, err)

	balance, err := k.GetUserMEVBalance(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr")
	require.NoError(t, err)
	require.Equal(t, "75000", balance)

	// Test total MEV pending operations
	err = k.SetTotalMEVPending(ctx, "75000")
	require.NoError(t, err)

	totalPending, err := k.GetTotalMEVPending(ctx)
	require.NoError(t, err)
	require.Equal(t, "75000", totalPending)

	// Test current height operations
	err = k.SetCurrentHeight(ctx, 12345)
	require.NoError(t, err)

	height, err := k.GetCurrentHeight(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(12345), height)

	// Test current time operations
	currentTime := time.Now().Unix()
	err = k.SetCurrentTime(ctx, currentTime)
	require.NoError(t, err)

	retrievedTime, err := k.GetCurrentTime(ctx)
	require.NoError(t, err)
	require.Equal(t, currentTime, retrievedTime)

	// Test last large tx time operations
	lastTxTime := time.Now().Unix()
	err = k.SetLastLargeTxTime(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", lastTxTime)
	require.NoError(t, err)

	retrievedLastTxTime, err := k.GetLastLargeTxTime(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr")
	require.NoError(t, err)
	require.Equal(t, lastTxTime, retrievedLastTxTime)

	// Test address holding operations
	err = k.SetAddressHolding(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", "1000000")
	require.NoError(t, err)

	holding, err := k.GetAddressHolding(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr")
	require.NoError(t, err)
	require.Equal(t, "1000000", holding)

	// Test previous inflation operations
	err = k.SetPreviousInflation(ctx, 1000)
	require.NoError(t, err)

	prevInflation, err := k.GetPreviousInflation(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1000), prevInflation)

	// Test user vesting index operations
	err = k.AddUserVestingSchedule(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", "schedule-1")
	require.NoError(t, err)

	err = k.AddUserVestingSchedule(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", "schedule-2")
	require.NoError(t, err)

	scheduleIDs, err := k.GetUserVestingIndex(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr")
	require.NoError(t, err)
	require.Len(t, scheduleIDs, 2)
	require.Contains(t, scheduleIDs, "schedule-1")
	require.Contains(t, scheduleIDs, "schedule-2")

	// Test user vote lock index operations
	err = k.AddUserVoteLock(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", "lock-1")
	require.NoError(t, err)

	err = k.AddUserVoteLock(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", "lock-2")
	require.NoError(t, err)

	lockIDs, err := k.GetUserVoteLockIndex(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr")
	require.NoError(t, err)
	require.Len(t, lockIDs, 2)
	require.Contains(t, lockIDs, "lock-1")
	require.Contains(t, lockIDs, "lock-2")
}
