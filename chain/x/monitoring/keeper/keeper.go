package keeper

import (
	"context"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/metrics"
	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// Key prefixes for KVStore
var (
	TransactionsKeyPrefix      = []byte{0x01}
	AlertsKeyPrefix            = []byte{0x02}
	AnomaliesKeyPrefix         = []byte{0x03}
	ValidatorUptimeKeyPrefix   = []byte{0x04}
	NetworkHealthKeyPrefix     = []byte{0x05}
	GasPriceTrackingKeyPrefix  = []byte{0x06}
	TVLMonitoringKeyPrefix     = []byte{0x07}
	FailedTxPatternsKeyPrefix  = []byte{0x08}
	SecurityEventsKeyPrefix    = []byte{0x09}
	LogsKeyPrefix              = []byte{0x0A}
	ParamsKeyPrefix            = []byte{0x0B}
)

// Keeper handles all monitoring operations
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
	metrics  *metrics.MonitoringMetrics

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
}

// NewKeeper creates a new monitoring keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *Keeper {
	ctx, cancel := context.WithCancel(context.Background())

	k := &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		metrics:  metrics.NewMonitoringMetrics(),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start background monitoring workers
	k.startBackgroundWorkers()

	return k
}

// Close stops all background workers
func (k *Keeper) Close() {
	k.cancel()
}

// startBackgroundWorkers starts all monitoring background workers
func (k *Keeper) startBackgroundWorkers() {
	// Background workers would be started here
	// For now, this is a placeholder for the actual implementation
}

// ============================================================================
// Params KVStore Methods
// ============================================================================

// GetParams returns the current parameters
func (k *Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ParamsKeyPrefix)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams updates the parameters
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(ParamsKeyPrefix, bz)
	return nil
}

// GetMetrics returns the Prometheus metrics
func (k *Keeper) GetMetrics() *metrics.MonitoringMetrics {
	return k.metrics
}

// ============================================================================
// Transaction Monitor KVStore Methods
// ============================================================================

// SetTransactionMonitorData stores transaction monitoring data
func (k *Keeper) SetTransactionMonitorData(ctx sdk.Context, data *types.TransactionMonitorData) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(data)
	if err != nil {
		return err
	}

	key := append(TransactionsKeyPrefix, []byte(data.TxHash)...)
	store.Set(key, bz)
	return nil
}

// GetTransactionMonitorData retrieves transaction monitoring data
func (k *Keeper) GetTransactionMonitorData(ctx sdk.Context, txHash string) (*types.TransactionMonitorData, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(TransactionsKeyPrefix, []byte(txHash)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("transaction monitor data not found: %s", txHash)
	}

	var data types.TransactionMonitorData
	if err := k.cdc.Unmarshal(bz, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetAllTransactionMonitorData retrieves all transaction monitoring data
func (k *Keeper) GetAllTransactionMonitorData(ctx sdk.Context) ([]*types.TransactionMonitorData, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, TransactionsKeyPrefix)
	defer iterator.Close()

	var transactions []*types.TransactionMonitorData
	for ; iterator.Valid(); iterator.Next() {
		var data types.TransactionMonitorData
		if err := k.cdc.Unmarshal(iterator.Value(), &data); err != nil {
			return nil, err
		}
		transactions = append(transactions, &data)
	}
	return transactions, nil
}

// DeleteTransactionMonitorData removes transaction monitoring data
func (k *Keeper) DeleteTransactionMonitorData(ctx sdk.Context, txHash string) {
	store := ctx.KVStore(k.storeKey)
	key := append(TransactionsKeyPrefix, []byte(txHash)...)
	store.Delete(key)
}

// ============================================================================
// Alert KVStore Methods
// ============================================================================

// SetAlert stores an alert
func (k *Keeper) SetAlert(ctx sdk.Context, alert *types.Alert) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(alert)
	if err != nil {
		return err
	}

	key := append(AlertsKeyPrefix, []byte(alert.ID)...)
	store.Set(key, bz)
	return nil
}

// GetAlert retrieves an alert
func (k *Keeper) GetAlert(ctx sdk.Context, id string) (*types.Alert, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(AlertsKeyPrefix, []byte(id)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("alert not found: %s", id)
	}

	var alert types.Alert
	if err := k.cdc.Unmarshal(bz, &alert); err != nil {
		return nil, err
	}
	return &alert, nil
}

// GetAllAlerts retrieves all alerts
func (k *Keeper) GetAllAlerts(ctx sdk.Context) ([]*types.Alert, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AlertsKeyPrefix)
	defer iterator.Close()

	var alerts []*types.Alert
	for ; iterator.Valid(); iterator.Next() {
		var alert types.Alert
		if err := k.cdc.Unmarshal(iterator.Value(), &alert); err != nil {
			return nil, err
		}
		alerts = append(alerts, &alert)
	}
	return alerts, nil
}

// DeleteAlert removes an alert
func (k *Keeper) DeleteAlert(ctx sdk.Context, id string) {
	store := ctx.KVStore(k.storeKey)
	key := append(AlertsKeyPrefix, []byte(id)...)
	store.Delete(key)
}

// ============================================================================
// Anomaly Detection KVStore Methods
// ============================================================================

// SetAnomalyDetection stores anomaly detection data
func (k *Keeper) SetAnomalyDetection(ctx sdk.Context, anomaly *types.AnomalyDetection) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(anomaly)
	if err != nil {
		return err
	}

	key := append(AnomaliesKeyPrefix, []byte(anomaly.ID)...)
	store.Set(key, bz)
	return nil
}

// GetAnomalyDetection retrieves anomaly detection data
func (k *Keeper) GetAnomalyDetection(ctx sdk.Context, id string) (*types.AnomalyDetection, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(AnomaliesKeyPrefix, []byte(id)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("anomaly detection not found: %s", id)
	}

	var anomaly types.AnomalyDetection
	if err := k.cdc.Unmarshal(bz, &anomaly); err != nil {
		return nil, err
	}
	return &anomaly, nil
}

// GetAllAnomalyDetections retrieves all anomaly detections
func (k *Keeper) GetAllAnomalyDetections(ctx sdk.Context) ([]*types.AnomalyDetection, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AnomaliesKeyPrefix)
	defer iterator.Close()

	var anomalies []*types.AnomalyDetection
	for ; iterator.Valid(); iterator.Next() {
		var anomaly types.AnomalyDetection
		if err := k.cdc.Unmarshal(iterator.Value(), &anomaly); err != nil {
			return nil, err
		}
		anomalies = append(anomalies, &anomaly)
	}
	return anomalies, nil
}

// DeleteAnomalyDetection removes anomaly detection data
func (k *Keeper) DeleteAnomalyDetection(ctx sdk.Context, id string) {
	store := ctx.KVStore(k.storeKey)
	key := append(AnomaliesKeyPrefix, []byte(id)...)
	store.Delete(key)
}

// ============================================================================
// Validator Uptime KVStore Methods
// ============================================================================

// SetValidatorUptime stores validator uptime data
func (k *Keeper) SetValidatorUptime(ctx sdk.Context, uptime *types.ValidatorUptime) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(uptime)
	if err != nil {
		return err
	}

	key := append(ValidatorUptimeKeyPrefix, []byte(uptime.ValidatorAddress)...)
	store.Set(key, bz)
	return nil
}

// GetValidatorUptime retrieves validator uptime data
func (k *Keeper) GetValidatorUptime(ctx sdk.Context, validatorAddress string) (*types.ValidatorUptime, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(ValidatorUptimeKeyPrefix, []byte(validatorAddress)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("validator uptime not found: %s", validatorAddress)
	}

	var uptime types.ValidatorUptime
	if err := k.cdc.Unmarshal(bz, &uptime); err != nil {
		return nil, err
	}
	return &uptime, nil
}

// GetAllValidatorUptimes retrieves all validator uptime data
func (k *Keeper) GetAllValidatorUptimes(ctx sdk.Context) ([]*types.ValidatorUptime, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, ValidatorUptimeKeyPrefix)
	defer iterator.Close()

	var uptimes []*types.ValidatorUptime
	for ; iterator.Valid(); iterator.Next() {
		var uptime types.ValidatorUptime
		if err := k.cdc.Unmarshal(iterator.Value(), &uptime); err != nil {
			return nil, err
		}
		uptimes = append(uptimes, &uptime)
	}
	return uptimes, nil
}

// DeleteValidatorUptime removes validator uptime data
func (k *Keeper) DeleteValidatorUptime(ctx sdk.Context, validatorAddress string) {
	store := ctx.KVStore(k.storeKey)
	key := append(ValidatorUptimeKeyPrefix, []byte(validatorAddress)...)
	store.Delete(key)
}

// ============================================================================
// Network Health KVStore Methods
// ============================================================================

// SetNetworkHealth stores network health metrics
func (k *Keeper) SetNetworkHealth(ctx sdk.Context, health *types.NetworkHealth) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(health)
	if err != nil {
		return err
	}

	store.Set(NetworkHealthKeyPrefix, bz)
	return nil
}

// GetNetworkHealth retrieves network health metrics
func (k *Keeper) GetNetworkHealth(ctx sdk.Context) (*types.NetworkHealth, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(NetworkHealthKeyPrefix)
	if bz == nil {
		return &types.NetworkHealth{}, nil
	}

	var health types.NetworkHealth
	if err := k.cdc.Unmarshal(bz, &health); err != nil {
		return nil, err
	}
	return &health, nil
}

// ============================================================================
// Gas Price Tracking KVStore Methods
// ============================================================================

// SetGasPriceTracking stores gas price tracking data
func (k *Keeper) SetGasPriceTracking(ctx sdk.Context, tracking *types.GasPriceTracking) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(tracking)
	if err != nil {
		return err
	}

	store.Set(GasPriceTrackingKeyPrefix, bz)
	return nil
}

// GetGasPriceTracking retrieves gas price tracking data
func (k *Keeper) GetGasPriceTracking(ctx sdk.Context) (*types.GasPriceTracking, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(GasPriceTrackingKeyPrefix)
	if bz == nil {
		return &types.GasPriceTracking{PriceHistory: []types.GasPricePoint{}}, nil
	}

	var tracking types.GasPriceTracking
	if err := k.cdc.Unmarshal(bz, &tracking); err != nil {
		return nil, err
	}
	return &tracking, nil
}

// ============================================================================
// TVL Monitoring KVStore Methods
// ============================================================================

// SetTVLMonitoring stores TVL monitoring data
func (k *Keeper) SetTVLMonitoring(ctx sdk.Context, tvl *types.TVLMonitoring) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(tvl)
	if err != nil {
		return err
	}

	store.Set(TVLMonitoringKeyPrefix, bz)
	return nil
}

// GetTVLMonitoring retrieves TVL monitoring data
func (k *Keeper) GetTVLMonitoring(ctx sdk.Context) (*types.TVLMonitoring, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(TVLMonitoringKeyPrefix)
	if bz == nil {
		return &types.TVLMonitoring{
			TVLByModule: make(map[string]uint64),
			TVLHistory:  []types.TVLPoint{},
		}, nil
	}

	var tvl types.TVLMonitoring
	if err := k.cdc.Unmarshal(bz, &tvl); err != nil {
		return nil, err
	}
	return &tvl, nil
}

// ============================================================================
// Failed Transaction Pattern KVStore Methods
// ============================================================================

// SetFailedTxPattern stores failed transaction pattern data
func (k *Keeper) SetFailedTxPattern(ctx sdk.Context, pattern *types.FailedTransactionPattern) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(pattern)
	if err != nil {
		return err
	}

	key := append(FailedTxPatternsKeyPrefix, []byte(pattern.ID)...)
	store.Set(key, bz)
	return nil
}

// GetFailedTxPattern retrieves failed transaction pattern data
func (k *Keeper) GetFailedTxPattern(ctx sdk.Context, id string) (*types.FailedTransactionPattern, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(FailedTxPatternsKeyPrefix, []byte(id)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("failed tx pattern not found: %s", id)
	}

	var pattern types.FailedTransactionPattern
	if err := k.cdc.Unmarshal(bz, &pattern); err != nil {
		return nil, err
	}
	return &pattern, nil
}

// GetAllFailedTxPatterns retrieves all failed transaction patterns
func (k *Keeper) GetAllFailedTxPatterns(ctx sdk.Context) ([]*types.FailedTransactionPattern, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, FailedTxPatternsKeyPrefix)
	defer iterator.Close()

	var patterns []*types.FailedTransactionPattern
	for ; iterator.Valid(); iterator.Next() {
		var pattern types.FailedTransactionPattern
		if err := k.cdc.Unmarshal(iterator.Value(), &pattern); err != nil {
			return nil, err
		}
		patterns = append(patterns, &pattern)
	}
	return patterns, nil
}

// DeleteFailedTxPattern removes failed transaction pattern data
func (k *Keeper) DeleteFailedTxPattern(ctx sdk.Context, id string) {
	store := ctx.KVStore(k.storeKey)
	key := append(FailedTxPatternsKeyPrefix, []byte(id)...)
	store.Delete(key)
}

// ============================================================================
// Security Event KVStore Methods
// ============================================================================

// SetSecurityEvent stores security event data
func (k *Keeper) SetSecurityEvent(ctx sdk.Context, event *types.SecurityEvent) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(event)
	if err != nil {
		return err
	}

	key := append(SecurityEventsKeyPrefix, []byte(event.ID)...)
	store.Set(key, bz)
	return nil
}

// GetSecurityEvent retrieves security event data
func (k *Keeper) GetSecurityEvent(ctx sdk.Context, id string) (*types.SecurityEvent, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(SecurityEventsKeyPrefix, []byte(id)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("security event not found: %s", id)
	}

	var event types.SecurityEvent
	if err := k.cdc.Unmarshal(bz, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// GetAllSecurityEvents retrieves all security events
func (k *Keeper) GetAllSecurityEvents(ctx sdk.Context) ([]*types.SecurityEvent, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, SecurityEventsKeyPrefix)
	defer iterator.Close()

	var events []*types.SecurityEvent
	for ; iterator.Valid(); iterator.Next() {
		var event types.SecurityEvent
		if err := k.cdc.Unmarshal(iterator.Value(), &event); err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, nil
}

// DeleteSecurityEvent removes security event data
func (k *Keeper) DeleteSecurityEvent(ctx sdk.Context, id string) {
	store := ctx.KVStore(k.storeKey)
	key := append(SecurityEventsKeyPrefix, []byte(id)...)
	store.Delete(key)
}

// ============================================================================
// Log Entry KVStore Methods
// ============================================================================

// SetLogEntry stores a log entry
func (k *Keeper) SetLogEntry(ctx sdk.Context, module string, log *types.LogEntry) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(log)
	if err != nil {
		return err
	}

	// Key: prefix + module + log ID
	key := append(LogsKeyPrefix, []byte(module)...)
	key = append(key, []byte(log.ID)...)
	store.Set(key, bz)
	return nil
}

// GetLogEntry retrieves a log entry
func (k *Keeper) GetLogEntry(ctx sdk.Context, module, id string) (*types.LogEntry, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(LogsKeyPrefix, []byte(module)...)
	key = append(key, []byte(id)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("log entry not found: %s/%s", module, id)
	}

	var log types.LogEntry
	if err := k.cdc.Unmarshal(bz, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// GetLogEntriesForModule retrieves all log entries for a module
func (k *Keeper) GetLogEntriesForModule(ctx sdk.Context, module string) ([]*types.LogEntry, error) {
	store := ctx.KVStore(k.storeKey)
	prefix := append(LogsKeyPrefix, []byte(module)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var logs []*types.LogEntry
	for ; iterator.Valid(); iterator.Next() {
		var log types.LogEntry
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

// GetAllLogEntries retrieves all log entries
func (k *Keeper) GetAllLogEntries(ctx sdk.Context) ([]*types.LogEntry, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, LogsKeyPrefix)
	defer iterator.Close()

	var logs []*types.LogEntry
	for ; iterator.Valid(); iterator.Next() {
		var log types.LogEntry
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

// DeleteLogEntry removes a log entry
func (k *Keeper) DeleteLogEntry(ctx sdk.Context, module, id string) {
	store := ctx.KVStore(k.storeKey)
	key := append(LogsKeyPrefix, []byte(module)...)
	key = append(key, []byte(id)...)
	store.Delete(key)
}

// ============================================================================
// Cleanup Helper Methods
// ============================================================================

// cleanupExpiredData removes old monitoring data based on retention policies
func (k *Keeper) CleanupExpiredData(ctx sdk.Context) {
	params := k.GetParams(ctx)
	now := time.Now()

	// Clean up old alerts
	if params.EnableAlerts {
		alerts, _ := k.GetAllAlerts(ctx)
		for _, alert := range alerts {
			if alert.Resolved && alert.ResolvedAt != nil {
				if now.Sub(*alert.ResolvedAt) > params.AlertRetentionPeriod {
					k.DeleteAlert(ctx, alert.ID)
				}
			}
		}
	}

	// Clean up old security events
	if params.EnableSIEM {
		events, _ := k.GetAllSecurityEvents(ctx)
		for _, event := range events {
			if now.Sub(event.Timestamp) > params.SIEMRetentionPeriod {
				k.DeleteSecurityEvent(ctx, event.ID)
			}
		}
	}

	// Clean up old log entries
	if params.EnableLogAggregation {
		logs, _ := k.GetAllLogEntries(ctx)
		for _, log := range logs {
			if now.Sub(log.Timestamp) > params.LogRetentionPeriod {
				k.DeleteLogEntry(ctx, log.Module, log.ID)
			}
		}
	}

	// Clean up old transactions (keep last 24 hours)
	transactions, _ := k.GetAllTransactionMonitorData(ctx)
	for _, tx := range transactions {
		if now.Sub(tx.Timestamp) > 24*time.Hour {
			k.DeleteTransactionMonitorData(ctx, tx.TxHash)
		}
	}

	// Clean up gas price history
	gasPriceTracking, _ := k.GetGasPriceTracking(ctx)
	if gasPriceTracking != nil && len(gasPriceTracking.PriceHistory) > params.GasPriceHistorySize {
		gasPriceTracking.PriceHistory = gasPriceTracking.PriceHistory[len(gasPriceTracking.PriceHistory)-params.GasPriceHistorySize:]
		k.SetGasPriceTracking(ctx, gasPriceTracking)
	}

	// Clean up TVL history
	tvlMonitoring, _ := k.GetTVLMonitoring(ctx)
	if tvlMonitoring != nil && len(tvlMonitoring.TVLHistory) > params.TVLHistorySize {
		tvlMonitoring.TVLHistory = tvlMonitoring.TVLHistory[len(tvlMonitoring.TVLHistory)-params.TVLHistorySize:]
		k.SetTVLMonitoring(ctx, tvlMonitoring)
	}
}

// generateID generates a unique ID for monitoring data
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// ============================================================================
// Background Worker Methods (placeholders)
// ============================================================================

// networkHealthWorker monitors network health
func (k *Keeper) networkHealthWorker() {
	// Placeholder for network health monitoring worker
}

// gasPriceWorker monitors gas prices
func (k *Keeper) gasPriceWorker() {
	// Placeholder for gas price monitoring worker
}

// tvlMonitoringWorker monitors TVL
func (k *Keeper) tvlMonitoringWorker() {
	// Placeholder for TVL monitoring worker
}

// validatorMonitoringWorker monitors validators
func (k *Keeper) validatorMonitoringWorker() {
	// Placeholder for validator monitoring worker
}

// failedTxAnalysisWorker analyzes failed transactions
func (k *Keeper) failedTxAnalysisWorker() {
	// Placeholder for failed transaction analysis worker
}

// explorerSyncWorker syncs with blockchain explorer
func (k *Keeper) explorerSyncWorker() {
	// Placeholder for explorer sync worker
}
