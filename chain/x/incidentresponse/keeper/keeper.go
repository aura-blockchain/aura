// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/incidentresponse/types"
)

// Keeper manages incident response functionality
type Keeper struct {
	mu sync.RWMutex

	// Incident storage
	incidents      map[string]*types.Incident
	nextIncidentID uint64

	// Chain pause state
	pauseState *types.ChainPauseState
	pauseVotes map[string][]string // pause request ID -> list of signers

	// Wallet limits
	walletLimits map[string]*types.WalletLimits

	// Parameters
	params types.IncidentResponseParams
}

// NewKeeper creates a new incident response keeper
func NewKeeper(params types.IncidentResponseParams) *Keeper {
	return &Keeper{
		incidents:      make(map[string]*types.Incident),
		pauseState:     &types.ChainPauseState{IsPaused: false, PauseLevel: types.PauseLevelNone},
		pauseVotes:     make(map[string][]string),
		walletLimits:   make(map[string]*types.WalletLimits),
		params:         params,
		nextIncidentID: 1,
	}
}

// ========================================
// Feature 1: Incident Reporting & Tracking
// ========================================

// ReportIncident creates a new security incident
func (k *Keeper) ReportIncident(
	ctx sdk.Context,
	title, description string,
	severity types.IncidentSeverity,
	reportedBy string,
	affectedSystems []string,
) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	incidentID := fmt.Sprintf("INC-%d", k.nextIncidentID)
	k.nextIncidentID++

	now := ctx.BlockTime()
	incident := &types.Incident{
		ID:              incidentID,
		Title:           title,
		Description:     description,
		Severity:        severity,
		Status:          types.StatusNew,
		ReportedBy:      reportedBy,
		ReportedAt:      now,
		UpdatedAt:       now,
		AffectedSystems: affectedSystems,
		ResponseTeam:    k.params.IncidentResponseTeam,
		Timeline: []types.IncidentTimelineEntry{
			{
				Timestamp:   now,
				Action:      "reported",
				Description: "Incident reported",
				Actor:       reportedBy,
			},
		},
	}

	k.incidents[incidentID] = incident

	// Auto-escalate critical incidents
	if severity == types.SeverityCritical {
		k.notifyEmergencyContacts(incident)
	}

	return incidentID, nil
}

// UpdateIncidentStatus updates the status of an incident
func (k *Keeper) UpdateIncidentStatus(
	ctx sdk.Context,
	incidentID string,
	status types.IncidentStatus,
	updatedBy, notes string,
) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	incident, exists := k.incidents[incidentID]
	if !exists {
		return types.ErrIncidentNotFound
	}

	now := ctx.BlockTime()
	incident.Status = status
	incident.UpdatedAt = now

	// Add timeline entry
	incident.Timeline = append(incident.Timeline, types.IncidentTimelineEntry{
		Timestamp:   now,
		Action:      string(status),
		Description: notes,
		Actor:       updatedBy,
	})

	// Mark resolved time
	if status == types.StatusResolved {
		incident.ResolvedAt = now
	}

	return nil
}

// GetIncident retrieves an incident by ID
func (k *Keeper) GetIncident(ctx sdk.Context, incidentID string) (*types.Incident, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	incident, exists := k.incidents[incidentID]
	if !exists {
		return nil, types.ErrIncidentNotFound
	}
	return incident, nil
}

// GetAllIncidents returns all incidents
func (k *Keeper) GetAllIncidents(ctx sdk.Context) []*types.Incident {
	k.mu.RLock()
	defer k.mu.RUnlock()

	incidents := make([]*types.Incident, 0, len(k.incidents))
	for _, incident := range k.incidents {
		incidents = append(incidents, incident)
	}
	return incidents
}

// ========================================
// Feature 2: Emergency Chain Pause
// ========================================

// RequestChainPause initiates an emergency chain pause
func (k *Keeper) RequestChainPause(
	ctx sdk.Context,
	requester string,
	pauseLevel types.PauseLevel,
	reason string,
	incidentID string,
	duration time.Duration,
) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.params.EmergencyPauseEnabled {
		return fmt.Errorf("emergency pause is not enabled")
	}

	if k.pauseState.IsPaused {
		return types.ErrChainAlreadyPaused
	}

	// Verify requester is authorized
	authorized := false
	for _, key := range k.params.PauseAuthorizedKeys {
		if key == requester {
			authorized = true
			break
		}
	}
	if !authorized {
		return types.ErrUnauthorizedPause
	}

	// Validate pause level
	if pauseLevel != types.PauseLevelFull &&
		pauseLevel != types.PauseLevelTransactions &&
		pauseLevel != types.PauseLevelModules {
		return types.ErrInvalidPauseLevel
	}

	// Validate duration
	if duration > k.params.MaxPauseDuration {
		return types.ErrMaxPauseDurationExceeded
	}

	// Create pause request ID
	pauseRequestID := fmt.Sprintf("pause-%s-%d", requester, ctx.BlockTime().Unix())

	// Record vote
	k.pauseVotes[pauseRequestID] = []string{requester}

	// Check if we have enough signers
	if k.params.PauseRequiredSigners == 1 {
		return k.executeChainPause(ctx, requester, pauseLevel, reason, incidentID, duration)
	}

	return nil
}

// ApproveChainPause approves a pending chain pause request
func (k *Keeper) ApproveChainPause(
	ctx sdk.Context,
	pauseRequestID string,
	approver string,
) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Verify approver is authorized
	authorized := false
	for _, key := range k.params.PauseAuthorizedKeys {
		if key == approver {
			authorized = true
			break
		}
	}
	if !authorized {
		return types.ErrUnauthorizedPause
	}

	votes, exists := k.pauseVotes[pauseRequestID]
	if !exists {
		return fmt.Errorf("pause request not found")
	}

	// Check if already voted
	for _, voter := range votes {
		if voter == approver {
			return fmt.Errorf("already approved this pause request")
		}
	}

	// Add vote
	k.pauseVotes[pauseRequestID] = append(votes, approver)

	// Check if we have enough approvals
	if uint32(len(k.pauseVotes[pauseRequestID])) >= k.params.PauseRequiredSigners {
		// Execute pause (would need pause details from original request)
		// For simplicity, we'll execute a full pause
		return k.executeChainPause(ctx, approver, types.PauseLevelFull, "Multi-sig approved", "", 1*time.Hour)
	}

	return nil
}

// executeChainPause performs the actual chain pause
func (k *Keeper) executeChainPause(
	ctx sdk.Context,
	pausedBy string,
	pauseLevel types.PauseLevel,
	reason string,
	incidentID string,
	duration time.Duration,
) error {
	now := ctx.BlockTime()
	k.pauseState = &types.ChainPauseState{
		IsPaused:        true,
		PauseLevel:      pauseLevel,
		PausedAt:        now,
		PausedBy:        pausedBy,
		Reason:          reason,
		IncidentID:      incidentID,
		EstimatedResume: now.Add(duration),
	}

	// If linked to an incident, update incident status
	if incidentID != "" {
		if incident, exists := k.incidents[incidentID]; exists {
			incident.Timeline = append(incident.Timeline, types.IncidentTimelineEntry{
				Timestamp:   now,
				Action:      "chain_paused",
				Description: fmt.Sprintf("Chain paused at level %s", pauseLevel),
				Actor:       pausedBy,
			})
		}
	}

	// Notify all stakeholders
	k.notifyChainPause(pauseLevel, reason)

	return nil
}

// ResumeChain resumes chain operations after a pause
func (k *Keeper) ResumeChain(ctx sdk.Context, resumedBy string, reason string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.pauseState.IsPaused {
		return types.ErrChainNotPaused
	}

	// Verify authorized
	authorized := false
	for _, key := range k.params.PauseAuthorizedKeys {
		if key == resumedBy {
			authorized = true
			break
		}
	}
	if !authorized {
		return types.ErrUnauthorizedPause
	}

	// Update pause state
	k.pauseState.IsPaused = false
	k.pauseState.PauseLevel = types.PauseLevelNone

	// Update incident if linked
	if k.pauseState.IncidentID != "" {
		if incident, exists := k.incidents[k.pauseState.IncidentID]; exists {
			incident.Timeline = append(incident.Timeline, types.IncidentTimelineEntry{
				Timestamp:   ctx.BlockTime(),
				Action:      "chain_resumed",
				Description: reason,
				Actor:       resumedBy,
			})
		}
	}

	// Notify all stakeholders
	k.notifyChainResume(reason)

	return nil
}

// GetChainPauseState returns the current pause state
func (k *Keeper) GetChainPauseState(ctx sdk.Context) *types.ChainPauseState {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.pauseState
}

// IsChainPaused checks if the chain is paused
func (k *Keeper) IsChainPaused() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.pauseState.IsPaused
}

// ========================================
// Feature 3: Hot Wallet Balance Limits
// ========================================

// SetWalletLimits configures limits for a hot wallet
func (k *Keeper) SetWalletLimits(
	ctx sdk.Context,
	address string,
	maxBalance, maxTransactionSize, dailyLimit string,
) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	limits := &types.WalletLimits{
		Address:            address,
		MaxBalance:         maxBalance,
		MaxTransactionSize: maxTransactionSize,
		DailyLimit:         dailyLimit,
		CurrentBalance:     "0",
		TodayTransferred:   "0",
		LastReset:          ctx.BlockTime(),
	}

	k.walletLimits[address] = limits
	return nil
}

// CheckWalletLimit validates a transaction against wallet limits
func (k *Keeper) CheckWalletLimit(ctx sdk.Context, address string, amount string, currentBalance string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.params.HotWalletLimitsEnabled {
		return nil
	}

	limits, exists := k.walletLimits[address]
	if !exists {
		// No specific limits, check global limits
		globalMax, _ := strconv.ParseInt(k.params.GlobalMaxHotWallet, 10, 64)
		currentBal, _ := strconv.ParseInt(currentBalance, 10, 64)
		transferAmt, _ := strconv.ParseInt(amount, 10, 64)

		if currentBal+transferAmt > globalMax {
			return types.ErrWalletLimitExceeded
		}
		return nil
	}

	// Check max balance
	maxBalance, _ := strconv.ParseInt(limits.MaxBalance, 10, 64)
	currentBal, _ := strconv.ParseInt(currentBalance, 10, 64)
	transferAmt, _ := strconv.ParseInt(amount, 10, 64)

	if currentBal+transferAmt > maxBalance {
		return types.ErrWalletLimitExceeded
	}

	// Check max transaction size
	maxTxSize, _ := strconv.ParseInt(limits.MaxTransactionSize, 10, 64)
	if transferAmt > maxTxSize {
		return types.ErrWalletLimitExceeded
	}

	// Check daily limit
	now := ctx.BlockTime()
	if now.Sub(limits.LastReset) > 24*time.Hour {
		limits.TodayTransferred = "0"
		limits.LastReset = now
	}

	todayTransferred, _ := strconv.ParseInt(limits.TodayTransferred, 10, 64)
	dailyLimit, _ := strconv.ParseInt(limits.DailyLimit, 10, 64)

	if todayTransferred+transferAmt > dailyLimit {
		return types.ErrWalletLimitExceeded
	}

	// Update transferred amount
	limits.TodayTransferred = strconv.FormatInt(todayTransferred+transferAmt, 10)

	return nil
}

// GetWalletLimits retrieves wallet limits
func (k *Keeper) GetWalletLimits(ctx sdk.Context, address string) (*types.WalletLimits, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	limits, exists := k.walletLimits[address]
	if !exists {
		return nil, fmt.Errorf("wallet limits not found")
	}
	return limits, nil
}

// ========================================
// Feature 4: Cold Storage Management
// ========================================

// GetColdStorageConfig returns cold storage configuration
func (k *Keeper) GetColdStorageConfig(ctx sdk.Context) types.ColdStorageConfig {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.params.ColdStorage
}

// ValidateColdStorageTransfer validates a transfer from cold storage
func (k *Keeper) ValidateColdStorageTransfer(signers []string, amount string) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if !k.params.ColdStorage.Enabled {
		return fmt.Errorf("cold storage is not enabled")
	}

	// Verify minimum signers
	if uint32(len(signers)) < k.params.ColdStorage.MultiSigThreshold {
		return types.ErrInsufficientSigners
	}

	// Verify all signers are authorized
	for _, signer := range signers {
		authorized := false
		for _, authorizedSigner := range k.params.ColdStorage.MultiSigSigners {
			if signer == authorizedSigner {
				authorized = true
				break
			}
		}
		if !authorized {
			return fmt.Errorf("unauthorized signer: %s", signer)
		}
	}

	return nil
}

// ========================================
// Feature 5: Post-Mortem Management
// ========================================

// CreatePostMortem creates a post-mortem analysis for an incident
func (k *Keeper) CreatePostMortem(
	ctx sdk.Context,
	incidentID string,
	createdBy string,
	summary, rootCause, impact, resolution string,
	lessonsLearned []string,
	actionItems []types.ActionItem,
) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	incident, exists := k.incidents[incidentID]
	if !exists {
		return types.ErrIncidentNotFound
	}

	postMortem := &types.PostMortem{
		CreatedAt:      ctx.BlockTime(),
		CreatedBy:      createdBy,
		Summary:        summary,
		RootCause:      rootCause,
		Impact:         impact,
		Resolution:     resolution,
		LessonsLearned: lessonsLearned,
		ActionItems:    actionItems,
	}

	incident.PostMortem = postMortem
	incident.Status = types.StatusPostMortem

	return nil
}

// CloseIncident closes an incident after post-mortem is complete
func (k *Keeper) CloseIncident(ctx sdk.Context, incidentID string, closedBy string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	incident, exists := k.incidents[incidentID]
	if !exists {
		return types.ErrIncidentNotFound
	}

	if incident.PostMortem == nil {
		return types.ErrPostMortemNotCompleted
	}

	incident.Status = types.StatusClosed
	incident.Timeline = append(incident.Timeline, types.IncidentTimelineEntry{
		Timestamp:   ctx.BlockTime(),
		Action:      "closed",
		Description: "Incident closed after post-mortem",
		Actor:       closedBy,
	})

	return nil
}

// ========================================
// Feature 6: Backup & Recovery
// ========================================

// TriggerBackup initiates a backup operation
func (k *Keeper) TriggerBackup(ctx sdk.Context, backupType string, triggeredBy string) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if !k.params.DisasterRecovery.Enabled {
		return "", fmt.Errorf("disaster recovery is not enabled")
	}

	backupID := fmt.Sprintf("backup-%s-%d", backupType, ctx.BlockTime().Unix())

	// In production, this would trigger actual backup operations
	// For now, we'll just return the backup ID

	return backupID, nil
}

// GetDisasterRecoveryPlan returns the disaster recovery configuration
func (k *Keeper) GetDisasterRecoveryPlan(ctx sdk.Context) types.DisasterRecoveryPlan {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.params.DisasterRecovery
}

// ========================================
// Feature 7: Validator Health Monitoring
// ========================================

// CheckValidatorHealth checks the health of validators
func (k *Keeper) CheckValidatorHealth(ctx sdk.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.params.BackupValidators.Enabled {
		return nil
	}

	// Update last health check time
	k.params.BackupValidators.LastHealthCheck = ctx.BlockTime()

	// In production, this would:
	// 1. Check validator uptime
	// 2. Check validator connectivity
	// 3. Trigger failover if needed
	// 4. Alert on issues

	return nil
}

// GetBackupValidatorConfig returns backup validator configuration
func (k *Keeper) GetBackupValidatorConfig(ctx sdk.Context) types.BackupValidatorConfig {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.params.BackupValidators
}

// ========================================
// Feature 8: Communication & Notifications
// ========================================

// notifyEmergencyContacts sends notifications to emergency contacts
func (k *Keeper) notifyEmergencyContacts(incident *types.Incident) {
	// In production, this would send actual notifications via:
	// - Email
	// - SMS
	// - Telegram
	// - PagerDuty
	// - Slack
	// - Status page updates

	// For now, we just log the notification
	fmt.Printf("Emergency notification sent for incident %s\n", incident.ID)
}

// notifyChainPause sends notifications about chain pause
func (k *Keeper) notifyChainPause(level types.PauseLevel, reason string) {
	// Send notifications to all configured channels
	fmt.Printf("Chain paused at level %s: %s\n", level, reason)
}

// notifyChainResume sends notifications about chain resume
func (k *Keeper) notifyChainResume(reason string) {
	// Send notifications to all configured channels
	fmt.Printf("Chain resumed: %s\n", reason)
}

// GetCommunicationPlan returns the communication plan
func (k *Keeper) GetCommunicationPlan(ctx sdk.Context) types.CommunicationPlan {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.params.Communication
}

// ========================================
// Feature 9: Insurance Integration
// ========================================

// TriggerInsuranceClaim triggers an insurance claim
func (k *Keeper) TriggerInsuranceClaim(
	ctx sdk.Context,
	incidentID string,
	amount string,
	signers []string,
) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.params.Insurance.Enabled {
		return "", fmt.Errorf("insurance integration is not enabled")
	}

	incident, exists := k.incidents[incidentID]
	if !exists {
		return "", types.ErrIncidentNotFound
	}

	// Verify required signers
	if len(signers) < len(k.params.Insurance.RequiredSigners) {
		return "", types.ErrInsufficientSigners
	}

	// In production, this would:
	// 1. Submit claim to insurance provider API
	// 2. Provide incident documentation
	// 3. Track claim status
	// 4. Handle claim approval/denial

	claimID := fmt.Sprintf("claim-%s-%d", incidentID, ctx.BlockTime().Unix())

	incident.Timeline = append(incident.Timeline, types.IncidentTimelineEntry{
		Timestamp:   ctx.BlockTime(),
		Action:      "insurance_claim_submitted",
		Description: fmt.Sprintf("Insurance claim %s submitted for amount %s", claimID, amount),
		Actor:       "system",
	})

	return claimID, nil
}

// GetInsuranceIntegration returns insurance integration config
func (k *Keeper) GetInsuranceIntegration(ctx sdk.Context) types.InsuranceIntegration {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.params.Insurance
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) (types.IncidentResponseParams, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.params, nil
}
