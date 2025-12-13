package optimization

import (
	"context"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/common/cache"
)

// OptimizedKeeper provides optimized keeper operations
type OptimizedKeeper struct {
	storeKey    storetypes.StoreKey
	cache       *cache.Cache
	memoizer    *Memoizer
	batchSize   int
	enableCache bool
}

// NewOptimizedKeeper creates a new optimized keeper
func NewOptimizedKeeper(
	storeKey storetypes.StoreKey,
	enableCache bool,
	cacheConfig cache.CacheConfig,
) *OptimizedKeeper {
	var c *cache.Cache
	if enableCache {
		c = cache.NewCache(cacheConfig)
	}

	return &OptimizedKeeper{
		storeKey:    storeKey,
		cache:       c,
		memoizer:    NewMemoizer(5 * time.Minute),
		batchSize:   100,
		enableCache: enableCache,
	}
}

// GetWithCache retrieves a value with caching
func (ok *OptimizedKeeper) GetWithCache(
	ctx context.Context,
	key string,
	loader func() (interface{}, error),
) (interface{}, error) {
	if !ok.enableCache {
		return loader()
	}

	return ok.cache.GetOrSet(key, 5*time.Minute, loader)
}

// MemoizedCompute performs a memoized computation
func (ok *OptimizedKeeper) MemoizedCompute(
	key string,
	fn func() (interface{}, error),
) (interface{}, error) {
	return ok.memoizer.Memoize(key, fn)
}

// BatchWrite performs batch writes
func (ok *OptimizedKeeper) BatchWrite(
	ctx context.Context,
	operations []BatchOperation,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(ok.storeKey)

	batch := NewBatchExecutor(store)
	for _, op := range operations {
		batch.Add(op.Key, op.Value, op.Op)

		if batch.Size() >= ok.batchSize {
			if err := batch.Execute(); err != nil {
				return err
			}
		}
	}

	return batch.Execute()
}

// InvalidateCache invalidates cache entries
func (ok *OptimizedKeeper) InvalidateCache(pattern string) {
	if ok.enableCache {
		ok.cache.InvalidatePattern(pattern)
	}
}

// QueryOptimizer provides query optimization utilities
type QueryOptimizer struct {
	enablePagination bool
	defaultPageSize  uint64
	maxPageSize      uint64
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{
		enablePagination: true,
		defaultPageSize:  100,
		maxPageSize:      1000,
	}
}

// OptimizePagination optimizes pagination parameters
func (qo *QueryOptimizer) OptimizePagination(limit, offset uint64) (uint64, uint64) {
	if limit == 0 {
		limit = qo.defaultPageSize
	}
	if limit > qo.maxPageSize {
		limit = qo.maxPageSize
	}
	return limit, offset
}

// ShouldCache determines if a query result should be cached
func (qo *QueryOptimizer) ShouldCache(resultSize int) bool {
	return resultSize < 1000
}

// ComputeIndexKey computes an optimized index key
func (qo *QueryOptimizer) ComputeIndexKey(fields ...string) string {
	var key string
	for i, field := range fields {
		if i > 0 {
			key += ":"
		}
		key += field
	}
	return key
}

// PerformanceMonitor tracks performance metrics
type PerformanceMonitor struct {
	metrics map[string]*OperationMetrics
}

// OperationMetrics tracks metrics for an operation
type OperationMetrics struct {
	Count        uint64
	TotalTime    time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	AverageTime  time.Duration
	LastExecuted time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics: make(map[string]*OperationMetrics),
	}
}

// Track tracks an operation
func (pm *PerformanceMonitor) Track(operation string, duration time.Duration) {
	metrics, ok := pm.metrics[operation]
	if !ok {
		metrics = &OperationMetrics{
			MinTime: duration,
			MaxTime: duration,
		}
		pm.metrics[operation] = metrics
	}

	metrics.Count++
	metrics.TotalTime += duration
	metrics.AverageTime = metrics.TotalTime / time.Duration(metrics.Count)
	metrics.LastExecuted = time.Now()

	if duration < metrics.MinTime {
		metrics.MinTime = duration
	}
	if duration > metrics.MaxTime {
		metrics.MaxTime = duration
	}
}

// GetMetrics retrieves metrics for an operation
func (pm *PerformanceMonitor) GetMetrics(operation string) (*OperationMetrics, bool) {
	metrics, ok := pm.metrics[operation]
	return metrics, ok
}

// ResetMetrics resets all metrics
func (pm *PerformanceMonitor) ResetMetrics() {
	pm.metrics = make(map[string]*OperationMetrics)
}

// TimedOperation executes an operation and tracks its performance
func (pm *PerformanceMonitor) TimedOperation(
	operation string,
	fn func() error,
) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	pm.Track(operation, duration)

	return err
}

// OptimizationConfig provides configuration for optimizations
type OptimizationConfig struct {
	EnableCache       bool
	EnableBatching    bool
	EnableMemoization bool
	EnableIndexing    bool
	CacheConfig       cache.CacheConfig
	BatchSize         int
}

// DefaultOptimizationConfig returns default optimization configuration
func DefaultOptimizationConfig() OptimizationConfig {
	return OptimizationConfig{
		EnableCache:       true,
		EnableBatching:    true,
		EnableMemoization: true,
		EnableIndexing:    true,
		CacheConfig:       cache.DefaultCacheConfig(),
		BatchSize:         100,
	}
}

// OptimizedStorage provides optimized storage operations
type OptimizedStorage struct {
	store       storetypes.KVStore
	cache       *cache.Cache
	indexes     map[string]*SecondaryIndex
	enableCache bool
}

// NewOptimizedStorage creates a new optimized storage
func NewOptimizedStorage(
	store storetypes.KVStore,
	enableCache bool,
	cacheConfig cache.CacheConfig,
) *OptimizedStorage {
	var c *cache.Cache
	if enableCache {
		c = cache.NewCache(cacheConfig)
	}

	return &OptimizedStorage{
		store:       store,
		cache:       c,
		indexes:     make(map[string]*SecondaryIndex),
		enableCache: enableCache,
	}
}

// Get retrieves a value with caching
func (os *OptimizedStorage) Get(key []byte) []byte {
	if !os.enableCache {
		return os.store.Get(key)
	}

	cacheKey := string(key)
	if value, found := os.cache.Get(cacheKey); found {
		if bz, ok := value.([]byte); ok {
			return bz
		}
	}

	value := os.store.Get(key)
	if value != nil {
		if err := os.cache.Set(cacheKey, value, 5*time.Minute); err != nil {
			fmt.Printf("cache set failed: %v\n", err)
		}
	}

	return value
}

// Set sets a value and updates cache
func (os *OptimizedStorage) Set(key []byte, value []byte) {
	os.store.Set(key, value)

	if os.enableCache {
		if err := os.cache.Set(string(key), value, 5*time.Minute); err != nil {
			fmt.Printf("cache set failed: %v\n", err)
		}
	}
}

// Delete deletes a value and invalidates cache
func (os *OptimizedStorage) Delete(key []byte) {
	os.store.Delete(key)

	if os.enableCache {
		os.cache.Delete(string(key))
	}
}

// AddIndex adds a secondary index
func (os *OptimizedStorage) AddIndex(name string, keyBuilder func(value interface{}) []byte) {
	os.indexes[name] = NewSecondaryIndex(os.store, name, keyBuilder)
}

// GetByIndex retrieves a value using a secondary index
func (os *OptimizedStorage) GetByIndex(indexName string, indexKey []byte) ([]byte, error) {
	index, ok := os.indexes[indexName]
	if !ok {
		return nil, fmt.Errorf("index %s not found", indexName)
	}

	primaryKey := index.Get(indexKey)
	if primaryKey == nil {
		return nil, fmt.Errorf("index key not found")
	}

	return os.Get(primaryKey), nil
}

// MetricsCollector collects optimization metrics
type MetricsCollector struct {
	cacheHits     uint64
	cacheMisses   uint64
	batchWrites   uint64
	queryTime     time.Duration
	queryCount    uint64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// RecordCacheHit records a cache hit
func (mc *MetricsCollector) RecordCacheHit() {
	mc.cacheHits++
}

// RecordCacheMiss records a cache miss
func (mc *MetricsCollector) RecordCacheMiss() {
	mc.cacheMisses++
}

// RecordBatchWrite records a batch write
func (mc *MetricsCollector) RecordBatchWrite(count int) {
	mc.batchWrites += uint64(count)
}

// RecordQuery records a query
func (mc *MetricsCollector) RecordQuery(duration time.Duration) {
	mc.queryTime += duration
	mc.queryCount++
}

// GetCacheHitRate returns the cache hit rate
func (mc *MetricsCollector) GetCacheHitRate() float64 {
	total := mc.cacheHits + mc.cacheMisses
	if total == 0 {
		return 0
	}
	return float64(mc.cacheHits) / float64(total)
}

// GetAverageQueryTime returns the average query time
func (mc *MetricsCollector) GetAverageQueryTime() time.Duration {
	if mc.queryCount == 0 {
		return 0
	}
	return mc.queryTime / time.Duration(mc.queryCount)
}
