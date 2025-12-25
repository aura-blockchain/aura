// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"sync"
	"time"

	storetypes "cosmossdk.io/store/types"
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

// NewRateLimiter creates a new rate limiter.
// currentTime must be ctx.BlockTime() from Cosmos SDK context for deterministic consensus.
func NewRateLimiter(maxRate, burstSize uint64, windowDuration time.Duration, currentTime time.Time) *RateLimiter {
	return &RateLimiter{
		maxRate:        maxRate,
		burstSize:      burstSize,
		tokens:         burstSize,
		lastRefill:     currentTime,
		windowDuration: windowDuration,
		windowStart:    currentTime,
	}
}

// Allow checks if a request should be allowed.
// currentTime must be ctx.BlockTime() from Cosmos SDK context for deterministic consensus.
// Uses integer math for deterministic consensus.
func (rl *RateLimiter) Allow(currentTime time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on elapsed time using integer math
	elapsed := currentTime.Sub(rl.lastRefill)
	// Convert to seconds using integer division (nanoseconds / 1e9)
	elapsedSeconds := uint64(elapsed.Nanoseconds() / 1e9)
	tokensToAdd := elapsedSeconds * rl.maxRate
	if tokensToAdd > 0 {
		rl.tokens = min(rl.tokens+tokensToAdd, rl.burstSize)
		rl.lastRefill = currentTime
	}

	// Check window-based rate limiting
	if currentTime.Sub(rl.windowStart) > rl.windowDuration {
		rl.windowStart = currentTime
		rl.requestsInWindow = 0
	}

	// Check both token bucket and window limit
	// windowDurationSeconds = windowDuration in nanoseconds / 1e9
	windowDurationSeconds := uint64(rl.windowDuration.Nanoseconds() / 1e9)
	if rl.tokens > 0 && rl.requestsInWindow < rl.maxRate*windowDurationSeconds {
		rl.tokens--
		rl.requestsInWindow++
		return true
	}

	return false
}

// Reset resets the rate limiter.
// currentTime must be ctx.BlockTime() from Cosmos SDK context for deterministic consensus.
func (rl *RateLimiter) Reset(currentTime time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.tokens = rl.burstSize
	rl.lastRefill = currentTime
	rl.windowStart = currentTime
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
		params.RateLimit.WindowDuration,
		ctx.BlockTime(),
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
	if !limiter.Allow(ctx.BlockTime()) {
		// Rate limit exceeded, increment violation count
		params, _ := k.GetParams(ctx)

		// Get or create rate limit entry
		entry, found := k.GetRateLimitEntry(ctx, peerID)
		if !found {
			entry = types.RateLimitEntry{
				PeerId:      peerID,
				WindowStart: ctx.BlockTime(),
			}
		}

		entry.RequestCount++

		// Check if we should ban this peer
		if entry.RequestCount > params.RateLimit.MaxRequestsPerSecond*10 {
			// Too many violations, ban the peer
			banDurationSecs := int64(params.RateLimit.BanDuration.Seconds())
			if err := k.BanPeer(ctx, peerID, banDurationSecs, "rate limit violations"); err != nil {
				return fmt.Errorf("failed to Seconds: %w", err)
			}

			// Penalize reputation
			k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty)
		}

		if err := k.SetRateLimitEntry(ctx, entry); err != nil {
			return fmt.Errorf("failed to Seconds: %w", err)
		}
		return types.ErrRateLimitExceeded
	}

	// Update rate limit entry
	entry, found := k.GetRateLimitEntry(ctx, peerID)
	if !found {
		entry = types.RateLimitEntry{
			PeerId:      peerID,
			WindowStart: ctx.BlockTime(),
		}
	}
	entry.RequestCount++
	if err := k.SetRateLimitEntry(ctx, entry); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

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

// NewBandwidthTracker creates a new bandwidth tracker.
// currentTime must be ctx.BlockTime() from Cosmos SDK context for deterministic consensus.
func NewBandwidthTracker(limit uint64, windowSize time.Duration, currentTime time.Time) *BandwidthTracker {
	return &BandwidthTracker{
		windowStart: currentTime,
		windowSize:  windowSize,
		limit:       limit,
	}
}

// RecordSent records bytes sent.
// currentTime must be ctx.BlockTime() from Cosmos SDK context for deterministic consensus.
func (bt *BandwidthTracker) RecordSent(bytes uint64, currentTime time.Time) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if currentTime.Sub(bt.windowStart) > bt.windowSize {
		bt.bytesSent = 0
		bt.bytesRecv = 0
		bt.windowStart = currentTime
	}

	bt.bytesSent += bytes
}

// RecordReceived records bytes received.
// currentTime must be ctx.BlockTime() from Cosmos SDK context for deterministic consensus.
func (bt *BandwidthTracker) RecordReceived(bytes uint64, currentTime time.Time) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if currentTime.Sub(bt.windowStart) > bt.windowSize {
		bt.bytesSent = 0
		bt.bytesRecv = 0
		bt.windowStart = currentTime
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
		params.RateLimit.WindowDuration,
		ctx.BlockTime(),
	)
	k.bandwidthTrackers[peerID] = tracker
	return tracker
}

// CheckBandwidthLimit checks if peer's bandwidth usage is within limits
func (k Keeper) CheckBandwidthLimit(ctx sdk.Context, peerID string, bytes uint64, isSending bool) error {
	tracker := k.GetBandwidthTracker(ctx, peerID)

	if isSending {
		tracker.RecordSent(bytes, ctx.BlockTime())
	} else {
		tracker.RecordReceived(bytes, ctx.BlockTime())
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
				WindowStart: ctx.BlockTime(),
			}
		}

		sent, recv := tracker.GetStats()
		entry.BytesSent = sent
		entry.BytesReceived = recv
		if err := k.SetRateLimitEntry(ctx, entry); err != nil {
			return fmt.Errorf("error in CheckBandwidthLimit for PeerId: %w", err)
		}

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

	toDelete := make([]string, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var entry types.RateLimitEntry
		if err := k.cdc.Unmarshal(iterator.Value(), &entry); err != nil {
			k.logger.Error("failed to unmarshal rate limit entry", "error", err)
			continue
		}

		// Check if window has expired
		if !entry.WindowStart.IsZero() {
			windowDuration := params.RateLimit.WindowDuration
			if ctx.BlockTime().Sub(entry.WindowStart) > windowDuration*2 {
				toDelete = append(toDelete, entry.PeerId)
			}
		}

		// Check if ban has expired
		if entry.IsBanned && entry.BanExpiresAt != nil {
			if ctx.BlockTime().After(*entry.BanExpiresAt) {
				if err := k.UnbanPeer(ctx, entry.PeerId); err != nil {
					k.logger.Error("failed to unban peer during cleanup", "peer", entry.PeerId, "err", err)
				}
			}
		}
	}

	// Delete expired entries
	for _, peerID := range toDelete {
		if err := store.Delete(types.GetRateLimitKey(peerID)); err != nil {
			k.logger.Error("failed to delete expired rate limit entry", "peer", peerID, "err", err)
		}
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
		return fmt.Errorf("error in DDosProtectionCheck: %w", err)
	}

	// 3. Check bandwidth limits
	if err := k.CheckBandwidthLimit(ctx, peerID, messageSize, false); err != nil {
		k.logger.Warn(fmt.Sprintf("Bandwidth check failed for peer %s: %v", peerID, err))
		return fmt.Errorf("error in DDosProtectionCheck for bandwidth: %w", err)
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
