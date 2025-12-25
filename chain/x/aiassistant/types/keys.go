// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "aiassistant"

	// StoreKey is the KVStore name
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_aiassistant"

	// DefaultStakeDenom is the default staking denomination
	DefaultStakeDenom = "uaura"
)

var (
	AssistantKeyPrefix                = []byte{0x01}
	LocaleKeyPrefix                   = []byte{0x02}
	ParamsKey                         = []byte{0x10}
	QueryUsageKeyPrefix               = []byte{0x03}
	QuotaKeyPrefix                    = []byte{0x04}
	ModelKeyPrefix                    = []byte{0x05}
	ModelVersionKeyPrefix             = []byte{0x06}
	CacheKeyPrefix                    = []byte{0x07}
	AuditLogKeyPrefix                 = []byte{0x08}
	AuditLogCounterKey                = []byte{0x09}
	AnalyticsKeyPrefix                = []byte{0x0A}
	AnalyticsSnapshotKeyPrefixBase    = []byte{0x0B}
)

func AssistantKey(address string) []byte {
	return append(AssistantKeyPrefix, []byte(address)...)
}

func LocaleAssistantKey(locale, assistant string) []byte {
	key := append(LocaleKeyPrefix, []byte(locale)...)
	key = append(key, byte(0x00))
	return append(key, []byte(assistant)...)
}

func QueryUsageKey(address string) []byte {
	return append(QueryUsageKeyPrefix, []byte(address)...)
}

func QuotaKey(address string) []byte {
	return append(QuotaKeyPrefix, []byte(address)...)
}

func ModelKey(modelHash string) []byte {
	return append(ModelKeyPrefix, []byte(modelHash)...)
}

func ModelVersionKey(name, version string) []byte {
	key := append(ModelVersionKeyPrefix, []byte(name)...)
	key = append(key, byte(0x00))
	return append(key, []byte(version)...)
}

func CacheKey(queryHash string) []byte {
	return append(CacheKeyPrefix, []byte(queryHash)...)
}

func AuditLogKey(id uint64) []byte {
	return append(AuditLogKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func AnalyticsKey(period string, timestamp time.Time) []byte {
	key := append(AnalyticsKeyPrefix, []byte(period)...)
	key = append(key, byte(0x00))
	return append(key, []byte(timestamp.Format("2006-01-02-15"))...)
}

func AnalyticsSnapshotKeyPrefix(period string) []byte {
	key := append(AnalyticsSnapshotKeyPrefixBase, []byte(period)...)
	return append(key, byte(0x00))
}

func AnalyticsSnapshotKey(period string, timestamp time.Time) []byte {
	key := AnalyticsSnapshotKeyPrefix(period)
	return append(key, []byte(timestamp.Format("2006-01-02-15"))...)
}
