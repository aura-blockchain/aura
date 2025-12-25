// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// InitGenesis initializes the network security module's state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, gs *types.GenesisState) error {
	// Validate genesis state
	if err := types.ValidateGenesisState(gs); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set module parameters
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Import trusted peers
	for _, peer := range gs.TrustedPeers {
		if err := k.SetTrustedPeer(ctx, peer); err != nil {
			k.logger.Error("failed to set trusted peer", "peer_id", peer.PeerId, "error", err)
			continue
		}
	}

	// Import peer reputations
	for _, reputation := range gs.Reputations {
		if err := k.SetReputation(ctx, reputation); err != nil {
			k.logger.Error("failed to set reputation", "peer_id", reputation.PeerId, "error", err)
			continue
		}
	}

	// Import rate limit entries
	for _, entry := range gs.RateLimits {
		if err := k.SetRateLimitEntry(ctx, entry); err != nil {
			k.logger.Error("failed to set rate limit entry", "peer_id", entry.PeerId, "error", err)
			continue
		}
	}

	// Import fork alerts
	for _, alert := range gs.ForkAlerts {
		if err := k.SetForkAlert(ctx, alert); err != nil {
			k.logger.Error("failed to set fork alert", "alert_id", alert.AlertId, "error", err)
			continue
		}
	}

	// Import partition alerts
	for _, alert := range gs.PartitionAlerts {
		if err := k.SetPartitionAlert(ctx, alert); err != nil {
			k.logger.Error("failed to set partition alert", "alert_id", alert.AlertId, "error", err)
			continue
		}
	}

	k.logger.Info("network security genesis state initialized",
		"trusted_peers", len(gs.TrustedPeers),
		"reputations", len(gs.Reputations),
		"rate_limits", len(gs.RateLimits),
		"fork_alerts", len(gs.ForkAlerts),
		"partition_alerts", len(gs.PartitionAlerts),
	)

	return nil
}

// ExportGenesis exports the network security module's state for genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Get current parameters
	params, err := k.GetParams(ctx)
	if err != nil {
		k.logger.Error("failed to get params during genesis export", "error", err)
		params = *types.DefaultParams()
	}

	// Export trusted peers
	trustedPeersVal := k.GetAllTrustedPeers(ctx)

	// Export reputations
	reputationsVal := k.GetAllReputations(ctx)

	// Export rate limits
	rateLimitsVal := k.GetAllRateLimits(ctx)

	// Export fork alerts (only active ones)
	forkAlertsVal := k.GetAllForkAlerts(ctx, false)

	// Export partition alerts (only active ones)
	partitionAlertsVal := k.GetAllPartitionAlerts(ctx, false)

	k.logger.Info("network security genesis state exported",
		"trusted_peers", len(trustedPeersVal),
		"reputations", len(reputationsVal),
		"rate_limits", len(rateLimitsVal),
		"fork_alerts", len(forkAlertsVal),
		"partition_alerts", len(partitionAlertsVal),
	)

	return &types.GenesisState{
		Params:          params,
		TrustedPeers:    trustedPeersVal,
		Reputations:     reputationsVal,
		RateLimits:      rateLimitsVal,
		ForkAlerts:      forkAlertsVal,
		PartitionAlerts: partitionAlertsVal,
	}
}

// GetAllRateLimits retrieves all rate limit entries
func (k Keeper) GetAllRateLimits(ctx sdk.Context) []types.RateLimitEntry {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.RateLimitPrefix, storetypes.PrefixEndBytes(types.RateLimitPrefix))
	if err != nil {
		return []types.RateLimitEntry{}
	}
	defer iterator.Close()

	entries := make([]types.RateLimitEntry, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var entry types.RateLimitEntry
		if err := k.cdc.Unmarshal(iterator.Value(), &entry); err != nil {
			k.logger.Error("failed to unmarshal rate limit entry during iteration", "error", err)
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}
