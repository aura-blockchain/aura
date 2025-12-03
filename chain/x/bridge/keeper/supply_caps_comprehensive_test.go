package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
)

// TestSupplyCaps_GetDailyMintedAmount tests daily mint amount tracking
func TestSupplyCaps_GetDailyMintedAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	denom := "wrapped.eth.usdc"

	// Initially should be zero
	amount := k.GetDailyMintedAmount(ctx, denom)
	require.Equal(t, sdkmath.ZeroInt(), amount, "Daily minted should start at zero")

	// Add 100k
	k.AddDailyMintedAmount(ctx, denom, sdkmath.NewInt(100000))
	amount = k.GetDailyMintedAmount(ctx, denom)
	require.Equal(t, sdkmath.NewInt(100000), amount, "Daily minted should be 100k")

	// Add 50k more
	k.AddDailyMintedAmount(ctx, denom, sdkmath.NewInt(50000))
	amount = k.GetDailyMintedAmount(ctx, denom)
	require.Equal(t, sdkmath.NewInt(150000), amount, "Daily minted should be 150k")
}

// TestSupplyCaps_GetHourlyMintedAmount tests hourly mint amount tracking
func TestSupplyCaps_GetHourlyMintedAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	denom := "wrapped.eth.usdc"

	// Initially should be zero
	amount := k.GetHourlyMintedAmount(ctx, denom)
	require.Equal(t, sdkmath.ZeroInt(), amount, "Hourly minted should start at zero")

	// Add 200k
	k.AddHourlyMintedAmount(ctx, denom, sdkmath.NewInt(200000))
	amount = k.GetHourlyMintedAmount(ctx, denom)
	require.Equal(t, sdkmath.NewInt(200000), amount, "Hourly minted should be 200k")

	// Add 100k more
	k.AddHourlyMintedAmount(ctx, denom, sdkmath.NewInt(100000))
	amount = k.GetHourlyMintedAmount(ctx, denom)
	require.Equal(t, sdkmath.NewInt(300000), amount, "Hourly minted should be 300k")
}

// TestSupplyCaps_MultipleTokens tests that tracking is separate per token
func TestSupplyCaps_MultipleTokens(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	denom1 := "wrapped.eth.usdc"
	denom2 := "wrapped.eth.usdt"

	// Add amounts for both tokens
	k.AddDailyMintedAmount(ctx, denom1, sdkmath.NewInt(100000))
	k.AddDailyMintedAmount(ctx, denom2, sdkmath.NewInt(200000))

	k.AddHourlyMintedAmount(ctx, denom1, sdkmath.NewInt(50000))
	k.AddHourlyMintedAmount(ctx, denom2, sdkmath.NewInt(75000))

	// Verify each token tracks separately
	daily1 := k.GetDailyMintedAmount(ctx, denom1)
	daily2 := k.GetDailyMintedAmount(ctx, denom2)
	require.Equal(t, sdkmath.NewInt(100000), daily1, "Denom1 daily should be 100k")
	require.Equal(t, sdkmath.NewInt(200000), daily2, "Denom2 daily should be 200k")

	hourly1 := k.GetHourlyMintedAmount(ctx, denom1)
	hourly2 := k.GetHourlyMintedAmount(ctx, denom2)
	require.Equal(t, sdkmath.NewInt(50000), hourly1, "Denom1 hourly should be 50k")
	require.Equal(t, sdkmath.NewInt(75000), hourly2, "Denom2 hourly should be 75k")
}

// TestSupplyCaps_EmptyDenom tests edge case of empty denom
func TestSupplyCaps_EmptyDenom(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	// GetDailyMintedAmount with empty denom should return zero
	daily := k.GetDailyMintedAmount(ctx, "")
	require.Equal(t, sdkmath.ZeroInt(), daily, "Empty denom should return zero")

	// GetHourlyMintedAmount with empty denom should return zero
	hourly := k.GetHourlyMintedAmount(ctx, "")
	require.Equal(t, sdkmath.ZeroInt(), hourly, "Empty denom should return zero")

	// AddDailyMintedAmount and AddHourlyMintedAmount should be no-ops with empty denom
	k.AddDailyMintedAmount(ctx, "", sdkmath.NewInt(100))
	k.AddHourlyMintedAmount(ctx, "", sdkmath.NewInt(100))

	// Verify nothing was stored
	daily = k.GetDailyMintedAmount(ctx, "")
	hourly = k.GetHourlyMintedAmount(ctx, "")
	require.Equal(t, sdkmath.ZeroInt(), daily, "Should still be zero")
	require.Equal(t, sdkmath.ZeroInt(), hourly, "Should still be zero")
}

// TestSupplyCaps_ZeroAmount tests that zero amounts are not tracked
func TestSupplyCaps_ZeroAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	denom := "wrapped.eth.usdc"

	// Try to add zero amount
	k.AddDailyMintedAmount(ctx, denom, sdkmath.ZeroInt())
	k.AddHourlyMintedAmount(ctx, denom, sdkmath.ZeroInt())

	// Verify nothing was stored
	daily := k.GetDailyMintedAmount(ctx, denom)
	hourly := k.GetHourlyMintedAmount(ctx, denom)
	require.Equal(t, sdkmath.ZeroInt(), daily, "Zero amount should not be tracked")
	require.Equal(t, sdkmath.ZeroInt(), hourly, "Zero amount should not be tracked")
}

// TestSupplyCaps_AddAndGet tests adding and retrieving mint amounts
func TestSupplyCaps_AddAndGet(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	denom := "wrapped.eth.usdc"

	// Add amounts in multiple calls
	k.AddDailyMintedAmount(ctx, denom, sdkmath.NewInt(1000))
	k.AddDailyMintedAmount(ctx, denom, sdkmath.NewInt(2000))
	k.AddDailyMintedAmount(ctx, denom, sdkmath.NewInt(3000))

	k.AddHourlyMintedAmount(ctx, denom, sdkmath.NewInt(500))
	k.AddHourlyMintedAmount(ctx, denom, sdkmath.NewInt(700))

	// Verify cumulative amounts
	daily := k.GetDailyMintedAmount(ctx, denom)
	hourly := k.GetHourlyMintedAmount(ctx, denom)

	require.Equal(t, sdkmath.NewInt(6000), daily, "Daily should be cumulative: 1000+2000+3000")
	require.Equal(t, sdkmath.NewInt(1200), hourly, "Hourly should be cumulative: 500+700")
}
