package optimization

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Memoizer provides memoization for expensive calculations
type Memoizer struct {
	cache map[string]*memoEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

type memoEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoizer creates a new memoizer with specified TTL
func NewMemoizer(ttl time.Duration) *Memoizer {
	return &Memoizer{
		cache: make(map[string]*memoEntry),
		ttl:   ttl,
	}
}

// Memoize memoizes a function call
func (m *Memoizer) Memoize(key string, fn func() (interface{}, error)) (interface{}, error) {
	m.mu.RLock()
	if entry, ok := m.cache[key]; ok {
		if time.Now().Before(entry.expiresAt) {
			m.mu.RUnlock()
			return entry.value, nil
		}
	}
	m.mu.RUnlock()

	value, err := fn()
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.cache[key] = &memoEntry{
		value:     value,
		expiresAt: time.Now().Add(m.ttl),
	}
	m.mu.Unlock()

	return value, nil
}

// MemoizeWithArgs memoizes a function with arguments
func (m *Memoizer) MemoizeWithArgs(args interface{}, fn func() (interface{}, error)) (interface{}, error) {
	key, err := m.argsToKey(args)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return m.Memoize(key, fn)
}

// Invalidate removes a cached entry
func (m *Memoizer) Invalidate(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cache, key)
}

// Clear clears all cached entries
func (m *Memoizer) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[string]*memoEntry)
}

// Cleanup removes expired entries
func (m *Memoizer) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, entry := range m.cache {
		if now.After(entry.expiresAt) {
			delete(m.cache, key)
		}
	}
}

// argsToKey converts arguments to a cache key
func (m *Memoizer) argsToKey(args interface{}) (string, error) {
	data, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// LazyValue provides lazy evaluation for expensive computations
type LazyValue struct {
	fn      func() (interface{}, error)
	value   interface{}
	err     error
	once    sync.Once
	fetched bool
	mu      sync.RWMutex
}

// NewLazyValue creates a new lazy value
func NewLazyValue(fn func() (interface{}, error)) *LazyValue {
	return &LazyValue{
		fn: fn,
	}
}

// Get retrieves the value, computing it if necessary
func (lv *LazyValue) Get() (interface{}, error) {
	lv.once.Do(func() {
		lv.value, lv.err = lv.fn()
		lv.mu.Lock()
		lv.fetched = true
		lv.mu.Unlock()
	})
	return lv.value, lv.err
}

// IsFetched returns whether the value has been fetched
func (lv *LazyValue) IsFetched() bool {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	return lv.fetched
}

// Reset resets the lazy value
func (lv *LazyValue) Reset() {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	lv.once = sync.Once{}
	lv.value = nil
	lv.err = nil
	lv.fetched = false
}

// ComputationCache provides caching with dependency tracking
type ComputationCache struct {
	cache        map[string]*cachedComputation
	dependencies map[string][]string
	mu           sync.RWMutex
}

type cachedComputation struct {
	value      interface{}
	computedAt time.Time
	deps       []string
}

// NewComputationCache creates a new computation cache
func NewComputationCache() *ComputationCache {
	return &ComputationCache{
		cache:        make(map[string]*cachedComputation),
		dependencies: make(map[string][]string),
	}
}

// Compute computes or retrieves a cached value
func (cc *ComputationCache) Compute(
	key string,
	deps []string,
	fn func() (interface{}, error),
) (interface{}, error) {
	cc.mu.RLock()
	if cached, ok := cc.cache[key]; ok {
		valid := true
		for _, dep := range cached.deps {
			if depCached, ok := cc.cache[dep]; ok {
				if depCached.computedAt.After(cached.computedAt) {
					valid = false
					break
				}
			}
		}
		if valid {
			cc.mu.RUnlock()
			return cached.value, nil
		}
	}
	cc.mu.RUnlock()

	value, err := fn()
	if err != nil {
		return nil, err
	}

	cc.mu.Lock()
	cc.cache[key] = &cachedComputation{
		value:      value,
		computedAt: time.Now(),
		deps:       deps,
	}

	for _, dep := range deps {
		cc.dependencies[dep] = append(cc.dependencies[dep], key)
	}
	cc.mu.Unlock()

	return value, nil
}

// Invalidate invalidates a key and all dependent computations
func (cc *ComputationCache) Invalidate(key string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	delete(cc.cache, key)

	if deps, ok := cc.dependencies[key]; ok {
		for _, dep := range deps {
			cc.invalidateRecursive(dep)
		}
		delete(cc.dependencies, key)
	}
}

// invalidateRecursive recursively invalidates dependent computations
func (cc *ComputationCache) invalidateRecursive(key string) {
	delete(cc.cache, key)

	if deps, ok := cc.dependencies[key]; ok {
		for _, dep := range deps {
			cc.invalidateRecursive(dep)
		}
		delete(cc.dependencies, key)
	}
}

// Clear clears all cached computations
func (cc *ComputationCache) Clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.cache = make(map[string]*cachedComputation)
	cc.dependencies = make(map[string][]string)
}

// ResultCache provides simple result caching with type safety
type ResultCache[K comparable, V any] struct {
	cache map[K]V
	mu    sync.RWMutex
}

// NewResultCache creates a new typed result cache
func NewResultCache[K comparable, V any]() *ResultCache[K, V] {
	return &ResultCache[K, V]{
		cache: make(map[K]V),
	}
}

// GetOrCompute retrieves a cached value or computes it
func (rc *ResultCache[K, V]) GetOrCompute(key K, fn func() (V, error)) (V, error) {
	rc.mu.RLock()
	if value, ok := rc.cache[key]; ok {
		rc.mu.RUnlock()
		return value, nil
	}
	rc.mu.RUnlock()

	value, err := fn()
	if err != nil {
		var zero V
		return zero, err
	}

	rc.mu.Lock()
	rc.cache[key] = value
	rc.mu.Unlock()

	return value, nil
}

// Set sets a value in the cache
func (rc *ResultCache[K, V]) Set(key K, value V) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache[key] = value
}

// Get retrieves a value from the cache
func (rc *ResultCache[K, V]) Get(key K) (V, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	value, ok := rc.cache[key]
	return value, ok
}

// Delete removes a value from the cache
func (rc *ResultCache[K, V]) Delete(key K) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	delete(rc.cache, key)
}

// Clear clears the cache
func (rc *ResultCache[K, V]) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache = make(map[K]V)
}
