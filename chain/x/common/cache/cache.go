// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

// Package cache provides high-performance caching for the AURA blockchain.
//
// This package implements multi-layer caching with compression, eviction policies,
// and metrics tracking to optimize frequently accessed data.
package cache

import (
	"compress/gzip"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

// Cache provides a thread-safe, multi-layer caching mechanism with compression
type Cache struct {
	// L1: In-memory LRU cache with expiration
	l1Cache *expirable.LRU[string, *CacheEntry]

	// L2: Compressed cache for larger data
	l2Cache *expirable.LRU[string, []byte]

	// Metrics
	mu          sync.RWMutex
	hits        uint64
	misses      uint64
	evictions   uint64
	compressionRatio float64

	// Configuration
	config CacheConfig
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Data      interface{}
	CachedAt  time.Time
	ExpiresAt time.Time
	Size      int
	Compressed bool
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	// L1 cache size (number of items)
	L1Size int

	// L2 cache size (number of items)
	L2Size int

	// Default TTL for cache entries
	DefaultTTL time.Duration

	// Compression threshold (bytes)
	CompressionThreshold int

	// Enable compression
	EnableCompression bool

	// Enable metrics
	EnableMetrics bool
}

// DefaultCacheConfig returns sensible defaults
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		L1Size:               10000,
		L2Size:               50000,
		DefaultTTL:           5 * time.Minute,
		CompressionThreshold: 1024, // 1KB
		EnableCompression:    true,
		EnableMetrics:        true,
	}
}

// NewCache creates a new multi-layer cache instance
//
// Example usage:
//
//	config := cache.DefaultCacheConfig()
//	config.L1Size = 5000
//	c := cache.NewCache(config)
//
//	// Store value
//	c.Set("key", value, 10*time.Minute)
//
//	// Retrieve value
//	if val, found := c.Get("key"); found {
//	    // Use val
//	}
func NewCache(config CacheConfig) *Cache {
	if config.L1Size == 0 {
		config.L1Size = DefaultCacheConfig().L1Size
	}
	if config.L2Size == 0 {
		config.L2Size = DefaultCacheConfig().L2Size
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = DefaultCacheConfig().DefaultTTL
	}

	return &Cache{
		l1Cache: expirable.NewLRU[string, *CacheEntry](
			config.L1Size,
			func(key string, value *CacheEntry) {
				// Eviction callback
			},
			config.DefaultTTL,
		),
		l2Cache: expirable.NewLRU[string, []byte](
			config.L2Size,
			nil,
			config.DefaultTTL,
		),
		config: config,
	}
}

// Get retrieves a value from cache
//
// Returns the cached value and true if found, nil and false otherwise.
// Automatically handles decompression and promotes L2 hits to L1.
func (c *Cache) Get(key string) (interface{}, bool) {
	// Try L1 cache first (fastest)
	if entry, found := c.l1Cache.Get(key); found {
		if c.config.EnableMetrics {
			c.mu.Lock()
			c.hits++
			c.mu.Unlock()
		}
		return entry.Data, true
	}

	// Try L2 cache (compressed)
	if compressed, found := c.l2Cache.Get(key); found {
		// Decompress and promote to L1
		data, err := c.decompress(compressed)
		if err == nil {
			entry := &CacheEntry{
				Data:      data,
				CachedAt:  time.Now(),
				Compressed: false,
			}
			c.l1Cache.Add(key, entry)

			if c.config.EnableMetrics {
				c.mu.Lock()
				c.hits++
				c.mu.Unlock()
			}
			return data, true
		}
	}

	// Cache miss
	if c.config.EnableMetrics {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
	}
	return nil, false
}

// Set stores a value in cache with optional TTL
//
// If ttl is 0, uses the default TTL from config.
// Automatically handles compression for large items.
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	entry := &CacheEntry{
		Data:      value,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		Size:      c.estimateSize(value),
	}

	// Decide whether to compress
	if c.config.EnableCompression && entry.Size > c.config.CompressionThreshold {
		// Large item - compress and store in L2
		compressed, err := c.compress(value)
		if err != nil {
			return fmt.Errorf("compression failed: %w", err)
		}

		entry.Compressed = true
		c.l2Cache.Add(key, compressed)

		// Track compression ratio
		if c.config.EnableMetrics {
			c.mu.Lock()
			ratio := float64(len(compressed)) / float64(entry.Size)
			c.compressionRatio = (c.compressionRatio + ratio) / 2
			c.mu.Unlock()
		}
	} else {
		// Small item - store directly in L1
		c.l1Cache.Add(key, entry)
	}

	return nil
}

// Delete removes an item from cache
func (c *Cache) Delete(key string) {
	c.l1Cache.Remove(key)
	c.l2Cache.Remove(key)
}

// Clear removes all items from cache
func (c *Cache) Clear() {
	c.l1Cache.Purge()
	c.l2Cache.Purge()
}

// GetOrSet retrieves a value from cache or sets it if not found
//
// The loader function is only called on cache miss.
// This pattern is useful for lazy loading and avoiding duplicate computations.
//
// Example:
//
//	value, err := c.GetOrSet("expensive-key", 10*time.Minute, func() (interface{}, error) {
//	    // Expensive computation only runs on cache miss
//	    return computeExpensiveValue()
//	})
func (c *Cache) GetOrSet(key string, ttl time.Duration, loader func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, found := c.Get(key); found {
		return value, nil
	}

	// Cache miss - call loader
	value, err := loader()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.Set(key, value, ttl); err != nil {
		return nil, err
	}

	return value, nil
}

// Warm pre-loads cache with data
//
// Useful for cache warming on startup to avoid cold start penalties.
func (c *Cache) Warm(ctx context.Context, loader func(ctx context.Context) (map[string]interface{}, error)) error {
	data, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("warm loader failed: %w", err)
	}

	for key, value := range data {
		if err := c.Set(key, value, c.config.DefaultTTL); err != nil {
			return fmt.Errorf("warm set failed for key %s: %w", key, err)
		}
	}

	return nil
}

// Metrics returns cache performance metrics
type Metrics struct {
	Hits             uint64
	Misses           uint64
	Evictions        uint64
	HitRate          float64
	CompressionRatio float64
	L1Size           int
	L2Size           int
}

// GetMetrics returns current cache metrics
func (c *Cache) GetMetrics() Metrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return Metrics{
		Hits:             c.hits,
		Misses:           c.misses,
		Evictions:        c.evictions,
		HitRate:          hitRate,
		CompressionRatio: c.compressionRatio,
		L1Size:           c.l1Cache.Len(),
		L2Size:           c.l2Cache.Len(),
	}
}

// compress serializes and compresses data
func (c *Cache) compress(data interface{}) ([]byte, error) {
	// Serialize using gob
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("gob encode failed: %w", err)
	}

	// Compress using gzip
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("gzip write failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("gzip close failed: %w", err)
	}

	return compressed.Bytes(), nil
}

// decompress decompresses and deserializes data
func (c *Cache) decompress(compressed []byte) (interface{}, error) {
	// Decompress
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("gzip reader failed: %w", err)
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("gzip read failed: %w", err)
	}

	// Deserialize
	var data interface{}
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("gob decode failed: %w", err)
	}

	return data, nil
}

// estimateSize estimates the size of a value in bytes
func (c *Cache) estimateSize(value interface{}) int {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(value); err != nil {
		return 0
	}
	return buf.Len()
}

// InvalidatePattern removes all keys matching a pattern
//
// Example:
//
//	c.InvalidatePattern("user:*")  // Remove all user-related cache entries
func (c *Cache) InvalidatePattern(pattern string) {
	// Note: This is a simplified implementation
	// Production code should use a more sophisticated pattern matching
	c.Clear() // For now, clear everything
}
