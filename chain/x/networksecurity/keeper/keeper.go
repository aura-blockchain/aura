package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
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
	authority string,
	logger log.Logger,
) Keeper {
	return Keeper{
		cdc:               cdc,
		storeService:      storeService,
		authority:         authority,
		logger:            logger,
		rateLimiters:      make(map[string]*RateLimiter),
		bandwidthTrackers: make(map[string]*BandwidthTracker),
		messageCache:      NewMessageCache(10000), // Cache up to 10k messages
	}
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns the module logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return k.logger
}

// GetParams returns the current module parameters
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return types.Params{}, err
	}

	if bz == nil {
		return types.DefaultParams(), nil
	}

	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params, nil
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&params)
	return store.Set(types.ParamsKey, bz)
}

// GetPeerInfo retrieves peer information
func (k Keeper) GetPeerInfo(ctx sdk.Context, peerID string) (types.PeerInfo, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetPeerInfoKey(peerID))
	if err != nil || bz == nil {
		return types.PeerInfo{}, false
	}

	var peerInfo types.PeerInfo
	k.cdc.MustUnmarshal(bz, &peerInfo)
	return peerInfo, true
}

// SetPeerInfo stores peer information
func (k Keeper) SetPeerInfo(ctx sdk.Context, peerInfo types.PeerInfo) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&peerInfo)
	return store.Set(types.GetPeerInfoKey(peerInfo.PeerId), bz)
}

// GetAllPeers retrieves all connected peers
func (k Keeper) GetAllPeers(ctx sdk.Context) []types.PeerInfo {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.PeerInfoPrefix, sdk.PrefixEndBytes(types.PeerInfoPrefix))
	if err != nil {
		return []types.PeerInfo{}
	}
	defer iterator.Close()

	var peers []types.PeerInfo
	for ; iterator.Valid(); iterator.Next() {
		var peer types.PeerInfo
		k.cdc.MustUnmarshal(iterator.Value(), &peer)
		peers = append(peers, peer)
	}
	return peers
}

// IsTrustedPeer checks if a peer is trusted
func (k Keeper) IsTrustedPeer(ctx sdk.Context, peerID string) bool {
	store := k.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.GetTrustedPeerKey(peerID))
	if err != nil {
		return false
	}
	return has
}

// GetTrustedPeer retrieves a trusted peer
func (k Keeper) GetTrustedPeer(ctx sdk.Context, peerID string) (types.TrustedPeer, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetTrustedPeerKey(peerID))
	if err != nil || bz == nil {
		return types.TrustedPeer{}, false
	}

	var peer types.TrustedPeer
	k.cdc.MustUnmarshal(bz, &peer)
	return peer, true
}

// SetTrustedPeer adds a trusted peer
func (k Keeper) SetTrustedPeer(ctx sdk.Context, peer types.TrustedPeer) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&peer)
	return store.Set(types.GetTrustedPeerKey(peer.PeerId), bz)
}

// RemoveTrustedPeer removes a trusted peer
func (k Keeper) RemoveTrustedPeer(ctx sdk.Context, peerID string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetTrustedPeerKey(peerID))
}

// GetAllTrustedPeers retrieves all trusted peers
func (k Keeper) GetAllTrustedPeers(ctx sdk.Context) []types.TrustedPeer {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.TrustedPeerPrefix, sdk.PrefixEndBytes(types.TrustedPeerPrefix))
	if err != nil {
		return []types.TrustedPeer{}
	}
	defer iterator.Close()

	var peers []types.TrustedPeer
	for ; iterator.Valid(); iterator.Next() {
		var peer types.TrustedPeer
		k.cdc.MustUnmarshal(iterator.Value(), &peer)
		peers = append(peers, peer)
	}
	return peers
}

// IsBanned checks if a peer is currently banned
func (k Keeper) IsBanned(ctx sdk.Context, peerID string) bool {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetBannedPeerKey(peerID))
	if err != nil || bz == nil {
		return false
	}

	var rateLimitEntry types.RateLimitEntry
	k.cdc.MustUnmarshal(bz, &rateLimitEntry)

	// Check if ban has expired
	if rateLimitEntry.IsBanned {
		if rateLimitEntry.BanExpiresAt != nil && ctx.BlockTime().After(*rateLimitEntry.BanExpiresAt) {
			// Ban expired, remove it
			k.UnbanPeer(ctx, peerID)
			return false
		}
		return true
	}
	return false
}

// BanPeer bans a peer until the specified time
func (k Keeper) BanPeer(ctx sdk.Context, peerID string, duration int64, reason string) error {
	rateLimitEntry, found := k.GetRateLimitEntry(ctx, peerID)
	if !found {
		rateLimitEntry = types.RateLimitEntry{
			PeerId: peerID,
		}
	}

	banExpiresAt := ctx.BlockTime().Add(sdk.NewInt(duration).ToDec().TruncateInt64())
	rateLimitEntry.IsBanned = true
	rateLimitEntry.BanExpiresAt = &banExpiresAt

	k.SetRateLimitEntry(ctx, rateLimitEntry)

	// Log the ban
	k.logger.Info(fmt.Sprintf("Peer %s banned for %d seconds. Reason: %s", peerID, duration, reason))

	return nil
}

// UnbanPeer removes a ban from a peer
func (k Keeper) UnbanPeer(ctx sdk.Context, peerID string) error {
	store := k.storeService.OpenKVStore(ctx)
	rateLimitEntry, found := k.GetRateLimitEntry(ctx, peerID)
	if !found {
		return nil
	}

	rateLimitEntry.IsBanned = false
	rateLimitEntry.BanExpiresAt = nil
	k.SetRateLimitEntry(ctx, rateLimitEntry)

	k.logger.Info(fmt.Sprintf("Peer %s unbanned", peerID))
	return store.Delete(types.GetBannedPeerKey(peerID))
}

// GetRateLimitEntry retrieves rate limit entry for a peer
func (k Keeper) GetRateLimitEntry(ctx sdk.Context, peerID string) (types.RateLimitEntry, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetRateLimitKey(peerID))
	if err != nil || bz == nil {
		return types.RateLimitEntry{}, false
	}

	var entry types.RateLimitEntry
	k.cdc.MustUnmarshal(bz, &entry)
	return entry, true
}

// SetRateLimitEntry stores rate limit entry
func (k Keeper) SetRateLimitEntry(ctx sdk.Context, entry types.RateLimitEntry) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&entry)
	return store.Set(types.GetRateLimitKey(entry.PeerId), bz)
}

// GetReputation retrieves reputation for a peer
func (k Keeper) GetReputation(ctx sdk.Context, peerID string) (types.NodeReputation, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetReputationKey(peerID))
	if err != nil || bz == nil {
		return types.NodeReputation{}, false
	}

	var reputation types.NodeReputation
	k.cdc.MustUnmarshal(bz, &reputation)
	return reputation, true
}

// SetReputation stores reputation for a peer
func (k Keeper) SetReputation(ctx sdk.Context, reputation types.NodeReputation) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&reputation)
	return store.Set(types.GetReputationKey(reputation.PeerId), bz)
}

// GetAllReputations retrieves all reputations
func (k Keeper) GetAllReputations(ctx sdk.Context) []types.NodeReputation {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ReputationPrefix, sdk.PrefixEndBytes(types.ReputationPrefix))
	if err != nil {
		return []types.NodeReputation{}
	}
	defer iterator.Close()

	var reputations []types.NodeReputation
	for ; iterator.Valid(); iterator.Next() {
		var rep types.NodeReputation
		k.cdc.MustUnmarshal(iterator.Value(), &rep)
		reputations = append(reputations, rep)
	}
	return reputations
}

// GetMempoolStats retrieves mempool statistics
func (k Keeper) GetMempoolStats(ctx sdk.Context) types.MempoolStats {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.MempoolStatsKey)
	if err != nil || bz == nil {
		return types.MempoolStats{}
	}

	var stats types.MempoolStats
	k.cdc.MustUnmarshal(bz, &stats)
	return stats
}

// SetMempoolStats stores mempool statistics
func (k Keeper) SetMempoolStats(ctx sdk.Context, stats types.MempoolStats) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&stats)
	return store.Set(types.MempoolStatsKey, bz)
}

// GetForkAlert retrieves a fork alert
func (k Keeper) GetForkAlert(ctx sdk.Context, alertID string) (types.ForkAlert, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetForkAlertKey(alertID))
	if err != nil || bz == nil {
		return types.ForkAlert{}, false
	}

	var alert types.ForkAlert
	k.cdc.MustUnmarshal(bz, &alert)
	return alert, true
}

// SetForkAlert stores a fork alert
func (k Keeper) SetForkAlert(ctx sdk.Context, alert types.ForkAlert) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&alert)
	return store.Set(types.GetForkAlertKey(alert.AlertId), bz)
}

// GetAllForkAlerts retrieves all fork alerts
func (k Keeper) GetAllForkAlerts(ctx sdk.Context, includeResolved bool) []types.ForkAlert {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ForkAlertPrefix, sdk.PrefixEndBytes(types.ForkAlertPrefix))
	if err != nil {
		return []types.ForkAlert{}
	}
	defer iterator.Close()

	var alerts []types.ForkAlert
	for ; iterator.Valid(); iterator.Next() {
		var alert types.ForkAlert
		k.cdc.MustUnmarshal(iterator.Value(), &alert)
		if includeResolved || !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// GetPartitionAlert retrieves a partition alert
func (k Keeper) GetPartitionAlert(ctx sdk.Context, alertID string) (types.PartitionAlert, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetPartitionAlertKey(alertID))
	if err != nil || bz == nil {
		return types.PartitionAlert{}, false
	}

	var alert types.PartitionAlert
	k.cdc.MustUnmarshal(bz, &alert)
	return alert, true
}

// SetPartitionAlert stores a partition alert
func (k Keeper) SetPartitionAlert(ctx sdk.Context, alert types.PartitionAlert) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&alert)
	return store.Set(types.GetPartitionAlertKey(alert.AlertId), bz)
}

// GetAllPartitionAlerts retrieves all partition alerts
func (k Keeper) GetAllPartitionAlerts(ctx sdk.Context, includeResolved bool) []types.PartitionAlert {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.PartitionAlertPrefix, sdk.PrefixEndBytes(types.PartitionAlertPrefix))
	if err != nil {
		return []types.PartitionAlert{}
	}
	defer iterator.Close()

	var alerts []types.PartitionAlert
	for ; iterator.Valid(); iterator.Next() {
		var alert types.PartitionAlert
		k.cdc.MustUnmarshal(iterator.Value(), &alert)
		if includeResolved || !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// GetConnectionCount retrieves connection count for an IP
func (k Keeper) GetConnectionCount(ctx sdk.Context, ipAddress string) uint32 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetConnectionCountKey(ipAddress))
	if err != nil || bz == nil {
		return 0
	}

	var count uint32
	k.cdc.MustUnmarshal(bz, &count)
	return count
}

// SetConnectionCount sets connection count for an IP
func (k Keeper) SetConnectionCount(ctx sdk.Context, ipAddress string, count uint32) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&count)
	return store.Set(types.GetConnectionCountKey(ipAddress), bz)
}

// IncrementConnectionCount increments connection count for an IP
func (k Keeper) IncrementConnectionCount(ctx sdk.Context, ipAddress string) error {
	count := k.GetConnectionCount(ctx, ipAddress)
	return k.SetConnectionCount(ctx, ipAddress, count+1)
}

// DecrementConnectionCount decrements connection count for an IP
func (k Keeper) DecrementConnectionCount(ctx sdk.Context, ipAddress string) error {
	count := k.GetConnectionCount(ctx, ipAddress)
	if count > 0 {
		return k.SetConnectionCount(ctx, ipAddress, count-1)
	}
	return nil
}

// NewMessageCache creates a new message cache for gossip deduplication

// Has checks if a message hash exists in the cache

// Add adds a message hash to the cache

// evictOldest removes the oldest entry from the cache
func (mc *MessageCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, timestamp := range mc.cache {
		if first || timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = timestamp
			first = false
		}
	}

	if oldestKey != "" {
		delete(mc.cache, oldestKey)
	}
}

// cleanup removes expired entries from the cache
func (mc *MessageCache) cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	for key, timestamp := range mc.cache {
		if now.Sub(timestamp) > mc.ttl {
			delete(mc.cache, key)
		}
	}
}

// startCleanup starts periodic cleanup of expired entries
func (mc *MessageCache) startCleanup() {
	// Run cleanup every minute
	mc.cleanupTimer = time.AfterFunc(time.Minute, func() {
		mc.cleanup()
		mc.startCleanup() // Reschedule
	})
}

// Stop stops the cleanup timer
func (mc *MessageCache) Stop() {
	if mc.cleanupTimer != nil {
		mc.cleanupTimer.Stop()
	}
}

// Size returns the current cache size

// Clear removes all entries from the cache
func (mc *MessageCache) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cache = make(map[string]time.Time)
}

// CheckGossipMessage checks if a gossip message is a duplicate
func (k Keeper) CheckGossipMessage(ctx sdk.Context, messageData []byte) (bool, string, error) {
	// Hash the message
	hash := sha256.Sum256(messageData)
	messageHash := hex.EncodeToString(hash[:])

	// Check cache
	if k.messageCache.Has(messageHash) {
		return false, messageHash, types.ErrDuplicateMessage
	}

	// Add to cache
	k.messageCache.Add(messageHash)

	return true, messageHash, nil
}

// ValidateGossipMessage performs comprehensive gossip message validation

// GetMessageCacheStats returns statistics about the message cache
func (k Keeper) GetMessageCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"size":     k.messageCache.Size(),
		"max_size": k.messageCache.maxSize,
		"ttl":      k.messageCache.ttl.String(),
	}
}
