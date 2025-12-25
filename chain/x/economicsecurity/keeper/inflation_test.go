// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// TestAdjustInflationRate_Success tests successful inflation rate adjustment
func TestAdjustInflationRate_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set initial params with inflation bounds
	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:           "1000000000", // Required for validation
		CirculatingSupply:   "100000000",  // Required for validation
		InflationRate:       500,          // 5.00%
		TargetInflationRate: 500,          // 5.00%
		MinInflationRate:    100,          // 1.00%
		MaxInflationRate:    1000,         // 10.00%
		LastInflationAdjustment: time.Now(),
		LastInflationCheck:      time.Now(),
	}
	require.NoError(t, keeper.SetParams(params))

	// Test adjustment within bounds
	authority := keeper.GetAuthority()
	newRate := uint64(600) // 6.00%
	reason := "Governance decision to increase inflation for ecosystem growth"

	oldRate, err := keeper.AdjustInflationRate(ctx, authority, newRate, reason)
	require.NoError(t, err)
	require.Equal(t, uint64(500), oldRate)

	// Verify rate was updated
	updatedParams, _ := keeper.GetParams(ctx)
	require.Equal(t, newRate, updatedParams.Tokenomics.InflationRate)

	// Verify previous rate was stored
	previousRate, err := keeper.GetPreviousInflation(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(500), previousRate)

	// Verify event was emitted
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	events := sdkCtx.EventManager().Events()
	require.Greater(t, len(events), 0)

	// Find inflation adjusted event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeInflationAdjusted {
			foundEvent = true
			// Verify event attributes
			attributes := event.Attributes
			require.Greater(t, len(attributes), 0)

			var hasOldRate, hasNewRate, hasReason, hasAuthority bool
			for _, attr := range attributes {
				switch attr.Key {
				case types.AttributeKeyOldRate:
					hasOldRate = true
					require.Equal(t, "500", attr.Value)
				case types.AttributeKeyNewRate:
					hasNewRate = true
					require.Equal(t, "600", attr.Value)
				case types.AttributeKeyReason:
					hasReason = true
					require.Equal(t, reason, attr.Value)
				case types.AttributeKeyAuthority:
					hasAuthority = true
					require.Equal(t, authority, attr.Value)
				}
			}

			require.True(t, hasOldRate, "event missing old_rate attribute")
			require.True(t, hasNewRate, "event missing new_rate attribute")
			require.True(t, hasReason, "event missing reason attribute")
			require.True(t, hasAuthority, "event missing authority attribute")
		}
	}
	require.True(t, foundEvent, "inflation_adjusted event not found")
}

// TestAdjustInflationRate_Unauthorized tests rejection of unauthorized adjustment
func TestAdjustInflationRate_Unauthorized(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set initial params
	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:           "1000000000",
		CirculatingSupply:   "100000000",
		InflationRate:       500,
		TargetInflationRate: 500,
		MinInflationRate:    100,
		MaxInflationRate:    1000,
	}
	require.NoError(t, keeper.SetParams(params))

	// Attempt adjustment with wrong authority
	unauthorizedAuthority := "cosmos1unauthorized"
	newRate := uint64(600)
	reason := "Unauthorized attempt"

	_, err := keeper.AdjustInflationRate(ctx, unauthorizedAuthority, newRate, reason)
	require.ErrorIs(t, err, types.ErrUnauthorized)

	// Verify rate was NOT changed
	updatedParams, _ := keeper.GetParams(ctx)
	require.Equal(t, uint64(500), updatedParams.Tokenomics.InflationRate)
}

// TestAdjustInflationRate_RateTooHigh tests rejection of rate above maximum
func TestAdjustInflationRate_RateTooHigh(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:           "1000000000",
		CirculatingSupply:   "100000000",
		InflationRate:       500,
		TargetInflationRate: 500,
		MinInflationRate:    100,
		MaxInflationRate:    1000, // 10.00% max
	}
	require.NoError(t, keeper.SetParams(params))

	authority := keeper.GetAuthority()
	newRate := uint64(1500) // 15.00% - exceeds max
	reason := "Attempting excessive inflation"

	_, err := keeper.AdjustInflationRate(ctx, authority, newRate, reason)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInflationRateTooHigh)

	// Verify rate was NOT changed
	updatedParams, _ := keeper.GetParams(ctx)
	require.Equal(t, uint64(500), updatedParams.Tokenomics.InflationRate)
}

// TestAdjustInflationRate_RateTooLow tests rejection of rate below minimum
func TestAdjustInflationRate_RateTooLow(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:           "1000000000",
		CirculatingSupply:   "100000000",
		InflationRate:       500,
		TargetInflationRate: 500,
		MinInflationRate:    100, // 1.00% min
		MaxInflationRate:    1000,
	}
	require.NoError(t, keeper.SetParams(params))

	authority := keeper.GetAuthority()
	newRate := uint64(50) // 0.50% - below min
	reason := "Attempting too low inflation"

	_, err := keeper.AdjustInflationRate(ctx, authority, newRate, reason)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInflationRateTooLow)

	// Verify rate was NOT changed
	updatedParams, _ := keeper.GetParams(ctx)
	require.Equal(t, uint64(500), updatedParams.Tokenomics.InflationRate)
}

// TestAdjustInflationRate_SameRate tests rejection when new rate equals current
func TestAdjustInflationRate_SameRate(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:           "1000000000",
		CirculatingSupply:   "100000000",
		InflationRate:       500,
		TargetInflationRate: 500,
		MinInflationRate:    100,
		MaxInflationRate:    1000,
	}
	require.NoError(t, keeper.SetParams(params))

	authority := keeper.GetAuthority()
	newRate := uint64(500) // Same as current
	reason := "Pointless adjustment"

	_, err := keeper.AdjustInflationRate(ctx, authority, newRate, reason)
	require.Error(t, err)
	require.Contains(t, err.Error(), "new rate must differ from current rate")
}

// TestAdjustInflationRate_NoReason tests rejection when reason is empty
func TestAdjustInflationRate_NoReason(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:           "1000000000",
		CirculatingSupply:   "100000000",
		InflationRate:       500,
		TargetInflationRate: 500,
		MinInflationRate:    100,
		MaxInflationRate:    1000,
	}
	require.NoError(t, keeper.SetParams(params))

	authority := keeper.GetAuthority()
	newRate := uint64(600)
	reason := "" // Empty reason

	_, err := keeper.AdjustInflationRate(ctx, authority, newRate, reason)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reason is required")
}

// TestGetInflationMetrics_Success tests successful retrieval of inflation metrics
func TestGetInflationMetrics_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set up params with known values
	lastAdjustmentTime := time.Now().Add(-12 * time.Hour)
	lastCheckTime := time.Now().Add(-6 * time.Hour)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:               "1000000000",
		CirculatingSupply:       "100000000",
		InflationRate:           600, // 6.00%
		TargetInflationRate:     500, // 5.00%
		MinInflationRate:        100,
		MaxInflationRate:        1000,
		LastInflationAdjustment: lastAdjustmentTime,
		LastInflationCheck:      lastCheckTime,
	}
	params.InflationCheckInterval = 100 // blocks
	require.NoError(t, keeper.SetParams(params))

	// Set previous inflation for 24h change calculation
	require.NoError(t, keeper.SetPreviousInflation(ctx, 500))

	// Get metrics
	currentRate, targetRate, change24h, lastAdj, nextCheck, err := keeper.GetInflationMetrics(ctx)
	require.NoError(t, err)

	// Verify metrics
	require.Equal(t, uint64(600), currentRate)
	require.Equal(t, uint64(500), targetRate)
	require.Equal(t, int64(100), change24h) // 600 - 500 = 100
	require.NotNil(t, lastAdj)
	require.NotNil(t, nextCheck)

	// Verify last adjustment time matches
	require.True(t, lastAdj.Equal(lastAdjustmentTime))

	// Verify next check is after last check
	require.True(t, nextCheck.After(lastCheckTime))
}

// TestGetInflationMetrics_NoHistory tests metrics when no historical data exists
func TestGetInflationMetrics_NoHistory(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:           "1000000000",
		CirculatingSupply:   "100000000",
		InflationRate:       500,
		TargetInflationRate: 500,
		MinInflationRate:    100,
		MaxInflationRate:    1000,
		// No LastInflationAdjustment or LastInflationCheck
	}
	params.InflationCheckInterval = 100
	require.NoError(t, keeper.SetParams(params))

	// No previous inflation stored
	currentRate, targetRate, change24h, _, nextCheck, err := keeper.GetInflationMetrics(ctx)
	require.NoError(t, err)

	require.Equal(t, uint64(500), currentRate)
	require.Equal(t, uint64(500), targetRate)
	require.Equal(t, int64(0), change24h) // No history = 0 change
	// lastAdj will be nil, so we don't check it
	require.NotNil(t, nextCheck) // Should still compute next check
}

// TestCalculateInflation24hChange_Increase tests positive rate change
func TestCalculateInflation24hChange_Increase(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:         "1000000000",
		CirculatingSupply: "100000000",
		InflationRate: 700, // Current: 7.00%
	}
	require.NoError(t, keeper.SetParams(params))

	// Previous rate was lower
	require.NoError(t, keeper.SetPreviousInflation(ctx, 500)) // Previous: 5.00%

	change, err := keeper.CalculateInflation24hChange(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(200), change) // 700 - 500 = +200
}

// TestCalculateInflation24hChange_Decrease tests negative rate change
func TestCalculateInflation24hChange_Decrease(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:         "1000000000",
		CirculatingSupply: "100000000",
		InflationRate: 400, // Current: 4.00%
	}
	require.NoError(t, keeper.SetParams(params))

	// Previous rate was higher
	require.NoError(t, keeper.SetPreviousInflation(ctx, 600)) // Previous: 6.00%

	change, err := keeper.CalculateInflation24hChange(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(-200), change) // 400 - 600 = -200
}

// TestCalculateInflation24hChange_NoChange tests zero change
func TestCalculateInflation24hChange_NoChange(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:         "1000000000",
		CirculatingSupply: "100000000",
		InflationRate: 500,
	}
	require.NoError(t, keeper.SetParams(params))

	require.NoError(t, keeper.SetPreviousInflation(ctx, 500))

	change, err := keeper.CalculateInflation24hChange(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), change)
}

// TestCalculateInflation24hChange_NoPreviousData tests when no history exists
func TestCalculateInflation24hChange_NoPreviousData(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:         "1000000000",
		CirculatingSupply: "100000000",
		InflationRate: 500,
	}
	require.NoError(t, keeper.SetParams(params))

	// No previous inflation stored (returns 0 by default)
	change, err := keeper.CalculateInflation24hChange(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), change) // No history = 0 change
}

// TestUpdateInflationCheckTimestamp tests timestamp update
func TestUpdateInflationCheckTimestamp(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:         "1000000000",
		CirculatingSupply: "100000000",
		InflationRate:     500,
		// Initially no LastInflationCheck
	}
	require.NoError(t, keeper.SetParams(params))

	// Update timestamp
	err := keeper.UpdateInflationCheckTimestamp(ctx)
	require.NoError(t, err)

	// Verify timestamp was set
	updatedParams, _ := keeper.GetParams(ctx)
	require.False(t, updatedParams.Tokenomics.LastInflationCheck.IsZero())

	// Verify it matches the block time
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	require.Equal(t, sdkCtx.BlockTime(), updatedParams.Tokenomics.LastInflationCheck)
}

// TestAdjustInflationRate_BoundaryValues tests edge cases at boundaries
func TestAdjustInflationRate_BoundaryValues(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:           "1000000000",
		CirculatingSupply:   "100000000",
		InflationRate:       500,
		TargetInflationRate: 500,
		MinInflationRate:    100,  // 1.00%
		MaxInflationRate:    1000, // 10.00%
	}
	require.NoError(t, keeper.SetParams(params))

	authority := keeper.GetAuthority()

	// Test setting to exact minimum (should succeed)
	_, err := keeper.AdjustInflationRate(ctx, authority, 100, "Set to minimum")
	require.NoError(t, err)

	params, _ = keeper.GetParams(ctx)
	require.Equal(t, uint64(100), params.Tokenomics.InflationRate)

	// Test setting to exact maximum (should succeed)
	_, err = keeper.AdjustInflationRate(ctx, authority, 1000, "Set to maximum")
	require.NoError(t, err)

	params, _ = keeper.GetParams(ctx)
	require.Equal(t, uint64(1000), params.Tokenomics.InflationRate)

	// Test one below minimum (should fail)
	_, err = keeper.AdjustInflationRate(ctx, authority, 99, "Below minimum")
	require.ErrorIs(t, err, types.ErrInflationRateTooLow)

	// Test one above maximum (should fail)
	_, err = keeper.AdjustInflationRate(ctx, authority, 1001, "Above maximum")
	require.ErrorIs(t, err, types.ErrInflationRateTooHigh)
}

// TestInflationFunctions_Integration tests full workflow
func TestInflationFunctions_Integration(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Initialize params
	params, _ := keeper.GetParams(ctx)
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:           "1000000000",
		CirculatingSupply:   "100000000",
		InflationRate:       500,  // 5.00%
		TargetInflationRate: 500,  // 5.00%
		MinInflationRate:    100,  // 1.00%
		MaxInflationRate:    1000, // 10.00%
		LastInflationAdjustment: time.Now(),
		LastInflationCheck:      time.Now(),
	}
	params.InflationCheckInterval = 100
	require.NoError(t, keeper.SetParams(params))

	authority := keeper.GetAuthority()

	// Step 1: Get initial metrics
	currentRate, targetRate, change24h, _, _, err := keeper.GetInflationMetrics(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(500), currentRate)
	require.Equal(t, uint64(500), targetRate)
	require.Equal(t, int64(0), change24h) // No history yet

	// Step 2: Adjust inflation rate
	oldRate, err := keeper.AdjustInflationRate(ctx, authority, 650, "Test adjustment")
	require.NoError(t, err)
	require.Equal(t, uint64(500), oldRate)

	// Step 3: Get updated metrics
	currentRate, targetRate, change24h, _, _, err = keeper.GetInflationMetrics(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(650), currentRate) // Updated
	require.Equal(t, uint64(500), targetRate)  // Target unchanged
	require.Equal(t, int64(150), change24h)    // 650 - 500 = 150

	// Step 4: Update check timestamp
	err = keeper.UpdateInflationCheckTimestamp(ctx)
	require.NoError(t, err)

	// Step 5: Verify timestamp was updated
	params, _ = keeper.GetParams(ctx)
	require.NotNil(t, params.Tokenomics.LastInflationCheck)

	// Step 6: Calculate 24h change directly
	change, err := keeper.CalculateInflation24hChange(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(150), change)
}
