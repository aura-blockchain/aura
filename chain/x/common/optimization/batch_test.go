package optimization

import (
	"context"
	"testing"
	"time"

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
	return nil
}

func (m *mockStore) ReverseIterator(start, end []byte) storetypes.Iterator {
	return nil
}

func TestParallelIterator(t *testing.T) {
	t.Skip("Requires SDK context")
}
