package keeper

import (
	"fmt"
	"sync"
	"time"

	storetypes "cosmossdk.io/store/types"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	mu               sync.Mutex
	maxRate          uint64        // Maximum requests per second
	burstSize        uint64        // Burst capacity
	tokens           uint64        // Current available tokens
	lastRefill       time.Time     // Last token refill time
	windowDuration   time.Duration // Rate limiting window
	requestsInWindow uint64        // Requests in current window
	windowStart      time.Time     // Current window start
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRate, burstSize uint64, windowDuration time.Duration) *RateLimiter {
	now := time.Now()
	return &RateLimiter{
		maxRate:        maxRate,
		burstSize:      burstSize,
		tokens:         burstSize,
		lastRefill:     now,
		windowDuration: windowDuration,
		windowStart:    now,
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Refill tokens based on elapsed time
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := uint64(elapsed.Seconds() * float64(rl.maxRate))
	if tokensToAdd > 0 {
		rl.tokens = min(rl.tokens+tokensToAdd, rl.burstSize)
		rl.lastRefill = now
	}

	// Check window-based rate limiting
	if now.Sub(rl.windowStart) > rl.windowDuration {
		rl.windowStart = now
		rl.requestsInWindow = 0
	}

	// Check both token bucket and window limit
	if rl.tokens > 0 && rl.requestsInWindow < rl.maxRate*uint64(rl.windowDuration.Seconds()) {
		rl.tokens--
		rl.requestsInWindow++
		return true
	}

	return false
}

// Reset resets the rate limiter
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.tokens = rl.burstSize
	rl.lastRefill = now
	rl.windowStart = now
	rl.requestsInWindow = 0
}

// GetRateLimiter gets or creates a rate limiter for a peer
func (k Keeper) GetRateLimiter(ctx sdk.Context, peerID string) *RateLimiter {
	if limiter, exists := k.rateLimiters[peerID]; exists {
		return limiter
	}

	params, _ := k.GetParams(ctx)
	limiter := NewRateLimiter(
		params.RateLimit.MaxRequestsPerSecond,
		params.RateLimit.BurstSize,
		params.RateLimit.WindowDuration.AsDuration(),
	)
	k.rateLimiters[peerID] = limiter
	return limiter
}

// CheckRateLimit checks if a peer's request should be allowed
func (k Keeper) CheckRateLimit(ctx sdk.Context, peerID string) error {
	// Check if peer is banned
	if k.IsBanned(ctx, peerID) {
		return types.ErrPeerBanned
	}

	// Get rate limiter
	limiter := k.GetRateLimiter(ctx, peerID)

	// Check if request is allowed
	if !limiter.Allow() {
		// Rate limit exceeded, increment violation count
		params, _ := k.GetParams(ctx)

		// Get or create rate limit entry
		entry, found := k.GetRateLimitEntry(ctx, peerID)
		if !found {
			entry = types.RateLimitEntry{
				PeerId:      peerID,
				WindowStart: timestamppb.New(ctx.BlockTime()),
			}
		}

		entry.RequestCount++

		// Check if we should ban this peer
		if entry.RequestCount > params.RateLimit.MaxRequestsPerSecond*10 {
			// Too many violations, ban the peer
			banDurationSecs := int64(params.RateLimit.BanDuration.AsDuration().Seconds())
			k.BanPeer(ctx, peerID, banDurationSecs, "rate limit violations")

			// Penalize reputation
			k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty)
		}

		k.SetRateLimitEntry(ctx, entry)
		return types.ErrRateLimitExceeded
	}

	// Update rate limit entry
	entry, found := k.GetRateLimitEntry(ctx, peerID)
	if !found {
		entry = types.RateLimitEntry{
			PeerId:      peerID,
			WindowStart: timestamppb.New(ctx.BlockTime()),
		}
	}
	entry.RequestCount++
	k.SetRateLimitEntry(ctx, entry)

	return nil
}

// BandwidthTracker tracks bandwidth usage per peer
type BandwidthTracker struct {
	mu          sync.Mutex
	bytesSent   uint64
	bytesRecv   uint64
	windowStart time.Time
	windowSize  time.Duration
	limit       uint64 // Bytes per second
}

// NewBandwidthTracker creates a new bandwidth tracker
func NewBandwidthTracker(limit uint64, windowSize time.Duration) *BandwidthTracker {
	return &BandwidthTracker{
		windowStart: time.Now(),
		windowSize:  windowSize,
		limit:       limit,
	}
}

// RecordSent records bytes sent
func (bt *BandwidthTracker) RecordSent(bytes uint64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	now := time.Now()
	if now.Sub(bt.windowStart) > bt.windowSize {
		bt.bytesSent = 0
		bt.bytesRecv = 0
		bt.windowStart = now
	}

	bt.bytesSent += bytes
}

// RecordReceived records bytes received
func (bt *BandwidthTracker) RecordReceived(bytes uint64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	now := time.Now()
	if now.Sub(bt.windowStart) > bt.windowSize {
		bt.bytesSent = 0
		bt.bytesRecv = 0
		bt.windowStart = now
	}

	bt.bytesRecv += bytes
}

// CheckLimit checks if bandwidth limit is exceeded
func (bt *BandwidthTracker) CheckLimit() bool {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	totalBytes := bt.bytesSent + bt.bytesRecv
	limitBytes := bt.limit * uint64(bt.windowSize.Seconds())
	return totalBytes <= limitBytes
}

// GetStats returns current bandwidth statistics
func (bt *BandwidthTracker) GetStats() (sent, recv uint64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	return bt.bytesSent, bt.bytesRecv
}

// GetBandwidthTracker gets or creates a bandwidth tracker for a peer
func (k Keeper) GetBandwidthTracker(ctx sdk.Context, peerID string) *BandwidthTracker {
	if tracker, exists := k.bandwidthTrackers[peerID]; exists {
		return tracker
	}

	params, _ := k.GetParams(ctx)
	tracker := NewBandwidthTracker(
		params.RateLimit.BandwidthLimitPerPeer,
		params.RateLimit.WindowDuration.AsDuration(),
	)
	k.bandwidthTrackers[peerID] = tracker
	return tracker
}

// CheckBandwidthLimit checks if peer's bandwidth usage is within limits
func (k Keeper) CheckBandwidthLimit(ctx sdk.Context, peerID string, bytes uint64, isSending bool) error {
	tracker := k.GetBandwidthTracker(ctx, peerID)

	if isSending {
		tracker.RecordSent(bytes)
	} else {
		tracker.RecordReceived(bytes)
	}

	if !tracker.CheckLimit() {
		params, _ := k.GetParams(ctx)

		// Bandwidth limit exceeded, penalize and potentially ban
		k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty/2)

		// Update rate limit entry with bandwidth stats
		entry, found := k.GetRateLimitEntry(ctx, peerID)
		if !found {
			entry = types.RateLimitEntry{
				PeerId:      peerID,
				WindowStart: timestamppb.New(ctx.BlockTime()),
			}
		}

		sent, recv := tracker.GetStats()
		entry.BytesSent = sent
		entry.BytesReceived = recv
		k.SetRateLimitEntry(ctx, entry)

		return types.ErrBandwidthLimitExceeded
	}

	return nil
}

// CleanupExpiredRateLimits removes expired rate limit entries and bans
func (k Keeper) CleanupExpiredRateLimits(ctx sdk.Context) {
	params, _ := k.GetParams(ctx)
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.RateLimitPrefix, storetypes.PrefixEndBytes(types.RateLimitPrefix))
	if err != nil {
		return
	}
	defer iterator.Close()

	var toDelete []string
	for ; iterator.Valid(); iterator.Next() {
		var entry types.RateLimitEntry
		k.cdc.MustUnmarshal(iterator.Value(), &entry)

		// Check if window has expired
		if entry.WindowStart != nil {
			windowDuration := params.RateLimit.WindowDuration.AsDuration()
			if ctx.BlockTime().Sub(entry.WindowStart.AsTime()) > windowDuration*2 {
				toDelete = append(toDelete, entry.PeerId)
			}
		}

		// Check if ban has expired
		if entry.IsBanned && entry.BanExpiresAt != nil {
			if ctx.BlockTime().After(entry.BanExpiresAt.AsTime()) {
				k.UnbanPeer(ctx, entry.PeerId)
			}
		}
	}

	// Delete expired entries
	for _, peerID := range toDelete {
		store.Delete(types.GetRateLimitKey(peerID))
		delete(k.rateLimiters, peerID)
		delete(k.bandwidthTrackers, peerID)
	}
}

// DDosProtectionCheck performs comprehensive DDoS protection checks
func (k Keeper) DDosProtectionCheck(ctx sdk.Context, peerID string, messageSize uint64) error {
	// 1. Check if peer is banned
	if k.IsBanned(ctx, peerID) {
		return types.ErrPeerBanned
	}

	// 2. Check rate limiting
	if err := k.CheckRateLimit(ctx, peerID); err != nil {
		k.logger.Warn(fmt.Sprintf("Rate limit check failed for peer %s: %v", peerID, err))
		return err
	}

	// 3. Check bandwidth limits
	if err := k.CheckBandwidthLimit(ctx, peerID, messageSize, false); err != nil {
		k.logger.Warn(fmt.Sprintf("Bandwidth check failed for peer %s: %v", peerID, err))
		return err
	}

	// 4. Check message size
	params, _ := k.GetParams(ctx)
	if messageSize > params.Gossip.MaxMessageSize {
		k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty)
		return types.ErrMessageTooLarge
	}

	return nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
