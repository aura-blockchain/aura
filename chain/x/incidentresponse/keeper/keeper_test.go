package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/incidentresponse/types"
	"github.com/stretchr/testify/require"
)

func TestNewKeeper(t *testing.T) {
	params := types.DefaultParams()
	k := NewKeeper(params)
	require.NotNil(t, k)
	require.False(t, k.IsChainPaused())
}

func TestReportIncident(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	incidentID, err := k.ReportIncident(
		"Database breach detected",
		"Unauthorized access to validator database",
		types.SeverityCritical,
		"security-team",
		[]string{"validator-db", "api-server"},
	)

	require.NoError(t, err)
	require.NotEmpty(t, incidentID)

	// Retrieve incident
	incident, err := k.GetIncident(incidentID)
	require.NoError(t, err)
	require.Equal(t, types.SeverityCritical, incident.Severity)
	require.Equal(t, types.StatusNew, incident.Status)
	require.Equal(t, "Database breach detected", incident.Title)
	require.Len(t, incident.Timeline, 1)
}

func TestUpdateIncidentStatus(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	// Create incident
	incidentID, _ := k.ReportIncident(
		"Test incident",
		"Test description",
		types.SeverityHigh,
		"admin",
		[]string{"system"},
	)

	// Update status
	err := k.UpdateIncidentStatus(
		incidentID,
		types.StatusInvestigation,
		"security-analyst",
		"Investigation started",
	)
	require.NoError(t, err)

	// Verify update
	incident, err := k.GetIncident(incidentID)
	require.NoError(t, err)
	require.Equal(t, types.StatusInvestigation, incident.Status)
	require.Len(t, incident.Timeline, 2)
}

func TestEmergencyChainPause(t *testing.T) {
	params := types.DefaultParams()
	params.PauseAuthorizedKeys = []string{"admin1", "admin2", "admin3"}
	params.PauseRequiredSigners = 1
	k := NewKeeper(params)

	// Request pause
	err := k.RequestChainPause(
		"admin1",
		types.PauseLevelFull,
		"Critical security vulnerability detected",
		"INC-001",
		1*time.Hour,
	)
	require.NoError(t, err)

	// Verify chain is paused
	require.True(t, k.IsChainPaused())
	pauseState := k.GetChainPauseState()
	require.Equal(t, types.PauseLevelFull, pauseState.PauseLevel)
	require.Equal(t, "admin1", pauseState.PausedBy)
}

func TestChainPauseUnauthorized(t *testing.T) {
	params := types.DefaultParams()
	params.PauseAuthorizedKeys = []string{"admin1"}
	k := NewKeeper(params)

	// Try to pause with unauthorized key
	err := k.RequestChainPause(
		"hacker",
		types.PauseLevelFull,
		"Malicious pause attempt",
		"",
		1*time.Hour,
	)
	require.Error(t, err)
	require.Equal(t, types.ErrUnauthorizedPause, err)
	require.False(t, k.IsChainPaused())
}

func TestChainPauseMultiSig(t *testing.T) {
	params := types.DefaultParams()
	params.PauseAuthorizedKeys = []string{"admin1", "admin2", "admin3"}
	params.PauseRequiredSigners = 2
	k := NewKeeper(params)

	// First approval - not enough
	err := k.RequestChainPause(
		"admin1",
		types.PauseLevelFull,
		"Test pause",
		"",
		1*time.Hour,
	)
	require.NoError(t, err)
	require.False(t, k.IsChainPaused()) // Need 2 signers

	// Note: In a real implementation, we'd track pause requests better
	// and test the multi-sig approval flow more thoroughly
	// For now, the test demonstrates the basic multi-sig concept
}

func TestResumeChain(t *testing.T) {
	params := types.DefaultParams()
	params.PauseAuthorizedKeys = []string{"admin1"}
	params.PauseRequiredSigners = 1
	k := NewKeeper(params)

	// Pause chain
	k.RequestChainPause(
		"admin1",
		types.PauseLevelFull,
		"Test pause",
		"",
		1*time.Hour,
	)
	require.True(t, k.IsChainPaused())

	// Resume chain
	err := k.ResumeChain("admin1", "Issue resolved")
	require.NoError(t, err)
	require.False(t, k.IsChainPaused())
}

func TestHotWalletLimits(t *testing.T) {
	params := types.DefaultParams()
	k := NewKeeper(params)

	// Set wallet limits
	err := k.SetWalletLimits(
		"wallet1",
		"1000000", // max balance
		"100000",  // max tx size
		"500000",  // daily limit
	)
	require.NoError(t, err)

	// Check valid transaction
	err = k.CheckWalletLimit("wallet1", "50000", "500000")
	require.NoError(t, err)

	// Check transaction exceeding max balance
	err = k.CheckWalletLimit("wallet1", "600000", "500000")
	require.Error(t, err)
	require.Equal(t, types.ErrWalletLimitExceeded, err)

	// Check transaction exceeding max tx size
	err = k.CheckWalletLimit("wallet1", "150000", "500000")
	require.Error(t, err)
	require.Equal(t, types.ErrWalletLimitExceeded, err)
}

func TestHotWalletDailyLimit(t *testing.T) {
	params := types.DefaultParams()
	k := NewKeeper(params)

	// Set wallet limits
	k.SetWalletLimits(
		"wallet1",
		"1000000",
		"100000",
		"200000", // daily limit
	)

	// First transaction
	err := k.CheckWalletLimit("wallet1", "100000", "500000")
	require.NoError(t, err)

	// Second transaction - should exceed daily limit
	err = k.CheckWalletLimit("wallet1", "150000", "500000")
	require.Error(t, err)
	require.Equal(t, types.ErrWalletLimitExceeded, err)
}

func TestColdStorageValidation(t *testing.T) {
	params := types.DefaultParams()
	params.ColdStorage.MultiSigSigners = []string{"signer1", "signer2", "signer3"}
	params.ColdStorage.MultiSigThreshold = 2
	k := NewKeeper(params)

	// Valid transfer with enough signers
	err := k.ValidateColdStorageTransfer(
		[]string{"signer1", "signer2"},
		"1000000",
	)
	require.NoError(t, err)

	// Invalid - not enough signers
	err = k.ValidateColdStorageTransfer(
		[]string{"signer1"},
		"1000000",
	)
	require.Error(t, err)
	require.Equal(t, types.ErrInsufficientSigners, err)

	// Invalid - unauthorized signer
	err = k.ValidateColdStorageTransfer(
		[]string{"signer1", "unauthorized"},
		"1000000",
	)
	require.Error(t, err)
}

func TestPostMortem(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	// Create incident
	incidentID, _ := k.ReportIncident(
		"Service outage",
		"API server down",
		types.SeverityHigh,
		"ops-team",
		[]string{"api-server"},
	)

	// Update to resolved
	k.UpdateIncidentStatus(incidentID, types.StatusResolved, "ops-team", "Service restored")

	// Create post-mortem
	actionItems := []types.ActionItem{
		{
			ID:          "action-1",
			Description: "Implement better monitoring",
			Assignee:    "ops-team",
			Priority:    "high",
			Status:      "pending",
			DueDate:     time.Now().Add(7 * 24 * time.Hour),
		},
	}

	err := k.CreatePostMortem(
		incidentID,
		"tech-lead",
		"Service outage due to memory leak",
		"Memory leak in API handler",
		"15 minute downtime, 1000 affected users",
		"Restarted service and deployed fix",
		[]string{"Implement memory monitoring", "Add auto-restart on OOM"},
		actionItems,
	)
	require.NoError(t, err)

	// Verify post-mortem
	incident, err := k.GetIncident(incidentID)
	require.NoError(t, err)
	require.NotNil(t, incident.PostMortem)
	require.Equal(t, "Memory leak in API handler", incident.PostMortem.RootCause)
	require.Len(t, incident.PostMortem.ActionItems, 1)
}

func TestCloseIncident(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	// Create incident
	incidentID, _ := k.ReportIncident(
		"Test incident",
		"Test description",
		types.SeverityLow,
		"admin",
		[]string{"system"},
	)

	// Try to close without post-mortem
	err := k.CloseIncident(incidentID, "admin")
	require.Error(t, err)
	require.Equal(t, types.ErrPostMortemNotCompleted, err)

	// Create post-mortem
	k.CreatePostMortem(
		incidentID,
		"admin",
		"Summary",
		"Root cause",
		"Impact",
		"Resolution",
		[]string{"Lesson 1"},
		[]types.ActionItem{},
	)

	// Now close incident
	err = k.CloseIncident(incidentID, "admin")
	require.NoError(t, err)

	// Verify closed
	incident, err := k.GetIncident(incidentID)
	require.NoError(t, err)
	require.Equal(t, types.StatusClosed, incident.Status)
}

func TestBackupTrigger(t *testing.T) {
	params := types.DefaultParams()
	params.DisasterRecovery.Enabled = true
	k := NewKeeper(params)

	backupID, err := k.TriggerBackup("state", "admin")
	require.NoError(t, err)
	require.NotEmpty(t, backupID)
}

func TestInsuranceClaim(t *testing.T) {
	params := types.DefaultParams()
	params.Insurance.Enabled = true
	params.Insurance.RequiredSigners = []string{"signer1", "signer2"}
	k := NewKeeper(params)

	// Create critical incident
	incidentID, _ := k.ReportIncident(
		"Major security breach",
		"Funds stolen",
		types.SeverityCritical,
		"security",
		[]string{"wallet"},
	)

	// Submit insurance claim
	claimID, err := k.TriggerInsuranceClaim(
		incidentID,
		"10000000",
		[]string{"signer1", "signer2"},
	)
	require.NoError(t, err)
	require.NotEmpty(t, claimID)

	// Verify incident timeline updated
	incident, _ := k.GetIncident(incidentID)
	hasClaimEntry := false
	for _, entry := range incident.Timeline {
		if entry.Action == "insurance_claim_submitted" {
			hasClaimEntry = true
			break
		}
	}
	require.True(t, hasClaimEntry)
}

func TestGetAllIncidents(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	// Create multiple incidents
	k.ReportIncident("Incident 1", "Desc 1", types.SeverityLow, "admin", []string{})
	k.ReportIncident("Incident 2", "Desc 2", types.SeverityMedium, "admin", []string{})
	k.ReportIncident("Incident 3", "Desc 3", types.SeverityHigh, "admin", []string{})

	incidents := k.GetAllIncidents()
	require.Len(t, incidents, 3)
}

func TestValidatorHealthCheck(t *testing.T) {
	params := types.DefaultParams()
	params.BackupValidators.Enabled = true
	k := NewKeeper(params)

	err := k.CheckValidatorHealth()
	require.NoError(t, err)

	config := k.GetBackupValidatorConfig()
	require.False(t, config.LastHealthCheck.IsZero())
}

func TestMaxPauseDuration(t *testing.T) {
	params := types.DefaultParams()
	params.PauseAuthorizedKeys = []string{"admin1"}
	params.PauseRequiredSigners = 1
	params.MaxPauseDuration = 1 * time.Hour
	k := NewKeeper(params)

	// Try to pause for longer than max duration
	err := k.RequestChainPause(
		"admin1",
		types.PauseLevelFull,
		"Test",
		"",
		25*time.Hour, // Exceeds max
	)
	require.Error(t, err)
	require.Equal(t, types.ErrMaxPauseDurationExceeded, err)
}

func TestInvalidPauseLevel(t *testing.T) {
	params := types.DefaultParams()
	params.PauseAuthorizedKeys = []string{"admin1"}
	k := NewKeeper(params)

	err := k.RequestChainPause(
		"admin1",
		types.PauseLevel("invalid"),
		"Test",
		"",
		1*time.Hour,
	)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidPauseLevel, err)
}

func TestGetParams(t *testing.T) {
	params := types.DefaultParams()
	k := NewKeeper(params)

	retrievedParams := k.GetParams()
	require.Equal(t, params.EmergencyPauseEnabled, retrievedParams.EmergencyPauseEnabled)
	require.Equal(t, params.MaxPauseDuration, retrievedParams.MaxPauseDuration)
}
