// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/metrics"
	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// KVStore key prefixes
var (
	AlertKeyPrefix            = []byte{0x01}
	TransactionKeyPrefix      = []byte{0x02}
	AnomalyKeyPrefix          = []byte{0x03}
	ValidatorUptimeKeyPrefix  = []byte{0x04}
	NetworkHealthKey          = []byte{0x05} // Single entry
	GasPriceTrackingKey       = []byte{0x06} // Single entry
	TVLMonitoringKey          = []byte{0x07} // Single entry
	FailedTxPatternKeyPrefix  = []byte{0x08}
	SecurityEventKeyPrefix    = []byte{0x09}
	LogEntryKeyPrefix         = []byte{0x0A}
	ParamsKey                 = []byte{0x0B} // Single entry
	ExplorerIntegrationKey    = []byte{0x0C} // Single entry
	LogCounterKey             = []byte{0x0D} // Single entry - log ID counter
)

// Keeper handles all monitoring operations with persistent KV store.
// All state is stored in the KV store for consensus safety.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	authority    string

	// Metrics (non-consensus, observability only)
	metrics *metrics.MonitoringMetrics

	// Network health collector (non-consensus, queries local RPC for metrics)
	healthCollector *NetworkHealthCollector

	// Cache for alert routes (non-consensus, performance optimization)
	// Invalidated at the start of each block or when routes are modified
	alertRoutesCache      []*AlertRoute
	alertRoutesCacheBlock int64
}

// NewKeeper creates a new monitoring keeper with KV store persistence
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeService:    storeService,
		authority:       authority,
		metrics:         metrics.NewMonitoringMetrics(),
		healthCollector: NewNetworkHealthCollector("http://localhost:26657"),
	}
}

// SetRPCEndpoint configures the RPC endpoint for network health collection.
// This should be called after keeper creation if the RPC port is non-standard.
func (k *Keeper) SetRPCEndpoint(endpoint string) {
	if k.healthCollector != nil {
		k.healthCollector.SetRPCEndpoint(endpoint)
	}
}

// GetAuthority returns the authority address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// GetMetrics returns the Prometheus metrics (non-consensus)
func (k *Keeper) GetMetrics() *metrics.MonitoringMetrics {
	return k.metrics
}

// ============================================================================
// Params Methods
// ============================================================================

// GetParams retrieves the params from the KV store
func (k *Keeper) GetParams(ctx context.Context) (types.Params, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(ParamsKey)
	if err != nil {
		return types.Params{}, err
	}
	if bz == nil {
		return types.DefaultParams(), nil
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.Params{}, err
	}

	return params, nil
}

// SetParams stores the params in the KV store
func (k *Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := types.ValidateParams(params); err != nil {
		return fmt.Errorf("error in SetParams for ValidateParams: %w", err)
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(&params)
	if err != nil {
		return fmt.Errorf("failed to marshal for ValidateParams: %w", err)
	}

	return store.Set(ParamsKey, bz)
}

// ============================================================================
// Alert CRUD Operations
// ============================================================================

// GetAlert retrieves an alert from the KV store
func (k *Keeper) GetAlert(ctx context.Context, alertID string) (*types.Alert, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(AlertKeyPrefix, []byte(alertID)...)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrAlertNotFound
	}

	var alert types.Alert
	if err := json.Unmarshal(bz, &alert); err != nil {
		return nil, err
	}

	return &alert, nil
}

// SetAlert stores an alert in the KV store
func (k *Keeper) SetAlert(ctx context.Context, alert *types.Alert) error {
	if alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(AlertKeyPrefix, []byte(alert.ID)...)

	bz, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return store.Set(key, bz)
}

// DeleteAlert removes an alert from the KV store
func (k *Keeper) DeleteAlert(ctx context.Context, alertID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(AlertKeyPrefix, []byte(alertID)...)
	return store.Delete(key)
}

// IterateAlerts iterates over all alerts in the KV store
func (k *Keeper) IterateAlerts(ctx context.Context, fn func(alert *types.Alert) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(AlertKeyPrefix, storetypes.PrefixEndBytes(AlertKeyPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var alert types.Alert
		if err := json.Unmarshal(iterator.Value(), &alert); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if fn(&alert) {
			break
		}
	}

	return nil
}

// GetAllAlerts retrieves all alerts from the KV store
func (k *Keeper) GetAllAlerts(ctx context.Context) ([]*types.Alert, error) {
	alerts := make([]*types.Alert, 0, 64)
	err := k.IterateAlerts(ctx, func(alert *types.Alert) bool {
		alerts = append(alerts, alert)
		return false
	})
	return alerts, err
}

// GetActiveAlerts returns all unresolved alerts
func (k *Keeper) GetActiveAlerts(ctx context.Context) ([]*types.Alert, error) {
	activeAlerts := make([]*types.Alert, 0, 64)
	err := k.IterateAlerts(ctx, func(alert *types.Alert) bool {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
		return false
	})
	return activeAlerts, err
}

// GetAlertsBySeverity returns alerts filtered by severity
func (k *Keeper) GetAlertsBySeverity(ctx context.Context, severity types.AlertSeverity) ([]*types.Alert, error) {
	filtered := make([]*types.Alert, 0, 64)
	err := k.IterateAlerts(ctx, func(alert *types.Alert) bool {
		if alert.Severity == severity {
			filtered = append(filtered, alert)
		}
		return false
	})
	return filtered, err
}

// GetAlertsByType returns alerts filtered by type
func (k *Keeper) GetAlertsByType(ctx context.Context, alertType types.AlertType) ([]*types.Alert, error) {
	filtered := make([]*types.Alert, 0, 64)
	err := k.IterateAlerts(ctx, func(alert *types.Alert) bool {
		if alert.Type == alertType {
			filtered = append(filtered, alert)
		}
		return false
	})
	return filtered, err
}

// ============================================================================
// Transaction Monitoring CRUD Operations
// ============================================================================

// NOTE: GetTransaction is defined in transaction_monitor.go

// SetTransaction stores a transaction in the KV store
func (k *Keeper) SetTransaction(ctx context.Context, tx *types.TransactionMonitorData) error {
	if tx == nil {
		return types.ErrInvalidTransaction
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(TransactionKeyPrefix, []byte(tx.TxHash)...)

	bz, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal for ErrInvalidTransaction: %w", err)
	}

	return store.Set(key, bz)
}

// DeleteTransaction removes a transaction from the KV store
func (k *Keeper) DeleteTransaction(ctx context.Context, txHash string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(TransactionKeyPrefix, []byte(txHash)...)
	return store.Delete(key)
}

// IterateTransactions iterates over all transactions in the KV store
func (k *Keeper) IterateTransactions(ctx context.Context, fn func(tx *types.TransactionMonitorData) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(TransactionKeyPrefix, storetypes.PrefixEndBytes(TransactionKeyPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tx types.TransactionMonitorData
		if err := json.Unmarshal(iterator.Value(), &tx); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if fn(&tx) {
			break
		}
	}

	return nil
}

// GetAllTransactions retrieves all transactions from the KV store
func (k *Keeper) GetAllTransactions(ctx context.Context) ([]*types.TransactionMonitorData, error) {
	transactions := make([]*types.TransactionMonitorData, 0, 64)
	err := k.IterateTransactions(ctx, func(tx *types.TransactionMonitorData) bool {
		transactions = append(transactions, tx)
		return false
	})
	return transactions, err
}

// ============================================================================
// Anomaly Detection CRUD Operations
// ============================================================================

// GetAnomaly retrieves an anomaly from the KV store
func (k *Keeper) GetAnomaly(ctx context.Context, anomalyID string) (*types.AnomalyDetection, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(AnomalyKeyPrefix, []byte(anomalyID)...)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrAnomalyDetectionFailed
	}

	var anomaly types.AnomalyDetection
	if err := json.Unmarshal(bz, &anomaly); err != nil {
		return nil, err
	}

	return &anomaly, nil
}

// SetAnomaly stores an anomaly in the KV store
func (k *Keeper) SetAnomaly(ctx context.Context, anomaly *types.AnomalyDetection) error {
	if anomaly == nil {
		return fmt.Errorf("anomaly cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(AnomalyKeyPrefix, []byte(anomaly.ID)...)

	bz, err := json.Marshal(anomaly)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return store.Set(key, bz)
}

// DeleteAnomaly removes an anomaly from the KV store
func (k *Keeper) DeleteAnomaly(ctx context.Context, anomalyID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(AnomalyKeyPrefix, []byte(anomalyID)...)
	return store.Delete(key)
}

// IterateAnomalies iterates over all anomalies in the KV store
func (k *Keeper) IterateAnomalies(ctx context.Context, fn func(anomaly *types.AnomalyDetection) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(AnomalyKeyPrefix, storetypes.PrefixEndBytes(AnomalyKeyPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var anomaly types.AnomalyDetection
		if err := json.Unmarshal(iterator.Value(), &anomaly); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if fn(&anomaly) {
			break
		}
	}

	return nil
}

// GetAllAnomalies retrieves all anomalies from the KV store
func (k *Keeper) GetAllAnomalies(ctx context.Context) ([]*types.AnomalyDetection, error) {
	anomalies := make([]*types.AnomalyDetection, 0, 64)
	err := k.IterateAnomalies(ctx, func(anomaly *types.AnomalyDetection) bool {
		anomalies = append(anomalies, anomaly)
		return false
	})
	return anomalies, err
}

// ============================================================================
// Validator Uptime CRUD Operations
// ============================================================================

// NOTE: GetValidatorUptime is defined in validator_monitor.go

// SetValidatorUptime stores validator uptime in the KV store
func (k *Keeper) SetValidatorUptime(ctx context.Context, uptime *types.ValidatorUptime) error {
	if uptime == nil {
		return fmt.Errorf("validator uptime cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(ValidatorUptimeKeyPrefix, []byte(uptime.ValidatorAddress)...)

	bz, err := json.Marshal(uptime)
	if err != nil {
		return fmt.Errorf("failed to marshal for SetValidatorUptime: %w", err)
	}

	return store.Set(key, bz)
}

// DeleteValidatorUptime removes validator uptime from the KV store
func (k *Keeper) DeleteValidatorUptime(ctx context.Context, validatorAddr string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(ValidatorUptimeKeyPrefix, []byte(validatorAddr)...)
	return store.Delete(key)
}

// IterateValidatorUptimes iterates over all validator uptimes in the KV store
func (k *Keeper) IterateValidatorUptimes(ctx context.Context, fn func(uptime *types.ValidatorUptime) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(ValidatorUptimeKeyPrefix, storetypes.PrefixEndBytes(ValidatorUptimeKeyPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator for ValidatorUptimeKeyPrefix: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var uptime types.ValidatorUptime
		if err := json.Unmarshal(iterator.Value(), &uptime); err != nil {
			return fmt.Errorf("failed to create iterator for ValidatorUptimeKeyPrefix: %w", err)
		}
		if fn(&uptime) {
			break
		}
	}

	return nil
}

// NOTE: GetAllValidatorUptimes is defined in validator_monitor.go

// ============================================================================
// Network Health Operations (Single Entry)
// ============================================================================

// NOTE: GetNetworkHealth is defined in network_health.go

// SetNetworkHealth stores the network health in the KV store
func (k *Keeper) SetNetworkHealth(ctx context.Context, health *types.NetworkHealth) error {
	if health == nil {
		return fmt.Errorf("network health cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)

	bz, err := json.Marshal(health)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return store.Set(NetworkHealthKey, bz)
}

// ============================================================================
// Gas Price Tracking Operations (Single Entry)
// ============================================================================

// NOTE: GetGasPriceTracking is defined in gas_price_tracker.go

// SetGasPriceTracking stores gas price tracking in the KV store
func (k *Keeper) SetGasPriceTracking(ctx context.Context, tracking *types.GasPriceTracking) error {
	if tracking == nil {
		return fmt.Errorf("gas price tracking cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)

	bz, err := json.Marshal(tracking)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return store.Set(GasPriceTrackingKey, bz)
}

// ============================================================================
// TVL Monitoring Operations (Single Entry)
// ============================================================================

// NOTE: GetTVLMonitoring is defined in tvl_monitor.go

// SetTVLMonitoring stores TVL monitoring in the KV store
func (k *Keeper) SetTVLMonitoring(ctx context.Context, monitoring *types.TVLMonitoring) error {
	if monitoring == nil {
		return fmt.Errorf("TVL monitoring cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)

	bz, err := json.Marshal(monitoring)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return store.Set(TVLMonitoringKey, bz)
}

// ============================================================================
// Failed Transaction Pattern CRUD Operations
// ============================================================================

// GetFailedTxPattern retrieves a failed transaction pattern from the KV store
func (k *Keeper) GetFailedTxPattern(ctx context.Context, patternID string) (*types.FailedTransactionPattern, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(FailedTxPatternKeyPrefix, []byte(patternID)...)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, fmt.Errorf("failed transaction pattern not found")
	}

	var pattern types.FailedTransactionPattern
	if err := json.Unmarshal(bz, &pattern); err != nil {
		return nil, err
	}

	return &pattern, nil
}

// SetFailedTxPattern stores a failed transaction pattern in the KV store
func (k *Keeper) SetFailedTxPattern(ctx context.Context, pattern *types.FailedTransactionPattern) error {
	if pattern == nil {
		return fmt.Errorf("failed transaction pattern cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(FailedTxPatternKeyPrefix, []byte(pattern.ID)...)

	bz, err := json.Marshal(pattern)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return store.Set(key, bz)
}

// DeleteFailedTxPattern removes a failed transaction pattern from the KV store
func (k *Keeper) DeleteFailedTxPattern(ctx context.Context, patternID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(FailedTxPatternKeyPrefix, []byte(patternID)...)
	return store.Delete(key)
}

// IterateFailedTxPatterns iterates over all failed transaction patterns in the KV store
func (k *Keeper) IterateFailedTxPatterns(ctx context.Context, fn func(pattern *types.FailedTransactionPattern) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(FailedTxPatternKeyPrefix, storetypes.PrefixEndBytes(FailedTxPatternKeyPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var pattern types.FailedTransactionPattern
		if err := json.Unmarshal(iterator.Value(), &pattern); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if fn(&pattern) {
			break
		}
	}

	return nil
}

// GetAllFailedTxPatterns retrieves all failed transaction patterns from the KV store
func (k *Keeper) GetAllFailedTxPatterns(ctx context.Context) ([]*types.FailedTransactionPattern, error) {
	patterns := make([]*types.FailedTransactionPattern, 0, 64)
	err := k.IterateFailedTxPatterns(ctx, func(pattern *types.FailedTransactionPattern) bool {
		patterns = append(patterns, pattern)
		return false
	})
	return patterns, err
}

// ============================================================================
// Security Event CRUD Operations
// ============================================================================

// GetSecurityEvent retrieves a security event from the KV store
func (k *Keeper) GetSecurityEvent(ctx context.Context, eventID string) (*types.SecurityEvent, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(SecurityEventKeyPrefix, []byte(eventID)...)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrSecurityEventInvalid
	}

	var event types.SecurityEvent
	if err := json.Unmarshal(bz, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// SetSecurityEvent stores a security event in the KV store
func (k *Keeper) SetSecurityEvent(ctx context.Context, event *types.SecurityEvent) error {
	if event == nil {
		return types.ErrSecurityEventInvalid
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(SecurityEventKeyPrefix, []byte(event.ID)...)

	bz, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal for ErrSecurityEventInvalid: %w", err)
	}

	return store.Set(key, bz)
}

// DeleteSecurityEvent removes a security event from the KV store
func (k *Keeper) DeleteSecurityEvent(ctx context.Context, eventID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(SecurityEventKeyPrefix, []byte(eventID)...)
	return store.Delete(key)
}

// IterateSecurityEvents iterates over all security events in the KV store
func (k *Keeper) IterateSecurityEvents(ctx context.Context, fn func(event *types.SecurityEvent) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(SecurityEventKeyPrefix, storetypes.PrefixEndBytes(SecurityEventKeyPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var event types.SecurityEvent
		if err := json.Unmarshal(iterator.Value(), &event); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if fn(&event) {
			break
		}
	}

	return nil
}

// GetAllSecurityEvents retrieves all security events from the KV store
func (k *Keeper) GetAllSecurityEvents(ctx context.Context) ([]*types.SecurityEvent, error) {
	events := make([]*types.SecurityEvent, 0, 64)
	err := k.IterateSecurityEvents(ctx, func(event *types.SecurityEvent) bool {
		events = append(events, event)
		return false
	})
	return events, err
}

// ============================================================================
// Log Entry CRUD Operations
// ============================================================================

// GetLogEntry retrieves a log entry from the KV store
func (k *Keeper) GetLogEntry(ctx context.Context, logID string) (*types.LogEntry, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(LogEntryKeyPrefix, []byte(logID)...)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrLogAggregationFailed
	}

	var entry types.LogEntry
	if err := json.Unmarshal(bz, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

// SetLogEntry stores a log entry in the KV store
func (k *Keeper) SetLogEntry(ctx context.Context, entry *types.LogEntry) error {
	if entry == nil {
		return types.ErrLogAggregationFailed
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(LogEntryKeyPrefix, []byte(entry.ID)...)

	bz, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return store.Set(key, bz)
}

// DeleteLogEntry removes a log entry from the KV store
func (k *Keeper) DeleteLogEntry(ctx context.Context, logID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(LogEntryKeyPrefix, []byte(logID)...)
	return store.Delete(key)
}

// IterateLogEntries iterates over all log entries in the KV store
func (k *Keeper) IterateLogEntries(ctx context.Context, fn func(entry *types.LogEntry) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(LogEntryKeyPrefix, storetypes.PrefixEndBytes(LogEntryKeyPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var entry types.LogEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if fn(&entry) {
			break
		}
	}

	return nil
}

// GetAllLogEntries retrieves all log entries from the KV store
func (k *Keeper) GetAllLogEntries(ctx context.Context) ([]*types.LogEntry, error) {
	entries := make([]*types.LogEntry, 0, 64)
	err := k.IterateLogEntries(ctx, func(entry *types.LogEntry) bool {
		entries = append(entries, entry)
		return false
	})
	return entries, err
}

// ============================================================================
// Explorer Integration Operations (Single Entry)
// ============================================================================

// NOTE: GetExplorerIntegration is defined in explorer_integration.go

// SetExplorerIntegration stores explorer integration in the KV store
func (k *Keeper) SetExplorerIntegration(ctx context.Context, integration *types.ExplorerIntegration) error {
	if integration == nil {
		return fmt.Errorf("explorer integration cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)

	bz, err := json.Marshal(integration)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return store.Set(ExplorerIntegrationKey, bz)
}

// ============================================================================
// Log Aggregation Methods
// ============================================================================

// LogEntry creates and stores a log entry with distributed tracing support
// Signature: LogEntry(ctx, level, module, message, fields, traceID, spanID)
func (k *Keeper) LogEntry(ctx context.Context, level types.LogLevel, module string, message string, fields map[string]interface{}, traceID string, spanID string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Generate unique log ID with timestamp, block height, and counter for uniqueness
	counter := k.getAndIncrementLogCounter(ctx)
	logID := fmt.Sprintf("log_%d_%d_%d", sdkCtx.BlockTime().UnixNano(), sdkCtx.BlockHeight(), counter)

	entry := &types.LogEntry{
		ID:        logID,
		Level:     level,
		Module:    module,
		Message:   message,
		Fields:    fields,
		Timestamp: sdkCtx.BlockTime(),
		TraceID:   traceID,
		SpanID:    spanID,
	}

	return k.SetLogEntry(ctx, entry)
}

// GetLogs retrieves log entries filtered by module with pagination limit
func (k *Keeper) GetLogs(ctx context.Context, module string, limit int) ([]*types.LogEntry, error) {
	logs := make([]*types.LogEntry, 0, 64)
	count := 0

	err := k.IterateLogEntries(ctx, func(entry *types.LogEntry) bool {
		if module == "" || entry.Module == module {
			logs = append(logs, entry)
			count++
			if limit > 0 && count >= limit {
				return true // Stop iteration
			}
		}
		return false
	})

	return logs, err
}

// GetErrorLogs retrieves only error-level log entries with pagination limit
func (k *Keeper) GetErrorLogs(ctx context.Context, limit int) ([]*types.LogEntry, error) {
	errorLogs := make([]*types.LogEntry, 0, 64)
	count := 0

	err := k.IterateLogEntries(ctx, func(entry *types.LogEntry) bool {
		if entry.Level == types.LogLevelError || entry.Level == types.LogLevelFatal {
			errorLogs = append(errorLogs, entry)
			count++
			if limit > 0 && count >= limit {
				return true // Stop iteration
			}
		}
		return false
	})

	return errorLogs, err
}

// GetLogsByTraceID retrieves all log entries for a specific distributed trace
func (k *Keeper) GetLogsByTraceID(ctx context.Context, traceID string) ([]*types.LogEntry, error) {
	tracedLogs := make([]*types.LogEntry, 0, 64)

	err := k.IterateLogEntries(ctx, func(entry *types.LogEntry) bool {
		if entry.TraceID == traceID {
			tracedLogs = append(tracedLogs, entry)
		}
		return false
	})

	return tracedLogs, err
}

// ============================================================================
// Failed Transaction Pattern Analysis Methods
// ============================================================================

// RecordFailedTransaction records a failed transaction and updates pattern tracking
// Signature: RecordFailedTransaction(ctx, tx, failureReason)
func (k *Keeper) RecordFailedTransaction(ctx context.Context, tx *types.TransactionMonitorData, failureReason string) error {
	if tx == nil {
		return types.ErrInvalidTransaction
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Store the failed transaction first
	if err := k.SetTransaction(ctx, tx); err != nil {
		return fmt.Errorf("error in RecordFailedTransaction for ErrInvalidTransaction: %w", err)
	}

	// Use failure reason as pattern ID for grouping
	patternID := failureReason

	// Try to get existing pattern
	pattern, err := k.GetFailedTxPattern(ctx, patternID)
	if err != nil {
		// Create new pattern
		pattern = &types.FailedTransactionPattern{
			ID:                patternID,
			Pattern:           failureReason, // Use failure reason as pattern identifier
			FailureReason:     failureReason,
			Occurrences:       1,
			FirstSeen:         sdkCtx.BlockTime(),
			LastSeen:          sdkCtx.BlockTime(),
			AffectedAddresses: []string{tx.Sender}, // Track affected sender addresses
			Severity:          types.SeverityMedium,
			Metadata: map[string]interface{}{
				"module":  tx.Module,
				"tx_hash": tx.TxHash,
			},
		}
	} else {
		// Update existing pattern
		pattern.Occurrences++
		pattern.LastSeen = sdkCtx.BlockTime()

		// Add sender to affected addresses if not already present
		found := false
		for _, addr := range pattern.AffectedAddresses {
			if addr == tx.Sender {
				found = true
				break
			}
		}
		if !found {
			pattern.AffectedAddresses = append(pattern.AffectedAddresses, tx.Sender)
		}

		// Limit affected addresses to last 100 to prevent unbounded growth
		if len(pattern.AffectedAddresses) > 100 {
			pattern.AffectedAddresses = pattern.AffectedAddresses[len(pattern.AffectedAddresses)-100:]
		}

		// Update severity based on occurrences
		if pattern.Occurrences > 100 {
			pattern.Severity = types.SeverityCritical
		} else if pattern.Occurrences > 50 {
			pattern.Severity = types.SeverityHigh
		}
	}

	return k.SetFailedTxPattern(ctx, pattern)
}

// GetFailedTransactionPatterns retrieves all failed transaction patterns
// Returns a slice of patterns, not a map
func (k *Keeper) GetFailedTransactionPatterns(ctx context.Context) ([]*types.FailedTransactionPattern, error) {
	return k.GetAllFailedTxPatterns(ctx)
}

// ============================================================================
// Helper Functions
// ============================================================================

// Counter key for generating unique IDs within a block
var CounterKeyPrefix = []byte{0x0F}

// getAndIncrementLogCounter gets the current log counter value and increments it atomically
func (k *Keeper) getAndIncrementLogCounter(ctx context.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)

	// Get current counter value
	bz, err := store.Get(LogCounterKey)
	var counter uint64
	if err != nil || bz == nil {
		counter = 0
	} else {
		counter = sdk.BigEndianToUint64(bz)
	}

	// Increment and store
	nextCounter := counter + 1
	if err := store.Set(LogCounterKey, sdk.Uint64ToBigEndian(nextCounter)); err != nil {
		// If increment fails, use timestamp-based fallback
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		return uint64(sdkCtx.BlockTime().UnixNano())
	}

	return counter
}

// generateID generates a unique ID with a prefix using block time and counter for consensus safety
func (k *Keeper) generateID(ctx context.Context, prefix string) string {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get and increment counter for this block to ensure uniqueness
	store := k.storeService.OpenKVStore(ctx)
	counterKey := append(CounterKeyPrefix, []byte(prefix)...)

	var counter uint64
	bz, err := store.Get(counterKey)
	if err == nil && bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}
	counter++

	// Store updated counter
	_ = store.Set(counterKey, sdk.Uint64ToBigEndian(counter))

	return fmt.Sprintf("%s_%d_%d", prefix, sdkCtx.BlockTime().UnixNano(), counter)
}

// ============================================================================
// Alert Routes Cache Management
// ============================================================================

// GetCachedAlertRoutes retrieves alert routes from cache or refreshes if stale
// This method is thread-safe for concurrent reads within the same block
func (k *Keeper) GetCachedAlertRoutes(ctx context.Context) ([]*AlertRoute, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentHeight := sdkCtx.BlockHeight()

	// Check if cache is valid for current block
	if k.alertRoutesCacheBlock == currentHeight && k.alertRoutesCache != nil {
		// Return cached routes (safe because they're read-only during the block)
		return k.alertRoutesCache, nil
	}

	// Cache is stale or empty, refresh from store
	return k.refreshAlertRoutesCache(ctx)
}

// refreshAlertRoutesCache loads all alert routes from store and caches them
func (k *Keeper) refreshAlertRoutesCache(ctx context.Context) ([]*AlertRoute, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Load all routes from store
	routes, err := k.GetAllAlertRoutes(ctx)
	if err != nil {
		return nil, err
	}

	// Update cache
	k.alertRoutesCache = routes
	k.alertRoutesCacheBlock = sdkCtx.BlockHeight()

	return routes, nil
}

// InvalidateAlertRoutesCache invalidates the alert routes cache
// Called when routes are modified (add/update/delete)
// Sets the cache block to -1 to force a refresh on next access
func (k *Keeper) InvalidateAlertRoutesCache() {
	k.alertRoutesCacheBlock = -1
}

// BeginBlocker is called at the start of each block to refresh caches and collect metrics
func (k *Keeper) BeginBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Pre-load alert routes cache for the block
	_, err := k.refreshAlertRoutesCache(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh alert routes cache: %w", err)
	}

	// Collect and update network health metrics
	// This is non-consensus (observability only) so errors are logged but don't halt the chain
	if k.healthCollector != nil {
		health, err := k.healthCollector.CollectNetworkHealth(ctx)
		if err != nil {
			sdkCtx.Logger().Debug("monitoring: failed to collect network health", "error", err)
		} else if health != nil {
			// Update the network health in store and metrics
			if updateErr := k.UpdateNetworkHealth(ctx, health); updateErr != nil {
				sdkCtx.Logger().Debug("monitoring: failed to update network health", "error", updateErr)
			}
		}
	}

	return nil
}
