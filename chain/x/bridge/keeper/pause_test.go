package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestRequireNotPaused_GlobalPause(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Initially not paused
	err := k.RequireNotPaused(ctx, "paw")
	require.NoError(t, err)

	// Set global pause
	params := k.GetParams(ctx)
	params.Paused = true
	k.SetParams(ctx, params)

	// Should be blocked
	err = k.RequireNotPaused(ctx, "paw")
	require.Error(t, err)
	require.Contains(t, err.Error(), "globally paused")
}

func TestRequireNotPaused_PerChainPause(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Set per-chain pause for paw
	params := k.GetParams(ctx)
	params.PausedChains = []string{"paw"}
	k.SetParams(ctx, params)

	// PAW should be blocked
	err := k.RequireNotPaused(ctx, "paw")
	require.Error(t, err)
	require.Contains(t, err.Error(), "paused for chain paw")

	// XAI should still work
	err = k.RequireNotPaused(ctx, "xai")
	require.NoError(t, err)
}

func TestRequireNotPaused_CaseInsensitive(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Pause "paw" (lowercase)
	params := k.GetParams(ctx)
	params.PausedChains = []string{"paw"}
	k.SetParams(ctx, params)

	// "PAW" (uppercase) should also be blocked
	err := k.RequireNotPaused(ctx, "PAW")
	require.Error(t, err)

	// " paw " (with spaces) should also be blocked
	err = k.RequireNotPaused(ctx, " paw ")
	require.Error(t, err)
}

func TestIsEmergencyPauseAuthorized(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	guardian1 := keepertest.GenTestAddr().String()
	guardian2 := keepertest.GenTestAddr().String()
	unauthorized := keepertest.GenTestAddr().String()

	// Set authorized guardians
	params := k.GetParams(ctx)
	params.EmergencyPauseAddresses = []string{guardian1, guardian2}
	k.SetParams(ctx, params)

	// Guardians should be authorized
	require.True(t, k.IsEmergencyPauseAuthorized(ctx, guardian1))
	require.True(t, k.IsEmergencyPauseAuthorized(ctx, guardian2))

	// Unauthorized address should not be authorized
	require.False(t, k.IsEmergencyPauseAuthorized(ctx, unauthorized))

	// Empty address should not be authorized
	require.False(t, k.IsEmergencyPauseAuthorized(ctx, ""))
}

func TestCheckAndTriggerAutoPause_Disabled(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Auto-pause disabled by default
	params := k.GetParams(ctx)
	require.False(t, params.AutoPauseEnabled)

	// Should not trigger even with large amount
	amount := sdkmath.NewInt(10_000_000_000)
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", amount)
	require.False(t, triggered)

	// Bridge should still be active
	params = k.GetParams(ctx)
	require.False(t, params.Paused)
}

func TestCheckAndTriggerAutoPause_BelowThreshold(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Enable auto-pause with threshold of 5 billion
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = true
	params.AutoPauseThreshold = "5000000000"
	k.SetParams(ctx, params)

	// Mint 1 billion (below threshold)
	amount := sdkmath.NewInt(1_000_000_000)
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", amount)
	require.False(t, triggered)

	// Bridge should still be active
	params = k.GetParams(ctx)
	require.False(t, params.Paused)
}

func TestCheckAndTriggerAutoPause_ExceedsThreshold(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Enable auto-pause with threshold of 5 billion
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = true
	params.AutoPauseThreshold = "5000000000"
	k.SetParams(ctx, params)

	// Record 4 billion minted
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(4_000_000_000))

	// Try to mint 2 billion more (total 6 billion > 5 billion threshold)
	amount := sdkmath.NewInt(2_000_000_000)
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", amount)
	require.True(t, triggered)

	// Bridge should be paused
	params = k.GetParams(ctx)
	require.True(t, params.Paused)
}

func TestCheckAndTriggerAutoPause_InvalidThreshold(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Enable auto-pause with invalid threshold
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = true
	params.AutoPauseThreshold = "invalid"
	k.SetParams(ctx, params)

	// Should not trigger with invalid threshold
	amount := sdkmath.NewInt(10_000_000_000)
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", amount)
	require.False(t, triggered)
}

func TestGetHourlyMintedAmount_Empty(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// No mints recorded
	hourly := k.GetHourlyMintedAmount(ctx, "test-denom")
	require.True(t, hourly.IsZero())
}

func TestGetHourlyMintedAmount_WithRecords(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Record some mints
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(1_000_000))
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(2_000_000))
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(3_000_000))

	// Should sum all mints
	hourly := k.GetHourlyMintedAmount(ctx, "test-denom")
	expected := sdkmath.NewInt(6_000_000)
	require.Equal(t, expected, hourly)
}

func TestRecordMintedAmount_EmitsEvent(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Record mint
	amount := sdkmath.NewInt(1_000_000)
	k.RecordMintedAmount(ctx, "test-denom", amount)

	// Check event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	// Find mint_recorded event
	found := false
	for _, event := range events {
		if event.Type == "mint_recorded" {
			found = true
			// Verify attributes
			for _, attr := range event.Attributes {
				if attr.Key == "denom" {
					require.Equal(t, "test-denom", attr.Value)
				}
				if attr.Key == "amount" {
					require.Equal(t, amount.String(), attr.Value)
				}
			}
		}
	}
	require.True(t, found, "mint_recorded event not found")
}

func TestRecordMintedAmount_ZeroAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx

	// Record zero amount (should be ignored)
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.ZeroInt())

	// Should not be recorded
	hourly := k.GetHourlyMintedAmount(ctx, "test-denom")
	require.True(t, hourly.IsZero())
}
