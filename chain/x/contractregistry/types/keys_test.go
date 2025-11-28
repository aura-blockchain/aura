package types

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContractInfoKey(t *testing.T) {
	contractAddr := "cosmos1contract"
	key := ContractInfoKey(contractAddr)

	require.True(t, len(key) > len(ContractInfoPrefix))
	require.Equal(t, ContractInfoPrefix, key[:len(ContractInfoPrefix)])
	require.Equal(t, []byte(contractAddr), key[len(ContractInfoPrefix):])
}

func TestContractMetricsKey(t *testing.T) {
	contractAddr := "cosmos1contract"
	key := ContractMetricsKey(contractAddr)

	require.True(t, len(key) > len(ContractMetricsPrefix))
	require.Equal(t, ContractMetricsPrefix, key[:len(ContractMetricsPrefix)])
}

func TestCreatorContractsKey(t *testing.T) {
	creator := "cosmos1creator"
	contract := "cosmos1contract"

	key := CreatorContractsKey(creator, contract)

	require.Contains(t, string(key), creator)
	require.Contains(t, string(key), contract)
	// Check the prefix is present (base prefix before adding creator)
	require.Equal(t, CreatorContractsKeyPrefix, key[:len(CreatorContractsKeyPrefix)])
}

func TestCreatorContractsIndexKey(t *testing.T) {
	creator := "cosmos1creator"
	// CreatorContractsPrefix is a function, call it to get the prefix
	prefix := CreatorContractsPrefix(creator)

	// Verify the prefix starts with the base prefix
	require.Equal(t, CreatorContractsKeyPrefix, prefix[:len(CreatorContractsKeyPrefix)])
	// Verify creator is in the key
	require.Contains(t, string(prefix), creator)
}

func TestTagContractsKey(t *testing.T) {
	tag := "defi"
	contract := "cosmos1contract"

	key := TagContractsKey(tag, contract)

	require.Contains(t, string(key), tag)
	require.Contains(t, string(key), contract)
	require.Equal(t, TagContractsKeyPrefix, key[:len(TagContractsKeyPrefix)])
}

func TestRateLimitKey(t *testing.T) {
	contractAddr := "cosmos1contract"
	userAddr := "cosmos1user"
	windowStart := int64(1234567890)

	key := RateLimitKey(contractAddr, userAddr, windowStart)

	require.Equal(t, RateLimitPrefix, key[:len(RateLimitPrefix)])

	// Extract window from key (last 8 bytes)
	windowBytes := key[len(key)-8:]
	extractedWindow := int64(binary.BigEndian.Uint64(windowBytes))
	require.Equal(t, windowStart, extractedWindow)
}

func TestAuditEntryKey(t *testing.T) {
	contractAddr := "cosmos1contract"
	id := uint64(123)

	key := AuditEntryKey(contractAddr, id)

	require.Equal(t, AuditEntryPrefix, key[:len(AuditEntryPrefix)])

	// Extract ID from key (last 8 bytes)
	idBytes := key[len(key)-8:]
	extractedID := binary.BigEndian.Uint64(idBytes)
	require.Equal(t, id, extractedID)
}

func TestMigrationKeys(t *testing.T) {
	oldAddr := "cosmos1old"
	newAddr := "cosmos1new"
	migrationID := uint64(42)

	// Test migration record key
	recordKey := MigrationRecordKey(migrationID)
	require.Equal(t, MigrationRecordPrefix, recordKey[:len(MigrationRecordPrefix)])

	idBytes := recordKey[len(MigrationRecordPrefix):]
	extractedID := binary.BigEndian.Uint64(idBytes)
	require.Equal(t, migrationID, extractedID)

	// Test migration from index key
	fromKey := MigrationFromIndexKey(oldAddr, migrationID)
	require.Equal(t, MigrationFromPrefix_Base, fromKey[:len(MigrationFromPrefix_Base)])

	// Test migration to index key
	toKey := MigrationToIndexKey(newAddr, migrationID)
	require.Equal(t, MigrationToPrefix_Base, toKey[:len(MigrationToPrefix_Base)])
}

func TestSecurityKeys(t *testing.T) {
	contractAddr := "cosmos1contract"

	// Test security score key
	scoreKey := SecurityScoreKey(contractAddr)
	require.Equal(t, SecurityScorePrefix, scoreKey[:len(SecurityScorePrefix)])

	// Test whitelist key
	whitelistKey := WhitelistKey(contractAddr)
	require.Equal(t, WhitelistPrefix, whitelistKey[:len(WhitelistPrefix)])

	// Test blacklist key
	blacklistKey := BlacklistKey(contractAddr)
	require.Equal(t, BlacklistPrefix, blacklistKey[:len(BlacklistPrefix)])
}

func TestKeyPrefixUniqueness(t *testing.T) {
	// Ensure all key prefixes are unique
	prefixes := [][]byte{
		ContractInfoPrefix,
		ContractMetricsPrefix,
		CreatorContractsKeyPrefix, // Use KeyPrefix instead of function
		TagContractsKeyPrefix,
		RateLimitPrefix,
		ParamsKey,
		AuditEntryPrefix,
		AuditSequencePrefix,
		VerificationPrefix,
		AuditReportPrefix,
		SecurityScorePrefix,
		WhitelistPrefix,
		BlacklistPrefix,
		MigrationRecordPrefix,
		MigrationFromPrefix_Base,
		MigrationToPrefix_Base,
	}

	seen := make(map[string]bool)
	for _, prefix := range prefixes {
		key := string(prefix)
		require.False(t, seen[key], "duplicate prefix found")
		seen[key] = true
	}
}
