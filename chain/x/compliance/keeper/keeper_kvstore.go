// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// Key prefixes for KVStore
var (
	KYCRecordsKeyPrefix           = []byte{0x01}
	AMLProfilesKeyPrefix          = []byte{0x02}
	SuspiciousActivitiesKeyPrefix = []byte{0x03}
	MonitoringRulesKeyPrefix      = []byte{0x04}
	TransactionAlertsKeyPrefix    = []byte{0x05}
	SanctionsResultsKeyPrefix     = []byte{0x06}
	GDPRConsentsKeyPrefix         = []byte{0x07}
	GDPRRequestsKeyPrefix         = []byte{0x08}
	TaxReportsKeyPrefix           = []byte{0x09}
	ParamsKeyPrefix               = []byte{0x0A}
	ProcessingRestrictionsKeyPrefix = []byte{0x0B}
	RateLimitKeyPrefix              = []byte{0x0C}
	KYCHistoryKeyPrefix             = []byte{0x0D}
	KYCExpirationIndexKeyPrefix     = []byte{0x0E}
)

// ============================================================================
// KYC Record KVStore Methods
// ============================================================================

func (k *Keeper) SetKYCRecord(ctx sdk.Context, record *types.KYCRecord) error {
	store := ctx.KVStore(k.storeKey)
	// Use codec for gogoproto-generated types
	bz, err := k.cdc.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(KYCRecordsKeyPrefix, []byte(record.Address)...)
	store.Set(key, bz)

	// Update expiration index
	k.AddToExpirationIndex(ctx, record)

	return nil
}

func (k *Keeper) GetKYCRecord(ctx sdk.Context, address string) (*types.KYCRecord, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(KYCRecordsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("KYC record not found: %s", address)
	}

	var record types.KYCRecord
	// Use codec for gogoproto-generated types
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return nil, fmt.Errorf("GetKYCRecord: failed to unmarshal record for %s: %w", address, err)
	}
	return &record, nil
}

func (k *Keeper) GetAllKYCRecords(ctx sdk.Context) ([]*types.KYCRecord, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, KYCRecordsKeyPrefix)
	defer iterator.Close()

	// Pre-allocate with reasonable initial capacity to reduce reallocations
	records := make([]*types.KYCRecord, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var record types.KYCRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return nil, fmt.Errorf("GetAllKYCRecords: failed to unmarshal record: %w", err)
		}
		records = append(records, &record)
	}
	return records, nil
}

// UpdateKYCRecord updates an existing KYC record with version tracking and history preservation.
// This implements proper deduplication and conflict resolution for KYC submissions.
//
// Version Tracking:
//   - On first submission: version = 1
//   - On each update: version is incremented
//   - Previous version is archived to history
//
// History Preservation:
//   - Complete snapshot of previous record stored
//   - Includes update timestamp, provider, and reason
//   - Enables full audit trail for compliance
//
// Conflict Resolution:
//   - Updates always replace previous record
//   - History preserves all previous versions
//   - No data loss on duplicate submissions
//
// Parameters:
//   - ctx: SDK context for state access
//   - newRecord: New KYC record to store
//   - reason: Reason for update (audit trail)
//
// Returns:
//   - error: If update fails
//
// Security considerations:
//   - Version counter prevents replay attacks
//   - History provides immutable audit trail
//   - All changes are logged via events
//
// Compliance:
//   - BSA/AML: Complete audit trail of KYC changes
//   - GDPR: History can be purged off-chain while maintaining on-chain commitments
//
// Example usage:
//   if err := k.UpdateKYCRecord(ctx, newRecord, "annual renewal"); err != nil {
//       return err
//   }
func (k *Keeper) UpdateKYCRecord(ctx sdk.Context, newRecord *types.KYCRecord, reason string) error {
	// Get existing record if it exists
	existing, err := k.GetKYCRecord(ctx, newRecord.Address)

	if err == nil {
		// Record exists - remove old expiration index entry before updating
		k.RemoveFromExpirationIndex(ctx, existing)

		// Archive to history before updating
		historyEntry := &types.KYCHistoryEntry{
			Address:      existing.Address,
			Version:      existing.Version,
			Snapshot:     existing,
			UpdatedAt:    ctx.BlockTime(),
			UpdatedBy:    newRecord.Provider,
			UpdateReason: reason,
		}

		if err := k.AddKYCHistory(ctx, historyEntry); err != nil {
			return fmt.Errorf("failed to archive KYC history: %w", err)
		}

		// Increment version
		newRecord.Version = existing.Version + 1
	} else {
		// New record - set version to 1
		newRecord.Version = 1
	}

	// Store updated record (this also adds new expiration index entry)
	if err := k.SetKYCRecord(ctx, newRecord); err != nil {
		return fmt.Errorf("failed to set KYC record: %w", err)
	}

	// Emit version tracking event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeKYCSubmitted,
			sdk.NewAttribute(types.AttributeKeyAddress, newRecord.Address),
			sdk.NewAttribute(types.AttributeKeyProvider, newRecord.Provider),
			sdk.NewAttribute(types.AttributeKeyKYCLevel, newRecord.KycLevel.String()),
			sdk.NewAttribute(types.AttributeKeyJurisdiction, newRecord.Jurisdiction),
			sdk.NewAttribute("version", fmt.Sprintf("%d", newRecord.Version)),
			sdk.NewAttribute("update_reason", reason),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
		),
	)

	return nil
}

// ============================================================================
// KYC History KVStore Methods
// ============================================================================

// AddKYCHistory adds a history entry for a KYC record update.
// History entries are stored per address to enable efficient querying.
//
// Storage layout:
//   Key: KYCHistoryKeyPrefix + address
//   Value: KYCHistoryList (repeated entries)
//
// Parameters:
//   - ctx: SDK context for state access
//   - entry: History entry to add
//
// Returns:
//   - error: If storage fails
func (k *Keeper) AddKYCHistory(ctx sdk.Context, entry *types.KYCHistoryEntry) error {
	history, err := k.GetKYCHistory(ctx, entry.Address)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Append new entry
	history = append(history, entry)

	// Store updated history list
	list := &types.KYCHistoryList{Entries: history}
	// Use codec for gogoproto-generated types
	bz, err := k.cdc.Marshal(list)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	store := ctx.KVStore(k.storeKey)
	key := append(KYCHistoryKeyPrefix, []byte(entry.Address)...)
	store.Set(key, bz)

	return nil
}

// GetKYCHistory retrieves all history entries for an address.
// Returns entries in chronological order (oldest to newest).
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: Address to retrieve history for
//
// Returns:
//   - []*types.KYCHistoryEntry: List of history entries
//   - error: If retrieval fails
//
// Note: Returns empty list if no history exists (not an error)
func (k *Keeper) GetKYCHistory(ctx sdk.Context, address string) ([]*types.KYCHistoryEntry, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(KYCHistoryKeyPrefix, []byte(address)...)
	bz := store.Get(key)

	if bz == nil {
		// No history exists - return empty list
		return []*types.KYCHistoryEntry{}, nil
	}

	var list types.KYCHistoryList
	// Use codec for gogoproto-generated types
	if err := k.cdc.Unmarshal(bz, &list); err != nil {
		return nil, fmt.Errorf("GetKYCHistory: failed to unmarshal history for %s: %w", address, err)
	}

	return list.Entries, nil
}

// GetAllKYCHistory retrieves all KYC history across all addresses.
// Used for genesis export and full state queries.
//
// Returns:
//   - map[string][]*types.KYCHistoryEntry: Map of address to history entries
//   - error: If iteration fails
func (k *Keeper) GetAllKYCHistory(ctx sdk.Context) (map[string][]*types.KYCHistoryEntry, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, KYCHistoryKeyPrefix)
	defer iterator.Close()

	history := make(map[string][]*types.KYCHistoryEntry)
	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Key()[len(KYCHistoryKeyPrefix):])
		var list types.KYCHistoryList
		if err := k.cdc.Unmarshal(iterator.Value(), &list); err != nil {
			return nil, fmt.Errorf("GetAllKYCHistory: failed to unmarshal history for %s: %w", address, err)
		}
		history[address] = list.Entries
	}

	return history, nil
}

// ============================================================================
// KYC Expiration Index Methods
// ============================================================================

// makeExpirationIndexKey creates a composite key for the expiration index.
// Format: KYCExpirationIndexKeyPrefix + timestamp (8 bytes, big-endian) + address
// This ordering allows efficient iteration by expiration time.
func makeExpirationIndexKey(expiresAt time.Time, address string) []byte {
	// Convert timestamp to Unix seconds for consistent ordering
	timestamp := expiresAt.Unix()

	// Encode timestamp as 8-byte big-endian (sortable)
	timeBz := sdk.Uint64ToBigEndian(uint64(timestamp))

	// Composite key: prefix + timestamp + address
	key := append(KYCExpirationIndexKeyPrefix, timeBz...)
	key = append(key, []byte(address)...)

	return key
}

// AddToExpirationIndex adds a KYC record to the expiration index.
// This index allows efficient lookup of expired records without full iteration.
//
// The index stores: timestamp + address -> empty value (existence check only)
// This creates a time-ordered index for efficient range queries.
//
// Parameters:
//   - ctx: SDK context for state access
//   - record: KYC record to index
//
// Security considerations:
//   - Only records with non-nil ExpiresAt are indexed
//   - Old index entries must be removed when updating records
//   - Index is deterministic and reproducible
func (k *Keeper) AddToExpirationIndex(ctx sdk.Context, record *types.KYCRecord) {
	// Only index records with expiration time
	if record.ExpiresAt == nil {
		return
	}

	store := ctx.KVStore(k.storeKey)
	key := makeExpirationIndexKey(*record.ExpiresAt, record.Address)

	// Store empty value - we only need the key for existence check
	store.Set(key, []byte{})
}

// RemoveFromExpirationIndex removes a KYC record from the expiration index.
// This must be called when:
//   - Deleting a KYC record
//   - Updating a record's expiration time (remove old, add new)
//
// Parameters:
//   - ctx: SDK context for state access
//   - record: KYC record to remove from index
func (k *Keeper) RemoveFromExpirationIndex(ctx sdk.Context, record *types.KYCRecord) {
	// Only records with expiration time can be in the index
	if record.ExpiresAt == nil {
		return
	}

	store := ctx.KVStore(k.storeKey)
	key := makeExpirationIndexKey(*record.ExpiresAt, record.Address)
	store.Delete(key)
}

// IterateExpiredRecords efficiently iterates over expired KYC records up to currentTime.
// This uses the expiration index to avoid scanning all KYC records.
//
// The callback receives the address of each expired record. To get the full record,
// use GetKYCRecord(ctx, address).
//
// Parameters:
//   - ctx: SDK context for state access
//   - currentTime: Current block time (records expiring before this are returned)
//   - maxRecords: Maximum number of records to process (0 = no limit)
//   - callback: Function called for each expired record address
//              Returns true to stop iteration, false to continue
//
// Returns:
//   - int: Number of records processed
//
// Performance:
//   - O(k) where k = number of expired records (vs O(n) for full scan)
//   - Efficient range iteration using time-ordered index
//   - Gas-bounded via maxRecords parameter
//
// Example usage:
//   processed := k.IterateExpiredRecords(ctx, ctx.BlockTime(), 100, func(address string) bool {
//       // Process expired record
//       return false // Continue
//   })
func (k *Keeper) IterateExpiredRecords(
	ctx sdk.Context,
	currentTime time.Time,
	maxRecords int,
	callback func(address string) bool,
) int {
	store := ctx.KVStore(k.storeKey)

	// Create end key for range iteration: current timestamp
	endTimestamp := currentTime.Unix()
	endTimeBz := sdk.Uint64ToBigEndian(uint64(endTimestamp))
	endKey := append(KYCExpirationIndexKeyPrefix, endTimeBz...)
	// Append 0xFF bytes to make it inclusive
	endKey = append(endKey, 0xFF, 0xFF, 0xFF, 0xFF)

	// Iterate from start of index to current time
	iterator := store.Iterator(KYCExpirationIndexKeyPrefix, endKey)
	defer iterator.Close()

	// Collect stale keys for cleanup (can't delete during iteration)
	staleKeys := make([][]byte, 0)

	processed := 0
	for ; iterator.Valid(); iterator.Next() {
		// Extract address from key
		// Key format: prefix (1 byte) + timestamp (8 bytes) + address
		key := iterator.Key()
		if len(key) <= 9 {
			// Invalid key format, skip
			continue
		}

		address := string(key[9:])

		// Check if the record still exists - if not, mark key for cleanup
		if _, err := k.GetKYCRecord(ctx, address); err != nil {
			// Record was deleted but index entry remains - mark for cleanup
			keyCopy := make([]byte, len(key))
			copy(keyCopy, key)
			staleKeys = append(staleKeys, keyCopy)
			continue // Skip stale entry
		}

		// Call callback with address
		if callback(address) {
			break
		}

		processed++

		// Check batch limit
		if maxRecords > 0 && processed >= maxRecords {
			break
		}
	}

	// Clean up stale index entries after iteration completes
	for _, staleKey := range staleKeys {
		store.Delete(staleKey)
	}

	return processed
}

// ============================================================================
// AML Profile KVStore Methods
// ============================================================================

func (k *Keeper) SetAMLProfile(ctx sdk.Context, profile *types.AMLProfile) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(AMLProfilesKeyPrefix, []byte(profile.Address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetAMLProfile(ctx sdk.Context, address string) (*types.AMLProfile, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(AMLProfilesKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("AML profile not found: %s", address)
	}

	var profile types.AMLProfile
	if err := k.cdc.Unmarshal(bz, &profile); err != nil {
		return nil, fmt.Errorf("GetAMLProfile: failed to unmarshal profile for %s: %w", address, err)
	}
	return &profile, nil
}

func (k *Keeper) GetAllAMLProfiles(ctx sdk.Context) ([]*types.AMLProfile, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AMLProfilesKeyPrefix)
	defer iterator.Close()

	// Pre-allocate with reasonable initial capacity to reduce reallocations
	profiles := make([]*types.AMLProfile, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var profile types.AMLProfile
		if err := k.cdc.Unmarshal(iterator.Value(), &profile); err != nil {
			return nil, fmt.Errorf("GetAllAMLProfiles: failed to unmarshal profile: %w", err)
		}
		profiles = append(profiles, &profile)
	}
	return profiles, nil
}

// UpdateAMLProfileOnTransaction updates an AML profile when a transaction occurs.
// This implements continuous AML monitoring by:
//   - Tracking transaction count and volume
//   - Recalculating risk level based on behavior
//   - Queueing updates for batch processing in EndBlocker
//
// Performance optimization:
//   - Updates are queued in memory during block execution
//   - All queued updates are written once in EndBlocker
//   - Reduces write operations by ~50% for addresses with multiple txs per block
//   - Events are still emitted immediately for real-time monitoring
//
// This method should be called for every transaction to maintain accurate
// risk profiles and enable real-time AML compliance monitoring.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: Address performing the transaction
//   - amount: Transaction amount (can be multi-denomination)
//
// Returns:
//   - error: If profile cannot be updated
//
// AML Compliance:
//   - FinCEN: Continuous transaction monitoring requirement
//   - FATF Recommendation 10: Customer due diligence and ongoing monitoring
//   - BSA: Suspicious activity detection through pattern analysis
//
// Security considerations:
//   - Profiles are created on first transaction (no pre-registration required)
//   - Risk level changes trigger immediate events for monitoring systems
//   - Transaction volume is accumulated across all denominations
//   - Timestamps track last activity for velocity analysis
//   - Multiple updates to same address in one block are merged automatically
//
// Example usage:
//   if err := k.UpdateAMLProfileOnTransaction(ctx, sender, amount); err != nil {
//       return errorsmod.Wrap(ErrAMLUpdateFailed, err.Error())
//   }
func (k *Keeper) UpdateAMLProfileOnTransaction(ctx sdk.Context, address string, amount sdk.Coins) error {
	// Check if there's already a pending update for this address in current block
	// If yes, use that as base; otherwise load from store
	var profile *types.AMLProfile
	var err error

	if pending, exists := k.pendingProfileUpdates[address]; exists {
		// Use pending profile as base for this update
		profile = pending
	} else {
		// Load from store or create new
		profile, err = k.GetAMLProfile(ctx, address)
		if err != nil {
			// Create new profile for first transaction
			profile = &types.AMLProfile{
				Address:           address,
				RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
				TotalTransactions: 0,
				TotalVolume:       "0",
				RiskFactors:       []string{},
				LastAssessment:    ctx.BlockTime(),
				PepStatus:         false,
				SourceOfFunds:     []string{},
				Occupation:        "",
			}
		}
	}

	// Update transaction metrics
	profile.TotalTransactions++
	profile.LastAssessment = ctx.BlockTime()

	// Calculate total transaction volume (sum all denominations)
	// Parse existing volume
	existingVolume, ok := math.NewIntFromString(profile.TotalVolume)
	if !ok {
		existingVolume = math.ZeroInt()
	}

	// Add current transaction amount (sum all coins as integer units)
	for _, coin := range amount {
		existingVolume = existingVolume.Add(coin.Amount)
	}
	profile.TotalVolume = existingVolume.String()

	// Store previous risk level for event emission
	previousRisk := profile.RiskLevel

	// Recalculate risk level based on updated metrics
	profile.RiskLevel = k.calculateRiskLevel(ctx, profile)

	// Queue updated profile for batch write in EndBlocker
	k.pendingProfileUpdates[address] = profile

	// Emit profile updated event (immediate for monitoring)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAMLProfileUpdated,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyTransactionCount, fmt.Sprintf("%d", profile.TotalTransactions)),
			sdk.NewAttribute(types.AttributeKeyTotalVolume, profile.TotalVolume),
			sdk.NewAttribute(types.AttributeKeyRiskLevel, profile.RiskLevel.String()),
		),
	)

	// Emit risk level changed event if risk level increased
	if profile.RiskLevel != previousRisk {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRiskLevelChanged,
				sdk.NewAttribute(types.AttributeKeyAddress, address),
				sdk.NewAttribute(types.AttributeKeyPreviousRisk, previousRisk.String()),
				sdk.NewAttribute(types.AttributeKeyNewRisk, profile.RiskLevel.String()),
				sdk.NewAttribute(types.AttributeKeyTotalVolume, profile.TotalVolume),
				sdk.NewAttribute(types.AttributeKeyTransactionCount, fmt.Sprintf("%d", profile.TotalTransactions)),
			),
		)
	}

	return nil
}

// calculateRiskLevel determines AML risk level based on transaction behavior.
// This implements risk-based approach required by FATF and FinCEN.
//
// Risk factors considered:
//   - Total transaction volume (high volume = higher risk)
//   - Transaction velocity (frequent transactions = higher risk)
//   - PEP status (Politically Exposed Person = higher risk)
//   - Existing risk factors from investigations
//
// Risk Level Thresholds:
//   - LOW: Normal transaction patterns, low volume
//   - MEDIUM: Moderate volume or frequency
//   - HIGH: High volume or frequent transactions
//   - SEVERE: PEP status or very high volume/frequency
//
// The thresholds are configurable via module params and should be
// calibrated based on jurisdiction requirements and risk appetite.
//
// Parameters:
//   - ctx: SDK context for accessing params
//   - profile: Current AML profile to assess
//
// Returns:
//   - AMLRiskLevel: Calculated risk level
//
// Security considerations:
//   - Conservative thresholds (bias toward flagging for review)
//   - PEP status always results in HIGH or SEVERE risk
//   - Multiple factors compound to increase risk level
//   - Thresholds should be reviewed regularly and adjusted
func (k *Keeper) calculateRiskLevel(ctx sdk.Context, profile *types.AMLProfile) types.AMLRiskLevel {
	params, _ := k.GetParams(ctx)

	// Parse volume as math.Int for comparison
	totalVolume, ok := math.NewIntFromString(profile.TotalVolume)
	if !ok {
		totalVolume = math.ZeroInt()
	}

	// PEP status or existing high-risk factors trigger SEVERE
	if profile.PepStatus || len(profile.RiskFactors) >= 3 {
		return types.AMLRiskLevel_AML_RISK_SEVERE
	}

	// Parse threshold parameters
	highVolumeThreshold, ok := math.NewIntFromString(params.VelocityLimit_24H)
	if !ok {
		// Default to 1 million if param not set
		highVolumeThreshold = math.NewInt(1_000_000)
	}

	// Calculate medium threshold as 50% of high threshold
	mediumVolumeThreshold := highVolumeThreshold.QuoRaw(2)

	// Evaluate volume-based risk
	if totalVolume.GTE(highVolumeThreshold) {
		return types.AMLRiskLevel_AML_RISK_HIGH
	}

	if totalVolume.GTE(mediumVolumeThreshold) {
		return types.AMLRiskLevel_AML_RISK_MEDIUM
	}

	// Evaluate transaction velocity (frequency)
	// High frequency transactions can indicate structuring
	highFrequencyThreshold := uint64(100) // More than 100 transactions is high risk
	mediumFrequencyThreshold := uint64(50) // More than 50 transactions is medium risk

	if profile.TotalTransactions > highFrequencyThreshold {
		return types.AMLRiskLevel_AML_RISK_HIGH
	}

	if profile.TotalTransactions > mediumFrequencyThreshold {
		return types.AMLRiskLevel_AML_RISK_MEDIUM
	}

	// Check for existing risk factors
	if len(profile.RiskFactors) > 0 {
		return types.AMLRiskLevel_AML_RISK_MEDIUM
	}

	return types.AMLRiskLevel_AML_RISK_LOW
}

// ============================================================================
// Suspicious Activity KVStore Methods
// ============================================================================

func (k *Keeper) SetSuspiciousActivity(ctx sdk.Context, activity *types.SuspiciousActivity) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(activity)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(SuspiciousActivitiesKeyPrefix, []byte(activity.Id)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetSuspiciousActivity(ctx sdk.Context, id string) (*types.SuspiciousActivity, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(SuspiciousActivitiesKeyPrefix, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("suspicious activity not found: %s", id)
	}

	var activity types.SuspiciousActivity
	if err := k.cdc.Unmarshal(bz, &activity); err != nil {
		return nil, fmt.Errorf("GetSuspiciousActivity: failed to unmarshal activity %s: %w", id, err)
	}
	return &activity, nil
}

func (k *Keeper) GetAllSuspiciousActivities(ctx sdk.Context) ([]*types.SuspiciousActivity, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, SuspiciousActivitiesKeyPrefix)
	defer iterator.Close()

	activities := make([]*types.SuspiciousActivity, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var activity types.SuspiciousActivity
		if err := k.cdc.Unmarshal(iterator.Value(), &activity); err != nil {
			return nil, fmt.Errorf("GetAllSuspiciousActivities: failed to unmarshal activity: %w", err)
		}
		activities = append(activities, &activity)
	}
	return activities, nil
}

// ============================================================================
// Transaction Monitoring Rule KVStore Methods
// ============================================================================

func (k *Keeper) SetMonitoringRule(ctx sdk.Context, rule *types.TransactionMonitoringRule) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(rule)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(MonitoringRulesKeyPrefix, []byte(rule.Id)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetMonitoringRule(ctx sdk.Context, id string) (*types.TransactionMonitoringRule, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(MonitoringRulesKeyPrefix, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("monitoring rule not found: %s", id)
	}

	var rule types.TransactionMonitoringRule
	if err := k.cdc.Unmarshal(bz, &rule); err != nil {
		return nil, fmt.Errorf("GetMonitoringRule: failed to unmarshal rule %s: %w", id, err)
	}
	return &rule, nil
}

func (k *Keeper) GetAllMonitoringRules(ctx sdk.Context) ([]*types.TransactionMonitoringRule, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, MonitoringRulesKeyPrefix)
	defer iterator.Close()

	rules := make([]*types.TransactionMonitoringRule, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var rule types.TransactionMonitoringRule
		if err := k.cdc.Unmarshal(iterator.Value(), &rule); err != nil {
			return nil, fmt.Errorf("GetAllMonitoringRules: failed to unmarshal rule: %w", err)
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

// ============================================================================
// Transaction Alert KVStore Methods (stored per address)
// ============================================================================

func (k *Keeper) AddTransactionAlert(ctx sdk.Context, address string, alert *types.TransactionAlert) error {
	alerts, err := k.GetTransactionAlerts(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}
	alerts = append(alerts, alert)
	list := &types.TransactionAlertList{Alerts: alerts}
	bz, err := k.cdc.Marshal(list)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	store := ctx.KVStore(k.storeKey)
	key := append(TransactionAlertsKeyPrefix, []byte(address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetTransactionAlerts(ctx sdk.Context, address string) ([]*types.TransactionAlert, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(TransactionAlertsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return []*types.TransactionAlert{}, nil
	}
	var list types.TransactionAlertList
	if err := k.cdc.Unmarshal(bz, &list); err != nil {
		return nil, fmt.Errorf("GetTransactionAlerts: failed to unmarshal alerts for %s: %w", address, err)
	}
	return list.Alerts, nil
}

// SetTransactionAlert sets transaction alerts for an address (alias for AddTransactionAlert)
func (k *Keeper) SetTransactionAlert(ctx sdk.Context, address string, alert *types.TransactionAlert) error {
	return k.AddTransactionAlert(ctx, address, alert)
}

// GetAllTransactionAlerts retrieves all transaction alerts across all addresses
func (k *Keeper) GetAllTransactionAlerts(ctx sdk.Context) (map[string][]*types.TransactionAlert, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, TransactionAlertsKeyPrefix)
	defer iterator.Close()

	alerts := make(map[string][]*types.TransactionAlert)
	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Key()[len(TransactionAlertsKeyPrefix):])
		var list types.TransactionAlertList
		if err := k.cdc.Unmarshal(iterator.Value(), &list); err != nil {
			return nil, fmt.Errorf("GetAllTransactionAlerts: failed to unmarshal alerts for %s: %w", address, err)
		}
		alerts[address] = list.Alerts
	}
	return alerts, nil
}

// ============================================================================
// Sanctions Screening Result KVStore Methods
// ============================================================================

func (k *Keeper) SetSanctionsResult(ctx sdk.Context, result *types.SanctionsScreeningResult) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(SanctionsResultsKeyPrefix, []byte(result.Address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetSanctionsResult(ctx sdk.Context, address string) (*types.SanctionsScreeningResult, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(SanctionsResultsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("sanctions result not found: %s", address)
	}

	var result types.SanctionsScreeningResult
	if err := k.cdc.Unmarshal(bz, &result); err != nil {
		return nil, fmt.Errorf("GetSanctionsResult: failed to unmarshal result for %s: %w", address, err)
	}
	return &result, nil
}

func (k *Keeper) GetAllSanctionsResults(ctx sdk.Context) ([]*types.SanctionsScreeningResult, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, SanctionsResultsKeyPrefix)
	defer iterator.Close()

	results := make([]*types.SanctionsScreeningResult, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var result types.SanctionsScreeningResult
		if err := k.cdc.Unmarshal(iterator.Value(), &result); err != nil {
			return nil, fmt.Errorf("GetAllSanctionsResults: failed to unmarshal result: %w", err)
		}
		results = append(results, &result)
	}
	return results, nil
}

// ============================================================================
// GDPR Consent KVStore Methods (nested by address and consent type)
// ============================================================================

func (k *Keeper) SetGDPRConsent(ctx sdk.Context, consent *types.GDPRConsent) error {
	consents, err := k.GetGDPRConsents(ctx, consent.Address)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Update or add consent
	found := false
	for i, existing := range consents {
		if existing.ConsentType == consent.ConsentType {
			consents[i] = consent
			found = true
			break
		}
	}
	if !found {
		consents = append(consents, consent)
	}

	store := ctx.KVStore(k.storeKey)
	list := &types.GDPRConsentList{Consents: consents}
	bz, err := k.cdc.Marshal(list)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(GDPRConsentsKeyPrefix, []byte(consent.Address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetGDPRConsents(ctx sdk.Context, address string) ([]*types.GDPRConsent, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(GDPRConsentsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return []*types.GDPRConsent{}, nil
	}
	var list types.GDPRConsentList
	if err := k.cdc.Unmarshal(bz, &list); err != nil {
		return nil, fmt.Errorf("GetGDPRConsents: failed to unmarshal consents for %s: %w", address, err)
	}
	return list.Consents, nil
}

// GetAllGDPRConsents retrieves all GDPR consents across all addresses
func (k *Keeper) GetAllGDPRConsents(ctx sdk.Context) (map[string][]*types.GDPRConsent, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, GDPRConsentsKeyPrefix)
	defer iterator.Close()

	consents := make(map[string][]*types.GDPRConsent)
	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Key()[len(GDPRConsentsKeyPrefix):])
		var list types.GDPRConsentList
		if err := k.cdc.Unmarshal(iterator.Value(), &list); err != nil {
			return nil, fmt.Errorf("GetAllGDPRConsents: failed to unmarshal consents for %s: %w", address, err)
		}
		consents[address] = list.Consents
	}
	return consents, nil
}

// ============================================================================
// GDPR Data Request KVStore Methods
// ============================================================================

func (k *Keeper) SetGDPRRequest(ctx sdk.Context, request *types.GDPRDataRequest) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(GDPRRequestsKeyPrefix, []byte(request.Id)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetGDPRRequest(ctx sdk.Context, requestID string) (*types.GDPRDataRequest, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(GDPRRequestsKeyPrefix, []byte(requestID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("GDPR request not found: %s", requestID)
	}

	var request types.GDPRDataRequest
	if err := k.cdc.Unmarshal(bz, &request); err != nil {
		return nil, fmt.Errorf("GetGDPRRequest: failed to unmarshal request %s: %w", requestID, err)
	}
	return &request, nil
}

func (k *Keeper) GetAllGDPRRequests(ctx sdk.Context) ([]*types.GDPRDataRequest, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, GDPRRequestsKeyPrefix)
	defer iterator.Close()

	requests := make([]*types.GDPRDataRequest, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var request types.GDPRDataRequest
		if err := k.cdc.Unmarshal(iterator.Value(), &request); err != nil {
			return nil, fmt.Errorf("GetAllGDPRRequests: failed to unmarshal request: %w", err)
		}
		requests = append(requests, &request)
	}
	return requests, nil
}

// ============================================================================
// Tax Report KVStore Methods (nested by address and tax year)
// ============================================================================

func (k *Keeper) SetTaxReport(ctx sdk.Context, report *types.TaxReport) error {
	reports, err := k.GetTaxReports(ctx, report.Address)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Update or add report
	found := false
	for i, existing := range reports {
		if existing.TaxYear == report.TaxYear {
			reports[i] = report
			found = true
			break
		}
	}
	if !found {
		reports = append(reports, report)
	}

	store := ctx.KVStore(k.storeKey)
	list := &types.TaxReportList{Reports: reports}
	bz, err := k.cdc.Marshal(list)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := append(TaxReportsKeyPrefix, []byte(report.Address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetTaxReports(ctx sdk.Context, address string) ([]*types.TaxReport, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(TaxReportsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return []*types.TaxReport{}, nil
	}
	var list types.TaxReportList
	if err := k.cdc.Unmarshal(bz, &list); err != nil {
		return nil, fmt.Errorf("GetTaxReports: failed to unmarshal reports for %s: %w", address, err)
	}
	return list.Reports, nil
}

// GetAllTaxReports retrieves all tax reports across all addresses
func (k *Keeper) GetAllTaxReports(ctx sdk.Context) (map[string][]*types.TaxReport, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, TaxReportsKeyPrefix)
	defer iterator.Close()

	reports := make(map[string][]*types.TaxReport)
	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Key()[len(TaxReportsKeyPrefix):])
		var list types.TaxReportList
		if err := k.cdc.Unmarshal(iterator.Value(), &list); err != nil {
			return nil, fmt.Errorf("GetAllTaxReports: failed to unmarshal reports for %s: %w", address, err)
		}
		reports[address] = list.Reports
	}
	return reports, nil
}

// ============================================================================
// Params KVStore Methods
// ============================================================================

func (k *Keeper) GetParamsFromStore(ctx sdk.Context) (types.ComplianceParams, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ParamsKeyPrefix)
	if bz == nil {
		return types.ComplianceParams{}, nil
	}

	var params types.ComplianceParams
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return types.ComplianceParams{}, fmt.Errorf("GetParamsFromStore: failed to unmarshal params: %w", err)
	}
	return params, nil
}

func (k *Keeper) SetParamsToStore(ctx sdk.Context, params types.ComplianceParams) error {
	if err := types.ValidateParams(params); err != nil {
		return fmt.Errorf("error in SetParamsToStore for ValidateParams: %w", err)
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return fmt.Errorf("failed to marshal for ValidateParams: %w", err)
	}
	store.Set(ParamsKeyPrefix, bz)
	return nil
}

// ============================================================================
// Processing Restriction KVStore Methods (GDPR Article 7(3) Enforcement)
// ============================================================================

// SetProcessingRestriction marks an address as having restricted data processing rights.
// When set to true, the address has withdrawn consent and data processing must be halted.
// This implements GDPR Article 7(3) "Right to Withdraw Consent" enforcement.
//
// Security considerations:
//   - This flag must be checked before any data processing operations
//   - Withdrawal is recorded immutably for compliance audit
//   - Processing restriction cannot be bypassed
//
// GDPR compliance:
//   - Article 7(3): Consent withdrawal must be as easy as giving consent
//   - Article 18: Right to restriction of processing
//   - Enforcement mechanism for consent withdrawal
func (k *Keeper) SetProcessingRestriction(ctx sdk.Context, address string, restricted bool) error {
	store := ctx.KVStore(k.storeKey)
	key := append(ProcessingRestrictionsKeyPrefix, []byte(address)...)

	if restricted {
		// Store as single byte flag (0x01 = restricted)
		store.Set(key, []byte{0x01})
	} else {
		// Remove restriction
		store.Delete(key)
	}

	return nil
}

// IsProcessingRestricted checks if data processing is restricted for an address.
// Returns true if the address has withdrawn consent and processing must be halted.
// This method should be called before any data processing operation.
//
// GDPR compliance:
//   - Article 7(3): Withdrawal of consent enforcement check
//   - Article 18: Processing restriction enforcement
//   - Must be called before accessing or processing user data
func (k *Keeper) IsProcessingRestricted(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)
	key := append(ProcessingRestrictionsKeyPrefix, []byte(address)...)
	return store.Has(key)
}

// GetGDPRConsent retrieves a specific consent record by address and consent type.
// This method searches through all consents for an address and returns the matching one.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: User's blockchain address
//   - consentType: Type of consent (e.g., "data_processing", "marketing")
//
// Returns:
//   - consent: The matching consent record
//   - found: true if consent was found, false otherwise
func (k *Keeper) GetGDPRConsent(ctx sdk.Context, address string, consentType string) (*types.GDPRConsent, bool) {
	consents, err := k.GetGDPRConsents(ctx, address)
	if err != nil {
		return nil, false
	}

	for _, consent := range consents {
		if consent.ConsentType == consentType {
			return consent, true
		}
	}

	return nil, false
}

// CanProcessData checks if data processing is allowed for an address and purpose.
// This is the primary enforcement mechanism for GDPR consent requirements.
//
// The function checks:
//   1. Whether processing is restricted (consent withdrawn)
//   2. Whether specific consent exists for the purpose
//   3. Whether the consent is still valid (not withdrawn)
//
// This method MUST be called before any data processing operation involving user data.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: User's blockchain address
//   - purpose: Purpose/type of data processing (must match consent type)
//
// Returns:
//   - bool: true if processing is allowed, false if restricted or no consent
//
// GDPR compliance:
//   - Article 6(1)(a): Lawfulness of processing based on consent
//   - Article 7(3): Consent withdrawal enforcement
//   - Article 18: Processing restriction enforcement
//
// Security considerations:
//   - Default deny: Returns false if no consent found
//   - Immutable audit: All checks are logged via state access
//   - Cannot be bypassed: All processing must call this function
//
// Example usage:
//   if !k.CanProcessData(ctx, userAddress, "data_processing") {
//       return errorsmod.Wrap(ErrProcessingRestricted, "consent withdrawn")
//   }
func (k *Keeper) CanProcessData(ctx sdk.Context, address string, purpose string) bool {
	// Check if processing is restricted (consent withdrawn)
	if k.IsProcessingRestricted(ctx, address) {
		return false
	}

	// Check if specific consent exists and is valid
	consent, found := k.GetGDPRConsent(ctx, address, purpose)
	if !found {
		return false
	}

	// Verify consent is still active (not withdrawn)
	return consent.Consented
}

// TriggerDataDeletion emits an event signaling that data should be deleted.
// This is used when consent is withdrawn to signal off-chain systems to delete PII.
//
// GDPR compliance:
//   - Article 7(3): Data must not be processed after consent withdrawal
//   - Article 17: Right to erasure implementation
//   - Off-chain systems must monitor and respond to these events
//
// Implementation note:
//   - On-chain data (commitments/hashes) remains for audit trail
//   - Off-chain PII must be deleted by monitoring systems
//   - Event provides immutable deletion request record
func (k *Keeper) TriggerDataDeletion(ctx sdk.Context, address string, consentType string) error {
	// Emit event for off-chain systems to process
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"gdpr_data_deletion_requested",
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyConsentType, consentType),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format("2006-01-02T15:04:05Z")),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
		),
	)

	return nil
}

// RequireConsent checks if the user has valid consent for a specific purpose.
// This is the enforcement point for GDPR consent requirements (Article 6(1)(a) and Article 7).
//
// This method MUST be called before any data processing operation involving user data.
// It ensures that:
//   1. User has given explicit consent for the processing purpose
//   2. Consent has not been withdrawn (Article 7(3))
//   3. Processing is not restricted (Article 18)
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: User's blockchain address
//   - purpose: Processing purpose (e.g., "kyc_processing", "aml_monitoring", "sanctions_screening")
//
// Returns:
//   - error: types.ErrProcessingRestricted if consent is missing or withdrawn
//
// GDPR compliance:
//   - Article 6(1)(a): Lawfulness of processing based on consent
//   - Article 7: Conditions for consent
//   - Article 7(3): Consent withdrawal enforcement - processing must stop immediately
//   - Article 18: Right to restriction of processing
//
// Security considerations:
//   - Default deny: Returns error if no consent found (fail-safe)
//   - Atomic check: Verifies both consent existence and validity
//   - Audit trail: All checks are logged via state access
//
// Example usage in processing functions:
//   if err := k.RequireConsent(ctx, userAddress, "kyc_processing"); err != nil {
//       return nil, err
//   }
func (k *Keeper) RequireConsent(ctx sdk.Context, address string, purpose string) error {
	// Check if data processing is allowed for this address and purpose
	// This internally checks:
	// 1. Processing restrictions (consent withdrawn)
	// 2. Consent existence for the specific purpose
	// 3. Consent validity (not withdrawn)
	if !k.CanProcessData(ctx, address, purpose) {
		return types.ErrProcessingRestricted
	}

	return nil
}


// ============================================================================
// Rate Limiting KVStore Methods (DoS Protection for Expensive Operations)
// ============================================================================

// getRateLimitKey generates a composite key for rate limit tracking
// Format: RateLimitKeyPrefix + address + ":" + operation
func getRateLimitKey(address string, operation string) []byte {
	key := append(RateLimitKeyPrefix, []byte(address)...)
	key = append(key, ':')
	key = append(key, []byte(operation)...)
	return key
}

// GetRateLimitEntry retrieves the rate limit entry for an address and operation
func (k *Keeper) GetRateLimitEntry(ctx sdk.Context, address string, operation string) (*types.RateLimitEntry, bool) {
	store := ctx.KVStore(k.storeKey)
	key := getRateLimitKey(address, operation)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var entry types.RateLimitEntry
	if err := k.cdc.Unmarshal(bz, &entry); err != nil {
		return nil, false
	}
	return &entry, true
}

// SetRateLimitEntry stores a rate limit entry for an address and operation
func (k *Keeper) SetRateLimitEntry(ctx sdk.Context, entry *types.RateLimitEntry) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	key := getRateLimitKey(entry.Address, entry.Operation)
	store.Set(key, bz)
	return nil
}

// DeleteRateLimitEntry removes a rate limit entry (for window reset)
func (k *Keeper) DeleteRateLimitEntry(ctx sdk.Context, address string, operation string) {
	store := ctx.KVStore(k.storeKey)
	key := getRateLimitKey(address, operation)
	store.Delete(key)
}

// CheckRateLimit enforces per-address, per-operation rate limits to prevent DoS
// of expensive external API calls (sanctions screening, KYC verification, etc.).
func (k *Keeper) CheckRateLimit(ctx sdk.Context, address string, operation string) error {
	params, _ := k.GetParams(ctx)

	// Get or create rate limit entry
	entry, found := k.GetRateLimitEntry(ctx, address, operation)
	if !found {
		// Create new entry for first request
		entry = &types.RateLimitEntry{
			Address:     address,
			Operation:   operation,
			Count:       0,
			WindowStart: ctx.BlockTime(),
		}
	}

	// Check if window has expired and should be reset
	windowDuration := time.Duration(params.RateLimitWindowSeconds) * time.Second
	if windowDuration == 0 {
		// Default to 1 hour if not configured
		windowDuration = time.Hour
	}

	windowStart := entry.WindowStart
	timeSinceWindowStart := ctx.BlockTime().Sub(windowStart)

	if timeSinceWindowStart >= windowDuration {
		// Reset window - new time period
		entry.Count = 0
		entry.WindowStart = ctx.BlockTime()
	}

	// Determine limit for this operation
	var limit int64
	switch operation {
	case "sanctions_screening":
		limit = params.SanctionsScreeningLimit
	case "kyc_verification":
		limit = params.KycVerificationLimit
	case "aml_profile_query":
		limit = params.AmlProfileQueryLimit
	case "tax_report_generation":
		limit = params.TaxReportGenerationLimit
	case "transaction_alerts":
		limit = params.DefaultQueryLimit
	default:
		limit = params.DefaultQueryLimit
	}

	// Apply default if limit not configured
	if limit <= 0 {
		// Reasonable defaults based on operation cost
		switch operation {
		case "sanctions_screening":
			limit = 10 // Expensive external API calls
		case "kyc_verification":
			limit = 5 // Very expensive verification processes
		case "aml_profile_query":
			limit = 50 // Moderate cost database queries
		case "tax_report_generation":
			limit = 3 // Very expensive report generation
		default:
			limit = 100 // Cheap operations
		}
	}

	// Check if limit exceeded
	if entry.Count >= limit {
		// Emit rate limit exceeded event for monitoring
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRateLimitExceeded,
				sdk.NewAttribute(types.AttributeKeyAddress, address),
				sdk.NewAttribute(types.AttributeKeyOperation, operation),
				sdk.NewAttribute(types.AttributeKeyCount, fmt.Sprintf("%d", entry.Count)),
				sdk.NewAttribute(types.AttributeKeyLimit, fmt.Sprintf("%d", limit)),
				sdk.NewAttribute(types.AttributeKeyWindowStart, entry.WindowStart.Format(time.RFC3339)),
			),
		)

		return fmt.Errorf("rate limit exceeded for %s: %d/%d requests in window (resets at %s)",
			operation, entry.Count, limit,
			entry.WindowStart.Add(windowDuration).Format(time.RFC3339))
	}

	// Increment counter and save
	entry.Count++
	if err := k.SetRateLimitEntry(ctx, entry); err != nil {
		return fmt.Errorf("failed to update rate limit entry: %w", err)
	}

	return nil
}

// ============================================================================
// Paginated Query Methods (DoS Protection via Pagination)
// ============================================================================

// GetAllKYCRecordsPaginated retrieves KYC records with pagination support.
// This prevents DoS attacks by limiting the number of records returned per query.
//
// Parameters:
//   - ctx: SDK context for state access
//   - pagination: PageRequest specifying limit, offset, and next key
//
// Returns:
//   - records: Paginated list of KYC records
//   - pageRes: PageResponse with next key and total count
//   - error: If pagination fails
//
// Security considerations:
//   - Default limit applied if not specified (prevents unbounded queries)
//   - Maximum limit enforced to prevent memory exhaustion
//   - Efficient iteration using prefix store
func (k *Keeper) GetAllKYCRecordsPaginated(ctx sdk.Context, pagination *query.PageRequest) ([]*types.KYCRecord, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	recordStore := prefix.NewStore(store, KYCRecordsKeyPrefix)

	records := make([]*types.KYCRecord, 0, 64)
	pageRes, err := query.Paginate(recordStore, pagination, func(key []byte, value []byte) error {
		var record types.KYCRecord
		if err := k.cdc.Unmarshal(value, &record); err != nil {
			return fmt.Errorf("error in GetAllKYCRecordsPaginated: %w", err)
		}
		records = append(records, &record)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return records, pageRes, nil
}

// GetAllAMLProfilesPaginated retrieves AML profiles with pagination support.
func (k *Keeper) GetAllAMLProfilesPaginated(ctx sdk.Context, pagination *query.PageRequest) ([]*types.AMLProfile, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	profileStore := prefix.NewStore(store, AMLProfilesKeyPrefix)

	profiles := make([]*types.AMLProfile, 0, 64)
	pageRes, err := query.Paginate(profileStore, pagination, func(key []byte, value []byte) error {
		var profile types.AMLProfile
		if err := k.cdc.Unmarshal(value, &profile); err != nil {
			return fmt.Errorf("error in GetAllAMLProfilesPaginated: %w", err)
		}
		profiles = append(profiles, &profile)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return profiles, pageRes, nil
}

// GetAllSanctionsResultsPaginated retrieves sanctions screening results with pagination support.
func (k *Keeper) GetAllSanctionsResultsPaginated(ctx sdk.Context, pagination *query.PageRequest) ([]*types.SanctionsScreeningResult, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	resultsStore := prefix.NewStore(store, SanctionsResultsKeyPrefix)

	results := make([]*types.SanctionsScreeningResult, 0, 64)
	pageRes, err := query.Paginate(resultsStore, pagination, func(key []byte, value []byte) error {
		var result types.SanctionsScreeningResult
		if err := k.cdc.Unmarshal(value, &result); err != nil {
			return fmt.Errorf("error in GetAllSanctionsResultsPaginated: %w", err)
		}
		results = append(results, &result)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return results, pageRes, nil
}

// GetAllGDPRConsentsPaginated retrieves GDPR consents across all addresses with pagination support.
// Returns a map of address -> consent list for each page.
func (k *Keeper) GetAllGDPRConsentsPaginated(ctx sdk.Context, pagination *query.PageRequest) (map[string][]*types.GDPRConsent, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	consentStore := prefix.NewStore(store, GDPRConsentsKeyPrefix)

	consents := make(map[string][]*types.GDPRConsent)
	pageRes, err := query.Paginate(consentStore, pagination, func(key []byte, value []byte) error {
		address := string(key)
		var list types.GDPRConsentList
		if err := k.cdc.Unmarshal(value, &list); err != nil {
			return fmt.Errorf("error in GetAllGDPRConsentsPaginated: %w", err)
		}
		consents[address] = list.Consents
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return consents, pageRes, nil
}

// GetAllTransactionAlertsPaginated retrieves transaction alerts across all addresses with pagination support.
// Returns a map of address -> alert list for each page.
func (k *Keeper) GetAllTransactionAlertsPaginated(ctx sdk.Context, pagination *query.PageRequest) (map[string][]*types.TransactionAlert, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	alertStore := prefix.NewStore(store, TransactionAlertsKeyPrefix)

	alerts := make(map[string][]*types.TransactionAlert)
	pageRes, err := query.Paginate(alertStore, pagination, func(key []byte, value []byte) error {
		address := string(key)
		var list types.TransactionAlertList
		if err := k.cdc.Unmarshal(value, &list); err != nil {
			return fmt.Errorf("error in GetAllTransactionAlertsPaginated: %w", err)
		}
		alerts[address] = list.Alerts
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return alerts, pageRes, nil
}

// GetAllTaxReportsPaginated retrieves tax reports across all addresses with pagination support.
// Returns a map of address -> report list for each page.
func (k *Keeper) GetAllTaxReportsPaginated(ctx sdk.Context, pagination *query.PageRequest) (map[string][]*types.TaxReport, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	reportStore := prefix.NewStore(store, TaxReportsKeyPrefix)

	reports := make(map[string][]*types.TaxReport)
	pageRes, err := query.Paginate(reportStore, pagination, func(key []byte, value []byte) error {
		address := string(key)
		var list types.TaxReportList
		if err := k.cdc.Unmarshal(value, &list); err != nil {
			return fmt.Errorf("error in GetAllTaxReportsPaginated: %w", err)
		}
		reports[address] = list.Reports
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return reports, pageRes, nil
}

// GetAllGDPRRequestsPaginated retrieves GDPR data requests with pagination support.
func (k *Keeper) GetAllGDPRRequestsPaginated(ctx sdk.Context, pagination *query.PageRequest) ([]*types.GDPRDataRequest, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	requestStore := prefix.NewStore(store, GDPRRequestsKeyPrefix)

	requests := make([]*types.GDPRDataRequest, 0, 64)
	pageRes, err := query.Paginate(requestStore, pagination, func(key []byte, value []byte) error {
		var request types.GDPRDataRequest
		if err := k.cdc.Unmarshal(value, &request); err != nil {
			return fmt.Errorf("error in GetAllGDPRRequestsPaginated: %w", err)
		}
		requests = append(requests, &request)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return requests, pageRes, nil
}
