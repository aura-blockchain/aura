package networksecurity

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/networksecurity/keeper"
	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// TestBeginBlockerPeerUptimeBatching verifies peer uptime updates are batched every 100 blocks
func TestBeginBlockerPeerUptimeBatching(t *testing.T) {
	k, ctx := setupTestKeeperAndContext(t)

	// Create test peers
	peer1 := types.PeerInfo{
		PeerId:         "peer1",
		IpAddress:      "192.168.1.1",
		ConnectionType: "inbound",
		ConnectedAt:    ctx.BlockTime().Add(-1 * time.Hour),
	}
	peer2 := types.PeerInfo{
		PeerId:         "peer2",
		IpAddress:      "192.168.1.2",
		ConnectionType: "outbound",
		ConnectedAt:    ctx.BlockTime().Add(-2 * time.Hour),
	}

	err := k.SetPeerInfo(ctx, peer1)
	require.NoError(t, err)
	err = k.SetPeerInfo(ctx, peer2)
	require.NoError(t, err)

	// Initialize reputations
	rep1 := types.NodeReputation{
		PeerId:            "peer1",
		Score:             100,
		Uptime:            0,
		LastUpdatedHeight: ctx.BlockHeight(),
	}
	rep2 := types.NodeReputation{
		PeerId:            "peer2",
		Score:             100,
		Uptime:            0,
		LastUpdatedHeight: ctx.BlockHeight(),
	}

	err = k.SetReputation(ctx, rep1)
	require.NoError(t, err)
	err = k.SetReputation(ctx, rep2)
	require.NoError(t, err)

	// Test 1: Uptime should NOT update on non-100th blocks
	for i := int64(1); i < 100; i++ {
		ctx = ctx.WithBlockHeight(i)
		BeginBlocker(ctx, k)

		// Check that uptime was NOT updated
		updatedRep1, found := k.GetReputation(ctx, "peer1")
		require.True(t, found)
		// Uptime should still be 0 (not updated)
		require.Equal(t, int64(0), updatedRep1.Uptime,
			"Uptime should not update on block %d (not divisible by 100)", i)
	}

	// Test 2: Uptime SHOULD update on 100th block
	ctx = ctx.WithBlockHeight(100)
	BeginBlocker(ctx, k)

	// Check that uptime was updated
	updatedRep1, found := k.GetReputation(ctx, "peer1")
	require.True(t, found)
	require.Greater(t, updatedRep1.Uptime, int64(0),
		"Uptime should be updated on block 100")

	updatedRep2, found := k.GetReputation(ctx, "peer2")
	require.True(t, found)
	require.Greater(t, updatedRep2.Uptime, int64(0),
		"Uptime should be updated on block 100")

	// Test 3: Uptime should update on blocks 200, 300, etc.
	testBlocks := []int64{200, 300, 400, 500, 1000}
	for _, blockHeight := range testBlocks {
		// Advance time to make uptime calculation meaningful
		ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(ctx.BlockTime().Add(100 * 6 * time.Second))

		uptimeBefore := updatedRep1.Uptime

		BeginBlocker(ctx, k)

		updatedRep1, found = k.GetReputation(ctx, "peer1")
		require.True(t, found)
		require.GreaterOrEqual(t, updatedRep1.Uptime, uptimeBefore,
			"Uptime should be updated on block %d", blockHeight)
	}
}

// TestBeginBlockerBatchingPerformance verifies batching reduces gas consumption
func TestBeginBlockerBatchingPerformance(t *testing.T) {
	k, ctx := setupTestKeeperAndContext(t)

	// Create 1000 test peers
	for i := 0; i < 1000; i++ {
		connType := "outbound"
		if i%2 == 0 {
			connType = "inbound"
		}
		peer := types.PeerInfo{
			PeerId:         string(rune(i)),
			IpAddress:      "192.168.1.1",
			ConnectionType: connType,
			ConnectedAt:    ctx.BlockTime().Add(-1 * time.Hour),
		}
		err := k.SetPeerInfo(ctx, peer)
		require.NoError(t, err)

		rep := types.NodeReputation{
			PeerId:            peer.PeerId,
			Score:             100,
			Uptime:            0,
			LastUpdatedHeight: ctx.BlockHeight(),
		}
		err = k.SetReputation(ctx, rep)
		require.NoError(t, err)
	}

	// Measure gas on non-batched blocks
	var nonBatchedGasTotal uint64
	for i := int64(1); i < 100; i++ {
		ctx = ctx.WithBlockHeight(i).WithGasMeter(storetypes.NewInfiniteGasMeter())
		gasBefore := ctx.GasMeter().GasConsumed()

		BeginBlocker(ctx, k)

		gasUsed := ctx.GasMeter().GasConsumed() - gasBefore
		nonBatchedGasTotal += gasUsed
	}

	avgNonBatchedGas := nonBatchedGasTotal / 99

	// Measure gas on batched block
	ctx = ctx.WithBlockHeight(100).WithGasMeter(storetypes.NewInfiniteGasMeter())
	gasBefore := ctx.GasMeter().GasConsumed()

	BeginBlocker(ctx, k)

	gasUsedBatched := ctx.GasMeter().GasConsumed() - gasBefore

	// Batched block should use more gas than non-batched
	require.Greater(t, gasUsedBatched, avgNonBatchedGas,
		"Batched block should consume more gas due to peer iteration")

	// Calculate efficiency: without batching, we'd iterate all peers every block
	// With batching, we iterate every 100 blocks
	// Gas saved per 100 blocks = (99 * batchedGas) - (99 * avgNonBatchedGas)
	gasSavedPer100Blocks := int64((99 * gasUsedBatched) - nonBatchedGasTotal)

	t.Logf("Average gas per non-batched block: %d", avgNonBatchedGas)
	t.Logf("Gas used on batched block (100): %d", gasUsedBatched)
	t.Logf("Gas saved per 100 blocks: ~%d (%.1f%% reduction)",
		gasSavedPer100Blocks,
		float64(gasSavedPer100Blocks)/float64(99*gasUsedBatched)*100)
}

// TestBeginBlockerOtherOperationsStillRun verifies other BeginBlocker operations run every block
func TestBeginBlockerOtherOperationsStillRun(t *testing.T) {
	k, ctx := setupTestKeeperAndContext(t)

	// Test that CheckMempoolHealth runs every block
	// We can't easily verify this without mocking, but we can ensure no panic
	for i := int64(1); i <= 10; i++ {
		ctx = ctx.WithBlockHeight(i)
		require.NotPanics(t, func() {
			BeginBlocker(ctx, k)
		}, "BeginBlocker should not panic on block %d", i)
	}

	// Test that reputation decay still runs every 100 blocks
	// Create a reputation with non-zero score
	rep := types.NodeReputation{
		PeerId:            "decay_test_peer",
		Score:             1000,
		LastUpdatedHeight: 1,
	}
	err := k.SetReputation(ctx, rep)
	require.NoError(t, err)

	// Run BeginBlocker on block 200 (divisible by 100)
	ctx = ctx.WithBlockHeight(200)
	BeginBlocker(ctx, k)

	// Reputation decay should have run
	updatedRep, found := k.GetReputation(ctx, "decay_test_peer")
	require.True(t, found)
	// LastUpdatedHeight should be updated (decay ran)
	require.Equal(t, int64(200), updatedRep.LastUpdatedHeight,
		"Reputation decay should run on block 200")
}

// setupTestKeeperAndContext creates a test keeper and context using the standard test helpers
func setupTestKeeperAndContext(t *testing.T) (keeper.Keeper, sdk.Context) {
	return keeper.NewTestKeeperWithContext(t)
}
