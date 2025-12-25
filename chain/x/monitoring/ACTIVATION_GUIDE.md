# Monitoring Module: ML/SIEM Activation Guide

**Status:** Components defined but dormant (not instantiated)
**Binary Impact:** ~100KB compiled code, 0 bytes runtime allocation
**Total Binary Size:** 172M (monitoring is <0.1% of total)

---

## Current Architecture

```go
// keeper/keeper.go - Current State
type Keeper struct {
    cdc          codec.BinaryCodec
    storeService store.KVStoreService
    authority    string

    // ACTIVE: Prometheus metrics for observability
    metrics *metrics.MonitoringMetrics

    // DORMANT: ML/SIEM components not instantiated
    // anomalyDetector *ml.AnomalyDetector       // Commented out
    // siemManager     *siem.SIEMManager          // Commented out
    // alertManager    *alerting.AlertManager     // Commented out
}
```

**What exists:**
- ✅ Storage schema (KV store keys for alerts, anomalies, security events)
- ✅ CRUD methods (Get/Set/Delete for all data types)
- ✅ Prometheus metrics (actively used)
- ✅ Query/msg server handlers (CLI commands work)

**What's dormant:**
- ○ ML anomaly detector engine
- ○ SIEM event correlation engine
- ○ Alert management engine

---

## Staged Activation Plan

### Phase 1: Testnet (Current)
**Goal:** Validate architecture, collect baseline metrics

**Active:**
- Prometheus metrics
- Data persistence (alerts, anomalies, events stored in KV store)
- Query endpoints

**Dormant:**
- ML training/detection
- SIEM correlation
- Alert routing

**Validation:**
```bash
# Check metrics endpoint
curl http://localhost:10030/metrics | grep monitoring_

# Query stored data
aurad query monitoring alerts
aurad query monitoring network-health
```

---

### Phase 2: Mainnet Launch (Alert Manager)
**Goal:** Enable operational alerting

**Changes Required:**

1. **Update keeper.go:**
```go
type Keeper struct {
    // ... existing fields ...

    // Add alert manager
    alertManager *alerting.AlertManager
}

func NewKeeper(...) *Keeper {
    k := &Keeper{
        // ... existing init ...
        alertManager: alerting.NewAlertManager(5 * time.Minute), // 5min cooldown
    }
    return k
}
```

2. **Wire to alert creation:**
```go
// keeper/alerts.go - Update CreateAlert method
func (k Keeper) CreateAlert(ctx context.Context, ...) (*types.Alert, error) {
    // Create via AlertManager (adds cooldown logic)
    alert, err := k.alertManager.CreateAlert(alertType, severity, message, details)
    if err != nil {
        return nil, err
    }

    // Persist to KV store (consensus-safe)
    if err := k.SetAlert(ctx, alert); err != nil {
        return nil, err
    }

    return alert, nil
}
```

3. **Add governance parameter:**
```go
// types/params.go
type Params struct {
    // ... existing params ...
    EnableAlerting    bool  `json:"enable_alerting"`
    AlertCooldownSecs int64 `json:"alert_cooldown_secs"`
}
```

**Testing:**
```bash
# Test alert creation
aurad tx monitoring create-alert \
  --type validator_down \
  --severity critical \
  --message "Validator X offline" \
  --from validator

# Verify cooldown works (second alert should fail within 5 min)
aurad tx monitoring create-alert \
  --type validator_down \
  --severity critical \
  --message "Validator X still offline" \
  --from validator
```

---

### Phase 3: Post-Launch (SIEM)
**Goal:** Security event correlation and threat analysis

**Changes Required:**

1. **Update keeper.go:**
```go
type Keeper struct {
    // ... existing fields ...
    siemManager *siem.SIEMManager
}

func NewKeeper(...) *Keeper {
    params, _ := k.GetParams(ctx)

    k := &Keeper{
        // ... existing init ...
        siemManager: siem.NewSIEMManager(params.SiemThreatThreshold),
    }
    return k
}
```

2. **Wire to security event recording:**
```go
// keeper/siem_integration.go (new file)
func (k Keeper) RecordSecurityEvent(ctx context.Context, ...) error {
    // Record in SIEM manager (in-memory correlation)
    event, err := k.siemManager.RecordSecurityEvent(
        ctx, eventType, severity, source, destination,
        description, rawData, indicators, threatLevel,
    )
    if err != nil {
        return err
    }

    // Persist to KV store (consensus-safe)
    return k.SetSecurityEvent(ctx, event)
}

// Query correlated events
func (k Keeper) GetHighThreatEvents(ctx context.Context) ([]*types.SecurityEvent, error) {
    return k.siemManager.GetHighThreatEvents(), nil
}
```

3. **Add governance parameters:**
```go
type Params struct {
    // ... existing params ...
    EnableSIEM           bool  `json:"enable_siem"`
    SiemThreatThreshold  int   `json:"siem_threat_threshold"` // 1-10
    SiemRetentionDays    int64 `json:"siem_retention_days"`
}
```

**Testing:**
```bash
# Record suspicious transaction
aurad tx monitoring record-security-event \
  --event-type suspicious_transaction \
  --severity high \
  --threat-level 8 \
  --source aura1abc... \
  --from validator

# Query high-threat events
aurad query monitoring security-events --threat-level 7
```

---

### Phase 4: Advanced (ML Anomaly Detection)
**Goal:** Automated threat detection

**Changes Required:**

1. **Update keeper.go:**
```go
type Keeper struct {
    // ... existing fields ...
    anomalyDetector *ml.AnomalyDetector
}

func NewKeeper(...) *Keeper {
    params, _ := k.GetParams(ctx)

    k := &Keeper{
        // ... existing init ...
        anomalyDetector: ml.NewAnomalyDetector(
            params.AnomalyThreshold,      // e.g., 0.75
            time.Duration(params.TrainingIntervalSecs) * time.Second,
        ),
    }
    return k
}
```

2. **Wire to transaction monitoring:**
```go
// keeper/transaction_monitor.go - Update MonitorTransaction
func (k Keeper) MonitorTransaction(ctx context.Context, tx *types.TransactionMonitorData) error {
    // Store transaction
    if err := k.SetTransaction(ctx, tx); err != nil {
        return err
    }

    // Run anomaly detection (if enabled)
    params, _ := k.GetParams(ctx)
    if params.EnableAnomalyDetection {
        detection, err := k.anomalyDetector.DetectTransactionAnomaly(ctx, tx)
        if err != nil {
            // Log but don't fail (non-consensus operation)
            k.Logger(ctx).Error("Anomaly detection failed", "error", err)
        } else if detection.IsAnomaly {
            // Store anomaly
            if err := k.SetAnomaly(ctx, detection); err != nil {
                k.Logger(ctx).Error("Failed to store anomaly", "error", err)
            }

            // Create alert for high-score anomalies
            if detection.Score > 0.9 {
                k.CreateAlert(ctx, types.AlertTypeAnomalousTransaction,
                    types.SeverityHigh, "High-score transaction anomaly detected", ...)
            }
        }
    }

    return nil
}
```

3. **Add EndBlock training trigger:**
```go
// keeper/keeper.go or module.go
func (k Keeper) EndBlock(ctx context.Context) error {
    params, _ := k.GetParams(ctx)

    if params.EnableAnomalyDetection {
        // Periodic training happens automatically in detector
        // (background goroutine runs every TrainingInterval)
    }

    return nil
}
```

4. **Add governance parameters:**
```go
type Params struct {
    // ... existing params ...
    EnableAnomalyDetection bool    `json:"enable_anomaly_detection"`
    AnomalyThreshold      float64 `json:"anomaly_threshold"`      // 0.0-1.0
    TrainingIntervalSecs  int64   `json:"training_interval_secs"` // seconds
}

func DefaultParams() Params {
    return Params{
        EnableAnomalyDetection: false,  // Start disabled
        AnomalyThreshold:      0.75,    // 75% threshold
        TrainingIntervalSecs:  86400,   // 24 hours
        // ... other defaults ...
    }
}
```

**Testing:**
```bash
# Enable via governance proposal
aurad tx gov submit-proposal param-change proposal.json

# proposal.json:
{
  "title": "Enable ML Anomaly Detection",
  "description": "Activate machine learning based anomaly detection",
  "changes": [{
    "subspace": "monitoring",
    "key": "EnableAnomalyDetection",
    "value": true
  }]
}

# Monitor anomaly detections
aurad query monitoring anomalies --limit 10
```

---

## Configuration Parameters

### Default Values (Production)

```go
type Params struct {
    // Alert Manager
    EnableAlerting       bool   `json:"enable_alerting"`        // true for mainnet
    AlertCooldownSecs    int64  `json:"alert_cooldown_secs"`    // 300 (5 minutes)

    // SIEM
    EnableSIEM           bool   `json:"enable_siem"`            // true for mainnet
    SiemThreatThreshold  int    `json:"siem_threat_threshold"`  // 7 (high threat)
    SiemRetentionDays    int64  `json:"siem_retention_days"`    // 90 days

    // ML Anomaly Detection
    EnableAnomalyDetection bool    `json:"enable_anomaly_detection"` // false initially
    AnomalyThreshold      float64 `json:"anomaly_threshold"`         // 0.75
    TrainingIntervalSecs  int64   `json:"training_interval_secs"`    // 86400 (24h)
}

func DefaultParams() Params {
    return Params{
        EnableAlerting:         true,   // Alerts on by default
        AlertCooldownSecs:      300,    // 5 minutes
        EnableSIEM:             true,   // SIEM on by default
        SiemThreatThreshold:    7,      // High threat level
        SiemRetentionDays:      90,     // 3 months
        EnableAnomalyDetection: false,  // ML off until proven stable
        AnomalyThreshold:       0.75,   // 75% threshold
        TrainingIntervalSecs:   86400,  // 24 hours
    }
}
```

---

## Performance Considerations

### Memory Usage (When Activated)

| Component | Memory Footprint | Notes |
|-----------|------------------|-------|
| AlertManager | ~100-500 KB | ~1000 alerts × ~500 bytes each |
| SIEMManager | ~500 KB - 1 MB | Event storage, configurable retention |
| AnomalyDetector | ~1-2 MB | 10k training samples × ~200 bytes |
| **Total** | **~2-4 MB** | Negligible for blockchain node |

### CPU Impact (When Activated)

| Operation | CPU Time | Frequency |
|-----------|----------|-----------|
| Anomaly detection | ~1ms | Per transaction |
| SIEM event recording | ~0.1ms | Per security event |
| Alert creation | ~0.1ms | Per alert |
| ML model training | ~100ms | Every 24 hours |

**Impact:** Negligible - all operations are non-consensus

---

## Monitoring Activation Status

### Metrics to Track

```promql
# Alert Manager status
monitoring_alerts_total{type="*"}
monitoring_alerts_active{severity="critical"}

# SIEM status
monitoring_security_events_total{event_type="*"}
monitoring_threat_level{level="high"}

# ML Anomaly Detection status
monitoring_anomaly_detections_total{type="*"}
monitoring_ml_model_accuracy
```

### Dashboard Updates

After activation, update Grafana dashboards:
1. Enable hidden panels for ML/SIEM metrics
2. Add alerting rules for high-threat events
3. Configure notification channels (Slack, PagerDuty, etc.)

---

## Rollback Procedure

If issues arise, disable via governance:

```bash
# Emergency disable (requires validator consensus)
aurad tx gov submit-proposal param-change disable-ml.json

# disable-ml.json:
{
  "title": "Disable ML Anomaly Detection",
  "description": "Emergency rollback due to performance issues",
  "changes": [{
    "subspace": "monitoring",
    "key": "EnableAnomalyDetection",
    "value": false
  }]
}
```

**Immediate effect:** Next block will skip ML processing, components remain in memory but idle.

---

## Success Criteria

**Phase 1 (Testnet - Current):**
- ✅ All tests pass
- ✅ Binary builds successfully
- ✅ Metrics endpoint functional
- ✅ Query commands work

**Phase 2 (Alert Manager):**
- Alert creation works
- Cooldown prevents spam
- CLI commands functional
- Grafana dashboards show alerts

**Phase 3 (SIEM):**
- Security events recorded
- Threat correlation works
- High-threat queries accurate
- Retention cleanup functional

**Phase 4 (ML):**
- Anomaly detection catches test cases
- Model retraining completes without errors
- Performance impact <1% (CPU, memory)
- False positive rate <5%

---

**Prepared by:** Claude Code Agent
**Last Updated:** 2025-12-25
**Status:** Ready for production activation when needed
