package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.NotNil(t, params)
	require.True(t, params.Enabled)
	require.NotNil(t, params.SchedulerConfig)
	require.NotNil(t, params.AutoScalingConfig)
	require.Equal(t, CacheStrategy_CACHE_STRATEGY_LRU, params.CacheStrategy)
	require.Equal(t, uint64(10000), params.MaxCacheSize)
	require.Equal(t, uint32(24), params.ExpiryHours)
	require.Equal(t, "AES-256-GCM", params.EncryptionAlgorithm)
	require.Equal(t, float64(10.0), params.ControlGroupPercentage)
	require.Equal(t, uint64(50), params.MinConfidenceScore)
	require.True(t, params.MetricsEnabled)
	require.False(t, params.DetailedLogging)
	require.Equal(t, uint32(3), params.MaxValidationAttempts)
	require.Equal(t, uint32(5), params.RetryDelaySeconds)

	// Validate default params
	require.NoError(t, ValidateParams(params))
}

func TestValidateParams_Valid(t *testing.T) {
	params := DefaultParams()
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_NilParams(t *testing.T) {
	err := ValidateParams(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "params cannot be nil")
}

func TestValidateParams_ControlGroupPercentage(t *testing.T) {
	tests := []struct {
		name       string
		percentage float64
		wantError  bool
	}{
		{"valid 0.0", 0.0, false},
		{"valid 50.0", 50.0, false},
		{"valid 100.0", 100.0, false},
		{"invalid negative", -1.0, true},
		{"invalid above 100", 100.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := DefaultParams()
			params.ControlGroupPercentage = tt.percentage
			err := ValidateParams(params)
			if tt.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), "control_group_percentage must be between 0.0 and 100.0")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateParams_MaxCacheSize(t *testing.T) {
	params := DefaultParams()
	params.MaxCacheSize = 0
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max_cache_size must be greater than 0")
}

func TestValidateParams_ExpiryHours(t *testing.T) {
	params := DefaultParams()
	params.ExpiryHours = 0
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expiry_hours must be greater than 0")
}

func TestIsOffPeakHour(t *testing.T) {
	config := &SchedulerConfig{
		OffPeakHours: []uint32{0, 1, 2, 3, 22, 23},
	}

	tests := []struct {
		hour     uint32
		expected bool
	}{
		{0, true},
		{1, true},
		{2, true},
		{3, true},
		{12, false},
		{22, true},
		{23, true},
	}

	for _, tt := range tests {
		t.Run("hour_"+string(rune(tt.hour+'0')), func(t *testing.T) {
			result := IsOffPeakHour(config, tt.hour)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldRunScheduler(t *testing.T) {
	tests := []struct {
		name     string
		config   *SchedulerConfig
		time     time.Time
		expected bool
	}{
		{
			name: "disabled scheduler",
			config: &SchedulerConfig{
				Enabled: false,
			},
			time:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name: "off-peak hour",
			config: &SchedulerConfig{
				Enabled:        true,
				OffPeakHours:   []uint32{2},
				AllowPeakHours: false,
			},
			time:     time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name: "peak hour without allow",
			config: &SchedulerConfig{
				Enabled:        true,
				OffPeakHours:   []uint32{2},
				AllowPeakHours: false,
			},
			time:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name: "peak hour with allow",
			config: &SchedulerConfig{
				Enabled:        true,
				OffPeakHours:   []uint32{2},
				AllowPeakHours: true,
			},
			time:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRunScheduler(tt.config, tt.time)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetInitialAmount(t *testing.T) {
	config := &AutoScalingConfig{
		InitialAmounts: map[string]uint64{
			"TX_TYPE_IR_COMPLETION": 20,
			"TX_TYPE_DEX_SWAP":      30,
		},
	}

	tests := []struct {
		name     string
		txType   TransactionType
		expected uint64
	}{
		{"existing type", TransactionType_TX_TYPE_IR_COMPLETION, 20},
		{"another existing type", TransactionType_TX_TYPE_DEX_SWAP, 30},
		{"non-existing type", TransactionType_TX_TYPE_VC_MINT, 10}, // default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetInitialAmount(config, tt.txType)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMaxAmount(t *testing.T) {
	config := &AutoScalingConfig{
		MaxAmounts: map[string]uint64{
			"TX_TYPE_IR_COMPLETION": 200,
			"TX_TYPE_DEX_SWAP":      300,
		},
	}

	tests := []struct {
		name     string
		txType   TransactionType
		expected uint64
	}{
		{"existing type", TransactionType_TX_TYPE_IR_COMPLETION, 200},
		{"another existing type", TransactionType_TX_TYPE_DEX_SWAP, 300},
		{"non-existing type", TransactionType_TX_TYPE_VC_MINT, 100}, // default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMaxAmount(config, tt.txType)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldScaleUp(t *testing.T) {
	config := &AutoScalingConfig{
		Enabled:            true,
		TargetCacheHitRate: 0.8,
	}

	tests := []struct {
		name     string
		metrics  *TypeMetrics
		expected bool
	}{
		{
			name: "high cache hit rate",
			metrics: &TypeMetrics{
				CacheHitRate: 0.85,
			},
			expected: true,
		},
		{
			name: "low cache hit rate",
			metrics: &TypeMetrics{
				CacheHitRate: 0.7,
			},
			expected: false,
		},
		{
			name: "exact target rate",
			metrics: &TypeMetrics{
				CacheHitRate: 0.8,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldScaleUp(config, tt.metrics)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldScaleDown(t *testing.T) {
	config := &AutoScalingConfig{
		Enabled:         true,
		MinCacheHitRate: 0.5,
	}

	tests := []struct {
		name     string
		metrics  *TypeMetrics
		expected bool
	}{
		{
			name: "low cache hit rate",
			metrics: &TypeMetrics{
				CacheHitRate: 0.4,
			},
			expected: true,
		},
		{
			name: "high cache hit rate",
			metrics: &TypeMetrics{
				CacheHitRate: 0.6,
			},
			expected: false,
		},
		{
			name: "exact min rate",
			metrics: &TypeMetrics{
				CacheHitRate: 0.5,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldScaleDown(config, tt.metrics)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateNewAmount(t *testing.T) {
	config := &AutoScalingConfig{
		ScaleUpFactor:   1.5,
		ScaleDownFactor: 0.7,
		InitialAmounts: map[string]uint64{
			"TX_TYPE_IR_COMPLETION": 10,
		},
		MaxAmounts: map[string]uint64{
			"TX_TYPE_IR_COMPLETION": 100,
		},
	}

	tests := []struct {
		name          string
		currentAmount uint64
		scaleUp       bool
		expected      uint64
	}{
		{"scale up normal", 50, true, 75},
		{"scale up with cap", 80, true, 100},
		{"scale down normal", 50, false, 35},
		{"scale down with floor", 12, false, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateNewAmount(config, tt.currentAmount, tt.scaleUp, TransactionType_TX_TYPE_IR_COMPLETION)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExpired(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name     string
		tx       *PreValidatedTransaction
		now      time.Time
		expected bool
	}{
		{
			name: "expired transaction",
			tx: &PreValidatedTransaction{
				ExpiresAt: past, // time.Time value, not pointer
			},
			now:      now,
			expected: true,
		},
		{
			name: "not expired transaction",
			tx: &PreValidatedTransaction{
				ExpiresAt: future, // time.Time value, not pointer
			},
			now:      now,
			expected: false,
		},
		{
			name: "zero time (no expiry)",
			tx: &PreValidatedTransaction{
				ExpiresAt: time.Time{}, // Zero time means no expiry
			},
			now:      now,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExpired(tt.tx, tt.now)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCanExecute(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name     string
		tx       *PreValidatedTransaction
		now      time.Time
		expected bool
	}{
		{
			name: "validated and not expired",
			tx: &PreValidatedTransaction{
				Status:    ValidationStatus_VALIDATION_STATUS_VALIDATED,
				ExpiresAt: future, // time.Time value, not pointer
			},
			now:      now,
			expected: true,
		},
		{
			name: "validated but expired",
			tx: &PreValidatedTransaction{
				Status:    ValidationStatus_VALIDATION_STATUS_VALIDATED,
				ExpiresAt: past, // time.Time value, not pointer
			},
			now:      now,
			expected: false,
		},
		{
			name: "pending status",
			tx: &PreValidatedTransaction{
				Status:    ValidationStatus_VALIDATION_STATUS_PENDING,
				ExpiresAt: future, // time.Time value, not pointer
			},
			now:      now,
			expected: false,
		},
		{
			name: "already executed",
			tx: &PreValidatedTransaction{
				Status:    ValidationStatus_VALIDATION_STATUS_EXECUTED,
				ExpiresAt: future, // time.Time value, not pointer
			},
			now:      now,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanExecute(tt.tx, tt.now)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestMarkExecuted(t *testing.T) {
	tx := &PreValidatedTransaction{
		Status: ValidationStatus_VALIDATION_STATUS_VALIDATED,
	}

	blockHeight := uint64(1000)
	executionTime := time.Now()

	MarkExecuted(tx, blockHeight, executionTime)

	require.Equal(t, ValidationStatus_VALIDATION_STATUS_EXECUTED, tx.Status)
	require.NotNil(t, tx.ExecutedAt)
	require.Equal(t, blockHeight, tx.ExecutedHeight)
	// ExecutedAt is *time.Time, so dereference it to compare
	require.Equal(t, executionTime.Unix(), tx.ExecutedAt.Unix())
}

func TestMarkExpired(t *testing.T) {
	tx := &PreValidatedTransaction{
		Status: ValidationStatus_VALIDATION_STATUS_VALIDATED,
	}

	MarkExpired(tx)

	require.Equal(t, ValidationStatus_VALIDATION_STATUS_EXPIRED, tx.Status)
}

func TestUpdateCacheHitRate(t *testing.T) {
	tests := []struct {
		name         string
		cacheHits    uint64
		cacheMisses  uint64
		expectedRate float64
	}{
		{"perfect hit rate", 100, 0, 1.0},
		{"50% hit rate", 50, 50, 0.5},
		{"no hits", 0, 100, 0.0},
		{"zero total", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PreValidationMetrics{
				TotalCacheHits:   tt.cacheHits,
				TotalCacheMisses: tt.cacheMisses,
			}

			UpdateCacheHitRate(metrics)

			require.Equal(t, tt.expectedRate, metrics.OverallCacheHitRate)
		})
	}
}

func TestRecordCacheHit(t *testing.T) {
	metrics := &PreValidationMetrics{}

	RecordCacheHit(metrics, TransactionType_TX_TYPE_IR_COMPLETION, 100)

	require.Equal(t, uint64(1), metrics.TotalCacheHits)
	require.Equal(t, 1.0, metrics.OverallCacheHitRate)

	// Check type-specific metrics
	typeMetrics := GetTypeMetrics(metrics, TransactionType_TX_TYPE_IR_COMPLETION)
	require.NotNil(t, typeMetrics)
	require.Equal(t, uint64(1), typeMetrics.CacheHits)
	require.Equal(t, uint64(1), typeMetrics.TotalExecuted)
	require.Equal(t, float64(100), typeMetrics.AvgTimeSavingsMs)

	// Record another hit
	RecordCacheHit(metrics, TransactionType_TX_TYPE_IR_COMPLETION, 200)

	require.Equal(t, uint64(2), metrics.TotalCacheHits)
	typeMetrics = GetTypeMetrics(metrics, TransactionType_TX_TYPE_IR_COMPLETION)
	require.Equal(t, uint64(2), typeMetrics.CacheHits)
	require.Equal(t, float64(150), typeMetrics.AvgTimeSavingsMs) // (100 + 200) / 2
}

func TestRecordCacheMiss(t *testing.T) {
	metrics := &PreValidationMetrics{}

	RecordCacheMiss(metrics, TransactionType_TX_TYPE_DEX_SWAP)

	require.Equal(t, uint64(1), metrics.TotalCacheMisses)
	require.Equal(t, 0.0, metrics.OverallCacheHitRate)

	// Check type-specific metrics
	typeMetrics := GetTypeMetrics(metrics, TransactionType_TX_TYPE_DEX_SWAP)
	require.NotNil(t, typeMetrics)
	require.Equal(t, uint64(1), typeMetrics.CacheMisses)
	require.Equal(t, 0.0, typeMetrics.CacheHitRate)
}

func TestGetTypeMetrics_NonExistent(t *testing.T) {
	metrics := &PreValidationMetrics{}

	typeMetrics := GetTypeMetrics(metrics, TransactionType_TX_TYPE_VC_MINT)
	require.Nil(t, typeMetrics)
}
