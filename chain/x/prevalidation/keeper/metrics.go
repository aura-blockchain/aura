package keeper

import (
	"math"
	"sort"
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UpdateMetrics updates the metrics based on current state
func (k *Keeper) UpdateMetrics() {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Update overall cache hit rate
	k.metrics.UpdateCacheHitRate()

	// Calculate time savings
	k.calculateTimeSavings()

	// Calculate energy savings
	k.calculateEnergySavings()

	// Update hourly metrics
	k.updateHourlyMetrics()

	// Prune old hourly metrics (keep only last 24 hours)
	k.pruneHourlyMetrics()
}

// calculateTimeSavings calculates time savings based on metrics
func (k *Keeper) calculateTimeSavings() {
	totalTimeSaved := uint64(0)
	totalExecuted := uint64(0)

	for _, typeMetrics := range k.metrics.MetricsByType {
		if typeMetrics.TotalExecuted > 0 {
			timeSaved := uint64(typeMetrics.AvgTimeSavingsMs * float64(typeMetrics.TotalExecuted))
			totalTimeSaved += timeSaved
			totalExecuted += typeMetrics.TotalExecuted
		}
	}

	k.metrics.TotalTimeSavedMs = totalTimeSaved

	if totalExecuted > 0 {
		k.metrics.AvgTimeSavingsMs = float64(totalTimeSaved) / float64(totalExecuted)
	}
}

// calculateEnergySavings calculates energy savings
func (k *Keeper) calculateEnergySavings() {
	params := k.GetParams()

	// Energy saved = (normal execution cost - pre-validation cost) * executions
	// Normal execution is more expensive than pre-validation
	energyPerExecution := params.EnergyCostPerExecutionKwh
	energyPerValidation := params.EnergyCostPerValidationKwh

	// Energy saved per transaction
	energySavedPerTx := energyPerExecution - energyPerValidation

	// Total energy saved
	totalEnergySaved := float64(k.metrics.TotalExecuted) * energySavedPerTx

	k.metrics.TotalEnergySavedKwh = totalEnergySaved
}

// updateHourlyMetrics updates metrics for the current hour
func (k *Keeper) updateHourlyMetrics() {
	currentTime := time.Unix(k.currentTime, 0)
	currentHour := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), 0, 0, 0, currentTime.Location())

	// Find or create current hour metrics
	var currentHourMetrics *types.HourlyMetrics
	if k.metrics.CurrentHour != nil && k.metrics.CurrentHour.Hour != nil {
		hourTime := k.metrics.CurrentHour.Hour.AsTime()
		if hourTime.Equal(currentHour) {
			currentHourMetrics = k.metrics.CurrentHour
		}
	}

	if currentHourMetrics == nil {
		currentHourMetrics = &types.HourlyMetrics{
			Hour: timestamppb.New(currentHour),
		}
		k.metrics.CurrentHour = currentHourMetrics
	}

	// Calculate metrics for current hour
	// This is a simplified version - in production, track incremental changes
	currentHourMetrics.CacheHits = k.metrics.TotalCacheHits
	currentHourMetrics.CacheMisses = k.metrics.TotalCacheMisses

	if k.metrics.TotalExecuted > 0 {
		currentHourMetrics.AvgTimeSavingsMs = k.metrics.AvgTimeSavingsMs
	}

	// Energy for current hour (simplified)
	params := k.GetParams()
	energySavedPerTx := params.EnergyCostPerExecutionKwh - params.EnergyCostPerValidationKwh
	currentHourMetrics.EnergySavedKwh = float64(currentHourMetrics.TransactionsExecuted) * energySavedPerTx
}

// pruneHourlyMetrics keeps only the last 24 hours of metrics
func (k *Keeper) pruneHourlyMetrics() {
	if len(k.metrics.Last24Hours) == 0 {
		return
	}

	currentTime := time.Unix(k.currentTime, 0)
	cutoffTime := currentTime.Add(-24 * time.Hour)

	pruned := []*types.HourlyMetrics{}
	for _, hourMetrics := range k.metrics.Last24Hours {
		if hourMetrics.Hour != nil && hourMetrics.Hour.AsTime().After(cutoffTime) {
			pruned = append(pruned, hourMetrics)
		}
	}

	k.metrics.Last24Hours = pruned
}

// RecordControlGroupExecution records execution time for a control group transaction
func (k *Keeper) RecordControlGroupExecution(executionTimeMs float64) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.metrics.ControlGroup == nil {
		k.metrics.ControlGroup = &types.ControlGroupMetrics{}
	}

	k.metrics.ControlGroup.TotalTransactions++

	// Update running average
	totalTime := k.metrics.ControlGroup.AvgExecutionTimeMs * float64(k.metrics.ControlGroup.TotalTransactions-1)
	totalTime += executionTimeMs
	k.metrics.ControlGroup.AvgExecutionTimeMs = totalTime / float64(k.metrics.ControlGroup.TotalTransactions)
}

// ShouldUseControlGroup determines if a transaction should be in the control group
func (k *Keeper) ShouldUseControlGroup(txHash []byte) bool {
	params := k.GetParams()

	if !params.MetricsEnabled || params.ControlGroupPercentage <= 0 {
		return false
	}

	// Use hash-based selection for deterministic control group assignment
	hashValue := uint64(0)
	for i := 0; i < 8 && i < len(txHash); i++ {
		hashValue = (hashValue << 8) | uint64(txHash[i])
	}

	// Check if hash falls within control group percentage
	threshold := uint64(float64(math.MaxUint64) * (params.ControlGroupPercentage / 100.0))
	return hashValue < threshold
}

// CalculatePercentiles calculates execution time percentiles for control group
func (k *Keeper) CalculatePercentiles(executionTimes []float64) {
	if len(executionTimes) == 0 {
		return
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	if k.metrics.ControlGroup == nil {
		k.metrics.ControlGroup = &types.ControlGroupMetrics{}
	}

	// Sort execution times
	sorted := make([]float64, len(executionTimes))
	copy(sorted, executionTimes)
	sort.Float64s(sorted)

	// Calculate median (50th percentile)
	medianIdx := len(sorted) / 2
	if len(sorted)%2 == 0 {
		k.metrics.ControlGroup.MedianExecutionTimeMs = (sorted[medianIdx-1] + sorted[medianIdx]) / 2
	} else {
		k.metrics.ControlGroup.MedianExecutionTimeMs = sorted[medianIdx]
	}

	// Calculate 95th percentile
	p95Idx := int(float64(len(sorted)) * 0.95)
	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	k.metrics.ControlGroup.P95ExecutionTimeMs = sorted[p95Idx]

	// Calculate 99th percentile
	p99Idx := int(float64(len(sorted)) * 0.99)
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}
	k.metrics.ControlGroup.P99ExecutionTimeMs = sorted[p99Idx]

	// Calculate standard deviation
	mean := k.metrics.ControlGroup.AvgExecutionTimeMs
	variance := 0.0
	for _, time := range sorted {
		diff := time - mean
		variance += diff * diff
	}
	variance /= float64(len(sorted))
	k.metrics.ControlGroup.ExecutionTimeStdDevMs = math.Sqrt(variance)
}

// GetMetricsSummary returns a human-readable summary of metrics
func (k *Keeper) GetMetricsSummary() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	summary := map[string]interface{}{
		"total_pre_validations": k.metrics.TotalPreValidations,
		"total_executed":        k.metrics.TotalExecuted,
		"total_expired":         k.metrics.TotalExpired,
		"cache_hits":            k.metrics.TotalCacheHits,
		"cache_misses":          k.metrics.TotalCacheMisses,
		"cache_hit_rate":        k.metrics.OverallCacheHitRate,
		"avg_time_savings_ms":   k.metrics.AvgTimeSavingsMs,
		"total_time_saved_ms":   k.metrics.TotalTimeSavedMs,
		"energy_saved_kwh":      k.metrics.TotalEnergySavedKwh,
	}

	// Add per-type metrics
	typeMetrics := make(map[string]interface{})
	for txTypeStr, metrics := range k.metrics.MetricsByType {
		typeMetrics[txTypeStr] = map[string]interface{}{
			"pre_validated":       metrics.TotalPreValidated,
			"executed":            metrics.TotalExecuted,
			"expired":             metrics.TotalExpired,
			"cache_hits":          metrics.CacheHits,
			"cache_misses":        metrics.CacheMisses,
			"hit_rate":            metrics.CacheHitRate,
			"avg_time_savings_ms": metrics.AvgTimeSavingsMs,
		}
	}
	summary["by_type"] = typeMetrics

	// Add control group metrics
	if k.metrics.ControlGroup != nil {
		summary["control_group"] = map[string]interface{}{
			"total_transactions":  k.metrics.ControlGroup.TotalTransactions,
			"avg_execution_ms":    k.metrics.ControlGroup.AvgExecutionTimeMs,
			"median_execution_ms": k.metrics.ControlGroup.MedianExecutionTimeMs,
			"p95_execution_ms":    k.metrics.ControlGroup.P95ExecutionTimeMs,
			"p99_execution_ms":    k.metrics.ControlGroup.P99ExecutionTimeMs,
			"std_dev_ms":          k.metrics.ControlGroup.ExecutionTimeStdDevMs,
		}
	}

	return summary
}

// EmitMetricsEvent creates a metrics update event
func (k *Keeper) EmitMetricsEvent() *types.EventMetricsUpdate {
	k.UpdateMetrics()

	return &types.EventMetricsUpdate{
		Metrics:   k.metrics,
		Timestamp: timestamppb.New(time.Unix(k.currentTime, 0)),
	}
}

// GetCacheStatistics returns detailed cache statistics
func (k *Keeper) GetCacheStatistics() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	params := k.GetParams()

	stats := map[string]interface{}{
		"cache_size":        len(k.preValidatedTxs),
		"max_cache_size":    params.MaxCacheSize,
		"cache_utilization": float64(len(k.preValidatedTxs)) / float64(params.MaxCacheSize),
		"cache_strategy":    params.CacheStrategy.String(),
	}

	// Count by status
	statusCounts := make(map[string]int)
	for _, tx := range k.preValidatedTxs {
		statusCounts[tx.Status.String()]++
	}
	stats["by_status"] = statusCounts

	// Count by type
	typeCounts := make(map[string]int)
	for _, tx := range k.preValidatedTxs {
		typeCounts[tx.TxType.String()]++
	}
	stats["by_type"] = typeCounts

	return stats
}

// ResetMetrics resets all metrics (for testing/admin)
func (k *Keeper) ResetMetrics() {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.metrics = &types.PreValidationMetrics{
		MetricsByType: make(map[string]*types.TypeMetrics),
		Last24Hours:   []*types.HourlyMetrics{},
		CurrentHour:   nil,
		ControlGroup:  &types.ControlGroupMetrics{},
	}
}

// ExportMetrics exports metrics in a format suitable for external monitoring systems
func (k *Keeper) ExportMetrics() map[string]float64 {
	k.mu.RLock()
	defer k.mu.RUnlock()

	metrics := make(map[string]float64)

	// Overall metrics
	metrics["prevalidation.total.created"] = float64(k.metrics.TotalPreValidations)
	metrics["prevalidation.total.executed"] = float64(k.metrics.TotalExecuted)
	metrics["prevalidation.total.expired"] = float64(k.metrics.TotalExpired)
	metrics["prevalidation.cache.hits"] = float64(k.metrics.TotalCacheHits)
	metrics["prevalidation.cache.misses"] = float64(k.metrics.TotalCacheMisses)
	metrics["prevalidation.cache.hit_rate"] = k.metrics.OverallCacheHitRate
	metrics["prevalidation.time_savings.avg_ms"] = k.metrics.AvgTimeSavingsMs
	metrics["prevalidation.time_savings.total_ms"] = float64(k.metrics.TotalTimeSavedMs)
	metrics["prevalidation.energy.saved_kwh"] = k.metrics.TotalEnergySavedKwh

	// Per-type metrics
	for txTypeStr, typeMetrics := range k.metrics.MetricsByType {
		prefix := "prevalidation." + txTypeStr
		metrics[prefix+".created"] = float64(typeMetrics.TotalPreValidated)
		metrics[prefix+".executed"] = float64(typeMetrics.TotalExecuted)
		metrics[prefix+".expired"] = float64(typeMetrics.TotalExpired)
		metrics[prefix+".cache.hits"] = float64(typeMetrics.CacheHits)
		metrics[prefix+".cache.misses"] = float64(typeMetrics.CacheMisses)
		metrics[prefix+".cache.hit_rate"] = typeMetrics.CacheHitRate
		metrics[prefix+".time_savings.avg_ms"] = typeMetrics.AvgTimeSavingsMs
	}

	// Control group metrics
	if k.metrics.ControlGroup != nil {
		metrics["prevalidation.control_group.transactions"] = float64(k.metrics.ControlGroup.TotalTransactions)
		metrics["prevalidation.control_group.avg_execution_ms"] = k.metrics.ControlGroup.AvgExecutionTimeMs
		metrics["prevalidation.control_group.median_execution_ms"] = k.metrics.ControlGroup.MedianExecutionTimeMs
		metrics["prevalidation.control_group.p95_execution_ms"] = k.metrics.ControlGroup.P95ExecutionTimeMs
		metrics["prevalidation.control_group.p99_execution_ms"] = k.metrics.ControlGroup.P99ExecutionTimeMs
	}

	return metrics
}
