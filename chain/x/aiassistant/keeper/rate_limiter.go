// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// RateLimitConfig defines rate limiting parameters for AI queries
type RateLimitConfig struct {
	MaxQueriesPerMinute uint64
	MaxQueriesPerHour   uint64
	MaxQueriesPerDay    uint64
	BurstAllowance      uint64
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxQueriesPerMinute: 60,
		MaxQueriesPerHour:   1000,
		MaxQueriesPerDay:    10000,
		BurstAllowance:      10,
	}
}

// QueryUsage tracks query usage for rate limiting
type QueryUsage struct {
	Address      string
	LastMinute   uint64
	LastHour     uint64
	LastDay      uint64
	LastReset    time.Time
	TotalQueries uint64
}

// CheckRateLimit validates if the user can perform an AI query
func (k Keeper) CheckRateLimit(ctx sdk.Context, address string) error {
	config := DefaultRateLimitConfig()

	usage := k.getQueryUsage(ctx, address)
	now := ctx.BlockTime()

	// Reset counters based on time windows
	if now.Sub(usage.LastReset) >= 24*time.Hour {
		usage.LastDay = 0
		usage.LastHour = 0
		usage.LastMinute = 0
		usage.LastReset = now
	} else if now.Sub(usage.LastReset) >= time.Hour {
		usage.LastHour = 0
		usage.LastMinute = 0
	} else if now.Sub(usage.LastReset) >= time.Minute {
		usage.LastMinute = 0
	}

	// Check rate limits
	if usage.LastMinute >= config.MaxQueriesPerMinute {
		return sdkerrors.Wrap(types.ErrInvalidParams,
			fmt.Sprintf("rate limit exceeded: %d queries per minute", config.MaxQueriesPerMinute))
	}
	if usage.LastHour >= config.MaxQueriesPerHour {
		return sdkerrors.Wrap(types.ErrInvalidParams,
			fmt.Sprintf("rate limit exceeded: %d queries per hour", config.MaxQueriesPerHour))
	}
	if usage.LastDay >= config.MaxQueriesPerDay {
		return sdkerrors.Wrap(types.ErrInvalidParams,
			fmt.Sprintf("rate limit exceeded: %d queries per day", config.MaxQueriesPerDay))
	}

	return nil
}

// IncrementQueryUsage increments query usage counters
func (k Keeper) IncrementQueryUsage(ctx sdk.Context, address string) {
	usage := k.getQueryUsage(ctx, address)

	usage.LastMinute++
	usage.LastHour++
	usage.LastDay++
	usage.TotalQueries++

	k.setQueryUsage(ctx, usage)
}

// getQueryUsage retrieves query usage for an address
func (k Keeper) getQueryUsage(ctx sdk.Context, address string) QueryUsage {
	store := ctx.KVStore(k.storeKey)
	key := types.QueryUsageKey(address)

	bz := store.Get(key)
	if len(bz) == 0 {
		return QueryUsage{
			Address:   address,
			LastReset: ctx.BlockTime(),
		}
	}

	var usage QueryUsage
	// Simple encoding - in production use protobuf
	if err := json.Unmarshal(bz, &usage); err != nil {
		return QueryUsage{
			Address:   address,
			LastReset: ctx.BlockTime(),
		}
	}

	return usage
}

// setQueryUsage stores query usage for an address
func (k Keeper) setQueryUsage(ctx sdk.Context, usage QueryUsage) {
	store := ctx.KVStore(k.storeKey)
	key := types.QueryUsageKey(usage.Address)

	bz, err := json.Marshal(&usage)
	if err != nil {
		return
	}

	store.Set(key, bz)
}

// GetQueryUsageStats returns query usage statistics
func (k Keeper) GetQueryUsageStats(ctx sdk.Context, address string) QueryUsage {
	return k.getQueryUsage(ctx, address)
}

// ResetRateLimit resets rate limit counters for an address (admin function)
func (k Keeper) ResetRateLimit(ctx sdk.Context, address string) error {
	store := ctx.KVStore(k.storeKey)
	key := types.QueryUsageKey(address)
	store.Delete(key)
	return nil
}

// ListQueryUsage returns all query usage records (for analytics)
func (k Keeper) ListQueryUsage(ctx sdk.Context) []QueryUsage {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.QueryUsageKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var usages []QueryUsage
	for ; iterator.Valid(); iterator.Next() {
		var usage QueryUsage
		if err := json.Unmarshal(iterator.Value(), &usage); err != nil {
			continue
		}
		usages = append(usages, usage)
	}

	return usages
}
