package optimization

import (
	"context"
	"fmt"
	"io"
	"sort"
	"sync"
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutor(t *testing.T) {
	store := &mockStore{data: make(map[string][]byte)}
	batch := NewBatchExecutor(store)

	batch.Set([]byte("key1"), []byte("value1"))
	batch.Set([]byte("key2"), []byte("value2"))
	batch.Delete([]byte("key3"))

	require.Equal(t, 3, batch.Size())

	err := batch.Execute()
	require.NoError(t, err)

	require.Equal(t, 0, batch.Size())
	require.Equal(t, []byte("value1"), store.Get([]byte("key1")))
	require.Equal(t, []byte("value2"), store.Get([]byte("key2")))
}

func TestBatchExecutorClear(t *testing.T) {
	store := &mockStore{data: make(map[string][]byte)}
	batch := NewBatchExecutor(store)

	batch.Set([]byte("key1"), []byte("value1"))
	batch.Set([]byte("key2"), []byte("value2"))

	require.Equal(t, 2, batch.Size())

	batch.Clear()
	require.Equal(t, 0, batch.Size())
}

func TestPrefetchCache(t *testing.T) {
	store := &mockStore{data: make(map[string][]byte)}
	store.Set([]byte("key1"), []byte("value1"))
	store.Set([]byte("key2"), []byte("value2"))

	cache := NewPrefetchCache(store)

	keys := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
	}
	cache.Prefetch(keys)

	value := cache.Get([]byte("key1"))
	require.Equal(t, []byte("value1"), value)

	value = cache.Get([]byte("key2"))
	require.Equal(t, []byte("value2"), value)
}

func TestPrefetchCacheClear(t *testing.T) {
	store := &mockStore{data: make(map[string][]byte)}
	store.Set([]byte("key1"), []byte("value1"))

	cache := NewPrefetchCache(store)
	cache.Prefetch([][]byte{[]byte("key1")})

	value := cache.Get([]byte("key1"))
	require.Equal(t, []byte("value1"), value)

	cache.Clear()

	store.Set([]byte("key1"), []byte("value2"))
	value = cache.Get([]byte("key1"))
	require.Equal(t, []byte("value2"), value)
}

type mockStore struct {
	data map[string][]byte
}

func (m *mockStore) Get(key []byte) []byte {
	return m.data[string(key)]
}

func (m *mockStore) Set(key []byte, value []byte) {
	m.data[string(key)] = value
}

func (m *mockStore) Delete(key []byte) {
	delete(m.data, string(key))
}

func (m *mockStore) Has(key []byte) bool {
	_, ok := m.data[string(key)]
	return ok
}

func (m *mockStore) Iterator(start, end []byte) storetypes.Iterator {
	// Collect keys in range
	var keys []string
	var values [][]byte

	// Get all keys and sort them
	allKeys := make([]string, 0, len(m.data))
	for k := range m.data {
		allKeys = append(allKeys, k)
	}
	sort.Strings(allKeys)

	// Filter by range
	for _, k := range allKeys {
		// Check if key is in range [start, end)
		if start != nil && k < string(start) {
			continue
		}
		if end != nil && k >= string(end) {
			continue
		}
		keys = append(keys, k)
		values = append(values, m.data[k])
	}

	return &mockIterator{
		keys:    keys,
		values:  values,
		current: 0,
	}
}

func (m *mockStore) ReverseIterator(start, end []byte) storetypes.Iterator {
	return nil
}

func (m *mockStore) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeDB
}

func (m *mockStore) CacheWrap() storetypes.CacheWrap {
	// Return a new mockStore with copied data for caching
	newData := make(map[string][]byte)
	for k, v := range m.data {
		newData[k] = v
	}
	return &mockStore{data: newData}
}

func (m *mockStore) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// For testing, ignore tracing and just return CacheWrap
	return m.CacheWrap()
}

func (m *mockStore) Write() {
	// No-op for mock implementation
}

func TestParallelIterator(t *testing.T) {
	store := &mockStore{data: make(map[string][]byte)}

	// Populate store with test data
	for i := 0; i < 100; i++ {
		key := []byte(fmt.Sprintf("key%03d", i))
		value := []byte(fmt.Sprintf("value%03d", i))
		store.Set(key, value)
	}

	// Test with default worker count
	parallel := NewParallelIterator(store, 0)
	require.NotNil(t, parallel)

	// Track processed keys
	processed := make(map[string]string)
	var mu sync.Mutex

	// Process all keys
	ctx := context.Background()
	err := parallel.IterateRange(ctx, []byte("key000"), []byte("key100"), func(key, value []byte) error {
		mu.Lock()
		processed[string(key)] = string(value)
		mu.Unlock()
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 100, len(processed), "all keys should be processed")

	// Verify all keys were processed correctly
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i)
		expectedValue := fmt.Sprintf("value%03d", i)
		actualValue, exists := processed[key]
		require.True(t, exists, "key %s should exist", key)
		require.Equal(t, expectedValue, actualValue, "value for key %s should match", key)
	}

	// Test with custom worker count
	parallel2 := NewParallelIterator(store, 8)
	processed2 := make(map[string]string)
	var mu2 sync.Mutex

	err = parallel2.IterateRange(ctx, []byte("key000"), []byte("key050"), func(key, value []byte) error {
		mu2.Lock()
		processed2[string(key)] = string(value)
		mu2.Unlock()
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 50, len(processed2), "should process first 50 keys")

	// Test error handling
	parallel3 := NewParallelIterator(store, 2)
	processed3 := 0
	var mu3 sync.Mutex

	err = parallel3.IterateRange(ctx, []byte("key000"), []byte("key100"), func(key, value []byte) error {
		mu3.Lock()
		processed3++
		mu3.Unlock()
		// Simulate error on key010
		if string(key) == "key010" {
			return fmt.Errorf("simulated error")
		}
		return nil
	})

	require.Error(t, err, "should propagate error from processor")
	require.Contains(t, err.Error(), "simulated error")
}

// mockIterator implements storetypes.Iterator for testing
type mockIterator struct {
	keys    []string
	values  [][]byte
	current int
}

func (m *mockIterator) Domain() (start, end []byte) {
	return nil, nil
}

func (m *mockIterator) Valid() bool {
	return m.current < len(m.keys)
}

func (m *mockIterator) Next() {
	m.current++
}

func (m *mockIterator) Key() []byte {
	if m.current < len(m.keys) {
		return []byte(m.keys[m.current])
	}
	return nil
}

func (m *mockIterator) Value() []byte {
	if m.current < len(m.values) {
		return m.values[m.current]
	}
	return nil
}

func (m *mockIterator) Close() error {
	return nil
}

func (m *mockIterator) Error() error {
	return nil
}
