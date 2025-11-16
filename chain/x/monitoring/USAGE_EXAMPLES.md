# Monitoring Module - Usage Examples

This document provides practical examples of how to use the Aura monitoring module in various scenarios.

## Table of Contents

1. [Basic Setup](#basic-setup)
2. [Transaction Monitoring](#transaction-monitoring)
3. [Alert Management](#alert-management)
4. [Validator Monitoring](#validator-monitoring)
5. [Network Health Tracking](#network-health-tracking)
6. [Gas Price Monitoring](#gas-price-monitoring)
7. [TVL Tracking](#tvl-tracking)
8. [Log Aggregation](#log-aggregation)
9. [SIEM Integration](#siem-integration)
10. [Anomaly Detection](#anomaly-detection)

## Basic Setup

### Initialize the Monitoring Keeper

```go
package main

import (
    "github.com/aequitas/aura/chain/x/monitoring/keeper"
    "github.com/aequitas/aura/chain/x/monitoring/types"
)

func main() {
    // Use default parameters
    params := types.DefaultParams()

    // Or customize parameters
    params.LargeTransactionThreshold = 5000000
    params.AnomalyThreshold = 0.8
    params.AlertRetentionPeriod = 60 * 24 * time.Hour

    // Create keeper
    k := keeper.NewKeeper(&params)
    defer k.Close()

    // Keeper is now ready to use
}
```

### Custom Configuration

```go
// Create highly customized parameters
params := types.Params{
    // Transaction Monitoring
    EnableTransactionMonitoring: true,
    LargeTransactionThreshold:   10000000, // 10M tokens

    // Alerts
    EnableAlerts:              true,
    AlertRetentionPeriod:      90 * 24 * time.Hour, // 90 days
    CriticalAlertCooldown:     10 * time.Minute,

    // Anomaly Detection
    EnableAnomalyDetection:    true,
    AnomalyThreshold:          0.85, // Higher threshold (less sensitive)
    MLModelUpdateInterval:     12 * time.Hour, // More frequent updates

    // Validator Monitoring
    EnableValidatorMonitoring: true,
    ValidatorUptimeWindow:     20000, // 20k blocks
    MaxConsecutiveMisses:      50,

    // Gas Price
    EnableGasPriceTracking:    true,
    GasPriceCheckInterval:     30 * time.Second,
    GasPriceSpikeThreshold:    3.0, // 3x average
    GasPriceHistorySize:       2880, // 48 hours at 1-min intervals
}

k := keeper.NewKeeper(&params)
```

## Transaction Monitoring

### Monitor a Single Transaction

```go
import "time"

tx := &types.TransactionMonitorData{
    TxHash:      "0x123abc...",
    Sender:      "aura1sender123...",
    Receiver:    "aura1receiver456...",
    Amount:      1500000,
    GasUsed:     75000,
    GasPrice:    150,
    Status:      "success",
    Timestamp:   time.Now(),
    BlockHeight: 12345,
    Module:      "bank",
}

err := k.MonitorTransaction(tx)
if err != nil {
    log.Printf("Failed to monitor transaction: %v", err)
}
```

### Monitor Failed Transaction

```go
failedTx := &types.TransactionMonitorData{
    TxHash:      "0x456def...",
    Sender:      "aura1sender123...",
    Receiver:    "aura1receiver456...",
    Amount:      500000,
    GasUsed:     25000,
    GasPrice:    100,
    Status:      "failed",
    Timestamp:   time.Now(),
    BlockHeight: 12346,
    Module:      "bank",
}

// Monitor the transaction
err := k.MonitorTransaction(failedTx)
if err != nil {
    log.Printf("Error: %v", err)
}

// Record the failure pattern
err = k.RecordFailedTransaction(failedTx, "insufficient_funds")
if err != nil {
    log.Printf("Error recording failure: %v", err)
}
```

### Retrieve Transaction Data

```go
// Get specific transaction
tx, err := k.GetTransaction("0x123abc...")
if err != nil {
    log.Printf("Transaction not found: %v", err)
} else {
    fmt.Printf("Amount: %d, Gas: %d\n", tx.Amount, tx.GasUsed)
}

// Get recent transactions
recentTxs := k.GetRecentTransactions(100)
fmt.Printf("Found %d recent transactions\n", len(recentTxs))

// Get large transactions
largeTxs := k.GetLargeTransactions()
for _, tx := range largeTxs {
    fmt.Printf("Large TX: %s, Amount: %d\n", tx.TxHash, tx.Amount)
}

// Get transaction statistics
stats := k.GetTransactionStats()
fmt.Printf("Total: %d, Success Rate: %.2f%%\n",
    stats["total_transactions"],
    stats["success_rate"])
```

## Alert Management

### Create Custom Alerts

```go
// Create a security alert
alert, err := k.CreateAlert(
    types.AlertTypeSecurityThreat,
    types.SeverityCritical,
    "Potential DDoS attack detected",
    map[string]interface{}{
        "source_ip":       "192.168.1.100",
        "requests_per_sec": 10000,
        "threshold":        1000,
        "duration":         "5m",
    },
)

if err != nil {
    log.Printf("Failed to create alert: %v", err)
} else {
    fmt.Printf("Alert created: %s\n", alert.ID)
}
```

### Acknowledge and Resolve Alerts

```go
// Get active alerts
activeAlerts := k.GetActiveAlerts()
for _, alert := range activeAlerts {
    if alert.Severity == types.SeverityCritical {
        // Acknowledge critical alerts
        err := k.AcknowledgeAlert(alert.ID, "ops@aura.network")
        if err != nil {
            log.Printf("Error acknowledging: %v", err)
            continue
        }

        // After investigation, resolve if appropriate
        time.Sleep(1 * time.Minute) // Investigation time
        err = k.ResolveAlert(alert.ID)
        if err != nil {
            log.Printf("Error resolving: %v", err)
        }
    }
}
```

### Filter and Query Alerts

```go
// Get alerts by severity
criticalAlerts := k.GetAlertsBySeverity(types.SeverityCritical)
fmt.Printf("Critical alerts: %d\n", len(criticalAlerts))

// Get alerts by type
securityAlerts := k.GetAlertsByType(types.AlertTypeSecurityThreat)
fmt.Printf("Security alerts: %d\n", len(securityAlerts))

// Get specific alert
alert, err := k.GetAlert("alert-123")
if err == nil {
    fmt.Printf("Alert: %s, Status: Resolved=%v, Ack=%v\n",
        alert.Message,
        alert.Resolved,
        alert.Acknowledged)
}
```

## Validator Monitoring

### Update Validator Uptime

```go
// Validator signed a block
err := k.UpdateValidatorUptime(
    "auravaloper1abc...",
    "ValidatorNode1",
    12345, // block height
    true,  // signed
)

// Validator missed a block
err = k.UpdateValidatorUptime(
    "auravaloper1abc...",
    "ValidatorNode1",
    12346,
    false, // missed
)
```

### Query Validator Status

```go
// Get specific validator
uptime, err := k.GetValidatorUptime("auravaloper1abc...")
if err == nil {
    fmt.Printf("Validator: %s\n", uptime.Moniker)
    fmt.Printf("Uptime: %.2f%%\n", uptime.UptimePercentage)
    fmt.Printf("Signed: %d, Missed: %d\n",
        uptime.SignedBlocks,
        uptime.MissedBlocks)
    fmt.Printf("Status: %s, Jailed: %v\n",
        uptime.Status,
        uptime.Jailed)
}

// Get all validators
allValidators := k.GetAllValidatorUptimes()
fmt.Printf("Total validators: %d\n", len(allValidators))

// Get jailed validators
jailedValidators := k.GetJailedValidators()
for _, v := range jailedValidators {
    fmt.Printf("Jailed: %s, Consecutive Misses: %d\n",
        v.Moniker,
        v.ConsecutiveMisses)
}

// Get validator statistics
stats := k.GetValidatorStats()
fmt.Printf("Active: %d, Jailed: %d, Avg Uptime: %.2f%%\n",
    stats["active_validators"],
    stats["jailed_validators"],
    stats["average_uptime"])
```

## Network Health Tracking

### Update Network Health

```go
health := &types.NetworkHealth{
    Timestamp:         time.Now(),
    BlockHeight:       12345,
    BlockTime:         6.2,
    TPS:               125.5,
    ActiveValidators:  75,
    TotalValidators:   100,
    NetworkHashRate:   1000000,
    PeerCount:         50,
    MempoolSize:       200,
    AverageGasPrice:   120,
    NetworkCongestion: 0.35,
    ConsensusHealth:   0.98,
}

err := k.UpdateNetworkHealth(health)
if err != nil {
    log.Printf("Failed to update health: %v", err)
}
```

### Monitor Network Health

```go
// Get current health
health := k.GetNetworkHealth()
fmt.Printf("Block Height: %d\n", health.BlockHeight)
fmt.Printf("TPS: %.2f\n", health.TPS)
fmt.Printf("Congestion: %.2f%%\n", health.NetworkCongestion*100)
fmt.Printf("Consensus Health: %.2f%%\n", health.ConsensusHealth*100)

// Check for issues
if health.NetworkCongestion > 0.8 {
    fmt.Println("WARNING: High network congestion!")
}
if health.ConsensusHealth < 0.9 {
    fmt.Println("WARNING: Low consensus health!")
}

// Get historical data
history := k.GetNetworkHealthHistory(24 * time.Hour)
fmt.Printf("Historical data points: %d\n", len(history))
```

## Gas Price Monitoring

### Track Gas Prices

```go
// Track current gas price
currentPrice := uint64(150)
err := k.TrackGasPrice(currentPrice)
if err != nil {
    log.Printf("Error tracking price: %v", err)
}

// Simulate price changes over time
prices := []uint64{150, 155, 160, 170, 450} // Last one is a spike
for _, price := range prices {
    k.TrackGasPrice(price)
    time.Sleep(1 * time.Minute)
}
```

### Analyze Gas Prices

```go
tracking := k.GetGasPriceTracking()

fmt.Printf("Current Price: %d\n", tracking.CurrentPrice)
fmt.Printf("Average Price: %d\n", tracking.AveragePrice)
fmt.Printf("Median Price: %d\n", tracking.MedianPrice)
fmt.Printf("Price Range: %d - %d\n", tracking.MinPrice, tracking.MaxPrice)
fmt.Printf("Volatility: %.4f\n", tracking.VolatilityScore)
fmt.Printf("Trend: %s\n", tracking.TrendDirection)

// Check price history
fmt.Printf("Price History (%d points):\n", len(tracking.PriceHistory))
for i, point := range tracking.PriceHistory {
    if i >= 10 { // Show first 10
        break
    }
    fmt.Printf("  %s: %d\n", point.Timestamp.Format("15:04:05"), point.Price)
}
```

## TVL Tracking

### Update TVL for Modules

```go
// Update TVL for different modules
k.UpdateTVL("dex", 5000000)
k.UpdateTVL("staking", 15000000)
k.UpdateTVL("lending", 8000000)
k.UpdateTVL("liquidity", 3000000)
```

### Monitor TVL

```go
tvl := k.GetTVLMonitoring()

fmt.Printf("Total TVL: %d\n", tvl.TotalTVL)
fmt.Printf("24h Change: %.2f%%\n", tvl.TVLChange24h)
fmt.Printf("7d Change: %.2f%%\n", tvl.TVLChange7d)

// Show TVL by module
fmt.Println("\nTVL by Module:")
for module, amount := range tvl.TVLByModule {
    percentage := float64(amount) / float64(tvl.TotalTVL) * 100
    fmt.Printf("  %s: %d (%.2f%%)\n", module, amount, percentage)
}

// Show largest pools
fmt.Println("\nTop Pools:")
for i, pool := range tvl.LargestPools {
    if i >= 5 { // Top 5
        break
    }
    fmt.Printf("  %d. %s: %d (%.2f%%)\n",
        i+1, pool.PoolName, pool.TVL, pool.Percentage)
}

// Get module-specific TVL
dexTVL, err := k.GetTVLByModule("dex")
if err == nil {
    fmt.Printf("\nDEX TVL: %d\n", dexTVL)
}
```

## Log Aggregation

### Log Various Events

```go
// Info log
k.LogEntry(
    types.LogLevelInfo,
    "transaction-processor",
    "Transaction processed successfully",
    map[string]interface{}{
        "tx_hash": "0x123...",
        "duration_ms": 45,
    },
    "trace-abc-123",
    "span-xyz-789",
)

// Error log
k.LogEntry(
    types.LogLevelError,
    "consensus",
    "Failed to reach consensus",
    map[string]interface{}{
        "block_height": 12345,
        "validators_voted": 65,
        "required_votes": 67,
    },
    "trace-def-456",
    "span-uvw-012",
)

// Warning log
k.LogEntry(
    types.LogLevelWarning,
    "network",
    "High peer churn detected",
    map[string]interface{}{
        "peers_joined": 10,
        "peers_left": 15,
        "interval": "5m",
    },
    "",
    "",
)
```

### Query Logs

```go
// Get all logs for a module
logs, err := k.GetLogs("consensus", 100)
if err == nil {
    fmt.Printf("Found %d logs\n", len(logs))
    for _, log := range logs {
        fmt.Printf("[%s] %s: %s\n",
            log.Level,
            log.Timestamp.Format("15:04:05"),
            log.Message)
    }
}

// Get error logs only
errorLogs, err := k.GetLogsByLevel("consensus", types.LogLevelError, 50)
fmt.Printf("Error logs: %d\n", len(errorLogs))

// Get all error logs across modules
allErrors := k.GetErrorLogs(100)
fmt.Printf("Total errors across all modules: %d\n", len(allErrors))

// Get logs by trace ID
traceLogs := k.GetLogsByTraceID("trace-abc-123")
fmt.Printf("Logs for trace: %d\n", len(traceLogs))

// Search logs
searchResults := k.SearchLogs("consensus", 50)
fmt.Printf("Search results: %d\n", len(searchResults))

// Get log statistics
stats := k.GetLogStats()
fmt.Printf("Total modules: %d\n", stats["total_modules"])
fmt.Printf("Total entries: %d\n", stats["total_entries"])

// Export logs for external system
exported, err := k.ExportLogs(
    "consensus",
    time.Now().Add(-24*time.Hour),
    time.Now(),
)
fmt.Printf("Exported %d logs\n", len(exported))
```

## SIEM Integration

### Record Security Events

```go
import "github.com/aequitas/aura/chain/x/monitoring/siem"

siemManager := siem.NewSIEMManager(7) // Threat threshold

// Record suspicious transaction
event, err := siemManager.RecordSecurityEvent(
    types.SecurityEventSuspiciousTransaction,
    types.SeverityHigh,
    "aura1attacker123...",
    "aura1victim456...",
    "Multiple high-value transactions from new address",
    map[string]interface{}{
        "tx_count": 50,
        "total_amount": 10000000,
        "time_window": "5m",
    },
    []string{"new_address", "high_frequency", "large_amount"},
    8, // Threat level
)

// Record DDoS attempt
ddosEvent, err := siemManager.RecordSecurityEvent(
    types.SecurityEventDDoSAttempt,
    types.SeverityCritical,
    "192.168.1.100",
    "",
    "DDoS attack detected from suspicious IP",
    map[string]interface{}{
        "requests_per_sec": 50000,
        "normal_rate": 500,
    },
    []string{"ddos", "rate_limit_exceeded"},
    9,
)
```

### Manage Security Events

```go
// Get unmitigated events
unmitigated := siemManager.GetUnmitigatedEvents()
fmt.Printf("Unmitigated events: %d\n", len(unmitigated))

// Mitigate an event
mitigationSteps := []string{
    "Blocked source IP",
    "Rate limiting applied",
    "Increased monitoring",
}
err := siemManager.MitigateSecurityEvent(ddosEvent.ID, mitigationSteps)

// Get high threat events
highThreat := siemManager.GetHighThreatEvents()
for _, event := range highThreat {
    fmt.Printf("HIGH THREAT: %s (Level %d)\n",
        event.Description,
        event.ThreatLevel)
}

// Get security statistics
stats := siemManager.GetSecurityStats()
fmt.Printf("Total Events: %d\n", stats["total_events"])
fmt.Printf("Mitigated: %d\n", stats["mitigated_events"])
fmt.Printf("High Threat: %d\n", stats["high_threat_events"])

// Analyze trends
trends := siemManager.AnalyzeThreatTrends(24 * time.Hour)
fmt.Printf("Recent events (24h): %d\n", trends["recent_events"])
fmt.Printf("Avg threat level: %.2f\n", trends["avg_threat_level"])
```

## Anomaly Detection

### Detect Transaction Anomalies

```go
// Anomaly detection is automatically performed when monitoring transactions
tx := &types.TransactionMonitorData{
    TxHash:      "0x789...",
    Sender:      "aura1sender...",
    Receiver:    "aura1receiver...",
    Amount:      50000000, // Very large amount
    GasUsed:     2000000,  // Very high gas
    GasPrice:    500,      // High gas price
    Status:      "success",
    Timestamp:   time.Date(2025, 1, 1, 3, 0, 0, 0, time.UTC), // 3 AM
    BlockHeight: 12345,
    Module:      "bank",
}

err := k.MonitorTransaction(tx)
// Anomaly will be detected if score is above threshold

// Get all detected anomalies
anomalies := k.GetAnomalies()
fmt.Printf("Total anomalies detected: %d\n", len(anomalies))

for _, anomaly := range anomalies {
    fmt.Printf("\nAnomaly: %s\n", anomaly.ID)
    fmt.Printf("Type: %s\n", anomaly.Type)
    fmt.Printf("Score: %.4f (threshold: %.4f)\n",
        anomaly.Score,
        anomaly.Threshold)
    fmt.Printf("Is Anomaly: %v\n", anomaly.IsAnomaly)
    fmt.Printf("Features: %v\n", anomaly.Features)
}

// Get anomalies by type
txAnomalies := k.GetAnomaliesByType(types.AnomalyTypeTransaction)
fmt.Printf("Transaction anomalies: %d\n", len(txAnomalies))
```

### Detect Network Anomalies

```go
// Create unusual network health state
unusualHealth := &types.NetworkHealth{
    Timestamp:         time.Now(),
    BlockHeight:       12345,
    BlockTime:         25.0,   // Very slow
    TPS:               5.0,    // Very low
    ActiveValidators:  20,     // Low count
    TotalValidators:   100,
    PeerCount:         5,      // Very low
    MempoolSize:       50000,  // Very high
    NetworkCongestion: 0.95,   // Almost full
    ConsensusHealth:   0.60,   // Poor health
}

detection, err := k.DetectNetworkAnomaly(unusualHealth)
if err == nil && detection.IsAnomaly {
    fmt.Printf("Network anomaly detected!\n")
    fmt.Printf("Score: %.4f\n", detection.Score)
    fmt.Printf("Features: %v\n", detection.Features)
}
```

## Best Practices

### 1. Error Handling

```go
// Always check errors
if err := k.MonitorTransaction(tx); err != nil {
    // Log the error
    k.LogEntry(
        types.LogLevelError,
        "monitoring",
        fmt.Sprintf("Failed to monitor transaction: %v", err),
        map[string]interface{}{"tx_hash": tx.TxHash},
        "",
        "",
    )
    // Handle appropriately
    return err
}
```

### 2. Resource Cleanup

```go
// Always close the keeper when done
defer k.Close()

// Or use context for automatic cleanup
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

### 3. Periodic Maintenance

```go
// Set up periodic maintenance tasks
ticker := time.NewTicker(1 * time.Hour)
defer ticker.Stop()

go func() {
    for range ticker.C {
        // Keeper automatically cleans up old data
        // But you can also manually trigger cleanup if needed
    }
}()
```

### 4. Metric Collection

```go
// Access Prometheus metrics
metrics := k.GetMetrics()

// Metrics are automatically updated by the keeper
// Just ensure Prometheus is scraping the endpoint
```

### 5. Alert Response

```go
// Set up alert monitoring
alerts := k.GetActiveAlerts()
for _, alert := range alerts {
    switch alert.Severity {
    case types.SeverityCritical:
        // Page on-call engineer
        sendPageAlert(alert)
        k.AcknowledgeAlert(alert.ID, "oncall@aura.network")

    case types.SeverityHigh:
        // Send to security team
        sendSecurityAlert(alert)

    case types.SeverityMedium:
        // Log for review
        k.LogEntry(types.LogLevelWarning, "alerts",
            alert.Message, alert.Details, "", "")
    }
}
```

## Complete Example Application

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/aequitas/aura/chain/x/monitoring/keeper"
    "github.com/aequitas/aura/chain/x/monitoring/types"
)

func main() {
    // Initialize monitoring
    params := types.DefaultParams()
    k := keeper.NewKeeper(&params)
    defer k.Close()

    log.Println("Monitoring system initialized")

    // Monitor a transaction
    tx := &types.TransactionMonitorData{
        TxHash:      "0xabc123",
        Sender:      "aura1sender",
        Receiver:    "aura1receiver",
        Amount:      2000000,
        GasUsed:     75000,
        GasPrice:    150,
        Status:      "success",
        Timestamp:   time.Now(),
        BlockHeight: 12345,
        Module:      "bank",
    }

    if err := k.MonitorTransaction(tx); err != nil {
        log.Fatalf("Transaction monitoring failed: %v", err)
    }

    // Update validator status
    if err := k.UpdateValidatorUptime("auravaloper1test", "Validator1", 12345, true); err != nil {
        log.Fatalf("Validator update failed: %v", err)
    }

    // Track gas price
    if err := k.TrackGasPrice(150); err != nil {
        log.Fatalf("Gas price tracking failed: %v", err)
    }

    // Update network health
    health := &types.NetworkHealth{
        Timestamp:         time.Now(),
        BlockHeight:       12345,
        BlockTime:         6.5,
        TPS:               100.0,
        ActiveValidators:  50,
        TotalValidators:   100,
        NetworkCongestion: 0.3,
        ConsensusHealth:   0.95,
    }

    if err := k.UpdateNetworkHealth(health); err != nil {
        log.Fatalf("Health update failed: %v", err)
    }

    // Check for alerts
    alerts := k.GetActiveAlerts()
    fmt.Printf("Active alerts: %d\n", len(alerts))

    // Get statistics
    txStats := k.GetTransactionStats()
    fmt.Printf("Transaction success rate: %.2f%%\n", txStats["success_rate"])

    valStats := k.GetValidatorStats()
    fmt.Printf("Active validators: %d\n", valStats["active_validators"])

    log.Println("Monitoring check complete")
}
```
