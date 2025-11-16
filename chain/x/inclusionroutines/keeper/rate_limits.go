package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// GetRateLimit retrieves the rate limit configuration for an IR
func (k *Keeper) GetRateLimit(irID string) (types.IRRateLimit, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	limit, ok := k.rateLimits[irID]
	return limit, ok
}

// SetRateLimit sets the rate limit configuration for an IR
func (k *Keeper) SetRateLimit(limit types.IRRateLimit) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if IR exists
	if _, exists := k.irs[limit.IRID]; !exists {
		return types.ErrIRNotFound
	}

	// Validate rate limit values
	if limit.PerWalletPerHour < 0 || limit.PerWalletPerDay < 0 || limit.PerBlockGlobal < 0 {
		return types.ErrInvalidRateLimit
	}

	k.rateLimits[limit.IRID] = limit
	return nil
}

// CheckRateLimit checks if a wallet has exceeded the rate limit for an IR
// Returns an error if the rate limit is exceeded
func (k *Keeper) CheckRateLimit(wallet, irID string) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Get rate limit configuration
	limit, hasLimit := k.rateLimits[irID]
	if !hasLimit {
		// If no specific rate limit is configured, use default from params
		params := k.GetParams()
		limit = types.IRRateLimit{
			IRID:             irID,
			PerWalletPerHour: params.DefaultRateLimitHour,
			PerWalletPerDay:  params.DefaultRateLimitHour * 24,
			PerBlockGlobal:   0, // No global limit by default
		}
	}

	now := time.Now()

	// Check hourly limit
	if limit.PerWalletPerHour > 0 {
		hourKey := fmt.Sprintf("%s:%s:hour:%s", irID, wallet, now.Format("2006-01-02-15"))
		hourUsage := k.rateLimitUsage[hourKey]
		if hourUsage >= limit.PerWalletPerHour {
			return fmt.Errorf("%w: hourly limit of %d exceeded for wallet %s on IR %s",
				types.ErrRateLimitExceeded, limit.PerWalletPerHour, wallet, irID)
		}
	}

	// Check daily limit
	if limit.PerWalletPerDay > 0 {
		dayKey := fmt.Sprintf("%s:%s:day:%s", irID, wallet, now.Format("2006-01-02"))
		dayUsage := k.rateLimitUsage[dayKey]
		if dayUsage >= limit.PerWalletPerDay {
			return fmt.Errorf("%w: daily limit of %d exceeded for wallet %s on IR %s",
				types.ErrRateLimitExceeded, limit.PerWalletPerDay, wallet, irID)
		}
	}

	// Check per-block global limit
	if limit.PerBlockGlobal > 0 {
		blockKey := fmt.Sprintf("%s:global:block:%d", irID, now.Unix()/10) // 10-second blocks
		blockUsage := k.rateLimitUsage[blockKey]
		if blockUsage >= limit.PerBlockGlobal {
			return fmt.Errorf("%w: global per-block limit of %d exceeded for IR %s",
				types.ErrRateLimitExceeded, limit.PerBlockGlobal, irID)
		}
	}

	return nil
}

// IncrementRateLimit increments the rate limit usage counters for a wallet and IR
func (k *Keeper) IncrementRateLimit(wallet, irID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Get rate limit configuration
	limit, hasLimit := k.rateLimits[irID]
	if !hasLimit {
		// If no specific rate limit is configured, use default from params
		params := k.GetParams()
		limit = types.IRRateLimit{
			IRID:             irID,
			PerWalletPerHour: params.DefaultRateLimitHour,
			PerWalletPerDay:  params.DefaultRateLimitHour * 24,
			PerBlockGlobal:   0,
		}
	}

	now := time.Now()

	// Increment hourly counter
	if limit.PerWalletPerHour > 0 {
		hourKey := fmt.Sprintf("%s:%s:hour:%s", irID, wallet, now.Format("2006-01-02-15"))
		k.rateLimitUsage[hourKey]++
	}

	// Increment daily counter
	if limit.PerWalletPerDay > 0 {
		dayKey := fmt.Sprintf("%s:%s:day:%s", irID, wallet, now.Format("2006-01-02"))
		k.rateLimitUsage[dayKey]++
	}

	// Increment per-block global counter
	if limit.PerBlockGlobal > 0 {
		blockKey := fmt.Sprintf("%s:global:block:%d", irID, now.Unix()/10)
		k.rateLimitUsage[blockKey]++
	}

	return nil
}

// CleanupExpiredRateLimits removes expired rate limit usage counters
// This should be called periodically to prevent memory growth
func (k *Keeper) CleanupExpiredRateLimits() {
	k.mu.Lock()
	defer k.mu.Unlock()

	now := time.Now()
	cutoffHour := now.Add(-2 * time.Hour).Format("2006-01-02-15")
	cutoffDay := now.Add(-48 * time.Hour).Format("2006-01-02")

	for key := range k.rateLimitUsage {
		// Check if this is an expired hourly counter
		if len(key) > 10 && key[len(key)-13:len(key)-10] == "our" {
			timestamp := key[len(key)-13:]
			if timestamp < cutoffHour {
				delete(k.rateLimitUsage, key)
			}
		}

		// Check if this is an expired daily counter
		if len(key) > 10 && key[len(key)-10:len(key)-7] == "day" {
			timestamp := key[len(key)-10:]
			if timestamp < cutoffDay {
				delete(k.rateLimitUsage, key)
			}
		}

		// Check if this is an expired block counter
		// Keep approximately 3 hours of block data (1000 blocks at ~10s each)
		if len(key) > 6 && key[len(key)-6:] == "block:" {
			// This is a simplified cleanup; in production, parse the actual block number
			delete(k.rateLimitUsage, key)
		}
	}
}

// GetRateLimitUsage returns the current usage for debugging/monitoring
func (k *Keeper) GetRateLimitUsage(wallet, irID string) (hourly, daily int32) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	now := time.Now()
	hourKey := fmt.Sprintf("%s:%s:hour:%s", irID, wallet, now.Format("2006-01-02-15"))
	dayKey := fmt.Sprintf("%s:%s:day:%s", irID, wallet, now.Format("2006-01-02"))

	return k.rateLimitUsage[hourKey], k.rateLimitUsage[dayKey]
}
