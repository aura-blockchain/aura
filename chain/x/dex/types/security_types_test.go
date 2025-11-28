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

	if params.MaxTradeSizePercent != math.LegacyNewDecWithPrec(20, 2).String() {
		t.Errorf("expected MaxTradeSizePercent to be %s, got %s",
			math.LegacyNewDecWithPrec(20, 2).String(), params.MaxTradeSizePercent)
	}

	if params.MaxPriceImpactPercent != math.LegacyNewDecWithPrec(10, 0).String() {
		t.Errorf("expected MaxPriceImpactPercent to be %s, got %s",
			math.LegacyNewDecWithPrec(10, 0).String(), params.MaxPriceImpactPercent)
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

	if params.MinPoolCreationLiquidity != math.NewInt(1000_000000).String() {
		t.Errorf("expected MinPoolCreationLiquidity to be %s, got %s",
			math.NewInt(1000_000000).String(), params.MinPoolCreationLiquidity)
	}

	if params.MinLiquidityBlocks != 5 {
		t.Errorf("expected MinLiquidityBlocks to be 5, got %d", params.MinLiquidityBlocks)
	}

	if params.WashTradeMinInterval != 60 {
		t.Errorf("expected WashTradeMinInterval to be 60, got %d", params.WashTradeMinInterval)
	}

	if params.MinTradeAmount != math.NewInt(1_000000).String() {
		t.Errorf("expected MinTradeAmount to be %s, got %s",
			math.NewInt(1_000000).String(), params.MinTradeAmount)
	}

	if params.MaxOrderVariance != math.LegacyNewDecWithPrec(50, 2).String() {
		t.Errorf("expected MaxOrderVariance to be %s, got %s",
			math.LegacyNewDecWithPrec(50, 2).String(), params.MaxOrderVariance)
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
		t.Error("DefaultSecurityParams should return consistent values")
	}

	if params1.MaxTradeSizePercent != params2.MaxTradeSizePercent {
		t.Error("DefaultSecurityParams should return consistent values")
	}

	if params1.CircuitBreakerEnabled != params2.CircuitBreakerEnabled {
		t.Error("DefaultSecurityParams should return consistent values")
	}

	if params1.MevProtectionEnabled != params2.MevProtectionEnabled {
		t.Error("DefaultSecurityParams should return consistent values")
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
