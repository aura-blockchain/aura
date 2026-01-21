// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// CachedResponse represents a cached AI response
type CachedResponse struct {
	CacheKey  string
	QueryHash string
	Response  string
	ModelHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	HitCount  uint64
	Size      uint64
	Status    CacheStatus
}

// CacheStatus defines cache entry status
type CacheStatus string

const (
	CacheStatusValid   CacheStatus = "valid"
	CacheStatusExpired CacheStatus = "expired"
	CacheStatusStale   CacheStatus = "stale"
)

// CacheConfig defines caching configuration
type CacheConfig struct {
	Enabled        bool
	DefaultTTL     time.Duration
	MaxCacheSize   uint64
	MaxEntrySize   uint64
	EvictionPolicy string
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:        true,
		DefaultTTL:     24 * time.Hour,
		MaxCacheSize:   1000000, // 1M entries
		MaxEntrySize:   1048576, // 1MB per entry
		EvictionPolicy: "lru",   // Least Recently Used
	}
}

// GenerateQueryHash creates a deterministic hash for a query
func GenerateQueryHash(query string, modelHash string, params map[string]string) string {
	hasher := sha256.New()
	hasher.Write([]byte(query))
	hasher.Write([]byte(modelHash))

	// Add sorted params for determinism
	for k, v := range params {
		hasher.Write([]byte(k))
		hasher.Write([]byte(v))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// GetCachedResponse retrieves a cached response if valid
func (k Keeper) GetCachedResponse(ctx sdk.Context, queryHash string) (CachedResponse, bool) {
	config := DefaultCacheConfig()
	if !config.Enabled {
		return CachedResponse{}, false
	}

	store := ctx.KVStore(k.storeKey)
	key := types.CacheKey(queryHash)

	bz := store.Get(key)
	if len(bz) == 0 {
		return CachedResponse{}, false
	}

	var cached CachedResponse
	if err := json.Unmarshal(bz, &cached); err != nil {
		return CachedResponse{}, false
	}

	// Check expiration
	if ctx.BlockTime().After(cached.ExpiresAt) {
		cached.Status = CacheStatusExpired
		k.deleteCachedResponse(ctx, queryHash)
		return CachedResponse{}, false
	}

	// Update hit count
	cached.HitCount++
	k.setCachedResponse(ctx, cached)

	return cached, true
}

// SetCachedResponse stores a response in cache
func (k Keeper) SetCachedResponse(ctx sdk.Context, queryHash, response, modelHash string) error {
	config := DefaultCacheConfig()
	if !config.Enabled {
		return nil
	}

	// Validate size
	responseSize := uint64(len(response))
	if responseSize > config.MaxEntrySize {
		return fmt.Errorf("response too large for cache: %d > %d", responseSize, config.MaxEntrySize)
	}

	cached := CachedResponse{
		CacheKey:  queryHash,
		QueryHash: queryHash,
		Response:  response,
		ModelHash: modelHash,
		CreatedAt: ctx.BlockTime(),
		ExpiresAt: ctx.BlockTime().Add(config.DefaultTTL),
		HitCount:  0,
		Size:      responseSize,
		Status:    CacheStatusValid,
	}

	k.setCachedResponse(ctx, cached)

	// Check cache size and evict if needed
	k.evictIfNeeded(ctx, config)

	return nil
}

// setCachedResponse stores cached response
func (k Keeper) setCachedResponse(ctx sdk.Context, cached CachedResponse) {
	store := ctx.KVStore(k.storeKey)
	key := types.CacheKey(cached.QueryHash)

	bz, err := json.Marshal(&cached)
	if err != nil {
		return
	}

	store.Set(key, bz)
}

// deleteCachedResponse removes a cached response
func (k Keeper) deleteCachedResponse(ctx sdk.Context, queryHash string) {
	store := ctx.KVStore(k.storeKey)
	key := types.CacheKey(queryHash)
	store.Delete(key)
}

// InvalidateCache invalidates cached responses for a model
func (k Keeper) InvalidateCache(ctx sdk.Context, modelHash string) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CacheKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	count := uint64(0)
	for ; iterator.Valid(); iterator.Next() {
		var cached CachedResponse
		if err := json.Unmarshal(iterator.Value(), &cached); err != nil {
			continue
		}

		if cached.ModelHash == modelHash {
			store.Delete(iterator.Key())
			count++
		}
	}

	return count
}

// ClearExpiredCache removes all expired cache entries
func (k Keeper) ClearExpiredCache(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CacheKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	count := uint64(0)
	now := ctx.BlockTime()

	for ; iterator.Valid(); iterator.Next() {
		var cached CachedResponse
		if err := json.Unmarshal(iterator.Value(), &cached); err != nil {
			continue
		}

		if now.After(cached.ExpiresAt) {
			store.Delete(iterator.Key())
			count++
		}
	}

	return count
}

// evictIfNeeded implements LRU eviction policy
func (k Keeper) evictIfNeeded(ctx sdk.Context, config CacheConfig) {
	// Count current entries
	count := k.GetCacheSize(ctx)
	if count <= config.MaxCacheSize {
		return
	}

	// Evict oldest/least used entries
	toEvict := count - config.MaxCacheSize
	k.evictLRU(ctx, toEvict)
}

// evictLRU evicts least recently used entries
func (k Keeper) evictLRU(ctx sdk.Context, count uint64) {
	if count == 0 {
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CacheKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	type cacheEntry struct {
		key      []byte
		hitCount uint64
		created  time.Time
	}

	var entries []cacheEntry
	for ; iterator.Valid(); iterator.Next() {
		var cached CachedResponse
		if err := json.Unmarshal(iterator.Value(), &cached); err != nil {
			continue
		}

		entries = append(entries, cacheEntry{
			key:      iterator.Key(),
			hitCount: cached.HitCount,
			created:  cached.CreatedAt,
		})
	}

	// Sort by hit count (ascending) and age (oldest first)
	// Simple bubble sort for small sets
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].hitCount > entries[j].hitCount ||
				(entries[i].hitCount == entries[j].hitCount && entries[i].created.After(entries[j].created)) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict first 'count' entries
	evicted := uint64(0)
	for i := 0; i < len(entries) && evicted < count; i++ {
		store.Delete(entries[i].key)
		evicted++
	}
}

// GetCacheSize returns number of cached entries
func (k Keeper) GetCacheSize(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CacheKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	count := uint64(0)
	for ; iterator.Valid(); iterator.Next() {
		count++
	}

	return count
}

// GetCacheStats returns cache statistics
func (k Keeper) GetCacheStats(ctx sdk.Context) CacheStats {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CacheKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	stats := CacheStats{}
	now := ctx.BlockTime()

	for ; iterator.Valid(); iterator.Next() {
		var cached CachedResponse
		if err := json.Unmarshal(iterator.Value(), &cached); err != nil {
			continue
		}

		stats.TotalEntries++
		stats.TotalSize += cached.Size
		stats.TotalHits += cached.HitCount

		if now.After(cached.ExpiresAt) {
			stats.ExpiredEntries++
		}
	}

	if stats.TotalEntries > 0 {
		stats.AverageHits = stats.TotalHits / stats.TotalEntries
	}

	return stats
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalEntries   uint64
	ExpiredEntries uint64
	TotalSize      uint64
	TotalHits      uint64
	AverageHits    uint64
}
