package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// =============================================================================
// Network Security Operations
// =============================================================================

// SetRateLimit stores a rate limit entry
func (k Keeper) SetRateLimit(ctx sdk.Context, rl *securitypb.RateLimitEntry) {
	store := k.GetStore(ctx)
	key := append(types.RateLimitKey, []byte(rl.PeerId)...)
	bz := k.cdc.MustMarshal(rl)
	store.Set(key, bz)
}

// GetRateLimit retrieves a rate limit entry
func (k Keeper) GetRateLimit(ctx sdk.Context, peerId string) (*securitypb.RateLimitEntry, bool) {
	store := k.GetStore(ctx)
	key := append(types.RateLimitKey, []byte(peerId)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var rl securitypb.RateLimitEntry
	k.cdc.MustUnmarshal(bz, &rl)
	return &rl, true
}

// GetAllRateLimits returns all rate limit entries
func (k Keeper) GetAllRateLimits(ctx sdk.Context) []*securitypb.RateLimitEntry {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.RateLimitKey)
	defer iterator.Close()

	var rateLimits []*securitypb.RateLimitEntry
	for ; iterator.Valid(); iterator.Next() {
		var rl securitypb.RateLimitEntry
		k.cdc.MustUnmarshal(iterator.Value(), &rl)
		rateLimits = append(rateLimits, &rl)
	}
	return rateLimits
}

// DeleteRateLimit removes a rate limit entry
func (k Keeper) DeleteRateLimit(ctx sdk.Context, peerId string) {
	store := k.GetStore(ctx)
	key := append(types.RateLimitKey, []byte(peerId)...)
	store.Delete(key)
}

// SetPeerReputation stores a peer reputation entry
func (k Keeper) SetPeerReputation(ctx sdk.Context, rep *securitypb.NodeReputation) {
	store := k.GetStore(ctx)
	key := append(types.ReputationKey, []byte(rep.PeerId)...)
	bz := k.cdc.MustMarshal(rep)
	store.Set(key, bz)
}

// GetPeerReputation retrieves a peer reputation entry
func (k Keeper) GetPeerReputation(ctx sdk.Context, peerID string) (*securitypb.NodeReputation, bool) {
	store := k.GetStore(ctx)
	key := append(types.ReputationKey, []byte(peerID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var rep securitypb.NodeReputation
	k.cdc.MustUnmarshal(bz, &rep)
	return &rep, true
}

// GetAllPeerReputations returns all peer reputations
func (k Keeper) GetAllPeerReputations(ctx sdk.Context) []*securitypb.NodeReputation {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ReputationKey)
	defer iterator.Close()

	var reputations []*securitypb.NodeReputation
	for ; iterator.Valid(); iterator.Next() {
		var rep securitypb.NodeReputation
		k.cdc.MustUnmarshal(iterator.Value(), &rep)
		reputations = append(reputations, &rep)
	}
	return reputations
}

// SetTrustedPeer stores a trusted peer entry
func (k Keeper) SetTrustedPeer(ctx sdk.Context, peer *securitypb.TrustedPeer) {
	store := k.GetStore(ctx)
	key := append(types.TrustedPeerKey, []byte(peer.PeerId)...)
	bz := k.cdc.MustMarshal(peer)
	store.Set(key, bz)
}

// GetTrustedPeer retrieves a trusted peer entry
func (k Keeper) GetTrustedPeer(ctx sdk.Context, peerID string) (*securitypb.TrustedPeer, bool) {
	store := k.GetStore(ctx)
	key := append(types.TrustedPeerKey, []byte(peerID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var peer securitypb.TrustedPeer
	k.cdc.MustUnmarshal(bz, &peer)
	return &peer, true
}

// GetAllTrustedPeers returns all trusted peers
func (k Keeper) GetAllTrustedPeers(ctx sdk.Context) []*securitypb.TrustedPeer {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.TrustedPeerKey)
	defer iterator.Close()

	var peers []*securitypb.TrustedPeer
	for ; iterator.Valid(); iterator.Next() {
		var peer securitypb.TrustedPeer
		k.cdc.MustUnmarshal(iterator.Value(), &peer)
		peers = append(peers, &peer)
	}
	return peers
}

// DeleteTrustedPeer removes a trusted peer
func (k Keeper) DeleteTrustedPeer(ctx sdk.Context, peerID string) {
	store := k.GetStore(ctx)
	key := append(types.TrustedPeerKey, []byte(peerID)...)
	store.Delete(key)
}

// SetBlacklistEntry stores a blacklist entry
func (k Keeper) SetBlacklistEntry(ctx sdk.Context, entry *types.BlacklistEntry) {
	store := k.GetStore(ctx)
	key := append(types.BlacklistKey, []byte(entry.Identifier)...)
	bz := k.cdc.MustMarshal(entry)
	store.Set(key, bz)
}

// GetBlacklistEntry retrieves a blacklist entry
func (k Keeper) GetBlacklistEntry(ctx sdk.Context, identifier string) (*types.BlacklistEntry, bool) {
	store := k.GetStore(ctx)
	key := append(types.BlacklistKey, []byte(identifier)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var entry types.BlacklistEntry
	k.cdc.MustUnmarshal(bz, &entry)
	return &entry, true
}

// GetAllBlacklistEntries returns all blacklist entries
func (k Keeper) GetAllBlacklistEntries(ctx sdk.Context) []*types.BlacklistEntry {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.BlacklistKey)
	defer iterator.Close()

	var entries []*types.BlacklistEntry
	for ; iterator.Valid(); iterator.Next() {
		var entry types.BlacklistEntry
		k.cdc.MustUnmarshal(iterator.Value(), &entry)
		entries = append(entries, &entry)
	}
	return entries
}

// DeleteBlacklistEntry removes a blacklist entry
func (k Keeper) DeleteBlacklistEntry(ctx sdk.Context, identifier string) {
	store := k.GetStore(ctx)
	key := append(types.BlacklistKey, []byte(identifier)...)
	store.Delete(key)
}

// IsBlacklisted checks if an identifier is blacklisted
func (k Keeper) IsBlacklisted(ctx sdk.Context, identifier string) bool {
	entry, found := k.GetBlacklistEntry(ctx, identifier)
	if !found {
		return false
	}
	// Check if blacklist has expired
	if !entry.Permanent && entry.ExpiresAt != nil && ctx.BlockTime().After(*entry.ExpiresAt) {
		k.DeleteBlacklistEntry(ctx, identifier)
		return false
	}
	return true
}

// SetForkAlert stores a fork alert
func (k Keeper) SetForkAlert(ctx sdk.Context, alert *securitypb.ForkAlert) {
	store := k.GetStore(ctx)
	key := append(types.ForkAlertKey, []byte(alert.AlertId)...)
	bz := k.cdc.MustMarshal(alert)
	store.Set(key, bz)
}

// GetAllForkAlerts returns all fork alerts
func (k Keeper) GetAllForkAlerts(ctx sdk.Context) []*securitypb.ForkAlert {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ForkAlertKey)
	defer iterator.Close()

	var alerts []*securitypb.ForkAlert
	for ; iterator.Valid(); iterator.Next() {
		var alert securitypb.ForkAlert
		k.cdc.MustUnmarshal(iterator.Value(), &alert)
		alerts = append(alerts, &alert)
	}
	return alerts
}

// SetPartitionAlert stores a partition alert
func (k Keeper) SetPartitionAlert(ctx sdk.Context, alert *securitypb.PartitionAlert) {
	store := k.GetStore(ctx)
	key := append(types.PartitionAlertKey, []byte(alert.AlertId)...)
	bz := k.cdc.MustMarshal(alert)
	store.Set(key, bz)
}

// GetAllPartitionAlerts returns all partition alerts
func (k Keeper) GetAllPartitionAlerts(ctx sdk.Context) []*securitypb.PartitionAlert {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.PartitionAlertKey)
	defer iterator.Close()

	var alerts []*securitypb.PartitionAlert
	for ; iterator.Valid(); iterator.Next() {
		var alert securitypb.PartitionAlert
		k.cdc.MustUnmarshal(iterator.Value(), &alert)
		alerts = append(alerts, &alert)
	}
	return alerts
}

// CheckRateLimit checks if a peer is within rate limits
func (k Keeper) CheckRateLimit(ctx sdk.Context, peerId string) error {
	params := k.GetParams(ctx)
	rl, found := k.GetRateLimit(ctx, peerId)

	blockTime := ctx.BlockTime()
	if !found {
		// Create new rate limit entry
		rl = &securitypb.RateLimitEntry{
			PeerId:       peerId,
			RequestCount: 1,
			WindowStart:  timestamppb.New(blockTime),
			IsBanned:     false,
		}
		k.SetRateLimit(ctx, rl)
		return nil
	}

	// Check if banned
	if rl.IsBanned {
		if rl.BanExpiresAt != nil && blockTime.Before(rl.BanExpiresAt.AsTime()) {
			return types.ErrRateLimitExceeded
		}
		// Unban
		rl.IsBanned = false
		rl.RequestCount = 0
	}

	// Check if we're in a new window (1 second)
	if rl.WindowStart != nil && blockTime.Sub(rl.WindowStart.AsTime()).Seconds() >= 1 {
		rl.WindowStart = timestamppb.New(blockTime)
		rl.RequestCount = 1
	} else {
		rl.RequestCount++
	}

	// Check if rate exceeded
	if params.Network.RateLimit != nil && rl.RequestCount > params.Network.RateLimit.MaxRequestsPerSecond {
		rl.IsBanned = true
		blockedUntil := blockTime.Add(3600 * 1e9) // 1 hour
		rl.BanExpiresAt = timestamppb.New(blockedUntil)
		k.SetRateLimit(ctx, rl)
		return types.ErrRateLimitExceeded
	}

	k.SetRateLimit(ctx, rl)
	return nil
}
