package keeper

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/incidentresponse/types"
)

func TestMsgServer_ReportIncident(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	tests := []struct {
		name    string
		msg     *types.MsgReportIncident
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid incident report",
			msg: &types.MsgReportIncident{
				Title:           "Database Breach",
				Description:     "Unauthorized access detected",
				Severity:        string(types.SeverityCritical),
				Reporter:        "security-team",
				AffectedSystems: []string{"db", "api"},
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "invalidargument",
		},
		{
			name: "empty title",
			msg: &types.MsgReportIncident{
				Title:           "",
				Description:     "Test",
				Severity:        string(types.SeverityHigh),
				Reporter:        "reporter",
				AffectedSystems: []string{"system"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.ReportIncident(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, strings.ToLower(err.Error()), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.IncidentId)

				// Verify incident was created
				incident, err := keeper.GetIncident(ctx, resp.IncidentId)
				require.NoError(t, err)
				require.Equal(t, tt.msg.Title, incident.Title)
				require.Equal(t, tt.msg.Description, incident.Description)
			}
		})
	}
}

func TestMsgServer_UpdateIncidentStatus(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	// Create an incident first
	incidentID, err := keeper.ReportIncident(
		ctx,
		"Test Incident",
		"Description",
		types.SeverityMedium,
		"reporter",
		[]string{"system"},
	)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     *types.MsgUpdateIncidentStatus
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid status update",
			msg: &types.MsgUpdateIncidentStatus{
				IncidentId: incidentID,
				Status:     string(types.StatusInvestigation),
				UpdatedBy:  "analyst",
				Notes:      "Investigation started",
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "nonexistent incident",
			msg: &types.MsgUpdateIncidentStatus{
				IncidentId: "INC-999",
				Status:     string(types.StatusResolved),
				UpdatedBy:  "analyst",
				Notes:      "Test",
			},
			wantErr: true,
			errMsg:  "incident not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.UpdateIncidentStatus(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				// Verify status was updated
				incident, err := keeper.GetIncident(ctx, tt.msg.IncidentId)
				require.NoError(t, err)
				require.Equal(t, types.IncidentStatus(tt.msg.Status), incident.Status)
			}
		})
	}
}

func TestMsgServer_RequestChainPause(t *testing.T) {
	tests := []struct {
		name    string
		msg     *types.MsgRequestChainPause
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pause request",
			msg: &types.MsgRequestChainPause{
				Requester:       "admin1",
				PauseLevel:      string(types.PauseLevelFull),
				Reason:          "Critical security issue",
				IncidentId:      "INC-001",
				DurationSeconds: 3600,
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "unauthorized requester",
			msg: &types.MsgRequestChainPause{
				Requester:       "hacker",
				PauseLevel:      string(types.PauseLevelFull),
				Reason:          "Test",
				IncidentId:      "INC-001",
				DurationSeconds: 3600,
			},
			wantErr: true,
			errMsg:  "permissiondenied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh keeper and context for each test case
			freshKeeper, freshCtx := setupKeeperForTest(t)
			freshMsgServer := NewMsgServerImpl(freshKeeper)

			// Setup params for this test case
			params := types.DefaultParams()
			params.EmergencyPauseEnabled = true
			params.PauseAuthorizedKeys = []string{"admin1", "admin2", "admin3"}
			params.PauseRequiredSigners = 1
			params.MaxPauseDuration = 24 * time.Hour
			err := freshKeeper.SetParams(freshCtx, params)
			require.NoError(t, err)

			resp, err := freshMsgServer.RequestChainPause(freshCtx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				if tt.errMsg == "permissiondenied" {
					require.Equal(t, codes.PermissionDenied, st.Code())
				} else {
					require.Equal(t, codes.InvalidArgument, st.Code())
				}
				require.Contains(t, strings.ToLower(err.Error()), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				// Verify chain is paused
				pauseState := freshKeeper.GetChainPauseState(freshCtx)
				require.True(t, pauseState.IsPaused)
				require.Equal(t, types.PauseLevel(tt.msg.PauseLevel), pauseState.PauseLevel)
			}
		})
	}
}

func TestMapUnauthorizedPause(t *testing.T) {
	err := mapUnauthorizedPause(types.ErrUnauthorizedPause)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}

func TestMsgServer_ResumeChain(t *testing.T) {
	tests := []struct {
		name    string
		msg     *types.MsgResumeChain
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid resume request",
			msg: &types.MsgResumeChain{
				Resumer: "admin1",
				Reason:  "Issue resolved",
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "invalidargument",
		},
		{
			name: "unauthorized resumer",
			msg: &types.MsgResumeChain{
				Resumer: "unauthorized",
				Reason:  "Test",
			},
			wantErr: true,
			errMsg:  "permissiondenied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			freshKeeper, freshCtx := setupKeeperForTest(t)
			msgServer := NewMsgServerImpl(freshKeeper)

			// Setup params with authorized keys
			params := types.DefaultParams()
			params.EmergencyPauseEnabled = true
			params.PauseAuthorizedKeys = []string{"admin1", "admin2"}
			params.PauseRequiredSigners = 1
			err := freshKeeper.SetParams(freshCtx, params)
			require.NoError(t, err)

			// Pause chain if a message is provided (nil message should still error)
			if tt.msg != nil {
				err = freshKeeper.RequestChainPause(freshCtx, "admin1", types.PauseLevelFull, "Test", "INC-001", 1*time.Hour)
				require.NoError(t, err)
			}

			resp, err := msgServer.ResumeChain(freshCtx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				if ok {
					if tt.errMsg == "permissiondenied" {
						require.Equal(t, codes.PermissionDenied, st.Code())
					} else {
						require.Equal(t, codes.InvalidArgument, st.Code())
					}
				}
				require.Contains(t, strings.ToLower(err.Error()), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				// Verify chain is resumed
				pauseState := freshKeeper.GetChainPauseState(freshCtx)
				require.False(t, pauseState.IsPaused)
			}
		})
	}
}

func TestMsgServer_SetWalletLimits(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	tests := []struct {
		name    string
		msg     *types.MsgSetWalletLimits
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid wallet limits",
			msg: &types.MsgSetWalletLimits{
				Address:            "aura1abc123",
				MaxBalance:         "10000000000",
				MaxTransactionSize: "1000000000",
				DailyLimit:         "5000000000",
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "empty address",
			msg: &types.MsgSetWalletLimits{
				Address:            "",
				MaxBalance:         "10000000000",
				MaxTransactionSize: "1000000000",
				DailyLimit:         "5000000000",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.SetWalletLimits(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				// Verify limits were set
				limits, err := keeper.GetWalletLimits(ctx, tt.msg.Address)
				require.NoError(t, err)
				require.Equal(t, tt.msg.MaxBalance, limits.MaxBalance)
			}
		})
	}
}

func TestMsgServer_CreatePostMortem(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	// Create an incident first
	incidentID, err := keeper.ReportIncident(
		ctx,
		"Test Incident",
		"Description",
		types.SeverityHigh,
		"reporter",
		[]string{"system"},
	)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     *types.MsgCreatePostMortem
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid post-mortem",
			msg: &types.MsgCreatePostMortem{
				IncidentId:     incidentID,
				Creator:        "analyst",
				Summary:        "Incident summary",
				RootCause:      "Configuration error",
				Impact:         "Minor service disruption",
				Resolution:     "Configuration fixed",
				LessonsLearned: []string{"Improve monitoring", "Better documentation"},
				ActionItems:    []types.ActionItem{},
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "nonexistent incident",
			msg: &types.MsgCreatePostMortem{
				IncidentId: "INC-999",
				Creator:    "analyst",
				Summary:    "Test",
				RootCause:  "Test",
				Impact:     "Test",
				Resolution: "Test",
			},
			wantErr: true,
			errMsg:  "incident not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.CreatePostMortem(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				// Verify post-mortem was created
				incident, err := keeper.GetIncident(ctx, tt.msg.IncidentId)
				require.NoError(t, err)
				require.NotNil(t, incident.PostMortem)
				require.Equal(t, tt.msg.Summary, incident.PostMortem.Summary)
			}
		})
	}
}

func TestMsgServer_CloseIncident(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	// Create an incident with post-mortem
	incidentID, err := keeper.ReportIncident(
		ctx,
		"Test Incident",
		"Description",
		types.SeverityMedium,
		"reporter",
		[]string{"system"},
	)
	require.NoError(t, err)

	err = keeper.CreatePostMortem(
		ctx,
		incidentID,
		"analyst",
		"Summary",
		"Root cause",
		"Impact",
		"Resolution",
		[]string{"Lesson 1"},
		[]types.ActionItem{},
	)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     *types.MsgCloseIncident
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid close incident",
			msg: &types.MsgCloseIncident{
				IncidentId: incidentID,
				Closer:     "admin",
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.CloseIncident(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				// Verify incident was closed
				incident, err := keeper.GetIncident(ctx, tt.msg.IncidentId)
				require.NoError(t, err)
				require.Equal(t, types.StatusClosed, incident.Status)
			}
		})
	}
}

func TestMsgServer_TriggerBackup(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	// Enable disaster recovery
	params := keeper.GetParams(ctx)
	params.DisasterRecovery.Enabled = true
	params.DisasterRecovery.BackupLocations = []string{"s3://backups"}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     *types.MsgTriggerBackup
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid backup trigger",
			msg: &types.MsgTriggerBackup{
				BackupType: "state",
				Requester:  "admin",
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.TriggerBackup(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.BackupId)
			}
		})
	}
}

func TestMsgServer_TriggerInsuranceClaim(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	// Create an incident first
	incidentID, err := keeper.ReportIncident(
		ctx,
		"Test Incident",
		"Description",
		types.SeverityCritical,
		"reporter",
		[]string{"system"},
	)
	require.NoError(t, err)

	// Enable insurance
	params := keeper.GetParams(ctx)
	params.Insurance.Enabled = true
	params.Insurance.RequiredSigners = []string{"admin1", "admin2"}
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     *types.MsgTriggerInsuranceClaim
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid insurance claim",
			msg: &types.MsgTriggerInsuranceClaim{
				IncidentId: incidentID,
				Amount:     "1000000000000",
				Signers:    []string{"admin1", "admin2"},
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "nonexistent incident",
			msg: &types.MsgTriggerInsuranceClaim{
				IncidentId: "INC-999",
				Amount:     "1000000000000",
				Signers:    []string{"admin1", "admin2"},
			},
			wantErr: true,
			errMsg:  "incident not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.TriggerInsuranceClaim(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.ClaimId)
			}
		})
	}
}
