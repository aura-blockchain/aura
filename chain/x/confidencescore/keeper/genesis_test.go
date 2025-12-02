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

	export := k.ExportGenesis(ctx)
	require.NotNil(t, export.Params)
	require.Len(t, export.UserRecords, 1)
	require.Equal(t, "aura1seed", export.UserRecords[0].WalletAddress)
}
