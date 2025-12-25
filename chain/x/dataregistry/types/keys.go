// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// ModuleName defines the module name
	ModuleName = "dataregistry"

	// StoreKey is the store key string for dataregistry
	StoreKey = ModuleName

	// RouterKey is the message route for dataregistry
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName
)

// State store key prefixes
var (
	// DataItemKeyPrefix is the prefix for data items
	DataItemKeyPrefix = []byte{0x01}

	// UserDataIndexKeyPrefix is the prefix for user data index
	UserDataIndexKeyPrefix = []byte{0x02}

	// DataItemCounterKey is the key for data item ID counter
	DataItemCounterKey = []byte{0x03}

	// TypeIndexKeyPrefix is the prefix for type-based index
	TypeIndexKeyPrefix = []byte{0x04}

	// GeoIndexKeyPrefix is the prefix for geo-location index
	GeoIndexKeyPrefix = []byte{0x05}

	// VerificationIndexKeyPrefix is the prefix for verification index
	VerificationIndexKeyPrefix = []byte{0x06}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x09}
)

// DataItemKey returns the key for a data item
func DataItemKey(dataID string) []byte {
	return append(DataItemKeyPrefix, []byte(dataID)...)
}

// UserDataIndexKey returns the key for user data index
func UserDataIndexKey(address string) []byte {
	return append(UserDataIndexKeyPrefix, []byte(address)...)
}

// TypeIndexKey returns the key for type-based index
func TypeIndexKey(dataType string) []byte {
	return append(TypeIndexKeyPrefix, []byte(dataType)...)
}

// GeoIndexKey returns the key for geo-location index
func GeoIndexKey(lat, lon float64) []byte {
	// Simple binning for geo-location (in production, use geohashing)
	latBin := int64(lat * 100)
	lonBin := int64(lon * 100)
	key := append(GeoIndexKeyPrefix, []byte{byte(latBin >> 8), byte(latBin)}...)
	return append(key, []byte{byte(lonBin >> 8), byte(lonBin)}...)
}

// VerificationIndexKey returns the key for verification index
func VerificationIndexKey(dataID string, verifier string) []byte {
	key := append(VerificationIndexKeyPrefix, []byte(dataID)...)
	return append(key, []byte(verifier)...)
}

// PrefixEndBytes returns the end bytes for a prefix, used for iteration
func PrefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end
		}
	}

	return nil
}
