package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

func TestInitGenesisWithState(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	genesis := types.GenesisState{
		Params: &inclusionroutinespb.Params{
			MaxIrPerLocale:       100,
			DefaultRateLimitHour: 10,
			SuspensionFee:        "1000",
			MinGovernanceDeposit: "5000",
		},
		Irs: []*inclusionroutinespb.IRDefinition{
			{
				Id:          "ir-gen-1",
				Name:        "Genesis IR 1",
				Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
				Description: "test genesis ir",
				Score:       50,
				LocaleTags:  []string{"global"},
				PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
				Version:     "1.0",
				Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
			},
		},
		Prerequisites: []*inclusionroutinespb.IRPrerequisite{
			{IrId: "ir-gen-1", RequiredIrIds: []string{}},
		},
		RateLimits: []*inclusionroutinespb.IRRateLimit{
			{IrId: "ir-gen-1", PerWalletPerHour: 1, PerWalletPerDay: 2},
		},
	}

	require.NoError(t, keeper.InitGenesis(ctx, genesis))

	ir, ok := keeper.GetIR(ctx, "ir-gen-1")
	require.True(t, ok)
	require.Equal(t, "Genesis IR 1", ir.Name)

	params := keeper.GetParams()
	require.Equal(t, int32(100), params.MaxIrPerLocale)
}

func TestExportGenesisReflectsState(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	ir := types.IRDefinition{
		Id:          "ir-export",
		Name:        "Export IR",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Description: "exportable",
		Score:       10,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		Version:     "1.0",
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, keeper.CreateIR(ctx, ir))
	require.NoError(t, keeper.SetRateLimit(ctx, types.IRRateLimit{IrId: ir.Id, PerWalletPerHour: 5}))

	export := keeper.ExportGenesis(ctx)
	require.NotNil(t, export.Params)
	require.Len(t, export.Irs, 1)
	require.Equal(t, ir.Id, export.Irs[0].Id)
	require.Len(t, export.RateLimits, 1)
}

func TestDefaultGenesisValidation(t *testing.T) {
	genesis := types.DefaultGenesisState()
	require.NoError(t, types.ValidateGenesisState(genesis))
}
