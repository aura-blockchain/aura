package keeper

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// GetOrCreateRateLimitConfig gets or creates a rate limit config for a user
func (k *Keeper) GetOrCreateRateLimitConfig(userAddress string) *authproto.RateLimitConfig {
	k.mu.Lock()
	defer k.mu.Unlock()

	config, ok := k.rateLimits[userAddress]
	if !ok {
		now := time.Now()
		config = &authproto.RateLimitConfig{
			UserAddress:        userAddress,
			RequestsPerMinute:  k.params.DefaultRequestsPerMinute,
			RequestsPerHour:    k.params.DefaultRequestsPerHour,
			RequestsPerDay:     k.params.DefaultRequestsPerDay,
			CurrentMinuteCount: 0,
			CurrentHourCount:   0,
			CurrentDayCount:    0,
			WindowStart:        &now,
		}
		k.rateLimits[userAddress] = config
	}

	return config
}

// GetRateLimitConfig retrieves the rate limit config for a user

// CheckRateLimit checks if a user has exceeded their rate limit
func (k *Keeper) CheckRateLimit(ctx context.Context, userAddress string) error {
	config := k.GetOrCreateRateLimitConfig(userAddress)

	// Reset windows if needed
	k.resetRateLimitWindowsIfNeeded(config)

	// Check if rate limited
	if types.IsRateLimited(config) {
		k.LogAudit(ctx, userAddress, "rate_limit_check", userAddress, "limited", map[string]string{
			"minute_count": fmt.Sprintf("%d/%d", config.CurrentMinuteCount, config.RequestsPerMinute),
			"hour_count":   fmt.Sprintf("%d/%d", config.CurrentHourCount, config.RequestsPerHour),
			"day_count":    fmt.Sprintf("%d/%d", config.CurrentDayCount, config.RequestsPerDay),
		}, "rate limit exceeded")
		return types.ErrRateLimitExceeded
	}

	// Increment counters
	k.mu.Lock()
	config.CurrentMinuteCount++
	config.CurrentHourCount++
	config.CurrentDayCount++
	k.mu.Unlock()

	return nil
}

// resetRateLimitWindowsIfNeeded resets rate limit counters if windows have passed
func (k *Keeper) resetRateLimitWindowsIfNeeded(config *authproto.RateLimitConfig) {
	k.mu.Lock()
	defer k.mu.Unlock()

	now := time.Now()
	windowStart := config.WindowStart

	if windowStart == nil {
		config.WindowStart = timestamppb.New(now)
		return
	}

	// Reset minute counter
	if now.Sub(windowStart.AsTime()) >= time.Minute {
		config.CurrentMinuteCount = 0
	}

	// Reset hour counter
	if now.Sub(windowStart.AsTime()) >= time.Hour {
		config.CurrentHourCount = 0
	}

	// Reset day counter and window start
	if now.Sub(windowStart.AsTime()) >= 24*time.Hour {
		config.CurrentDayCount = 0
		config.WindowStart = timestamppb.New(now)
	}
}

// SetCustomRateLimit sets a custom rate limit for a user
func (k *Keeper) SetCustomRateLimit(ctx context.Context, setter, userAddress string, requestsPerMinute, requestsPerHour, requestsPerDay uint64) error {
	// Validate setter has permission
	if err := k.RequirePermission(ctx, setter, types.PermissionAdmin); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	config := k.rateLimits[userAddress]
	if config == nil {
		now := time.Now()
		config = &authproto.RateLimitConfig{
			UserAddress:        userAddress,
			CurrentMinuteCount: 0,
			CurrentHourCount:   0,
			CurrentDayCount:    0,
			WindowStart:        &now,
		}
		k.rateLimits[userAddress] = config
	}

	config.RequestsPerMinute = requestsPerMinute
	config.RequestsPerHour = requestsPerHour
	config.RequestsPerDay = requestsPerDay

	k.LogAudit(ctx, setter, "set_custom_rate_limit", userAddress, "success", map[string]string{
		"requests_per_minute": fmt.Sprintf("%d", requestsPerMinute),
		"requests_per_hour":   fmt.Sprintf("%d", requestsPerHour),
		"requests_per_day":    fmt.Sprintf("%d", requestsPerDay),
	}, "")

	return nil
}

// ResetRateLimit resets the rate limit counters for a user
func (k *Keeper) ResetRateLimit(ctx context.Context, resetter, userAddress string) error {
	// Validate resetter has permission
	if err := k.RequirePermission(ctx, resetter, types.PermissionAdmin); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	config, ok := k.rateLimits[userAddress]
	if !ok {
		return types.ErrInvalidRateLimitConfig
	}

	now := time.Now()
	config.CurrentMinuteCount = 0
	config.CurrentHourCount = 0
	config.CurrentDayCount = 0
	config.WindowStart = &now

	k.LogAudit(ctx, resetter, "reset_rate_limit", userAddress, "success", nil, "")

	return nil
}

// GetRateLimitStatus returns the current rate limit status for a user
func (k *Keeper) GetRateLimitStatus(userAddress string) (*authproto.RateLimitConfig, bool) {
	config := k.GetOrCreateRateLimitConfig(userAddress)
	isLimited := types.IsRateLimited(config)
	return config, isLimited
}

// IncrementRateLimit increments the rate limit counter for a user
func (k *Keeper) IncrementRateLimit(userAddress string) {
	config := k.GetOrCreateRateLimitConfig(userAddress)

	k.resetRateLimitWindowsIfNeeded(config)

	k.mu.Lock()
	config.CurrentMinuteCount++
	config.CurrentHourCount++
	config.CurrentDayCount++
	k.mu.Unlock()
}

// BypassRateLimit temporarily bypasses rate limiting for a user
func (k *Keeper) BypassRateLimit(ctx context.Context, setter, userAddress string, durationSeconds uint64) error {
	// Validate setter has admin permission
	if err := k.RequirePermission(ctx, setter, types.PermissionAdmin); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	config, ok := k.rateLimits[userAddress]
	if !ok {
		now := time.Now()
		config = &authproto.RateLimitConfig{
			UserAddress: userAddress,
			WindowStart: &now,
		}
		k.rateLimits[userAddress] = config
	}

	// Set extremely high limits for the duration
	config.RequestsPerMinute = 1000000
	config.RequestsPerHour = 60000000
	config.RequestsPerDay = 1440000000

	k.LogAudit(ctx, setter, "bypass_rate_limit", userAddress, "success", map[string]string{
		"duration_seconds": fmt.Sprintf("%d", durationSeconds),
	}, "")

	// In a real implementation, we would schedule a task to reset the limits after the duration

	return nil
}

// GetTopRateLimitedUsers returns users with the highest request counts
func (k *Keeper) GetTopRateLimitedUsers(limit int) []*authproto.RateLimitConfig {
	k.mu.RLock()
	defer k.mu.RUnlock()

	configs := make([]*authproto.RateLimitConfig, 0, len(k.rateLimits))
	for _, config := range k.rateLimits {
		configs = append(configs, config)
	}

	// Sort by total requests (simple bubble sort for small datasets)
	for i := 0; i < len(configs); i++ {
		for j := i + 1; j < len(configs); j++ {
			if configs[i].CurrentDayCount < configs[j].CurrentDayCount {
				configs[i], configs[j] = configs[j], configs[i]
			}
		}
	}

	if limit > 0 && limit < len(configs) {
		configs = configs[:limit]
	}

	return configs
}

// CleanupInactiveRateLimits removes rate limit configs that haven't been used recently
func (k *Keeper) CleanupInactiveRateLimits(inactiveDays int) int {
	k.mu.Lock()
	defer k.mu.Unlock()

	count := 0
	now := time.Now()
	threshold := time.Duration(inactiveDays) * 24 * time.Hour

	for userAddress, config := range k.rateLimits {
		if config.WindowStart != nil && now.Sub(config.WindowStart.AsTime()) > threshold {
			delete(k.rateLimits, userAddress)
			count++
		}
	}

	return count
}
