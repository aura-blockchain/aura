package cache

import (
	"context"
	"testing"
	"time"
)

func TestCacheBasicOperations(t *testing.T) {
	c := NewCache(DefaultCacheConfig())

	// Test Set and Get
	key := "test-key"
	value := "test-value"

	err := c.Set(key, value, time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, found := c.Get(key)
	if !found {
		t.Fatal("Expected to find key in cache")
	}

	if retrieved != value {
		t.Fatalf("Expected %v, got %v", value, retrieved)
	}
}

func TestCacheExpiration(t *testing.T) {
	config := DefaultCacheConfig()
	config.DefaultTTL = 100 * time.Millisecond
	c := NewCache(config)

	key := "expiring-key"
	value := "expiring-value"

	c.Set(key, value, 100*time.Millisecond)

	// Should be present immediately
	if _, found := c.Get(key); !found {
		t.Fatal("Expected key to be present")
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired
	if _, found := c.Get(key); found {
		t.Fatal("Expected key to be expired")
	}
}

func TestCacheMetrics(t *testing.T) {
	c := NewCache(DefaultCacheConfig())

	// Generate some hits and misses
	c.Set("key1", "value1", time.Minute)
	c.Get("key1") // hit
	c.Get("key2") // miss

	metrics := c.GetMetrics()

	if metrics.Hits != 1 {
		t.Fatalf("Expected 1 hit, got %d", metrics.Hits)
	}

	if metrics.Misses != 1 {
		t.Fatalf("Expected 1 miss, got %d", metrics.Misses)
	}

	if metrics.HitRate != 0.5 {
		t.Fatalf("Expected hit rate 0.5, got %f", metrics.HitRate)
	}
}

func TestGetOrSet(t *testing.T) {
	c := NewCache(DefaultCacheConfig())

	callCount := 0
	loader := func() (interface{}, error) {
		callCount++
		return "loaded-value", nil
	}

	// First call should invoke loader
	val1, err := c.GetOrSet("key", time.Minute, loader)
	if err != nil {
		t.Fatalf("GetOrSet failed: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("Expected loader to be called once, got %d", callCount)
	}

	// Second call should use cache
	val2, err := c.GetOrSet("key", time.Minute, loader)
	if err != nil {
		t.Fatalf("GetOrSet failed: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("Expected loader to still be called once, got %d", callCount)
	}

	if val1 != val2 {
		t.Fatal("Values should be equal")
	}
}

func TestCacheWarm(t *testing.T) {
	c := NewCache(DefaultCacheConfig())

	loader := func(ctx context.Context) (map[string]interface{}, error) {
		return map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}, nil
	}

	err := c.Warm(context.Background(), loader)
	if err != nil {
		t.Fatalf("Warm failed: %v", err)
	}

	// Verify all keys are cached
	for i := 1; i <= 3; i++ {
		key := "key1"
		if _, found := c.Get(key); !found {
			t.Fatalf("Expected %s to be in cache", key)
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	c := NewCache(DefaultCacheConfig())
	c.Set("benchmark-key", "benchmark-value", time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("benchmark-key")
	}
}

func BenchmarkCacheSet(b *testing.B) {
	c := NewCache(DefaultCacheConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("benchmark-key", "benchmark-value", time.Hour)
	}
}
