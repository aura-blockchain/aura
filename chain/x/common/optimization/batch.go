// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package optimization

import (
	"context"
	"fmt"
	"sync"

	storetypes "cosmossdk.io/store/types"
)

// BatchOperation represents a single operation in a batch
type BatchOperation struct {
	Key   []byte
	Value []byte
	Op    OpType
}

// OpType represents the type of operation
type OpType int

const (
	OpSet OpType = iota
	OpDelete
)

// BatchExecutor handles batch operations for improved performance
type BatchExecutor struct {
	store storetypes.KVStore
	ops   []BatchOperation
	mu    sync.Mutex
}

// NewBatchExecutor creates a new batch executor
func NewBatchExecutor(store storetypes.KVStore) *BatchExecutor {
	return &BatchExecutor{
		store: store,
		ops:   make([]BatchOperation, 0, 100),
	}
}

// Add adds an operation to the batch
func (b *BatchExecutor) Add(key []byte, value []byte, op OpType) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ops = append(b.ops, BatchOperation{
		Key:   key,
		Value: value,
		Op:    op,
	})
}

// Set adds a set operation to the batch
func (b *BatchExecutor) Set(key []byte, value []byte) {
	b.Add(key, value, OpSet)
}

// Delete adds a delete operation to the batch
func (b *BatchExecutor) Delete(key []byte) {
	b.Add(key, nil, OpDelete)
}

// Execute executes all batched operations
func (b *BatchExecutor) Execute() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, op := range b.ops {
		switch op.Op {
		case OpSet:
			b.store.Set(op.Key, op.Value)
		case OpDelete:
			b.store.Delete(op.Key)
		default:
			return fmt.Errorf("unknown operation type: %d", op.Op)
		}
	}

	b.ops = b.ops[:0]
	return nil
}

// Size returns the number of pending operations
func (b *BatchExecutor) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.ops)
}

// Clear clears all pending operations
func (b *BatchExecutor) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ops = b.ops[:0]
}

// ParallelIterator provides parallel iteration over store keys
type ParallelIterator struct {
	store      storetypes.KVStore
	workerPool int
}

// NewParallelIterator creates a new parallel iterator
func NewParallelIterator(store storetypes.KVStore, workers int) *ParallelIterator {
	if workers <= 0 {
		workers = 4
	}
	return &ParallelIterator{
		store:      store,
		workerPool: workers,
	}
}

// IterateRange iterates over a range of keys in parallel
func (p *ParallelIterator) IterateRange(
	ctx context.Context,
	start, end []byte,
	processor func(key, value []byte) error,
) error {
	iterator := p.store.Iterator(start, end)
	defer iterator.Close()

	type item struct {
		key   []byte
		value []byte
	}

	items := make(chan item, 100)
	errors := make(chan error, 1)
	var wg sync.WaitGroup

	for i := 0; i < p.workerPool; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for it := range items {
				if err := processor(it.key, it.value); err != nil {
					select {
					case errors <- err:
					default:
					}
					return
				}
			}
		}()
	}

	go func() {
		defer close(items)
		for ; iterator.Valid(); iterator.Next() {
			select {
			case <-ctx.Done():
				return
			case items <- item{key: iterator.Key(), value: iterator.Value()}:
			}
		}
	}()

	wg.Wait()
	close(errors)

	if err := <-errors; err != nil {
		return err
	}

	return ctx.Err()
}

// PrefetchCache provides prefetching for predictable access patterns
type PrefetchCache struct {
	store      storetypes.KVStore
	cache      map[string][]byte
	prefetched map[string]bool
	mu         sync.RWMutex
}

// NewPrefetchCache creates a new prefetch cache
func NewPrefetchCache(store storetypes.KVStore) *PrefetchCache {
	return &PrefetchCache{
		store:      store,
		cache:      make(map[string][]byte),
		prefetched: make(map[string]bool),
	}
}

// Prefetch prefetches a set of keys
func (p *PrefetchCache) Prefetch(keys [][]byte) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, key := range keys {
		if p.prefetched[string(key)] {
			continue
		}
		value := p.store.Get(key)
		if value != nil {
			p.cache[string(key)] = value
		}
		p.prefetched[string(key)] = true
	}
}

// Get retrieves a value, using cache if available
func (p *PrefetchCache) Get(key []byte) []byte {
	p.mu.RLock()
	if value, ok := p.cache[string(key)]; ok {
		p.mu.RUnlock()
		return value
	}
	p.mu.RUnlock()

	return p.store.Get(key)
}

// Clear clears the cache
func (p *PrefetchCache) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache = make(map[string][]byte)
	p.prefetched = make(map[string]bool)
}
