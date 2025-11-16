package networksecurity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/keeper"
	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// InitGenesis initializes the module's state from a genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs *types.GenesisState) {
	// Validate genesis state
	if err := gs.Validate(); err != nil {
		panic(err)
	}

	// Set params
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	// Set trusted peers
	for _, peer := range gs.TrustedPeers {
		if err := k.SetTrustedPeer(ctx, peer); err != nil {
			panic(err)
		}
	}

	// Set reputations
	for _, reputation := range gs.Reputations {
		if err := k.SetReputation(ctx, reputation); err != nil {
			panic(err)
		}
	}

	// Set rate limits
	for _, rateLimit := range gs.RateLimits {
		if err := k.SetRateLimitEntry(ctx, rateLimit); err != nil {
			panic(err)
		}
	}

	// Set fork alerts
	for _, alert := range gs.ForkAlerts {
		if err := k.SetForkAlert(ctx, alert); err != nil {
			panic(err)
		}
	}

	// Set partition alerts
	for _, alert := range gs.PartitionAlerts {
		if err := k.SetPartitionAlert(ctx, alert); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the module's state to a genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:          params,
		TrustedPeers:    k.GetAllTrustedPeers(ctx),
		Reputations:     k.GetAllReputations(ctx),
		RateLimits:      []types.RateLimitEntry{}, // Don't export rate limits
		ForkAlerts:      k.GetAllForkAlerts(ctx, true),
		PartitionAlerts: k.GetAllPartitionAlerts(ctx, true),
	}
}
