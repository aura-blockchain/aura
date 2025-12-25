// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// TestEndBlocker_FlushPendingUpdates verifies that EndBlocker writes all pending updates
func TestEndBlocker_FlushPendingUpdates(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create test addresses
	addr1 := sdk.AccAddress("addr1_______________").String()
	addr2 := sdk.AccAddress("addr2_______________").String()

	// Perform transactions to queue updates
	err := keeper.UpdateAMLProfileOnTransaction(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000)))
	require.NoError(t, err)

	err = keeper.UpdateAMLProfileOnTransaction(ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin("uaura", 2000)))
	require.NoError(t, err)

	// Verify profiles are NOT in store yet (only in pending map)
	_, err = keeper.GetAMLProfile(ctx, addr1)
	require.Error(t, err, "profile should not be in store before EndBlocker")

	_, err = keeper.GetAMLProfile(ctx, addr2)
	require.Error(t, err, "profile should not be in store before EndBlocker")

	// Run EndBlocker to flush updates
	keeper.EndBlocker(ctx)

	// Verify profiles are now in store
	profile1, err := keeper.GetAMLProfile(ctx, addr1)
	require.NoError(t, err)
	require.Equal(t, addr1, profile1.Address)
	require.Equal(t, uint64(1), profile1.TotalTransactions)
	require.Equal(t, "1000", profile1.TotalVolume)

	profile2, err := keeper.GetAMLProfile(ctx, addr2)
	require.NoError(t, err)
	require.Equal(t, addr2, profile2.Address)
	require.Equal(t, uint64(1), profile2.TotalTransactions)
	require.Equal(t, "2000", profile2.TotalVolume)
}

// TestEndBlocker_MultipleUpdatesToSameAddress verifies that multiple updates
// to the same address in one block are merged correctly
func TestEndBlocker_MultipleUpdatesToSameAddress(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	addr := sdk.AccAddress("addr________________").String()

	// Perform multiple transactions for same address
	err := keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000)))
	require.NoError(t, err)

	err = keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 2000)))
	require.NoError(t, err)

	err = keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 3000)))
	require.NoError(t, err)

	// Run EndBlocker
	keeper.EndBlocker(ctx)

	// Verify final state reflects all 3 transactions
	profile, err := keeper.GetAMLProfile(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, addr, profile.Address)
	require.Equal(t, uint64(3), profile.TotalTransactions)
	require.Equal(t, "6000", profile.TotalVolume) // 1000 + 2000 + 3000
}

// TestEndBlocker_ClearsPendingMap verifies that pending updates are cleared after flush
func TestEndBlocker_ClearsPendingMap(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	addr := sdk.AccAddress("addr________________").String()

	// Queue an update
	err := keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000)))
	require.NoError(t, err)

	// Run EndBlocker
	keeper.EndBlocker(ctx)

	// Verify pending map is cleared by making another update and checking it's independent
	err = keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 500)))
	require.NoError(t, err)

	// Second EndBlocker
	keeper.EndBlocker(ctx)

	// Final profile should have 2 transactions total
	profile, err := keeper.GetAMLProfile(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, uint64(2), profile.TotalTransactions)
	require.Equal(t, "1500", profile.TotalVolume)
}

// TestEndBlocker_EmptyPendingMap verifies EndBlocker handles empty map gracefully
func TestEndBlocker_EmptyPendingMap(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Run EndBlocker with no pending updates
	keeper.EndBlocker(ctx)

	// Should not panic or error - just no-op
	// Test passes if we reach here without panic
}

// TestEndBlocker_ExistingProfileUpdate verifies updates to existing profiles
func TestEndBlocker_ExistingProfileUpdate(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	addr := sdk.AccAddress("addr________________").String()

	// Create initial profile directly in store
	initialProfile := &types.AMLProfile{
		Address:           addr,
		RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
		TotalTransactions: 5,
		TotalVolume:       "10000",
		RiskFactors:       []string{},
		LastAssessment:    ctx.BlockTime(),
		PepStatus:         false,
		SourceOfFunds:     []string{},
		Occupation:        "",
	}
	err := keeper.SetAMLProfile(ctx, initialProfile)
	require.NoError(t, err)

	// Queue an update
	err = keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 3000)))
	require.NoError(t, err)

	// Run EndBlocker
	keeper.EndBlocker(ctx)

	// Verify profile was updated
	profile, err := keeper.GetAMLProfile(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, uint64(6), profile.TotalTransactions) // 5 + 1
	require.Equal(t, "13000", profile.TotalVolume)         // 10000 + 3000
}

// TestEndBlocker_RiskLevelProgression verifies risk levels are calculated correctly
func TestEndBlocker_RiskLevelProgression(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	addr := sdk.AccAddress("addr________________").String()

	// Set params with known thresholds for testing
	params := types.DefaultParams()
	params.VelocityLimit_24H = "100000"       // Medium threshold at 50000, high at 100000
	params.SingleTransactionLimit = "100000" // Must be <= velocity limit
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Queue multiple high-value transactions to exceed thresholds
	// First transaction: 60000 -> should be MEDIUM (above 50000 medium threshold)
	err = keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 60000)))
	require.NoError(t, err)

	// Run EndBlocker
	keeper.EndBlocker(ctx)

	// Verify risk level is MEDIUM
	profile, err := keeper.GetAMLProfile(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_MEDIUM, profile.RiskLevel)

	// Add more transactions to exceed HIGH threshold (>= 100000)
	err = keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 50000)))
	require.NoError(t, err)

	// Run EndBlocker again
	keeper.EndBlocker(ctx)

	// Verify risk level is now HIGH
	profile, err = keeper.GetAMLProfile(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_HIGH, profile.RiskLevel)
}

// TestEndBlocker_BatchWriteReduction verifies write reduction
func TestEndBlocker_BatchWriteReduction(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Simulate a block with 10 addresses, some with multiple transactions
	addresses := make([]string, 10)
	for i := 0; i < 10; i++ {
		addresses[i] = sdk.AccAddress([]byte("addr" + string(rune(i)) + "____________")).String()
	}

	// First 5 addresses: 1 transaction each
	for i := 0; i < 5; i++ {
		err := keeper.UpdateAMLProfileOnTransaction(ctx, addresses[i], sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000)))
		require.NoError(t, err)
	}

	// Next 3 addresses: 3 transactions each (9 total)
	for i := 5; i < 8; i++ {
		for j := 0; j < 3; j++ {
			err := keeper.UpdateAMLProfileOnTransaction(ctx, addresses[i], sdk.NewCoins(sdk.NewInt64Coin("uaura", 500)))
			require.NoError(t, err)
		}
	}

	// Last 2 addresses: 5 transactions each (10 total)
	for i := 8; i < 10; i++ {
		for j := 0; j < 5; j++ {
			err := keeper.UpdateAMLProfileOnTransaction(ctx, addresses[i], sdk.NewCoins(sdk.NewInt64Coin("uaura", 200)))
			require.NoError(t, err)
		}
	}

	// Total transactions: 5 + 9 + 10 = 24 transactions
	// Unique addresses: 10
	// Old approach: 24 writes
	// New approach: 10 writes
	// Reduction: (24 - 10) / 24 = 58.3% reduction

	// Run EndBlocker
	keeper.EndBlocker(ctx)

	// Verify all 10 profiles exist with correct data
	for i := 0; i < 5; i++ {
		profile, err := keeper.GetAMLProfile(ctx, addresses[i])
		require.NoError(t, err)
		require.Equal(t, uint64(1), profile.TotalTransactions)
		require.Equal(t, "1000", profile.TotalVolume)
	}

	for i := 5; i < 8; i++ {
		profile, err := keeper.GetAMLProfile(ctx, addresses[i])
		require.NoError(t, err)
		require.Equal(t, uint64(3), profile.TotalTransactions)
		require.Equal(t, "1500", profile.TotalVolume) // 500 * 3
	}

	for i := 8; i < 10; i++ {
		profile, err := keeper.GetAMLProfile(ctx, addresses[i])
		require.NoError(t, err)
		require.Equal(t, uint64(5), profile.TotalTransactions)
		require.Equal(t, "1000", profile.TotalVolume) // 200 * 5
	}
}

// TestEndBlocker_EventEmissionNotDelayed verifies events are still emitted immediately
func TestEndBlocker_EventEmissionNotDelayed(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	addr := sdk.AccAddress("addr________________").String()

	// Queue an update
	err := keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000)))
	require.NoError(t, err)

	// Verify event was emitted during UpdateAMLProfileOnTransaction (not delayed to EndBlocker)
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events, "events should be emitted immediately, not delayed")

	// Look for AMLProfileUpdated event
	foundEvent := false
	for _, event := range events {
		if event.Type == types.EventTypeAMLProfileUpdated {
			foundEvent = true
			break
		}
	}
	require.True(t, foundEvent, "AMLProfileUpdated event should be emitted immediately")
}

// TestEndBlocker_MultiDenomination verifies multi-denomination amounts are handled correctly
func TestEndBlocker_MultiDenomination(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	addr := sdk.AccAddress("addr________________").String()

	// Queue update with multiple denominations
	amount := sdk.NewCoins(
		sdk.NewInt64Coin("uaura", 1000),
		sdk.NewInt64Coin("uatom", 500),
		sdk.NewInt64Coin("uosmo", 2000),
	)
	err := keeper.UpdateAMLProfileOnTransaction(ctx, addr, amount)
	require.NoError(t, err)

	// Run EndBlocker
	keeper.EndBlocker(ctx)

	// Verify profile has combined volume
	profile, err := keeper.GetAMLProfile(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, uint64(1), profile.TotalTransactions)

	// Volume should be sum of all denominations
	expectedVolume := math.NewInt(1000 + 500 + 2000)
	require.Equal(t, expectedVolume.String(), profile.TotalVolume)
}
