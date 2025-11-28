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
	if gs.Params != nil {
		if err := k.SetParams(ctx, *gs.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Import trusted peers
	for _, peer := range gs.TrustedPeers {
		if peer != nil {
			if err := k.SetTrustedPeer(ctx, *peer); err != nil {
				k.logger.Error("failed to set trusted peer", "peer_id", peer.PeerId, "error", err)
				continue
			}
		}
	}

	// Import peer reputations
	for _, reputation := range gs.Reputations {
		if reputation != nil {
			if err := k.SetReputation(ctx, *reputation); err != nil {
				k.logger.Error("failed to set reputation", "peer_id", reputation.PeerId, "error", err)
				continue
			}
		}
	}

	// Import rate limit entries
	for _, entry := range gs.RateLimits {
		if entry != nil {
			if err := k.SetRateLimitEntry(ctx, *entry); err != nil {
				k.logger.Error("failed to set rate limit entry", "peer_id", entry.PeerId, "error", err)
				continue
			}
		}
	}

	// Import fork alerts
	for _, alert := range gs.ForkAlerts {
		if alert != nil {
			if err := k.SetForkAlert(ctx, *alert); err != nil {
				k.logger.Error("failed to set fork alert", "alert_id", alert.AlertId, "error", err)
				continue
			}
		}
	}

	// Import partition alerts
	for _, alert := range gs.PartitionAlerts {
		if alert != nil {
			if err := k.SetPartitionAlert(ctx, *alert); err != nil {
				k.logger.Error("failed to set partition alert", "alert_id", alert.AlertId, "error", err)
				continue
			}
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
	trustedPeers := make([]*types.TrustedPeer, len(trustedPeersVal))
	for i := range trustedPeersVal {
		trustedPeers[i] = &trustedPeersVal[i]
	}

	// Export reputations
	reputationsVal := k.GetAllReputations(ctx)
	reputations := make([]*types.NodeReputation, len(reputationsVal))
	for i := range reputationsVal {
		reputations[i] = &reputationsVal[i]
	}

	// Export rate limits
	rateLimitsVal := k.GetAllRateLimits(ctx)
	rateLimits := make([]*types.RateLimitEntry, len(rateLimitsVal))
	for i := range rateLimitsVal {
		rateLimits[i] = &rateLimitsVal[i]
	}

	// Export fork alerts (only active ones)
	forkAlertsVal := k.GetAllForkAlerts(ctx, false)
	forkAlerts := make([]*types.ForkAlert, len(forkAlertsVal))
	for i := range forkAlertsVal {
		forkAlerts[i] = &forkAlertsVal[i]
	}

	// Export partition alerts (only active ones)
	partitionAlertsVal := k.GetAllPartitionAlerts(ctx, false)
	partitionAlerts := make([]*types.PartitionAlert, len(partitionAlertsVal))
	for i := range partitionAlertsVal {
		partitionAlerts[i] = &partitionAlertsVal[i]
	}

	k.logger.Info("network security genesis state exported",
		"trusted_peers", len(trustedPeers),
		"reputations", len(reputations),
		"rate_limits", len(rateLimits),
		"fork_alerts", len(forkAlerts),
		"partition_alerts", len(partitionAlerts),
	)

	return &types.GenesisState{
		Params:          &params,
		TrustedPeers:    trustedPeers,
		Reputations:     reputations,
		RateLimits:      rateLimits,
		ForkAlerts:      forkAlerts,
		PartitionAlerts: partitionAlerts,
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

	var entries []types.RateLimitEntry
	for ; iterator.Valid(); iterator.Next() {
		var entry types.RateLimitEntry
		k.cdc.MustUnmarshal(iterator.Value(), &entry)
		entries = append(entries, entry)
	}
	return entries
}
