package keeper

import (
	"fmt"

	"context"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// InitGenesis initializes the monitoring module state from genesis
func (k *Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	if data == nil {
		return nil
	}

	// Set parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		return fmt.Errorf("error in InitGenesis: %w", err)
	}

	// Note: Monitoring module primarily tracks runtime metrics and doesn't
	// need to restore transient state from genesis. The module will start
	// collecting fresh data upon initialization.

	return nil
}

// ExportGenesis exports the monitoring module state to genesis
func (k *Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	// Get current parameters
	params, err := k.GetParams(ctx)
	if err != nil {
		// Return default params if error
		params = types.DefaultParams()
	}

	// Note: We only export parameters, not runtime monitoring data.
	// Runtime data (transactions, alerts, metrics) is transient and
	// should not be persisted across chain restarts.

	return &types.GenesisState{
		Params: params,
	}
}
