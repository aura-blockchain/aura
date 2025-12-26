// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyConstants(t *testing.T) {
	require.Equal(t, "dataregistry", ModuleName)
	require.Equal(t, ModuleName, StoreKey)
	require.Equal(t, ModuleName, RouterKey)
	require.Equal(t, "mem_dataregistry", MemStoreKey)
}

func TestKeyPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		prefix   []byte
		expected byte
	}{
		{"DataItemKeyPrefix", DataItemKeyPrefix, 0x01},
		{"UserDataIndexKeyPrefix", UserDataIndexKeyPrefix, 0x02},
		{"DataItemCounterKey", DataItemCounterKey, 0x03},
		{"TypeIndexKeyPrefix", TypeIndexKeyPrefix, 0x04},
		{"GeoIndexKeyPrefix", GeoIndexKeyPrefix, 0x05},
		{"VerificationIndexKeyPrefix", VerificationIndexKeyPrefix, 0x06},
		{"ParamsKey", ParamsKey, 0x09},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Len(t, tt.prefix, 1)
			require.Equal(t, tt.expected, tt.prefix[0])
		})
	}
}

func TestDataItemKey(t *testing.T) {
	tests := []struct {
		name     string
		dataID   string
		wantLen  int
	}{
		{"normal ID", "item-123", 9},
		{"empty ID", "", 1},
		{"long ID", "very-long-data-item-identifier", 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := DataItemKey(tt.dataID)
			require.Len(t, key, tt.wantLen)
			require.Equal(t, DataItemKeyPrefix[0], key[0])
		})
	}
}

func TestUserDataIndexKey(t *testing.T) {
	key := UserDataIndexKey("aura1user123")
	require.Equal(t, UserDataIndexKeyPrefix[0], key[0])
	require.Contains(t, string(key), "aura1user123")
}

func TestTypeIndexKey(t *testing.T) {
	key := TypeIndexKey("PHOTO")
	require.Equal(t, TypeIndexKeyPrefix[0], key[0])
	require.Contains(t, string(key), "PHOTO")
}

func TestGeoIndexKey(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"positive coords", 40.7128, -74.0060},
		{"zero coords", 0.0, 0.0},
		{"negative coords", -33.8688, 151.2093},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GeoIndexKey(tt.lat, tt.lon)
			require.Equal(t, GeoIndexKeyPrefix[0], key[0])
			require.Greater(t, len(key), 1)
		})
	}
}

func TestVerificationIndexKey(t *testing.T) {
	key := VerificationIndexKey("item-1", "verifier-1")
	require.Equal(t, VerificationIndexKeyPrefix[0], key[0])
	require.Contains(t, string(key), "item-1")
	require.Contains(t, string(key), "verifier-1")
}

func TestPrefixEndBytes(t *testing.T) {
	tests := []struct {
		name   string
		prefix []byte
		want   []byte
	}{
		{"single byte", []byte{0x01}, []byte{0x02}},
		{"multi byte", []byte{0x01, 0x02}, []byte{0x01, 0x03}},
		{"overflow last", []byte{0x01, 0xFF}, []byte{0x02, 0x00}},
		{"all FF", []byte{0xFF, 0xFF}, nil},
		{"empty", []byte{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrefixEndBytes(tt.prefix)
			require.Equal(t, tt.want, got)
		})
	}
}
