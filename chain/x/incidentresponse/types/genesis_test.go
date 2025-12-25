// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	genesis := DefaultGenesisState()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.Params)
	require.NotNil(t, genesis.PauseState)
	require.NotNil(t, genesis.Incidents)
	require.NotNil(t, genesis.WalletLimits)
	require.Equal(t, uint64(1), genesis.NextIncidentID)
	require.False(t, genesis.PauseState.IsPaused)
	require.Equal(t, PauseLevelNone, genesis.PauseState.PauseLevel)

	// Verify genesis is valid
	err := genesis.Validate()
	require.NoError(t, err)
}

func TestGenesisState_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		genesis GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid genesis state",
			genesis: GenesisState{
				Params: &IncidentResponseParams{
					EmergencyPauseEnabled:  true,
					PauseAuthorizedKeys:    []string{"admin1", "admin2"},
					PauseRequiredSigners:   2,
					MaxPauseDuration:       24 * time.Hour,
					HotWalletLimitsEnabled: true,
					GlobalMaxHotWallet:     "10000000000",
					GlobalDailyLimit:       "1000000000",
					ColdStorage: ColdStorageConfig{
						Enabled:           true,
						MultiSigThreshold: 3,
						MultiSigSigners:   []string{"signer1", "signer2", "signer3"},
						MinimumBalance:    "50000000000",
						MaxHotWalletRatio: 0.2,
					},
					DisasterRecovery: DisasterRecoveryPlan{
						Enabled:         true,
						BackupInterval:  6 * time.Hour,
						BackupLocations: []string{"s3://backup"},
					},
				},
				Incidents: []*Incident{
					{
						ID:              "INC-1",
						Title:           "Test Incident",
						Description:     "Test Description",
						Severity:        SeverityHigh,
						Status:          StatusNew,
						ReportedBy:      "reporter",
						ReportedAt:      now,
						UpdatedAt:       now,
						AffectedSystems: []string{"system1"},
						Timeline:        []IncidentTimelineEntry{},
					},
				},
				PauseState: &ChainPauseState{
					IsPaused:   false,
					PauseLevel: PauseLevelNone,
				},
				WalletLimits: []*WalletLimits{
					{
						Address:            "aura1abc123",
						MaxBalance:         "10000000000",
						MaxTransactionSize: "1000000000",
						DailyLimit:         "5000000000",
						CurrentBalance:     "0",
						TodayTransferred:   "0",
						LastReset:          now,
					},
				},
				NextIncidentID: 2,
			},
			wantErr: false,
		},
		{
			name: "invalid params",
			genesis: GenesisState{
				Params: &IncidentResponseParams{
					EmergencyPauseEnabled: true,
					PauseAuthorizedKeys:   []string{}, // Invalid: empty when enabled
					PauseRequiredSigners:  0,
					MaxPauseDuration:      24 * time.Hour,
				},
				Incidents:      []*Incident{},
				PauseState:     &ChainPauseState{IsPaused: false, PauseLevel: PauseLevelNone},
				WalletLimits:   []*WalletLimits{},
				NextIncidentID: 1,
			},
			wantErr: true,
			errMsg:  "invalid params",
		},
		{
			name: "nil incident",
			genesis: GenesisState{
				Params: nil,
				Incidents: []*Incident{
					nil, // Invalid: nil incident
				},
				PauseState:     &ChainPauseState{IsPaused: false, PauseLevel: PauseLevelNone},
				WalletLimits:   []*WalletLimits{},
				NextIncidentID: 1,
			},
			wantErr: true,
			errMsg:  "incident 0 is nil",
		},
		{
			name: "empty incident ID",
			genesis: GenesisState{
				Params: nil,
				Incidents: []*Incident{
					{
						ID:              "", // Invalid: empty ID
						Title:           "Test",
						Description:     "Test",
						Severity:        SeverityHigh,
						Status:          StatusNew,
						ReportedBy:      "reporter",
						ReportedAt:      now,
						UpdatedAt:       now,
						AffectedSystems: []string{"system"},
						Timeline:        []IncidentTimelineEntry{},
					},
				},
				PauseState:     &ChainPauseState{IsPaused: false, PauseLevel: PauseLevelNone},
				WalletLimits:   []*WalletLimits{},
				NextIncidentID: 1,
			},
			wantErr: true,
			errMsg:  "has empty ID",
		},
		{
			name: "duplicate incident ID",
			genesis: GenesisState{
				Params: nil,
				Incidents: []*Incident{
					{
						ID:              "INC-1",
						Title:           "Incident 1",
						Description:     "Test",
						Severity:        SeverityHigh,
						Status:          StatusNew,
						ReportedBy:      "reporter",
						ReportedAt:      now,
						UpdatedAt:       now,
						AffectedSystems: []string{"system"},
						Timeline:        []IncidentTimelineEntry{},
					},
					{
						ID:              "INC-1", // Duplicate
						Title:           "Incident 2",
						Description:     "Test",
						Severity:        SeverityMedium,
						Status:          StatusNew,
						ReportedBy:      "reporter",
						ReportedAt:      now,
						UpdatedAt:       now,
						AffectedSystems: []string{"system"},
						Timeline:        []IncidentTimelineEntry{},
					},
				},
				PauseState:     &ChainPauseState{IsPaused: false, PauseLevel: PauseLevelNone},
				WalletLimits:   []*WalletLimits{},
				NextIncidentID: 2,
			},
			wantErr: true,
			errMsg:  "duplicate incident ID",
		},
		{
			name: "empty incident title",
			genesis: GenesisState{
				Params: nil,
				Incidents: []*Incident{
					{
						ID:              "INC-1",
						Title:           "", // Invalid: empty title
						Description:     "Test",
						Severity:        SeverityHigh,
						Status:          StatusNew,
						ReportedBy:      "reporter",
						ReportedAt:      now,
						UpdatedAt:       now,
						AffectedSystems: []string{"system"},
						Timeline:        []IncidentTimelineEntry{},
					},
				},
				PauseState:     &ChainPauseState{IsPaused: false, PauseLevel: PauseLevelNone},
				WalletLimits:   []*WalletLimits{},
				NextIncidentID: 2,
			},
			wantErr: true,
			errMsg:  "has empty title",
		},
		{
			name: "empty reported_by",
			genesis: GenesisState{
				Params: nil,
				Incidents: []*Incident{
					{
						ID:              "INC-1",
						Title:           "Test",
						Description:     "Test",
						Severity:        SeverityHigh,
						Status:          StatusNew,
						ReportedBy:      "", // Invalid: empty reported_by
						ReportedAt:      now,
						UpdatedAt:       now,
						AffectedSystems: []string{"system"},
						Timeline:        []IncidentTimelineEntry{},
					},
				},
				PauseState:     &ChainPauseState{IsPaused: false, PauseLevel: PauseLevelNone},
				WalletLimits:   []*WalletLimits{},
				NextIncidentID: 2,
			},
			wantErr: true,
			errMsg:  "has empty reported_by",
		},
		{
			name: "paused chain without paused_by",
			genesis: GenesisState{
				Params:    nil,
				Incidents: []*Incident{},
				PauseState: &ChainPauseState{
					IsPaused:   true,
					PauseLevel: PauseLevelFull,
					PausedBy:   "", // Invalid: must have paused_by
				},
				WalletLimits:   []*WalletLimits{},
				NextIncidentID: 1,
			},
			wantErr: true,
			errMsg:  "must have paused_by set",
		},
		{
			name: "paused chain with PauseLevelNone",
			genesis: GenesisState{
				Params:    nil,
				Incidents: []*Incident{},
				PauseState: &ChainPauseState{
					IsPaused:   true,
					PauseLevel: PauseLevelNone, // Invalid: must have valid pause level
					PausedBy:   "admin",
				},
				WalletLimits:   []*WalletLimits{},
				NextIncidentID: 1,
			},
			wantErr: true,
			errMsg:  "must have valid pause level",
		},
		{
			name: "nil wallet limit",
			genesis: GenesisState{
				Params:    nil,
				Incidents: []*Incident{},
				PauseState: &ChainPauseState{
					IsPaused:   false,
					PauseLevel: PauseLevelNone,
				},
				WalletLimits: []*WalletLimits{
					nil, // Invalid: nil wallet limit
				},
				NextIncidentID: 1,
			},
			wantErr: true,
			errMsg:  "wallet limit 0 is nil",
		},
		{
			name: "empty wallet address",
			genesis: GenesisState{
				Params:    nil,
				Incidents: []*Incident{},
				PauseState: &ChainPauseState{
					IsPaused:   false,
					PauseLevel: PauseLevelNone,
				},
				WalletLimits: []*WalletLimits{
					{
						Address:            "", // Invalid: empty address
						MaxBalance:         "10000000000",
						MaxTransactionSize: "1000000000",
						DailyLimit:         "5000000000",
					},
				},
				NextIncidentID: 1,
			},
			wantErr: true,
			errMsg:  "has empty address",
		},
		{
			name: "duplicate wallet address",
			genesis: GenesisState{
				Params:    nil,
				Incidents: []*Incident{},
				PauseState: &ChainPauseState{
					IsPaused:   false,
					PauseLevel: PauseLevelNone,
				},
				WalletLimits: []*WalletLimits{
					{
						Address:            "aura1abc123",
						MaxBalance:         "10000000000",
						MaxTransactionSize: "1000000000",
						DailyLimit:         "5000000000",
					},
					{
						Address:            "aura1abc123", // Duplicate
						MaxBalance:         "20000000000",
						MaxTransactionSize: "2000000000",
						DailyLimit:         "10000000000",
					},
				},
				NextIncidentID: 1,
			},
			wantErr: true,
			errMsg:  "duplicate wallet limit for address",
		},
		{
			name: "zero next incident ID",
			genesis: GenesisState{
				Params:         nil,
				Incidents:      []*Incident{},
				PauseState:     &ChainPauseState{IsPaused: false, PauseLevel: PauseLevelNone},
				WalletLimits:   []*WalletLimits{},
				NextIncidentID: 0, // Invalid: must be > 0
			},
			wantErr: true,
			errMsg:  "must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.genesis.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenesisState_Validate_NilPauseState(t *testing.T) {
	genesis := GenesisState{
		Params:         nil,
		Incidents:      []*Incident{},
		PauseState:     nil, // nil pause state should be handled
		WalletLimits:   []*WalletLimits{},
		NextIncidentID: 1,
	}

	err := genesis.Validate()
	require.NoError(t, err)
}

func TestGenesisState_Validate_MultipleIncidents(t *testing.T) {
	now := time.Now()

	genesis := GenesisState{
		Params: nil,
		Incidents: []*Incident{
			{
				ID:              "INC-1",
				Title:           "Incident 1",
				Description:     "Description 1",
				Severity:        SeverityCritical,
				Status:          StatusNew,
				ReportedBy:      "reporter1",
				ReportedAt:      now,
				UpdatedAt:       now,
				AffectedSystems: []string{"system1"},
				Timeline:        []IncidentTimelineEntry{},
			},
			{
				ID:              "INC-2",
				Title:           "Incident 2",
				Description:     "Description 2",
				Severity:        SeverityHigh,
				Status:          StatusInvestigation,
				ReportedBy:      "reporter2",
				ReportedAt:      now,
				UpdatedAt:       now,
				AffectedSystems: []string{"system2"},
				Timeline:        []IncidentTimelineEntry{},
			},
			{
				ID:              "INC-3",
				Title:           "Incident 3",
				Description:     "Description 3",
				Severity:        SeverityMedium,
				Status:          StatusResolved,
				ReportedBy:      "reporter3",
				ReportedAt:      now,
				UpdatedAt:       now,
				ResolvedAt:      now.Add(1 * time.Hour),
				AffectedSystems: []string{"system3"},
				Timeline:        []IncidentTimelineEntry{},
			},
		},
		PauseState: &ChainPauseState{
			IsPaused:   false,
			PauseLevel: PauseLevelNone,
		},
		WalletLimits:   []*WalletLimits{},
		NextIncidentID: 4,
	}

	err := genesis.Validate()
	require.NoError(t, err)
}

func TestGenesisState_Validate_IncidentWithPostMortem(t *testing.T) {
	now := time.Now()

	genesis := GenesisState{
		Params: nil,
		Incidents: []*Incident{
			{
				ID:              "INC-1",
				Title:           "Critical Incident",
				Description:     "Description",
				Severity:        SeverityCritical,
				Status:          StatusPostMortem,
				ReportedBy:      "reporter",
				ReportedAt:      now,
				UpdatedAt:       now,
				ResolvedAt:      now.Add(2 * time.Hour),
				AffectedSystems: []string{"system1"},
				Timeline:        []IncidentTimelineEntry{},
				PostMortem: &PostMortem{
					CreatedAt:      now.Add(2 * time.Hour),
					CreatedBy:      "analyst",
					Summary:        "Summary",
					RootCause:      "Root cause",
					Impact:         "Impact",
					Resolution:     "Resolution",
					LessonsLearned: []string{"Lesson 1", "Lesson 2"},
					ActionItems:    []ActionItem{},
				},
			},
		},
		PauseState: &ChainPauseState{
			IsPaused:   false,
			PauseLevel: PauseLevelNone,
		},
		WalletLimits:   []*WalletLimits{},
		NextIncidentID: 2,
	}

	err := genesis.Validate()
	require.NoError(t, err)
}
