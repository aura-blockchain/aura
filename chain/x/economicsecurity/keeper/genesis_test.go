package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestInitGenesis(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create comprehensive genesis state
	genesis := types.GenesisState{
		Params: types.DefaultParams(),
		VestingSchedules: []*types.VestingSchedule{
			{
				ScheduleId:          "team-vesting-1",
				BeneficiaryAddress:  "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				TotalAmount:         "10000000",
				VestedAmount:        "1000000",
				VestingDuration:     31536000, // 1 year
				CliffDuration:       7776000,  // 90 days
				StartTime:           timestamppb.New(time.Now()),
				VestingType:         types.VestingTypeLinear,
			},
			{
				ScheduleId:          "advisor-vesting-1",
				BeneficiaryAddress:  "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
				TotalAmount:         "5000000",
				VestedAmount:        "0",
				VestingDuration:     15768000, // 6 months
				CliffDuration:       2592000,  // 30 days
				StartTime:           timestamppb.New(time.Now()),
				VestingType:         types.VestingTypeCliffThenLinear,
			},
		},
		VoteLocks: []*types.VoteLock{
			{
				LockId:      "vote-lock-1",
				Owner:       "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				Amount:      "1000000",
				LockStart:   timestamppb.New(time.Now()),
				LockEnd:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
				VotingPower: "1500000", // 1.5x multiplier for long lock
			},
		},
		PendingTreasuryTxs: []*types.PendingTreasuryTx{
			{
				TxId:         "treasury-tx-1",
				Recipient:    "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
				Amount:       "100000",
				Proposer:     "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				Signatures:   []string{"sig1", "sig2"},
				CreatedAt:    timestamppb.New(time.Now()),
				ExecutableAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
			},
		},
		InflationAlerts: []*types.InflationAlert{
			{
				AlertId:              "alert-1",
				AlertType:            types.InflationAlertTypeRapidChange,
				Severity:             types.AlertSeverityWarning,
				CurrentInflationRate: 1200,
				TriggeredAt:          timestamppb.New(time.Now()),
				Message:              "Inflation spike detected",
			},
		},
		LargeTxRecords: []*types.LargeTxRecord{
			{
				TxHash:             "hash-1",
				Sender:             "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				Recipient:          "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
				Amount:             "5000000",
				Timestamp:          timestamppb.New(time.Now()),
				BlockHeight:        12345,
				PercentageOfSupply: 250, // 2.5%
			},
		},
		LastLargeTxTimes: map[string]int64{
			"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn": time.Now().Unix(),
		},
		UserMevBalances: map[string]string{
			"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn": "75000",
			"aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh": "25000",
		},
	}

	// Initialize genesis
	err := k.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify vesting schedules were loaded
	schedule1, err := k.GetVestingSchedule(ctx, "team-vesting-1")
	require.NoError(t, err)
	require.Equal(t, "10000000", schedule1.TotalAmount)
	require.Equal(t, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", schedule1.BeneficiaryAddress)

	schedule2, err := k.GetVestingSchedule(ctx, "advisor-vesting-1")
	require.NoError(t, err)
	require.Equal(t, "5000000", schedule2.TotalAmount)

	// Verify vote locks were loaded
	lock, err := k.GetVoteLock(ctx, "vote-lock-1")
	require.NoError(t, err)
	require.Equal(t, "1000000", lock.Amount)
	require.Equal(t, "1500000", lock.VotingPower)

	// Verify treasury tx was loaded
	tx, err := k.GetPendingTreasuryTx(ctx, "treasury-tx-1")
	require.NoError(t, err)
	require.Equal(t, "100000", tx.Amount)
	require.Len(t, tx.Signatures, 2)

	// Verify inflation alert was loaded
	alert, err := k.GetInflationAlert(ctx, "alert-1")
	require.NoError(t, err)
	require.Equal(t, types.InflationAlertTypeRapidChange, alert.AlertType)
	require.Equal(t, uint64(1200), alert.CurrentInflationRate)

	// Verify large tx record was loaded
	record, err := k.GetLargeTxRecord(ctx, "hash-1")
	require.NoError(t, err)
	require.Equal(t, "5000000", record.Amount)
	require.Equal(t, uint64(250), record.PercentageOfSupply)

	// Verify last large tx times were loaded
	lastTime, err := k.GetLastLargeTxTime(ctx, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")
	require.NoError(t, err)
	require.Greater(t, lastTime, int64(0))

	// Verify MEV balances were loaded
	mevBalance1, err := k.GetUserMEVBalance(ctx, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")
	require.NoError(t, err)
	require.Equal(t, "75000", mevBalance1)

	mevBalance2, err := k.GetUserMEVBalance(ctx, "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh")
	require.NoError(t, err)
	require.Equal(t, "25000", mevBalance2)

	// Verify params were set
	params := k.GetParams()
	require.NotNil(t, params.Tokenomics)
}

func TestExportGenesis(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set up test data
	schedule := &types.VestingSchedule{
		ScheduleId:          "export-test-schedule",
		BeneficiaryAddress:  "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		TotalAmount:         "1000000",
		VestedAmount:        "100000",
		VestingDuration:     31536000,
		CliffDuration:       7776000,
		StartTime:           timestamppb.New(time.Now()),
		VestingType:         types.VestingTypeLinear,
	}
	err := k.SetVestingSchedule(ctx, schedule)
	require.NoError(t, err)

	lock := &types.VoteLock{
		LockId:      "export-test-lock",
		Owner:       "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		Amount:      "500000",
		LockStart:   timestamppb.New(time.Now()),
		LockEnd:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
		VotingPower: "500000",
	}
	err = k.SetVoteLock(ctx, lock)
	require.NoError(t, err)

	tx := &types.PendingTreasuryTx{
		TxId:         "export-test-tx",
		Recipient:    "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
		Amount:       "100000",
		Proposer:     "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		Signatures:   []string{},
		CreatedAt:    timestamppb.New(time.Now()),
		ExecutableAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
	}
	err = k.SetPendingTreasuryTx(ctx, tx)
	require.NoError(t, err)

	alert := &types.InflationAlert{
		AlertId:              "export-test-alert",
		AlertType:            types.InflationAlertTypeBelowTarget,
		Severity:             types.AlertSeverityInfo,
		CurrentInflationRate: 800,
		TriggeredAt:          timestamppb.New(time.Now()),
		Message:              "Test alert",
	}
	err = k.SetInflationAlert(ctx, alert)
	require.NoError(t, err)

	record := &types.LargeTxRecord{
		TxHash:             "export-test-hash",
		Sender:             "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		Recipient:          "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
		Amount:             "1000000",
		Timestamp:          timestamppb.New(time.Now()),
		BlockHeight:        12345,
		PercentageOfSupply: 100,
	}
	err = k.SetLargeTxRecord(ctx, record)
	require.NoError(t, err)

	err = k.SetUserMEVBalance(ctx, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", "50000")
	require.NoError(t, err)

	// Export genesis
	exported, err := k.ExportGenesis(ctx)
	require.NoError(t, err)

	// Verify exported data
	require.NotNil(t, exported.Params)
	require.Len(t, exported.VestingSchedules, 1)
	require.Equal(t, "export-test-schedule", exported.VestingSchedules[0].ScheduleId)

	require.Len(t, exported.VoteLocks, 1)
	require.Equal(t, "export-test-lock", exported.VoteLocks[0].LockId)

	require.Len(t, exported.PendingTreasuryTxs, 1)
	require.Equal(t, "export-test-tx", exported.PendingTreasuryTxs[0].TxId)

	require.Len(t, exported.InflationAlerts, 1)
	require.Equal(t, "export-test-alert", exported.InflationAlerts[0].AlertId)

	require.Len(t, exported.LargeTxRecords, 1)
	require.Equal(t, "export-test-hash", exported.LargeTxRecords[0].TxHash)

	require.Len(t, exported.UserMevBalances, 1)
	require.Equal(t, "50000", exported.UserMevBalances["aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"])
}

func TestGenesisRoundTrip(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create complete genesis state
	originalGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		VestingSchedules: []*types.VestingSchedule{
			{
				ScheduleId:          "roundtrip-schedule",
				BeneficiaryAddress:  "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				TotalAmount:         "10000000",
				VestedAmount:        "1000000",
				VestingDuration:     31536000,
				CliffDuration:       7776000,
				StartTime:           timestamppb.New(time.Now()),
				VestingType:         types.VestingTypeLinear,
			},
		},
		VoteLocks: []*types.VoteLock{
			{
				LockId:      "roundtrip-lock",
				Owner:       "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				Amount:      "1000000",
				LockStart:   timestamppb.New(time.Now()),
				LockEnd:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
				VotingPower: "1500000",
			},
		},
		PendingTreasuryTxs: []*types.PendingTreasuryTx{
			{
				TxId:         "roundtrip-tx",
				Recipient:    "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
				Amount:       "100000",
				Proposer:     "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				Signatures:   []string{"sig1"},
				CreatedAt:    timestamppb.New(time.Now()),
				ExecutableAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
			},
		},
		InflationAlerts: []*types.InflationAlert{
			{
				AlertId:              "roundtrip-alert",
				AlertType:            types.InflationAlertTypeRapidChange,
				Severity:             types.AlertSeverityCritical,
				CurrentInflationRate: 1500,
				TriggeredAt:          timestamppb.New(time.Now()),
				Message:              "Critical inflation alert",
			},
		},
		LargeTxRecords: []*types.LargeTxRecord{
			{
				TxHash:             "roundtrip-hash",
				Sender:             "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				Recipient:          "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
				Amount:             "5000000",
				Timestamp:          timestamppb.New(time.Now()),
				BlockHeight:        67890,
				PercentageOfSupply: 500,
			},
		},
		LastLargeTxTimes: map[string]int64{
			"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn": time.Now().Unix(),
		},
		UserMevBalances: map[string]string{
			"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn": "123456",
			"aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh": "654321",
		},
	}

	// Initialize with original genesis
	err := k.InitGenesis(ctx, originalGenesis)
	require.NoError(t, err)

	// Export genesis
	exportedGenesis, err := k.ExportGenesis(ctx)
	require.NoError(t, err)

	// Verify all data matches
	require.NotNil(t, exportedGenesis.Params)

	// Verify vesting schedules
	require.Len(t, exportedGenesis.VestingSchedules, 1)
	require.Equal(t, "roundtrip-schedule", exportedGenesis.VestingSchedules[0].ScheduleId)
	require.Equal(t, "10000000", exportedGenesis.VestingSchedules[0].TotalAmount)

	// Verify vote locks
	require.Len(t, exportedGenesis.VoteLocks, 1)
	require.Equal(t, "roundtrip-lock", exportedGenesis.VoteLocks[0].LockId)
	require.Equal(t, "1000000", exportedGenesis.VoteLocks[0].Amount)

	// Verify treasury txs
	require.Len(t, exportedGenesis.PendingTreasuryTxs, 1)
	require.Equal(t, "roundtrip-tx", exportedGenesis.PendingTreasuryTxs[0].TxId)

	// Verify inflation alerts
	require.Len(t, exportedGenesis.InflationAlerts, 1)
	require.Equal(t, "roundtrip-alert", exportedGenesis.InflationAlerts[0].AlertId)

	// Verify large tx records
	require.Len(t, exportedGenesis.LargeTxRecords, 1)
	require.Equal(t, "roundtrip-hash", exportedGenesis.LargeTxRecords[0].TxHash)

	// Verify MEV balances
	require.Len(t, exportedGenesis.UserMevBalances, 2)
	require.Equal(t, "123456", exportedGenesis.UserMevBalances["aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"])
	require.Equal(t, "654321", exportedGenesis.UserMevBalances["aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"])
}

func TestDefaultGenesis(t *testing.T) {
	// Test that default genesis is valid
	require.NotPanics(t, func() {
		genesis := types.DefaultGenesis()
		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)

		// Validate default genesis
		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})
}

func TestInitGenesis_WithMultipleSchedules(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create genesis with multiple schedules for same user
	genesis := types.GenesisState{
		Params: types.DefaultParams(),
		VestingSchedules: []*types.VestingSchedule{
			{
				ScheduleId:          "team-vesting-1",
				BeneficiaryAddress:  "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				TotalAmount:         "5000000",
				VestedAmount:        "500000",
				VestingDuration:     31536000,
				CliffDuration:       7776000,
				StartTime:           timestamppb.New(time.Now()),
				VestingType:         types.VestingTypeLinear,
			},
			{
				ScheduleId:          "team-vesting-2",
				BeneficiaryAddress:  "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				TotalAmount:         "3000000",
				VestedAmount:        "0",
				VestingDuration:     15768000,
				CliffDuration:       2592000,
				StartTime:           timestamppb.New(time.Now().Add(-30 * 24 * time.Hour)),
				VestingType:         types.VestingTypeCliffThenLinear,
			},
			{
				ScheduleId:          "advisor-vesting-1",
				BeneficiaryAddress:  "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
				TotalAmount:         "2000000",
				VestedAmount:        "200000",
				VestingDuration:     15768000,
				CliffDuration:       2592000,
				StartTime:           timestamppb.New(time.Now()),
				VestingType:         types.VestingTypeLinear,
			},
		},
	}

	// Initialize genesis
	err := k.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify all schedules were loaded
	schedule1, err := k.GetVestingSchedule(ctx, "team-vesting-1")
	require.NoError(t, err)
	require.Equal(t, "5000000", schedule1.TotalAmount)

	schedule2, err := k.GetVestingSchedule(ctx, "team-vesting-2")
	require.NoError(t, err)
	require.Equal(t, "3000000", schedule2.TotalAmount)

	schedule3, err := k.GetVestingSchedule(ctx, "advisor-vesting-1")
	require.NoError(t, err)
	require.Equal(t, "2000000", schedule3.TotalAmount)

	// Verify user vesting index has multiple schedules
	scheduleIDs, err := k.GetUserVestingIndex(ctx, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")
	require.NoError(t, err)
	require.Len(t, scheduleIDs, 2)
	require.Contains(t, scheduleIDs, "team-vesting-1")
	require.Contains(t, scheduleIDs, "team-vesting-2")

	// Verify second user has one schedule
	scheduleIDs2, err := k.GetUserVestingIndex(ctx, "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh")
	require.NoError(t, err)
	require.Len(t, scheduleIDs2, 1)
	require.Contains(t, scheduleIDs2, "advisor-vesting-1")
}
