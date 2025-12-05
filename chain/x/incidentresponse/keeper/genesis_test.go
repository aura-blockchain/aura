package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/incidentresponse/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	tests := []struct {
		name    string
		genesis types.GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid genesis with incidents",
			genesis: types.GenesisState{
				Params: &types.IncidentResponseParams{
					EmergencyPauseEnabled:  true,
					PauseAuthorizedKeys:    []string{"admin1", "admin2"},
					PauseRequiredSigners:   2,
					MaxPauseDuration:       24 * time.Hour,
					HotWalletLimitsEnabled: true,
					GlobalMaxHotWallet:     "10000000000",
					GlobalDailyLimit:       "1000000000",
					ColdStorage: types.ColdStorageConfig{
						Enabled:           true,
						MultiSigThreshold: 3,
						MultiSigSigners:   []string{"signer1", "signer2", "signer3"},
						MinimumBalance:    "50000000000",
						MaxHotWalletRatio: 0.2,
					},
					BackupValidators: types.BackupValidatorConfig{
						Enabled:           true,
						AutoFailover:      true,
						FailoverThreshold: 3,
						HeartbeatInterval: 30 * time.Second,
					},
					Communication: types.CommunicationPlan{
						Enabled:        true,
						UpdateInterval: 30 * time.Minute,
					},
					DisasterRecovery: types.DisasterRecoveryPlan{
						Enabled:           true,
						BackupInterval:    6 * time.Hour,
						BackupLocations:   []string{"s3://backup"},
						RPO:               15 * time.Minute,
						RTO:               2 * time.Hour,
						SnapshotRetention: 7,
						ValidatorBackups:  true,
						StateBackups:      true,
						KeyBackups:        false,
					},
					Insurance: types.InsuranceIntegration{
						Enabled:        false,
						AutoClaim:      false,
						ClaimThreshold: "1000000000000",
					},
					IncidentResponseTeam: []string{"team1", "team2"},
				},
				Incidents: []*types.Incident{
					{
						ID:              "INC-1",
						Title:           "Test Incident 1",
						Description:     "Description 1",
						Severity:        types.SeverityHigh,
						Status:          types.StatusNew,
						ReportedBy:      "reporter1",
						ReportedAt:      time.Now(),
						UpdatedAt:       time.Now(),
						AffectedSystems: []string{"system1"},
						ResponseTeam:    []string{"team1"},
						Timeline:        []types.IncidentTimelineEntry{},
					},
					{
						ID:              "INC-2",
						Title:           "Test Incident 2",
						Description:     "Description 2",
						Severity:        types.SeverityMedium,
						Status:          types.StatusInvestigation,
						ReportedBy:      "reporter2",
						ReportedAt:      time.Now(),
						UpdatedAt:       time.Now(),
						AffectedSystems: []string{"system2"},
						ResponseTeam:    []string{"team2"},
						Timeline:        []types.IncidentTimelineEntry{},
					},
				},
				PauseState: &types.ChainPauseState{
					IsPaused:   false,
					PauseLevel: types.PauseLevelNone,
				},
				WalletLimits: []*types.WalletLimits{
					{
						Address:            "aura1abc123",
						MaxBalance:         "10000000000",
						MaxTransactionSize: "1000000000",
						DailyLimit:         "5000000000",
						CurrentBalance:     "0",
						TodayTransferred:   "0",
						LastReset:          time.Now(),
					},
				},
				NextIncidentID: 3,
			},
			wantErr: false,
		},
		{
			name: "default genesis",
			genesis: types.GenesisState{
				Params:         nil,
				Incidents:      []*types.Incident{},
				PauseState:     &types.ChainPauseState{IsPaused: false, PauseLevel: types.PauseLevelNone},
				WalletLimits:   []*types.WalletLimits{},
				NextIncidentID: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid genesis - duplicate incident IDs",
			genesis: types.GenesisState{
				Params: nil,
				Incidents: []*types.Incident{
					{
						ID:              "INC-1",
						Title:           "Incident 1",
						Description:     "Description",
						Severity:        types.SeverityHigh,
						Status:          types.StatusNew,
						ReportedBy:      "reporter",
						ReportedAt:      time.Now(),
						UpdatedAt:       time.Now(),
						AffectedSystems: []string{"system"},
						Timeline:        []types.IncidentTimelineEntry{},
					},
					{
						ID:              "INC-1", // Duplicate
						Title:           "Incident 2",
						Description:     "Description",
						Severity:        types.SeverityMedium,
						Status:          types.StatusNew,
						ReportedBy:      "reporter",
						ReportedAt:      time.Now(),
						UpdatedAt:       time.Now(),
						AffectedSystems: []string{"system"},
						Timeline:        []types.IncidentTimelineEntry{},
					},
				},
				PauseState:     &types.ChainPauseState{IsPaused: false, PauseLevel: types.PauseLevelNone},
				WalletLimits:   []*types.WalletLimits{},
				NextIncidentID: 2,
			},
			wantErr: true,
			errMsg:  "duplicate incident ID",
		},
		{
			name: "invalid genesis - zero next incident ID",
			genesis: types.GenesisState{
				Params:         nil,
				Incidents:      []*types.Incident{},
				PauseState:     &types.ChainPauseState{IsPaused: false, PauseLevel: types.PauseLevelNone},
				WalletLimits:   []*types.WalletLimits{},
				NextIncidentID: 0,
			},
			wantErr: true,
			errMsg:  "must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.InitGenesis(ctx, tt.genesis)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)

				// Verify params were set
				params := keeper.GetParams(ctx)
				require.NotNil(t, params)

				// Verify incidents were loaded
				if len(tt.genesis.Incidents) > 0 {
					for _, incident := range tt.genesis.Incidents {
						retrieved, err := keeper.GetIncident(ctx, incident.ID)
						require.NoError(t, err)
						require.Equal(t, incident.Title, retrieved.Title)
					}
				}

				// Verify pause state was set
				pauseState := keeper.GetChainPauseState(ctx)
				require.NotNil(t, pauseState)
				require.Equal(t, tt.genesis.PauseState.IsPaused, pauseState.IsPaused)

				// Verify wallet limits were loaded
				if len(tt.genesis.WalletLimits) > 0 {
					for _, limit := range tt.genesis.WalletLimits {
						retrieved, err := keeper.GetWalletLimits(ctx, limit.Address)
						require.NoError(t, err)
						require.Equal(t, limit.MaxBalance, retrieved.MaxBalance)
					}
				}
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set up test data
	params := types.DefaultParams()
	params.EmergencyPauseEnabled = true
	params.PauseAuthorizedKeys = []string{"admin1", "admin2"}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create test incidents
	incident1ID, err := keeper.ReportIncident(
		ctx,
		"Incident 1",
		"Description 1",
		types.SeverityCritical,
		"reporter1",
		[]string{"system1"},
	)
	require.NoError(t, err)

	incident2ID, err := keeper.ReportIncident(
		ctx,
		"Incident 2",
		"Description 2",
		types.SeverityHigh,
		"reporter2",
		[]string{"system2"},
	)
	require.NoError(t, err)

	// Set wallet limits
	err = keeper.SetWalletLimits(
		ctx,
		"aura1test123",
		"10000000000",
		"1000000000",
		"5000000000",
	)
	require.NoError(t, err)

	// Export genesis
	exported := keeper.ExportGenesis(ctx)

	// Verify exported data
	require.NotNil(t, exported.Params)
	require.True(t, exported.Params.EmergencyPauseEnabled)
	require.Equal(t, []string{"admin1", "admin2"}, exported.Params.PauseAuthorizedKeys)

	require.Len(t, exported.Incidents, 2)
	require.Contains(t, []string{incident1ID, incident2ID}, exported.Incidents[0].ID)
	require.Contains(t, []string{incident1ID, incident2ID}, exported.Incidents[1].ID)

	require.NotNil(t, exported.PauseState)
	require.False(t, exported.PauseState.IsPaused)

	require.Len(t, exported.WalletLimits, 1)
	require.Equal(t, "aura1test123", exported.WalletLimits[0].Address)

	require.Greater(t, exported.NextIncidentID, uint64(0))
}

func TestGenesisRoundTrip(t *testing.T) {
	keeper1, ctx1 := setupKeeperForTest(t)

	// Set up initial state
	params := types.DefaultParams()
	params.EmergencyPauseEnabled = true
	params.PauseAuthorizedKeys = []string{"admin1", "admin2", "admin3"}
	params.PauseRequiredSigners = 2
	err := keeper1.SetParams(ctx1, params)
	require.NoError(t, err)

	// Create incidents
	incidentID1, err := keeper1.ReportIncident(
		ctx1,
		"Security Breach",
		"Unauthorized access detected",
		types.SeverityCritical,
		"security-team",
		[]string{"database", "api-server"},
	)
	require.NoError(t, err)

	err = keeper1.UpdateIncidentStatus(
		ctx1,
		incidentID1,
		types.StatusInvestigation,
		"analyst",
		"Investigation started",
	)
	require.NoError(t, err)

	// Set wallet limits
	err = keeper1.SetWalletLimits(
		ctx1,
		"aura1wallet123",
		"20000000000",
		"2000000000",
		"10000000000",
	)
	require.NoError(t, err)

	// Export genesis from keeper1
	exported := keeper1.ExportGenesis(ctx1)

	// Create a new keeper and import the exported genesis
	keeper2, ctx2 := setupKeeperForTest(t)
	err = keeper2.InitGenesis(ctx2, exported)
	require.NoError(t, err)

	// Verify all data was preserved
	params2 := keeper2.GetParams(ctx2)
	require.Equal(t, params.EmergencyPauseEnabled, params2.EmergencyPauseEnabled)
	require.Equal(t, params.PauseAuthorizedKeys, params2.PauseAuthorizedKeys)
	require.Equal(t, params.PauseRequiredSigners, params2.PauseRequiredSigners)

	// Verify incident
	incident, err := keeper2.GetIncident(ctx2, incidentID1)
	require.NoError(t, err)
	require.Equal(t, "Security Breach", incident.Title)
	require.Equal(t, types.StatusInvestigation, incident.Status)
	require.Len(t, incident.Timeline, 2) // reported + status update

	// Verify wallet limits
	limits, err := keeper2.GetWalletLimits(ctx2, "aura1wallet123")
	require.NoError(t, err)
	require.Equal(t, "20000000000", limits.MaxBalance)

	// Export again and verify consistency
	exported2 := keeper2.ExportGenesis(ctx2)
	require.Equal(t, len(exported.Incidents), len(exported2.Incidents))
	require.Equal(t, len(exported.WalletLimits), len(exported2.WalletLimits))
	require.Equal(t, exported.NextIncidentID, exported2.NextIncidentID)
}

func TestInitGenesis_WithPausedChain(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	now := time.Now()
	genesis := types.GenesisState{
		Params: &types.IncidentResponseParams{
			EmergencyPauseEnabled: true,
			PauseAuthorizedKeys:   []string{"admin1"},
			PauseRequiredSigners:  1,
			MaxPauseDuration:      24 * time.Hour,
			DisasterRecovery: types.DisasterRecoveryPlan{
				Enabled:         true,
				BackupInterval:  6 * time.Hour,
				BackupLocations: []string{"s3://backup"},
			},
		},
		Incidents: []*types.Incident{},
		PauseState: &types.ChainPauseState{
			IsPaused:        true,
			PauseLevel:      types.PauseLevelFull,
			PausedAt:        now,
			PausedBy:        "admin1",
			Reason:          "Emergency maintenance",
			IncidentID:      "INC-123",
			EstimatedResume: now.Add(2 * time.Hour),
		},
		WalletLimits:   []*types.WalletLimits{},
		NextIncidentID: 1,
	}

	err := keeper.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify chain is paused
	pauseState := keeper.GetChainPauseState(ctx)
	require.True(t, pauseState.IsPaused)
	require.Equal(t, types.PauseLevelFull, pauseState.PauseLevel)
	require.Equal(t, "admin1", pauseState.PausedBy)
	require.Equal(t, "Emergency maintenance", pauseState.Reason)
	require.Equal(t, "INC-123", pauseState.IncidentID)
}

func TestInitGenesis_WithMultipleWalletLimits(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	genesis := types.GenesisState{
		Params: &types.IncidentResponseParams{
			HotWalletLimitsEnabled: true,
			GlobalMaxHotWallet:     "10000000000",
			GlobalDailyLimit:       "1000000000",
		},
		Incidents: []*types.Incident{},
		PauseState: &types.ChainPauseState{
			IsPaused:   false,
			PauseLevel: types.PauseLevelNone,
		},
		WalletLimits: []*types.WalletLimits{
			{
				Address:            "aura1wallet1",
				MaxBalance:         "5000000000",
				MaxTransactionSize: "500000000",
				DailyLimit:         "2000000000",
				CurrentBalance:     "0",
				TodayTransferred:   "0",
				LastReset:          time.Now(),
			},
			{
				Address:            "aura1wallet2",
				MaxBalance:         "10000000000",
				MaxTransactionSize: "1000000000",
				DailyLimit:         "5000000000",
				CurrentBalance:     "0",
				TodayTransferred:   "0",
				LastReset:          time.Now(),
			},
			{
				Address:            "aura1wallet3",
				MaxBalance:         "3000000000",
				MaxTransactionSize: "300000000",
				DailyLimit:         "1000000000",
				CurrentBalance:     "0",
				TodayTransferred:   "0",
				LastReset:          time.Now(),
			},
		},
		NextIncidentID: 1,
	}

	err := keeper.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify all wallet limits were loaded
	for _, expectedLimit := range genesis.WalletLimits {
		actualLimit, err := keeper.GetWalletLimits(ctx, expectedLimit.Address)
		require.NoError(t, err)
		require.Equal(t, expectedLimit.MaxBalance, actualLimit.MaxBalance)
		require.Equal(t, expectedLimit.MaxTransactionSize, actualLimit.MaxTransactionSize)
		require.Equal(t, expectedLimit.DailyLimit, actualLimit.DailyLimit)
	}
}
