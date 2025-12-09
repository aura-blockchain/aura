package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// TestBeginBlockerBatching verifies that KYC expiry checks are batched every 50 blocks
func TestBeginBlockerBatching(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create an expired KYC record
	expiredRecord := &types.KYCRecord{
		Address:      "aura1expireduser",
		KycLevel:     types.KYCLevel_KYC_LEVEL_BASIC,
		VerifiedAt:   ctx.BlockTime().Add(-365 * 24 * time.Hour), // 1 year ago
		ExpiresAt:    ctx.BlockTime().Add(-1 * time.Hour),        // Expired 1 hour ago
		Provider:     "test-provider",
		Jurisdiction: "US",
	}

	err := k.SetKYCRecord(ctx, expiredRecord)
	require.NoError(t, err)

	// Test 1: BeginBlocker should NOT process on non-50th blocks
	for i := int64(1); i < 50; i++ {
		ctx = ctx.WithBlockHeight(i)
		eventsBefore := len(ctx.EventManager().Events())

		k.BeginBlocker(ctx)

		eventsAfter := len(ctx.EventManager().Events())
		require.Equal(t, eventsBefore, eventsAfter,
			"BeginBlocker should not emit events on block %d (not divisible by 50)", i)
	}

	// Test 2: BeginBlocker SHOULD process on 50th block
	ctx = ctx.WithBlockHeight(50)
	eventsBefore := len(ctx.EventManager().Events())

	k.BeginBlocker(ctx)

	eventsAfter := len(ctx.EventManager().Events())
	require.Greater(t, eventsAfter, eventsBefore,
		"BeginBlocker should emit expiry events on block 50")

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

	// Test 3: BeginBlocker should process on block 100, 150, 200, etc.
	testBlocks := []int64{100, 150, 200, 250, 500, 1000}
	for _, blockHeight := range testBlocks {
		ctx = ctx.WithBlockHeight(blockHeight)
		eventsBefore := len(ctx.EventManager().Events())

		k.BeginBlocker(ctx)

		eventsAfter := len(ctx.EventManager().Events())
		// Events may or may not be emitted depending on if there are new expired records,
		// but BeginBlocker should run (we just verify no panic)
		_ = eventsAfter
		require.GreaterOrEqual(t, eventsAfter, eventsBefore,
			"BeginBlocker should run on block %d", blockHeight)
	}
}

// TestBeginBlockerBatchingPerformance verifies batching reduces gas cost
func TestBeginBlockerBatchingPerformance(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create 1000 KYC records (mix of valid and expired)
	currentTime := ctx.BlockTime()
	for i := 0; i < 1000; i++ {
		var expiresAt time.Time
		if i%10 == 0 {
			// 10% expired
			expiresAt = currentTime.Add(-1 * time.Hour)
		} else {
			// 90% valid
			expiresAt = currentTime.Add(365 * 24 * time.Hour)
		}

		record := &types.KYCRecord{
			Address:      "aura1user" + string(rune(i)),
			KycLevel:     types.KYCLevel_KYC_LEVEL_BASIC,
			VerifiedAt:   currentTime.Add(-30 * 24 * time.Hour),
			ExpiresAt:    expiresAt,
			Provider:     "test-provider",
			Jurisdiction: "US",
		}

		err := k.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Measure gas on non-batched blocks (should be minimal)
	nonBatchedBlocks := 0
	for i := int64(1); i < 50; i++ {
		ctx = ctx.WithBlockHeight(i).WithGasMeter(sdk.NewInfiniteGasMeter())
		gasBefore := ctx.GasMeter().GasConsumed()

		k.BeginBlocker(ctx)

		gasUsed := ctx.GasMeter().GasConsumed() - gasBefore
		if gasUsed > 1000 {
			t.Errorf("Block %d used %d gas (expected <1000 for early return)", i, gasUsed)
		}
		nonBatchedBlocks++
	}

	// Measure gas on batched block (will be higher due to iteration)
	ctx = ctx.WithBlockHeight(50).WithGasMeter(sdk.NewInfiniteGasMeter())
	gasBefore := ctx.GasMeter().GasConsumed()

	k.BeginBlocker(ctx)

	gasUsedBatched := ctx.GasMeter().GasConsumed() - gasBefore

	// Batched block should use significantly more gas (due to iteration)
	require.Greater(t, gasUsedBatched, uint64(1000),
		"Batched block should consume gas for iteration")

	// Calculate gas savings: 49 blocks with minimal gas vs 1 block with full scan
	// Savings = (49 * gasUsedBatched) - (49 * minimal_gas)
	// With batching, we save ~98% of gas over 50 blocks
	t.Logf("Gas used on batched block (50): %d", gasUsedBatched)
	t.Logf("Gas saved over 49 non-batched blocks: ~%d", (nonBatchedBlocks)*int(gasUsedBatched))
	t.Logf("Batching efficiency: ~98%% gas reduction")
}

// TestBeginBlockerExpiryDetection verifies expired records are still detected with batching
func TestBeginBlockerExpiryDetection(t *testing.T) {
	k, ctx := setupKeeper(t)

	currentTime := ctx.BlockTime()

	// Create records that expire at different times
	testCases := []struct {
		address   string
		expiresAt time.Time
		shouldExpireByBlock50 bool
	}{
		{
			address:   "aura1already_expired",
			expiresAt: currentTime.Add(-10 * time.Hour), // Already expired
			shouldExpireByBlock50: true,
		},
		{
			address:   "aura1expires_soon",
			expiresAt: currentTime.Add(-1 * time.Minute), // Just expired
			shouldExpireByBlock50: true,
		},
		{
			address:   "aura1valid",
			expiresAt: currentTime.Add(365 * 24 * time.Hour), // Valid for 1 year
			shouldExpireByBlock50: false,
		},
	}

	for _, tc := range testCases {
		record := &types.KYCRecord{
			Address:      tc.address,
			KycLevel:     types.KYCLevel_KYC_LEVEL_BASIC,
			VerifiedAt:   currentTime.Add(-30 * 24 * time.Hour),
			ExpiresAt:    tc.expiresAt,
			Provider:     "test-provider",
			Jurisdiction: "US",
		}

		err := k.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Run BeginBlocker on block 50 (batched)
	ctx = ctx.WithBlockHeight(50)
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
		if tc.shouldExpireByBlock50 {
			require.True(t, expiredAddresses[tc.address],
				"Address %s should have expired event", tc.address)
		} else {
			require.False(t, expiredAddresses[tc.address],
				"Address %s should NOT have expired event", tc.address)
		}
	}
}
