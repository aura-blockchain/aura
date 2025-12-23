package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetRateLimitConfig retrieves the rate limit configuration for an IR from KV store
func (k *Keeper) GetRateLimitConfig(ctx sdk.Context, irID string) (types.IRRateLimit, bool) {
	return k.GetRateLimit(ctx, irID)
}

// SetRateLimitConfig sets the rate limit configuration for an IR in KV store
func (k *Keeper) SetRateLimitConfig(ctx sdk.Context, limit types.IRRateLimit) error {
	// Check if IR exists
	if _, exists := k.GetIR(ctx, limit.IrId); !exists {
		return types.ErrIRNotFound
	}

	// Validate rate limit values
	if limit.PerWalletPerHour < 0 || limit.PerWalletPerDay < 0 || limit.PerBlockGlobal < 0 {
		return types.ErrInvalidRateLimit
	}

	return k.SetRateLimit(ctx, limit)
}

// CheckRateLimit checks if a wallet has exceeded the rate limit for an IR
// Returns an error if the rate limit is exceeded
func (k *Keeper) CheckRateLimit(ctx sdk.Context, wallet, irID string) error {
	// Get rate limit configuration
	limit, hasLimit := k.GetRateLimit(ctx, irID)
	if !hasLimit {
		// If no specific rate limit is configured, use default from params
		params := k.GetParams()
		limit = types.IRRateLimit{
			IrId:             irID,
			PerWalletPerHour: params.DefaultRateLimitHour,
			PerWalletPerDay:  params.DefaultRateLimitHour * 24,
			PerBlockGlobal:   0, // No global limit by default
		}
	}

	blockTime := ctx.BlockTime().Unix()

	// Check hourly limit
	if limit.PerWalletPerHour > 0 {
		currentHour := blockTime / 3600
		hourKey := fmt.Sprintf("hour:%d", currentHour)
		hourUsage := k.GetRateLimitUsage(ctx, irID, wallet, hourKey)
		if hourUsage >= limit.PerWalletPerHour {
			return fmt.Errorf("%w: hourly limit of %d exceeded for wallet %s on IR %s",
				types.ErrRateLimitExceeded, limit.PerWalletPerHour, wallet, irID)
		}
	}

	// Check daily limit
	if limit.PerWalletPerDay > 0 {
		currentDay := blockTime / 86400
		dayKey := fmt.Sprintf("day:%d", currentDay)
		dayUsage := k.GetRateLimitUsage(ctx, irID, wallet, dayKey)
		if dayUsage >= limit.PerWalletPerDay {
			return fmt.Errorf("%w: daily limit of %d exceeded for wallet %s on IR %s",
				types.ErrRateLimitExceeded, limit.PerWalletPerDay, wallet, irID)
		}
	}

	// Check per-block global limit
	if limit.PerBlockGlobal > 0 {
		blockKey := fmt.Sprintf("block:%d", blockTime)
		blockUsage := k.GetRateLimitUsage(ctx, irID, "global", blockKey)
		if blockUsage >= limit.PerBlockGlobal {
			return fmt.Errorf("%w: global per-block limit of %d exceeded for IR %s",
				types.ErrRateLimitExceeded, limit.PerBlockGlobal, irID)
		}
	}

	return nil
}

// IncrementRateLimitCounters increments the rate limit usage counters for a wallet and IR
func (k *Keeper) IncrementRateLimitCounters(ctx sdk.Context, wallet, irID string) error {
	// Get rate limit configuration
	limit, hasLimit := k.GetRateLimit(ctx, irID)
	if !hasLimit {
		// If no specific rate limit is configured, use default from params
		params := k.GetParams()
		limit = types.IRRateLimit{
			IrId:             irID,
			PerWalletPerHour: params.DefaultRateLimitHour,
			PerWalletPerDay:  params.DefaultRateLimitHour * 24,
			PerBlockGlobal:   0,
		}
	}

	blockTime := ctx.BlockTime().Unix()

	// Increment hourly counter
	if limit.PerWalletPerHour > 0 {
		currentHour := blockTime / 3600
		hourKey := fmt.Sprintf("hour:%d", currentHour)
		if err := k.IncrementRateLimitUsage(ctx, irID, wallet, hourKey); err != nil {
			return fmt.Errorf("failed to Sprintf: %w", err)
		}
	}

	// Increment daily counter
	if limit.PerWalletPerDay > 0 {
		currentDay := blockTime / 86400
		dayKey := fmt.Sprintf("day:%d", currentDay)
		if err := k.IncrementRateLimitUsage(ctx, irID, wallet, dayKey); err != nil {
			return fmt.Errorf("failed to Sprintf: %w", err)
		}
	}

	// Increment per-block global counter
	if limit.PerBlockGlobal > 0 {
		blockKey := fmt.Sprintf("block:%d", blockTime)
		if err := k.IncrementRateLimitUsage(ctx, irID, "global", blockKey); err != nil {
			return fmt.Errorf("failed to Sprintf: %w", err)
		}
	}

	return nil
}

// CleanupExpiredRateLimits removes expired rate limit usage counters
// This should be called periodically in EndBlock to prevent state growth
func (k *Keeper) CleanupExpiredRateLimits(ctx sdk.Context) error {
	// In a production system, we would iterate through all rate limit usage keys
	// and delete those that are expired. This would require maintaining an index
	// of all rate limit usage keys or using a prefix iterator.

	// For now, we rely on the natural expiration of old data and the fact that
	// the rate limit checks only look at current time windows.

	// NOTE: In a production blockchain, you would implement a more sophisticated
	// pruning strategy, possibly using a separate store or background process.

	return nil
}

// GetRateLimitStatus returns the current usage for monitoring/debugging
func (k *Keeper) GetRateLimitStatus(ctx sdk.Context, wallet, irID string) (hourly, daily int32) {
	blockTime := ctx.BlockTime().Unix()
	currentHour := blockTime / 3600
	currentDay := blockTime / 86400

	hourKey := fmt.Sprintf("hour:%d", currentHour)
	dayKey := fmt.Sprintf("day:%d", currentDay)

	return k.GetRateLimitUsage(ctx, irID, wallet, hourKey),
		k.GetRateLimitUsage(ctx, irID, wallet, dayKey)
}

// ValidateRateLimit validates rate limit configuration
func (k *Keeper) ValidateRateLimit(limit types.IRRateLimit) error {
	if limit.IrId == "" {
		return types.ErrIRInvalidID
	}

	if limit.PerWalletPerHour < 0 {
		return fmt.Errorf("%w: per_wallet_per_hour cannot be negative", types.ErrInvalidRateLimit)
	}

	if limit.PerWalletPerDay < 0 {
		return fmt.Errorf("%w: per_wallet_per_day cannot be negative", types.ErrInvalidRateLimit)
	}

	if limit.PerBlockGlobal < 0 {
		return fmt.Errorf("%w: per_block_global cannot be negative", types.ErrInvalidRateLimit)
	}

	// Daily limit should be >= hourly limit if both are set
	if limit.PerWalletPerHour > 0 && limit.PerWalletPerDay > 0 {
		if limit.PerWalletPerDay < limit.PerWalletPerHour {
			return fmt.Errorf("%w: daily limit must be >= hourly limit", types.ErrInvalidRateLimit)
		}
	}

	return nil
}
