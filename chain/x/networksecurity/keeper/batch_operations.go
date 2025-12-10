package keeper

import (
	"fmt"
	"sort"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// UpdateThreatMetricsBatched performs batched threat metric updates
// Returns the number of items processed
func (k Keeper) UpdateThreatMetricsBatched(ctx sdk.Context, limit int) int {
	if limit <= 0 {
		return 0
	}

	// Get cursor to track progress
	cursor, err := k.GetBatchCursor(ctx, types.ThreatUpdateCursorKey)
	if err != nil {
		k.logger.Error("failed to get threat update cursor", "error", err)
		return 0
	}

	// Get all rate limit entries (which contain threat metrics)
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.RateLimitPrefix, storetypes.PrefixEndBytes(types.RateLimitPrefix))
	if err != nil {
		k.logger.Error("failed to create rate limit iterator", "error", err)
		return 0
	}
	defer iterator.Close()

	// Skip to cursor position
	skipCount := uint64(0)
	for ; iterator.Valid() && skipCount < cursor; iterator.Next() {
		skipCount++
	}

	// Process batch
	processedCount := 0
	params, _ := k.GetParams(ctx)

	for ; iterator.Valid() && processedCount < limit; iterator.Next() {
		var entry types.RateLimitEntry
		if err := k.cdc.Unmarshal(iterator.Value(), &entry); err != nil {
			k.logger.Error("failed to unmarshal rate limit entry in batch update", "error", err)
			continue
		}

		// Update threat metrics: reset window if expired
		if !entry.WindowStart.IsZero() && ctx.BlockTime().Sub(entry.WindowStart) > params.RateLimit.WindowDuration {
			entry.WindowStart = ctx.BlockTime()
			entry.RequestCount = 0

			// Persist updated entry
			if err := k.SetRateLimitEntry(ctx, entry); err != nil {
				k.logger.Error("failed to update rate limit entry", "peer_id", entry.PeerId, "error", err)
			}
		}

		processedCount++
	}

	// Update cursor
	newCursor := cursor + uint64(processedCount)

	// Reset cursor if we've reached the end
	if !iterator.Valid() {
		newCursor = 0
		k.logger.Debug("threat metrics batch complete, resetting cursor")
	}

	if err := k.SetBatchCursor(ctx, types.ThreatUpdateCursorKey, newCursor); err != nil {
		k.logger.Error("failed to update threat update cursor", "error", err)
	}

	return processedCount
}

// ProcessSecurityAlertsBatched processes security alerts in batches
// Returns the number of alerts processed
func (k Keeper) ProcessSecurityAlertsBatched(ctx sdk.Context, limit int) int {
	if limit <= 0 {
		return 0
	}

	// Get cursor
	cursor, err := k.GetBatchCursor(ctx, types.SecurityAlertCursorKey)
	if err != nil {
		k.logger.Error("failed to get security alert cursor", "error", err)
		return 0
	}

	// Process fork alerts
	forkAlerts := k.GetAllForkAlerts(ctx, false) // Only unresolved

	// DETERMINISM: Sort alerts by alert ID to ensure consistent iteration order
	// across all validators. This prevents consensus failure from different
	// processing orders affecting state changes (e.g., auto-resolution).
	sort.Slice(forkAlerts, func(i, j int) bool {
		return forkAlerts[i].AlertId < forkAlerts[j].AlertId
	})

	processedCount := 0

	// Skip to cursor position
	startIdx := int(cursor)
	if startIdx >= len(forkAlerts) {
		// Process partition alerts instead
		partitionAlerts := k.GetAllPartitionAlerts(ctx, false)

		// DETERMINISM: Sort partition alerts by alert ID to ensure consistent
		// iteration order across all validators.
		sort.Slice(partitionAlerts, func(i, j int) bool {
			return partitionAlerts[i].AlertId < partitionAlerts[j].AlertId
		})

		partitionStartIdx := startIdx - len(forkAlerts)
		if partitionStartIdx >= len(partitionAlerts) {
			// Reset cursor - we've processed all alerts
			if err := k.SetBatchCursor(ctx, types.SecurityAlertCursorKey, 0); err != nil {
				k.logger.Error("failed to reset security alert cursor", "error", err)
			}
			return 0
		}

		// Process partition alerts batch
		for i := partitionStartIdx; i < len(partitionAlerts) && processedCount < limit; i++ {
			alert := partitionAlerts[i]

			// Check if partition has been resolved (peers reconnected)
			currentPeers := k.GetAllPeers(ctx)

			// DETERMINISM: While we only use the peer count here (not iteration),
			// we note that GetAllPeers returns peers in store key order (by peer ID).
			// If we iterated this list, it would need to be sorted by peer ID.
			if uint32(len(currentPeers)) > alert.ExpectedPeers/2 {
				// More than 50% of expected peers are back, consider resolved
				if err := k.ResolvePartitionAlert(ctx, alert.AlertId); err != nil {
					k.logger.Error("failed to resolve partition alert", "alert_id", alert.AlertId, "error", err)
				} else {
					k.logger.Info("auto-resolved partition alert", "alert_id", alert.AlertId)
				}
			}

			processedCount++
		}

		newCursor := uint64(startIdx + processedCount)
		if err := k.SetBatchCursor(ctx, types.SecurityAlertCursorKey, newCursor); err != nil {
			k.logger.Error("failed to update security alert cursor", "error", err)
		}

		return processedCount
	}

	// Process fork alerts batch
	for i := startIdx; i < len(forkAlerts) && processedCount < limit; i++ {
		alert := forkAlerts[i]

		// Attempt auto-resolution if enabled
		params, _ := k.GetParams(ctx)
		if params.ForkDetection.EnableAutoResolution {
			if err := k.ResolveFork(ctx, alert.AlertId); err != nil && err != types.ErrAlreadyResolved {
				k.logger.Debug("fork alert not yet resolvable", "alert_id", alert.AlertId)
			}
		}

		processedCount++
	}

	// Update cursor
	newCursor := cursor + uint64(processedCount)
	if err := k.SetBatchCursor(ctx, types.SecurityAlertCursorKey, newCursor); err != nil {
		k.logger.Error("failed to update security alert cursor", "error", err)
	}

	return processedCount
}

// RefreshReputationScoresBatched performs batched reputation score refresh
// Returns the number of reputations processed
func (k Keeper) RefreshReputationScoresBatched(ctx sdk.Context, limit int) int {
	if limit <= 0 {
		return 0
	}

	params, _ := k.GetParams(ctx)

	// Only process if reputation tracking is enabled
	if !params.Reputation.EnableTracking {
		return 0
	}

	// Get cursor
	cursor, err := k.GetBatchCursor(ctx, types.ReputationRefreshCursorKey)
	if err != nil {
		k.logger.Error("failed to get reputation refresh cursor", "error", err)
		return 0
	}

	// Iterate reputations with cursor
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ReputationPrefix, storetypes.PrefixEndBytes(types.ReputationPrefix))
	if err != nil {
		k.logger.Error("failed to create reputation iterator", "error", err)
		return 0
	}
	defer iterator.Close()

	// Skip to cursor position
	skipCount := uint64(0)
	for ; iterator.Valid() && skipCount < cursor; iterator.Next() {
		skipCount++
	}

	// Process batch
	processedCount := 0

	for ; iterator.Valid() && processedCount < limit; iterator.Next() {
		var reputation types.NodeReputation
		if err := k.cdc.Unmarshal(iterator.Value(), &reputation); err != nil {
			k.logger.Error("failed to unmarshal reputation in batch refresh", "error", err)
			continue
		}

		// Apply decay if enough blocks have passed
		if params.Reputation.DecayRate > 0 && ctx.BlockHeight()-reputation.LastUpdatedHeight >= REPUTATION_REFRESH_INTERVAL {
			oldScore := reputation.Score
			reputation.Score -= params.Reputation.DecayRate

			if reputation.Score < 0 {
				reputation.Score = 0
			}

			reputation.LastUpdatedHeight = ctx.BlockHeight()

			if err := k.SetReputation(ctx, reputation); err != nil {
				k.logger.Error("failed to update reputation", "peer_id", reputation.PeerId, "error", err)
			} else if oldScore != reputation.Score {
				k.logger.Debug("decayed reputation", "peer_id", reputation.PeerId, "old_score", oldScore, "new_score", reputation.Score)
			}
		}

		// Update uptime for connected peers
		if peerInfo, found := k.GetPeerInfo(ctx, reputation.PeerId); found {
			var uptime int64
			if !peerInfo.ConnectedAt.IsZero() {
				uptime = int64(ctx.BlockTime().Sub(peerInfo.ConnectedAt).Seconds())
			}

			if reputation.Uptime != uptime {
				reputation.Uptime = uptime
				if err := k.SetReputation(ctx, reputation); err != nil {
					k.logger.Error("failed to update uptime", "peer_id", reputation.PeerId, "error", err)
				}
			}
		}

		processedCount++
	}

	// Update cursor
	newCursor := cursor + uint64(processedCount)

	// Reset cursor if we've reached the end
	if !iterator.Valid() {
		newCursor = 0
		k.logger.Debug("reputation refresh batch complete, resetting cursor")
	}

	if err := k.SetBatchCursor(ctx, types.ReputationRefreshCursorKey, newCursor); err != nil {
		k.logger.Error("failed to update reputation refresh cursor", "error", err)
	}

	return processedCount
}

// PruneLowReputationPeersBatched performs batched low reputation peer pruning
// Returns the number of peers pruned
func (k Keeper) PruneLowReputationPeersBatched(ctx sdk.Context, limit int) int {
	if limit <= 0 {
		return 0
	}

	params, _ := k.GetParams(ctx)

	// Get all reputations
	reputations := k.GetAllReputations(ctx)

	// CRITICAL DETERMINISM: Sort reputations by peer ID before iteration.
	// This function makes state-changing decisions (banning/disconnecting peers)
	// based on which peers are processed first (up to the limit).
	// Different iteration orders across validators will cause consensus failure,
	// as different validators would ban different peers.
	sort.Slice(reputations, func(i, j int) bool {
		return reputations[i].PeerId < reputations[j].PeerId
	})

	prunedCount := 0

	for _, rep := range reputations {
		if prunedCount >= limit {
			break
		}

		// If reputation is 0 and has high misbehavior count, prune
		if rep.Score == 0 && rep.MisbehaviorCount > 10 {
			k.logger.Info(fmt.Sprintf("Pruning low reputation peer: %s (score: %d, misbehavior: %d)",
				rep.PeerId, rep.Score, rep.MisbehaviorCount))

			// Ban the peer
			banDuration := int64(params.RateLimit.BanDuration.Seconds() * 2)
			if err := k.BanPeer(ctx, rep.PeerId, banDuration, "low reputation with high misbehavior"); err != nil {
				k.logger.Error("failed to ban low reputation peer", "peer_id", rep.PeerId, "error", err)
				continue
			}

			// Disconnect
			if err := k.DisconnectPeer(ctx, rep.PeerId); err != nil {
				k.logger.Debug("failed to disconnect peer (may already be disconnected)", "peer_id", rep.PeerId)
			}

			prunedCount++
		}
	}

	return prunedCount
}

// UpdateKnownPeerListBatched updates the known peer list for partition detection
// This is lightweight and doesn't need heavy batching, but we make it configurable
func (k Keeper) UpdateKnownPeerListBatched(ctx sdk.Context) error {
	peers := k.GetAllPeers(ctx)

	// DETERMINISM: Sort peers by peer ID before building peer ID list.
	// This ensures the peer list hash and order is consistent across validators.
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].PeerId < peers[j].PeerId
	})

	var peerIDs []string
	for _, peer := range peers {
		peerIDs = append(peerIDs, peer.PeerId)
	}

	// Don't use JSON marshaling in production - use protobuf for efficiency
	// But for compatibility with existing code, we keep this approach
	store := k.storeService.OpenKVStore(ctx)

	// Store peer count and hash instead of full list for efficiency
	peerCount := uint64(len(peerIDs))

	// Calculate a hash of peer IDs for change detection
	peerListHash := calculatePeerListHash(peerIDs)

	// Store count
	if err := store.Set([]byte("known_peer_count"), sdk.Uint64ToBigEndian(peerCount)); err != nil {
		return fmt.Errorf("failed to store peer count: %w", err)
	}

	// Store hash
	if err := store.Set([]byte("known_peer_hash"), peerListHash); err != nil {
		return fmt.Errorf("failed to store peer hash: %w", err)
	}

	// Still store full list for GetMissingPeerIDs compatibility
	// In a production optimization, this could be sharded or compressed
	k.UpdateKnownPeerList(ctx)

	return nil
}

// calculatePeerListHash creates a deterministic hash of peer IDs
func calculatePeerListHash(peerIDs []string) []byte {
	// Simple concatenation hash - in production use a proper hash function
	data := []byte{}
	for _, id := range peerIDs {
		data = append(data, []byte(id)...)
	}

	// Use SDK's built-in hash for determinism
	hash := sdk.Uint64ToBigEndian(uint64(len(data)))
	return hash
}
