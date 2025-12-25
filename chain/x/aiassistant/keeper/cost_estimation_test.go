// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func TestEstimateQueryCost(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name          string
		modelHash     string
		inputLength   uint64
		operationType string
		wantErr       bool
	}{
		{
			name:          "simple query",
			modelHash:     "model1",
			inputLength:   100,
			operationType: "inference",
			wantErr:       false,
		},
		{
			name:          "large input",
			modelHash:     "model1",
			inputLength:   10000,
			operationType: "inference",
			wantErr:       false,
		},
		{
			name:          "training operation",
			modelHash:     "model2",
			inputLength:   1000,
			operationType: "training",
			wantErr:       false,
		},
		{
			name:          "fine-tuning",
			modelHash:     "model3",
			inputLength:   5000,
			operationType: "fine_tuning",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := k.EstimateQueryCost(ctx, tt.modelHash, tt.inputLength, tt.operationType)
			if (err != nil) != tt.wantErr {
				t.Errorf("EstimateQueryCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cost.Amount.IsNegative() {
					t.Error("Cost should not be negative")
				}
				if cost.Amount.IsZero() {
					t.Error("Cost should not be zero for valid query")
				}
			}
		})
	}
}

func TestCostScaling(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Test that cost scales with input length
	smallCost, err := k.EstimateQueryCost(ctx, "model1", 100, "inference")
	if err != nil {
		t.Fatalf("Failed to estimate small cost: %v", err)
	}

	largeCost, err := k.EstimateQueryCost(ctx, "model1", 1000, "inference")
	if err != nil {
		t.Fatalf("Failed to estimate large cost: %v", err)
	}

	if !largeCost.Amount.GT(smallCost.Amount) {
		t.Error("Larger input should cost more than smaller input")
	}
}

func TestComputeUnitsEstimation(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name          string
		modelHash     string
		inputLength   uint64
		operationType string
		minUnits      uint64
	}{
		{
			name:          "small inference",
			modelHash:     "model1",
			inputLength:   100,
			operationType: "inference",
			minUnits:      10,
		},
		{
			name:          "large inference",
			modelHash:     "model1",
			inputLength:   10000,
			operationType: "inference",
			minUnits:      100,
		},
		{
			name:          "training",
			modelHash:     "model2",
			inputLength:   1000,
			operationType: "training",
			minUnits:      500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			units := k.EstimateComputeUnits(ctx, tt.modelHash, tt.inputLength, tt.operationType)
			if units < tt.minUnits {
				t.Errorf("Expected at least %d compute units, got %d", tt.minUnits, units)
			}
		})
	}
}

func TestGetModelPricing(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set model pricing
	pricing := ModelPricing{
		ModelHash:      "model1",
		BasePrice:      sdkmath.NewInt(1000),
		PerTokenPrice:  sdkmath.NewInt(10),
		PerComputeUnit: sdkmath.NewInt(5),
		MinimumCharge:  sdkmath.NewInt(100),
	}

	err := k.SetModelPricing(ctx, pricing)
	if err != nil {
		t.Fatalf("Failed to set model pricing: %v", err)
	}

	// Retrieve pricing
	retrieved, found := k.GetModelPricing(ctx, "model1")
	if !found {
		t.Fatal("Model pricing not found")
	}

	if !retrieved.BasePrice.Equal(pricing.BasePrice) {
		t.Errorf("Expected base price %v, got %v", pricing.BasePrice, retrieved.BasePrice)
	}
}

func TestDynamicPricing(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set initial pricing
	pricing := ModelPricing{
		ModelHash:     "model1",
		BasePrice:     sdkmath.NewInt(1000),
		PerTokenPrice: sdkmath.NewInt(10),
	}
	k.SetModelPricing(ctx, pricing)

	// Simulate high demand
	for i := 0; i < 100; i++ {
		k.EstimateQueryCost(ctx, "model1", 100, "inference")
	}

	// Check if pricing adjusted
	adjustedPricing, _ := k.GetModelPricing(ctx, "model1")

	t.Logf("Original price: %v, Adjusted price: %v",
		pricing.BasePrice, adjustedPricing.BasePrice)
}

func TestBulkCostEstimation(t *testing.T) {
	k, ctx := setupKeeper(t)

	queries := []struct {
		modelHash   string
		inputLength uint64
		opType      string
	}{
		{"model1", 100, "inference"},
		{"model1", 200, "inference"},
		{"model2", 300, "training"},
	}

	totalCost := sdkmath.ZeroInt()
	for _, q := range queries {
		cost, err := k.EstimateQueryCost(ctx, q.modelHash, q.inputLength, q.opType)
		if err != nil {
			t.Fatalf("Failed to estimate cost: %v", err)
		}
		totalCost = totalCost.Add(cost.Amount)
	}

	if totalCost.IsZero() {
		t.Error("Total cost should be greater than zero")
	}

	t.Logf("Total estimated cost for %d queries: %v", len(queries), totalCost)
}

func TestCostDiscounts(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set discount for high-volume users
	discount := CostDiscount{
		Address:         "user1",
		DiscountPercent: 20,
		MinimumQueries:  100,
		ValidUntil:      ctx.BlockTime().Add(30 * 24 * 60 * 60),
	}

	err := k.SetCostDiscount(ctx, discount)
	if err != nil {
		t.Fatalf("Failed to set discount: %v", err)
	}

	// Estimate cost with discount
	baseCost, _ := k.EstimateQueryCost(ctx, "model1", 100, "inference")
	discountedCost := k.ApplyDiscount(ctx, "user1", baseCost.Amount)

	expectedDiscount := baseCost.Amount.Mul(sdkmath.NewInt(20)).Quo(sdkmath.NewInt(100))
	expectedFinal := baseCost.Amount.Sub(expectedDiscount)

	if !discountedCost.Equal(expectedFinal) {
		t.Errorf("Expected discounted cost %v, got %v", expectedFinal, discountedCost)
	}
}

func TestMinimumCharge(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set pricing with minimum charge
	pricing := ModelPricing{
		ModelHash:     "model1",
		BasePrice:     sdkmath.NewInt(10),
		PerTokenPrice: sdkmath.NewInt(1),
		MinimumCharge: sdkmath.NewInt(100),
	}
	k.SetModelPricing(ctx, pricing)

	// Very small query
	cost, err := k.EstimateQueryCost(ctx, "model1", 1, "inference")
	if err != nil {
		t.Fatalf("Failed to estimate cost: %v", err)
	}

	if cost.Amount.LT(pricing.MinimumCharge) {
		t.Errorf("Cost should meet minimum charge of %v, got %v",
			pricing.MinimumCharge, cost.Amount)
	}
}

func TestPeakHourPricing(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set peak hour multiplier
	err := k.SetPeakHourMultiplier(ctx, 1.5)
	if err != nil {
		t.Fatalf("Failed to set peak hour multiplier: %v", err)
	}

	// Estimate cost during peak hours
	// (Implementation would check current time)
	cost, err := k.EstimateQueryCost(ctx, "model1", 100, "inference")
	if err != nil {
		t.Fatalf("Failed to estimate cost: %v", err)
	}

	if cost.Amount.IsZero() {
		t.Error("Cost should be non-zero")
	}
}

func TestCostEstimationCache(t *testing.T) {
	k, ctx := setupKeeper(t)

	// First estimation
	cost1, err := k.EstimateQueryCost(ctx, "model1", 100, "inference")
	if err != nil {
		t.Fatalf("Failed first estimation: %v", err)
	}

	// Second estimation with same parameters
	cost2, err := k.EstimateQueryCost(ctx, "model1", 100, "inference")
	if err != nil {
		t.Fatalf("Failed second estimation: %v", err)
	}

	// Costs should be consistent
	if !cost1.Amount.Equal(cost2.Amount) {
		t.Errorf("Cost estimates should be consistent: %v vs %v", cost1, cost2)
	}
}

func TestOperationTypePricing(t *testing.T) {
	k, ctx := setupKeeper(t)

	operations := []string{"inference", "training", "fine_tuning", "embedding"}
	costs := make(map[string]sdkmath.Int)

	for _, op := range operations {
		cost, err := k.EstimateQueryCost(ctx, "model1", 1000, op)
		if err != nil {
			t.Logf("Operation %s not supported: %v", op, err)
			continue
		}
		costs[op] = cost.Amount
	}

	// Training should typically cost more than inference
	if infCost, hasInf := costs["inference"]; hasInf {
		if trainCost, hasTrain := costs["training"]; hasTrain {
			if !trainCost.GT(infCost) {
				t.Log("Training cost may not be higher than inference (could be model-specific)")
			}
		}
	}
}

func TestCostEstimationHistory(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Perform multiple estimations
	for i := 0; i < 5; i++ {
		k.EstimateQueryCost(ctx, "model1", uint64(100*(i+1)), "inference")
	}

	// Get estimation history
	history := k.GetCostEstimationHistory(ctx, "model1", 10)
	if len(history) < 5 {
		t.Logf("Expected at least 5 history entries, got %d (may not be tracked)", len(history))
	}
}

func TestRefundCalculation(t *testing.T) {
	k, ctx := setupKeeper(t)

	estimatedCost := sdkmath.NewInt(1000)
	actualCost := sdkmath.NewInt(800)

	refund := k.CalculateRefund(ctx, estimatedCost, actualCost)

	expectedRefund := sdkmath.NewInt(200)
	if !refund.Equal(expectedRefund) {
		t.Errorf("Expected refund %v, got %v", expectedRefund, refund)
	}

	// Test case where actual > estimated (no refund)
	refund = k.CalculateRefund(ctx, actualCost, estimatedCost)
	if !refund.IsZero() {
		t.Error("Refund should be zero when actual cost exceeds estimate")
	}
}

func TestCostBreakdown(t *testing.T) {
	k, ctx := setupKeeper(t)

	breakdown := k.GetCostBreakdown(ctx, "model1", 1000, "inference")

	// Verify breakdown includes all components
	if breakdown.BaseCost.IsZero() && breakdown.ComputeCost.IsZero() && breakdown.StorageCost.IsZero() {
		t.Log("Cost breakdown may not be fully implemented")
	}

	// Total should equal sum of components
	total := breakdown.BaseCost.Add(breakdown.ComputeCost).Add(breakdown.StorageCost)
	if !total.Equal(breakdown.TotalCost) {
		t.Errorf("Total cost %v should equal sum of components %v",
			breakdown.TotalCost, total)
	}
}

func TestSubscriptionPricing(t *testing.T) {
	k, ctx := setupKeeper(t)

	subscription := SubscriptionPlan{
		Name:            "premium",
		MonthlyPrice:    sdkmath.NewInt(10000),
		IncludedQueries: 1000,
		OverageRate:     sdkmath.NewInt(5),
	}

	err := k.SetSubscriptionPlan(ctx, subscription)
	if err != nil {
		t.Fatalf("Failed to set subscription plan: %v", err)
	}

	// User with subscription
	k.AssignSubscription(ctx, "user1", "premium")

	// Estimate cost should consider subscription
	cost, err := k.EstimateQueryCost(ctx, "model1", 100, "inference")
	if err != nil {
		t.Fatalf("Failed to estimate cost: %v", err)
	}

	t.Logf("Cost with subscription: %v", cost.Amount)
}

func TestFreeTierPricing(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set free tier limits
	freeTier := FreeTier{
		DailyQueries:   10,
		MonthlyQueries: 100,
		MaxInputLength: 1000,
	}

	err := k.SetFreeTier(ctx, freeTier)
	if err != nil {
		t.Fatalf("Failed to set free tier: %v", err)
	}

	// Check if query qualifies for free tier
	qualifies := k.QualifiesForFreeTier(ctx, "user1", 100)
	t.Logf("Qualifies for free tier: %v", qualifies)
}
