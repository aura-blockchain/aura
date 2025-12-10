package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// TestIntSqrt tests the deterministic integer square root function
func TestIntSqrt(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected uint64
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"perfect square", 16, 4},
		{"perfect square large", 10000, 100},
		{"non-perfect square", 17, 4}, // floor(sqrt(17)) = 4
		{"non-perfect square", 99, 9}, // floor(sqrt(99)) = 9
		{"large number", 1000000, 1000},
		{"very large", 18446744073709551615, 4294967295}, // sqrt(2^64-1) = 2^32-1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := intSqrt(tt.input)
			assert.Equal(t, tt.expected, result, "intSqrt(%d) should equal %d", tt.input, tt.expected)
		})
	}
}

// TestIntSqrtDeterminism verifies that intSqrt produces consistent results
func TestIntSqrtDeterminism(t *testing.T) {
	testValues := []uint64{0, 1, 2, 15, 16, 17, 100, 1000, 10000, 999999}

	for _, val := range testValues {
		// Run multiple times to ensure consistency
		first := intSqrt(val)
		for i := 0; i < 100; i++ {
			result := intSqrt(val)
			require.Equal(t, first, result, "intSqrt(%d) must be deterministic", val)
		}
	}
}

// TestCalculateGasPriceStatsDeterminism verifies volatility calculation is deterministic
func TestCalculateGasPriceStatsDeterminism(t *testing.T) {
	keeper, _ := setupTestKeeper(t)

	// Create tracking data with known prices
	tracking := &types.GasPriceTracking{
		PriceHistory: []types.GasPricePoint{
			{Price: 100},
			{Price: 110},
			{Price: 105},
			{Price: 108},
			{Price: 112},
			{Price: 98},
			{Price: 102},
			{Price: 115},
			{Price: 107},
			{Price: 103},
		},
	}

	// Calculate stats multiple times
	keeper.calculateGasPriceStats(tracking)
	firstVolatility := tracking.VolatilityScore
	firstAvg := tracking.AveragePrice
	firstMedian := tracking.MedianPrice
	firstMin := tracking.MinPrice
	firstMax := tracking.MaxPrice
	firstTrend := tracking.TrendDirection

	// Verify determinism by recalculating
	for i := 0; i < 100; i++ {
		keeper.calculateGasPriceStats(tracking)
		assert.Equal(t, firstVolatility, tracking.VolatilityScore, "VolatilityScore must be deterministic")
		assert.Equal(t, firstAvg, tracking.AveragePrice, "AveragePrice must be deterministic")
		assert.Equal(t, firstMedian, tracking.MedianPrice, "MedianPrice must be deterministic")
		assert.Equal(t, firstMin, tracking.MinPrice, "MinPrice must be deterministic")
		assert.Equal(t, firstMax, tracking.MaxPrice, "MaxPrice must be deterministic")
		assert.Equal(t, firstTrend, tracking.TrendDirection, "TrendDirection must be deterministic")
	}

	// Verify reasonable values
	assert.Equal(t, uint64(106), firstAvg, "Average should be 1060/10 = 106")
	assert.Equal(t, uint64(106), firstMedian, "Median should be (105+107)/2 = 106")
	assert.Equal(t, uint64(98), firstMin, "Min should be 98")
	assert.Equal(t, uint64(115), firstMax, "Max should be 115")
	assert.Greater(t, firstVolatility, uint64(0), "Volatility should be > 0")
	assert.Less(t, firstVolatility, uint64(10000), "Volatility should be < 100% (10000 basis points)")
}

// TestCalculateGasPriceStatsZeroAverage tests edge case with zero average
func TestCalculateGasPriceStatsZeroAverage(t *testing.T) {
	keeper, _ := setupTestKeeper(t)

	tracking := &types.GasPriceTracking{
		PriceHistory: []types.GasPricePoint{},
	}

	keeper.calculateGasPriceStats(tracking)
	assert.Equal(t, uint64(0), tracking.VolatilityScore, "Volatility should be 0 for empty history")
}

// TestCalculateGasPriceTrendDeterminism verifies trend calculation is deterministic
func TestCalculateGasPriceTrendDeterminism(t *testing.T) {
	keeper, _ := setupTestKeeper(t)

	testCases := []struct {
		name     string
		prices   []uint64
		expected string
	}{
		{
			name:     "increasing trend",
			prices:   []uint64{100, 102, 105, 108, 110, 112, 115, 118, 120, 122, 125, 128, 130, 135, 140},
			expected: "increasing",
		},
		{
			name:     "decreasing trend",
			prices:   []uint64{140, 135, 130, 128, 125, 122, 120, 118, 115, 112, 110, 108, 105, 102, 100},
			expected: "decreasing",
		},
		{
			name:     "stable trend",
			prices:   []uint64{100, 101, 100, 102, 101, 100, 101, 102, 101, 100, 101, 100, 102, 101, 100},
			expected: "stable",
		},
		{
			name:     "insufficient history",
			prices:   []uint64{100, 110, 105},
			expected: "stable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			history := make([]types.GasPricePoint, len(tc.prices))
			for i, price := range tc.prices {
				history[i] = types.GasPricePoint{Price: price}
			}

			// Calculate trend multiple times to verify determinism
			firstResult := keeper.calculateGasPriceTrend(history)
			for i := 0; i < 100; i++ {
				result := keeper.calculateGasPriceTrend(history)
				require.Equal(t, firstResult, result, "Trend calculation must be deterministic")
			}

			assert.Equal(t, tc.expected, firstResult, "Trend should match expected direction")
		})
	}
}

// TestVolatilityScoreBasisPoints verifies the volatility score is in basis points
func TestVolatilityScoreBasisPoints(t *testing.T) {
	keeper, _ := setupTestKeeper(t)

	// Test case: prices with known standard deviation
	// Prices: 100, 100, 100, 100, 100 -> stdDev = 0, CV = 0
	tracking := &types.GasPriceTracking{
		PriceHistory: []types.GasPricePoint{
			{Price: 100},
			{Price: 100},
			{Price: 100},
			{Price: 100},
			{Price: 100},
		},
	}

	keeper.calculateGasPriceStats(tracking)
	assert.Equal(t, uint64(0), tracking.VolatilityScore, "Zero variance should result in zero volatility")

	// Test case: prices with larger variance
	// Average = 100, with prices varying ±20
	tracking2 := &types.GasPriceTracking{
		PriceHistory: []types.GasPricePoint{
			{Price: 80},
			{Price: 90},
			{Price: 100},
			{Price: 110},
			{Price: 120},
		},
	}

	keeper.calculateGasPriceStats(tracking2)
	assert.Greater(t, tracking2.VolatilityScore, uint64(1000), "Should have significant volatility (>10%)")
	assert.Less(t, tracking2.VolatilityScore, uint64(2000), "Should have reasonable volatility (<20%)")
}

// TestGasPriceTrackerOverflow tests overflow protection
func TestGasPriceTrackerOverflow(t *testing.T) {
	keeper, _ := setupTestKeeper(t)

	// Test with very large prices to ensure no overflow
	tracking := &types.GasPriceTracking{
		PriceHistory: []types.GasPricePoint{
			{Price: 1000000000000}, // 1 trillion
			{Price: 1100000000000},
			{Price: 1050000000000},
			{Price: 1080000000000},
			{Price: 1120000000000},
		},
	}

	// Should not panic
	keeper.calculateGasPriceStats(tracking)

	// Verify calculations completed
	assert.Greater(t, tracking.AveragePrice, uint64(0))
	assert.Greater(t, tracking.MaxPrice, tracking.MinPrice)
}
