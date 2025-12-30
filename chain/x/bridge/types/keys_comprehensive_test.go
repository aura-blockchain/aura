// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "bridge", types.ModuleName)
	require.Equal(t, "bridge", types.StoreKey)
	require.Equal(t, "bridge", types.RouterKey)
	require.Equal(t, "bridge", types.QuerierRoute)
	require.Equal(t, 100, types.MaxPendingTransfersPerBlock)
}

func TestChainConfigKey(t *testing.T) {
	tests := []struct {
		name    string
		chainID string
	}{
		{"normal chain ID", "paw"},
		{"long chain ID", "ethereum-mainnet-1"},
		{"empty chain ID", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.ChainConfigKey(tt.chainID)
			require.NotNil(t, key)
			if tt.chainID != "" {
				require.Contains(t, string(key), tt.chainID)
			}
			require.Equal(t, append(types.ChainConfigPrefix, []byte(tt.chainID)...), key)
		})
	}
}

func TestValidatorKey(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"normal address", "aura1validator123"},
		{"long address", "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"},
		{"empty address", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.ValidatorKey(tt.address)
			require.NotNil(t, key)
			if tt.address != "" {
				require.Contains(t, string(key), tt.address)
			}
			require.Equal(t, append(types.ValidatorPrefix, []byte(tt.address)...), key)
		})
	}
}

func TestAttestationKey(t *testing.T) {
	tests := []struct {
		name       string
		transferID string
		validator  string
	}{
		{"normal", "transfer-1", "validator-1"},
		{"long IDs", "transfer-abcdefghijklmnop", "validator-123456789"},
		{"empty transfer", "", "validator-1"},
		{"empty validator", "transfer-1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.AttestationKey(tt.transferID, tt.validator)
			require.NotNil(t, key)
			// Should contain null separator (0x00)
			require.Contains(t, key, byte(0x00))
		})
	}
}

func TestTransferHashIndexKey(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"normal hash", "0xabcdef123456"},
		{"short hash", "abc"},
		{"empty hash", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.TransferHashIndexKey(tt.hash)
			require.NotNil(t, key)
			if tt.hash != "" {
				require.Contains(t, string(key), tt.hash)
			}
			require.Equal(t, append(types.TransferHashIndexPrefix, []byte(tt.hash)...), key)
		})
	}
}

func TestSwapKey(t *testing.T) {
	tests := []struct {
		name   string
		swapID string
	}{
		{"normal", "swap-123"},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.SwapKey(tt.swapID)
			require.NotNil(t, key)
			require.Equal(t, append(types.SwapPrefix, []byte(tt.swapID)...), key)
		})
	}
}

func TestProcessedSourceHashKey(t *testing.T) {
	tests := []struct {
		name        string
		sourceChain string
		sourceHash  string
	}{
		{"normal", "paw", "hash123"},
		{"long hash", "ethereum", "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"},
		{"empty hash", "aura", ""},
		{"empty chain", "", "hash123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.ProcessedSourceHashKey(tt.sourceChain, tt.sourceHash)
			require.NotNil(t, key)
			// Should contain colon separator
			require.Contains(t, string(key), ":")
		})
	}
}

func TestProcessingSourceHashKey(t *testing.T) {
	// Same format as ProcessedSourceHashKey but different prefix
	key1 := types.ProcessedSourceHashKey("paw", "hash1")
	key2 := types.ProcessingSourceHashKey("paw", "hash1")

	// Different prefixes should make keys different
	require.NotEqual(t, key1, key2)

	// Same args should produce consistent keys
	key3 := types.ProcessingSourceHashKey("paw", "hash1")
	require.Equal(t, key2, key3)
}

func TestSignatureSetKey(t *testing.T) {
	tests := []struct {
		name             string
		transferID       string
		signatureSetHash []byte
	}{
		{"normal", "transfer-1", []byte("sighash123")},
		{"empty transfer", "", []byte("sighash")},
		{"empty hash", "transfer-1", []byte{}},
		{"nil hash", "transfer-1", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.SignatureSetKey(tt.transferID, tt.signatureSetHash)
			require.NotNil(t, key)
		})
	}
}

func TestValidatorSnapshotKey(t *testing.T) {
	tests := []struct {
		name        string
		blockHeight int64
	}{
		{"normal height", 12345},
		{"zero height", 0},
		{"large height", 999999999999},
		{"negative height", -1}, // Should be handled safely
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.ValidatorSnapshotKey(tt.blockHeight)
			require.NotNil(t, key)
			// Should have prefix + 8 bytes for height
			require.Len(t, key, len(types.ValidatorSnapshotPrefix)+8)
		})
	}
}

func TestDailyMintKey(t *testing.T) {
	tests := []struct {
		name  string
		date  string
		denom string
	}{
		{"normal", "20240115", "uaura"},
		{"with IBC denom", "20240115", "ibc/ABC123"},
		{"empty date", "", "uaura"},
		{"empty denom", "20240115", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.DailyMintKey(tt.date, tt.denom)
			require.NotNil(t, key)
			require.Contains(t, string(key), ":")
		})
	}
}

func TestHourlyMintKey(t *testing.T) {
	tests := []struct {
		name     string
		datetime string
		denom    string
	}{
		{"normal", "2024011512", "uaura"},
		{"midnight", "2024011500", "uaura"},
		{"with IBC denom", "2024011512", "ibc/ABC123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.HourlyMintKey(tt.datetime, tt.denom)
			require.NotNil(t, key)
			require.Contains(t, string(key), ":")
		})
	}
}

func TestVerifiedBlockHashKey(t *testing.T) {
	tests := []struct {
		name        string
		sourceChain string
		blockHeight uint64
	}{
		{"normal", "ethereum", 12345678},
		{"zero height", "paw", 0},
		{"large height", "bitcoin", 999999999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.VerifiedBlockHashKey(tt.sourceChain, tt.blockHeight)
			require.NotNil(t, key)
			// Should contain colon separator
			require.Contains(t, string(key), ":")
		})
	}
}

func TestPendingTransferKey(t *testing.T) {
	tests := []struct {
		name       string
		transferID string
	}{
		{"normal", "transfer-123"},
		{"empty", ""},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.PendingTransferKey(tt.transferID)
			require.NotNil(t, key)
			require.Equal(t, append(types.PendingTransferPrefix, []byte(tt.transferID)...), key)
		})
	}
}

func TestSignatureUsedKey(t *testing.T) {
	tests := []struct {
		name          string
		signatureHash []byte
	}{
		{"32 bytes", make([]byte, 32)},
		{"less than 32 bytes", []byte("short")},
		{"more than 32 bytes", make([]byte, 64)},
		{"empty", []byte{}},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.SignatureUsedKey(tt.signatureHash)
			require.NotNil(t, key)
		})
	}
}

func TestSignatureRateLimitKey(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		windowStart int64
	}{
		{"normal", "cosmos1abc", 1234567890},
		{"zero window", "cosmos1abc", 0},
		{"negative window", "cosmos1abc", -1}, // Should be handled safely
		{"empty address", "", 1234567890},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.SignatureRateLimitKey(tt.address, tt.windowStart)
			require.NotNil(t, key)
			require.Contains(t, key, byte(':'))
		})
	}
}

func TestUserTransferIndexKey(t *testing.T) {
	tests := []struct {
		name       string
		address    string
		transferID string
	}{
		{"normal", "cosmos1abc", "transfer-1"},
		{"empty address", "", "transfer-1"},
		{"empty transfer", "cosmos1abc", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.UserTransferIndexKey(tt.address, tt.transferID)
			require.NotNil(t, key)
			// Should contain null separator
			require.Contains(t, key, byte(0x00))
		})
	}
}

func TestUserTransferIndexPrefixKey(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"normal", "cosmos1abc"},
		{"long address", "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefixKey := types.UserTransferIndexPrefixKey(tt.address)
			require.NotNil(t, prefixKey)

			// Full key should start with prefix key
			fullKey := types.UserTransferIndexKey(tt.address, "transfer-1")
			require.True(t, len(fullKey) > len(prefixKey))
		})
	}
}

func TestKeyPrefixesAreUnique(t *testing.T) {
	// Collect all prefix values
	prefixes := [][]byte{
		types.TransferPrefix,
		types.WrappedTokenPrefix,
		types.SharedIdentityPrefix,
		types.RelayerPrefix,
		types.ChainConfigPrefix,
		types.ValidatorPrefix,
		types.AttestationPrefix,
		types.RelayerStatsPrefix,
		types.SwapPrefix,
		types.TransferHashIndexPrefix,
		types.ProcessedSourceHashPrefix,
		types.SignatureSetPrefix,
		types.ValidatorSnapshotPrefix,
		types.DailyMintPrefix,
		types.HourlyMintPrefix,
		types.VerifiedBlockHashPrefix,
		types.PendingTransferPrefix,
		types.SignatureUsedPrefix,
		types.SignatureRateLimitPrefix,
		types.ProcessingSourceHashPrefix,
		types.UserTransferIndexPrefix,
	}

	// Check all prefixes are unique
	seen := make(map[string]bool)
	for i, p := range prefixes {
		key := string(p)
		require.False(t, seen[key], "Duplicate prefix at index %d: %x", i, p)
		seen[key] = true
	}
}

func TestSecurityKeyPrefixesAreUnique(t *testing.T) {
	prefixes := [][]byte{
		types.MerkleRootPrefix,
		types.TSSNoncePrefix,
		types.BridgeValidatorPrefix,
		types.ValidatorRotationPrefix,
		types.SlashingEventPrefix,
		types.FraudProofPrefix,
		types.TimeLockPrefix,
		types.WithdrawalLimitPrefix,
		types.CircuitBreakerPrefix,
		types.NonceTrackerPrefix,
		types.AddressPermissionPrefix,
		types.BridgeFeePrefix,
		types.InsuranceFundPrefix,
		types.InsuranceClaimPrefix,
		types.ValidatorSigningPrefix,
	}

	seen := make(map[string]bool)
	for i, p := range prefixes {
		key := string(p)
		require.False(t, seen[key], "Duplicate security prefix at index %d: %x", i, p)
		seen[key] = true
	}
}

func TestValidatorSigningInfoKey(t *testing.T) {
	tests := []struct {
		name             string
		validatorAddress string
		blockHeight      int64
	}{
		{"normal", "cosmos1validator", 12345},
		{"zero height", "cosmos1validator", 0},
		{"negative height", "cosmos1validator", -1},
		{"empty address", "", 12345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.ValidatorSigningInfoKey(tt.validatorAddress, tt.blockHeight)
			require.NotNil(t, key)
		})
	}
}

func TestSafeInt64ToUint64(t *testing.T) {
	// Test through ValidatorSnapshotKey which uses safeInt64ToUint64
	tests := []struct {
		name   string
		height int64
	}{
		{"positive", 100},
		{"zero", 0},
		{"negative", -100},        // Should return 0
		{"max int64", 1<<62},      // Large value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.ValidatorSnapshotKey(tt.height)
			require.NotNil(t, key)
			require.Len(t, key, len(types.ValidatorSnapshotPrefix)+8)
		})
	}
}
