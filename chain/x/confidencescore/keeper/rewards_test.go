// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestCalculatePoIReward(t *testing.T) {
	_, k := setupConfKeeperWithTime(t)

	tests := []struct {
		name           string
		auraPrice      math.LegacyDec
		expectedReward math.Int
	}{
		{
			name:           "Tier 1: Price $0.05",
			auraPrice:      math.LegacyNewDecWithPrec(5, 2), // $0.05
			expectedReward: math.NewInt(500_000_000),        // 500 AURA
		},
		{
			name:           "Tier 1: Price $0.10",
			auraPrice:      math.LegacyNewDecWithPrec(10, 2), // $0.10
			expectedReward: math.NewInt(500_000_000),         // 500 AURA
		},
		{
			name:           "Tier 2: Price $0.15",
			auraPrice:      math.LegacyNewDecWithPrec(15, 2), // $0.15
			expectedReward: math.NewInt(250_000_000),         // 250 AURA
		},
		{
			name:           "Tier 2: Price $0.25",
			auraPrice:      math.LegacyNewDecWithPrec(25, 2), // $0.25
			expectedReward: math.NewInt(250_000_000),         // 250 AURA
		},
		{
			name:           "Tier 3: Price $0.35",
			auraPrice:      math.LegacyNewDecWithPrec(35, 2), // $0.35
			expectedReward: math.NewInt(100_000_000),         // 100 AURA
		},
		{
			name:           "Tier 3: Price $0.45",
			auraPrice:      math.LegacyNewDecWithPrec(45, 2), // $0.45
			expectedReward: math.NewInt(100_000_000),         // 100 AURA
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward := k.CalculatePoIReward(tt.auraPrice)
			if !reward.Equal(tt.expectedReward) {
				t.Errorf("expected reward %s, got %s", tt.expectedReward, reward)
			}
		})
	}
}

func TestCalculatePoIReward_Tier4(t *testing.T) {
	_, k := setupConfKeeperWithTime(t)

	// Tier 4: Price >= $0.50, reward = $50 / price
	tests := []struct {
		name      string
		auraPrice math.LegacyDec
		maxReward math.Int // Approximate max (should not exceed this)
	}{
		{
			name:      "Price $0.50",
			auraPrice: math.LegacyNewDecWithPrec(50, 2), // $0.50
			maxReward: math.NewInt(100_000_000),         // 100 AURA = $50 / $0.50
		},
		{
			name:      "Price $1.00",
			auraPrice: math.LegacyNewDec(1),    // $1.00
			maxReward: math.NewInt(50_000_000), // 50 AURA = $50 / $1.00
		},
		{
			name:      "Price $2.00",
			auraPrice: math.LegacyNewDec(2),    // $2.00
			maxReward: math.NewInt(25_000_000), // 25 AURA = $50 / $2.00
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward := k.CalculatePoIReward(tt.auraPrice)

			// For tier 4, check that it's approximately correct
			// Allow small variance due to decimal math
			diff := reward.Sub(tt.maxReward).Abs()
			tolerance := math.NewInt(1000) // Allow 1000 uaura tolerance

			if diff.GT(tolerance) {
				t.Errorf("expected reward close to %s, got %s (diff: %s)",
					tt.maxReward, reward, diff)
			}
		})
	}
}

func TestSplitPoIReward(t *testing.T) {
	_, k := setupConfKeeperWithTime(t)

	tests := []struct {
		name         string
		totalReward  math.Int
		userPercent  uint64
		expectedUser math.Int
		expectedNode math.Int
	}{
		{
			name:         "50/50 split",
			totalReward:  math.NewInt(1000),
			userPercent:  50,
			expectedUser: math.NewInt(500),
			expectedNode: math.NewInt(500),
		},
		{
			name:         "70/30 split",
			totalReward:  math.NewInt(1000),
			userPercent:  70,
			expectedUser: math.NewInt(700),
			expectedNode: math.NewInt(300),
		},
		{
			name:         "100/0 split",
			totalReward:  math.NewInt(1000),
			userPercent:  100,
			expectedUser: math.NewInt(1000),
			expectedNode: math.NewInt(0),
		},
		{
			name:         "0/100 split",
			totalReward:  math.NewInt(1000),
			userPercent:  0,
			expectedUser: math.NewInt(0),
			expectedNode: math.NewInt(1000),
		},
		{
			name:         "Invalid percentage > 100 (defaults to 50/50)",
			totalReward:  math.NewInt(1000),
			userPercent:  150,
			expectedUser: math.NewInt(500),
			expectedNode: math.NewInt(500),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userReward, nodeReward := k.SplitPoIReward(tt.totalReward, tt.userPercent)

			if !userReward.Equal(tt.expectedUser) {
				t.Errorf("expected user reward %s, got %s", tt.expectedUser, userReward)
			}

			if !nodeReward.Equal(tt.expectedNode) {
				t.Errorf("expected node reward %s, got %s", tt.expectedNode, nodeReward)
			}

			// Verify rewards sum to total
			sum := userReward.Add(nodeReward)
			if !sum.Equal(tt.totalReward) {
				t.Errorf("rewards don't sum to total: %s + %s != %s",
					userReward, nodeReward, tt.totalReward)
			}
		})
	}
}

func TestCalculateVBTBoost(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)

	tests := []struct {
		name               string
		completionTime     int64
		expectedMultiplier math.LegacyDec
	}{
		{
			name:               "Completed in 25% of expected time (1800s / 3600s)",
			completionTime:     1800,
			expectedMultiplier: math.LegacyNewDecWithPrec(20, 1), // 2.0x
		},
		{
			name:               "Completed in 60% of expected time (2200s / 3600s)",
			completionTime:     2200,
			expectedMultiplier: math.LegacyNewDecWithPrec(15, 1), // 1.5x
		},
		{
			name:               "Completed in 90% of expected time (3200s / 3600s)",
			completionTime:     3200,
			expectedMultiplier: math.LegacyNewDecWithPrec(125, 2), // 1.25x
		},
		{
			name:               "Completed in 100% of expected time",
			completionTime:     3600,
			expectedMultiplier: math.LegacyNewDecWithPrec(125, 2), // 1.25x
		},
		{
			name:               "Completed slower than expected (4000s / 3600s)",
			completionTime:     4000,
			expectedMultiplier: math.LegacyOneDec(), // 1.0x
		},
		{
			name:               "Invalid time (0 or negative)",
			completionTime:     0,
			expectedMultiplier: math.LegacyOneDec(), // 1.0x
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multiplier := k.CalculateVBTBoost(ctx, tt.completionTime, "IR-001")

			if !multiplier.Equal(tt.expectedMultiplier) {
				t.Errorf("expected multiplier %s, got %s",
					tt.expectedMultiplier, multiplier)
			}
		})
	}
}

func TestCalculateVBTBoost_Disabled(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)

	// Get params and disable velocity bonus
	params, _ := k.GetParams(ctx)
	params.VelocityBonusEnabled = false
	require.NoError(t, k.SetParams(params))

	multiplier := k.CalculateVBTBoost(ctx, 1800, "IR-001")

	if !multiplier.Equal(math.LegacyOneDec()) {
		t.Errorf("expected 1.0x when disabled, got %s", multiplier)
	}
}

func TestApplyVBTBoost(t *testing.T) {
	_, k := setupConfKeeperWithTime(t)

	tests := []struct {
		name           string
		baseReward     math.Int
		multiplier     math.LegacyDec
		expectedReward math.Int
	}{
		{
			name:           "2.0x boost",
			baseReward:     math.NewInt(100_000_000),
			multiplier:     math.LegacyNewDecWithPrec(20, 1),
			expectedReward: math.NewInt(200_000_000),
		},
		{
			name:           "1.5x boost",
			baseReward:     math.NewInt(100_000_000),
			multiplier:     math.LegacyNewDecWithPrec(15, 1),
			expectedReward: math.NewInt(150_000_000),
		},
		{
			name:           "1.25x boost",
			baseReward:     math.NewInt(100_000_000),
			multiplier:     math.LegacyNewDecWithPrec(125, 2),
			expectedReward: math.NewInt(125_000_000),
		},
		{
			name:           "No boost (1.0x)",
			baseReward:     math.NewInt(100_000_000),
			multiplier:     math.LegacyOneDec(),
			expectedReward: math.NewInt(100_000_000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boostedReward := k.ApplyVBTBoost(tt.baseReward, tt.multiplier)

			if !boostedReward.Equal(tt.expectedReward) {
				t.Errorf("expected reward %s, got %s",
					tt.expectedReward, boostedReward)
			}
		})
	}
}

func TestGetCurrentRewardAmount(t *testing.T) {
	_, k := setupConfKeeperWithTime(t)

	auraPrice := math.LegacyNewDecWithPrec(15, 2) // $0.15
	reward := k.GetCurrentRewardAmount(auraPrice)

	expected := math.NewInt(250_000_000) // Tier 2: 250 AURA
	if !reward.Equal(expected) {
		t.Errorf("expected reward %s, got %s", expected, reward)
	}
}

func TestGetRewardTierInfo(t *testing.T) {
	_, k := setupConfKeeperWithTime(t)

	tests := []struct {
		name           string
		auraPrice      math.LegacyDec
		expectedTier   string
		expectedReward math.Int
	}{
		{
			name:           "Bootstrap tier",
			auraPrice:      math.LegacyNewDecWithPrec(5, 2),
			expectedTier:   "Bootstrap (<$0.11)",
			expectedReward: math.NewInt(500_000_000),
		},
		{
			name:           "Early Growth tier",
			auraPrice:      math.LegacyNewDecWithPrec(20, 2),
			expectedTier:   "Early Growth ($0.11-$0.30)",
			expectedReward: math.NewInt(250_000_000),
		},
		{
			name:           "Growth tier",
			auraPrice:      math.LegacyNewDecWithPrec(40, 2),
			expectedTier:   "Growth ($0.30-$0.50)",
			expectedReward: math.NewInt(100_000_000),
		},
		{
			name:         "Established tier",
			auraPrice:    math.LegacyNewDec(1),
			expectedTier: "Established (≥$0.50)",
			// expectedReward is dynamic, don't check exact value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tierName, rewardAmount := k.GetRewardTierInfo(tt.auraPrice)

			if tierName != tt.expectedTier {
				t.Errorf("expected tier %s, got %s", tt.expectedTier, tierName)
			}

			// Only check reward for non-Established tiers
			if tt.expectedTier != "Established (≥$0.50)" {
				if !rewardAmount.Equal(tt.expectedReward) {
					t.Errorf("expected reward %s, got %s",
						tt.expectedReward, rewardAmount)
				}
			} else {
				// For Established tier, just verify it's positive
				if rewardAmount.LTE(math.ZeroInt()) {
					t.Error("expected positive reward for Established tier")
				}
			}
		})
	}
}

func TestGetTotalRewardsDistributed(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100)

	walletAddr := "aura1test"

	// Currently returns 0 as placeholder (rewards not tracked in proto yet)
	total := k.GetTotalRewardsDistributed(ctx, walletAddr)

	if total != 0 {
		t.Errorf("expected 0 (placeholder), got %d", total)
	}
}

func TestGetRewardTiers(t *testing.T) {
	tiers := GetRewardTiers()

	if len(tiers) != 4 {
		t.Fatalf("expected 4 tiers, got %d", len(tiers))
	}

	// Verify tier 1
	if !tiers[0].MaxPrice.Equal(math.LegacyNewDecWithPrec(11, 2)) {
		t.Error("tier 1 max price should be $0.11")
	}
	if !tiers[0].RewardAmount.Equal(math.NewInt(500_000_000)) {
		t.Error("tier 1 reward should be 500 AURA")
	}
	if tiers[0].UseUSDCap {
		t.Error("tier 1 should not use USD cap")
	}

	// Verify tier 2
	if !tiers[1].MaxPrice.Equal(math.LegacyNewDecWithPrec(30, 2)) {
		t.Error("tier 2 max price should be $0.30")
	}
	if !tiers[1].RewardAmount.Equal(math.NewInt(250_000_000)) {
		t.Error("tier 2 reward should be 250 AURA")
	}

	// Verify tier 3
	if !tiers[2].MaxPrice.Equal(math.LegacyNewDecWithPrec(50, 2)) {
		t.Error("tier 3 max price should be $0.50")
	}
	if !tiers[2].RewardAmount.Equal(math.NewInt(100_000_000)) {
		t.Error("tier 3 reward should be 100 AURA")
	}

	// Verify tier 4 (unlimited)
	if !tiers[3].MaxPrice.IsZero() {
		t.Error("tier 4 max price should be 0 (unlimited)")
	}
	if !tiers[3].UseUSDCap {
		t.Error("tier 4 should use USD cap")
	}
	if !tiers[3].USDCap.Equal(math.LegacyNewDec(50)) {
		t.Error("tier 4 USD cap should be $50")
	}
}
