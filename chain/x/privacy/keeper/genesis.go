package keeper

import (
	"context"

	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// InitGenesis initializes the privacy module state from genesis
func (k *Keeper) InitGenesis(ctx context.Context, data *privacyproto.GenesisState) error {
	if data == nil {
		return nil
	}

	// Set parameters
	if data.Params != nil {
		if err := k.SetParams(ctx, data.Params); err != nil {
			k.Logger(ctx).Error("failed to set params", "error", err)
			return err
		}
	}

	k.Logger(ctx).Info("privacy module initialized from genesis")
	return nil
}

// ExportGenesis exports the privacy module state to genesis
func (k *Keeper) ExportGenesis(ctx context.Context) *privacyproto.GenesisState {
	// Get parameters
	params, err := k.GetParams(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to get params during export", "error", err)
		params = &privacyproto.Params{}
	}

	return &privacyproto.GenesisState{
		Params:             params,
		MixingPools:        []*privacyproto.MixingPool{},
		RegisteredViewKeys: []*privacyproto.ViewKey{},
	}
}
