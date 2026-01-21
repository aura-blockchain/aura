// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	t.Run("init with nil genesis", func(t *testing.T) {
		k, ctx := NewTestKeeper(t)

		err := k.InitGenesis(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("init with default genesis", func(t *testing.T) {
		k, ctx := NewTestKeeper(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify params are set
		params, err := k.GetParams(ctx)
		require.NoError(t, err)
		require.Equal(t, genesis.Params.EnableTransactionMonitoring, params.EnableTransactionMonitoring)
	})

	t.Run("init with custom params", func(t *testing.T) {
		k, ctx := NewTestKeeper(t)

		// Start with default params and modify specific fields
		customParams := types.DefaultParams()
		customParams.EnableTransactionMonitoring = true
		customParams.EnableAlerts = true
		customParams.EnableAnomalyDetection = true
		customParams.MetricsRetentionPeriod = 86400 * time.Second

		genesis := &types.GenesisState{
			Params: customParams,
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify custom params
		params, err := k.GetParams(ctx)
		require.NoError(t, err)
		require.True(t, params.EnableTransactionMonitoring)
		require.True(t, params.EnableAlerts)
		require.True(t, params.EnableAnomalyDetection)
		require.Equal(t, 86400*time.Second, params.MetricsRetentionPeriod)
	})

	t.Run("init with invalid params fails", func(t *testing.T) {
		k, ctx := NewTestKeeper(t)

		genesis := &types.GenesisState{
			Params: types.Params{
				LargeTransactionThreshold: 0, // Invalid - must be positive
			},
		}

		err := k.InitGenesis(ctx, genesis)
		require.Error(t, err)
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("export empty state", func(t *testing.T) {
		k, ctx := NewTestKeeper(t)

		genesis := k.ExportGenesis(ctx)

		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
	})

	t.Run("export after init preserves params", func(t *testing.T) {
		k, ctx := NewTestKeeper(t)

		// Start with default params and modify specific fields
		customParams := types.DefaultParams()
		customParams.EnableTransactionMonitoring = true
		customParams.EnableAlerts = false
		customParams.EnableAnomalyDetection = true
		customParams.EnablePrometheusMetrics = true
		customParams.MetricsRetentionPeriod = 7200 * time.Second

		originalGenesis := &types.GenesisState{
			Params: customParams,
		}

		err := k.InitGenesis(ctx, originalGenesis)
		require.NoError(t, err)

		exported := k.ExportGenesis(ctx)

		require.Equal(t, originalGenesis.Params.EnableTransactionMonitoring, exported.Params.EnableTransactionMonitoring)
		require.Equal(t, originalGenesis.Params.EnableAlerts, exported.Params.EnableAlerts)
		require.Equal(t, originalGenesis.Params.EnableAnomalyDetection, exported.Params.EnableAnomalyDetection)
		require.Equal(t, originalGenesis.Params.MetricsRetentionPeriod, exported.Params.MetricsRetentionPeriod)
	})
}

func TestGenesisRoundTrip(t *testing.T) {
	t.Run("init then export produces same state", func(t *testing.T) {
		k, ctx := NewTestKeeper(t)

		// Start with default params and modify specific fields
		customParams := types.DefaultParams()
		customParams.EnableTransactionMonitoring = true
		customParams.EnableAlerts = true
		customParams.EnableAnomalyDetection = false
		customParams.EnablePrometheusMetrics = true
		customParams.MetricsRetentionPeriod = 3600 * time.Second

		originalGenesis := &types.GenesisState{
			Params: customParams,
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
		k1, ctx1 := NewTestKeeper(t)
		k2, ctx2 := NewTestKeeper(t)

		genesis := types.DefaultGenesisState()
		genesis.Params.MetricsRetentionPeriod = 9999 * time.Second

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
		k, ctx := NewTestKeeper(t)

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
		require.Greater(t, genesis.Params.MetricsRetentionPeriod, time.Duration(0))
		require.Greater(t, genesis.Params.LargeTransactionThreshold, uint64(0))
	})

	t.Run("can init with default genesis", func(t *testing.T) {
		k, ctx := NewTestKeeper(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify params were set
		params, err := k.GetParams(ctx)
		require.NoError(t, err)
		require.Equal(t, genesis.Params, params)
	})

	t.Run("default genesis validation passes", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		// Should not panic or error
		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
	})
}
