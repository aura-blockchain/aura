package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// UpdateReputation updates the reputation score for a peer
func (k Keeper) UpdateReputation(ctx sdk.Context, peerID string, score int64, reason string) error {
	params, _ := k.GetParams(ctx)

	reputation, found := k.GetReputation(ctx, peerID)
	if !found {
		reputation = types.NodeReputation{
			PeerId:            peerID,
			Score:             params.Reputation.InitialScore,
			LastUpdatedHeight: ctx.BlockHeight(),
		}
	}

	// Update score
	reputation.Score = score

	// Enforce bounds
	if reputation.Score < 0 {
		reputation.Score = 0
	}
	if reputation.Score > params.Reputation.MaxScore {
		reputation.Score = params.Reputation.MaxScore
	}

	reputation.LastUpdatedHeight = ctx.BlockHeight()

	if err := k.SetReputation(ctx, reputation); err != nil {
		return fmt.Errorf("error in UpdateReputation: %w", err)
	}

	k.logger.Info(fmt.Sprintf("Updated reputation for peer %s to %d. Reason: %s", peerID, score, reason))

	return nil
}

// PenalizeReputation reduces reputation score for a peer
func (k Keeper) PenalizeReputation(ctx sdk.Context, peerID string, penalty int64) {
	reputation, found := k.GetReputation(ctx, peerID)
	if !found {
		params, _ := k.GetParams(ctx)
		reputation = types.NodeReputation{
			PeerId:            peerID,
			Score:             params.Reputation.InitialScore,
			LastUpdatedHeight: ctx.BlockHeight(),
		}
	}

	oldScore := reputation.Score
	reputation.Score -= penalty
	reputation.MisbehaviorCount++

	if reputation.Score < 0 {
		reputation.Score = 0
	}

	reputation.LastUpdatedHeight = ctx.BlockHeight()
	if err := k.SetReputation(ctx, reputation); err != nil {
		return
	}

	k.logger.Warn(fmt.Sprintf("Penalized peer %s: %d -> %d (penalty: %d, misbehavior count: %d)",
		peerID, oldScore, reputation.Score, penalty, reputation.MisbehaviorCount))
}

// RewardReputation increases reputation score for a peer
func (k Keeper) RewardReputation(ctx sdk.Context, peerID string, reward int64) {
	params, _ := k.GetParams(ctx)

	reputation, found := k.GetReputation(ctx, peerID)
	if !found {
		reputation = types.NodeReputation{
			PeerId:            peerID,
			Score:             params.Reputation.InitialScore,
			LastUpdatedHeight: ctx.BlockHeight(),
		}
	}

	oldScore := reputation.Score
	reputation.Score += reward

	if reputation.Score > params.Reputation.MaxScore {
		reputation.Score = params.Reputation.MaxScore
	}

	reputation.LastUpdatedHeight = ctx.BlockHeight()
	if err := k.SetReputation(ctx, reputation); err != nil {
		return
	}

	k.logger.Debug(fmt.Sprintf("Rewarded peer %s: %d -> %d (reward: %d)",
		peerID, oldScore, reputation.Score, reward))
}

// DecayReputations applies decay to all reputation scores
func (k Keeper) DecayReputations(ctx sdk.Context) {
	params, _ := k.GetParams(ctx)

	if !params.Reputation.EnableTracking || params.Reputation.DecayRate == 0 {
		return
	}

	reputations := k.GetAllReputations(ctx)

	for _, reputation := range reputations {
		// Only decay if last update was at least 100 blocks ago
		if ctx.BlockHeight()-reputation.LastUpdatedHeight >= 100 {
			reputation.Score -= params.Reputation.DecayRate

			if reputation.Score < 0 {
				reputation.Score = 0
			}

			reputation.LastUpdatedHeight = ctx.BlockHeight()
			if err := k.SetReputation(ctx, reputation); err != nil {
				k.logger.Error("failed to apply reputation decay", "peer", reputation.PeerId, "err", err)
			}
		}
	}
}

// UpdatePeerUptime updates the uptime for a connected peer
func (k Keeper) UpdatePeerUptime(ctx sdk.Context, peerID string) {
	peerInfo, found := k.GetPeerInfo(ctx, peerID)
	if !found {
		return
	}

	// Calculate uptime in seconds
	var uptime float64
	if !peerInfo.ConnectedAt.IsZero() {
		uptime = ctx.BlockTime().Sub(peerInfo.ConnectedAt).Seconds()
	}

	reputation, found := k.GetReputation(ctx, peerID)
	if found {
		reputation.Uptime = int64(uptime)
		if err := k.SetReputation(ctx, reputation); err != nil {
			k.logger.Error("failed to update reputation uptime", "peer", peerID, "err", err)
		}
	}
}

// GetHealthyPeers returns peers with reputation above threshold
func (k Keeper) GetHealthyPeers(ctx sdk.Context) []types.PeerInfo {
	params, _ := k.GetParams(ctx)
	allPeers := k.GetAllPeers(ctx)

	healthyPeers := make([]types.PeerInfo, 0, 64)
	for _, peer := range allPeers {
		if reputation, found := k.GetReputation(ctx, peer.PeerId); found {
			if reputation.Score >= params.Reputation.MinScoreToConnect {
				healthyPeers = append(healthyPeers, peer)
			}
		} else {
			// No reputation data, consider healthy
			healthyPeers = append(healthyPeers, peer)
		}
	}

	return healthyPeers
}

// GetBannedPeers returns all currently banned peers
func (k Keeper) GetBannedPeers(ctx sdk.Context) []string {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.RateLimitPrefix, storetypes.PrefixEndBytes(types.RateLimitPrefix))
	if err != nil {
		return []string{}
	}
	defer iterator.Close()

	bannedPeers := make([]string, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var entry types.RateLimitEntry
		if err := k.cdc.Unmarshal(iterator.Value(), &entry); err != nil {
			k.logger.Error("failed to unmarshal rate limit entry", "error", err)
			continue
		}

		if entry.IsBanned {
			// Check if ban is still active
			if entry.BanExpiresAt == nil || ctx.BlockTime().Before(*entry.BanExpiresAt) {
				bannedPeers = append(bannedPeers, entry.PeerId)
			}
		}
	}

	return bannedPeers
}

// CalculateAverageReputation calculates average reputation across all peers
func (k Keeper) CalculateAverageReputation(ctx sdk.Context) int64 {
	reputations := k.GetAllReputations(ctx)
	if len(reputations) == 0 {
		return 0
	}

	var total int64
	for _, rep := range reputations {
		total += rep.Score
	}

	return total / int64(len(reputations))
}

// GetReputationStats returns reputation statistics
func (k Keeper) GetReputationStats(ctx sdk.Context) (avgScore, minScore, maxScore int64, totalPeers int) {
	reputations := k.GetAllReputations(ctx)
	totalPeers = len(reputations)

	if totalPeers == 0 {
		return 0, 0, 0, 0
	}

	var total int64
	minScore = int64(^uint64(0) >> 1) // Max int64
	maxScore = int64(0)

	for _, rep := range reputations {
		total += rep.Score
		if rep.Score < minScore {
			minScore = rep.Score
		}
		if rep.Score > maxScore {
			maxScore = rep.Score
		}
	}

	avgScore = total / int64(totalPeers)
	return
}

// PruneLowReputationPeers disconnects peers with very low reputation
func (k Keeper) PruneLowReputationPeers(ctx sdk.Context) {
	params, _ := k.GetParams(ctx)

	reputations := k.GetAllReputations(ctx)
	for _, rep := range reputations {
		// If reputation is 0 and has high misbehavior count, disconnect
		if rep.Score == 0 && rep.MisbehaviorCount > 10 {
			k.logger.Info(fmt.Sprintf("Pruning low reputation peer: %s (score: %d, misbehavior: %d)",
				rep.PeerId, rep.Score, rep.MisbehaviorCount))

			// Ban the peer
			banDuration := int64(params.RateLimit.BanDuration.Seconds() * 2)
			if err := k.BanPeer(ctx, rep.PeerId, banDuration, "low reputation with high misbehavior"); err != nil {
				k.logger.Error("failed to ban low reputation peer", "peer", rep.PeerId, "err", err)
			}

			// Disconnect
			if err := k.DisconnectPeer(ctx, rep.PeerId); err != nil {
				k.logger.Error("failed to disconnect low reputation peer", "peer", rep.PeerId, "err", err)
			}
		}
	}
}

// ReputationBasedConnectionPriority returns connection priority for a peer
func (k Keeper) ReputationBasedConnectionPriority(ctx sdk.Context, peerID string) int64 {
	reputation, found := k.GetReputation(ctx, peerID)
	if !found {
		params, _ := k.GetParams(ctx)
		return params.Reputation.InitialScore
	}

	// Higher reputation = higher priority
	return reputation.Score
}

// TrackPeerBehavior tracks and updates peer behavior metrics
func (k Keeper) TrackPeerBehavior(ctx sdk.Context, peerID string, behaviorType string, isGood bool) {
	reputation, found := k.GetReputation(ctx, peerID)
	if !found {
		params, _ := k.GetParams(ctx)
		reputation = types.NodeReputation{
			PeerId:            peerID,
			Score:             params.Reputation.InitialScore,
			LastUpdatedHeight: ctx.BlockHeight(),
		}
	}

	params, _ := k.GetParams(ctx)

	switch behaviorType {
	case "message":
		reputation.MessagesReceived++
		if isGood {
			reputation.ValidMessages++
			// Small reward every 100 valid messages
			if reputation.ValidMessages%100 == 0 {
				k.RewardReputation(ctx, peerID, params.Reputation.GoodBehaviorReward)
			}
		} else {
			reputation.InvalidMessages++
			k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty)
		}

	case "connection":
		if isGood {
			// Peer maintained stable connection
			k.UpdatePeerUptime(ctx, peerID)
		} else {
			// Peer connection issues
			k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty/2)
		}

	case "rate_limit":
		if !isGood {
			// Rate limit violation
			k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty)
		}

	case "bandwidth":
		if !isGood {
			// Bandwidth abuse
			k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty/2)
		}
	}

	reputation.LastUpdatedHeight = ctx.BlockHeight()
	if err := k.SetReputation(ctx, reputation); err != nil {
		k.logger.Error("failed to persist reputation update", "peer", peerID, "err", err)
	}
}
