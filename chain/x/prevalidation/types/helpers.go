package types

import (
	"time"

	"cosmossdk.io/math"
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
	return metrics.CacheHitRate.GT(a.TargetCacheHitRate)
}

// ShouldScaleDown determines if we should scale down based on metrics
func ShouldScaleDown(a *AutoScalingConfig, metrics *TypeMetrics) bool {
	if !a.Enabled {
		return false
	}
	return metrics.CacheHitRate.LT(a.MinCacheHitRate)
}

// CalculateNewAmount calculates the new amount after scaling
func CalculateNewAmount(a *AutoScalingConfig, currentAmount uint64, scaleUp bool, txType TransactionType) uint64 {
	var newAmount uint64
	if scaleUp {
		currentDec := math.LegacyNewDec(int64(currentAmount))
		scaledDec := currentDec.Mul(a.ScaleUpFactor)
		scaledInt64 := scaledDec.TruncateInt64() // Convert to int64, truncate decimal
		if scaledInt64 < 0 {
			newAmount = 0
		} else {
			newAmount = uint64(scaledInt64)
		}
		maxAmount := GetMaxAmount(a, txType)
		if newAmount > maxAmount {
			newAmount = maxAmount
		}
	} else {
		currentDec := math.LegacyNewDec(int64(currentAmount))
		scaledDec := currentDec.Mul(a.ScaleDownFactor)
		scaledInt64 := scaledDec.TruncateInt64()
		if scaledInt64 < 0 {
			newAmount = 0
		} else {
			newAmount = uint64(scaledInt64)
		}
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
	// ExpiresAt is a time.Time value (non-nullable in proto)
	// Zero time means no expiry
	if p.ExpiresAt.IsZero() {
		return false
	}
	return currentTime.After(p.ExpiresAt)
}

// CanExecute checks if a pre-validated transaction can be executed
func CanExecute(p *PreValidatedTransaction, currentTime time.Time) bool {
	if p.Status != ValidationStatus_VALIDATION_STATUS_VALIDATED {
		return false
	}
	return !IsExpired(p, currentTime)
}

// MarkExecuted marks the transaction as executed
func MarkExecuted(p *PreValidatedTransaction, blockHeight uint64, executionTime time.Time) {
	p.Status = ValidationStatus_VALIDATION_STATUS_EXECUTED
	p.ExecutedAt = &executionTime
	p.ExecutedHeight = blockHeight
}

// MarkExpired marks the transaction as expired
func MarkExpired(p *PreValidatedTransaction) {
	p.Status = ValidationStatus_VALIDATION_STATUS_EXPIRED
}

// Helper functions for PreValidationMetrics

// UpdateCacheHitRate updates the cache hit rate
func UpdateCacheHitRate(m *PreValidationMetrics) {
	total := m.TotalCacheHits + m.TotalCacheMisses
	if total == 0 {
		m.OverallCacheHitRate = math.LegacyNewDec(0)
		return
	}
	hits := math.LegacyNewDec(int64(m.TotalCacheHits))
	totalDec := math.LegacyNewDec(int64(total))
	m.OverallCacheHitRate = hits.Quo(totalDec)
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
		m.MetricsByType[key] = &TypeMetrics{
			TxType:           txType,
			AvgTimeSavingsMs: math.LegacyNewDec(0),
			CacheHitRate:     math.LegacyNewDec(0),
		}
	}

	typeMetrics := m.MetricsByType[key]
	if typeMetrics.AvgTimeSavingsMs.IsNil() {
		typeMetrics.AvgTimeSavingsMs = math.LegacyNewDec(0)
	}
	if typeMetrics.CacheHitRate.IsNil() {
		typeMetrics.CacheHitRate = math.LegacyNewDec(0)
	}
	typeMetrics.CacheHits++
	typeMetrics.TotalExecuted++

	// Update average time savings using running average formula
	prevCount := math.LegacyNewDec(int64(typeMetrics.CacheHits - 1))
	totalSavings := typeMetrics.AvgTimeSavingsMs.Mul(prevCount)
	timeSavedDec := math.LegacyNewDec(int64(timeSavedMs))
	totalSavings = totalSavings.Add(timeSavedDec)
	cacheHitsDec := math.LegacyNewDec(int64(typeMetrics.CacheHits))
	typeMetrics.AvgTimeSavingsMs = totalSavings.Quo(cacheHitsDec)

	// Update cache hit rate for this type
	totalTypeRequests := typeMetrics.CacheHits + typeMetrics.CacheMisses
	if totalTypeRequests > 0 {
		hits := math.LegacyNewDec(int64(typeMetrics.CacheHits))
		totalDec := math.LegacyNewDec(int64(totalTypeRequests))
		typeMetrics.CacheHitRate = hits.Quo(totalDec)
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
		hits := math.LegacyNewDec(int64(typeMetrics.CacheHits))
		totalDec := math.LegacyNewDec(int64(totalTypeRequests))
		typeMetrics.CacheHitRate = hits.Quo(totalDec)
	}
}

// GetTypeMetrics returns metrics for a specific transaction type
func GetTypeMetrics(m *PreValidationMetrics, txType TransactionType) *TypeMetrics {
	if m.MetricsByType == nil {
		return nil
	}
	return m.MetricsByType[txType.String()]
}
