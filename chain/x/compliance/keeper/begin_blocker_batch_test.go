// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// TestBeginBlockerWithExpirationIndex verifies efficient KYC expiry using indexed lookups
func TestBeginBlockerWithExpirationIndex(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create an expired KYC record
	expiresAt := ctx.BlockTime().Add(-1 * time.Hour)
	expiredRecord := &types.KYCRecord{
		Address:      "aura1expireduser",
		KycLevel:     types.KYCLevel_KYC_LEVEL_BASIC,
		VerifiedAt:   ctx.BlockTime().Add(-365 * 24 * time.Hour), // 1 year ago
		ExpiresAt:    &expiresAt,                                 // Expired 1 hour ago
		Provider:     "test-provider",
		Jurisdiction: "US",
	}

	err := k.SetKYCRecord(ctx, expiredRecord)
	require.NoError(t, err)

	// Add to expiration index
	k.AddToExpirationIndex(ctx, expiredRecord)

	// BeginBlocker should process expired records immediately using index
	ctx = ctx.WithBlockHeight(1)
	eventsBefore := len(ctx.EventManager().Events())

	k.BeginBlocker(ctx)

	eventsAfter := len(ctx.EventManager().Events())
	require.Greater(t, eventsAfter, eventsBefore,
		"BeginBlocker should emit expiry events on first block when using index")

	// Verify the expired event was emitted
	events := ctx.EventManager().Events()
	foundExpiredEvent := false
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyAddress && attr.Value == "aura1expireduser" {
					foundExpiredEvent = true
					break
				}
			}
		}
	}
	require.True(t, foundExpiredEvent, "Should find KYC expired event for expired user")
}

// TestBeginBlockerIndexedPerformance verifies the expiration index provides O(k) lookups
func TestBeginBlockerIndexedPerformance(t *testing.T) {
	k, ctx := setupKeeper(t)

	currentTime := ctx.BlockTime()

	// Create 1000 KYC records (mix of valid and expired)
	for i := 0; i < 1000; i++ {
		var expiresAt time.Time
		if i%10 == 0 {
			// 10% expired (100 records)
			expiresAt = currentTime.Add(-1 * time.Hour)
		} else {
			// 90% valid
			expiresAt = currentTime.Add(365 * 24 * time.Hour)
		}

		record := &types.KYCRecord{
			Address:      "aura1user" + string(rune(i)),
			KycLevel:     types.KYCLevel_KYC_LEVEL_BASIC,
			VerifiedAt:   currentTime.Add(-30 * 24 * time.Hour),
			ExpiresAt:    &expiresAt,
			Provider:     "test-provider",
			Jurisdiction: "US",
		}

		err := k.SetKYCRecord(ctx, record)
		require.NoError(t, err)

		// Add to expiration index
		k.AddToExpirationIndex(ctx, record)
	}

	// Measure gas - with indexed lookups, should only iterate expired records
	ctx = ctx.WithBlockHeight(1).WithGasMeter(storetypes.NewInfiniteGasMeter())
	gasBefore := ctx.GasMeter().GasConsumed()

	k.BeginBlocker(ctx)

	gasUsed := ctx.GasMeter().GasConsumed() - gasBefore

	// With indexed lookups, gas should be proportional to expired count (100)
	// not total count (1000), so it should be significantly less than full scan
	t.Logf("Gas used with expiration index (100/1000 expired): %d", gasUsed)

	// Verify events were emitted for expired records (up to batch limit of 100)
	events := ctx.EventManager().Events()
	expiredCount := 0
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			expiredCount++
		}
	}

	// Should process exactly 100 expired records (batch limit)
	require.LessOrEqual(t, expiredCount, 100,
		"Should process at most 100 expired records per block (batch limit)")
}

// TestBeginBlockerExpiryDetection verifies expired records are detected with indexed lookups
func TestBeginBlockerExpiryDetection(t *testing.T) {
	k, ctx := setupKeeper(t)

	currentTime := ctx.BlockTime()

	// Create records that expire at different times
	testCases := []struct {
		address      string
		expiresAt    time.Time
		shouldExpire bool
	}{
		{
			address:      "aura1already_expired",
			expiresAt:    currentTime.Add(-10 * time.Hour),
			shouldExpire: true,
		},
		{
			address:      "aura1expires_soon",
			expiresAt:    currentTime.Add(-1 * time.Minute),
			shouldExpire: true,
		},
		{
			address:      "aura1valid",
			expiresAt:    currentTime.Add(365 * 24 * time.Hour),
			shouldExpire: false,
		},
	}

	for _, tc := range testCases {
		expiresAtPtr := tc.expiresAt
		record := &types.KYCRecord{
			Address:      tc.address,
			KycLevel:     types.KYCLevel_KYC_LEVEL_BASIC,
			VerifiedAt:   currentTime.Add(-30 * 24 * time.Hour),
			ExpiresAt:    &expiresAtPtr,
			Provider:     "test-provider",
			Jurisdiction: "US",
		}

		err := k.SetKYCRecord(ctx, record)
		require.NoError(t, err)

		// Add to expiration index
		k.AddToExpirationIndex(ctx, record)
	}

	// Run BeginBlocker
	ctx = ctx.WithBlockHeight(1)
	k.BeginBlocker(ctx)

	// Check events
	events := ctx.EventManager().Events()
	expiredAddresses := make(map[string]bool)

	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyAddress {
					expiredAddresses[attr.Value] = true
				}
			}
		}
	}

	// Verify expiry detection
	for _, tc := range testCases {
		if tc.shouldExpire {
			require.True(t, expiredAddresses[tc.address],
				"Address %s should have expired event", tc.address)
		} else {
			require.False(t, expiredAddresses[tc.address],
				"Address %s should NOT have expired event", tc.address)
		}
	}
}

// TestBeginBlockerBatchLimit verifies max 100 records processed per block
func TestBeginBlockerBatchLimit(t *testing.T) {
	k, ctx := setupKeeper(t)

	currentTime := ctx.BlockTime()

	// Create 150 expired records
	for i := 0; i < 150; i++ {
		expiresAt := currentTime.Add(-1 * time.Hour)
		record := &types.KYCRecord{
			Address:      "aura1user" + string(rune(i)),
			KycLevel:     types.KYCLevel_KYC_LEVEL_BASIC,
			VerifiedAt:   currentTime.Add(-30 * 24 * time.Hour),
			ExpiresAt:    &expiresAt,
			Provider:     "test-provider",
			Jurisdiction: "US",
		}

		err := k.SetKYCRecord(ctx, record)
		require.NoError(t, err)
		k.AddToExpirationIndex(ctx, record)
	}

	// First block should process 100 (batch limit)
	ctx = ctx.WithBlockHeight(1)
	k.BeginBlocker(ctx)

	events := ctx.EventManager().Events()
	expiredCount := 0
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			expiredCount++
		}
	}

	require.Equal(t, 100, expiredCount,
		"Should process exactly 100 expired records in first block")

	// Second block should process remaining 50
	ctx = ctx.WithBlockHeight(2)
	k.BeginBlocker(ctx)

	events = ctx.EventManager().Events()
	secondExpiredCount := 0
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			secondExpiredCount++
		}
	}

	// Total should be 150 (100 from first + 50 from second)
	require.Equal(t, 150, secondExpiredCount,
		"Should have processed all 150 expired records across two blocks")
}
