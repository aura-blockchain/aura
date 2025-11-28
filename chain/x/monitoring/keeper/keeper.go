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
)

// Keeper handles all monitoring operations with persistent KV store.
// All state is stored in the KV store for consensus safety.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	authority    string

	// Metrics (non-consensus, observability only)
	metrics *metrics.MonitoringMetrics
}

// NewKeeper creates a new monitoring keeper with KV store persistence
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		metrics:      metrics.NewMonitoringMetrics(),
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
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
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
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := types.ValidateParams(params); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(&params)
	if err != nil {
		return err
	}

	return store.Set(ParamsKey, bz)
}

// ============================================================================
// Alert CRUD Operations
// ============================================================================

// GetAlert retrieves an alert from the KV store
func (k Keeper) GetAlert(ctx context.Context, alertID string) (*types.Alert, error) {
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
func (k Keeper) SetAlert(ctx context.Context, alert *types.Alert) error {
	if alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(AlertKeyPrefix, []byte(alert.ID)...)

	bz, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	return store.Set(key, bz)
}

// DeleteAlert removes an alert from the KV store
func (k Keeper) DeleteAlert(ctx context.Context, alertID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(AlertKeyPrefix, []byte(alertID)...)
	return store.Delete(key)
}

// IterateAlerts iterates over all alerts in the KV store
func (k Keeper) IterateAlerts(ctx context.Context, fn func(alert *types.Alert) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(AlertKeyPrefix, storetypes.PrefixEndBytes(AlertKeyPrefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var alert types.Alert
		if err := json.Unmarshal(iterator.Value(), &alert); err != nil {
			return err
		}
		if fn(&alert) {
			break
		}
	}

	return nil
}

// GetAllAlerts retrieves all alerts from the KV store
func (k Keeper) GetAllAlerts(ctx context.Context) ([]*types.Alert, error) {
	var alerts []*types.Alert
	err := k.IterateAlerts(ctx, func(alert *types.Alert) bool {
		alerts = append(alerts, alert)
		return false
	})
	return alerts, err
}

// GetActiveAlerts returns all unresolved alerts
func (k Keeper) GetActiveAlerts(ctx context.Context) ([]*types.Alert, error) {
	var activeAlerts []*types.Alert
	err := k.IterateAlerts(ctx, func(alert *types.Alert) bool {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
		return false
	})
	return activeAlerts, err
}

// GetAlertsBySeverity returns alerts filtered by severity
func (k Keeper) GetAlertsBySeverity(ctx context.Context, severity types.AlertSeverity) ([]*types.Alert, error) {
	var filtered []*types.Alert
	err := k.IterateAlerts(ctx, func(alert *types.Alert) bool {
		if alert.Severity == severity {
			filtered = append(filtered, alert)
		}
		return false
	})
	return filtered, err
}

// GetAlertsByType returns alerts filtered by type
func (k Keeper) GetAlertsByType(ctx context.Context, alertType types.AlertType) ([]*types.Alert, error) {
	var filtered []*types.Alert
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
func (k Keeper) SetTransaction(ctx context.Context, tx *types.TransactionMonitorData) error {
	if tx == nil {
		return types.ErrInvalidTransaction
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(TransactionKeyPrefix, []byte(tx.TxHash)...)

	bz, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	return store.Set(key, bz)
}

// DeleteTransaction removes a transaction from the KV store
func (k Keeper) DeleteTransaction(ctx context.Context, txHash string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(TransactionKeyPrefix, []byte(txHash)...)
	return store.Delete(key)
}

// IterateTransactions iterates over all transactions in the KV store
func (k Keeper) IterateTransactions(ctx context.Context, fn func(tx *types.TransactionMonitorData) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(TransactionKeyPrefix, storetypes.PrefixEndBytes(TransactionKeyPrefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tx types.TransactionMonitorData
		if err := json.Unmarshal(iterator.Value(), &tx); err != nil {
			return err
		}
		if fn(&tx) {
			break
		}
	}

	return nil
}

// GetAllTransactions retrieves all transactions from the KV store
func (k Keeper) GetAllTransactions(ctx context.Context) ([]*types.TransactionMonitorData, error) {
	var transactions []*types.TransactionMonitorData
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
func (k Keeper) GetAnomaly(ctx context.Context, anomalyID string) (*types.AnomalyDetection, error) {
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
func (k Keeper) SetAnomaly(ctx context.Context, anomaly *types.AnomalyDetection) error {
	if anomaly == nil {
		return fmt.Errorf("anomaly cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(AnomalyKeyPrefix, []byte(anomaly.ID)...)

	bz, err := json.Marshal(anomaly)
	if err != nil {
		return err
	}

	return store.Set(key, bz)
}

// DeleteAnomaly removes an anomaly from the KV store
func (k Keeper) DeleteAnomaly(ctx context.Context, anomalyID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(AnomalyKeyPrefix, []byte(anomalyID)...)
	return store.Delete(key)
}

// IterateAnomalies iterates over all anomalies in the KV store
func (k Keeper) IterateAnomalies(ctx context.Context, fn func(anomaly *types.AnomalyDetection) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(AnomalyKeyPrefix, storetypes.PrefixEndBytes(AnomalyKeyPrefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var anomaly types.AnomalyDetection
		if err := json.Unmarshal(iterator.Value(), &anomaly); err != nil {
			return err
		}
		if fn(&anomaly) {
			break
		}
	}

	return nil
}

// GetAllAnomalies retrieves all anomalies from the KV store
func (k Keeper) GetAllAnomalies(ctx context.Context) ([]*types.AnomalyDetection, error) {
	var anomalies []*types.AnomalyDetection
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
func (k Keeper) SetValidatorUptime(ctx context.Context, uptime *types.ValidatorUptime) error {
	if uptime == nil {
		return fmt.Errorf("validator uptime cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(ValidatorUptimeKeyPrefix, []byte(uptime.ValidatorAddress)...)

	bz, err := json.Marshal(uptime)
	if err != nil {
		return err
	}

	return store.Set(key, bz)
}

// DeleteValidatorUptime removes validator uptime from the KV store
func (k Keeper) DeleteValidatorUptime(ctx context.Context, validatorAddr string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(ValidatorUptimeKeyPrefix, []byte(validatorAddr)...)
	return store.Delete(key)
}

// IterateValidatorUptimes iterates over all validator uptimes in the KV store
func (k Keeper) IterateValidatorUptimes(ctx context.Context, fn func(uptime *types.ValidatorUptime) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(ValidatorUptimeKeyPrefix, storetypes.PrefixEndBytes(ValidatorUptimeKeyPrefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var uptime types.ValidatorUptime
		if err := json.Unmarshal(iterator.Value(), &uptime); err != nil {
			return err
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
func (k Keeper) SetNetworkHealth(ctx context.Context, health *types.NetworkHealth) error {
	if health == nil {
		return fmt.Errorf("network health cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)

	bz, err := json.Marshal(health)
	if err != nil {
		return err
	}

	return store.Set(NetworkHealthKey, bz)
}

// ============================================================================
// Gas Price Tracking Operations (Single Entry)
// ============================================================================

// NOTE: GetGasPriceTracking is defined in gas_price_tracker.go

// SetGasPriceTracking stores gas price tracking in the KV store
func (k Keeper) SetGasPriceTracking(ctx context.Context, tracking *types.GasPriceTracking) error {
	if tracking == nil {
		return fmt.Errorf("gas price tracking cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)

	bz, err := json.Marshal(tracking)
	if err != nil {
		return err
	}

	return store.Set(GasPriceTrackingKey, bz)
}

// ============================================================================
// TVL Monitoring Operations (Single Entry)
// ============================================================================

// NOTE: GetTVLMonitoring is defined in tvl_monitor.go

// SetTVLMonitoring stores TVL monitoring in the KV store
func (k Keeper) SetTVLMonitoring(ctx context.Context, monitoring *types.TVLMonitoring) error {
	if monitoring == nil {
		return fmt.Errorf("TVL monitoring cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)

	bz, err := json.Marshal(monitoring)
	if err != nil {
		return err
	}

	return store.Set(TVLMonitoringKey, bz)
}

// ============================================================================
// Failed Transaction Pattern CRUD Operations
// ============================================================================

// GetFailedTxPattern retrieves a failed transaction pattern from the KV store
func (k Keeper) GetFailedTxPattern(ctx context.Context, patternID string) (*types.FailedTransactionPattern, error) {
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
func (k Keeper) SetFailedTxPattern(ctx context.Context, pattern *types.FailedTransactionPattern) error {
	if pattern == nil {
		return fmt.Errorf("failed transaction pattern cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(FailedTxPatternKeyPrefix, []byte(pattern.ID)...)

	bz, err := json.Marshal(pattern)
	if err != nil {
		return err
	}

	return store.Set(key, bz)
}

// DeleteFailedTxPattern removes a failed transaction pattern from the KV store
func (k Keeper) DeleteFailedTxPattern(ctx context.Context, patternID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(FailedTxPatternKeyPrefix, []byte(patternID)...)
	return store.Delete(key)
}

// IterateFailedTxPatterns iterates over all failed transaction patterns in the KV store
func (k Keeper) IterateFailedTxPatterns(ctx context.Context, fn func(pattern *types.FailedTransactionPattern) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(FailedTxPatternKeyPrefix, storetypes.PrefixEndBytes(FailedTxPatternKeyPrefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var pattern types.FailedTransactionPattern
		if err := json.Unmarshal(iterator.Value(), &pattern); err != nil {
			return err
		}
		if fn(&pattern) {
			break
		}
	}

	return nil
}

// GetAllFailedTxPatterns retrieves all failed transaction patterns from the KV store
func (k Keeper) GetAllFailedTxPatterns(ctx context.Context) ([]*types.FailedTransactionPattern, error) {
	var patterns []*types.FailedTransactionPattern
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
func (k Keeper) GetSecurityEvent(ctx context.Context, eventID string) (*types.SecurityEvent, error) {
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
func (k Keeper) SetSecurityEvent(ctx context.Context, event *types.SecurityEvent) error {
	if event == nil {
		return types.ErrSecurityEventInvalid
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(SecurityEventKeyPrefix, []byte(event.ID)...)

	bz, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return store.Set(key, bz)
}

// DeleteSecurityEvent removes a security event from the KV store
func (k Keeper) DeleteSecurityEvent(ctx context.Context, eventID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(SecurityEventKeyPrefix, []byte(eventID)...)
	return store.Delete(key)
}

// IterateSecurityEvents iterates over all security events in the KV store
func (k Keeper) IterateSecurityEvents(ctx context.Context, fn func(event *types.SecurityEvent) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(SecurityEventKeyPrefix, storetypes.PrefixEndBytes(SecurityEventKeyPrefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var event types.SecurityEvent
		if err := json.Unmarshal(iterator.Value(), &event); err != nil {
			return err
		}
		if fn(&event) {
			break
		}
	}

	return nil
}

// GetAllSecurityEvents retrieves all security events from the KV store
func (k Keeper) GetAllSecurityEvents(ctx context.Context) ([]*types.SecurityEvent, error) {
	var events []*types.SecurityEvent
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
func (k Keeper) GetLogEntry(ctx context.Context, logID string) (*types.LogEntry, error) {
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
func (k Keeper) SetLogEntry(ctx context.Context, entry *types.LogEntry) error {
	if entry == nil {
		return types.ErrLogAggregationFailed
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(LogEntryKeyPrefix, []byte(entry.ID)...)

	bz, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return store.Set(key, bz)
}

// DeleteLogEntry removes a log entry from the KV store
func (k Keeper) DeleteLogEntry(ctx context.Context, logID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(LogEntryKeyPrefix, []byte(logID)...)
	return store.Delete(key)
}

// IterateLogEntries iterates over all log entries in the KV store
func (k Keeper) IterateLogEntries(ctx context.Context, fn func(entry *types.LogEntry) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(LogEntryKeyPrefix, storetypes.PrefixEndBytes(LogEntryKeyPrefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var entry types.LogEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			return err
		}
		if fn(&entry) {
			break
		}
	}

	return nil
}

// GetAllLogEntries retrieves all log entries from the KV store
func (k Keeper) GetAllLogEntries(ctx context.Context) ([]*types.LogEntry, error) {
	var entries []*types.LogEntry
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
func (k Keeper) SetExplorerIntegration(ctx context.Context, integration *types.ExplorerIntegration) error {
	if integration == nil {
		return fmt.Errorf("explorer integration cannot be nil")
	}

	store := k.storeService.OpenKVStore(ctx)

	bz, err := json.Marshal(integration)
	if err != nil {
		return err
	}

	return store.Set(ExplorerIntegrationKey, bz)
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateID generates a unique ID with a prefix using block time for consensus safety
func (k Keeper) generateID(ctx context.Context, prefix string) string {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return fmt.Sprintf("%s_%d", prefix, sdkCtx.BlockTime().UnixNano())
}
