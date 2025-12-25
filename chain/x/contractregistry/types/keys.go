// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"encoding/binary"
	"math"
)

const (
	// ModuleName defines the module name
	ModuleName = "contractregistry"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	// ContractInfoPrefix is the prefix for contract info
	ContractInfoPrefix = []byte{0x01}

	// ContractMetadataKeyPrefix is the prefix for contract metadata
	ContractMetadataKeyPrefix = []byte{0x01}

	// ContractAddressIndexKeyPrefix is the prefix for contract address index
	ContractAddressIndexKeyPrefix = []byte{0x02}

	// SecurityScoreKeyPrefix is the prefix for security scores
	SecurityScoreKeyPrefix = []byte{0x03}

	// WhitelistKeyPrefix is the prefix for whitelist entries
	WhitelistKeyPrefix = []byte{0x04}

	// BlacklistKeyPrefix is the prefix for blacklist entries
	BlacklistKeyPrefix = []byte{0x05}

	// AuditReportKeyPrefix is the prefix for audit reports
	AuditReportKeyPrefix = []byte{0x06}

	// AuditReportCountKeyPrefix is the prefix for audit report counts
	AuditReportCountKeyPrefix = []byte{0x07}

	// VerificationKeyPrefix is the prefix for contract verifications
	VerificationKeyPrefix = []byte{0x08}

	// MigrationRecordKeyPrefix is the prefix for migration records
	MigrationRecordKeyPrefix = []byte{0x09}

	// MigrationCounterKey is the key for migration ID counter
	MigrationCounterKey = []byte{0x0A}

	// MigrationFromIndexKeyPrefix is the prefix for migration from index
	MigrationFromIndexKeyPrefix = []byte{0x0B}

	// MigrationToIndexKeyPrefix is the prefix for migration to index
	MigrationToIndexKeyPrefix = []byte{0x0C}

	// AuditEntriesKeyPrefix is the prefix for audit trail entries
	AuditEntriesKeyPrefix = []byte{0x0D}

	// AuditSequenceKeyPrefix is the prefix for audit sequence numbers
	AuditSequenceKeyPrefix = []byte{0x0E}

	// ContractMetricsKeyPrefix is the prefix for contract metrics
	ContractMetricsKeyPrefix = []byte{0x0F}

	// VerificationResultKeyPrefix is the prefix for verification results
	VerificationResultKeyPrefix = []byte{0x10}

	// SecurityScoreHistoryKeyPrefix is the prefix for security score history
	SecurityScoreHistoryKeyPrefix = []byte{0x11}

	// Audit trail index prefixes
	AuditActorIndexKeyPrefix = []byte{0x12}
	AuditTypeIndexKeyPrefix  = []byte{0x13}
	AuditRetentionPolicyKey  = []byte{0x14}
	AuditGlobalSequenceKey   = []byte{0x15}

	// Creator contracts index
	CreatorContractsKeyPrefix = []byte{0x16}

	// Contract metrics prefix (if not exists)
	ContractMetricsPrefix = ContractMetricsKeyPrefix

	// WhitelistPrefix returns the prefix for all whitelist entries
	WhitelistPrefix = WhitelistKeyPrefix

	// BlacklistPrefix returns the prefix for all blacklist entries
	BlacklistPrefix = BlacklistKeyPrefix
)

// ContractInfoKey returns the key for a contract's info
func ContractInfoKey(contractAddr string) []byte {
	return append(ContractInfoPrefix, []byte(contractAddr)...)
}

// ContractMetadataKey returns the key for a contract's metadata
func ContractMetadataKey(contractAddr string) []byte {
	return append(ContractMetadataKeyPrefix, []byte(contractAddr)...)
}

// SecurityScoreKey returns the key for a contract's security score
func SecurityScoreKey(contractAddr string) []byte {
	return append(SecurityScoreKeyPrefix, []byte(contractAddr)...)
}

// WhitelistKey returns the key for a whitelist entry
func WhitelistKey(contractAddr string) []byte {
	return append(WhitelistKeyPrefix, []byte(contractAddr)...)
}

// BlacklistKey returns the key for a blacklist entry
func BlacklistKey(contractAddr string) []byte {
	return append(BlacklistKeyPrefix, []byte(contractAddr)...)
}

// AuditReportKey returns the key for an audit report
func AuditReportKey(contractAddr string, id uint64) []byte {
	key := append(AuditReportKeyPrefix, []byte(contractAddr)...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return append(key, bz...)
}

// AuditReportCountKey returns the key for audit report count
func AuditReportCountKey(contractAddr string) []byte {
	return append(AuditReportCountKeyPrefix, []byte(contractAddr)...)
}

// VerificationKey returns the key for contract verification
func VerificationKey(contractAddr string) []byte {
	return append(VerificationKeyPrefix, []byte(contractAddr)...)
}

// MigrationRecordKey returns the key for a migration record
func MigrationRecordKey(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return append(MigrationRecordKeyPrefix, bz...)
}

// MigrationFromIndexKey returns the key for migration from index
func MigrationFromIndexKey(contractAddr string, id uint64) []byte {
	key := append(MigrationFromIndexKeyPrefix, []byte(contractAddr)...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return append(key, bz...)
}

// MigrationToIndexKey returns the key for migration to index
func MigrationToIndexKey(contractAddr string, id uint64) []byte {
	key := append(MigrationToIndexKeyPrefix, []byte(contractAddr)...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return append(key, bz...)
}

// MigrationFromPrefix returns the prefix for all migrations from a contract
func MigrationFromPrefix(contractAddr string) []byte {
	return append(MigrationFromIndexKeyPrefix, []byte(contractAddr)...)
}

// MigrationToPrefix returns the prefix for all migrations to a contract
func MigrationToPrefix(contractAddr string) []byte {
	return append(MigrationToIndexKeyPrefix, []byte(contractAddr)...)
}

// AuditEntryKey returns the key for an audit entry
func AuditEntryKey(contractAddr string, seq uint64) []byte {
	key := append(AuditEntriesKeyPrefix, []byte(contractAddr)...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	return append(key, bz...)
}

// AuditEntriesPrefix returns the prefix for all audit entries of a contract
func AuditEntriesPrefix(contractAddr string) []byte {
	return append(AuditEntriesKeyPrefix, []byte(contractAddr)...)
}

// AuditSequenceKey returns the key for audit sequence number
func AuditSequenceKey(contractAddr string) []byte {
	return append(AuditSequenceKeyPrefix, []byte(contractAddr)...)
}

// ContractMetricsKey returns the key for contract metrics
func ContractMetricsKey(contractAddr string) []byte {
	return append(ContractMetricsKeyPrefix, []byte(contractAddr)...)
}

// VerificationResultKey returns the key for a contract's verification result
func VerificationResultKey(contractAddr string) []byte {
	return append(VerificationResultKeyPrefix, []byte(contractAddr)...)
}

// SecurityScoreHistoryKey returns the key for security score history
func SecurityScoreHistoryKey(contractAddr string, timestamp uint64) []byte {
	key := append(SecurityScoreHistoryKeyPrefix, []byte(contractAddr)...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, timestamp)
	return append(key, bz...)
}

// SecurityScoreHistoryPrefix returns the prefix for all security score history entries
func SecurityScoreHistoryPrefix(contractAddr string) []byte {
	return append(SecurityScoreHistoryKeyPrefix, []byte(contractAddr)...)
}

// AuditActorIndexKey returns the key for actor index
func AuditActorIndexKey(actor string, seq uint64) []byte {
	key := append(AuditActorIndexKeyPrefix, []byte(actor)...)
	key = append(key, byte('/'))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	return append(key, bz...)
}

// AuditActorPrefix returns the prefix for all audit entries by actor
func AuditActorPrefix(actor string) []byte {
	key := append(AuditActorIndexKeyPrefix, []byte(actor)...)
	return append(key, byte('/'))
}

// AuditTypeIndexKey returns the key for event type index
func AuditTypeIndexKey(eventType string, seq uint64) []byte {
	key := append(AuditTypeIndexKeyPrefix, []byte(eventType)...)
	key = append(key, byte('/'))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	return append(key, bz...)
}

// AuditTypePrefix returns the prefix for all audit entries by type
func AuditTypePrefix(eventType string) []byte {
	key := append(AuditTypeIndexKeyPrefix, []byte(eventType)...)
	return append(key, byte('/'))
}

// CreatorContractsKey returns the key for contracts by creator
func CreatorContractsKey(creator string, contractAddr string) []byte {
	key := append(CreatorContractsKeyPrefix, []byte(creator)...)
	key = append(key, byte('/'))
	return append(key, []byte(contractAddr)...)
}

// CreatorContractsPrefix returns the prefix for all contracts by creator
func CreatorContractsPrefix(creator string) []byte {
	key := append(CreatorContractsKeyPrefix, []byte(creator)...)
	return append(key, byte('/'))
}

// Additional key prefixes for testing compatibility
var (
	// TagContractsKeyPrefix is the prefix for contracts by tag
	TagContractsKeyPrefix = []byte{0x17}

	// RateLimitKeyPrefix is the prefix for rate limit entries
	RateLimitKeyPrefix = []byte{0x18}

	// AuditEntryPrefix is the prefix for audit entries (alias for compatibility)
	AuditEntryPrefix = AuditEntriesKeyPrefix

	// AuditSequencePrefix is the prefix for audit sequences (alias for compatibility)
	AuditSequencePrefix = AuditSequenceKeyPrefix

	// VerificationPrefix is the prefix for verifications (alias for compatibility)
	VerificationPrefix = VerificationKeyPrefix

	// AuditReportPrefix is the prefix for audit reports (alias for compatibility)
	AuditReportPrefix = AuditReportKeyPrefix

	// SecurityScorePrefix is the prefix for security scores (alias for compatibility)
	SecurityScorePrefix = SecurityScoreKeyPrefix

	// MigrationRecordPrefix is the prefix for migration records (alias for compatibility)
	MigrationRecordPrefix = MigrationRecordKeyPrefix

	// MigrationFromPrefix_Base is the base prefix for migration from index
	MigrationFromPrefix_Base = MigrationFromIndexKeyPrefix

	// MigrationToPrefix_Base is the base prefix for migration to index
	MigrationToPrefix_Base = MigrationToIndexKeyPrefix

	// CreatorContractsPrefix is the prefix for creator contracts index (alias)
	CreatorContractsPrefix_Base = CreatorContractsKeyPrefix

	// RateLimitPrefix is the prefix for rate limits
	RateLimitPrefix = RateLimitKeyPrefix

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x19}
)

// TagContractsKey returns the key for contracts by tag
func TagContractsKey(tag string, contractAddr string) []byte {
	key := append(TagContractsKeyPrefix, []byte(tag)...)
	key = append(key, byte('/'))
	return append(key, []byte(contractAddr)...)
}

// TagContractsPrefix returns the prefix for all contracts by tag
func TagContractsPrefix(tag string) []byte {
	key := append(TagContractsKeyPrefix, []byte(tag)...)
	return append(key, byte('/'))
}

// RateLimitKey returns the key for rate limit tracking
func RateLimitKey(contractAddr string, userAddr string, windowStart int64) []byte {
	key := append(RateLimitKeyPrefix, []byte(contractAddr)...)
	key = append(key, byte('/'))
	key = append(key, []byte(userAddr)...)
	key = append(key, byte('/'))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, safeInt64ToUint64(windowStart))
	return append(key, bz...)
}

// RateLimitPrefixForContract returns the prefix for all rate limit entries for a contract and user
func RateLimitPrefixForContract(contractAddr string, userAddr string) []byte {
	key := append(RateLimitKeyPrefix, []byte(contractAddr)...)
	key = append(key, byte('/'))
	key = append(key, []byte(userAddr)...)
	return append(key, byte('/'))
}

func safeInt64ToUint64(v int64) uint64 {
	if v < 0 {
		return 0
	}
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return uint64(v)
}
