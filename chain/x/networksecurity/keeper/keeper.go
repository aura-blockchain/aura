// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	storeKey     storetypes.StoreKey
	authority    string
	logger       log.Logger

	// Rate limiters map (peer_id -> rate limiter)
	rateLimiters map[string]*RateLimiter

	// Bandwidth trackers map (peer_id -> bandwidth tracker)
	bandwidthTrackers map[string]*BandwidthTracker

	// Gossip message cache for deduplication
	messageCache *MessageCache
}

// MessageCache implements LRU cache for gossip message deduplication

// NewKeeper creates a new network security Keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	storeKey storetypes.StoreKey,
	authority string,
	logger log.Logger,
) Keeper {
	return Keeper{
		cdc:               cdc,
		storeService:      storeService,
		storeKey:          storeKey,
		authority:         authority,
		logger:            logger,
		rateLimiters:      make(map[string]*RateLimiter),
		bandwidthTrackers: make(map[string]*BandwidthTracker),
		messageCache:      NewMessageCache(10000), // Cache up to 10k messages
	}
}

// GetAuthority returns the module's authority
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns the module logger
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return k.logger
}

// GetParams returns the current module parameters
func (k *Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return types.Params{}, err
	}

	if bz == nil {
		return *types.DefaultParams(), nil
	}

	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		k.logger.Error("failed to unmarshal network security params", "error", err)
		return types.Params{}, types.ErrCorruptedState
	}
	return params, nil
}

// SetParams sets the module parameters
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := types.ValidateParams(&params); err != nil {
		return fmt.Errorf("error in SetParams for ValidateParams: %w", err)
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&params)
	return store.Set(types.ParamsKey, bz)
}

// GetPeerInfo retrieves peer information
func (k *Keeper) GetPeerInfo(ctx sdk.Context, peerID string) (types.PeerInfo, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetPeerInfoKey(peerID))
	if err != nil || bz == nil {
		return types.PeerInfo{}, false
	}

	var peerInfo types.PeerInfo
	if err := k.cdc.Unmarshal(bz, &peerInfo); err != nil {
		k.logger.Error("failed to unmarshal peer info", "peer_id", peerID, "error", err)
		return types.PeerInfo{}, false
	}
	return peerInfo, true
}

// SetPeerInfo stores peer information
func (k *Keeper) SetPeerInfo(ctx sdk.Context, peerInfo types.PeerInfo) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&peerInfo)
	return store.Set(types.GetPeerInfoKey(peerInfo.PeerId), bz)
}

// GetAllPeers retrieves all connected peers
func (k *Keeper) GetAllPeers(ctx sdk.Context) []types.PeerInfo {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.PeerInfoPrefix, storetypes.PrefixEndBytes(types.PeerInfoPrefix))
	if err != nil {
		return []types.PeerInfo{}
	}
	defer iterator.Close()

	peers := make([]types.PeerInfo, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var peer types.PeerInfo
		if err := k.cdc.Unmarshal(iterator.Value(), &peer); err != nil {
			k.logger.Error("failed to unmarshal peer in GetAllPeers, skipping", "error", err)
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

// IsTrustedPeer checks if a peer is trusted
func (k *Keeper) IsTrustedPeer(ctx sdk.Context, peerID string) bool {
	store := k.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.GetTrustedPeerKey(peerID))
	if err != nil {
		return false
	}
	return has
}

// GetTrustedPeer retrieves a trusted peer
func (k *Keeper) GetTrustedPeer(ctx sdk.Context, peerID string) (types.TrustedPeer, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetTrustedPeerKey(peerID))
	if err != nil || bz == nil {
		return types.TrustedPeer{}, false
	}

	var peer types.TrustedPeer
	if err := k.cdc.Unmarshal(bz, &peer); err != nil {
		k.logger.Error("failed to unmarshal trusted peer", "peer_id", peerID, "error", err)
		return types.TrustedPeer{}, false
	}
	return peer, true
}

// SetTrustedPeer adds a trusted peer
func (k *Keeper) SetTrustedPeer(ctx sdk.Context, peer types.TrustedPeer) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&peer)
	return store.Set(types.GetTrustedPeerKey(peer.PeerId), bz)
}

// RemoveTrustedPeer removes a trusted peer
func (k *Keeper) RemoveTrustedPeer(ctx sdk.Context, peerID string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetTrustedPeerKey(peerID))
}

// GetAllTrustedPeers retrieves all trusted peers
func (k *Keeper) GetAllTrustedPeers(ctx sdk.Context) []types.TrustedPeer {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.TrustedPeerPrefix, storetypes.PrefixEndBytes(types.TrustedPeerPrefix))
	if err != nil {
		return []types.TrustedPeer{}
	}
	defer iterator.Close()

	peers := make([]types.TrustedPeer, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var peer types.TrustedPeer
		if err := k.cdc.Unmarshal(iterator.Value(), &peer); err != nil {
			k.logger.Error("failed to unmarshal trusted peer in GetAllTrustedPeers, skipping", "error", err)
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

// IsBanned checks if a peer is currently banned
func (k *Keeper) IsBanned(ctx sdk.Context, peerID string) bool {
	rateLimitEntry, found := k.GetRateLimitEntry(ctx, peerID)
	if !found {
		return false
	}

	// Check if ban has expired
	if rateLimitEntry.IsBanned {
		if rateLimitEntry.BanExpiresAt != nil && ctx.BlockTime().After(*rateLimitEntry.BanExpiresAt) {
			// Ban expired, remove it
			if err := k.UnbanPeer(ctx, peerID); err != nil {
				k.logger.Error("failed to unban peer during IsBanned", "peer", peerID, "err", err)
			}
			return false
		}
		return true
	}
	return false
}

// BanPeer bans a peer until the specified time
func (k *Keeper) BanPeer(ctx sdk.Context, peerID string, duration int64, reason string) error {
	rateLimitEntry, found := k.GetRateLimitEntry(ctx, peerID)
	if !found {
		rateLimitEntry = types.RateLimitEntry{
			PeerId: peerID,
		}
	}

	banTime := ctx.BlockTime().Add(time.Duration(duration))
	rateLimitEntry.IsBanned = true
	rateLimitEntry.BanExpiresAt = &banTime

	if err := k.SetRateLimitEntry(ctx, rateLimitEntry); err != nil {
		return fmt.Errorf("error in BanPeer for PeerId: %w", err)
	}

	// Log the ban
	k.logger.Info(fmt.Sprintf("Peer %s banned for %d seconds. Reason: %s", peerID, duration, reason))

	return nil
}

// UnbanPeer removes a ban from a peer
func (k *Keeper) UnbanPeer(ctx sdk.Context, peerID string) error {
	rateLimitEntry, found := k.GetRateLimitEntry(ctx, peerID)
	if !found {
		return nil
	}

	rateLimitEntry.IsBanned = false
	rateLimitEntry.BanExpiresAt = nil
	if err := k.SetRateLimitEntry(ctx, rateLimitEntry); err != nil {
		return fmt.Errorf("error in UnbanPeer: %w", err)
	}

	k.logger.Info(fmt.Sprintf("Peer %s unbanned", peerID))
	return nil
}

// GetRateLimitEntry retrieves rate limit entry for a peer
func (k *Keeper) GetRateLimitEntry(ctx sdk.Context, peerID string) (types.RateLimitEntry, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetRateLimitKey(peerID))
	if err != nil || bz == nil {
		return types.RateLimitEntry{}, false
	}

	var entry types.RateLimitEntry
	if err := k.cdc.Unmarshal(bz, &entry); err != nil {
		k.logger.Error("failed to unmarshal rate limit entry", "peer_id", peerID, "error", err)
		return types.RateLimitEntry{}, false
	}
	return entry, true
}

// SetRateLimitEntry stores rate limit entry
func (k *Keeper) SetRateLimitEntry(ctx sdk.Context, entry types.RateLimitEntry) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&entry)
	return store.Set(types.GetRateLimitKey(entry.PeerId), bz)
}

// GetReputation retrieves reputation for a peer
func (k *Keeper) GetReputation(ctx sdk.Context, peerID string) (types.NodeReputation, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetReputationKey(peerID))
	if err != nil || bz == nil {
		return types.NodeReputation{}, false
	}

	var reputation types.NodeReputation
	if err := k.cdc.Unmarshal(bz, &reputation); err != nil {
		k.logger.Error("failed to unmarshal node reputation", "peer_id", peerID, "error", err)
		return types.NodeReputation{}, false
	}
	return reputation, true
}

// SetReputation stores reputation for a peer
func (k *Keeper) SetReputation(ctx sdk.Context, reputation types.NodeReputation) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&reputation)
	return store.Set(types.GetReputationKey(reputation.PeerId), bz)
}

// GetAllReputations retrieves all reputations
func (k *Keeper) GetAllReputations(ctx sdk.Context) []types.NodeReputation {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ReputationPrefix, storetypes.PrefixEndBytes(types.ReputationPrefix))
	if err != nil {
		return []types.NodeReputation{}
	}
	defer iterator.Close()

	reputations := make([]types.NodeReputation, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var rep types.NodeReputation
		if err := k.cdc.Unmarshal(iterator.Value(), &rep); err != nil {
			k.logger.Error("failed to unmarshal reputation in GetAllReputations, skipping", "error", err)
			continue
		}
		reputations = append(reputations, rep)
	}
	return reputations
}

// GetMempoolStats retrieves mempool statistics
func (k *Keeper) GetMempoolStats(ctx sdk.Context) types.MempoolStats {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.MempoolStatsKey)
	if err != nil || bz == nil {
		return types.MempoolStats{}
	}

	var stats types.MempoolStats
	if err := k.cdc.Unmarshal(bz, &stats); err != nil {
		k.logger.Error("failed to unmarshal mempool stats", "error", err)
		return types.MempoolStats{}
	}
	return stats
}

// SetMempoolStats stores mempool statistics
func (k *Keeper) SetMempoolStats(ctx sdk.Context, stats types.MempoolStats) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&stats)
	return store.Set(types.MempoolStatsKey, bz)
}

// GetForkAlert retrieves a fork alert
func (k *Keeper) GetForkAlert(ctx sdk.Context, alertID string) (types.ForkAlert, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetForkAlertKey(alertID))
	if err != nil || bz == nil {
		return types.ForkAlert{}, false
	}

	var alert types.ForkAlert
	if err := k.cdc.Unmarshal(bz, &alert); err != nil {
		k.logger.Error("failed to unmarshal fork alert", "alert_id", alertID, "error", err)
		return types.ForkAlert{}, false
	}
	return alert, true
}

// SetForkAlert stores a fork alert
func (k *Keeper) SetForkAlert(ctx sdk.Context, alert types.ForkAlert) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&alert)
	return store.Set(types.GetForkAlertKey(alert.AlertId), bz)
}

// GetAllForkAlerts retrieves all fork alerts
func (k *Keeper) GetAllForkAlerts(ctx sdk.Context, includeResolved bool) []types.ForkAlert {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ForkAlertPrefix, storetypes.PrefixEndBytes(types.ForkAlertPrefix))
	if err != nil {
		return []types.ForkAlert{}
	}
	defer iterator.Close()

	alerts := make([]types.ForkAlert, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var alert types.ForkAlert
		if err := k.cdc.Unmarshal(iterator.Value(), &alert); err != nil {
			k.logger.Error("failed to unmarshal fork alert in GetAllForkAlerts, skipping", "error", err)
			continue
		}
		if includeResolved || !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// GetPartitionAlert retrieves a partition alert
func (k *Keeper) GetPartitionAlert(ctx sdk.Context, alertID string) (types.PartitionAlert, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetPartitionAlertKey(alertID))
	if err != nil || bz == nil {
		return types.PartitionAlert{}, false
	}

	var alert types.PartitionAlert
	if err := k.cdc.Unmarshal(bz, &alert); err != nil {
		k.logger.Error("failed to unmarshal partition alert", "alert_id", alertID, "error", err)
		return types.PartitionAlert{}, false
	}
	return alert, true
}

// SetPartitionAlert stores a partition alert
func (k *Keeper) SetPartitionAlert(ctx sdk.Context, alert types.PartitionAlert) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&alert)
	return store.Set(types.GetPartitionAlertKey(alert.AlertId), bz)
}

// GetAllPartitionAlerts retrieves all partition alerts
func (k *Keeper) GetAllPartitionAlerts(ctx sdk.Context, includeResolved bool) []types.PartitionAlert {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.PartitionAlertPrefix, storetypes.PrefixEndBytes(types.PartitionAlertPrefix))
	if err != nil {
		return []types.PartitionAlert{}
	}
	defer iterator.Close()

	alerts := make([]types.PartitionAlert, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var alert types.PartitionAlert
		if err := k.cdc.Unmarshal(iterator.Value(), &alert); err != nil {
			k.logger.Error("failed to unmarshal partition alert in GetAllPartitionAlerts, skipping", "error", err)
			continue
		}
		if includeResolved || !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// GetConnectionCount retrieves connection count for an IP
func (k *Keeper) GetConnectionCount(ctx sdk.Context, ipAddress string) uint32 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetConnectionCountKey(ipAddress))
	if err != nil || bz == nil {
		return 0
	}

	var count uint32
	// Decode binary uint32
	if len(bz) >= 4 {
		count = uint32(bz[0]) | uint32(bz[1])<<8 | uint32(bz[2])<<16 | uint32(bz[3])<<24
	}
	return count
}

// SetConnectionCount sets connection count for an IP
// SetConnectionCount sets connection count for an IP
func (k *Keeper) SetConnectionCount(ctx sdk.Context, ipAddress string, count uint32) error {
	store := k.storeService.OpenKVStore(ctx)
	// Encode as binary uint32
	bz := make([]byte, 4)
	bz[0] = byte(count)
	bz[1] = byte(count >> 8)
	bz[2] = byte(count >> 16)
	bz[3] = byte(count >> 24)
	return store.Set(types.GetConnectionCountKey(ipAddress), bz)
}

// IncrementConnectionCount increments connection count for an IP
func (k *Keeper) IncrementConnectionCount(ctx sdk.Context, ipAddress string) error {
	count := k.GetConnectionCount(ctx, ipAddress)
	return k.SetConnectionCount(ctx, ipAddress, count+1)
}

// DecrementConnectionCount decrements connection count for an IP
func (k *Keeper) DecrementConnectionCount(ctx sdk.Context, ipAddress string) error {
	count := k.GetConnectionCount(ctx, ipAddress)
	if count > 0 {
		return k.SetConnectionCount(ctx, ipAddress, count-1)
	}
	return nil
}

// NewMessageCache creates a new message cache for gossip deduplication

// Has checks if a message hash exists in the cache

// Add adds a message hash to the cache

// CheckGossipMessage hashes the provided message payload and reports whether it
// has already been seen. The boolean return value is true when the payload was
// added to the cache, and false when ErrDuplicateMessage indicates a cache hit.
func (k *Keeper) CheckGossipMessage(ctx sdk.Context, messageData []byte) (bool, string, error) {
	// Hash the message
	hash := sha256.Sum256(messageData)
	messageHash := hex.EncodeToString(hash[:])

	// Check cache
	if k.messageCache.Has(messageHash) {
		telemetry.IncrCounter(float32(1), "networksecurity", "gossip", "cache_hit")
		return false, messageHash, types.ErrDuplicateMessage
	}

	// Add to cache with empty senderID (not available in this context)
	k.messageCache.Add(ctx, messageHash, "")
	telemetry.IncrCounter(float32(1), "networksecurity", "gossip", "cache_miss")

	return true, messageHash, nil
}

// ValidateGossipMessage performs comprehensive gossip message validation

// GetMessageCacheStats returns statistics about the message cache
func (k *Keeper) GetMessageCacheStats() MessageCacheStats {
	return k.messageCache.Stats()
}

// Batch processing constants for BeginBlocker performance optimization
const (
	MAX_THREAT_UPDATES_PER_BLOCK = 50
	MAX_ALERTS_PER_BLOCK         = 20
	REPUTATION_REFRESH_INTERVAL  = 100 // blocks
)

// GetBatchCursor retrieves a batch processing cursor from state
func (k *Keeper) GetBatchCursor(ctx sdk.Context, cursorKey []byte) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(cursorKey)
	if err != nil || bz == nil {
		return 0, nil
	}

	if len(bz) != 8 {
		return 0, nil
	}
	return sdk.BigEndianToUint64(bz), nil
}

// SetBatchCursor stores a batch processing cursor in state
func (k *Keeper) SetBatchCursor(ctx sdk.Context, cursorKey []byte, cursor uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := sdk.Uint64ToBigEndian(cursor)
	return store.Set(cursorKey, bz)
}

// KVStoreService returns the keeper's store service (for testing)
func (k *Keeper) KVStoreService() store.KVStoreService {
	return k.storeService
}
