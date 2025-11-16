package types

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper functions for SchedulerConfig

// IsOffPeakHour checks if the given hour is an off-peak hour
func IsOffPeakHour(s *SchedulerConfig, hour uint32) bool {
	for _, offPeakHour := range s.OffPeakHours {
		if offPeakHour == hour {
			return true
		}
	}
	return false
}

// ShouldRunScheduler checks if the scheduler should run at the given time
func ShouldRunScheduler(s *SchedulerConfig, t time.Time) bool {
	if !s.Enabled {
		return false
	}

	hour := uint32(t.Hour())
	isOffPeak := IsOffPeakHour(s, hour)

	return isOffPeak || s.AllowPeakHours
}

// Helper functions for AutoScalingConfig

// GetInitialAmount returns the initial amount for a transaction type
func GetInitialAmount(a *AutoScalingConfig, txType TransactionType) uint64 {
	if amount, ok := a.InitialAmounts[txType.String()]; ok {
		return amount
	}
	return 10 // Default fallback
}

// GetMaxAmount returns the maximum amount for a transaction type
func GetMaxAmount(a *AutoScalingConfig, txType TransactionType) uint64 {
	if amount, ok := a.MaxAmounts[txType.String()]; ok {
		return amount
	}
	return 100 // Default fallback
}

// ShouldScaleUp determines if we should scale up based on metrics
func ShouldScaleUp(a *AutoScalingConfig, metrics *TypeMetrics) bool {
	if !a.Enabled {
		return false
	}
	return metrics.CacheHitRate > a.TargetCacheHitRate
}

// ShouldScaleDown determines if we should scale down based on metrics
func ShouldScaleDown(a *AutoScalingConfig, metrics *TypeMetrics) bool {
	if !a.Enabled {
		return false
	}
	return metrics.CacheHitRate < a.MinCacheHitRate
}

// CalculateNewAmount calculates the new amount after scaling
func CalculateNewAmount(a *AutoScalingConfig, currentAmount uint64, scaleUp bool, txType TransactionType) uint64 {
	var newAmount uint64
	if scaleUp {
		newAmount = uint64(float64(currentAmount) * a.ScaleUpFactor)
		maxAmount := GetMaxAmount(a, txType)
		if newAmount > maxAmount {
			newAmount = maxAmount
		}
	} else {
		newAmount = uint64(float64(currentAmount) * a.ScaleDownFactor)
		initialAmount := GetInitialAmount(a, txType)
		if newAmount < initialAmount {
			newAmount = initialAmount
		}
	}
	return newAmount
}

// Helper functions for PreValidatedTransaction

// IsExpired checks if a pre-validated transaction has expired
func IsExpired(p *PreValidatedTransaction, currentTime time.Time) bool {
	if p.ExpiresAt == nil {
		return false
	}
	return currentTime.After(p.ExpiresAt.AsTime())
}

// CanExecute checks if a pre-validated transaction can be executed
func CanExecute(p *PreValidatedTransaction, currentTime time.Time) bool {
	if p.Status != ValidationStatusValidated {
		return false
	}
	return !IsExpired(p, currentTime)
}

// MarkExecuted marks the transaction as executed
func MarkExecuted(p *PreValidatedTransaction, blockHeight uint64, executionTime time.Time) {
	p.Status = ValidationStatusExecuted
	p.ExecutedAt = timestamppb.New(executionTime)
	p.ExecutedHeight = blockHeight
}

// MarkExpired marks the transaction as expired
func MarkExpired(p *PreValidatedTransaction) {
	p.Status = ValidationStatusExpired
}

// Helper functions for PreValidationMetrics

// UpdateCacheHitRate updates the cache hit rate
func UpdateCacheHitRate(m *PreValidationMetrics) {
	total := m.TotalCacheHits + m.TotalCacheMisses
	if total > 0 {
		m.OverallCacheHitRate = float64(m.TotalCacheHits) / float64(total)
	}
}

// RecordCacheHit records a cache hit
func RecordCacheHit(m *PreValidationMetrics, txType TransactionType, timeSavedMs uint64) {
	m.TotalCacheHits++
	UpdateCacheHitRate(m)

	// Update type-specific metrics
	if m.MetricsByType == nil {
		m.MetricsByType = make(map[string]*TypeMetrics)
	}

	key := txType.String()
	if _, ok := m.MetricsByType[key]; !ok {
		m.MetricsByType[key] = &TypeMetrics{TxType: txType}
	}

	typeMetrics := m.MetricsByType[key]
	typeMetrics.CacheHits++
	typeMetrics.TotalExecuted++

	// Update average time savings
	totalSavings := typeMetrics.AvgTimeSavingsMs * float64(typeMetrics.CacheHits-1)
	totalSavings += float64(timeSavedMs)
	typeMetrics.AvgTimeSavingsMs = totalSavings / float64(typeMetrics.CacheHits)

	// Update cache hit rate for this type
	totalTypeRequests := typeMetrics.CacheHits + typeMetrics.CacheMisses
	if totalTypeRequests > 0 {
		typeMetrics.CacheHitRate = float64(typeMetrics.CacheHits) / float64(totalTypeRequests)
	}
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(m *PreValidationMetrics, txType TransactionType) {
	m.TotalCacheMisses++
	UpdateCacheHitRate(m)

	// Update type-specific metrics
	if m.MetricsByType == nil {
		m.MetricsByType = make(map[string]*TypeMetrics)
	}

	key := txType.String()
	if _, ok := m.MetricsByType[key]; !ok {
		m.MetricsByType[key] = &TypeMetrics{TxType: txType}
	}

	typeMetrics := m.MetricsByType[key]
	typeMetrics.CacheMisses++

	// Update cache hit rate for this type
	totalTypeRequests := typeMetrics.CacheHits + typeMetrics.CacheMisses
	if totalTypeRequests > 0 {
		typeMetrics.CacheHitRate = float64(typeMetrics.CacheHits) / float64(totalTypeRequests)
	}
}

// GetTypeMetrics returns metrics for a specific transaction type
func GetTypeMetrics(m *PreValidationMetrics, txType TransactionType) *TypeMetrics {
	if m.MetricsByType == nil {
		return nil
	}
	return m.MetricsByType[txType.String()]
}
