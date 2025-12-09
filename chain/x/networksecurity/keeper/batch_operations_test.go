package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/networksecurity/keeper"
	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// TestUpdateThreatMetricsBatched verifies threat metric updates are batched correctly
func TestUpdateThreatMetricsBatched(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create test rate limit entries
	params := types.DefaultParams()
	k.SetParams(ctx, *params)

	entries := []types.RateLimitEntry{
		{
			PeerId:       "peer1",
			RequestCount: 100,
			WindowStart:  ctx.BlockTime().Add(-2 * params.RateLimit.WindowDuration),
		},
		{
			PeerId:       "peer2",
			RequestCount: 50,
			WindowStart:  ctx.BlockTime().Add(-2 * params.RateLimit.WindowDuration),
		},
		{
			PeerId:       "peer3",
			RequestCount: 75,
			WindowStart:  ctx.BlockTime().Add(-2 * params.RateLimit.WindowDuration),
		},
		{
			PeerId:       "peer4",
			RequestCount: 25,
			WindowStart:  ctx.BlockTime().Add(-2 * params.RateLimit.WindowDuration),
		},
		{
			PeerId:       "peer5",
			RequestCount: 10,
			WindowStart:  ctx.BlockTime().Add(-2 * params.RateLimit.WindowDuration),
		},
	}

	for _, entry := range entries {
		k.SetRateLimitEntry(ctx, entry)
	}

	// Process first batch (limit 2)
	processed := k.UpdateThreatMetricsBatched(ctx, 2)
	require.Equal(t, 2, processed, "should process exactly 2 entries")

	// Verify cursor moved
	cursor, err := k.GetBatchCursor(ctx, types.ThreatUpdateCursorKey)
	require.NoError(t, err)
	require.Equal(t, uint64(2), cursor, "cursor should be at position 2")

	// Process second batch (limit 2)
	processed = k.UpdateThreatMetricsBatched(ctx, 2)
	require.Equal(t, 2, processed, "should process 2 more entries")

	// Verify cursor moved
	cursor, err = k.GetBatchCursor(ctx, types.ThreatUpdateCursorKey)
	require.NoError(t, err)
	require.Equal(t, uint64(4), cursor, "cursor should be at position 4")

	// Process remaining (limit 2, but only 1 left)
	processed = k.UpdateThreatMetricsBatched(ctx, 2)
	require.Equal(t, 1, processed, "should process last entry")

	// Verify cursor reset
	cursor, err = k.GetBatchCursor(ctx, types.ThreatUpdateCursorKey)
	require.NoError(t, err)
	require.Equal(t, uint64(0), cursor, "cursor should reset to 0 after completing all entries")
}

// TestProcessSecurityAlertsBatched verifies security alerts are processed in batches
func TestProcessSecurityAlertsBatched(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.DefaultParams()
	params.ForkDetection.EnableAutoResolution = true
	k.SetParams(ctx, *params)

	// Create test fork alerts
	for i := 0; i < 5; i++ {
		alert := types.ForkAlert{
			AlertId:     generateTestAlertID("fork", int64(i)),
			BlockHeight: int64(100 + i),
			ChainAHash:  []byte("hashA"),
			ChainBHash:  []byte("hashB"),
			DetectedAt:  ctx.BlockTime(),
			Resolved:    false,
		}
		k.SetForkAlert(ctx, alert)
	}

	// Create test partition alerts
	for i := 0; i < 3; i++ {
		alert := types.PartitionAlert{
			AlertId:        generateTestAlertID("partition", int64(i)),
			ConnectedPeers: 2,
			ExpectedPeers:  10,
			MissingPeerIds: []string{"peer1", "peer2"},
			DetectedAt:     ctx.BlockTime(),
			Resolved:       false,
		}
		k.SetPartitionAlert(ctx, alert)
	}

	// Process first batch (limit 3)
	processed := k.ProcessSecurityAlertsBatched(ctx, 3)
	require.Equal(t, 3, processed, "should process 3 fork alerts")

	// Process second batch (limit 3)
	processed = k.ProcessSecurityAlertsBatched(ctx, 3)
	require.Equal(t, 2, processed, "should process remaining 2 fork alerts")

	// Process third batch (should process partition alerts)
	processed = k.ProcessSecurityAlertsBatched(ctx, 3)
	require.Equal(t, 3, processed, "should process 3 partition alerts")
}

// TestRefreshReputationScoresBatched verifies reputation scores are refreshed in batches
func TestRefreshReputationScoresBatched(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.DefaultParams()
	params.Reputation.EnableTracking = true
	params.Reputation.DecayRate = 5
	k.SetParams(ctx, *params)

	// Create test reputations with old timestamps
	oldHeight := ctx.BlockHeight() - keeper.REPUTATION_REFRESH_INTERVAL - 10

	reputations := []types.NodeReputation{
		{
			PeerId:            "peer1",
			Score:             100,
			LastUpdatedHeight: oldHeight,
		},
		{
			PeerId:            "peer2",
			Score:             80,
			LastUpdatedHeight: oldHeight,
		},
		{
			PeerId:            "peer3",
			Score:             60,
			LastUpdatedHeight: oldHeight,
		},
		{
			PeerId:            "peer4",
			Score:             40,
			LastUpdatedHeight: oldHeight,
		},
	}

	for _, rep := range reputations {
		k.SetReputation(ctx, rep)
	}

	// Process first batch (limit 2)
	processed := k.RefreshReputationScoresBatched(ctx, 2)
	require.Equal(t, 2, processed, "should process 2 reputations")

	// Verify scores were decayed
	rep1, found := k.GetReputation(ctx, "peer1")
	require.True(t, found)
	require.Equal(t, int64(95), rep1.Score, "score should be decayed by 5")

	rep2, found := k.GetReputation(ctx, "peer2")
	require.True(t, found)
	require.Equal(t, int64(75), rep2.Score, "score should be decayed by 5")

	// Process remaining batch
	processed = k.RefreshReputationScoresBatched(ctx, 10)
	require.Equal(t, 2, processed, "should process remaining 2 reputations")

	// Verify all scores were decayed
	rep3, found := k.GetReputation(ctx, "peer3")
	require.True(t, found)
	require.Equal(t, int64(55), rep3.Score)

	rep4, found := k.GetReputation(ctx, "peer4")
	require.True(t, found)
	require.Equal(t, int64(35), rep4.Score)
}

// TestRefreshReputationScoresBatchedWithUptime verifies uptime is updated
func TestRefreshReputationScoresBatchedWithUptime(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.DefaultParams()
	params.Reputation.EnableTracking = true
	k.SetParams(ctx, *params)

	// Create peer info with connection time
	peerInfo := types.PeerInfo{
		PeerId:      "peer1",
		IpAddress:   "192.168.1.1",
		ConnectedAt: ctx.BlockTime().Add(-1 * time.Hour),
	}
	k.SetPeerInfo(ctx, peerInfo)

	// Create reputation
	rep := types.NodeReputation{
		PeerId:            "peer1",
		Score:             100,
		Uptime:            0,
		LastUpdatedHeight: ctx.BlockHeight() - keeper.REPUTATION_REFRESH_INTERVAL - 1,
	}
	k.SetReputation(ctx, rep)

	// Process batch
	processed := k.RefreshReputationScoresBatched(ctx, 10)
	require.Equal(t, 1, processed)

	// Verify uptime was updated
	updatedRep, found := k.GetReputation(ctx, "peer1")
	require.True(t, found)
	require.Greater(t, updatedRep.Uptime, int64(0), "uptime should be updated")
	require.Equal(t, int64(3600), updatedRep.Uptime, "uptime should be 1 hour in seconds")
}

// TestPruneLowReputationPeersBatched verifies low reputation peers are pruned in batches
func TestPruneLowReputationPeersBatched(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.DefaultParams()
	k.SetParams(ctx, *params)

	// Create peers with low reputation and high misbehavior
	for i := 0; i < 5; i++ {
		peerID := generateTestPeerID(i)

		// Create peer info
		peerInfo := types.PeerInfo{
			PeerId:      peerID,
			IpAddress:   generateTestIP(i),
			ConnectedAt: ctx.BlockTime(),
		}
		k.SetPeerInfo(ctx, peerInfo)

		// Create reputation
		rep := types.NodeReputation{
			PeerId:            peerID,
			Score:             0,
			MisbehaviorCount:  15,
			LastUpdatedHeight: ctx.BlockHeight(),
		}
		k.SetReputation(ctx, rep)
	}

	// Prune first batch (limit 2)
	pruned := k.PruneLowReputationPeersBatched(ctx, 2)
	require.Equal(t, 2, pruned, "should prune 2 peers")

	// Verify peers are banned
	require.True(t, k.IsBanned(ctx, generateTestPeerID(0)))
	require.True(t, k.IsBanned(ctx, generateTestPeerID(1)))

	// Prune remaining
	pruned = k.PruneLowReputationPeersBatched(ctx, 10)
	require.Equal(t, 3, pruned, "should prune remaining 3 peers")
}

// TestUpdateKnownPeerListBatched verifies known peer list is updated efficiently
func TestUpdateKnownPeerListBatched(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create test peers
	for i := 0; i < 10; i++ {
		peerInfo := types.PeerInfo{
			PeerId:      generateTestPeerID(i),
			IpAddress:   generateTestIP(i),
			ConnectedAt: ctx.BlockTime(),
		}
		k.SetPeerInfo(ctx, peerInfo)
	}

	// Update known peer list
	err := k.UpdateKnownPeerListBatched(ctx)
	require.NoError(t, err)

	// Verify peer count was stored
	store := k.KVStoreService().OpenKVStore(ctx)
	countBz, err := store.Get([]byte("known_peer_count"))
	require.NoError(t, err)
	require.NotNil(t, countBz)

	// Verify peer hash was stored
	hashBz, err := store.Get([]byte("known_peer_hash"))
	require.NoError(t, err)
	require.NotNil(t, hashBz)
}

// TestBatchOperationsUnderLoad verifies batch operations handle high load
func TestBatchOperationsUnderLoad(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.DefaultParams()
	params.Reputation.EnableTracking = true
	k.SetParams(ctx, *params)

	// Create many entries to simulate high load
	numEntries := 1000

	// Create rate limit entries
	for i := 0; i < numEntries; i++ {
		entry := types.RateLimitEntry{
			PeerId:       generateTestPeerID(i),
			RequestCount: uint64(i),
			WindowStart:  ctx.BlockTime().Add(-2 * params.RateLimit.WindowDuration),
		}
		k.SetRateLimitEntry(ctx, entry)
	}

	// Create reputations
	for i := 0; i < numEntries; i++ {
		rep := types.NodeReputation{
			PeerId:            generateTestPeerID(i),
			Score:             100,
			LastUpdatedHeight: ctx.BlockHeight() - keeper.REPUTATION_REFRESH_INTERVAL - 1,
		}
		k.SetReputation(ctx, rep)
	}

	// Process multiple batches to ensure progress
	totalProcessed := 0
	maxIterations := 30 // Should process 1000 entries with batches of ~50

	for i := 0; i < maxIterations; i++ {
		processed := k.UpdateThreatMetricsBatched(ctx, keeper.MAX_THREAT_UPDATES_PER_BLOCK)
		totalProcessed += processed
		if processed == 0 {
			break // Completed all entries
		}
	}

	require.Greater(t, totalProcessed, 0, "should process entries")
	require.LessOrEqual(t, totalProcessed, numEntries, "should not process more than available")

	// Verify cursor was managed correctly
	cursor, err := k.GetBatchCursor(ctx, types.ThreatUpdateCursorKey)
	require.NoError(t, err)
	require.Equal(t, uint64(0), cursor, "cursor should reset after processing all entries")
}

// TestBatchOperationsEmptyState verifies batch operations handle empty state
func TestBatchOperationsEmptyState(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.DefaultParams()
	k.SetParams(ctx, *params)

	// Test with no data
	processed := k.UpdateThreatMetricsBatched(ctx, 10)
	require.Equal(t, 0, processed, "should process 0 entries when no data exists")

	processed = k.ProcessSecurityAlertsBatched(ctx, 10)
	require.Equal(t, 0, processed, "should process 0 alerts when no data exists")

	processed = k.RefreshReputationScoresBatched(ctx, 10)
	require.Equal(t, 0, processed, "should process 0 reputations when no data exists")

	pruned := k.PruneLowReputationPeersBatched(ctx, 10)
	require.Equal(t, 0, pruned, "should prune 0 peers when no data exists")
}

// TestBatchCursorPersistence verifies cursors persist across operations
func TestBatchCursorPersistence(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set cursor
	err := k.SetBatchCursor(ctx, types.ThreatUpdateCursorKey, 42)
	require.NoError(t, err)

	// Retrieve cursor
	cursor, err := k.GetBatchCursor(ctx, types.ThreatUpdateCursorKey)
	require.NoError(t, err)
	require.Equal(t, uint64(42), cursor)

	// Update cursor
	err = k.SetBatchCursor(ctx, types.ThreatUpdateCursorKey, 100)
	require.NoError(t, err)

	// Verify update
	cursor, err = k.GetBatchCursor(ctx, types.ThreatUpdateCursorKey)
	require.NoError(t, err)
	require.Equal(t, uint64(100), cursor)
}

// Helper functions

func generateTestAlertID(alertType string, index int64) string {
	return alertType + string(rune(index))
}

func generateTestPeerID(index int) string {
	return "peer" + string(rune('0'+index))
}

func generateTestIP(index int) string {
	return "192.168.1." + string(rune('0'+index))
}
