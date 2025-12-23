package keeper

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/common/determinism"
	"github.com/aequitas/aura/chain/x/incidentresponse/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

// Keeper manages incident response functionality
// ALL STATE IS PERSISTED IN KV STORE - NO IN-MEMORY FALLBACKS
type KeeperKV struct {
	mu       sync.RWMutex
	store    *Store
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

// NewKeeperKV creates a new incident response keeper with KV persistence
func NewKeeperKV(storeKey storetypes.StoreKey, cdc codec.BinaryCodec) *KeeperKV {
	s := NewStore(storeKey, cdc)
	return &KeeperKV{
		store:    &s,
		storeKey: storeKey,
		cdc:      cdc,
	}
}

// requireStore panics if store is not initialized (production safety check)
func (k *KeeperKV) requireStore() {
	if k.store == nil {
		panic("incidentresponse keeper: KV store not initialized")
	}
}

// ========================================
// Feature 1: Incident Reporting & Tracking
// ========================================

// ReportIncident creates a new security incident
func (k *KeeperKV) ReportIncident(
	ctx context.Context,
	title, description string,
	severity types.IncidentSeverity,
	reportedBy string,
	affectedSystems []string,
) (string, error) {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	// Get next incident ID
	nextID := k.store.GetNextIncidentID(ctx)
	incidentID := fmt.Sprintf("INC-%d", nextID)
	k.store.SetNextIncidentID(ctx, nextID+1)

	// Get params to get response team
	params, ok := k.store.GetParams(ctx)
	if !ok {
		params = &types.IncidentResponseParams{}
	}

	now := determinism.GetBlockTime(ctx)
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
		ResponseTeam:    params.IncidentResponseTeam,
		Timeline: []types.IncidentTimelineEntry{
			{
				Timestamp:   now,
				Action:      "reported",
				Description: "Incident reported",
				Actor:       reportedBy,
			},
		},
	}

	if err := k.store.SetIncident(ctx, incident); err != nil {
		return "", err
	}

	// Auto-escalate critical incidents
	if severity == types.SeverityCritical {
		k.notifyEmergencyContacts(incident)
	}

	return incidentID, nil
}

// UpdateIncidentStatus updates the status of an incident
func (k *KeeperKV) UpdateIncidentStatus(
	ctx context.Context,
	incidentID string,
	status types.IncidentStatus,
	updatedBy, notes string,
) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	incident, exists := k.store.GetIncident(ctx, incidentID)
	if !exists {
		return types.ErrIncidentNotFound
	}

	now := determinism.GetBlockTime(ctx)
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

	return k.store.SetIncident(ctx, incident)
}

// GetIncident retrieves an incident by ID
func (k *KeeperKV) GetIncident(ctx context.Context, incidentID string) (*types.Incident, error) {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	incident, exists := k.store.GetIncident(ctx, incidentID)
	if !exists {
		return nil, types.ErrIncidentNotFound
	}
	return incident, nil
}

// GetAllIncidents returns all incidents
func (k *KeeperKV) GetAllIncidents(ctx context.Context) []*types.Incident {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.store.IterateIncidents(ctx)
}

// ========================================
// Feature 2: Emergency Chain Pause
// ========================================

// RequestChainPause initiates an emergency chain pause
func (k *KeeperKV) RequestChainPause(
	ctx context.Context,
	requester string,
	pauseLevel types.PauseLevel,
	reason string,
	incidentID string,
	duration time.Duration,
) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	params, ok := k.store.GetParams(ctx)
	if !ok {
		params = &types.IncidentResponseParams{}
	}

	if !params.EmergencyPauseEnabled {
		return fmt.Errorf("emergency pause is not enabled")
	}

	pauseState, ok := k.store.GetPauseState(ctx)
	if !ok {
		pauseState = &types.ChainPauseState{IsPaused: false, PauseLevel: types.PauseLevelNone}
	}

	if pauseState.IsPaused {
		return types.ErrChainAlreadyPaused
	}

	// Verify requester is authorized
	authorized := false
	for _, key := range params.PauseAuthorizedKeys {
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
	if duration > params.MaxPauseDuration {
		return types.ErrMaxPauseDurationExceeded
	}

	// Create pause request ID
	pauseRequestID := fmt.Sprintf("pause-%s-%d", requester, determinism.GetBlockTime(ctx).Unix())

	// Record vote
	k.store.SetPauseVote(ctx, pauseRequestID, requester)

	// Check if we have enough signers
	if params.PauseRequiredSigners == 1 {
		return k.executeChainPause(ctx, requester, pauseLevel, reason, incidentID, duration)
	}

	return nil
}

// ApproveChainPause approves a pending chain pause request
func (k *KeeperKV) ApproveChainPause(
	ctx context.Context,
	pauseRequestID string,
	approver string,
) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	params, ok := k.store.GetParams(ctx)
	if !ok {
		params = &types.IncidentResponseParams{}
	}

	// Verify approver is authorized
	authorized := false
	for _, key := range params.PauseAuthorizedKeys {
		if key == approver {
			authorized = true
			break
		}
	}
	if !authorized {
		return types.ErrUnauthorizedPause
	}

	votes := k.store.GetPauseVotes(ctx, pauseRequestID)
	if len(votes) == 0 {
		return fmt.Errorf("pause request not found")
	}

	// Check if already voted
	for _, voter := range votes {
		if voter == approver {
			return fmt.Errorf("already approved this pause request")
		}
	}

	// Add vote
	k.store.SetPauseVote(ctx, pauseRequestID, approver)

	// Get updated votes
	votes = k.store.GetPauseVotes(ctx, pauseRequestID)

	// Check if we have enough approvals
	if uint32(len(votes)) >= params.PauseRequiredSigners {
		// Execute pause (would need pause details from original request)
		// For simplicity, we'll execute a full pause
		return k.executeChainPause(ctx, approver, types.PauseLevelFull, "Multi-sig approved", "", 1*time.Hour)
	}

	return nil
}

// executeChainPause performs the actual chain pause
func (k *KeeperKV) executeChainPause(
	ctx context.Context,
	pausedBy string,
	pauseLevel types.PauseLevel,
	reason string,
	incidentID string,
	duration time.Duration,
) error {
	now := determinism.GetBlockTime(ctx)
	pauseState := &types.ChainPauseState{
		IsPaused:        true,
		PauseLevel:      pauseLevel,
		PausedAt:        now,
		PausedBy:        pausedBy,
		Reason:          reason,
		IncidentID:      incidentID,
		EstimatedResume: now.Add(duration),
	}

	if err := k.store.SetPauseState(ctx, pauseState); err != nil {
		return fmt.Errorf("error in executeChainPause for IncidentID: %w", err)
	}

	// If linked to an incident, update incident status
	if incidentID != "" {
		if incident, exists := k.store.GetIncident(ctx, incidentID); exists {
			incident.Timeline = append(incident.Timeline, types.IncidentTimelineEntry{
				Timestamp:   now,
				Action:      "chain_paused",
				Description: fmt.Sprintf("Chain paused at level %s", pauseLevel),
				Actor:       pausedBy,
			})
			if err := k.store.SetIncident(ctx, incident); err != nil {
				return fmt.Errorf("failed to update incident %s: %w", incidentID, err)
			}
		}
	}

	// Notify all stakeholders
	k.notifyChainPause(pauseLevel, reason)

	return nil
}

// ResumeChain resumes chain operations after a pause
func (k *KeeperKV) ResumeChain(ctx context.Context, resumedBy string, reason string) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	pauseState, ok := k.store.GetPauseState(ctx)
	if !ok || !pauseState.IsPaused {
		return types.ErrChainNotPaused
	}

	params, ok := k.store.GetParams(ctx)
	if !ok {
		params = &types.IncidentResponseParams{}
	}

	// Verify authorized
	authorized := false
	for _, key := range params.PauseAuthorizedKeys {
		if key == resumedBy {
			authorized = true
			break
		}
	}
	if !authorized {
		return types.ErrUnauthorizedPause
	}

	// Update pause state
	pauseState.IsPaused = false
	pauseState.PauseLevel = types.PauseLevelNone

	if err := k.store.SetPauseState(ctx, pauseState); err != nil {
		return fmt.Errorf("error in ResumeChain: %w", err)
	}

	// Update incident if linked
	if pauseState.IncidentID != "" {
		if incident, exists := k.store.GetIncident(ctx, pauseState.IncidentID); exists {
			incident.Timeline = append(incident.Timeline, types.IncidentTimelineEntry{
				Timestamp:   determinism.GetBlockTime(ctx),
				Action:      "chain_resumed",
				Description: reason,
				Actor:       resumedBy,
			})
			if err := k.store.SetIncident(ctx, incident); err != nil {
				return fmt.Errorf("failed to update incident %s: %w", pauseState.IncidentID, err)
			}
		}
	}

	// Notify all stakeholders
	k.notifyChainResume(reason)

	return nil
}

// GetChainPauseState returns the current pause state
func (k *KeeperKV) GetChainPauseState(ctx context.Context) *types.ChainPauseState {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	pauseState, ok := k.store.GetPauseState(ctx)
	if !ok {
		return &types.ChainPauseState{IsPaused: false, PauseLevel: types.PauseLevelNone}
	}
	return pauseState
}

// IsChainPaused checks if the chain is paused
func (k *KeeperKV) IsChainPaused(ctx context.Context) bool {
	pauseState := k.GetChainPauseState(ctx)
	return pauseState.IsPaused
}

// ========================================
// Feature 3: Hot Wallet Balance Limits
// ========================================

// SetWalletLimits configures limits for a hot wallet
func (k *KeeperKV) SetWalletLimits(
	ctx context.Context,
	address string,
	maxBalance, maxTransactionSize, dailyLimit string,
) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	limits := &types.WalletLimits{
		Address:            address,
		MaxBalance:         maxBalance,
		MaxTransactionSize: maxTransactionSize,
		DailyLimit:         dailyLimit,
		CurrentBalance:     "0",
		TodayTransferred:   "0",
		LastReset:          determinism.GetBlockTime(ctx),
	}

	return k.store.SetWalletLimit(ctx, limits)
}

// CheckWalletLimit validates a transaction against wallet limits
func (k *KeeperKV) CheckWalletLimit(ctx context.Context, address string, amount string, currentBalance string) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	params, ok := k.store.GetParams(ctx)
	if !ok || !params.HotWalletLimitsEnabled {
		return nil
	}

	limits, exists := k.store.GetWalletLimit(ctx, address)
	if !exists {
		// No specific limits, check global limits
		globalMax, _ := strconv.ParseInt(params.GlobalMaxHotWallet, 10, 64)
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
	now := determinism.GetBlockTime(ctx)
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
	if err := k.store.SetWalletLimit(ctx, limits); err != nil {
		return fmt.Errorf("failed to update wallet limits for %s: %w", address, err)
	}

	return nil
}

// GetWalletLimits retrieves wallet limits
func (k *KeeperKV) GetWalletLimits(ctx context.Context, address string) (*types.WalletLimits, error) {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	limits, exists := k.store.GetWalletLimit(ctx, address)
	if !exists {
		return nil, fmt.Errorf("wallet limits not found")
	}
	return limits, nil
}

// ========================================
// Feature 4: Cold Storage Management
// ========================================

// GetColdStorageConfig returns cold storage configuration
func (k *KeeperKV) GetColdStorageConfig(ctx context.Context) types.ColdStorageConfig {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	params, ok := k.store.GetParams(ctx)
	if !ok {
		return types.ColdStorageConfig{}
	}
	return params.ColdStorage
}

// ValidateColdStorageTransfer validates a transfer from cold storage
func (k *KeeperKV) ValidateColdStorageTransfer(ctx context.Context, signers []string, amount string) error {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	params, ok := k.store.GetParams(ctx)
	if !ok || !params.ColdStorage.Enabled {
		return fmt.Errorf("cold storage is not enabled")
	}

	// Verify minimum signers
	if uint32(len(signers)) < params.ColdStorage.MultiSigThreshold {
		return types.ErrInsufficientSigners
	}

	// Verify all signers are authorized
	for _, signer := range signers {
		authorized := false
		for _, authorizedSigner := range params.ColdStorage.MultiSigSigners {
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
func (k *KeeperKV) CreatePostMortem(
	ctx context.Context,
	incidentID string,
	createdBy string,
	summary, rootCause, impact, resolution string,
	lessonsLearned []string,
	actionItems []types.ActionItem,
) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	incident, exists := k.store.GetIncident(ctx, incidentID)
	if !exists {
		return types.ErrIncidentNotFound
	}

	postMortem := &types.PostMortem{
		CreatedAt:      determinism.GetBlockTime(ctx),
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

	return k.store.SetIncident(ctx, incident)
}

// CloseIncident closes an incident after post-mortem is complete
func (k *KeeperKV) CloseIncident(ctx context.Context, incidentID string, closedBy string) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	incident, exists := k.store.GetIncident(ctx, incidentID)
	if !exists {
		return types.ErrIncidentNotFound
	}

	if incident.PostMortem == nil {
		return types.ErrPostMortemNotCompleted
	}

	incident.Status = types.StatusClosed
	incident.Timeline = append(incident.Timeline, types.IncidentTimelineEntry{
		Timestamp:   determinism.GetBlockTime(ctx),
		Action:      "closed",
		Description: "Incident closed after post-mortem",
		Actor:       closedBy,
	})

	return k.store.SetIncident(ctx, incident)
}

// ========================================
// Feature 6: Backup & Recovery
// ========================================

// TriggerBackup initiates a backup operation
func (k *KeeperKV) TriggerBackup(ctx context.Context, backupType string, triggeredBy string) (string, error) {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	params, ok := k.store.GetParams(ctx)
	if !ok || !params.DisasterRecovery.Enabled {
		return "", fmt.Errorf("disaster recovery is not enabled")
	}

	backupID := fmt.Sprintf("backup-%s-%d", backupType, determinism.GetBlockTime(ctx).Unix())

	// In production, this would trigger actual backup operations
	// For now, we'll just return the backup ID

	return backupID, nil
}

// GetDisasterRecoveryPlan returns the disaster recovery configuration
func (k *KeeperKV) GetDisasterRecoveryPlan(ctx context.Context) types.DisasterRecoveryPlan {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	params, ok := k.store.GetParams(ctx)
	if !ok {
		return types.DisasterRecoveryPlan{}
	}
	return params.DisasterRecovery
}

// ========================================
// Feature 7: Validator Health Monitoring
// ========================================

// CheckValidatorHealth checks the health of validators
func (k *KeeperKV) CheckValidatorHealth(ctx context.Context) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	params, ok := k.store.GetParams(ctx)
	if !ok || !params.BackupValidators.Enabled {
		return nil
	}

	// Update last health check time
	params.BackupValidators.LastHealthCheck = determinism.GetBlockTime(ctx)
	if err := k.store.SetParams(ctx, params); err != nil {
		return fmt.Errorf("failed to update backup validator params: %w", err)
	}

	// In production, this would:
	// 1. Check validator uptime
	// 2. Check validator connectivity
	// 3. Trigger failover if needed
	// 4. Alert on issues

	return nil
}

// GetBackupValidatorConfig returns backup validator configuration
func (k *KeeperKV) GetBackupValidatorConfig(ctx context.Context) types.BackupValidatorConfig {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	params, ok := k.store.GetParams(ctx)
	if !ok {
		return types.BackupValidatorConfig{}
	}
	return params.BackupValidators
}

// ========================================
// Feature 8: Communication & Notifications
// ========================================

// notifyEmergencyContacts sends notifications to emergency contacts
func (k *KeeperKV) notifyEmergencyContacts(incident *types.Incident) {
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
func (k *KeeperKV) notifyChainPause(level types.PauseLevel, reason string) {
	// Send notifications to all configured channels
	fmt.Printf("Chain paused at level %s: %s\n", level, reason)
}

// notifyChainResume sends notifications about chain resume
func (k *KeeperKV) notifyChainResume(reason string) {
	// Send notifications to all configured channels
	fmt.Printf("Chain resumed: %s\n", reason)
}

// GetCommunicationPlan returns the communication plan
func (k *KeeperKV) GetCommunicationPlan(ctx context.Context) types.CommunicationPlan {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	params, ok := k.store.GetParams(ctx)
	if !ok {
		return types.CommunicationPlan{}
	}
	return params.Communication
}

// ========================================
// Feature 9: Insurance Integration
// ========================================

// TriggerInsuranceClaim triggers an insurance claim
func (k *KeeperKV) TriggerInsuranceClaim(
	ctx context.Context,
	incidentID string,
	amount string,
	signers []string,
) (string, error) {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	params, ok := k.store.GetParams(ctx)
	if !ok || !params.Insurance.Enabled {
		return "", fmt.Errorf("insurance integration is not enabled")
	}

	incident, exists := k.store.GetIncident(ctx, incidentID)
	if !exists {
		return "", types.ErrIncidentNotFound
	}

	// Verify required signers
	if len(signers) < len(params.Insurance.RequiredSigners) {
		return "", types.ErrInsufficientSigners
	}

	// In production, this would:
	// 1. Submit claim to insurance provider API
	// 2. Provide incident documentation
	// 3. Track claim status
	// 4. Handle claim approval/denial

	claimID := fmt.Sprintf("claim-%s-%d", incidentID, determinism.GetBlockTime(ctx).Unix())

	incident.Timeline = append(incident.Timeline, types.IncidentTimelineEntry{
		Timestamp:   determinism.GetBlockTime(ctx),
		Action:      "insurance_claim_submitted",
		Description: fmt.Sprintf("Insurance claim %s submitted for amount %s", claimID, amount),
		Actor:       "system",
	})

	if err := k.store.SetIncident(ctx, incident); err != nil {
		return "", fmt.Errorf("failed to persist insurance claim timeline for %s: %w", incidentID, err)
	}

	return claimID, nil
}

// GetInsuranceIntegration returns insurance integration config
func (k *KeeperKV) GetInsuranceIntegration(ctx context.Context) types.InsuranceIntegration {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	params, ok := k.store.GetParams(ctx)
	if !ok {
		return types.InsuranceIntegration{}
	}
	return params.Insurance
}

// GetParams returns the module parameters
func (k *KeeperKV) GetParams(ctx context.Context) types.IncidentResponseParams {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	params, ok := k.store.GetParams(ctx)
	if !ok {
		return types.DefaultParams()
	}
	return *params
}

// SetParams sets the module parameters
func (k *KeeperKV) SetParams(ctx context.Context, params types.IncidentResponseParams) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	return k.store.SetParams(ctx, &params)
}
