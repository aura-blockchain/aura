// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "economicsecurity", ModuleName)
	require.Equal(t, ModuleName, StoreKey)
	require.Equal(t, ModuleName, RouterKey)
	require.Equal(t, ModuleName, QuerierRoute)
	require.Equal(t, "mem_economicsecurity", MemStoreKey)
}

func TestKeyPrefixes_Unique(t *testing.T) {
	prefixes := [][]byte{
		ParamsKey,
		DynamicFeeConfigPrefix,
		MEVConfigPrefix,
		WhaleProtectionPrefix,
		VestingSchedulePrefix,
		VoteLockPrefix,
		PendingTreasuryTxPrefix,
		RewardDistributionPrefix,
		UserVestingIndexPrefix,
		UserVoteLockIndexPrefix,
		WhaleActivityPrefix,
		InflationAlertPrefix,
		LargeTxRecordPrefix,
		LastLargeTxTimePrefix,
		AddressHoldingPrefix,
		UserMEVBalancePrefix,
		TotalMEVPendingKey,
		TotalBurnedKey,
		PreviousInflationKey,
		CurrentHeightKey,
		CurrentTimeKey,
		InflationAlertCounterKey,
		LargeTxRecordCounterKey,
	}

	// Ensure all prefixes are unique
	seen := make(map[byte]int)
	for i, p := range prefixes {
		if prevIdx, exists := seen[p[0]]; exists {
			t.Errorf("prefix at index %d has same value as prefix at index %d (value: 0x%02X)", i, prevIdx, p[0])
		}
		seen[p[0]] = i
	}
}

func TestGetDynamicFeeConfigKey(t *testing.T) {
	id := "fee123"
	key := GetDynamicFeeConfigKey(id)

	require.NotEmpty(t, key)
	require.Equal(t, DynamicFeeConfigPrefix[0], key[0])
	require.Contains(t, string(key), id)
}

func TestGetMEVConfigKey(t *testing.T) {
	id := "mev456"
	key := GetMEVConfigKey(id)

	require.NotEmpty(t, key)
	require.Equal(t, MEVConfigPrefix[0], key[0])
	require.Contains(t, string(key), id)
}

func TestGetWhaleProtectionKey(t *testing.T) {
	address := "aura1whale123"
	key := GetWhaleProtectionKey(address)

	require.NotEmpty(t, key)
	require.Equal(t, WhaleProtectionPrefix[0], key[0])
	require.Contains(t, string(key), address)
}

func TestGetVestingScheduleKey(t *testing.T) {
	scheduleID := "schedule789"
	key := GetVestingScheduleKey(scheduleID)

	require.NotEmpty(t, key)
	require.Equal(t, VestingSchedulePrefix[0], key[0])
	require.Contains(t, string(key), scheduleID)
}

func TestGetVoteLockKey(t *testing.T) {
	lockID := "lock123"
	key := GetVoteLockKey(lockID)

	require.NotEmpty(t, key)
	require.Equal(t, VoteLockPrefix[0], key[0])
	require.Contains(t, string(key), lockID)
}

func TestGetPendingTreasuryTxKey(t *testing.T) {
	txID := "tx456"
	key := GetPendingTreasuryTxKey(txID)

	require.NotEmpty(t, key)
	require.Equal(t, PendingTreasuryTxPrefix[0], key[0])
	require.Contains(t, string(key), txID)
}

func TestGetRewardDistributionKey(t *testing.T) {
	distID := "dist789"
	key := GetRewardDistributionKey(distID)

	require.NotEmpty(t, key)
	require.Equal(t, RewardDistributionPrefix[0], key[0])
	require.Contains(t, string(key), distID)
}

func TestGetUserVestingIndexKey(t *testing.T) {
	address := "aura1user123"
	key := GetUserVestingIndexKey(address)

	require.NotEmpty(t, key)
	require.Equal(t, UserVestingIndexPrefix[0], key[0])
	require.Contains(t, string(key), address)
}

func TestGetUserVoteLockIndexKey(t *testing.T) {
	address := "aura1voter456"
	key := GetUserVoteLockIndexKey(address)

	require.NotEmpty(t, key)
	require.Equal(t, UserVoteLockIndexPrefix[0], key[0])
	require.Contains(t, string(key), address)
}

func TestGetWhaleActivityKey(t *testing.T) {
	address := "aura1whaleact"
	key := GetWhaleActivityKey(address)

	require.NotEmpty(t, key)
	require.Equal(t, WhaleActivityPrefix[0], key[0])
	require.Contains(t, string(key), address)
}

func TestGetInflationAlertKey(t *testing.T) {
	alertID := "alert123"
	key := GetInflationAlertKey(alertID)

	require.NotEmpty(t, key)
	require.Equal(t, InflationAlertPrefix[0], key[0])
	require.Contains(t, string(key), alertID)
}

func TestGetLargeTxRecordKey(t *testing.T) {
	recordID := "record456"
	key := GetLargeTxRecordKey(recordID)

	require.NotEmpty(t, key)
	require.Equal(t, LargeTxRecordPrefix[0], key[0])
	require.Contains(t, string(key), recordID)
}

func TestGetLastLargeTxTimeKey(t *testing.T) {
	address := "aura1lasttx"
	key := GetLastLargeTxTimeKey(address)

	require.NotEmpty(t, key)
	require.Equal(t, LastLargeTxTimePrefix[0], key[0])
	require.Contains(t, string(key), address)
}

func TestGetAddressHoldingKey(t *testing.T) {
	address := "aura1holding"
	key := GetAddressHoldingKey(address)

	require.NotEmpty(t, key)
	require.Equal(t, AddressHoldingPrefix[0], key[0])
	require.Contains(t, string(key), address)
}

func TestGetUserMEVBalanceKey(t *testing.T) {
	address := "aura1mevbal"
	key := GetUserMEVBalanceKey(address)

	require.NotEmpty(t, key)
	require.Equal(t, UserMEVBalancePrefix[0], key[0])
	require.Contains(t, string(key), address)
}

func TestKeyFunctions_EmptyInput(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string) []byte
	}{
		{"GetDynamicFeeConfigKey", GetDynamicFeeConfigKey},
		{"GetMEVConfigKey", GetMEVConfigKey},
		{"GetWhaleProtectionKey", GetWhaleProtectionKey},
		{"GetVestingScheduleKey", GetVestingScheduleKey},
		{"GetVoteLockKey", GetVoteLockKey},
		{"GetPendingTreasuryTxKey", GetPendingTreasuryTxKey},
		{"GetRewardDistributionKey", GetRewardDistributionKey},
		{"GetUserVestingIndexKey", GetUserVestingIndexKey},
		{"GetUserVoteLockIndexKey", GetUserVoteLockIndexKey},
		{"GetWhaleActivityKey", GetWhaleActivityKey},
		{"GetInflationAlertKey", GetInflationAlertKey},
		{"GetLargeTxRecordKey", GetLargeTxRecordKey},
		{"GetLastLargeTxTimeKey", GetLastLargeTxTimeKey},
		{"GetAddressHoldingKey", GetAddressHoldingKey},
		{"GetUserMEVBalanceKey", GetUserMEVBalanceKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.fn("")
			require.NotEmpty(t, key)
			require.Len(t, key, 1) // Should only contain the prefix
		})
	}
}

func TestKeyFunctions_UniqueKeys(t *testing.T) {
	id1 := "id1"
	id2 := "id2"

	// Same function, different IDs should produce different keys
	key1 := GetDynamicFeeConfigKey(id1)
	key2 := GetDynamicFeeConfigKey(id2)
	require.NotEqual(t, key1, key2)

	key1 = GetVestingScheduleKey(id1)
	key2 = GetVestingScheduleKey(id2)
	require.NotEqual(t, key1, key2)
}
