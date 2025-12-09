package networksecurity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/keeper"
)

// BeginBlocker performs begin block operations for the network security module
// with rate-limited batch processing to prevent consensus timeouts
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Rate-limited operations with configurable batch sizes
	// This ensures BeginBlocker completes in <200ms under any load

	// 1. Update threat metrics in batches (every block, limited to MAX_THREAT_UPDATES_PER_BLOCK)
	processedThreats := k.UpdateThreatMetricsBatched(ctx, keeper.MAX_THREAT_UPDATES_PER_BLOCK)
	if processedThreats > 0 {
		k.Logger(ctx).Debug("processed threat metrics", "count", processedThreats)
	}

	// 2. Process security alerts in batches (every block, limited to MAX_ALERTS_PER_BLOCK)
	processedAlerts := k.ProcessSecurityAlertsBatched(ctx, keeper.MAX_ALERTS_PER_BLOCK)
	if processedAlerts > 0 {
		k.Logger(ctx).Debug("processed security alerts", "count", processedAlerts)
	}

	// 3. Refresh reputation scores in batches (every REPUTATION_REFRESH_INTERVAL blocks)
	// This includes decay and uptime updates
	if ctx.BlockHeight()%keeper.REPUTATION_REFRESH_INTERVAL == 0 {
		processedReps := k.RefreshReputationScoresBatched(ctx, keeper.MAX_THREAT_UPDATES_PER_BLOCK)
		if processedReps > 0 {
			k.Logger(ctx).Debug("refreshed reputation scores", "count", processedReps)
		}
	}

	// 4. Cleanup expired rate limits and bans (every 50 blocks, already lightweight)
	if ctx.BlockHeight()%50 == 0 {
		k.CleanupExpiredRateLimits(ctx)
	}

	// 5. Cleanup message cache (every 200 blocks, already lightweight)
	if ctx.BlockHeight()%200 == 0 {
		k.CleanupMessageCache(ctx)
	}

	// 6. Check mempool health (every block, lightweight - just reads stats)
	k.CheckMempoolHealth(ctx)

	// 7. Cleanup resolved alerts (every 1000 blocks, infrequent)
	if ctx.BlockHeight()%1000 == 0 {
		k.CleanupResolvedAlerts(ctx)
	}

	// 8. Prune low reputation peers (batched, every 500 blocks)
	if ctx.BlockHeight()%500 == 0 {
		prunedCount := k.PruneLowReputationPeersBatched(ctx, 10) // Limit to 10 peers per batch
		if prunedCount > 0 {
			k.Logger(ctx).Info("pruned low reputation peers", "count", prunedCount)
		}
	}

	// 9. Update known peer list for partition detection (every 100 blocks, optimized)
	if ctx.BlockHeight()%100 == 0 {
		if err := k.UpdateKnownPeerListBatched(ctx); err != nil {
			k.Logger(ctx).Error("failed to update known peer list", "error", err)
		}
	}

	// 10. Perform lightweight network health checks (already optimized, no unbounded loops)
	// This only checks current state, doesn't iterate all entities
	k.CheckMempoolHealth(ctx)

	// Note: Heavy operations like PerformNetworkHealthCheck have been decomposed:
	// - Partition detection: handled by known peer list updates (every 100 blocks)
	// - Sybil/Eclipse checks: run on-demand when new peers connect (in ValidateNewConnection)
	// - Peer diversity: checked on-demand, not in hot path
	//
	// This ensures BeginBlocker stays under 200ms even with thousands of peers/alerts
}

// EndBlocker performs end block operations for the network security module
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Cleanup mempool if needed
	k.CleanupMempool(ctx)
}
