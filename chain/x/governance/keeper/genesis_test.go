package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/aequitas/aura/chain/x/governance/types"
)

func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey("governance")
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	keeper := NewKeeper(cdc, storeKey)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger()).WithKVGasConfig(storetypes.GasConfig{})

	return keeper, ctx
}

func TestInitGenesis(t *testing.T) {
	tests := []struct {
		name    string
		genesis types.GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid genesis with all data",
			genesis: types.GenesisState{
				Params: &types.GovernanceParams{
					MinDeposit:        "1000",
					MaxDepositPeriod:  durationpb.New(172800000000000),  // 2 days in nanoseconds
					VotingPeriod:      durationpb.New(604800000000000),  // 7 days in nanoseconds
					Quorum:            "0.334",
					Threshold:         "0.5",
					VetoThreshold:     "0.334",
					MinInitialDeposit: "100",
					BurnVoteQuorum:    true,
					BurnVoteVeto:      true,
					BurnProposalDeposit: false,
				},
			},
			wantErr: false,
		},
		{
			name: "default genesis",
			genesis: types.GenesisState{
				Params: nil,
			},
			wantErr: false,
		},
		{
			name: "invalid genesis - nil params in validation",
			genesis: types.GenesisState{
				Params: nil,
			},
			wantErr: false, // InitGenesis uses defaults for nil params
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupKeeper(t)

			// Validate genesis state first
			err := tt.genesis.Validate()
			if tt.wantErr && tt.errMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}

			err = keeper.InitGenesis(ctx, tt.genesis)
			require.NoError(t, err)

			// Verify params were set
			p := keeper.GetParams(ctx)
			require.NotNil(t, p)
			require.NotEmpty(t, p.MinDeposit)
		})
	}
}

func TestExportGenesis(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Set custom params
	params := &types.GovernanceParams{
		MinDeposit:        "2000",
		MaxDepositPeriod:  durationpb.New(259200000000000),  // 3 days
		VotingPeriod:      durationpb.New(1209600000000000), // 14 days
		Quorum:            "0.4",
		Threshold:         "0.6",
		VetoThreshold:     "0.33",
		MinInitialDeposit: "200",
		BurnVoteQuorum:    false,
		BurnVoteVeto:      true,
		BurnProposalDeposit: true,
	}
	keeper.SetParams(ctx, params)

	// Export genesis
	exported := keeper.ExportGenesis(ctx)

	// Verify exported data
	require.NotNil(t, exported.Params)
	require.Equal(t, "2000", exported.Params.MinDeposit)
	require.Equal(t, "0.4", exported.Params.Quorum)
	require.Equal(t, "0.6", exported.Params.Threshold)
	require.Equal(t, "0.33", exported.Params.VetoThreshold)
	require.True(t, exported.Params.BurnVoteVeto)
	require.False(t, exported.Params.BurnVoteQuorum)
}

func TestGenesisRoundTrip(t *testing.T) {
	// Create first keeper with initial state
	keeper1, ctx1 := setupKeeper(t)

	// Set custom params
	params := &types.GovernanceParams{
		MinDeposit:        "5000",
		MaxDepositPeriod:  durationpb.New(345600000000000),  // 4 days
		VotingPeriod:      durationpb.New(1814400000000000), // 21 days
		Quorum:            "0.35",
		Threshold:         "0.55",
		VetoThreshold:     "0.35",
		MinInitialDeposit: "500",
		BurnVoteQuorum:    true,
		BurnVoteVeto:      false,
		BurnProposalDeposit: true,
	}
	keeper1.SetParams(ctx1, params)

	// Export genesis from keeper1
	exported := keeper1.ExportGenesis(ctx1)

	// Create a new keeper and import the exported genesis
	keeper2, ctx2 := setupKeeper(t)
	err := keeper2.InitGenesis(ctx2, exported)
	require.NoError(t, err)

	// Verify all data was preserved
	params1 := keeper1.GetParams(ctx1)
	params2 := keeper2.GetParams(ctx2)
	require.Equal(t, params1.MinDeposit, params2.MinDeposit)
	require.Equal(t, params1.Quorum, params2.Quorum)
	require.Equal(t, params1.Threshold, params2.Threshold)
	require.Equal(t, params1.VetoThreshold, params2.VetoThreshold)
	require.Equal(t, params1.BurnVoteQuorum, params2.BurnVoteQuorum)
	require.Equal(t, params1.BurnVoteVeto, params2.BurnVoteVeto)
	require.Equal(t, params1.BurnProposalDeposit, params2.BurnProposalDeposit)

	// Export again and verify consistency
	exported2 := keeper2.ExportGenesis(ctx2)
	require.Equal(t, exported.Params.MinDeposit, exported2.Params.MinDeposit)
	require.Equal(t, exported.Params.Quorum, exported2.Params.Quorum)
	require.Equal(t, exported.Params.Threshold, exported2.Params.Threshold)
}

func TestDefaultGenesis(t *testing.T) {
	// Test that default genesis is valid
	defaultGen := types.DefaultGenesis()
	require.NotNil(t, defaultGen)

	// Validate default genesis
	err := defaultGen.Validate()
	require.NoError(t, err)

	// Verify default params
	require.NotNil(t, defaultGen.Params)
	require.NotEmpty(t, defaultGen.Params.MinDeposit)
	require.NotEmpty(t, defaultGen.Params.Quorum)
	require.NotEmpty(t, defaultGen.Params.Threshold)
	require.NotEmpty(t, defaultGen.Params.VetoThreshold)
	require.NotNil(t, defaultGen.Params.MaxDepositPeriod)
	require.NotNil(t, defaultGen.Params.VotingPeriod)

	// Test importing default genesis
	keeper, ctx := setupKeeper(t)
	err = keeper.InitGenesis(ctx, *defaultGen)
	require.NoError(t, err)

	// Verify keeper state after importing default genesis
	p := keeper.GetParams(ctx)
	require.NotNil(t, p)
	require.NotEmpty(t, p.MinDeposit)
	require.NotEmpty(t, p.Quorum)
}

func TestInitGenesis_WithCustomParams(t *testing.T) {
	genesis := types.GenesisState{
		Params: &types.GovernanceParams{
			MinDeposit:        "10000",
			MaxDepositPeriod:  durationpb.New(432000000000000),  // 5 days
			VotingPeriod:      durationpb.New(2419200000000000), // 28 days
			Quorum:            "0.5",
			Threshold:         "0.667",
			VetoThreshold:     "0.4",
			MinInitialDeposit: "1000",
			BurnVoteQuorum:    false,
			BurnVoteVeto:      false,
			BurnProposalDeposit: false,
		},
	}

	keeper, ctx := setupKeeper(t)
	err := keeper.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify custom params were set
	params := keeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, "10000", params.MinDeposit)
	require.Equal(t, "0.5", params.Quorum)
	require.Equal(t, "0.667", params.Threshold)
	require.Equal(t, "0.4", params.VetoThreshold)
	require.Equal(t, "1000", params.MinInitialDeposit)
	require.False(t, params.BurnVoteQuorum)
	require.False(t, params.BurnVoteVeto)
	require.False(t, params.BurnProposalDeposit)
}

func TestInitGenesis_NilParams(t *testing.T) {
	genesis := types.GenesisState{
		Params: nil,
	}

	keeper, ctx := setupKeeper(t)
	err := keeper.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify default params were set
	params := keeper.GetParams(ctx)
	require.NotNil(t, params)
	require.NotEmpty(t, params.MinDeposit)
	require.NotEmpty(t, params.Quorum)
}

func TestExportGenesis_DefaultState(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Don't set any params, should use defaults
	exported := keeper.ExportGenesis(ctx)

	// Verify default params are exported
	require.NotNil(t, exported.Params)
	require.NotEmpty(t, exported.Params.MinDeposit)
	require.NotEmpty(t, exported.Params.Quorum)
	require.NotEmpty(t, exported.Params.Threshold)
	require.NotEmpty(t, exported.Params.VetoThreshold)
}

func TestGenesisRoundTrip_MultipleIterations(t *testing.T) {
	// Test that multiple round trips preserve data integrity
	keeper1, ctx1 := setupKeeper(t)

	initialParams := &types.GovernanceParams{
		MinDeposit:        "3000",
		MaxDepositPeriod:  durationpb.New(259200000000000),  // 3 days
		VotingPeriod:      durationpb.New(1209600000000000), // 14 days
		Quorum:            "0.45",
		Threshold:         "0.65",
		VetoThreshold:     "0.38",
		MinInitialDeposit: "300",
		BurnVoteQuorum:    true,
		BurnVoteVeto:      true,
		BurnProposalDeposit: false,
	}
	keeper1.SetParams(ctx1, initialParams)

	// First export
	exported1 := keeper1.ExportGenesis(ctx1)

	// Import to keeper2
	keeper2, ctx2 := setupKeeper(t)
	err := keeper2.InitGenesis(ctx2, exported1)
	require.NoError(t, err)

	// Second export
	exported2 := keeper2.ExportGenesis(ctx2)

	// Import to keeper3
	keeper3, ctx3 := setupKeeper(t)
	err = keeper3.InitGenesis(ctx3, exported2)
	require.NoError(t, err)

	// Third export
	exported3 := keeper3.ExportGenesis(ctx3)

	// Verify all exports are consistent
	require.Equal(t, exported1.Params.MinDeposit, exported2.Params.MinDeposit)
	require.Equal(t, exported2.Params.MinDeposit, exported3.Params.MinDeposit)
	require.Equal(t, exported1.Params.Quorum, exported2.Params.Quorum)
	require.Equal(t, exported2.Params.Quorum, exported3.Params.Quorum)
	require.Equal(t, exported1.Params.Threshold, exported2.Params.Threshold)
	require.Equal(t, exported2.Params.Threshold, exported3.Params.Threshold)
}
