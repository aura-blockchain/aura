// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/incidentresponse/types"
)

// InitGenesis initializes the module state from genesis
func (k *KeeperKV) InitGenesis(ctx context.Context, genesis types.GenesisState) error {
	k.requireStore()
	k.mu.Lock()
	defer k.mu.Unlock()

	// Validate genesis state
	if err := genesis.Validate(); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set params
	if genesis.Params != nil {
		if err := k.store.SetParams(ctx, genesis.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Load incidents
	for _, incident := range genesis.Incidents {
		if err := k.store.SetIncident(ctx, incident); err != nil {
			return fmt.Errorf("failed to set incident %s: %w", incident.ID, err)
		}
	}

	// Set pause state
	if genesis.PauseState != nil {
		if err := k.store.SetPauseState(ctx, genesis.PauseState); err != nil {
			return fmt.Errorf("failed to set pause state: %w", err)
		}
	}

	// Load wallet limits
	for _, limit := range genesis.WalletLimits {
		if err := k.store.SetWalletLimit(ctx, limit); err != nil {
			return fmt.Errorf("failed to set wallet limit for %s: %w", limit.Address, err)
		}
	}

	// Set next incident ID
	k.store.SetNextIncidentID(ctx, genesis.NextIncidentID)

	return nil
}

// ExportGenesis exports the module state to genesis
func (k *KeeperKV) ExportGenesis(ctx context.Context) types.GenesisState {
	k.requireStore()
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Get params
	params, err := k.store.GetParams(ctx)
	if err != nil {
		defaultParams := types.DefaultParams()
		params = defaultParams
	}

	// Get all incidents
	incidents := k.store.IterateIncidents(ctx)

	// Get pause state
	pauseState, ok := k.store.GetPauseState(ctx)
	if !ok {
		pauseState = &types.ChainPauseState{
			IsPaused:   false,
			PauseLevel: types.PauseLevelNone,
		}
	}

	// Get all wallet limits
	walletLimits := k.store.IterateWalletLimits(ctx)

	// Get next incident ID
	nextIncidentID := k.store.GetNextIncidentID(ctx)

	return types.GenesisState{
		Params:         &params,
		Incidents:      incidents,
		PauseState:     pauseState,
		WalletLimits:   walletLimits,
		NextIncidentID: nextIncidentID,
	}
}
