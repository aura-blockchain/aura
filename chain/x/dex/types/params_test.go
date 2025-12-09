package types

import (
	"testing"

	"cosmossdk.io/math"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	// Validate default values
	expectedTradingFee := math.LegacyMustNewDecFromStr("0.003")
	if !params.TradingFee.Equal(expectedTradingFee) {
		t.Errorf("expected TradingFee to be 0.003, got %s", params.TradingFee)
	}

	expectedProtocolFee := math.LegacyMustNewDecFromStr("0.0005")
	if !params.ProtocolFee.Equal(expectedProtocolFee) {
		t.Errorf("expected ProtocolFee to be 0.0005, got %s", params.ProtocolFee)
	}

	if params.MaxSlippageBps != 10000 {
		t.Errorf("expected MaxSlippageBps to be 10000, got %d", params.MaxSlippageBps)
	}

	expectedMinSwap := math.NewInt(1000000)
	if !params.MinSwapAmount.Equal(expectedMinSwap) {
		t.Errorf("expected MinSwapAmount to be 1000000, got %s", params.MinSwapAmount)
	}

	if !params.IrBoostEnabled {
		t.Error("expected IrBoostEnabled to be true")
	}

	if params.IrBoostPercent != 40 {
		t.Errorf("expected IrBoostPercent to be 40, got %d", params.IrBoostPercent)
	}

	if !params.BondingCurveEnabled {
		t.Error("expected BondingCurveEnabled to be true")
	}

	if !params.BuybackBurnEnabled {
		t.Error("expected BuybackBurnEnabled to be true")
	}

	if params.BuybackPercent != 100 {
		t.Errorf("expected BuybackPercent to be 100, got %d", params.BuybackPercent)
	}

	if params.MinLiquidityTiers == nil {
		t.Error("expected MinLiquidityTiers to be initialized")
	}
}

func TestDefaultGenesis(t *testing.T) {
	genesis := DefaultGenesis()

	if genesis == nil {
		t.Fatal("DefaultGenesis should not return nil")
	}

	// Validate default genesis state (Params is a value type, not pointer)
	// Just check it's initialized with valid values
	if genesis.Params.TradingFee.IsNil() {
		t.Fatal("expected Params.TradingFee to be set")
	}

	if genesis.LiquidityPools == nil {
		t.Error("expected LiquidityPools to be initialized")
	}

	if len(genesis.LiquidityPools) != 0 {
		t.Errorf("expected empty LiquidityPools, got length %d", len(genesis.LiquidityPools))
	}

	if genesis.SwapOrders == nil {
		t.Error("expected SwapOrders to be initialized")
	}

	if len(genesis.SwapOrders) != 0 {
		t.Errorf("expected empty SwapOrders, got length %d", len(genesis.SwapOrders))
	}

	if genesis.Orderbooks == nil {
		t.Error("expected Orderbooks to be initialized")
	}

	if len(genesis.Orderbooks) != 0 {
		t.Errorf("expected empty Orderbooks, got length %d", len(genesis.Orderbooks))
	}

	if genesis.MarketPrices == nil {
		t.Error("expected MarketPrices to be initialized")
	}

	if len(genesis.MarketPrices) != 0 {
		t.Errorf("expected empty MarketPrices, got length %d", len(genesis.MarketPrices))
	}

	if genesis.SwapStats == nil {
		t.Error("expected SwapStats to be initialized")
	}

	if len(genesis.SwapStats) != 0 {
		t.Errorf("expected empty SwapStats, got length %d", len(genesis.SwapStats))
	}
}

func TestParamsConsistency(t *testing.T) {
	params1 := DefaultParams()
	params2 := DefaultParams()

	// Verify that multiple calls return consistent values
	if params1.TradingFee != params2.TradingFee {
		t.Error("DefaultParams should return consistent values")
	}

	if params1.ProtocolFee != params2.ProtocolFee {
		t.Error("DefaultParams should return consistent values")
	}

	if params1.MaxSlippageBps != params2.MaxSlippageBps {
		t.Error("DefaultParams should return consistent values")
	}
}
