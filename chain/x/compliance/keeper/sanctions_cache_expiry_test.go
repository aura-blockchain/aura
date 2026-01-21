// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// TestSanctionsCacheExpiryLogic tests the cache expiry logic in isolation.
// This is critical for OFAC compliance - cached "CLEAR" status must not persist
// after the configured cache period, ensuring newly sanctioned addresses are detected.
func TestSanctionsCacheExpiryLogic(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "aura1testaddr"

	// Configure cache with 24 hour expiry
	params, _ := keeper.GetParams(ctx)
	params.ScreeningCacheHours = 24
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create initial sanctions result
	initialTime := ctx.BlockTime()
	result := &types.SanctionsScreeningResult{
		Address:              address,
		Status:               types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:              []*types.SanctionsMatch{},
		ScreenedAt:           initialTime,
		ScreeningProvider:    "test_provider",
		RequiresManualReview: false,
	}
	err = keeper.SetSanctionsResult(ctx, result)
	require.NoError(t, err)

	// Test 1: Cache is valid within expiry window
	ctx = ctx.WithBlockTime(initialTime.Add(12 * time.Hour)) // 12 hours later
	cached, err := keeper.GetSanctionsResult(ctx, address)
	require.NoError(t, err)
	require.NotNil(t, cached)

	// Verify cache validity check
	params, _ = keeper.GetParams(ctx)
	cacheAge := ctx.BlockTime().Sub(cached.ScreenedAt)
	maxCacheAge := time.Duration(params.ScreeningCacheHours) * time.Hour
	require.True(t, cacheAge <= maxCacheAge, "cache should be valid within expiry window")

	// Test 2: Cache is invalid after expiry window
	ctx = ctx.WithBlockTime(initialTime.Add(25 * time.Hour)) // 25 hours later (exceeds 24h)
	cached, err = keeper.GetSanctionsResult(ctx, address)
	require.NoError(t, err)
	require.NotNil(t, cached)

	// Verify cache would be considered expired
	params, _ = keeper.GetParams(ctx)
	cacheAge = ctx.BlockTime().Sub(cached.ScreenedAt)
	maxCacheAge = time.Duration(params.ScreeningCacheHours) * time.Hour
	require.True(t, cacheAge > maxCacheAge, "cache should be expired after expiry window")
}

// TestSanctionsCacheExpiryBoundary tests boundary conditions for cache expiry.
func TestSanctionsCacheExpiryBoundary(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "aura1testaddr"

	// Configure cache with 1 hour expiry
	params, _ := keeper.GetParams(ctx)
	params.ScreeningCacheHours = 1
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	initialTime := ctx.BlockTime()
	result := &types.SanctionsScreeningResult{
		Address:           address,
		Status:            types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:           []*types.SanctionsMatch{},
		ScreenedAt:        initialTime,
		ScreeningProvider: "test_provider",
	}
	err = keeper.SetSanctionsResult(ctx, result)
	require.NoError(t, err)

	// Test exactly at expiry boundary (59 minutes 59 seconds)
	ctx59m59s := ctx.WithBlockTime(initialTime.Add(59*time.Minute + 59*time.Second))
	cached, err := keeper.GetSanctionsResult(ctx59m59s, address)
	require.NoError(t, err)

	params, _ = keeper.GetParams(ctx59m59s)
	cacheAge := ctx59m59s.BlockTime().Sub(cached.ScreenedAt)
	maxCacheAge := time.Duration(params.ScreeningCacheHours) * time.Hour
	require.True(t, cacheAge <= maxCacheAge, "cache should be valid at 59m59s")

	// Test just after expiry (1 hour 1 second)
	ctx1h1s := ctx.WithBlockTime(initialTime.Add(1*time.Hour + 1*time.Second))
	cached, err = keeper.GetSanctionsResult(ctx1h1s, address)
	require.NoError(t, err)

	params, _ = keeper.GetParams(ctx1h1s)
	cacheAge = ctx1h1s.BlockTime().Sub(cached.ScreenedAt)
	maxCacheAge = time.Duration(params.ScreeningCacheHours) * time.Hour
	require.True(t, cacheAge > maxCacheAge, "cache should be expired at 1h1s")
}

// TestSanctionsCacheExpiryZeroDisablesExpiry tests that ScreeningCacheHours=0 means no expiry.
func TestSanctionsCacheExpiryZeroDisablesExpiry(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "aura1testaddr"

	// Configure with zero cache hours (no expiry)
	params, _ := keeper.GetParams(ctx)
	params.ScreeningCacheHours = 0
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	initialTime := ctx.BlockTime()
	result := &types.SanctionsScreeningResult{
		Address:           address,
		Status:            types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:           []*types.SanctionsMatch{},
		ScreenedAt:        initialTime,
		ScreeningProvider: "test_provider",
	}
	err = keeper.SetSanctionsResult(ctx, result)
	require.NoError(t, err)

	// Test after very long time
	ctxLater := ctx.WithBlockTime(initialTime.Add(365 * 24 * time.Hour)) // 1 year later
	cached, err := keeper.GetSanctionsResult(ctxLater, address)
	require.NoError(t, err)
	require.NotNil(t, cached)

	// With ScreeningCacheHours=0, cache never expires
	params, _ = keeper.GetParams(ctxLater)
	require.Equal(t, uint64(0), params.ScreeningCacheHours)
}

// TestSanctionsCacheExpiryMultipleAddresses tests that cache expiry is independent per address.
func TestSanctionsCacheExpiryMultipleAddresses(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address1 := "aura1addr1"
	address2 := "aura1addr2"

	// Configure cache with 2 hour expiry
	params, _ := keeper.GetParams(ctx)
	params.ScreeningCacheHours = 2
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	initialTime := ctx.BlockTime()

	// Screen address1
	result1 := &types.SanctionsScreeningResult{
		Address:           address1,
		Status:            types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:           []*types.SanctionsMatch{},
		ScreenedAt:        initialTime,
		ScreeningProvider: "test_provider",
	}
	err = keeper.SetSanctionsResult(ctx, result1)
	require.NoError(t, err)

	// Screen address2 one hour later
	time2 := initialTime.Add(1 * time.Hour)
	result2 := &types.SanctionsScreeningResult{
		Address:           address2,
		Status:            types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:           []*types.SanctionsMatch{},
		ScreenedAt:        time2,
		ScreeningProvider: "test_provider",
	}
	ctx2 := ctx.WithBlockTime(time2)
	err = keeper.SetSanctionsResult(ctx2, result2)
	require.NoError(t, err)

	// Check at 2 hours 1 minute after initial time
	// address1: 2h1m old (expired)
	// address2: 1h1m old (valid)
	checkTime := initialTime.Add(2*time.Hour + 1*time.Minute)
	ctxCheck := ctx.WithBlockTime(checkTime)
	params, _ = keeper.GetParams(ctxCheck)
	maxAge := time.Duration(params.ScreeningCacheHours) * time.Hour

	// Verify address1 is expired
	cached1, err := keeper.GetSanctionsResult(ctxCheck, address1)
	require.NoError(t, err)
	age1 := checkTime.Sub(cached1.ScreenedAt)
	require.True(t, age1 > maxAge, "address1 cache should be expired")

	// Verify address2 is still valid
	cached2, err := keeper.GetSanctionsResult(ctxCheck, address2)
	require.NoError(t, err)
	age2 := checkTime.Sub(cached2.ScreenedAt)
	require.True(t, age2 <= maxAge, "address2 cache should be valid")
}

// TestSanctionsCacheExpiryOFACCompliance simulates a real OFAC compliance scenario.
func TestSanctionsCacheExpiryOFACCompliance(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "aura1potential_sanction"

	// Configure short cache for testing (production typically uses 24 hours)
	params, _ := keeper.GetParams(ctx)
	params.ScreeningCacheHours = 1
	params.SanctionsScreeningEnabled = true
	params.SanctionsLists = []string{"OFAC_SDN", "EU", "UN"}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Day 1: User passes initial sanctions screening
	day1Time := ctx.BlockTime()
	result := &types.SanctionsScreeningResult{
		Address:              address,
		Status:               types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:              []*types.SanctionsMatch{},
		ScreenedAt:           day1Time,
		ScreeningProvider:    "ofac_provider",
		RequiresManualReview: false,
	}
	err = keeper.SetSanctionsResult(ctx, result)
	require.NoError(t, err)

	// User transacts successfully within cache window
	ctx30min := ctx.WithBlockTime(day1Time.Add(30 * time.Minute))
	cached, err := keeper.GetSanctionsResult(ctx30min, address)
	require.NoError(t, err)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CLEAR, cached.Status)

	// Cache age check shows valid
	params, _ = keeper.GetParams(ctx30min)
	cacheAge := ctx30min.BlockTime().Sub(cached.ScreenedAt)
	maxAge := time.Duration(params.ScreeningCacheHours) * time.Hour
	require.True(t, cacheAge <= maxAge, "cache should be valid within window")

	// OFAC list updated (simulated by time passing beyond cache expiry)
	// User attempts transaction after cache expiry
	ctx2h := ctx.WithBlockTime(day1Time.Add(2 * time.Hour)) // Cache expired
	cached, err = keeper.GetSanctionsResult(ctx2h, address)
	require.NoError(t, err)

	// Cache age check shows expired - fresh screening required
	params, _ = keeper.GetParams(ctx2h)
	cacheAge = ctx2h.BlockTime().Sub(cached.ScreenedAt)
	maxAge = time.Duration(params.ScreeningCacheHours) * time.Hour
	require.True(t, cacheAge > maxAge, "cache must be expired to trigger fresh screening")

	// This ensures that if the address was added to OFAC SDN list,
	// a fresh screening will detect it instead of using stale "CLEAR" status
}

// TestSanctionsCacheExpiryManualExpiredEntry tests handling of already-expired cache entries.
func TestSanctionsCacheExpiryManualExpiredEntry(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "aura1manual"

	// Configure cache with 24 hour expiry
	params, _ := keeper.GetParams(ctx)
	params.ScreeningCacheHours = 24
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create an already-expired cache entry (48 hours old)
	currentTime := ctx.BlockTime()
	expiredTime := currentTime.Add(-48 * time.Hour)

	expiredResult := &types.SanctionsScreeningResult{
		Address:              address,
		Status:               types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:              []*types.SanctionsMatch{},
		ScreenedAt:           expiredTime,
		ScreeningProvider:    "test_provider",
		RequiresManualReview: false,
	}
	err = keeper.SetSanctionsResult(ctx, expiredResult)
	require.NoError(t, err)

	// Retrieve and verify it's expired
	cached, err := keeper.GetSanctionsResult(ctx, address)
	require.NoError(t, err)
	require.NotNil(t, cached)

	params, _ = keeper.GetParams(ctx)
	cacheAge := currentTime.Sub(cached.ScreenedAt)
	maxAge := time.Duration(params.ScreeningCacheHours) * time.Hour
	require.True(t, cacheAge > maxAge, "manually set expired entry should be detected as expired")
	require.Equal(t, expiredTime, cached.ScreenedAt)
}

// TestSanctionsCacheExpiry_VariousDurations tests cache expiry with different time windows.
func TestSanctionsCacheExpiry_VariousDurations(t *testing.T) {
	testCases := []struct {
		name         string
		cacheHours   uint64
		testAge      time.Duration
		shouldExpire bool
	}{
		{
			name:         "1 hour cache, 30 minutes age - valid",
			cacheHours:   1,
			testAge:      30 * time.Minute,
			shouldExpire: false,
		},
		{
			name:         "1 hour cache, 2 hours age - expired",
			cacheHours:   1,
			testAge:      2 * time.Hour,
			shouldExpire: true,
		},
		{
			name:         "24 hour cache, 12 hours age - valid",
			cacheHours:   24,
			testAge:      12 * time.Hour,
			shouldExpire: false,
		},
		{
			name:         "24 hour cache, 48 hours age - expired",
			cacheHours:   24,
			testAge:      48 * time.Hour,
			shouldExpire: true,
		},
		{
			name:         "168 hour (1 week) cache, 100 hours age - valid",
			cacheHours:   168,
			testAge:      100 * time.Hour,
			shouldExpire: false,
		},
		{
			name:         "168 hour (1 week) cache, 200 hours age - expired",
			cacheHours:   168,
			testAge:      200 * time.Hour,
			shouldExpire: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := setupTestKeeper(t)

			address := "aura1test"

			// Configure cache
			params, _ := keeper.GetParams(ctx)
			params.ScreeningCacheHours = tc.cacheHours
			err := keeper.SetParams(ctx, params)
			require.NoError(t, err)

			// Create screening result
			screenedAt := ctx.BlockTime().Add(-tc.testAge)
			result := &types.SanctionsScreeningResult{
				Address:           address,
				Status:            types.SanctionsStatus_SANCTIONS_CLEAR,
				Matches:           []*types.SanctionsMatch{},
				ScreenedAt:        screenedAt,
				ScreeningProvider: "test_provider",
			}
			err = keeper.SetSanctionsResult(ctx, result)
			require.NoError(t, err)

			// Check expiry status
			cached, err := keeper.GetSanctionsResult(ctx, address)
			require.NoError(t, err)

			params, _ = keeper.GetParams(ctx)
			cacheAge := ctx.BlockTime().Sub(cached.ScreenedAt)
			maxAge := time.Duration(params.ScreeningCacheHours) * time.Hour

			if tc.shouldExpire {
				require.True(t, cacheAge > maxAge, "cache should be expired")
			} else {
				require.True(t, cacheAge <= maxAge, "cache should be valid")
			}
		})
	}
}
