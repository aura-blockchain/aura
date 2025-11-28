# Pre-Validation Module

Energy-efficient transaction optimization through off-peak pre-validation.

## Quick Start

### Overview

The Pre-Validation module optimizes transaction processing by:
1. Pre-validating common transactions during off-peak hours (2am-6am)
2. Storing them encrypted as smart contracts
3. Executing instantly when needed
4. Auto-scaling based on usage patterns

**Benefits:**
- ⚡ 50-500ms time savings per transaction
- 🌱 ~0.9 Wh energy savings per transaction
- 📈 Auto-scales to match demand
- 🔒 AES-256-GCM encryption
- 📊 Real-time metrics & monitoring

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Pre-Validation Module                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Scheduler   │  │ Auto-Scaling │  │   Metrics    │      │
│  │              │  │              │  │              │      │
│  │ Off-Peak Run │  │ Hit Rate     │  │ Cache Hits   │      │
│  │ 2am-6am      │  │ Monitoring   │  │ Time Savings │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                               │
│  ┌──────────────────────────────────────────────────┐       │
│  │           Pre-Validated Transaction Cache         │       │
│  │                                                    │       │
│  │  Strategy: FIFO | LRU | LFU | Adaptive           │       │
│  │  Max Size: 10,000 (configurable)                 │       │
│  │  Expiry: 72 hours (configurable)                 │       │
│  └──────────────────────────────────────────────────┘       │
│                                                               │
│  ┌──────────────────────────────────────────────────┐       │
│  │              Template System                      │       │
│  │                                                    │       │
│  │  IR Completion • DEX Swap • LP Ops • VC Mint     │       │
│  │  Bridge • Score Update • Identity Change          │       │
│  └──────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. Keeper (`keeper/keeper.go`)
Core state management and transaction handling.

**Key Methods:**
- `CreatePreValidatedTransaction()`: Create new pre-validation
- `GetPreValidatedTransaction()`: Retrieve by ID
- `FindPreValidatedTransaction()`: Search by criteria
- `ExecutePreValidatedTransaction()`: Execute pre-validated tx
- `RegisterTemplate()`: Add validation template

### 2. Scheduler (`keeper/scheduler.go`)
Manages off-peak pre-validation runs.

**Key Methods:**
- `ShouldRunScheduler()`: Check if should run now
- `RunScheduler()`: Execute pre-validation run
- `ForceSchedulerRun()`: Admin override

### 3. Auto-Scaling (`keeper/autoscaling.go`)
Adjusts pre-validation amounts based on metrics.

**Key Methods:**
- `RunAutoScaling()`: Execute scaling algorithm
- `GetAutoScalingRecommendations()`: Get recommendations
- `AdjustTypeAmount()`: Manual adjustment
- `CalculateOptimalAmounts()`: Calculate optimal config

### 4. Metrics (`keeper/metrics.go`)
Tracks performance and generates insights.

**Key Methods:**
- `UpdateMetrics()`: Refresh all metrics
- `RecordCacheHit()`: Record successful cache use
- `RecordCacheMiss()`: Record cache miss
- `GetMetricsSummary()`: Get summary
- `ExportMetrics()`: Prometheus format

## Transaction Types

| Type | Initial | Max | Priority | Min Score | Gas |
|------|---------|-----|----------|-----------|-----|
| IR Completion | 100 | 1,000 | 100 | 100 | 50k |
| DEX Swap | 50 | 500 | 80 | 50 | 100k |
| LP Deposit | 30 | 300 | 60 | 50 | 80k |
| LP Withdrawal | 30 | 300 | 60 | 50 | 80k |
| VC Mint | 20 | 200 | 70 | 500 | 120k |
| Bridge Transfer | 15 | 150 | 50 | 200 | 150k |
| Score Update | 25 | 250 | 90 | 0 | 60k |
| Identity Change | 10 | 100 | 40 | 1,000 | 90k |

## Configuration

### Default Parameters

```go
Enabled: true
MaxCacheSize: 10,000
ExpiryHours: 72
EncryptionAlgorithm: "AES-256-GCM"
ControlGroupPercentage: 5.0
MinConfidenceScore: 100
CacheStrategy: Adaptive

SchedulerConfig:
  OffPeakHours: [2, 3, 4, 5, 6]  // 2am-6am UTC
  RunIntervalMinutes: 30
  MaxPerRun: 1,000

AutoScalingConfig:
  TargetCacheHitRate: 0.80  // 80%
  MinCacheHitRate: 0.50     // 50%
  ScaleUpFactor: 1.5
  ScaleDownFactor: 0.75
  CooldownMinutes: 60
```

### Customization

**Change Off-Peak Hours:**
```go
params.SchedulerConfig.OffPeakHours = []uint32{0, 1, 2, 3, 4, 5, 6, 7}
params.SchedulerConfig.Timezone = "America/New_York"
```

**Adjust Cache Size:**
```go
params.MaxCacheSize = 50000  // Increase for high volume
```

**Tune Auto-Scaling:**
```go
params.AutoScalingConfig.TargetCacheHitRate = 0.85  // Higher target
params.AutoScalingConfig.ScaleUpFactor = 2.0        // More aggressive
```

## Usage Examples

### Creating a Pre-Validated Transaction

```go
keeper := NewKeeper(paramsStore)

tx, err := keeper.CreatePreValidatedTransaction(
    types.TxTypeIRCompletion,
    "ir-completion-basic",
    []byte("transaction-data"),
    "aura1...",  // signer
    50000,       // estimated gas
    map[string]string{"ir_id": "routine-123"},
)
```

### Finding a Matching Pre-Validation

```go
tx, found := keeper.FindPreValidatedTransaction(
    types.TxTypeIRCompletion,
    "aura1...",  // signer
    map[string]string{"ir_id": "routine-123"},
)

if found {
    // Use pre-validated transaction
    data, err := keeper.ExecutePreValidatedTransaction(tx.Id)
}
```

### Registering a Custom Template

```go
template := &types.ValidationTemplate{
    Id:              "usdc-aura-high-volume",
    TxType:          types.TxTypeDexSwap,
    Name:            "USDC-AURA High Volume",
    Description:     "Optimized for high-volume USDC-AURA swaps",
    ValidationRules: `{"min_amount": 1000, "max_slippage": 0.01}`,
    ParameterSchema: `{"from": "USDC", "to": "AURA", "amount": "uint64"}`,
    GasFormula:      "95000",
    PriorityWeight:  150,  // Higher priority
    Active:          true,
}

keeper.RegisterTemplate(template)
```

### Getting Metrics

```go
// Summary
summary := keeper.GetMetricsSummary()
fmt.Printf("Cache Hit Rate: %.2f%%\n", summary["cache_hit_rate"].(float64) * 100)
fmt.Printf("Time Saved: %dms\n", summary["total_time_saved_ms"].(uint64))

// Detailed metrics
metrics := keeper.GetMetrics()
for txType, typeMetrics := range metrics.MetricsByType {
    fmt.Printf("%s: %.2f%% hit rate\n", txType, typeMetrics.CacheHitRate * 100)
}

// Cache statistics
stats := keeper.GetCacheStatistics()
fmt.Printf("Cache Utilization: %.2f%%\n", stats["cache_utilization"].(float64) * 100)
```

### Running Auto-Scaling

```go
// Automatic (runs periodically)
events, err := keeper.RunAutoScaling()
for _, event := range events {
    fmt.Printf("Scaled %s from %d to %d (hit rate: %.2f%%)\n",
        event.TxType, event.PreviousAmount, event.NewAmount,
        event.CacheHitRate * 100)
}

// Get recommendations only
recommendations := keeper.GetAutoScalingRecommendations()
for _, rec := range recommendations {
    if rec.Decision != "no_change" {
        fmt.Printf("%s: %s (%s)\n", rec.TxType, rec.Decision, rec.Reason)
    }
}
```

## Events

The module emits the following events:

### EventPreValidationCreated
Emitted when a new pre-validation is created.
```go
{
  TxId: "pvtx:abc123...",
  TxType: TxTypeIRCompletion,
  TemplateId: "ir-completion-basic",
  Signer: "aura1...",
  BlockHeight: 12345,
  EstimatedGas: 50000
}
```

### EventPreValidationExecuted
Emitted when a pre-validated transaction is executed.
```go
{
  TxId: "pvtx:abc123...",
  TxType: TxTypeIRCompletion,
  Signer: "aura1...",
  TimeSavedMs: 75,
  FromCache: true
}
```

### EventSchedulerRun
Emitted when the scheduler completes a run.
```go
{
  StartedAt: timestamp,
  CompletedAt: timestamp,
  PreValidationsCreated: 280,
  CreatedByType: {
    "IR_COMPLETION": 100,
    "DEX_SWAP": 50,
    ...
  },
  IsOffPeak: true
}
```

### EventAutoScaling
Emitted when auto-scaling adjusts amounts.
```go
{
  TxType: TxTypeIRCompletion,
  PreviousAmount: 100,
  NewAmount: 150,
  CacheHitRate: 0.87,
  Reason: "cache hit rate above target - increasing amount"
}
```

## Testing

### Unit Tests

```bash
cd chain/x/prevalidation/keeper
go test -v
```

**Test Coverage:**
- Keeper initialization
- Transaction lifecycle
- Encryption/decryption
- Template management
- Cache eviction strategies
- Expiration handling
- Metrics recording
- Confidence score validation

### Integration Testing

```go
// Mock confidence score keeper
mockCS := NewMockConfidenceScoreKeeper()
mockCS.SetUserScore("aura1test", 500)

keeper := setupTestKeeper()
keeper.SetConfidenceScoreKeeper(mockCS)

// Test full flow
tx, _ := keeper.CreatePreValidatedTransaction(...)
tx.Status = ValidationStatusValidated
data, _ := keeper.ExecutePreValidatedTransaction(tx.Id)
```

## Monitoring

### Prometheus Metrics

Export metrics for monitoring:
```go
metrics := keeper.ExportMetrics()
// Returns map[string]float64 with keys like:
// - prevalidation.total.created
// - prevalidation.total.executed
// - prevalidation.cache.hit_rate
// - prevalidation.time_savings.avg_ms
// - prevalidation.energy.saved_kwh
// - prevalidation.IR_COMPLETION.cache.hit_rate
```

### Recommended Dashboards

**Cache Performance:**
- Hit rate over time (line chart)
- Hit/miss ratio by type (stacked bar)
- Cache utilization gauge

**Scaling Behavior:**
- Type amounts over time (multi-line)
- Scaling events (timeline)
- Hit rate vs target (dual-axis)

**Efficiency:**
- Cumulative time saved (area chart)
- Energy savings (line chart)
- Cost savings (calculated)

**Control Group:**
- Execution time distribution (histogram)
- Pre-validated vs normal comparison (box plot)
- Percentiles over time (line chart)

## Troubleshooting

### Low Cache Hit Rate

**Symptoms:** Hit rate < 60%

**Possible Causes:**
- Transaction patterns don't match templates
- Amounts too low for demand
- Templates too specific

**Solutions:**
1. Review template parameter schemas
2. Add more general templates
3. Increase amounts for affected types
4. Check auto-scaling is enabled

### High Expiration Rate

**Symptoms:** >30% of pre-validations expire unused

**Possible Causes:**
- Amounts too high for demand
- Expiry time too long
- Transaction patterns changed

**Solutions:**
1. Scale down amounts
2. Reduce expiry hours
3. Update templates to match patterns
4. Enable auto-scaling

### Scheduler Not Running

**Symptoms:** No scheduler events

**Possible Causes:**
- Disabled in config
- Not in off-peak hours
- Cooldown period active

**Solutions:**
1. Check `params.Enabled` and `params.SchedulerConfig.Enabled`
2. Verify current hour in `OffPeakHours`
3. Use `ForceSchedulerRun()` to test
4. Check logs for errors

### Memory Issues

**Symptoms:** High memory usage, OOM errors

**Possible Causes:**
- Cache size too large
- Not enough cleanup
- Memory leak

**Solutions:**
1. Reduce `MaxCacheSize`
2. Reduce `ExpiryHours`
3. Enable more aggressive eviction
4. Run `CleanupExpiredTransactions()` more frequently

## Best Practices

### Production Deployment

1. **Start Conservative**
   - Use default amounts initially
   - Monitor for 48 hours before scaling up
   - Keep control group at 5-10%

2. **Monitor Closely**
   - Set up alerts for hit rate < 50%
   - Watch expiration rates
   - Track energy savings

3. **Tune Gradually**
   - Adjust one parameter at a time
   - Wait for evaluation period before next change
   - Document all changes

4. **Security**
   - Rotate encryption keys periodically
   - Audit template changes
   - Monitor for unusual patterns

### Development

1. **Enable Detailed Logging**
   ```go
   params.DetailedLogging = true
   ```

2. **Use Shorter Intervals**
   ```go
   params.SchedulerConfig.RunIntervalMinutes = 5
   params.AutoScalingConfig.CooldownMinutes = 10
   ```

3. **Allow Peak Hours**
   ```go
   params.SchedulerConfig.AllowPeakHours = true
   ```

4. **Larger Control Group**
   ```go
   params.ControlGroupPercentage = 20.0
   ```

## API Reference

See [PREVALIDATION.md](../../../docs/modules/PREVALIDATION.md) for detailed API documentation.

## Contributing

When adding new features:

1. Update proto definitions first
2. Regenerate Go code: `buf generate`
3. Implement keeper methods
4. Add unit tests
5. Update documentation
6. Add integration tests

## License

Same as parent Aura project.

## Support

- Documentation: `/docs/modules/PREVALIDATION.md`
- Implementation Guide: `/PREVALIDATION_IMPLEMENTATION.md`
- Issues: GitHub issues
- Discussions: GitHub discussions
