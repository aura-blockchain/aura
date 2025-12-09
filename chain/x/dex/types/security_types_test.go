package types

import (
	"testing"

	"cosmossdk.io/math"
)

func TestDefaultSecurityParams(t *testing.T) {
	params := DefaultSecurityParams()

	if params == nil {
		t.Fatal("DefaultSecurityParams should not return nil")
	}

	// Validate default security values
	if params.MinBlockDelay != 2 {
		t.Errorf("expected MinBlockDelay to be 2, got %d", params.MinBlockDelay)
	}

	expectedMaxTradeSize := math.LegacyNewDecWithPrec(20, 2)
	if !params.MaxTradeSizePercent.Equal(expectedMaxTradeSize) {
		t.Errorf("expected MaxTradeSizePercent to be %s, got %s",
			expectedMaxTradeSize, params.MaxTradeSizePercent)
	}

	expectedMaxPriceImpact := math.LegacyNewDecWithPrec(10, 0)
	if !params.MaxPriceImpactPercent.Equal(expectedMaxPriceImpact) {
		t.Errorf("expected MaxPriceImpactPercent to be %s, got %s",
			expectedMaxPriceImpact, params.MaxPriceImpactPercent)
	}

	if params.LiquidityLockupSeconds != 86400 {
		t.Errorf("expected LiquidityLockupSeconds to be 86400, got %d", params.LiquidityLockupSeconds)
	}

	if params.PoolCreationCooldown != 3600 {
		t.Errorf("expected PoolCreationCooldown to be 3600, got %d", params.PoolCreationCooldown)
	}

	if params.MaxPoolsPerCreator != 10 {
		t.Errorf("expected MaxPoolsPerCreator to be 10, got %d", params.MaxPoolsPerCreator)
	}

	if params.TwapWindowBlocks != 100 {
		t.Errorf("expected TwapWindowBlocks to be 100, got %d", params.TwapWindowBlocks)
	}

	expectedMinPoolLiq := math.NewInt(1000_000000)
	if !params.MinPoolCreationLiquidity.Equal(expectedMinPoolLiq) {
		t.Errorf("expected MinPoolCreationLiquidity to be %s, got %s",
			expectedMinPoolLiq, params.MinPoolCreationLiquidity)
	}

	if params.MinLiquidityBlocks != 5 {
		t.Errorf("expected MinLiquidityBlocks to be 5, got %d", params.MinLiquidityBlocks)
	}

	if params.WashTradeMinInterval != 60 {
		t.Errorf("expected WashTradeMinInterval to be 60, got %d", params.WashTradeMinInterval)
	}

	expectedMinTrade := math.NewInt(1_000000)
	if !params.MinTradeAmount.Equal(expectedMinTrade) {
		t.Errorf("expected MinTradeAmount to be %s, got %s",
			expectedMinTrade, params.MinTradeAmount)
	}

	expectedMaxVariance := math.LegacyNewDecWithPrec(50, 2)
	if !params.MaxOrderVariance.Equal(expectedMaxVariance) {
		t.Errorf("expected MaxOrderVariance to be %s, got %s",
			expectedMaxVariance, params.MaxOrderVariance)
	}

	if !params.CircuitBreakerEnabled {
		t.Error("expected CircuitBreakerEnabled to be true")
	}

	if !params.MevProtectionEnabled {
		t.Error("expected MevProtectionEnabled to be true")
	}

	if params.MaxSwapsPerBlock != 5 {
		t.Errorf("expected MaxSwapsPerBlock to be 5, got %d", params.MaxSwapsPerBlock)
	}
}

func TestSecurityParamsConsistency(t *testing.T) {
	params1 := DefaultSecurityParams()
	params2 := DefaultSecurityParams()

	// Verify that multiple calls return consistent values
	if params1.MinBlockDelay != params2.MinBlockDelay {
		t.Errorf("DefaultSecurityParams should return consistent MinBlockDelay values, got %d and %d",
			params1.MinBlockDelay, params2.MinBlockDelay)
	}

	if !params1.MaxTradeSizePercent.Equal(params2.MaxTradeSizePercent) {
		t.Errorf("DefaultSecurityParams should return consistent MaxTradeSizePercent values, got %s and %s",
			params1.MaxTradeSizePercent, params2.MaxTradeSizePercent)
	}

	if !params1.MaxPriceImpactPercent.Equal(params2.MaxPriceImpactPercent) {
		t.Errorf("DefaultSecurityParams should return consistent MaxPriceImpactPercent values, got %s and %s",
			params1.MaxPriceImpactPercent, params2.MaxPriceImpactPercent)
	}

	if !params1.MinPoolCreationLiquidity.Equal(params2.MinPoolCreationLiquidity) {
		t.Errorf("DefaultSecurityParams should return consistent MinPoolCreationLiquidity values, got %s and %s",
			params1.MinPoolCreationLiquidity, params2.MinPoolCreationLiquidity)
	}

	if !params1.MinTradeAmount.Equal(params2.MinTradeAmount) {
		t.Errorf("DefaultSecurityParams should return consistent MinTradeAmount values, got %s and %s",
			params1.MinTradeAmount, params2.MinTradeAmount)
	}

	if !params1.MaxOrderVariance.Equal(params2.MaxOrderVariance) {
		t.Errorf("DefaultSecurityParams should return consistent MaxOrderVariance values, got %s and %s",
			params1.MaxOrderVariance, params2.MaxOrderVariance)
	}

	if params1.CircuitBreakerEnabled != params2.CircuitBreakerEnabled {
		t.Errorf("DefaultSecurityParams should return consistent CircuitBreakerEnabled values, got %v and %v",
			params1.CircuitBreakerEnabled, params2.CircuitBreakerEnabled)
	}

	if params1.MevProtectionEnabled != params2.MevProtectionEnabled {
		t.Errorf("DefaultSecurityParams should return consistent MevProtectionEnabled values, got %v and %v",
			params1.MevProtectionEnabled, params2.MevProtectionEnabled)
	}
}

func TestSecurityParamsTimeouts(t *testing.T) {
	params := DefaultSecurityParams()

	// Test that timeout values are reasonable
	if params.LiquidityLockupSeconds < 0 {
		t.Error("LiquidityLockupSeconds should not be negative")
	}

	if params.PoolCreationCooldown < 0 {
		t.Error("PoolCreationCooldown should not be negative")
	}

	if params.WashTradeMinInterval < 0 {
		t.Error("WashTradeMinInterval should not be negative")
	}

	// Test that 24 hours = 86400 seconds
	expectedDaySeconds := int64(24 * 60 * 60)
	if params.LiquidityLockupSeconds != expectedDaySeconds {
		t.Errorf("expected LiquidityLockupSeconds to be %d (24 hours), got %d",
			expectedDaySeconds, params.LiquidityLockupSeconds)
	}

	// Test that 1 hour = 3600 seconds
	expectedHourSeconds := int64(60 * 60)
	if params.PoolCreationCooldown != expectedHourSeconds {
		t.Errorf("expected PoolCreationCooldown to be %d (1 hour), got %d",
			expectedHourSeconds, params.PoolCreationCooldown)
	}
}

func TestSecurityParamsLimits(t *testing.T) {
	params := DefaultSecurityParams()

	// Test that limits are within reasonable ranges
	if params.MaxPoolsPerCreator <= 0 {
		t.Error("MaxPoolsPerCreator should be positive")
	}

	if params.MaxPoolsPerCreator > 100 {
		t.Error("MaxPoolsPerCreator seems too high")
	}

	if params.MinBlockDelay <= 0 {
		t.Error("MinBlockDelay should be positive")
	}

	if params.TwapWindowBlocks <= 0 {
		t.Error("TwapWindowBlocks should be positive")
	}

	if params.MinLiquidityBlocks <= 0 {
		t.Error("MinLiquidityBlocks should be positive")
	}

	if params.MaxSwapsPerBlock <= 0 {
		t.Error("MaxSwapsPerBlock should be positive")
	}
}
