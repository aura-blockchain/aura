package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/params"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInitGenesis(t *testing.T) {
	tests := []struct {
		name    string
		genesis types.GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid genesis with all data",
			genesis: types.GenesisState{
				Params: &types.Params{
					Tokenomics: &types.TokenomicsParams{
						MaxSupply:         "1000000000",
						CirculatingSupply: "100000000",
						InflationRate:     500,
						MaxInflationRate:  2000,
						MinInflationRate:  100,
						BurnRate:          50,
					},
					Vesting: &types.VestingParams{
						MinVestingDuration: durationpb.New(30 * 24 * time.Hour),
						MaxVestingDuration: durationpb.New(365 * 24 * time.Hour),
						CliffDuration:      durationpb.New(90 * 24 * time.Hour),
					},
					Treasury: &types.TreasuryParams{
						MinApprovals:      3,
						ApprovalThreshold: "0.66",
						MaxDailySpend:     "1000000",
					},
					Governance: &types.GovernanceParams{
						VoteLockDuration:   durationpb.New(7 * 24 * time.Hour),
						MinVoteLockAmount:  "1000",
						MaxVoteLockAmount:  "1000000000",
						VotingPowerDecay:   100,
					},
					WhaleProtection: &types.WhaleProtectionParams{
						MaxHoldingPercent:     "5.0",
						MaxTxPercent:          "1.0",
						LargeTxThreshold:      "10000000",
						LargeTxCooldown:       durationpb.New(1 * time.Hour),
						WhaleDetectionEnabled: true,
					},
					MevProtection: &types.MevProtectionParams{
						MevRedistributionEnabled: true,
						MevBurnPercent:           "50",
						MevRewardsPercent:        "50",
					},
					DynamicFees: &types.DynamicFeesParams{
						BaseFee:             "100",
						MaxFee:              "10000",
						FeeAdjustmentFactor: "1.5",
					},
				},
				VestingSchedules: []*types.VestingSchedule{
					{
						ScheduleId:         "vesting-1",
						BeneficiaryAddress: "aura1beneficiary1",
						TotalAmount:        "1000000",
						ReleasedAmount:     "100000",
						StartTime:          timestamppb.Now(),
						EndTime:            timestamppb.New(time.Now().Add(365 * 24 * time.Hour)),
						CliffTime:          timestamppb.New(time.Now().Add(90 * 24 * time.Hour)),
						IsRevocable:        true,
					},
					{
						ScheduleId:         "vesting-2",
						BeneficiaryAddress: "aura1beneficiary2",
						TotalAmount:        "2000000",
						ReleasedAmount:     "0",
						StartTime:          timestamppb.Now(),
						EndTime:            timestamppb.New(time.Now().Add(180 * 24 * time.Hour)),
						IsRevocable:        false,
					},
				},
				VoteLocks: []*types.VoteLock{
					{
						LockId:     "lock-1",
						Owner:      "aura1voter1",
						Amount:     "50000",
						LockedAt:   timestamppb.Now(),
						UnlocksAt:  timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
						ProposalId: "prop-1",
					},
				},
				PendingTreasuryTxs: []*types.PendingTreasuryTx{
					{
						TxId:          "tx-1",
						Recipient:     "aura1recipient",
						Amount:        "100000",
						Purpose:       "Development grant",
						Approvals:     []string{"admin1", "admin2", "admin3"},
						RequiredApprovals: 3,
						SubmittedAt:   timestamppb.Now(),
						ExpiresAt:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
					},
				},
				InflationAlerts: []*types.InflationAlert{
					{
						AlertTime:        timestamppb.Now(),
						CurrentRate:      2100,
						MaxRate:          2000,
						CirculatingSupply: "110000000",
						Message:          "Inflation rate exceeded maximum",
					},
				},
				LargeTxRecords: []*types.LargeTxRecord{
					{
						TxHash:    "hash1",
						From:      "aura1whale",
						Amount:    "15000000",
						Timestamp: timestamppb.Now(),
						Flagged:   true,
					},
				},
				LastLargeTxTimes: map[string]int64{
					"aura1whale": time.Now().Unix() - 3600,
				},
				UserMevBalances: map[string]string{
					"aura1user1": "50000",
					"aura1user2": "75000",
				},
			},
			wantErr: false,
		},
		{
			name: "default genesis",
			genesis: types.GenesisState{
				Params:             nil,
				VestingSchedules:   []*types.VestingSchedule{},
				VoteLocks:          []*types.VoteLock{},
				PendingTreasuryTxs: []*types.PendingTreasuryTx{},
				InflationAlerts:    []*types.InflationAlert{},
				LargeTxRecords:     []*types.LargeTxRecord{},
				LastLargeTxTimes:   make(map[string]int64),
				UserMevBalances:    make(map[string]string),
			},
			wantErr: false,
		},
		{
			name: "invalid genesis - duplicate vesting schedule IDs",
			genesis: types.GenesisState{
				Params: nil,
				VestingSchedules: []*types.VestingSchedule{
					{
						ScheduleId:         "vesting-1",
						BeneficiaryAddress: "aura1beneficiary1",
						TotalAmount:        "1000000",
						ReleasedAmount:     "0",
						StartTime:          timestamppb.Now(),
						EndTime:            timestamppb.New(time.Now().Add(365 * 24 * time.Hour)),
					},
					{
						ScheduleId:         "vesting-1", // Duplicate
						BeneficiaryAddress: "aura1beneficiary2",
						TotalAmount:        "2000000",
						ReleasedAmount:     "0",
						StartTime:          timestamppb.Now(),
						EndTime:            timestamppb.New(time.Now().Add(180 * 24 * time.Hour)),
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate schedule ID",
		},
		{
			name: "invalid genesis - empty vesting schedule ID",
			genesis: types.GenesisState{
				Params: nil,
				VestingSchedules: []*types.VestingSchedule{
					{
						ScheduleId:         "",
						BeneficiaryAddress: "aura1beneficiary",
						TotalAmount:        "1000000",
						StartTime:          timestamppb.Now(),
						EndTime:            timestamppb.New(time.Now().Add(365 * 24 * time.Hour)),
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid schedule ID",
		},
		{
			name: "invalid genesis - duplicate vote lock IDs",
			genesis: types.GenesisState{
				Params:           nil,
				VestingSchedules: []*types.VestingSchedule{},
				VoteLocks: []*types.VoteLock{
					{
						LockId:    "lock-1",
						Owner:     "aura1voter1",
						Amount:    "50000",
						LockedAt:  timestamppb.Now(),
						UnlocksAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
					},
					{
						LockId:    "lock-1", // Duplicate
						Owner:     "aura1voter2",
						Amount:    "60000",
						LockedAt:  timestamppb.Now(),
						UnlocksAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate lock ID",
		},
		{
			name: "invalid genesis - duplicate treasury tx IDs",
			genesis: types.GenesisState{
				Params:           nil,
				VestingSchedules: []*types.VestingSchedule{},
				VoteLocks:        []*types.VoteLock{},
				PendingTreasuryTxs: []*types.PendingTreasuryTx{
					{
						TxId:              "tx-1",
						Recipient:         "aura1recipient1",
						Amount:            "100000",
						Purpose:           "Grant 1",
						RequiredApprovals: 3,
						SubmittedAt:       timestamppb.Now(),
					},
					{
						TxId:              "tx-1", // Duplicate
						Recipient:         "aura1recipient2",
						Amount:            "200000",
						Purpose:           "Grant 2",
						RequiredApprovals: 3,
						SubmittedAt:       timestamppb.Now(),
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate transaction ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate genesis state first
			if tt.wantErr {
				err := types.ValidateGenesis(&tt.genesis)
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			store := params.NewStore(*types.DefaultParams())
			keeper := NewKeeper(store)

			err := keeper.InitGenesis(tt.genesis)
			require.NoError(t, err)

			// Verify params were set
			p := keeper.GetParams()
			require.NotNil(t, p)

			// Verify vesting schedules
			if len(tt.genesis.VestingSchedules) > 0 {
				for _, schedule := range tt.genesis.VestingSchedules {
					retrieved := keeper.GetVestingSchedule(schedule.ScheduleId)
					require.NotNil(t, retrieved)
					require.Equal(t, schedule.BeneficiaryAddress, retrieved.BeneficiaryAddress)
					require.Equal(t, schedule.TotalAmount, retrieved.TotalAmount)
				}
			}

			// Verify vote locks
			if len(tt.genesis.VoteLocks) > 0 {
				for _, lock := range tt.genesis.VoteLocks {
					retrieved := keeper.GetVoteLock(lock.LockId)
					require.NotNil(t, retrieved)
					require.Equal(t, lock.Owner, retrieved.Owner)
					require.Equal(t, lock.Amount, retrieved.Amount)
				}
			}

			// Verify pending treasury txs
			if len(tt.genesis.PendingTreasuryTxs) > 0 {
				for _, tx := range tt.genesis.PendingTreasuryTxs {
					retrieved := keeper.GetPendingTreasuryTx(tx.TxId)
					require.NotNil(t, retrieved)
					require.Equal(t, tx.Recipient, retrieved.Recipient)
					require.Equal(t, tx.Amount, retrieved.Amount)
				}
			}

			// Verify inflation alerts
			if len(tt.genesis.InflationAlerts) > 0 {
				alerts := keeper.GetInflationAlerts(100)
				require.Len(t, alerts, len(tt.genesis.InflationAlerts))
			}

			// Verify large tx records
			if len(tt.genesis.LargeTxRecords) > 0 {
				records := keeper.GetLargeTxRecords(100)
				require.Len(t, records, len(tt.genesis.LargeTxRecords))
			}

			// Verify user MEV balances
			if len(tt.genesis.UserMevBalances) > 0 {
				for addr, expectedBalance := range tt.genesis.UserMevBalances {
					actualBalance := keeper.GetUserMEVBalance(addr)
					require.Equal(t, expectedBalance, actualBalance)
				}
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)

	// Create some test data
	now := time.Now()

	// Add vesting schedule
	err := keeper.CreateVestingSchedule(
		"vesting-1",
		"aura1beneficiary",
		"1000000",
		now,
		now.Add(365*24*time.Hour),
		now.Add(90*24*time.Hour),
		true,
	)
	require.NoError(t, err)

	// Add vote lock
	err = keeper.LockVotingTokens("aura1voter", "50000", "prop-1", 7*24*time.Hour)
	require.NoError(t, err)

	// Add pending treasury tx
	err = keeper.SubmitTreasuryTransaction(
		"tx-1",
		"aura1recipient",
		"100000",
		"Development grant",
		3,
		30*24*time.Hour,
	)
	require.NoError(t, err)

	// Add inflation alert
	keeper.RecordInflationAlert(2100, 2000, "110000000")

	// Add large tx record
	keeper.RecordLargeTx("hash1", "aura1whale", "15000000", true)

	// Add MEV balance
	keeper.AddMEVRewards("aura1user", "50000")

	// Export genesis
	exported := keeper.ExportGenesis()

	// Verify exported data
	require.NotNil(t, exported.Params)
	require.Greater(t, exported.Params.Tokenomics.MaxSupply, "0")

	require.Len(t, exported.VestingSchedules, 1)
	require.Equal(t, "vesting-1", exported.VestingSchedules[0].ScheduleId)
	require.Equal(t, "aura1beneficiary", exported.VestingSchedules[0].BeneficiaryAddress)

	require.Len(t, exported.VoteLocks, 1)
	require.Equal(t, "aura1voter", exported.VoteLocks[0].Owner)

	require.Len(t, exported.PendingTreasuryTxs, 1)
	require.Equal(t, "tx-1", exported.PendingTreasuryTxs[0].TxId)

	require.Len(t, exported.InflationAlerts, 1)
	require.Equal(t, uint64(2100), exported.InflationAlerts[0].CurrentRate)

	require.Len(t, exported.LargeTxRecords, 1)
	require.Equal(t, "hash1", exported.LargeTxRecords[0].TxHash)

	require.Contains(t, exported.UserMevBalances, "aura1user")
	require.Equal(t, "50000", exported.UserMevBalances["aura1user"])
}

func TestGenesisRoundTrip(t *testing.T) {
	// Create first keeper with initial state
	store1 := params.NewStore(*types.DefaultParams())
	keeper1 := NewKeeper(store1)

	now := time.Now()

	// Create comprehensive test data
	err := keeper1.CreateVestingSchedule(
		"vesting-round-trip",
		"aura1beneficiary123",
		"5000000",
		now,
		now.Add(365*24*time.Hour),
		now.Add(90*24*time.Hour),
		true,
	)
	require.NoError(t, err)

	err = keeper1.LockVotingTokens("aura1voter123", "100000", "prop-123", 14*24*time.Hour)
	require.NoError(t, err)

	err = keeper1.SubmitTreasuryTransaction(
		"tx-round-trip",
		"aura1recipient123",
		"250000",
		"Round trip test grant",
		3,
		60*24*time.Hour,
	)
	require.NoError(t, err)

	keeper1.RecordInflationAlert(1800, 2000, "105000000")
	keeper1.RecordLargeTx("hash-round-trip", "aura1whale123", "20000000", true)
	keeper1.AddMEVRewards("aura1mevuser", "75000")

	// Export genesis from keeper1
	exported := keeper1.ExportGenesis()

	// Create a new keeper and import the exported genesis
	store2 := params.NewStore(*types.DefaultParams())
	keeper2 := NewKeeper(store2)
	err = keeper2.InitGenesis(exported)
	require.NoError(t, err)

	// Verify all data was preserved
	params1 := keeper1.GetParams()
	params2 := keeper2.GetParams()
	require.Equal(t, params1.Tokenomics.MaxSupply, params2.Tokenomics.MaxSupply)
	require.Equal(t, params1.Tokenomics.InflationRate, params2.Tokenomics.InflationRate)

	// Verify vesting schedule
	schedule := keeper2.GetVestingSchedule("vesting-round-trip")
	require.NotNil(t, schedule)
	require.Equal(t, "aura1beneficiary123", schedule.BeneficiaryAddress)
	require.Equal(t, "5000000", schedule.TotalAmount)

	// Verify vote lock
	locks := keeper2.GetUserVoteLocks("aura1voter123")
	require.Len(t, locks, 1)
	require.Equal(t, "100000", locks[0].Amount)

	// Verify treasury tx
	tx := keeper2.GetPendingTreasuryTx("tx-round-trip")
	require.NotNil(t, tx)
	require.Equal(t, "aura1recipient123", tx.Recipient)
	require.Equal(t, "250000", tx.Amount)

	// Verify inflation alerts
	alerts := keeper2.GetInflationAlerts(10)
	require.Len(t, alerts, 1)
	require.Equal(t, uint64(1800), alerts[0].CurrentRate)

	// Verify large tx records
	records := keeper2.GetLargeTxRecords(10)
	require.Len(t, records, 1)
	require.Equal(t, "hash-round-trip", records[0].TxHash)

	// Verify MEV balances
	mevBalance := keeper2.GetUserMEVBalance("aura1mevuser")
	require.Equal(t, "75000", mevBalance)

	// Export again and verify consistency
	exported2 := keeper2.ExportGenesis()
	require.Equal(t, len(exported.VestingSchedules), len(exported2.VestingSchedules))
	require.Equal(t, len(exported.VoteLocks), len(exported2.VoteLocks))
	require.Equal(t, len(exported.PendingTreasuryTxs), len(exported2.PendingTreasuryTxs))
	require.Equal(t, len(exported.InflationAlerts), len(exported2.InflationAlerts))
	require.Equal(t, len(exported.LargeTxRecords), len(exported2.LargeTxRecords))
	require.Equal(t, len(exported.UserMevBalances), len(exported2.UserMevBalances))
}

func TestDefaultGenesis(t *testing.T) {
	// Test that default genesis is valid
	defaultGen := types.DefaultGenesis()
	require.NotNil(t, defaultGen)

	// Validate default genesis
	err := types.ValidateGenesis(defaultGen)
	require.NoError(t, err)

	// Verify default params
	require.NotNil(t, defaultGen.Params)
	require.NotEmpty(t, defaultGen.Params.Tokenomics.MaxSupply)
	require.Greater(t, defaultGen.Params.Tokenomics.MaxInflationRate, uint64(0))

	// Verify default collections are empty
	require.Empty(t, defaultGen.VestingSchedules)
	require.Empty(t, defaultGen.VoteLocks)
	require.Empty(t, defaultGen.PendingTreasuryTxs)
	require.Empty(t, defaultGen.InflationAlerts)
	require.Empty(t, defaultGen.LargeTxRecords)
	require.NotNil(t, defaultGen.LastLargeTxTimes)
	require.NotNil(t, defaultGen.UserMevBalances)

	// Test importing default genesis
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)
	err = keeper.InitGenesis(*defaultGen)
	require.NoError(t, err)

	// Verify keeper state after importing default genesis
	p := keeper.GetParams()
	require.NotNil(t, p)
	require.NotEmpty(t, p.Tokenomics.MaxSupply)
}

func TestInitGenesis_WithMultipleSchedules(t *testing.T) {
	now := time.Now()
	genesis := types.GenesisState{
		Params: nil,
		VestingSchedules: []*types.VestingSchedule{
			{
				ScheduleId:         "schedule-1",
				BeneficiaryAddress: "aura1beneficiary1",
				TotalAmount:        "1000000",
				ReleasedAmount:     "100000",
				StartTime:          timestamppb.New(now),
				EndTime:            timestamppb.New(now.Add(365 * 24 * time.Hour)),
				CliffTime:          timestamppb.New(now.Add(90 * 24 * time.Hour)),
				IsRevocable:        true,
			},
			{
				ScheduleId:         "schedule-2",
				BeneficiaryAddress: "aura1beneficiary1", // Same beneficiary
				TotalAmount:        "2000000",
				ReleasedAmount:     "0",
				StartTime:          timestamppb.New(now),
				EndTime:            timestamppb.New(now.Add(180 * 24 * time.Hour)),
				IsRevocable:        false,
			},
			{
				ScheduleId:         "schedule-3",
				BeneficiaryAddress: "aura1beneficiary2",
				TotalAmount:        "500000",
				ReleasedAmount:     "250000",
				StartTime:          timestamppb.New(now.Add(-180 * 24 * time.Hour)),
				EndTime:            timestamppb.New(now.Add(180 * 24 * time.Hour)),
				IsRevocable:        false,
			},
		},
	}

	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)

	err := keeper.InitGenesis(genesis)
	require.NoError(t, err)

	// Verify all schedules were loaded
	schedule1 := keeper.GetVestingSchedule("schedule-1")
	require.NotNil(t, schedule1)
	require.Equal(t, "1000000", schedule1.TotalAmount)

	schedule2 := keeper.GetVestingSchedule("schedule-2")
	require.NotNil(t, schedule2)
	require.Equal(t, "2000000", schedule2.TotalAmount)

	schedule3 := keeper.GetVestingSchedule("schedule-3")
	require.NotNil(t, schedule3)
	require.Equal(t, "500000", schedule3.TotalAmount)

	// Verify user index was built correctly
	schedules1 := keeper.GetUserVestingSchedules("aura1beneficiary1")
	require.Len(t, schedules1, 2)

	schedules2 := keeper.GetUserVestingSchedules("aura1beneficiary2")
	require.Len(t, schedules2, 1)
}
