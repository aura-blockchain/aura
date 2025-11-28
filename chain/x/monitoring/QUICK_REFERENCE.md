# Monitoring Module KV Store - Quick Reference

## Keeper Structure (NEW)

```go
type Keeper struct {
    cdc          codec.BinaryCodec
    storeService store.KVStoreService
    authority    string
    metrics      *metrics.MonitoringMetrics  // Non-consensus only
}
```

## Constructor

```go
keeper.NewKeeper(
    cdc codec.BinaryCodec,
    storeService store.KVStoreService,
    authority string,
)
```

## KV Store Key Prefixes

| Prefix | Type | Description |
|--------|------|-------------|
| `0x01` | Multi | Alerts |
| `0x02` | Multi | Transactions |
| `0x03` | Multi | Anomalies |
| `0x04` | Multi | Validator Uptimes |
| `0x05` | Single | Network Health |
| `0x06` | Single | Gas Price Tracking |
| `0x07` | Single | TVL Monitoring |
| `0x08` | Multi | Failed Tx Patterns |
| `0x09` | Multi | Security Events |
| `0x0A` | Multi | Log Entries |
| `0x0B` | Single | Params |
| `0x0C` | Single | Explorer Integration |

## Common Operations

### Alert Management

```go
// Create/Update alert
alert := &types.Alert{
    ID:       k.generateID(ctx, "alert"),
    Type:     types.AlertTypeLargeTransaction,
    Severity: types.SeverityHigh,
    Message:  "Alert message",
    // ... other fields
}
err := k.SetAlert(ctx, alert)

// Get alert
alert, err := k.GetAlert(ctx, alertID)

// Get all active alerts
activeAlerts, err := k.GetActiveAlerts(ctx)

// Get alerts by severity
criticalAlerts, err := k.GetAlertsBySeverity(ctx, types.SeverityCritical)

// Delete alert
err := k.DeleteAlert(ctx, alertID)

// Iterate all alerts
err := k.IterateAlerts(ctx, func(alert *types.Alert) bool {
    // Process alert
    return false // continue iteration
})
```

### Transaction Monitoring

```go
// Store transaction
tx := &types.TransactionMonitorData{
    TxHash:      "0x123...",
    Sender:      "cosmos1...",
    Amount:      1000000,
    BlockHeight: sdkCtx.BlockHeight(),
    // ... other fields
}
err := k.SetTransaction(ctx, tx)

// Get transaction
tx, err := k.GetTransaction(ctx, txHash)

// Get all transactions
txs, err := k.GetAllTransactions(ctx)

// Iterate with filtering
err := k.IterateTransactions(ctx, func(tx *types.TransactionMonitorData) bool {
    if tx.Status == "failed" {
        // Process failed tx
    }
    return false
})
```

### Anomaly Detection

```go
// Store anomaly
anomaly := &types.AnomalyDetection{
    ID:        k.generateID(ctx, "anomaly"),
    Type:      types.AnomalyTypeTransaction,
    Score:     0.95,
    Threshold: 0.80,
    IsAnomaly: true,
    // ... other fields
}
err := k.SetAnomaly(ctx, anomaly)

// Get anomaly
anomaly, err := k.GetAnomaly(ctx, anomalyID)

// Get all anomalies
anomalies, err := k.GetAllAnomalies(ctx)
```

### Validator Uptime

```go
// Update validator uptime
uptime := &types.ValidatorUptime{
    ValidatorAddress: "cosmosvaloper1...",
    TotalBlocks:      10000,
    SignedBlocks:     9950,
    MissedBlocks:     50,
    UptimePercentage: 99.5,
    // ... other fields
}
err := k.SetValidatorUptime(ctx, uptime)

// Get validator uptime
uptime, err := k.GetValidatorUptime(ctx, validatorAddr)

// Get all uptimes
uptimes, err := k.GetAllValidatorUptimes(ctx)
```

### Network Health (Single Entry)

```go
// Update network health
health := &types.NetworkHealth{
    Timestamp:        sdkCtx.BlockTime(),
    BlockHeight:      sdkCtx.BlockHeight(),
    BlockTime:        6.5,
    TPS:              100.0,
    ActiveValidators: 125,
    NetworkCongestion: 0.3,
    // ... other fields
}
err := k.SetNetworkHealth(ctx, health)

// Get current network health
health, err := k.GetNetworkHealth(ctx)
```

### Gas Price Tracking (Single Entry)

```go
// Update gas price tracking
tracking, err := k.GetGasPriceTracking(ctx)
if err != nil {
    return err
}

tracking.CurrentPrice = 1000
tracking.AveragePrice = 950
tracking.PriceHistory = append(tracking.PriceHistory, types.GasPricePoint{
    Timestamp: sdkCtx.BlockTime(),
    Price:     1000,
})

err = k.SetGasPriceTracking(ctx, tracking)
```

### TVL Monitoring (Single Entry)

```go
// Update TVL
monitoring, err := k.GetTVLMonitoring(ctx)
if err != nil {
    return err
}

monitoring.TotalTVL = 50000000
monitoring.TVLByModule["dex"] = 30000000
monitoring.TVLByModule["lending"] = 20000000

err = k.SetTVLMonitoring(ctx, monitoring)
```

### Failed Transaction Patterns

```go
// Store pattern
pattern := &types.FailedTransactionPattern{
    ID:              k.generateID(ctx, "pattern"),
    Pattern:         "insufficient_funds",
    FailureReason:   "Insufficient balance",
    Occurrences:     42,
    Severity:        types.SeverityMedium,
    // ... other fields
}
err := k.SetFailedTxPattern(ctx, pattern)

// Get all patterns
patterns, err := k.GetAllFailedTxPatterns(ctx)
```

### Security Events

```go
// Store security event
event := &types.SecurityEvent{
    ID:          k.generateID(ctx, "sec-event"),
    EventType:   types.SecurityEventSuspiciousTransaction,
    Severity:    types.SeverityHigh,
    Description: "Unusual transaction pattern detected",
    ThreatLevel: 7,
    // ... other fields
}
err := k.SetSecurityEvent(ctx, event)

// Get all events
events, err := k.GetAllSecurityEvents(ctx)
```

### Log Entries

```go
// Store log entry
entry := &types.LogEntry{
    ID:        k.generateID(ctx, "log"),
    Level:     types.LogLevelError,
    Module:    "monitoring",
    Message:   "Error processing transaction",
    Timestamp: sdkCtx.BlockTime(),
    // ... other fields
}
err := k.SetLogEntry(ctx, entry)

// Iterate logs with filtering
err := k.IterateLogEntries(ctx, func(entry *types.LogEntry) bool {
    if entry.Level == types.LogLevelError {
        // Process error log
    }
    return false
})
```

### Parameters

```go
// Get params
params, err := k.GetParams(ctx)

// Update params
params.EnableAlerts = true
params.LargeTransactionThreshold = 1000000
err = k.SetParams(ctx, params)
```

## ID Generation (Consensus-Safe)

```go
// Generate consensus-safe ID using block time
id := k.generateID(ctx, "alert")
// Returns: "alert_1234567890123456789"
```

## Migration Checklist

- [x] Remove all in-memory maps
- [x] Remove sync.RWMutex
- [x] Remove sync.WaitGroup
- [x] Remove background workers
- [x] Implement KV store CRUD operations
- [x] Use consensus-safe ID generation
- [ ] Update other keeper files to use new methods
- [ ] Update module constructor
- [ ] Update genesis import/export
- [ ] Add comprehensive tests
- [ ] Remove legacy code

## Verification

```bash
# No in-memory state
grep -r "sync.RWMutex\|sync.Mutex\|map\[string\]" keeper/keeper.go

# Proper KV store usage
grep -c "storeService.OpenKVStore" keeper/keeper.go  # Should be 38+

# No background workers
grep -r "go k\.\|sync.WaitGroup" keeper/keeper.go
```

## Key Benefits

1. **Consensus Safe**: All nodes have identical state
2. **Persistent**: State survives restarts
3. **Deterministic**: Uses block time, not wall-clock
4. **Scalable**: No memory bloat
5. **Standard**: Follows Cosmos SDK best practices
