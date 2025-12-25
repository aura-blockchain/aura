// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package optimization

import (
	"bytes"
	"encoding/binary"
	"fmt"

	storetypes "cosmossdk.io/store/types"
)

// SecondaryIndex provides secondary indexing for efficient lookups
type SecondaryIndex struct {
	store      storetypes.KVStore
	indexName  string
	keyBuilder func(value interface{}) []byte
}

// NewSecondaryIndex creates a new secondary index
func NewSecondaryIndex(
	store storetypes.KVStore,
	indexName string,
	keyBuilder func(value interface{}) []byte,
) *SecondaryIndex {
	return &SecondaryIndex{
		store:      store,
		indexName:  indexName,
		keyBuilder: keyBuilder,
	}
}

// Add adds an entry to the secondary index
func (si *SecondaryIndex) Add(indexKey []byte, primaryKey []byte) {
	key := si.buildIndexKey(indexKey)
	si.store.Set(key, primaryKey)
}

// Get retrieves a primary key from the secondary index
func (si *SecondaryIndex) Get(indexKey []byte) []byte {
	key := si.buildIndexKey(indexKey)
	return si.store.Get(key)
}

// Delete removes an entry from the secondary index
func (si *SecondaryIndex) Delete(indexKey []byte) {
	key := si.buildIndexKey(indexKey)
	si.store.Delete(key)
}

// buildIndexKey builds the full index key with prefix
func (si *SecondaryIndex) buildIndexKey(indexKey []byte) []byte {
	prefix := []byte(fmt.Sprintf("idx:%s:", si.indexName))
	return append(prefix, indexKey...)
}

// MultiValueIndex allows multiple values for a single index key
type MultiValueIndex struct {
	store     storetypes.KVStore
	indexName string
}

// NewMultiValueIndex creates a new multi-value index
func NewMultiValueIndex(store storetypes.KVStore, indexName string) *MultiValueIndex {
	return &MultiValueIndex{
		store:     store,
		indexName: indexName,
	}
}

// Add adds a value to the index
func (mvi *MultiValueIndex) Add(indexKey []byte, value []byte) {
	key := mvi.buildKey(indexKey, value)
	mvi.store.Set(key, []byte{1})
}

// Remove removes a value from the index
func (mvi *MultiValueIndex) Remove(indexKey []byte, value []byte) {
	key := mvi.buildKey(indexKey, value)
	mvi.store.Delete(key)
}

// GetAll retrieves all values for an index key
func (mvi *MultiValueIndex) GetAll(indexKey []byte) [][]byte {
	prefix := mvi.buildPrefix(indexKey)
	iterator := storetypes.KVStorePrefixIterator(mvi.store, prefix)
	defer iterator.Close()

	var values [][]byte
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		value := key[len(prefix):]
		values = append(values, value)
	}

	return values
}

// buildKey builds the full key for a value
func (mvi *MultiValueIndex) buildKey(indexKey []byte, value []byte) []byte {
	prefix := mvi.buildPrefix(indexKey)
	return append(prefix, value...)
}

// buildPrefix builds the prefix for an index key
func (mvi *MultiValueIndex) buildPrefix(indexKey []byte) []byte {
	prefix := []byte(fmt.Sprintf("midx:%s:", mvi.indexName))
	return append(prefix, append(indexKey, []byte(":")...)...)
}

// CompositeIndex provides indexing on multiple fields
type CompositeIndex struct {
	store     storetypes.KVStore
	indexName string
	fields    []string
}

// NewCompositeIndex creates a new composite index
func NewCompositeIndex(store storetypes.KVStore, indexName string, fields []string) *CompositeIndex {
	return &CompositeIndex{
		store:     store,
		indexName: indexName,
		fields:    fields,
	}
}

// Add adds an entry to the composite index
func (ci *CompositeIndex) Add(fieldValues [][]byte, primaryKey []byte) error {
	if len(fieldValues) != len(ci.fields) {
		return fmt.Errorf("expected %d field values, got %d", len(ci.fields), len(fieldValues))
	}

	key := ci.buildKey(fieldValues)
	ci.store.Set(key, primaryKey)
	return nil
}

// Get retrieves a primary key using the composite index
func (ci *CompositeIndex) Get(fieldValues [][]byte) ([]byte, error) {
	if len(fieldValues) != len(ci.fields) {
		return nil, fmt.Errorf("expected %d field values, got %d", len(ci.fields), len(fieldValues))
	}

	key := ci.buildKey(fieldValues)
	return ci.store.Get(key), nil
}

// Delete removes an entry from the composite index
func (ci *CompositeIndex) Delete(fieldValues [][]byte) error {
	if len(fieldValues) != len(ci.fields) {
		return fmt.Errorf("expected %d field values, got %d", len(ci.fields), len(fieldValues))
	}

	key := ci.buildKey(fieldValues)
	ci.store.Delete(key)
	return nil
}

// buildKey builds the composite key
func (ci *CompositeIndex) buildKey(fieldValues [][]byte) []byte {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("cidx:%s:", ci.indexName))

	for i, value := range fieldValues {
		if i > 0 {
			buf.WriteByte(':')
		}
		buf.Write(value)
	}

	return buf.Bytes()
}

// RangeIndex provides efficient range queries
type RangeIndex struct {
	store     storetypes.KVStore
	indexName string
}

// NewRangeIndex creates a new range index
func NewRangeIndex(store storetypes.KVStore, indexName string) *RangeIndex {
	return &RangeIndex{
		store:     store,
		indexName: indexName,
	}
}

// Add adds an entry to the range index
func (ri *RangeIndex) Add(sortValue uint64, primaryKey []byte) {
	key := ri.buildKey(sortValue)
	ri.store.Set(key, primaryKey)
}

// GetRange retrieves all entries in a range
func (ri *RangeIndex) GetRange(start, end uint64) [][]byte {
	startKey := ri.buildKey(start)
	endKey := ri.buildKey(end + 1)

	iterator := ri.store.Iterator(startKey, endKey)
	defer iterator.Close()

	var results [][]byte
	for ; iterator.Valid(); iterator.Next() {
		results = append(results, iterator.Value())
	}

	return results
}

// Delete removes an entry from the range index
func (ri *RangeIndex) Delete(sortValue uint64) {
	key := ri.buildKey(sortValue)
	ri.store.Delete(key)
}

// buildKey builds the range index key
func (ri *RangeIndex) buildKey(sortValue uint64) []byte {
	prefix := []byte(fmt.Sprintf("ridx:%s:", ri.indexName))
	valueBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(valueBuf, sortValue)
	return append(prefix, valueBuf...)
}
