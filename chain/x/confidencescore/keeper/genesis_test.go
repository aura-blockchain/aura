// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

func TestInitGenesis(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	genesis := types.GenesisState{
		Params: types.DefaultGenesisState().Params,
		UserRecords: []*confidencescorepb.UserConfidenceRecord{
			{
				WalletAddress: "aura1test",
				TotalScore:    100,
				Status:        types.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
			},
		},
	}

	require.NoError(t, k.InitGenesis(ctx, genesis))

	record, ok := k.GetUserRecord(ctx, "aura1test")
	require.True(t, ok)
	require.Equal(t, uint64(100), record.TotalScore)
}

func TestExportGenesis(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Seed a user record
	require.NoError(t, k.SetUserRecord(ctx, types.UserConfidenceRecord{
		WalletAddress: "aura1seed",
		TotalScore:    200,
		Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
	}))

	export, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, export.Params)
	require.Len(t, export.UserRecords, 1)
	require.Equal(t, "aura1seed", export.UserRecords[0].WalletAddress)
}

func TestExportGenesisRoundTrip(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Create complex state with user records and slash records
	userRecord := types.UserConfidenceRecord{
		WalletAddress: "aura1test",
		TotalScore:    15000, // Above verification threshold of 10,000
		Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
		CompletedIrs: []*confidencescorepb.IRCompletion{
			{
				IrId:       "ir1",
				BaseScore:  100,
				FinalScore: 100,
				Arena:      "Biometric",
			},
		},
	}
	require.NoError(t, k.SetUserRecord(ctx, userRecord))

	slashRecord := confidencescorepb.SlashRecord{
		WalletAddress: "aura1test",
		Reason:        confidencescorepb.SlashReason_SLASH_REASON_FRAUD_DETECTED,
		SlashAmount:   50,
		SlashHeight:   1234567890,
		RelatedIrId:   "ir1",
		SlashTxHash:   "ABC123DEF456",
		Authority:     "authority",
	}
	require.NoError(t, k.AddSlashRecord(ctx, slashRecord))

	// Export genesis
	exported, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, exported)
	require.Len(t, exported.UserRecords, 1)
	require.Len(t, exported.SlashRecords, 1)

	// Create new keeper and import
	ctx2, k2 := setupConfKeeper(t)
	require.NoError(t, k2.InitGenesis(ctx2, *exported))

	// Export from new keeper and verify round-trip
	exported2, err := k2.ExportGenesis(ctx2)
	require.NoError(t, err)
	require.Equal(t, len(exported.UserRecords), len(exported2.UserRecords))
	require.Equal(t, len(exported.SlashRecords), len(exported2.SlashRecords))

	// Verify specific values
	require.Equal(t, exported.UserRecords[0].WalletAddress, exported2.UserRecords[0].WalletAddress)
	require.Equal(t, exported.UserRecords[0].TotalScore, exported2.UserRecords[0].TotalScore)
	require.Equal(t, exported.SlashRecords[0].WalletAddress, exported2.SlashRecords[0].WalletAddress)
}

func TestExportGenesisEmptyState(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Export from empty state
	exported, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, exported)
	require.NotNil(t, exported.Params)
	require.Empty(t, exported.UserRecords)
	require.Empty(t, exported.SlashRecords)
	require.Empty(t, exported.Completions)
	require.Empty(t, exported.History)
}

func TestExportGenesisMultipleRecords(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	// Create multiple user records
	for i := 0; i < 10; i++ {
		require.NoError(t, k.SetUserRecord(ctx, types.UserConfidenceRecord{
			WalletAddress: "aura1test" + string(rune('0'+i)),
			TotalScore:    uint64(100 * (i + 1)),
			Status:        types.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
		}))
	}

	// Create multiple slash records
	for i := 0; i < 5; i++ {
		require.NoError(t, k.AddSlashRecord(ctx, confidencescorepb.SlashRecord{
			WalletAddress: "aura1test" + string(rune('0'+i)),
			Reason:        confidencescorepb.SlashReason_SLASH_REASON_FRAUD_DETECTED,
			SlashAmount:   uint64(10 * (i + 1)),
			SlashHeight:   uint64(1234567890 + i),
			RelatedIrId:   "ir1",
			SlashTxHash:   "ABCDEF" + string(rune('0'+i)),
			Authority:     "authority",
		}))
	}

	// Export and verify all records are present
	exported, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Len(t, exported.UserRecords, 10)
	require.Len(t, exported.SlashRecords, 5)

	// Verify no data loss
	for i, record := range exported.UserRecords {
		require.NotEmpty(t, record.WalletAddress)
		require.Greater(t, record.TotalScore, uint64(0))
		require.Equal(t, types.VerificationStatus_VERIFICATION_STATUS_VERIFIED, record.Status)
		_ = i // Use i to avoid unused variable warning
	}

	for _, record := range exported.SlashRecords {
		require.NotEmpty(t, record.WalletAddress)
		require.Greater(t, record.SlashAmount, uint64(0))
		require.NotEqual(t, confidencescorepb.SlashReason_SLASH_REASON_UNSPECIFIED, record.Reason)
	}
}
