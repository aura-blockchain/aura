// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	require.NotNil(t, params)
	require.True(t, params.EmergencyPauseEnabled)
	require.Equal(t, uint32(3), params.PauseRequiredSigners)
	require.Equal(t, 24*time.Hour, params.MaxPauseDuration)
	require.True(t, params.HotWalletLimitsEnabled)
	require.NotEmpty(t, params.GlobalMaxHotWallet)
	require.NotEmpty(t, params.GlobalDailyLimit)

	// Validate default params
	err := params.ValidateBasic()
	require.NoError(t, err)
}

func TestIncidentResponseParams_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		params  IncidentResponseParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid params",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  true,
				PauseAuthorizedKeys:    []string{"admin1", "admin2", "admin3"},
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
			wantErr: false,
		},
		{
			name: "empty authorized keys when pause enabled",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  true,
				PauseAuthorizedKeys:    []string{}, // Invalid
				PauseRequiredSigners:   1,
				MaxPauseDuration:       24 * time.Hour,
				HotWalletLimitsEnabled: false,
			},
			wantErr: true,
			errMsg:  "cannot be empty when emergency pause is enabled",
		},
		{
			name: "zero required signers",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  true,
				PauseAuthorizedKeys:    []string{"admin1"},
				PauseRequiredSigners:   0, // Invalid
				MaxPauseDuration:       24 * time.Hour,
				HotWalletLimitsEnabled: false,
			},
			wantErr: true,
			errMsg:  "must be greater than 0",
		},
		{
			name: "required signers exceeds authorized keys",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  true,
				PauseAuthorizedKeys:    []string{"admin1", "admin2"},
				PauseRequiredSigners:   3, // Invalid: exceeds authorized keys
				MaxPauseDuration:       24 * time.Hour,
				HotWalletLimitsEnabled: false,
			},
			wantErr: true,
			errMsg:  "cannot exceed number of authorized keys",
		},
		{
			name: "cold storage with zero multisig threshold",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  false,
				HotWalletLimitsEnabled: false,
				ColdStorage: ColdStorageConfig{
					Enabled:           true,
					MultiSigThreshold: 0, // Invalid
					MinimumBalance:    "1000000",
					MaxHotWalletRatio: 0.5,
				},
			},
			wantErr: true,
			errMsg:  "must be greater than 0",
		},
		{
			name: "cold storage with invalid hot wallet ratio (negative)",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  false,
				HotWalletLimitsEnabled: false,
				ColdStorage: ColdStorageConfig{
					Enabled:           true,
					MultiSigThreshold: 3,
					MinimumBalance:    "1000000",
					MaxHotWalletRatio: -0.1, // Invalid
				},
			},
			wantErr: true,
			errMsg:  "must be between 0 and 1",
		},
		{
			name: "cold storage with invalid hot wallet ratio (>1)",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  false,
				HotWalletLimitsEnabled: false,
				ColdStorage: ColdStorageConfig{
					Enabled:           true,
					MultiSigThreshold: 3,
					MinimumBalance:    "1000000",
					MaxHotWalletRatio: 1.5, // Invalid
				},
			},
			wantErr: true,
			errMsg:  "must be between 0 and 1",
		},
		{
			name: "disaster recovery with zero backup interval",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  false,
				HotWalletLimitsEnabled: false,
				DisasterRecovery: DisasterRecoveryPlan{
					Enabled:         true,
					BackupInterval:  0, // Invalid
					BackupLocations: []string{"s3://backup"},
				},
			},
			wantErr: true,
			errMsg:  "must be greater than 0",
		},
		{
			name: "disaster recovery with empty backup locations",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  false,
				HotWalletLimitsEnabled: false,
				DisasterRecovery: DisasterRecoveryPlan{
					Enabled:         true,
					BackupInterval:  6 * time.Hour,
					BackupLocations: []string{}, // Invalid
				},
			},
			wantErr: true,
			errMsg:  "at least one backup location must be specified",
		},
		{
			name: "pause disabled with empty authorized keys",
			params: IncidentResponseParams{
				EmergencyPauseEnabled:  false, // Disabled, so empty keys are OK
				PauseAuthorizedKeys:    []string{},
				PauseRequiredSigners:   0,
				MaxPauseDuration:       24 * time.Hour,
				HotWalletLimitsEnabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.ValidateBasic()

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

func TestIncidentSeverity_Validation(t *testing.T) {
	validSeverities := []IncidentSeverity{
		SeverityLow,
		SeverityMedium,
		SeverityHigh,
		SeverityCritical,
	}

	for _, severity := range validSeverities {
		require.NotEmpty(t, string(severity))
	}

	require.Equal(t, "low", string(SeverityLow))
	require.Equal(t, "medium", string(SeverityMedium))
	require.Equal(t, "high", string(SeverityHigh))
	require.Equal(t, "critical", string(SeverityCritical))
}

func TestIncidentStatus_Validation(t *testing.T) {
	validStatuses := []IncidentStatus{
		StatusNew,
		StatusInvestigation,
		StatusContained,
		StatusResolved,
		StatusPostMortem,
		StatusClosed,
	}

	for _, status := range validStatuses {
		require.NotEmpty(t, string(status))
	}

	require.Equal(t, "new", string(StatusNew))
	require.Equal(t, "investigating", string(StatusInvestigation))
	require.Equal(t, "contained", string(StatusContained))
	require.Equal(t, "resolved", string(StatusResolved))
	require.Equal(t, "post_mortem", string(StatusPostMortem))
	require.Equal(t, "closed", string(StatusClosed))
}

func TestPauseLevel_Validation(t *testing.T) {
	validLevels := []PauseLevel{
		PauseLevelNone,
		PauseLevelTransactions,
		PauseLevelModules,
		PauseLevelFull,
	}

	for _, level := range validLevels {
		require.NotEmpty(t, string(level))
	}

	require.Equal(t, "none", string(PauseLevelNone))
	require.Equal(t, "transactions", string(PauseLevelTransactions))
	require.Equal(t, "modules", string(PauseLevelModules))
	require.Equal(t, "full", string(PauseLevelFull))
}

func TestIncident_Structure(t *testing.T) {
	now := time.Now()

	incident := &Incident{
		ID:              "INC-123",
		Title:           "Test Incident",
		Description:     "Test Description",
		Severity:        SeverityHigh,
		Status:          StatusNew,
		ReportedBy:      "reporter",
		ReportedAt:      now,
		UpdatedAt:       now,
		AffectedSystems: []string{"system1", "system2"},
		ResponseTeam:    []string{"team1", "team2"},
		Timeline: []IncidentTimelineEntry{
			{
				Timestamp:   now,
				Action:      "reported",
				Description: "Incident reported",
				Actor:       "reporter",
			},
		},
	}

	require.Equal(t, "INC-123", incident.ID)
	require.Equal(t, "Test Incident", incident.Title)
	require.Equal(t, SeverityHigh, incident.Severity)
	require.Equal(t, StatusNew, incident.Status)
	require.Len(t, incident.AffectedSystems, 2)
	require.Len(t, incident.ResponseTeam, 2)
	require.Len(t, incident.Timeline, 1)
}

func TestWalletLimits_Structure(t *testing.T) {
	now := time.Now()

	limits := &WalletLimits{
		Address:            "aura1abc123",
		MaxBalance:         "10000000000",
		MaxTransactionSize: "1000000000",
		DailyLimit:         "5000000000",
		CurrentBalance:     "1000000",
		TodayTransferred:   "500000",
		LastReset:          now,
	}

	require.Equal(t, "aura1abc123", limits.Address)
	require.Equal(t, "10000000000", limits.MaxBalance)
	require.Equal(t, "1000000000", limits.MaxTransactionSize)
	require.Equal(t, "5000000000", limits.DailyLimit)
}

func TestChainPauseState_Structure(t *testing.T) {
	now := time.Now()

	pauseState := &ChainPauseState{
		IsPaused:        true,
		PauseLevel:      PauseLevelFull,
		PausedAt:        now,
		PausedBy:        "admin",
		Reason:          "Security incident",
		IncidentID:      "INC-456",
		PausedModules:   []string{"module1", "module2"},
		EstimatedResume: now.Add(2 * time.Hour),
	}

	require.True(t, pauseState.IsPaused)
	require.Equal(t, PauseLevelFull, pauseState.PauseLevel)
	require.Equal(t, "admin", pauseState.PausedBy)
	require.Equal(t, "INC-456", pauseState.IncidentID)
	require.Len(t, pauseState.PausedModules, 2)
}

func TestPostMortem_Structure(t *testing.T) {
	now := time.Now()

	postMortem := &PostMortem{
		CreatedAt:      now,
		CreatedBy:      "analyst",
		Summary:        "Incident summary",
		RootCause:      "Configuration error",
		Impact:         "Minor service disruption",
		Resolution:     "Configuration fixed",
		LessonsLearned: []string{"Lesson 1", "Lesson 2"},
		ActionItems: []ActionItem{
			{
				ID:          "ACTION-1",
				Description: "Update documentation",
				Assignee:    "team-lead",
				Priority:    "high",
				Status:      "pending",
				DueDate:     now.Add(7 * 24 * time.Hour),
			},
		},
	}

	require.Equal(t, "analyst", postMortem.CreatedBy)
	require.Equal(t, "Incident summary", postMortem.Summary)
	require.Len(t, postMortem.LessonsLearned, 2)
	require.Len(t, postMortem.ActionItems, 1)
	require.Equal(t, "ACTION-1", postMortem.ActionItems[0].ID)
}

func TestColdStorageConfig_Structure(t *testing.T) {
	now := time.Now()

	config := ColdStorageConfig{
		Enabled:           true,
		MultiSigThreshold: 5,
		MultiSigSigners:   []string{"signer1", "signer2", "signer3", "signer4", "signer5"},
		TimeLockedUntil:   now.Add(30 * 24 * time.Hour),
		MinimumBalance:    "50000000000",
		MaxHotWalletRatio: 0.2,
	}

	require.True(t, config.Enabled)
	require.Equal(t, uint32(5), config.MultiSigThreshold)
	require.Len(t, config.MultiSigSigners, 5)
	require.Equal(t, 0.2, config.MaxHotWalletRatio)
}

func TestBackupValidatorConfig_Structure(t *testing.T) {
	now := time.Now()

	config := BackupValidatorConfig{
		Enabled:           true,
		PrimaryValidators: []string{"val1", "val2", "val3"},
		BackupValidators:  []string{"backup1", "backup2"},
		AutoFailover:      true,
		FailoverThreshold: 3,
		HeartbeatInterval: 30 * time.Second,
		LastHealthCheck:   now,
	}

	require.True(t, config.Enabled)
	require.Len(t, config.PrimaryValidators, 3)
	require.Len(t, config.BackupValidators, 2)
	require.True(t, config.AutoFailover)
	require.Equal(t, 3, config.FailoverThreshold)
}

func TestDisasterRecoveryPlan_Structure(t *testing.T) {
	now := time.Now()

	plan := DisasterRecoveryPlan{
		Enabled:           true,
		BackupInterval:    6 * time.Hour,
		BackupLocations:   []string{"s3://backup1", "s3://backup2"},
		LastBackupTime:    now,
		RPO:               15 * time.Minute,
		RTO:               2 * time.Hour,
		SnapshotRetention: 7,
		ValidatorBackups:  true,
		StateBackups:      true,
		KeyBackups:        false,
	}

	require.True(t, plan.Enabled)
	require.Equal(t, 6*time.Hour, plan.BackupInterval)
	require.Len(t, plan.BackupLocations, 2)
	require.Equal(t, 15*time.Minute, plan.RPO)
	require.Equal(t, 2*time.Hour, plan.RTO)
	require.True(t, plan.ValidatorBackups)
	require.False(t, plan.KeyBackups)
}

func TestInsuranceIntegration_Structure(t *testing.T) {
	integration := InsuranceIntegration{
		Enabled:         true,
		Provider:        "InsuranceCo",
		PolicyNumber:    "POL-12345",
		CoverageAmount:  "1000000000000",
		ClaimEndpoint:   "https://insurance.example.com/claims",
		RequiredSigners: []string{"admin1", "admin2", "admin3"},
		AutoClaim:       false,
		ClaimThreshold:  "1000000000000",
	}

	require.True(t, integration.Enabled)
	require.Equal(t, "InsuranceCo", integration.Provider)
	require.Equal(t, "POL-12345", integration.PolicyNumber)
	require.Len(t, integration.RequiredSigners, 3)
	require.False(t, integration.AutoClaim)
}

func TestCommunicationPlan_Structure(t *testing.T) {
	plan := CommunicationPlan{
		Enabled:              true,
		NotificationChannels: []string{"email", "sms", "telegram", "slack"},
		EscalationContacts: []Contact{
			{
				Name:     "John Doe",
				Role:     "Security Lead",
				Email:    "john@example.com",
				Phone:    "+1234567890",
				Telegram: "@johndoe",
				Priority: 1,
			},
			{
				Name:     "Jane Smith",
				Role:     "CTO",
				Email:    "jane@example.com",
				Phone:    "+0987654321",
				Priority: 2,
			},
		},
		StatusPageURL:  "https://status.example.com",
		UpdateInterval: 30 * time.Minute,
	}

	require.True(t, plan.Enabled)
	require.Len(t, plan.NotificationChannels, 4)
	require.Len(t, plan.EscalationContacts, 2)
	require.Equal(t, "Security Lead", plan.EscalationContacts[0].Role)
	require.Equal(t, 1, plan.EscalationContacts[0].Priority)
}

func TestErrors(t *testing.T) {
	errors := []error{
		ErrIncidentNotFound,
		ErrUnauthorizedPause,
		ErrChainAlreadyPaused,
		ErrChainNotPaused,
		ErrWalletLimitExceeded,
		ErrInsufficientSigners,
		ErrInvalidPauseLevel,
		ErrMaxPauseDurationExceeded,
		ErrPostMortemNotCompleted,
	}

	for _, err := range errors {
		require.NotNil(t, err)
		require.NotEmpty(t, err.Error())
	}
}
