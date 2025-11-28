package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/aequitas/aura/chain/x/identitychange/types"
)

// ============================
// IDENTITY ANALYTICS
// ============================

// IdentityAnalytics contains comprehensive analytics for an identity
type IdentityAnalytics struct {
	Did                   string
	ConfidenceScore       int64
	Reputation            int64
	TotalChanges          int64
	AccountAge            int64
	LastActivityHeight    int64
	RiskScore             int64
	TrustLevel            string
	PendingRequestsCount  int64
	AppliedRequestsCount  int64
	RejectedRequestsCount int64
}

// GetIdentityAnalytics returns comprehensive identity analytics
func (k *Keeper) GetIdentityAnalytics(ctx sdk.Context, did string) (*IdentityAnalytics, error) {
	record, exists := k.GetIdentityRecord(ctx, did)
	if !exists {
		return nil, fmt.Errorf("DID not found: %s", did)
	}

	// Calculate metrics
	changeCount := len(k.ListHistory(ctx, did))
	reputation := k.CalculateReputation(ctx, did)

	// Count requests by status
	pendingCount := int64(0)
	appliedCount := int64(0)
	rejectedCount := int64(0)

	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.RequestStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err == nil {
		defer iterator.Close()
		for ; iterator.Valid(); iterator.Next() {
			var req types.IdentityChangeRequest
			if err := k.cdc.Unmarshal(iterator.Value(), &req); err != nil {
				continue
			}
			if req.TargetDid == did {
				switch req.Status {
				case types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION:
					pendingCount++
				case types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED:
					appliedCount++
				case types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED:
					rejectedCount++
				}
			}
		}
	}

	accountAge := ctx.BlockHeight() - record.LastChangedHeight
	if accountAge < 0 {
		accountAge = 0
	}

	analytics := &IdentityAnalytics{
		Did:                   did,
		ConfidenceScore:       record.ConfidenceScore,
		Reputation:            reputation,
		TotalChanges:          int64(changeCount),
		AccountAge:            accountAge,
		LastActivityHeight:    record.LastChangedHeight,
		RiskScore:             k.calculateRiskScore(ctx, did),
		TrustLevel:            k.calculateTrustLevel(reputation),
		PendingRequestsCount:  pendingCount,
		AppliedRequestsCount:  appliedCount,
		RejectedRequestsCount: rejectedCount,
	}

	return analytics, nil
}

// calculateRiskScore calculates risk score based on activity patterns
func (k *Keeper) calculateRiskScore(ctx sdk.Context, did string) int64 {
	// Low score = low risk, high score = high risk
	riskScore := int64(0)

	// Frequent changes increase risk
	recentHistory := k.ListHistory(ctx, did)
	recentChanges := 0
	currentHeight := ctx.BlockHeight()

	// Count changes in the last 10000 blocks (approximately 1 day at 6s blocks)
	for _, h := range recentHistory {
		if currentHeight-h.ChangedHeight < 10000 {
			recentChanges++
		}
	}

	if recentChanges > 5 {
		riskScore += int64(recentChanges) * 5
	}

	// Check pending verifications
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.RequestStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err == nil {
		defer iterator.Close()
		for ; iterator.Valid(); iterator.Next() {
			var req types.IdentityChangeRequest
			if err := k.cdc.Unmarshal(iterator.Value(), &req); err != nil {
				continue
			}
			if req.TargetDid == did && req.Status == types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION {
				riskScore += 10
			}
		}
	}

	// Cap at 100
	if riskScore > 100 {
		riskScore = 100
	}

	return riskScore
}

// calculateTrustLevel categorizes trust level based on reputation
func (k *Keeper) calculateTrustLevel(reputation int64) string {
	if reputation >= 90 {
		return "high"
	} else if reputation >= 60 {
		return "medium"
	} else if reputation >= 30 {
		return "low"
	}
	return "untrusted"
}

// ============================
// IDENTITY REPUTATION
// ============================

// CalculateReputation calculates overall identity reputation
func (k *Keeper) CalculateReputation(ctx sdk.Context, did string) int64 {
	record, exists := k.GetIdentityRecord(ctx, did)
	if !exists {
		return 0
	}

	// Base reputation from confidence score
	reputation := record.ConfidenceScore

	// Penalty from recent changes (frequent changes reduce reputation)
	recentHistory := k.ListHistory(ctx, did)
	recentChanges := 0
	currentHeight := ctx.BlockHeight()

	// Count changes in the last 10000 blocks
	for _, h := range recentHistory {
		if currentHeight-h.ChangedHeight < 10000 {
			recentChanges++
		}
	}

	if recentChanges > 5 {
		reputation -= int64(recentChanges-5) * 10
	}

	// Ensure non-negative
	if reputation < 0 {
		reputation = 0
	}

	return reputation
}

// UpdateReputation updates identity reputation
func (k *Keeper) UpdateReputation(ctx sdk.Context, did string, delta int64) error {
	record, exists := k.GetIdentityRecord(ctx, did)
	if !exists {
		return fmt.Errorf("DID not found: %s", did)
	}

	record.ConfidenceScore += delta
	if record.ConfidenceScore < 0 {
		record.ConfidenceScore = 0
	}

	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return err
	}

	// Track the change in history
	history := types.IdentityChangeHistory{
		RequestId:           fmt.Sprintf("reputation_update_%d", ctx.BlockHeight()),
		TargetDid:           did,
		PrevConfidenceScore: record.ConfidenceScore - delta,
		NewConfidenceScore:  record.ConfidenceScore,
		TransitionReason:    fmt.Sprintf("Reputation changed by %d", delta),
		ChangedHeight:       ctx.BlockHeight(),
	}

	return k.AddHistory(ctx, history)
}

// ============================
// IDENTITY HISTORY TRACKING
// ============================

// GetIdentityHistoryPaginated returns paginated history for a DID
func (k *Keeper) GetIdentityHistoryPaginated(ctx sdk.Context, did string, limit int, offset int) []types.IdentityChangeHistory {
	allHistory := k.ListHistory(ctx, did)

	// Apply offset
	if offset >= len(allHistory) {
		return []types.IdentityChangeHistory{}
	}
	allHistory = allHistory[offset:]

	// Apply limit
	if limit > 0 && len(allHistory) > limit {
		allHistory = allHistory[:limit]
	}

	return allHistory
}

// TrackIdentityChange adds a change to history
func (k *Keeper) TrackIdentityChange(ctx sdk.Context, did, requestID, reason string) error {
	record, exists := k.GetIdentityRecord(ctx, did)
	if !exists {
		return fmt.Errorf("DID not found: %s", did)
	}

	historyEntry := types.IdentityChangeHistory{
		RequestId:           requestID,
		TargetDid:           did,
		PrevConfidenceScore: record.ConfidenceScore,
		NewConfidenceScore:  record.ConfidenceScore,
		TransitionReason:    reason,
		ChangedHeight:       ctx.BlockHeight(),
	}

	return k.AddHistory(ctx, historyEntry)
}

// ============================
// BATCH OPERATIONS
// ============================

// GetAllRecords retrieves all identity records
func (k *Keeper) GetAllRecords(ctx sdk.Context) []types.IdentityRecord {
	store := k.storeService.OpenKVStore(ctx)

	prefix := []byte(types.RecordStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return []types.IdentityRecord{}
	}
	defer iterator.Close()

	records := []types.IdentityRecord{}
	for ; iterator.Valid(); iterator.Next() {
		var record types.IdentityRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}

	return records
}

// GetAllRequests retrieves all identity change requests
func (k *Keeper) GetAllRequests(ctx sdk.Context) []types.IdentityChangeRequest {
	store := k.storeService.OpenKVStore(ctx)

	prefix := []byte(types.RequestStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return []types.IdentityChangeRequest{}
	}
	defer iterator.Close()

	requests := []types.IdentityChangeRequest{}
	for ; iterator.Valid(); iterator.Next() {
		var request types.IdentityChangeRequest
		if err := k.cdc.Unmarshal(iterator.Value(), &request); err != nil {
			continue
		}
		requests = append(requests, request)
	}

	return requests
}

// GetRequestsByStatus retrieves all requests with a specific status
func (k *Keeper) GetRequestsByStatus(ctx sdk.Context, status types.IdentityChangeStatus) []types.IdentityChangeRequest {
	allRequests := k.GetAllRequests(ctx)
	filtered := []types.IdentityChangeRequest{}

	for _, req := range allRequests {
		if req.Status == status {
			filtered = append(filtered, req)
		}
	}

	return filtered
}

// GetRequestsByDID retrieves all requests for a specific DID
func (k *Keeper) GetRequestsByDID(ctx sdk.Context, did string) []types.IdentityChangeRequest {
	allRequests := k.GetAllRequests(ctx)
	filtered := []types.IdentityChangeRequest{}

	for _, req := range allRequests {
		if req.TargetDid == did {
			filtered = append(filtered, req)
		}
	}

	return filtered
}

// ============================
// UTILITY FUNCTIONS
// ============================

// GenerateRequestID generates a unique request ID
func (k *Keeper) GenerateRequestID(ctx sdk.Context, requester, targetDID string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s-%s-%d-%d",
		requester,
		targetDID,
		ctx.BlockHeight(),
		time.Now().UnixNano())))
	return hex.EncodeToString(h.Sum(nil))[:32]
}

// ValidateIdentityOwnership verifies that an address owns a DID
func (k *Keeper) ValidateIdentityOwnership(ctx sdk.Context, did, owner string) error {
	record, exists := k.GetIdentityRecord(ctx, did)
	if !exists {
		return fmt.Errorf("DID not found: %s", did)
	}

	if record.Owner != owner {
		return fmt.Errorf("unauthorized: %s does not own DID %s", owner, did)
	}

	return nil
}

// IsRequestExpired checks if a request has expired based on params
func (k *Keeper) IsRequestExpired(ctx sdk.Context, request types.IdentityChangeRequest) bool {
	params := k.GetParams()

	if params.StalenessHeightThreshold <= 0 {
		return false
	}

	if request.Status != types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION &&
		request.Status != types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY {
		return false
	}

	currentHeight := ctx.BlockHeight()
	return currentHeight-request.CreatedHeight > params.StalenessHeightThreshold
}

// CleanupExpiredRequests removes expired requests
func (k *Keeper) CleanupExpiredRequests(ctx sdk.Context) error {
	allRequests := k.GetAllRequests(ctx)
	store := k.storeService.OpenKVStore(ctx)

	for _, req := range allRequests {
		if k.IsRequestExpired(ctx, req) {
			// Update status to rejected
			req.Status = types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED
			req.Reason = "Request expired due to staleness threshold"

			bz, err := k.cdc.Marshal(&req)
			if err != nil {
				continue
			}

			if err := store.Set([]byte(types.RequestStoreKey(req.RequestId)), bz); err != nil {
				continue
			}
		}
	}

	return nil
}

// ============================
// STATISTICS AND METRICS
// ============================

// GetModuleStats returns module-wide statistics
func (k *Keeper) GetModuleStats(ctx sdk.Context) map[string]int64 {
	stats := make(map[string]int64)

	allRequests := k.GetAllRequests(ctx)
	allRecords := k.GetAllRecords(ctx)

	stats["total_records"] = int64(len(allRecords))
	stats["total_requests"] = int64(len(allRequests))

	// Count by status
	stats["pending_requests"] = 0
	stats["ready_requests"] = 0
	stats["applied_requests"] = 0
	stats["rejected_requests"] = 0

	for _, req := range allRequests {
		switch req.Status {
		case types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION:
			stats["pending_requests"]++
		case types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY:
			stats["ready_requests"]++
		case types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED:
			stats["applied_requests"]++
		case types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED:
			stats["rejected_requests"]++
		}
	}

	// Calculate average confidence score
	totalConfidence := int64(0)
	for _, record := range allRecords {
		totalConfidence += record.ConfidenceScore
	}
	if len(allRecords) > 0 {
		stats["avg_confidence_score"] = totalConfidence / int64(len(allRecords))
	} else {
		stats["avg_confidence_score"] = 0
	}

	stats["suspended"] = 0
	if k.IsSuspended(ctx) {
		stats["suspended"] = 1
	}

	return stats
}

// GetTopIdentitiesByConfidence returns the top N identities by confidence score
func (k *Keeper) GetTopIdentitiesByConfidence(ctx sdk.Context, limit int) []types.IdentityRecord {
	allRecords := k.GetAllRecords(ctx)

	// Simple bubble sort for small datasets (production would use more efficient sorting)
	for i := 0; i < len(allRecords); i++ {
		for j := i + 1; j < len(allRecords); j++ {
			if allRecords[j].ConfidenceScore > allRecords[i].ConfidenceScore {
				allRecords[i], allRecords[j] = allRecords[j], allRecords[i]
			}
		}
	}

	if limit > 0 && len(allRecords) > limit {
		allRecords = allRecords[:limit]
	}

	return allRecords
}

// GetLowConfidenceIdentities returns identities below a threshold
func (k *Keeper) GetLowConfidenceIdentities(ctx sdk.Context, threshold int64) []types.IdentityRecord {
	allRecords := k.GetAllRecords(ctx)
	lowConfidence := []types.IdentityRecord{}

	for _, record := range allRecords {
		if record.ConfidenceScore < threshold {
			lowConfidence = append(lowConfidence, record)
		}
	}

	return lowConfidence
}
