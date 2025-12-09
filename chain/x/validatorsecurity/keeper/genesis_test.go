package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// Ensure sdk.Context is used (for type checking)
var _ sdk.Context

func TestInitGenesis(t *testing.T) {
	t.Run("init with default genesis", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify params were set
		params := k.GetParams(ctx)
		require.NotNil(t, params)
	})

	t.Run("init with validator security info", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &types.GenesisState{
			Params: *types.DefaultParams(),
			Validators: []types.ValidatorSecurityInfo{
				{
					ValidatorAddress:    "validator1",
					HotKey:              "hotkey1",
					ColdKey:             "coldkey1",
					KeysSeparated:       true,
					Region:              "us-east",
					CountryCode:         "US",
					SentryNodeAddresses: []string{"sentry1", "sentry2"},
					IsJailed:            false,
					IsTombstoned:        false,
					MissedBlocksCounter: 0,
					IndexOffset:         1000,
				},
				{
					ValidatorAddress:    "validator2",
					HotKey:              "hotkey2",
					ColdKey:             "coldkey2",
					KeysSeparated:       true,
					Region:              "eu-west",
					CountryCode:         "EU",
					SentryNodeAddresses: []string{"sentry3"},
					IsJailed:            false,
					IsTombstoned:        false,
					MissedBlocksCounter: 0,
					IndexOffset:         1000,
				},
			},
			DoubleSignEvidences: []types.DoubleSignEvidence{},
			DowntimeInfractions: []types.DowntimeInfraction{},
			Alerts:              []types.ValidatorAlert{},
			SentryNodes:         []types.SentryNodeInfo{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify validators were imported
		allValidators := k.GetAllValidators(ctx)
		require.Len(t, allValidators, 2)
	})

	t.Run("init with double sign evidences", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &types.GenesisState{
			Params:     *types.DefaultParams(),
			Validators: []types.ValidatorSecurityInfo{},
			DoubleSignEvidences: []types.DoubleSignEvidence{
				{
					ValidatorAddress: "validator1",
					Height:           100,
					VoteA:            []byte("vote_a"),
					VoteB:            []byte("vote_b"),
					SlashFraction:    sdkmath.LegacyMustNewDecFromStr("0.05"),
				},
			},
			DowntimeInfractions: []types.DowntimeInfraction{},
			Alerts:              []types.ValidatorAlert{},
			SentryNodes:         []types.SentryNodeInfo{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify evidence was imported
		evidences := k.GetAllDoubleSignEvidences(ctx)
		require.Len(t, evidences, 1)
	})

	t.Run("init with downtime infractions", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &types.GenesisState{
			Params:              *types.DefaultParams(),
			Validators:          []types.ValidatorSecurityInfo{},
			DoubleSignEvidences: []types.DoubleSignEvidence{},
			DowntimeInfractions: []types.DowntimeInfraction{
				{
					ValidatorAddress: "validator1",
					MissedBlocks:     50,
					WindowSize:       1000,
					SlashFraction:    sdkmath.LegacyMustNewDecFromStr("0.0001"),
				},
				{
					ValidatorAddress: "validator2",
					MissedBlocks:     75,
					WindowSize:       1000,
					SlashFraction:    sdkmath.LegacyMustNewDecFromStr("0.0001"),
				},
			},
			Alerts:      []types.ValidatorAlert{},
			SentryNodes: []types.SentryNodeInfo{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify infractions were imported
		infractions := k.GetAllDowntimeInfractions(ctx)
		require.Len(t, infractions, 2)
	})

	t.Run("init with alerts", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &types.GenesisState{
			Params:              *types.DefaultParams(),
			Validators:          []types.ValidatorSecurityInfo{},
			DoubleSignEvidences: []types.DoubleSignEvidence{},
			DowntimeInfractions: []types.DowntimeInfraction{},
			Alerts: []types.ValidatorAlert{
				{
					Id:               "alert1",
					ValidatorAddress: "validator1",
					AlertType:        types.ValidatorAlert_DOWNTIME,
					Severity:         types.ValidatorAlert_WARNING,
					Message:          "Validator offline",
					Acknowledged:     false,
				},
				{
					Id:               "alert2",
					ValidatorAddress: "validator2",
					AlertType:        types.ValidatorAlert_DOUBLE_SIGN,
					Severity:         types.ValidatorAlert_CRITICAL,
					Message:          "Double sign detected",
					Acknowledged:     true,
				},
			},
			SentryNodes: []types.SentryNodeInfo{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify alerts were imported
		allAlerts := k.GetAllValidatorAlerts(ctx)
		require.Len(t, allAlerts, 2)
	})

	t.Run("init with sentry nodes", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &types.GenesisState{
			Params:              *types.DefaultParams(),
			Validators:          []types.ValidatorSecurityInfo{},
			DoubleSignEvidences: []types.DoubleSignEvidence{},
			DowntimeInfractions: []types.DowntimeInfraction{},
			Alerts:              []types.ValidatorAlert{},
			SentryNodes: []types.SentryNodeInfo{
				{
					Address:          "sentry1",
					ValidatorAddress: "validator1",
					IpAddress:        "192.168.1.1",
					Port:             26656,
					IsActive:         true,
				},
				{
					Address:          "sentry2",
					ValidatorAddress: "validator1",
					IpAddress:        "192.168.1.2",
					Port:             26656,
					IsActive:         true,
				},
			},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify sentry nodes were imported
		sentryNodes := k.GetAllSentryNodes(ctx)
		require.Len(t, sentryNodes, 2)
	})

	t.Run("init with invalid genesis fails", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &types.GenesisState{
			Params: types.ValidatorSecurityParams{
				SignedBlocksWindow: -100, // Invalid - negative value
			},
			Validators:          []types.ValidatorSecurityInfo{},
			DoubleSignEvidences: []types.DoubleSignEvidence{},
			DowntimeInfractions: []types.DowntimeInfraction{},
			Alerts:              []types.ValidatorAlert{},
			SentryNodes:         []types.SentryNodeInfo{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.Error(t, err)
	})

	t.Run("init with geo distribution", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		params := types.DefaultParams()
		params.EnableGeoDistribution = true

		genesis := &types.GenesisState{
			Params: *params,
			Validators: []types.ValidatorSecurityInfo{
				{
					ValidatorAddress: "validator1",
					Region:           "us-east",
					CountryCode:      "US",
					IsJailed:         false,
					IsTombstoned:     false,
				},
				{
					ValidatorAddress: "validator2",
					Region:           "eu-west",
					CountryCode:      "EU",
					IsJailed:         false,
					IsTombstoned:     false,
				},
			},
			DoubleSignEvidences: []types.DoubleSignEvidence{},
			DowntimeInfractions: []types.DowntimeInfraction{},
			Alerts:              []types.ValidatorAlert{},
			SentryNodes:         []types.SentryNodeInfo{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("export empty state", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := k.ExportGenesis(ctx)

		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
		require.Empty(t, genesis.Validators)
		require.Empty(t, genesis.DoubleSignEvidences)
		require.Empty(t, genesis.DowntimeInfractions)
	})

	t.Run("export with data", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		// Initialize with data
		initGenesis := &types.GenesisState{
			Params: types.DefaultParams(),
			Validators: []*types.ValidatorSecurityInfo{
				{ValidatorAddress: "validator1", HotKey: "hk1", ColdKey: "ck1", KeysSeparated: true, IsJailed: false, IsTombstoned: false},
				{ValidatorAddress: "validator2", HotKey: "hk2", ColdKey: "ck2", KeysSeparated: true, IsJailed: false, IsTombstoned: false},
			},
			DoubleSignEvidences: []*types.DoubleSignEvidence{
				&types.DoubleSignEvidence{ValidatorAddress: "validator1", Height: 100, SlashFraction: "0.05"},
			},
			DowntimeInfractions: []*types.DowntimeInfraction{
				&types.DowntimeInfraction{ValidatorAddress: "validator2", MissedBlocks: 50, WindowSize: 1000, SlashFraction: "0.0001"},
			},
			Alerts: []*types.ValidatorAlert{
				&types.ValidatorAlert{Id: "alert1", ValidatorAddress: "validator1", AlertType: types.ValidatorAlert_DOWNTIME, Severity: types.ValidatorAlert_WARNING, Acknowledged: false},
			},
			SentryNodes: []*types.SentryNodeInfo{
				&types.SentryNodeInfo{Address: "sentry1", ValidatorAddress: "validator1", IsActive: true},
			},
		}

		err := k.InitGenesis(ctx, initGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesis(ctx)

		require.Len(t, exported.Validators, 2)
		require.Len(t, exported.DoubleSignEvidences, 1)
		require.Len(t, exported.DowntimeInfractions, 1)
		require.Len(t, exported.SentryNodes, 1)
	})

	t.Run("export only includes active alerts", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		initGenesis := &types.GenesisState{
			Params:              types.DefaultParams(),
			Validators:          make([]*types.ValidatorSecurityInfo, 0),
			DoubleSignEvidences: make([]*types.DoubleSignEvidence, 0),
			DowntimeInfractions: make([]*types.DowntimeInfraction, 0),
			Alerts: []*types.ValidatorAlert{
				{Id: "alert1", ValidatorAddress: "val1", AlertType: types.ValidatorAlert_DOWNTIME, Severity: types.ValidatorAlert_INFO, Acknowledged: false},
				{Id: "alert2", ValidatorAddress: "val1", AlertType: types.ValidatorAlert_DOWNTIME, Severity: types.ValidatorAlert_INFO, Acknowledged: true},
				{Id: "alert3", ValidatorAddress: "val1", AlertType: types.ValidatorAlert_DOWNTIME, Severity: types.ValidatorAlert_INFO, Acknowledged: false},
			},
			SentryNodes: make([]*types.SentryNodeInfo, 0),
		}

		err := k.InitGenesis(ctx, initGenesis)
		require.NoError(t, err)

		exported := k.ExportGenesis(ctx)

		// Only active (unacknowledged) alerts should be exported
		for _, alert := range exported.Alerts {
			require.False(t, alert.Acknowledged)
		}
	})
}

func TestGenesisRoundTrip(t *testing.T) {
	t.Run("init then export produces same state", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		// Use DefaultParams() which has all required fields populated
		originalGenesis := &types.GenesisState{
			Params: types.DefaultParams(),
			Validators: []*types.ValidatorSecurityInfo{
				{
					ValidatorAddress: "validator1",
					HotKey:           "hotkey1",
					ColdKey:          "coldkey1",
					KeysSeparated:    true,
					Region:           "us-east",
					CountryCode:      "US",
					IsJailed:         false,
					IsTombstoned:     false,
				},
			},
			DoubleSignEvidences: []*types.DoubleSignEvidence{
				{ValidatorAddress: "validator1", Height: 100, SlashFraction: "0.05"},
			},
			DowntimeInfractions: []*types.DowntimeInfraction{
				{ValidatorAddress: "validator1", MissedBlocks: 50, WindowSize: 1000, SlashFraction: "0.0001"},
			},
			Alerts: []*types.ValidatorAlert{
				{Id: "alert1", ValidatorAddress: "validator1", AlertType: types.ValidatorAlert_DOWNTIME, Severity: types.ValidatorAlert_WARNING, Acknowledged: false},
			},
			SentryNodes: []*types.SentryNodeInfo{
				{Address: "sentry1", ValidatorAddress: "validator1", IsActive: true},
			},
		}

		// Import
		err := k.InitGenesis(ctx, originalGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesis(ctx)

		// Verify counts match
		require.Len(t, exported.Validators, len(originalGenesis.Validators))
		require.Len(t, exported.DoubleSignEvidences, len(originalGenesis.DoubleSignEvidences))
		require.Len(t, exported.DowntimeInfractions, len(originalGenesis.DowntimeInfractions))
		require.Len(t, exported.SentryNodes, len(originalGenesis.SentryNodes))
	})

	t.Run("multiple round trips are deterministic", func(t *testing.T) {
		k1, ctx1 := setupTestKeeper(t)
		k2, ctx2 := setupTestKeeper(t)

		genesis := types.DefaultGenesisState()
		genesis.Validators = []*types.ValidatorSecurityInfo{
			{ValidatorAddress: "val1", IsJailed: false, IsTombstoned: false},
		}

		// First round trip
		err := k1.InitGenesis(ctx1, genesis)
		require.NoError(t, err)
		export1 := k1.ExportGenesis(ctx1)

		// Second round trip
		err = k2.InitGenesis(ctx2, export1)
		require.NoError(t, err)
		export2 := k2.ExportGenesis(ctx2)

		// Verify exports match
		require.Len(t, export2.Validators, len(export1.Validators))
	})
}

func TestDefaultGenesis(t *testing.T) {
	t.Run("default genesis is valid", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		err := types.ValidateGenesisState(genesis)
		require.NoError(t, err)

		require.NotNil(t, genesis.Params)
		require.NotNil(t, genesis.Validators)
		require.NotNil(t, genesis.DoubleSignEvidences)
		require.NotNil(t, genesis.DowntimeInfractions)
	})

	t.Run("can init with default genesis", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify params were set
		params := k.GetParams(ctx)
		require.NotNil(t, params)
	})

	t.Run("default params are reasonable", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		require.Greater(t, genesis.Params.SignedBlocksWindow, int64(0))
		require.NotEmpty(t, genesis.Params.MinSignedPerWindow)
		require.NotEmpty(t, genesis.Params.DowntimeSlashFraction)
		require.NotEmpty(t, genesis.Params.DoubleSignSlashFraction)
	})
}

func setupTestKeeper(t *testing.T) (Keeper, sdk.Context) {
	// Create test keeper and SDK context
	return NewTestKeeper(t)
}
