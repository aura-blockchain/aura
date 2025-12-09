package networksecurity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/keeper"
)

// BeginBlocker performs begin block operations for the network security module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	// 1. Perform network health checks
	k.PerformNetworkHealthCheck(ctx)

	// 2. Decay reputation scores
	if ctx.BlockHeight()%100 == 0 { // Every 100 blocks
		k.DecayReputations(ctx)
	}

	// 3. Cleanup expired rate limits and bans
	if ctx.BlockHeight()%50 == 0 { // Every 50 blocks
		k.CleanupExpiredRateLimits(ctx)
	}

	// 4. Cleanup message cache
	if ctx.BlockHeight()%200 == 0 { // Every 200 blocks
		k.CleanupMessageCache(ctx)
	}

	// 5. Update peer uptimes (batched every 100 blocks to reduce gas)
	// Uptime is a derived metric that doesn't need per-block precision
	if ctx.BlockHeight()%100 == 0 {
		peers := k.GetAllPeers(ctx)
		for _, peer := range peers {
			k.UpdatePeerUptime(ctx, peer.PeerId)
		}
	}

	// 6. Check mempool health
	k.CheckMempoolHealth(ctx)

	// 7. Cleanup resolved alerts
	if ctx.BlockHeight()%1000 == 0 { // Every 1000 blocks
		k.CleanupResolvedAlerts(ctx)
	}

	// 8. Prune low reputation peers
	if ctx.BlockHeight()%500 == 0 { // Every 500 blocks
		k.PruneLowReputationPeers(ctx)
	}

	// 9. Update known peer list for partition detection
	if ctx.BlockHeight()%100 == 0 { // Every 100 blocks
		k.UpdateKnownPeerList(ctx)
	}
}

// EndBlocker performs end block operations for the network security module
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Cleanup mempool if needed
	k.CleanupMempool(ctx)
}
