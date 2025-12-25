// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/aequitas/aura/chain/x/incidentresponse/types"
	"github.com/stretchr/testify/require"
)

func TestQueryServer_GetIncident(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create test incident
	incidentID, err := keeper.ReportIncident(
		ctx,
		"Test Incident",
		"Test Description",
		types.SeverityHigh,
		"reporter",
		[]string{"system1"},
	)
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     *types.QueryGetIncidentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid query",
			req: &types.QueryGetIncidentRequest{
				IncidentId: incidentID,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "empty incident ID",
			req: &types.QueryGetIncidentRequest{
				IncidentId: "",
			},
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name: "nonexistent incident",
			req: &types.QueryGetIncidentRequest{
				IncidentId: "INC-999",
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetIncident(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Incident)
				require.Equal(t, "Test Incident", resp.Incident.Title)
			}
		})
	}
}

func TestQueryServer_GetAllIncidents(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Create test incidents with different severities and statuses
	incidents := []struct {
		title    string
		severity types.IncidentSeverity
		status   types.IncidentStatus
	}{
		{"Critical Incident", types.SeverityCritical, types.StatusNew},
		{"High Incident", types.SeverityHigh, types.StatusInvestigation},
		{"Medium Incident", types.SeverityMedium, types.StatusContained},
		{"Low Incident", types.SeverityLow, types.StatusResolved},
	}

	for _, inc := range incidents {
		id, err := keeper.ReportIncident(
			ctx,
			inc.title,
			"Description",
			inc.severity,
			"reporter",
			[]string{"system"},
		)
		require.NoError(t, err)

		// Update status if not New
		if inc.status != types.StatusNew {
			err = keeper.UpdateIncidentStatus(ctx, id, inc.status, "updater", "Update")
			require.NoError(t, err)
		}
	}

	tests := []struct {
		name          string
		req           *types.QueryGetAllIncidentsRequest
		wantErr       bool
		expectedCount int
	}{
		{
			name: "get all incidents",
			req: &types.QueryGetAllIncidentsRequest{
				Status:   "",
				Severity: "",
			},
			wantErr:       false,
			expectedCount: 4,
		},
		{
			name: "filter by status",
			req: &types.QueryGetAllIncidentsRequest{
				Status:   string(types.StatusInvestigation),
				Severity: "",
			},
			wantErr:       false,
			expectedCount: 1,
		},
		{
			name: "filter by severity",
			req: &types.QueryGetAllIncidentsRequest{
				Status:   "",
				Severity: string(types.SeverityCritical),
			},
			wantErr:       false,
			expectedCount: 1,
		},
		{
			name: "filter by both",
			req: &types.QueryGetAllIncidentsRequest{
				Status:   string(types.StatusNew),
				Severity: string(types.SeverityCritical),
			},
			wantErr:       false,
			expectedCount: 1,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetAllIncidents(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Len(t, resp.Incidents, tt.expectedCount)
			}
		})
	}
}

func TestQueryServer_GetChainPauseState(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	tests := []struct {
		name       string
		req        *types.QueryGetChainPauseStateRequest
		wantErr    bool
		isPaused   bool
		pauseLevel types.PauseLevel
	}{
		{
			name:       "query not paused state",
			req:        &types.QueryGetChainPauseStateRequest{},
			wantErr:    false,
			isPaused:   false,
			pauseLevel: types.PauseLevelNone,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetChainPauseState(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.PauseState)
				require.Equal(t, tt.isPaused, resp.PauseState.IsPaused)
				require.Equal(t, tt.pauseLevel, resp.PauseState.PauseLevel)
			}
		})
	}
}

func TestQueryServer_GetWalletLimits(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	// Set up wallet limits
	testAddress := "aura1abc123"
	err := keeper.SetWalletLimits(
		ctx,
		testAddress,
		"10000000000",
		"1000000000",
		"5000000000",
	)
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     *types.QueryGetWalletLimitsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid query",
			req: &types.QueryGetWalletLimitsRequest{
				Address: testAddress,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "empty address",
			req: &types.QueryGetWalletLimitsRequest{
				Address: "",
			},
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name: "nonexistent wallet",
			req: &types.QueryGetWalletLimitsRequest{
				Address: "aura1nonexistent",
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetWalletLimits(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Limits)
				require.Equal(t, testAddress, resp.Limits.Address)
				require.Equal(t, "10000000000", resp.Limits.MaxBalance)
			}
		})
	}
}

func TestQueryServer_GetParams(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	tests := []struct {
		name    string
		req     *types.QueryGetParamsRequest
		wantErr bool
	}{
		{
			name:    "valid query",
			req:     &types.QueryGetParamsRequest{},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetParams(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Params)
			}
		})
	}
}

func TestQueryServer_GetColdStorageConfig(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	tests := []struct {
		name    string
		req     *types.QueryGetColdStorageConfigRequest
		wantErr bool
	}{
		{
			name:    "valid query",
			req:     &types.QueryGetColdStorageConfigRequest{},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetColdStorageConfig(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Config)
			}
		})
	}
}

func TestQueryServer_GetBackupValidatorConfig(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	tests := []struct {
		name    string
		req     *types.QueryGetBackupValidatorConfigRequest
		wantErr bool
	}{
		{
			name:    "valid query",
			req:     &types.QueryGetBackupValidatorConfigRequest{},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetBackupValidatorConfig(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Config)
			}
		})
	}
}

func TestQueryServer_GetDisasterRecoveryPlan(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	tests := []struct {
		name    string
		req     *types.QueryGetDisasterRecoveryPlanRequest
		wantErr bool
	}{
		{
			name:    "valid query",
			req:     &types.QueryGetDisasterRecoveryPlanRequest{},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetDisasterRecoveryPlan(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Plan)
			}
		})
	}
}

func TestQueryServer_GetCommunicationPlan(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	tests := []struct {
		name    string
		req     *types.QueryGetCommunicationPlanRequest
		wantErr bool
	}{
		{
			name:    "valid query",
			req:     &types.QueryGetCommunicationPlanRequest{},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetCommunicationPlan(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Plan)
			}
		})
	}
}

func TestQueryServer_GetInsuranceIntegration(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServerImpl(keeper)

	tests := []struct {
		name    string
		req     *types.QueryGetInsuranceIntegrationRequest
		wantErr bool
	}{
		{
			name:    "valid query",
			req:     &types.QueryGetInsuranceIntegrationRequest{},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.GetInsuranceIntegration(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Integration)
			}
		})
	}
}
