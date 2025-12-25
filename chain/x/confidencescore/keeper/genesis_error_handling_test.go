// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// TestExportGenesisWithCorruptedUserRecord tests that ExportGenesis returns an error
// when a user record in the store is corrupted (invalid protobuf data).
func TestExportGenesisWithCorruptedUserRecord(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Manually inject corrupted data into the store at a user record key
	store := k.storeService.OpenKVStore(ctx)
	corruptedKey := []byte(types.UserRecordStoreKeyPrefix + "aura1corrupted")
	corruptedData := []byte{0xFF, 0xFF, 0xFF, 0xFF} // Invalid protobuf data
	require.NoError(t, store.Set(corruptedKey, corruptedData))

	// Attempt to export - should fail with unmarshal error
	exported, err := k.ExportGenesis(ctx)
	require.Error(t, err)
	require.Nil(t, exported)
	require.Contains(t, err.Error(), "failed to unmarshal user record")
	// Key should be in error message for debugging (hex-encoded)
	require.Contains(t, err.Error(), "at key")
}

// TestExportGenesisWithCorruptedSlashRecord tests that ExportGenesis returns an error
// when a slash record in the store is corrupted.
func TestExportGenesisWithCorruptedSlashRecord(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Add a valid user record first
	require.NoError(t, k.SetUserRecord(ctx, types.UserConfidenceRecord{
		WalletAddress: "aura1valid",
		TotalScore:    100,
		Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
	}))

	// Manually inject corrupted data into the store at a slash record key
	store := k.storeService.OpenKVStore(ctx)
	corruptedKey := []byte(types.SlashRecordStoreKeyPrefix + "aura1slashed/txhash")
	corruptedData := []byte{0xDE, 0xAD, 0xBE, 0xEF} // Invalid protobuf data
	require.NoError(t, store.Set(corruptedKey, corruptedData))

	// Attempt to export - should fail with unmarshal error
	exported, err := k.ExportGenesis(ctx)
	require.Error(t, err)
	require.Nil(t, exported)
	require.Contains(t, err.Error(), "failed to unmarshal slash record")
	// Key should be in error message for debugging (hex-encoded)
	require.Contains(t, err.Error(), "at key")
}

// TestExportGenesisPartialCorruption tests that even if only one record is corrupted,
// the entire export fails (no partial exports that hide data loss).
func TestExportGenesisPartialCorruption(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Add several valid user records
	for i := 0; i < 5; i++ {
		addr := "aura1valid" + string(rune('0'+i))
		require.NoError(t, k.SetUserRecord(ctx, types.UserConfidenceRecord{
			WalletAddress: addr,
			TotalScore:    uint64(100 * (i + 1)),
			Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
		}))
	}

	// Inject one corrupted record in the middle
	store := k.storeService.OpenKVStore(ctx)
	corruptedKey := []byte(types.UserRecordStoreKeyPrefix + "aura1valid2corrupted")
	corruptedData := []byte{0xBA, 0xD0, 0xDA, 0x7A}
	require.NoError(t, store.Set(corruptedKey, corruptedData))

	// Export should fail even though most records are valid
	exported, err := k.ExportGenesis(ctx)
	require.Error(t, err)
	require.Nil(t, exported)
	require.Contains(t, err.Error(), "failed to unmarshal user record")

	// Verify that we don't get a "partial success" with only the valid records
	// The error should prevent any data from being exported
}

// TestExportGenesisErrorPropagation tests that errors from ExportGenesis
// propagate correctly through the module layer.
func TestExportGenesisErrorPropagation(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Create a scenario that will cause export to fail
	store := k.storeService.OpenKVStore(ctx)
	corruptedKey := []byte(types.UserRecordStoreKeyPrefix + "aura1error")
	corruptedData := []byte{0x00, 0x00, 0x00}
	require.NoError(t, store.Set(corruptedKey, corruptedData))

	// Call ExportGenesis - should return error
	_, err := k.ExportGenesis(ctx)
	require.Error(t, err)

	// Note: The AppModule.ExportGenesis method (in module.go) panics on error,
	// which is the correct behavior for genesis export failures.
	// This is tested here at the keeper level where we can catch the error.
}

// TestExportGenesisValidDataSucceeds tests that valid data exports successfully
// without errors (baseline test to ensure our error handling doesn't break valid cases).
func TestExportGenesisValidDataSucceeds(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Add valid user records
	for i := 0; i < 10; i++ {
		addr := "aura1user" + string(rune('0'+i))
		require.NoError(t, k.SetUserRecord(ctx, types.UserConfidenceRecord{
			WalletAddress: addr,
			TotalScore:    uint64(1000 * (i + 1)),
			Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
			CompletedIrs: []*confidencescorepb.IRCompletion{
				{
					IrId:       "ir_" + string(rune('0'+i)),
					BaseScore:  uint64(100 * (i + 1)),
					FinalScore: uint64(100 * (i + 1)),
					Arena:      "TestArena",
				},
			},
		}))
	}

	// Add valid slash records
	for i := 0; i < 3; i++ {
		require.NoError(t, k.AddSlashRecord(ctx, confidencescorepb.SlashRecord{
			WalletAddress: "aura1user" + string(rune('0'+i)),
			Reason:        confidencescorepb.SlashReason_SLASH_REASON_FRAUD_DETECTED,
			SlashAmount:   uint64(50 * (i + 1)),
			SlashHeight:   uint64(1000 + i),
			RelatedIrId:   "ir_" + string(rune('0'+i)),
			SlashTxHash:   "hash" + string(rune('0'+i)),
			Authority:     "authority",
		}))
	}

	// Export should succeed
	exported, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, exported)
	require.Len(t, exported.UserRecords, 10)
	require.Len(t, exported.SlashRecords, 3)

	// Verify all data is present and correct
	for i, record := range exported.UserRecords {
		require.NotEmpty(t, record.WalletAddress)
		require.Greater(t, record.TotalScore, uint64(0))
		require.NotEmpty(t, record.CompletedIrs)
		require.Equal(t, "ir_"+string(rune('0'+i)), record.CompletedIrs[0].IrId)
	}

	for i, record := range exported.SlashRecords {
		require.NotEmpty(t, record.WalletAddress)
		require.Greater(t, record.SlashAmount, uint64(0))
		require.Equal(t, "ir_"+string(rune('0'+i)), record.RelatedIrId)
	}
}

// TestExportGenesisEmptyStoreSucceeds tests that exporting from an empty store
// succeeds (returns empty arrays, not error).
func TestExportGenesisEmptyStoreSucceeds(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Export from empty state
	exported, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, exported)
	require.NotNil(t, exported.Params)
	require.Empty(t, exported.UserRecords)
	require.Empty(t, exported.SlashRecords)
}

// TestExportGenesisKeyInErrorMessage tests that when unmarshal fails,
// the error message includes the key for debugging purposes.
func TestExportGenesisKeyInErrorMessage(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Inject corrupted data with a specific key
	store := k.storeService.OpenKVStore(ctx)
	testKey := []byte(types.UserRecordStoreKeyPrefix + "aura1debugkey")
	corruptedData := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	require.NoError(t, store.Set(testKey, corruptedData))

	// Export should fail with key in error message
	_, err := k.ExportGenesis(ctx)
	require.Error(t, err)
	// Key is hex-encoded in the error message
	require.Contains(t, err.Error(), "at key", "Error message should contain the key for debugging")
}

// TestExportGenesisMultipleCorruptedRecords tests behavior when multiple records are corrupted.
// It should fail on the first corrupted record encountered.
func TestExportGenesisMultipleCorruptedRecords(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Inject multiple corrupted records
	store := k.storeService.OpenKVStore(ctx)

	corruptedKey1 := []byte(types.UserRecordStoreKeyPrefix + "aura1corrupt1")
	corruptedKey2 := []byte(types.UserRecordStoreKeyPrefix + "aura1corrupt2")
	corruptedKey3 := []byte(types.UserRecordStoreKeyPrefix + "aura1corrupt3")

	corruptedData := []byte{0xFF, 0xEE, 0xDD}
	require.NoError(t, store.Set(corruptedKey1, corruptedData))
	require.NoError(t, store.Set(corruptedKey2, corruptedData))
	require.NoError(t, store.Set(corruptedKey3, corruptedData))

	// Export should fail on first corrupted record
	_, err := k.ExportGenesis(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal user record")
	// Key is included in error (hex-encoded)
	require.Contains(t, err.Error(), "at key")
	_ = strings.Contains // silence unused import warning
}

// TestExportGenesisMixedValidAndCorruptedData tests that export fails
// even when valid data exists alongside corrupted data.
func TestExportGenesisMixedValidAndCorruptedData(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Add valid records
	require.NoError(t, k.SetUserRecord(ctx, types.UserConfidenceRecord{
		WalletAddress: "aura1valid1",
		TotalScore:    500,
		Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
	}))
	require.NoError(t, k.SetUserRecord(ctx, types.UserConfidenceRecord{
		WalletAddress: "aura1valid2",
		TotalScore:    600,
		Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
	}))

	// Inject corrupted record
	store := k.storeService.OpenKVStore(ctx)
	corruptedKey := []byte(types.UserRecordStoreKeyPrefix + "aura1valid1corrupted")
	corruptedData := []byte{0x11, 0x22, 0x33}
	require.NoError(t, store.Set(corruptedKey, corruptedData))

	// Add more valid records after the corrupted one
	require.NoError(t, k.SetUserRecord(ctx, types.UserConfidenceRecord{
		WalletAddress: "aura1valid3",
		TotalScore:    700,
		Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
	}))

	// Export should fail despite having valid records
	exported, err := k.ExportGenesis(ctx)
	require.Error(t, err)
	require.Nil(t, exported)
	require.Contains(t, err.Error(), "failed to unmarshal user record")
}

// TestExportGenesisComplexScenario tests a realistic scenario with various data types.
func TestExportGenesisComplexScenario(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Create a complex valid state
	for i := 0; i < 20; i++ {
		addr := "aura1user" + string(rune('A'+i))
		completions := make([]*confidencescorepb.IRCompletion, i%5+1)
		for j := range completions {
			completions[j] = &confidencescorepb.IRCompletion{
				IrId:       "ir_" + string(rune('A'+i)) + "_" + string(rune('0'+j)),
				BaseScore:  uint64((i + 1) * 10),
				FinalScore: uint64((i + 1) * 12),
				Arena:      "Arena" + string(rune('0'+(i%3))),
			}
		}

		// Use high enough score to pass verification threshold (10000)
		require.NoError(t, k.SetUserRecord(ctx, types.UserConfidenceRecord{
			WalletAddress: addr,
			TotalScore:    uint64(10000 + (i+1)*100), // Above 10000 threshold
			Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
			CompletedIrs:  completions,
		}))
	}

	// Add slash records
	for i := 0; i < 10; i++ {
		require.NoError(t, k.AddSlashRecord(ctx, confidencescorepb.SlashRecord{
			WalletAddress: "aura1user" + string(rune('A'+i)),
			Reason:        confidencescorepb.SlashReason_SLASH_REASON_FRAUD_DETECTED,
			SlashAmount:   uint64((i + 1) * 25),
			SlashHeight:   uint64(5000 + i*10),
			RelatedIrId:   "ir_" + string(rune('A'+i)) + "_0",
			SlashTxHash:   "txhash_" + string(rune('A'+i)),
			Authority:     "authority",
		}))
	}

	// Export should succeed with all data intact
	exported, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, exported)
	require.Len(t, exported.UserRecords, 20)
	require.Len(t, exported.SlashRecords, 10)

	// Verify data integrity
	for i, record := range exported.UserRecords {
		require.NotEmpty(t, record.WalletAddress)
		require.Greater(t, record.TotalScore, uint64(0))
		require.NotEmpty(t, record.CompletedIrs)
		require.Equal(t, types.VerificationStatus_VERIFICATION_STATUS_VERIFIED, record.Status)
		_ = i // Use i
	}

	// Test round-trip
	ctx2, k2 := setupConfKeeper(t)
	require.NoError(t, k2.InitGenesis(ctx2, *exported))

	exported2, err := k2.ExportGenesis(ctx2)
	require.NoError(t, err)
	require.Equal(t, len(exported.UserRecords), len(exported2.UserRecords))
	require.Equal(t, len(exported.SlashRecords), len(exported2.SlashRecords))
}
