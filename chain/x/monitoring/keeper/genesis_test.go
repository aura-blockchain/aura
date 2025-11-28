package keeper

import (
	"context"
	"testing"

	"github.com/aequitas/aura/chain/x/monitoring/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	t.Run("init with nil genesis", func(t *testing.T) {
		k, ctx := setupKeeper(t)

		err := k.InitGenesis(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("init with default genesis", func(t *testing.T) {
		k, ctx := setupKeeper(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify params are set
		params := k.GetParams()
		require.NotNil(t, params)
		require.Equal(t, genesis.Params.EnableTransactionMonitoring, params.EnableTransactionMonitoring)
	})

	t.Run("init with custom params", func(t *testing.T) {
		k, ctx := setupKeeper(t)

		genesis := &types.GenesisState{
			Params: types.Params{
				EnableTransactionMonitoring: true,
				EnablePerformanceMonitoring: true,
				EnableSecurityMonitoring:    true,
				EnableHealthChecks:          true,
				MetricsRetentionPeriod:      86400,
				AlertThresholdLatency:       1000,
				AlertThresholdErrorRate:     5,
			},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify custom params
		params := k.GetParams()
		require.True(t, params.EnableTransactionMonitoring)
		require.True(t, params.EnablePerformanceMonitoring)
		require.True(t, params.EnableSecurityMonitoring)
		require.Equal(t, uint64(86400), params.MetricsRetentionPeriod)
		require.Equal(t, uint64(1000), params.AlertThresholdLatency)
	})

	t.Run("init with invalid params fails", func(t *testing.T) {
		k, ctx := setupKeeper(t)

		genesis := &types.GenesisState{
			Params: types.Params{
				MetricsRetentionPeriod: 0, // Invalid - must be positive
			},
		}

		err := k.InitGenesis(ctx, genesis)
		require.Error(t, err)
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("export empty state", func(t *testing.T) {
		k, ctx := setupKeeper(t)

		genesis := k.ExportGenesis(ctx)

		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
	})

	t.Run("export after init preserves params", func(t *testing.T) {
		k, ctx := setupKeeper(t)

		originalGenesis := &types.GenesisState{
			Params: types.Params{
				EnableTransactionMonitoring: true,
				EnablePerformanceMonitoring: false,
				EnableSecurityMonitoring:    true,
				EnableHealthChecks:          true,
				MetricsRetentionPeriod:      7200,
				AlertThresholdLatency:       500,
				AlertThresholdErrorRate:     3,
			},
		}

		err := k.InitGenesis(ctx, originalGenesis)
		require.NoError(t, err)

		exported := k.ExportGenesis(ctx)

		require.Equal(t, originalGenesis.Params.EnableTransactionMonitoring, exported.Params.EnableTransactionMonitoring)
		require.Equal(t, originalGenesis.Params.EnablePerformanceMonitoring, exported.Params.EnablePerformanceMonitoring)
		require.Equal(t, originalGenesis.Params.MetricsRetentionPeriod, exported.Params.MetricsRetentionPeriod)
		require.Equal(t, originalGenesis.Params.AlertThresholdLatency, exported.Params.AlertThresholdLatency)
	})
}

func TestGenesisRoundTrip(t *testing.T) {
	t.Run("init then export produces same state", func(t *testing.T) {
		k, ctx := setupKeeper(t)

		originalGenesis := &types.GenesisState{
			Params: types.Params{
				EnableTransactionMonitoring: true,
				EnablePerformanceMonitoring: true,
				EnableSecurityMonitoring:    false,
				EnableHealthChecks:          true,
				MetricsRetentionPeriod:      3600,
				AlertThresholdLatency:       750,
				AlertThresholdErrorRate:     10,
			},
		}

		// Import
		err := k.InitGenesis(ctx, originalGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesis(ctx)

		// Verify all params match
		require.Equal(t, originalGenesis.Params, exported.Params)
	})

	t.Run("multiple round trips are deterministic", func(t *testing.T) {
		k1, ctx1 := setupKeeper(t)
		k2, ctx2 := setupKeeper(t)

		genesis := types.DefaultGenesisState()
		genesis.Params.MetricsRetentionPeriod = 9999

		// First round trip
		err := k1.InitGenesis(ctx1, genesis)
		require.NoError(t, err)
		export1 := k1.ExportGenesis(ctx1)

		// Second round trip
		err = k2.InitGenesis(ctx2, export1)
		require.NoError(t, err)
		export2 := k2.ExportGenesis(ctx2)

		// Verify exports match
		require.Equal(t, export1.Params, export2.Params)
	})

	t.Run("handles default genesis correctly", func(t *testing.T) {
		k, ctx := setupKeeper(t)

		defaultGenesis := types.DefaultGenesisState()

		err := k.InitGenesis(ctx, defaultGenesis)
		require.NoError(t, err)

		exported := k.ExportGenesis(ctx)

		// Verify exported has same params as default
		require.Equal(t, defaultGenesis.Params, exported.Params)
	})
}

func TestDefaultGenesis(t *testing.T) {
	t.Run("default genesis is valid", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)

		// Verify sensible defaults
		require.Greater(t, genesis.Params.MetricsRetentionPeriod, uint64(0))
		require.Greater(t, genesis.Params.AlertThresholdLatency, uint64(0))
	})

	t.Run("can init with default genesis", func(t *testing.T) {
		k, ctx := setupKeeper(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify params were set
		params := k.GetParams()
		require.Equal(t, genesis.Params, params)
	})

	t.Run("default genesis validation passes", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		// Should not panic or error
		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
	})
}

// setupKeeper creates a keeper for testing
func setupKeeper(t *testing.T) (*Keeper, context.Context) {
	// This is a simplified setup - adjust based on your actual keeper setup
	k := NewTestKeeper(t)
	ctx := context.Background()
	return k, ctx
}
